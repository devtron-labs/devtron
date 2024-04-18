package devtronResource

import "github.com/devtron-labs/devtron/pkg/devtronResource/repository"

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
