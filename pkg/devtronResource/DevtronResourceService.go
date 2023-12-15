package devtronResource

import (
	"encoding/json"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
	"math"
	"net/http"
	"strings"
	"time"
)

type DevtronResourceService interface {
	GetDevtronResourceList(onlyIsExposed bool) ([]*bean.DevtronResourceBean, error)
	GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error)
	CreateOrUpdateResourceObject(reqBean *bean.DevtronResourceObjectBean) error
	GetResourceDependencies(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error)
	CreateOrUpdateResourceDependencies(req *bean.DevtronResourceObjectBean) error
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
	appRepository                        appRepository.AppRepository
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

func (impl *DevtronResourceServiceImpl) GetResourceObject(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error) {
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		return nil, err
	}
	referencedPaths, schemaWithUpdatedRefData, err := getReferencedPathsAndUpdatedSchema(resourceSchema.Schema)
	if err != nil {
		impl.logger.Errorw("err, getReferencedPathsAndUpdatedSchema", "err", err, "schema", resourceSchema.Schema)
		return nil, err
	}
	//we need to get metadata from the resource schema because it is the only part which is being used at UI.
	//In future iterations, this should be removed and made generic for the user to work on the whole object.
	responseSchema, err := impl.getUpdatedSchemaWithAllRefObjectValues(schemaWithUpdatedRefData, referencedPaths)
	if err != nil {
		impl.logger.Errorw("error, getUpdatedSchemaWithAllRefObjectValues", "err", err,
			"schemaWithUpdatedRefData", schemaWithUpdatedRefData, "referencedPaths", referencedPaths)
		return nil, err
	}
	existingResourceObject, err := impl.getExistingDevtronObject(req.OldObjectId, resourceSchema.Id, req.Name)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", req)
		return nil, err
	}
	resourceObject := &bean.DevtronResourceObjectBean{
		DevtronResourceObjectDescriptorBean: req,
		Schema:                              responseSchema,
	}
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		metadataObject := gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectMetadataPath)
		resourceObject.ObjectData = metadataObject.Raw
	}
	return resourceObject, nil
}

func (impl *DevtronResourceServiceImpl) CreateOrUpdateResourceObject(reqBean *bean.DevtronResourceObjectBean) error {
	//getting schema latest from the db (not getting it from FE for edge cases when schema has got updated
	//just before an object update is requested)
	devtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(reqBean.Kind, reqBean.SubKind, reqBean.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema from db", "err", err, "request", reqBean)
		return err
	}
	devtronResourceObject, err := impl.getExistingDevtronObject(reqBean.OldObjectId, devtronResourceSchema.Id, reqBean.Name)
	if err != nil {
		impl.logger.Errorw("error in getting object by id or name", "err", err, "request", reqBean)
		return err
	}
	return impl.createOrUpdateDevtronResourceObject(reqBean, devtronResourceSchema, devtronResourceObject, bean.ResourceObjectMetadataPath, false)
}

func (impl *DevtronResourceServiceImpl) GetResourceDependencies(req *bean.DevtronResourceObjectDescriptorBean) (*bean.DevtronResourceObjectBean, error) {
	response := &bean.DevtronResourceObjectBean{
		Dependencies:      make([]*bean.DevtronResourceDependencyBean, 0),
		ChildDependencies: make([]*bean.DevtronResourceDependencyBean, 0),
	}

	resourceSchemaOfRequestObject, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		return nil, err
	}
	existingResourceObject, err := impl.getExistingDevtronObject(req.OldObjectId, resourceSchemaOfRequestObject.Id, req.Name)
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

