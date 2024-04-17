package devtronResource

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	repositoryAdapter "github.com/devtron-labs/devtron/pkg/devtronResource/repository/adapter"
	"github.com/devtron-labs/devtron/util/response/pagination"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slices"
	"math"
	"net/http"
	"time"

	apiBean "github.com/devtron-labs/devtron/api/bean"
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
	//    - false for detailed data   // provides the complete bean.DevtronResourceObjectBean
	ListResourceObjectByKindAndVersion(kind, subKind, version string, isLite, fetchChild bool) (pagination.PaginatedResponse[*bean.DevtronResourceObjectGetAPIBean], error)
	// GetResourceObject will get the bean.DevtronResourceObjectBean based on the given bean.DevtronResourceObjectDescriptorBean
	GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectGetAPIBean, error)
	// CreateResourceObject creates resource object corresponding to kind,version according to bean.DevtronResourceObjectBean
	CreateResourceObject(ctx context.Context, reqBean *bean.DevtronResourceObjectBean) error
	CreateOrUpdateResourceObject(ctx context.Context, reqBean *bean.DevtronResourceObjectBean) error
	// PatchResourceObject supports json patch operation corresponding to kind,subKind,version on json object data takes in ([]PatchQuery in DevtronResourceObjectBean), returns error if any
	PatchResourceObject(ctx context.Context, req *bean.DevtronResourceObjectBean) (*bean.SuccessResponse, error)
	// DeleteResourceObject deletes resource object corresponding to kind,version, id or name
	DeleteResourceObject(ctx context.Context, req *bean.DevtronResourceObjectDescriptorBean) (*bean.SuccessResponse, error)
	GetResourceDependencies(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error)
	CreateOrUpdateResourceDependencies(ctx context.Context, req *bean.DevtronResourceObjectBean) error
	GetSchema(req *bean.DevtronResourceBean) (*bean.DevtronResourceBean, error)
	UpdateSchema(req *bean.DevtronResourceSchemaRequestBean, dryRun bool) (*bean.UpdateSchemaResponseBean, error)
	DeleteObjectAndItsDependency(oldObjectId int, kind, subKind bean.DevtronResourceKind,
		version bean.DevtronResourceVersion, updatedBy int32) error
	FindNumberOfApplicationsWithDependenciesMapped() (int, error)
}

type DevtronResourceServiceImpl struct {
	logger                               *zap.SugaredLogger
	devtronResourceRepository            repository.DevtronResourceRepository
	devtronResourceSchemaRepository      repository.DevtronResourceSchemaRepository
	devtronResourceObjectRepository      repository.DevtronResourceObjectRepository
	devtronResourceSchemaAuditRepository repository.DevtronResourceSchemaAuditRepository
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository
	appRepository                        appRepository.AppRepository //TODO: remove repo dependency
	pipelineRepository                   pipelineConfig.PipelineRepository
	userRepository                       repository2.UserRepository
	appListingRepository                 repository3.AppListingRepository
	devtronResourcesMapById              map[int]*repository.DevtronResource       //map of id and its object
	devtronResourcesMapByKind            map[string]*repository.DevtronResource    //map of kind and its object
	devtronResourcesSchemaMapById        map[int]*repository.DevtronResourceSchema //map of id and its object
}

func NewDevtronResourceServiceImpl(logger *zap.SugaredLogger,
	devtronResourceRepository repository.DevtronResourceRepository,
	devtronResourceSchemaRepository repository.DevtronResourceSchemaRepository,
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository,
	devtronResourceSchemaAuditRepository repository.DevtronResourceSchemaAuditRepository,
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository,
	appRepository appRepository.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appListingRepository repository3.AppListingRepository,
	userRepository repository2.UserRepository) (*DevtronResourceServiceImpl, error) {
	impl := &DevtronResourceServiceImpl{
		logger:                               logger,
		devtronResourceRepository:            devtronResourceRepository,
		devtronResourceSchemaRepository:      devtronResourceSchemaRepository,
		devtronResourceObjectRepository:      devtronResourceObjectRepository,
		devtronResourceSchemaAuditRepository: devtronResourceSchemaAuditRepository,
		devtronResourceObjectAuditRepository: devtronResourceObjectAuditRepository,
		appRepository:                        appRepository,
		pipelineRepository:                   pipelineRepository,
		userRepository:                       userRepository,
		appListingRepository:                 appListingRepository,
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

func (impl *DevtronResourceServiceImpl) ListResourceObjectByKindAndVersion(kind, subKind, version string, isLite, fetchChild bool) (pagination.PaginatedResponse[*bean.DevtronResourceObjectGetAPIBean], error) {
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

	f := listApiResourceKindFunc(kind)
	if f == nil {
		impl.logger.Errorw("error kind type not supported", "err", err, "kind", kind)
		return response, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrComponent, bean.InvalidResourceKindOrComponent)
	}
	response.Data, err = f(impl, resourceObjects, childResourceObjects, resourceObjectIndexChildMap, isLite)
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
		childDependencies := getSpecificDependenciesInObjectDataFromJsonString(resourceObject.ObjectData, bean.DevtronResourceDependencyTypeChild)
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
		childDependencies := getSpecificDependenciesInObjectDataFromJsonString(resourceObject.ObjectData, bean.DevtronResourceDependencyTypeChild)
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

func (impl *DevtronResourceServiceImpl) listReleaseTracks(resourceObjects, childObjects []*repository.DevtronResourceObject, resourceObjectIndexChildMap map[int][]int,
	isLite bool) ([]*bean.DevtronResourceObjectGetAPIBean, error) {
	resp := make([]*bean.DevtronResourceObjectGetAPIBean, 0, len(resourceObjects))
	for i := range resourceObjects {
		resourceData := &bean.DevtronResourceObjectGetAPIBean{
			DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{},
			DevtronResourceObjectBasicDataBean: &bean.DevtronResourceObjectBasicDataBean{
				Overview: &bean.ResourceOverview{},
			},
		}
		resourceData.IdType = bean.IdType(gjson.Get(resourceObjects[i].ObjectData, bean.ResourceObjectIdTypePath).String())
		if resourceData.IdType == bean.ResourceObjectIdType {
			resourceData.OldObjectId = resourceObjects[i].Id
		} else {
			resourceData.OldObjectId = resourceObjects[i].OldObjectId
		}
		resourceData.Overview.Description = gjson.Get(resourceObjects[i].ObjectData, bean.ResourceObjectDescriptionPath).String()
		resourceData.Name = gjson.Get(resourceObjects[i].ObjectData, bean.ResourceObjectNamePath).String()
		childIndexes := resourceObjectIndexChildMap[i]
		for _, childIndex := range childIndexes {
			childObject := childObjects[childIndex]
			childData := &bean.DevtronResourceObjectGetAPIBean{
				DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{},
				DevtronResourceObjectBasicDataBean:  &bean.DevtronResourceObjectBasicDataBean{},
			}
			if !isLite {
				err := impl.updateCompleteReleaseDataInResourceObj(nil, childObject, childData)
				if err != nil {
					impl.logger.Errorw("error in getting detailed resource data", "resourceObjectId", resourceObjects[i].Id, "err", err)
					return nil, err
				}
			} else {
				err := impl.updateReleaseOverviewDataInResourceObj(nil, childObject, childData)
				if err != nil {
					impl.logger.Errorw("error in getting overview data", "err", err)
					return nil, err
				}
			}
			resourceData.ChildObjects = append(resourceData.ChildObjects, childData)
		}
		err := impl.updateReleaseTrackOverviewDataInResourceObj(nil, resourceObjects[i], resourceData)
		if err != nil {
			impl.logger.Errorw("error in getting detailed resource data", "resourceObjectId", resourceObjects[i].Id, "err", err)
			return nil, err
		}
		resp = append(resp, resourceData)
	}
	return resp, nil
}

func (impl *DevtronResourceServiceImpl) GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectGetAPIBean, error) {
	resp := &bean.DevtronResourceObjectGetAPIBean{
		DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{},
		DevtronResourceObjectBasicDataBean:  &bean.DevtronResourceObjectBasicDataBean{},
	}
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, err
	}
	resp.Schema = resourceSchema.Schema
	existingResourceObject, err := impl.getExistingDevtronObject(req.Id, req.OldObjectId, resourceSchema.Id, req.Name, req.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
		return nil, err
	}
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
		f := getApiResourceKindUIComponentFunc(req.Kind, component.ToString()) //getting function for component requested from UI
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
	return resourceObject, nil
}

