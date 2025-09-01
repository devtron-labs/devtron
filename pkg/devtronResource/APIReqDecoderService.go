package devtronResource

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/util"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

// APIReqDecoderService is common service used for getting decoded devtronResource and history related api params.
type APIReqDecoderService interface {
	GetFilterCriteriaParamsForDeploymentHistory(filterCriteria []string) (*bean2.DeploymentHistoryGetReqDecoderBean, error)
}

type APIReqDecoderServiceImpl struct {
	logger             *zap.SugaredLogger
	pipelineRepository pipelineConfig.PipelineRepository
}

func NewAPIReqDecoderServiceImpl(logger *zap.SugaredLogger,
	pipelineRepository pipelineConfig.PipelineRepository) *APIReqDecoderServiceImpl {
	return &APIReqDecoderServiceImpl{
		logger:             logger,
		pipelineRepository: pipelineRepository,
	}
}

func (impl *APIReqDecoderServiceImpl) GetFilterCriteriaParamsForDeploymentHistory(filterCriteria []string) (*bean2.DeploymentHistoryGetReqDecoderBean, error) {
	resp := &bean2.DeploymentHistoryGetReqDecoderBean{}
	for _, criteria := range filterCriteria {
		criteriaDecoder, err := util.DecodeFilterCriteriaString(criteria)
		if err != nil {
			impl.logger.Errorw("error encountered in applyFilterCriteriaOnResourceObjects", "filterCriteria", filterCriteria)
			return nil, err
		}
		switch criteriaDecoder.Kind {
		case bean2.DevtronResourceApplication:
			if criteriaDecoder.SubKind == bean2.DevtronResourceDevtronApplication {
				resp.AppId, err = strconv.Atoi(criteriaDecoder.Value)
			}
		case bean2.DevtronResourceEnvironment:
			resp.EnvId, err = strconv.Atoi(criteriaDecoder.Value)
		case bean2.DevtronResourceCdPipeline:
			resp.PipelineId, err = strconv.Atoi(criteriaDecoder.Value)
		}
	}
	if (resp.AppId == 0 || resp.EnvId == 0) && resp.PipelineId == 0 {
		//currently this method only supports history for a specific pipeline
		return nil, &util2.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			Code:            "400",
			InternalMessage: "missing required identifiers: either (appId and envId) or pipelineId must be provided to fetch history",
			UserMessage:     "Please provide either both App ID and Env ID, or a Pipeline ID to view history.",
		}
	}
	if resp.PipelineId == 0 {
		missingResources := []string{}
		if resp.AppId == 0 {
			missingResources = append(missingResources, "appId")
		}
		if resp.EnvId == 0 {
			missingResources = append(missingResources, "EnvId")
		}
		if len(missingResources) > 0 {
			return nil, &util2.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				Code:            "400",
				InternalMessage: "missing required resources: " + strings.Join(missingResources, ", "),
				UserMessage:     "missing required resources: " + strings.Join(missingResources, ", "),
			}
		}
		pipelineObj, err := impl.pipelineRepository.FindActiveByAppIdAndEnvId(resp.AppId, resp.EnvId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline", "appId", resp.AppId, "envId", resp.EnvId, "err", err)
			return nil, err
		}
		resp.PipelineId = pipelineObj.Id
	} else {
		pipelineObj, err := impl.pipelineRepository.FindById(resp.PipelineId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline", "appId", resp.AppId, "envId", resp.EnvId, "err", err)
			return nil, err
		}
		resp.AppId = pipelineObj.AppId
		resp.EnvId = pipelineObj.EnvironmentId
	}
	return resp, nil
}
