package workflowStatus

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"slices"
	"strings"
	"time"
)

type WorkFlowStageStatusService interface {
	SaveCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error
	UpdateCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error
	SaveCDWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) (*pipelineConfig.CdWorkflowRunner, error)
	UpdateCdWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) error
	GetCiWorkflowStagesByWorkflowIds(wfIds []int) ([]*repository.WorkflowExecutionStage, error)
	GetPrePostWorkflowStagesByWorkflowIdAndType(wfId int, wfType string) ([]*repository.WorkflowExecutionStage, error)
	GetPrePostWorkflowStagesByWorkflowRunnerIdsList(wfIdWfTypeMap map[int]bean4.CdWorkflowWithArtifact) (map[int]map[string][]*bean2.WorkflowStageDto, error)
}

type WorkFlowStageStatusServiceImpl struct {
	logger                   *zap.SugaredLogger
	workflowStatusRepository repository.WorkflowStageRepository
	ciWorkflowRepository     pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository     pipelineConfig.CdWorkflowRepository
	transactionManager       sql.TransactionWrapper
}

func NewWorkflowStageFlowStatusServiceImpl(logger *zap.SugaredLogger,
	workflowStatusRepository repository.WorkflowStageRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	transactionManager sql.TransactionWrapper) *WorkFlowStageStatusServiceImpl {
	return &WorkFlowStageStatusServiceImpl{
		logger:                   logger,
		workflowStatusRepository: workflowStatusRepository,
		ciWorkflowRepository:     ciWorkflowRepository,
		cdWorkflowRepository:     cdWorkflowRepository,
		transactionManager:       transactionManager,
	}
}

func (impl *WorkFlowStageStatusServiceImpl) SaveCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "workflowName", wf.Name, "error", err)
		return err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil {
			impl.logger.Errorw("error in rolling back transaction", "workflowName", wf.Name, "error", dbErr)
		}
	}()
	wf.Status = cdWorkflow.WorkflowWaitingToStart
	err = impl.ciWorkflowRepository.SaveWorkFlowWithTx(wf, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wf, "error", err)
		return err
	}
	pipelineStageStatus := adapter.GetDefaultPipelineStatusForWorkflow(wf.Id, bean.CI_WORKFLOW_TYPE.String())
	pipelineStageStatus, err = impl.workflowStatusRepository.SaveWorkflowStages(pipelineStageStatus, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow stages", "workflowName", wf.Name, "error", err)
		return err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "workflowName", wf.Name, "error", err)
		return err
	}
	return nil

}

func (impl *WorkFlowStageStatusServiceImpl) UpdateCiWorkflowWithStage(wf *pipelineConfig.CiWorkflow) error {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "workflowName", wf.Name, "error", err)
		return err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil {
			impl.logger.Errorw("error in rolling back transaction", "workflowName", wf.Name, "error", dbErr)
		}
	}()

	if checkIfWorkflowIsWaitingToStart(wf.Status, wf.PodStatus) {
		wf.Status = cdWorkflow.WorkflowWaitingToStart
	}

	pipelineStageStatus, updatedWfStatus := impl.getUpdatedPipelineStagesForWorkflow(wf.Id, bean.CI_WORKFLOW_TYPE.String(), wf.Status, wf.PodStatus, wf.Message, wf.PodName)
	pipelineStageStatus, err = impl.workflowStatusRepository.UpdateWorkflowStages(pipelineStageStatus, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow stages", "workflowName", wf.Name, "error", err)
		return err
	}

	// update workflow with updated wf status
	wf.Status = updatedWfStatus
	err = impl.ciWorkflowRepository.UpdateWorkFlowWithTx(wf, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wf, "error", err)
		return err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "workflowName", wf.Name, "error", err)
		return err
	}
	return nil

}

