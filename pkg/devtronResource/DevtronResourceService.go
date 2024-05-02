package devtronResource

import (
	"context"
	"encoding/json"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	helper2 "github.com/devtron-labs/devtron/internal/sql/repository/helper"
	read2 "github.com/devtron-labs/devtron/pkg/appWorkflow/read"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/audit"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/in"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/devtronResource/release"
	"github.com/devtron-labs/devtron/util/response/pagination"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slices"
	"math"
	"net/http"
	"strings"
	"time"

	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

type DevtronResourceService interface {
	GetDevtronResourceList(onlyIsExposed bool) ([]*bean.DevtronResourceBean, error)
	// ListResourceObjectByKindAndVersion will list out all the resource objects by kind, subKind and version
	//
	// Query Flag:
	//
	// 1. isLite
	//    - true for lightweight data // provides the bean.DevtronResourceObjectDescriptorBean only
	//    - false for detailed data   // provides the complete bean.DevtronResourceObjectBasicDataBean
	// 2. fetchChild
	//    - true to resource data along with children data 	// includes the ChildObjects also
	//    - false for resource data only   					// doesn't include the ChildObjects
	// 3. filterCondition
	//    - supports filtering on resource object list b=in format resource|type|comma seperated values
	//    - release-track|id|12,13,14
	ListResourceObjectByKindAndVersion(kind, subKind, version string, isLite, fetchChild bool, filterCriteria []string) (pagination.PaginatedResponse[*bean.DevtronResourceObjectGetAPIBean], error)
	// GetResourceObject will get the bean.DevtronResourceObjectBean based on the given bean.DevtronResourceObjectDescriptorBean
	GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectGetAPIBean, error)
	// CreateResourceObject creates resource object corresponding to kind,version according to bean.DevtronResourceObjectBean
	CreateResourceObject(ctx context.Context, reqBean *bean.DtResourceObjectCreateReqBean) error
	// CreateOrUpdateResourceObject is only used for catalog
	CreateOrUpdateResourceObject(ctx context.Context, reqBean *bean.DtResourceObjectCatalogReqBean) error
	// PatchResourceObject supports json patch operation corresponding to kind,subKind,version on json object data takes in ([]PatchQuery in DevtronResourceObjectBean), returns error if any
	PatchResourceObject(ctx context.Context, req *bean.DtResourceObjectPatchReqBean) (*bean.SuccessResponse, error)
	// DeleteResourceObject deletes resource object corresponding to kind,version, id or name
	DeleteResourceObject(ctx context.Context, req *bean.DevtronResourceObjectDescriptorBean) (*bean.SuccessResponse, error)
	// GetResourceDependencies will get the bean.DevtronResourceObjectBean based on the given bean.DevtronResourceObjectDescriptorBean
	// It provides the dependencies and child dependencies []bean.DevtronResourceDependencyBean
	GetResourceDependencies(req *bean.DevtronResourceObjectDescriptorBean, query *apiBean.GetDependencyQueryParams) (*bean.DevtronResourceObjectBean, error)
	GetDependencyConfigOptions(req *bean.DevtronResourceObjectDescriptorBean, query *apiBean.GetConfigOptionsQueryParams) (any, error)
	CreateOrUpdateResourceDependencies(ctx context.Context, req *bean.DtResourceObjectDependenciesReqBean) error
	// PatchResourceDependencies supports json patch operation corresponding to kind,subKind,version on json object data takes in ([]PatchQuery in DevtronResourceObjectBean), returns error if any
	PatchResourceDependencies(ctx context.Context, req *bean.DevtronResourceDependencyPatchAPIBean) (*bean.SuccessResponse, error)

	FindNumberOfApplicationsWithDependenciesMapped() (int, error)

	SetDevtronResourcesAndSchemaMap() error
}

type DevtronResourceServiceImpl struct {
	logger                               *zap.SugaredLogger
	devtronResourceRepository            repository.DevtronResourceRepository
	devtronResourceSchemaRepository      repository.DevtronResourceSchemaRepository
	devtronResourceObjectRepository      repository.DevtronResourceObjectRepository
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository
	appRepository                        appRepository.AppRepository //TODO: remove repo dependency
	pipelineRepository                   pipelineConfig.PipelineRepository
	userRepository                       repository2.UserRepository
	appListingRepository                 repository3.AppListingRepository
	ciArtifactRepository                 repository3.CiArtifactRepository
	clusterRepository                    clusterRepository.ClusterRepository
	devtronResourcesMapById              map[int]*repository.DevtronResource       //map of id and its object
	devtronResourcesMapByKind            map[string]*repository.DevtronResource    //map of kind and its object
	devtronResourcesSchemaMapById        map[int]*repository.DevtronResourceSchema //map of id and its object
	dtResourceInternalProcessingService  in.InternalProcessingService
	dtResourceReadService                read.ReadService
	dtResourceObjectAuditService         audit.ObjectAuditService
	appArtifactManager                   pipeline.AppArtifactManager
	appWorkflowDataReadService           read2.AppWorkflowDataReadService
	releasePolicyEvaluationService       release.PolicyEvaluationService
}

func NewDevtronResourceServiceImpl(logger *zap.SugaredLogger,
	devtronResourceRepository repository.DevtronResourceRepository,
	devtronResourceSchemaRepository repository.DevtronResourceSchemaRepository,
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository,
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository,
	appRepository appRepository.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appListingRepository repository3.AppListingRepository,
	userRepository repository2.UserRepository,
	ciArtifactRepository repository3.CiArtifactRepository,
	clusterRepository clusterRepository.ClusterRepository,
	dtResourceInternalProcessingService in.InternalProcessingService,
	dtResourceReadService read.ReadService,
	dtResourceObjectAuditService audit.ObjectAuditService,
	appArtifactManager pipeline.AppArtifactManager,
	appWorkflowDataReadService read2.AppWorkflowDataReadService,
	releasePolicyEvaluationService release.PolicyEvaluationService) (*DevtronResourceServiceImpl, error) {
	impl := &DevtronResourceServiceImpl{
		logger:                               logger,
		devtronResourceRepository:            devtronResourceRepository,
		devtronResourceSchemaRepository:      devtronResourceSchemaRepository,
		devtronResourceObjectRepository:      devtronResourceObjectRepository,
		devtronResourceObjectAuditRepository: devtronResourceObjectAuditRepository,
		appRepository:                        appRepository,
		pipelineRepository:                   pipelineRepository,
		userRepository:                       userRepository,
		ciArtifactRepository:                 ciArtifactRepository,
		appListingRepository:                 appListingRepository,
		clusterRepository:                    clusterRepository,
		dtResourceInternalProcessingService:  dtResourceInternalProcessingService,
		dtResourceReadService:                dtResourceReadService,
		dtResourceObjectAuditService:         dtResourceObjectAuditService,
		appArtifactManager:                   appArtifactManager,
		appWorkflowDataReadService:           appWorkflowDataReadService,
		releasePolicyEvaluationService:       releasePolicyEvaluationService,
	}
	err := impl.SetDevtronResourcesAndSchemaMap()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

func (impl *DevtronResourceServiceImpl) SetDevtronResourcesAndSchemaMap() error {
	devtronResources, err := impl.devtronResourceRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error in getting devtron resources, NewDevtronResourceServiceImpl", "err", err)
		return err
	}
	devtronResourcesMap := make(map[int]*repository.DevtronResource)
	devtronResourcesMapByKind := make(map[string]*repository.DevtronResource)
	for _, devtronResource := range devtronResources {
		devtronResourcesMap[devtronResource.Id] = devtronResource
		devtronResourcesMapByKind[devtronResource.Kind] = devtronResource
	}
	devtronResourceSchemas, err := impl.devtronResourceSchemaRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error in getting devtron resource schemas, NewDevtronResourceServiceImpl", "err", err)
		return err
	}
	devtronResourceSchemasMap := make(map[int]*repository.DevtronResourceSchema)
	for _, devtronResourceSchema := range devtronResourceSchemas {
		devtronResourceSchemasMap[devtronResourceSchema.Id] = devtronResourceSchema
	}
	impl.devtronResourcesMapById = devtronResourcesMap
	impl.devtronResourcesMapByKind = devtronResourcesMapByKind
	impl.devtronResourcesSchemaMapById = devtronResourceSchemasMap
	return nil
}

//
//Get resource list and related method starts

func (impl *DevtronResourceServiceImpl) GetDevtronResourceList(onlyIsExposed bool) ([]*bean.DevtronResourceBean, error) {
	//getting all resource details from cache only as resource crud is not available as of now
	devtronResourceSchemas := impl.devtronResourcesSchemaMapById
	devtronResources := impl.devtronResourcesMapById
	response := make([]*bean.DevtronResourceBean, 0, len(devtronResources))
	resourceIdAndObjectIndexMap := make(map[int]int, len(devtronResources))
	i := 0
	for _, devtronResource := range devtronResources {
		if onlyIsExposed && !devtronResource.IsExposed {
			continue
		}
		response = append(response, &bean.DevtronResourceBean{
			DevtronResourceId: devtronResource.Id,
			Kind:              devtronResource.Kind,
			DisplayName:       devtronResource.DisplayName,
			Description:       devtronResource.Description,
			LastUpdatedOn:     devtronResource.UpdatedOn,
		})
		resourceIdAndObjectIndexMap[devtronResource.Id] = i
		i++
	}
	for _, devtronResourceSchema := range devtronResourceSchemas {
		//getting index where resource of this schema is present
		index := resourceIdAndObjectIndexMap[devtronResourceSchema.DevtronResourceId]
		response[index].VersionSchemaDetails = append(response[index].VersionSchemaDetails, &bean.DevtronResourceSchemaBean{
			DevtronResourceSchemaId: devtronResourceSchema.Id,
			Version:                 devtronResourceSchema.Version,
		})
	}
	return response, nil
}

//Get resource list and related method ends
//

//
//list resource object and related method starts

