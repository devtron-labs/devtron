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

package resourceTree

import (
	"context"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func (impl *ServiceImpl) FetchResourceTreeWithDrift(ctx context.Context, appId int, envId int, cdPipeline *pipelineConfig.Pipeline,
	deploymentConfig *commonBean.DeploymentConfig) (map[string]interface{}, error) {
	return nil, nil
}
