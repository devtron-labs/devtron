package read

import (
	"context"
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	read2 "github.com/devtron-labs/devtron/pkg/deployment/manifest/configMapAndSecret/read"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/adaptors"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type DeploymentTemplateHistoryReadService interface {
	GetHistoryForDeployedTemplateById(ctx context.Context, id int, pipelineId int) (*bean.HistoryDetailDto, error)
	CheckIfHistoryExistsForPipelineIdAndWfrId(pipelineId, wfrId int) (historyId int, exists bool, err error)
	CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (deploymentTemplateHistoryId int, exists bool, err error)
	GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*bean.DeployedHistoryComponentMetadataDto, error)
	// used for rollback
	GetDeployedHistoryByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (*bean.HistoryDetailDto, error)

	GetTemplateHistoryModelForDeployedTemplateById(deploymentTemplateHistoryId, pipelineId int) (*repository2.DeploymentTemplateHistory, error)
	GetDeployedOnByDeploymentTemplateAndPipelineId(deploymentTemplateHistoryId, pipelineId int) (time.Time, error)

	GetDeployedConfigurationByWfrId(ctx context.Context, pipelineId, wfrId int) ([]*bean.DeploymentConfigurationDto, error)
	GetDeployedHistoryComponentList(pipelineId, baseConfigId int, historyComponent, historyComponentName string) ([]*bean.DeployedHistoryComponentMetadataDto, error)
	GetDeployedHistoryComponentDetail(ctx context.Context, pipelineId, id int, historyComponent, historyComponentName string, userHasAdminAccess bool) (*bean.HistoryDetailDto, error)
	GetAllDeployedConfigurationByPipelineIdAndLatestWfrId(ctx context.Context, pipelineId int, userHasAdminAccess bool) (*bean.AllDeploymentConfigurationDetail, error)
	GetAllDeployedConfigurationByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int, userHasAdminAccess bool) (*bean.AllDeploymentConfigurationDetail, error)
	GetLatestDeployedArtifactByPipelineId(pipelineId int) (*repository3.CiArtifact, error)
}

type DeploymentTemplateHistoryReadServiceImpl struct {
	logger                              *zap.SugaredLogger
	deploymentTemplateHistoryRepository repository2.DeploymentTemplateHistoryRepository
	scopedVariableManager               variables.ScopedVariableManager
	strategyHistoryService              history.PipelineStrategyHistoryService
	cdWorkflowRepository                pipelineConfig.CdWorkflowRepository
	configMapHistoryReadService         read2.ConfigMapHistoryReadService
}

