package app

import (
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	. "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	status2 "github.com/devtron-labs/devtron/pkg/app/status"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/gitops"
	"github.com/devtron-labs/devtron/util/ChartsUtil"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type ManifestPushService interface {
	PushChart(manifestPushConfig *bean.ManifestPushTemplate, ctx context.Context) bean.ManifestPushResponse
}

type GitOpsPushService interface {
	ManifestPushService
}

type GitOpsManifestPushServiceImpl struct {
	logger                           *zap.SugaredLogger
	chartTemplateService             util.ChartTemplateService
	chartDeploymentService        util.ChartDeploymentService
	chartService                     chartService.ChartService
	gitOpsConfigRepository           repository.GitOpsConfigRepository
	gitFactory                       *GitFactory
	pipelineStatusTimelineService    status2.PipelineStatusTimelineService
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
	acdConfig                        *argocdServer.ACDConfig
	gitOpsConfigService           gitops.GitOpsConfigService
}

func NewGitOpsManifestPushServiceImpl(
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	chartDeploymentService util.ChartDeploymentService,
	chartService chartService.ChartService,
	gitOpsConfigRepository repository.GitOpsConfigRepository,
	gitFactory *GitFactory,
	pipelineStatusTimelineService status2.PipelineStatusTimelineService,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	acdConfig *argocdServer.ACDConfig,
	gitOpsConfigService gitops.GitOpsConfigService,
) *GitOpsManifestPushServiceImpl {
	return &GitOpsManifestPushServiceImpl{
		logger:                           logger,
		chartTemplateService:             chartTemplateService,
		chartDeploymentService:        chartDeploymentService,
		chartService:                     chartService,
		gitOpsConfigRepository:           gitOpsConfigRepository,
		gitFactory:                       gitFactory,
		pipelineStatusTimelineService:    pipelineStatusTimelineService,
		pipelineStatusTimelineRepository: pipelineStatusTimelineRepository,
		acdConfig:                        acdConfig,
		gitOpsConfigService:           gitOpsConfigService,
	}
}

func (impl *GitOpsManifestPushServiceImpl) PushChart(manifestPushTemplate *bean.ManifestPushTemplate, ctx context.Context) bean.ManifestPushResponse {
	manifestPushResponse := bean.ManifestPushResponse{}
	activeGlobalGitOpsConfig, err := impl.gitOpsConfigService.GetGitOpsConfigActive()
	if err != nil {
		impl.logger.Errorw("error in fetching active gitOps config", "err", err)
		if util.IsErrNoRows(err) {
			errMsg := fmt.Errorf("Gitops integration is not installed/configured. Please install/configure gitops.")
			manifestPushResponse.Error = errMsg
			impl.SaveTimelineForError(manifestPushTemplate, errMsg)
			return manifestPushResponse
		}
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}

	if ChartsUtil.IsGitOpsRepoNotConfigured(manifestPushTemplate.RepoUrl) || manifestPushTemplate.GitOpsRepoMigrationRequired {
		if activeGlobalGitOpsConfig.AllowCustomRepository || manifestPushTemplate.IsCustomGitRepository {
			errMsg := fmt.Errorf("GitOps repository is not configured! Please configure gitops repository for application first.")
			manifestPushResponse.Error = errMsg
			impl.SaveTimelineForError(manifestPushTemplate, errMsg)
			return manifestPushResponse
		}

		gitOpsRepoName := impl.chartTemplateService.GetGitOpsRepoName(manifestPushTemplate.AppName)
		chartGitAttr, err := impl.chartTemplateService.CreateGitRepositoryForApp(gitOpsRepoName, manifestPushTemplate.UserId)
		if err != nil {
			impl.logger.Errorw("error in pushing chart to git ", "gitOpsRepoName", gitOpsRepoName, "err", err)
			errMsg := fmt.Errorf("No repository configured for Gitops! Error while creating git repository: '%s'", gitOpsRepoName)
			manifestPushResponse.Error = errMsg
			impl.SaveTimelineForError(manifestPushTemplate, errMsg)
			return manifestPushResponse
		}
		err = impl.chartDeploymentService.RegisterInArgo(chartGitAttr, manifestPushTemplate.UserId, ctx, false)
		if err != nil {
			impl.logger.Errorw("error in registering app in acd", "err", err)
			errMsg := fmt.Errorf("Error in registering repository '%s' in ArgoCd", gitOpsRepoName)
			manifestPushResponse.Error = errMsg
			impl.SaveTimelineForError(manifestPushTemplate, errMsg)
			return manifestPushResponse
		}
		manifestPushTemplate.RepoUrl = chartGitAttr.RepoUrl
		chartGitAttr.ChartLocation = manifestPushTemplate.ChartLocation
		err = impl.chartService.UpdateGitRepoUrlInCharts(manifestPushTemplate.AppId, chartGitAttr, manifestPushTemplate.UserId)
		if err != nil {
			impl.logger.Errorw("error in updating git repo url in charts", "err", err)
			errMsg := fmt.Errorf("No repository configured for Gitops! Error while creating git repository: '%s'", gitOpsRepoName)
			manifestPushResponse.Error = errMsg
			impl.SaveTimelineForError(manifestPushTemplate, errMsg)
			return manifestPushResponse
		}
	}

	validateRequest := gitops.ValidateCustomGitRepoURLRequest{
		GitRepoURL: manifestPushTemplate.RepoUrl,
	}
	detailedErrorGitOpsConfigResponse := impl.gitOpsConfigService.ValidateCustomGitRepoURL(validateRequest)
	if len(detailedErrorGitOpsConfigResponse.StageErrorMap) != 0 {
		errMsg := impl.gitOpsConfigService.ExtractErrorsFromGitOpsConfigResponse(detailedErrorGitOpsConfigResponse)
		impl.logger.Errorw("invalid gitOps repo URL", "err", errMsg)
		manifestPushResponse.Error = errMsg //as the pipeline_status_timeline.status_detail is of type TEXT; the length of errMsg won't be an issue
		impl.SaveTimelineForError(manifestPushTemplate, errMsg)
		return manifestPushResponse
	}

	err = impl.PushChartToGitRepo(manifestPushTemplate, ctx)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git", "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}
	commitHash, commitTime, err := impl.CommitValuesToGit(manifestPushTemplate, ctx)
	if err != nil {
		impl.logger.Errorw("error in committing values to git", "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}
	manifestPushResponse.CommitHash = commitHash
	manifestPushResponse.CommitTime = commitTime

	dbConnection := impl.pipelineStatusTimelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in transaction begin in saving gitops timeline", "err", err)
		manifestPushResponse.Error = err
		return manifestPushResponse
	}

	gitCommitTimeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(manifestPushTemplate.WorkflowRunnerId, 0, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", manifestPushTemplate.UserId, time.Now())

	timelines := []*pipelineConfig.PipelineStatusTimeline{gitCommitTimeline}
	if !impl.acdConfig.ArgoCDAutoSyncEnabled {
		// if manual sync is enabled, add ARGOCD_SYNC_INITIATED_TIMELINE
		argoCDSyncInitiatedTimeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(manifestPushTemplate.WorkflowRunnerId, 0, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED, "argocd sync initiated.", manifestPushTemplate.UserId, time.Now())
		timelines = append(timelines, argoCDSyncInitiatedTimeline)
	}
	timelineErr := impl.pipelineStatusTimelineService.SaveTimelines(timelines, tx)
	if timelineErr != nil {
		impl.logger.Errorw("Error in saving git commit success timeline", err, timelineErr)
	}
	tx.Commit()

	return manifestPushResponse
}

func (impl *GitOpsManifestPushServiceImpl) PushChartToGitRepo(manifestPushTemplate *bean.ManifestPushTemplate, ctx context.Context) error {

	_, span := otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.GetGitOpsRepoName")
	// CHART COMMIT and PUSH STARTS HERE, it will push latest version, if found modified on deployment template and overrides
	gitOpsRepoName := util.GetGitRepoNameFromGitRepoUrl(manifestPushTemplate.RepoUrl)
	span.End()
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartService.CheckChartExists")
	err := impl.chartService.CheckChartExists(manifestPushTemplate.ChartRefId)
	span.End()
	if err != nil {
		impl.logger.Errorw("err in getting chart info", "err", err)
		return err
	}
	err = impl.chartTemplateService.PushChartToGitRepo(gitOpsRepoName, manifestPushTemplate.ChartReferenceTemplate, manifestPushTemplate.ChartVersion, manifestPushTemplate.BuiltChartPath, manifestPushTemplate.RepoUrl, manifestPushTemplate.UserId)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git", "err", err)
		return err
	}
	return nil
}

func (impl *GitOpsManifestPushServiceImpl) CommitValuesToGit(manifestPushTemplate *bean.ManifestPushTemplate, ctx context.Context) (commitHash string, commitTime time.Time, err error) {
	commitHash = ""
	commitTime = time.Time{}
	chartRepoName := util.GetGitRepoNameFromGitRepoUrl(manifestPushTemplate.RepoUrl)
	_, span := otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit")
	//getting username & emailId for commit author data
	userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(manifestPushTemplate.UserId)
	span.End()
	chartGitAttr := &util.ChartConfig{
		FileName:       fmt.Sprintf("_%d-values.yaml", manifestPushTemplate.TargetEnvironmentName),
		FileContent:    string(manifestPushTemplate.MergedValues),
		ChartName:      manifestPushTemplate.ChartName,
		ChartLocation:  manifestPushTemplate.ChartLocation,
		ChartRepoName:  chartRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", manifestPushTemplate.PipelineOverrideId, manifestPushTemplate.TargetEnvironmentName),
		UserName:       userName,
		UserEmailId:    userEmailId,
	}
	gitOpsConfigBitbucket, err := impl.gitOpsConfigRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			return commitHash, commitTime, err
		}
	}
	gitOpsConfig := &bean2.GitOpsConfigDto{BitBucketWorkspaceId: gitOpsConfigBitbucket.BitBucketWorkspaceId}
	_, span = otel.Tracer("orchestrator").Start(ctx, "gitFactory.Client.CommitValues")
	commitHash, commitTime, err = impl.gitFactory.Client.CommitValues(chartGitAttr, gitOpsConfig)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return commitHash, commitTime, err
	}
	if commitTime.IsZero() {
		commitTime = time.Now()
	}
	span.End()
	if err != nil {
		return commitHash, commitTime, err
	}
	return commitHash, commitTime, nil
}

func (impl *GitOpsManifestPushServiceImpl) SaveTimelineForError(manifestPushTemplate *bean.ManifestPushTemplate, gitCommitErr error) {
	timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(manifestPushTemplate.WorkflowRunnerId, 0, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", gitCommitErr), manifestPushTemplate.UserId, time.Now())
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
	if timelineErr != nil {
		impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
	}
}
