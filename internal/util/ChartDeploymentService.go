package util

import (
	"context"
	repository3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	repository4 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	"go.uber.org/zap"
	"strings"
	"time"
)

type ChartDeploymentService interface {
	RegisterInArgo(chartGitAttribute *ChartGitAttribute, userId int32, ctx context.Context) error
}

type ChartDeploymentServiceImpl struct {
	logger                   *zap.SugaredLogger
	repositoryService        repository4.ServiceClient
	chartTemplateService     ChartTemplateService
	RegisterInArgoRetryCount int
}

func NewChartDeploymentServiceImpl(logger *zap.SugaredLogger, repositoryService repository4.ServiceClient, chartTemplateService ChartTemplateService) *ChartDeploymentServiceImpl {
	return &ChartDeploymentServiceImpl{
		logger:                   logger,
		repositoryService:        repositoryService,
		chartTemplateService:     chartTemplateService,
		RegisterInArgoRetryCount: 3,
	}
}

func (impl *ChartDeploymentServiceImpl) handleArgoRepoCreationError(retryCount int, repoUrl string, userId int32, argoCdErr error) (int, error) {
	argoRepoCreationErrorMessage := "Unable to resolve 'HEAD' to a commit SHA" // This error occurs inconsistently; ArgoCD requires 80-120s after last commit for create repository operation
	notArgoRepoCreationErrorMessage := !strings.Contains(argoCdErr.Error(), argoRepoCreationErrorMessage)

	emptyRepoErrorMessage := []string{"failed to get index: 404 Not Found", "remote repository is empty"} // ArgoCD can't register empty repo and throws these error message in such cases
	if retryCount >= impl.RegisterInArgoRetryCount &&
		notArgoRepoCreationErrorMessage &&
		!strings.Contains(argoCdErr.Error(), emptyRepoErrorMessage[0]) &&
		!strings.Contains(argoCdErr.Error(), emptyRepoErrorMessage[1]) {
		return 0, argoCdErr
	}
	if notArgoRepoCreationErrorMessage {
		// - found empty repository, create some file in repository
		gitOpsRepoName := GetGitRepoNameFromGitRepoUrl(repoUrl)
		err := impl.chartTemplateService.CreateReadmeInGitRepo(gitOpsRepoName, userId)
		if err != nil {
			impl.logger.Errorw("error in creating file in git repo", "err", err)
			return 0, err
		}
	}
	return retryCount + 1, nil
}

func (impl *ChartDeploymentServiceImpl) RegisterInArgo(chartGitAttribute *ChartGitAttribute, userId int32, ctx context.Context) error {

	retryCount := 0
	// label to register git repository in ArgoCd
	// ArgoCd requires approx 80 to 120 sec after the last commit to allow create-repository action
	// hence this operation needed to be perform with retry
CreateArgoRepositoryWithRetry:
	repo := &v1alpha1.Repository{
		Repo: chartGitAttribute.RepoUrl,
	}
	repo, err := impl.repositoryService.Create(ctx, &repository3.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository ", "err", err)
		retryCount, err = impl.handleArgoRepoCreationError(retryCount, chartGitAttribute.RepoUrl, userId, err)
		if err != nil {
			impl.logger.Errorw("error in RegisterInArgo with retry operation", "err", err)
			return err
		}
		impl.logger.Errorw("retrying RegisterInArgo operation", "retry count", retryCount)
		time.Sleep(10 * time.Second)
		goto CreateArgoRepositoryWithRetry
	}
	impl.logger.Infow("repo registered in argo", "name", chartGitAttribute.RepoUrl)
	return err
}
