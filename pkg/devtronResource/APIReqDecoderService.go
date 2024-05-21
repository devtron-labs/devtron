package devtronResource

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/util"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// APIReqDecoderService is common service used for getting decoded devtronResource and history related api params.
type APIReqDecoderService interface {
	GetFilterCriteriaParamsForDeploymentHistory(filterCriteria []string) (appId,
		environmentId, pipelineId, filterByReleaseId int, err error)
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

func (impl *APIReqDecoderServiceImpl) GetFilterCriteriaParamsForDeploymentHistory(filterCriteria []string) (appId,
	environmentId, pipelineId, filterByReleaseId int, err error) {
	for _, criteria := range filterCriteria {
		criteriaDecoder, err := util.DecodeFilterCriteriaString(criteria)
		if err != nil {
			impl.logger.Errorw("error encountered in applyFilterCriteriaOnResourceObjects", "filterCriteria", filterCriteria, "err", bean2.InvalidFilterCriteria)
			return appId, environmentId, pipelineId, filterByReleaseId, err
		}
		switch criteriaDecoder.Kind {
		case bean2.DevtronResourceApplication:
			if criteriaDecoder.SubKind == bean2.DevtronResourceDevtronApplication {
				appId, err = strconv.Atoi(criteriaDecoder.Value)
			}
		case bean2.DevtronResourceEnvironment:
			environmentId, err = strconv.Atoi(criteriaDecoder.Value)
		case bean2.DevtronResourceCdPipeline:
			pipelineId, err = strconv.Atoi(criteriaDecoder.Value)
		case bean2.DevtronResourceRelease:
			filterByReleaseId, err = strconv.Atoi(criteriaDecoder.Value)
		}
	}
	if (appId == 0 || environmentId == 0) && pipelineId == 0 {
		//currently this method only supports history for a specific pipeline
		return appId, environmentId, pipelineId, filterByReleaseId, util2.GetApiErrorAdapter(http.StatusBadRequest, "400", bean2.InvalidFilterCriteria, bean2.InvalidFilterCriteria)
	}
	if pipelineId == 0 {
		pipelineObj, err := impl.pipelineRepository.FindActiveByAppIdAndEnvId(appId, environmentId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline", "err", err, "appId", appId, "envId", environmentId)
			return appId, environmentId, pipelineId, filterByReleaseId, err
		}
		pipelineId = pipelineObj.Id
	} else {
		pipelineObj, err := impl.pipelineRepository.FindById(pipelineId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline", "err", err, "appId", appId, "envId", environmentId)
			return appId, environmentId, pipelineId, filterByReleaseId, err
		}
		appId = pipelineObj.AppId
		environmentId = pipelineObj.EnvironmentId
	}
	return appId, environmentId, pipelineId, filterByReleaseId, nil
}
