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
	logger               *zap.SugaredLogger
	repositoryService    repository4.ServiceClient
	chartTemplateService ChartTemplateService
}

func NewChartDeploymentServiceImpl(logger *zap.SugaredLogger, repositoryService repository4.ServiceClient, chartTemplateService ChartTemplateService) *ChartDeploymentServiceImpl {
	return &ChartDeploymentServiceImpl{
		logger:               logger,
		repositoryService:    repositoryService,
		chartTemplateService: chartTemplateService,
	}
}

func (impl *ChartDeploymentServiceImpl) RegisterInArgo(chartGitAttribute *ChartGitAttribute, userId int32, ctx context.Context) error {
	repo := &v1alpha1.Repository{
		Repo: chartGitAttribute.RepoUrl,
	}
	retryCount := 3
	// label to register git repository in ArgoCd
	// ArgoCd requires approx 80 to 120 sec after the last commit to allow create-repository action
	// hence this operation needed to be perform with retry
CreateArgoRepositoryWithRetry:
	repo, err := impl.repositoryService.Create(ctx, &repository3.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository ", "retry count", retryCount, "err", err)
		if strings.Contains(err.Error(), "Unable to resolve 'HEAD' to a commit SHA") {
			retryCount -= 1
			time.Sleep(10 * time.Second)
			goto CreateArgoRepositoryWithRetry
		}
		emptyRepoErrorMessage := []string{"failed to get index: 404 Not Found", "remote repository is empty"}
		if !strings.Contains(err.Error(), emptyRepoErrorMessage[0]) && !strings.Contains(err.Error(), emptyRepoErrorMessage[1]) {
			return err
		}
		// - found empty repository, create some file in repository
		gitOpsRepoName := GetGitRepoNameFromGitRepoUrl(chartGitAttribute.RepoUrl)
		err = impl.chartTemplateService.CreateReadmeInGitRepo(gitOpsRepoName, userId)
		if err != nil {
			impl.logger.Errorw("error in creating file in git repo", "err", err)
			return err
		}
		retryCount -= 1
		time.Sleep(10 * time.Second)
		goto CreateArgoRepositoryWithRetry
	}
	impl.logger.Infow("repo registered in argo", "name", chartGitAttribute.RepoUrl)
	return err
}
