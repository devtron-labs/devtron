package devtronResource

import (
	"context"
	"encoding/json"
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	bean5 "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	helper2 "github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	serviceBean "github.com/devtron-labs/devtron/pkg/bean"
	adapter2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	util3 "github.com/devtron-labs/devtron/pkg/devtronResource/util"
	pipelineStageBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	bean4 "github.com/devtron-labs/devtron/pkg/policyGovernance/devtronResource/release/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	adapter3 "github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	util2 "github.com/devtron-labs/devtron/pkg/workflow/cd/util"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/exp/slices"
	"math"
	"net/http"
	"sort"
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
		if dep.TypeOfDependency != bean.DevtronResourceDependencyTypeLevel { //not including level since we will be relying on new levels in request
			mapOfExistingDeps[helper.GetKeyForADependencyMap(dep.OldObjectId, dep.DevtronResourceSchemaId)] = i
		}
	}

	//updating config
	for i := range req.Dependencies {
		dep := req.Dependencies[i]
		reqDependenciesMaxIndex = math.Max(reqDependenciesMaxIndex, float64(dep.Index))
		if existingDepIndex, ok := mapOfExistingDeps[helper.GetKeyForADependencyMap(dep.OldObjectId, dep.DevtronResourceSchemaId)]; ok {
			//get config from existing dep and update in request dep config
			dep.Config = existingDependencies[existingDepIndex].Config
			if dep.Config.ArtifactConfig != nil && dep.Config.ArtifactConfig.ArtifactId > 0 {
				dep.ChildInheritance = []*bean.ChildInheritance{{ResourceId: impl.devtronResourcesMapByKind[bean.DevtronResourceCdPipeline.ToString()].Id, Selector: adapter.GetDefaultCdPipelineSelector()}}
			}
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
		if gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceConfigStatusPath).Exists() {
			var status bean.ReleaseStatus
			configStatus := bean.ReleaseConfigStatus(gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceConfigStatusStatusPath).String())
			rolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceRolloutStatusPath).String())
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
				Comment:  gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceConfigStatusCommentPath).String(),
				IsLocked: gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceConfigStatusIsLockedPath).Bool(),
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
		if gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceObjectReleaseNotePath).Exists() {
			resourceObject.Overview.Note = &bean.NoteBean{
				Value: gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceObjectReleaseNotePath).String(),
			}
			var audit *repository.DevtronResourceObjectAudit
			var err error
			audit, err = impl.devtronResourceObjectAuditRepository.FindLatestAuditByOpPath(existingResourceObject.Id, bean.ReleaseResourceObjectReleaseNotePath)
			if err != nil || audit == nil || audit.Id == 0 {
				impl.logger.Warnw("error in getting audit by path", "err", err, "objectId", existingResourceObject.Id, "path", bean.ReleaseResourceObjectReleaseNotePath)
				//it might be possible that if audit is not found then these field's data is populated through clone action, getting its audit
				audit, err = impl.devtronResourceObjectAuditRepository.FindLatestAuditByOpType(existingResourceObject.Id, repository.AuditOperationTypeClone)
				if err != nil {
					impl.logger.Errorw("error in getting audit by type", "err", err, "objectId", existingResourceObject.Id, "opType", repository.AuditOperationTypeClone)
				}
			}
			if audit != nil && audit.Id >= 0 {
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
				ReleaseVersion: gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceObjectReleaseVersionPath).String(),
				CreatedBy: &bean.UserSchema{
					Id:   int32(gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectCreatedByIdPath).Int()),
					Name: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectCreatedByNamePath).String(),
					Icon: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectCreatedByIconPath).Bool(),
				},
				FirstReleasedOn: gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceObjectFirstReleasedOnPath).Time(),
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
				ReleaseVersion: gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceObjectReleaseVersionPath).String(),
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

func (impl *DevtronResourceServiceImpl) updateOverviewAndConfigStatusDataForGetApiResourceObj(resourceSchema *repository.DevtronResourceSchema,
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
	return nil
}

func validateCreateReleaseRequest(reqBean *bean.DtResourceObjectCreateReqBean) error {
	if reqBean.Overview != nil {
		err := helper.CheckIfReleaseVersionIsValid(reqBean.Overview.ReleaseVersion)
		if err != nil {
			return err
		}
	} else if reqBean.ParentConfig == nil {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceParentConfigNotFound, bean.ResourceParentConfigNotFound)
	} else if reqBean.ParentConfig.Id == 0 && len(reqBean.ParentConfig.Identifier) == 0 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigData, bean.InvalidResourceParentConfigData)
	} else if reqBean.ParentConfig.ResourceKind != bean.DevtronResourceReleaseTrack {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigKind, bean.InvalidResourceParentConfigKind)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) populateDefaultValuesForCreateReleaseRequest(reqBean *bean.DtResourceObjectCreateReqBean) (err error) {
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
	if len(reqBean.Identifier) == 0 {
		reqBean.Identifier, err = impl.getIdentifierForCreateReleaseRequest(reqBean.DevtronResourceObjectDescriptorBean, reqBean.ParentConfig, reqBean.Overview.ReleaseVersion)
		if err != nil {
			impl.logger.Errorw("error encountered in populateDefaultValuesForCreateReleaseRequest", "err", err)
			return err
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getIdentifierForCreateReleaseRequest(descriptorBean *bean.DevtronResourceObjectDescriptorBean,
	parentConfig *bean.ResourceIdentifier, releaseVersion string) (identifier string, err error) {
	return impl.getIdentifierForReleaseByParentDescriptorBean(releaseVersion, parentConfig)
}

func (impl *DevtronResourceServiceImpl) getIdentifierForReleaseByParentDescriptorBean(releaseVersion string, parentConfig *bean.ResourceIdentifier) (string, error) {
	// identifier for release is {releaseTrackName-version}
	if len(parentConfig.Identifier) != 0 {
		// identifier for parent is same as name for release-track,
		return fmt.Sprintf("%s-%s", parentConfig.Identifier, releaseVersion), nil
	} else if (parentConfig.Id) != 0 {
		_, existingObject, err := impl.getResourceSchemaAndExistingObject(&bean.DevtronResourceObjectDescriptorBean{
			SchemaId: parentConfig.SchemaId,
			Kind:     parentConfig.ResourceKind.ToString(),
			SubKind:  parentConfig.ResourceSubKind.ToString(),
			Version:  parentConfig.ResourceVersion.ToString(),
			Id:       parentConfig.Id})
		if err != nil {
			impl.logger.Errorw("error encountered in getIdentifierForCreateReleaseRequest", "id", parentConfig.Id, "err", err)
			return "", err
		}
		return fmt.Sprintf("%s-%s", existingObject.Identifier, releaseVersion), nil
	}
	return "", util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidParentConfigIdOrIdentifier, bean.InvalidParentConfigIdOrIdentifier)
}
func (impl *DevtronResourceServiceImpl) updateUserProvidedDataInReleaseObj(objectData string, reqBean *bean.DtResourceObjectInternalBean) (string, error) {
	var err error
	isConfigStatusPresentInExistingObj := len(gjson.Get(objectData, bean.ReleaseResourceConfigStatusStatusPath).String()) > 0
	if reqBean.ConfigStatus == nil && !isConfigStatusPresentInExistingObj {
		reqBean.ConfigStatus = &bean.ConfigStatus{
			Status: bean.DraftReleaseConfigStatus,
		}
		objectData, err = sjson.Set(objectData, bean.ReleaseResourceConfigStatusPath, adapter.BuildConfigStatusSchemaData(reqBean.ConfigStatus))
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
		objectData, err = sjson.Set(objectData, bean.ReleaseResourceObjectReleaseVersionPath, overview.ReleaseVersion)
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
		objectData, err = sjson.Set(objectData, bean.ReleaseResourceObjectReleaseNotePath, overview.Note.Value)
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
	releaseVersion := gjson.Get(object.ObjectData, bean.ReleaseResourceObjectReleaseVersionPath).String()
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
	//sorting release objects on basis of created time, need to be maintained from query after sort options introduction
	sort.Slice(resourceObjects, func(i, j int) bool {
		return resourceObjects[i].CreatedOn.After(resourceObjects[j].CreatedOn)
	})
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
			err := impl.updateOverviewAndConfigStatusDataForGetApiResourceObj(resourceSchema, resourceObject, resourceData)
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
		//TODO: do bulk operation
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
		criteriaDecoder, err := util3.DecodeFilterCriteriaString(criteria)
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
	resourceSchema *repository.DevtronResourceSchema, resourceObject *repository.DevtronResourceObject, response *bean.DtResourceObjectDependenciesReqBean) error {
	dependenciesOfParent, err := impl.getDependenciesInObjectDataFromJsonString(resourceObject.DevtronResourceSchemaId, resourceObject.ObjectData, query.IsLite)
	if err != nil {
		impl.logger.Errorw("error in getting dependencies from json object", "err", err)
		return err
	}
	filterDependencyTypes := []bean.DevtronResourceDependencyType{
		bean.DevtronResourceDependencyTypeLevel,
		bean.DevtronResourceDependencyTypeUpstream,
	}
	appIdsToGetMetadata := helper.GetDependencyOldObjectIdsForSpecificType(dependenciesOfParent, bean.DevtronResourceDependencyTypeUpstream)
	dependencyFilterKeys, err := impl.getFilterKeysFromDependenciesInfo(query.DependenciesInfo)
	if err != nil {
		return err
	}
	appMetadataObj, appIdNameMap, err := impl.getMapOfAppMetadata(appIdsToGetMetadata)
	if err != nil {
		return err
	}
	metadataObj := &bean.DependencyMetaDataBean{
		MapOfAppsMetadata: appMetadataObj,
	}
	response.Dependencies = impl.getFilteredDependenciesWithMetaData(dependenciesOfParent, filterDependencyTypes, dependencyFilterKeys, metadataObj, appIdNameMap)
	return nil
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
	sourceReleaseConfig := gjson.Get(configDataJsonObj, bean.ReleaseResourceArtifactSourceReleaseConfigPath).String()
	sourceReleaseConfigObj := &bean.DtResourceObjectInternalDescriptorBean{}
	if len(sourceReleaseConfig) != 0 {
		err := json.Unmarshal([]byte(sourceReleaseConfig), sourceReleaseConfigObj)
		if err != nil {
			impl.logger.Errorw("error encountered in un-marshaling sourceReleaseConfig", "sourceReleaseConfig", sourceReleaseConfig, "err", err)
			return err
		}
		if sourceReleaseConfigObj.Id > 0 { //it might be possible that the id is not present as dependencies creation will save empty values
			obj, err := impl.getExistingDevtronObject(sourceReleaseConfigObj.Id, 0, sourceReleaseConfigObj.DevtronResourceSchemaId, sourceReleaseConfigObj.Identifier)
			if err != nil {
				impl.logger.Errorw("error encountered in updateReleaseDependencyConfigDataInObj", "sourceReleaseConfigObjId", sourceReleaseConfigObj.Id, "err", err)
				return err
			}
			sourceReleaseConfigObj.ReleaseVersion = gjson.Get(obj.ObjectData, bean.ReleaseResourceObjectReleaseVersionPath).String()
			sourceReleaseConfigObj.Name = gjson.Get(obj.ObjectData, bean.ResourceObjectNamePath).String()
		}
	}
	sourceAppWfId := int(gjson.Get(configDataJsonObj, bean.ReleaseResourceArtifactSourceAppWfIdPath).Int())
	artifactId := int(gjson.Get(configDataJsonObj, bean.ReleaseResourceDependencyConfigArtifactIdKey).Int())

	if !isLite {
		// getting artifact git commit data and image at runtime by artifact id instead of setting this schema, this has to be modified when commit source is also kept in schema (eg ci trigger is introduced)
		artifact, err := impl.ciArtifactRepository.Get(artifactId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error encountered in updateReleaseDependencyConfigDataInObj", "artifactId", artifactId, "err", err)
			return err
		}
		configData.ReleaseInstruction = gjson.Get(configDataJsonObj, bean.ReleaseResourceDependencyConfigReleaseInstructionKey).String()
		configData.CiWorkflowId = int(gjson.Get(configDataJsonObj, bean.ReleaseResourceDependencyConfigCiWorkflowKey).Int())
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
			ArtifactId:          artifactId,
			Image:               artifact.Image,
			RegistryType:        gjson.Get(configDataJsonObj, bean.ReleaseResourceDependencyConfigRegistryTypeKey).String(),
			RegistryName:        gjson.Get(configDataJsonObj, bean.ReleaseResourceDependencyConfigRegistryNameKey).String(),
			SourceAppWorkflowId: sourceAppWfId,
			CommitSource:        gitCommitData,
			SourceReleaseConfig: sourceReleaseConfigObj,
		}
	} else { //adding basic artifact config data for liter for internal calls
		configData.ArtifactConfig = &bean.ArtifactConfig{
			ArtifactId:          artifactId,
			SourceReleaseConfig: sourceReleaseConfigObj,
			SourceAppWorkflowId: sourceAppWfId,
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
				err = util.GetApiErrorAdapter(http.StatusNotFound, "404", bean.InvalidResourceKindOrVersion, bean.InvalidResourceKindOrVersion)
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
					err = util.GetApiErrorAdapter(http.StatusNotFound, "404", bean.InvalidQueryDependencyInfo, bean.InvalidQueryDependencyInfo)
				}
				return nil, err
			}
			identifierString = strconv.Itoa(id)
		} else {
			identifierString = bean.AllIdentifierQueryString
		}
		dependencyFilterKey := helper.GetFilterKeyObjectFromId(dependencyResourceSchema.Id, identifierString)
		if !slices.Contains(dependencyFilterKeys, dependencyFilterKey) {
			dependencyFilterKeys = append(dependencyFilterKeys, dependencyFilterKey)
		}
	}
	return dependencyFilterKeys, nil
}

