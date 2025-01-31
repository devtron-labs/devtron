package workflowStatus

import (
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/util"
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
	ConvertDBWorkflowStageToMap(workflowStages []*repository.WorkflowExecutionStage, wfId int, status, podStatus, message, wfType string, startTime, endTime time.Time) map[string][]*bean2.WorkflowStageDto
}

type WorkFlowStageStatusServiceImpl struct {
	logger                   *zap.SugaredLogger
	workflowStatusRepository repository.WorkflowStageRepository
	ciWorkflowRepository     pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository     pipelineConfig.CdWorkflowRepository
	transactionManager       sql.TransactionWrapper
	config                   *types.CiCdConfig
}

func NewWorkflowStageFlowStatusServiceImpl(logger *zap.SugaredLogger,
	workflowStatusRepository repository.WorkflowStageRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	transactionManager sql.TransactionWrapper,
	config *types.CiCdConfig,
) *WorkFlowStageStatusServiceImpl {
	wfStageServiceImpl := &WorkFlowStageStatusServiceImpl{
		logger:                   logger,
		workflowStatusRepository: workflowStatusRepository,
		ciWorkflowRepository:     ciWorkflowRepository,
		cdWorkflowRepository:     cdWorkflowRepository,
		transactionManager:       transactionManager,
	}
	ciCdConfig, err := types.GetCiCdConfig()
	if err != nil {
		return nil
	}
	wfStageServiceImpl.config = ciCdConfig
	return wfStageServiceImpl
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
	if impl.config.EnableWorkflowExecutionStage {
		wf.Status = cdWorkflow.WorkflowWaitingToStart
	}
	err = impl.ciWorkflowRepository.SaveWorkFlowWithTx(wf, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wf, "error", err)
		return err
	}

	if impl.config.EnableWorkflowExecutionStage {
		pipelineStageStatus := adapter.GetDefaultPipelineStatusForWorkflow(wf.Id, bean.CI_WORKFLOW_TYPE.String())
		pipelineStageStatus, err = impl.workflowStatusRepository.SaveWorkflowStages(pipelineStageStatus, tx)
		if err != nil {
			impl.logger.Errorw("error in saving workflow stages", "workflowName", wf.Name, "error", err)
			return err
		}
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

	if impl.config.EnableWorkflowExecutionStage {
		pipelineStageStatus, updatedWfStatus, updatedPodStatus := impl.getUpdatedPipelineStagesForWorkflow(wf.Id, bean.CI_WORKFLOW_TYPE.String(), wf.Status, wf.PodStatus, wf.Message, wf.PodName)
		pipelineStageStatus, err = impl.workflowStatusRepository.UpdateWorkflowStages(pipelineStageStatus, tx)
		if err != nil {
			impl.logger.Errorw("error in saving workflow stages", "workflowName", wf.Name, "error", err)
			return err
		}

		// update workflow with updated wf status
		wf.Status = updatedWfStatus
		wf.PodStatus = updatedPodStatus
	}

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

func (impl *WorkFlowStageStatusServiceImpl) getUpdatedPipelineStagesForWorkflow(wfId int, wfType string, wfStatus string, podStatus string, message string, podName string) ([]*repository.WorkflowExecutionStage, string, string) {
	// implementation
	currentWorkflowStages, err := impl.workflowStatusRepository.GetWorkflowStagesByWorkflowIdAndType(wfId, wfType)
	if err != nil {
		impl.logger.Errorw("error in getting workflow stages", "workflowId", wfId, "error", err)
		return nil, wfStatus, podStatus
	}
	if len(currentWorkflowStages) == 0 {
		return []*repository.WorkflowExecutionStage{}, wfStatus, podStatus
	}

	var currentWfDBstatus, currentPodStatus string

	if wfType == bean.CI_WORKFLOW_TYPE.String() {
		//get current status from db
		dbWf, err := impl.ciWorkflowRepository.FindById(wfId)
		if err != nil {
			impl.logger.Errorw("error in getting workflow", "wfId", wfId, "error", err)
			return nil, wfStatus, podStatus
		}
		currentWfDBstatus = dbWf.Status
		currentPodStatus = dbWf.PodStatus
	} else {
		dbWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfId)
		if err != nil {
			impl.logger.Errorw("error in getting workflow runner", "wfId", wfId, "error", err)
			return nil, wfStatus, podStatus
		}
		currentWfDBstatus = dbWfr.Status
		currentPodStatus = dbWfr.PodStatus
	}

	impl.logger.Infow("step-1", "wfId", wfId, "wfType", wfType, "wfStatus", wfStatus, "currentWfDBstatus", currentWfDBstatus, "podStatus", podStatus, "currentPodStatus", currentPodStatus, "message", message)
	currentWorkflowStages, updatedPodStatus := impl.updatePodStages(currentWorkflowStages, podStatus, currentPodStatus, message, podName)
	impl.logger.Infow("step-2", "updatedPodStatus", updatedPodStatus, "updated pod stages", currentWorkflowStages)
	currentWorkflowStages, updatedWfStatus := impl.updateWorkflowStagesToDevtronStatus(currentWorkflowStages, wfStatus, currentWfDBstatus, message, podStatus)
	impl.logger.Infow("step-3", "updatedWfStatus", updatedWfStatus, "updatedPodStatus", updatedPodStatus, "updated workflow stages", currentWorkflowStages)

	return currentWorkflowStages, updatedWfStatus, updatedPodStatus
}

