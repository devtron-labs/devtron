package devtronResource

import (
	"encoding/json"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	slices "golang.org/x/exp/slices"
	"math"
	"net/http"
	"strings"
	"time"
)

func (impl *DevtronResourceServiceImpl) handleReleaseDependencyUpdateRequest(req *bean.DtResourceObjectInternalBean,
	existingObj *repository.DevtronResourceObject) {
	//getting dependencies of existingObj
	existingDependencies, err := impl.getDependenciesInObjectDataFromJsonString(existingObj.DevtronResourceSchemaId, existingObj.ObjectData, false)
	if err != nil {
		impl.logger.Errorw("error, getDependenciesInObjectDataFromJsonString", "err", err, "existingObj", existingObj)
		//TODO: handler error
		return
	}
	//we need to get parent dependency of release from existing list and add it to update req
	//and config of dependencies already present since FE does not send them in the request

	mapOfExistingDeps := make(map[string]int) //map of "id-schemaId" and index of dependency
	existingDepParentTypeIndex := 0           //index of parent type dependency in existing dependencies
	var reqDependenciesMaxIndex float64

	for i, dep := range existingDependencies {
		if dep.TypeOfDependency == bean.DevtronResourceDependencyTypeParent {
			existingDepParentTypeIndex = i
		}
		//TODO: level will fail here, fix
		mapOfExistingDeps[helper.GetKeyForADependencyMap(dep.OldObjectId, dep.DevtronResourceSchemaId)] = i
	}

	//updating config
	for i := range req.Dependencies {
		dep := req.Dependencies[i]
		reqDependenciesMaxIndex = math.Max(reqDependenciesMaxIndex, float64(dep.Index))
		if existingDepIndex, ok := mapOfExistingDeps[helper.GetKeyForADependencyMap(dep.OldObjectId, dep.DevtronResourceSchemaId)]; ok {
			//get config from existing dep and update in request dep config
			dep.Config = existingDependencies[existingDepIndex].Config
		}
	}
	//adding parent config in request dependencies
	existingParentDep := existingDependencies[existingDepParentTypeIndex]
	existingParentDep.Index = int(reqDependenciesMaxIndex + 1) //updating index of parent index
	req.Dependencies = append(req.Dependencies, existingParentDep)
	marshaledDependencies, err := json.Marshal(req.Dependencies)
	if err != nil {
		impl.logger.Errorw("error in marshaling dependencies", "err", err, "request", req)
		//TODO: handle error
		return
	}
	req.ObjectData = string(marshaledDependencies)
}