func (impl *DevtronResourceServiceImpl) getArtifactResponseForDependency(dependency *bean.DevtronResourceDependencyBean, toFetchArtifactData bool,
	mapOfArtifactIdAndReleases map[int][]*bean.DtResourceObjectOverviewDescriptorBean, appWorkflowId int, searchArtifactTag,
	searchImageTag string, artifactIdsForReleaseTrackFilter []int, limit, offset int, userId int32) (bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse], error) {
	artifactResponse := bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse]{
		Id:              dependency.OldObjectId,
		Identifier:      dependency.Identifier,
		ResourceKind:    dependency.ResourceKind,
		ResourceVersion: dependency.ResourceVersion,
	}
	if toFetchArtifactData {
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
			//no workflow components, not looking for artifact further
			return artifactResponse, nil
		}
		workflowFilterMap := make(map[string]int) //map of "componentType-componentId" and appWorkflowId
		var ciPipelineIds, externalCiPipelineIds, cdPipelineIds []int
		for appWfId, workflowComponents := range workflowComponentsMap {
			if workflowComponents.CiPipelineId != 0 {
				ciPipelineIds = append(ciPipelineIds, workflowComponents.CiPipelineId)
				workflowFilterMap[fmt.Sprintf("%s-%d", appWorkflow.CIPIPELINE, workflowComponents.CiPipelineId)] = appWfId
			}
			if workflowComponents.ExternalCiPipelineId != 0 {
				externalCiPipelineIds = append(externalCiPipelineIds, workflowComponents.ExternalCiPipelineId)
				workflowFilterMap[fmt.Sprintf("%s-%d", appWorkflow.WEBHOOK, workflowComponents.ExternalCiPipelineId)] = appWfId
			}
			for _, cdPipelineId := range workflowComponents.CdPipelineIds {
				workflowFilterMap[fmt.Sprintf("%s-%d", appWorkflow.CDPIPELINE, cdPipelineId)] = appWfId
				cdPipelineIds = append(cdPipelineIds, cdPipelineId)
			}
		}
		request := &bean3.WorkflowComponentsBean{
			AppId:                 dependency.OldObjectId,
			CiPipelineIds:         ciPipelineIds,
			ExternalCiPipelineIds: externalCiPipelineIds,
			CdPipelineIds:         cdPipelineIds,
			SearchArtifactTag:     searchArtifactTag,
			SearchImageTag:        searchImageTag,
			ArtifactIds:           artifactIdsForReleaseTrackFilter,
			Limit:                 limit,
			Offset:                offset,
			UserId:                userId,
		}
		data, err := impl.appArtifactManager.RetrieveArtifactsForAppWorkflows(request, workflowFilterMap)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in getting artifact list for", "request", request, "err", err)
			return artifactResponse, err
		}
		// Note: Overriding bean.CiArtifactResponse.TagsEditable as we are not supporting Image Tags edit from UI in V1
		// TODO: to be removed when supported in UI
		data.TagsEditable = false
		if len(data.CiArtifacts) > 0 {
			for i := range data.CiArtifacts {
				ciArtifact := data.CiArtifacts[i]
				if releases, ok := mapOfArtifactIdAndReleases[ciArtifact.Id]; ok {
					ciArtifact.ConfiguredInReleases = releases
				}
				data.CiArtifacts[i] = ciArtifact
			}
			artifactResponse.Data = &data
		}
	}
	return artifactResponse, nil
}

func getReleaseConfigOptionsFilterCriteriaData(query *apiBean.GetConfigOptionsQueryParams) (appWorkflowId int, releaseTrackFilter *bean.FilterCriteriaDecoder, err error) {
	for _, filterCriteria := range query.FilterCriteria {
		criteriaDecoder, err := util3.DecodeFilterCriteriaString(filterCriteria)
		if err != nil {
			return appWorkflowId, nil, err
		}
		switch criteriaDecoder.Resource {
		case bean.DevtronResourceAppWorkflow:
			if criteriaDecoder.Type != bean.IdQueryString {
				return appWorkflowId, nil, fmt.Errorf("invalid filterCriteria: AppWorkflow")
			}
			appWorkflowId, err = strconv.Atoi(criteriaDecoder.Value)
			if err != nil {
				return appWorkflowId, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
			}
		case bean.DevtronResourceReleaseTrack:
			if criteriaDecoder.Type == bean.IdQueryString {
				releaseTrackId, err := strconv.Atoi(criteriaDecoder.Value)
				if err != nil {
					return appWorkflowId, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
				}
				criteriaDecoder.ValueInt = releaseTrackId
			}
			releaseTrackFilter = criteriaDecoder
		default:
			return appWorkflowId, nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidFilterCriteria, bean.InvalidFilterCriteria)
		}
	}
	return appWorkflowId, releaseTrackFilter, nil
}

func getReleaseConfigOptionsSearchKeyData(query *apiBean.GetConfigOptionsQueryParams) (searchArtifactTag, searchImageTag string, err error) {
	searchDecoder, err := util3.DecodeSearchKeyString(query.SearchKey)
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

func (impl *DevtronResourceServiceImpl) updateReleaseArtifactListInResponseObject(reqBean *bean.DevtronResourceObjectDescriptorBean,
	resourceObject *repository.DevtronResourceObject, query *apiBean.GetConfigOptionsQueryParams) ([]bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse], error) {
	response := make([]bean.DependencyConfigOptions[*serviceBean.CiArtifactResponse], 0)
	filterDependencyTypes := []bean.DevtronResourceDependencyType{
		bean.DevtronResourceDependencyTypeUpstream,
	}
	dependenciesWithMetaData, err := impl.getReleaseDependenciesData(resourceObject, filterDependencyTypes, query.DependenciesInfo, true)
	if err != nil {
		return nil, err
	}
	appWorkflowId, releaseTrackFilter, err := getReleaseConfigOptionsFilterCriteriaData(query)
	if err != nil {
		impl.logger.Errorw("error encountered in decodeFilterCriteriaForConfigOptions", "filterCriteria", query.FilterCriteria, "err", bean.InvalidFilterCriteria)
		return nil, err
	}
	toFetchArtifactData := true
	mapOfArtifactIdAndReleases := make(map[int][]*bean.DtResourceObjectOverviewDescriptorBean)
	var artifactIdsForReleaseTrackFilter []int
	if releaseTrackFilter != nil {
		//this filter enables user to get configured images in releases of this release track where same workflow is present
		releaseTrackDescriptorBean := &bean.DevtronResourceObjectDescriptorBean{
			Kind:       bean.DevtronResourceReleaseTrack.ToString(),
			Version:    bean.DevtronResourceVersionAlpha1.ToString(),
			Id:         releaseTrackFilter.ValueInt,
			Identifier: releaseTrackFilter.Value, //todo: make conditional on basis of filter criteria type
		}
		_, releaseTrack, err := impl.getResourceSchemaAndExistingObject(releaseTrackDescriptorBean)
		if err != nil {
			impl.logger.Errorw("error, getResourceSchemaAndExistingObject", "err", err, "descriptorBean", releaseTrackDescriptorBean)
			return nil, err
		}
		releases, err := impl.devtronResourceObjectRepository.GetChildObjectsByParentArgAndSchemaId(releaseTrack.Id, bean.IdDbColumnKey, releaseTrack.DevtronResourceSchemaId)
		if err != nil {
			impl.logger.Errorw("error, GetChildObjectsByParentArgAndSchemaId", "err", err, "descriptorBean", err)
			return nil, err
		}
		for _, release := range releases {
			rolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(release.ObjectData, bean.ReleaseResourceRolloutStatusPath).String())
			if rolloutStatus == bean.PartiallyDeployedReleaseRolloutStatus || rolloutStatus == bean.CompletelyDeployedReleaseRolloutStatus {
				//getting dependencies
				dependencies, err := impl.getDependenciesInObjectDataFromJsonString(release.DevtronResourceSchemaId, release.ObjectData, true)
				if err != nil {
					impl.logger.Errorw("error, getDependenciesInObjectDataFromJsonString", "err", err, "objectData", release.ObjectData)
					return nil, err
				}
				overviewDescBean := impl.getReleaseOverviewDescriptorBeanFromObject(release)
				for _, dependency := range dependencies {
					artifactId := 0
					if dependency.Config != nil && dependency.Config.ArtifactConfig != nil {
						artifactId = dependency.Config.ArtifactConfig.ArtifactId
					}
					if artifactId > 0 { //just appending all selected artifacts and app workflow filter will be added later as old artifacts configured do not have appWorkflowId saved with them
						mapOfArtifactIdAndReleases[artifactId] = append(mapOfArtifactIdAndReleases[artifactId], overviewDescBean)
						artifactIdsForReleaseTrackFilter = append(artifactIdsForReleaseTrackFilter, artifactId)
					}
				}
			}
		}
		toFetchArtifactData = len(artifactIdsForReleaseTrackFilter) != 0
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
		artifactResponse, err := impl.getArtifactResponseForDependency(dependency, toFetchArtifactData, mapOfArtifactIdAndReleases, appWorkflowId, searchArtifactTag,
			searchImageTag, artifactIdsForReleaseTrackFilter, query.Limit, query.Offset, reqBean.UserId)
		if err != nil {
			return nil, err
		}
		response = append(response, artifactResponse)
	}
	return response, nil
}

func (impl *DevtronResourceServiceImpl) updateReleaseDependencyChildObjectsInObj(dependencyString string) ([]*bean.ChildObject, error) {
	childObjects := make([]*bean.ChildObject, 0)
	oldObjectId := int(gjson.Get(dependencyString, bean.IdKey).Int())
	childInheritance, err := getChildInheritanceData(dependencyString)
	if err != nil {
		return nil, err
	}
	envs, err := impl.getEnvironmentsForApplicationDependency(childInheritance, oldObjectId)
	if err != nil {
		impl.logger.Errorw("error encountered in updateReleaseDependencyChildObjectsInObj", "id", oldObjectId, "err", err)
		return nil, err
	}
	if len(envs) > 0 {
		childObject := adapter.BuildChildObject(envs, bean.EnvironmentChildObjectType)
		childObjects = append(childObjects, childObject)
	}
	return childObjects, nil
}

func (impl *DevtronResourceServiceImpl) checkIfReleaseResourcePatchValid(objectData string, queries []bean.PatchQuery) error {
	operationsPaths := make([]bean4.PolicyReleaseOperationPath, 0, len(queries))
	for _, query := range queries {
		operationsPaths = append(operationsPaths, bean4.PatchQueryPathReleasePolicyOpPathMap[query.Path])
	}
	isValid, err := impl.checkIfReleaseResourceOperationValid(objectData, bean4.PolicyReleaseOperationTypePatch, operationsPaths)
	if err != nil {
		impl.logger.Errorw("err, checkIfReleaseResourcePatchValid", "err", err, "objectData", objectData)
		return err
	}
	if !isValid {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ActionPolicyInValidDueToStatusErrMessage, bean.ActionPolicyInValidDueToStatusErrMessage)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) checkIfReleaseResourceTaskRunValid(req *bean.DevtronResourceTaskExecutionBean, existingResourceObject *repository.DevtronResourceObject) error {
	isValid, err := impl.checkIfReleaseResourceOperationValid(existingResourceObject.ObjectData, bean4.PolicyReleaseOperationTypeDeploymentTrigger, nil)
	if err != nil {
		impl.logger.Errorw("err, checkIfReleaseResourceTaskRunValid", "err", err, "objectData", existingResourceObject.ObjectData)
		return err
	}
	if !isValid {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ActionPolicyInValidDueToStatusErrMessage, bean.ActionPolicyInValidDueToStatusErrMessage)
	}

	taskRunInfo, err := impl.fetchReleaseTaskRunInfoWithFilters(&bean.TaskInfoPostApiBean{DevtronResourceObjectDescriptorBean: req.DevtronResourceObjectDescriptorBean}, &apiBean.GetTaskRunInfoQueryParams{IsLite: true}, existingResourceObject)
	if err != nil {
		impl.logger.Errorw("error encountered in checkIfReleaseResourceTaskRunValid", "descriptorBean", req.DevtronResourceObjectDescriptorBean, "err", err)
		return err
	}
	err = checkIfTaskExecutionAllowed(req.Tasks, taskRunInfo.Data)
	if err != nil {
		impl.logger.Errorw("error encountered in checkIfReleaseResourceTaskRunValid", "taskRunInfo", taskRunInfo, "err", err)
		return err
	}

	return nil
}

