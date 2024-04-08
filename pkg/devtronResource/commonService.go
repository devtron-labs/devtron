package devtronResource

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"strings"
)

func (impl *DevtronResourceServiceImpl) updateCatalogDataInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectBean) (err error) {
	referencedPaths, schemaWithUpdatedRefData, err := getReferencedPathsAndUpdatedSchema(resourceSchema.Schema)
	if err != nil {
		impl.logger.Errorw("err, getReferencedPathsAndUpdatedSchema", "err", err, "schema", resourceSchema.Schema)
		return err
	}
	schema, err := impl.getUpdatedSchemaWithAllRefObjectValues(schemaWithUpdatedRefData, referencedPaths)
	if err != nil {
		impl.logger.Errorw("error, getUpdatedSchemaWithAllRefObjectValues", "err", err,
			"schemaWithUpdatedRefData", schemaWithUpdatedRefData, "referencedPaths", referencedPaths)
		return err
	}
	resourceObject.Schema = schema
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		metadataObject := gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectMetadataPath)
		resourceObject.ObjectData = metadataObject.Raw
	}
	return nil
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
