package trigger

import (
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/security"
	"go.uber.org/zap"
)

type FeasibilityManagerImpl struct {
	logger           *zap.SugaredLogger
	imageScanService security.ImageScanService
}

func NewFeasibilityManagerImpl(logger *zap.SugaredLogger,
	imageScanService security.ImageScanService) *FeasibilityManagerImpl {
	return &FeasibilityManagerImpl{
		logger:           logger,
		imageScanService: imageScanService,
	}
}

type FeasibilityManager interface {
	CheckFeasibility(triggerRequirementRequest *bean.TriggerRequirementRequestDto) error
}

func (impl *FeasibilityManagerImpl) CheckFeasibility(triggerRequirementRequest *bean.TriggerRequirementRequestDto) error {
	//checking vulnerability for deploying image
	isVulnerable, err := impl.imageScanService.GetArtifactVulnerabilityStatus(triggerRequirementRequest.Artifact, triggerRequirementRequest.Pipeline, triggerRequirementRequest.Context)
	if err != nil {
		impl.logger.Errorw("error in getting Artifact vulnerability status, TriggerAutomaticDeployment", "err", err)
		return bean.GetOperationPerformError(err.Error())
	}
	if isVulnerable {
		return bean.GetVulnerabilityFoundError(triggerRequirementRequest.Artifact.ImageDigest)
	}
	return nil
}