func checkIfTaskExecutionAllowed(tasks []*bean.Task, taskInfo []bean.DtReleaseTaskRunInfo) error {
	levelIndexVsDeploymentAllowedMap := make(map[int]*bool, len(taskInfo))
	for _, info := range taskInfo {
		levelIndexVsDeploymentAllowedMap[info.Level] = info.TaskRunAllowed
	}
	for _, task := range tasks {
		if val, ok := levelIndexVsDeploymentAllowedMap[task.LevelIndex]; !ok {
			return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidLevelIndexOrLevelIndexChangedMessage, bean.InvalidLevelIndexOrLevelIndexChangedMessage)
		} else if ok && !*val {
			return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.StageTaskExecutionNotAllowedMessage, bean.StageTaskExecutionNotAllowedMessage)
		}
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) checkIfReleaseResourceDeletionValid(objectData string) error {
	isValid, err := impl.checkIfReleaseResourceOperationValid(objectData, bean4.PolicyReleaseOperationTypeDelete, nil)
	if err != nil {
		impl.logger.Errorw("err, checkIfReleaseResourceDeletionValid", "err", err, "objectData", objectData)
		return err
	}
	if !isValid {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ActionPolicyInValidDueToStatusErrMessage, bean.ActionPolicyInValidDueToStatusErrMessage)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) checkIfReleaseResourceOperationValid(objectData string, operationType bean4.PolicyReleaseOperationType,
	operationPaths []bean4.PolicyReleaseOperationPath) (bool, error) {
	stateFrom, err := adapter.GetPolicyDefinitionStateFromReleaseObject(objectData)
	if err != nil {
		impl.logger.Errorw("error, GetPolicyDefinitionStateFromReleaseObject", "err", err, "objectData", objectData)
		return false, err
	}
	return impl.releasePolicyEvaluationService.EvaluateReleaseActionRequest(operationType, operationPaths, stateFrom)
}

// performFeasibilityChecks performs feasibility checks for a task for kind release
// STEP 1: get required map using bulk operations and use the required maps appVsArtifactIdMap, artifactIdVsArtifactMap, pipelineIdVsPipelineMap to fetch the data.
// STEP 2: Run feasibility for every task by getting required data from map formed in Step 1 And 2, err is not blocking here
func (impl *DevtronResourceServiceImpl) performFeasibilityChecks(ctx context.Context, req *bean.DevtronResourceTaskExecutionBean, objectData string, dRSchemaId int) ([]*bean.TaskExecutionResponseBean, error) {
	tasks := req.Tasks
	taskExecutionResponse := make([]*bean.TaskExecutionResponseBean, 0, len(tasks))
	valid := helper.ValidateTasksPayload(tasks)
	if !valid {
		return taskExecutionResponse, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.NoTaskFoundMessage, bean.NoTaskFoundMessage)
	}
	appVsArtifactIdMap, artifactIdVsArtifactMap, pipelineIdVsPipelineMap, _, err := impl.getArtifactAndPipelineMapFromTasks(req.Tasks, objectData, dRSchemaId)
	if err != nil {
		impl.logger.Errorw("error encountered in performFeasibilityChecks", "objectData", objectData, "err", err)
		return nil, err
	}
	// performing feasibility for every task.
	for _, task := range tasks {
		// details fetch
		pipeline := pipelineIdVsPipelineMap[task.PipelineId]
		if pipeline == nil {
			// handling case when pipeline is deleted or not found, for feasibility
			continue
		}
		env := pipeline.Environment
		app := pipeline.App
		artifact := artifactIdVsArtifactMap[appVsArtifactIdMap[task.AppId]]
		// request from details
		scope := resourceQualifiers.BuildScope(task.AppId, env.Id, env.ClusterId, app.TeamId, env.Default)
		triggerRequest := adapter2.GetTriggerRequest(pipeline, artifact, req.UserId, adapter2.GetTriggerContext(ctx), task.CdWorkflowType)
		triggerRequirementRequest := adapter2.GetTriggerRequirementRequest(scope, triggerRequest, resourceFilter.WorkflowTypeToReferenceType(task.CdWorkflowType), task.CdWorkflowType.GetDeploymentStageType())
		// checking feasibility
		_, _, deploymentWindowByPassed, err := impl.triggerService.CheckFeasibility(triggerRequirementRequest)
		// err is not blocking here as feasibility breaks returns error with message and error code, just if err is nil we will convert that to success status
		err = helper.ConvertErrorAccordingToFeasibility(err, deploymentWindowByPassed)
		taskExecutionResponse = append(taskExecutionResponse, adapter.BuildTaskExecutionResponseBean(task.AppId, env.Id, app.AppName, env.Name, env.IsVirtualEnvironment, err, nil))
	}
	return taskExecutionResponse, nil
}

// executeDeploymentsForDependencies will execute deployments(pre-cd, deploy, post-cd) for release
// STEP 1: Get required map using bulk operations and use the required maps appVsArtifactIdMap, _, pipelineIdVsPipelineMap to fetch the data.
// STEP 2: Start a transaction
// STEP 3: bulk create cd workflow and cd workflow runners
// STEP 4: bulk create devtron resource task run objects
// STEP 5: Publish cd bulk event to nats for every task
func (impl *DevtronResourceServiceImpl) executeDeploymentsForDependencies(ctx context.Context, req *bean.DevtronResourceTaskExecutionBean, existingObject *repository.DevtronResourceObject) ([]*bean.TaskExecutionResponseBean, error) {
	tasks := req.Tasks
	taskExecutionResponse := make([]*bean.TaskExecutionResponseBean, 0, len(tasks))
	valid := helper.ValidateTasksPayload(tasks)
	if !valid {
		return taskExecutionResponse, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.NoTaskFoundMessage, bean.NoTaskFoundMessage)
	}
	objectData := existingObject.ObjectData
	dRSchemaId := existingObject.DevtronResourceSchemaId
	req.Id = existingObject.Id
	req.OldObjectId = existingObject.OldObjectId

	// recording here for overall consistency
	triggeredTime := time.Now()
	triggeredBy := req.UserId
	// setting time in request
	req.TriggeredTime = triggeredTime

	appVsArtifactIdMap, _, pipelineIdVsPipelineMap, appIdVsDrSchemaDetails, err := impl.getArtifactAndPipelineMapFromTasks(req.Tasks, objectData, dRSchemaId)
	if err != nil {
		impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "objectData", objectData, "err", err)
		return nil, err
	}
	// Starting a transaction
	tx, err := impl.dtResourceTaskRunRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "err", err)
		return nil, err
	}
	defer impl.dtResourceTaskRunRepository.RollbackTx(tx)

	// bulk create cd workflow
	// bulk create cd workflow runner
	pipelineCiArtifactKeyVsWorkflowIdMap, cdWorkflowIdVsRunnerIdMap, err := impl.bulkCreateCdWorkFlowAndRunners(tx, tasks, triggeredBy, triggeredTime, appVsArtifactIdMap, pipelineIdVsPipelineMap)
	if err != nil {
		impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "err", err)
		return nil, err
	}
	// create task run in bulk with identifiers
	savedTaskRuns, err := impl.bulkCreateDevtronResourceTaskRunObjects(tx, req, appVsArtifactIdMap, pipelineCiArtifactKeyVsWorkflowIdMap, cdWorkflowIdVsRunnerIdMap, appIdVsDrSchemaDetails, existingObject)
	if err != nil {
		impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "err", err)
		return nil, err
	}

	// committing transaction
	err = impl.dtResourceTaskRunRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in commiting transaction in executeDeploymentsForDependencies", "err", err)
		return nil, err
	}
	checkForRolloutStatusPatch := false
	// performing bulk deployment
	var currentlyExecutedPipelineIds []int
	for _, task := range tasks {
		// details fetch
		pipeline := pipelineIdVsPipelineMap[task.PipelineId]
		if pipeline == nil {
			// handling case when pipeline is deleted or not found
			continue
		}
		env := pipeline.Environment
		app := pipeline.App
		artifactId := appVsArtifactIdMap[task.AppId]
		cdWorkflowId := pipelineCiArtifactKeyVsWorkflowIdMap[util2.GetKeyForPipelineIdAndArtifact(task.PipelineId, appVsArtifactIdMap[task.AppId])]
		cdWorkflowRunnerId := cdWorkflowIdVsRunnerIdMap[cdWorkflowId]

		triggerErr := impl.cdPipelineEventPublishService.PublishBulkTriggerTopicEvent(task.PipelineId, task.AppId, artifactId, req.UserId, task.CdWorkflowType, cdWorkflowId, cdWorkflowRunnerId)
		if triggerErr != nil {
			impl.logger.Errorw("error encountered in executeDeploymentsForDependencies: PublishBulkTriggerTopicEvent", "err", triggerErr, "pipelineId", task.PipelineId, "artifactId", artifactId)
		}
		triggerErr = helper.ConvertErrorAccordingToDeployment(triggerErr)
		taskExecutionResponse = append(taskExecutionResponse, adapter.BuildTaskExecutionResponseBean(task.AppId, env.Id, app.AppName, env.Name, env.IsVirtualEnvironment, nil, triggerErr))
		checkForRolloutStatusPatch = true
		currentlyExecutedPipelineIds = append(currentlyExecutedPipelineIds, task.PipelineId)
	}

	currentRolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(objectData, bean.ReleaseResourceRolloutStatusPath).String())
	if checkForRolloutStatusPatch && !currentRolloutStatus.IsPartiallyDeployed() {
		found, err := impl.isAnyNewPipelineTriggered(req, existingObject, savedTaskRuns, currentlyExecutedPipelineIds)
		if err != nil {
			impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "err", err)
			return nil, err
		}
		if found {
			//check if firstReleasedOn is present, if not then update it
			firstReleasedOnExists := len(gjson.Get(objectData, bean.ReleaseResourceObjectFirstReleasedOnPath).String()) != 0
			if !firstReleasedOnExists {
				objectDataNew, err := helper.PatchResourceObjectDataAtAPath(objectData, bean.ReleaseResourceObjectFirstReleasedOnPath, triggeredTime)
				if err != nil {
					impl.logger.Errorw("error, PatchResourceObjectData", "err", err, "releaseObj", objectData, "path", bean.ReleaseResourceObjectFirstReleasedOnPath, "value", triggeredTime)
					return nil, err
				}
				existingObject.ObjectData = objectDataNew
			}
			// updated existing object (rollout status to partially deployed)
			err = impl.updateRolloutStatusInExistingObject(existingObject,
				bean.PartiallyDeployedReleaseRolloutStatus, triggeredBy, triggeredTime)
			if err != nil {
				impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "err", err)
				return nil, err
			}
		}
	}
	return taskExecutionResponse, nil
}

func (impl *DevtronResourceServiceImpl) isAnyNewPipelineTriggered(req *bean.DevtronResourceTaskExecutionBean,
	existingObject *repository.DevtronResourceObject, savedTaskRuns []*repository.DevtronResourceTaskRun, currentlyExecutedPipelineIds []int) (bool, error) {
	dtResourceId := existingObject.DevtronResourceId
	dRSchemaId := existingObject.DevtronResourceSchemaId
	releaseObjectId := req.GetResourceIdByIdType()

	rsIdentifier := helper.GetTaskRunSourceIdentifier(releaseObjectId, req.IdType, dtResourceId, dRSchemaId)
	var appDependencyIdentifiers []string
	var excludedTaskRunIds []int
	for _, savedTaskRun := range savedTaskRuns {
		excludedTaskRunIds = append(excludedTaskRunIds, savedTaskRun.Id)
		if !slices.Contains(appDependencyIdentifiers, savedTaskRun.RunSourceDependencyIdentifier) {
			appDependencyIdentifiers = append(appDependencyIdentifiers, savedTaskRun.RunSourceDependencyIdentifier)
		}
	}
	previousCdWfrIds, err := impl.getExecutedCdWfrIdsFromTaskRun(rsIdentifier, appDependencyIdentifiers, excludedTaskRunIds...)
	previouslyExecutedPipelineIds, err := impl.ciCdPipelineOrchestrator.GetCdPipelineIdsForRunnerIds(previousCdWfrIds)
	if err != nil {
		impl.logger.Errorw("error in getting previously executed pipelineIds", "err", err, "previousCdWfrIds", previousCdWfrIds)
		return false, err
	}
	for _, currentlyExecutedPipelineId := range currentlyExecutedPipelineIds {
		if !slices.Contains(previouslyExecutedPipelineIds, currentlyExecutedPipelineId) {
			return true, nil
		}
	}
	return false, nil
}

