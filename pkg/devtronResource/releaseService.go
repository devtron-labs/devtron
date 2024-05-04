package devtronResource

import (
	"encoding/json"
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/internal/util"
	serviceBean "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	slices "golang.org/x/exp/slices"
	"math"
	"net/http"
	"strconv"
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
			var status bean.ReleaseStatus
			configStatus := bean.ReleaseConfigStatus(gjson.Get(existingResourceObject.ObjectData, bean.ResourceConfigStatusStatusPath).String())
			rolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(existingResourceObject.ObjectData, bean.ResourceReleaseRolloutStatusPath).String())
			switch configStatus {
			case bean.DraftReleaseConfigStatus:
				status = bean.DraftReleaseStatus
			case bean.HoldReleaseConfigStatus:
				status = bean.HoldReleaseStatus
			case bean.RescindReleaseConfigStatus:
				status = bean.RescindReleaseStatus
			case bean.CorruptedReleaseConfigStatus:
				status = bean.CorruptedReleaseStatus
			case bean.ReadyForReleaseConfigStatus:
				switch rolloutStatus {
				case bean.PartiallyDeployedReleaseRolloutStatus:
					status = bean.PartiallyReleasedReleaseStatus
				case bean.CompletelyDeployedReleaseRolloutStatus:
					status = bean.CompletelyReleasedReleaseRolloutStatus
				default:
					status = bean.ReadyForReleaseStatus
				}
			default:
				status = bean.CorruptedReleaseStatus
			}
			resourceObject.ReleaseStatus = &bean.ReleaseStatusApiBean{
				Status:   status,
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
	err = impl.updateCatalogDataForGetApiResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting catalogue data", "err", err)
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
			Status: bean.DraftReleaseConfigStatus,
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
		resourceObject := resourceObjects[i]
		resourceSchema := impl.devtronResourcesSchemaMapById[resourceObject.DevtronResourceSchemaId]
		if !isLite {
			err := impl.updateCompleteReleaseDataForGetApiResourceObj(resourceSchema, resourceObject, resourceData)
			if err != nil {
				impl.logger.Errorw("error in getting detailed resource data", "resourceObjectId", resourceObjects[i].Id, "err", err)
				return nil, err
			}
		} else {
			err := impl.updateReleaseVersionAndParentConfigInResourceObj(resourceSchema, resourceObject, resourceData)
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
		criteriaDecoder, err := helper.DecodeFilterCriteriaString(criteria)
		if err != nil {
			impl.logger.Errorw("error encountered in applyFilterCriteriaOnResourceObjects", "filterCriteria", filterCriteria, "err", bean.InvalidFilterCriteria)
			return nil, err
		}
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

func (impl *DevtronResourceServiceImpl) getReleaseDependenciesData(resourceObject *repository.DevtronResourceObject, filterDependencyTypes []bean.DevtronResourceDependencyType, dependenciesInfo []string, isLite bool) ([]*bean.DevtronResourceDependencyBean, error) {
	dependenciesOfParent, err := impl.getDependenciesInObjectDataFromJsonString(resourceObject.DevtronResourceSchemaId, resourceObject.ObjectData, isLite)
	if err != nil {
		impl.logger.Errorw("error in getting dependencies from json object", "err", err)
		return nil, err
	}
	appIdsToGetMetadata := helper.GetDependencyOldObjectIdsForSpecificType(dependenciesOfParent, bean.DevtronResourceDependencyTypeUpstream)
	dependencyFilterKeys, err := impl.getFilterKeysFromDependenciesInfo(dependenciesInfo)
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
	dependenciesWithMetaData := impl.getFilteredDependenciesWithMetaData(dependenciesOfParent, filterDependencyTypes, dependencyFilterKeys, metadataObj, appIdNameMap)
	return dependenciesWithMetaData, nil
}

func (impl *DevtronResourceServiceImpl) updateReleaseDependencyConfigDataInObj(configDataJsonObj string, configData *bean.DependencyConfigBean, isLite bool) error {
	if configData.DevtronAppDependencyConfig == nil {
		configData.DevtronAppDependencyConfig = &bean.DevtronAppDependencyConfig{}
	}
	if !isLite {
		artifactId := int(gjson.Get(configDataJsonObj, bean.DependencyConfigArtifactIdKey).Int())
		// getting artifact git commit data and image at runtime by artifact id instead of setting this schema, this has to be modified when commit source is also kept in schema (eg ci trigger is introduced)
		artifact, err := impl.ciArtifactRepository.Get(artifactId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error encountered in updateReleaseDependencyConfigDataInObj", "artifactId", artifactId, "err", err)
			return err
		}
		configData.ReleaseInstruction = gjson.Get(configDataJsonObj, bean.DependencyConfigReleaseInstructionKey).String()
		configData.CiWorkflowId = int(gjson.Get(configDataJsonObj, bean.DependencyConfigCiWorkflowKey).Int())
		gitCommitData := make([]bean.GitCommitData, 0)
		if artifactId > 0 {
			materialInfo, err := artifact.GetMaterialInfo()
			if err != nil {
				impl.logger.Errorw("error encountered in updateReleaseDependencyConfigDataInObj", "artifactId", artifact.Id, "err", err)
				return err
			}

			for _, material := range materialInfo {
				for _, modification := range material.Modifications {
					gitCommitData = append(gitCommitData, adapter.BuildGitCommit(modification.Author, modification.Branch, modification.Message, modification.ModifiedTime, modification.Revision, modification.Tag, adapter.BuildWebHookMaterialInfo(modification.WebhookData.Id, modification.WebhookData.EventActionType, modification.WebhookData.Data), material.Material.GitConfiguration.URL))
				}
			}
		}
		configData.ArtifactConfig = &bean.ArtifactConfig{
			ArtifactId:   artifactId,
			Image:        artifact.Image,
			RegistryType: gjson.Get(configDataJsonObj, bean.DependencyConfigRegistryTypeKey).String(),
			RegistryName: gjson.Get(configDataJsonObj, bean.DependencyConfigRegistryNameKey).String(),
			CommitSource: gitCommitData,
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getFilterKeysFromDependenciesInfo(dependenciesInfo []string) ([]string, error) {
	dependencyFilterKeys := make([]bean.FilterKeyObject, len(dependenciesInfo))
	for _, dependencyInfo := range dependenciesInfo {
		resourceIdentifier, err := helper.DecodeDependencyInfoString(dependencyInfo)
		if err != nil {
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
		var identifierString string
		if resourceIdentifier.Identifier != bean.AllIdentifierQueryString {
			f := getFuncToGetResourceIdAndIdTypeFromIdentifier(resourceIdentifier.ResourceKind.ToString(),
				resourceIdentifier.ResourceSubKind.ToString(), resourceIdentifier.ResourceVersion.ToString())
			if f == nil {
				err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryDependencyInfo, bean.InvalidQueryDependencyInfo)
				return nil, err
			}
			id, _, err := f(impl, resourceIdentifier)
			if err != nil {
				if util.IsErrNoRows(err) {
					err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryDependencyInfo, bean.InvalidQueryDependencyInfo)
				}
				return nil, err
			}
			identifierString = strconv.Itoa(id)
		} else {
			identifierString = bean.AllIdentifierQueryString
		}
		dependencyFilterKey := helper.GetDependencyIdentifierMap(dependencyResourceSchema.Id, identifierString)
		if !slices.Contains(dependencyFilterKeys, dependencyFilterKey) {
			dependencyFilterKeys = append(dependencyFilterKeys, dependencyFilterKey)
		}
	}
	return dependencyFilterKeys, nil
}

func (impl *DevtronResourceServiceImpl) getArtifactResponseForDependency(dependency *bean.DevtronResourceDependencyBean, appWorkflowId int,
	searchArtifactTag, searchImageTag string, limit, offset int, userId int32) (bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse], error) {
	artifactResponse := bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse]{
		Id:              dependency.OldObjectId,
		Identifier:      dependency.Identifier,
		ResourceKind:    dependency.ResourceKind,
		ResourceVersion: dependency.ResourceVersion,
	}
	var appWorkflowIds []int
	if appWorkflowId > 0 {
		appWorkflowIds = append(appWorkflowIds, appWorkflowId)
	}
	workflowComponentsMap, err := impl.appWorkflowDataReadService.FindWorkflowComponentsToAppIdMapping(dependency.OldObjectId, appWorkflowIds)
	if err != nil {
		impl.logger.Errorw("error in getting workflowComponentsMap", "err", err,
			"appId", dependency.OldObjectId, "appWorkflowId", appWorkflowId)
		return artifactResponse, err
	}
	if workflowComponentsMap == nil || len(workflowComponentsMap) == 0 {
		return artifactResponse, nil
	}
	var ciPipelineIds, externalCiPipelineIds, cdPipelineIds []int
	for _, workflowComponents := range workflowComponentsMap {
		if workflowComponents.CiPipelineId != 0 {
			ciPipelineIds = append(ciPipelineIds, workflowComponents.CiPipelineId)
		}
		if workflowComponents.ExternalCiPipelineId != 0 {
			externalCiPipelineIds = append(externalCiPipelineIds, workflowComponents.ExternalCiPipelineId)
		}
		cdPipelineIds = append(cdPipelineIds, workflowComponents.CdPipelineIds...)
	}
	request := &bean3.WorkflowComponentsBean{
		AppId:                 dependency.OldObjectId,
		CiPipelineIds:         ciPipelineIds,
		ExternalCiPipelineIds: externalCiPipelineIds,
		CdPipelineIds:         cdPipelineIds,
		SearchArtifactTag:     searchArtifactTag,
		SearchImageTag:        searchImageTag,
		Limit:                 limit,
		Offset:                offset,
		UserId:                userId,
	}
	data, err := impl.appArtifactManager.RetrieveArtifactsForAppWorkflows(request)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting artifact list for", "request", request, "err", err)
		return artifactResponse, err
	}
	// Note: Overriding bean.CiArtifactResponse.TagsEditable as we are not supporting Image Tags edit from UI in V1
	// TODO: to be removed when supported in UI
	data.TagsEditable = false
	if len(data.CiArtifacts) > 0 {
		artifactResponse.Data = &data
	}
	return artifactResponse, nil
}

func getReleaseConfigOptionsFilterCriteriaData(query *apiBean.GetConfigOptionsQueryParams) (appWorkflowId int, err error) {
	criteriaDecoder, err := helper.DecodeFilterCriteriaString(query.FilterCriteria)
	if err != nil {
		return appWorkflowId, err
	}
	if criteriaDecoder.Resource != bean.DevtronResourceAppWorkflow || criteriaDecoder.Type != bean.Id {
		return 0, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
	}
	appWorkflowId, err = strconv.Atoi(criteriaDecoder.Value)
	if err != nil {
		return appWorkflowId, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
	}
	return appWorkflowId, nil
}

func getReleaseConfigOptionsSearchKeyData(query *apiBean.GetConfigOptionsQueryParams) (searchArtifactTag, searchImageTag string, err error) {
	searchDecoder, err := helper.DecodeSearchKeyString(query.SearchKey)
	if err != nil {
		return searchArtifactTag, searchImageTag, err
	}
	if searchDecoder.SearchBy == bean.ArtifactTag {
		searchArtifactTag = searchDecoder.Value
	} else if searchDecoder.SearchBy == bean.ImageTag {
		searchImageTag = searchDecoder.Value
	} else {
		return searchArtifactTag, searchImageTag, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidSearchKey, bean.InvalidSearchKey)
	}
	return searchArtifactTag, searchImageTag, nil
}