func (impl *DevtronResourceServiceImpl) ListResourceObjectByKindAndVersion(kind, subKind, version string, isLite, fetchChild bool, filterCriteria []string) (pagination.PaginatedResponse[*bean.DevtronResourceObjectGetAPIBean], error) {
	response := pagination.NewPaginatedResponse[*bean.DevtronResourceObjectGetAPIBean]()
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(kind, subKind, version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "kind", kind, "subKind", subKind, "version", version)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return response, err
	}
	resourceObjects, err := impl.devtronResourceObjectRepository.GetAllWithSchemaId(resourceSchema.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting objects by resourceSchemaId", "err", err, "resourceSchemaId", resourceSchema.Id)
		return response, err
	}

	if len(filterCriteria) > 0 {
		// getting func for applying filter criteria on all resource objects based on kind
		f := getFuncToApplyFilterResourceKind(kind, subKind, version)
		if f == nil {
			return response, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
		}
		//applying filters and getting filtered resource objects
		resourceObjects, err = f(impl, kind, subKind, version, resourceObjects, filterCriteria)
		if err != nil {
			impl.logger.Errorw("error in decodeFilterCriteria", "filterCriteria", filterCriteria, "err", err)
			return response, err
		}
	}
	// filters application ends resource objects are filtered.
	var childResourceObjects []*repository.DevtronResourceObject
	resourceObjectIndexChildMap := make(map[int][]int)
	if fetchChild {
		childResourceObjects, resourceObjectIndexChildMap, err = impl.fetchChildObjectsAndIndexMapForMultipleObjects(resourceObjects)
		if err != nil {
			impl.logger.Errorw("error, fetchChildObjectsAndIndexMapForMultipleObjects", "err", err, "kind", kind, "subKind", subKind, "version", version)
			return response, err
		}
	}
	response.UpdateTotalCount(len(resourceObjects))
	response.UpdateOffset(0)
	response.UpdateSize(len(resourceObjects))

	f1 := getFuncToListApiResourceKind(kind)
	if f1 == nil {
		impl.logger.Errorw("error kind type not supported", "err", err, "kind", kind)
		return response, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrComponent, bean.InvalidResourceKindOrComponent)
	}
	response.Data, err = f1(impl, resourceObjects, childResourceObjects, resourceObjectIndexChildMap, isLite)
	if err != nil {
		impl.logger.Errorw("error in getting list response", "err", err, "kind", kind, "subKind", subKind, "version", version)
		return response, err
	}
	return response, nil
}

func (impl *DevtronResourceServiceImpl) fetchChildObjectsAndIndexMapForMultipleObjects(resourceObjects []*repository.DevtronResourceObject) ([]*repository.DevtronResourceObject, map[int][]int, error) {
	var childResourceObjects []*repository.DevtronResourceObject
	resourceObjectIndexChildMap := make(map[int][]int)
	allSchemaIdsInChildObjects := make(map[int]bool, 0)
	schemaIdObjectIdsMap := make(map[int][]int)
	schemaIdOldObjectIdsMap := make(map[int][]int)
	for _, resourceObject := range resourceObjects {
		childDependencies, err := impl.getSpecificDependenciesInObjectDataFromJsonString(resourceObject.DevtronResourceSchemaId, resourceObject.ObjectData, bean.DevtronResourceDependencyTypeChild)
		if err != nil {
			return nil, nil, err
		}
		for _, childDependency := range childDependencies {
			allSchemaIdsInChildObjects[childDependency.DevtronResourceSchemaId] = true
			if childDependency.IdType == bean.ResourceObjectIdType {
				schemaIdObjectIdsMap[childDependency.DevtronResourceSchemaId] = append(schemaIdObjectIdsMap[childDependency.DevtronResourceSchemaId],
					childDependency.OldObjectId)
			} else if childDependency.IdType == bean.OldObjectId {
				schemaIdOldObjectIdsMap[childDependency.DevtronResourceSchemaId] = append(schemaIdOldObjectIdsMap[childDependency.DevtronResourceSchemaId],
					childDependency.OldObjectId)
			}
		}
	}
	for schemaId := range allSchemaIdsInChildObjects {
		objectIds := schemaIdObjectIdsMap[schemaId]
		oldObjectIds := schemaIdOldObjectIdsMap[schemaId]
		childObjects, err := impl.devtronResourceObjectRepository.GetAllObjectByIdsOrOldObjectIds(objectIds, oldObjectIds, schemaId)
		if err != nil {
			impl.logger.Errorw("error, GetAllObjectByIdsOrOldObjectIds", "err", err, "objectIds", objectIds, "oldObjectIds", oldObjectIds, "schemaId", schemaId)
			return childResourceObjects, resourceObjectIndexChildMap, err
		}
		childResourceObjects = append(childResourceObjects, childObjects...)
	}
	childObjectIdObjectsMap := make(map[string]int)    //map of "objectId-schemaId" and index of object in array
	childOldObjectIdObjectsMap := make(map[string]int) //map of "oldObjectId-schemaId" and index of object in array
	for i, childResourceObject := range childResourceObjects {
		childObjectIdObjectsMap[fmt.Sprintf("%d-%d", childResourceObject.Id, childResourceObject.DevtronResourceSchemaId)] = i
		if childResourceObject.OldObjectId > 0 {
			childOldObjectIdObjectsMap[fmt.Sprintf("%d-%d", childResourceObject.OldObjectId, childResourceObject.DevtronResourceSchemaId)] = i
		}
	}
	for i, resourceObject := range resourceObjects {
		childDependencies, err := impl.getSpecificDependenciesInObjectDataFromJsonString(resourceObject.DevtronResourceSchemaId, resourceObject.ObjectData, bean.DevtronResourceDependencyTypeChild)
		if err != nil {
			return nil, nil, err
		}
		for _, childDependency := range childDependencies {
			if childDependency.IdType == bean.ResourceObjectIdType {
				if indexOfChild, ok := childObjectIdObjectsMap[fmt.Sprintf("%d-%d", childDependency.OldObjectId, childDependency.DevtronResourceSchemaId)]; ok {
					resourceObjectIndexChildMap[i] = append(resourceObjectIndexChildMap[i], indexOfChild)
				}
			} else if childDependency.IdType == bean.OldObjectId {
				if indexOfChild, ok := childObjectIdObjectsMap[fmt.Sprintf("%d-%d", childDependency.OldObjectId, childDependency.DevtronResourceSchemaId)]; ok {
					resourceObjectIndexChildMap[i] = append(resourceObjectIndexChildMap[i], indexOfChild)
				}
			}
		}
	}
	return childResourceObjects, resourceObjectIndexChildMap, nil
}

//list resource object and related method starts
//

//
//Get resource object and related method starts

func (impl *DevtronResourceServiceImpl) GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectGetAPIBean, error) {
	adapter.SetIdTypeAndResourceIdBasedOnKind(req, req.OldObjectId)
	resp := &bean.DevtronResourceObjectGetAPIBean{
		DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{},
		DevtronResourceObjectBasicDataBean:  &bean.DevtronResourceObjectBasicDataBean{},
	}
	resourceSchema, existingResourceObject, err := impl.getResourceSchemaAndExistingObject(req)
	if err != nil {
		return nil, err
	}
	resp.Schema = resourceSchema.Schema
	if existingResourceObject == nil || existingResourceObject.Id == 0 {
		if req.Kind == bean.DevtronResourceRelease.ToString() || req.Kind == bean.DevtronResourceReleaseTrack.ToString() {
			impl.logger.Warnw("invalid get request, object not found", "req", req)
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
		}
	}
	resourceObject := &bean.DevtronResourceObjectGetAPIBean{
		DevtronResourceObjectDescriptorBean: req,
		DevtronResourceObjectBasicDataBean:  &bean.DevtronResourceObjectBasicDataBean{},
	}
	if req.UIComponents == nil || len(req.UIComponents) == 0 {
		// if no components are defined, fetch the complete data
		req.UIComponents = []bean.DevtronResourceUIComponent{bean.UIComponentAll}
	}
	for _, component := range req.UIComponents {
		f := getFuncForGetApiResourceKindUIComponent(req.Kind, component.ToString()) //getting function for component requested from UI
		if f == nil {
			impl.logger.Errorw("error component type not supported", "err", err, "kind", req.Kind, "component", component)
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrComponent, bean.InvalidResourceKindOrComponent)
		}
		err = f(impl, resourceSchema, existingResourceObject, resourceObject)
		if err != nil {
			impl.logger.Errorw("error, GetResourceObject", "err", err, "kind", req.Kind, "component", component)
			return nil, err
		}
	}
	return adapter.RemoveInternalOnlyFieldsFromGetResourceObjectBean(resourceObject), nil
}

//Get resource object and related method ends
//

//
//Create resource object and related method starts

func (impl *DevtronResourceServiceImpl) CreateResourceObject(ctx context.Context, reqBean *bean.DtResourceObjectCreateReqBean) error {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "CreateResourceObject")
	defer span.End()
	adapter.SetIdTypeAndResourceIdBasedOnKind(reqBean.DevtronResourceObjectDescriptorBean, reqBean.OldObjectId)
	err := validateCreateResourceRequest(reqBean)
	if err != nil {
		return err
	}
	err = impl.populateDefaultValuesToRequestBean(reqBean)
	if err != nil {
		return err
	}
	//getting schema latest from the db (not getting it from FE for edge cases when schema has got updated
	//just before an object update is requested)
	devtronResourceSchema, existingResourceObjectFound, err := impl.getResourceSchemaAndCheckIfObjectFound(reqBean.DevtronResourceObjectDescriptorBean)
	if err != nil {
		return err
	}
	if existingResourceObjectFound {
		impl.logger.Errorw("error encountered in CreateResourceObject", "request", reqBean, "err", err)
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceAlreadyExistsMessage, bean.ResourceAlreadyExistsMessage)
	}
	resourceObjReq := adapter.GetRequirementReqForCreateRequest(reqBean, "", false)
	return impl.createOrUpdateDevtronResourceObject(newCtx, resourceObjReq, devtronResourceSchema, nil, nil)
}

func (impl *DevtronResourceServiceImpl) getResourceSchemaAndCheckIfObjectFound(reqBean *bean.DevtronResourceObjectDescriptorBean) (*repository.DevtronResourceSchema, bool, error) {
	devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema from db", "err", err, "request", reqBean)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, false, err
	}
	exists, err := impl.dtResourceReadService.CheckIfExistingDevtronObject(reqBean.Id, devtronResourceSchema.Id, reqBean.IdType, reqBean.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", reqBean)
		return devtronResourceSchema, false, err
	}
	return devtronResourceSchema, exists, nil
}

