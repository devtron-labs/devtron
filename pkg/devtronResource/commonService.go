package devtronResource

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
	"net/http"
	"strings"
	"time"
)

func getResourceObjectIdAndType(existingResourceObject *repository.DevtronResourceObject) (objectId int, idType bean.IdType) {
	idType = bean.IdType(gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectIdTypePath).String())
	if idType == bean.ResourceObjectIdType {
		objectId = existingResourceObject.Id
	} else {
		objectId = existingResourceObject.OldObjectId
	}
	return objectId, idType
}

func (impl *DevtronResourceServiceImpl) getParentConfigVariablesFromDependencies(objectData string) (parentConfig *bean.ResourceParentConfig, parentResourceSchemaId int) {
	if gjson.Get(objectData, bean.ResourceObjectDependenciesPath).Exists() {
		var parentResourceObjectId int
		gjson.Get(objectData, bean.ResourceObjectDependenciesPath).ForEach(
			func(key, value gjson.Result) bool {
				if gjson.Get(value.Raw, bean.TypeOfDependencyKey).String() == bean.DevtronResourceDependencyTypeParent.ToString() {
					parentResourceSchemaId = int(gjson.Get(value.Raw, bean.DevtronResourceSchemaIdKey).Int())
					parentResourceObjectId = int(gjson.Get(value.Raw, bean.IdKey).Int())
					return false
				}
				return true
			},
		)
		if parentResourceSchemaId > 0 {
			parentResourceSchema := impl.devtronResourcesSchemaMapById[parentResourceSchemaId]
			if parentResourceSchema != nil {
				kind, subKind := impl.getKindSubKindOfResourceBySchemaObject(parentResourceSchema)
				parentConfig = &bean.ResourceParentConfig{
					DevtronResourceTypeReq: bean.DevtronResourceTypeReq{
						ResourceKind:    bean.DevtronResourceKind(kind),
						ResourceSubKind: bean.DevtronResourceKind(subKind),
						ResourceVersion: bean.DevtronResourceVersion(parentResourceSchema.Version),
					},
					Id: parentResourceObjectId,
				}
			}
		}
	}
	return parentConfig, parentResourceSchemaId
}

func (impl *DevtronResourceServiceImpl) updateParentConfigInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, _ = getResourceObjectIdAndType(existingResourceObject)
		resourceObject.Name = gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectNamePath).String()
		resourceObject.ParentConfig, _ = impl.getParentConfigVariablesFromDependencies(existingResourceObject.ObjectData)
	}
	return nil
}

func getCreatedOnTime(objectData string) (createdOn time.Time, err error) {
	createdOnStr := gjson.Get(objectData, bean.ResourceObjectCreatedOnPath).String()
	if len(createdOnStr) != 0 {
		createdOn, err = time.Parse(time.RFC3339, createdOnStr)
		if err != nil {
			return createdOn, err
		}
	}
	return createdOn, nil
}

func getOverviewTags(objectData string) map[string]string {
	tagsResults := gjson.Get(objectData, bean.ResourceObjectTagsPath).Map()
	tags := make(map[string]string, len(tagsResults))
	for key, value := range tagsResults {
		tags[key] = value.String()
	}
	return tags
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

func (impl *DevtronResourceServiceImpl) getUpdatedSchemaWithUserRefDetails(resourceKind, responseSchema string) (string, error) {
	userModel, err := impl.userRepository.GetAllExcludingApiTokenUser()
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

func (impl *DevtronResourceServiceImpl) getUserSchemaDataById(userId int32) (*bean.UserSchema, error) {
	userDetails, err := impl.userRepository.GetById(userId)
	if err != nil {
		impl.logger.Errorw("found error on getting user data ", "userId", userId)
		return nil, err
	}
	return adapter.BuildUserSchemaData(userDetails.Id, userDetails.EmailId), nil
}

func validateSchemaAndObjectData(schema, objectData string) (*gojsonschema.Result, error) {
	//validate user provided json with the schema
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(objectData)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return result, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.SchemaValidationFailedErrorUserMessage, err.Error())
	} else if !result.Valid() {
		errStr := ""
		for _, errResult := range result.Errors() {
			errStr += fmt.Sprintln(errResult.String())
		}
		return result, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.SchemaValidationFailedErrorUserMessage, errStr)
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