func (impl *DevtronResourceServiceImpl) CreateOrUpdateResourceDependencies(req *bean.DevtronResourceObjectBean) error {
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
		err = impl.createOrUpdateDevtronResourceObject(request, allRequestSchemas[i], devtronResourceSchema, bean.ResourceObjectDependenciesPath, true)
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
		return 0, &util.ApiError{
			HttpStatusCode:  404,
			Code:            "404",
			UserMessage:     "no resource objects found",
			InternalMessage: err.Error(),
		}
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

func (impl *DevtronResourceServiceImpl) createOrUpdateDevtronResourceObject(reqBean *bean.DevtronResourceObjectBean,
	devtronResourceSchema *repository.DevtronResourceSchema, devtronResourceObject *repository.DevtronResourceObject,
	objectDataPath string, skipJsonSchemaValidation bool) (err error) {
	schema := ""
	if devtronResourceSchema != nil {
		schema = devtronResourceSchema.Schema
	}
	devtronResourceObjectPresentAlready := devtronResourceObject != nil && devtronResourceObject.Id > 0
	initialObjectData := ""
	if devtronResourceObjectPresentAlready {
		initialObjectData = devtronResourceObject.ObjectData
	}
	//we need to put the object got from UI at a path(possible values currently - overview.metadata or dependencies) since only this part is controlled from UI currently
	objectDataGeneral, err := sjson.Set(initialObjectData, objectDataPath, json.RawMessage(reqBean.ObjectData))
	if err != nil {
		impl.logger.Errorw("error in setting version in schema", "err", err, "request", reqBean)
		return err
	}
	objectDataGeneral, err = impl.setDevtronManagedFieldsInObjectData(objectDataGeneral, reqBean.DevtronResourceObjectDescriptorBean)
	if err != nil {
		impl.logger.Errorw("error, setDevtronManagedFieldsInObjectData", "err", err, "req", reqBean)
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

	if devtronResourceObjectPresentAlready {
		//object already exists, update the same
		devtronResourceObject.ObjectData = objectDataGeneral
		devtronResourceObject.UpdatedBy = reqBean.UserId
		devtronResourceObject.UpdatedOn = time.Now()
		_, err = impl.devtronResourceObjectRepository.Update(devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in updating", "err", err, "req", devtronResourceObject)
			return err
		}
	} else {
		//object does not exist, create new
		devtronResourceObject = &repository.DevtronResourceObject{
			OldObjectId:             reqBean.OldObjectId,
			Name:                    reqBean.Name,
			DevtronResourceId:       devtronResourceSchema.DevtronResourceId,
			DevtronResourceSchemaId: devtronResourceSchema.Id,
			ObjectData:              objectDataGeneral,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: reqBean.UserId,
				UpdatedOn: time.Now(),
				UpdatedBy: reqBean.UserId,
			},
		}
		_, err = impl.devtronResourceObjectRepository.Save(devtronResourceObject)
		if err != nil {
			impl.logger.Errorw("error in saving", "err", err, "req", devtronResourceObject)
			return err
		}
	}
	//saving audit
	impl.saveAudit(devtronResourceObject, devtronResourceObjectPresentAlready)
	return nil
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
		return &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: internalMessage,
			UserMessage:     bean.BadRequestDependenciesErrorMessage,
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getUpdatedDependenciesRequestData(req *bean.DevtronResourceObjectBean) ([]*bean.DevtronResourceObjectBean,
	[]*repository.DevtronResourceSchema, map[string]*repository.DevtronResourceObject, error) {
	parentDevtronResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(req.Kind, req.SubKind, req.Version)
	if err != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
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

func (impl *DevtronResourceServiceImpl) saveAudit(devtronResourceObject *repository.DevtronResourceObject,
	devtronResourceObjectPresentAlready bool) {
	var auditAction repository.AuditOperationType
	if devtronResourceObjectPresentAlready {
		auditAction = repository.AuditOperationTypeUpdate
	} else {
		auditAction = repository.AuditOperationTypeCreate
	}
	//save audit
	auditModel := &repository.DevtronResourceObjectAudit{
		DevtronResourceObjectId: devtronResourceObject.Id,
		ObjectData:              devtronResourceObject.ObjectData,
		AuditOperation:          auditAction,
		AuditLog: sql.AuditLog{
			CreatedOn: devtronResourceObject.CreatedOn,
			CreatedBy: devtronResourceObject.CreatedBy,
			UpdatedBy: devtronResourceObject.UpdatedBy,
			UpdatedOn: devtronResourceObject.UpdatedOn,
		},
	}
	err := impl.devtronResourceObjectAuditRepository.Save(auditModel)
	if err != nil { //only logging not propagating to user
		impl.logger.Warnw("error in saving devtronResourceObject audit", "err", err, "auditModel", auditModel)
	}
}

// getExistingDevtronObject : this method gets existing object if present in the db.
// If not present, returns nil object along with nil error (pg.ErrNoRows error is handled in this method only)
func (impl *DevtronResourceServiceImpl) getExistingDevtronObject(oldObjectId, devtronResourceSchemaId int, name string) (*repository.DevtronResourceObject, error) {
	var existingResourceObject *repository.DevtronResourceObject
	var err error
	if oldObjectId > 0 {
		existingResourceObject, err = impl.devtronResourceObjectRepository.FindByOldObjectId(oldObjectId, devtronResourceSchemaId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting object by id or name", "err", err, "oldObjectId", oldObjectId, "devtronResourceSchemaId", devtronResourceSchemaId)
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

func (impl *DevtronResourceServiceImpl) setDevtronManagedFieldsInObjectData(objectData string, reqBean *bean.DevtronResourceObjectDescriptorBean) (string, error) {
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
	objectData, err = sjson.Set(objectData, bean.ResourceObjectIdPath, reqBean.OldObjectId)
	if err != nil {
		impl.logger.Errorw("error in setting id in schema", "err", err, "request", reqBean)
		return objectData, err
	}
	return objectData, nil
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

func (impl *DevtronResourceServiceImpl) getUpdatedSchemaWithAllRefObjectValues(schema string, referencedPaths map[string]bool) (string, error) {
	//we need to get metadata from the resource schema because it is the only part which is being used at UI.
	//In future iterations, this should be removed and made generic for the user to work on the whole object.
	responseSchemaResult := gjson.Get(schema, bean.ResourceSchemaMetadataPath)
	responseSchema := responseSchemaResult.String()
	var err error
	for refPath := range referencedPaths {
		refPathSplit := strings.Split(refPath, "/")
		if len(refPathSplit) < 3 {
			return responseSchema, fmt.Errorf("invalid schema found, references not mentioned correctly")
		}
		resourceKind := refPathSplit[2]
		if resourceKind == string(bean.DevtronResourceUser) {
			responseSchema, err = impl.getUpdatedSchemaWithUserRefDetails(resourceKind, responseSchema)
			if err != nil {
				impl.logger.Errorw("error, getUpdatedSchemaWithUserRefDetails", "err", err)
			}
		} else {
			impl.logger.Errorw("error while extracting kind of resource; kind not supported as of now", "resource kind", resourceKind)
			return responseSchema, errors.New(fmt.Sprintf("%s kind is not supported", resourceKind))
		}
	}
	return responseSchema, nil
}

func (impl *DevtronResourceServiceImpl) getUpdatedSchemaWithUserRefDetails(resourceKind, responseSchema string) (string, error) {
	userModel, err := impl.userRepository.GetAllActiveUsers()
	if err != nil {
		impl.logger.Errorw("error while fetching all users", "err", err, "resource kind", resourceKind)
		return responseSchema, err
	}
	//creating enums and enumNames
	enums := make([]map[string]interface{}, 0, len(userModel))
	enumNames := make([]interface{}, 0, len(userModel))
	for _, user := range userModel {
		enum := make(map[string]interface{})
		enum[bean.IdKey] = user.Id
		enum[bean.NameKey] = user.EmailId
		enum[bean.IconKey] = true // to get image from user profile when it is done
		enums = append(enums, enum)
		//currently we are referring only object, in future if reference is some field(got from refPathSplit) then we will pass its data as it is
		enumNames = append(enumNames, user.EmailId)
	}
	//updating schema with enum and enumNames
	referencesUpdatePathCommon := fmt.Sprintf("%s.%s", bean.ReferencesKey, resourceKind)
	referenceUpdatePathEnum := fmt.Sprintf("%s.%s", referencesUpdatePathCommon, bean.EnumKey)
	referenceUpdatePathEnumNames := fmt.Sprintf("%s.%s", referencesUpdatePathCommon, bean.EnumNamesKey)
	responseSchema, err = sjson.Set(responseSchema, referenceUpdatePathEnum, enums)
	if err != nil {
		impl.logger.Errorw("error in setting references enum in resourceSchema", "err", err)
		return responseSchema, err
	}
	responseSchema, err = sjson.Set(responseSchema, referenceUpdatePathEnumNames, enumNames)
	if err != nil {
		impl.logger.Errorw("error in setting references enumNames in resourceSchema", "err", err)
		return responseSchema, err
	}
	return responseSchema, nil
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
	var pipelineMetadataDtos []*bean2.EnvironmentForDependency
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
		cdPipelineIdsPresentAlready := make([]int, 0, len(*childDependenciesOfParent))
		var maxIndex float64
		cdPipelineResourceSchemaId := 0
		for _, devtronResourceSchema := range impl.devtronResourcesSchemaMapById {
			if devtronResourceSchema != nil {
				if devtronResourceSchema.DevtronResourceId == cdPipelineResourceId {
					cdPipelineResourceSchemaId = devtronResourceSchema.Id
				}
			}
		}
		for _, childDependency := range *childDependenciesOfParent {
			maxIndex = math.Max(maxIndex, float64(childDependency.Index))
			if childDependency.DevtronResourceId == cdPipelineResourceId {
				cdPipelineIdsPresentAlready = append(cdPipelineIdsPresentAlready, childDependency.OldObjectId)
			}
		}
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
			maxIndex++
			childDependency := &bean.DevtronResourceDependencyBean{
				Name:                    pipelineToBeAdded.Name,
				OldObjectId:             pipelineToBeAdded.Id,
				Index:                   int(maxIndex),
				TypeOfDependency:        bean.DevtronResourceDependencyTypeChild,
				DevtronResourceId:       cdPipelineResourceId,
				DevtronResourceSchemaId: cdPipelineResourceSchemaId,
				Dependencies:            make([]*bean.DevtronResourceDependencyBean, 0),
			}

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

func appendDependencyArgDetails(argValues *[]interface{}, argTypes *[]string, schemaIds *[]int, name string, oldObjectId, schemaId int) {
	argValue, argType := getArgTypeAndValueForADependency(name, oldObjectId)
	*argValues = append(*argValues, argValue)
	*argTypes = append(*argTypes, argType)
	*schemaIds = append(*schemaIds, schemaId)
}

func appendDbObjectArgDetails(argValues *[]interface{}, argTypes *[]string, schemaIds *[]int, name string, oldObjectId, schemaId int) {
	argValue, argType := getArgTypeAndValueForObject(name, oldObjectId)
	*argValues = append(*argValues, argValue)
	*argTypes = append(*argTypes, argType)
	*schemaIds = append(*schemaIds, schemaId)
}

func getReferencedPathsAndUpdatedSchema(schema string) (map[string]bool, string, error) {
	referencedPaths := make(map[string]bool)
	schemaJsonMap := make(map[string]interface{})
	schemaWithUpdatedRefData := ""
	err := json.Unmarshal([]byte(schema), &schemaJsonMap)
	if err != nil {
		return referencedPaths, schemaWithUpdatedRefData, err
	}
	getRefTypeInJsonAndAddRefKey(schemaJsonMap, referencedPaths)
	//marshaling new schema with $ref keys
	responseSchemaByte, err := json.Marshal(schemaJsonMap)
	if err != nil {
		return referencedPaths, schemaWithUpdatedRefData, err
	}
	schemaWithUpdatedRefData = string(responseSchemaByte)
	return referencedPaths, schemaWithUpdatedRefData, nil
}

func validateSchemaAndObjectData(schema, objectData string) (*gojsonschema.Result, error) {
	//validate user provided json with the schema
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(objectData)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return result, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: err.Error(),
			UserMessage:     bean.SchemaValidationFailedErrorUserMessage,
		}
	} else if !result.Valid() {
		errStr := ""
		for _, errResult := range result.Errors() {
			errStr += fmt.Sprintln(errResult.String())
		}
		return result, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: errStr,
			UserMessage:     bean.SchemaValidationFailedErrorUserMessage,
		}
	}
	return result, nil
}

func getRefTypeInJsonAndAddRefKey(schemaJsonMap map[string]interface{}, referencedObjects map[string]bool) {
	for key, value := range schemaJsonMap {
		if key == bean.RefTypeKey {
			valStr, ok := value.(string)
			if ok && strings.HasPrefix(valStr, bean.ReferencesPrefix) {
				schemaJsonMap[bean.RefKey] = valStr //adding $ref key for FE schema parsing
				delete(schemaJsonMap, bean.TypeKey) //deleting type because FE will be using $ref and thus type will be invalid
				referencedObjects[valStr] = true
			}
		} else {
			schemaUpdatedWithRef := resolveValForIteration(value, referencedObjects)
			schemaJsonMap[key] = schemaUpdatedWithRef
		}
	}
}

func resolveValForIteration(value interface{}, referencedObjects map[string]bool) interface{} {
	schemaUpdatedWithRef := value
	if valNew, ok := value.(map[string]interface{}); ok {
		getRefTypeInJsonAndAddRefKey(valNew, referencedObjects)
		schemaUpdatedWithRef = valNew
	} else if valArr, ok := value.([]interface{}); ok {
		for index, val := range valArr {
			schemaUpdatedWithRefNew := resolveValForIteration(val, referencedObjects)
			valArr[index] = schemaUpdatedWithRefNew
		}
		schemaUpdatedWithRef = valArr
	}
	return schemaUpdatedWithRef
}

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

func getDependenciesInObjectDataFromJsonString(dependency string) []*bean.DevtronResourceDependencyBean {
	dependenciesResult := gjson.Get(dependency, bean.ResourceObjectDependenciesPath)
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
	}
}
