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

package devtronApps

import (
	"context"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/fluxcd"
	"github.com/devtron-labs/devtron/pkg/app"
	"go.opentelemetry.io/otel"
	"time"
)

const (
	gitRepositoryReconcileInterval = 1 * time.Minute
	helmReleaseReconcileInterval   = 1 * time.Minute
)

func (impl *HandlerServiceImpl) deployFluxCdApp(ctx context.Context, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.deployFluxCdApp")
	defer span.End()
	clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(overrideRequest.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster", "clusterId", overrideRequest.ClusterId, "error", err)
		return err
	}
	req := &fluxcd.DeploymentRequest{
		ClusterConfig:    clusterConfig,
		DeploymentConfig: valuesOverrideResponse.DeploymentConfig,
		IsAppCreated:     valuesOverrideResponse.Pipeline != nil && valuesOverrideResponse.Pipeline.DeploymentAppCreated,
	}
	err = impl.fluxCdDeploymentService.DeployFluxCdApp(newCtx, req)
	if err != nil {
		impl.logger.Errorw("error in deploying FluxCdApp", "err", err)
		return err
	}
	if !req.IsAppCreated {
		//setting deploymentAppCreated = true
		_, err = impl.updatePipeline(valuesOverrideResponse.Pipeline, overrideRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in update cd pipeline for deployment app created or not", "err", err)
			return err
		}
	}
	return nil
}