func NewDeploymentTemplateHistoryReadServiceImpl(
	logger *zap.SugaredLogger,
	deploymentTemplateHistoryRepository repository2.DeploymentTemplateHistoryRepository,
	scopedVariableManager variables.ScopedVariableManager,
	strategyHistoryService history.PipelineStrategyHistoryService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	configMapHistoryReadService read2.ConfigMapHistoryReadService,
) *DeploymentTemplateHistoryReadServiceImpl {
	return &DeploymentTemplateHistoryReadServiceImpl{
		logger:                              logger,
		deploymentTemplateHistoryRepository: deploymentTemplateHistoryRepository,
		scopedVariableManager:               scopedVariableManager,
		strategyHistoryService:              strategyHistoryService,
		cdWorkflowRepository:                cdWorkflowRepository,
		configMapHistoryReadService:         configMapHistoryReadService,
	}
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetHistoryForDeployedTemplateById(ctx context.Context, id int, pipelineId int) (*bean.HistoryDetailDto, error) {
	history, err := impl.deploymentTemplateHistoryRepository.GetHistoryForDeployedTemplateById(id, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment template history", "err", err, "id", id, "pipelineId", pipelineId)
		return nil, err
	}

	isSuperAdmin, err := util.GetIsSuperAdminFromContext(ctx)
	if err != nil {
		return nil, err
	}
	reference := repository.HistoryReference{
		HistoryReferenceId:   history.Id,
		HistoryReferenceType: repository.HistoryReferenceTypeDeploymentTemplate,
	}
	variableSnapshotMap, resolvedTemplate, err := impl.scopedVariableManager.GetVariableSnapshotAndResolveTemplate(history.Template, parsers.JsonVariableTemplate, reference, isSuperAdmin, false)
	if err != nil {
		impl.logger.Errorw("error while resolving template from history", "err", err, "id", id, "pipelineID", pipelineId)
	}
	return adaptors.GetHistoryDetailDto(history, variableSnapshotMap, resolvedTemplate), nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) CheckIfHistoryExistsForPipelineIdAndWfrId(pipelineId, wfrId int) (historyId int, exists bool, err error) {
	impl.logger.Debugw("received request, CheckIfHistoryExistsForPipelineIdAndWfrId", "pipelineId", pipelineId, "wfrId", wfrId)

	//checking if history exists for pipelineId and wfrId
	history, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return 0, false, err
	} else if err == pg.ErrNoRows {
		return 0, false, nil
	}
	return history.Id, true, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetDeployedHistoryByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (*bean.HistoryDetailDto, error) {
	impl.logger.Debugw("received request, GetDeployedHistoryByPipelineIdAndWfrId", "pipelineId", pipelineId, "wfrId", wfrId)

	//checking if history exists for pipelineId and wfrId
	history, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}

	isSuperAdmin, err := util.GetIsSuperAdminFromContext(ctx)
	if err != nil {
		return nil, err
	}
	reference := repository.HistoryReference{
		HistoryReferenceId:   history.Id,
		HistoryReferenceType: repository.HistoryReferenceTypeDeploymentTemplate,
	}
	variableSnapshotMap, resolvedTemplate, err := impl.scopedVariableManager.GetVariableSnapshotAndResolveTemplate(history.Template, parsers.JsonVariableTemplate, reference, isSuperAdmin, false)
	if err != nil {
		impl.logger.Errorw("error while resolving template from history", "err", err, "wfrId", wfrId, "pipelineID", pipelineId)
	}

	return adaptors.GetHistoryDetailDto(history, variableSnapshotMap, resolvedTemplate), nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (deploymentTemplateHistoryId int, exists bool, err error) {
	history, err := impl.deploymentTemplateHistoryRepository.GetDeployedHistoryForPipelineIdOnTime(pipelineId, deployedOn)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in checking if history exists for pipelineId and deployedOn", "err", err, "pipelineId", pipelineId, "deployedOn", deployedOn)
		return deploymentTemplateHistoryId, exists, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return deploymentTemplateHistoryId, exists, nil
	}
	deploymentTemplateHistoryId = history.Id
	exists = true
	return deploymentTemplateHistoryId, exists, err
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetTemplateHistoryModelForDeployedTemplateById(deploymentTemplateHistoryId, pipelineId int) (*repository2.DeploymentTemplateHistory, error) {
	history, err := impl.deploymentTemplateHistoryRepository.GetHistoryForDeployedTemplateById(deploymentTemplateHistoryId, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment template history", "err", err, "deploymentTemplateHistoryId", deploymentTemplateHistoryId, "pipelineId", pipelineId)
		return nil, err
	}
	return history, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*bean.DeployedHistoryComponentMetadataDto, error) {
	impl.logger.Debugw("received request, GetDeployedHistoryList", "pipelineId", pipelineId, "baseConfigId", baseConfigId)

	//checking if history exists for pipelineId and wfrId
	histories, err := impl.deploymentTemplateHistoryRepository.GetDeployedHistoryList(pipelineId, baseConfigId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting history list for pipelineId and baseConfigId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historyList []*bean.DeployedHistoryComponentMetadataDto
	for _, history := range histories {
		historyList = append(historyList, &bean.DeployedHistoryComponentMetadataDto{
			Id:               history.Id,
			DeployedOn:       history.DeployedOn,
			DeployedBy:       history.DeployedByEmailId,
			DeploymentStatus: history.DeploymentStatus,
		})
	}
	return historyList, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetDeployedOnByDeploymentTemplateAndPipelineId(deploymentTemplateHistoryId, pipelineId int) (time.Time, error) {
	deployedOn, err := impl.deploymentTemplateHistoryRepository.GetDeployedOnByDeploymentTemplateAndPipelineId(deploymentTemplateHistoryId, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment template history", "err", err, "deploymentTemplateHistoryId", deploymentTemplateHistoryId, "pipelineId", pipelineId)
		return deployedOn, err
	}
	return deployedOn, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetLatestDeployedArtifactByPipelineId(pipelineId int) (*repository3.CiArtifact, error) {
	wfr, err := impl.cdWorkflowRepository.FindLatestByPipelineIdAndRunnerType(pipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Infow("error in getting latest deploy stage wfr by pipelineId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return wfr.CdWorkflow.CiArtifact, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetDeployedConfigurationByWfrId(ctx context.Context, pipelineId, wfrId int) ([]*bean.DeploymentConfigurationDto, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "DeploymentTemplateHistoryReadServiceImpl.GetDeployedConfigurationByWfrId")
	defer span.End()
	var deployedConfigurations []*bean.DeploymentConfigurationDto
	//checking if deployment template configuration for this pipelineId and wfrId exists or not
	templateHistoryId, exists, err := impl.CheckIfHistoryExistsForPipelineIdAndWfrId(pipelineId, wfrId)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for deployment template", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}

	deploymentTemplateConfiguration := &bean.DeploymentConfigurationDto{
		Name: bean.DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT,
	}
	if exists {
		deploymentTemplateConfiguration.Id = templateHistoryId
	}
	deployedConfigurations = append(deployedConfigurations, deploymentTemplateConfiguration)

	//checking if pipeline strategy configuration for this pipelineId and wfrId exists or not

	strategyHistoryId, exists, err := impl.strategyHistoryService.CheckIfHistoryExistsForPipelineIdAndWfrId(newCtx, pipelineId, wfrId)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for pipeline strategy", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	pipelineStrategyConfiguration := &bean.DeploymentConfigurationDto{
		Name: bean.PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT,
	}
	if exists {
		pipelineStrategyConfiguration.Id = strategyHistoryId
		deployedConfigurations = append(deployedConfigurations, pipelineStrategyConfiguration)
	}

	//checking if configmap history data exists and get its details
	configmapHistory, exists, names, err := impl.configMapHistoryReadService.GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId, repository2.CONFIGMAP_TYPE)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for configmap", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	if exists {
		configmapConfiguration := &bean.DeploymentConfigurationDto{
			Id:                  configmapHistory.Id,
			Name:                bean.CONFIGMAP_TYPE_HISTORY_COMPONENT,
			ChildComponentNames: names,
		}
		deployedConfigurations = append(deployedConfigurations, configmapConfiguration)
	}

	//checking if secret history data exists and get its details
	secretHistory, exists, names, err := impl.configMapHistoryReadService.GetDeployedHistoryByPipelineIdAndWfrId(pipelineId, wfrId, repository2.SECRET_TYPE)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for secret", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	if exists {
		secretConfiguration := &bean.DeploymentConfigurationDto{
			Id:                  secretHistory.Id,
			Name:                bean.SECRET_TYPE_HISTORY_COMPONENT,
			ChildComponentNames: names,
		}
		deployedConfigurations = append(deployedConfigurations, secretConfiguration)
	}
	return deployedConfigurations, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetDeployedHistoryComponentList(pipelineId, baseConfigId int, historyComponent, historyComponentName string) ([]*bean.DeployedHistoryComponentMetadataDto, error) {
	var historyList []*bean.DeployedHistoryComponentMetadataDto
	var err error
	if historyComponent == string(bean.DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.GetDeployedHistoryList(pipelineId, baseConfigId)
	} else if historyComponent == string(bean.PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.strategyHistoryService.GetDeployedHistoryList(pipelineId, baseConfigId)
	} else if historyComponent == string(bean.CONFIGMAP_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.configMapHistoryReadService.GetDeployedHistoryList(pipelineId, baseConfigId, repository2.CONFIGMAP_TYPE, historyComponentName)
	} else if historyComponent == string(bean.SECRET_TYPE_HISTORY_COMPONENT) {
		historyList, err = impl.configMapHistoryReadService.GetDeployedHistoryList(pipelineId, baseConfigId, repository2.SECRET_TYPE, historyComponentName)
	} else {
		return nil, errors.New(fmt.Sprintf("history of %s not supported", historyComponent))
	}
	if err != nil {
		impl.logger.Errorw("error in getting deployed history component list", "err", err, "pipelineId", pipelineId, "historyComponent", historyComponent, "componentName", historyComponentName)
		return nil, err
	}
	return historyList, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetDeployedHistoryComponentDetail(ctx context.Context, pipelineId, id int, historyComponent, historyComponentName string, userHasAdminAccess bool) (*bean.HistoryDetailDto, error) {
	history := &bean.HistoryDetailDto{}
	var err error
	if historyComponent == string(bean.DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT) {
		history, err = impl.GetHistoryForDeployedTemplateById(ctx, id, pipelineId)
	} else if historyComponent == string(bean.PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT) {
		history, err = impl.strategyHistoryService.GetHistoryForDeployedStrategyById(id, pipelineId)
	} else if historyComponent == string(bean.CONFIGMAP_TYPE_HISTORY_COMPONENT) {
		history, err = impl.configMapHistoryReadService.GetHistoryForDeployedCMCSById(ctx, id, pipelineId, repository2.CONFIGMAP_TYPE, historyComponentName, userHasAdminAccess)
	} else if historyComponent == string(bean.SECRET_TYPE_HISTORY_COMPONENT) {
		history, err = impl.configMapHistoryReadService.GetHistoryForDeployedCMCSById(ctx, id, pipelineId, repository2.SECRET_TYPE, historyComponentName, userHasAdminAccess)
	} else {
		return nil, errors.New(fmt.Sprintf("history of %s not supported", historyComponent))
	}
	if err != nil {
		impl.logger.Errorw("error in getting deployed history component detail", "err", err, "pipelineId", pipelineId, "id", id, "historyComponent", historyComponent, "componentName", historyComponentName)
		return nil, err
	}
	return history, nil
}

func (impl *DeploymentTemplateHistoryReadServiceImpl) GetAllDeployedConfigurationByPipelineIdAndLatestWfrId(ctx context.Context, pipelineId int, userHasAdminAccess bool) (*bean.AllDeploymentConfigurationDetail, error) {
	//getting latest wfr from pipelineId
	wfr, err := impl.cdWorkflowRepository.FindLatestByPipelineIdAndRunnerType(pipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest deploy stage wfr by pipelineId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	deployedConfig, err := impl.GetAllDeployedConfigurationByPipelineIdAndWfrId(ctx, pipelineId, wfr.Id, userHasAdminAccess)
	if err != nil {
		impl.logger.Errorw("error in getting GetAllDeployedConfigurationByPipelineIdAndWfrId", "err", err, "pipelineID", pipelineId, "wfrId", wfr.Id)
		return nil, err
	}
	deployedConfig.WfrId = wfr.Id
	return deployedConfig, nil
}
func (impl *DeploymentTemplateHistoryReadServiceImpl) GetAllDeployedConfigurationByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int, userHasAdminAccess bool) (*bean.AllDeploymentConfigurationDetail, error) {
	//getting history of deployment template for latest deployment
	deploymentTemplateHistory, err := impl.GetDeployedHistoryByPipelineIdAndWfrId(ctx, pipelineId, wfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment template history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	//getting history of config map for latest deployment
	configMapHistory, err := impl.configMapHistoryReadService.GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(ctx, pipelineId, wfrId, repository2.CONFIGMAP_TYPE, userHasAdminAccess)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting config map history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	//getting history of secret for latest deployment
	secretHistory, err := impl.configMapHistoryReadService.GetDeployedHistoryDetailForCMCSByPipelineIdAndWfrId(ctx, pipelineId, wfrId, repository2.SECRET_TYPE, userHasAdminAccess)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting secret history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	//getting history of pipeline strategy for latest deployment
	strategyHistory, err := impl.strategyHistoryService.GetLatestDeployedHistoryByPipelineIdAndWfrId(ctx, pipelineId, wfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting strategy history by pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	allDeploymentConfigurationHistoryDetail := &bean.AllDeploymentConfigurationDetail{
		DeploymentTemplateConfig: deploymentTemplateHistory,
		ConfigMapConfig:          configMapHistory,
		SecretConfig:             secretHistory,
		StrategyConfig:           strategyHistory,
	}
	return allDeploymentConfigurationHistoryDetail, nil
}
