package resourceFilter

import (
	"encoding/json"
	"errors"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ResourceFilterService interface {
	//CRUD methods
	ListFilters() ([]*FilterMetaDataBean, error)
	GetFilterById(id int) (*FilterRequestResponseBean, error)
	UpdateFilter(userId int32, filterRequest *FilterRequestResponseBean) error
	CreateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterMetaDataBean, error)
	DeleteFilter(userId int32, id int) error

	//GetFiltersByAppIdEnvId
	GetFiltersByAppIdEnvId(appId, envId int) ([]*FilterRequestResponseBean, error)

	CheckForResource(scope resourceQualifiers.Scope, metadata ExpressionMetadata) (FilterState, error)
}

type ResourceFilterServiceImpl struct {
	logger                   *zap.SugaredLogger
	qualifyingMappingService resourceQualifiers.QualifierMappingService
	resourceFilterRepository ResourceFilterRepository
	resourceFilterEvaluator  ResourceFilterEvaluator
	appRepository            appRepository.AppRepository
	teamRepository           team.TeamRepository
	clusterRepository        clusterRepository.ClusterRepository
	environmentRepository    clusterRepository.EnvironmentRepository
}

func NewResourceFilterServiceImpl(logger *zap.SugaredLogger,
	qualifyingMappingService resourceQualifiers.QualifierMappingService,
	resourceFilterRepository ResourceFilterRepository,
	resourceFilterEvaluator ResourceFilterEvaluator,
	appRepository appRepository.AppRepository,
	teamRepository team.TeamRepository,
	clusterRepository clusterRepository.ClusterRepository,
	environmentRepository clusterRepository.EnvironmentRepository,
) *ResourceFilterServiceImpl {
	return &ResourceFilterServiceImpl{
		logger:                   logger,
		qualifyingMappingService: qualifyingMappingService,
		resourceFilterRepository: resourceFilterRepository,
		resourceFilterEvaluator:  resourceFilterEvaluator,
		appRepository:            appRepository,
		teamRepository:           teamRepository,
		clusterRepository:        clusterRepository,
		environmentRepository:    environmentRepository,
	}
}

func (impl *ResourceFilterServiceImpl) ListFilters() ([]*FilterMetaDataBean, error) {
	filtersList, err := impl.resourceFilterRepository.ListAll()
	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Info(NoResourceFiltersFound)
			return nil, nil
		}
		impl.logger.Errorw("error in fetching all active filters", "err", err)
		return nil, err
	}
	filtersResp := make([]*FilterMetaDataBean, len(filtersList))
	for i, filter := range filtersList {
		filtersResp[i] = &FilterMetaDataBean{
			Id:           filter.Id,
			Description:  filter.Description,
			Name:         filter.Name,
			TargetObject: filter.TargetObject,
		}
	}
	return filtersResp, err
}

func (impl *ResourceFilterServiceImpl) GetFilterById(id int) (*FilterRequestResponseBean, error) {
	return nil, nil
}

func (impl *ResourceFilterServiceImpl) CreateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterMetaDataBean, error) {
	if filterRequest == nil || len(filterRequest.QualifierSelector.EnvironmentSelectors) == 0 || len(filterRequest.QualifierSelector.ApplicationSelectors) == 0 {
		return nil, errors.New(AppAndEnvSelectorRequiredMessage)
	}

	//TODO: evaluate filterRequest.Conditions
	tx, err := impl.resourceFilterRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting db transaction", "err", err)
		return nil, err
	}
	defer impl.resourceFilterRepository.RollbackTx(tx)
	currentTime := time.Now()
	auditLog := sql.AuditLog{
		CreatedOn: currentTime,
		UpdatedOn: currentTime,
		CreatedBy: userId,
		UpdatedBy: userId,
	}
	conditionExpression, err := getJsonStringFromResourceCondition(filterRequest.Conditions)
	if err != nil {
		impl.logger.Errorw("error in converting resourceFilterConditions to json string", "err", err, "resourceFilterConditions", filterRequest.Conditions)
		return nil, err
	}
	filterDataBean := &ResourceFilter{
		Name:                filterRequest.Name,
		Description:         filterRequest.Description,
		Deleted:             false,
		TargetObject:        filterRequest.TargetObject,
		ConditionExpression: conditionExpression,
		AuditLog:            auditLog,
	}

	createdFilterDataBean, err := impl.resourceFilterRepository.CreateResourceFilter(tx, filterDataBean)
	if err != nil {
		impl.logger.Errorw("error in saving resourceFilter in db", "err", err, "resourceFilter", filterDataBean)
		return nil, err
	}

	err = impl.saveQualifierMappings(tx, userId, createdFilterDataBean.Id, filterRequest.QualifierSelector)
	if err != nil {
		impl.logger.Errorw("error in saveQualifierMappings", "err", err, "QualifierSelector", filterRequest.QualifierSelector)
		return nil, err
	}

	err = impl.resourceFilterRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing CreateFilter", "err", err, "filterRequest", filterRequest)
		return nil, err
	}
	filterRequest.Id = createdFilterDataBean.Id
	return filterRequest.FilterMetaDataBean, nil
}

