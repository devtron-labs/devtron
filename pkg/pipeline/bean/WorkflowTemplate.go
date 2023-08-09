package bean

import (
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
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
	BlobStorageConfigured  bool
	BlobStorageS3Config    *blob_storage.BlobStorageS3Config
	CloudProvider          blob_storage.BlobStorageType
	AzureBlobConfig        *blob_storage.AzureBlobConfig
	GcpBlobConfig          *blob_storage.GcpBlobConfig
	CloudStorageKey        string
	TerminationGracePeriod int
	PrePostDeploySteps     []*StepObject
	RefPlugins             []*RefPluginObject
}

type JobManifestTemplate struct {
	NameSpace               string                 `json:"Namespace"`
	Container               v1.Container           `json:"Container"`
	ConfigMaps              []bean.ConfigSecretMap `json:"ConfigMaps"`
	ConfigSecrets           []bean.ConfigSecretMap `json:"ConfigSecrets"`
	Volumes                 []v1.Volume            `json:"Volumes"`
	Toleration              []v1.Toleration        `json:"Toleration"`
	Affinity                v1.Affinity            `json:"Affinity"`
	NodeSelector            map[string]string      `json:"NodeSelector"`
	ActiveDeadlineSeconds   *int64                 `json:"ActiveDeadlineSeconds"`
	TTLSecondsAfterFinished *int32                 `json:"TTLSecondsAfterFinished"`
}
