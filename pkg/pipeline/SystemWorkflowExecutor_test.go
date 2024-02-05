package pipeline

import (
	"fmt"
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/common-lib-private/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func TestSystemWorkflowExecute(t *testing.T) {
	t.SkipNow()
	logger, loggerErr := util.NewSugardLogger()
	assert.Nil(t, loggerErr)
	cdConfig, err := types.GetCiCdConfig()
	assert.Nil(t, err)
	runtimeConfig, err := client.GetRuntimeConfig()
	assert.Nil(t, err)
	k8sUtil := k8s.NewK8sUtil(logger, runtimeConfig)
	workflowExecutorImpl := executors.NewSystemWorkflowExecutorImpl(logger, k8sUtil)

	t.Run("validate not configured blob storage", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		executeAndValidateJobTemplate(t, workflowExecutorImpl, workflowTemplate)
	})

	t.Run("validate cm and cs", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		cms, secrets := getEnvSpecificCmCs(t)
		workflowTemplate.ConfigMaps = cms
		workflowTemplate.Secrets = secrets
		templatesList := executeAndValidateJobTemplate(t, workflowExecutorImpl, workflowTemplate)
		validateCmTemplates(t, templatesList, cms)
		validateSecretTemplates(t, templatesList, secrets)
	})
	t.Run("validate external cm and cs", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		cms, secrets := getEnvSpecificExternalCmCs(t)
		workflowTemplate.ConfigMaps = cms
		workflowTemplate.Secrets = secrets
		templatesList := executeAndValidateJobTemplate(t, workflowExecutorImpl, workflowTemplate)
		validateCmTemplates(t, templatesList, cms)
		validateSecretTemplates(t, templatesList, secrets)
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
		templateList := executeAndValidateJobTemplate(t, workflowExecutorImpl, workflowTemplate)
		jobTemplate, err := getJobTemplate(templateList)
		assert.Nil(t, err)
		err = workflowExecutorImpl.TerminateWorkflow(jobTemplate.Name, jobTemplate.Namespace, workflowTemplate.ClusterConfig)
		assert.Nil(t, err)
	})

	t.Run("terminate with invalid name", func(t *testing.T) {
		workflowTemplate := getBaseWorkflowTemplate(cdConfig)
		workflowTemplate.BlobStorageConfigured = false
		templateList := executeAndValidateJobTemplate(t, workflowExecutorImpl, workflowTemplate)
		jobTemplate, err := getJobTemplate(templateList)
		assert.Nil(t, err)
		err = workflowExecutorImpl.TerminateWorkflow("invalid-name", jobTemplate.Namespace, workflowTemplate.ClusterConfig)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "cannot find workflow invalid-name")
	})
}

func validateCmTemplates(t *testing.T, templatesList *unstructured.UnstructuredList, configMaps []bean2.ConfigSecretMap) {
	jobTemplate, err := getJobTemplate(templatesList)
	assert.Nil(t, err)
	for _, configMap := range configMaps {
		cmName := configMap.Name
		for _, templateItem := range templatesList.Items {
			if templateItem.GetKind() == "ConfigMap" && templateItem.GetName() == cmName {
				var cmFromTemplate v12.ConfigMap
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(templateItem.Object, &cmFromTemplate)
				assert.Nil(t, err)
				validateCmData(t, configMap.Data, cmFromTemplate.Data)
				verifyJobOwnerRef(t, jobTemplate, cmFromTemplate.OwnerReferences)
			}
		}
	}
}

func verifyJobOwnerRef(t *testing.T, jobTemplate v1.Job, ownerReferences []metav1.OwnerReference) {
	assert.Equal(t, 1, len(ownerReferences))
	ownerReference := ownerReferences[0]
	assert.Equal(t, jobTemplate.Name, ownerReference.Name)
	assert.Equal(t, jobTemplate.Kind, ownerReference.Kind)
	assert.Equal(t, jobTemplate.APIVersion, ownerReference.APIVersion)
	assert.Equal(t, jobTemplate.UID, ownerReference.UID)
	assert.NotNil(t, ownerReference.BlockOwnerDeletion)
	assert.True(t, *ownerReference.BlockOwnerDeletion)
	assert.NotNil(t, ownerReference.Controller)
	assert.True(t, *ownerReference.Controller)
}

func validateSecretTemplates(t *testing.T, templatesList *unstructured.UnstructuredList, secrets []bean2.ConfigSecretMap) {
	jobTemplate, err := getJobTemplate(templatesList)
	assert.Nil(t, err)
	for _, secret := range secrets {
		secretName := secret.Name
		for _, templateItem := range templatesList.Items {
			if templateItem.GetKind() == "Secret" && templateItem.GetName() == secretName {
				var secretFromTemplate v12.Secret
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(templateItem.Object, &secretFromTemplate)
				assert.Nil(t, err)
				validateSecretData(t, secret.Data, secretFromTemplate.Data)
				verifyJobOwnerRef(t, jobTemplate, secretFromTemplate.OwnerReferences)
			}
		}
	}
}

func executeAndValidateJobTemplate(t *testing.T, workflowExecutorImpl *executors.SystemWorkflowExecutorImpl, workflowTemplate bean.WorkflowTemplate) *unstructured.UnstructuredList {
	templatesList, err := workflowExecutorImpl.ExecuteWorkflow(workflowTemplate)
	assert.Nil(t, err)
	jobTemplate, err := getJobTemplate(templatesList)
	assert.Nil(t, err)
	validateJobTemplate(t, jobTemplate, workflowTemplate)
	return templatesList
}

func validateJobTemplate(t *testing.T, jobTemplate v1.Job, workflowTemplate bean.WorkflowTemplate) {
	objectMeta := jobTemplate.ObjectMeta
	assert.True(t, strings.Contains(objectMeta.Name, fmt.Sprintf(executors.WORKFLOW_GENERATE_NAME_REGEX, workflowTemplate.WorkflowNamePrefix)))
	wfLabels := objectMeta.Labels
	assert.Equal(t, executors.DEVTRON_WORKFLOW_LABEL_VALUE, wfLabels[executors.DEVTRON_WORKFLOW_LABEL_KEY])
	jobSpec := jobTemplate.Spec
	activeDeadlineSeconds := jobSpec.ActiveDeadlineSeconds
	assert.NotNil(t, activeDeadlineSeconds)
	assert.Equal(t, *workflowTemplate.ActiveDeadlineSeconds, *activeDeadlineSeconds)
	ttlSecondsAfterFinished := jobSpec.TTLSecondsAfterFinished
	assert.NotNil(t, ttlSecondsAfterFinished)
	assert.Equal(t, *workflowTemplate.TTLValue, *ttlSecondsAfterFinished)
	assert.NotNil(t, jobSpec.Suspend)
	assert.False(t, *jobSpec.Suspend)
}

func getJobTemplate(templatesList *unstructured.UnstructuredList) (v1.Job, error) {
	var jobTemplate v1.Job
	var err error
	for _, templateItem := range templatesList.Items {
		if templateItem.GetKind() == k8sCommonBean.JobKind {
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(templateItem.Object, &jobTemplate)
			break
		}
	}
	return jobTemplate, err
}