func (impl *DevtronResourceServiceImpl) updateReleaseArtifactListInResponseObject(reqBean *bean.DevtronResourceObjectDescriptorBean, resourceObject *repository.DevtronResourceObject, query *apiBean.GetConfigOptionsQueryParams) ([]bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse], error) {
	response := make([]bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse], 0)
	filterDependencyTypes := []bean.DevtronResourceDependencyType{
		bean.DevtronResourceDependencyTypeUpstream,
	}
	dependenciesWithMetaData, err := impl.getReleaseDependenciesData(resourceObject, filterDependencyTypes, query.DependenciesInfo, true)
	if err != nil {
		return nil, err
	}
	var appWorkflowId int
	if len(query.FilterCriteria) > 0 {
		appWorkflowId, err = getReleaseConfigOptionsFilterCriteriaData(query)
		if err != nil {
			impl.logger.Errorw("error encountered in decodeFilterCriteriaForConfigOptions", "filterCriteria", query.FilterCriteria, "err", bean.InvalidFilterCriteria)
			return nil, err
		}
	}
	var searchArtifactTag, searchImageTag string
	if len(query.SearchKey) > 0 {
		searchArtifactTag, searchImageTag, err = getReleaseConfigOptionsSearchKeyData(query)
		if err != nil {
			impl.logger.Errorw("error encountered in decodeSearchKeyForConfigOptions", "searchKey", query.SearchKey, "err", bean.InvalidSearchKey)
			return nil, err
		}
	}
	for _, dependency := range dependenciesWithMetaData {
		artifactResponse, err := impl.getArtifactResponseForDependency(dependency, appWorkflowId, searchArtifactTag,
			searchImageTag, query.Limit, query.Offset, reqBean.UserId)
		if err != nil {
			return nil, err
		}
		response = append(response, artifactResponse)
	}
	return response, nil
}