func (impl *DevtronResourceServiceImpl) CreateResourceObject(ctx context.Context, reqBean *bean.DevtronResourceObjectBean) error {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "CreateResourceObject")
	defer span.End()
	err := validateCreateResourceRequest(reqBean)
	if err != nil {
		return err
	}
	err = impl.populateDefaultValuesToRequestBean(reqBean)
	if err != nil {
		return err
	}
	err = setIdentifierValueBasedOnKind(reqBean)
	if err != nil {
		impl.logger.Errorw("error encountered in CreateResourceObject", "req", reqBean, "err", err)
		return err
	}
	//getting schema latest from the db (not getting it from FE for edge cases when schema has got updated
	//just before an object update is requested)
	devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema from db", "err", err, "request", reqBean)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return err
	}
	exists, err := impl.checkIfExistingDevtronObject(reqBean.Id, devtronResourceSchema.Id, reqBean.Name, reqBean.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", reqBean)
		return err
	}
	if exists {
		impl.logger.Errorw("error encountered in CreateResourceObject", "request", reqBean, "err", err)
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceAlreadyExistsMessage, bean.ResourceAlreadyExistsMessage)
	}
	resourceObjReq := adapter.GetResourceObjectRequirementRequest(reqBean, "", false)
	return impl.createOrUpdateDevtronResourceObject(newCtx, resourceObjReq, devtronResourceSchema, nil, nil)
}

func (impl *DevtronResourceServiceImpl) CreateOrUpdateResourceObject(ctx context.Context, reqBean *bean.DevtronResourceObjectBean) error {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "CreateOrUpdateResourceObject")
	defer span.End()
	//getting schema latest from the db (not getting it from FE for edge cases when schema has got updated
	//just before an object update is requested)
	devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema from db", "err", err, "request", reqBean)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return err
	}
	devtronResourceObject, err := impl.getExistingDevtronObject(reqBean.Id, reqBean.OldObjectId, devtronResourceSchema.Id, reqBean.Name, reqBean.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", reqBean)
		return err
	}
	resourceObjReq := adapter.GetResourceObjectRequirementRequest(reqBean, bean.ResourceObjectMetadataPath, false)
	return impl.createOrUpdateDevtronResourceObject(newCtx, resourceObjReq, devtronResourceSchema, devtronResourceObject, nil)
}

func (impl *DevtronResourceServiceImpl) PatchResourceObject(ctx context.Context, req *bean.DevtronResourceObjectBean) (*bean.SuccessResponse, error) {
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, err
	}
	existingResourceObject, err := impl.getExistingDevtronObject(req.Id, req.OldObjectId, resourceSchema.Id, req.Name, req.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
		return nil, err
	}
	// performing json patch operations
	objectData := existingResourceObject.ObjectData
	auditPaths := make([]string, 0, len(req.PatchQuery))
	jsonPath := ""
	for _, query := range req.PatchQuery {
		objectData, jsonPath, err = impl.performPatchOperation(objectData, query)
		if err != nil {
			impl.logger.Errorw("error encountered in PatchResourceObject", "query", query, "err", err)
			return nil, err
		}
		auditPaths = append(auditPaths, jsonPath)
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
	impl.saveAudit(existingResourceObject, repository.AuditOperationTypePatch, auditPaths)
	return adapter.GetSuccessPassResponse(), nil
}

func (impl *DevtronResourceServiceImpl) DeleteResourceObject(ctx context.Context, req *bean.DevtronResourceObjectDescriptorBean) (*bean.SuccessResponse, error) {
	devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, err
	}
	exists, err := impl.checkIfExistingDevtronObject(req.Id, devtronResourceSchema.Id, req.Name, req.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
		return nil, err
	}
	if !exists {
		impl.logger.Errorw("error encountered in DeleteResourceObject", "request", req, "err", err)
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
	}
	devtronResourceObject, err := impl.deleteDevtronResourceObject(req.Id, devtronResourceSchema.Id, req.Name, "")
	if err != nil {
		impl.logger.Errorw("error in DeleteResourceObject", "id", req.Id, "devtronResourceSchemaId", devtronResourceSchema.Id, "name", req.Name)
		return nil, err
	}

	impl.saveAudit(devtronResourceObject, repository.AuditOperationTypeDeleted, nil)

	return adapter.GetSuccessPassResponse(), nil
}

func (impl *DevtronResourceServiceImpl) GetResourceDependencies(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error) {
	response := &bean.DevtronResourceObjectBean{
		Dependencies:      make([]*bean.DevtronResourceDependencyBean, 0),
		ChildDependencies: make([]*bean.DevtronResourceDependencyBean, 0),
	}

	resourceSchemaOfRequestObject, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, err
	}
	existingResourceObject, err := impl.getExistingDevtronObject(req.Id, req.OldObjectId, resourceSchemaOfRequestObject.Id, req.Name, req.Identifier)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
		return nil, err
	}
	if existingResourceObject == nil || existingResourceObject.Id < 1 {
		//Since we have not added a migration for saving resource objects its always possible that resource object is
		//not added but dependency is added and those resource objects should be included in downstream
		existingResourceObject = &repository.DevtronResourceObject{
			OldObjectId:             req.OldObjectId,
			Name:                    req.Name,
			DevtronResourceId:       resourceSchemaOfRequestObject.DevtronResourceId,
			DevtronResourceSchemaId: resourceSchemaOfRequestObject.Id,
			ObjectData:              bean.EmptyJsonObject,
		}
	}

	dependenciesOfParent := getDependenciesInObjectDataFromJsonString(existingResourceObject.ObjectData)

	argValuesToGetDownstream := make([]interface{}, 0, len(dependenciesOfParent)+1)
	argTypesToGetDownstream := make([]string, 0, len(dependenciesOfParent)+1)
	schemaIdsOfArgsToGetDownstream := make([]int, 0, len(dependenciesOfParent)+1)

	// adding request data for getting downstream args of request resource object
	appendDependencyArgDetails(&argValuesToGetDownstream, &argTypesToGetDownstream, &schemaIdsOfArgsToGetDownstream,
		req.Name, req.OldObjectId, resourceSchemaOfRequestObject.Id)

	nonChildDependenciesOfParent, mapOfNonChildDependenciesAndIndex, childDependenciesOfParent, mapOfChildDependenciesAndIndex,
		appIdsToGetMetadata, pipelineIdsToGetMetadata, maxIndexInNonChildDependencies, err :=
		impl.separateNonChildAndChildDependencies(dependenciesOfParent, &argValuesToGetDownstream, &argTypesToGetDownstream, &schemaIdsOfArgsToGetDownstream)

	err = impl.addChildCdPipelinesNotPresentInObjects(&childDependenciesOfParent, mapOfChildDependenciesAndIndex, &pipelineIdsToGetMetadata, existingResourceObject,
		&argValuesToGetDownstream, &argTypesToGetDownstream, &schemaIdsOfArgsToGetDownstream)
	if err != nil {
		impl.logger.Errorw("error, addChildCdPipelinesNotPresentInObjects", "err", err, "childDependencies", childDependenciesOfParent)
		return nil, err
	}

	err = impl.updateChildDependenciesWithOwnDependenciesData(req.Name, req.OldObjectId, resourceSchemaOfRequestObject.Id,
		mapOfChildDependenciesAndIndex, childDependenciesOfParent, &appIdsToGetMetadata, &pipelineIdsToGetMetadata)
	if err != nil {
		impl.logger.Errorw("error, updateChildDependenciesWithOwnDependenciesData", "err", err, "parentName", req.Name,
			"parentOldObjectId", req.OldObjectId, "parentSchemaId", resourceSchemaOfRequestObject.Id)
		return nil, err
	}

	downstreamDependencyObjects, err := impl.getDownstreamDependencyObjects(argValuesToGetDownstream, argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream)
	if err != nil {
		impl.logger.Errorw("err, getDownstreamDependencyObjects", "err", err, "argValues", argValuesToGetDownstream,
			"argTypes", argTypesToGetDownstream, "schemaIds", schemaIdsOfArgsToGetDownstream)
		return nil, err
	}

	indexesToCheckInDownstreamObjectForChildDependency, err :=
		impl.updateNonChildDependenciesWithDownstreamDependencies(downstreamDependencyObjects, mapOfNonChildDependenciesAndIndex, &nonChildDependenciesOfParent,
			&appIdsToGetMetadata, &pipelineIdsToGetMetadata, maxIndexInNonChildDependencies)
	if err != nil {
		impl.logger.Errorw("error, updateNonChildDependenciesWithDownstreamDependencies", "err", err,
			"downstreamDependencyObjects", downstreamDependencyObjects)
		return nil, err
	}

	impl.updateChildDependenciesWithDownstreamDependencies(indexesToCheckInDownstreamObjectForChildDependency,
		downstreamDependencyObjects, &pipelineIdsToGetMetadata, mapOfNonChildDependenciesAndIndex, mapOfChildDependenciesAndIndex,
		nonChildDependenciesOfParent, childDependenciesOfParent)
	mapOfAppsMetadata, mapOfCdPipelinesMetadata, err := impl.getMapOfAppAndCdPipelineMetadata(appIdsToGetMetadata, pipelineIdsToGetMetadata)
	if err != nil {
		impl.logger.Errorw("error, getMapOfAppAndCdPipelineMetadata", "err", "appIds", appIdsToGetMetadata,
			"pipelineIds", pipelineIdsToGetMetadata)
		return nil, err
	}
	nonChildDependenciesOfParent = impl.getUpdatedDependencyArrayWithMetadata(nonChildDependenciesOfParent, mapOfAppsMetadata, mapOfCdPipelinesMetadata)
	childDependenciesOfParent = impl.getUpdatedDependencyArrayWithMetadata(childDependenciesOfParent, mapOfAppsMetadata, mapOfCdPipelinesMetadata)
	response.Dependencies = nonChildDependenciesOfParent
	response.ChildDependencies = childDependenciesOfParent
	return response, nil
}