func (impl *DevtronResourceServiceImpl) updateReleaseConfigStatusForGetApiResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, _ = helper.GetResourceObjectIdAndType(existingResourceObject)
		resourceObject.Name = gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectNamePath).String()
		if gjson.Get(existingResourceObject.ObjectData, bean.ResourceConfigStatusPath).Exists() {
			resourceObject.ConfigStatus = &bean.ConfigStatus{
				Status:   bean.Status(gjson.Get(existingResourceObject.ObjectData, bean.ResourceConfigStatusStatusPath).String()),
				Comment:  gjson.Get(existingResourceObject.ObjectData, bean.ResourceConfigStatusCommentPath).String(),
				IsLocked: gjson.Get(existingResourceObject.ObjectData, bean.ResourceConfigStatusIsLockedPath).Bool(),
			}
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) updateReleaseNoteForGetApiResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, _ = helper.GetResourceObjectIdAndType(existingResourceObject)
		resourceObject.Name = gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectNamePath).String()
		if gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectReleaseNotePath).Exists() {
			resourceObject.Overview.Note = &bean.NoteBean{
				Value: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectReleaseNotePath).String(),
			}
			audit, err := impl.devtronResourceObjectAuditRepository.FindLatestAuditByOpPath(existingResourceObject.Id, bean.ResourceObjectReleaseNotePath)
			if err != nil {
				impl.logger.Errorw("error in getting audit ")
			} else if audit != nil && audit.Id >= 0 {
				userSchema, err := impl.getUserSchemaDataById(audit.UpdatedBy)
				if err != nil {
					impl.logger.Errorw("error in getting user schema, updateReleaseNoteInResourceObj", "err", err, "userId", audit.UpdatedBy)
				} else if userSchema == nil {
					userSchema = &bean.UserSchema{} //setting not null since possible that updatedBy user is deleted and could not be found now which can break UI
				}
				resourceObject.Overview.Note.UpdatedOn = audit.UpdatedOn
				resourceObject.Overview.Note.UpdatedBy = userSchema
			}
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) updateReleaseOverviewDataForGetApiResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, resourceObject.IdType = helper.GetResourceObjectIdAndType(existingResourceObject)
		resourceObject.Name = gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectNamePath).String()
		if gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectOverviewPath).Exists() {
			resourceObject.Overview = &bean.ResourceOverview{
				Description:    gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectDescriptionPath).String(),
				ReleaseVersion: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectReleaseVersionPath).String(),
				CreatedBy: &bean.UserSchema{
					Id:   int32(gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectCreatedByIdPath).Int()),
					Name: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectCreatedByNamePath).String(),
					Icon: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectCreatedByIconPath).Bool(),
				},
			}
			resourceObject.Overview.CreatedOn, err = helper.GetCreatedOnTime(existingResourceObject.ObjectData)
			if err != nil {
				impl.logger.Errorw("error in time conversion", "err", err)
				return err
			}
			resourceObject.Overview.Tags = helper.GetOverviewTags(existingResourceObject.ObjectData)
		}
	}
	// get parent config data for overview component
	err = impl.updateParentConfigInResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting note", "err", err)
		return err
	}
	return nil
}

// updateReleaseVersionAndParentConfigInResourceObj  updates only id,idType,name ,releaseVersion And parentConfig from object data
func (impl *DevtronResourceServiceImpl) updateReleaseVersionAndParentConfigInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, resourceObject.IdType = helper.GetResourceObjectIdAndType(existingResourceObject)
		resourceObject.Name = gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectNamePath).String()
		if gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectOverviewPath).Exists() {
			resourceObject.Overview = &bean.ResourceOverview{
				ReleaseVersion: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectReleaseVersionPath).String(),
			}
		}
	}
	// get parent config data for overview component
	err = impl.updateParentConfigInResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting note", "err", err)
		return err
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) updateCompleteReleaseDataForGetApiResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	err = impl.updateReleaseOverviewDataForGetApiResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting overview data", "err", err)
		return err
	}
	err = impl.updateReleaseConfigStatusForGetApiResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting config status data", "err", err)
		return err
	}
	err = impl.updateReleaseNoteForGetApiResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting note", "err", err)
		return err
	}
	return nil
}

