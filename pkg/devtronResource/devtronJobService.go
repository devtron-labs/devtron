package devtronResource

import "github.com/devtron-labs/devtron/pkg/devtronResource/repository"

func (impl *DevtronResourceServiceImpl) getJobIdentifierByAppId(appId int) (string, error) {
	app, err := impl.appRepository.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error in finding app by app id", "err", err, "appId", appId)
		return "", err
	}
	return app.DisplayName, nil
}

func (impl *DevtronResourceServiceImpl) buildIdentifierForDevtronJobResourceObj(object *repository.DevtronResourceObject) (string, error) {
	return impl.getJobIdentifierByAppId(object.OldObjectId)
}
