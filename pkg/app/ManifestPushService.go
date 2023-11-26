package app

import (
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	. "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	status2 "github.com/devtron-labs/devtron/pkg/app/status"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/chartRepo"
	"github.com/devtron-labs/devtron/pkg/gitops"
	"github.com/devtron-labs/devtron/pkg/sql"
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
	logger                        *zap.SugaredLogger
	chartTemplateService          util.ChartTemplateService
	chartService                  chartService.ChartService
	gitOpsConfigRepository        repository.GitOpsConfigRepository
	gitFactory                    *GitFactory
	pipelineStatusTimelineService status2.PipelineStatusTimelineService
	gitOpsConfigService           gitops.GitOpsConfigService
}

func NewGitOpsManifestPushServiceImpl(
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	chartService chartService.ChartService,
	gitOpsConfigRepository repository.GitOpsConfigRepository,
	gitFactory *GitFactory,
	pipelineStatusTimelineService status2.PipelineStatusTimelineService,
	gitOpsConfigService gitops.GitOpsConfigService,
) *GitOpsManifestPushServiceImpl {
	return &GitOpsManifestPushServiceImpl{
		logger:                        logger,
		chartTemplateService:          chartTemplateService,
		chartService:                  chartService,
		gitOpsConfigRepository:        gitOpsConfigRepository,
		gitFactory:                    gitFactory,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
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

	if manifestPushTemplate.RepoUrl == chartRepo.GIT_REPO_NOT_CONFIGURED {
		if activeGlobalGitOpsConfig.AllowCustomRepository {
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

	timeline := getTimelineObject(manifestPushTemplate, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.")
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
	impl.logger.Errorw("Error in saving git commit success timeline", err, timelineErr)

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
	timeline := getTimelineObject(manifestPushTemplate, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", gitCommitErr))
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
	if timelineErr != nil {
		impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
	}
}

func getTimelineObject(manifestPushTemplate *bean.ManifestPushTemplate, status string, statusDetail string) *pipelineConfig.PipelineStatusTimeline {
	return &pipelineConfig.PipelineStatusTimeline{
		CdWorkflowRunnerId: manifestPushTemplate.WorkflowRunnerId,
		Status:             status,
		StatusDetail:       statusDetail,
		StatusTime:         time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: manifestPushTemplate.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: manifestPushTemplate.UserId,
			UpdatedOn: time.Now(),
		},
	}
}
