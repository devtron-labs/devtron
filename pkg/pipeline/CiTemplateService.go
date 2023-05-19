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
	FindTemplateOverrideByCiPipelineIds(ciPipelineIds []int) (ciTemplateBeans []*bean.CiTemplateBean, err error)
	FindTemplateOverrideByCiPipelineId(ciPipelineId int) (*bean.CiTemplateBean, error)
	Update(ciTemplateBean *bean.CiTemplateBean) error
	FindByAppIds(appIds []int) (map[int]*bean.CiTemplateBean, error)
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

	buildConfig := ciTemplateBean.CiBuildConfig
	err := impl.CiBuildConfigService.Save(ciTemplateId, ciTemplateOverrideId, buildConfig, ciTemplateBean.UserId)
	if err != nil {
		impl.Logger.Errorw("error occurred while saving ci build config", "config", buildConfig, "err", err)
	}
	if ciTemplateOverride == nil {
		ciTemplate.CiBuildConfigId = buildConfig.Id
		err := impl.CiTemplateRepository.Save(ciTemplate)
		if err != nil {
			impl.Logger.Errorw("error in saving ci template in db ", "template", ciTemplate, "err", err)
			//TODO delete template from gocd otherwise dangling+ no create in future
			return err
		}
		ciTemplateId = ciTemplate.Id
	} else {
		ciTemplateOverride.CiBuildConfigId = buildConfig.Id
		_, err := impl.CiTemplateOverrideRepository.Save(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in saving template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
		ciTemplateOverrideId = ciTemplateOverride.Id
	}

	return err
}

func (impl CiTemplateServiceImpl) FindByAppId(appId int) (ciTemplateBean *bean.CiTemplateBean, err error) {
	ciTemplate, err := impl.CiTemplateRepository.FindByAppId(appId)
	if err != nil {
		return nil, err
	}
	ciBuildConfig := ciTemplate.CiBuildConfig
	ciBuildConfigBean, err := bean.ConvertDbBuildConfigToBean(ciBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
			ciBuildConfig, "error", err)
	}
	if ciBuildConfigBean == nil {
		ciBuildConfigBean, err = bean.OverrideCiBuildConfig(ciTemplate.DockerfilePath, ciTemplate.Args, "", ciTemplate.DockerBuildOptions, ciTemplate.TargetPlatform, nil)
		if err != nil {
			impl.Logger.Errorw("error occurred while parsing ci build config", "err", err)
		}
	}
	ciBuildConfigBean.GitMaterialId = ciTemplate.GitMaterialId
	ciBuildConfigBean.BuildContextGitMaterialId = ciTemplate.BuildContextGitMaterialId
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
		ciBuildConfigBean, err := impl.extractBuildConfigBean(templateOverride)
		if err != nil {
			return templateBeanOverrides, err
		}
		overrideBean := &bean.CiTemplateBean{
			CiTemplateOverride: templateOverride,
			CiBuildConfig:      ciBuildConfigBean,
		}
		templateBeanOverrides = append(templateBeanOverrides, overrideBean)
	}
	return templateBeanOverrides, nil
}

func (impl CiTemplateServiceImpl) FindTemplateOverrideByCiPipelineIds(ciPipelineIds []int) (ciTemplateBeans []*bean.CiTemplateBean, err error) {
	templateOverrides, err := impl.CiTemplateOverrideRepository.FindByCiPipelineIds(ciPipelineIds)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting ciTemplateOverrides by appId", "err", err, "ciPipelineIds", ciPipelineIds)
		return nil, err
	}
	var templateBeanOverrides []*bean.CiTemplateBean
	for _, templateOverride := range templateOverrides {
		ciBuildConfigBean, err := impl.extractBuildConfigBean(templateOverride)
		if err != nil {
			return templateBeanOverrides, err
		}
		overrideBean := &bean.CiTemplateBean{
			CiTemplateOverride: templateOverride,
			CiBuildConfig:      ciBuildConfigBean,
		}
		templateBeanOverrides = append(templateBeanOverrides, overrideBean)
	}
	return templateBeanOverrides, nil
}

