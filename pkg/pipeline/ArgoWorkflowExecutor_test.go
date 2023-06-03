package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestExecuteWorkflow(t *testing.T) {
	logger, loggerErr := util.NewSugardLogger()
	assert.Nil(t, loggerErr)
	//cdConfig, err := GetCdConfig()
	//assert.Nil(t, err)
	workflowExecutorImpl := NewArgoWorkflowExecutorImpl(logger)
	baseWorkflowTemplate := getBaseWorkflowTemplate()

	t.Run("validate not configured blob storage", func(t *testing.T) {
		workflowTemplate := deepCopyWfTemplate(baseWorkflowTemplate)
		workflowTemplate.BlobStorageConfigured = false
		cdTemplate := executeWorkflowAndGetCdTemplate(t, workflowExecutorImpl, workflowTemplate)
		archiveLocation := cdTemplate.ArchiveLocation
		assert.Equal(t, workflowTemplate.BlobStorageConfigured, *archiveLocation.ArchiveLogs)
	})

	t.Run("validate S3 blob storage", func(t *testing.T) {
		workflowTemplate := deepCopyWfTemplate(baseWorkflowTemplate)
		workflowTemplate.BlobStorageConfigured = true
		workflowTemplate.CloudStorageKey = "cloud-storage-key"
		s3BlobStorage := getS3BlobStorage()
		workflowTemplate.BlobStorageS3Config = s3BlobStorage
		cdTemplate := executeWorkflowAndGetCdTemplate(t, workflowExecutorImpl, workflowTemplate)
		archiveLocation := cdTemplate.ArchiveLocation
		assert.Equal(t, workflowTemplate.BlobStorageConfigured, *archiveLocation.ArchiveLogs)
		s3Artifact := archiveLocation.S3
		assert.Equal(t, workflowTemplate.CloudStorageKey, s3Artifact.Key)
		assert.Equal(t, s3BlobStorage.CiLogBucketName, s3Artifact.Bucket)
		assert.Equal(t, S3_ENDPOINT_URL, s3Artifact.Endpoint)
		assert.Equal(t, s3BlobStorage.IsInSecure, s3Artifact.Insecure)
		accessKeySecret := s3Artifact.AccessKeySecret
		secretKeySecret := s3Artifact.SecretKeySecret
		assert.NotNil(t, accessKeySecret)
		assert.NotNil(t, secretKeySecret)
		assert.True(t, reflect.DeepEqual(accessKeySecret, ACCESS_KEY_SELECTOR))
		assert.True(t, reflect.DeepEqual(secretKeySecret, SECRET_KEY_SELECTOR))
	})

	t.Run("validate gcp blob storage", func(t *testing.T) {
		workflowTemplate := deepCopyWfTemplate(baseWorkflowTemplate)
		workflowTemplate.BlobStorageConfigured = true
		workflowTemplate.CloudStorageKey = "cloud-storage-key"
		gcpBlobStorage := getGcpBlobStorage()
		workflowTemplate.GcpBlobConfig = gcpBlobStorage
		cdTemplate := executeWorkflowAndGetCdTemplate(t, workflowExecutorImpl, workflowTemplate)
		archiveLocation := cdTemplate.ArchiveLocation
		gcsArtifact := archiveLocation.GCS
		assert.Equal(t, workflowTemplate.BlobStorageConfigured, *archiveLocation.ArchiveLogs)
		assert.Equal(t, workflowTemplate.CloudStorageKey, gcsArtifact.Key)
		assert.Equal(t, gcpBlobStorage.LogBucketName, gcsArtifact.Bucket)
		secretKeySecret := gcsArtifact.ServiceAccountKeySecret
		assert.NotNil(t, secretKeySecret)
		assert.True(t, reflect.DeepEqual(secretKeySecret, SECRET_KEY_SELECTOR))
	})
}

func executeWorkflowAndGetCdTemplate(t *testing.T, workflowExecutorImpl *ArgoWorkflowExecutorImpl, workflowTemplate bean.WorkflowTemplate) v1alpha1.Template {
	unstructuredWorkflow, err := workflowExecutorImpl.ExecuteWorkflow(workflowTemplate)
	assert.Nil(t, err)
	cdWorkflow, err := getWorkflow(unstructuredWorkflow)
	assert.Nil(t, err)
	validateCdWorkflowSpec(t, cdWorkflow, workflowTemplate)
	cdTemplate := verifyCdTemplate(t, cdWorkflow, workflowTemplate)
	return cdTemplate
}

