package devtronApps

import (
	"context"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	bean4 "github.com/devtron-labs/devtron/pkg/app/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/coreEntities/argoApplication/helper"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/publish"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"helm.sh/helm/v3/pkg/chart"
)

func (impl *TriggerServiceImpl) getEnrichedWorkflowRunner(overrideRequest *bean3.ValuesOverrideRequest, artifact *repository3.CiArtifact, wfrId int) *pipelineConfig.CdWorkflowRunner {
	return nil
}

func (impl *TriggerServiceImpl) postDeployHook(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, referenceChartByte []byte, err error) {
	impl.logger.Debugw("no post deploy hook registered")
}

func (impl *TriggerServiceImpl) isDevtronAsyncArgoCdInstallModeEnabledForApp(appId, envId int, forceSync bool) (bool, error) {
	return impl.globalEnvVariables.EnableAsyncArgoCdInstallDevtronChart && !forceSync, nil
}

func (impl *TriggerServiceImpl) getClusterGRPCConfig(cluster repository2.Cluster) *gRPC.ClusterConfig {
	clusterConfig := helper.ConvertClusterBeanToGrpcConfig(cluster)
	return clusterConfig
}

func (impl *TriggerServiceImpl) overrideReferenceChartByteForHelmTypeApp(valuesOverrideResponse *app.ValuesOverrideResponse,
	chartMetaData *chart.Metadata, referenceTemplatePath string, referenceChartByte []byte) ([]byte, error) {
	return referenceChartByte, nil
}

func (impl *TriggerServiceImpl) getManifestPushService(storageType string) publish.ManifestPushService {
	var manifestPushService publish.ManifestPushService
	if storageType == bean2.ManifestStorageGit {
		manifestPushService = impl.gitOpsManifestPushService
	}
	return manifestPushService
}

func (impl *TriggerServiceImpl) preStageHandlingForTriggerStageInBulk(triggerRequest *bean.TriggerRequest) error {
	return nil
}

func (impl *TriggerServiceImpl) manifestGenerationFailedTimelineHandling(triggerEvent bean.TriggerEvent, overrideRequest *bean3.ValuesOverrideRequest, err error) {
}

func (impl *TriggerServiceImpl) getHelmManifestForTriggerRelease(ctx context.Context, triggerEvent bean.TriggerEvent, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string) ([]byte, error) {
	return nil, nil
}

func (impl *TriggerServiceImpl) buildManifestPushTemplateForNonGitStorageType(overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string, err error, manifestPushConfig *repository.ManifestPushConfig,
	manifestPushTemplate *bean4.ManifestPushTemplate) error {
	return nil
}

func (impl *TriggerServiceImpl) triggerReleaseSuccessHandling(triggerEvent bean.TriggerEvent, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, helmManifest []byte) error {
	return nil
}