func (impl *DevtronResourceServiceImpl) updateReleaseDependencyChildObjectsInObj(dependencyString string, isLite bool, id int) ([]*bean.ChildObject, error) {
	childObjects := make([]*bean.ChildObject, 0)
	if !isLite && id > 0 {
		childInheritance, err := getChildInheritanceData(dependencyString)
		if err != nil {
			return nil, err
		}
		envs, err := impl.getEnvironmentsForApplicationDependency(childInheritance, id)
		if err != nil {
			impl.logger.Errorw("error encountered in updateReleaseDependencyChildObjectsInObj", "id", id, "err", err)
			return nil, err
		}
		if len(envs) > 0 {
			childObject := adapter.BuildChildObject(envs, bean.EnvironmentChildObjectType)
			childObjects = append(childObjects, childObject)
		}
	}
	return childObjects, nil
}

func (impl *DevtronResourceServiceImpl) performReleaseResourcePatchOperation(objectData string, queries []bean.PatchQuery) (string, []string, error) {
	var err error
	auditPaths := make([]string, 0, len(queries))
	toPerformPolicyCheck := false
	newObjectData := objectData //will be using this for patches and mutation since we need old object for policy check
	for _, query := range queries {
		newObjectData, err = impl.patchQueryForReleaseObject(newObjectData, query)
		if err != nil {
			impl.logger.Errorw("error in patch operation, release track", "err", err, "objectData", "query", query)
			return "", nil, err
		}
		auditPaths = append(auditPaths, bean.PatchQueryPathAuditPathMap[query.Path])
		if query.Path == bean.StatusQueryPath || query.Path == bean.LockQueryPath {
			toPerformPolicyCheck = true
		}
	}
	if toPerformPolicyCheck {
		newObjectData, err = impl.checkReleasePatchPolicyAndAutoAction(objectData, newObjectData)
		if err != nil {
			impl.logger.Errorw("error, checkReleasePatchPolicyAndAutoAction", "err", err, "objectData", objectData, "newObjectData", newObjectData)
			return "", nil, err
		}
	}
	return newObjectData, auditPaths, nil
}

