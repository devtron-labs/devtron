package resourceFilter

import (
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
	CreateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterRequestResponseBean, error)
	DeleteFilter(userId int32, id int) error

	//GetFiltersByScope
	GetFiltersByScope(scope resourceQualifiers.Scope) ([]*FilterRequestResponseBean, error)

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
	ceLEvaluatorService      CELEvaluatorService
}

func NewResourceFilterServiceImpl(logger *zap.SugaredLogger,
	qualifyingMappingService resourceQualifiers.QualifierMappingService,
	resourceFilterRepository ResourceFilterRepository,
	resourceFilterEvaluator ResourceFilterEvaluator,
	appRepository appRepository.AppRepository,
	teamRepository team.TeamRepository,
	clusterRepository clusterRepository.ClusterRepository,
	environmentRepository clusterRepository.EnvironmentRepository,
	ceLEvaluatorService CELEvaluatorService,
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
		ceLEvaluatorService:      ceLEvaluatorService,
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
	if id == 0 {
		return nil, errors.New("filter not found for given id")
	}
	filter, err := impl.resourceFilterRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching filter by id", "err", err, "filterId", id)
		if err == pg.ErrNoRows {
			err = errors.New("filter not found for given id")
		}
		return nil, err
	}
	resp, err := convertToFilterBean(filter)
	if err != nil {
		impl.logger.Errorw("error in convertToFilterBean", "err", err, "filter.ConditionExpression", filter.ConditionExpression)
		return nil, err
	}

	qualifierMappings, err := impl.qualifyingMappingService.GetQualifierMappingsForFilterById(id)
	if err != nil {
		impl.logger.Errorw("error in GetQualifierMappingsForFilterById", "err", err, "filterId", id)
		return nil, err
	}
	qualifierSelector, err := impl.makeQualifierSelector(qualifierMappings)
	if err != nil {
		impl.logger.Errorw("error in makeQualifierSelector", "error", err, "filterId", id)
		return nil, err
	}
	resp.QualifierSelector = qualifierSelector

	return resp, nil
}

func (impl *ResourceFilterServiceImpl) CreateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterRequestResponseBean, error) {
	if filterRequest == nil || len(filterRequest.QualifierSelector.EnvironmentSelectors) == 0 || len(filterRequest.QualifierSelector.ApplicationSelectors) == 0 {
		return nil, errors.New(AppAndEnvSelectorRequiredMessage)
	}

	//validating given condition expressions
	validateResp, errored := impl.ceLEvaluatorService.ValidateCELRequest(ValidateRequestResponse{Conditions: filterRequest.Conditions})
	if errored {
		filterRequest.Conditions = validateResp.Conditions
		impl.logger.Errorw("error in validating expression", "Conditions", validateResp.Conditions)
		return filterRequest, errors.New(InvalidExpressions)
	}
	//validation done
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
	return filterRequest, nil
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
		impl.logger.Errorw("error in updating filter", "resourceFilter", resourceFilter, "err", err)
		return err
	}
	err = impl.qualifyingMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.Filter, resourceFilter.Id, sql.AuditLog{UpdatedBy: userId, UpdatedOn: currentTime}, tx)
	if err != nil {
		impl.logger.Errorw("error in DeleteAllQualifierMappingsByResourceTypeAndId", "resourceType", resourceQualifiers.Filter, "resourceId", resourceFilter.Id, "err", err)
		return err
	}
	err = impl.saveQualifierMappings(tx, userId, resourceFilter.Id, filterRequest.QualifierSelector)
	if err != nil {
		impl.logger.Errorw("error in saveQualifierMappings for resourceFilter", "resourceFilterId", resourceFilter.Id, "qualifierMappings", filterRequest.QualifierSelector, "err", err)
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
		impl.logger.Errorw("error in UpdateFilter", "err", err, "resourceFilter", resourceFilter)
		return err
	}
	err = impl.qualifyingMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.Filter, id, sql.AuditLog{UpdatedBy: userId, UpdatedOn: currentTime}, tx)
	if err != nil {
		impl.logger.Errorw("error in DeleteAllQualifierMappingsByResourceTypeAndId", "err", err, "resourceType", resourceQualifiers.Filter, "resourceId", id)
		return err
	}
	return nil
}

