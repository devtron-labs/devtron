package pipeline

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
)

type CiBuildConfigService interface {
	Save(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean) error
	Update(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean) (*bean.CiBuildConfigBean, error)
	Delete(ciBuildConfigId int) error
}

type CiBuildConfigServiceImpl struct {
	Logger                  *zap.SugaredLogger
	CiBuildConfigRepository pipelineConfig.CiBuildConfigRepository
}

func NewCiBuildConfigServiceImpl(logger *zap.SugaredLogger, ciBuildConfigRepository pipelineConfig.CiBuildConfigRepository) *CiBuildConfigServiceImpl {
	return &CiBuildConfigServiceImpl{
		Logger:                  logger,
		CiBuildConfigRepository: ciBuildConfigRepository,
	}
}

func (impl *CiBuildConfigServiceImpl) Save(templateId int, overrideTemplateId int, ciBuildConfigBean *bean.CiBuildConfigBean) error {
	ciBuildConfig, err := bean.ConvertBuildConfigBeanToDbEntity(templateId, overrideTemplateId, ciBuildConfigBean)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting build config to db entity", "templateId", templateId,
			"overrideTemplateId", overrideTemplateId, "ciBuildConfigBean", ciBuildConfigBean, "err", err)
		return errors.New("error while saving build config")
	}
	err = impl.CiBuildConfigRepository.Save(ciBuildConfig)
	if err != nil {
		return errors.New("error while saving build config")
	}
	return nil
}

func (impl *CiBuildConfigServiceImpl) Update(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean) (*bean.CiBuildConfigBean, error) {
	if ciBuildConfig == nil || ciBuildConfig.Id == 0 {
		impl.Logger.Warnw("not updating build config as object is empty", "ciBuildConfig", ciBuildConfig)
		return nil, nil
	}
	ciBuildConfigEntity, err := bean.ConvertBuildConfigBeanToDbEntity(templateId, overrideTemplateId, ciBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting build config to db entity", "templateId", templateId,
			"overrideTemplateId", overrideTemplateId, "ciBuildConfig", ciBuildConfig, "err", err)
		return nil, errors.New("error while saving build config")
	}
	err = impl.CiBuildConfigRepository.Update(ciBuildConfigEntity)
	if err != nil {
		return nil, errors.New("error while updating build config")
	}
	return ciBuildConfig, nil
}

func (impl *CiBuildConfigServiceImpl) Delete(ciBuildConfigId int) error {
	return impl.CiBuildConfigRepository.Delete(ciBuildConfigId)
}
