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
	//IsExt                  bool
}
