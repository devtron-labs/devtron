package devtronApps

import (
	"context"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	bean6 "github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean4 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/helper"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
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

func (impl *TriggerServiceImpl) getHelmHistoryLimitAndChartMetadataForHelmAppCreation(ctx context.Context,
	valuesOverrideResponse *app.ValuesOverrideResponse) (*chart.Metadata, int32, *gRPC.ReleaseIdentifier, error) {
	pipelineModel := valuesOverrideResponse.Pipeline
	envOverride := valuesOverrideResponse.EnvOverride

	var chartMetaData *chart.Metadata
	releaseName := pipelineModel.DeploymentAppName
	//getting cluster by id
	cluster, err := impl.clusterRepository.FindById(envOverride.Environment.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by id", "clusterId", envOverride.Environment.ClusterId, "err", err)
		return nil, 0, nil, err
	} else if cluster == nil {
		impl.logger.Errorw("error in getting cluster by id, found nil object", "clusterId", envOverride.Environment.ClusterId)
		return nil, 0, nil, err
	}

	clusterConfig := helper.ConvertClusterBeanToGrpcConfig(*cluster)

	releaseIdentifier := &gRPC.ReleaseIdentifier{
		ReleaseName:      releaseName,
		ReleaseNamespace: envOverride.Namespace,
		ClusterConfig:    clusterConfig,
	}

	var helmRevisionHistory int32
	if valuesOverrideResponse.DeploymentConfig.ReleaseMode == util.PIPELINE_RELEASE_MODE_LINK {
		detail, err := impl.helmAppClient.GetReleaseDetails(ctx, releaseIdentifier)
		if err != nil {
			impl.logger.Errorw("error in fetching release details", "clusterId", clusterConfig.ClusterId, "namespace", envOverride.Namespace, "releaseName", releaseName, "err", err)
			return nil, 0, nil, err
		}
		chartMetaData = &chart.Metadata{
			Name:    detail.ChartName,
			Version: detail.ChartVersion,
		}
		//not modifying revision history in case of linked release
		helmRevisionHistory = impl.helmAppService.GetRevisionHistoryMaxValue(bean6.SOURCE_LINKED_HELM_APP)
	} else {
		chartMetaData = &chart.Metadata{
			Name:    pipelineModel.App.AppName,
			Version: envOverride.Chart.ChartVersion,
		}
		helmRevisionHistory = impl.helmAppService.GetRevisionHistoryMaxValue(bean6.SOURCE_DEVTRON_APP)
	}

	return chartMetaData, helmRevisionHistory, releaseIdentifier, nil
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