func (impl *DevtronResourceServiceImpl) checkReleasePatchPolicyAndAutoAction(oldObjectData, newObjectData string) (string, error) {
	stateFrom := adapter.GetPolicyDefinitionStateFromReleaseObject(oldObjectData)
	stateTo := adapter.GetPolicyDefinitionStateFromReleaseObject(newObjectData)

	isValid, autoAction, err := impl.releasePolicyEvaluationService.EvaluateReleaseStatusChangeAndGetAutoAction(stateTo, stateFrom)
	if err != nil {
		impl.logger.Errorw("error, EvaluateReleaseStatusChangeAndGetAutoAction", "err", err, "oldObjectData", oldObjectData, "newObjectData", newObjectData)
		return newObjectData, err
	}
	if !isValid {
		impl.logger.Errorw("error in EvaluateReleaseStatusChangeAndGetAutoAction : invalid action", "oldObjectData", oldObjectData, "newObjectData", newObjectData)
		return newObjectData, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: bean.InvalidPatchOperation,
			UserMessage:     bean.InvalidPatchOperation,
		}
	}
	if autoAction != nil {
		if autoAction.LockStatus == nil {
			autoAction.LockStatus = stateTo.LockStatus //not handling stateTo lock status nil because it is coming from adapter and always assuming it to be present
		}
		//policy valid, apply auto action if needed
		if autoAction.ConfigStatus != stateTo.ConfigStatus {
			query := bean.PatchQuery{
				Path:  bean.StatusQueryPath,
				Value: autoAction.ConfigStatus,
			}
			newObjectData, err = impl.patchQueryForReleaseObject(newObjectData, query)
			if err != nil {
				impl.logger.Errorw("error in auto action patch operation, release track", "err", err, "objectData", "query", query)
				return newObjectData, err
			}
		}
		if *autoAction.LockStatus != *stateTo.LockStatus {
			query := bean.PatchQuery{
				Path:  bean.LockQueryPath,
				Value: *autoAction.LockStatus,
			}
			newObjectData, err = impl.patchQueryForReleaseObject(newObjectData, query)
			if err != nil {
				impl.logger.Errorw("error in auto action patch operation, release track", "err", err, "objectData", "query", query)
				return newObjectData, err
			}
		}
	}
	return newObjectData, nil
}

