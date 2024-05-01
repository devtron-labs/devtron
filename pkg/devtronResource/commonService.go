package devtronResource

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
	"strings"
)

func (impl *DevtronResourceServiceImpl) updateCatalogDataForGetApiResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	referencedPaths, schemaWithUpdatedRefData, err := helper.GetReferencedPathsAndUpdatedSchema(resourceSchema.Schema)
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
		resourceObject.CatalogData = metadataObject.Raw
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getParentConfigVariablesFromDependencies(objectData string) (parentConfig *bean.ResourceIdentifier, parentResourceObject *repository.DevtronResourceObject, err error) {
	if gjson.Get(objectData, bean.ResourceObjectDependenciesPath).Exists() {
		var parentResourceObjectId, parentResourceSchemaId int
		var parentResourceObjectIdType bean.IdType
		gjson.Get(objectData, bean.ResourceObjectDependenciesPath).ForEach(
			func(key, value gjson.Result) bool {
				if gjson.Get(value.Raw, bean.TypeOfDependencyKey).String() == bean.DevtronResourceDependencyTypeParent.ToString() {
					parentResourceSchemaId = int(gjson.Get(value.Raw, bean.DevtronResourceSchemaIdKey).Int())
					parentResourceObjectId = int(gjson.Get(value.Raw, bean.IdKey).Int())
					parentResourceObjectIdType = bean.IdType(gjson.Get(value.Raw, bean.IdTypeKey).String())
					return false
				}
				return true
			},
		)
		if parentResourceSchemaId > 0 {
			kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(parentResourceSchemaId, impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
			if err != nil {
				impl.logger.Errorw("error in getting kind and subKind by devtronResourceSchemaId", "err", err, "devtronResourceSchemaId", parentResourceSchemaId)
				return nil, nil, err
			}
			//resolved conflicts here, in case of issue check change
			parentConfig = &bean.ResourceIdentifier{
				Id: parentResourceObjectId,
			}
			parentConfig.ResourceKind = helper.BuildExtendedResourceKindUsingKindAndSubKind(kind, subKind)
			parentConfig.ResourceVersion = bean.DevtronResourceVersion(version)
			if parentResourceObjectIdType == bean.ResourceObjectIdType {
				parentResourceObject, err = impl.devtronResourceObjectRepository.FindByIdAndSchemaId(parentResourceObjectId, parentResourceSchemaId)
				if err != nil {
					impl.logger.Errorw("error in getting release track object", "err", err,
						"idType", parentResourceObjectIdType, "id", parentResourceObjectId, "schemaId", parentResourceSchemaId)
					if util.IsErrNoRows(err) {
						err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigId, bean.InvalidResourceParentConfigId)
					}
					return nil, nil, err
				}
			} else if parentResourceObjectIdType == bean.OldObjectId {
				parentResourceObject, err = impl.devtronResourceObjectRepository.FindByOldObjectId(parentResourceObjectId, parentResourceSchemaId)
				if err != nil {
					impl.logger.Errorw("error in getting release track object", "err", err,
						"idType", parentResourceObjectIdType, "id", parentResourceObjectId, "schemaId", parentResourceSchemaId)
					if util.IsErrNoRows(err) {
						err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigId, bean.InvalidResourceParentConfigId)
					}
					return nil, nil, err
				}
			}
		}
	}
	return parentConfig, parentResourceObject, nil
}

func (impl *DevtronResourceServiceImpl) updateParentConfigInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		var parentResourceObject *repository.DevtronResourceObject
		resourceObject.OldObjectId, _ = helper.GetResourceObjectIdAndType(existingResourceObject)
		resourceObject.Name = gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectNamePath).String()
		resourceObject.ParentConfig, parentResourceObject, err = impl.getParentConfigVariablesFromDependencies(existingResourceObject.ObjectData)
		if err != nil {
			impl.logger.Errorw("error in getting parentConfig for", "err", err, "id", resourceObject.Id)
			return err
		}
		if len(parentResourceObject.Identifier) == 0 {
			err = impl.migrateDataForResourceObjectIdentifier(parentResourceObject)
			if err != nil {
				impl.logger.Warnw("error in service migrateDataForResourceObjectIdentifier", "err", err, "existingResourceObjectId", existingResourceObject.Id)
			}
		}
		resourceObject.ParentConfig.Identifier = parentResourceObject.Identifier
	}
	return nil
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