// BulkCreateCdWorkFlowAndRunners created cd workflows and cd workflow runner sin bulk for all tasks
func (impl *DevtronResourceServiceImpl) bulkCreateCdWorkFlowAndRunners(tx *pg.Tx, tasks []*bean.Task, triggeredBy int32, triggeredTime time.Time, appVsArtifactIdMap map[int]int, pipelineIdVsPipelineMap map[int]*pipelineConfig.Pipeline) (pipelineCiArtifactKeyVsWorkflowIdMap map[string]int, cdWorkflowIdVsRunnerIdMap map[int]int, err error) {
	lenOfTasks := len(tasks)
	cdWorkflows := make([]*bean2.CdWorkflowDto, 0, lenOfTasks)
	cdWorkflowsRunner := make([]*bean2.CdWorkflowRunnerDto, 0, lenOfTasks)

	// building cd workflow dtos for bulk cd workflow creation
	for _, task := range tasks {
		pipeline := pipelineIdVsPipelineMap[task.PipelineId]
		if pipeline == nil {
			continue
		}
		cdWorkflows = append(cdWorkflows, adapter3.BuildCdWorkflowDto(task.PipelineId, appVsArtifactIdMap[task.AppId], triggeredBy))
	}
	if len(cdWorkflows) == 0 {
		//handling cases for deleted pipeline deployment triggered
		return pipelineCiArtifactKeyVsWorkflowIdMap, cdWorkflowIdVsRunnerIdMap, nil
	}
	// bulk cd workflow creation
	pipelineCiArtifactKeyVsWorkflowIdMap, err = impl.cdWorkflowService.CreateBulkCdWorkflow(tx, cdWorkflows, triggeredTime)
	if err != nil {
		impl.logger.Errorw("error encountered in BulkCreateCdWorkFlowAndRunners", "err", err, "cdWorkflows", cdWorkflows)
		return pipelineCiArtifactKeyVsWorkflowIdMap, cdWorkflowIdVsRunnerIdMap, err
	}

	// building cd workflow runners dtos for bulk cd workflow runners creation
	for _, task := range tasks {
		pipeline := pipelineIdVsPipelineMap[task.PipelineId]
		if pipeline == nil {
			continue
		}
		env := pipeline.Environment
		//fetching  cd workflow id from logic -
		// STEP 1: get artifact id from task app id
		// STEP 2: get cd workflow id from task pipeline id and artifact id
		cdWorkflowId := pipelineCiArtifactKeyVsWorkflowIdMap[util2.GetKeyForPipelineIdAndArtifact(pipeline.Id, appVsArtifactIdMap[task.AppId])]

		cdWorkflowsRunner = append(cdWorkflowsRunner, impl.triggerService.GetCdWorkflowRunnerWithEnvConfig(task.CdWorkflowType, pipeline, env.Namespace, cdWorkflowId, triggeredBy, triggeredTime))
	}
	// bulk creating runners
	cdWorkflowIdVsRunnerIdMap, err = impl.cdWorkflowRunnerService.CreateBulkCdWorkflowRunners(tx, cdWorkflowsRunner)
	if err != nil {
		impl.logger.Errorw("error encountered in BulkCreateCdWorkFlowAndRunners", "err", err, "cdWorkflows", cdWorkflows)
		return pipelineCiArtifactKeyVsWorkflowIdMap, cdWorkflowIdVsRunnerIdMap, err
	}
	return pipelineCiArtifactKeyVsWorkflowIdMap, cdWorkflowIdVsRunnerIdMap, nil
}

// getArtifactAndPipelineMapFromTasks returns three maps appVsArtifactIdMap, artifactIdVsArtifactMap , artifactIdVsArtifactMap
// STEP 1: appVsArtifactIdMap maps app vs artifact id mapped in resource object artifact config
// STEP 2: GetArtifactsIds in bulk and appId Vs Artifact id Map
// STEP 3: GetPipelineId vs Pipeline Details map( pipeline here has env, cluster ,app)
func (impl *DevtronResourceServiceImpl) getArtifactAndPipelineMapFromTasks(tasks []*bean.Task, objectData string, dRSchemaId int) (appVsArtifactIdMap map[int]int, artifactIdVsArtifactMap map[int]*repository2.CiArtifact, pipelineIdVsPipelineMap map[int]*pipelineConfig.Pipeline, appIdVsDrSchemaDetail map[int]*bean.DependencyDetail, err error) {
	taskLength := len(tasks)
	appVsArtifactIdMap = make(map[int]int, taskLength)
	appIds := make([]int, 0, taskLength)
	pipelineIds := make([]int, 0, taskLength)
	for _, task := range tasks {
		//initially setting all appId to artifact id as 0, will be updated after getting from dependencies (devtron resource object)
		appVsArtifactIdMap[task.AppId] = 0
		appIds = append(appIds, task.AppId)
		pipelineIds = append(pipelineIds, task.PipelineId)
	}
	// updated appVsArtifactId map
	artifactIds, appIdVsDrSchemaDetail, err := impl.updateArtifactIdAndReturnIdsForDependencies(dRSchemaId, objectData, appVsArtifactIdMap)
	if err != nil {
		impl.logger.Errorw("error encountered in performFeasibilityChecks", "objectData", objectData, "appVsArtifactIdMap", appVsArtifactIdMap, "err", err)
		return appVsArtifactIdMap, artifactIdVsArtifactMap, pipelineIdVsPipelineMap, appIdVsDrSchemaDetail, err
	}
	// getting all artifacts here with mapping artifact id to artifact
	artifactIdVsArtifactMap, err = impl.getArtifactIdVsArtifactMapForIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error encountered in performFeasibilityChecks", "objectData", objectData, "artifactIds", artifactIds, "err", err)
		return appVsArtifactIdMap, artifactIdVsArtifactMap, pipelineIdVsPipelineMap, appIdVsDrSchemaDetail, err
	}
	// getting all pipelines here with mapping pipeline id to pipeline
	pipelineIdVsPipelineMap, err = impl.getPipelineIdPipelineMapForPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error encountered in performFeasibilityChecks", "objectData", objectData, "pipelineIdVsPipelineMap", pipelineIdVsPipelineMap, "err", err)
		return appVsArtifactIdMap, artifactIdVsArtifactMap, pipelineIdVsPipelineMap, appIdVsDrSchemaDetail, err
	}
	return appVsArtifactIdMap, artifactIdVsArtifactMap, pipelineIdVsPipelineMap, appIdVsDrSchemaDetail, nil

}

func (impl *DevtronResourceServiceImpl) updateArtifactIdAndReturnIdsForDependencies(devtronResourceSchemaId int, objectData string, appVsArtifactIdMap map[int]int) ([]int, map[int]*bean.DependencyDetail, error) {
	dependenciesResult := gjson.Get(objectData, bean.ResourceObjectDependenciesPath)
	dependenciesResultArr := dependenciesResult.Array()
	artifactsIds := make([]int, 0, len(dependenciesResultArr))
	appIdVsDrSchemaDetail := make(map[int]*bean.DependencyDetail)
	kind, subKind, version, err := helper.GetKindSubKindAndVersionOfResourceBySchemaId(devtronResourceSchemaId,
		impl.devtronResourcesSchemaMapById, impl.devtronResourcesMapById)
	if err != nil {
		return artifactsIds, appIdVsDrSchemaDetail, err
	}
	parentResourceType := &bean.DevtronResourceTypeReq{
		ResourceKind:    bean.DevtronResourceKind(kind),
		ResourceSubKind: bean.DevtronResourceKind(subKind),
		ResourceVersion: bean.DevtronResourceVersion(version),
	}
	for _, dependencyResult := range dependenciesResultArr {
		dependencyBean, err := impl.getDependencyBeanFromJsonString(parentResourceType, dependencyResult.String(), false)
		if err != nil {
			return artifactsIds, appIdVsDrSchemaDetail, err
		}
		// currently checking schema id to be of kind devtron-applicationas we create images in application only currently.
		if _, ok := appVsArtifactIdMap[dependencyBean.OldObjectId]; ok && helper.IsApplicationDependency(dependencyBean.DevtronResourceTypeReq) {
			//updating artifact id here
			appVsArtifactIdMap[dependencyBean.OldObjectId] = dependencyBean.Config.ArtifactConfig.ArtifactId
			// this is new map to get id, id-type, devtronResourceId,devtronSchemaId,kept it here as we need to calculate for given app ids
			appIdVsDrSchemaDetail[dependencyBean.OldObjectId] = adapter.BuildDependencyDetail(dependencyBean.OldObjectId, dependencyBean.IdType, dependencyBean.DevtronResourceId, dependencyBean.DevtronResourceSchemaId)
			artifactsIds = append(artifactsIds, dependencyBean.Config.ArtifactConfig.ArtifactId)
		}
	}
	return artifactsIds, appIdVsDrSchemaDetail, nil
}

func (impl *DevtronResourceServiceImpl) getPipelineIdPipelineMapForPipelineIds(pipelineIds []int) (map[int]*pipelineConfig.Pipeline, error) {
	pipelineIdByPipelineMap := make(map[int]*pipelineConfig.Pipeline, len(pipelineIds))
	cdPipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error encountered in GetAppIdVsAppDetailsMapForAppIds", "err", err)
		return pipelineIdByPipelineMap, err
	}
	for _, pipeline := range cdPipelines {
		pipelineIdByPipelineMap[pipeline.Id] = pipeline
	}
	return pipelineIdByPipelineMap, nil

}

func (impl *DevtronResourceServiceImpl) getArtifactIdVsArtifactMapForIds(artifactsIds []int) (map[int]*repository2.CiArtifact, error) {
	artifactIdVsArtifactMap := make(map[int]*repository2.CiArtifact, len(artifactsIds))
	artifacts, err := impl.ciArtifactRepository.GetByIds(artifactsIds)
	if err != nil {
		impl.logger.Errorw("error encountered in getArtifactIdVsArtifactMapForIds", "err", err)
		return artifactIdVsArtifactMap, err
	}
	for _, artifact := range artifacts {
		artifactIdVsArtifactMap[artifact.Id] = artifact
	}
	return artifactIdVsArtifactMap, nil
}

func (impl *DevtronResourceServiceImpl) getAppAndCdWfrIdsForDependencies(releaseRunSourceIdentifier string, dependencies []*bean.DevtronResourceDependencyBean) (appIds, cdWfrIds []int, err error) {
	var appDependencyIdentifiers []string
	for _, dependency := range dependencies {
		// iterating in every inheritance and getting child inheritances(for eg cd) and getting corresponding details) for now it is ["*"] we will fetch all cd (pipelines) for that dependency
		findAll := false
		for _, inheritance := range dependency.ChildInheritance {
			// collecting selectors here currently only ["*"] is present so will find all CD Pipelines for an app but can be modified in future
			findAll = slices.Contains(inheritance.Selector, bean.DefaultCdPipelineSelector)
		}
		if findAll {
			appDependencyIdentifier := helper.GetTaskRunSourceDependencyIdentifier(dependency.OldObjectId, dependency.IdType, dependency.DevtronResourceId, dependency.DevtronResourceSchemaId)
			appDependencyIdentifiers = append(appDependencyIdentifiers, appDependencyIdentifier)
			appIds = append(appIds, dependency.OldObjectId)
		} else {
			// if specific, will have to find corresponding ids and get the details
			// currently we don't store specific childInheritance
			impl.logger.Errorw("invalid childInheritance selector data", "appId", dependency.OldObjectId)
			return nil, nil, fmt.Errorf("invalid childInheritance selector! Implementation not supported")
		}
	}
	if len(appDependencyIdentifiers) != 0 {
		cdWfrIds, err = impl.getExecutedCdWfrIdsFromTaskRun(releaseRunSourceIdentifier, appDependencyIdentifiers)
		if err != nil {
			impl.logger.Errorw("invalid childInheritance selector data", "err", err, "appDependencyIdentifiers", appDependencyIdentifiers)
			return nil, nil, err
		}
	}
	return appIds, cdWfrIds, nil
}

func (impl *DevtronResourceServiceImpl) getExecutedCdWfrIdsFromTaskRun(releaseRunSourceIdentifier string,
	appDependencyIdentifiers []string, excludedTaskRunIds ...int) (cdWfrIds []int, err error) {
	taskTypes := []bean.TaskType{bean.TaskTypeDeployment, bean.TaskTypePreDeployment, bean.TaskTypePostDeployment}
	dtResourceTaskRuns, err := impl.dtResourceTaskRunRepository.GetByRunSourceAndTaskTypes(releaseRunSourceIdentifier,
		appDependencyIdentifiers, taskTypes, excludedTaskRunIds)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, dtResourceTaskRun := range dtResourceTaskRuns {
		cdWfrIds = append(cdWfrIds, dtResourceTaskRun.TaskTypeIdentifier)
	}
	return cdWfrIds, nil
}

// getReleaseDeploymentInfoForDependencies gets every data point for dependencies in CdPipelineReleaseInfo object
func (impl *DevtronResourceServiceImpl) getReleaseDeploymentInfoForDependencies(releaseRunSourceIdentifier string, dependencies []*bean.DevtronResourceDependencyBean) ([]*bean.CdPipelineReleaseInfo, map[string]*bean.CdPipelineReleaseInfo, error) {
	appIds, cdWfrIds, err := impl.getAppAndCdWfrIdsForDependencies(releaseRunSourceIdentifier, dependencies)
	if err != nil {
		impl.logger.Errorw("error encountered in getReleaseDeploymentInfoForDependencies", "releaseRunSourceIdentifier", releaseRunSourceIdentifier, "err", err)
		return nil, nil, err
	}
	pipelinesInfo, pipelineIdAppIdKeyVsReleaseInfo, err := impl.ciCdPipelineOrchestrator.GetCdPipelinesReleaseInfoForApp(appIds, cdWfrIds)
	if err != nil {
		impl.logger.Errorw("error encountered in getReleaseDeploymentInfoForDependencies", "appIds", appIds, "cdWfrIds", cdWfrIds, "err", err)
		return nil, nil, err
	}
	return pipelinesInfo, pipelineIdAppIdKeyVsReleaseInfo, nil
}

