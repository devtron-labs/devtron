package devtronResource

import "github.com/devtron-labs/devtron/pkg/devtronResource/repository"

func (impl *DevtronResourceServiceImpl) buildIdentifierFormHelmAppResourceObj(object *repository.DevtronResourceObject) (string, error) {
	return impl.getAppIdentifierByAppId(object.OldObjectId)
}
