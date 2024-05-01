package read

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appWorkflow/bean"
	"go.uber.org/zap"
)

type AppWorkflowDataReadService interface {
	FindCDPipelineIdsAndCdPipelineIdToWfIdMapping(wfIds []int) ([]int, map[int]int, error)
	FindWorkflowComponentsToAppIdMapping(appId int, wfIds []int) (map[int]*bean.WorkflowComponents, error)
}

type AppWorkflowDataReadServiceImpl struct {
	appWorkflowRepository appWorkflow.AppWorkflowRepository
	logger                *zap.SugaredLogger
}

func NewAppWorkflowDataReadServiceImpl(
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	logger *zap.SugaredLogger) *AppWorkflowDataReadServiceImpl {
	return &AppWorkflowDataReadServiceImpl{
		appWorkflowRepository: appWorkflowRepository,
		logger:                logger,
	}
}

func (impl AppWorkflowDataReadServiceImpl) FindCDPipelineIdsAndCdPipelineIdToWfIdMapping(wfIds []int) ([]int, map[int]int, error) {

	wfMappings, err := impl.appWorkflowRepository.FindByWorkflowIds(wfIds)
	if err != nil {
		impl.logger.Errorw("error in fetching all workflow mappings by workflowId", "workflowIds", wfIds, "err", err)
		return nil, nil, bean.WorkflowMappingsNotFoundError{WorkflowIds: wfIds}
	}

	cdPipelineIds := make([]int, 0)
	cdPipelineIdToWorkflowIdMapping := make(map[int]int)
	for _, wfMapping := range wfMappings {
		if wfMapping.Type == appWorkflow.CDPIPELINE {
			cdPipelineIds = append(cdPipelineIds, wfMapping.ComponentId)
			cdPipelineIdToWorkflowIdMapping[wfMapping.ComponentId] = wfMapping.AppWorkflowId
		}
	}
	return cdPipelineIds, cdPipelineIdToWorkflowIdMapping, nil
}

func (impl AppWorkflowDataReadServiceImpl) FindWorkflowComponentsToAppIdMapping(appId int, wfIds []int) (map[int]*bean.WorkflowComponents, error) {
	appWorkflowMappings, err := impl.appWorkflowRepository.FindFilteredWFMappingsByAppId(appId, wfIds...)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching all workflow mappings by appId", "appIds", appId, "err", err)
		return nil, err
	}
	wfComponentDetails := make(map[int]*bean.WorkflowComponents)
	for _, appWfMapping := range appWorkflowMappings {
		if _, ok := wfComponentDetails[appWfMapping.AppWorkflowId]; !ok {
			wfComponentDetails[appWfMapping.AppWorkflowId] = &bean.WorkflowComponents{}
		}
		if appWfMapping.Type == bean.CI_PIPELINE_TYPE {
			wfComponentDetails[appWfMapping.AppWorkflowId].CiPipelineId = appWfMapping.ComponentId
		} else if appWfMapping.Type == bean.WEBHOOK_TYPE {
			wfComponentDetails[appWfMapping.AppWorkflowId].ExternalCiPipelineId = appWfMapping.ComponentId
		} else if appWfMapping.Type == bean.CD_PIPELINE_TYPE {
			wfComponentDetails[appWfMapping.AppWorkflowId].CdPipelineIds = append(wfComponentDetails[appWfMapping.AppWorkflowId].CdPipelineIds, appWfMapping.ComponentId)
		}
	}
	return wfComponentDetails, nil
}