// getReleaseDeploymentInfoForDependenciesFromMap gets every data point for dependencies in CdPipelineReleaseInfo object from map
func (impl *DevtronResourceServiceImpl) getReleaseDeploymentInfoForDependenciesFromMap(dependencies []*bean.DevtronResourceDependencyBean, pipelineIdAppIdKeyVsReleaseInfo map[string]*bean.CdPipelineReleaseInfo) ([]*bean.CdPipelineReleaseInfo, error) {
	pipelinesInfo := make([]*bean.CdPipelineReleaseInfo, 0)
	appIds := make([]int, 0, len(dependencies))
	for _, dependency := range dependencies {
		// not checking child inheritance as callee has already filtered all child inheritance case.
		appIds = append(appIds, dependency.OldObjectId)
	}
	// getting info from pipelineAppIdVsReleaseInfoMap if app id is in appIds calculated from level dependencies
	for key, info := range pipelineIdAppIdKeyVsReleaseInfo {
		appId, err := helper.GetAppIdFromPipelineIdAppIdKey(key)
		if err != nil {
			impl.logger.Errorw("error encountered in getReleaseDeploymentInfoForDependenciesFromMap", "key", key, "err", err)
			return nil, err
		}
		if slices.Contains(appIds, appId) {
			pipelinesInfo = append(pipelinesInfo, info)
		}
	}

	return pipelinesInfo, nil
}

func (impl *DevtronResourceServiceImpl) getEnvironmentsForApplicationDependency(childInheritance []*bean.ChildInheritance, appId int) ([]*bean.CdPipelineEnvironment, error) {
	// iterating in every inheritance and getting child inheritances(for eg cd) and getting corresponding details) for now it is ["*"] we will fetch all cd (env) for that dependency
	envs := make([]*bean.CdPipelineEnvironment, 0)
	findAll := false
	for _, inheritance := range childInheritance {
		// collecting selectors here currently only ["all"] is present so will find all env names for an app but can be modified in future
		findAll = slices.Contains(inheritance.Selector, bean.DefaultCdPipelineSelector)
	}
	if findAll {
		pipelines, err := impl.pipelineRepository.FindEnvNameAndIdByAppId(appId)
		if err != nil {
			impl.logger.Errorw("error encountered in getEnvironmentsForApplicationDependency", "err", err)
			return envs, err
		}
		for _, pipeline := range pipelines {
			env := adapter.BuildCdPipelineEnvironmentBasicData(pipeline.Environment.Name, pipeline.DeploymentAppType, pipeline.EnvironmentId, pipeline.Id)
			envs = append(envs, env)
		}
	} else {
		// if specific, will have to find corresponding ids and get the details
		// currently we don't store specific childInheritance
	}

	return envs, nil
}

func (impl *DevtronResourceServiceImpl) fetchReleaseTaskRunInfo(req *bean.DevtronResourceObjectDescriptorBean, query *apiBean.GetTaskRunInfoQueryParams,
	existingResourceObject *repository.DevtronResourceObject) ([]bean.DtReleaseTaskRunInfo, error) {
	if existingResourceObject == nil || existingResourceObject.Id == 0 {
		impl.logger.Warnw("invalid get request, object not found", "req", req)
		return nil, util.GetApiErrorAdapter(http.StatusNotFound, "404", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
	}
	response := make([]bean.DtReleaseTaskRunInfo, 0)

	req.Id = existingResourceObject.Id
	req.OldObjectId = existingResourceObject.OldObjectId
	resourceId := req.GetResourceIdByIdType()
	rsIdentifier := helper.GetTaskRunSourceIdentifier(resourceId, req.IdType, existingResourceObject.DevtronResourceId, existingResourceObject.DevtronResourceSchemaId)

	levelFilterCondition := bean.NewDependencyFilterCondition().
		WithFilterByTypes(bean.DevtronResourceDependencyTypeLevel)
	if query.LevelIndex != 0 {
		levelFilterCondition = levelFilterCondition.WithFilterByIndexes(query.LevelIndex)
	}
	levelDependencies := GetDependenciesBeanFromObjectData(existingResourceObject.ObjectData, levelFilterCondition)
	var levelToAppDependenciesMap map[int][]*bean.DevtronResourceDependencyBean
	if !query.IsLite {
		applicationFilterCondition := bean.NewDependencyFilterCondition().
			WithFilterByTypes(bean.DevtronResourceDependencyTypeUpstream).
			WithFilterByDependentOnIndex(query.LevelIndex).
			WithChildInheritance()
		if query.LevelIndex != 0 {
			applicationFilterCondition = applicationFilterCondition.WithFilterByDependentOnIndex(query.LevelIndex)
		}
		applicationDependencies := GetDependenciesBeanFromObjectData(existingResourceObject.ObjectData, applicationFilterCondition)
		levelToAppDependenciesMap = adapter.MapDependenciesByDependentOnIndex(applicationDependencies)
	}
	for _, levelDependency := range levelDependencies {
		dtReleaseTaskRunInfo := bean.DtReleaseTaskRunInfo{
			Level: levelDependency.Index,
		}
		if query.IsLite {
			taskRunAllowed := true
			lastStageResponse := adapter.GetLastReleaseTaskRunInfo(response)
			if lastStageResponse != nil && !lastStageResponse.IsTaskRunAllowed() {
				taskRunAllowed = false
				dtReleaseTaskRunInfo.TaskRunAllowed = &taskRunAllowed
			} else {
				previousLevelIndex := getPreviousLevelDependency(levelDependencies, levelDependency.Index)
				if previousLevelIndex != 0 {
					previousAppFilterCondition := bean.NewDependencyFilterCondition().
						WithFilterByTypes(bean.DevtronResourceDependencyTypeUpstream).
						WithFilterByDependentOnIndex(previousLevelIndex).
						WithChildInheritance()
					previousLevelAppDependencies := GetDependenciesBeanFromObjectData(existingResourceObject.ObjectData, previousAppFilterCondition)
					if len(previousLevelAppDependencies) != 0 {
						appIds, cdWfrIds, err := impl.getAppAndCdWfrIdsForDependencies(rsIdentifier, previousLevelAppDependencies)
						if err != nil {
							impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "rsIdentifier", rsIdentifier, "previousLevelIndex", previousLevelIndex, "err", err)
							return nil, err
						}
						if len(cdWfrIds) == 0 {
							taskRunAllowed = false
						} else {
							taskRunAllowed, err = impl.ciCdPipelineOrchestrator.IsEachAppDeployedOnAtLeastOneEnvWithRunnerIds(appIds, cdWfrIds)
							if err != nil {
								impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "appIds", appIds, "cdWfrIds", cdWfrIds, "err", err)
								return nil, err
							}
						}
					} else {
						taskRunAllowed = false
					}
				}
				dtReleaseTaskRunInfo.TaskRunAllowed = &taskRunAllowed
			}
		} else {
			dependencies := make([]*bean.CdPipelineReleaseInfo, 0)
			if levelToAppDependenciesMap != nil && levelToAppDependenciesMap[levelDependency.Index] != nil {
				dependencyBean, _, err := impl.getReleaseDeploymentInfoForDependencies(rsIdentifier, levelToAppDependenciesMap[levelDependency.Index])
				if err != nil {
					impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "rsIdentifier", rsIdentifier, "stage", levelDependency.Index, "err", err)
					return nil, err
				}
				dependencies = append(dependencies, dependencyBean...)
			}
			dtReleaseTaskRunInfo.Dependencies = dependencies
		}
		response = append(response, dtReleaseTaskRunInfo)
	}
	// If the last stage of the release has TaskRunAllowed; then verify the release rollout status
	// If all the applications in all stages are deployed to their respective environments,
	// Mark the rollout status -> bean.CompletelyDeployedReleaseRolloutStatus
	lastStageResponse := adapter.GetLastReleaseTaskRunInfo(response)
	if query.IsLite && lastStageResponse != nil && lastStageResponse.IsTaskRunAllowed() {
		err := impl.markRolloutStatusIfAllDependenciesGotSucceed(existingResourceObject, rsIdentifier, nil)
		if err != nil {
			impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "rsIdentifier", rsIdentifier, "existingResourceObjectId", existingResourceObject.Id, "err", err)
			return nil, err
		}
	}
	return response, nil
}

