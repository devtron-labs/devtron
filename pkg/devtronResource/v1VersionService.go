/*
 * Copyright (c) 2024. Devtron Inc.
 */

package devtronResource

import (
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"math"
	"strings"
)

// TODO: check if we can move this
func appendDependencyArgDetails(argValues *[]interface{}, argTypes *[]string, schemaIds *[]int, oldObjectId, schemaId int) {
	argValue, argType := getArgTypeAndValueForADependency(oldObjectId)
	*argValues = append(*argValues, argValue)
	*argTypes = append(*argTypes, argType)
	*schemaIds = append(*schemaIds, schemaId)
}

func (impl *DevtronResourceServiceImpl) updateV1ResourceDataForGetDependenciesApi(req *bean.DevtronResourceObjectDescriptorBean, query *bean2.GetDependencyQueryParams,
	resourceSchema *repository.DevtronResourceSchema, resourceObject *repository.DevtronResourceObject, response *bean.DtResourceObjectDependenciesReqBean) error {
	devtronAppResourceId := impl.devtronResourcesMapByKind[bean.DevtronResourceDevtronApplication.ToString()].Id
	cdPipelineResourceId := impl.devtronResourcesMapByKind[bean.DevtronResourceCdPipeline.ToString()].Id
	filterDownstreamByResourceIds := []int{devtronAppResourceId, cdPipelineResourceId} //this should be propagated through request when UI has become mature
	dependenciesOfParent, err := impl.getDependenciesInObjectDataFromJsonString(resourceObject.DevtronResourceSchemaId, resourceObject.ObjectData, true)
	if err != nil {
		impl.logger.Errorw("error in getting dependencies from json object", "err", err)
		return err
	}
	argValuesToGetDownstream := make([]interface{}, 0, len(dependenciesOfParent)+1)
	argTypesToGetDownstream := make([]string, 0, len(dependenciesOfParent)+1)
	schemaIdsOfArgsToGetDownstream := make([]int, 0, len(dependenciesOfParent)+1)

	// adding request data for getting downstream args of request resource object
	appendDependencyArgDetails(&argValuesToGetDownstream, &argTypesToGetDownstream, &schemaIdsOfArgsToGetDownstream, req.OldObjectId, resourceSchema.Id)

	nonChildDependenciesOfParent, mapOfNonChildDependenciesAndIndex, childDependenciesOfParent, mapOfChildDependenciesAndIndex,
		appIdsToGetMetadata, pipelineIdsToGetMetadata, maxIndexInNonChildDependencies, err :=
		impl.separateNonChildAndChildDependenciesForV1Schema(dependenciesOfParent, &argValuesToGetDownstream, &argTypesToGetDownstream, &schemaIdsOfArgsToGetDownstream)

	err = impl.addChildCdPipelinesNotPresentInObjects(&childDependenciesOfParent, mapOfChildDependenciesAndIndex, &pipelineIdsToGetMetadata, resourceObject,
		&argValuesToGetDownstream, &argTypesToGetDownstream, &schemaIdsOfArgsToGetDownstream)
	if err != nil {
		impl.logger.Errorw("error, addChildCdPipelinesNotPresentInObjects", "err", err, "childDependencies", childDependenciesOfParent)
		return err
	}

	err = impl.updateChildDependenciesWithOwnDependenciesData(req.OldObjectId, resourceSchema.Id, mapOfChildDependenciesAndIndex, childDependenciesOfParent, &appIdsToGetMetadata, &pipelineIdsToGetMetadata)
	if err != nil {
		impl.logger.Errorw("error, updateChildDependenciesWithOwnDependenciesData", "err", err,
			"parentOldObjectId", req.OldObjectId, "parentSchemaId", resourceSchema.Id)
		return err
	}

	downstreamDependencyObjects, err := impl.getDownstreamDependencyObjects(argValuesToGetDownstream, argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream, filterDownstreamByResourceIds)
	if err != nil {
		impl.logger.Errorw("err, getDownstreamDependencyObjects", "err", err, "argValues", argValuesToGetDownstream,
			"argTypes", argTypesToGetDownstream, "schemaIds", schemaIdsOfArgsToGetDownstream)
		return err
	}

	indexesToCheckInDownstreamObjectForChildDependency, err :=
		impl.updateNonChildDependenciesWithDownstreamDependencies(downstreamDependencyObjects, mapOfNonChildDependenciesAndIndex, &nonChildDependenciesOfParent,
			&appIdsToGetMetadata, &pipelineIdsToGetMetadata, maxIndexInNonChildDependencies)
	if err != nil {
		impl.logger.Errorw("error, updateNonChildDependenciesWithDownstreamDependencies", "err", err,
			"downstreamDependencyObjects", downstreamDependencyObjects)
		return err
	}

	err = impl.updateChildDependenciesWithDownstreamDependencies(indexesToCheckInDownstreamObjectForChildDependency,
		downstreamDependencyObjects, &pipelineIdsToGetMetadata, mapOfNonChildDependenciesAndIndex, mapOfChildDependenciesAndIndex,
		nonChildDependenciesOfParent, childDependenciesOfParent)
	if err != nil {
		impl.logger.Errorw("error in updating child dependency data", "err", err)
		return err
	}
	mapOfAppsMetadata, mapOfCdPipelinesMetadata, err := impl.getMapOfAppAndCdPipelineMetadata(appIdsToGetMetadata, pipelineIdsToGetMetadata)
	if err != nil {
		impl.logger.Errorw("error, getMapOfAppAndCdPipelineMetadata", "err", "appIds", appIdsToGetMetadata,
			"pipelineIds", pipelineIdsToGetMetadata)
		return err
	}
	metaDataObj := &bean.DependencyMetaDataBean{
		MapOfAppsMetadata:        mapOfAppsMetadata,
		MapOfCdPipelinesMetadata: mapOfCdPipelinesMetadata,
	}
	nonChildDependenciesOfParent = impl.getUpdatedDependencyArrayWithMetadata(nonChildDependenciesOfParent, metaDataObj)
	childDependenciesOfParent = impl.getUpdatedDependencyArrayWithMetadata(childDependenciesOfParent, metaDataObj)
	response.Dependencies = nonChildDependenciesOfParent
	response.ChildDependencies = childDependenciesOfParent
	return nil
}