func (impl *DevtronResourceServiceImpl) CreateOrUpdateResourceDependencies(ctx context.Context, req *bean.DevtronResourceObjectBean) error {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "CreateOrUpdateResourceDependencies")
	defer span.End()
	err := impl.validateDependencies(req)
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
		keyToGetSchema := getKeyForADependencyMap(request.Name, request.OldObjectId, request.SchemaId)
		devtronResourceSchema := existingObjectsMap[keyToGetSchema]
		resourceObjReq := adapter.GetResourceObjectRequirementRequest(request, bean.ResourceObjectDependenciesPath, true)
		err = impl.createOrUpdateDevtronResourceObject(newCtx, resourceObjReq, allRequestSchemas[i], devtronResourceSchema, []string{bean.ResourceObjectDependenciesPath})
		if err != nil {
			impl.logger.Errorw("error, createOrUpdateDevtronResourceObject", "err", err, "request", request)
			return err
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) DeleteObjectAndItsDependency(oldObjectId int, kind, subKind bean.DevtronResourceKind,
	version bean.DevtronResourceVersion, updatedBy int32) error {
	dbConnection := impl.devtronResourceObjectRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in getting transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(kind.ToString(), subKind.ToString(), version.ToString())
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "kind", kind, "subKind", subKind, "version", version)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return err
	}
	err = impl.devtronResourceObjectRepository.DeleteObject(tx, oldObjectId, resourceSchema.DevtronResourceId, updatedBy)
	if err != nil {
		impl.logger.Errorw("error, DeleteObject", "err", err, "oldObjectId", oldObjectId, "devtronResourceId", resourceSchema.DevtronResourceId)
		return err
	}
	err = impl.devtronResourceObjectRepository.DeleteDependencyInObjectData(tx, oldObjectId, resourceSchema.DevtronResourceId, updatedBy)
	if err != nil {
		impl.logger.Errorw("error, DeleteDependencyInObjectData", "err", err, "oldObjectId", oldObjectId, "devtronResourceId", resourceSchema.DevtronResourceId)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction, DeleteObjectAndItsDependency", "err", err)
		return err
	}
	return nil
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
		dependencies := getDependenciesInObjectDataFromJsonString(objectData)
		if len(dependencies) > 0 {
			countOfApplicationsWithDependenciesMapped += 1
		}
	}
	return countOfApplicationsWithDependenciesMapped, nil
}

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
		result, err := validateSchemaAndObjectData(schema, objectDataGeneral)
		if err != nil {
			impl.logger.Errorw("error in validating resource object json against schema", "result", result, "request", reqBean, "schema", schema, "objectData", objectDataGeneral)
			return err
		}
	}
	var auditAction repository.AuditOperationType
	if devtronResourceObjectPresentAlready {
		//object already exists, update the same
		if len(reqBean.Name) != 0 && reqBean.Name != devtronResourceObject.Name {
			// for release type resource the resource name can be updated
			devtronResourceObject.Name = reqBean.Name
		}
		devtronResourceObject.ObjectData = objectDataGeneral
		devtronResourceObject.UpdateAuditLog(reqBean.UserId)
		err = impl.devtronResourceObjectRepository.Update(tx, devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in updating", "err", err, "req", devtronResourceObject)
			return err
		}
		auditAction = repository.AuditOperationTypeUpdate
	} else {
		if reqBean.ParentConfig != nil {
			objectDataGeneral, err = impl.addParentDependencyToChildResourceObj(newCtx, reqBean.ParentConfig, objectDataGeneral, reqBean.Version)
			if err != nil {
				impl.logger.Errorw("error in updating parent resource object", "err", err, "parentConfig", reqBean.ParentConfig)
				return err
			}
		}
		//object does not exist, create new
		devtronResourceObject = &repository.DevtronResourceObject{
			Name:                    reqBean.Name,
			DevtronResourceId:       devtronResourceSchema.DevtronResourceId,
			DevtronResourceSchemaId: devtronResourceSchema.Id,
			ObjectData:              objectDataGeneral,
			Identifier:              reqBean.Identifier,
		}
		// for IdType -> bean.ResourceObjectIdType; DevtronResourceObject.OldObjectId is not present
		if reqBean.IdType != bean.ResourceObjectIdType {
			devtronResourceObject.OldObjectId = reqBean.OldObjectId
		}
		devtronResourceObject.CreateAuditLog(reqBean.UserId)
		err = impl.devtronResourceObjectRepository.Save(tx, devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in saving", "err", err, "req", devtronResourceObject)
			return err
		}
		auditAction = repository.AuditOperationTypeCreate
		if reqBean.ParentConfig != nil {
			err = impl.addChildDependencyToParentResourceObj(newCtx, tx, reqBean.ParentConfig, devtronResourceObject, reqBean.Version, reqBean.IdType)
			if err != nil {
				impl.logger.Errorw("error in updating parent resource object", "err", err, "parentConfig", reqBean.ParentConfig)
				return err
			}
		}
	}
	//saving audit
	impl.saveAudit(devtronResourceObject, auditAction, auditPaths)
	err = impl.devtronResourceObjectRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing tx createOrUpdateDevtronResourceObject", "err", err)
		return err
	}
	return nil
}
func (impl *DevtronResourceServiceImpl) deleteDevtronResourceObject(id, devtronResourceSchemaId int, name, identifier string) (*repository.DevtronResourceObject, error) {
	var updatedResourceObject *repository.DevtronResourceObject
	var err error
	if id > 0 {
		updatedResourceObject, err = impl.devtronResourceObjectRepository.SoftDeleteById(id, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in SoftDeleteById", "err", err, "id", id, "devtronResourceSchemaId", devtronResourceSchemaId)
			return nil, err
		}
	} else if len(name) > 0 {
		updatedResourceObject, err = impl.devtronResourceObjectRepository.SoftDeleteByName(name, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in SoftDeleteByName", "err", err, "name", name, "devtronResourceSchemaId", devtronResourceSchemaId)
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

func (impl *DevtronResourceServiceImpl) validateDependencies(req *bean.DevtronResourceObjectBean) error {
	allDependenciesToBeValidated := make([]*bean.DevtronResourceDependencyBean, 0, len(req.Dependencies)+2*len(req.ChildDependencies))
	allDependenciesToBeValidated = req.Dependencies
	for _, childDependency := range req.ChildDependencies {
		allDependenciesToBeValidated = append(allDependenciesToBeValidated, childDependency)
		//here assuming that dependencies of childDependency further don't have their own dependencies, i.e. only one level of nesting in resources
		allDependenciesToBeValidated = append(allDependenciesToBeValidated, childDependency.Dependencies...)
	}
	mapOfSchemaIdAndDependencyIds := make(map[int][]int) //map of devtronResourceSchemaId and all its dependencies present in request
	for _, dependency := range allDependenciesToBeValidated {
		mapOfSchemaIdAndDependencyIds[dependency.DevtronResourceSchemaId] =
			append(mapOfSchemaIdAndDependencyIds[dependency.DevtronResourceSchemaId], dependency.OldObjectId)
	}
	invalidSchemaIds := make([]int, 0, len(mapOfSchemaIdAndDependencyIds))
	var invalidAppIds []int
	var invalidCdPipelineIds []int
	var err error
	for devtronResourceSchemaId, dependencyIds := range mapOfSchemaIdAndDependencyIds {
		if devtronResourceSchema, ok := impl.devtronResourcesSchemaMapById[devtronResourceSchemaId]; ok {
			switch devtronResourceSchema.DevtronResource.Kind {
			case bean.DevtronResourceDevtronApplication.ToString():
				invalidAppIds, err = impl.getAppsAndReturnNotFoundIds(dependencyIds)
				if err != nil {
					impl.logger.Errorw("error, getAppsAndReturnNotFoundIds", "err", err, "appIds", dependencyIds)
					return err
				}
			case bean.DevtronResourceCdPipeline.ToString():
				invalidCdPipelineIds, err = impl.getCdPipelinesAndReturnNotFoundIds(dependencyIds)
				if err != nil {
					impl.logger.Errorw("error, getCdPipelinesAndReturnNotFoundIds", "err", err, "pipelineIds", dependencyIds)
					return err
				}
			default:
				invalidSchemaIds = append(invalidSchemaIds, devtronResourceSchemaId)
			}
		} else {
			invalidSchemaIds = append(invalidSchemaIds, devtronResourceSchemaId)
		}
	}
	internalMessage := ""
	isRequestInvalid := false
	if len(invalidSchemaIds) > 0 {
		isRequestInvalid = true
		internalMessage += fmt.Sprintf("invalid schemaIds : %v\n", invalidSchemaIds)
	}
	if len(invalidAppIds) > 0 {
		isRequestInvalid = true
		internalMessage += fmt.Sprintf("invalid appIds : %v\n", invalidAppIds)
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

func (impl *DevtronResourceServiceImpl) getUpdatedDependenciesRequestData(req *bean.DevtronResourceObjectBean) ([]*bean.DevtronResourceObjectBean,
	[]*repository.DevtronResourceSchema, map[string]*repository.DevtronResourceObject, error) {
	parentDevtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		if util.IsErrNoRows(err) {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
		}
		return nil, nil, nil, err
	}
	for i := range req.Dependencies {
		req.Dependencies[i].Metadata = nil //emptying in case UI sends the data back
		//since child dependencies are included separately in the payload and downstream are not declared explicitly setting this as upstream
		req.Dependencies[i].TypeOfDependency = bean.DevtronResourceDependencyTypeUpstream
	}
	allRequests := make([]*bean.DevtronResourceObjectBean, 0, len(req.ChildDependencies)+1)
	allRequestSchemas := make([]*repository.DevtronResourceSchema, 0, len(req.ChildDependencies)+1)

	allArgValues := make([]interface{}, 0, len(req.ChildDependencies)+1)
	allArgTypes := make([]string, 0, len(req.ChildDependencies)+1)
	devtronSchemaIdsForAllArgs := make([]int, 0, len(req.ChildDependencies)+1)

	//adding oldObjectId and Name for main request
	appendDbObjectArgDetails(&allArgValues, &allArgTypes, &devtronSchemaIdsForAllArgs, req.Name, req.OldObjectId, parentDevtronResourceSchema.Id)

	for _, childDependency := range req.ChildDependencies {
		childDependency.Metadata = nil //emptying in case UI sends the data back
		for i := range childDependency.Dependencies {
			childDependency.Dependencies[i].Metadata = nil                                                //emptying in case UI sends the data back
			childDependency.Dependencies[i].TypeOfDependency = bean.DevtronResourceDependencyTypeUpstream //assuming one level of nesting
		}
		//adding info of parent dependency in this child dependency's dependencies
		childDependency.Dependencies = append(childDependency.Dependencies, &bean.DevtronResourceDependencyBean{
			OldObjectId:             req.OldObjectId,
			Name:                    req.Name,
			DevtronResourceSchemaId: parentDevtronResourceSchema.Id,
			DevtronResourceId:       parentDevtronResourceSchema.DevtronResourceId,
			TypeOfDependency:        bean.DevtronResourceDependencyTypeParent,
		})

		//getting devtronResourceSchema for this child dependency
		devtronResourceSchema := impl.devtronResourcesSchemaMapById[childDependency.DevtronResourceSchemaId]
		kind, subKind := impl.getKindSubKindOfResourceBySchemaObject(devtronResourceSchema)
		marshaledDependencies, err := json.Marshal(childDependency.Dependencies)
		if err != nil {
			impl.logger.Errorw("error in marshaling dependencies", "err", err, "request", req)
			return nil, nil, nil, err
		}
		reqForChildDependency := &bean.DevtronResourceObjectBean{
			DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{
				Kind:        kind,
				SubKind:     subKind,
				Version:     devtronResourceSchema.Version,
				OldObjectId: childDependency.OldObjectId,
				Name:        childDependency.Name,
				SchemaId:    childDependency.DevtronResourceSchemaId,
			},
			Dependencies: childDependency.Dependencies,
			ObjectData:   string(marshaledDependencies),
		}
		allRequestSchemas = append(allRequestSchemas, devtronResourceSchema)
		allRequests = append(allRequests, reqForChildDependency)

		//need to add this child dependency in parent
		childDependency.Dependencies = nil //since we only need to add child dependency for parent-child relationship and not keeping nested dependencies in every object
		childDependency.TypeOfDependency = bean.DevtronResourceDependencyTypeChild
		req.Dependencies = append(req.Dependencies, childDependency)

		//adding oldObjectIds or names for getting existing objects
		appendDbObjectArgDetails(&allArgValues, &allArgTypes, &devtronSchemaIdsForAllArgs, childDependency.Name, childDependency.OldObjectId, childDependency.DevtronResourceSchemaId)
	}

	marshaledDependencies, err := json.Marshal(req.Dependencies)
	if err != nil {
		impl.logger.Errorw("error in marshaling dependencies", "err", err, "request", req)
		return nil, nil, nil, err
	}
	req.ObjectData = string(marshaledDependencies)
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

func (impl *DevtronResourceServiceImpl) saveAudit(devtronResourceObject *repository.DevtronResourceObject, auditAction repository.AuditOperationType, auditPath []string) {
	auditModel := repositoryAdapter.GetResourceObjectAudit(devtronResourceObject, auditAction, auditPath)
	err := impl.devtronResourceObjectAuditRepository.Save(auditModel)
	if err != nil { //only logging not propagating to user
		impl.logger.Warnw("error in saving devtronResourceObject audit", "err", err, "auditModel", auditModel)
	}
}

// getExistingDevtronObject : this method gets existing object if present in the db.
// If not present, returns nil object along with nil error (pg.ErrNoRows error is handled in this method only)
func (impl *DevtronResourceServiceImpl) getExistingDevtronObject(id, oldObjectId, devtronResourceSchemaId int, name, identifier string) (*repository.DevtronResourceObject, error) {
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
	} else if len(name) > 0 {
		existingResourceObject, err = impl.devtronResourceObjectRepository.FindByObjectName(name, devtronResourceSchemaId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "name", name, "devtronResourceSchemaId", devtronResourceSchemaId)
			return nil, err
		}
	}
	return existingResourceObject, nil
}

// checkIfExistingDevtronObject : this method check if it is existing object in the db.
func (impl *DevtronResourceServiceImpl) checkIfExistingDevtronObject(id, devtronResourceSchemaId int, name, identifier string) (bool, error) {
	var exists bool
	var err error
	if id > 0 {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistById(id, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in checking object exists by id or name", "err", err, "id", id, "devtronResourceSchemaId", devtronResourceSchemaId)
			return false, err
		}
	} else if len(name) > 0 {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistByName(name, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in checking object exists by id or name", "err", err, "name", name, "devtronResourceSchemaId", devtronResourceSchemaId)
			return false, err
		}
	} else if len(identifier) > 0 {
		exists, err = impl.devtronResourceObjectRepository.CheckIfExistByIdentifier(identifier, devtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error in checking object exists by identifier", "err", err, "identifier", identifier, "devtronResourceSchemaId", devtronResourceSchemaId)
			return false, err
		}
	}
	return exists, nil
}

func (impl *DevtronResourceServiceImpl) getParentResourceObject(ctx context.Context, parentConfig *bean.ResourceParentConfig, version string) (*repository.DevtronResourceObject, error) {
	_, span := otel.Tracer("DevtronResourceService").Start(ctx, "getParentResourceObject")
	defer span.End()
	if parentConfig == nil || parentConfig.Data == nil {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceParentConfigDataNotFound, bean.ResourceParentConfigDataNotFound)
	}

	if parentConfig.Data.Id > 0 {
		parentResourceObject, err := impl.devtronResourceObjectRepository.FindById(parentConfig.Data.Id)
		if err != nil {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "id", parentConfig.Data.Id)
			if util.IsErrNoRows(err) {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigId, bean.InvalidResourceParentConfigId)
			}
		}
		return parentResourceObject, err
	} else if len(parentConfig.Data.Name) > 0 {
		parentResource := impl.devtronResourcesMapByKind[parentConfig.Type.ToString()]
		kind, subKind := impl.getKindSubKindOfResource(parentResource)
		// TODO Asutosh: discuss and update this logic
		// NOTE: considering that the child resource and the parent resource will share the same version;
		// taking the child version for searching parent resource
		resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(kind, subKind, version)
		if err != nil {
			impl.logger.Errorw("error in getting parent devtronResourceSchema", "err", err, "kind", kind, "subKind", subKind, "version", version)
			if util.IsErrNoRows(err) {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
			}
			return nil, err
		}
		parentResourceObject, err := impl.devtronResourceObjectRepository.FindByObjectName(parentConfig.Data.Name, resourceSchema.Id)
		if err != nil {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "id", parentConfig.Data.Id)
			if util.IsErrNoRows(err) {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigId, bean.InvalidResourceParentConfigId)
			}
		}
		return parentResourceObject, err
	} else {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigData, bean.InvalidResourceParentConfigData)
	}
}

