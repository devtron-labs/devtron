package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type WorkflowTemplate struct {
	WorkflowId       int
	WorkflowRunnerId int
	v1.PodSpec
	ConfigMaps             []bean.ConfigSecretMap
	Secrets                []bean.ConfigSecretMap
	TTLValue               *int32
	WorkflowRequestJson    string
	WorkflowNamePrefix     string
	WfControllerInstanceID string
	ClusterConfig          *rest.Config
	Namespace              string
	ArchiveLogs            bool
}

type JobManifestTemplate struct {
	Container     v1.Container           `json:"Container"`
	ConfigMaps    []bean.ConfigSecretMap `json:"ConfigMaps"`
	ConfigSecrets []bean.ConfigSecretMap `json:"ConfigSecrets"`
}
