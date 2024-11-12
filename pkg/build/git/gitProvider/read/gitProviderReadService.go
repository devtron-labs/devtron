package read

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
	"go.uber.org/zap"
)

type GitProviderReadService interface {
}

type GitProviderReadServiceImpl struct {
	logger                *zap.SugaredLogger
	gitProviderRepository repository.GitProviderRepository
}

func NewGitProviderReadService(logger *zap.SugaredLogger,
	gitProviderRepository repository.GitProviderRepository) *GitProviderReadServiceImpl {
	return &GitProviderReadServiceImpl{
		logger:                logger,
		gitProviderRepository: gitProviderRepository,
	}
}
