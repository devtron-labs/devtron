package devtronApps

import (
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
)

type FeasibilityManager interface {
	CheckFeasibility(triggerRequirementRequest *bean.TriggerRequirementRequestDto) error
}

func (impl *TriggerServiceImpl) CheckFeasibility(triggerRequirementRequest *bean.TriggerRequirementRequestDto) error {
	// have not implemented right now, will be implemented in future for security vulnerability
	return nil
}
