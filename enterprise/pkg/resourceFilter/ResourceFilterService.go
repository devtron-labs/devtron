package resourceFilter

import (
	"errors"
	"fmt"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"strings"
	"time"
)

type ResourceFilterService interface {
	//CRUD methods
	ListFilters() ([]*FilterMetaDataBean, error)
	GetFilterById(id int) (*FilterRequestResponseBean, error)
	UpdateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterRequestResponseBean, error)
	CreateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterRequestResponseBean, error)
	DeleteFilter(userId int32, id int) error

	GetFiltersByScope(scope resourceQualifiers.Scope) ([]*FilterMetaDataBean, error)
	CheckForResource(filters []*FilterMetaDataBean, metadata ExpressionMetadata) (FilterState, map[int]FilterState, error)

	//filter evaluation audit
	CreateFilterEvaluationAudit(subjectType SubjectType, subjectIds []int, refType ReferenceType, refId int, filters []*FilterMetaDataBean, filterIdVsState map[int]FilterState) (*ResourceFilterEvaluationAudit, error)
	UpdateFilterEvaluationAuditRef(id int, refType ReferenceType, refId int) error
}

type ResourceFilterServiceImpl struct {
	logger                               *zap.SugaredLogger
	qualifyingMappingService             resourceQualifiers.QualifierMappingService
	resourceFilterRepository             ResourceFilterRepository
	resourceFilterEvaluator              ResourceFilterEvaluator
	appRepository                        appRepository.AppRepository
	teamRepository                       team.TeamRepository
	clusterRepository                    clusterRepository.ClusterRepository
	environmentRepository                clusterRepository.EnvironmentRepository
	ceLEvaluatorService                  CELEvaluatorService
	devtronResourceSearchableKeyService  devtronResource.DevtronResourceService
	resourceFilterEvaluationAuditService FilterEvaluationAuditService
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
	devtronResourceSearchableKeyService devtronResource.DevtronResourceService,
	resourceFilterEvaluationAuditService FilterEvaluationAuditService,
) *ResourceFilterServiceImpl {
	return &ResourceFilterServiceImpl{
		logger:                               logger,
		qualifyingMappingService:             qualifyingMappingService,
		resourceFilterRepository:             resourceFilterRepository,
		resourceFilterEvaluator:              resourceFilterEvaluator,
		appRepository:                        appRepository,
		teamRepository:                       teamRepository,
		clusterRepository:                    clusterRepository,
		environmentRepository:                environmentRepository,
		ceLEvaluatorService:                  ceLEvaluatorService,
		devtronResourceSearchableKeyService:  devtronResourceSearchableKeyService,
		resourceFilterEvaluationAuditService: resourceFilterEvaluationAuditService,
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
		resourceConditions, err := getResourceConditionFromJsonString(filter.ConditionExpression)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling ConditionExpression to Conditions", "err", err, "conditionExpression", filter.ConditionExpression, "filterId", filter.Id)
			return nil, err
		}
		filtersResp[i] = &FilterMetaDataBean{
			Id:           filter.Id,
			Description:  filter.Description,
			Name:         filter.Name,
			TargetObject: filter.TargetObject,
			Conditions:   resourceConditions,
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
			err = errors.New(FilterNotFound)
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
	if strings.Contains(filterRequest.Name, " ") {
		return nil, errors.New("spaces are not allowed in name")
	}
	appSelectorsExist := len(filterRequest.QualifierSelector.ApplicationSelectors) > 0
	envSelectorExist := len(filterRequest.QualifierSelector.EnvironmentSelectors) > 0

	if util.XORBool(appSelectorsExist, envSelectorExist) {
		return filterRequest, errors.New("invalid qualifier selectors")
	}

	//validating given condition expressions
	validateResp, errored := impl.ceLEvaluatorService.ValidateCELRequest(ValidateRequestResponse{Conditions: filterRequest.Conditions})
	if errored {
		filterRequest.Conditions = validateResp.Conditions
		impl.logger.Errorw("error in validating expression", "Conditions", validateResp.Conditions)
		return filterRequest, errors.New(InvalidExpressions)
	}
	//validation done

	//unique name validation
	filterNames, err := impl.resourceFilterRepository.GetDistinctFilterNames()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}

	for _, name := range filterNames {
		if name == filterRequest.Name {
			return nil, errors.New("filter already exists with this name")
		}
	}
	//unique name validation done

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
		Deleted:             pointer.Bool(false),
		TargetObject:        filterRequest.TargetObject,
		ConditionExpression: conditionExpression,
		AuditLog:            auditLog,
	}

	createdFilterDataBean, err := impl.resourceFilterRepository.CreateResourceFilter(tx, filterDataBean)
	if err != nil {
		impl.logger.Errorw("error in saving resourceFilter in db", "err", err, "resourceFilter", filterDataBean)
		return nil, err
	}

	if appSelectorsExist && envSelectorExist {
		err = impl.saveQualifierMappings(tx, userId, createdFilterDataBean.Id, filterRequest.QualifierSelector)
		if err != nil {
			impl.logger.Errorw("error in saveQualifierMappings", "err", err, "QualifierSelector", filterRequest.QualifierSelector)
			return nil, err
		}
	}

	err = impl.resourceFilterRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing CreateFilter", "err", err, "filterRequest", filterRequest)
		return nil, err
	}
	filterRequest.Id = createdFilterDataBean.Id
	return filterRequest, nil
}