func (impl *DevtronResourceServiceImpl) addChildDependencyToParentResourceObj(ctx context.Context, tx *pg.Tx, parentConfig *bean.ResourceParentConfig,
	childResourceObject *repository.DevtronResourceObject, version string, idType bean.IdType) (err error) {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "addChildDependencyToParentResourceObj")
	defer span.End()
	parentResourceObject, err := impl.getParentResourceObject(newCtx, parentConfig, version)
	if err != nil {
		impl.logger.Errorw("error in getting parent resource object by id or name", "err", err, "parentConfig", parentConfig)
		return err
	}
	dependenciesOfParent := getDependenciesInObjectDataFromJsonString(parentResourceObject.ObjectData)
	resourceIdsPresentAlready, maxIndex := getExistingDependencyIdsForResourceType(dependenciesOfParent, childResourceObject.DevtronResourceId)
	if slices.Contains(resourceIdsPresentAlready, childResourceObject.Id) {
		// dependency exists
		return nil
	}
	// generate dependency data
	childDependency := adapter.BuildDependencyData(childResourceObject.Name,
		childResourceObject.Id, childResourceObject.DevtronResourceId,
		childResourceObject.DevtronResourceSchemaId, maxIndex,
		bean.DevtronResourceDependencyTypeChild, idType)
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