func (impl *WorkFlowStageStatusServiceImpl) updatePodStages(currentWorkflowStages []*repository.WorkflowExecutionStage, podStatus string, currentPodStatus string, message string, podName string) ([]*repository.WorkflowExecutionStage, string) {
	updatedPodStatus := currentPodStatus
	if !slices.Contains(cdWorkflow.WfrTerminalStatusList, currentPodStatus) {
		updatedPodStatus = podStatus
	}
	//update pod stage status by using convertPodStatusToDevtronStatus
	for _, stage := range currentWorkflowStages {
		if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_POD {
			// add pod name in stage metadata if not empty
			if len(podName) > 0 {
				marshalledMetadata, _ := json.Marshal(map[string]string{"podName": podName})
				stage.Metadata = string(marshalledMetadata)
			}
			switch podStatus {
			case string(v1alpha1.NodePending):
				if !stage.Status.IsTerminal() {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED
				}
			case string(v1alpha1.NodeRunning):
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_UNKNOWN {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_RUNNING
					stage.StartTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			case string(v1alpha1.NodeSucceeded):
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_UNKNOWN {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_SUCCEEDED
					stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
				}
			case string(v1alpha1.NodeFailed), string(v1alpha1.NodeError):
				if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED ||
					stage.Status == bean2.WORKFLOW_STAGE_STATUS_UNKNOWN {
					stage.Message = message
					stage.Status = bean2.WORKFLOW_STAGE_STATUS_FAILED
					stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
						stage.StartTime = time.Now().Format(bean3.LayoutRFC3339)
					}
				}
			default:
				impl.logger.Errorw("unknown pod status", "podStatus", podStatus, "message", message)
				stage.Message = message
				stage.Status = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN
				stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
			}
		}
	}
	return currentWorkflowStages, updatedPodStatus
}