func (impl *ResourceFilterServiceImpl) UpdateFilter(userId int32, filterRequest *FilterRequestResponseBean) error {
	//if mappings are edited delete all the existing mappings and create new mappings
	conditionExpression, err := getJsonStringFromResourceCondition(filterRequest.Conditions)
	if err != nil {
		impl.logger.Errorw("error in converting resourceFilterConditions to json string", "err", err, "resourceFilterConditions", filterRequest.Conditions)
		return err
	}

	tx, err := impl.resourceFilterRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "err", err)
		return err
	}
	defer impl.resourceFilterRepository.RollbackTx(tx)

	resourceFilter, err := impl.resourceFilterRepository.GetById(filterRequest.Id)
	if err != nil || resourceFilter == nil {
		if err == pg.ErrNoRows {
			return errors.New("filter with given id not found")
		}
		return err
	}
	currentTime := time.Now()
	resourceFilter.UpdatedBy = userId
	resourceFilter.Name = filterRequest.Name
	resourceFilter.Description = filterRequest.Description
	resourceFilter.UpdatedOn = currentTime
	resourceFilter.Deleted = false
	resourceFilter.TargetObject = filterRequest.TargetObject
	resourceFilter.ConditionExpression = conditionExpression
	err = impl.resourceFilterRepository.UpdateFilter(tx, resourceFilter)
	if err != nil {
		//TODO: add error log
		return err
	}
	err = impl.qualifyingMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.Filter, resourceFilter.Id, sql.AuditLog{UpdatedBy: userId, UpdatedOn: currentTime}, tx)
	if err != nil {
		//TODO: add error log
		return err
	}
	err = impl.saveQualifierMappings(tx, userId, resourceFilter.Id, filterRequest.QualifierSelector)
	if err != nil {
		//TODO: add error log
		return err
	}
	return nil
}

func (impl *ResourceFilterServiceImpl) DeleteFilter(userId int32, id int) error {
	tx, err := impl.resourceFilterRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "err", err)
		return err
	}
	defer impl.resourceFilterRepository.RollbackTx(tx)

	resourceFilter, err := impl.resourceFilterRepository.GetById(id)
	if err != nil || resourceFilter == nil {
		if err == pg.ErrNoRows {
			return errors.New("filter with given id not found")
		}
		return err
	}
	currentTime := time.Now()
	resourceFilter.UpdatedBy = userId
	resourceFilter.UpdatedOn = currentTime
	resourceFilter.Deleted = true
	err = impl.resourceFilterRepository.UpdateFilter(tx, resourceFilter)
	if err != nil {
		//TODO: log error
		return err
	}
	err = impl.qualifyingMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.Filter, id, sql.AuditLog{UpdatedBy: userId, UpdatedOn: currentTime}, tx)
	if err != nil {
		//TODO: log error
		return err
	}
	return nil
}

func (impl *ResourceFilterServiceImpl) GetFiltersByAppIdEnvId(appId, envId int) ([]*FilterRequestResponseBean, error) {
	return nil, nil
}