func (impl *DevtronResourceServiceImpl) addParentDependencyToChildResourceObj(ctx context.Context, parentConfig *bean.ResourceParentConfig,
	objectDataGeneral string, version string) (string, error) {
	newCtx, span := otel.Tracer("DevtronResourceService").Start(ctx, "addParentDependencyToChildResourceObj")
	defer span.End()
	parentResourceObject, err := impl.getParentResourceObject(newCtx, parentConfig, version)
	if err != nil {
		impl.logger.Errorw("error in getting parent resource object by id or name", "err", err, "parentConfig", parentConfig)
		return objectDataGeneral, err
	}
	// generate dependency data
	parentObjectId, parentIdType := getResourceObjectIdAndType(parentResourceObject)
	parentDependency := adapter.BuildDependencyData(parentResourceObject.Name,
		parentObjectId, parentResourceObject.DevtronResourceId,
		parentResourceObject.DevtronResourceSchemaId, 0,
		bean.DevtronResourceDependencyTypeParent, parentIdType)

	// patch updated dependency data
	objectDataGeneral, err = sjson.Set(objectDataGeneral, bean.ResourceObjectDependenciesPath, []*bean.DevtronResourceDependencyBean{parentDependency})
	if err != nil {
		impl.logger.Errorw("error in setting parent dependencies in child resource object", "err", err, "parentDependency", parentDependency)
		return objectDataGeneral, err
	}
	return objectDataGeneral, nil
}

func getResourceObjectIdValue(reqBean *bean.DevtronResourceObjectBean) int {
	if reqBean.IdType == bean.ResourceObjectIdType {
		return reqBean.Id
	} else {
		return reqBean.OldObjectId
	}
}

func (impl *DevtronResourceServiceImpl) setDevtronManagedFieldsInObjectData(objectData string, reqBean *bean.DevtronResourceObjectBean) (string, error) {
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
	objectData, err = sjson.Set(objectData, bean.ResourceObjectIdPath, getResourceObjectIdValue(reqBean))
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

func (impl *DevtronResourceServiceImpl) performPatchOperation(objectData string, query bean.PatchQuery) (string, string, error) {
	var err error
	switch query.Path {
	case bean.DescriptionQueryPath:
		objectData, err = patchResourceObjectDataAtAPath(objectData, bean.ResourceObjectDescriptionPath, query.Value)
		return objectData, bean.ResourceObjectDescriptionPath, err
	case bean.StatusQueryPath:
		objectData, err = patchResourceObjectDataAtAPath(objectData, bean.ResourceConfigStatusPath, query.Value)
		return objectData, bean.ResourceConfigStatusPath, err
	case bean.NoteQueryPath:
		objectData, err = patchResourceObjectDataAtAPath(objectData, bean.ResourceObjectReleaseNotePath, query.Value)
		return objectData, bean.ResourceObjectReleaseNotePath, err
	case bean.TagsQueryPath:
		objectData, err = patchResourceObjectDataAtAPath(objectData, bean.ResourceObjectTagsPath, query.Value)
		return objectData, bean.ResourceObjectTagsPath, err
	case bean.LockQueryPath:
		objectData, err = patchResourceObjectDataAtAPath(objectData, bean.ResourceConfigStatusIsLockedPath, query.Value)
		return objectData, bean.ResourceConfigStatusIsLockedPath, err
	case bean.ReleaseInstructionQueryPath:
		objectData, err = patchResourceObjectDataAtAPath(objectData, bean.ResourceObjectReleaseInstructionPath, query.Value)
		return objectData, bean.ResourceObjectReleaseInstructionPath, err
	case bean.NameQueryPath:
		objectData, err = patchResourceObjectDataAtAPath(objectData, bean.ResourceObjectNamePath, query.Value)
		return objectData, bean.ResourceObjectNamePath, err
	default:
		return objectData, "", util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.PatchPathNotSupportedError, bean.PatchPathNotSupportedError)
	}
	return objectData, "", err

}

func patchResourceObjectDataAtAPath(objectData string, path string, value interface{}) (string, error) {
	return sjson.Set(objectData, path, value)
}

func (impl *DevtronResourceServiceImpl) getMetadataForADependency(resourceSchemaId, oldObjectId int, mapOfAppsMetadata, mapOfCdPipelinesMetadata map[int]interface{}) interface{} {
	var metadata interface{}
	if schema, ok := impl.devtronResourcesSchemaMapById[resourceSchemaId]; ok {
		if schema.DevtronResource.Kind == bean.DevtronResourceDevtronApplication.ToString() {
			metadata = mapOfAppsMetadata[oldObjectId]
		} else if schema.DevtronResource.Kind == bean.DevtronResourceCdPipeline.ToString() {
			metadata = mapOfCdPipelinesMetadata[oldObjectId]
		}
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return metadata
}

func (impl *DevtronResourceServiceImpl) getDownstreamDependencyObjects(argValuesToGetDownstream []interface{},
	argTypesToGetDownstream []string, schemaIdsOfArgsToGetDownstream []int) ([]*repository.DevtronResourceObject, error) {
	downstreamDependencyObjects := make([]*repository.DevtronResourceObject, 0, len(argValuesToGetDownstream))
	var err error
	if len(argValuesToGetDownstream) > 0 {
		downstreamDependencyObjects, err = impl.devtronResourceObjectRepository.GetDownstreamObjectsByParentArgAndSchemaIds(argValuesToGetDownstream,
			argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting downstream objects by parent old object ids and schema ids", "err", err, "oldObjectIds", argValuesToGetDownstream,
				"schemaIds", schemaIdsOfArgsToGetDownstream)
			return nil, err
		}
	}
	return downstreamDependencyObjects, nil
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
			existingObjectsMap[getKeyForADependencyMap(oldObject.Name, oldObject.OldObjectId, oldObject.DevtronResourceSchemaId)] = oldObject
		}
	}
	return existingObjectsMap, nil
}

