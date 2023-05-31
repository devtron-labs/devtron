package util

import (
	"context"
	repository3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	repository4 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	"go.uber.org/zap"
)

type ChartDeploymentService interface {
	RegisterInArgo(chartGitAttribute *ChartGitAttribute, ctx context.Context) error
}

type ChartDeploymentServiceImpl struct {
	logger            *zap.SugaredLogger
	repositoryService repository4.ServiceClient
}

func NewChartDeploymentServiceImpl(logger *zap.SugaredLogger, repositoryService repository4.ServiceClient) *ChartDeploymentServiceImpl {
	return &ChartDeploymentServiceImpl{
		logger:            logger,
		repositoryService: repositoryService,
	}
}

func (impl *ChartDeploymentServiceImpl) RegisterInArgo(chartGitAttribute *ChartGitAttribute, ctx context.Context) error {
	repo := &v1alpha1.Repository{
		Repo: chartGitAttribute.RepoUrl,
	}
	repo, err := impl.repositoryService.Create(ctx, &repository3.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository ", "err", err)
		return err
	}
	impl.logger.Infow("repo registered in argo", "name", chartGitAttribute.RepoUrl)
	return err
}