func (impl *DevtronResourceServiceImpl) getDependenciesInObjectDataFromJsonString(devtronResourceSchemaId int, objectData string, isLite bool) ([]*bean.DevtronResourceDependencyBean, error) {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(devtronResourceSchemaId,
		impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
	if err != nil {
		impl.logger.Errorw("error in getting kind and subKind by devtronResourceSchemaId", "err", err, "devtronResourceSchemaId", devtronResourceSchemaId)
		return nil, err
	}
	parentResourceType := &bean.DevtronResourceTypeReq{
		ResourceKind:    bean.DevtronResourceKind(kind),
		ResourceSubKind: bean.DevtronResourceKind(subKind),
		ResourceVersion: bean.DevtronResourceVersion(version),
	}
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean, err := impl.getDependencyBeanForGetDependenciesApi(parentResourceType, dependencyResult.String(), isLite)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies, nil
}

func (impl *DevtronResourceServiceImpl) getUpdatedDependencyArrayWithMetadata(dependencies []*bean.DevtronResourceDependencyBean, metaDataObj *bean.DependencyMetaDataBean) []*bean.DevtronResourceDependencyBean {
	for _, dependency := range dependencies {
		dependency.Metadata = impl.getMetadataForAnyDependency(dependency.DevtronResourceSchemaId,
			dependency.OldObjectId, metaDataObj)
		for _, nestedDependency := range dependency.Dependencies {
			nestedDependency.Metadata = impl.getMetadataForAnyDependency(nestedDependency.DevtronResourceSchemaId,
				nestedDependency.OldObjectId, metaDataObj)
		}
	}
	return dependencies
}

func (impl *DevtronResourceServiceImpl) getMetadataForAnyDependency(resourceSchemaId, oldObjectId int, metaDataObj *bean.DependencyMetaDataBean) interface{} {
	var metadata interface{}
	if schema, ok := impl.devtronResourcesSchemaMapById[resourceSchemaId]; ok {
		kind, subKind := helper.GetKindSubKindOfResourceBySchemaObject(schema, impl.devtronResourcesMapById)
		f := getFuncToUpdateMetadataInDependency(kind, subKind, schema.Version)
		if f != nil {
			metadata = f(oldObjectId, metaDataObj)
		}
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return metadata
}

func (impl *DevtronResourceServiceImpl) getFilteredDependenciesWithMetaData(dependenciesOfParent []*bean.DevtronResourceDependencyBean,
	filterTypes []bean.DevtronResourceDependencyType, dependencyFilterKeys []bean.FilterKeyObject, metadataObj *bean.DependencyMetaDataBean, oldObjectIdToIdentifierMap map[int]string) []*bean.DevtronResourceDependencyBean {
	filteredDependenciesOfParent := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesOfParent))
	for _, dependencyOfParent := range dependenciesOfParent {
		if slices.Contains(filterTypes, dependencyOfParent.TypeOfDependency) {
			if pass := validateDependencyFilterCondition(dependencyOfParent, dependencyFilterKeys); !pass {
				continue
			}
			filteredDependenciesOfParent = append(filteredDependenciesOfParent, dependencyOfParent)
			dependencyOfParent.Metadata = impl.getMetadataForAnyDependency(dependencyOfParent.DevtronResourceSchemaId,
				dependencyOfParent.OldObjectId, metadataObj)
			dependencyOfParent.Identifier = oldObjectIdToIdentifierMap[dependencyOfParent.OldObjectId]
		}
	}
	return filteredDependenciesOfParent
}

func validateDependencyFilterCondition(dependencyOfParent *bean.DevtronResourceDependencyBean, dependencyFilterKeys []bean.FilterKeyObject) bool {
	dependencyFilterKey := helper.GetDependencyIdentifierMap(dependencyOfParent.DevtronResourceSchemaId, strconv.Itoa(dependencyOfParent.OldObjectId))
	allDependencyFilterKey := helper.GetDependencyIdentifierMap(dependencyOfParent.DevtronResourceSchemaId, bean.AllIdentifierQueryString)
	if dependencyFilterKeys == nil || len(dependencyFilterKeys) == 0 {
		return true
	} else if slices.Contains(dependencyFilterKeys, dependencyFilterKey) {
		return true
	} else if slices.Contains(dependencyFilterKeys, allDependencyFilterKey) {
		return true
	}
	return false
}

func (impl *DevtronResourceServiceImpl) getSpecificDependenciesInObjectDataFromJsonString(devtronResourceSchemaId int, objectData string, typeOfDependency bean.DevtronResourceDependencyType) ([]*bean.DevtronResourceDependencyBean, error) {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(devtronResourceSchemaId,
		impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
	if err != nil {
		return nil, err
	}
	parentResourceType := &bean.DevtronResourceTypeReq{
		ResourceKind:    bean.DevtronResourceKind(kind),
		ResourceSubKind: bean.DevtronResourceKind(subKind),
		ResourceVersion: bean.DevtronResourceVersion(version),
	}
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependencyResult.String(), true)
		if err != nil {
			return nil, err
		}
		if dependencyBean.TypeOfDependency != typeOfDependency {
			continue
		}
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies, nil
}

func DecodeFilterCriteriaString(criteria string) (*bean.FilterCriteriaDecoder, error) {
	objs := strings.Split(criteria, "|")
	if len(objs) != 3 {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
	}
	criteriaDecoder := adapter.BuildFilterCriteriaDecoder(objs[0], objs[1], objs[2])
	return criteriaDecoder, nil
}

func DecodeSearchKeyString(searchKey string) (*bean.SearchCriteriaDecoder, error) {
	objs := strings.Split(searchKey, "|")
	if len(objs) != 2 {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidSearchKey, bean.InvalidSearchKey)
	}
	searchDecoder := adapter.BuildSearchCriteriaDecoder(objs[0], objs[1])
	return searchDecoder, nil
}
