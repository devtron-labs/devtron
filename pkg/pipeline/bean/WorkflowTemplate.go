package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	v1 "k8s.io/api/core/v1"
)

type WorkflowTemplate struct {
	v1.PodSpec
	ConfigMaps          []bean.ConfigSecretMap
	Secrets             []bean.ConfigSecretMap
	TTLValue            int32
	WorkflowRequestJson string
}
