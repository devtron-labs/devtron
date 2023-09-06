package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/util"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"net/url"
)

const (
	STEP_NAME_REGEX              = "create-env-%s-gb-%d"
	TEMPLATE_NAME_REGEX          = "%s-gb-%d"
	WORKFLOW_MINIO_CRED          = "workflow-minio-cred"
	CRED_ACCESS_KEY              = "accessKey"
	CRED_SECRET_KEY              = "secretKey"
	S3_ENDPOINT_URL              = "s3.amazonaws.com"
	DEVTRON_WORKFLOW_LABEL_KEY   = "devtron.ai/workflow-purpose"
	DEVTRON_WORKFLOW_LABEL_VALUE = "cd"
	WORKFLOW_GENERATE_NAME_REGEX = "%s-"
	RESOURCE_CREATE_ACTION       = "create"
)

var ACCESS_KEY_SELECTOR = &v12.SecretKeySelector{Key: CRED_ACCESS_KEY, LocalObjectReference: v12.LocalObjectReference{Name: WORKFLOW_MINIO_CRED}}
var SECRET_KEY_SELECTOR = &v12.SecretKeySelector{Key: CRED_SECRET_KEY, LocalObjectReference: v12.LocalObjectReference{Name: WORKFLOW_MINIO_CRED}}

type WorkflowExecutor interface {
	ExecuteWorkflow(workflowTemplate bean.WorkflowTemplate) (*unstructured.UnstructuredList, error)
	TerminateWorkflow(workflowName string, namespace string, clusterConfig *rest.Config) error
}

type ArgoWorkflowExecutor interface {
	WorkflowExecutor
}

type ArgoWorkflowExecutorImpl struct {
	logger *zap.SugaredLogger
}

func NewArgoWorkflowExecutorImpl(logger *zap.SugaredLogger) *ArgoWorkflowExecutorImpl {
	return &ArgoWorkflowExecutorImpl{logger: logger}
}

func (impl *ArgoWorkflowExecutorImpl) TerminateWorkflow(workflowName string, namespace string, clusterConfig *rest.Config) error {
	impl.logger.Debugw("terminating wf", "name", workflowName)
	wfClient, err := impl.getClientInstance(namespace, clusterConfig)
	if err != nil {
		impl.logger.Errorw("cannot build wf client", "wfName", workflowName, "err", err)
		return err
	}
	_, err = wfClient.Get(context.Background(), workflowName, v1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("cannot find workflow", "name", workflowName, "err", err)
		return errors.New("cannot find workflow " + workflowName)
	}
	err = util.TerminateWorkflow(context.Background(), wfClient, workflowName)
	return err
}

func (impl *ArgoWorkflowExecutorImpl) ExecuteWorkflow(workflowTemplate bean.WorkflowTemplate) (*unstructured.UnstructuredList, error) {

	entryPoint := workflowTemplate.WorkflowType
	// get cm and cs argo step templates
	templates, err := impl.getArgoTemplates(workflowTemplate.ConfigMaps, workflowTemplate.Secrets, workflowTemplate.WorkflowType == bean.CI_WORKFLOW_NAME)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching argo templates and steps", "err", err)
		return nil, err
	}
	if len(templates) > 0 {
		//TODO KB: put this logic inside WFTemplate.getEntrypoint
		entryPoint = workflowTemplate.GetEntrypoint()
	}

	wfContainer := workflowTemplate.Containers[0]
	ciCdTemplate := v1alpha1.Template{
		Name:      workflowTemplate.WorkflowType,
		Container: &wfContainer,
		ActiveDeadlineSeconds: &intstr.IntOrString{
			IntVal: int32(*workflowTemplate.ActiveDeadlineSeconds),
		},
	}
	impl.updateBlobStorageConfig(workflowTemplate, &ciCdTemplate)
	templates = append(templates, ciCdTemplate)

	//TODO KB: use func wfTemplate.createObjectMetadata
	objectMeta := workflowTemplate.CreateObjectMetadata()

	var (
		ciCdWorkflow = v1alpha1.Workflow{
			ObjectMeta: *objectMeta,
			Spec: v1alpha1.WorkflowSpec{
				ServiceAccountName: workflowTemplate.ServiceAccountName,
				NodeSelector:       workflowTemplate.NodeSelector,
				Tolerations:        workflowTemplate.Tolerations,
				Entrypoint:         entryPoint,
				TTLStrategy: &v1alpha1.TTLStrategy{
					SecondsAfterCompletion: workflowTemplate.TTLValue,
				},
				Templates: templates,
				Volumes:   workflowTemplate.Volumes,
			},
		}
	)

	wfTemplate, err := json.Marshal(ciCdWorkflow)
	if err != nil {
		impl.logger.Errorw("error occurred while marshalling json", "err", err)
		return nil, err
	}
	impl.logger.Debugw("workflow request to submit", "wf", string(wfTemplate))

	wfClient, err := impl.getClientInstance(workflowTemplate.Namespace, workflowTemplate.ClusterConfig)
	if err != nil {
		impl.logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}

	createdWf, err := wfClient.Create(context.Background(), &ciCdWorkflow, v1.CreateOptions{})
	if err != nil {
		impl.logger.Errorw("error in wf trigger", "err", err)
		return nil, err
	}
	impl.logger.Debugw("workflow submitted: ", "name", createdWf.Name)
	return impl.convertToUnstructured(createdWf), nil
}