func validateCdWorkflowSpec(t *testing.T, cdWorkflow v1alpha1.Workflow, workflowTemplate bean.WorkflowTemplate) {
	assert.Equal(t, "", cdWorkflow.Namespace)
	objectMeta := cdWorkflow.ObjectMeta
	assert.Equal(t, fmt.Sprintf(WORKFLOW_GENERATE_NAME_REGEX, workflowTemplate.WorkflowNamePrefix), objectMeta.GenerateName)
	wfLabels := objectMeta.Labels
	assert.Equal(t, DEVTRON_WORKFLOW_LABEL_VALUE, wfLabels[DEVTRON_WORKFLOW_LABEL_KEY])
	workflowSpec := cdWorkflow.GetWorkflowSpec()
	assert.Equal(t, workflowTemplate.ServiceAccountName, workflowSpec.ServiceAccountName)
	assert.Equal(t, "", workflowSpec.Entrypoint)
	assert.True(t, reflect.DeepEqual(workflowTemplate.NodeSelector, workflowSpec.NodeSelector))
	assert.True(t, reflect.DeepEqual(workflowTemplate.Tolerations, workflowSpec.Tolerations))
	assert.True(t, reflect.DeepEqual(workflowTemplate.Volumes, workflowSpec.Volumes))
	ttlStrategy := workflowSpec.GetTTLStrategy()
	assert.NotNil(t, ttlStrategy)
	assert.Equal(t, *workflowTemplate.TTLValue, *ttlStrategy.SecondsAfterCompletion)
}

func getWorkflow(unstructuredWorkflow *unstructured.UnstructuredList) (v1alpha1.Workflow, error) {
	var cdWorkflow v1alpha1.Workflow
	var err error
	unstructuredItems := unstructuredWorkflow.Items
	for _, unstructuredItem := range unstructuredItems {
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredItem.Object, &cdWorkflow)
	}
	return cdWorkflow, err
}

func getS3BlobStorage() *blob_storage.BlobStorageS3Config {
	return &blob_storage.BlobStorageS3Config{
		AccessKey:       "accessKey",
		Passkey:         "passKey",
		EndpointUrl:     "",
		IsInSecure:      false,
		CiLogBucketName: "bucketName",
		CiLogRegion:     "ciLogRegion",
	}
}

func getGcpBlobStorage() *blob_storage.GcpBlobConfig {
	return &blob_storage.GcpBlobConfig{
		LogBucketName: "logBucketName",
	}
}

func deepCopyWfTemplate(baseWorkflowTemplate bean.WorkflowTemplate) bean.WorkflowTemplate {
	origJSON, err := json.Marshal(baseWorkflowTemplate)
	clone := bean.WorkflowTemplate{}
	if err = json.Unmarshal(origJSON, &clone); err != nil {
		return clone
	}
	return clone
}

func getBaseWorkflowTemplate() bean.WorkflowTemplate {
	workflowTemplate := bean.WorkflowTemplate{}
	workflowTemplate.WfControllerInstanceID = "random-controller-id"
	workflowTemplate.Namespace = "default"
	return workflowTemplate
}

func verifyCmCsTemplates(argoTemplates []v1alpha1.Template, workflowTemplate bean.WorkflowTemplate) {

}

func getCdTemplate(workflow v1alpha1.Workflow) v1alpha1.Template {
	var cdTemplate v1alpha1.Template
	templates := workflow.Spec.Templates
	for _, template := range templates {
		templateName := template.Name
		if templateName == CD_WORKFLOW_NAME {
			cdTemplate = template
		}
	}
	return cdTemplate
}

func verifyCdTemplate(t *testing.T, workflow v1alpha1.Workflow, workflowTemplate bean.WorkflowTemplate) v1alpha1.Template {
	//fetch template with name cd in template and verify content of that template
	cdTemplate := getCdTemplate(workflow)
	assert.NotNil(t, cdTemplate)
	templateContainer := cdTemplate.Container
	mainContainer := workflowTemplate.Containers[0]
	assert.True(t, reflect.DeepEqual(templateContainer, mainContainer))
	activeDeadlineSeconds := cdTemplate.ActiveDeadlineSeconds
	assert.NotNil(t, activeDeadlineSeconds)
	assert.Equal(t, *workflowTemplate.ActiveDeadlineSeconds, activeDeadlineSeconds.IntVal)
	return cdTemplate
}
