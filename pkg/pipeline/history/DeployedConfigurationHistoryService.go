package history

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeployedConfigurationHistoryService interface {
	GetDeployedConfigurationByWfrId(pipelineId, wfrId int) ([]*DeploymentConfigurationDto, error)
	GetDeployedHistoryComponentList(pipelineId, baseConfigId int, historyComponent, historyComponentName string) ([]*DeployedHistoryComponentMetadataDto, error)
	GetDeployedHistoryComponentDetail(pipelineId, id int, historyComponent, historyComponentName string, userHasAdminAccess bool) (*HistoryDetailDto, error)
	GetAllDeployedConfigurationByPipelineIdAndLatestWfrId(pipelineId int, userHasAdminAccess bool) (*AllDeploymentConfigurationDetail, error)
	GetAllDeployedConfigurationByPipelineIdAndWfrId(pipelineId, wfrId int, userHasAdminAccess bool) (*AllDeploymentConfigurationDetail, error)
}

type DeployedConfigurationHistoryServiceImpl struct {
	logger                           *zap.SugaredLogger
	userService                      user.UserService
	deploymentTemplateHistoryService DeploymentTemplateHistoryService
	strategyHistoryService           PipelineStrategyHistoryService
	configMapHistoryService          ConfigMapHistoryService
	cdWorkflowRepository             pipelineConfig.CdWorkflowRepository
}

func NewDeployedConfigurationHistoryServiceImpl(logger *zap.SugaredLogger,
	userService user.UserService, deploymentTemplateHistoryService DeploymentTemplateHistoryService,
	strategyHistoryService PipelineStrategyHistoryService, configMapHistoryService ConfigMapHistoryService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *DeployedConfigurationHistoryServiceImpl {
	return &DeployedConfigurationHistoryServiceImpl{
		logger:                           logger,
		userService:                      userService,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		strategyHistoryService:           strategyHistoryService,
		configMapHistoryService:          configMapHistoryService,
		cdWorkflowRepository:             cdWorkflowRepository,
	}
}

func (impl *DeployedConfigurationHistoryServiceImpl) GetDeployedConfigurationByWfrId(pipelineId, wfrId int) ([]*DeploymentConfigurationDto, error) {
	var deployedConfigurations []*DeploymentConfigurationDto
	//checking if deployment template configuration for this pipelineId and wfrId exists or not
	templateHistoryId, exists, err := impl.deploymentTemplateHistoryService.CheckIfHistoryExistsForPipelineIdAndWfrId(pipelineId, wfrId)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for deployment template", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	deploymentTemplateConfiguration := &DeploymentConfigurationDto{
		Name: DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT,
	}
	if exists {
		deploymentTemplateConfiguration.Id = templateHistoryId
	}
	deployedConfigurations = append(deployedConfigurations, deploymentTemplateConfiguration)

	//checking if pipeline strategy configuration for this pipelineId and wfrId exists or not
	strategyHistoryId, exists, err := impl.strategyHistoryService.CheckIfHistoryExistsForPipelineIdAndWfrId(pipelineId, wfrId)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for pipeline strategy", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	pipelineStrategyConfiguration := &DeploymentConfigurationDto{
		Name: PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT,
	}
	if exists {
		pipelineStrategyConfiguration.Id = strategyHistoryId
		deployedConfigurations = append(deployedConfigurations, pipelineStrategyConfiguration)
	}

	//checking if configmap history data exists and get its details
	configmapHistory, exists, names, err := impl.configMapHistoryService.GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId, repository.CONFIGMAP_TYPE)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for configmap", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	if exists {
		configmapConfiguration := &DeploymentConfigurationDto{
			Id:                  configmapHistory.Id,
			Name:                CONFIGMAP_TYPE_HISTORY_COMPONENT,
			ChildComponentNames: names,
		}
		deployedConfigurations = append(deployedConfigurations, configmapConfiguration)
	}

	//checking if secret history data exists and get its details
	secretHistory, exists, names, err := impl.configMapHistoryService.GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId, repository.SECRET_TYPE)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for secret", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	if exists {
		secretConfiguration := &DeploymentConfigurationDto{
			Id:                  secretHistory.Id,
			Name:                SECRET_TYPE_HISTORY_COMPONENT,
			ChildComponentNames: names,
		}
		deployedConfigurations = append(deployedConfigurations, secretConfiguration)
	}
	return deployedConfigurations, nil
}

