package pipeline

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
	"time"
)

type CiBuildConfigService interface {
	Save(templateId int, overrideTemplateId int, ciBuildConfigBean *bean.CiBuildConfigBean, userId int32) error
	UpdateOrSave(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean, userId int32) (*bean.CiBuildConfigBean, error)
	Delete(ciBuildConfigId int) error
	GetCountByBuildType() map[bean.CiBuildType]int
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

func (impl *CiBuildConfigServiceImpl) Save(templateId int, overrideTemplateId int, ciBuildConfigBean *bean.CiBuildConfigBean, userId int32) error {
	ciBuildConfigEntity, err := bean.ConvertBuildConfigBeanToDbEntity(templateId, overrideTemplateId, ciBuildConfigBean, userId)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting build config to db entity", "templateId", templateId,
			"overrideTemplateId", overrideTemplateId, "ciBuildConfigBean", ciBuildConfigBean, "err", err)
		return errors.New("error while saving build config")
	}
	ciBuildConfigEntity.CreatedOn = time.Now()
	ciBuildConfigEntity.CreatedBy = userId
	ciBuildConfigEntity.Id = 0
	err = impl.CiBuildConfigRepository.Save(ciBuildConfigEntity)
	ciBuildConfigBean.Id = ciBuildConfigEntity.Id
	if err != nil {
		return errors.New("error while saving build config")
	}
	return nil
}

func (impl *CiBuildConfigServiceImpl) UpdateOrSave(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean, userId int32) (*bean.CiBuildConfigBean, error) {
	if ciBuildConfig == nil {
		impl.Logger.Warnw("not updating build config as object is empty", "ciBuildConfig", ciBuildConfig)
		return nil, nil
	}
	ciBuildConfigEntity, err := bean.ConvertBuildConfigBeanToDbEntity(templateId, overrideTemplateId, ciBuildConfig, userId)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting build config to db entity", "templateId", templateId,
			"overrideTemplateId", overrideTemplateId, "ciBuildConfig", ciBuildConfig, "err", err)
		return nil, errors.New("error while saving build config")
	}
	if ciBuildConfig.Id == 0 {
		ciBuildConfigEntity.CreatedOn = time.Now()
		ciBuildConfigEntity.CreatedBy = userId
		err = impl.CiBuildConfigRepository.Save(ciBuildConfigEntity)
		ciBuildConfig.Id = ciBuildConfigEntity.Id
	} else {
		err = impl.CiBuildConfigRepository.Update(ciBuildConfigEntity)
	}
	if err != nil {
		impl.Logger.Errorw("error occurred while updating/saving ciBuildConfig", "entity", ciBuildConfigEntity, "err", err)
		return nil, errors.New("error while updating build config")
	}
	return ciBuildConfig, nil
}

func (impl *CiBuildConfigServiceImpl) Delete(ciBuildConfigId int) error {
	return impl.CiBuildConfigRepository.Delete(ciBuildConfigId)
}

func (impl *CiBuildConfigServiceImpl) GetCountByBuildType() map[bean.CiBuildType]int {
	result := make(map[bean.CiBuildType]int)
	buildTypeVsCount, err := impl.CiBuildConfigRepository.GetCountByBuildType()
	if err != nil {
		return result
	}
	for buildType, count := range buildTypeVsCount {
		result[bean.CiBuildType(buildType)] = count
	}
	return result
}