// Each case has 2 steps to do
// step-1: update latest status field if its not terminal already
// step-2: accordingly update stage status
func (impl *WorkFlowStageStatusServiceImpl) updateWorkflowStagesToDevtronStatus(currentWorkflowStages []*repository.WorkflowExecutionStage, wfStatus string, currentWfDBstatus, wfMessage string, podStatus string) ([]*repository.WorkflowExecutionStage, string) {
	// implementation
	updatedWfStatus := currentWfDBstatus
	//todo for switch case use enums
	switch strings.ToLower(podStatus) {
	case strings.ToLower(string(v1alpha1.NodePending)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, cdWorkflow.WorkflowWaitingToStart)

		// update workflow preparation stage and pod status if terminal
		for _, stage := range currentWorkflowStages {
			if stage.StageName == bean2.WORKFLOW_PREPARATION && !stage.Status.IsTerminal() {
				extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
				if extractedStatus != bean2.WORKFLOW_STAGE_STATUS_NOT_STARTED {
					stage.Status = extractedStatus
				}
			}

			//also mark pod status as terminal if wfstatus is terminal
			if stage.StageName == bean2.POD_EXECUTION && slices.Contains(cdWorkflow.WfrTerminalStatusList, wfStatus) {
				stage.Status = adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
				stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
			}
		}
	case strings.ToLower(string(v1alpha1.NodeRunning)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, constants.Running)

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
						extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
						if extractedStatus.IsTerminal() {
							stage.Status = extractedStatus
							stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
						}
					}
				}
			}
		}
	case strings.ToLower(string(v1alpha1.NodeSucceeded)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, cdWorkflow.WorkflowSucceeded)

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
	case strings.ToLower(string(v1alpha1.NodeFailed)), strings.ToLower(string(v1alpha1.NodeError)):
		updatedWfStatus = util.ComputeWorkflowStatus(currentWfDBstatus, wfStatus, cdWorkflow.WorkflowFailed)

		//if pod is failed, update execution stage
		for _, stage := range currentWorkflowStages {
			if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
						if extractedStatus.IsTerminal() {
							stage.Status = extractedStatus
							stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
							if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_TIMEOUT {
								updatedWfStatus = cdWorkflow.WorkflowTimedOut
							}
							if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_ABORTED {
								updatedWfStatus = cdWorkflow.WorkflowCancel
							}
						}
					}
				} else if stage.StageName == bean2.WORKFLOW_PREPARATION && !stage.Status.IsTerminal() {
					extractedStatus := adapter.ConvertStatusToDevtronStatus(wfStatus, wfMessage)
					if extractedStatus.IsTerminal() {
						stage.Status = extractedStatus
						stage.EndTime = time.Now().Format(bean3.LayoutRFC3339)
						if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_TIMEOUT {
							updatedWfStatus = cdWorkflow.WorkflowTimedOut
						}
						if extractedStatus == bean2.WORKFLOW_STAGE_STATUS_ABORTED {
							updatedWfStatus = cdWorkflow.WorkflowCancel
						}
					}
				}
			}
		}
	default:
		impl.logger.Errorw("unknown pod status", "podStatus", podStatus)
		//mark workflow stage status as unknown
		for _, stage := range currentWorkflowStages {
			if stage.StatusType == bean2.WORKFLOW_STAGE_STATUS_TYPE_WORKFLOW {
				//mark execution stage as completed
				if stage.StageName == bean2.WORKFLOW_EXECUTION {
					if stage.Status == bean2.WORKFLOW_STAGE_STATUS_RUNNING {
						stage.Status = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN
						updatedWfStatus = bean2.WORKFLOW_STAGE_STATUS_UNKNOWN.ToString()
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
			resp[wfId] = impl.ConvertDBWorkflowStageToMap(preCdDbData, wfId, wf.Status, wf.PodStatus, wf.Message, wf.WorkflowType, wf.StartedOn, wf.FinishedOn)
		} else if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_POST.String() {
			resp[wfId] = impl.ConvertDBWorkflowStageToMap(postCdDbData, wfId, wf.Status, wf.PodStatus, wf.Message, wf.WorkflowType, wf.StartedOn, wf.FinishedOn)
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
	if impl.config.EnableWorkflowExecutionStage {
		wfr.Status = cdWorkflow.WorkflowWaitingToStart
	}
	wfr, err = impl.cdWorkflowRepository.SaveWorkFlowRunnerWithTx(wfr, tx)
	if err != nil {
		impl.logger.Errorw("error in saving workflow", "payload", wfr, "error", err)
		return wfr, err
	}

	if impl.config.EnableWorkflowExecutionStage {
		pipelineStageStatus := adapter.GetDefaultPipelineStatusForWorkflow(wfr.Id, wfr.WorkflowType.String())
		pipelineStageStatus, err = impl.workflowStatusRepository.SaveWorkflowStages(pipelineStageStatus, tx)
		if err != nil {
			impl.logger.Errorw("error in saving workflow stages", "workflowName", wfr.Name, "error", err)
			return wfr, err
		}
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

	if impl.config.EnableWorkflowExecutionStage {
		if wfr.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE || wfr.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
			pipelineStageStatus, updatedWfStatus, updatedPodStatus := impl.getUpdatedPipelineStagesForWorkflow(wfr.Id, wfr.WorkflowType.String(), wfr.Status, wfr.PodStatus, wfr.Message, wfr.PodName)
			pipelineStageStatus, err = impl.workflowStatusRepository.UpdateWorkflowStages(pipelineStageStatus, tx)
			if err != nil {
				impl.logger.Errorw("error in saving workflow stages", "workflowName", wfr.Name, "error", err)
				return err
			}
			wfr.Status = updatedWfStatus
			wfr.PodStatus = updatedPodStatus
		} else {
			impl.logger.Infow("workflow type not supported to update stage data", "workflowName", wfr.Name, "workflowType", wfr.WorkflowType)
		}
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

func (impl *WorkFlowStageStatusServiceImpl) ConvertDBWorkflowStageToMap(workflowStages []*repository.WorkflowExecutionStage, wfId int, status, podStatus, message, wfType string, startTime, endTime time.Time) map[string][]*bean2.WorkflowStageDto {
	wfMap := make(map[string][]*bean2.WorkflowStageDto)
	foundInDb := false
	if !impl.config.EnableWorkflowExecutionStage {
		// if flag is not enabled then return empty map
		return map[string][]*bean2.WorkflowStageDto{}
	}
	for _, wfStage := range workflowStages {
		if wfStage.WorkflowId == wfId {
			wfMap[wfStage.StatusType.ToString()] = append(wfMap[wfStage.StatusType.ToString()], adapter.ConvertDBWorkflowStageToDto(wfStage))
			foundInDb = true
		}
	}

	if !foundInDb {
		//for old data where stages are not saved in db return empty map
		return map[string][]*bean2.WorkflowStageDto{}
	}

	return wfMap

}
