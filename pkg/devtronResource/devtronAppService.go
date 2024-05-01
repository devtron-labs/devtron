package devtronResource

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
)

func (impl *DevtronResourceServiceImpl) getAppIdentifierByAppId(appId int) (string, error) {
	app, err := impl.appRepository.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error in finding app by app id", "err", err, "appId", appId)
		return "", err
	}
	return app.AppName, nil
}

func (impl *DevtronResourceServiceImpl) buildIdentifierForDevtronAppResourceObj(object *repository.DevtronResourceObject) (string, error) {
	return impl.getAppIdentifierByAppId(object.OldObjectId)
}

func (impl *DevtronResourceServiceImpl) getIdAndIdTypeFromIdentifierForDevtronApps(resourceIdentifier *bean.ResourceIdentifier) (int, bean.IdType, error) {
	if resourceIdentifier.Id > 0 {
		return resourceIdentifier.Id, bean.OldObjectId, nil
	} else {
		appId, err := impl.appRepository.FindIdByNameAndAppType(resourceIdentifier.Identifier, helper.CustomApp)
		if err != nil {
			impl.logger.Errorw("error in finding app by app name", "err", err, "appName", resourceIdentifier.Identifier)
			return 0, "", err
		}
		return appId, bean.OldObjectId, nil
	}
}

func (impl *DevtronResourceServiceImpl) getMapOfAppMetadata(appIdsToGetMetadata []int) (map[int]interface{}, map[int]string, error) {
	mapOfAppsMetadata := make(map[int]interface{})
	mapOfAppIdName := make(map[int]string)
	var apps []*app.App
	var err error
	if len(appIdsToGetMetadata) > 0 {
		apps, err = impl.appRepository.FindAppAndProjectByIdsIn(appIdsToGetMetadata)
		if err != nil {
			impl.logger.Errorw("error in getting apps by ids", "err", err, "ids", appIdsToGetMetadata)
			return nil, nil, err
		}
	}
	for _, app := range apps {
		mapOfAppsMetadata[app.Id] = &struct {
			AppName string `json:"appName"`
			AppId   int    `json:"appId"`
		}{
			AppName: app.AppName,
			AppId:   app.Id,
		}
		//TODO: merge below logic with v1 dep get
		mapOfAppIdName[app.Id] = app.AppName
	}
	return mapOfAppsMetadata, mapOfAppIdName, nil
}

func updateAppMetaDataInDependencyObj(oldObjectId int, metaDataObj *bean.DependencyMetaDataBean) interface{} {
	return metaDataObj.MapOfAppsMetadata[oldObjectId]
}
