package bean

import (
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
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

func (workflowTemplate *WorkflowTemplate) GetEntrypoint() string {
	switch workflowTemplate.WorkflowType {
	case pipeline.CI_WORKFLOW_NAME:
		return pipeline.CI_WORKFLOW_WITH_STAGES
	case pipeline.CD_WORKFLOW_NAME:
		return pipeline.CD_WORKFLOW_WITH_STAGES
	default:
		return ""
	}
}

func (workflowTemplate *WorkflowTemplate) CreateObjectMetadata() *v12.ObjectMeta {

	switch workflowTemplate.WorkflowType {
	case pipeline.CiStage:
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Labels:       map[string]string{"devtron.ai/workflow-purpose": "ci"},
		}
	case pipeline.CdStage:
		return &v12.ObjectMeta{
			GenerateName: workflowTemplate.WorkflowNamePrefix + "-",
			Annotations:  map[string]string{"workflows.argoproj.io/controller-instanceid": workflowTemplate.WfControllerInstanceID},
			Labels:       map[string]string{"devtron.ai/workflow-purpose": "cd"},
		}
	default:
		return nil
	}
}
