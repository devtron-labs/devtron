/*
 * Copyright (c) 2024. Devtron Inc.
 */

package devtronResource

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
	"net/http"
)

type DevtronResourceSchemaService interface {
	GetSchema(req *bean.DevtronResourceBean) (*bean.DevtronResourceBean, error)
	UpdateSchema(req *bean.DevtronResourceSchemaRequestBean, dryRun bool) (*bean.UpdateSchemaResponseBean, error)
}

type DevtronResourceSchemaServiceImpl struct {
	logger                               *zap.SugaredLogger
	devtronResourceRepository            repository.DevtronResourceRepository
	devtronResourceSchemaRepository      repository.DevtronResourceSchemaRepository
	devtronResourceSchemaAuditRepository repository.DevtronResourceSchemaAuditRepository
	devtronResourceObjectRepository      repository.DevtronResourceObjectRepository
	devtronResourceService               DevtronResourceService
}

func NewDevtronResourceSchemaServiceImpl(logger *zap.SugaredLogger,
	devtronResourceRepository repository.DevtronResourceRepository,
	devtronResourceSchemaRepository repository.DevtronResourceSchemaRepository,
	devtronResourceSchemaAuditRepository repository.DevtronResourceSchemaAuditRepository,
	devtronResourceObjectRepository repository.DevtronResourceObjectRepository,
	devtronResourceService DevtronResourceService) *DevtronResourceSchemaServiceImpl {
	return &DevtronResourceSchemaServiceImpl{
		logger:                               logger,
		devtronResourceRepository:            devtronResourceRepository,
		devtronResourceSchemaRepository:      devtronResourceSchemaRepository,
		devtronResourceSchemaAuditRepository: devtronResourceSchemaAuditRepository,
		devtronResourceObjectRepository:      devtronResourceObjectRepository,
		devtronResourceService:               devtronResourceService,
	}
}

func (impl *DevtronResourceSchemaServiceImpl) GetSchema(req *bean.DevtronResourceBean) (*bean.DevtronResourceBean, error) {
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindAllByResourceId(req.DevtronResourceId)
	if err != nil && resourceSchema != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		return nil, err
	}
	if resourceSchema == nil {
		impl.logger.Errorw("no schema found ", "err", err, "request", req)
		return nil, util.GetApiErrorAdapter(http.StatusNotFound, "404", fmt.Sprintf("No schema found for resourceId = %v", req.DevtronResourceId), "schema not found")
	}
	resource, err := impl.devtronResourceRepository.GetById(req.DevtronResourceId)
	if err != nil {
		impl.logger.Errorw("error in getting devtron resource", "err", err, "request", req)
		return nil, err
	}
	versionSchemaDetails := make([]*bean.DevtronResourceSchemaBean, 0)
	for _, schema := range resourceSchema {
		schemaMetadataResult := gjson.Get(schema.Schema, bean.ResourceSchemaMetadataPath)
		schemaMetadata := schemaMetadataResult.String()

		sampleSchemaMetadataResult := gjson.Get(schema.SampleSchema, bean.ResourceSchemaMetadataPath)
		sampleSchemaMetadata := sampleSchemaMetadataResult.String()

		schemaVersion := &bean.DevtronResourceSchemaBean{
			DevtronResourceSchemaId: schema.Id,
			Version:                 schema.Version,
			Schema:                  schemaMetadata,
			SampleSchema:            sampleSchemaMetadata,
		}
		versionSchemaDetails = append(versionSchemaDetails, schemaVersion)
	}
	resourceSchemaDetails := &bean.DevtronResourceBean{
		DevtronResourceId:    req.DevtronResourceId,
		DisplayName:          resource.DisplayName,
		Kind:                 resource.Kind,
		Description:          resource.Description,
		VersionSchemaDetails: versionSchemaDetails,
	}
	return resourceSchemaDetails, nil
}

