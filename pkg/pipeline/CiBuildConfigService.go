package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
)

type CiBuildConfigService interface {
	Get(ciBuildConfigId int) (*bean.CiBuildConfig, error)
	Save(ciBuildConfig *bean.CiBuildConfig) (*bean.CiBuildConfig, error)
	Update(ciBuildConfig *bean.CiBuildConfig) (*bean.CiBuildConfig, error)
	Delete(ciBuildConfigId int) error
}

type CiBuildConfigImpl struct {
	Logger                  *zap.SugaredLogger
	CiBuildConfigRepository *pipelineConfig.CiBuildConfigRepository
}

func NewCiBuildConfigImpl(logger *zap.SugaredLogger, ciBuildConfigRepository *pipelineConfig.CiBuildConfigRepository) *CiBuildConfigImpl {
	return &CiBuildConfigImpl{
		Logger:                  logger,
		CiBuildConfigRepository: ciBuildConfigRepository,
	}
}

func (impl *CiBuildConfigImpl) Get(ciBuildConfigId int) (*bean.CiBuildConfig, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *CiBuildConfigImpl) Save(ciBuildConfig *bean.CiBuildConfig) (*bean.CiBuildConfig, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *CiBuildConfigImpl) Update(ciBuildConfig *bean.CiBuildConfig) (*bean.CiBuildConfig, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *CiBuildConfigImpl) Delete(ciBuildConfigId int) error {
	//TODO implement me
	panic("implement me")
}