func (impl *DeployedConfigurationHistoryServiceImpl) GetDeployedHistoryComponentList(pipelineId, baseConfigId int, historyComponent, historyComponentName string) ([]*DeployedHistoryComponentMetadataDto, error) {
	var historyList []*DeployedHistoryComponentMetadataDto
	var err error
	if historyComponent == string(DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.deploymentTemplateHistoryService.GetDeployedHistoryList(pipelineId, baseConfigId)
	} else if historyComponent == string(PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.strategyHistoryService.GetDeployedHistoryList(pipelineId, baseConfigId)
	} else if historyComponent == string(CONFIGMAP_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.configMapHistoryService.GetDeployedHistoryList(pipelineId, baseConfigId, repository.CONFIGMAP_TYPE, historyComponentName)
	} else if historyComponent == string(SECRET_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.configMapHistoryService.GetDeployedHistoryList(pipelineId, baseConfigId, repository.SECRET_TYPE, historyComponentName)
	} else {
		return nil, errors.New(fmt.Sprintf("history of %s not supported", historyComponent))
	}
	if err != nil {
		impl.logger.Errorw("error in getting deployed history component list", "err", err, "pipelineId", pipelineId, "historyComponent", historyComponent, "componentName", historyComponentName)
		return nil, err
	}
	return historyList, nil
}

func (impl *DeployedConfigurationHistoryServiceImpl) GetDeployedHistoryComponentDetail(pipelineId, id int, historyComponent, historyComponentName string, userHasAdminAccess bool) (*HistoryDetailDto, error) {
	history := &HistoryDetailDto{}
	var err error
	if historyComponent == string(DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT) {
		history, err = impl.deploymentTemplateHistoryService.GetHistoryForDeployedTemplateById(id, pipelineId)
	} else if historyComponent == string(PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT) {
		history, err = impl.strategyHistoryService.GetHistoryForDeployedStrategyById(id, pipelineId)
	} else if historyComponent == string(CONFIGMAP_TYPE_HISTORY_COMPONENT) {
		history, err = impl.configMapHistoryService.GetHistoryForDeployedCMCSById(id, pipelineId, repository.CONFIGMAP_TYPE, historyComponentName, userHasAdminAccess)
	} else if historyComponent == string(SECRET_TYPE_HISTORY_COMPONENT) {
		history, err = impl.configMapHistoryService.GetHistoryForDeployedCMCSById(id, pipelineId, repository.SECRET_TYPE, historyComponentName, userHasAdminAccess)
	} else {
		return nil, errors.New(fmt.Sprintf("history of %s not supported", historyComponent))
	}
	if err != nil {
		impl.logger.Errorw("error in getting deployed history component detail", "err", err, "pipelineId", pipelineId, "id", id, "historyComponent", historyComponent, "componentName", historyComponentName)
		return nil, err
	}
	return history, nil
}

func (impl *DeployedConfigurationHistoryServiceImpl) GetAllDeployedConfigurationByPipelineIdAndLatestWfrId(pipelineId int, userHasAdminAccess bool) (*AllDeploymentConfigurationDetail, error) {
	//getting latest wfr from pipelineId
	wfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest deploy stage wfr by pipelineId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	deployedConfig, err := impl.GetAllDeployedConfigurationByPipelineIdAndWfrId(pipelineId, wfr.Id, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getting GetAllDeployedConfigurationByPipelineIdAndWfrId", "err", err, "pipelineID", pipelineId, "wfrId", wfr.Id)
		return nil, err
	}
	deployedConfig.WfrId = wfr.Id
	return deployedConfig, nil
}
func (impl *DeployedConfigurationHistoryServiceImpl) GetAllDeployedConfigurationByPipelineIdAndWfrId(pipelineId, wfrId int, userHasAdminAccess bool) (*AllDeploymentConfigurationDetail, error) {
	//getting history of deployment template for latest deployment
	deploymentTemplateHistory, err := impl.deploymentTemplateHistoryService.GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment template history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	//getting history of config map for latest deployment
	configMapHistory, err := impl.configMapHistoryService.GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(pipelineId, wfrId, repository.CONFIGMAP_TYPE, userHasAdminAccess)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting config map history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	//getting history of secret for latest deployment
	secretHistory, err := impl.configMapHistoryService.GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(pipelineId, wfrId, repository.SECRET_TYPE, userHasAdminAccess)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting secret history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	//getting history of pipeline strategy for latest deployment
	strategyHistory, err := impl.strategyHistoryService.GetLatestDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting strategy history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	allDeploymentConfigurationHistoryDetail := &AllDeploymentConfigurationDetail{
		DeploymentTemplateConfig: deploymentTemplateHistory,
		ConfigMapConfig:          configMapHistory,
		SecretConfig:             secretHistory,
		StrategyConfig:           strategyHistory,
	}
	return allDeploymentConfigurationHistoryDetail, nil
}
