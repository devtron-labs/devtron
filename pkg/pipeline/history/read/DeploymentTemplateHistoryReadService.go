package read

import (
	"context"
	"errors"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/adaptors"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
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
}

type DeploymentTemplateHistoryReadServiceImpl struct {
	logger                              *zap.SugaredLogger
	deploymentTemplateHistoryRepository repository2.DeploymentTemplateHistoryRepository
	scopedVariableManager               variables.ScopedVariableManager
}

func NewDeploymentTemplateHistoryReadServiceImpl(
	logger *zap.SugaredLogger,
	deploymentTemplateHistoryRepository repository2.DeploymentTemplateHistoryRepository,
	scopedVariableManager variables.ScopedVariableManager,
) *DeploymentTemplateHistoryReadServiceImpl {
	return &DeploymentTemplateHistoryReadServiceImpl{
		logger:                              logger,
		deploymentTemplateHistoryRepository: deploymentTemplateHistoryRepository,
		scopedVariableManager:               scopedVariableManager,
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
