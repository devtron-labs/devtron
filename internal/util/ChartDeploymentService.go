package util

import (
	"context"
	repository3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	repository4 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	"go.uber.org/zap"
	"strings"
)

type ChartDeploymentService interface {
	RegisterInArgo(chartGitAttribute *ChartGitAttribute, userId int32, ctx context.Context, skipRetry bool) error
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

func (impl *ChartDeploymentServiceImpl) RegisterInArgo(chartGitAttribute *ChartGitAttribute, userId int32, ctx context.Context, skipRetry bool) error {
	repo := &v1alpha1.Repository{
		Repo: chartGitAttribute.RepoUrl,
	}
	// TODO Asutosh
	repo, err := impl.repositoryService.Create(ctx, &repository3.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil && strings.Contains(err.Error(), "Unable to resolve 'HEAD' to a commit SHA") {
		// - retry register in argo
		impl.logger.Infow("retrying argocd repo creation", "current err", err)
		repo, err = impl.repositoryService.Create(ctx, &repository3.RepoCreateRequest{Repo: &v1alpha1.Repository{
			Repo: chartGitAttribute.RepoUrl,
		}, Upsert: true})
		if err != nil {
			impl.logger.Errorw("retrying argocd repo creation", "current err", err)
		}
	}
	if err != nil && !strings.Contains(err.Error(), "Unable to resolve 'HEAD' to a commit SHA") {
		impl.logger.Errorw("error in creating argo Repository ", "err", err)
		if skipRetry {
			return err
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
		// - retry register in argo
		err = impl.RegisterInArgo(chartGitAttribute, userId, ctx, true)
		if err != nil {
			impl.logger.Errorw("error in re-try register in argo", "err", err)
			return err
		}
	}

	impl.logger.Infow("repo registered in argo", "name", chartGitAttribute.RepoUrl)
	return err
}