func (impl *DevtronResourceServiceImpl) fetchReleaseTaskRunInfoWithFilters(req *bean.TaskInfoPostApiBean, query *apiBean.GetTaskRunInfoQueryParams,
	existingResourceObject *repository.DevtronResourceObject) (*bean.DeploymentTaskInfoResponse, error) {
	if existingResourceObject == nil || existingResourceObject.Id == 0 {
		impl.logger.Warnw("invalid get request, object not found", "req", req)
		return nil, util.GetApiErrorAdapter(http.StatusNotFound, "404", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
	}
	response := make([]bean.DtReleaseTaskRunInfo, 0)

	req.Id = existingResourceObject.Id
	req.OldObjectId = existingResourceObject.OldObjectId
	resourceId := req.GetResourceIdByIdType()
	rsIdentifier := helper.GetTaskRunSourceIdentifier(resourceId, req.IdType, existingResourceObject.DevtronResourceId, existingResourceObject.DevtronResourceSchemaId)

	// getting all release Info and corresponding map to use in further processing at level/stage or showAll(rollout status)
	pipelineIdAppIdKeyVsReleaseInfo, allCdPipelineReleaseInfo, taskInfoCount, err := impl.fetchAllReleaseInfoStatusWithMap(existingResourceObject, rsIdentifier)
	if err != nil {
		impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfoWithFilters", "rsIdentifier", rsIdentifier, "err", err)
		return nil, err
	}

	if query.IsLite {
		response, err = impl.getOnlyLevelDataForTaskInfo(existingResourceObject.ObjectData, pipelineIdAppIdKeyVsReleaseInfo, query.LevelIndex)
		if err != nil {
			impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfoWithFilters", "rsIdentifier", rsIdentifier, "err", err)
			return nil, err
		}
	} else {
		var appDevtronResourceSchemaId int
		filterConditionReq := bean.NewFilterConditionInternalBean()
		// filters decoding
		if len(req.FilterCriteria) > 0 {
			// setting filter values from filters
			filterConditionReq, err = impl.getFilterConditionBeanFromDecodingFilters(req)
			if err != nil {
				impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfoWithFilters", "err", err)
				return nil, err
			}

			// finding application/devtron-application schema id for filters check in dependency
			applicationSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(bean.DevtronResourceApplication.ToString(), bean.DevtronResourceDevtronApplication.ToString(), bean.DevtronResourceVersion1.ToString())
			if err != nil {
				impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfoWithFilters", "err", err)
				return nil, err
			}
			appDevtronResourceSchemaId = applicationSchema.Id
		} else {
			filterConditionReq.RequestWithoutFilters = true
		}

		if query.ShowAll {
			// rollout status page api
			// apply filters and return updated dependencies
			updatedCdPipelineReleaseInfo := impl.applyFiltersToDependencies(filterConditionReq, allCdPipelineReleaseInfo)
			return &bean.DeploymentTaskInfoResponse{TaskInfoCount: taskInfoCount, Data: []bean.DtReleaseTaskRunInfo{{Dependencies: updatedCdPipelineReleaseInfo}}}, nil

		}
		// deploy page api
		response, err = impl.getLevelDataWithDependenciesForTaskInfo(filterConditionReq, existingResourceObject.ObjectData, rsIdentifier, query.LevelIndex, appDevtronResourceSchemaId, pipelineIdAppIdKeyVsReleaseInfo)
		if err != nil {
			impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfoWithFilters", "rsIdentifier", rsIdentifier, "err", err)
			return nil, err
		}
	}
	return &bean.DeploymentTaskInfoResponse{TaskInfoCount: taskInfoCount, Data: response}, nil
}

// setFilterConditionInRequest decodes the filters provided and sets teh filters in request for further processing
func (impl *DevtronResourceServiceImpl) getFilterConditionBeanFromDecodingFilters(req *bean.TaskInfoPostApiBean) (*bean.FilterConditionInternalBean, error) {
	appIdsFilters, envIdsFilters, deploymentStatus, rolloutStatus, err := impl.getAppIdsAndEnvIdsFromFilterCriteria(req.FilterCriteria)
	if err != nil {
		impl.logger.Errorw("error encountered in setFilterConditionInRequest", "id", req.Id, "err", err)
		return nil, err
	}

	return adapter.BuildFilterConditionInternalBean(appIdsFilters, envIdsFilters, rolloutStatus, deploymentStatus), nil
}

// getOnlyLevelDataForTaskInfo gets only level data with task run allowed operation.(signifies lite mode)
func (impl *DevtronResourceServiceImpl) getOnlyLevelDataForTaskInfo(objectData string, pipelineIdAppIdKeyVsReleaseInfo map[string]*bean.CdPipelineReleaseInfo, levelIndex int) ([]bean.DtReleaseTaskRunInfo, error) {
	response := make([]bean.DtReleaseTaskRunInfo, 0)
	levelDependencies := impl.getLevelDependenciesFromObjectData(objectData, levelIndex)
	var err error
	for _, levelDependency := range levelDependencies {
		dtReleaseTaskRunInfo := bean.DtReleaseTaskRunInfo{
			Level: levelDependency.Index,
		}
		taskRunAllowed := true
		lastStageResponse := adapter.GetLastReleaseTaskRunInfo(response)
		if lastStageResponse != nil && !lastStageResponse.IsTaskRunAllowed() {
			taskRunAllowed = false
			dtReleaseTaskRunInfo.TaskRunAllowed = &taskRunAllowed
		} else {
			appIds := make([]int, 0)
			previousLevelIndex := getPreviousLevelDependency(levelDependencies, levelDependency.Index)
			if previousLevelIndex != 0 {
				previousAppFilterCondition := bean.NewDependencyFilterCondition().
					WithFilterByTypes(bean.DevtronResourceDependencyTypeUpstream).
					WithFilterByDependentOnIndex(previousLevelIndex).
					WithChildInheritance()
				previousLevelAppDependencies := GetDependenciesBeanFromObjectData(objectData, previousAppFilterCondition)
				if len(previousLevelAppDependencies) != 0 {
					for _, dependency := range previousLevelAppDependencies {
						// not considering child inheritance as map as already filtered out child inheritance
						appIds = append(appIds, dependency.OldObjectId)
					}

					taskRunAllowed, err = impl.isEachAppDeployedOnAtLeastOneEnvWithMap(appIds, pipelineIdAppIdKeyVsReleaseInfo)
					if err != nil {
						impl.logger.Errorw("error encountered in getOnlyLevelDataForTaskInfo", "appIds", appIds, "err", err)
						return nil, err
					}
				} else {
					taskRunAllowed = false
				}
			}
			dtReleaseTaskRunInfo.TaskRunAllowed = &taskRunAllowed
		}
		response = append(response, dtReleaseTaskRunInfo)
	}
	return response, nil
}

func (impl *DevtronResourceServiceImpl) isEachAppDeployedOnAtLeastOneEnvWithMap(appIds []int, pipelineIdAppIdKeyVsReleaseInfo map[string]*bean.CdPipelineReleaseInfo) (bool, error) {
	appIdToSuccessCriteriaFlag := make(map[int]bool, len(appIds))
	for key, info := range pipelineIdAppIdKeyVsReleaseInfo {
		appId, err := helper.GetAppIdFromPipelineIdAppIdKey(key)
		if err != nil {
			impl.logger.Errorw("error encountered in IsEachAppDeployedOnAtLeastOneEnvWithMap", "key", key, "err", err)
			return false, err
		}
		if appIdToSuccessCriteriaFlag[appId] {
			continue
		}
		// if this appId from map is not in provided appIds continue as we don't need to calculate for thsi appId
		if !slices.Contains(appIds, appId) {
			continue
		}
		// if not of (deployment exist and status is succeeded)
		if info.ExistingStages.Deploy && !helper.IsStatusSucceeded(info.DeployStatus) {
			continue
		}
		// if not of (pre exist and status is succeeded)
		if info.ExistingStages.Pre && !helper.IsStatusSucceeded(info.PreStatus) {
			continue
		}
		// if not of (post exist and status is succeeded)
		if info.ExistingStages.Post && !helper.IsStatusSucceeded(info.PostStatus) {
			continue
		}
		appIdToSuccessCriteriaFlag[appId] = true
	}
	if len(appIds) != len(appIdToSuccessCriteriaFlag) {
		return false, nil
	}
	return true, nil
}

// getLevelDataWithDependenciesForTaskInfo get level wise data with dependencies in it, supports level Index key if not 0 get level of that index with its dependencies
func (impl *DevtronResourceServiceImpl) getLevelDataWithDependenciesForTaskInfo(req *bean.FilterConditionInternalBean, objectData, rsIdentifier string, levelIndex, appDevtronResourceSchemaId int, pipelineIdAppIdKeyVsReleaseInfo map[string]*bean.CdPipelineReleaseInfo) ([]bean.DtReleaseTaskRunInfo, error) {
	response := make([]bean.DtReleaseTaskRunInfo, 0)
	var levelToAppDependenciesMap map[int][]*bean.DevtronResourceDependencyBean

	levelDependencies := impl.getLevelDependenciesFromObjectData(objectData, levelIndex)
	applicationFilterCondition := bean.NewDependencyFilterCondition().
		WithFilterByTypes(bean.DevtronResourceDependencyTypeUpstream).
		WithFilterByDependentOnIndex(levelIndex).
		WithFilterByIdAndSchemaId(req.AppIds, appDevtronResourceSchemaId).
		WithChildInheritance()
	applicationDependencies := GetDependenciesBeanFromObjectData(objectData, applicationFilterCondition)
	levelToAppDependenciesMap = adapter.MapDependenciesByDependentOnIndex(applicationDependencies)

	for _, levelDependency := range levelDependencies {
		dtReleaseTaskRunInfo := bean.DtReleaseTaskRunInfo{
			Level: levelDependency.Index,
		}
		dependencies := make([]*bean.CdPipelineReleaseInfo, 0)
		if levelToAppDependenciesMap != nil && levelToAppDependenciesMap[levelDependency.Index] != nil {
			dependencyBean, err := impl.getReleaseDeploymentInfoForDependenciesFromMap(levelToAppDependenciesMap[levelDependency.Index], pipelineIdAppIdKeyVsReleaseInfo)
			if err != nil {
				impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "rsIdentifier", rsIdentifier, "stage", levelDependency.Index, "err", err)
				return nil, err
			}
			// applying filters for rollout , env and deployment status (app ids filters is already handled, one level while fetching applicationDependencies
			dependencyBean = impl.applyFiltersToDependencies(req, dependencyBean)
			dependencies = append(dependencies, dependencyBean...)
		}
		dtReleaseTaskRunInfo.Dependencies = dependencies

		response = append(response, dtReleaseTaskRunInfo)
	}
	return response, nil
}

// fetchAllReleaseInfoStatusWithMap will fetch all release info with mappings
func (impl *DevtronResourceServiceImpl) fetchAllReleaseInfoStatusWithMap(existingResourceObject *repository.DevtronResourceObject, rsIdentifier string) (map[string]*bean.CdPipelineReleaseInfo, []*bean.CdPipelineReleaseInfo, *bean.TaskInfoCount, error) {
	var taskInfoCount *bean.TaskInfoCount
	var dependencyBean []*bean.CdPipelineReleaseInfo
	var pipelineIdAppIdKeyVsReleaseInfo map[string]*bean.CdPipelineReleaseInfo
	var err error
	allApplicationDependencies := getAllApplicationDependenciesFromObjectData(existingResourceObject.ObjectData)
	if len(allApplicationDependencies) != 0 {
		// will get map here for all application dependencies and all data as well
		dependencyBean, pipelineIdAppIdKeyVsReleaseInfo, err = impl.getReleaseDeploymentInfoForDependencies(rsIdentifier, allApplicationDependencies)
		if err != nil {
			impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "rsIdentifier", rsIdentifier, "err", err)
			return nil, nil, nil, err
		}
		// If all the applications in all stages are deployed to their respective environments,
		// Mark the rollout status -> bean.CompletelyDeployedReleaseRolloutStatus
		err = impl.markRolloutStatusIfAllDependenciesGotSucceedFromMap(existingResourceObject, pipelineIdAppIdKeyVsReleaseInfo, allApplicationDependencies)
		if err != nil {
			impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfoWithFilters", "rsIdentifier", rsIdentifier, "existingResourceObjectId", existingResourceObject.Id, "err", err)
			return nil, nil, nil, err
		}

		stageWiseDeploymentStatusCount, releaseDeploymentStatusCount := getStageWiseAndReleaseDeploymentStatusCountFromPipelineInfo(dependencyBean)
		taskInfoCount = adapter.BuildTaskInfoCount(releaseDeploymentStatusCount, stageWiseDeploymentStatusCount)
	}
	return pipelineIdAppIdKeyVsReleaseInfo, dependencyBean, taskInfoCount, nil
}

func (impl *DevtronResourceServiceImpl) applyFiltersToDependencies(req *bean.FilterConditionInternalBean, cdPipelineReleaseInfo []*bean.CdPipelineReleaseInfo) []*bean.CdPipelineReleaseInfo {
	if req.RequestWithoutFilters {
		return cdPipelineReleaseInfo
	}
	updatedCdPipelineReleaseInfo := make([]*bean.CdPipelineReleaseInfo, 0, len(cdPipelineReleaseInfo))
	for _, info := range cdPipelineReleaseInfo {
		if len(req.AppIds) != 0 && !slices.Contains(req.AppIds, info.AppId) {
			//continue in case app id filters len is greater than 0 and does not contain app id.
			continue
		}
		if len(req.EnvIds) != 0 && !slices.Contains(req.EnvIds, info.EnvId) {
			//continue in case env id filters len is greater than 0 and does not contain env id.
			continue
		}
		if len(req.ReleaseDeploymentStatus) != 0 && !slices.Contains(req.ReleaseDeploymentStatus, info.ReleaseDeploymentStatus.ToString()) {
			//continue in case ReleaseDeploymentStatus filters len is greater than 0 and does not contain ReleaseDeploymentStatus.
			continue
		}
		//pre
		if values, ok := req.StageWiseDeploymentStatus[bean3.CD_WORKFLOW_TYPE_PRE]; ok && len(values) > 0 {
			if !info.ExistingStages.Pre || !slices.Contains(values, info.PreStatus) {
				//continue in case pre stage filters ln is greater than 0 and does not contain pre status of info.
				continue
			}

		}
		//deploy
		if values, ok := req.StageWiseDeploymentStatus[bean3.CD_WORKFLOW_TYPE_DEPLOY]; ok && len(values) > 0 {
			if !info.ExistingStages.Deploy || !slices.Contains(values, info.DeployStatus) {
				//continue in case deploy stage filters ln is greater than 0 and does not contain deploy status of info.
				continue
			}
		}
		//post
		if values, ok := req.StageWiseDeploymentStatus[bean3.CD_WORKFLOW_TYPE_POST]; ok && len(values) > 0 {
			if !info.ExistingStages.Post || !slices.Contains(values, info.PostStatus) {
				//continue in case post stage filters ln is greater than 0 and does not contain post status of info.
				continue
			}
		}

		updatedCdPipelineReleaseInfo = append(updatedCdPipelineReleaseInfo, info)
	}

	return updatedCdPipelineReleaseInfo
}

func (impl *DevtronResourceServiceImpl) getLevelDependenciesFromObjectData(objectData string, levelIndex int) []*bean.DevtronResourceDependencyBean {
	levelFilterCondition := bean.NewDependencyFilterCondition().
		WithFilterByTypes(bean.DevtronResourceDependencyTypeLevel)
	if levelIndex != 0 {
		levelFilterCondition = levelFilterCondition.WithFilterByIndexes(levelIndex)
	}
	return GetDependenciesBeanFromObjectData(objectData, levelFilterCondition)
}

func processReleaseDeploymentStatusVsCountMapForResponse(rolloutStatusVsCountMap map[bean.ReleaseDeploymentStatus]int) *bean.ReleaseDeploymentStatusCount {
	completedCount := 0
	yetToTriggerCount := 0
	failedCount := 0
	ongoingCount := 0

	if val, ok := rolloutStatusVsCountMap[bean.Completed]; ok {
		completedCount = val
	}
	if val, ok := rolloutStatusVsCountMap[bean.YetToTrigger]; ok {
		yetToTriggerCount = val
	}
	if val, ok := rolloutStatusVsCountMap[bean.Failed]; ok {
		failedCount = val
	}
	if val, ok := rolloutStatusVsCountMap[bean.Ongoing]; ok {
		ongoingCount = val
	}
	return adapter.BuildReleaseDeploymentStatus(ongoingCount, yetToTriggerCount, failedCount, completedCount)
}

func getStageWiseAndReleaseDeploymentStatusCountFromPipelineInfo(pipelinesInfo []*bean.CdPipelineReleaseInfo) (*bean.StageWiseStatusCount, *bean.ReleaseDeploymentStatusCount) {
	preStatusVsCountMap := make(map[string]int, len(pipelinesInfo))
	deployStatusVsCountMap := make(map[string]int, len(pipelinesInfo))
	postStatusVsCountMap := make(map[string]int, len(pipelinesInfo))
	releaseDeploymentStatusVsCountMap := make(map[bean.ReleaseDeploymentStatus]int, len(pipelinesInfo))
	for _, pipeline := range pipelinesInfo {
		if pipeline.ExistingStages.Pre {
			preStatusVsCountMap[pipeline.PreStatus] = preStatusVsCountMap[pipeline.PreStatus] + 1
		}
		if pipeline.ExistingStages.Post {
			postStatusVsCountMap[pipeline.PostStatus] = postStatusVsCountMap[pipeline.PostStatus] + 1
		}
		if pipeline.ExistingStages.Deploy {
			// cd pipeline will always exist added this check intentionally for validation
			deployStatusVsCountMap[pipeline.DeployStatus] = deployStatusVsCountMap[pipeline.DeployStatus] + 1
		}
		releaseDeploymentStatusVsCountMap[pipeline.ReleaseDeploymentStatus] = releaseDeploymentStatusVsCountMap[pipeline.ReleaseDeploymentStatus] + 1
	}
	preStatusCount := processPreOrPostDeploymentVsCountMapForResponse(preStatusVsCountMap)
	deployStatusCount := processDeploymentVsCountMapForResponse(deployStatusVsCountMap)
	postStatusCount := processPreOrPostDeploymentVsCountMapForResponse(postStatusVsCountMap)
	releaseDeploymentCount := processReleaseDeploymentStatusVsCountMapForResponse(releaseDeploymentStatusVsCountMap)
	return adapter.BuildStageWiseStatusCount(preStatusCount, deployStatusCount, postStatusCount), releaseDeploymentCount
}

