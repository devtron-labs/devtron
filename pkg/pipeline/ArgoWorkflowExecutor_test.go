package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	blob_storage "github.com/devtron-labs/common-lib-private/blob-storage"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestExecuteWorkflow(t *testing.T) {
	t.SkipNow()
	logger, loggerErr := util.NewSugardLogger()
	assert.Nil(t, loggerErr)
	cdConfig, err := types.GetCiCdConfig()
	assert.Nil(t, err)
	workflowExecutorImpl := executors.NewArgoWorkflowExecutorImpl(logger)

	t.Run("validate not configured blob storage", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		cdTemplate := executeWorkflowAndGetCdTemplate(t, workflowExecutorImpl, workflowTemplate)
		archiveLocation := cdTemplate.ArchiveLocation
		assert.Equal(t, workflowTemplate.BlobStorageConfigured, *archiveLocation.ArchiveLogs)
	})

	t.Run("validate S3 blob storage", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
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
		assert.Equal(t, executors.S3_ENDPOINT_URL, s3Artifact.Endpoint)
		assert.Equal(t, s3BlobStorage.IsInSecure, *s3Artifact.Insecure)
		accessKeySecret := s3Artifact.AccessKeySecret
		secretKeySecret := s3Artifact.SecretKeySecret
		assert.NotNil(t, accessKeySecret)
		assert.NotNil(t, secretKeySecret)
		assert.True(t, reflect.DeepEqual(accessKeySecret, executors.ACCESS_KEY_SELECTOR))
		assert.True(t, reflect.DeepEqual(secretKeySecret, executors.SECRET_KEY_SELECTOR))
	})

	t.Run("validate s3 blob storage with endpoint", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = true
		s3BlobStorageWithEndpoint := getS3BlobStorageWithEndpoint()
		workflowTemplate.BlobStorageS3Config = s3BlobStorageWithEndpoint
		cdTemplate := executeWorkflowAndGetCdTemplate(t, workflowExecutorImpl, workflowTemplate)
		archiveLocation := cdTemplate.ArchiveLocation
		assert.Equal(t, workflowTemplate.BlobStorageConfigured, *archiveLocation.ArchiveLogs)
		s3Artifact := archiveLocation.S3
		assert.True(t, strings.Contains(s3BlobStorageWithEndpoint.EndpointUrl, s3Artifact.Endpoint))
	})

	t.Run("validate gcp blob storage", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
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
		assert.True(t, reflect.DeepEqual(secretKeySecret, executors.SECRET_KEY_SELECTOR))
	})
	t.Run("validate env specific cm and secret", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		cms, secrets := getEnvSpecificCmCs(t)
		workflowTemplate.ConfigMaps = cms
		workflowTemplate.Secrets = secrets
		workflow := executeAndGetWorkflow(t, workflowExecutorImpl, workflowTemplate)
		verifyCmCsTemplates(t, workflow, workflowTemplate)
	})

	t.Run("validate external cm and cs", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		cms, secrets := getEnvSpecificExternalCmCs(t)
		workflowTemplate.ConfigMaps = cms
		workflowTemplate.Secrets = secrets
		workflow := executeAndGetWorkflow(t, workflowExecutorImpl, workflowTemplate)
		verifyCmCsTemplates(t, workflow, workflowTemplate)
	})

	t.Run("invalid config map", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		var configMaps []bean2.ConfigSecretMap
		cm := bean2.ConfigSecretMap{}
		cm.Name = "env-specific-cm-" + strconv.Itoa(rand.Intn(10000)) + "-" + strconv.Itoa(rand.Intn(10000))
		cm.Type = "environment"
		cm.Data = []byte("")
		configMaps = append(configMaps, cm)
		workflowTemplate.ConfigMaps = configMaps
		_, err := workflowExecutorImpl.ExecuteWorkflow(workflowTemplate)
		assert.NotNil(t, err)
	})

	t.Run("invalid cluster host", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		clusterConfig := workflowTemplate.ClusterConfig
		clusterConfig.Host = ""
		_, err := workflowExecutorImpl.ExecuteWorkflow(workflowTemplate)
		assert.NotNil(t, err)
	})

	t.Run("terminate workflow", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		cdWorkflow := executeAndGetWorkflow(t, workflowExecutorImpl, workflowTemplate)
		err := workflowExecutorImpl.TerminateWorkflow(cdWorkflow.Name, cdWorkflow.Namespace, workflowTemplate.ClusterConfig)
		assert.Nil(t, err)
	})

	t.Run("terminate with invalid name", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		cdWorkflow := executeAndGetWorkflow(t, workflowExecutorImpl, workflowTemplate)
		err := workflowExecutorImpl.TerminateWorkflow("invalid-name", cdWorkflow.Namespace, workflowTemplate.ClusterConfig)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "cannot find workflow invalid-name")
	})
}