// populateDefaultValuesToRequestBean is used to fill the default values of some fields for Create Resource request only.
func (impl *DevtronResourceServiceImpl) populateDefaultValuesToRequestBean(reqBean *bean.DtResourceObjectCreateReqBean) error {
	f := getFuncToPopulateDefaultValuesForCreateResourceRequest(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if f != nil {
		return f(impl, reqBean)
	}
	return nil
}

func validateCreateResourceRequest(reqBean *bean.DtResourceObjectCreateReqBean) error {
	f := getFuncToValidateCreateResourceRequest(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if f != nil {
		return f(reqBean)
	}
	return nil
}

//Create resource object and related method ends
//

//
//Create/Update(old method used for catalogue) resource object and related method starts

func (impl *DevtronResourceServiceImpl) CreateOrUpdateResourceObject(ctx context.Context, reqBean *bean.DtResourceObjectCatalogReqBean) error {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "CreateOrUpdateResourceObject")
	defer span.End()
	adapter.SetIdTypeAndResourceIdBasedOnKind(reqBean.DevtronResourceObjectDescriptorBean, reqBean.OldObjectId)
	//getting schema latest from the db (not getting it from FE for edge cases when schema has got updated
	//just before an object update is requested)
	devtronResourceSchema, devtronResourceObject, err := impl.getResourceSchemaAndExistingObject(reqBean.DevtronResourceObjectDescriptorBean)
	if err != nil {
		return err
	}
	resourceObjReq := adapter.GetRequirementRequestForCatalogRequest(reqBean, false)
	return impl.createOrUpdateDevtronResourceObject(newCtx, resourceObjReq, devtronResourceSchema, devtronResourceObject, nil)
}

//Create/Update(old method used for catalogue) resource object and related method ends
//

//
//Patch resource object and related method starts

func (impl *DevtronResourceServiceImpl) PatchResourceObject(ctx context.Context, req *bean.DtResourceObjectPatchReqBean) (*bean.SuccessResponse, error) {
	adapter.SetIdTypeAndResourceIdBasedOnKind(req.DevtronResourceObjectDescriptorBean, req.OldObjectId)
	_, existingResourceObject, err := impl.getResourceSchemaAndExistingObject(req.DevtronResourceObjectDescriptorBean)
	if err != nil {
		return nil, err
	}
	// performing json patch operations
	objectData := existingResourceObject.ObjectData
	newObjectData, auditPaths, err := impl.performResourcePatchOperation(req.DevtronResourceObjectDescriptorBean, objectData, req.PatchQuery)
	if err != nil {
		impl.logger.Errorw("error encountered in PatchResourceObject", "query", req.PatchQuery, "err", err)
		return nil, err
	}
	//updating final object data in resource object
	existingResourceObject.ObjectData = newObjectData
	existingResourceObject.UpdatedBy = req.UserId
	existingResourceObject.UpdatedOn = time.Now()
	err = impl.devtronResourceObjectRepository.Update(nil, existingResourceObject)
	if err != nil {
		impl.logger.Errorw("error encountered in PatchResourceObject", "err", err, "req", existingResourceObject)
		return nil, err
	}
	impl.dtResourceObjectAuditService.SaveAudit(existingResourceObject, repository.AuditOperationTypePatch, auditPaths)
	return adapter.GetSuccessPassResponse(), nil
}

func (impl *DevtronResourceServiceImpl) performResourcePatchOperation(descriptorBean *bean.DevtronResourceObjectDescriptorBean, objectData string, queries []bean.PatchQuery) (string, []string, error) {
	newObjectData := ""
	auditPaths := make([]string, 0, len(queries))
	var err error
	f := getFuncToPerformPatchOperation(descriptorBean.Kind, descriptorBean.SubKind, descriptorBean.Version)
	if f == nil {
		return newObjectData, auditPaths, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
	}
	newObjectData, auditPaths, err = f(impl, objectData, queries)
	if err != nil {
		impl.logger.Errorw("error in performing patch operation", "err", err, "descriptorBean", descriptorBean, "patchQuery", queries)
		return newObjectData, auditPaths, err
	}
	return newObjectData, auditPaths, nil
}

//Patch resource object and related method ends
//

//
//Delete resource object and related method starts

func (impl *DevtronResourceServiceImpl) DeleteResourceObject(ctx context.Context, req *bean.DevtronResourceObjectDescriptorBean) (*bean.SuccessResponse, error) {
	adapter.SetIdTypeAndResourceIdBasedOnKind(req, req.OldObjectId)
	//getting object
	_, existingObj, err := impl.getResourceSchemaAndExistingObject(req)
	if err != nil {
		impl.logger.Errorw("error in getting existing resource object", "err", err, "req", req)
		return nil, err
	}
	f := getFuncToValidateResourceObjectDelete(req.Kind, req.SubKind, req.Version)
	if f != nil {
		isValid, err := f(impl, existingObj)
		if err != nil {
			impl.logger.Errorw("error in validation delete object request", "err", err, "req", req)
			return nil, err
		}
		if !isValid {
			impl.logger.Errorw("invalid delete request", "err", err, "req", req)
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidDeleteRequest, bean.InvalidDeleteRequest)
		}
	}
	err = impl.dtResourceInternalProcessingService.DeleteObjectAndItsDependency(req)
	if err != nil {
		impl.logger.Errorw("error in DeleteResourceObject", "request", req)
		return nil, err
	}
	return adapter.GetSuccessPassResponse(), nil
}

//Delete resource object and related method ends
//

//
//Get resource object dependencies and related method starts

func (impl *DevtronResourceServiceImpl) GetResourceDependencies(req *bean.DevtronResourceObjectDescriptorBean, query *apiBean.GetDependencyQueryParams) (*bean.DevtronResourceObjectBean, error) {
	adapter.SetIdTypeAndResourceIdBasedOnKind(req, req.OldObjectId)
	resourceSchemaOfRequestObject, existingResourceObject, err := impl.getResourceSchemaAndExistingObject(req)
	if err != nil {
		return nil, err
	}
	if existingResourceObject == nil || existingResourceObject.Id < 1 {
		//Since we have not added a migration for saving resource objects its always possible that resource object is
		//not added but dependency is added and those resource objects should be included in downstream
		existingResourceObject = &repository.DevtronResourceObject{
			OldObjectId:             req.OldObjectId,
			DevtronResourceId:       resourceSchemaOfRequestObject.DevtronResourceId,
			DevtronResourceSchemaId: resourceSchemaOfRequestObject.Id,
			ObjectData:              bean.EmptyJsonObject,
		}
	}
	response, err := impl.getResourceDependenciesData(req, query, resourceSchemaOfRequestObject, existingResourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting resource dependencies data", "err", err, "query", query)
		return nil, err
	}
	return adapter.RemoveInternalOnlyFieldsFromResourceObjectBean(response), nil
}

// getDependencyBeanForGetDependenciesApi is used for get resource dependencies by extra child objects data which is not present in schema
func (impl *DevtronResourceServiceImpl) getDependencyBeanForGetDependenciesApi(parentResourceType *bean.DevtronResourceTypeReq, dependency string, isLite bool) (*bean.DevtronResourceDependencyBean, error) {
	dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependency, isLite)
	if err != nil {
		impl.logger.Errorw("error encountered in getDependencyBeanForGetDependenciesApi", "err", err, "dependency", dependency)
		return nil, err
	}
	oldObjectId := int(gjson.Get(dependency, bean.IdKey).Int())
	childObjects, err := impl.getChildObjectsByParentResourceType(parentResourceType, dependency, isLite, oldObjectId)
	if err != nil {
		impl.logger.Errorw("error encountered in getDependencyBeanForGetDependenciesApi", "err", err, "dependency", dependency)

		return nil, err
	}
	// setting child objects only for ui (get resource dependencies api), not stored in schema
	dependencyBean.ChildObjects = childObjects

	return dependencyBean, nil
}

func (impl *DevtronResourceServiceImpl) getResourceDependenciesData(req *bean.DevtronResourceObjectDescriptorBean, query *apiBean.GetDependencyQueryParams,
	resourceSchema *repository.DevtronResourceSchema, resourceObject *repository.DevtronResourceObject) (*bean.DevtronResourceObjectBean, error) {
	response := &bean.DevtronResourceObjectBean{
		Dependencies:      make([]*bean.DevtronResourceDependencyBean, 0),
		ChildDependencies: make([]*bean.DevtronResourceDependencyBean, 0),
	}
	f := getFuncToUpdateResourceDependenciesDataInResponseObj(req.Kind, req.SubKind, req.Version)
	if f != nil {
		return f(impl, req, query, resourceSchema, resourceObject, response)
	}
	return response, nil
}

//Get resource object dependencies and related method ends
//

// Get dependency config option starts
//

func (impl *DevtronResourceServiceImpl) GetDependencyConfigOptions(req *bean.DevtronResourceObjectDescriptorBean, query *apiBean.GetConfigOptionsQueryParams) (any, error) {
	adapter.SetIdTypeAndResourceIdBasedOnKind(req, req.OldObjectId)
	resourceSchema, existingResourceObject, err := impl.getResourceSchemaAndExistingObject(req)
	if err != nil {
		return nil, err
	}
	if existingResourceObject == nil || existingResourceObject.Id == 0 {
		if req.Kind == bean.DevtronResourceRelease.ToString() || req.Kind == bean.DevtronResourceReleaseTrack.ToString() {
			impl.logger.Warnw("invalid get request, object not found", "req", req)
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
		} else {
			existingResourceObject = &repository.DevtronResourceObject{
				OldObjectId:             req.OldObjectId,
				DevtronResourceId:       resourceSchema.DevtronResourceId,
				DevtronResourceSchemaId: resourceSchema.Id,
				ObjectData:              bean.EmptyJsonObject,
			}
		}
	}
	var response any
	switch query.ConfigOption {
	case apiBean.ArtifactConfig:
		f := getFuncToGetArtifactConfigOptions(req.Kind, req.SubKind, req.Version)
		if f == nil {
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
		}
		response, err = f(impl, req, existingResourceObject, query)
		if err != nil {
			return nil, err
		}
	case apiBean.CommitConfig:
	default:
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryConfigOption, bean.InvalidQueryConfigOption)
	}
	return response, nil
}

//
// GetDependencyConfigOptions ends

//
//Create/Update resource object dependencies and related method starts

