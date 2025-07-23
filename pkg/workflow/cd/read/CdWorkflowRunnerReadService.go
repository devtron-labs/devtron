package read

import (
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/adapter"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/workflowStatusLatest"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CdWorkflowRunnerReadService interface {
	FindWorkflowRunnerById(wfrId int) (*bean.CdWorkflowRunnerDto, error)
	CheckIfWfrLatest(wfrId, pipelineId int) (isLatest bool, err error)
	GetWfrStatusForLatestRunners(pipelineIds []int, pipelines []*pipelineConfig.Pipeline) ([]*pipelineConfig.CdWorkflowStatus, error)
}

type CdWorkflowRunnerReadServiceImpl struct {
	logger                      *zap.SugaredLogger
	cdWorkflowRepository        pipelineConfig.CdWorkflowRepository
	WorkflowStatusLatestService workflowStatusLatest.WorkflowStatusLatestService
	pipelineStageRepository     repository2.PipelineStageRepository
}

func NewCdWorkflowRunnerReadServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	WorkflowStatusLatestService workflowStatusLatest.WorkflowStatusLatestService,
	pipelineStageRepository repository2.PipelineStageRepository) *CdWorkflowRunnerReadServiceImpl {
	return &CdWorkflowRunnerReadServiceImpl{
		logger:                      logger,
		cdWorkflowRepository:        cdWorkflowRepository,
		WorkflowStatusLatestService: WorkflowStatusLatestService,
		pipelineStageRepository:     pipelineStageRepository,
	}
}

func (impl *CdWorkflowRunnerReadServiceImpl) FindWorkflowRunnerById(wfrId int) (*bean.CdWorkflowRunnerDto, error) {
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow runner by id", "err", err, "id", wfrId)
		return nil, err
	}
	return adapter.ConvertCdWorkflowRunnerDbObjToDto(cdWfr), nil

}

func (impl *CdWorkflowRunnerReadServiceImpl) CheckIfWfrLatest(wfrId, pipelineId int) (isLatest bool, err error) {
	isLatest, err = impl.cdWorkflowRepository.IsLatestCDWfr(wfrId, pipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in checking latest cd workflow runner", "err", err)
		return false, err
	}
	return isLatest, nil
}

func (impl *CdWorkflowRunnerReadServiceImpl) GetWfrStatusForLatestRunners(pipelineIds []int, pipelines []*pipelineConfig.Pipeline) ([]*pipelineConfig.CdWorkflowStatus, error) {
	// fetching the latest pipeline from the index table - cdWorkflowLatest
	var result []*pipelineConfig.CdWorkflowStatus
	cdWorkflowLatest, err := impl.WorkflowStatusLatestService.GetCdWorkflowLatestByPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting latest by pipelineId", "pipelineId", pipelineIds, "err", err)
		return nil, err
	}

	var pipelineIdToCiPipelineIdMap map[int]int
	for _, item := range pipelines {
		pipelineIdToCiPipelineIdMap[item.Id] = item.CiPipelineId
	}

	for _, item := range cdWorkflowLatest {
		result = append(result, &pipelineConfig.CdWorkflowStatus{
			CiPipelineId: pipelineIdToCiPipelineIdMap[item.PipelineId],
			PipelineId:   item.PipelineId,
			WorkflowType: item.WorkflowType,
			WfrId:        item.WorkflowRunnerId,
		})
	}

	var cdWorfklowLatestMap map[int]map[bean2.WorkflowType]bool
	for _, item := range cdWorkflowLatest {
		if _, ok := cdWorfklowLatestMap[item.PipelineId]; !ok {
			cdWorfklowLatestMap[item.PipelineId] = make(map[bean2.WorkflowType]bool)
		}
		cdWorfklowLatestMap[item.PipelineId][bean2.WorkflowType(item.WorkflowType)] = true
	}

	pipelineStage, err := impl.pipelineStageRepository.GetAllCdStagesByCdPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline stages", "pipelineId", pipelineIds, "err", err)
		return nil, err
	}
	pipelineStageMap := make(map[int]map[bean2.WorkflowType]bool)
	for _, item := range pipelineStage {
		if _, ok := pipelineStageMap[item.CdPipelineId]; !ok {
			pipelineStageMap[item.CdPipelineId] = make(map[bean2.WorkflowType]bool)
		}
		if item.Type == repository2.PIPELINE_STAGE_TYPE_PRE_CD {
			pipelineStageMap[item.CdPipelineId][bean2.CD_WORKFLOW_TYPE_PRE] = true
		} else if item.Type == repository2.PIPELINE_STAGE_TYPE_POST_CD {
			pipelineStageMap[item.CdPipelineId][bean2.CD_WORKFLOW_TYPE_POST] = true
		}
	}

	// calculating all the pipelines not present in the index table cdWorkflowLatest
	var pipelinesAbsentInCache map[int]bean2.WorkflowType
	for _, item := range pipelines {
		if _, ok := cdWorfklowLatestMap[item.Id]; !ok {
			pipelinesAbsentInCache[item.Id] = bean2.CD_WORKFLOW_TYPE_PRE
			pipelinesAbsentInCache[item.Id] = bean2.CD_WORKFLOW_TYPE_DEPLOY
			pipelinesAbsentInCache[item.Id] = bean2.CD_WORKFLOW_TYPE_POST
		} else {
			if _, ok := pipelineStageMap[item.Id][bean2.CD_WORKFLOW_TYPE_PRE]; ok {
				if val, ok := cdWorfklowLatestMap[item.Id][bean2.CD_WORKFLOW_TYPE_PRE]; !ok || !val {
					pipelinesAbsentInCache[item.Id] = bean2.CD_WORKFLOW_TYPE_PRE
				}
			}
			if _, ok := pipelineStageMap[item.Id][bean2.CD_WORKFLOW_TYPE_POST]; ok {
				if val, ok := cdWorfklowLatestMap[item.Id][bean2.CD_WORKFLOW_TYPE_POST]; !ok || !val {
					pipelinesAbsentInCache[item.Id] = bean2.CD_WORKFLOW_TYPE_POST
				}
			}
			if val, ok := cdWorfklowLatestMap[item.Id][bean2.CD_WORKFLOW_TYPE_DEPLOY]; !ok || !val {
				pipelinesAbsentInCache[item.Id] = bean2.CD_WORKFLOW_TYPE_POST
			}
		}
	}
	if len(pipelinesAbsentInCache) > 0 {
		remainingRunners, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelinesAbsentInCache)
		if err != nil {
			impl.logger.Errorw("error in fetching all cd stages latest entity", "pipelinesAbsentInCache", pipelinesAbsentInCache, "err", err)
			return nil, err
		}
		result = append(result, remainingRunners...)
	}
	return result, nil
}
