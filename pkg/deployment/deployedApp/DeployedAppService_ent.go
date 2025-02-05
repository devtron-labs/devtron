package deployedApp

import (
	"context"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
)

func (impl *DeployedAppServiceImpl) getTemplate(stopRequest *bean.StopAppRequest) (string, error) {
	return "", nil
}
func (impl *DeployedAppServiceImpl) checkForFeasibilityBeforeStartStop(appId, envId int, userId int32) error {
	return nil
}

func (impl *DeployedAppServiceImpl) StopStartAppV1(ctx context.Context, stopRequest *bean.StopAppRequest) (int, error) {
	return 0, nil
}

func (impl *DeployedAppServiceImpl) HibernationPatch(ctx context.Context, appId, envId int) (*bean.HibernationPatchResponse, error) {
	return nil, nil
}