func (impl *DevtronResourceServiceImpl) CreateOrUpdateResourceDependencies(ctx context.Context, req *bean.DtResourceObjectDependenciesReqBean) error {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "CreateOrUpdateResourceDependencies")
	defer span.End()
	adapter.SetIdTypeAndResourceIdBasedOnKind(req.DevtronResourceObjectDescriptorBean, req.OldObjectId)
	err := impl.setDefaultDataAndValidateDependencies(req)
	if err != nil {
		impl.logger.Errorw("validation error, CreateOrUpdateResourceDependencies", "err", err, "req", req)
		return err
	}
	allRequests, allRequestSchemas, existingObjectsMap, err := impl.getUpdatedDependenciesRequestData(req)
	if err != nil {
		impl.logger.Errorw("error, getUpdatedDependenciesRequestData", "err", err, "req", req)
		return err
	}
	for i := range allRequests {
		request := allRequests[i]
		resourceObjReq, err := adapter.GetRequirementRequestForDependenciesRequest(request, true)
		if err != nil {
			impl.logger.Errorw("error, GetRequirementRequestForDependenciesRequest", "err", err, "request", request)
			return err
		}
		keyToGetSchema := helper.GetKeyForADependencyMap(request.OldObjectId, request.SchemaId)
		devtronResourceObject := existingObjectsMap[keyToGetSchema]
		devtronResourceObjectPresentAlready := devtronResourceObject != nil && devtronResourceObject.Id > 0
		if devtronResourceObjectPresentAlready {
			//update request
			f := getFuncToHandleExistingObjInDependenciesUpdateReq(request.Kind, request.SubKind, request.Version)
			if f == nil {
				//not returning since all resource type do not have update req handle functions
			} else {
				f(impl, resourceObjReq.ReqBean, devtronResourceObject)
			}
		}
		err = impl.createOrUpdateDevtronResourceObject(newCtx, resourceObjReq, allRequestSchemas[i], devtronResourceObject, []string{bean.ResourceObjectDependenciesPath})
		if err != nil {
			impl.logger.Errorw("error, createOrUpdateDevtronResourceObject", "err", err, "request", request)
			return err
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) setDefaultDataAndValidateDependencies(req *bean.DtResourceObjectDependenciesReqBean) error {
	if len(req.Dependencies) == 0 {
		impl.logger.Errorw("invalid request, no dependency in request", "req", req)
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.BadRequestDependenciesErrorMessage, bean.InvalidNoDependencyRequest)
	}
	allDependenciesToBeValidated := make([]*bean.DevtronResourceDependencyBean, 0, len(req.Dependencies)+2*len(req.ChildDependencies))
	for i := range req.Dependencies {
		dep := req.Dependencies[i]
		adapter.RemoveRedundantFieldsAndSetDefaultForDependency(dep, false)
		if helper.CheckIfDependencyTypeToBeValidated(dep.TypeOfDependency) {
			allDependenciesToBeValidated = append(allDependenciesToBeValidated, dep)
		}
	}
	for j := range req.ChildDependencies {
		childDependency := req.ChildDependencies[j]
		adapter.RemoveRedundantFieldsAndSetDefaultForDependency(childDependency, true)
		if helper.CheckIfDependencyTypeToBeValidated(childDependency.TypeOfDependency) {
			allDependenciesToBeValidated = append(allDependenciesToBeValidated, childDependency)
		}
		//here assuming that dependencies of childDependency further don't have their own dependencies, i.e. only one level of nesting in resources
		for k := range childDependency.Dependencies {
			depOfChildDep := childDependency.Dependencies[k]
			adapter.RemoveRedundantFieldsAndSetDefaultForDependency(depOfChildDep, false)
			if helper.CheckIfDependencyTypeToBeValidated(depOfChildDep.TypeOfDependency) {
				allDependenciesToBeValidated = append(allDependenciesToBeValidated, depOfChildDep)
			}
		}
	}

	mapOfSchemaIdAndDependencyData := make(map[int]bean.IdIdentifierIndex)
	for i, dependency := range allDependenciesToBeValidated {
		err := impl.validateDependencyResourceType(allDependenciesToBeValidated[i])
		if err != nil {
			return err
		}

		allDependenciesToBeValidated[i].IdType = bean.OldObjectId // currently only apps and cd pipelines are expected in dependencies PUT request
		if dependency.OldObjectId > 0 {
			data := mapOfSchemaIdAndDependencyData[dependency.DevtronResourceSchemaId]
			data.Ids = append(data.Ids, dependency.OldObjectId)
			mapOfSchemaIdAndDependencyData[dependency.DevtronResourceSchemaId] = data
		} else {
			data := mapOfSchemaIdAndDependencyData[dependency.DevtronResourceSchemaId]
			data.Identifiers = append(data.Identifiers, dependency.Identifier)
			data.IdentifiersIndex = append(data.IdentifiersIndex, i)
			mapOfSchemaIdAndDependencyData[dependency.DevtronResourceSchemaId] = data
		}
	}
	internalMessage := ""
	isRequestInvalid := false

	invalidSchemaIds := make([]int, 0, len(mapOfSchemaIdAndDependencyData))
	var invalidAppIds []int
	var invalidAppNames []string
	var invalidCdPipelineIds []int
	var err error
	for devtronResourceSchemaId, depData := range mapOfSchemaIdAndDependencyData {
		if devtronResourceSchema, ok := impl.devtronResourcesSchemaMapById[devtronResourceSchemaId]; ok {
			switch devtronResourceSchema.DevtronResource.Kind {
			case bean.DevtronResourceDevtronApplication.ToString():
				var mapOfAppNameId map[string]int
				mapOfAppNameId, invalidAppIds, invalidAppNames, err = impl.getAppsMapAndReturnNotFoundIdsAndNames(depData.Ids, depData.Identifiers)
				if err != nil {
					impl.logger.Errorw("error, getAppsAndReturnNotFoundIds", "err", err, "appIds", depData.Ids)
					return err
				}
				if len(invalidAppIds) == 0 && len(invalidAppNames) == 0 {
					for i, identifier := range depData.Identifiers {
						dependencyIndexToBeUpdated := depData.IdentifiersIndex[i]
						allDependenciesToBeValidated[dependencyIndexToBeUpdated].Identifier = ""
						allDependenciesToBeValidated[dependencyIndexToBeUpdated].OldObjectId = mapOfAppNameId[identifier]
					}
				}
			case bean.DevtronResourceCdPipeline.ToString():
				if len(depData.Identifiers) > 0 {
					isRequestInvalid = true
					internalMessage += fmt.Sprintf("identifiers not supported in cd pipelines as of now : %v\n", depData.Identifiers)
				}
				pipelineIds := depData.Ids
				invalidCdPipelineIds, err = impl.getCdPipelinesAndReturnNotFoundIds(pipelineIds)
				if err != nil {
					impl.logger.Errorw("error, getCdPipelinesAndReturnNotFoundIds", "err", err, "pipelineIds", pipelineIds)
					return err
				}
			default:
				invalidSchemaIds = append(invalidSchemaIds, devtronResourceSchemaId)
			}
		} else {
			invalidSchemaIds = append(invalidSchemaIds, devtronResourceSchemaId)
		}
	}
	if len(invalidSchemaIds) > 0 {
		isRequestInvalid = true
		internalMessage += fmt.Sprintf("invalid schemaIds : %v\n", invalidSchemaIds)
	}
	if len(invalidAppIds) > 0 {
		isRequestInvalid = true
		internalMessage += fmt.Sprintf("invalid appIds : %v\n", invalidAppIds)
	}
	if len(invalidAppNames) > 0 {
		isRequestInvalid = true
		internalMessage += fmt.Sprintf("invalid appNames : %v\n", invalidAppNames)
	}
	if len(invalidCdPipelineIds) > 0 {
		isRequestInvalid = true
		internalMessage += fmt.Sprintf("invalid cdPipelineIds : %v\n", invalidCdPipelineIds)
	}
	if isRequestInvalid {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.BadRequestDependenciesErrorMessage, internalMessage)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getAppsMapAndReturnNotFoundIdsAndNames(appIds []int, appNames []string) (map[string]int, []int, []string, error) {
	invalidAppIds := make([]int, 0, len(appIds))
	invalidAppNames := make([]string, 0, len(appNames))
	mapOfAppIds := make(map[int]bool)
	mapOfAppNames := make(map[string]int) //map of appName and its id
	apps, err := impl.appRepository.FindAppsByIdsOrNames(appIds, appNames)
	if err != nil {
		impl.logger.Errorw("error in getting apps by ids or names", "err", err, "ids", appIds, "names", appNames)
		return mapOfAppNames, invalidAppIds, invalidAppNames, err
	}
	for _, app := range apps {
		mapOfAppIds[app.Id] = true
		mapOfAppNames[app.AppName] = app.Id
	}
	for _, appId := range appIds {
		if _, ok := mapOfAppIds[appId]; !ok {
			invalidAppIds = append(invalidAppIds, appId)
		}
	}
	for _, appName := range appNames {
		if _, ok := mapOfAppNames[appName]; !ok {
			invalidAppNames = append(invalidAppNames, appName)
		}
	}

	return mapOfAppNames, invalidAppIds, invalidAppNames, nil
}

func (impl *DevtronResourceServiceImpl) getCdPipelinesAndReturnNotFoundIds(pipelineIds []int) ([]int, error) {
	invalidCdPipelineIds := make([]int, 0, len(pipelineIds))
	mapOfCdPipelines := make(map[int]*pipelineConfig.Pipeline)
	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting cd pipelines by ids", "err", err, "ids", pipelineIds)
		return nil, err
	}
	for _, pipeline := range pipelines {
		mapOfCdPipelines[pipeline.Id] = pipeline
	}
	if len(mapOfCdPipelines) != len(pipelineIds) {
		for _, dependencyId := range pipelineIds {
			if _, ok := mapOfCdPipelines[dependencyId]; !ok {
				invalidCdPipelineIds = append(invalidCdPipelineIds, dependencyId)
			}
		}
	}
	return invalidCdPipelineIds, nil
}

func (impl *DevtronResourceServiceImpl) getUpdatedDependenciesRequestData(req *bean.DtResourceObjectDependenciesReqBean) ([]*bean.DtResourceObjectDependenciesReqBean,
	[]*repository.DevtronResourceSchema, map[string]*repository.DevtronResourceObject, error) {
	parentDevtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, nil, nil, err
	}
	allRequests := make([]*bean.DtResourceObjectDependenciesReqBean, 0, len(req.ChildDependencies)+1)
	allRequestSchemas := make([]*repository.DevtronResourceSchema, 0, len(req.ChildDependencies)+1)
	allArgValues := make([]interface{}, 0, len(req.ChildDependencies)+1)
	allArgTypes := make([]string, 0, len(req.ChildDependencies)+1)
	devtronSchemaIdsForAllArgs := make([]int, 0, len(req.ChildDependencies)+1)

	//adding oldObjectId and Name for main request
	isTypeRelease := req.Kind == bean.DevtronResourceRelease.ToString()
	appendDbObjectArgDetails(&allArgValues, &allArgTypes, &devtronSchemaIdsForAllArgs, req.Id, req.OldObjectId, parentDevtronResourceSchema.Id, isTypeRelease)

	for j := range req.ChildDependencies {
		childDependency := req.ChildDependencies[j]
		//adding info of parent dependency in this child dependency's dependencies
		childDependency.Dependencies = append(childDependency.Dependencies, &bean.DevtronResourceDependencyBean{
			OldObjectId:             req.OldObjectId,
			DevtronResourceSchemaId: parentDevtronResourceSchema.Id,
			DevtronResourceId:       parentDevtronResourceSchema.DevtronResourceId,
			TypeOfDependency:        bean.DevtronResourceDependencyTypeParent,
		})

		//getting devtronResourceSchema for this child dependency
		devtronResourceSchema := impl.devtronResourcesSchemaMapById[childDependency.DevtronResourceSchemaId]
		kind, subKind := helper.GetKindSubKindOfResourceBySchemaObject(devtronResourceSchema, impl.devtronResourcesMapById)
		reqForChildDependency := &bean.DtResourceObjectDependenciesReqBean{
			DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{
				Kind:        kind,
				SubKind:     subKind,
				Version:     devtronResourceSchema.Version,
				OldObjectId: childDependency.OldObjectId,
				SchemaId:    childDependency.DevtronResourceSchemaId,
			},
			Dependencies: childDependency.Dependencies,
		}
		allRequestSchemas = append(allRequestSchemas, devtronResourceSchema)
		allRequests = append(allRequests, reqForChildDependency)

		//need to add this child dependency in parent
		childDependency.Dependencies = nil //since we only need to add child dependency for parent-child relationship and not keeping nested dependencies in every object
		childDependency.TypeOfDependency = bean.DevtronResourceDependencyTypeChild
		req.Dependencies = append(req.Dependencies, childDependency)

		//adding oldObjectIds or names for getting existing objects
		appendDbObjectArgDetails(&allArgValues, &allArgTypes, &devtronSchemaIdsForAllArgs, 0, childDependency.OldObjectId, childDependency.DevtronResourceSchemaId, false)
	}
	req.SchemaId = parentDevtronResourceSchema.Id
	//adding our initial request to allRequest
	allRequests = append(allRequests, req)
	allRequestSchemas = append(allRequestSchemas, parentDevtronResourceSchema)

	existingObjectsMap, err := impl.getExistingObjectsMap(allArgValues, allArgTypes, devtronSchemaIdsForAllArgs)
	if err != nil {
		impl.logger.Errorw("error, getExistingObjectsMap", "err", err)
		return nil, nil, nil, err
	}
	return allRequests, allRequestSchemas, existingObjectsMap, nil
}