func (impl *ResourceFilterServiceImpl) UpdateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterRequestResponseBean, error) {
	appSelectorsExist := len(filterRequest.QualifierSelector.ApplicationSelectors) > 0
	envSelectorExist := len(filterRequest.QualifierSelector.EnvironmentSelectors) > 0
	if util.XORBool(appSelectorsExist, envSelectorExist) {
		return filterRequest, errors.New("invalid qualifier selectors")
	}
	//validating given condition expressions
	validateResp, errored := impl.ceLEvaluatorService.ValidateCELRequest(ValidateRequestResponse{Conditions: filterRequest.Conditions})
	if errored {
		filterRequest.Conditions = validateResp.Conditions
		impl.logger.Errorw("error in validating expression", "Conditions", validateResp.Conditions)
		return filterRequest, errors.New(InvalidExpressions)
	}
	//validation done

	if strings.Contains(filterRequest.Name, " ") {
		return filterRequest, errors.New("spaces are not allowed in name")
	}

	//if mappings are edited delete all the existing mappings and create new mappings
	conditionExpression, err := getJsonStringFromResourceCondition(filterRequest.Conditions)
	if err != nil {
		impl.logger.Errorw("error in converting resourceFilterConditions to json string", "err", err, "resourceFilterConditions", filterRequest.Conditions)
		return filterRequest, err
	}

	tx, err := impl.resourceFilterRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "err", err)
		return filterRequest, err
	}
	defer impl.resourceFilterRepository.RollbackTx(tx)

	resourceFilter, err := impl.resourceFilterRepository.GetById(filterRequest.Id)
	if err != nil || resourceFilter == nil {
		if err == pg.ErrNoRows {
			return filterRequest, errors.New("filter with given id not found")
		}
		return filterRequest, err
	}

	//validate if update request have different name stored in db
	if resourceFilter.Name != filterRequest.Name {
		//unique name validation
		filterNames, err := impl.resourceFilterRepository.GetDistinctFilterNames()
		if err != nil && err != pg.ErrNoRows {
			return filterRequest, err
		}

		for _, name := range filterNames {
			if name == filterRequest.Name {
				return filterRequest, errors.New("filter already exists with this name")
			}
		}
		//unique name validation done
	}

	currentTime := time.Now()
	resourceFilter.UpdatedBy = userId
	resourceFilter.Name = filterRequest.Name
	resourceFilter.Description = filterRequest.Description
	resourceFilter.UpdatedOn = currentTime
	resourceFilter.Deleted = pointer.Bool(false)
	resourceFilter.TargetObject = filterRequest.TargetObject
	resourceFilter.ConditionExpression = conditionExpression
	err = impl.resourceFilterRepository.UpdateFilter(tx, resourceFilter)
	if err != nil {
		impl.logger.Errorw("error in updating filter", "resourceFilter", resourceFilter, "err", err)
		return filterRequest, err
	}
	err = impl.qualifyingMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.Filter, resourceFilter.Id, sql.AuditLog{UpdatedBy: userId, UpdatedOn: currentTime}, tx)
	if err != nil {
		impl.logger.Errorw("error in DeleteAllQualifierMappingsByResourceTypeAndId", "resourceType", resourceQualifiers.Filter, "resourceId", resourceFilter.Id, "err", err)
		return filterRequest, err
	}

	if appSelectorsExist && envSelectorExist {
		err = impl.saveQualifierMappings(tx, userId, resourceFilter.Id, filterRequest.QualifierSelector)
		if err != nil {
			impl.logger.Errorw("error in saveQualifierMappings for resourceFilter", "resourceFilterId", resourceFilter.Id, "qualifierMappings", filterRequest.QualifierSelector, "err", err)
			return filterRequest, err
		}
	}

	err = impl.resourceFilterRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err, "resourceFilterId", filterRequest.Id)
		return filterRequest, err
	}
	return filterRequest, nil
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
	resourceFilter.Deleted = pointer.Bool(true)
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
	err = impl.resourceFilterRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err, "resourceFilterId", id)
		return err
	}
	return nil
}