func (impl *ArgoWorkflowExecutorImpl) convertToUnstructured(cdWorkflow interface{}) *unstructured.UnstructuredList {
	unstructedObjMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&cdWorkflow)
	if err != nil {
		return nil
	}
	unstructuredObj := unstructured.Unstructured{Object: unstructedObjMap}
	unstructuredList := &unstructured.UnstructuredList{Items: []unstructured.Unstructured{unstructuredObj}}
	return unstructuredList
}

func (impl *ArgoWorkflowExecutorImpl) updateBlobStorageConfig(workflowTemplate bean.WorkflowTemplate, cdTemplate *v1alpha1.Template) {
	cdTemplate.ArchiveLocation = &v1alpha1.ArtifactLocation{
		ArchiveLogs: &workflowTemplate.ArchiveLogs,
	}
	if workflowTemplate.BlobStorageConfigured {
		var s3Artifact *v1alpha1.S3Artifact
		var gcsArtifact *v1alpha1.GCSArtifact
		blobStorageS3Config := workflowTemplate.BlobStorageS3Config
		gcpBlobConfig := workflowTemplate.GcpBlobConfig
		cloudStorageKey := workflowTemplate.CloudStorageKey
		if blobStorageS3Config != nil {
			s3CompatibleEndpointUrl := blobStorageS3Config.EndpointUrl
			if s3CompatibleEndpointUrl == "" {
				s3CompatibleEndpointUrl = "s3.amazonaws.com"
			} else {
				parsedUrl, err := url.Parse(s3CompatibleEndpointUrl)
				if err != nil {
					impl.logger.Errorw("error occurred while parsing s3CompatibleEndpointUrl, ", "s3CompatibleEndpointUrl", s3CompatibleEndpointUrl, "err", err)
				} else {
					s3CompatibleEndpointUrl = parsedUrl.Host
				}
			}
			isInsecure := blobStorageS3Config.IsInSecure
			var accessKeySelector *v12.SecretKeySelector
			var secretKeySelector *v12.SecretKeySelector
			if blobStorageS3Config.AccessKey != "" {
				accessKeySelector = &v12.SecretKeySelector{
					Key: CRED_ACCESS_KEY,
					LocalObjectReference: v12.LocalObjectReference{
						Name: WORKFLOW_MINIO_CRED,
					},
				}
				secretKeySelector = &v12.SecretKeySelector{
					Key: CRED_SECRET_KEY,
					LocalObjectReference: v12.LocalObjectReference{
						Name: WORKFLOW_MINIO_CRED,
					},
				}
			}
			s3Artifact = &v1alpha1.S3Artifact{
				Key: cloudStorageKey,
				S3Bucket: v1alpha1.S3Bucket{
					Endpoint:        s3CompatibleEndpointUrl,
					AccessKeySecret: accessKeySelector,
					SecretKeySecret: secretKeySelector,
					Bucket:          blobStorageS3Config.CiLogBucketName,
					Insecure:        &isInsecure,
				},
			}
			if blobStorageS3Config.CiLogRegion != "" {
				//TODO checking for Azure
				s3Artifact.Region = blobStorageS3Config.CiLogRegion
			}
		} else if gcpBlobConfig != nil {
			gcsArtifact = &v1alpha1.GCSArtifact{
				Key: cloudStorageKey,
				GCSBucket: v1alpha1.GCSBucket{
					Bucket: gcpBlobConfig.LogBucketName,
					ServiceAccountKeySecret: &v12.SecretKeySelector{
						Key: CRED_SECRET_KEY,
						LocalObjectReference: v12.LocalObjectReference{
							Name: WORKFLOW_MINIO_CRED,
						},
					},
				},
			}
		}

		// set in ArchiveLocation
		cdTemplate.ArchiveLocation.S3 = s3Artifact
		cdTemplate.ArchiveLocation.GCS = gcsArtifact
	}
}

