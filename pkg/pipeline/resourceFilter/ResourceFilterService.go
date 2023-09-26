package resourceFilter

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

const (
	NoResourceFiltersFound               = "no active resource filters found"
	AppAndEnvSelectorRequiredMessage     = "both application and environment selectors are required"
	InvalidExpressions                   = "one or more expressions are invalid"
	AllProjectsValue                     = "0"
	AllProjectsInt                       = 0
	AllExistingAndFutureProdEnvsValue    = "0"
	AllExistingAndFutureProdEnvsInt      = 0
	AllExistingAndFutureNonProdEnvsValue = "-1"
	AllExistingAndFutureNonProdEnvsInt   = -1
)

type IdentifierType int

const (
	ProjectIdentifier     = 0
	AppIdentifier         = 1
	ClusterIdentifier     = 2
	EnvironmentIdentifier = 3
)

type FilterResponseBean struct {
	Id           int                `json:"id"`
	TargetObject FilterTargetObject `json:"targetObject" validate:"required"`
	Description  string             `json:"description" `
	Name         string             `json:"name" validate:"required"`
}

type FilterRequestResponseBean struct {
	*FilterResponseBean
	Conditions        []ResourceCondition `json:"conditions"`
	QualifierSelector QualifierSelector   `json:"qualifierSelector"`
}

type ResourceCondition struct {
	ConditionType ResourceConditionType `json:"conditionType"`
	Expression    string                `json:"expression"`
	ErrorMsg      string                `json:"errorMsg"`
}

type QualifierSelector struct {
	ApplicationSelectors []ApplicationSelector `json:"applicationSelectors"`
	EnvironmentSelectors []EnvironmentSelector `json:"environmentSelectors"`
}

type ApplicationSelector struct {
	ProjectName  string   `json:"projectName"`
	Applications []string `json:"applications"`
}

type EnvironmentSelector struct {
	ClusterName  string   `json:"clusterName"`
	Environments []string `json:"environments"`
}

type ResourceFilterService interface {
	//CRUD methods
	ListFilters() ([]*FilterResponseBean, error)
	GetFilterById(id int) (*FilterRequestResponseBean, error)
	UpdateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterResponseBean, error)
	CreateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterResponseBean, error)
	DeleteFilter(userId int32, id int) error

	//GetFiltersByAppIdEnvId
	GetFiltersByAppIdEnvId(appId, envId int) ([]*FilterRequestResponseBean, error)
}

type ResourceFilterServiceImpl struct {
	logger                   *zap.SugaredLogger
	qualifyingMappingService resourceQualifiers.QualifierMappingService
	resourceFilterRepository ResourceFilterRepository
}

func NewScopedVariableServiceImpl(logger *zap.SugaredLogger,
	qualifyingMappingService resourceQualifiers.QualifierMappingService,
	resourceFilterRepository ResourceFilterRepository) *ResourceFilterServiceImpl {
	return &ResourceFilterServiceImpl{
		logger:                   logger,
		qualifyingMappingService: qualifyingMappingService,
		resourceFilterRepository: resourceFilterRepository,
	}
}

func (impl *ResourceFilterServiceImpl) ListFilters() ([]*FilterResponseBean, error) {
	filtersList, err := impl.resourceFilterRepository.ListAll()
	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Info(NoResourceFiltersFound)
			return nil, nil
		}
		impl.logger.Errorw("error in fetching all active filters", "err", err)
		return nil, err
	}
	filtersResp := make([]*FilterResponseBean, len(filtersList))
	for i, filter := range filtersList {
		filtersResp[i] = &FilterResponseBean{
			Id:          filter.Id,
			Description: filter.Description,
			Name:        filter.Name,
		}
	}
	return filtersResp, err
}

func (impl *ResourceFilterServiceImpl) GetFilterById(id int) (*FilterRequestResponseBean, error) {
	return nil, nil
}

func (impl *ResourceFilterServiceImpl) CreateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterResponseBean, error) {
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
	return filterRequest.FilterResponseBean, nil
}

func (impl *ResourceFilterServiceImpl) UpdateFilter(userId int32, filterRequest *FilterRequestResponseBean) (*FilterResponseBean, error) {
	//if mappings are edited delete all the existing mappings and create new mappings
	return nil, nil
}

func (impl *ResourceFilterServiceImpl) DeleteFilter(userId int32, id int) error {
	return nil
}

func (impl *ResourceFilterServiceImpl) GetFiltersByAppIdEnvId(appId, envId int) ([]*FilterRequestResponseBean, error) {
	return nil, nil
}

func (impl *ResourceFilterServiceImpl) saveQualifierMappings(tx *pg.Tx, userId int32, resourceFilterId int, qualifierSelector QualifierSelector) error {
	qualifierMappings := make([]*resourceQualifiers.QualifierMapping, 0)
	//TODO: build these maps
	projectNameToIdMap := make(map[string]int)
	appNameToIdMap := make(map[string]int)
	clusterNameToIdMap := make(map[string]int)
	envNameToIdMap := make(map[string]int)
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
	_, err := impl.qualifyingMappingService.CreateQualifierMappings(qualifierMappings, tx)
	return err
}

func getJsonStringFromResourceCondition(resourceConditions []ResourceCondition) (string, error) {

	jsonBytes, err := json.Marshal(resourceConditions)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
