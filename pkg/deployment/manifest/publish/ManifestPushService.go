/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package publish

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/app/status"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	gitOpsBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/config/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type ManifestPushService interface {
	PushChart(ctx context.Context, manifestPushTemplate *bean.ManifestPushTemplate) bean.ManifestPushResponse
}

type GitOpsPushService interface {
	ManifestPushService
}

type GitOpsManifestPushServiceImpl struct {
	logger                        *zap.SugaredLogger
	pipelineStatusTimelineService status.PipelineStatusTimelineService
	pipelineOverrideRepository    chartConfig.PipelineOverrideRepository
	acdConfig                     *argocdServer.ACDConfig
	chartRefService               chartRef.ChartRefService
	gitOpsConfigReadService       config.GitOpsConfigReadService
	chartService                  chartService.ChartService
	gitOperationService           git.GitOperationService
	argoClientWrapperService      argocdServer.ArgoClientWrapperService
	*sql.TransactionUtilImpl
}

func NewGitOpsManifestPushServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	acdConfig *argocdServer.ACDConfig,
	chartRefService chartRef.ChartRefService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	chartService chartService.ChartService,
	gitOperationService git.GitOperationService,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	transactionUtilImpl *sql.TransactionUtilImpl) *GitOpsManifestPushServiceImpl {
	return &GitOpsManifestPushServiceImpl{
		logger:                        logger,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		pipelineOverrideRepository:    pipelineOverrideRepository,
		acdConfig:                     acdConfig,
		chartRefService:               chartRefService,
		gitOpsConfigReadService:       gitOpsConfigReadService,
		chartService:                  chartService,
		gitOperationService:           gitOperationService,
		argoClientWrapperService:      argoClientWrapperService,
		TransactionUtilImpl:           transactionUtilImpl,
	}
}

func (impl *GitOpsManifestPushServiceImpl) createRepoForGitOperation(manifestPushTemplate bean.ManifestPushTemplate, ctx context.Context) (string, error) {
	// custom GitOps repo url doesn't support migration
	if manifestPushTemplate.IsCustomGitRepository {
		return manifestPushTemplate.RepoUrl, nil
	}
	gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(manifestPushTemplate.AppName)
	chartGitAttr, err := impl.gitOperationService.CreateGitRepositoryForDevtronApp(ctx, gitOpsRepoName, manifestPushTemplate.UserId)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git ", "gitOpsRepoName", gitOpsRepoName, "err", err)
		return "", fmt.Errorf("No repository configured for Gitops! Error while creating git repository: '%s'", gitOpsRepoName)
	}
	chartGitAttr.ChartLocation = manifestPushTemplate.ChartLocation
	err = impl.argoClientWrapperService.RegisterGitOpsRepoInArgoWithRetry(ctx, chartGitAttr.RepoUrl, manifestPushTemplate.UserId)
	if err != nil {
		impl.logger.Errorw("error in registering app in acd", "err", err)
		return "", fmt.Errorf("Error in registering repository '%s' in ArgoCd", gitOpsRepoName)
	}
	return chartGitAttr.RepoUrl, nil
}

func (impl *GitOpsManifestPushServiceImpl) validateManifestPushRequest(globalGitOpsConfigStatus gitOpsBean.GitOpsConfigurationStatus, manifestPushTemplate bean.ManifestPushTemplate) error {
	if !globalGitOpsConfigStatus.IsGitOpsConfigured {
		return fmt.Errorf("Gitops integration is not installed/configured. Please install/configure gitops.")
	}
	if gitOps.IsGitOpsRepoNotConfigured(manifestPushTemplate.RepoUrl) {
		if globalGitOpsConfigStatus.AllowCustomRepository {
			return fmt.Errorf("GitOps repository is not configured! Please configure gitops repository for application first.")
		}
	}
	return nil
}

