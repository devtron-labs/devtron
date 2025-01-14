package audit

import (
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository/audit"
)

func getInfraConfigEntTriggerAudit(config *v1.InfraConfig) ([]*audit.InfraConfigTriggerHistory, error) {
	return make([]*audit.InfraConfigTriggerHistory, 0), nil
}
