/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deployedApp

import (
	"context"
	bean6 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
)

func (impl *DeployedAppServiceImpl) getTemplate(stopRequest *bean.StopAppRequest) (string, error) {
	return "", nil
}
func (impl *DeployedAppServiceImpl) checkForFeasibilityBeforeStartStop(ctx context.Context, appId, envId int, userMetadata *bean6.UserMetadata) error {
	return nil
}

func (impl *DeployedAppServiceImpl) StopStartAppV1(ctx context.Context, stopRequest *bean.StopAppRequest, userMetadata *bean6.UserMetadata) (int, error) {
	return 0, nil
}

func (impl *DeployedAppServiceImpl) HibernationPatch(ctx context.Context, appId, envId int, userMetadata *bean6.UserMetadata) (*bean.HibernationPatchResponse, error) {
	return nil, nil
}
