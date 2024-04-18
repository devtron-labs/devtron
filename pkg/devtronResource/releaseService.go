package devtronResource

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net/http"
	"time"
)

func (impl *DevtronResourceServiceImpl) updateReleaseConfigStatusInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, _ = getResourceObjectIdAndType(existingResourceObject)
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

func (impl *DevtronResourceServiceImpl) updateReleaseNoteInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, _ = getResourceObjectIdAndType(existingResourceObject)
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

func (impl *DevtronResourceServiceImpl) updateReleaseOverviewDataInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, resourceObject.IdType = getResourceObjectIdAndType(existingResourceObject)
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
			resourceObject.Overview.CreatedOn, err = getCreatedOnTime(existingResourceObject.ObjectData)
			if err != nil {
				impl.logger.Errorw("error in time conversion", "err", err)
				return err
			}
			resourceObject.Overview.Tags = getOverviewTags(existingResourceObject.ObjectData)
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

func (impl *DevtronResourceServiceImpl) updateCompleteReleaseDataInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	err = impl.updateReleaseOverviewDataInResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting overview data", "err", err)
		return err
	}
	err = impl.updateReleaseConfigStatusInResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting config status data", "err", err)
		return err
	}
	err = impl.updateReleaseNoteInResourceObj(resourceSchema, existingResourceObject, resourceObject)
	if err != nil {
		impl.logger.Errorw("error in getting note", "err", err)
		return err
	}
	return nil
}

func validateCreateReleaseRequest(reqBean *bean.DevtronResourceObjectBean) error {
	if reqBean.Overview == nil || len(reqBean.Overview.ReleaseVersion) == 0 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ReleaseVersionNotFound, bean.ReleaseVersionNotFound)
	} else if reqBean.ParentConfig == nil || reqBean.ParentConfig.Data == nil {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceParentConfigNotFound, bean.ResourceParentConfigNotFound)
	} else if reqBean.ParentConfig.Data.Id == 0 && len(reqBean.ParentConfig.Data.Name) == 0 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigData, bean.InvalidResourceParentConfigData)
	} else if reqBean.ParentConfig.Type != bean.DevtronResourceReleaseTrack {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidResourceParentConfigType, bean.InvalidResourceParentConfigType)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) populateDefaultValuesForCreateReleaseRequest(reqBean *bean.DevtronResourceObjectBean) error {
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

func (impl *DevtronResourceServiceImpl) updateUserProvidedDataInReleaseObj(objectData string, reqBean *bean.DevtronResourceObjectBean) (string, error) {
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
	releaseTrackConfig := impl.getParentConfigVariablesFromDependencies(object.ObjectData)
	return fmt.Sprintf("%s-%s", releaseTrackConfig.Data.Name, releaseVersion), nil
}
