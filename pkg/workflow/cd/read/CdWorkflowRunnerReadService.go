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

	pipelineIdToCiPipelineIdMap := make(map[int]int)
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

	cdWorfklowLatestMap := make(map[int][]bean2.WorkflowType)
	for _, item := range cdWorkflowLatest {
		if _, ok := cdWorfklowLatestMap[item.PipelineId]; !ok {
			cdWorfklowLatestMap[item.PipelineId] = make([]bean2.WorkflowType, 0)
		}
		cdWorfklowLatestMap[item.PipelineId] = append(cdWorfklowLatestMap[item.PipelineId], bean2.WorkflowType(item.WorkflowType))
	}

	pipelineStage, err := impl.pipelineStageRepository.GetAllCdStagesByCdPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline stages", "pipelineId", pipelineIds, "err", err)
		return nil, err
	}
	pipelineStageMap := make(map[int][]bean2.WorkflowType)
	for _, item := range pipelineStage {
		if _, ok := pipelineStageMap[item.CdPipelineId]; !ok {
			pipelineStageMap[item.CdPipelineId] = make([]bean2.WorkflowType, 0)
		}
		if item.Type == repository2.PIPELINE_STAGE_TYPE_PRE_CD {
			pipelineStageMap[item.CdPipelineId] = append(pipelineStageMap[item.CdPipelineId], bean2.CD_WORKFLOW_TYPE_PRE)
		} else if item.Type == repository2.PIPELINE_STAGE_TYPE_POST_CD {
			pipelineStageMap[item.CdPipelineId] = append(pipelineStageMap[item.CdPipelineId], bean2.CD_WORKFLOW_TYPE_POST)
		}
	}

	// calculating all the pipelines not present in the index table cdWorkflowLatest
	absentPipelineIds := make([]int, 0)
	for _, item := range pipelines {
		var isPreCDConfigured, isPostCDConfigured bool
		if configuredStages, ok := pipelineStageMap[item.Id]; ok {
			for _, stage := range configuredStages {
				if stage == bean2.CD_WORKFLOW_TYPE_PRE {
					isPreCDConfigured = true
				} else if stage == bean2.CD_WORKFLOW_TYPE_POST {
					isPostCDConfigured = true
				}
			}
		}

		if _, ok := cdWorfklowLatestMap[item.Id]; !ok {
			absentPipelineIds = append(absentPipelineIds, item.Id)
		} else {
			isPreCDStageAbsent, isPostCdStageAbsent, isDeployStageAbsent := true, true, true
			for _, stage := range cdWorfklowLatestMap[item.Id] {
				switch stage {
				case bean2.CD_WORKFLOW_TYPE_PRE:
					isPreCDStageAbsent = false
				case bean2.CD_WORKFLOW_TYPE_POST:
					isPostCdStageAbsent = false
				case bean2.CD_WORKFLOW_TYPE_DEPLOY:
					isDeployStageAbsent = false
				}
			}
			if isDeployStageAbsent || (isPreCDConfigured && isPreCDStageAbsent) || (isPostCDConfigured && isPostCdStageAbsent) {
				absentPipelineIds = append(absentPipelineIds, item.Id)
			}
		}
	}
	if len(absentPipelineIds) > 0 {
		remainingRunners, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(absentPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in fetching all cd stages latest entity", "pipelineIds", absentPipelineIds, "err", err)
			return nil, err
		}
		result = append(result, remainingRunners...)
	}
	return result, nil
}
