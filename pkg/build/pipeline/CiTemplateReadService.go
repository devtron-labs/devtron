package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiTemplateReadService interface {
	FindByAppId(appId int) (*bean.CiTemplateBean, error)
	FindTemplateOverrideByAppId(appId int) ([]*bean.CiTemplateBean, error)
	FindTemplateOverrideByCiPipelineIds(ciPipelineIds []int) ([]*bean.CiTemplateBean, error)
	FindTemplateOverrideByCiPipelineId(ciPipelineId int) (*bean.CiTemplateBean, error)
	GetAppliedDockerConfigForCiPipeline(ciPipelineId, appId int, isOverridden bool) (*types.DockerArtifactStoreBean, error)
	GetBaseDockerConfigForCiPipeline(appId int) (*types.DockerArtifactStoreBean, error)
	FindByAppIds(appIds []int) (map[int]*bean.CiTemplateBean, error)
}

type CiTemplateReadServiceImpl struct {
	Logger                       *zap.SugaredLogger
	CiTemplateRepository         pipelineConfig.CiTemplateRepository
	CiTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository
}

func NewCiTemplateReadServiceImpl(logger *zap.SugaredLogger, ciTemplateRepository pipelineConfig.CiTemplateRepository, ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository) *CiTemplateReadServiceImpl {
	return &CiTemplateReadServiceImpl{
		Logger:                       logger,
		CiTemplateRepository:         ciTemplateRepository,
		CiTemplateOverrideRepository: ciTemplateOverrideRepository,
	}
}

func (impl *CiTemplateReadServiceImpl) FindByAppId(appId int) (ciTemplateBean *bean.CiTemplateBean, err error) {
	ciTemplate, err := impl.CiTemplateRepository.FindByAppId(appId)
	if err != nil {
		return nil, err
	}
	ciBuildConfig := ciTemplate.CiBuildConfig
	ciBuildConfigBean, err := adapter.ConvertDbBuildConfigToBean(ciBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
			ciBuildConfig, "error", err)
	}
	if ciBuildConfigBean == nil {
		ciBuildConfigBean, err = adapter.OverrideCiBuildConfig(ciTemplate.DockerfilePath, ciTemplate.Args, "", ciTemplate.DockerBuildOptions, ciTemplate.TargetPlatform, nil)
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

func (impl *CiTemplateReadServiceImpl) FindTemplateOverrideByAppId(appId int) (ciTemplateBeans []*bean.CiTemplateBean, err error) {
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

func (impl *CiTemplateReadServiceImpl) FindTemplateOverrideByCiPipelineIds(ciPipelineIds []int) (ciTemplateBeans []*bean.CiTemplateBean, err error) {
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

func (impl *CiTemplateReadServiceImpl) extractBuildConfigBean(templateOverride *pipelineConfig.CiTemplateOverride) (*bean.CiBuildConfigBean, error) {
	ciBuildConfigBean, err := adapter.ConvertDbBuildConfigToBean(templateOverride.CiBuildConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
			templateOverride.CiBuildConfig, "error", err)
		return nil, err
	}
	if ciBuildConfigBean == nil {
		ciBuildConfigBean, err = adapter.OverrideCiBuildConfig(templateOverride.DockerfilePath, "", "", "", "", nil)
		if err != nil {
			impl.Logger.Errorw("error occurred while parsing ci build config", "err", err)
		}
	}
	ciBuildConfigBean.GitMaterialId = templateOverride.GitMaterialId
	ciBuildConfigBean.BuildContextGitMaterialId = templateOverride.BuildContextGitMaterialId
	return ciBuildConfigBean, nil
}

func (impl *CiTemplateReadServiceImpl) FindTemplateOverrideByCiPipelineId(ciPipelineId int) (*bean.CiTemplateBean, error) {
	templateOverride, err := impl.CiTemplateOverrideRepository.FindByCiPipelineId(ciPipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting ciTemplateOverrides by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return nil, err
	}
	ciBuildConfigBean, err := impl.extractBuildConfigBean(templateOverride)
	return &bean.CiTemplateBean{CiTemplateOverride: templateOverride, CiBuildConfig: ciBuildConfigBean}, err
}

func (impl *CiTemplateReadServiceImpl) GetAppliedDockerConfigForCiPipeline(ciPipelineId, appId int, isOverridden bool) (*types.DockerArtifactStoreBean, error) {
	if !isOverridden {
		return impl.GetBaseDockerConfigForCiPipeline(appId)
	}
	templateOverride, err := impl.CiTemplateOverrideRepository.FindByCiPipelineId(ciPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting ciTemplateOverrides by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return nil, err
	}
	if util.IsErrNoRows(err) {
		return impl.GetBaseDockerConfigForCiPipeline(appId)
	}
	return adapter.GetDockerConfigBean(templateOverride.DockerRegistry), nil
}

func (impl *CiTemplateReadServiceImpl) GetBaseDockerConfigForCiPipeline(appId int) (*types.DockerArtifactStoreBean, error) {
	ciTemplate, err := impl.CiTemplateRepository.FindByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting ciTemplate by appId", "err", err, "appId", appId)
		return nil, err
	}
	return adapter.GetDockerConfigBean(ciTemplate.DockerRegistry), nil
}

func (impl *CiTemplateReadServiceImpl) FindByAppIds(appIds []int) (map[int]*bean.CiTemplateBean, error) {
	ciTemplates, err := impl.CiTemplateRepository.FindByAppIds(appIds)
	if err != nil {
		return nil, err
	}
	ciTemplateMap := make(map[int]*bean.CiTemplateBean)
	for _, ciTemplate := range ciTemplates {
		ciBuildConfig := ciTemplate.CiBuildConfig
		ciBuildConfigBean, err := adapter.ConvertDbBuildConfigToBean(ciBuildConfig)
		if err != nil {
			impl.Logger.Errorw("error occurred while converting dbBuildConfig to bean", "ciBuildConfig",
				ciBuildConfig, "error", err)
		}
		if ciBuildConfigBean == nil {
			ciBuildConfigBean, err = adapter.OverrideCiBuildConfig(ciTemplate.DockerfilePath, ciTemplate.Args, "", ciTemplate.DockerBuildOptions, ciTemplate.TargetPlatform, nil)
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
