package devtronResource

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
)

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
	dependencyFilterKey := helper.GetFilterKeyObjectFromId(dependencyOfParent.DevtronResourceSchemaId, strconv.Itoa(dependencyOfParent.OldObjectId))
	allDependencyFilterKey := helper.GetFilterKeyObjectFromId(dependencyOfParent.DevtronResourceSchemaId, bean.AllIdentifierQueryString)
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

// getDependencyBeanForGetDependenciesApi is used for get resource dependencies by extra child objects data which is not present in schema
func (impl *DevtronResourceServiceImpl) getDependencyBeanForGetDependenciesApi(parentResourceType *bean.DevtronResourceTypeReq, dependency string, isLite bool) (*bean.DevtronResourceDependencyBean, error) {
	dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependency, isLite)
	if err != nil {
		impl.logger.Errorw("error encountered in getDependencyBeanForGetDependenciesApi", "err", err, "dependency", dependency)
		return nil, err
	}
	if !isLite {
		childObjects, err := impl.getChildObjectsByParentResourceType(parentResourceType, dependency)
		if err != nil {
			impl.logger.Errorw("error encountered in getDependencyBeanForGetDependenciesApi", "err", err, "dependency", dependency)
			return nil, err
		}
		// setting child objects only for ui (get resource dependencies api), not stored in schema
		dependencyBean.ChildObjects = childObjects
	}
	return dependencyBean, nil
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

func (impl *DevtronResourceServiceImpl) getDependencyBeanWithChildInheritance(parentResourceType *bean.DevtronResourceTypeReq, dependency string, isLite bool) (*bean.DevtronResourceDependencyBean, error) {
	dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependency, isLite)
	if err != nil {
		impl.logger.Errorw("error encountered in getDependencyBeanWithChildInheritance", "dependency", dependency, "err", err)
		return nil, err
	}
	if dependencyBean.Config.ArtifactConfig != nil && dependencyBean.Config.ArtifactConfig.ArtifactId != 0 {
		dependencyBean.ChildInheritance = []*bean.ChildInheritance{{ResourceId: impl.devtronResourcesMapByKind[bean.DevtronResourceCdPipeline.ToString()].Id, Selector: adapter.GetDefaultCdPipelineSelector()}}
	}
	return dependencyBean, nil
}

// GetDependenciesBeanFromObjectData is used for by internal services for getting minimal data with filter conditions bean.DependencyFilterCondition.
func GetDependenciesBeanFromObjectData(objectData string, filterCondition *bean.DependencyFilterCondition) []*bean.DevtronResourceDependencyBean {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	dependencies := make([]*bean.DevtronResourceDependencyBean, 0, len(dependenciesResultArr))
	for _, dependencyResult := range dependenciesResultArr {
		dependency := dependencyResult.String()
		typeResult := gjson.Get(dependency, bean.TypeOfDependencyKey)
		typeOfDependency := bean.DevtronResourceDependencyType(typeResult.String())
		if !slices.Contains(filterCondition.GetFilterByTypes(), typeOfDependency) {
			continue
		}
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
		if filterCondition.GetFilterByDependentOnIndex() != 0 && !slices.Contains(dependentOnIndexArray, filterCondition.GetFilterByDependentOnIndex()) {
			continue
		}
		indexResult := gjson.Get(dependency, bean.IndexKey)
		index := int(indexResult.Int())
		if len(filterCondition.GetFilterByIndexes()) != 0 && !slices.Contains(filterCondition.GetFilterByIndexes(), index) {
			continue
		}

		schemaIdResult := gjson.Get(dependency, bean.DevtronResourceSchemaIdKey)
		schemaId := int(schemaIdResult.Int())

		oldObjectIdResult := gjson.Get(dependency, bean.IdKey)
		oldObjectId := int(oldObjectIdResult.Int())

		if len(filterCondition.GetFilterByFilterByIdAndSchemaId()) > 0 && !slices.Contains(filterCondition.GetFilterByFilterByIdAndSchemaId(), bean.IdAndSchemaIdFilter{Id: oldObjectId, DevtronResourceSchemaId: schemaId}) {
			continue
		}

		dependencyBean := bean.NewDevtronResourceDependencyBean().
			WithDependentOnIndexes(dependentOnIndexArray...).
			WithTypeOfDependency(typeOfDependency).
			WithIndex(index).
			WithDevtronResourceSchemaId(schemaId).
			WithOldObjectId(oldObjectId)

		devtronResourceIdResult := gjson.Get(dependency, bean.DevtronResourceIdKey)
		devtronResourceId := int(devtronResourceIdResult.Int())
		dependencyBean = dependencyBean.WithDevtronResourceId(devtronResourceId)

		idTypeResult := gjson.Get(dependency, bean.IdTypeKey)
		idType := bean.IdType(idTypeResult.String())
		dependencyBean = dependencyBean.WithIdType(idType)

		dependentOnIndexResult := gjson.Get(dependency, bean.DependentOnIndexKey)
		dependentOnIndex := int(dependentOnIndexResult.Int())
		dependencyBean = dependencyBean.WithDependentOnIndex(dependentOnIndex)

		dependentOnParentIndexResult := gjson.Get(dependency, bean.DependentOnParentIndexKey)
		dependentOnParentIndex := int(dependentOnParentIndexResult.Int())
		dependencyBean = dependencyBean.WithDependentOnParentIndex(dependentOnParentIndex)

		//not handling for nested dependencies

		if filterCondition.GetChildInheritance() {
			childInheritance, err := getChildInheritanceData(dependency)
			if err == nil {
				dependencyBean.WithChildInheritance(childInheritance...)
			}
		}
		dependencies = append(dependencies, dependencyBean)
	}
	return dependencies
}

func (impl *DevtronResourceServiceImpl) getChildObjectsByParentResourceType(parentResourceType *bean.DevtronResourceTypeReq, dependency string) (childObjects []*bean.ChildObject, err error) {
	f := getFuncToUpdateChildObjectsData(parentResourceType.ResourceKind.ToString(),
		parentResourceType.ResourceSubKind.ToString(), parentResourceType.ResourceVersion.ToString())
	if f != nil {
		childObjects, err = f(impl, dependency)
		if err != nil {
			return nil, err
		}
	}
	return childObjects, nil
}