func (impl *WorkFlowStageStatusServiceImpl) getUpdatedPipelineStagesForWorkflow(wfId int, wfType string, wfStatus string, podStatus string, message string, podName string) ([]*repository.WorkflowExecutionStage, string) {
	// implementation
	updatedWfStatus := wfStatus
	currentWorkflowStages, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdAndType(wfId, wfType)
	if err != nil {
		impl.logger.Errorw("error in getting workflow stages", "workflowId", wfId, "error", err)
		return nil, updatedWfStatus
	}
	if len(currentWorkflowStages) == 0 {
		return []*repository.WorkflowExecutionStage{}, updatedWfStatus
	}
	impl.logger.Infow("step-1", "wfId", wfId, "wfType", wfType, "wfStatus", wfStatus, "podStatus", podStatus, "message", message)
	currentWorkflowStages = impl.updatePodStages(currentWorkflowStages, podStatus, message, podName)
	impl.logger.Infow("step-2", "updated pod stages", currentWorkflowStages)
	currentWorkflowStages, updatedWfStatus = impl.updateWorkflowStagesToDevtronStatus(currentWorkflowStages, wfStatus, message, podStatus)
	impl.logger.Infow("step-3", "updated workflow stages", currentWorkflowStages)

	if len(updatedWfStatus) == 0 {
		//case when current wfStatus and received wfStatus both are terminal, then keep the status as it is in DB
		dbWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfId)
		if err != nil {
			impl.logger.Errorw("error in getting workflow runner", "wfId", wfId, "error", err)
			return nil, updatedWfStatus
		}
		updatedWfStatus = dbWfr.Status
	}

	return currentWorkflowStages, updatedWfStatus
}

func (impl *WorkFlowStageStatusServiceImpl) updatePodStages(currentWorkflowStages []*repository.WorkflowExecutionStage, podStatus string, message string, podName string) []*repository.WorkflowExecutionStage {
	//update pod stage status by using convertPodStatusToDevtronStatus
	for _, stage := range currentWorkflowStages {
		if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_POD {
			// add pod name in stage metadata if not empty
			if len(podName) > 0 {
				marshalledMetadata, _ := json.Marshal(map[string]string{"podName": podName})
				stage.Metadata = string(marshalledMetadata)
			}
			switch podStatus {
			case "Pending":
				//only update message as we create this entry when pod is created
				stage.Message = message
			case "Running":
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_RUNNING
					stage.StartTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			case "Succeeded":
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_SUCCEEDED
					stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			case "Failed", "Error":
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {

					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_FAILED
					stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			default:
				impl.logger.Errorw("unknown pod status", "podStatus", podStatus)
				stage.Message = message
				stage.Status = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN
				stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
			}
		}
	}
	return currentWorkflowStages
}

// Each case has 2 steps to do
// step-1: update latest status field if its not terminal already
// step-2: accordingly update stage status
func (impl *WorkFlowStageStatusServiceImpl) updateWorkflowStagesToDevtronStatus(currentWorkflowStages []*repository.WorkflowExecutionStage, wfStatus string, wfMessage string, podStatus string) ([]*repository.WorkflowExecutionStage, string) {
	// implementation
	updatedWfStatus := ""
	//todo for switch case use enums
	switch strings.ToLower(podStatus) {
	case "pending":
		if !slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
			updatedWfStatus = cdWorkflow.WorkflowWaitingToStart
		}

		// update workflow preparation stage and pod status if terminal
		for _, stage := range currentWorkflowStages {
			if stage.StageName == bean2.WORKFLOW_PREPARATION && !stage.Status.IsTerminal() {
				extractedStatus := adapter.ConvertWfStatusToDevtronStatus(wfStatus, wfMessage)
				if extractedStatus != bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
					stage.Status = extractedStatus
				}
			}

			//also mark pod status as terminal if wfstatus is terminal
			if stage.StageName == bean2.POD_EXECUTION && slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
				stage.Status = adapter.ConvertWfStatusToDevtronStatus(wfStatus, wfMessage)
				stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
			}
		}
	case "running":
		if !slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
			updatedWfStatus = constants.Running
		}
		//if pod is running, update preparation and execution stages
		for _, stage := range currentWorkflowStages {
			if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark preparation stage as completed
				if stage.StageName == bean2.WORKFLOW_PREPARATION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_SUCCEEDED
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}

				//mark execution stage as started
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_RUNNING
						stage.StartTime = time.Now().Format(bean3.LayoutRFC3339)
					} else if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						extractedStatus := adapter.ConvertWfStatusToDevtronStatus(wfStatus, wfMessage)
						if extractedStatus.IsTerminal() {
							stage.Status = extractedStatus
							stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
						}
					}
				}
			}
		}
	case "succeeded":
		if !slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
			updatedWfStatus = cdWorkflow.WorkflowSucceeded
		}
		//if pod is succeeded, update execution stage
		for _, stage := range currentWorkflowStages {
			if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_SUCCEEDED
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}
			}
		}
	case "failed", "error":
		if !slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
			updatedWfStatus = cdWorkflow.WorkflowFailed
		}

		//if pod is failed, update execution stage
		for _, stage := range currentWorkflowStages {
			if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						extractedStatus := adapter.ConvertWfStatusToDevtronStatus(wfStatus, wfMessage)
						if extractedStatus.IsTerminal() {
							stage.Status = extractedStatus
							stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
							if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_TIMEOUT {
								updatedWfStatus = cdWorkflow.WorkflowTimedOut
							}
						}
					}
				} else if stage.StageName == bean2.WORKFLOW_PREPARATION && !stage.Status.IsTerminal() {
					extractedStatus := adapter.ConvertWfStatusToDevtronStatus(wfStatus, wfMessage)
					if extractedStatus.IsTerminal() {
						stage.Status = extractedStatus
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}
			}
		}
	default:
		impl.logger.Errorw("unknown pod status", "podStatus", podStatus)
		//mark workflow stage status as unknown and end it
		for _, stage := range currentWorkflowStages {
			if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}
			}
		}
	}

	return currentWorkflowStages, updatedWfStatus
}

