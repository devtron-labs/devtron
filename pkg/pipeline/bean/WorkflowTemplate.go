package bean

import (
	blob_storage "github.com/devtron-labs/common-lib-private/blob-storage"
	"github.com/devtron-labs/devtron/api/bean"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	PrePostDeploySteps     []*StepObject
	RefPlugins             []*RefPluginObject
	TerminationGracePeriod int
	WorkflowType           string
}

const (
	CI_WORKFLOW_NAME           = "ci"
	CI_WORKFLOW_WITH_STAGES    = "ci-stages-with-env"
	CiStage                    = "CI"
	CdStage                    = "CD"
	CD_WORKFLOW_NAME           = "cd"
	CD_WORKFLOW_WITH_STAGES    = "cd-stages-with-env"
	HELM_JOB_REF_TEMPLATE_NAME = "helm-job-template"
	JOB_CHART_API_VERSION      = "v2"
	JOB_CHART_NAME             = "helm-job"
	JOB_CHART_VERSION          = "0.1.0"
)

func (workflowTemplate *WorkflowTemplate) GetEntrypoint() string {
	switch workflowTemplate.WorkflowType {
	case CI_WORKFLOW_NAME:
		return CI_WORKFLOW_WITH_STAGES
	case CD_WORKFLOW_NAME:
		return CD_WORKFLOW_WITH_STAGES
	default:
		return ""
	}
}

func (workflowTemplate *WorkflowTemplate) CreateObjectMetadata() *v12.ObjectMeta {

	switch workflowTemplate.WorkflowType {
	case CI_WORKFLOW_NAME:
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Labels:       map[string]string{"devtron.ai/workflow-purpose": "ci"},
		}
	case CD_WORKFLOW_NAME:
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Annotations:  map[string]string{"workflows.argoproj.io/controller-instanceid": workflowTemplate.WfControllerInstanceID},
			Labels:       map[string]string{"devtron.ai/workflow-purpose": "cd"},
		}
	default:
		return nil
	}
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
