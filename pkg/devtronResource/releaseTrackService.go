package devtronResource

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net/http"
	"time"
)

func (impl *DevtronResourceServiceImpl) updateReleaseTrackOverviewDataInResourceObj(resourceSchema *repository.DevtronResourceSchema,
	existingResourceObject *repository.DevtronResourceObject, resourceObject *bean.DevtronResourceObjectGetAPIBean) (err error) {
	//checking if resource object exists
	if existingResourceObject != nil && existingResourceObject.Id > 0 {
		//getting metadata out of this object
		resourceObject.OldObjectId, resourceObject.IdType = getResourceObjectIdAndType(existingResourceObject)
		resourceObject.Name = gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectNamePath).String()
		if gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectOverviewPath).Exists() {
			resourceObject.Overview = &bean.ResourceOverview{
				Description: gjson.Get(existingResourceObject.ObjectData, bean.ResourceObjectDescriptionPath).String(),
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
	return nil
}

func validateCreateReleaseTrackRequest(reqBean *bean.DevtronResourceObjectBean) error {
	if len(reqBean.Name) == 0 {
		return util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.ResourceNameNotFound, bean.ResourceNameNotFound)
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) populateDefaultValuesForCreateReleaseTrackRequest(reqBean *bean.DevtronResourceObjectBean) error {
	if reqBean.Overview != nil && reqBean.Overview.CreatedBy == nil {
		createdByDetails, err := impl.getUserSchemaDataById(reqBean.UserId)
		// considering the user details are already verified; this error indicates to an internal db error.
		if err != nil {
			impl.logger.Errorw("error encountered in populateDefaultValuesForCreateReleaseTrackRequest", "userId", reqBean.UserId, "err", err)
			return err
		}
		reqBean.Overview.CreatedBy = createdByDetails
		reqBean.Overview.CreatedOn = time.Now()
	}
	return nil
}

func (impl *DevtronResourceServiceImpl) updateUserProvidedDataInReleaseTrackObj(objectData string, reqBean *bean.DevtronResourceObjectBean) (string, error) {
	var err error
	if reqBean.Overview != nil {
		objectData, err = impl.setReleaseTrackOverviewFieldsInObjectData(objectData, reqBean.Overview)
		if err != nil {
			impl.logger.Errorw("error in setting overview data in schema", "err", err, "request", reqBean)
			return objectData, err
		}
	}
	return objectData, nil
}

func (impl *DevtronResourceServiceImpl) setReleaseTrackOverviewFieldsInObjectData(objectData string, overview *bean.ResourceOverview) (string, error) {
	var err error
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
	if overview.Tags != nil {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectTagsPath, overview.Tags)
		if err != nil {
			impl.logger.Errorw("error in setting description in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	return objectData, nil
}

func (impl *DevtronResourceServiceImpl) buildIdentifierForReleaseTrackResourceObj(object *repository.DevtronResourceObject) (string, error) {
	return gjson.Get(object.ObjectData, bean.ResourceObjectNamePath).String(), nil
}