func (impl *DevtronResourceServiceImpl) getMapOfAppAndCdPipelineMetadata(appIdsToGetMetadata, pipelineIdsToGetMetadata []int) (mapOfAppsMetadata map[int]interface{}, mapOfCdPipelinesMetadata map[int]interface{}, err error) {
	mapOfAppsMetadata, _, err = impl.getMapOfAppMetadata(appIdsToGetMetadata)
	if err != nil {
		return nil, nil, err
	}
	mapOfCdPipelinesMetadata, err = impl.getMapOfCdPipelineMetadata(pipelineIdsToGetMetadata)
	if err != nil {
		return nil, nil, err
	}
	return mapOfAppsMetadata, mapOfCdPipelinesMetadata, nil
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

func (impl *DevtronResourceServiceImpl) separateNonChildAndChildDependenciesForV1Schema(dependenciesOfParent []*bean.DevtronResourceDependencyBean,
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
			mapOfNonChildDependenciesAndIndex[helper.GetKeyForADependencyMap(dependencyOfParent.OldObjectId, dependencyOfParent.DevtronResourceSchemaId)] = len(nonChildDependenciesOfParent)
			nonChildDependenciesOfParent = append(nonChildDependenciesOfParent, dependencyOfParent)
		case bean.DevtronResourceDependencyTypeChild:
			appendDependencyArgDetails(argValuesToGetDownstream, argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream, dependencyOfParent.OldObjectId, dependencyOfParent.DevtronResourceSchemaId)
			mapOfChildDependenciesAndIndex[helper.GetKeyForADependencyMap(dependencyOfParent.OldObjectId, dependencyOfParent.DevtronResourceSchemaId)] = len(childDependenciesOfParent)
			childDependenciesOfParent = append(childDependenciesOfParent, dependencyOfParent)
		default: //since we are not storing downstream dependencies or any other type, returning error from here for now
			return nil, nil, nil, nil, nil, nil, int(maxIndexInNonChildDependencies), fmt.Errorf("invalid dependency mapping found")
		}
		impl.updateAppIdAndPipelineIdForADependency(dependencyOfParent, &appIdsToGetMetadata, &pipelineIdsToGetMetadata)
	}
	return nonChildDependenciesOfParent, mapOfNonChildDependenciesAndIndex, childDependenciesOfParent, mapOfChildDependenciesAndIndex,
		appIdsToGetMetadata, pipelineIdsToGetMetadata, int(maxIndexInNonChildDependencies), nil
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
			maxIndex++
			childDependency := adapter.CreateDependencyData(pipelineToBeAdded.Id, cdPipelineResourceId, cdPipelineResourceSchemaId, maxIndex, bean.DevtronResourceDependencyTypeChild, "")
			appendDependencyArgDetails(argValuesToGetDownstream, argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream, childDependency.OldObjectId, childDependency.DevtronResourceSchemaId)
			mapOfChildDependenciesAndIndex[helper.GetKeyForADependencyMap(childDependency.OldObjectId, childDependency.DevtronResourceSchemaId)] = len(*childDependenciesOfParent)
			*childDependenciesOfParent = append(*childDependenciesOfParent, childDependency)
			*pipelineIdsToGetMetadata = append(*pipelineIdsToGetMetadata, pipelineToBeAdded.Id)
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) updateChildDependenciesWithOwnDependenciesData(parentOldObjectId, parentSchemaId int, mapOfChildDependenciesAndIndex map[string]int, childDependenciesOfParent []*bean.DevtronResourceDependencyBean, appIdsToGetMetadata, pipelineIdsToGetMetadata *[]int) error {
	parentArgValue, parentArgType := getArgTypeAndValueForADependency(parentOldObjectId)
	childObjects, err := impl.devtronResourceObjectRepository.GetChildObjectsByParentArgAndSchemaId(parentArgValue, parentArgType, parentSchemaId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error, GetChildObjectsByParentArgAndSchemaId", "err", err, "argValue", parentArgValue, "argType", parentArgType,
			"schemaId", parentSchemaId)
		return err
	}
	for _, childObject := range childObjects {
		objectData := childObject.ObjectData
		nestedDependencies, err := impl.getDependenciesInObjectDataFromJsonString(childObject.DevtronResourceSchemaId, objectData, true)
		if err != nil {
			impl.logger.Errorw("error in getting dependencies from json object", "err", err)
			return err
		}
		keyForChildDependency := helper.GetKeyForADependencyMap(childObject.OldObjectId, childObject.DevtronResourceSchemaId)
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

func (impl *DevtronResourceServiceImpl) getDownstreamDependencyObjects(argValuesToGetDownstream []interface{},
	argTypesToGetDownstream []string, schemaIdsOfArgsToGetDownstream, filterDownstreamByResourceIds []int) ([]*repository.DevtronResourceObject, error) {
	downstreamDependencyObjects := make([]*repository.DevtronResourceObject, 0, len(argValuesToGetDownstream))
	var err error
	if len(argValuesToGetDownstream) > 0 {
		downstreamDependencyObjects, err = impl.devtronResourceObjectRepository.GetDownstreamObjectsByParentArgAndSchemaIds(argValuesToGetDownstream,
			argTypesToGetDownstream, schemaIdsOfArgsToGetDownstream, filterDownstreamByResourceIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting downstream objects by parent old object ids and schema ids", "err", err, "oldObjectIds", argValuesToGetDownstream,
				"schemaIds", schemaIdsOfArgsToGetDownstream)
			return nil, err
		}
	}
	return downstreamDependencyObjects, nil
}

