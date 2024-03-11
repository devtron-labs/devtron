package read

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
	"net/http"
)

type AppWorkflowDataReadService interface {
	FindCDPipelineIdsAndCdPipelineIdTowfIdMapping(wfIds []int) ([]int, map[int]int, error)
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

func (impl AppWorkflowDataReadServiceImpl) FindCDPipelineIdsAndCdPipelineIdTowfIdMapping(wfIds []int) ([]int, map[int]int, error) {
	wfMappings, err := impl.appWorkflowRepository.FindByWorkflowIds(wfIds)
	if err != nil {
		impl.logger.Errorw("error in fetching all workflow mappings by workflowId", "workflowIds", wfIds, "err", err)
		return nil, nil, &util.ApiError{
			HttpStatusCode:  http.StatusUnprocessableEntity,
			InternalMessage: "workflowMappings not found for given workflowId",
			UserMessage:     "workflowMappings not found for given workflowId",
		}
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