func (impl *GitOpsManifestPushServiceImpl) PushChart(ctx context.Context, manifestPushTemplate *bean.ManifestPushTemplate) bean.ManifestPushResponse {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "GitOpsManifestPushServiceImpl.PushChart")
	defer span.End()
	manifestPushResponse := bean.ManifestPushResponse{}
	// 1. Fetch Global GitOps Details
	globalGitOpsConfigStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error in fetching active gitOps config", "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}
	// 2. Validate Repository for Git Operation
	errMsg := impl.validateManifestPushRequest(*globalGitOpsConfigStatus, *manifestPushTemplate)
	if errMsg != nil {
		manifestPushResponse.Error = errMsg
		impl.SaveTimelineForError(manifestPushTemplate, errMsg)
		return manifestPushResponse
	}
	// 3. Create Git Repo if required
	if gitOps.IsGitOpsRepoNotConfigured(manifestPushTemplate.RepoUrl) {
		newGitRepoUrl, errMsg := impl.createRepoForGitOperation(*manifestPushTemplate, newCtx)
		if errMsg != nil {
			manifestPushResponse.Error = errMsg
			impl.SaveTimelineForError(manifestPushTemplate, errMsg)
			return manifestPushResponse
		}
		manifestPushTemplate.RepoUrl = newGitRepoUrl
		manifestPushResponse.NewGitRepoUrl = newGitRepoUrl
		// below function will override gitRepoUrl for charts even if user has already configured gitOps repoURL
		err = impl.chartService.OverrideGitOpsRepoUrl(manifestPushTemplate.AppId, newGitRepoUrl, manifestPushTemplate.UserId)
		if err != nil {
			impl.logger.Errorw("error in updating git repo url in charts", "err", err)
			manifestPushResponse.Error = fmt.Errorf("No repository configured for Gitops! Error while migrating gitops repository: '%s'", newGitRepoUrl)
			impl.SaveTimelineForError(manifestPushTemplate, manifestPushResponse.Error)
			return manifestPushResponse
		}

	}
	// 4. Push Chart to Git Repository
	err = impl.pushChartToGitRepo(newCtx, manifestPushTemplate)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git", "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}

	// 5. Commit chart values to Git Repository
	commitHash, commitTime, err := impl.commitValuesToGit(newCtx, manifestPushTemplate)
	if err != nil {
		impl.logger.Errorw("error in committing values to git", "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}
	manifestPushResponse.CommitHash = commitHash
	manifestPushResponse.CommitTime = commitTime
	// 6. Update commit details in PipelineConfigOverride and Deployment Status Timelines
	tx, err := impl.TransactionUtilImpl.StartTx()
	defer impl.TransactionUtilImpl.RollbackTx(tx)
	if err != nil {
		impl.logger.Errorw("error in transaction begin in saving gitops timeline", "err", err)
		manifestPushResponse.Error = err
		return manifestPushResponse
	}
	err = impl.pipelineOverrideRepository.UpdateCommitDetails(newCtx, tx, manifestPushTemplate.PipelineOverrideId, manifestPushResponse.CommitHash, manifestPushResponse.CommitTime, manifestPushTemplate.UserId)
	if err != nil {
		impl.logger.Errorw("error in updating commit details to PipelineConfigOverride", "pipelineOverrideId", manifestPushTemplate.PipelineOverrideId, "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}
	gitCommitTimeline := impl.pipelineStatusTimelineService.NewDevtronAppPipelineStatusTimelineDbObject(manifestPushTemplate.WorkflowRunnerId, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, pipelineConfig.TIMELINE_DESCRIPTION_ARGOCD_GIT_COMMIT, manifestPushTemplate.UserId)
	timelines := []*pipelineConfig.PipelineStatusTimeline{gitCommitTimeline}
	if impl.acdConfig.IsManualSyncEnabled() {
		// if manual sync is enabled, add ARGOCD_SYNC_INITIATED_TIMELINE
		argoCDSyncInitiatedTimeline := impl.pipelineStatusTimelineService.NewDevtronAppPipelineStatusTimelineDbObject(manifestPushTemplate.WorkflowRunnerId, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED, pipelineConfig.TIMELINE_DESCRIPTION_ARGOCD_SYNC_INITIATED, manifestPushTemplate.UserId)
		timelines = append(timelines, argoCDSyncInitiatedTimeline)
	}
	timelineErr := impl.pipelineStatusTimelineService.SaveTimelinesIfNotAlreadyPresent(timelines, tx)
	if timelineErr != nil {
		impl.logger.Errorw("Error in saving git commit success timeline", err, timelineErr)
	}
	err = impl.TransactionUtilImpl.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to save gitops timeline", "err", err)
		manifestPushResponse.Error = err
		return manifestPushResponse
	}
	return manifestPushResponse
}

func (impl *GitOpsManifestPushServiceImpl) pushChartToGitRepo(ctx context.Context, manifestPushTemplate *bean.ManifestPushTemplate) error {

	_, span := otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.getGitOpsRepoName")
	// CHART COMMIT and PUSH STARTS HERE, it will push latest version, if found modified on deployment template and overrides
	gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(manifestPushTemplate.RepoUrl)
	span.End()
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartService.CheckChartExists")
	err := impl.chartRefService.CheckChartExists(manifestPushTemplate.ChartRefId)
	span.End()
	if err != nil {
		impl.logger.Errorw("err in getting chart info", "err", err)
		return err
	}
	err = impl.gitOperationService.PushChartToGitRepo(ctx, gitOpsRepoName, manifestPushTemplate.ChartReferenceTemplate, manifestPushTemplate.ChartVersion, manifestPushTemplate.BuiltChartPath, manifestPushTemplate.RepoUrl, manifestPushTemplate.UserId)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git", "err", err)
		return err
	}
	return nil
}

func (impl *GitOpsManifestPushServiceImpl) commitValuesToGit(ctx context.Context, manifestPushTemplate *bean.ManifestPushTemplate) (commitHash string, commitTime time.Time, err error) {
	commitHash = ""
	commitTime = time.Time{}
	chartRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(manifestPushTemplate.RepoUrl)
	_, span := otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit")
	//getting username & emailId for commit author data
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(manifestPushTemplate.UserId)
	span.End()
	chartGitAttr := &git.ChartConfig{
		FileName:       fmt.Sprintf("_%d-values.yaml", manifestPushTemplate.TargetEnvironmentName),
		FileContent:    manifestPushTemplate.MergedValues,
		ChartName:      manifestPushTemplate.ChartName,
		ChartLocation:  manifestPushTemplate.ChartLocation,
		ChartRepoName:  chartRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", manifestPushTemplate.PipelineOverrideId, manifestPushTemplate.TargetEnvironmentName),
		UserName:       userName,
		UserEmailId:    userEmailId,
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "gitOperationService.CommitValues")
	commitHash, commitTime, err = impl.gitOperationService.CommitValues(ctx, chartGitAttr)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return commitHash, commitTime, err
	}
	return commitHash, commitTime, nil
}

func (impl *GitOpsManifestPushServiceImpl) SaveTimelineForError(manifestPushTemplate *bean.ManifestPushTemplate, gitCommitErr error) {
	timeline := impl.pipelineStatusTimelineService.NewDevtronAppPipelineStatusTimelineDbObject(manifestPushTemplate.WorkflowRunnerId, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", gitCommitErr), manifestPushTemplate.UserId)
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
	if timelineErr != nil {
		impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
	}
}