func (impl *DevtronResourceServiceImpl) updateNonChildDependenciesWithDownstreamDependencies(downstreamDependencyObjects []*repository.DevtronResourceObject,
	mapOfNonChildDependenciesAndIndex map[string]int, nonChildDependenciesOfParent *[]*bean.DevtronResourceDependencyBean,
	appIdsToGetMetadata, pipelineIdsToGetMetadata *[]int, maxIndexInNonChildDependencies int) ([]int, error) {
	indexesToCheckInDownstreamObjectForChildDependency := make([]int, 0, len(downstreamDependencyObjects))
	for i, downstreamObj := range downstreamDependencyObjects {
		resourceSchemaId := downstreamObj.DevtronResourceSchemaId
		if schema, ok := impl.devtronResourcesSchemaMapById[resourceSchemaId]; ok {
			if schema.DevtronResource.Kind == bean.DevtronResourceDevtronApplication.ToString() {
				mapOfNonChildDependenciesAndIndex[helper.GetKeyForADependencyMap(downstreamObj.OldObjectId, downstreamObj.DevtronResourceSchemaId)] = len(*nonChildDependenciesOfParent)
				maxIndexInNonChildDependencies++ //increasing max index by one, will use this value directly in downstream dependency index
				//this downstream obj is of devtron app meaning that this obj is downstream of app directly
				*nonChildDependenciesOfParent = append(*nonChildDependenciesOfParent, &bean.DevtronResourceDependencyBean{
					OldObjectId:             downstreamObj.OldObjectId,
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

func (impl *DevtronResourceServiceImpl) updateChildDependenciesWithDownstreamDependencies(indexesToCheckInDownstreamObjectForChildDependency []int,
	downstreamDependencyObjects []*repository.DevtronResourceObject, pipelineIdsToGetMetadata *[]int,
	mapOfNonChildDependenciesAndIndex, mapOfChildDependenciesAndIndex map[string]int,
	nonChildDependenciesOfParent, childDependenciesOfParent []*bean.DevtronResourceDependencyBean) error {
	for _, i := range indexesToCheckInDownstreamObjectForChildDependency {
		downstreamObj := downstreamDependencyObjects[i]
		downstreamObjDependencies, err := impl.getDependenciesInObjectDataFromJsonString(downstreamObj.DevtronResourceSchemaId, downstreamObj.ObjectData, true)
		if err != nil {
			impl.logger.Errorw("error in getting dependencies from json object", "err", err)
			return err
		}
		keyForDownstreamObjInParent := ""
		keysForDownstreamDependenciesInChild := make([]string, 0, len(downstreamObjDependencies))
		for _, downstreamDependency := range downstreamObjDependencies {
			keyForMapOfDependency := helper.GetKeyForADependencyMap(downstreamDependency.OldObjectId, downstreamDependency.DevtronResourceSchemaId)
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
	return nil
}

func getArgTypeAndValueForADependency(oldObjectId int) (argValue interface{}, argType string) {
	if oldObjectId > 0 {
		argValue = oldObjectId
		argType = bean.IdKey //here we are sending arg as id because in the json object we are keeping this as id only and have named this as oldObjectId outside the json for easier understanding
	}
	return argValue, argType
}

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