func getEnvSpecificCmCs(t *testing.T) ([]bean2.ConfigSecretMap, []bean2.ConfigSecretMap) {
	var configMaps, secrets []bean2.ConfigSecretMap
	cmMap := make(map[string]string)
	cmMap["key1"] = "value1"
	cmDataJson, err := json.Marshal(cmMap)
	assert.Nil(t, err)

	secretMap := make(map[string]string)
	secretMap["secret1"] = "value1"
	secretDataJson, err := json.Marshal(secretMap)
	assert.Nil(t, err)

	cm := bean2.ConfigSecretMap{}
	cm.Name = "env-specific-cm-" + strconv.Itoa(rand.Intn(10000)) + "-" + strconv.Itoa(rand.Intn(10000))
	cm.Type = "environment"
	cm.Data = cmDataJson
	configMaps = append(configMaps, cm)

	secret := bean2.ConfigSecretMap{}
	secret.Name = "env-specific-secret-" + strconv.Itoa(rand.Intn(10000)) + "-" + strconv.Itoa(rand.Intn(10000))
	secret.Type = "environment"
	secret.Data = secretDataJson
	secret.SecretData = secretDataJson
	secrets = append(secrets, secret)

	return configMaps, secrets
}

func getEnvSpecificExternalCmCs(t *testing.T) ([]bean2.ConfigSecretMap, []bean2.ConfigSecretMap) {
	var configMaps, secrets []bean2.ConfigSecretMap
	cm := bean2.ConfigSecretMap{}
	cm.Name = "env-specific-cm-external"
	cm.Type = "environment"
	cm.External = true
	configMaps = append(configMaps, cm)

	secret := bean2.ConfigSecretMap{}
	secret.Name = "env-specific-secret-external"
	secret.Type = "environment"
	secret.External = true
	secrets = append(secrets, secret)

	return configMaps, secrets
}

func executeWorkflowAndGetCdTemplate(t *testing.T, workflowExecutorImpl *executors.ArgoWorkflowExecutorImpl, workflowTemplate bean.WorkflowTemplate) v1alpha1.Template {
	cdWorkflow := executeAndGetWorkflow(t, workflowExecutorImpl, workflowTemplate)
	cdTemplate := verifyCdTemplate(t, cdWorkflow, workflowTemplate)
	return cdTemplate
}

func executeAndGetWorkflow(t *testing.T, workflowExecutorImpl *executors.ArgoWorkflowExecutorImpl, workflowTemplate bean.WorkflowTemplate) v1alpha1.Workflow {
	unstructuredWorkflow, err := workflowExecutorImpl.ExecuteWorkflow(workflowTemplate)
	assert.Nil(t, err)
	cdWorkflow, err := getWorkflow(unstructuredWorkflow)
	assert.Nil(t, err)
	validateCdWorkflowSpec(t, cdWorkflow, workflowTemplate)
	return cdWorkflow
}

func validateCdWorkflowSpec(t *testing.T, cdWorkflow v1alpha1.Workflow, workflowTemplate bean.WorkflowTemplate) {
	assert.Equal(t, "default", cdWorkflow.Namespace)
	objectMeta := cdWorkflow.ObjectMeta
	assert.Equal(t, fmt.Sprintf(executors.WORKFLOW_GENERATE_NAME_REGEX, workflowTemplate.WorkflowNamePrefix), objectMeta.GenerateName)
	wfLabels := objectMeta.Labels
	assert.Equal(t, executors.DEVTRON_WORKFLOW_LABEL_VALUE, wfLabels[executors.DEVTRON_WORKFLOW_LABEL_KEY])
	workflowSpec := cdWorkflow.GetWorkflowSpec()
	assert.Equal(t, workflowTemplate.ServiceAccountName, workflowSpec.ServiceAccountName)
	//assert.Equal(t, "", workflowSpec.Entrypoint)
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

func getS3BlobStorageWithEndpoint() *blob_storage.BlobStorageS3Config {
	s3BlobStorage := getS3BlobStorage()
	s3BlobStorage.EndpointUrl = "https://minio.url:9090"
	return s3BlobStorage
}

func getGcpBlobStorage() *blob_storage.GcpBlobConfig {
	return &blob_storage.GcpBlobConfig{
		LogBucketName: "logBucketName",
	}
}

func getBaseWorkflowTemplate(cdConfig *types.CiCdConfig) bean.WorkflowTemplate {

	workflowTemplate := bean.WorkflowTemplate{}
	workflowTemplate.WfControllerInstanceID = "random-controller-id"
	workflowTemplate.Namespace = "default"
	workflowTemplate.ActiveDeadlineSeconds = pointer.Int64Ptr(3600)
	workflowTemplate.RestartPolicy = v12.RestartPolicyNever
	clusterConfig := deepCopyClusterConfig(*cdConfig.ClusterConfig)
	workflowTemplate.ClusterConfig = &clusterConfig
	workflowTemplate.WorkflowNamePrefix = "workflow-mock-prefix-" + strconv.Itoa(rand.Intn(1000))
	workflowTemplate.TTLValue = pointer.Int32Ptr(3600)
	var containers []v12.Container
	containers = append(containers, getMainContainer())
	workflowTemplate.Containers = containers
	return workflowTemplate
}

func deepCopyClusterConfig(clusterConfig rest.Config) rest.Config {
	return clusterConfig
}

func getMainContainer() v12.Container {
	return v12.Container{
		Name:  common.MainContainerName,
		Image: "random-image",
	}
}

func verifyCmCsTemplates(t *testing.T, workflow v1alpha1.Workflow, workflowTemplate bean.WorkflowTemplate) {
	cms := workflowTemplate.ConfigMaps
	for _, configMap := range cms {
		verifyConfigMapTemplate(t, configMap, workflow)
	}
	for _, secret := range workflowTemplate.Secrets {
		verifySecretTemplate(t, secret, workflow)
	}
}

func verifySecretTemplate(t *testing.T, secret bean2.ConfigSecretMap, workflow v1alpha1.Workflow) {
	templates := workflow.GetTemplates()
	for _, template := range templates {
		resourceTemplate := template.Resource
		if resourceTemplate == nil {
			continue
		}
		resourceAction := resourceTemplate.Action
		assert.Equal(t, executors.RESOURCE_CREATE_ACTION, resourceAction)
		assert.True(t, resourceTemplate.SetOwnerReference)
		secretJson := resourceTemplate.Manifest
		secretFromTemplateMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(secretJson), &secretFromTemplateMap)
		secretFromTemplate := v12.Secret{}
		assert.Nil(t, err)
		if "Secret" == secretFromTemplateMap["Kind"] {
			err = json.Unmarshal([]byte(secretJson), &secretFromTemplate)
			assert.Equal(t, secret.Name, secretFromTemplate.Name)
			validateSecretData(t, secret.SecretData, secretFromTemplate.Data)
			verifyOwnerRef(t, secretFromTemplate.OwnerReferences)
		}
	}
}