func (impl *ResourceFilterServiceImpl) getIdsMaps(qualifierSelector QualifierSelector) (map[string]int, map[string]int, map[string]int, map[string]int, error) {
	teams := make([]string, 0)
	apps := make([]string, 0)
	envs := make([]string, 0)
	clusters := make([]string, 0)

	teamsMap := make(map[string]int)
	appsMap := make(map[string]int)
	envsMap := make(map[string]int)
	clustersMap := make(map[string]int)
	for _, appSelector := range qualifierSelector.ApplicationSelectors {
		if appSelector.ProjectName != AllProjectsValue {
			teams = append(teams, appSelector.ProjectName)
		}
		for _, app := range appSelector.Applications {
			apps = append(apps, app)
		}
	}

	for _, envSelector := range qualifierSelector.EnvironmentSelectors {
		if envSelector.ClusterName != AllExistingAndFutureProdEnvsValue && envSelector.ClusterName != AllExistingAndFutureNonProdEnvsValue {
			clusters = append(clusters, envSelector.ClusterName)
		}
		for _, env := range envSelector.Environments {
			envs = append(envs, env)
		}
	}

	if len(apps) > 0 {
		appObjs, err := impl.appRepository.FindByNames(apps)
		if err != nil {
			if err == pg.ErrNoRows {
				//TODO: log error
				err = errors.New("none of the selected apps are active")
				return teamsMap, appsMap, clustersMap, envsMap, err
			}
		}

		for _, appObj := range appObjs {
			appsMap[appObj.AppName] = appObj.Id
		}
	}

	if len(teams) > 0 {
		teamObjs, err := impl.teamRepository.FindByNames(teams)
		if err != nil {
			if err == pg.ErrNoRows {
				//TODO: log error
				err = errors.New("none of the selected projects are active")
				return teamsMap, appsMap, clustersMap, envsMap, err
			}
		}
		for _, teamObj := range teamObjs {
			teamsMap[teamObj.Name] = teamObj.Id
		}
	}

	if len(envs) > 0 {
		envObjs, err := impl.environmentRepository.FindByNames(envs)
		if err != nil {
			if err == pg.ErrNoRows {
				//TODO: log error
				err = errors.New("none of the apps selected environments are active")
				return teamsMap, appsMap, clustersMap, envsMap, err
			}
		}
		for _, envObj := range envObjs {
			envsMap[envObj.Name] = envObj.Id
		}
	}

	if len(clusters) > 0 {
		clusterObjs, err := impl.clusterRepository.FindByNames(clusters)
		if err != nil {
			if err == pg.ErrNoRows {
				//TODO: log error
				err = errors.New("none of the selected clusters are active")
				return teamsMap, appsMap, clustersMap, envsMap, err
			}
		}
		for _, clusterObj := range clusterObjs {
			clustersMap[clusterObj.ClusterName] = clusterObj.Id
		}
	}
	return teamsMap, appsMap, clustersMap, envsMap, nil
}

func (impl *ResourceFilterServiceImpl) CheckForResource(scope resourceQualifiers.Scope, metadata ExpressionMetadata) (FilterState, error) {
	// fetch filters for given scope, use FilterEvaluator.Evaluate to check for access
	filters, err := impl.GetFiltersByAppIdEnvId(scope.AppId, scope.EnvId)
	if err != nil {
		return ERROR, err
	}
	for _, filter := range filters {
		allowed, err := impl.resourceFilterEvaluator.EvaluateFilter(filter, metadata)
		if err != nil {
			return ERROR, nil
		}
		if !allowed {
			return BLOCK, nil
		}
	}
	return ALLOW, nil
}