func validateCreateReleaseRequest(reqBean *bean.DtResourceObjectCreateReqBean) error {
	if reqBean.Overview == nil || len(reqBean.Overview.ReleaseVersion) == 0 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ReleaseVersionNotFound, bean.ReleaseVersionNotFound)
	} else if reqBean.ParentConfig == nil {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceParentConfigNotFound, bean.ResourceParentConfigNotFound)
	} else if reqBean.ParentConfig.Id == 0 && len(reqBean.ParentConfig.Identifier) == 0 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigData, bean.InvalidResourceParentConfigData)
	} else if reqBean.ParentConfig.ResourceKind != bean.DevtronResourceReleaseTrack {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigKind, bean.InvalidResourceParentConfigKind)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) populateDefaultValuesForCreateReleaseRequest(reqBean *bean.DtResourceObjectCreateReqBean) error {
	if reqBean.Overview != nil && reqBean.Overview.CreatedBy == nil {
		createdByDetails, err := impl.getUserSchemaDataById(reqBean.UserId)
		// considering the user details are already verified; this error indicates to an internal db error.
		if err != nil {
			impl.logger.Errorw("error encountered in populateDefaultValuesForCreateReleaseRequest", "userId", reqBean.UserId, "err", err)
			return err
		}
		reqBean.Overview.CreatedBy = createdByDetails
		reqBean.Overview.CreatedOn = time.Now()
	}
	if len(reqBean.Name) == 0 {
		// setting default name for kind release if not provided by the user
		reqBean.Name = helper.GetDefaultReleaseNameIfNotProvided(reqBean)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) updateUserProvidedDataInReleaseObj(objectData string, reqBean *bean.DtResourceObjectInternalBean) (string, error) {
	var err error
	if reqBean.ConfigStatus == nil {
		reqBean.ConfigStatus = &bean.ConfigStatus{
			Status: bean.DraftStatus,
		}
	}
	if reqBean.ConfigStatus != nil {
		objectData, err = sjson.Set(objectData, bean.ResourceConfigStatusPath, adapter.BuildConfigStatusSchemaData(reqBean.ConfigStatus))
		if err != nil {
			impl.logger.Errorw("error in setting id in schema", "err", err, "request", reqBean)
			return objectData, err
		}
	}
	if reqBean.Overview != nil {
		objectData, err = impl.setReleaseOverviewFieldsInObjectData(objectData, reqBean.Overview)
		if err != nil {
			impl.logger.Errorw("error in setting overview data in schema", "err", err, "request", reqBean)
			return objectData, err
		}
	}
	return objectData, nil
}

func (impl *DevtronResourceServiceImpl) setReleaseOverviewFieldsInObjectData(objectData string, overview *bean.ResourceOverview) (string, error) {
	var err error
	if overview.ReleaseVersion != "" {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectReleaseVersionPath, overview.ReleaseVersion)
		if err != nil {
			impl.logger.Errorw("error in setting description in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	if overview.Description != "" {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectDescriptionPath, overview.Description)
		if err != nil {
			impl.logger.Errorw("error in setting description in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	if overview.CreatedBy != nil && overview.CreatedBy.Id > 0 {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectCreatedByPath, overview.CreatedBy)
		if err != nil {
			impl.logger.Errorw("error in setting createdBy in schema", "err", err, "overview", overview)
			return objectData, err
		}
		objectData, err = sjson.Set(objectData, bean.ResourceObjectCreatedOnPath, overview.CreatedOn)
		if err != nil {
			impl.logger.Errorw("error in setting createdOn in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	if overview.Note != nil {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectReleaseNotePath, overview.Note.Value)
		if err != nil {
			impl.logger.Errorw("error in setting description in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	if overview.Tags != nil {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectTagsPath, overview.Tags)
		if err != nil {
			impl.logger.Errorw("error in setting description in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	return objectData, nil
}

func (impl *DevtronResourceServiceImpl) buildIdentifierForReleaseResourceObj(object *repository.DevtronResourceObject) (string, error) {
	releaseVersion := gjson.Get(object.ObjectData, bean.ResourceObjectReleaseVersionPath).String()
	_, releaseTrackObject, err := impl.getParentConfigVariablesFromDependencies(object.ObjectData)
	if err != nil {
		impl.logger.Errorw("error in getting release track object", "err", err, "resourceObjectId", object.Id)
		return "", err
	}
	releaseTrackName := gjson.Get(releaseTrackObject.ObjectData, bean.ResourceObjectNamePath).String()
	return fmt.Sprintf("%s-%s", releaseTrackName, releaseVersion), nil
}

func (impl *DevtronResourceServiceImpl) listRelease(resourceObjects, childObjects []*repository.DevtronResourceObject, resourceObjectIndexChildMap map[int][]int,
	isLite bool) ([]*bean.DevtronResourceObjectGetAPIBean, error) {
	resp := make([]*bean.DevtronResourceObjectGetAPIBean, 0, len(resourceObjects))
	for i := range resourceObjects {
		resourceData := adapter.BuildDevtronResourceObjectGetAPIBean()
		if !isLite {
			err := impl.updateCompleteReleaseDataForGetApiResourceObj(nil, resourceObjects[i], resourceData)
			if err != nil {
				impl.logger.Errorw("error in getting detailed resource data", "resourceObjectId", resourceObjects[i].Id, "err", err)
				return nil, err
			}
		} else {
			err := impl.updateReleaseVersionAndParentConfigInResourceObj(nil, resourceObjects[i], resourceData)
			if err != nil {
				impl.logger.Errorw("error in getting overview data", "err", err)
				return nil, err
			}
		}
		resp = append(resp, resourceData)
	}
	return resp, nil
}

func (impl *DevtronResourceServiceImpl) getFilteredReleaseObjectsForReleaseTrackIds(resourceObjects []*repository.DevtronResourceObject, releaseTrackIds []int) ([]*repository.DevtronResourceObject, error) {
	finalResourceObjects := make([]*repository.DevtronResourceObject, 0, len(resourceObjects))
	for _, resourceObject := range resourceObjects {
		parentConfig, _, err := impl.getParentConfigVariablesFromDependencies(resourceObject.ObjectData)
		if err != nil {
			impl.logger.Errorw("error in getting parentConfig for", "err", err, "id", resourceObject.Id)
			return nil, err
		}
		if parentConfig != nil && slices.Contains(releaseTrackIds, parentConfig.Id) {
			finalResourceObjects = append(finalResourceObjects, resourceObject)
		}
	}
	return finalResourceObjects, nil
}

// applyFilterCriteriaOnResourceObjects takes in resourceObjects and returns filtered resource objects after applying all the objects
func (impl *DevtronResourceServiceImpl) applyFilterCriteriaOnReleaseResourceObjects(kind string, subKind string, version string, resourceObjects []*repository.DevtronResourceObject, filterCriteria []string) ([]*repository.DevtronResourceObject, error) {
	for _, criteria := range filterCriteria {
		// criteria will be in the form of resourceType|identifierType|commaSeperatedValues, will be invalid filterCriteria and error would be returned if not provided in this format.
		objs := strings.Split(criteria, "|")
		if len(objs) != 3 {
			impl.logger.Errorw("error encountered in applyFilterCriteriaOnResourceObjects", "filterCriteria", filterCriteria, "err", bean.InvalidFilterCriteria)
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
		}
		criteriaDecoder := adapter.BuildFilterCriteriaDecoder(objs[0], objs[1], objs[2])
		f1 := getFuncToExtractConditionsFromFilterCriteria(kind, subKind, version, criteriaDecoder.Resource)
		if f1 == nil {
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrComponent, bean.InvalidResourceKindOrComponent)
		}
		ids, err := f1(impl, criteriaDecoder)
		if err != nil {
			impl.logger.Errorw("error in applyFilterCriteriaOnResourceObjects", "criteriaDecoder", criteriaDecoder, "err", err)
			return nil, err
		}
		f2 := getFuncForProcessingFiltersOnResourceObjects(kind, subKind, version, criteriaDecoder.Resource)
		if f2 == nil {
			return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrComponent, bean.InvalidResourceKindOrComponent)
		}
		resourceObjects, err = f2(impl, resourceObjects, ids)
		if err != nil {
			impl.logger.Errorw("error in applyFilterCriteriaOnResourceObjects", "ids", ids, "err", err)
			return nil, err
		}
	}
	return resourceObjects, nil
}

func (impl *DevtronResourceServiceImpl) updateReleaseDataForGetDependenciesApi(req *bean.DevtronResourceObjectDescriptorBean, query *apiBean.GetDependencyQueryParams,
	resourceSchema *repository.DevtronResourceSchema, resourceObject *repository.DevtronResourceObject, response *bean.DevtronResourceObjectBean) (*bean.DevtronResourceObjectBean, error) {
	dependenciesOfParent, err := impl.getDependenciesInObjectDataFromJsonString(resourceObject.DevtronResourceSchemaId, resourceObject.ObjectData, query.IsLite)
	if err != nil {
		impl.logger.Errorw("error in getting dependencies from json object", "err", err)
		return nil, err
	}
	filterDependencyTypes := []bean.DevtronResourceDependencyType{
		bean.DevtronResourceDependencyTypeLevel,
		bean.DevtronResourceDependencyTypeUpstream,
	}
	appIdsToGetMetadata := helper.GetDependencyOldObjectIdsForSpecificType(dependenciesOfParent, bean.DevtronResourceDependencyTypeUpstream)
	dependencyFilterKeys, err := impl.getFilterKeysFromDependenciesInfo(query.DependenciesInfo)
	if err != nil {
		return nil, err
	}
	appMetadataObj, appIdNameMap, err := impl.getMapOfAppMetadata(appIdsToGetMetadata)
	if err != nil {
		return nil, err
	}
	metadataObj := &bean.DependencyMetaDataBean{
		MapOfAppsMetadata: appMetadataObj,
	}
	response.Dependencies = impl.getFilteredDependenciesWithMetaData(dependenciesOfParent, filterDependencyTypes, dependencyFilterKeys, metadataObj, appIdNameMap)
	return response, nil
}

func updateReleaseDependencyConfigDataInObj(configDataJsonObj string, configData *bean.DependencyConfigBean, isLite bool) error {
	if configData.DevtronAppDependencyConfig == nil {
		configData.DevtronAppDependencyConfig = &bean.DevtronAppDependencyConfig{}
	}
	configData.ArtifactConfig = &bean.ArtifactConfig{
		ArtifactId:   int(gjson.Get(configDataJsonObj, bean.DependencyConfigArtifactIdKey).Int()),
		Image:        gjson.Get(configDataJsonObj, bean.DependencyConfigImageKey).String(),
		RegistryType: gjson.Get(configDataJsonObj, bean.DependencyConfigRegistryTypeKey).String(),
		RegistryName: gjson.Get(configDataJsonObj, bean.DependencyConfigRegistryNameKey).String(),
	}
	configData.ReleaseInstruction = gjson.Get(configDataJsonObj, bean.DependencyConfigReleaseInstructionKey).String()
	//TODO: review isLite data criteria
	if !isLite {
		configData.CiWorkflowId = int(gjson.Get(configDataJsonObj, bean.DependencyConfigCiWorkflowKey).Int())
		configData.ArtifactConfig.CommitSource = make([]bean.GitCommitData, 0)
		gitCommitInfo := gjson.Get(configDataJsonObj, bean.DependencyConfigCommitSourceKey).Raw
		if len(gitCommitInfo) != 0 {
			err := json.Unmarshal([]byte(gitCommitInfo), &configData.ArtifactConfig.CommitSource)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getFilterKeysFromDependenciesInfo(dependenciesInfo []string) ([]string, error) {
	dependencyFilterKeys := make([]string, len(dependenciesInfo))
	for _, dependencyInfo := range dependenciesInfo {
		resourceIdentifier, err := helper.DecodeDependencyInfoString(dependencyInfo)
		if err != nil {
			return nil, err
		}
		f := getFuncToGetResourceIdAndIdTypeFromIdentifier(resourceIdentifier.ResourceKind.ToString(),
			resourceIdentifier.ResourceSubKind.ToString(), resourceIdentifier.ResourceVersion.ToString())
		if f == nil {
			err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryDependencyInfo, bean.InvalidQueryDependencyInfo)
			return nil, err
		}
		dependencyResourceSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(resourceIdentifier.ResourceKind.ToString(),
			resourceIdentifier.ResourceSubKind.ToString(), resourceIdentifier.ResourceVersion.ToString())
		if err != nil {
			impl.logger.Errorw("error in getting devtronResourceSchema", "err", err, "kind", resourceIdentifier.ResourceKind.ToString(),
				"subKind", resourceIdentifier.ResourceSubKind.ToString(), "version", resourceIdentifier.ResourceVersion.ToString())
			if util.IsErrNoRows(err) {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
			}
			return nil, err
		}
		id, _, err := f(impl, resourceIdentifier)
		if err != nil {
			if util.IsErrNoRows(err) {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryDependencyInfo, bean.InvalidQueryDependencyInfo)
			}
			return nil, err
		}
		dependencyFilterKeys = append(dependencyFilterKeys, helper.GetKeyForADependencyMap(id, dependencyResourceSchema.Id))
	}
	return dependencyFilterKeys, nil
}