func verifyConfigMapTemplate(t *testing.T, configMap bean2.ConfigSecretMap, workflow v1alpha1.Workflow) {
	templates := workflow.GetTemplates()
	for _, template := range templates {
		resourceTemplate := template.Resource
		if resourceTemplate == nil {
			continue
		}
		resourceAction := resourceTemplate.Action
		assert.Equal(t, executors.RESOURCE_CREATE_ACTION, resourceAction)
		assert.True(t, resourceTemplate.SetOwnerReference)
		configMapJson := resourceTemplate.Manifest
		cmFromTemplateMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(configMapJson), &cmFromTemplateMap)
		assert.Nil(t, err)
		if "ConfigMap" == cmFromTemplateMap["Kind"] {
			cmFromTemplate := v12.ConfigMap{}
			err = json.Unmarshal([]byte(configMapJson), &cmFromTemplate)
			assert.Nil(t, err)
			assert.Equal(t, configMap.Name, cmFromTemplate.Name)
			validateCmData(t, configMap.Data, cmFromTemplate.Data)
			verifyOwnerRef(t, cmFromTemplate.OwnerReferences)
		}
	}
}

func validateSecretData(t *testing.T, data []byte, secret map[string][]byte) {
	secretMap := make(map[string]string)
	for key, secretValueBytes := range secret {
		secretMap[key] = string(secretValueBytes)
	}
	validateCmData(t, data, secretMap)
}

func validateCmData(t *testing.T, data []byte, cmMap map[string]string) {
	cmData := make(map[string]string)
	err := json.Unmarshal(data, &cmData)
	assert.Nil(t, err)
	assert.Equal(t, cmData, cmMap)
}

func verifyOwnerRef(t *testing.T, ownerReferences []metav1.OwnerReference) {
	assert.Equal(t, 1, len(ownerReferences))
	ownerReference := ownerReferences[0]
	assert.True(t, reflect.DeepEqual(ownerReference, executors.ArgoWorkflowOwnerRef))
}

func getCdTemplate(workflow v1alpha1.Workflow) v1alpha1.Template {
	var cdTemplate v1alpha1.Template
	templates := workflow.Spec.Templates
	for _, template := range templates {
		templateName := template.Name
		if templateName == bean.CD_WORKFLOW_NAME {
			cdTemplate = template
		}
	}
	return cdTemplate
}

func verifyCdTemplate(t *testing.T, workflow v1alpha1.Workflow, workflowTemplate bean.WorkflowTemplate) v1alpha1.Template {
	cdTemplate := getCdTemplate(workflow)
	assert.NotNil(t, cdTemplate)
	templateContainer := cdTemplate.Container
	mainContainer := workflowTemplate.Containers[0]
	assert.True(t, reflect.DeepEqual(*templateContainer, mainContainer))
	activeDeadlineSeconds := cdTemplate.ActiveDeadlineSeconds
	assert.NotNil(t, activeDeadlineSeconds)
	assert.Equal(t, *workflowTemplate.ActiveDeadlineSeconds, int64(activeDeadlineSeconds.IntVal))
	return cdTemplate
}
