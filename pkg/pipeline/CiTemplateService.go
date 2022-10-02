package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiTemplateService interface {
	Save(ciTemplateBean *bean.CiTemplateBean) error
	FindByAppId(appId int) (ciTemplateBean *bean.CiTemplateBean, err error)
	FindTemplateOverrideByAppId(appId int) (ciTemplateBeans []*bean.CiTemplateBean, err error)
	FindTemplateOverrideByCiPipelineId(ciPipelineId int) (*bean.CiTemplateBean, error)
	Update(ciTemplateBean *bean.CiTemplateBean) error
	FindByDockerRegistryId(dockerRegistryId string) (ciTemplates []*pipelineConfig.CiTemplate, err error)
	FindNumberOfAppsWithDockerConfigured(appIds []int) (int, error)
}
type CiTemplateServiceImpl struct {
	Logger                       *zap.SugaredLogger
	CiBuildConfigService         CiBuildConfigService
	CiTemplateRepository         pipelineConfig.CiTemplateRepository
	CiTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository
}

func NewCiTemplateServiceImpl(logger *zap.SugaredLogger, ciBuildConfigService CiBuildConfigService,
	ciTemplateRepository pipelineConfig.CiTemplateRepository, ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository) *CiTemplateServiceImpl {
	return &CiTemplateServiceImpl{
		Logger:                       logger,
		CiBuildConfigService:         ciBuildConfigService,
		CiTemplateRepository:         ciTemplateRepository,
		CiTemplateOverrideRepository: ciTemplateOverrideRepository,
	}
}

func (impl CiTemplateServiceImpl) Save(ciTemplateBean *bean.CiTemplateBean) error {
	ciTemplate := ciTemplateBean.CiTemplate
	ciTemplateOverride := ciTemplateBean.CiTemplateOverride
	ciTemplateId := 0
	ciTemplateOverrideId := 0
	if ciTemplateOverride == nil {
		err := impl.CiTemplateRepository.Save(ciTemplate)
		if err != nil {
			impl.Logger.Errorw("error in saving ci template in db ", "template", ciTemplate, "err", err)
			//TODO delete template from gocd otherwise dangling+ no create in future
			return err
		}
		ciTemplateId = ciTemplate.Id
	} else {
		_, err := impl.CiTemplateOverrideRepository.Save(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in saving template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
		ciTemplateOverrideId = ciTemplateOverride.Id
	}
	buildConfig := ciTemplateBean.CiBuildConfig
	err := impl.CiBuildConfigService.Save(ciTemplateId, ciTemplateOverrideId, buildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while saving ci build config", "config", buildConfig, "err", err)
	}
	return err
}

func (impl CiTemplateServiceImpl) FindByAppId(appId int) (ciTemplateBean *bean.CiTemplateBean, err error) {
	ciTemplate, err := impl.CiTemplateRepository.FindByAppId(appId)
	if err != nil {
		return nil, err
	}
	//dockerArgs := map[string]string{}
	//if err := json.Unmarshal([]byte(template.Args), &dockerArgs); err != nil {
	//	impl.logger.Debugw("error in json unmarshal", "app", appId, "err", err)
	//	return nil, err
	//}
	ciBuildConfig := ciTemplate.CiBuildConfig
	ciBuildConfigBean, err := bean.ConvertDbBuildConfigToBean(ciBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
			ciBuildConfig, "error", err)
	}
	return &bean.CiTemplateBean{
		CiTemplate:    ciTemplate,
		CiBuildConfig: ciBuildConfigBean,
	}, err
}

func (impl CiTemplateServiceImpl) FindTemplateOverrideByAppId(appId int) (ciTemplateBeans []*bean.CiTemplateBean, err error) {
	templateOverrides, err := impl.CiTemplateOverrideRepository.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting ciTemplateOverrides by appId", "err", err, "appId", appId)
		return nil, err
	}
	var templateBeanOverrides []*bean.CiTemplateBean
	for _, templateOverride := range templateOverrides {
		ciBuildConfigBean, err := bean.ConvertDbBuildConfigToBean(templateOverride.CiBuildConfig)
		if err != nil {
			impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
				templateOverride.CiBuildConfig, "error", err)
			return nil, err
		}
		overrideBean := &bean.CiTemplateBean{
			CiTemplateOverride: templateOverride,
			CiBuildConfig:      ciBuildConfigBean,
		}
		templateBeanOverrides = append(templateBeanOverrides, overrideBean)
	}
	return templateBeanOverrides, nil
}

func (impl CiTemplateServiceImpl) FindTemplateOverrideByCiPipelineId(ciPipelineId int) (*bean.CiTemplateBean, error) {
	templateOverride, err := impl.CiTemplateOverrideRepository.FindByCiPipelineId(ciPipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting ciTemplateOverrides by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return nil, err
	}
	ciBuildConfig := templateOverride.CiBuildConfig
	ciBuildConfigBean, err := bean.ConvertDbBuildConfigToBean(ciBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
			ciBuildConfig, "error", err)
	}
	return &bean.CiTemplateBean{CiTemplateOverride: templateOverride, CiBuildConfig: ciBuildConfigBean}, err
}

func (impl CiTemplateServiceImpl) Update(ciTemplateBean *bean.CiTemplateBean) error {
	ciTemplate := ciTemplateBean.CiTemplate
	ciTemplateOverride := ciTemplateBean.CiTemplateOverride
	ciTemplateId := 0
	ciTemplateOverrideId := 0
	if ciTemplateOverride == nil {
		err := impl.CiTemplateRepository.Update(ciTemplate)
		if err != nil {
			impl.Logger.Errorw("error in updating ci template in db", "template", ciTemplate, "err", err)
			return err
		}
		ciTemplateId = ciTemplate.Id
	} else {
		_, err := impl.CiTemplateOverrideRepository.Update(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in updating template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
		ciTemplateOverrideId = ciTemplateOverride.Id
	}
	ciBuildConfig := ciTemplateBean.CiBuildConfig
	_, err := impl.CiBuildConfigService.Update(ciTemplateId, ciTemplateOverrideId, ciBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error in updating ci build config in db", "ciBuildConfig", ciBuildConfig, "err", err)
	}
	return err
}

func (impl CiTemplateServiceImpl) FindByDockerRegistryId(dockerRegistryId string) (ciTemplates []*pipelineConfig.CiTemplate, err error) {
	//TODO implement me
	panic("implement me")
}

func (impl CiTemplateServiceImpl) FindNumberOfAppsWithDockerConfigured(appIds []int) (int, error) {
	//TODO implement me
	panic("implement me")
}