func (impl *ResourceFilterServiceImpl) GetFiltersByScope(scope resourceQualifiers.Scope) ([]*FilterRequestResponseBean, error) {
	// fetch all the qualifier mappings, club them by filterIds, check for each filter whether it is eligible or not, then fetch filter details
	var filters []*FilterRequestResponseBean
	qualifierMappings, err := impl.qualifyingMappingService.GetQualifierMappingsForFilter(scope)
	if err != nil {
		return filters, err
	}
	eligibleFilterIds := impl.extractEligibleFilters(scope, qualifierMappings)
	resourceFilters, err := impl.resourceFilterRepository.GetByIds(eligibleFilterIds)
	if err != nil {
		return filters, err
	}
	filters, err = convertToResponseBeans(resourceFilters)
	if err != nil {
		impl.logger.Errorw("error occurred while converting db dtos to beans", "scope", scope, "err", err)
	}
	return filters, err
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
				impl.logger.Errorw("error in finding apps with appNames", "appNames", apps)
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
				impl.logger.Errorw("error in finding teams with teamNames", "teamNames", teams)
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
				impl.logger.Errorw("error in finding envs with envNames", "envNames", envs)
				err = errors.New("none of the selected environments are active")
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
				impl.logger.Errorw("error in finding clusters with clusterNames", "clusterNames", clusters)
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
	filters, err := impl.GetFiltersByScope(scope)
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
	projectNameToIdMap, appNameToIdMap, clusterNameToIdMap, envNameToIdMap, err := impl.getIdsMaps(qualifierSelector)
	if err != nil {
		impl.logger.Errorw("error in making name to id maps for apps,envs,projects,clusters", "qualifierSelector", qualifierSelector, "err", err)
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
	if err != nil {
		impl.logger.Errorw("error in CreateQualifierMappings", "qualifierMappings", qualifierMappings, "err", err)
	}
	return err
}

func (impl *ResourceFilterServiceImpl) convertToFilterMappings(qualifierMappings []*resourceQualifiers.QualifierMapping) map[int][]*resourceQualifiers.QualifierMapping {
	filterIdVsMappings := make(map[int][]*resourceQualifiers.QualifierMapping, 0)
	for _, qualifierMapping := range qualifierMappings {
		filterId := qualifierMapping.QualifierId
		filterMappings := filterIdVsMappings[filterId]
		filterMappings = append(filterMappings, qualifierMapping)
		filterIdVsMappings[filterId] = filterMappings
	}
	return filterIdVsMappings
}

func (impl *ResourceFilterServiceImpl) extractEligibleFilters(scope resourceQualifiers.Scope, qualifierMappings []*resourceQualifiers.QualifierMapping) []int {
	filterIdVsMappings := impl.convertToFilterMappings(qualifierMappings)
	var eligibleFilterIds []int
	for filterId, filterMappings := range filterIdVsMappings {
		eligible := impl.checkForFilterEligibility(scope, filterMappings)
		if eligible {
			eligibleFilterIds = append(eligibleFilterIds, filterId)
		}
	}
	return eligibleFilterIds
}

func (impl *ResourceFilterServiceImpl) checkForFilterEligibility(scope resourceQualifiers.Scope, filterMappings []*resourceQualifiers.QualifierMapping) bool {

	//club app qualifiers, shortlist project qualifiers or app qualifiers, if not found, return false
	appAllowed := impl.checkForAppQualifier(scope, filterMappings)
	// club env qualifiers, shortlist cluster qualifiers or env qualifiers, if not found, return false
	envAllowed := impl.checkForEnvQualifier(scope, filterMappings)

	eligible := appAllowed && envAllowed
	return eligible
}

func (impl *ResourceFilterServiceImpl) checkForEnvQualifier(scope resourceQualifiers.Scope, filterMappings []*resourceQualifiers.QualifierMapping) bool {
	var envAllowed bool
	var envFilterQualifier *resourceQualifiers.QualifierMapping
	envFilterQualifier = impl.filterEnvQualifier(filterMappings)
	if envFilterQualifier == nil {
		return envAllowed
	}
	envIdentifierValueInt := envFilterQualifier.IdentifierValueInt
	if envFilterQualifier.IdentifierKey == ClusterIdentifier {
		envAllowed = envIdentifierValueInt == AllExistingAndFutureProdEnvsInt ||
			envIdentifierValueInt == AllExistingAndFutureNonProdEnvsInt || envIdentifierValueInt == scope.ClusterId
	} else {
		// check for env identifier value
		envAllowed = envIdentifierValueInt == scope.EnvId
	}
	return envAllowed
}

func (impl *ResourceFilterServiceImpl) checkForAppQualifier(scope resourceQualifiers.Scope, filterMappings []*resourceQualifiers.QualifierMapping) bool {
	var appAllowed bool
	var appFilterQualifier *resourceQualifiers.QualifierMapping
	appFilterQualifier = impl.filterAppQualifier(filterMappings)
	if appFilterQualifier == nil {
		return appAllowed
	}
	appIdentifierValueInt := appFilterQualifier.IdentifierValueInt
	if appFilterQualifier.IdentifierKey == ProjectIdentifier {
		appAllowed = appIdentifierValueInt == AllProjectsInt || appIdentifierValueInt == scope.ProjectId
	} else {
		// check for app identifier value
		appAllowed = appIdentifierValueInt == scope.AppId
	}
	return appAllowed
}

func (impl *ResourceFilterServiceImpl) filterAppQualifier(qualifierMappings []*resourceQualifiers.QualifierMapping) *resourceQualifiers.QualifierMapping {
	for _, qualifierMapping := range qualifierMappings {
		identifierKey := qualifierMapping.IdentifierKey
		if identifierKey == ProjectIdentifier || identifierKey == AppIdentifier {
			return qualifierMapping
		}
	}
	return nil
}

func (impl *ResourceFilterServiceImpl) filterEnvQualifier(qualifierMappings []*resourceQualifiers.QualifierMapping) *resourceQualifiers.QualifierMapping {
	for _, qualifierMapping := range qualifierMappings {
		identifierKey := qualifierMapping.IdentifierKey
		if identifierKey == ClusterIdentifier || identifierKey == EnvironmentIdentifier {
			return qualifierMapping
		}
	}
	return nil
}

func (impl *ResourceFilterServiceImpl) makeQualifierSelector(qualifierMappings []*resourceQualifiers.QualifierMapping) (QualifierSelector, error) {
	appSelectors, envSelectors := make([]ApplicationSelector, 0), make([]EnvironmentSelector, 0)
	appIds, envIds := make([]int, 0), make([]int, 0)
	resp := QualifierSelector{}
	for _, qualifierMapping := range qualifierMappings {
		if qualifierMapping.IdentifierKey == ProjectIdentifier || qualifierMapping.IdentifierKey == AppIdentifier {
			appSelector := ApplicationSelector{}
			if qualifierMapping.IdentifierKey == ProjectIdentifier {
				appSelector.ProjectName = qualifierMapping.IdentifierValueString
				appSelector.Applications = make([]string, 0)
			} else {
				appIds = append(appIds, qualifierMapping.IdentifierValueInt)
			}

		} else if qualifierMapping.IdentifierKey == ClusterIdentifier || qualifierMapping.IdentifierKey == EnvironmentIdentifier {
			if qualifierMapping.IdentifierKey == ClusterIdentifier {
				envSelector := EnvironmentSelector{}
				envSelector.ClusterName = qualifierMapping.IdentifierValueString
				envSelector.Environments = make([]string, 0)
			} else {
				envIds = append(envIds, qualifierMapping.IdentifierValueInt)
			}
		}
	}

	apps, err := impl.appRepository.FindAppAndProjectByIdsOrderByTeam(appIds)
	if err != nil {
		impl.logger.Errorw("error in fetching apps by appIds", "err", err, "appIds", appIds)
		return resp, err
	}

	envs, err := impl.environmentRepository.FindByIdsOrderByCluster(envIds)
	if err != nil {
		impl.logger.Errorw("error in fetching envs by envIds", "err", err, "envIds", envIds)
		return resp, err
	}
	appSelectors, envSelectors = impl.appendAppAndEnvSelectors(appSelectors, envSelectors, apps, envs)
	resp.ApplicationSelectors = appSelectors
	resp.EnvironmentSelectors = envSelectors
	return resp, nil
}

func (impl *ResourceFilterServiceImpl) appendAppAndEnvSelectors(appSelectors []ApplicationSelector, envSelectors []EnvironmentSelector, apps []*appRepository.App, envs []*clusterRepository.Environment) ([]ApplicationSelector, []EnvironmentSelector) {
	if len(apps) > 0 {
		prev := 0
		appSelector := ApplicationSelector{
			ProjectName:  apps[0].Team.Name,
			Applications: []string{apps[0].AppName},
		}
		for _, app := range apps {
			if apps[prev].TeamId != app.TeamId {
				appSelectors = append(appSelectors, appSelector)
				appSelector = ApplicationSelector{
					ProjectName:  app.Team.Name,
					Applications: []string{app.AppName},
				}
			} else {
				appSelector.Applications = append(appSelector.Applications, app.AppName)
			}
		}
		appSelectors = append(appSelectors, appSelector)
	}

	if len(envs) > 0 {
		prev := 0
		envSelector := EnvironmentSelector{
			ClusterName:  envs[0].Cluster.ClusterName,
			Environments: []string{envs[0].Name},
		}
		for _, env := range envs {
			if envs[prev].ClusterId != env.ClusterId {
				envSelectors = append(envSelectors, envSelector)
				envSelector = EnvironmentSelector{
					ClusterName:  env.Cluster.ClusterName,
					Environments: []string{env.Name},
				}
			} else {
				envSelector.Environments = append(envSelector.Environments, env.Name)
			}
		}
		envSelectors = append(envSelectors, envSelector)
	}

	return appSelectors, envSelectors
}
