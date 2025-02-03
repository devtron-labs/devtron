package read

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/deployment/common/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
)

type DeploymentConfigReadService interface {
	GetDeploymentAppTypeForCDInBulk(pipelines []*pipelineConfig.Pipeline, appIdToGitOpsConfiguredMap map[int]bool) (map[int]*bean.DeploymentConfigMin, error)
}

type DeploymentConfigReadServiceImpl struct {
	logger                      *zap.SugaredLogger
	deploymentServiceTypeConfig *util.DeploymentServiceTypeConfig
	deploymentConfigRepository  deploymentConfig.Repository
}

func NewDeploymentConfigReadServiceImpl(logger *zap.SugaredLogger,
	deploymentConfigRepository deploymentConfig.Repository,
	envVariables *util.EnvironmentVariables) *DeploymentConfigReadServiceImpl {
	return &DeploymentConfigReadServiceImpl{
		logger:                      logger,
		deploymentServiceTypeConfig: envVariables.DeploymentServiceTypeConfig,
		deploymentConfigRepository:  deploymentConfigRepository,
	}
}

func (impl *DeploymentConfigReadServiceImpl) GetDeploymentAppTypeForCDInBulk(pipelines []*pipelineConfig.Pipeline, appIdToGitOpsConfiguredMap map[int]bool) (map[int]*bean.DeploymentConfigMin, error) {
	resp := make(map[int]*bean.DeploymentConfigMin, len(pipelines)) //map of pipelineId and deploymentAppType
	if impl.deploymentServiceTypeConfig.UseDeploymentConfigData {
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
	}
	for _, pipeline := range pipelines {
		if _, ok := resp[pipeline.Id]; !ok {
			isGitOpsRepoConfigured := appIdToGitOpsConfiguredMap[pipeline.AppId]
			// not found in map, either flag is disabled or config not migrated yet. Getting from old data
			resp[pipeline.Id] = adapter.NewDeploymentConfigMin(pipeline.DeploymentAppType, util2.PIPELINE_RELEASE_MODE_CREATE, isGitOpsRepoConfigured)
		}
	}
	return resp, nil
}