func (impl *DevtronResourceServiceImpl) getAppsAndReturnNotFoundIds(appIds []int) ([]int, error) {
	invalidAppIds := make([]int, 0, len(appIds))
	mapOfApps := make(map[int]*appRepository.App)
	apps, err := impl.appRepository.FindAppAndProjectByIdsIn(appIds)
	if err != nil {
		impl.logger.Errorw("error in getting apps by ids", "err", err, "ids", appIds)
		return invalidAppIds, err
	}
	for _, app := range apps {
		mapOfApps[app.Id] = app
	}
	if len(mapOfApps) != len(appIds) {
		for _, dependencyId := range appIds {
			if _, ok := mapOfApps[dependencyId]; !ok {
				invalidAppIds = append(invalidAppIds, dependencyId)
			}
		}
	}
	return invalidAppIds, nil
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

func (impl *DevtronResourceServiceImpl) getKindSubKindOfResourceBySchemaObject(devtronResourceSchema *repository.DevtronResourceSchema) (string, string) {
	kind, subKind := "", ""
	if devtronResourceSchema != nil {
		devtronResource := devtronResourceSchema.DevtronResource
		return impl.getKindSubKindOfResource(&devtronResource)
	}
	return kind, subKind
}

func (impl *DevtronResourceServiceImpl) getKindSubKindOfResource(devtronResource *repository.DevtronResource) (string, string) {
	kind, subKind := "", ""
	if devtronResource != nil {
		if devtronResource.ParentKindId > 0 {
			devtronParentResource := impl.devtronResourcesMapById[devtronResource.ParentKindId]
			if devtronParentResource != nil {
				kind = devtronParentResource.Kind
				subKind = devtronResource.Kind
			}
		} else {
			kind = devtronResource.Kind
		}
	}
	return kind, subKind
}

func (impl *DevtronResourceServiceImpl) getMapOfAppAndCdPipelineMetadata(appIdsToGetMetadata, pipelineIdsToGetMetadata []int) (map[int]interface{}, map[int]interface{}, error) {
	mapOfAppsMetadata := make(map[int]interface{})
	mapOfCdPipelinesMetadata := make(map[int]interface{})
	var apps []*appRepository.App
	var err error
	if len(appIdsToGetMetadata) > 0 {
		apps, err = impl.appRepository.FindAppAndProjectByIdsIn(appIdsToGetMetadata)
		if err != nil {
			impl.logger.Errorw("error in getting apps by ids", "err", err, "ids", appIdsToGetMetadata)
			return nil, nil, err
		}
	}
	for _, app := range apps {
		mapOfAppsMetadata[app.Id] = &struct {
			AppName string `json:"appName"`
			AppId   int    `json:"appId"`
		}{
			AppName: app.AppName,
			AppId:   app.Id,
		}
	}
	var pipelineMetadataDtos []*apiBean.EnvironmentForDependency
	if len(pipelineIdsToGetMetadata) > 0 {
		pipelineMetadataDtos, err = impl.appListingRepository.FetchDependencyMetadataByPipelineIds(pipelineIdsToGetMetadata)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipelines by ids", "err", err, "ids", pipelineIdsToGetMetadata)
			return nil, nil, err
		}
	}
	for _, pipelineMetadata := range pipelineMetadataDtos {
		mapOfCdPipelinesMetadata[pipelineMetadata.PipelineId] = pipelineMetadata
	}
	return mapOfAppsMetadata, mapOfCdPipelinesMetadata, nil
}

func (impl *DevtronResourceServiceImpl) getUpdatedDependencyArrayWithMetadata(dependencies []*bean.DevtronResourceDependencyBean, mapOfAppsMetadata, mapOfCdPipelinesMetadata map[int]interface{}) []*bean.DevtronResourceDependencyBean {
	for _, dependency := range dependencies {
		dependency.Metadata = impl.getMetadataForADependency(dependency.DevtronResourceSchemaId, dependency.OldObjectId,
			mapOfAppsMetadata, mapOfCdPipelinesMetadata)
		for _, nestedDependency := range dependency.Dependencies {
			nestedDependency.Metadata = impl.getMetadataForADependency(nestedDependency.DevtronResourceSchemaId, nestedDependency.OldObjectId,
				mapOfAppsMetadata, mapOfCdPipelinesMetadata)
		}
	}
	return dependencies
}

func (impl *DevtronResourceServiceImpl) updateAppIdAndPipelineIdForADependency(dependency *bean.DevtronResourceDependencyBean,
	appIdsToGetMetadata, pipelineIdsToGetMetadata *[]int) {
	resourceSchemaId := dependency.DevtronResourceSchemaId
	if schema, ok := impl.devtronResourcesSchemaMapById[resourceSchemaId]; ok {
		if schema.DevtronResource.Kind == bean.DevtronResourceDevtronApplication.ToString() {
			*appIdsToGetMetadata = append(*appIdsToGetMetadata, dependency.OldObjectId)
		} else if schema.DevtronResource.Kind == bean.DevtronResourceCdPipeline.ToString() {
			*pipelineIdsToGetMetadata = append(*pipelineIdsToGetMetadata, dependency.OldObjectId)
		}
	}
}

func (impl *DevtronResourceServiceImpl) separateNonChildAndChildDependencies(dependenciesOfParent []*bean.DevtronResourceDependencyBean,
	argValuesToGetDownstream *[]interface{}, argTypesToGetDownstream *[]string, schemaIdsOfArgsToGetDownstream *[]int) ([]*bean.DevtronResourceDependencyBean,
	map[string]int, []*bean.DevtronResourceDependencyBean, map[string]int, []int, []int, int, error) {

	nonChildDependenciesOfParent := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesOfParent))
	mapOfNonChildDependenciesAndIndex := make(map[string]int, len(dependenciesOfParent)) //map of key : ["oldObjectId-schemaId" or "schemaName-schemaId"] and index of obj in array
	childDependenciesOfParent := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesOfParent))
	mapOfChildDependenciesAndIndex := make(map[string]int, len(dependenciesOfParent)) //map of key : ["oldObjectId-schemaId" or "schemaName-schemaId"] and index of obj in array

	var maxIndexInNonChildDependencies float64

	appIdsToGetMetadata := make([]int, 0, len(dependenciesOfParent))
	pipelineIdsToGetMetadata := make([]int, 0, 2*len(dependenciesOfParent))

	for _, dependencyOfParent := range dependenciesOfParent {
		dependencyOfParent.Metadata = nil //emptying metadata in case someone sends it with reference to get api response
		switch dependencyOfParent.TypeOfDependency {
		case bean.DevtronResourceDependencyTypeUpstream:
			maxIndexInNonChildDependencies = math.Max(maxIndexInNonChildDependencies, float64(dependencyOfParent.Index))
			mapOfNonChildDependenciesAndIndex[getKeyForADependencyMap(dependencyOfParent.Name, dependencyOfParent.OldObjectId, dependencyOfParent.DevtronResourceSchemaId)] = len(nonChildDependenciesOfParent)
			nonChildDependenciesOfParent = append(nonChildDependenciesOfParent, dependencyOfParent)
		case bean.DevtronResourceDependencyTypeChild:
			appendDependencyArgDetails(argValuesToGetDownstream, argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream,
				dependencyOfParent.Name, dependencyOfParent.OldObjectId, dependencyOfParent.DevtronResourceSchemaId)
			mapOfChildDependenciesAndIndex[getKeyForADependencyMap(dependencyOfParent.Name, dependencyOfParent.OldObjectId,
				dependencyOfParent.DevtronResourceSchemaId)] = len(childDependenciesOfParent)
			childDependenciesOfParent = append(childDependenciesOfParent, dependencyOfParent)
		default: //since we are not storing downstream dependencies or any other type, returning error from here for now
			return nil, nil, nil, nil, nil, nil, int(maxIndexInNonChildDependencies), fmt.Errorf("invalid dependency mapping found")
		}
		impl.updateAppIdAndPipelineIdForADependency(dependencyOfParent, &appIdsToGetMetadata, &pipelineIdsToGetMetadata)
	}
	return nonChildDependenciesOfParent, mapOfNonChildDependenciesAndIndex, childDependenciesOfParent, mapOfChildDependenciesAndIndex,
		appIdsToGetMetadata, pipelineIdsToGetMetadata, int(maxIndexInNonChildDependencies), nil
}