func (impl *DevtronResourceServiceImpl) getExistingObjectsMap(allArgValues []interface{},
	allArgTypes []string, devtronSchemaIdsForAllArgs []int) (map[string]*repository.DevtronResourceObject, error) {
	existingObjectsMap := make(map[string]*repository.DevtronResourceObject, len(allArgValues))
	if len(allArgValues) > 0 {
		oldObjects, err := impl.devtronResourceObjectRepository.GetObjectsByArgAndSchemaIds(allArgValues, allArgTypes, devtronSchemaIdsForAllArgs)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting devtron schemas by old object id or name and schema id array", "err", err,
				"allArgValues", allArgValues, "allArgTypes", allArgTypes, "schemaIds", devtronSchemaIdsForAllArgs)
			return nil, err
		}
		for _, oldObject := range oldObjects {
			existingObjectsMap[helper.GetKeyForADependencyMap(oldObject.OldObjectId, oldObject.DevtronResourceSchemaId)] = oldObject
		}
	}
	return existingObjectsMap, nil
}

func (impl *DevtronResourceServiceImpl) validateDependencyResourceType(dep *bean.DevtronResourceDependencyBean) error {
	isValidDependencyResourceType := false
	//first checking the resourceId and schemaId
	_, ok1 := impl.devtronResourcesMapById[dep.DevtronResourceId]
	_, ok2 := impl.devtronResourcesSchemaMapById[dep.DevtronResourceSchemaId]
	isValidDependencyResourceType = ok1 && ok2
	if !isValidDependencyResourceType { //ids not valid, check if data received from resourceType is correct
		dtResource, dtSchema, err := impl.getResourceAndSchemaFromResourceType(dep.DevtronResourceTypeReq)
		if err != nil {
			//both ids and resource type data invalid
			return err
		}
		//valid resource type data, set ids
		dep.DevtronResourceId = dtResource.Id
		dep.DevtronResourceSchemaId = dtSchema.Id
	}
	//setting resource type as nil so that it does not get saved in json
	dep.DevtronResourceTypeReq = nil
	return nil
}

func (impl *DevtronResourceServiceImpl) getResourceAndSchemaFromResourceType(resourceType *bean.DevtronResourceTypeReq) (dtResource *repository.DevtronResource,
	dtSchema *repository.DevtronResourceSchema, err error) {
	kindSplits := strings.Split(resourceType.ResourceKind.ToString(), "/")
	var ok bool
	if len(kindSplits) == 1 {
		//no subKind
		kind := kindSplits[0]
		dtResource, ok = impl.devtronResourcesMapByKind[kind]
		if !ok {
			impl.logger.Errorw("invalid resourceType", "resourceType", resourceType)
			return dtResource, dtSchema, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
		}
		resourceType.ResourceKind = bean.DevtronResourceKind(kindSplits[0])
	} else if len(kindSplits) == 2 {
		kind := kindSplits[0]
		subKind := kindSplits[1]
		_, ok = impl.devtronResourcesMapByKind[kind]
		if !ok {
			impl.logger.Errorw("invalid resourceType", "resourceType", resourceType)
			return dtResource, dtSchema, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
		}
		dtResource, ok = impl.devtronResourcesMapByKind[subKind]
		if !ok {
			impl.logger.Errorw("invalid resourceType", "resourceType", resourceType)
			return dtResource, dtSchema, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
		}
	}
	//check version
	if supportedVersionsMap, ok :=
		bean.DevtronResourceSupportedVersionMap[bean.DevtronResourceKind(dtResource.Kind)]; !ok || !supportedVersionsMap[resourceType.ResourceVersion] {
		impl.logger.Errorw("invalid resourceType", "resourceType", resourceType)
		return dtResource, dtSchema, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceVersion, bean.InvalidResourceVersion)
	}
	dtSchema, err = impl.getSchemaByResourceIdAndVersion(dtResource.Id, resourceType.ResourceVersion)
	if err != nil {
		impl.logger.Errorw("invalid resourceType", "resourceType", resourceType)
		return dtResource, dtSchema, err
	}
	return dtResource, dtSchema, nil
}

func (impl *DevtronResourceServiceImpl) getSchemaByResourceIdAndVersion(resourceId int,
	version bean.DevtronResourceVersion) (*repository.DevtronResourceSchema, error) {
	for _, schema := range impl.devtronResourcesSchemaMapById {
		if schema.DevtronResourceId == resourceId && schema.Version == version.ToString() {
			return schema, nil
		}
	}
	return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceVersion, bean.InvalidResourceVersion)
}

func appendDbObjectArgDetails(argValues *[]interface{}, argTypes *[]string, schemaIds *[]int, id, oldObjectId, schemaId int, isTypeRelease bool) {
	argValue, argType := getArgTypeAndValueForObject(oldObjectId)
	//TODO : make generic
	if isTypeRelease {
		argType = bean.IdDbColumnKey
		argValue = id
	}
	*argValues = append(*argValues, argValue)
	*argTypes = append(*argTypes, argType)
	*schemaIds = append(*schemaIds, schemaId)
}

func getArgTypeAndValueForObject(oldObjectId int) (argValue interface{}, argType string) {
	if oldObjectId > 0 {
		argValue = oldObjectId
		argType = bean.OldObjectIdDbColumnKey
	}
	return argValue, argType
}

//Create/Update resource object dependencies and related method ends
//

//
//Patch resource object dependencies and related method starts

func (impl *DevtronResourceServiceImpl) PatchResourceDependencies(ctx context.Context, req *bean.DevtronResourceDependencyPatchAPIBean) (*bean.SuccessResponse, error) {
	adapter.SetIdTypeAndResourceIdBasedOnKind(req.DevtronResourceObjectDescriptorBean, req.OldObjectId)
	_, existingResourceObject, err := impl.getResourceSchemaAndExistingObject(req.DevtronResourceObjectDescriptorBean)
	if err != nil {
		return nil, err
	}
	// performing json patch operations
	objectData := existingResourceObject.ObjectData
	auditPaths := make([]string, 0)
	for _, patchReq := range req.DependencyPatch {
		// performing json patch operations
		patchQuery := patchReq.PatchQuery
		dependencyInfo := patchReq.DependencyInfo
		jsonPath := ""
		for _, query := range patchQuery {
			objectData, jsonPath, err = impl.performDependencyPatchOperation(existingResourceObject.DevtronResourceSchemaId, objectData, query, dependencyInfo)
			if err != nil {
				impl.logger.Errorw("error encountered in PatchResourceObject", "query", query, "err", err)
				return nil, err
			}
			auditPaths = append(auditPaths, jsonPath)
		}
	}
	//updating final object data in resource object
	existingResourceObject.ObjectData = objectData
	existingResourceObject.UpdatedBy = req.UserId
	existingResourceObject.UpdatedOn = time.Now()
	err = impl.devtronResourceObjectRepository.Update(nil, existingResourceObject)
	if err != nil {
		impl.logger.Errorw("error encountered in PatchResourceObject", "err", err, "req", existingResourceObject)
		return nil, err
	}
	impl.dtResourceObjectAuditService.SaveAudit(existingResourceObject, repository.AuditOperationTypePatch, auditPaths)
	return adapter.GetSuccessPassResponse(), nil
}

func (impl *DevtronResourceServiceImpl) performDependencyPatchOperation(devtronResourceSchemaId int, objectData string, query bean.PatchQuery, dependencyInfo *bean.DependencyInfo) (string, string, error) {
	switch query.Path {
	case bean.ApplicationQueryPath:
		//currently only remove is supported
		if query.Operation != bean.Remove || dependencyInfo == nil {
			return "", "", util.GetApiErrorAdapter(http.StatusNotFound, "400", bean.InvalidPatchOperation, bean.InvalidPatchOperation)
		}
		return impl.removeApplicationDependency(devtronResourceSchemaId, objectData, dependencyInfo)
	case bean.ReleaseInstructionQueryPath:
		return impl.patchReleaseInstructionForApplication(devtronResourceSchemaId, objectData, dependencyInfo, query.Value.(string))
	case bean.ImageQueryPath:
		return impl.patchConfigImageForApplication(devtronResourceSchemaId, objectData, dependencyInfo, query.Value)
	default:
		return "", "", util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.PatchPathNotSupportedError, bean.PatchPathNotSupportedError)
	}
}

func (impl *DevtronResourceServiceImpl) patchReleaseInstructionForApplication(devtronResourceSchemaId int, objectData string, dependencyInfo *bean.DependencyInfo, value string) (string, string, error) {
	id, err := impl.findAppIdFromDependencyInfo(dependencyInfo)
	if err != nil {
		impl.logger.Errorw("error encountered in removeApplicationDependency", "dependencyInfo", dependencyInfo, "err", err)
		return objectData, "", err
	}
	updatedDependencies, indexChanged, err := impl.patchReleaseInstructionForADependency(devtronResourceSchemaId, objectData, id, value)
	if err != nil {
		impl.logger.Errorw("error in removeApplicationDependency", "err", err, "id", id, "value", value)
		return objectData, "", err
	}
	objectData, err = sjson.Set(objectData, bean.ResourceObjectDependenciesPath, updatedDependencies)
	if err != nil {
		impl.logger.Errorw("error in removeApplicationDependency", "err", err, "updatedDependencies", updatedDependencies)
		return objectData, "", err
	}
	return objectData, fmt.Sprintf("%s/%v/%s", "dependencies", indexChanged, "config/releaseInstruction"), nil
}