func (impl *ResourceFilterServiceImpl) CheckForResource(filters []*FilterMetaDataBean, metadata ExpressionMetadata) (FilterState, map[int]FilterState, error) {
	filterIdVsState := make(map[int]FilterState)
	finalState := ALLOW
	for _, filter := range filters {
		allowed, err := impl.resourceFilterEvaluator.EvaluateFilter(filter, metadata)
		if err != nil {
			finalState = ERROR
			filterIdVsState[filter.Id] = ERROR
			//return ERROR, nil
		}
		if !allowed {
			finalState = BLOCK
			filterIdVsState[filter.Id] = BLOCK
			//return BLOCK, nil
		}
	}
	return finalState, filterIdVsState, nil
}

func (impl *ResourceFilterServiceImpl) GetFiltersByScope(scope resourceQualifiers.Scope) ([]*FilterMetaDataBean, error) {
	// fetch all the qualifier mappings, club them by filterIds, check for each filter whether it is eligible or not, then fetch filter details
	var filters []*FilterMetaDataBean
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

// private methods
func (impl *ResourceFilterServiceImpl) validateOwnershipAndGetIdMaps(qualifierSelector QualifierSelector) (map[string]int, map[string]int, map[string]int, map[string]int, error) {
	teams := make([]string, 0)
	apps := make([]string, 0)
	envs := make([]string, 0)
	clusters := make([]string, 0)

	//stores requested app to team mappings
	appToTeamMap := make(map[string]string)
	//stores requested env to cluster mappings
	envToClusterMap := make(map[string]string)

	//stores name vs id maps
	teamsMap := make(map[string]int)
	appsMap := make(map[string]int)
	envsMap := make(map[string]int)
	clustersMap := make(map[string]int)
	for _, appSelector := range qualifierSelector.ApplicationSelectors {
		if appSelector.ProjectName != resourceQualifiers.AllProjectsValue {
			teams = append(teams, appSelector.ProjectName)
		}
		for _, app := range appSelector.Applications {
			apps = append(apps, app)
			appToTeamMap[app] = appSelector.ProjectName
		}
	}

	for _, envSelector := range qualifierSelector.EnvironmentSelectors {
		if envSelector.ClusterName != resourceQualifiers.AllExistingAndFutureProdEnvsValue && envSelector.ClusterName != resourceQualifiers.AllExistingAndFutureNonProdEnvsValue {
			clusters = append(clusters, envSelector.ClusterName)
		}
		for _, env := range envSelector.Environments {
			envs = append(envs, env)
			envToClusterMap[env] = envSelector.ClusterName
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
			actualTeamId := appObj.TeamId
			requestedTeamName := appToTeamMap[appObj.AppName]
			if actualTeamId != teamsMap[requestedTeamName] {
				return teamsMap, appsMap, clustersMap, envsMap, errors.New(fmt.Sprintf("invalid request payload,msg : app '%s' doesn't belong to project '%s'", appObj.AppName, requestedTeamName))
			}
			appsMap[appObj.AppName] = appObj.Id
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
			actualClusterId := envObj.ClusterId
			requestedClusterName := envToClusterMap[envObj.Name]
			if actualClusterId != clustersMap[requestedClusterName] {
				return teamsMap, appsMap, clustersMap, envsMap, errors.New(fmt.Sprintf("invalid request payload,msg : environment '%s' doesn't belong to cluster '%s'", envObj.Name, requestedClusterName))
			}
			envsMap[envObj.Name] = envObj.Id
		}
	}

	return teamsMap, appsMap, clustersMap, envsMap, nil
}

func (impl *ResourceFilterServiceImpl) saveQualifierMappings(tx *pg.Tx, userId int32, resourceFilterId int, qualifierSelector QualifierSelector) error {
	projectNameToIdMap, appNameToIdMap, clusterNameToIdMap, envNameToIdMap, err := impl.validateOwnershipAndGetIdMaps(qualifierSelector)
	if err != nil {
		impl.logger.Errorw("error in validateOwnershipAndGetIdMaps", "qualifierSelector", qualifierSelector, "err", err)
		return err
	}
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	qualifierMappings, err := qualifierSelector.BuildQualifierMappings(resourceFilterId, projectNameToIdMap, appNameToIdMap, clusterNameToIdMap, envNameToIdMap, searchableKeyNameIdMap, userId)
	if err != nil {
		impl.logger.Errorw("error in ExtractQualifierMappings", "qualifierSelector", qualifierSelector, "err", err)
		return err
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
		filterId := qualifierMapping.ResourceId
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
	envFilterQualifiers := impl.filterEnvQualifier(filterMappings)
	for _, envFilterQualifier := range envFilterQualifiers {
		if allowed := impl.isEnvAllowed(scope, envFilterQualifier); allowed {
			return allowed
		}
	}
	return false
}

func (impl *ResourceFilterServiceImpl) isEnvAllowed(scope resourceQualifiers.Scope, envFilterQualifier *resourceQualifiers.QualifierMapping) bool {
	var envAllowed bool
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	envIdentifierValueInt := envFilterQualifier.IdentifierValueInt
	if envFilterQualifier.IdentifierKey == GetIdentifierKey(ClusterIdentifier, searchableKeyNameIdMap) {
		envIdentifierScopeValue := resourceQualifiers.GetEnvIdentifierValue(scope)
		envAllowed = envIdentifierValueInt == envIdentifierScopeValue || envIdentifierValueInt == scope.ClusterId
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
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	appIdentifierValueInt := appFilterQualifier.IdentifierValueInt
	if appFilterQualifier.IdentifierKey == GetIdentifierKey(ProjectIdentifier, searchableKeyNameIdMap) {
		appAllowed = appIdentifierValueInt == resourceQualifiers.AllProjectsInt || appIdentifierValueInt == scope.ProjectId
	} else {
		// check for app identifier value
		appAllowed = appIdentifierValueInt == scope.AppId
	}
	return appAllowed
}

func (impl *ResourceFilterServiceImpl) filterAppQualifier(qualifierMappings []*resourceQualifiers.QualifierMapping) *resourceQualifiers.QualifierMapping {
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	for _, qualifierMapping := range qualifierMappings {
		identifierKey := qualifierMapping.IdentifierKey
		if identifierKey == GetIdentifierKey(ProjectIdentifier, searchableKeyNameIdMap) || identifierKey == GetIdentifierKey(AppIdentifier, searchableKeyNameIdMap) {
			return qualifierMapping
		}
	}
	return nil
}

func (impl *ResourceFilterServiceImpl) filterEnvQualifier(qualifierMappings []*resourceQualifiers.QualifierMapping) []*resourceQualifiers.QualifierMapping {
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	res := make([]*resourceQualifiers.QualifierMapping, 0)
	for _, qualifierMapping := range qualifierMappings {
		identifierKey := qualifierMapping.IdentifierKey
		if identifierKey == GetIdentifierKey(ClusterIdentifier, searchableKeyNameIdMap) || identifierKey == GetIdentifierKey(EnvironmentIdentifier, searchableKeyNameIdMap) {
			res = append(res, qualifierMapping)
		}
	}
	return res
}

func (impl *ResourceFilterServiceImpl) makeQualifierSelector(qualifierMappings []*resourceQualifiers.QualifierMapping) (QualifierSelector, error) {
	appSelectors, envSelectors := make([]ApplicationSelector, 0), make([]EnvironmentSelector, 0)
	appIds, envIds := make([]int, 0), make([]int, 0)
	resp := QualifierSelector{}
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	for _, qualifierMapping := range qualifierMappings {
		if qualifierMapping.IdentifierKey == GetIdentifierKey(ProjectIdentifier, searchableKeyNameIdMap) || qualifierMapping.IdentifierKey == GetIdentifierKey(AppIdentifier, searchableKeyNameIdMap) {
			appSelector := ApplicationSelector{}
			if qualifierMapping.IdentifierKey == GetIdentifierKey(ProjectIdentifier, searchableKeyNameIdMap) {
				appSelector.ProjectName = qualifierMapping.IdentifierValueString
				appSelector.Applications = make([]string, 0)
				appSelectors = append(appSelectors, appSelector)
			} else {
				appIds = append(appIds, qualifierMapping.IdentifierValueInt)
			}

		} else if qualifierMapping.IdentifierKey == GetIdentifierKey(ClusterIdentifier, searchableKeyNameIdMap) || qualifierMapping.IdentifierKey == GetIdentifierKey(EnvironmentIdentifier, searchableKeyNameIdMap) {
			if qualifierMapping.IdentifierKey == GetIdentifierKey(ClusterIdentifier, searchableKeyNameIdMap) {
				envSelector := EnvironmentSelector{}
				envSelector.ClusterName = qualifierMapping.IdentifierValueString
				envSelector.Environments = make([]string, 0)
				envSelectors = append(envSelectors, envSelector)
			} else {
				envIds = append(envIds, qualifierMapping.IdentifierValueInt)
			}
		}
	}

	appSelectors, envSelectors, err := impl.updateAppAndEnvSelectors(appSelectors, envSelectors, appIds, envIds)
	if err != nil {
		impl.logger.Errorw("error in fetching apps by appIds or envs by envIds", "err", err, "appIds", appIds)
		return resp, err
	}
	resp.ApplicationSelectors = appSelectors
	resp.EnvironmentSelectors = envSelectors
	return resp, nil
}

func (impl *ResourceFilterServiceImpl) updateAppAndEnvSelectors(appSelectors []ApplicationSelector, envSelectors []EnvironmentSelector, appIds []int, envIds []int) ([]ApplicationSelector, []EnvironmentSelector, error) {
	apps, envs, err := impl.fetchAppsAndEnvs(appIds, envIds)
	if err != nil {
		impl.logger.Errorw("error in fetching apps by appIds or envs by envIds", "err", err, "appIds", appIds)
		return nil, nil, err
	}
	if len(apps) > 0 {
		prev := 0
		appSelector := ApplicationSelector{
			ProjectName:  apps[0].Team.Name,
			Applications: []string{apps[0].AppName},
		}
		for i := 1; i < len(apps); i++ {
			app := apps[i]
			if apps[prev].TeamId != app.TeamId {
				appSelectors = append(appSelectors, appSelector)
				appSelector = ApplicationSelector{
					ProjectName:  app.Team.Name,
					Applications: []string{app.AppName},
				}
				prev = i
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
		for i := 1; i < len(envs); i++ {
			env := envs[i]
			if envs[prev].ClusterId != env.ClusterId {
				envSelectors = append(envSelectors, envSelector)
				envSelector = EnvironmentSelector{
					ClusterName:  env.Cluster.ClusterName,
					Environments: []string{env.Name},
				}
				prev = i
			} else {
				envSelector.Environments = append(envSelector.Environments, env.Name)
			}
		}
		envSelectors = append(envSelectors, envSelector)
	}

	return appSelectors, envSelectors, nil
}

func (impl *ResourceFilterServiceImpl) fetchAppsAndEnvs(appIds []int, envIds []int) ([]*appRepository.App, []*clusterRepository.Environment, error) {
	apps, err := impl.appRepository.FindAppAndProjectByIdsOrderByTeam(appIds)
	if err != nil {
		impl.logger.Errorw("error in fetching apps by appIds", "err", err, "appIds", appIds)
		return apps, nil, err
	}

	envs, err := impl.environmentRepository.FindByIdsOrderByCluster(envIds)
	if err != nil {
		impl.logger.Errorw("error in fetching envs by envIds", "err", err, "envIds", envIds)
		return apps, envs, err
	}
	return apps, envs, nil
}

func (impl *ResourceFilterServiceImpl) CreateFilterEvaluationAudit(subjectType SubjectType, subjectIds []int, refType ReferenceType, refId int, filters []*FilterMetaDataBean, filterIdVsState map[int]FilterState) (*ResourceFilterEvaluationAudit, error) {
	return impl.resourceFilterEvaluationAuditService.CreateFilterEvaluation(subjectType, subjectIds, refType, refId, filters, filterIdVsState)
}

func (impl *ResourceFilterServiceImpl) UpdateFilterEvaluationAuditRef(id int, refType ReferenceType, refId int) error {
	return impl.resourceFilterEvaluationAuditService.UpdateFilterEvaluationAuditRef(id, refType, refId)
}