func (impl *DevtronResourceServiceImpl) updateNonChildDependenciesWithDownstreamDependencies(downstreamDependencyObjects []*repository.DevtronResourceObject,
	mapOfNonChildDependenciesAndIndex map[string]int, nonChildDependenciesOfParent *[]*bean.DevtronResourceDependencyBean,
	appIdsToGetMetadata, pipelineIdsToGetMetadata *[]int, maxIndexInNonChildDependencies int) ([]int, error) {
	indexesToCheckInDownstreamObjectForChildDependency := make([]int, 0, len(downstreamDependencyObjects))
	for i, downstreamObj := range downstreamDependencyObjects {
		resourceSchemaId := downstreamObj.DevtronResourceSchemaId
		if schema, ok := impl.devtronResourcesSchemaMapById[resourceSchemaId]; ok {
			if schema.DevtronResource.Kind == bean.DevtronResourceDevtronApplication.ToString() {
				mapOfNonChildDependenciesAndIndex[getKeyForADependencyMap(downstreamObj.Name,
					downstreamObj.OldObjectId, downstreamObj.DevtronResourceSchemaId)] = len(*nonChildDependenciesOfParent)
				maxIndexInNonChildDependencies++ //increasing max index by one, will use this value directly in downstream dependency index
				//this downstream obj is of devtron app meaning that this obj is downstream of app directly
				*nonChildDependenciesOfParent = append(*nonChildDependenciesOfParent, &bean.DevtronResourceDependencyBean{
					OldObjectId:             downstreamObj.OldObjectId,
					Name:                    downstreamObj.Name,
					TypeOfDependency:        bean.DevtronResourceDependencyTypeDownStream,
					DevtronResourceId:       schema.DevtronResourceId,
					DevtronResourceSchemaId: schema.Id,
					Index:                   maxIndexInNonChildDependencies,
				})
				*appIdsToGetMetadata = append(*appIdsToGetMetadata, downstreamObj.OldObjectId)
			} else if schema.DevtronResource.Kind == bean.DevtronResourceCdPipeline.ToString() {
				//here we are assuming that if the type of this downstream is not devtron app then this is cd pipeline(only possible child dependency in parent resource)
				//and these indexes are processed for downstream of child dependency in parent resource, in future this process will be the main flow, and we'll need to add handling for all type in generic manner
				indexesToCheckInDownstreamObjectForChildDependency = append(indexesToCheckInDownstreamObjectForChildDependency, i)
				*pipelineIdsToGetMetadata = append(*pipelineIdsToGetMetadata, downstreamObj.OldObjectId)
			} else {
				return nil, fmt.Errorf("invalid dependency mapping found")
			}
		}
	}
	return indexesToCheckInDownstreamObjectForChildDependency, nil
}