func (impl *DevtronResourceSchemaServiceImpl) UpdateSchema(req *bean.DevtronResourceSchemaRequestBean, dryRun bool) (*bean.UpdateSchemaResponseBean, error) {
	pathsToRemove := make([]string, 0)
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindById(req.DevtronResourceSchemaId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		return nil, err
	}
	if resourceSchema == nil || resourceSchema.Id == 0 {
		impl.logger.Errorw("no schema found ", "err", err, "request", req)
		return nil, util.GetApiErrorAdapter(http.StatusNotFound, "404", "schema not found", err.Error())
	}
	oldSchema := resourceSchema.Schema
	newSchema := req.Schema

	fullNewSchema, err := sjson.SetRaw(oldSchema, bean.ResourceSchemaMetadataPath, newSchema)
	if err != nil {
		impl.logger.Errorw("error in setting metadata to old schema", "err", err, "oldSchema", oldSchema, "newSchema", newSchema)
		return nil, err
	}

	// json structure validation
	jsonSchemaMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(fullNewSchema), &jsonSchemaMap)
	if err != nil {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", "invalid JSON Schema structure", err.Error())
	}

	// json schema validation
	sl := gojsonschema.NewSchemaLoader()
	schemaLoader := gojsonschema.NewStringLoader(newSchema)
	_, err = sl.Compile(schemaLoader)
	if err != nil {
		impl.logger.Errorw("error validating new schema, schema not valid", "err", err, "fullNewSchema", fullNewSchema)
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", "Provided schema is not valid", err.Error())
	}
	jsonSchemaMap = make(map[string]interface{})
	err = json.Unmarshal([]byte(newSchema), &jsonSchemaMap)
	if err != nil {
		return nil, err
	}
	invalidRefPaths := make([]string, 0)
	findInvalidPaths(jsonSchemaMap, "", &invalidRefPaths)
	if len(invalidRefPaths) > 0 {
		impl.logger.Errorw("error, found invalid paths in schema", "invalidPaths", invalidRefPaths)
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", fmt.Sprintf("invalid refType, paths = %s", invalidRefPaths), fmt.Sprintf("invalid refType, paths = %s", invalidRefPaths))
	}

	// fetch diff b/w json schemas
	pathsToRemove, err = helper.ExtractDiffPaths([]byte(oldSchema), []byte(fullNewSchema))
	if err != nil {
		impl.logger.Errorw("error in extracting pathList", "err", err, "oldSchema", oldSchema, "fullNewSchema", fullNewSchema)
		return nil, err
	}
	if dryRun {
		return &bean.UpdateSchemaResponseBean{
			Message:       bean.DryRunSuccessfullMessage,
			PathsToRemove: pathsToRemove,
		}, nil
	}

	// start db txn
	dbConnection := impl.devtronResourceSchemaRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in starting db transaction", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// removed keys from existing objects
	if len(pathsToRemove) > 0 {
		err = impl.devtronResourceObjectRepository.DeleteKeysFromObjectData(tx, pathsToRemove, resourceSchema.DevtronResourceId, req.UserId)
		if err != nil {
			impl.logger.Errorw("error in deleting keys from json in object data", "err", err, "pathList", pathsToRemove, "resourceId", resourceSchema.DevtronResourceId, "userId", req.UserId)
			return nil, err
		}
	}

	resourceSchemaModel := &repository.DevtronResourceSchema{
		Id:     req.DevtronResourceSchemaId,
		Schema: fullNewSchema,
	}

	// update schema in db
	err = impl.devtronResourceSchemaRepository.UpdateSchema(tx, resourceSchemaModel, req.UserId)
	if err != nil {
		impl.logger.Errorw("error in updating schema in db", "err", err, "resourceSchemaModel", resourceSchemaModel, "userId", req.UserId)
		return nil, err
	}

	devtronResource := &repository.DevtronResource{
		Id:          resourceSchema.DevtronResourceId,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}
	err = impl.devtronResourceRepository.UpdateNameAndDescription(tx, devtronResource, req.UserId)
	if err != nil {
		impl.logger.Errorw("error in updating schema in db", "err", err, "resourceSchemaModel", resourceSchemaModel, "userId", req.UserId)
		return nil, err
	}

	// commit db txn
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing schema change", "err", err, "request", req)
		return nil, err
	}

	//save audit
	auditModel := &repository.DevtronResourceSchemaAudit{
		DevtronResourceSchemaId: resourceSchemaModel.Id,
		Schema:                  resourceSchemaModel.Schema,
		AuditOperation:          repository.AuditOperationTypeUpdate,
		AuditLog: sql.AuditLog{
			CreatedOn: resourceSchemaModel.CreatedOn,
			CreatedBy: resourceSchemaModel.CreatedBy,
			UpdatedBy: resourceSchemaModel.UpdatedBy,
			UpdatedOn: resourceSchemaModel.UpdatedOn,
		},
	}
	err = impl.devtronResourceSchemaAuditRepository.Save(auditModel)
	if err != nil { //only logging not propagating to user
		impl.logger.Warnw("error in saving devtronResourceSchema audit", "err", err, "auditModel", auditModel)
	}

	//reloading cache
	err = impl.devtronResourceService.SetDevtronResourcesAndSchemaMap()
	if err != nil {
		impl.logger.Errorw("error, SetDevtronResourcesAndSchemaMap", "err", err)
		return nil, err
	}
	return &bean.UpdateSchemaResponseBean{
		Message:       bean.SchemaUpdateSuccessMessage,
		PathsToRemove: pathsToRemove,
	}, nil
}

func findInvalidPaths(schemaJsonMap map[string]interface{}, currentPath string, invalidRefPaths *[]string) {
	for key, value := range schemaJsonMap {
		newPath := fmt.Sprintf("%s.%s", currentPath, key)
		if key == bean.RefTypeKey {
			valStr, ok := value.(string)
			if (ok && valStr != bean.RefTypePath) || !ok {
				*invalidRefPaths = append(*invalidRefPaths, newPath)
			}
		} else if key == bean.ResourceObjectDependenciesPath {
			*invalidRefPaths = append(*invalidRefPaths, newPath)
		} else {
			if valNew, ok := value.(map[string]interface{}); ok {
				findInvalidPaths(valNew, newPath, invalidRefPaths)
			}
		}
	}
}