func (impl *ResourceFilterServiceImpl) saveQualifierMappings(tx *pg.Tx, userId int32, resourceFilterId int, qualifierSelector QualifierSelector) error {
	qualifierMappings := make([]*resourceQualifiers.QualifierMapping, 0)
	//TODO: build these maps
	projectNameToIdMap, appNameToIdMap, clusterNameToIdMap, envNameToIdMap, err := impl.getIdsMaps(qualifierSelector)
	if err != nil {
		//TODO: log error
		return err
	}
	//apps
	currentTime := time.Now()
	auditLog := sql.AuditLog{
		CreatedOn: currentTime,
		UpdatedOn: currentTime,
		CreatedBy: userId,
		UpdatedBy: userId,
	}
	//case-1) all existing and future applications -> will get empty ApplicationSelector , db entry (proj,0,"0")
	if len(qualifierSelector.ApplicationSelectors) == 1 && qualifierSelector.ApplicationSelectors[0].ProjectName == AllProjectsValue {
		allExistingAndFutureAppsQualifierMapping := &resourceQualifiers.QualifierMapping{
			ResourceId:   resourceFilterId,
			ResourceType: resourceQualifiers.Filter,

			//qualifierId: get qualifierId
			//IdentifierKey: get identifier key for proj
			IdentifierKey:         ProjectIdentifier,
			Active:                true,
			IdentifierValueInt:    AllProjectsInt,
			IdentifierValueString: AllProjectsValue,
			AuditLog:              auditLog,
		}
		qualifierMappings = append(qualifierMappings, allExistingAndFutureAppsQualifierMapping)
	} else {

		for _, appSelector := range qualifierSelector.ApplicationSelectors {
			//case-2) all existing and future apps in a project ->  will get projectName and empty applications array
			if len(appSelector.Applications) == 0 {
				allExistingAppsQualifierMapping := &resourceQualifiers.QualifierMapping{
					ResourceId:            resourceFilterId,
					ResourceType:          resourceQualifiers.Filter,
					IdentifierKey:         ProjectIdentifier,
					Active:                true,
					IdentifierValueInt:    projectNameToIdMap[appSelector.ProjectName],
					IdentifierValueString: appSelector.ProjectName,
					AuditLog:              auditLog,
				}
				qualifierMappings = append(qualifierMappings, allExistingAppsQualifierMapping)
			}
			//case-3) all existing applications -> will get all apps in payload
			//case-4) particular apps -> will get ApplicationSelectors array
			//case-5) all existing apps in a project -> will get projectName and all applications array
			for _, appName := range appSelector.Applications {
				appQualifierMapping := &resourceQualifiers.QualifierMapping{
					ResourceId:            resourceFilterId,
					ResourceType:          resourceQualifiers.Filter,
					IdentifierKey:         AppIdentifier,
					Active:                true,
					IdentifierValueInt:    appNameToIdMap[appName],
					IdentifierValueString: appName,
					AuditLog:              auditLog,
				}
				qualifierMappings = append(qualifierMappings, appQualifierMapping)
			}
		}
	}

	//envs
	//1) all existing and future prod envs -> get single EnvironmentSelector with clusterName as "0"(prod) (cluster,0,"0")
	//2) all existing and future non-prod envs -> get single EnvironmentSelector with clusterName as "-1"(non-prod) (cluster,-1,"-1")
	if len(qualifierSelector.ApplicationSelectors) == 1 && (qualifierSelector.EnvironmentSelectors[0].ClusterName == AllExistingAndFutureProdEnvsValue || qualifierSelector.EnvironmentSelectors[0].ClusterName == AllExistingAndFutureNonProdEnvsValue) {
		envSelector := qualifierSelector.EnvironmentSelectors[0]
		allExistingAndFutureEnvQualifierMapping := &resourceQualifiers.QualifierMapping{
			ResourceId:    resourceFilterId,
			ResourceType:  resourceQualifiers.Filter,
			IdentifierKey: ClusterIdentifier,
			Active:        true,
			AuditLog:      auditLog,
		}
		if envSelector.ClusterName == AllExistingAndFutureProdEnvsValue {
			allExistingAndFutureEnvQualifierMapping.IdentifierValueInt = AllExistingAndFutureProdEnvsInt
			allExistingAndFutureEnvQualifierMapping.IdentifierValueString = AllExistingAndFutureProdEnvsValue
		} else {
			allExistingAndFutureEnvQualifierMapping.IdentifierValueInt = AllExistingAndFutureNonProdEnvsInt
			allExistingAndFutureEnvQualifierMapping.IdentifierValueString = AllExistingAndFutureNonProdEnvsValue
		}
		qualifierMappings = append(qualifierMappings, allExistingAndFutureEnvQualifierMapping)
	} else {
		for _, envSelector := range qualifierSelector.EnvironmentSelectors {
			//3) all existing and future envs of a cluster ->  get clusterName and empty environments list (cluster,clusterId,clusterName)
			if len(envSelector.Environments) == 0 {
				allCurrentAndFutureEnvsInClusterQualifierMapping := &resourceQualifiers.QualifierMapping{
					ResourceId:            resourceFilterId,
					ResourceType:          resourceQualifiers.Filter,
					IdentifierKey:         ClusterIdentifier,
					IdentifierValueInt:    clusterNameToIdMap[envSelector.ClusterName],
					IdentifierValueString: envSelector.ClusterName,
					Active:                true,
					AuditLog:              auditLog,
				}
				qualifierMappings = append(qualifierMappings, allCurrentAndFutureEnvsInClusterQualifierMapping)
			}
			//4) all existing envs of a cluster -> get clusterName and all the envs list
			//5) particular envs , will get EnvironmentSelector array
			for _, env := range envSelector.Environments {
				envQualifierMapping := &resourceQualifiers.QualifierMapping{
					ResourceId:            resourceFilterId,
					ResourceType:          resourceQualifiers.Filter,
					IdentifierKey:         EnvironmentIdentifier,
					IdentifierValueInt:    envNameToIdMap[env],
					IdentifierValueString: env,
					Active:                true,
					AuditLog:              auditLog,
				}
				qualifierMappings = append(qualifierMappings, envQualifierMapping)
			}
		}
	}
	_, err = impl.qualifyingMappingService.CreateQualifierMappings(qualifierMappings, tx)
	return err
}

func getJsonStringFromResourceCondition(resourceConditions []ResourceCondition) (string, error) {

	jsonBytes, err := json.Marshal(resourceConditions)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