func (impl CiTemplateServiceImpl) extractBuildConfigBean(templateOverride *pipelineConfig.CiTemplateOverride) (*bean.CiBuildConfigBean, error) {
	ciBuildConfigBean, err := bean.ConvertDbBuildConfigToBean(templateOverride.CiBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
			templateOverride.CiBuildConfig, "error", err)
		return nil, err
	}
	if ciBuildConfigBean == nil {
		ciBuildConfigBean, err = bean.OverrideCiBuildConfig(templateOverride.DockerfilePath, "", "", "", "", nil)
		if err != nil {
			impl.Logger.Errorw("error occurred while parsing ci build config", "err", err)
		}
	}
	ciBuildConfigBean.GitMaterialId = templateOverride.GitMaterialId
	ciBuildConfigBean.BuildContextGitMaterialId = templateOverride.BuildContextGitMaterialId
	return ciBuildConfigBean, nil
}

func (impl CiTemplateServiceImpl) FindTemplateOverrideByCiPipelineId(ciPipelineId int) (*bean.CiTemplateBean, error) {
	templateOverride, err := impl.CiTemplateOverrideRepository.FindByCiPipelineId(ciPipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting ciTemplateOverrides by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return nil, err
	}
	ciBuildConfigBean, err := impl.extractBuildConfigBean(templateOverride)
	return &bean.CiTemplateBean{CiTemplateOverride: templateOverride, CiBuildConfig: ciBuildConfigBean}, err
}

func (impl CiTemplateServiceImpl) Update(ciTemplateBean *bean.CiTemplateBean) error {
	ciTemplate := ciTemplateBean.CiTemplate
	ciTemplateOverride := ciTemplateBean.CiTemplateOverride
	ciTemplateId := 0
	ciTemplateOverrideId := 0
	ciBuildConfig := ciTemplateBean.CiBuildConfig
	if ciTemplateOverride == nil {
		ciTemplateId = ciTemplate.Id
	} else {
		ciTemplateOverrideId = ciTemplateOverride.Id
	}
	_, err := impl.CiBuildConfigService.UpdateOrSave(ciTemplateId, ciTemplateOverrideId, ciBuildConfig, ciTemplateBean.UserId)
	if err != nil {
		impl.Logger.Errorw("error in updating ci build config in db", "ciBuildConfig", ciBuildConfig, "err", err)
	}
	if ciTemplateOverride == nil {
		ciTemplate.CiBuildConfigId = ciBuildConfig.Id
		err := impl.CiTemplateRepository.Update(ciTemplate)
		if err != nil {
			impl.Logger.Errorw("error in updating ci template in db", "template", ciTemplate, "err", err)
			return err
		}
	} else {
		ciTemplateOverride.CiBuildConfigId = ciBuildConfig.Id
		_, err := impl.CiTemplateOverrideRepository.Update(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in updating template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
	}
	return err
}

func (impl CiTemplateServiceImpl) FindByAppIds(appIds []int) (map[int]*bean.CiTemplateBean, error) {
	ciTemplates, err := impl.CiTemplateRepository.FindByAppIds(appIds)
	if err != nil {
		return nil, err
	}
	ciTemplateMap := make(map[int]*bean.CiTemplateBean)
	for _, ciTemplate := range ciTemplates {
		ciBuildConfig := ciTemplate.CiBuildConfig
		ciBuildConfigBean, err := bean.ConvertDbBuildConfigToBean(ciBuildConfig)
		if err != nil {
			impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
				ciBuildConfig, "error", err)
		}
		if ciBuildConfigBean == nil {
			ciBuildConfigBean, err = bean.OverrideCiBuildConfig(ciTemplate.DockerfilePath, ciTemplate.Args, "", ciTemplate.DockerBuildOptions, ciTemplate.TargetPlatform, nil)
			if err != nil {
				impl.Logger.Errorw("error occurred while parsing ci build config", "err", err)
			}
		}
		ciBuildConfigBean.GitMaterialId = ciTemplate.GitMaterialId
		ciTemplateBean := &bean.CiTemplateBean{
			CiTemplate:    ciTemplate,
			CiBuildConfig: ciBuildConfigBean,
		}
		ciTemplateMap[ciTemplate.AppId] = ciTemplateBean
	}
	return ciTemplateMap, nil
}
