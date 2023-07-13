package app

import (
	"context"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	status2 "github.com/devtron-labs/devtron/pkg/app/status"
	"go.uber.org/zap"
	"time"
)

type HelmRepoPushService interface {
	ManifestPushService
}

type HelmRepoPushServiceImpl struct {
	logger                        *zap.SugaredLogger
	helmAppClient                 client.HelmAppClient
	pipelineStatusTimelineService status2.PipelineStatusTimelineService
}

func NewHelmRepoPushServiceImpl(
	logger *zap.SugaredLogger,
	helmAppClient client.HelmAppClient,
	pipelineStatusTimelineService status2.PipelineStatusTimelineService,
) *HelmRepoPushServiceImpl {
	return &HelmRepoPushServiceImpl{
		logger:                        logger,
		helmAppClient:                 helmAppClient,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
	}
}

func (impl *HelmRepoPushServiceImpl) PushChart(manifestPushTemplate *bean.ManifestPushTemplate, ctx context.Context) bean.ManifestPushResponse {

	var manifestPushResponse bean.ManifestPushResponse
	ociHelmRepoPushRequest := getOciPushTemplate(manifestPushTemplate)
	helmManifestResponse, err := impl.helmAppClient.PushHelmChartToOCIRegistry(ctx, ociHelmRepoPushRequest)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to helm oci repo", "err", err)
		manifestPushResponse.Error = fmt.Errorf("Could not push helm package to \"%v\"", ociHelmRepoPushRequest.RepoURL)
		timeline := getTimelineObject(manifestPushTemplate, pipelineConfig.TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO_FAILED, fmt.Sprintf("Could not push helm package to- \"%v\"", ociHelmRepoPushRequest.RepoURL))
		timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
		if timelineErr != nil {
			impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
		}

		return manifestPushResponse
	}
	manifestPushResponse.CommitHash = helmManifestResponse.Result.Digest
	manifestPushResponse.CommitTime = time.Now()

	timeline := getTimelineObject(manifestPushTemplate, pipelineConfig.TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO, "helm packaged successfully pushed to helm repo")
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
	if timelineErr != nil {
		impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
	}

	return manifestPushResponse

}

func getOciPushTemplate(manifestPushTemplate *bean.ManifestPushTemplate) *client.OCIRegistryRequest {
	return &client.OCIRegistryRequest{
		Chart:        *manifestPushTemplate.BuiltChartBytes,
		ChartName:    manifestPushTemplate.ChartName,
		ChartVersion: manifestPushTemplate.ChartVersion,
		IsInsecure:   true,
		Username:     manifestPushTemplate.ContainerRegistryConfig.Username,
		Password:     manifestPushTemplate.ContainerRegistryConfig.Password,
		AwsRegion:    manifestPushTemplate.ContainerRegistryConfig.AwsRegion,
		AccessKey:    manifestPushTemplate.ContainerRegistryConfig.AccessKey,
		SecretKey:    manifestPushTemplate.ContainerRegistryConfig.SecretKey,
		RegistryURL:  manifestPushTemplate.ContainerRegistryConfig.RegistryUrl,
		RepoURL:      manifestPushTemplate.RepoUrl,
	}
}
