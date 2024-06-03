/*
 * Copyright (c) 2024. Devtron Inc.
 */

package cd

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CdWorkflowService interface {
	CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error)
	UpdateWorkFlow(dto *bean.CdWorkflowDto) error
	CreateBulkCdWorkflow(tx *pg.Tx, cdWorkflowDtos []*bean.CdWorkflowDto, creationTime time.Time) (map[string]int, error)
}

type CdWorkflowServiceImpl struct {
	logger               *zap.SugaredLogger
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewCdWorkflowServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *CdWorkflowServiceImpl {
	return &CdWorkflowServiceImpl{
		logger:               logger,
		cdWorkflowRepository: cdWorkflowRepository,
	}
}

func (impl *CdWorkflowServiceImpl) CheckIfLatestWf(pipelineId, cdWfId int) (latest bool, err error) {
	latest, err = impl.cdWorkflowRepository.IsLatestWf(pipelineId, cdWfId)
	if err != nil {
		impl.logger.Errorw("error in checking if wf is latest", "pipelineId", pipelineId, "cdWfId", cdWfId, "err", err)
		return false, err
	}
	return latest, nil
}

func (impl *CdWorkflowServiceImpl) UpdateWorkFlow(dto *bean.CdWorkflowDto) error {
	dbObj := adapter.ConvertCdWorkflowDtoToDbObj(dto)
	err := impl.cdWorkflowRepository.UpdateWorkFlow(dbObj)
	if err != nil {
		impl.logger.Errorw("error in updating workflow", "err", err, "req", dto)
		return err
	}
	return nil
}

func (impl *CdWorkflowServiceImpl) CreateBulkCdWorkflow(tx *pg.Tx, cdWorkflowDtos []*bean.CdWorkflowDto, creationTime time.Time) (map[string]int, error) {
	cdWorkFlows := make([]*pipelineConfig.CdWorkflow, 0, len(cdWorkflowDtos))
	for _, dto := range cdWorkflowDtos {
		cdWorkFlows = append(cdWorkFlows, adapter.ConvertCdWorkflowDtoToDbObjForCreation(dto, creationTime))
	}
	err := impl.cdWorkflowRepository.BulkSaveWorkflow(tx, cdWorkFlows)
	if err != nil {
		impl.logger.Errorw("error encountered in CreateBulkCdWorkflow", "cdWorkFlows", cdWorkFlows, "err", err)
		return nil, err
	}
	return util.GetPipelineArtifactIdKeyVsCdWorkflowIdMap(cdWorkFlows), nil
}