func (impl *DevtronResourceServiceImpl) updateChildDependenciesWithOwnDependenciesData(parentName string, parentOldObjectId,
	parentSchemaId int, mapOfChildDependenciesAndIndex map[string]int, childDependenciesOfParent []*bean.DevtronResourceDependencyBean,
	appIdsToGetMetadata, pipelineIdsToGetMetadata *[]int) error {
	parentArgValue, parentArgType := getArgTypeAndValueForADependency(parentName, parentOldObjectId)
	childObjects, err := impl.devtronResourceObjectRepository.GetChildObjectsByParentArgAndSchemaId(parentArgValue, parentArgType, parentSchemaId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error, GetChildObjectsByParentArgAndSchemaId", "err", err, "argValue", parentArgValue, "argType", parentArgType,
			"schemaId", parentSchemaId)
		return err
	}
	for _, childObject := range childObjects {
		objectData := childObject.ObjectData
		nestedDependencies := getDependenciesInObjectDataFromJsonString(objectData)
		keyForChildDependency := getKeyForADependencyMap(childObject.Name, childObject.OldObjectId, childObject.DevtronResourceSchemaId)
		indexOfChildDependency := mapOfChildDependenciesAndIndex[keyForChildDependency]
		for _, nestedDependency := range nestedDependencies {
			if nestedDependency.TypeOfDependency == bean.DevtronResourceDependencyTypeParent {
				continue
			}
			nestedDependency.Metadata = nil //emptying metadata in case someone sends it with reference to get api response
			childDependenciesOfParent[indexOfChildDependency].Dependencies =
				append(childDependenciesOfParent[indexOfChildDependency].Dependencies, nestedDependency)
			impl.updateAppIdAndPipelineIdForADependency(nestedDependency, appIdsToGetMetadata, pipelineIdsToGetMetadata)
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) updateChildDependenciesWithDownstreamDependencies(indexesToCheckInDownstreamObjectForChildDependency []int,
	downstreamDependencyObjects []*repository.DevtronResourceObject, pipelineIdsToGetMetadata *[]int,
	mapOfNonChildDependenciesAndIndex, mapOfChildDependenciesAndIndex map[string]int,
	nonChildDependenciesOfParent, childDependenciesOfParent []*bean.DevtronResourceDependencyBean) {
	for _, i := range indexesToCheckInDownstreamObjectForChildDependency {
		downstreamObj := downstreamDependencyObjects[i]
		downstreamObjDependencies := getDependenciesInObjectDataFromJsonString(downstreamObj.ObjectData)
		keyForDownstreamObjInParent := ""
		keysForDownstreamDependenciesInChild := make([]string, 0, len(downstreamObjDependencies))
		for _, downstreamDependency := range downstreamObjDependencies {
			keyForMapOfDependency := getKeyForADependencyMap(downstreamDependency.Name, downstreamDependency.OldObjectId,
				downstreamDependency.DevtronResourceSchemaId)
			if downstreamDependency.TypeOfDependency == bean.DevtronResourceDependencyTypeParent {
				keyForDownstreamObjInParent = keyForMapOfDependency
			} else {
				keysForDownstreamDependenciesInChild = append(keysForDownstreamDependenciesInChild, keyForMapOfDependency)
			}
			*pipelineIdsToGetMetadata = append(*pipelineIdsToGetMetadata, downstreamDependency.OldObjectId)
		}
		//getting parent index
		indexOfDownstreamDependencyInParent := mapOfNonChildDependenciesAndIndex[keyForDownstreamObjInParent]
		for _, keyForDownstreamChildDependencies := range keysForDownstreamDependenciesInChild {
			//getting index of child dependency where this object is to be added as downstream dependency
			if indexOfChildDependency, ok := mapOfChildDependenciesAndIndex[keyForDownstreamChildDependencies]; ok {
				downstreamDependencyInChild := &bean.DevtronResourceDependencyBean{
					OldObjectId:             downstreamObj.OldObjectId,
					Name:                    downstreamObj.Name,
					DependentOnParentIndex:  nonChildDependenciesOfParent[indexOfDownstreamDependencyInParent].Index,
					TypeOfDependency:        bean.DevtronResourceDependencyTypeDownStream,
					DevtronResourceId:       downstreamObj.DevtronResourceId,
					DevtronResourceSchemaId: downstreamObj.DevtronResourceSchemaId,
				}
				childDependenciesOfParent[indexOfChildDependency].Dependencies =
					append(childDependenciesOfParent[indexOfChildDependency].Dependencies, downstreamDependencyInChild)
			}
		}
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

func (impl *DevtronResourceServiceImpl) addChildCdPipelinesNotPresentInObjects(childDependenciesOfParent *[]*bean.DevtronResourceDependencyBean,
	mapOfChildDependenciesAndIndex map[string]int, pipelineIdsToGetMetadata *[]int, parentResourceObject *repository.DevtronResourceObject,
	argValuesToGetDownstream *[]interface{}, argTypesToGetDownstream *[]string, schemaIdsOfArgsToGetDownstream *[]int) error {
	devtronAppResource := impl.devtronResourcesMapByKind[bean.DevtronResourceDevtronApplication.ToString()]
	devtronAppResourceId := 0
	if devtronAppResource != nil {
		devtronAppResourceId = devtronAppResource.Id
	}

	if parentResourceObject != nil && parentResourceObject.DevtronResourceId == devtronAppResourceId {
		cdPipelineResource := impl.devtronResourcesMapByKind[bean.DevtronResourceCdPipeline.ToString()]
		cdPipelineResourceId := 0
		if cdPipelineResource != nil {
			cdPipelineResourceId = cdPipelineResource.Id
		}
		cdPipelineResourceSchemaId := 0
		for _, devtronResourceSchema := range impl.devtronResourcesSchemaMapById {
			if devtronResourceSchema != nil {
				if devtronResourceSchema.DevtronResourceId == cdPipelineResourceId {
					cdPipelineResourceSchemaId = devtronResourceSchema.Id
				}
			}
		}
		cdPipelineIdsPresentAlready, maxIndex := getExistingDependencyIdsForResourceType(*childDependenciesOfParent, cdPipelineResourceId)
		var pipelinesToBeAdded []*pipelineConfig.Pipeline
		var err error
		if len(cdPipelineIdsPresentAlready) > 0 {
			pipelinesToBeAdded, err = impl.pipelineRepository.FindByIdsNotInAndAppId(cdPipelineIdsPresentAlready, parentResourceObject.OldObjectId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error, FindByIdsNotInAndAppId", "err", err, "pipelineIdsPresent", cdPipelineIdsPresentAlready, "appId", parentResourceObject.OldObjectId)
				return err
			}
		} else {
			pipelinesToBeAdded, err = impl.pipelineRepository.FindActiveByAppId(parentResourceObject.OldObjectId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error, FindActiveByAppId", "err", err, "appId", parentResourceObject.OldObjectId)
				return err
			}
		}
		for _, pipelineToBeAdded := range pipelinesToBeAdded {
			childDependency := adapter.BuildDependencyData(pipelineToBeAdded.Name, pipelineToBeAdded.Id, cdPipelineResourceId, cdPipelineResourceSchemaId, maxIndex, bean.DevtronResourceDependencyTypeChild, "")
			appendDependencyArgDetails(argValuesToGetDownstream, argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream,
				childDependency.Name, childDependency.OldObjectId, childDependency.DevtronResourceSchemaId)
			mapOfChildDependenciesAndIndex[getKeyForADependencyMap(childDependency.Name, childDependency.OldObjectId,
				childDependency.DevtronResourceSchemaId)] = len(*childDependenciesOfParent)
			*childDependenciesOfParent = append(*childDependenciesOfParent, childDependency)
			*pipelineIdsToGetMetadata = append(*pipelineIdsToGetMetadata, pipelineToBeAdded.Id)
		}
	}
	return nil
}

// populateDefaultValuesToRequestBean is used to fill the default values of some fields for Create Resource request only.
func (impl *DevtronResourceServiceImpl) populateDefaultValuesToRequestBean(reqBean *bean.DevtronResourceObjectBean) error {
	f := getFuncToPopulateDefaultValuesForCreateResourceRequest(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if f != nil {
		return f(impl, reqBean)
	}
	return nil
}

func validateCreateResourceRequest(reqBean *bean.DevtronResourceObjectBean) error {
	f := getFuncToValidateCreateResourceRequest(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if f != nil {
		return f(reqBean)
	}
	return nil
}

func setIdentifierValueBasedOnKind(req *bean.DevtronResourceObjectBean) error {
	f := getFuncToGetResourceKindIdentifier(req.Kind, req.SubKind, req.Version) //getting function for component requested from UI
	if f == nil {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKind, bean.InvalidResourceKind)
	}
	req.Identifier = f(req)
	return nil

}

func (impl *DevtronResourceServiceImpl) setUserProvidedFieldsInObjectData(objectData string, reqBean *bean.DevtronResourceObjectBean) (string, error) {
	var err error
	f := getFuncToSetUserProvidedDataInResourceObject(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if f != nil {
		objectData, err = f(impl, objectData, reqBean)
	}
	return objectData, err
}

// TODO: check if we can move this
func appendDependencyArgDetails(argValues *[]interface{}, argTypes *[]string, schemaIds *[]int, name string, oldObjectId, schemaId int) {
	argValue, argType := getArgTypeAndValueForADependency(name, oldObjectId)
	*argValues = append(*argValues, argValue)
	*argTypes = append(*argTypes, argType)
	*schemaIds = append(*schemaIds, schemaId)
}

// TODO: check if we can move this
func appendDbObjectArgDetails(argValues *[]interface{}, argTypes *[]string, schemaIds *[]int, name string, oldObjectId, schemaId int) {
	argValue, argType := getArgTypeAndValueForObject(name, oldObjectId)
	*argValues = append(*argValues, argValue)
	*argTypes = append(*argTypes, argType)
	*schemaIds = append(*schemaIds, schemaId)
}

// TODO: check if we can move this
func getKeyForADependencyMap(name string, oldObjectId, devtronResourceSchemaId int) string {
	key := "" // key can be "oldObjectId-schemaId" or "name-schemaId"
	if oldObjectId > 0 {
		key = fmt.Sprintf("%d-%d", oldObjectId, devtronResourceSchemaId)
	} else if len(name) > 0 {
		//assuming here that name will not match id since it is unusual and one in a million case, if this needs to overcome then need to have different additional identifier in the key
		key = fmt.Sprintf("%s-%d", name, devtronResourceSchemaId)
	}
	return key
}

// TODO: check if we can move this
func getArgTypeAndValueForADependency(name string, oldObjectId int) (argValue interface{}, argType string) {
	if oldObjectId > 0 {
		argValue = oldObjectId
		argType = bean.IdKey //here we are sending arg as id because in the json object we are keeping this as id only and have named this as oldObjectId outside the json for easier understanding
	} else if len(name) > 0 {
		argValue = name
		argType = bean.NameKey
	}
	return argValue, argType
}

// TODO: check if we can move this
func getArgTypeAndValueForObject(name string, oldObjectId int) (argValue interface{}, argType string) {
	if oldObjectId > 0 {
		argValue = oldObjectId
		argType = bean.OldObjectIdDbColumnKey
	} else if len(name) > 0 {
		argValue = name
		argType = bean.NameDbColumnKey
	}
	return argValue, argType
}

func getSpecificDependenciesInObjectDataFromJsonString(objectData string, typeOfDependency bean.DevtronResourceDependencyType) []*bean.DevtronResourceDependencyBean {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean := getDependencyBeanFromJsonString(dependencyResult.String())
		if dependencyBean.TypeOfDependency != typeOfDependency {
			continue
		}
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies
}

func getDependenciesInObjectDataFromJsonString(objectData string) []*bean.DevtronResourceDependencyBean {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean := getDependencyBeanFromJsonString(dependencyResult.String())
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies
}

func getDependencyBeanFromJsonString(dependency string) *bean.DevtronResourceDependencyBean {
	typeResult := gjson.Get(dependency, bean.TypeOfDependencyKey)
	typeOfDependency := typeResult.String()
	devtronResourceIdResult := gjson.Get(dependency, bean.DevtronResourceIdKey)
	devtronResourceId := int(devtronResourceIdResult.Int())
	schemaIdResult := gjson.Get(dependency, bean.DevtronResourceSchemaIdKey)
	schemaId := int(schemaIdResult.Int())
	oldObjectIdResult := gjson.Get(dependency, bean.IdKey)
	oldObjectId := int(oldObjectIdResult.Int())
	idTypeResult := gjson.Get(dependency, bean.IdTypeKey)
	idType := bean.IdType(idTypeResult.String())
	nameResult := gjson.Get(dependency, bean.NameKey)
	name := nameResult.String()
	indexResult := gjson.Get(dependency, bean.IndexKey)
	index := int(indexResult.Int())
	dependentOnIndexResult := gjson.Get(dependency, bean.DependentOnIndexKey)
	dependentOnIndex := int(dependentOnIndexResult.Int())
	dependentOnParentIndexResult := gjson.Get(dependency, bean.DependentOnParentIndexKey)
	dependentOnParentIndex := int(dependentOnParentIndexResult.Int())
	//not handling for nested dependencies

	return &bean.DevtronResourceDependencyBean{
		Name:                    name,
		OldObjectId:             oldObjectId,
		DevtronResourceId:       devtronResourceId,
		DevtronResourceSchemaId: schemaId,
		DependentOnIndex:        dependentOnIndex,
		DependentOnParentIndex:  dependentOnParentIndex,
		TypeOfDependency:        bean.DevtronResourceDependencyType(typeOfDependency),
		Index:                   index,
		IdType:                  idType,
	}
}