func processPreOrPostDeploymentVsCountMapForResponse(preOrPostStatusVsCountMap map[string]int) *bean.PrePostStatusCount {
	notTriggered := 0
	failed := 0
	inProgress := 0
	succeeded := 0
	others := 0
	if val, ok := preOrPostStatusVsCountMap[pipelineStageBean.NotTriggered]; ok {
		notTriggered = val
	}
	if val, ok := preOrPostStatusVsCountMap[pipelineConfig.WorkflowFailed]; ok {
		failed = val
	}
	if val, ok := preOrPostStatusVsCountMap[pipelineConfig.WorkflowAborted]; ok {
		failed = failed + val
	}
	if val, ok := preOrPostStatusVsCountMap[executors.WorkflowCancel]; ok {
		failed = failed + val
	}
	if val, ok := preOrPostStatusVsCountMap[bean5.Degraded]; ok {
		failed = failed + val
	}
	if val, ok := preOrPostStatusVsCountMap[bean.Error]; ok {
		failed = failed + val
	}
	if val, ok := preOrPostStatusVsCountMap[pipelineConfig.WorkflowInProgress]; ok {
		inProgress = val
	}
	if val, ok := preOrPostStatusVsCountMap[pipelineConfig.WorkflowStarting]; ok {
		inProgress = inProgress + val
	}
	if val, ok := preOrPostStatusVsCountMap[bean.RunningStatus]; ok {
		inProgress = inProgress + val
	}
	if val, ok := preOrPostStatusVsCountMap[pipelineConfig.WorkflowSucceeded]; ok {
		succeeded = val
	}
	if val, ok := preOrPostStatusVsCountMap[bean5.Healthy]; ok {
		succeeded = succeeded + val
	}
	if val, ok := preOrPostStatusVsCountMap[bean.Unknown]; ok {
		others = val
	}
	if val, ok := preOrPostStatusVsCountMap[bean.Missing]; ok {
		others = others + val
	}
	return adapter.BuildPreOrPostDeploymentCount(notTriggered, failed, succeeded, inProgress, others)

}

func processDeploymentVsCountMapForResponse(deployStatusVsCountMap map[string]int) *bean.DeploymentCount {
	notTriggered := 0
	failed := 0
	inProgress := 0
	succeeded := 0
	queued := 0
	others := 0
	if val, ok := deployStatusVsCountMap[pipelineStageBean.NotTriggered]; ok {
		notTriggered = val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowFailed]; ok {
		failed = val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowAborted]; ok {
		failed = failed + val
	}
	if val, ok := deployStatusVsCountMap[executors.WorkflowCancel]; ok {
		failed = failed + val
	}
	if val, ok := deployStatusVsCountMap[bean5.Degraded]; ok {
		failed = failed + val
	}
	if val, ok := deployStatusVsCountMap[bean.Error]; ok {
		failed = failed + val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowInProgress]; ok {
		inProgress = val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowInitiated]; ok {
		inProgress = inProgress + val
	}
	if val, ok := deployStatusVsCountMap[bean.RunningStatus]; ok {
		inProgress = inProgress + val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowStarting]; ok {
		inProgress = inProgress + val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowSucceeded]; ok {
		succeeded = val
	}
	if val, ok := deployStatusVsCountMap[bean5.Healthy]; ok {
		succeeded = succeeded + val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowInQueue]; ok {
		queued = val
	}
	if val, ok := deployStatusVsCountMap[bean.Unknown]; ok {
		others = val
	}
	if val, ok := deployStatusVsCountMap[bean.Missing]; ok {
		others = others + val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowUnableToFetchState]; ok {
		others = others + val
	}
	if val, ok := deployStatusVsCountMap[pipelineConfig.WorkflowTimedOut]; ok {
		others = others + val
	}
	return adapter.BuildDeploymentCount(notTriggered, failed, succeeded, queued, inProgress, others)

}

// markRolloutStatusIfAllDependenciesGotSucceed get allApplicationDependencies if not provided.
// If all the applications in all stages are deployed to their respective environments,
// Mark the rollout status -> bean.CompletelyDeployedReleaseRolloutStatus
func (impl *DevtronResourceServiceImpl) markRolloutStatusIfAllDependenciesGotSucceed(existingResourceObject *repository.DevtronResourceObject, rsIdentifier string, allApplicationDependencies []*bean.DevtronResourceDependencyBean) (err error) {
	rolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceRolloutStatusPath).String())
	getAllApplicationDependencies := false
	if len(allApplicationDependencies) == 0 {
		getAllApplicationDependencies = true
	}
	if !rolloutStatus.IsCompletelyDeployed() {
		if getAllApplicationDependencies {
			allApplicationDependencies = getAllApplicationDependenciesFromObjectData(existingResourceObject.ObjectData)
		}
		if len(allApplicationDependencies) != 0 {
			appIds, cdWfrIds, err := impl.getAppAndCdWfrIdsForDependencies(rsIdentifier, allApplicationDependencies)
			if err != nil {
				impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "rsIdentifier", rsIdentifier, "err", err)
				return err
			}
			isReleaseCompleted := false
			if len(cdWfrIds) != 0 {
				isReleaseCompleted, err = impl.ciCdPipelineOrchestrator.IsAppsDeployedOnAllEnvWithRunnerIds(appIds, cdWfrIds)
				if err != nil {
					impl.logger.Errorw("error encountered in fetchReleaseTaskRunInfo", "appIds", appIds, "cdWfrIds", cdWfrIds, "err", err)
					return err
				}
			}
			if isReleaseCompleted {
				existingResourceObject.ObjectData, err = sjson.Set(existingResourceObject.ObjectData, bean.ReleaseResourceRolloutStatusPath, bean.CompletelyDeployedReleaseRolloutStatus)
				if err != nil {
					impl.logger.Errorw("error in json set release rollout status completed", "err", err)
					return err
				}
				err = impl.devtronResourceObjectRepository.Update(nil, existingResourceObject)
				if err != nil {
					impl.logger.Errorw("error in updating release rollout status completed", "err", err, "existingResourceObject", existingResourceObject.Id)
					return err
				}
				// updated existing object (rollout status to completely deployed)
				err = impl.updateRolloutStatusInExistingObject(existingResourceObject,
					bean.CompletelyDeployedReleaseRolloutStatus, 1, time.Now())
				if err != nil {
					impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "err", err)
					return err
				}
			}
		}
	}
	return nil
}

// markRolloutStatusIfAllDependenciesGotSucceedFromMap get allApplicationDependencies if not provided.
// If all the applications in all stages are deployed to their respective environments, checks from map
// Mark the rollout status -> bean.CompletelyDeployedReleaseRolloutStatus
func (impl *DevtronResourceServiceImpl) markRolloutStatusIfAllDependenciesGotSucceedFromMap(existingResourceObject *repository.DevtronResourceObject, pipelineIdAppIdKeyVsReleaseInfo map[string]*bean.CdPipelineReleaseInfo, allApplicationDependencies []*bean.DevtronResourceDependencyBean) (err error) {
	rolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(existingResourceObject.ObjectData, bean.ReleaseResourceRolloutStatusPath).String())
	if !rolloutStatus.IsCompletelyDeployed() && len(allApplicationDependencies) != 0 {
		appIds := make([]int, 0, len(allApplicationDependencies))
		// ignoring child inheritance check here as callee has already handled it( GET call will be deprecated in future not handling for that)
		for _, dependency := range allApplicationDependencies {
			appIds = append(appIds, dependency.OldObjectId)
		}
		isReleaseCompleted, err := impl.isAppsDeployedOnAllEnvWithRunnerFromMap(appIds, pipelineIdAppIdKeyVsReleaseInfo)
		if err != nil {
			impl.logger.Errorw("error encountered in markRolloutStatusIfAllDependenciesGotSucceedFromMap", "appIds", appIds, "err", err)
			return err
		}

		if isReleaseCompleted {
			existingResourceObject.ObjectData, err = sjson.Set(existingResourceObject.ObjectData, bean.ReleaseResourceRolloutStatusPath, bean.CompletelyDeployedReleaseRolloutStatus)
			if err != nil {
				impl.logger.Errorw("error in json set release rollout status completed", "err", err)
				return err
			}
			err = impl.devtronResourceObjectRepository.Update(nil, existingResourceObject)
			if err != nil {
				impl.logger.Errorw("error in updating release rollout status completed", "err", err, "existingResourceObject", existingResourceObject.Id)
				return err
			}
			// updated existing object (rollout status to completely deployed)
			err = impl.updateRolloutStatusInExistingObject(existingResourceObject,
				bean.CompletelyDeployedReleaseRolloutStatus, 1, time.Now())
			if err != nil {
				impl.logger.Errorw("error encountered in markRolloutStatusIfAllDependenciesGotSucceedFromMap", "err", err)
				return err
			}
		}
	}
	return nil
}

// isAppsDeployedOnAllEnvWithRunnerFromMap checks from map that every app is deployed on every stage successfully from
func (impl *DevtronResourceServiceImpl) isAppsDeployedOnAllEnvWithRunnerFromMap(appIds []int, pipelineIdAppIdKeyVsReleaseInfo map[string]*bean.CdPipelineReleaseInfo) (bool, error) {
	for key, info := range pipelineIdAppIdKeyVsReleaseInfo {
		appId, err := helper.GetAppIdFromPipelineIdAppIdKey(key)
		if err != nil {
			impl.logger.Errorw("error encountered in isAppsDeployedOnAllEnvWithRunnerFromMap", "key", key, "err", err)
			return false, err
		}
		// if this appId from map is not in provided appIds continue as we don't need to calculate for thsi appId
		if !slices.Contains(appIds, appId) {
			continue
		}
		// if not of (deployment exist and status is succeeded)
		if info.ExistingStages.Deploy && !helper.IsStatusSucceeded(info.DeployStatus) {
			return false, nil
		}
		// if not of (pre exist and status is succeeded)
		if info.ExistingStages.Pre && !helper.IsStatusSucceeded(info.PreStatus) {
			return false, nil
		}
		// if not of (post exist and status is succeeded)
		if info.ExistingStages.Post && !helper.IsStatusSucceeded(info.PostStatus) {
			return false, nil
		}
	}
	return true, nil
}

// getAppIdsAndEnvIdsFromFilterCriteria decodes filters and return app ids ,envIds, deployment status with stage, rollout Status, gets ids for identifiers if identifiers are given
func (impl *DevtronResourceServiceImpl) getAppIdsAndEnvIdsFromFilterCriteria(filters []string) ([]int, []int, map[bean3.WorkflowType][]string, []string, error) {
	// filters decoding
	appIdsFilters, appIdentifierFilters, envIdsFilters, envIdentifierFilters, deploymentStatus, rolloutStatus, err := util3.DecodeFiltersForDeployAndRolloutStatus(filters)
	if err != nil {
		impl.logger.Errorw("error encountered in getAppIdsAndEnvIdsFromFilterCriteria", "err", err, "filters", filters)
		return appIdsFilters, envIdsFilters, deploymentStatus, rolloutStatus, err
	}
	// evaluating app ids from app identifiers as we are maintaining and processing everything in ids
	if len(appIdentifierFilters) > 0 {
		appIds, err := impl.appRepository.FindIdsByNamesAndAppType(appIdentifierFilters, helper2.CustomApp)
		if err != nil {
			impl.logger.Errorw("error encountered in getAppIdsAndEnvIdsFromFilterCriteria", "err", err, "appIdentifierFilters", appIdentifierFilters)
			return appIdsFilters, envIdsFilters, deploymentStatus, rolloutStatus, err
		}
		appIdsFilters = append(appIdsFilters, appIds...)
	}
	if len(envIdentifierFilters) > 0 {
		envIds, err := impl.envService.FindIdsByNames(envIdentifierFilters)
		if err != nil {
			impl.logger.Errorw("error encountered in getAppIdsAndEnvIdsFromFilterCriteria", "err", err, "envIdentifierFilters", envIdentifierFilters)
			return appIdsFilters, envIdsFilters, deploymentStatus, rolloutStatus, err
		}
		envIdsFilters = append(envIdsFilters, envIds...)
	}
	return appIdsFilters, envIdsFilters, deploymentStatus, rolloutStatus, err
}

// getAllApplicationDependenciesFromObjectData gets all upstream dependencies from object data json
func getAllApplicationDependenciesFromObjectData(objectData string) []*bean.DevtronResourceDependencyBean {
	applicationFilterCondition := bean.NewDependencyFilterCondition().
		WithFilterByTypes(bean.DevtronResourceDependencyTypeUpstream).
		WithChildInheritance()

	return GetDependenciesBeanFromObjectData(objectData, applicationFilterCondition)
}

func getPreviousLevelDependency(levelDependencies []*bean.DevtronResourceDependencyBean, currentIndex int) int {
	previousIndex := 0
	for _, levelDependency := range levelDependencies {
		if currentIndex > levelDependency.Index && levelDependency.Index > previousIndex {
			previousIndex = levelDependency.Index
		}
	}
	return previousIndex
}

