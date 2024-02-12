package app

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	status2 "github.com/devtron-labs/devtron/pkg/app/status"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
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
	pipelineStatusTimelineService    status2.PipelineStatusTimelineService
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
	acdConfig                        *argocdServer.ACDConfig
	chartRefService                  chartRef.ChartRefService
	chartService                     chartService.ChartService
	gitOpsConfigReadService          config.GitOpsConfigReadService
	gitOperationService              git.GitOperationService
	argoClientWrapperService         argocdServer.ArgoClientWrapperService
}

func NewGitOpsManifestPushServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService status2.PipelineStatusTimelineService,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	acdConfig *argocdServer.ACDConfig,
	chartRefService chartRef.ChartRefService,
	chartService chartService.ChartService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOperationService git.GitOperationService,
	argoClientWrapperService argocdServer.ArgoClientWrapperService) *GitOpsManifestPushServiceImpl {
	return &GitOpsManifestPushServiceImpl{
		logger:                           logger,
		pipelineStatusTimelineService:    pipelineStatusTimelineService,
		pipelineStatusTimelineRepository: pipelineStatusTimelineRepository,
		acdConfig:                        acdConfig,
		chartRefService:                  chartRefService,
		chartService:                     chartService,
		gitOpsConfigReadService:          gitOpsConfigReadService,
		gitOperationService:              gitOperationService,
		argoClientWrapperService:         argoClientWrapperService,
	}
}

func (impl *GitOpsManifestPushServiceImpl) ValidateRepoForGitOperation(manifestPushTemplate *bean.ManifestPushTemplate, ctx context.Context) error {
	// 1. Fetch Global GitOps Details
	activeGlobalGitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsConfigActive()
	if err != nil {
		impl.logger.Errorw("error in fetching active gitOps config", "err", err)
		if util.IsErrNoRows(err) {
			return fmt.Errorf("Gitops integration is not installed/configured. Please install/configure gitops.")
		}
		return err
	}

	// 2. Create Git Repo if required
	if gitOps.IsGitOpsRepoNotConfigured(manifestPushTemplate.RepoUrl) || manifestPushTemplate.GitOpsRepoMigrationRequired {
		if activeGlobalGitOpsConfig.AllowCustomRepository || manifestPushTemplate.IsCustomGitRepository {
			return fmt.Errorf("GitOps repository is not configured! Please configure gitops repository for application first.")
		}

		gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(manifestPushTemplate.AppName)
		chartGitAttr, err := impl.gitOperationService.CreateGitRepositoryForApp(gitOpsRepoName, manifestPushTemplate.UserId)
		if err != nil {
			impl.logger.Errorw("error in pushing chart to git ", "gitOpsRepoName", gitOpsRepoName, "err", err)
			return fmt.Errorf("No repository configured for Gitops! Error while creating git repository: '%s'", gitOpsRepoName)
		}
		err = impl.argoClientWrapperService.RegisterGitOpsRepoInArgo(ctx, chartGitAttr.RepoUrl, manifestPushTemplate.UserId)
		if err != nil {
			impl.logger.Errorw("error in registering app in acd", "err", err)
			return fmt.Errorf("Error in registering repository '%s' in ArgoCd", gitOpsRepoName)
		}
		manifestPushTemplate.RepoUrl = chartGitAttr.RepoUrl
		chartGitAttr.ChartLocation = manifestPushTemplate.ChartLocation
		// below function will override gitRepoUrl for charts even if user has already configured gitOps repoURL
		err = impl.chartService.OverrideGitOpsRepoUrl(manifestPushTemplate.AppId, chartGitAttr.RepoUrl, chartGitAttr.ChartLocation, manifestPushTemplate.UserId)
		if err != nil {
			impl.logger.Errorw("error in updating git repo url in charts", "err", err)
			return fmt.Errorf("No repository configured for Gitops! Error while creating git repository: '%s'", gitOpsRepoName)
		}
	}
	return nil
}

func (impl *GitOpsManifestPushServiceImpl) PushChart(manifestPushTemplate *bean.ManifestPushTemplate, ctx context.Context) bean.ManifestPushResponse {
	manifestPushResponse := bean.ManifestPushResponse{}

	// 1. Validate Repository for Git Operation
	errMsg := impl.ValidateRepoForGitOperation(manifestPushTemplate, ctx)
	if errMsg != nil {
		manifestPushResponse.Error = errMsg
		impl.SaveTimelineForError(manifestPushTemplate, errMsg)
		return manifestPushResponse
	}

	// 2. Push Chart to Git Repository
	err := impl.PushChartToGitRepo(manifestPushTemplate, ctx)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git", "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}

	// 3. Commit chart values to Git Repository
	commitHash, commitTime, err := impl.CommitValuesToGit(manifestPushTemplate, ctx)
	if err != nil {
		impl.logger.Errorw("error in committing values to git", "err", err)
		manifestPushResponse.Error = err
		impl.SaveTimelineForError(manifestPushTemplate, err)
		return manifestPushResponse
	}
	manifestPushResponse.CommitHash = commitHash
	manifestPushResponse.CommitTime = commitTime

	// 4. Update Deployment Status Timeline
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
	gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(manifestPushTemplate.RepoUrl)
	span.End()
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartService.CheckChartExists")
	err := impl.chartRefService.CheckChartExists(manifestPushTemplate.ChartRefId)
	span.End()
	if err != nil {
		impl.logger.Errorw("err in getting chart info", "err", err)
		return err
	}
	err = impl.gitOperationService.PushChartToGitRepo(gitOpsRepoName, manifestPushTemplate.ChartReferenceTemplate, manifestPushTemplate.ChartVersion, manifestPushTemplate.BuiltChartPath, manifestPushTemplate.RepoUrl, manifestPushTemplate.UserId)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git", "err", err)
		return err
	}
	return nil
}

func (impl *GitOpsManifestPushServiceImpl) CommitValuesToGit(manifestPushTemplate *bean.ManifestPushTemplate, ctx context.Context) (commitHash string, commitTime time.Time, err error) {
	commitHash = ""
	commitTime = time.Time{}
	chartRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(manifestPushTemplate.RepoUrl)
	_, span := otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit")
	//getting username & emailId for commit author data
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(manifestPushTemplate.UserId)
	span.End()
	chartGitAttr := &git.ChartConfig{
		FileName:       fmt.Sprintf("_%d-values.yaml", manifestPushTemplate.TargetEnvironmentName),
		FileContent:    string(manifestPushTemplate.MergedValues),
		ChartName:      manifestPushTemplate.ChartName,
		ChartLocation:  manifestPushTemplate.ChartLocation,
		ChartRepoName:  chartRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", manifestPushTemplate.PipelineOverrideId, manifestPushTemplate.TargetEnvironmentName),
		UserName:       userName,
		UserEmailId:    userEmailId,
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "gitOperationService.CommitValues")
	commitHash, commitTime, err = impl.gitOperationService.CommitValues(chartGitAttr)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
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
