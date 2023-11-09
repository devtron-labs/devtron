package executors

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	types2 "github.com/devtron-labs/devtron/pkg/pipeline/types"
	"go.uber.org/zap"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
)

const (
	WORKFLOW_JOB_BACKOFF_LIMIT = 0
	WORKFLOW_JOB_FINALIZER     = "foregroundDeletion"
)

type SystemWorkflowExecutor interface {
	WorkflowExecutor
}

type SystemWorkflowExecutorImpl struct {
	logger  *zap.SugaredLogger
	k8sUtil *k8s.K8sUtil
}

func NewSystemWorkflowExecutorImpl(logger *zap.SugaredLogger, k8sUtil *k8s.K8sUtil) *SystemWorkflowExecutorImpl {
	return &SystemWorkflowExecutorImpl{logger: logger, k8sUtil: k8sUtil}
}

func (impl *SystemWorkflowExecutorImpl) ExecuteWorkflow(workflowTemplate bean.WorkflowTemplate) (*unstructured.UnstructuredList, error) {
	templatesList := &unstructured.UnstructuredList{}
	//create job template with suspended state
	jobTemplate := impl.getJobTemplate(workflowTemplate)
	_, clientset, err := impl.k8sUtil.GetK8sConfigAndClientsByRestConfig(workflowTemplate.ClusterConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while creating k8s client", "WorkflowRunnerId", workflowTemplate.WorkflowRunnerId, "err", err)
		return nil, err
	}
	ctx := context.Background()
	createdJob, err := clientset.BatchV1().Jobs(workflowTemplate.Namespace).Create(ctx, jobTemplate, v12.CreateOptions{})
	if err != nil {
		impl.logger.Errorw("error occurred while creating k8s job", "WorkflowRunnerId", workflowTemplate.WorkflowRunnerId, "err", err)
		return nil, err
	}

	//create cm and secrets with owner reference
	err = impl.createCmAndSecrets(workflowTemplate, createdJob, templatesList)
	if err != nil {
		impl.logger.Errorw("error occurred while creating cm and secret", "WorkflowRunnerId", workflowTemplate.WorkflowRunnerId, "err", err)
		return nil, err
	}

	//change job state to running
	_, err = clientset.BatchV1().Jobs(workflowTemplate.Namespace).Patch(ctx, createdJob.Name, types.StrategicMergePatchType, []byte(`{"spec":{"suspend": false}}`), v12.PatchOptions{})
	if err != nil {
		impl.logger.Errorw("error occurred while updating job suspended status", "WorkflowRunnerId", workflowTemplate.WorkflowRunnerId, "err", err)
		return nil, err
	}
	createdJob.Kind = jobTemplate.Kind
	createdJob.APIVersion = jobTemplate.APIVersion
	createdJob.Spec.Suspend = pointer.BoolPtr(false)
	impl.addToUnstructuredList(createdJob, templatesList)
	return templatesList, nil
}

func (impl *SystemWorkflowExecutorImpl) addToUnstructuredList(template interface{}, templateList *unstructured.UnstructuredList) {
	unstructuredObjMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&template)
	if err != nil {
		return
	}
	templateList.Items = append(templateList.Items, unstructured.Unstructured{Object: unstructuredObjMap})
}

func (impl *SystemWorkflowExecutorImpl) TerminateWorkflow(workflowName string, namespace string, clusterConfig *rest.Config) error {
	_, clientset, err := impl.k8sUtil.GetK8sConfigAndClientsByRestConfig(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while creating k8s client", "workflowName", workflowName, "namespace", namespace, "err", err)
		return err
	}
	err = clientset.BatchV1().Jobs(namespace).Delete(context.Background(), workflowName, v12.DeleteOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			err = fmt.Errorf("cannot find workflow %s", workflowName)
		}
		impl.logger.Errorw("error occurred while deleting workflow", "workflowName", workflowName, "namespace", namespace, "err", err)
	}
	return err
}

func (impl *SystemWorkflowExecutorImpl) GetWorkflow(workflowName string, namespace string, clusterConfig *rest.Config) (*unstructured.UnstructuredList, error) {
	templatesList := &unstructured.UnstructuredList{}
	_, clientset, err := impl.k8sUtil.GetK8sConfigAndClientsByRestConfig(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while creating k8s client", "workflowName", workflowName, "namespace", namespace, "err", err)
		return nil, err
	}
	wf, err := clientset.BatchV1().Jobs(namespace).Get(context.Background(), workflowName, v12.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			err = fmt.Errorf("cannot find workflow %s", workflowName)
		}
		return nil, err
	}
	impl.addToUnstructuredList(wf, templatesList)
	return templatesList, nil
}

// This will work for

func (impl *SystemWorkflowExecutorImpl) GetWorkflowStatus(workflowName string, namespace string, clusterConfig *rest.Config) (*types2.WorkflowStatus, error) {

	_, clientset, err := impl.k8sUtil.GetK8sConfigAndClientsByRestConfig(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while creating k8s client", "workflowName", workflowName, "namespace", namespace, "err", err)
		return nil, err
	}
	wf, err := clientset.BatchV1().Jobs(namespace).Get(context.Background(), workflowName, v12.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			err = fmt.Errorf("cannot find workflow %s", workflowName)
		}
		return nil, err
	}
	status := ""
	if len(wf.Status.Conditions) > 0 {
		status = string(wf.Status.Conditions[0].Type)
	}
	wfStatus := &types2.WorkflowStatus{
		Status: status,
	}
	return wfStatus, nil
}