func (impl *ArgoWorkflowExecutorImpl) getArgoTemplates(configMaps []bean2.ConfigSecretMap, secrets []bean2.ConfigSecretMap, isCi bool) ([]v1alpha1.Template, error) {
	var templates []v1alpha1.Template
	var steps []v1alpha1.ParallelSteps
	cmIndex := 0
	csIndex := 0
	for _, configMap := range configMaps {
		if configMap.External {
			continue
		}
		parallelStep, argoTemplate, err := impl.appendCMCSToStepAndTemplate(false, configMap, cmIndex)
		if err != nil {
			return templates, err
		}
		steps = append(steps, parallelStep)
		templates = append(templates, argoTemplate)
		cmIndex++
	}
	for _, secret := range secrets {
		if secret.External {
			continue
		}
		parallelStep, argoTemplate, err := impl.appendCMCSToStepAndTemplate(true, secret, csIndex)
		if err != nil {
			return templates, err
		}
		steps = append(steps, parallelStep)
		templates = append(templates, argoTemplate)
		csIndex++
	}
	if len(templates) <= 0 {
		return templates, nil
	}
	stepName := bean.CD_WORKFLOW_NAME
	templateName := bean.CD_WORKFLOW_WITH_STAGES
	if isCi {
		stepName = bean.CI_WORKFLOW_NAME
		templateName = bean.CI_WORKFLOW_WITH_STAGES
	}

	steps = append(steps, v1alpha1.ParallelSteps{
		Steps: []v1alpha1.WorkflowStep{
			{
				Name:     "run-wf",
				Template: stepName,
			},
		},
	})

	templates = append(templates, v1alpha1.Template{
		Name:  templateName,
		Steps: steps,
	})

	return templates, nil
}

func (impl *ArgoWorkflowExecutorImpl) appendCMCSToStepAndTemplate(isSecret bool, configSecretMap bean2.ConfigSecretMap, cmSecretIndex int) (v1alpha1.ParallelSteps, v1alpha1.Template, error) {
	var parallelStep v1alpha1.ParallelSteps
	var argoTemplate v1alpha1.Template
	configDataMap, err := configSecretMap.GetDataMap()
	if err != nil {
		impl.logger.Errorw("error occurred while extracting data map", "Data", configSecretMap.Data, "err", err)
		return parallelStep, argoTemplate, err
	}

	var cmSecretJson string
	configMapSecretDto := ConfigMapSecretDto{Name: configSecretMap.Name, Data: configDataMap, OwnerRef: ArgoWorkflowOwnerRef}
	if isSecret {
		cmSecretJson, err = GetSecretJson(configMapSecretDto)
	} else {
		cmSecretJson, err = GetConfigMapJson(configMapSecretDto)
	}
	if err != nil {
		impl.logger.Errorw("error occurred while extracting cm/secret json", "configSecretName", configSecretMap.Name, "err", err)
		return parallelStep, argoTemplate, err
	}
	parallelStep, argoTemplate = impl.createStepAndTemplate(isSecret, cmSecretIndex, cmSecretJson)
	return parallelStep, argoTemplate, nil
}

func (impl *ArgoWorkflowExecutorImpl) createStepAndTemplate(isSecret bool, cmSecretIndex int, cmSecretJson string) (v1alpha1.ParallelSteps, v1alpha1.Template) {
	stepName := fmt.Sprintf(STEP_NAME_REGEX, "cm", cmSecretIndex)
	templateName := fmt.Sprintf(TEMPLATE_NAME_REGEX, "cm", cmSecretIndex)
	if isSecret {
		stepName = fmt.Sprintf(STEP_NAME_REGEX, "secret", cmSecretIndex)
		templateName = fmt.Sprintf(TEMPLATE_NAME_REGEX, "secret", cmSecretIndex)
	}
	parallelStep := v1alpha1.ParallelSteps{
		Steps: []v1alpha1.WorkflowStep{
			{
				Name:     stepName,
				Template: templateName,
			},
		},
	}
	argoTemplate := v1alpha1.Template{
		Name: templateName,
		Resource: &v1alpha1.ResourceTemplate{
			Action:            RESOURCE_CREATE_ACTION,
			SetOwnerReference: true,
			Manifest:          string(cmSecretJson),
		},
	}
	return parallelStep, argoTemplate
}

func (impl *ArgoWorkflowExecutorImpl) getClientInstance(namespace string, clusterConfig *rest.Config) (v1alpha12.WorkflowInterface, error) {
	clientSet, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while creating client from config", "err", err)
		return nil, err
	}
	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
	return wfClient, nil
}
