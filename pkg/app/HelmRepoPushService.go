package app

import (
	"context"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	status2 "github.com/devtron-labs/devtron/pkg/app/status"
	"go.uber.org/zap"
	"path"
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
		if ociHelmRepoPushRequest.RegistryCredential != nil {
			repoURL := path.Join(ociHelmRepoPushRequest.RegistryCredential.RegistryUrl, ociHelmRepoPushRequest.RegistryCredential.RepoName)
			manifestPushResponse.Error = fmt.Errorf("Could not push helm package to \"%v\"", repoURL)
			timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(manifestPushTemplate.WorkflowRunnerId, 0, pipelineConfig.TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO_FAILED, fmt.Sprintf("Could not push helm package to \"%v\"", repoURL), manifestPushTemplate.UserId, time.Now())
			timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
			if timelineErr != nil {
				impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
			}
		} else {
			manifestPushResponse.Error = err
			timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(manifestPushTemplate.WorkflowRunnerId, 0, pipelineConfig.TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO_FAILED, err.Error(), manifestPushTemplate.UserId, time.Now())
			timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
			if timelineErr != nil {
				impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
			}
		}
		return manifestPushResponse
	}
	manifestPushResponse.CommitHash = helmManifestResponse.PushResult.Digest
	manifestPushResponse.CommitTime = time.Now()
	timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(manifestPushTemplate.WorkflowRunnerId, 0, pipelineConfig.TIMELINE_STATUS_MANIFEST_PUSHED_TO_HELM_REPO, "helm packaged successfully pushed to helm repo", manifestPushTemplate.UserId, time.Now())
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
		RegistryCredential: &client.RegistryCredential{
			Username:     manifestPushTemplate.ContainerRegistryConfig.Username,
			Password:     manifestPushTemplate.ContainerRegistryConfig.Password,
			AwsRegion:    manifestPushTemplate.ContainerRegistryConfig.AwsRegion,
			AccessKey:    manifestPushTemplate.ContainerRegistryConfig.AccessKey,
			SecretKey:    manifestPushTemplate.ContainerRegistryConfig.SecretKey,
			RegistryUrl:  manifestPushTemplate.ContainerRegistryConfig.RegistryUrl,
			RegistryType: manifestPushTemplate.ContainerRegistryConfig.RegistryType,
			IsPublic:     manifestPushTemplate.ContainerRegistryConfig.IsPublic,
			RepoName:     manifestPushTemplate.ContainerRegistryConfig.RepoName,
		},
	}
}