func (impl *DevtronResourceServiceImpl) performReleaseResourcePatchOperation(objectData string,
	queries []bean.PatchQuery) (*bean.SuccessResponse, string, []string, error) {
	var err error
	auditPaths := make([]string, 0, len(queries))
	patchContainsConfigOrLockStatusQuery := false
	newObjectData := objectData //will be using this for patches and mutation since we need old object for policy check
	for _, query := range queries {
		newObjectData, err = impl.patchQueryForReleaseObject(newObjectData, query)
		if err != nil {
			impl.logger.Errorw("error in patch operation, release track", "err", err, "objectData", "query", query)
			return nil, "", nil, err
		}
		auditPaths = append(auditPaths, bean.PatchQueryPathAuditPathMap[query.Path])
		if query.Path == bean.ReleaseStatusQueryPath || query.Path == bean.ReleaseLockQueryPath {
			patchContainsConfigOrLockStatusQuery = true
		}
	}
	toPerformStatusPolicyCheck := patchContainsConfigOrLockStatusQuery //keeping policy flag different as in future it can not solely depend on status and lock
	if toPerformStatusPolicyCheck {
		newObjectData, err = impl.checkReleasePatchPolicyAndAutoAction(objectData, newObjectData)
		if err != nil {
			impl.logger.Errorw("error, checkReleasePatchPolicyAndAutoAction", "err", err, "objectData", objectData, "newObjectData", newObjectData)
			return nil, "", nil, err
		}
	}
	successResp := adapter.GetSuccessPassResponse()
	if patchContainsConfigOrLockStatusQuery {
		successResp = getSuccessResponseForReleaseStatusOrLockPatch(objectData, newObjectData)
	}
	return successResp, newObjectData, auditPaths, nil
}

func getSuccessResponseForReleaseStatusOrLockPatch(oldObjectData, newObjectData string) *bean.SuccessResponse {
	oldConfigStatus := bean.ReleaseConfigStatus(gjson.Get(oldObjectData, bean.ReleaseResourceConfigStatusStatusPath).String())
	newConfigStatus := bean.ReleaseConfigStatus(gjson.Get(newObjectData, bean.ReleaseResourceConfigStatusStatusPath).String())
	oldLockStatus := gjson.Get(oldObjectData, bean.ReleaseResourceConfigStatusIsLockedPath).Bool()
	newLockStatus := gjson.Get(newObjectData, bean.ReleaseResourceConfigStatusIsLockedPath).Bool()
	configStatusChanged := oldConfigStatus != newConfigStatus
	lockStatusChanged := oldLockStatus != newLockStatus
	if configStatusChanged && lockStatusChanged {
		return adapter.GetReleaseConfigAndLockStatusChangeSuccessResponse(newConfigStatus, newLockStatus)
	} else if configStatusChanged {
		return adapter.GetReleaseConfigStatusChangeSuccessResponse(newConfigStatus)
	} else if lockStatusChanged {
		return adapter.GetReleaseLockStatusChangeSuccessResponse(newLockStatus)
	}
	return adapter.GetSuccessPassResponse()
}

func getReleaseStatusChangePolicyErrResponse(oldObjectData, newObjectData string) error {
	oldConfigStatus := bean.ReleaseConfigStatus(gjson.Get(oldObjectData, bean.ReleaseResourceConfigStatusStatusPath).String())
	newConfigStatus := bean.ReleaseConfigStatus(gjson.Get(newObjectData, bean.ReleaseResourceConfigStatusStatusPath).String())
	oldDepAppCount, oldDepArtifactStatus := adapter.GetReleaseAppCountAndDepArtifactStatusFromResourceObjData(oldObjectData)
	if newConfigStatus == bean.ReadyForReleaseConfigStatus && oldConfigStatus == bean.DraftReleaseConfigStatus {
		if oldDepAppCount == 0 {
			return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ReleaseStatusReadyForReleaseNoAppErrMessage, bean.InvalidPatchOperation)
		} else if oldDepArtifactStatus != bean.AllSelectedDependencyArtifactStatus { // app present but no or partial images selected
			return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ReleaseStatusPatchErrMessage, bean.ReleaseStatusReadyForReleaseNoOrPartialImageErrMessage)
		}
	}
	return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidPatchOperation, bean.InvalidPatchOperation)
}

func (impl *DevtronResourceServiceImpl) checkReleasePatchPolicyAndAutoAction(oldObjectData, newObjectData string) (string, error) {
	stateFrom, err := adapter.GetPolicyDefinitionStateFromReleaseObject(oldObjectData)
	if err != nil {
		impl.logger.Errorw("error, GetPolicyDefinitionStateFromReleaseObject", "err", err, "oldObjectData", oldObjectData)
		return "", err
	}
	stateTo, err := adapter.GetPolicyDefinitionStateFromReleaseObject(newObjectData)
	if err != nil {
		impl.logger.Errorw("error, GetPolicyDefinitionStateFromReleaseObject", "err", err, "newObjectData", newObjectData)
		return "", err
	}

	isValid, autoAction, err := impl.releasePolicyEvaluationService.EvaluateReleaseStatusChangeAndGetAutoAction(stateTo, stateFrom)
	if err != nil {
		impl.logger.Errorw("error, EvaluateReleaseStatusChangeAndGetAutoAction", "err", err, "oldObjectData", oldObjectData, "newObjectData", newObjectData)
		return newObjectData, err
	}
	if !isValid {
		impl.logger.Errorw("error in EvaluateReleaseStatusChangeAndGetAutoAction : invalid action", "oldObjectData", oldObjectData, "newObjectData", newObjectData)

		return newObjectData, getReleaseStatusChangePolicyErrResponse(oldObjectData, newObjectData)
	}
	if autoAction != nil {
		patchQueries, err := adapter.GetPatchQueryForPolicyAutoAction(autoAction, stateTo)
		if err != nil {
			impl.logger.Errorw("error, GetPatchQueryForPolicyAutoAction", "autoAction", autoAction, "stateTo", stateTo, "newObjectData", newObjectData)
			return "", err
		}
		for _, patchQuery := range patchQueries {
			newObjectData, err = impl.patchQueryForReleaseObject(newObjectData, patchQuery)
			if err != nil {
				impl.logger.Errorw("error in auto action patch operation, release track", "err", err, "objectData", "query", patchQuery)
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
	case bean.ReleaseStatusQueryPath:
		objectData, err = impl.patchConfigStatus(objectData, query.Value)
	case bean.ReleaseNoteQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ReleaseResourceObjectReleaseNotePath, query.Value)
	case bean.TagsQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ResourceObjectTagsPath, query.Value)
	case bean.ReleaseLockQueryPath:
		objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ReleaseResourceConfigStatusIsLockedPath, query.Value)
	case bean.NameQueryPath:
		if nameStr, ok := query.Value.(string); !ok || len(nameStr) == 0 {
			return objectData, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.PatchValueNotSupportedError, bean.PatchValueNotSupportedError)
		}
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
	if (configStatus.Status == bean.HoldReleaseConfigStatus || configStatus.Status == bean.RescindReleaseConfigStatus) &&
		len(configStatus.Comment) == 0 {
		return objectData, util.GetApiErrorAdapter(http.StatusBadRequest, "400",
			bean.ReleaseStatusHoldOrRescindPatchNoCommentErrMessage, bean.ReleaseStatusHoldOrRescindPatchNoCommentErrMessage)
	}
	objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ReleaseResourceConfigStatusCommentPath, configStatus.Comment)
	if err != nil {
		return objectData, err
	}
	objectData, err = helper.PatchResourceObjectDataAtAPath(objectData, bean.ReleaseResourceConfigStatusStatusPath, configStatus.Status)

	return objectData, err
}

func (impl *DevtronResourceServiceImpl) validateReleaseDelete(object *repository.DevtronResourceObject) (bool, error) {
	if object == nil || object.Id == 0 {
		return false, util.GetApiErrorAdapter(http.StatusNotFound, "404", bean.ResourceDoesNotExistMessage, bean.ResourceDoesNotExistMessage)
	}
	//getting release rollout status
	rolloutStatus := bean.ReleaseRolloutStatus(gjson.Get(object.ObjectData, bean.ReleaseResourceRolloutStatusPath).String())
	return !rolloutStatus.IsPartiallyDeployed() &&
		!rolloutStatus.IsCompletelyDeployed(), nil
}

func (impl *DevtronResourceServiceImpl) updateRolloutStatusInExistingObject(existingObject *repository.DevtronResourceObject,
	rolloutStatus bean.ReleaseRolloutStatus, triggeredBy int32, triggeredTime time.Time) error {
	newObjectData, err := helper.PatchResourceObjectDataAtAPath(existingObject.ObjectData, bean.ReleaseResourceRolloutStatusPath, rolloutStatus)
	if err != nil {
		impl.logger.Errorw("error encountered in updateRolloutStatusInExistingObject", "err", err)
		return err
	}
	//updating final object data in resource object
	existingObject.ObjectData = newObjectData
	existingObject.UpdatedBy = triggeredBy
	existingObject.UpdatedOn = triggeredTime
	// made it not transaction as we need to commit transaction for cd workflow runners as event has already been sent to nats.
	err = impl.devtronResourceObjectRepository.Update(nil, existingObject)
	if err != nil {
		// made this non-blocking as only roll out status change was not completed but deployment has been triggered as events has been published
		impl.logger.Errorw("error encountered in executeDeploymentsForDependencies", "err", err, "newObjectData", newObjectData)
	}
	impl.dtResourceObjectAuditService.SaveAudit(existingObject, repository.AuditOperationTypeUpdate, []string{bean.ReleaseResourceRolloutStatusPath})
	return nil
}

func (impl *DevtronResourceServiceImpl) setDefaultValueAndValidateForReleaseClone(req *bean.DtResourceObjectCloneReqBean,
	parentConfig *bean.ResourceIdentifier) error {
	err := helper.CheckIfReleaseVersionIsValid(req.Overview.ReleaseVersion)
	if err != nil {
		return err
	}
	adapter.SetIdTypeAndResourceIdBasedOnKind(req.DevtronResourceObjectDescriptorBean, req.OldObjectId)
	identifier, err := impl.getIdentifierForReleaseByParentDescriptorBean(req.Overview.ReleaseVersion, parentConfig)
	if err != nil {
		impl.logger.Errorw("error, getIdentifierForReleaseByParentDescriptorBean", "err", err, "parentConfig", parentConfig, "releaseVersion", req.Overview.ReleaseVersion)
		return err
	}
	req.Identifier = identifier
	if len(req.Name) == 0 {
		req.Name = req.Overview.ReleaseVersion
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) getPathUpdateMapForReleaseClone(req *bean.DtResourceObjectCloneReqBean,
	createdOn time.Time) (map[string]interface{}, error) {
	userObj, err := impl.userRepository.GetById(req.UserId)
	if err != nil {
		impl.logger.Errorw("error in getting user", "err", err, "userId", req.UserId)
		return nil, err
	}
	replaceDataMap := map[string]interface{}{
		bean.ResourceObjectIdPath:                     req.Id,                      //reset Id
		bean.ResourceObjectIdentifierPath:             req.Identifier,              //reset identifier
		bean.ResourceObjectNamePath:                   req.Name,                    //reset name
		bean.ResourceObjectTagsPath:                   req.Overview.Tags,           //reset tags
		bean.ReleaseResourceObjectReleaseVersionPath:  req.Overview.ReleaseVersion, //reset releaseVersion
		bean.ResourceObjectDescriptionPath:            req.Overview.Description,    //reset description
		bean.ReleaseResourceObjectReleaseNotePath:     "",                          //reset note
		bean.ResourceObjectCreatedByIdPath:            userObj.Id,                  //reset created by
		bean.ResourceObjectCreatedByNamePath:          userObj.EmailId,
		bean.ResourceObjectCreatedOnPath:              createdOn,                 //reset created on
		bean.ReleaseResourceConfigStatusPath:          bean.DefaultConfigStatus,  //reset config status
		bean.ReleaseResourceRolloutStatusPath:         bean.DefaultRolloutStatus, //reset rollout status
		bean.ReleaseResourceObjectFirstReleasedOnPath: time.Time{},               //reset first release on time
	}
	return replaceDataMap, nil
}

func (impl *DevtronResourceServiceImpl) getReleaseOverviewDescriptorBeanFromObject(object *repository.DevtronResourceObject) *bean.DtResourceObjectOverviewDescriptorBean {
	resp := &bean.DtResourceObjectOverviewDescriptorBean{
		DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{
			Kind:        bean.DevtronResourceRelease.ToString(),
			Version:     bean.DevtronResourceVersionAlpha1.ToString(),
			OldObjectId: object.Id,
			Identifier:  object.Identifier,
			Name:        gjson.Get(object.ObjectData, bean.ResourceObjectNamePath).String(),
		},
	}
	if gjson.Get(object.ObjectData, bean.ResourceObjectOverviewPath).Exists() {
		resp.ResourceOverview = &bean.ResourceOverview{
			ReleaseVersion: gjson.Get(object.ObjectData, bean.ReleaseResourceObjectReleaseVersionPath).String(),
		}
	}
	return resp
}