func (impl *DevtronResourceServiceImpl) patchReleaseInstructionForADependency(devtronResourceSchemaId int, objectData string, id int, releaseInstruction string) ([]*bean.DevtronResourceDependencyBean, int, error) {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	var indexChanged int
	kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(devtronResourceSchemaId,
		impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
	if err != nil {
		return nil, 0, err
	}
	parentResourceType := &bean.DevtronResourceTypeReq{
		ResourceKind:    bean.DevtronResourceKind(kind),
		ResourceSubKind: bean.DevtronResourceKind(subKind),
		ResourceVersion: bean.DevtronResourceVersion(version),
	}
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependencyResult.String(), false)
		if err != nil {
			return nil, 0, err
		}
		if dependencyBean.OldObjectId == id {
			dependencyBean.Config.ReleaseInstruction = releaseInstruction
			indexChanged = dependencyBean.Index
		}
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies, indexChanged, nil
}

func (impl *DevtronResourceServiceImpl) patchConfigImageForApplication(devtronResourceSchemaId int, objectData string, dependencyInfo *bean.DependencyInfo, value interface{}) (string, string, error) {
	id, err := impl.findAppIdFromDependencyInfo(dependencyInfo)
	if err != nil {
		impl.logger.Errorw("error encountered in removeApplicationDependency", "dependencyInfo", dependencyInfo, "err", err)
		return objectData, "", err
	}
	artifactConfig, err := getArtifactConfigFromQueryValue(value)
	if err != nil {
		impl.logger.Errorw("error encountered in patchConfigImageForApplication", "value", value, "err", err)
		return objectData, "", err
	}
	updatedDependencies, indexChanged, err := impl.patchImageForADependency(devtronResourceSchemaId, objectData, id, artifactConfig)
	if err != nil {
		impl.logger.Errorw("error in removeApplicationDependency", "err", err, "id", id, "value", value)
		return objectData, "", err
	}
	objectData, err = sjson.Set(objectData, bean.ResourceObjectDependenciesPath, updatedDependencies)
	if err != nil {
		impl.logger.Errorw("error in patchConfigImageForApplication", "err", err, "updatedDependencies", updatedDependencies)
		return objectData, "", err
	}
	return objectData, fmt.Sprintf("%s/%v/%s", "dependencies", indexChanged, "config/image"), nil
}

func (impl *DevtronResourceServiceImpl) patchImageForADependency(devtronResourceSchemaId int, objectData string, id int, artifactConfig *bean.ArtifactConfig) ([]*bean.DevtronResourceDependencyBean, int, error) {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	indexChanged := 0
	kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(devtronResourceSchemaId, impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
	if err != nil {
		return nil, 0, err
	}
	parentResourceType := &bean.DevtronResourceTypeReq{
		ResourceKind:    bean.DevtronResourceKind(kind),
		ResourceSubKind: bean.DevtronResourceKind(subKind),
		ResourceVersion: bean.DevtronResourceVersion(version),
	}
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependencyResult.String(), false)
		if err != nil {
			return nil, 0, err
		}
		if dependencyBean.OldObjectId == id {
			dependencyBean.Config.ArtifactConfig = &bean.ArtifactConfig{
				ArtifactId:   artifactConfig.ArtifactId,
				RegistryName: artifactConfig.RegistryName,
				RegistryType: artifactConfig.RegistryType,
			}
			dependencyBean.ChildInheritance = []*bean.ChildInheritance{{ResourceId: impl.devtronResourcesMapByKind[bean.DevtronResourceCdPipeline.ToString()].Id, Selector: adapter.GetDefaultCdPipelineSelector()}}
			indexChanged = dependencyBean.Index
		}
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies, indexChanged, nil
}

func getArtifactConfigFromQueryValue(value interface{}) (*bean.ArtifactConfig, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var artifactConfig bean.ArtifactConfig
	err = json.Unmarshal(data, &artifactConfig)
	if err != nil {
		return nil, err
	}
	return &artifactConfig, nil
}

// removeApplicationDependency remove an application dependency from existing resource object and return updated json object data with error if any, identifier in cse of application is app_name
func (impl *DevtronResourceServiceImpl) removeApplicationDependency(devtronResourceSchemaId int, objectData string, dependencyInfo *bean.DependencyInfo) (string, string, error) {
	id, err := impl.findAppIdFromDependencyInfo(dependencyInfo)
	if err != nil {
		impl.logger.Errorw("error encountered in removeApplicationDependency", "dependencyInfo", dependencyInfo, "err", err)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusNotFound, "404", bean.ApplicationDependencyNotFoundError, bean.ApplicationDependencyNotFoundError)
		}
		return objectData, "", err
	}
	updatedDependencies, indexOfDependencyRemoved, err := impl.getUpdatedDependenciesRemovingParticularWithId(devtronResourceSchemaId, objectData, id)
	if err != nil {
		impl.logger.Errorw("error in removeApplicationDependency", "err", err, "id", id)
		return objectData, "", err
	}
	isDependent := helper.CheckIfDependencyIsDependentOnRemovedDependency(updatedDependencies, indexOfDependencyRemoved)
	if isDependent {
		return objectData, "", util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ApplicationDependencyFoundError, bean.ApplicationDependencyFoundError)
	}
	objectData, err = sjson.Set(objectData, bean.ResourceObjectDependenciesPath, updatedDependencies)
	if err != nil {
		impl.logger.Errorw("error in removeApplicationDependency", "err", err, "updatedDependencies", updatedDependencies)
		return objectData, "", err
	}

	return objectData, fmt.Sprintf("%s/%v", bean.ResourceObjectDependenciesPath, indexOfDependencyRemoved), nil
}

func (impl *DevtronResourceServiceImpl) getUpdatedDependenciesRemovingParticularWithId(devtronResourceSchemaId int, objectData string, id int) ([]*bean.DevtronResourceDependencyBean, int, error) {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	indexOfDependencyRemoved := 0
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(devtronResourceSchemaId,
		impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
	if err != nil {
		return nil, 0, err
	}
	parentResourceType := &bean.DevtronResourceTypeReq{
		ResourceKind:    bean.DevtronResourceKind(kind),
		ResourceSubKind: bean.DevtronResourceKind(subKind),
		ResourceVersion: bean.DevtronResourceVersion(version),
	}
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependencyResult.String(), false)
		if err != nil {
			return nil, 0, err
		}
		if dependencyBean.OldObjectId == id {
			indexOfDependencyRemoved = dependencyBean.Index
			continue
		}
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies, indexOfDependencyRemoved, nil
}

//Patch resource object dependencies and related method ends
//

//
// create update common methods starts

func (impl *DevtronResourceServiceImpl) createOrUpdateDevtronResourceObject(ctx context.Context, requirementReq *bean.ResourceObjectRequirementRequest,
	devtronResourceSchema *repository.DevtronResourceSchema, devtronResourceObject *repository.DevtronResourceObject, auditPaths []string) (err error) {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "createOrUpdateDevtronResourceObject")
	defer span.End()
	tx, err := impl.devtronResourceObjectRepository.StartTx()
	defer impl.devtronResourceObjectRepository.RollbackTx(tx)
	if err != nil {
		impl.logger.Errorw("error encountered in db tx, createOrUpdateDevtronResourceObject", "err", err)
		return err
	}
	reqBean := requirementReq.ReqBean
	objectDataPath := requirementReq.ObjectDataPath
	skipJsonSchemaValidation := requirementReq.SkipJsonSchemaValidation
	var objectDataGeneral string
	schema := ""
	if devtronResourceSchema != nil {
		schema = devtronResourceSchema.Schema
	}
	devtronResourceObjectPresentAlready := devtronResourceObject != nil && devtronResourceObject.Id > 0
	initialObjectData := ""
	if devtronResourceObjectPresentAlready {
		initialObjectData = devtronResourceObject.ObjectData
	}

	if reqBean.ObjectData != "" {
		//we need to put the object got from UI at a path(possible values currently - overview.metadata or dependencies) since only this part is controlled from UI currently
		objectDataGeneral, err = sjson.Set(initialObjectData, objectDataPath, json.RawMessage(reqBean.ObjectData))
		if err != nil {
			impl.logger.Errorw("error in setting version in schema", "err", err, "request", reqBean)
			return err
		}
	}
	objectDataGeneral, err = impl.setDevtronManagedFieldsInObjectData(objectDataGeneral, reqBean)
	if err != nil {
		impl.logger.Errorw("error, setDevtronManagedFieldsInObjectData", "err", err, "req", reqBean)
		return err
	}
	objectDataGeneral, err = impl.setUserProvidedFieldsInObjectData(objectDataGeneral, reqBean)
	if err != nil {
		impl.logger.Errorw("error, setUserProvidedFieldsInObjectData", "err", err, "req", reqBean)
		return err
	}

	// below check is added because it might be possible that user might not have added catalog data and only updating dependencies.
	// In this case, the validation for catalog data will fail.
	if !skipJsonSchemaValidation {
		//validate user provided json with the schema
		result, err := helper.ValidateSchemaAndObjectData(schema, objectDataGeneral)
		if err != nil {
			impl.logger.Errorw("error in validating resource object json against schema", "result", result, "request", reqBean, "schema", schema, "objectData", objectDataGeneral)
			return err
		}
	}
	var auditAction repository.AuditOperationType
	if devtronResourceObjectPresentAlready {
		//object already exists, update the same
		devtronResourceObject.ObjectData = objectDataGeneral
		if len(devtronResourceObject.Identifier) == 0 {
			reqBean.Identifier, err = impl.getResourceObjectIdentifier(reqBean.Kind, reqBean.SubKind, reqBean.Version, devtronResourceObject)
			if err != nil {
				impl.logger.Errorw("error in building identifier for ResourceObject", "err", err,
					"kind", reqBean.Kind, "subKind", reqBean.SubKind, "Version", reqBean.Version, "oldObjectId", devtronResourceObject.OldObjectId)
				return err
			}
			devtronResourceObject.Identifier = reqBean.Identifier
		}
		devtronResourceObject.UpdateAuditLog(reqBean.UserId)
		err = impl.devtronResourceObjectRepository.Update(tx, devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in updating", "err", err, "req", devtronResourceObject)
			return err
		}
		auditAction = repository.AuditOperationTypeUpdate
	} else {
		if reqBean.ParentConfig != nil {
			err = helper.UpdateKindAndSubKindParentConfig(reqBean.ParentConfig)
			if err != nil {
				impl.logger.Errorw("error in updating kind, subKind for parent resource config", "err", err)
				return err
			}
			objectDataGeneral, err = impl.addParentDependencyToChildResourceObj(newCtx, reqBean.ParentConfig, objectDataGeneral)
			if err != nil {
				impl.logger.Errorw("error in updating parent resource object", "err", err, "parentConfig", reqBean.ParentConfig)
				return err
			}
		}
		//object does not exist, create new
		devtronResourceObject = &repository.DevtronResourceObject{
			DevtronResourceId:       devtronResourceSchema.DevtronResourceId,
			DevtronResourceSchemaId: devtronResourceSchema.Id,
			ObjectData:              objectDataGeneral,
		}
		// for IdType -> bean.ResourceObjectIdType; DevtronResourceObject.OldObjectId is not present
		if reqBean.IdType != bean.ResourceObjectIdType {
			devtronResourceObject.OldObjectId = reqBean.OldObjectId
		}
		reqBean.Identifier, err = impl.getResourceObjectIdentifier(reqBean.Kind, reqBean.SubKind, reqBean.Version, devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in building identifier for ResourceObject", "err", err,
				"kind", reqBean.Kind, "subKind", reqBean.SubKind, "Version", reqBean.Version, "oldObjectId", devtronResourceObject.OldObjectId)
			return err
		}
		devtronResourceObject.Identifier = reqBean.Identifier
		devtronResourceObject.CreateAuditLog(reqBean.UserId)
		err = impl.devtronResourceObjectRepository.Save(tx, devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in saving", "err", err, "req", devtronResourceObject)
			return err
		}
		auditAction = repository.AuditOperationTypeCreate
		if reqBean.ParentConfig != nil {
			err = impl.addChildDependencyToParentResourceObj(newCtx, tx, reqBean.ParentConfig, devtronResourceObject, reqBean.IdType)
			if err != nil {
				impl.logger.Errorw("error in updating parent resource object", "err", err, "parentConfig", reqBean.ParentConfig)
				return err
			}
		}
	}
	//saving audit
	impl.dtResourceObjectAuditService.SaveAudit(devtronResourceObject, auditAction, auditPaths)
	err = impl.devtronResourceObjectRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing tx createOrUpdateDevtronResourceObject", "err", err)
		return err
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) setDevtronManagedFieldsInObjectData(objectData string, reqBean *bean.DtResourceObjectInternalBean) (string, error) {
	var err error
	kindForSchema := reqBean.Kind
	if len(reqBean.SubKind) > 0 {
		kindForSchema += fmt.Sprintf("/%s", reqBean.SubKind)
	}
	objectData, err = sjson.Set(objectData, bean.KindKey, kindForSchema)
	if err != nil {
		impl.logger.Errorw("error in setting kind in schema", "err", err, "request", reqBean)
		return objectData, err
	}
	objectData, err = sjson.Set(objectData, bean.VersionKey, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in setting version in schema", "err", err, "request", reqBean)
		return objectData, err
	}
	if reqBean.IdType != "" {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectIdTypePath, reqBean.IdType)
		if err != nil {
			impl.logger.Errorw("error in setting id type in schema", "err", err, "request", reqBean)
			return objectData, err
		}
	}
	objectData, err = sjson.Set(objectData, bean.ResourceObjectIdPath, reqBean.DevtronResourceObjectDescriptorBean.GetResourceIdByIdType())
	if err != nil {
		impl.logger.Errorw("error in setting id in schema", "err", err, "request", reqBean)
		return objectData, err
	}
	if reqBean.Name != "" {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectNamePath, reqBean.Name)
		if err != nil {
			impl.logger.Errorw("error in setting id in schema", "err", err, "request", reqBean)
			return objectData, err
		}
	}
	return objectData, nil
}

