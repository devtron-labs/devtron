package devtronResource

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
	"reflect"
	"strings"
)

func (impl *DevtronResourceServiceImpl) GetSchema(req *bean.DevtronResourceBean) (*bean.DevtronResourceBean, error) {
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindAllByResourceId(req.DevtronResourceId)
	if err != nil && resourceSchema != nil {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		return nil, err
	}
	if resourceSchema == nil {
		impl.logger.Errorw("no schema found ", "err", err, "request", req)
		return nil, &util.ApiError{
			HttpStatusCode:  404,
			Code:            "404",
			UserMessage:     fmt.Sprintf("No schema found for resourceId = %v", req.DevtronResourceId),
			InternalMessage: "schema not found",
		}
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

func (impl *DevtronResourceServiceImpl) UpdateSchema(req *bean.DevtronResourceSchemaRequestBean, dryRun bool) (*bean.UpdateSchemaResponseBean, error) {
	pathsToRemove := make([]string, 0)
	resourceSchema, err := impl.devtronResourceSchemaRepository.FindById(req.DevtronResourceSchemaId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "request", req)
		return nil, err
	}
	if resourceSchema == nil || resourceSchema.Id == 0 {
		impl.logger.Errorw("no schema found ", "err", err, "request", req)
		return nil, &util.ApiError{
			HttpStatusCode:  404,
			Code:            "404",
			UserMessage:     "schema not found",
			InternalMessage: err.Error(),
		}
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
		return nil, &util.ApiError{
			HttpStatusCode:  400,
			Code:            "400",
			UserMessage:     "invalid JSON Schema structure",
			InternalMessage: err.Error(),
		}
	}

	// json schema validation
	sl := gojsonschema.NewSchemaLoader()
	schemaLoader := gojsonschema.NewStringLoader(newSchema)
	_, err = sl.Compile(schemaLoader)
	if err != nil {
		impl.logger.Errorw("error validating new schema, schema not valid", "err", err, "fullNewSchema", fullNewSchema)
		return nil, &util.ApiError{
			HttpStatusCode:  400,
			Code:            "400",
			UserMessage:     "Provided schema is not valid",
			InternalMessage: err.Error(),
		}
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
		return nil, &util.ApiError{
			HttpStatusCode:  400,
			Code:            "400",
			UserMessage:     fmt.Sprintf("invalid refType, paths = %s", invalidRefPaths),
			InternalMessage: fmt.Sprintf("invalid refType, paths = %s", invalidRefPaths),
		}
	}

	// fetch diff b/w json schemas
	pathsToRemove, err = impl.extractDiffPaths([]byte(oldSchema), []byte(fullNewSchema))
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
	err = impl.SetDevtronResourcesAndSchemaMap()
	if err != nil {
		impl.logger.Errorw("error, SetDevtronResourcesAndSchemaMap", "err", err)
		return nil, err
	}
	return &bean.UpdateSchemaResponseBean{
		Message:       bean.SchemaUpdateSuccessMessage,
		PathsToRemove: pathsToRemove,
	}, nil
}

func (impl *DevtronResourceServiceImpl) CompareJSON(json1, json2 []byte) ([]string, error) {
	var m1 interface{}
	var m2 interface{}

	err := json.Unmarshal(json1, &m1)
	if err != nil {
		impl.logger.Errorw("error in unmarshaling json", "err", err, "json1", json1)
		return nil, err
	}
	err = json.Unmarshal(json2, &m2)
	if err != nil {
		impl.logger.Errorw("error in unmarshaling json", "err", err, "json2", json2)
		return nil, err
	}

	pathList := make([]string, 0)
	impl.compareMaps(m1.(map[string]interface{}), m2.(map[string]interface{}), "", &pathList)
	return pathList, nil
}

func (impl *DevtronResourceServiceImpl) compareMaps(m1, m2 map[string]interface{}, currentPath string, pathList *[]string) {
	for k, v1 := range m1 {
		newPath := fmt.Sprintf("%s,%s", currentPath, k)
		if v2, ok := m2[k]; ok {
			switch v1.(type) {
			case []interface{}:
				if k == bean.Enum && !reflect.DeepEqual(v1, v2) {
					*pathList = append(*pathList, currentPath)
				}
			case map[string]interface{}:
				switch v2.(type) {
				case map[string]interface{}:
					impl.compareMaps(v1.(map[string]interface{}), v2.(map[string]interface{}), newPath, pathList)
				default:
				}
			default:
				if !reflect.DeepEqual(v1, v2) {
					*pathList = append(*pathList, currentPath)
				}
			}
		} else {
			if k != bean.Required {
				*pathList = append(*pathList, newPath)
			}
		}
	}
}

func (impl *DevtronResourceServiceImpl) extractDiffPaths(json1, json2 []byte) ([]string, error) {
	pathList, err := impl.CompareJSON(json1, json2)
	if err != nil {
		impl.logger.Errorw("error in comparing json", "err", err, "json1", json1, "json2", json2)
		return nil, err
	}
	pathsToRemove := make([]string, 0)
	for _, path := range pathList {
		if len(path) > 0 {
			path = path[1:]
		}
		pathSplit := strings.Split(path, ",")
		if len(pathSplit) > 0 && (pathSplit[len(pathSplit)-1] == bean.Items || pathSplit[len(pathSplit)-1] == bean.AdditionalProperties) {
			pathSplit = pathSplit[:len(pathSplit)-1]
		}

		// remove properties attribute from path array
		idx := 0
		propertiesCount := 0
		for _, e := range pathSplit {
			if e != bean.Properties {
				times := propertiesCount / 2
				for time := 0; time < times; time++ {
					pathSplit[idx] = bean.Properties
					idx++
				}
				pathSplit[idx] = e
				propertiesCount = 0
				idx++
			} else {
				propertiesCount++
			}
		}

		pathSplit = pathSplit[:idx]

		path = strings.Join(pathSplit, ",")
		pathsToRemove = append(pathsToRemove, path)
	}
	return pathsToRemove, nil
}