func (impl *SystemWorkflowExecutorImpl) getJobTemplate(workflowTemplate bean.WorkflowTemplate) *v1.Job {

	workflowLabels := map[string]string{DEVTRON_WORKFLOW_LABEL_KEY: DEVTRON_WORKFLOW_LABEL_VALUE, "devtron.ai/purpose": "workflow", "workflowType": workflowTemplate.WorkflowType}

	//setting TerminationGracePeriodSeconds in PodSpec
	//which ensures Pod has enough time to execute cleanup on SIGTERM event
	workflowTemplate.PodSpec.TerminationGracePeriodSeconds = pointer.Int64(int64(workflowTemplate.TerminationGracePeriod))
	workflowJob := v1.Job{
		TypeMeta: v12.TypeMeta{
			Kind:       k8sCommonBean.JobKind,
			APIVersion: "batch/v1",
		},
		ObjectMeta: v12.ObjectMeta{
			GenerateName: fmt.Sprintf(WORKFLOW_GENERATE_NAME_REGEX, workflowTemplate.WorkflowNamePrefix),
			//Annotations:  map[string]string{"workflows.argoproj.io/controller-instanceid": workflowTemplate.WfControllerInstanceID},
			Labels:     workflowLabels,
			Finalizers: []string{WORKFLOW_JOB_FINALIZER},
		},
		Spec: v1.JobSpec{
			BackoffLimit:            pointer.Int32Ptr(WORKFLOW_JOB_BACKOFF_LIMIT),
			ActiveDeadlineSeconds:   workflowTemplate.ActiveDeadlineSeconds,
			TTLSecondsAfterFinished: workflowTemplate.TTLValue,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v12.ObjectMeta{
					Labels: workflowLabels,
				},
				Spec: workflowTemplate.PodSpec,
			},
			Suspend: pointer.BoolPtr(true),
		},
	}
	return &workflowJob
}

func (impl *SystemWorkflowExecutorImpl) getCmAndSecrets(workflowTemplate bean.WorkflowTemplate, createdJob *v1.Job) ([]corev1.ConfigMap, []corev1.Secret, error) {
	var configMaps []corev1.ConfigMap
	var secrets []corev1.Secret
	configMapDataArray := workflowTemplate.ConfigMaps
	for _, configSecretMap := range configMapDataArray {
		if configSecretMap.External {
			continue
		}
		configDataMap, err := configSecretMap.GetDataMap()
		if err != nil {
			impl.logger.Errorw("error occurred while extracting data map", "Data", configSecretMap.Data, "err", err)
			return configMaps, secrets, err
		}
		configMapSecretDto := types2.ConfigMapSecretDto{Name: configSecretMap.Name, Data: configDataMap, OwnerRef: impl.createJobOwnerRefVal(createdJob)}
		configMap := GetConfigMapBody(configMapSecretDto)
		configMaps = append(configMaps, configMap)
	}
	secretMaps := workflowTemplate.Secrets
	for _, secretMapData := range secretMaps {
		if secretMapData.External {
			continue
		}
		dataMap, err := secretMapData.GetDataMap()
		if err != nil {
			impl.logger.Errorw("error occurred while extracting data map", "Data", secretMapData.Data, "err", err)
			return configMaps, secrets, err
		}
		configMapSecretDto := types2.ConfigMapSecretDto{Name: secretMapData.Name, Data: dataMap, OwnerRef: impl.createJobOwnerRefVal(createdJob)}
		secretBody := GetSecretBody(configMapSecretDto)
		secrets = append(secrets, secretBody)
	}
	return configMaps, secrets, nil
}

func (impl *SystemWorkflowExecutorImpl) createJobOwnerRefVal(createdJob *v1.Job) v12.OwnerReference {
	return v12.OwnerReference{UID: createdJob.UID, Name: createdJob.Name, Kind: k8sCommonBean.JobKind, APIVersion: "batch/v1", BlockOwnerDeletion: pointer.BoolPtr(true), Controller: pointer.BoolPtr(true)}
}

func (impl *SystemWorkflowExecutorImpl) createCmAndSecrets(template bean.WorkflowTemplate, createdJob *v1.Job, templateList *unstructured.UnstructuredList) error {
	client, err := impl.k8sUtil.GetCoreV1ClientByRestConfig(template.ClusterConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while creating k8s client", "WorkflowRunnerId", template.WorkflowRunnerId, "err", err)
		return err
	}
	configMaps, secrets, err := impl.getCmAndSecrets(template, createdJob)
	if err != nil {
		return err
	}
	for _, configMap := range configMaps {
		impl.addToUnstructuredList(configMap, templateList)
		_, err = impl.k8sUtil.CreateConfigMap(createdJob.Namespace, &configMap, client)
		if err != nil {
			impl.logger.Errorw("error occurred while creating cm, but ignoring", "err", err)
		}
	}
	for _, secret := range secrets {
		impl.addToUnstructuredList(secret, templateList)
		_, err = impl.k8sUtil.CreateSecretData(createdJob.Namespace, &secret, client)
		if err != nil {
			impl.logger.Errorw("error occurred while creating secret, but ignoring", "err", err)
		}
	}
	return nil
}