func (impl *DevtronResourceServiceImpl) setUserProvidedFieldsInObjectData(objectData string, reqBean *bean.DtResourceObjectInternalBean) (string, error) {
	var err error
	f := getFuncToSetUserProvidedDataInResourceObject(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if f != nil {
		objectData, err = f(impl, objectData, reqBean)
	}
	return objectData, err
}

func (impl *DevtronResourceServiceImpl) addParentDependencyToChildResourceObj(ctx context.Context, parentConfig *bean.ResourceIdentifier,
	objectDataGeneral string) (string, error) {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "addParentDependencyToChildResourceObj")
	defer span.End()
	parentResourceObject, err := impl.getParentResourceObject(newCtx, parentConfig)
	if err != nil {
		impl.logger.Errorw("error in getting parent resource object by id or name", "err", err, "parentConfig", parentConfig)
		return objectDataGeneral, err
	}
	// generate dependency data
	parentObjectId, parentIdType := helper.GetResourceObjectIdAndType(parentResourceObject)
	parentDependency := adapter.BuildDependencyData(parentObjectId, parentResourceObject.DevtronResourceId, parentResourceObject.DevtronResourceSchemaId, 0, bean.DevtronResourceDependencyTypeParent, parentIdType)

	// patch updated dependency data
	objectDataGeneral, err = sjson.Set(objectDataGeneral, bean.ResourceObjectDependenciesPath, []*bean.DevtronResourceDependencyBean{parentDependency})
	if err != nil {
		impl.logger.Errorw("error in setting parent dependencies in child resource object", "err", err, "parentDependency", parentDependency)
		return objectDataGeneral, err
	}
	return objectDataGeneral, nil
}

func (impl *DevtronResourceServiceImpl) addChildDependencyToParentResourceObj(ctx context.Context, tx *pg.Tx, parentConfig *bean.ResourceIdentifier,
	childResourceObject *repository.DevtronResourceObject, idType bean.IdType) (err error) {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "addChildDependencyToParentResourceObj")
	defer span.End()
	parentResourceObject, err := impl.getParentResourceObject(newCtx, parentConfig)
	if err != nil {
		impl.logger.Errorw("error in getting parent resource object by id or name", "err", err, "parentConfig", parentConfig)
		return err
	}
	dependenciesOfParent, err := impl.getDependenciesInObjectDataFromJsonString(parentResourceObject.DevtronResourceSchemaId, parentResourceObject.ObjectData, true)
	if err != nil {
		impl.logger.Errorw("error in getting dependencies from json object", "err", err)
		return err
	}
	resourceIdsPresentAlready, maxIndex := getExistingDependencyIdsForResourceType(dependenciesOfParent, childResourceObject.DevtronResourceId)
	if slices.Contains(resourceIdsPresentAlready, childResourceObject.Id) {
		// dependency exists
		return nil
	}
	// generate dependency data
	childDependency := adapter.BuildDependencyData(childResourceObject.Id, childResourceObject.DevtronResourceId, childResourceObject.DevtronResourceSchemaId, maxIndex, bean.DevtronResourceDependencyTypeChild, idType)
	dependenciesOfParent = append(dependenciesOfParent, childDependency)
	// patch updated dependency data
	parentResourceObject.ObjectData, err = sjson.Set(parentResourceObject.ObjectData, bean.ResourceObjectDependenciesPath, dependenciesOfParent)
	if err != nil {
		impl.logger.Errorw("error in setting child dependencies in parent resource object", "err", err, "parentResourceObjectId", parentResourceObject.Id)
		return err
	}
	// update dependency data to db
	err = impl.devtronResourceObjectRepository.Update(tx, parentResourceObject)
	if err != nil {
		impl.logger.Errorw("error in updating child dependencies into parent resource object", "err", err, "parentResourceObjectId", parentResourceObject.Id)
		return err
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getParentResourceObject(ctx context.Context, parentConfig *bean.ResourceIdentifier) (*repository.DevtronResourceObject, error) {
	_, span := otel.Tracer("DevtronResourceService").Start(ctx, "getParentResourceObject")
	defer span.End()
	if parentConfig == nil {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceParentConfigDataNotFound, bean.ResourceParentConfigDataNotFound)
	}
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(parentConfig.ResourceKind.ToString(), parentConfig.ResourceSubKind.ToString(), parentConfig.ResourceVersion.ToString())
	if err != nil {
		impl.logger.Errorw("error in getting parent devtronResourceSchema", "err", err, "kind", parentConfig.ResourceKind, "subKind", parentConfig.ResourceSubKind, "version", parentConfig.ResourceVersion)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, err
	}
	if parentConfig.Id > 0 {
		parentResourceObject, err := impl.devtronResourceObjectRepository.FindByIdAndSchemaId(parentConfig.Id, resourceSchema.Id)
		if err != nil {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "id", parentConfig.Id)
			if util.IsErrNoRows(err) {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigId, bean.InvalidResourceParentConfigId)
			}
		}
		return parentResourceObject, err
	} else if len(parentConfig.Identifier) > 0 {
		parentResourceObject, err := impl.devtronResourceObjectRepository.FindByObjectIdentifier(parentConfig.Identifier, resourceSchema.Id)
		if err != nil {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "id", parentConfig.Id)
			if util.IsErrNoRows(err) {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigId, bean.InvalidResourceParentConfigId)
			}
		}
		return parentResourceObject, err
	} else {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigData, bean.InvalidResourceParentConfigData)
	}
}

func getExistingDependencyIdsForResourceType(childDependenciesOfParent []*bean.DevtronResourceDependencyBean, devtronResourceId int) ([]int, float64) {
	dependenciesPresentAlready := make([]int, 0, len(childDependenciesOfParent))
	var maxIndex float64
	for _, childDependency := range childDependenciesOfParent {
		maxIndex = math.Max(maxIndex, float64(childDependency.Index))
		if childDependency.DevtronResourceId == devtronResourceId {
			dependenciesPresentAlready = append(dependenciesPresentAlready, childDependency.OldObjectId)
		}
	}
	return dependenciesPresentAlready, maxIndex
}

// create update common methods ends
//

func (impl *DevtronResourceServiceImpl) getDevtronResourceIdsFromIdentifiers(identifiers []string) ([]int, error) {
	ids, err := impl.devtronResourceObjectRepository.GetIdsByIdentifiers(identifiers)
	if err != nil {
		impl.logger.Errorw("error encountered in getDevtronResourceIdsFromIdentifiers", "err", err, "identifiers", identifiers)
		return nil, err
	}
	return ids, nil
}

func (impl *DevtronResourceServiceImpl) getResourceSchemaAndExistingObject(req *bean.DevtronResourceObjectDescriptorBean) (*repository.DevtronResourceSchema, *repository.DevtronResourceObject, error) {
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, nil, err
	}
	existingResourceObject, err := impl.getExistingDevtronObject(req.Id, req.OldObjectId, resourceSchema.Id, req.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
		return resourceSchema, nil, err
	}
	return resourceSchema, existingResourceObject, nil
}

func (impl *DevtronResourceServiceImpl) getExistingDevtronObject(id, oldObjectId, devtronResourceSchemaId int, identifier string) (*repository.DevtronResourceObject, error) {
	var existingResourceObject *repository.DevtronResourceObject
	var err error
	if id > 0 {
		existingResourceObject, err = impl.devtronResourceObjectRepository.FindByIdAndSchemaId(id, devtronResourceSchemaId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "id", id, "devtronResourceSchemaId", devtronResourceSchemaId)
			return nil, err
		}
	} else if oldObjectId > 0 {
		existingResourceObject, err = impl.devtronResourceObjectRepository.FindByOldObjectId(oldObjectId, devtronResourceSchemaId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "oldObjectId", oldObjectId, "devtronResourceSchemaId", devtronResourceSchemaId)
			return nil, err
		}
	} else if len(identifier) > 0 {
		existingResourceObject, err = impl.devtronResourceObjectRepository.FindByObjectIdentifier(identifier, devtronResourceSchemaId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by identifier", "err", err, "identifier", identifier, "devtronResourceSchemaId", devtronResourceSchemaId)
			return nil, err
		}
	}
	// migrate data for resource object identifier
	if existingResourceObject != nil {
		err = impl.migrateDataForResourceObjectIdentifier(existingResourceObject)
		if err != nil {
			impl.logger.Warnw("error in service migrateDataForResourceObjectIdentifier", "err", err, "existingResourceObjectId", existingResourceObject.Id)
		}
	}
	return existingResourceObject, nil
}