func (impl *WorkFlowStageStatusServiceImpl) GetCiWorkflowStagesByWorkflowIds(wfIds []int) ([]*repository.WorkflowExecutionStage, error) {
	// implementation

	dbData, err := impl.workflowStatusRepository.GetCiWorkflowStagesByWorkflowIds(wfIds)
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow stages", "error", err)
		return nil, err
	}
	if len(dbData) == 0 {
		return []*repository.WorkflowExecutionStage{}, nil
	} else {
		return dbData, nil
	}
}

func (impl *WorkFlowStageStatusServiceImpl) GetPrePostWorkflowStagesByWorkflowIdAndType(wfId int, wfType string) ([]*repository.WorkflowExecutionStage, error) {
	// implementation

	dbData, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdAndWtype(wfId, wfType)
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow stages", "error", err)
		return nil, err
	}
	if len(dbData) == 0 {
		return []*repository.WorkflowExecutionStage{}, nil
	} else {
		return dbData, nil
	}
}

func (impl *WorkFlowStageStatusServiceImpl) GetPrePostWorkflowStagesByWorkflowRunnerIdsList(wfIdWfTypeMap map[int]bean4.CdWorkflowWithArtifact) (map[int]map[string][]*bean2.WorkflowStageDto, error) {
	// implementation
	resp := map[int]map[string][]*bean2.WorkflowStageDto{}
	if len(wfIdWfTypeMap) == 0 {
		return resp, nil
	}
	//first create a map of pre-runner ids and post-runner ids
	prePostRunnerIds := map[string][]int{}
	for wfId, wf := range wfIdWfTypeMap {
		if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE.String() {
			prePostRunnerIds[bean.CD_WORKFLOW_TYPE_PRE.String()] = append(prePostRunnerIds[bean.CD_WORKFLOW_TYPE_PRE.String()], wfId)
		} else if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_POST.String() {
			prePostRunnerIds[bean.CD_WORKFLOW_TYPE_POST.String()] = append(prePostRunnerIds[bean.CD_WORKFLOW_TYPE_POST.String()], wfId)
		}
	}

	preCdDbData, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdsAndWtype(prePostRunnerIds[bean.CD_WORKFLOW_TYPE_PRE.String()], bean.CD_WORKFLOW_TYPE_PRE.String())
	if err != nil {
		impl.logger.Errorw("error in getting pre-ci workflow stages", "error", err)
		return resp, err
	}
	//do the above for post cd
	postCdDbData, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdsAndWtype(prePostRunnerIds[bean.CD_WORKFLOW_TYPE_POST.String()], bean.CD_WORKFLOW_TYPE_POST.String())
	if err != nil {
		impl.logger.Errorw("error in getting post-ci workflow stages", "error", err)
		return resp, err
	}
	//iterate over prePostRunnerIds and create response structure using ConvertDBWorkflowStageToMap function
	for wfId, wf := range wfIdWfTypeMap {
		if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE.String() {
			resp[wfId] = adapter.ConvertDBWorkflowStageToMap(preCdDbData, wfId, wf.Status, wf.PodStatus, wf.Message, wf.WorkflowType, wf.StartedOn, wf.FinishedOn)
		} else if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_POST.String() {
			resp[wfId] = adapter.ConvertDBWorkflowStageToMap(postCdDbData, wfId, wf.Status, wf.PodStatus, wf.Message, wf.WorkflowType, wf.StartedOn, wf.FinishedOn)
		}
	}
	return resp, nil
}

