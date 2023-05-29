package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type WorkflowTemplate struct {
	WorkflowId       int `json:"workflowId"`
	WorkflowRunnerId int `json:"workflowRunnerId"`
	v1.PodSpec
	ConfigMaps             []bean.ConfigSecretMap `json:"configMaps"`
	Secrets                []bean.ConfigSecretMap `json:"configSecrets"`
	TTLValue               *int32                 `json:"ttlValue"`
	WorkflowRequestJson    string                 `json:"workflowRequestJson"`
	WorkflowNamePrefix     string                 `json:"workflowNamePrefix"`
	WfControllerInstanceID string                 `json:"wfControllerInstanceID"`
	ClusterConfig          *rest.Config           `json:"-"`
	Namespace              string                 `json:"namespace"`
	ArchiveLogs            bool                   `json:"archiveLogs"`
}

type JobManifestTemplate struct {
	App                     string                 `json:"app"`
	NameSpace               string                 `json:"NameSpace"`
	Container               v1.Container           `json:"Container"`
	ConfigMaps              []bean.ConfigSecretMap `json:"ConfigMaps"`
	ConfigSecrets           []bean.ConfigSecretMap `json:"ConfigSecrets"`
	Toleration              []v1.Toleration        `json:"Toleration"`
	Affinity                v1.Affinity            `json:"Affinity"`
	NodeSelector            map[string]string      `json:"NodeSelector"`
	ActiveDeadlineSeconds   *int32                 `json:"ActiveDeadlineSeconds"`
	TTLSecondsAfterFinished *int32                 `json:"TTLSecondsAfterFinished"`
}
