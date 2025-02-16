package read

import (
	"fmt"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	interalUtil "github.com/devtron-labs/devtron/internal/util"
	serviceBean "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/common/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
)

type DeploymentConfigReadService interface {
	GetDeploymentConfigMinForAppAndEnv(appId, envId int) (*bean.DeploymentConfigMin, error)
	GetDeploymentAppTypeForCDInBulk(pipelines []*serviceBean.CDPipelineMinConfig, appIdToGitOpsConfiguredMap map[int]bool) (map[int]*bean.DeploymentConfigMin, error)
}

type DeploymentConfigReadServiceImpl struct {
	logger                      *zap.SugaredLogger
	deploymentConfigRepository  deploymentConfig.Repository
	deploymentServiceTypeConfig *util.DeploymentServiceTypeConfig
}

func NewDeploymentConfigReadServiceImpl(logger *zap.SugaredLogger,
	deploymentConfigRepository deploymentConfig.Repository,
	envVariables *util.EnvironmentVariables) *DeploymentConfigReadServiceImpl {
	return &DeploymentConfigReadServiceImpl{
		logger:                      logger,
		deploymentConfigRepository:  deploymentConfigRepository,
		deploymentServiceTypeConfig: envVariables.DeploymentServiceTypeConfig,
	}
}

func (impl *DeploymentConfigReadServiceImpl) GetDeploymentConfigMinForAppAndEnv(appId, envId int) (*bean.DeploymentConfigMin, error) {
	deploymentDetail := &bean.DeploymentConfigMin{}
	config, err := impl.deploymentConfigRepository.GetByAppIdAndEnvId(appId, envId)
	if err != nil && !interalUtil.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
		return deploymentDetail, err
	} else if interalUtil.IsErrNoRows(err) {
		deploymentDetail.ReleaseMode = interalUtil.PIPELINE_RELEASE_MODE_CREATE
		return deploymentDetail, nil
	}
	configBean, err := adapter.ConvertDeploymentConfigDbObjToDTO(config)
	if err != nil {
		impl.logger.Errorw("error, ConvertDeploymentConfigDbObjToDTO", "config", config, "err", err)
		return nil, err
	}
	if configBean != nil {
		deploymentDetail.DeploymentAppType = configBean.DeploymentAppType
		deploymentDetail.ReleaseMode = configBean.ReleaseMode
		deploymentDetail.GitRepoUrl = configBean.GetRepoURL()
		deploymentDetail.IsGitOpsRepoConfigured = !apiGitOpsBean.IsGitOpsRepoNotConfigured(configBean.GetRepoURL())
	}
	return deploymentDetail, nil
}

func (impl *DeploymentConfigReadServiceImpl) GetDeploymentAppTypeForCDInBulk(pipelines []*serviceBean.CDPipelineMinConfig, appIdToGitOpsConfiguredMap map[int]bool) (map[int]*bean.DeploymentConfigMin, error) {
	resp := make(map[int]*bean.DeploymentConfigMin, len(pipelines)) //map of pipelineId and deploymentAppType
	appIdEnvIdMapping := make(map[int][]int, len(pipelines))
	appIdEnvIdKeyPipelineIdMap := make(map[string]int, len(pipelines))
	for _, pipeline := range pipelines {
		appIdEnvIdMapping[pipeline.AppId] = append(appIdEnvIdMapping[pipeline.AppId], pipeline.EnvironmentId)
		appIdEnvIdKeyPipelineIdMap[fmt.Sprintf("%d-%d", pipeline.AppId, pipeline.EnvironmentId)] = pipeline.Id
	}
	configs, err := impl.deploymentConfigRepository.GetAppAndEnvLevelConfigsInBulk(appIdEnvIdMapping)
	if err != nil {
		impl.logger.Errorw("error, GetAppAndEnvLevelConfigsInBulk", "appIdEnvIdMapping", appIdEnvIdMapping, "err", err)
		return nil, err
	}
	for _, config := range configs {
		configBean, err := adapter.ConvertDeploymentConfigDbObjToDTO(config)
		if err != nil {
			impl.logger.Errorw("error, ConvertDeploymentConfigDbObjToDTO", "config", config, "err", err)
			return nil, err
		}
		pipelineId := appIdEnvIdKeyPipelineIdMap[fmt.Sprintf("%d-%d", configBean.AppId, configBean.EnvironmentId)]
		isGitOpsRepoConfigured := configBean.IsPipelineGitOpsRepoConfigured(appIdToGitOpsConfiguredMap[configBean.AppId])
		resp[pipelineId] = adapter.NewDeploymentConfigMin(configBean.DeploymentAppType, configBean.ReleaseMode, isGitOpsRepoConfigured)
	}
	for _, pipeline := range pipelines {
		if _, ok := resp[pipeline.Id]; !ok {
			isGitOpsRepoConfigured := appIdToGitOpsConfiguredMap[pipeline.AppId]
			// not found in map, either flag is disabled or config not migrated yet. Getting from old data
			resp[pipeline.Id] = adapter.NewDeploymentConfigMin(pipeline.DeploymentAppType, interalUtil.PIPELINE_RELEASE_MODE_CREATE, isGitOpsRepoConfigured)
		}
	}
	return resp, nil
}