func (impl *WorkFlowStageStatusServiceImpl) SaveCDWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) (*pipelineConfig.CdWorkflowRunner, error) {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "workflowName", wfr.Name, "error", err)
		return wfr, err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil {
			impl.logger.Errorw("error in rolling back transaction", "workflowName", wfr.Name, "error", dbErr)
		}
	}()
	wfr.Status = cdWorkflow.WorkflowWaitingToStart
	wfr, err = impl.cdWorkflowRepository.SaveWorkFlowRunnerWithTx(wfr, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wfr, "error", err)
		return wfr, err
	}
	pipelineStageStatus := adapter.GetDefaultPipelineStatusForWorkflow(wfr.Id, wfr.WorkflowType.String())
	pipelineStageStatus, err = impl.workflowStatusRepository.SaveWorkflowStages(pipelineStageStatus, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow stages", "workflowName", wfr.Name, "error", err)
		return wfr, err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "workflowName", wfr.Name, "error", err)
		return wfr, err
	}
	return wfr, nil
}

func (impl *WorkFlowStageStatusServiceImpl) UpdateCdWorkflowRunnerWithStage(wfr *pipelineConfig.CdWorkflowRunner) error {
	// implementation
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "workflowName", wfr.Name, "error", err)
		return err
	}

	defer func() {
		dbErr := impl.transactionManager.RollbackTx(tx)
		if dbErr != nil {
			impl.logger.Errorw("error in rolling back transaction", "workflowName", wfr.Name, "error", dbErr)
		}
	}()

	if (wfr.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE || wfr.WorkflowType == bean.CD_WORKFLOW_TYPE_POST) && checkIfWorkflowIsWaitingToStart(wfr.Status, wfr.PodStatus) {
		wfr.Status = cdWorkflow.WorkflowWaitingToStart
	}

	if wfr.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE || wfr.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		pipelineStageStatus, updatedWfStatus := impl.getUpdatedPipelineStagesForWorkflow(wfr.Id, wfr.WorkflowType.String(), wfr.Status, wfr.PodStatus, wfr.Message, wfr.PodName)
		pipelineStageStatus, err = impl.workflowStatusRepository.UpdateWorkflowStages(pipelineStageStatus, tx)
		if err != nil {
			impl.logger.Errorw("error in saving workflow stages", "workflowName", wfr.Name, "error", err)
			return err
		}
		wfr.Status = updatedWfStatus
	} else {
		impl.logger.Infow("workflow type not supported to update stage data", "workflowName", wfr.Name, "workflowType", wfr.WorkflowType)
	}

	//update workflow runner now with updatedWfStatus if applicable
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunnerWithTx(wfr, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wfr, "error", err)
		return err
	}

	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "workflowName", wfr.Name, "error", err)
		return err
	}
	return nil

}

func checkIfWorkflowIsWaitingToStart(wfStatus string, podStatus string) bool {
	// implementation
	return strings.ToLower(podStatus) == "pending" && (strings.ToLower(wfStatus) == strings.ToLower(cdWorkflow.WorkflowWaitingToStart) || strings.ToLower(wfStatus) == "running" || strings.ToLower(wfStatus) == "starting")
}
