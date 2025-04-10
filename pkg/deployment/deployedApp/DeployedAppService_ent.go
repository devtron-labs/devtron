package deployedApp

import (
	"context"
	bean6 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
)

func (impl *DeployedAppServiceImpl) getTemplate(stopRequest *bean.StopAppRequest) (string, error) {
	return "", nil
}
func (impl *DeployedAppServiceImpl) checkForFeasibilityBeforeStartStop(appId, envId int, userMetadata *bean6.UserMetadata) error {
	return nil
}

func (impl *DeployedAppServiceImpl) StopStartAppV1(ctx context.Context, stopRequest *bean.StopAppRequest, userMetadata *bean6.UserMetadata) (int, error) {
	return 0, nil
}

func (impl *DeployedAppServiceImpl) HibernationPatch(ctx context.Context, appId, envId int, userMetadata *bean6.UserMetadata) (*bean.HibernationPatchResponse, error) {
	return nil, nil
}
