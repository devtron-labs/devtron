package devtronResource

import "github.com/devtron-labs/devtron/pkg/devtronResource/repository"

func (impl *DevtronResourceServiceImpl) buildIdentifierForHelmAppResourceObj(object *repository.DevtronResourceObject) (string, error) {
	return impl.getAppIdentifierByAppId(object.OldObjectId)
}
