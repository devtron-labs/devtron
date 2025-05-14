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
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	bean4 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/helper"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/publish"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"helm.sh/helm/v3/pkg/chart"
)

func (impl *HandlerServiceImpl) getEnrichedWorkflowRunner(overrideRequest *bean3.ValuesOverrideRequest, artifact *repository3.CiArtifact, wfrId int) *pipelineConfig.CdWorkflowRunner {
	return nil
}

func (impl *HandlerServiceImpl) postDeployHook(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, referenceChartByte []byte, err error) {
	impl.logger.Debugw("no post deploy hook registered")
}

func (impl *HandlerServiceImpl) isDevtronAsyncArgoCdInstallModeEnabledForApp(appId, envId int, forceSync bool) (bool, error) {
	return impl.globalEnvVariables.EnableAsyncArgoCdInstallDevtronChart && !forceSync, nil
}

func (impl *HandlerServiceImpl) getClusterGRPCConfig(cluster repository2.Cluster) *gRPC.ClusterConfig {
	clusterConfig := helper.ConvertClusterBeanToGrpcConfig(cluster)
	return clusterConfig
}

func (impl *HandlerServiceImpl) overrideReferenceChartByteForHelmTypeApp(valuesOverrideResponse *app.ValuesOverrideResponse,
	chartMetaData *chart.Metadata, referenceTemplatePath string, referenceChartByte []byte) ([]byte, error) {
	return referenceChartByte, nil
}

func (impl *HandlerServiceImpl) getManifestPushService(storageType string) publish.ManifestPushService {
	var manifestPushService publish.ManifestPushService
	if storageType == bean2.ManifestStorageGit {
		manifestPushService = impl.gitOpsManifestPushService
	}
	return manifestPushService
}

func (impl *HandlerServiceImpl) preStageHandlingForTriggerStageInBulk(triggerRequest *bean.TriggerRequest) error {
	return nil
}

func (impl *HandlerServiceImpl) manifestGenerationFailedTimelineHandling(triggerEvent bean.TriggerEvent, overrideRequest *bean3.ValuesOverrideRequest, err error) {
}

func (impl *HandlerServiceImpl) getHelmManifestForTriggerRelease(ctx context.Context, triggerEvent bean.TriggerEvent, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string) ([]byte, error) {
	return nil, nil
}

func (impl *HandlerServiceImpl) buildManifestPushTemplateForNonGitStorageType(overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string, err error, manifestPushConfig *repository.ManifestPushConfig,
	manifestPushTemplate *bean4.ManifestPushTemplate) error {
	return nil
}

func (impl *HandlerServiceImpl) triggerReleaseSuccessHandling(triggerEvent bean.TriggerEvent, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, helmManifest []byte) error {
	return nil
}