func (impl *DevtronResourceServiceImpl) patchQueryForReleaseObject(objectData string, query bean.PatchQuery) (string, error) {
	var err error
	switch query.Path {
	case bean.DescriptionQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceObjectDescriptionPath, query.Value)
	case bean.StatusQueryPath:
		objectData, err = impl.patchConfigStatus(objectData, query.Value)
	case bean.NoteQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceObjectReleaseNotePath, query.Value)
	case bean.TagsQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceObjectTagsPath, query.Value)
	case bean.LockQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceConfigStatusIsLockedPath, query.Value)
	case bean.NameQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceObjectNamePath, query.Value)
	default:
		err = util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.PatchPathNotSupportedError, bean.PatchPathNotSupportedError)
	}
	return objectData, err
}

func (impl *DevtronResourceServiceImpl) patchConfigStatus(objectData string, value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		impl.logger.Errorw("error encountered in patchConfigStatus", "value", value, "err", err)
		return objectData, err
	}
	var configStatus bean.ConfigStatus
	err = json.Unmarshal(data, &configStatus)
	if err != nil {
		impl.logger.Errorw("error encountered in patchConfigStatus", "value ", value, "err", err)
		return objectData, err
	}
	objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceConfigStatusCommentPath, configStatus.Comment)
	if err != nil {
		return objectData, err
	}
	objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceConfigStatusStatusPath, configStatus.Status)
	return objectData, err
}

func (impl *DevtronResourceServiceImpl) validateReleaseDelete(object *repository.DevtronResourceObject) (bool, error) {
	if object == nil || object.Id == 0 {
		return false, util.GetApiErrorAdapter(http.StatusNotFound, "404", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
	}
	//getting release rollout status
	rolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(object.ObjectData, bean.ResourceReleaseRolloutStatusPath).String())
	return rolloutStatus != bean.PartiallyDeployedReleaseRolloutStatus &&
		rolloutStatus != bean.CompletelyDeployedReleaseRolloutStatus, nil
}