func (impl *DevtronResourceServiceImpl) FindNumberOfApplicationsWithDependenciesMapped() (int, error) {
	resourceObjects, err := impl.devtronResourceObjectRepository.FindAllObjects()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching all resource objects", "err", err)
		return 0, err
	}
	if err == pg.ErrNoRows {
		return 0, util.GetApiErrorAdapter(http.StatusNotFound, "404", "no resource objects found", err.Error())
	}
	countOfApplicationsWithDependenciesMapped := 0
	for _, object := range resourceObjects {
		objectData := object.ObjectData
		dependencies, err := impl.getDependenciesInObjectDataFromJsonString(object.DevtronResourceSchemaId, objectData, true)
		if err != nil {
			impl.logger.Errorw("error in getting dependencies from json object", "err", err)
			return 0, err
		}
		if len(dependencies) > 0 {
			countOfApplicationsWithDependenciesMapped += 1
		}
	}
	return countOfApplicationsWithDependenciesMapped, nil
}

func (impl *DevtronResourceServiceImpl) deleteDevtronResourceObject(id, devtronResourceSchemaId int, identifier string) (*repository.DevtronResourceObject, error) {
	var updatedResourceObject *repository.DevtronResourceObject
	var err error
	if id > 0 {
		updatedResourceObject, err = impl.devtronResourceObjectRepository.SoftDeleteById(id, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in SoftDeleteById", "err", err, "id", id, "devtronResourceSchemaId", devtronResourceSchemaId)
			return nil, err
		}
	} else if len(identifier) > 0 {
		updatedResourceObject, err = impl.devtronResourceObjectRepository.SoftDeleteByIdentifier(identifier, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in SoftDeleteByIdentifier", "err", err, "identifier", identifier, "devtronResourceSchemaId", devtronResourceSchemaId)
			return nil, err
		}
	}
	return updatedResourceObject, nil
}

// getExistingDevtronObject : this method gets existing object if present in the db.
// If not present, returns nil object along with nil error (pg.ErrNoRows error is handled in this method only)

func (impl *DevtronResourceServiceImpl) getResourceObjectIdentifier(kind, subKind, version string, existingResourceObject *repository.DevtronResourceObject) (string, error) {
	f := getFuncToBuildIdentifierForResourceObj(kind, subKind, version) //getting function for component requested from UI
	if f == nil {
		return "", util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
	}
	return f(impl, existingResourceObject)
}

func (impl *DevtronResourceServiceImpl) migrateDataForResourceObjectIdentifier(existingResourceObject *repository.DevtronResourceObject) error {
	if len(existingResourceObject.Identifier) != 0 {
		return nil
	}
	kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(existingResourceObject.DevtronResourceSchemaId,
		impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
	if err != nil {
		impl.logger.Errorw("error in getting kind and subKind by devtronResourceSchemaId", "err", err, "devtronResourceSchemaId", existingResourceObject.DevtronResourceSchemaId)
		return err
	}
	identifier, err := impl.getResourceObjectIdentifier(kind, subKind, version, existingResourceObject)
	if err != nil {
		impl.logger.Errorw("error in building identifier for ResourceObject", "err", err,
			"kind", kind, "subKind", subKind, "version", version, "existingResourceObjectId", existingResourceObject.Id)
		return err
	}
	err = impl.devtronResourceObjectRepository.UpdateIdentifier(existingResourceObject.Id, existingResourceObject.DevtronResourceSchemaId, identifier)
	if err != nil {
		impl.logger.Errorw("error in updating identifier for ResourceObject", "err", err,
			"existingResourceObjectId", existingResourceObject.Id)
		return err
	}
	existingResourceObject.Identifier = identifier
	return nil
}

func (impl *DevtronResourceServiceImpl) findAppIdFromDependencyInfo(dependencyInfo *bean.DependencyInfo) (int, error) {
	id := dependencyInfo.Id
	var err error
	if id == 0 {
		// identifier is app_name for application dependency
		id, err = impl.appRepository.FindIdByNameAndAppType(dependencyInfo.Identifier, helper2.CustomApp)
	}
	return id, err
}

func (impl *DevtronResourceServiceImpl) getDependencyBeanFromJsonString(parentResourceType *bean.DevtronResourceTypeReq, dependency string, isLite bool) (*bean.DevtronResourceDependencyBean, error) {
	typeResult := gjson.Get(dependency, bean.TypeOfDependencyKey)
	typeOfDependency := typeResult.String()
	devtronResourceIdResult := gjson.Get(dependency, bean.DevtronResourceIdKey)
	devtronResourceId := int(devtronResourceIdResult.Int())
	schemaIdResult := gjson.Get(dependency, bean.DevtronResourceSchemaIdKey)
	schemaId := int(schemaIdResult.Int())
	var resourceKind bean.DevtronResourceKind
	var resourceVersion bean.DevtronResourceVersion
	if schemaId != 0 {
		kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(schemaId,
			impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
		if err != nil {
			impl.logger.Errorw("error in getting kind and subKind by devtronResourceSchemaId", "err", err, "devtronResourceSchemaId", schemaId)
			return nil, err
		}
		resourceKind = helper.BuildExtendedResourceKindUsingKindAndSubKind(kind, subKind)
		resourceVersion = bean.DevtronResourceVersion(version)
	}
	oldObjectIdResult := gjson.Get(dependency, bean.IdKey)
	oldObjectId := int(oldObjectIdResult.Int())
	idTypeResult := gjson.Get(dependency, bean.IdTypeKey)
	idType := bean.IdType(idTypeResult.String())
	indexResult := gjson.Get(dependency, bean.IndexKey)
	index := int(indexResult.Int())
	dependentOnIndexResult := gjson.Get(dependency, bean.DependentOnIndexKey)
	dependentOnIndex := int(dependentOnIndexResult.Int())
	dependentOnIndexArrayResult := gjson.Get(dependency, bean.DependentOnIndexesKey)
	var dependentOnIndexArray []int
	if dependentOnIndexArrayResult.IsArray() {
		dependentOnIndexArrayResult.ForEach(
			func(key, value gjson.Result) bool {
				dependentOnIndexArray = append(dependentOnIndexArray, int(value.Int()))
				return true
			},
		)
	}
	dependentOnParentIndexResult := gjson.Get(dependency, bean.DependentOnParentIndexKey)
	dependentOnParentIndex := int(dependentOnParentIndexResult.Int())
	//not handling for nested dependencies
	configDataJsonObj := gjson.Get(dependency, bean.ConfigKey)
	configData, err := impl.getConfigDataByParentResourceType(parentResourceType, configDataJsonObj.String(), isLite)
	if err != nil {
		return nil, err
	}
	return &bean.DevtronResourceDependencyBean{
		OldObjectId:             oldObjectId,
		DevtronResourceId:       devtronResourceId,
		DevtronResourceSchemaId: schemaId,
		DevtronResourceTypeReq: &bean.DevtronResourceTypeReq{
			ResourceKind:    resourceKind,
			ResourceVersion: resourceVersion,
		},
		DependentOnIndex:       dependentOnIndex,
		DependentOnIndexes:     dependentOnIndexArray,
		DependentOnParentIndex: dependentOnParentIndex,
		TypeOfDependency:       bean.DevtronResourceDependencyType(typeOfDependency),
		Config:                 configData,
		Index:                  index,
		IdType:                 idType,
	}, nil
}

func (impl *DevtronResourceServiceImpl) getConfigDataByParentResourceType(parentResourceType *bean.DevtronResourceTypeReq, configDataJsonObj string, isLite bool) (configData *bean.DependencyConfigBean, err error) {
	f := getFuncToUpdateDependencyConfigData(parentResourceType.ResourceKind.ToString(),
		parentResourceType.ResourceSubKind.ToString(), parentResourceType.ResourceVersion.ToString())
	if f != nil {
		configData = &bean.DependencyConfigBean{}
		err = f(impl, configDataJsonObj, configData, isLite)
		if err != nil {
			return nil, err
		}
	}
	return configData, nil
}

func (impl *DevtronResourceServiceImpl) getChildObjectsByParentResourceType(parentResourceType *bean.DevtronResourceTypeReq, dependency string, isLite bool, id int) (childObjects []*bean.ChildObject, err error) {
	f := getFuncToUpdateChildObjectsData(parentResourceType.ResourceKind.ToString(),
		parentResourceType.ResourceSubKind.ToString(), parentResourceType.ResourceVersion.ToString())
	if f != nil {
		childObjects, err = f(impl, dependency, isLite, id)
		if err != nil {
			return nil, err
		}
	}
	return childObjects, nil
}

func getChildInheritanceData(dependency string) ([]*bean.ChildInheritance, error) {
	childInheritance := make([]*bean.ChildInheritance, 0)
	inheritance := gjson.Get(dependency, bean.DependencyChildInheritanceKey).String()
	if len(inheritance) != 0 {
		err := json.Unmarshal([]byte(inheritance), &childInheritance)
		if err != nil {
			return childInheritance, err
		}
	}
	return childInheritance, nil
}

func (impl *DevtronResourceServiceImpl) getEnvironmentsForApplicationDependency(childInheritance []*bean.ChildInheritance, appId int) ([]*bean.Environment, error) {
	// iterating in every inheritance and getting child inheritances(for eg cd) and getting corresponding details) for now it is ["*"] we will fetch all cd (env) for that dependency
	envs := make([]*bean.Environment, 0)
	findAll := false
	for _, inheritance := range childInheritance {
		// collecting selectors here currently only ["all"] is present so will find all env names for an app but can be modified in future
		findAll = slices.Contains(inheritance.Selector, bean.DefaultCdPipelineSelector)
	}
	if findAll {
		pipelines, err := impl.pipelineRepository.FindEnvNameAndIdByAppId(appId)
		if err != nil {
			impl.logger.Errorw("error encountered in getEnvironmentsForApplicationDependency", "err", err)
			return envs, err
		}
		for _, pipeline := range pipelines {
			env := adapter.BuildEnvironmentBasicData(pipeline.Environment.Name, pipeline.EnvironmentId)
			envs = append(envs, env)
		}
	} else {
		// if specific , will have to find corresponding ids and get the details
	}

	return envs, nil
}
