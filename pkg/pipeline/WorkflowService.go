/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/util"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util/k8s"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strings"
)

// TODO: move isCi/isJob to workflowRequest

type WorkflowService interface {
	SubmitWorkflow(workflowRequest *WorkflowRequest) error
	//DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(name string, namespace string, isExt bool, environment *repository.Environment) (*v1alpha1.Workflow, error)
	//ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	//UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	TerminateWorkflow(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExt bool, environment *repository.Environment) error
}

type WorkflowServiceImpl struct {
	Logger                 *zap.SugaredLogger
	config                 *rest.Config
	ciCdConfig             *CiCdConfig
	appService             app.AppService
	envRepository          repository.EnvironmentRepository
	globalCMCSService      GlobalCMCSService
	argoWorkflowExecutor   ArgoWorkflowExecutor
	systemWorkflowExecutor SystemWorkflowExecutor
	k8sUtil                *k8s.K8sUtil
	k8sCommonService       k8s2.K8sCommonService
}

// TODO: Move to bean

func NewWorkflowServiceImpl(Logger *zap.SugaredLogger, envRepository repository.EnvironmentRepository, ciCdConfig *CiCdConfig,
	appService app.AppService, globalCMCSService GlobalCMCSService, argoWorkflowExecutor ArgoWorkflowExecutor,
	k8sUtil *k8s.K8sUtil,
	systemWorkflowExecutor SystemWorkflowExecutor, k8sCommonService k8s2.K8sCommonService) (*WorkflowServiceImpl, error) {
	commonWorkflowService := &WorkflowServiceImpl{Logger: Logger,
		ciCdConfig:             ciCdConfig,
		appService:             appService,
		envRepository:          envRepository,
		globalCMCSService:      globalCMCSService,
		argoWorkflowExecutor:   argoWorkflowExecutor,
		k8sUtil:                k8sUtil,
		systemWorkflowExecutor: systemWorkflowExecutor,
		k8sCommonService:       k8sCommonService,
	}
	restConfig, err := k8sUtil.GetK8sInClusterRestConfig()
	if err != nil {
		Logger.Errorw("error in getting in cluster rest config", "err", err)
		return nil, err
	}
	commonWorkflowService.config = restConfig
	return commonWorkflowService, nil
}

const (
	BLOB_STORAGE_AZURE             = "AZURE"
	BLOB_STORAGE_S3                = "S3"
	BLOB_STORAGE_GCP               = "GCP"
	BLOB_STORAGE_MINIO             = "MINIO"
	CI_WORKFLOW_NAME               = "ci"
	CI_WORKFLOW_WITH_STAGES        = "ci-stages-with-env"
	CI_NODE_SELECTOR_APP_LABEL_KEY = "devtron.ai/node-selector"
	CI_NODE_PVC_ALL_ENV            = "devtron.ai/ci-pvc-all"
	CI_NODE_PVC_PIPELINE_PREFIX    = "devtron.ai/ci-pvc"
	PRE                            = "PRE"
	POST                           = "POST"
	CiStage                        = "CI"
	CdStage                        = "CD"
	CD_WORKFLOW_NAME               = "cd"
	CD_WORKFLOW_WITH_STAGES        = "cd-stages-with-env"
)

func (impl *WorkflowServiceImpl) SubmitWorkflow(workflowRequest *WorkflowRequest) error {
	//Create ciCdTriggerEvent
	//containerEnvVariables := []v12.EnvVar{}
	//if workflowRequest.isCi {
	//	containerEnvVariables = []v12.EnvVar{{Name: "IMAGE_SCANNER_ENDPOINT", Value: impl.ciCdConfig.ImageScannerEndpoint}}
	//}
	//if impl.ciCdConfig.CloudProvider == BLOB_STORAGE_S3 && impl.ciCdConfig.BlobStorageS3AccessKey != "" {
	//	miniCred := []v12.EnvVar{{Name: "AWS_ACCESS_KEY_ID", Value: impl.ciCdConfig.BlobStorageS3AccessKey}, {Name: "AWS_SECRET_ACCESS_KEY", Value: impl.ciCdConfig.BlobStorageS3SecretKey}}
	//	containerEnvVariables = append(containerEnvVariables, miniCred...)
	//}
	//if (workflowRequest.StageType == PRE && pipeline.RunPreStageInEnv) || (workflowRequest.StageType == POST && pipeline.RunPostStageInEnv) {
	//	workflowRequest.IsExtRun = true
	//}
	//if env != nil && env.Id != 0 && isCi {
	//	workflowRequest.EnvironmentId = env.Id
	//	workflowRequest.IsExtRun = true
	//}
	//
	//pvc := appLabels[strings.ToLower(fmt.Sprintf("%s-%s", CI_NODE_PVC_PIPELINE_PREFIX, workflowRequest.PipelineName))]
	//if len(pvc) == 0 {
	//	pvc = appLabels[CI_NODE_PVC_ALL_ENV]
	//}
	//if len(pvc) != 0 {
	//	workflowRequest.IsPvcMounted = true
	//	workflowRequest.IgnoreDockerCachePush = true
	//	workflowRequest.IgnoreDockerCachePull = true
	//}
	//eventType := cdStage
	//if isCi {
	//	eventType = ciEvent
	//}
	//// key will be used for log archival through in-app logging
	//if isCi {
	//	workflowRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", impl.ciCdConfig.CiDefaultBuildLogsKeyPrefix, workflowRequest.WorkflowPrefixForLog)
	//} else {
	//	workflowRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", impl.ciCdConfig.CdDefaultBuildLogsKeyPrefix, workflowRequest.WorkflowPrefixForLog)
	//}
	//workflowRequest.InAppLoggingEnabled = impl.ciCdConfig.InAppLoggingEnabled || (workflowRequest.WorkflowExecutor == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM)
	//ciCdTriggerEvent := CiCdTriggerEvent{
	//	Type:                  eventType,
	//	CommonWorkflowRequest: workflowRequest,
	//}
	//workflowJson, err := json.Marshal(&ciCdTriggerEvent)
	//if err != nil {
	//	impl.Logger.Errorw("error occurred while marshalling ciCdTriggerEvent", "error", err)
	//	return err
	//}
	// END ciCdTriggerEvent
	workflowTemplate, err := impl.createWorkflowTemplate(workflowRequest)
	if err != nil {
		return err
	}
	//privileged := true
	//storageConfigured := workflowRequest.BlobStorageConfigured
	//ttl := int32(impl.ciCdConfig.BuildLogTTLValue)
	//workflowTemplate.TTLValue = &ttl
	//workflowTemplate.WorkflowId = workflowRequest.WorkflowId
	//if !isCi {
	//	workflowTemplate.WorkflowRunnerId = workflowRequest.WorkflowRunnerId
	//	workflowTemplate.PrePostDeploySteps = workflowRequest.PrePostDeploySteps
	//}
	//workflowTemplate.WorkflowRequestJson = string(workflowJson)
	//workflowTemplate.RefPlugins = workflowRequest.RefPlugins
	//var globalCmCsConfigs []*bean3.GlobalCMCSDto
	//var workflowConfigMaps []bean.ConfigSecretMap
	//var workflowSecrets []bean.ConfigSecretMap
	//if !workflowRequest.IsExtRun {
	//	// inject global variables only if IsExtRun is false
	//	if isCi {
	//		globalCmCsConfigs, err = impl.globalCMCSService.FindAllActiveByPipelineType(repository2.PIPELINE_TYPE_CI)
	//	} else {
	//		globalCmCsConfigs, err = impl.globalCMCSService.FindAllActiveByPipelineType(repository2.PIPELINE_TYPE_CD)
	//	}
	//	if err != nil {
	//		impl.Logger.Errorw("error in getting all global cm/cs config", "err", err)
	//		return err
	//	}
	//	if isCi {
	//		for i := range globalCmCsConfigs {
	//			globalCmCsConfigs[i].Name = strings.ToLower(globalCmCsConfigs[i].Name) + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
	//		}
	//	} else {
	//		for i := range globalCmCsConfigs {
	//			globalCmCsConfigs[i].Name = fmt.Sprintf("%s-%s-%s", strings.ToLower(globalCmCsConfigs[i].Name), strconv.Itoa(workflowRequest.WorkflowRunnerId), CD_WORKFLOW_NAME)
	//		}
	//	}
	//	workflowConfigMaps, workflowSecrets, err = GetFromGlobalCmCsDtos(globalCmCsConfigs)
	//	if err != nil {
	//		impl.Logger.Errorw("error in creating templates for global secrets", "err", err)
	//	}
	//}
	//var cdPipelineLevelConfigMaps, cdPipelineLevelSecrets map[string]bool
	//if !isCi {
	//	cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err = impl.getConfiguredCmCs(pipeline, workflowRequest.StageType)
	//	if err != nil {
	//		impl.Logger.Errorw("error occurred while fetching pipeline configured cm and cs", "pipelineId", pipeline.Id, "err", err)
	//		return err
	//	}
	//}
	//var existingConfigMap *bean.ConfigMapJson
	//var existingSecrets *bean.ConfigSecretJson
	//if !isCi || isJob {
	//	existingConfigMap, existingSecrets, err = impl.appService.GetCmSecretNew(workflowRequest.AppId, workflowRequest.EnvironmentId, isJob)
	//	if err != nil {
	//		impl.Logger.Errorw("failed to get configmap data", "err", err)
	//		return err
	//	}
	//	impl.Logger.Debugw("existing cm", "cm", existingConfigMap, "secrets", existingSecrets)
	//}
	//
	//if isCi && isJob {
	//	for _, cm := range existingConfigMap.Maps {
	//		if !cm.External {
	//			cm.Name = cm.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
	//		}
	//		workflowConfigMaps = append(workflowConfigMaps, cm)
	//	}
	//
	//	for _, secret := range existingSecrets.Secrets {
	//		if !secret.External {
	//			secret.Name = secret.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
	//		}
	//		workflowSecrets = append(workflowSecrets, *secret)
	//	}
	//} else if !isCi {
	//
	//	for _, cm := range existingConfigMap.Maps {
	//		if _, ok := cdPipelineLevelConfigMaps[cm.Name]; ok {
	//			if !cm.External {
	//				cm.Name = cm.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
	//			}
	//			workflowConfigMaps = append(workflowConfigMaps, cm)
	//		}
	//	}
	//	for _, secret := range existingSecrets.Secrets {
	//		if _, ok := cdPipelineLevelSecrets[secret.Name]; ok {
	//			if !secret.External {
	//				secret.Name = secret.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
	//			}
	//			workflowSecrets = append(workflowSecrets, *secret)
	//		}
	//	}
	//}
	//workflowTemplate.ConfigMaps = workflowConfigMaps
	//workflowTemplate.Secrets = workflowSecrets
	//if isCi {
	//	workflowTemplate.ServiceAccountName = impl.ciCdConfig.CiWorkflowServiceAccount
	//	if impl.ciCdConfig.CiTaintKey != "" || impl.ciCdConfig.CiTaintValue != "" {
	//		workflowTemplate.Tolerations = []v12.Toleration{{Key: impl.ciCdConfig.CiTaintKey, Value: impl.ciCdConfig.CiTaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
	//	}
	//	// In the future, we will give support for NodeSelector for job currently we need to have a node without dedicated NodeLabel to run job
	//	if len(impl.ciCdConfig.NodeLabel) > 0 && !(isJob && workflowRequest.IsExtRun) {
	//		workflowTemplate.NodeSelector = impl.ciCdConfig.NodeLabel
	//	}
	//} else {
	//	workflowTemplate.ServiceAccountName = impl.ciCdConfig.CdWorkflowServiceAccount
	//	workflowTemplate.NodeSelector = map[string]string{impl.ciCdConfig.CdTaintKey: impl.ciCdConfig.CdTaintValue}
	//	workflowTemplate.Tolerations = []v12.Toleration{{Key: impl.ciCdConfig.CdTaintKey, Value: impl.ciCdConfig.CdTaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
	//	if len(impl.ciCdConfig.NodeLabel) > 0 {
	//		workflowTemplate.NodeSelector = impl.ciCdConfig.NodeLabel
	//	}
	//}
	//workflowTemplate.Volumes = ExtractVolumesFromCmCs(workflowConfigMaps, workflowSecrets)
	//workflowTemplate.ArchiveLogs = storageConfigured
	//workflowTemplate.ArchiveLogs = workflowTemplate.ArchiveLogs && !workflowRequest.InAppLoggingEnabled
	//workflowTemplate.RestartPolicy = v12.RestartPolicyNever
	//var limitCpu, limitMem, reqCpu, reqMem string
	//if isCi {
	//	limitCpu = impl.ciCdConfig.CiLimitCpu
	//	limitMem = impl.ciCdConfig.CiLimitMem
	//	reqCpu = impl.ciCdConfig.CiReqCpu
	//	reqMem = impl.ciCdConfig.CiReqMem
	//} else {
	//	limitCpu = impl.ciCdConfig.CdLimitCpu
	//	limitMem = impl.ciCdConfig.CdLimitMem
	//	reqCpu = impl.ciCdConfig.CdReqCpu
	//	reqMem = impl.ciCdConfig.CdReqMem
	//}
	//
	//eventEnv := v12.EnvVar{Name: "CI_CD_EVENT", Value: string(workflowJson)}
	//inAppLoggingEnv := v12.EnvVar{Name: "IN_APP_LOGGING", Value: strconv.FormatBool(workflowRequest.InAppLoggingEnabled)}
	//containerEnvVariables = append(containerEnvVariables, eventEnv, inAppLoggingEnv)
	//workflowImage := workflowRequest.CdImage
	//if isCi {
	//	workflowImage = workflowRequest.CiImage
	//}
	//workflowMainContainer := v12.Container{
	//	Env:   containerEnvVariables,
	//	Name:  common.MainContainerName,
	//	Image: workflowImage,
	//	SecurityContext: &v12.SecurityContext{
	//		Privileged: &privileged,
	//	},
	//	Resources: v12.ResourceRequirements{
	//		Limits: v12.ResourceList{
	//			v12.ResourceCPU:    resource.MustParse(limitCpu),
	//			v12.ResourceMemory: resource.MustParse(limitMem),
	//		},
	//		Requests: v12.ResourceList{
	//			v12.ResourceCPU:    resource.MustParse(reqCpu),
	//			v12.ResourceMemory: resource.MustParse(reqMem),
	//		},
	//	},
	//}
	//if len(pvc) != 0 && isCi {
	//	buildPvcCachePath := impl.ciCdConfig.BuildPvcCachePath
	//	buildxPvcCachePath := impl.ciCdConfig.BuildxPvcCachePath
	//	defaultPvcCachePath := impl.ciCdConfig.DefaultPvcCachePath
	//
	//	workflowTemplate.Volumes = append(workflowTemplate.Volumes, v12.Volume{
	//		Name: "root-vol",
	//		VolumeSource: v12.VolumeSource{
	//			PersistentVolumeClaim: &v12.PersistentVolumeClaimVolumeSource{
	//				ClaimName: pvc,
	//				ReadOnly:  false,
	//			},
	//		},
	//	})
	//	workflowMainContainer.VolumeMounts = append(workflowMainContainer.VolumeMounts,
	//		v12.VolumeMount{
	//			Name:      "root-vol",
	//			MountPath: buildPvcCachePath,
	//		},
	//		v12.VolumeMount{
	//			Name:      "root-vol",
	//			MountPath: buildxPvcCachePath,
	//		},
	//		v12.VolumeMount{
	//			Name:      "root-vol",
	//			MountPath: defaultPvcCachePath,
	//		})
	//}
	//UpdateContainerEnvsFromCmCs(&workflowMainContainer, workflowConfigMaps, workflowSecrets)
	//impl.updateBlobStorageConfig(workflowRequest, &workflowTemplate, storageConfigured, workflowRequest.BlobStorageLogsKey)
	//workflowTemplate.Containers = []v12.Container{workflowMainContainer}
	//workflowTemplate.WorkflowNamePrefix = workflowRequest.WorkflowNamePrefix
	//if !isCi {
	//	workflowTemplate.WfControllerInstanceID = impl.ciCdConfig.WfControllerInstanceID
	//	workflowTemplate.TerminationGracePeriod = impl.ciCdConfig.TerminationGracePeriod
	//}
	//workflowTemplate.ActiveDeadlineSeconds = &workflowRequest.ActiveDeadlineSeconds
	//workflowTemplate.Namespace = workflowRequest.Namespace
	//if workflowRequest.IsExtRun {
	//	configMap := env.Cluster.Config
	//	bearerToken := configMap[k8s.BearerToken]
	//	clusterConfig := &k8s.ClusterConfig{
	//		ClusterName:           env.Cluster.ClusterName,
	//		BearerToken:           bearerToken,
	//		Host:                  env.Cluster.ServerUrl,
	//		InsecureSkipTLSVerify: true,
	//	}
	//	restConfig, err2 := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	//	if err2 != nil {
	//		impl.Logger.Errorw("error in getting rest config from cluster config", "err", err2, "appId", workflowRequest.AppId)
	//		return err2
	//	}
	//	workflowTemplate.ClusterConfig = restConfig
	//} else {
	//	workflowTemplate.ClusterConfig = impl.config
	//}
	workflowExecutor := impl.getWorkflowExecutor(workflowRequest.WorkflowExecutor)
	if workflowExecutor == nil {
		return errors.New("workflow executor not found")
	}
	_, err = workflowExecutor.ExecuteWorkflow(workflowTemplate)
	return err
}

func (impl *WorkflowServiceImpl) createWorkflowTemplate(workflowRequest *WorkflowRequest) (bean3.WorkflowTemplate, error) {
	workflowJson, err := workflowRequest.GetWorkflowJsonAndPVC(impl.ciCdConfig)
	workflowTemplate := workflowRequest.GetWorkflowTemplate(workflowJson, impl.ciCdConfig)
	workflowConfigMaps, workflowSecrets, err := impl.appendGlobalCMCS(workflowRequest, workflowRequest.GetEventTypeForWorkflowRequest(), workflowRequest.GetGlobalCmCsNamePrefix())
	pipelineLevelConfigMaps, pipelineLevelSecrets, err := impl.getConfiguredCmCs(workflowRequest.Pipeline, workflowRequest.StageType)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching pipeline configured cm and cs", "pipelineId", workflowRequest.Pipeline.Id, "err", err)
		return bean3.WorkflowTemplate{}, err
	}
	workflowConfigMaps, workflowSecrets, err = impl.addExistingCmCsInWorkflow(workflowRequest, workflowRequest.CheckForJob(), workflowRequest.CheckForJob(), workflowConfigMaps, workflowSecrets, pipelineLevelConfigMaps, pipelineLevelSecrets, workflowRequest.GetExistingCmCsNamePrefix())

	workflowTemplate.ConfigMaps = workflowConfigMaps
	workflowTemplate.Secrets = workflowSecrets
	workflowTemplate.Volumes = ExtractVolumesFromCmCs(workflowConfigMaps, workflowSecrets)

	nodeConstraints := workflowRequest.GetNodeConstraints(impl.ciCdConfig)
	workflowRequest.AddNodeConstraintsFromConfig(&workflowTemplate, nodeConstraints)
	workflowMainContainer := workflowRequest.GetWorkflowMainContainer(impl.ciCdConfig, workflowJson, workflowTemplate, workflowConfigMaps, workflowSecrets)
	workflowTemplate.Containers = []v12.Container{workflowMainContainer}
	impl.updateBlobStorageConfig(workflowRequest, &workflowTemplate)

	if workflowRequest.Type == bean3.CD_WORKFLOW_PIPELINE_TYPE {
		workflowTemplate.WfControllerInstanceID = impl.ciCdConfig.WfControllerInstanceID
		workflowTemplate.TerminationGracePeriod = impl.ciCdConfig.TerminationGracePeriod
	}

	clusterConfig, err := impl.getClusterConfig(workflowRequest)
	workflowTemplate.ClusterConfig = clusterConfig
	workflowTemplate.WorkflowType = workflowRequest.GetEventTypeForWorkflowRequest()
	return workflowTemplate, nil
}

func (impl *WorkflowServiceImpl) getClusterConfig(workflowRequest *WorkflowRequest) (*rest.Config, error) {
	env := workflowRequest.Env
	if workflowRequest.IsExtRun {
		configMap := env.Cluster.Config
		bearerToken := configMap[k8s.BearerToken]
		clusterConfig := &k8s.ClusterConfig{
			ClusterName:           env.Cluster.ClusterName,
			BearerToken:           bearerToken,
			Host:                  env.Cluster.ServerUrl,
			InsecureSkipTLSVerify: true,
		}
		restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
		if err != nil {
			impl.Logger.Errorw("error in getting rest config from cluster config", "err", err, "appId", workflowRequest.AppId)
			return nil, err
		}
		return restConfig, nil
	}
	return impl.config, nil

}

//
//func (impl *WorkflowServiceImpl) createWorkflowTemplateForCi(workflowRequest *bean3.WorkflowRequest, workflowJson []byte, config *CiCdConfig, nodeLabel map[string]string) (bean3.WorkflowTemplate, error) {
//	storageConfigured := workflowRequest.BlobStorageConfigured
//	workflowTemplate := GetWorkflowTemplate(workflowRequest, workflowJson, config)
//	workflowConfigMaps, workflowSecrets, err := impl.appendGlobalCMCS(workflowRequest, repository2.PIPELINE_TYPE_CI, strconv.Itoa(workflowRequest.WorkflowId)+"-"+CI_WORKFLOW_NAME)
//	workflowTemplate.ConfigMaps = workflowConfigMaps
//	workflowTemplate.Secrets = workflowSecrets
//	AddNodeConstraintsFromConfig(&workflowTemplate, impl.ciCdConfig.CiWorkflowServiceAccount, impl.ciCdConfig.CiTaintKey, impl.ciCdConfig.CiTaintValue, nodeLabel, false, workflowRequest, storageConfigured, workflowConfigMaps, workflowSecrets)
//
//}
//
//func (impl *WorkflowServiceImpl) createWorkflowTemplateForCd(workflowRequest *bean3.WorkflowRequest, workflowJson []byte, config *CiCdConfig, nodeLabel map[string]string) (bean3.WorkflowTemplate, error) {
//	storageConfigured := workflowRequest.BlobStorageConfigured
//	workflowTemplate := GetWorkflowTemplate(workflowRequest, workflowJson, config)
//
//	workflowConfigMaps, workflowSecrets, err := impl.appendGlobalCMCS(workflowRequest, repository2.PIPELINE_TYPE_CD, strconv.Itoa(workflowRequest.WorkflowRunnerId)+"-"+CD_WORKFLOW_NAME)
//	cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err := impl.getConfiguredCmCs(workflowRequest.pipeline, workflowRequest.StageType)
//	if err != nil {
//		impl.Logger.Errorw("error occurred while fetching pipeline configured cm and cs", "pipelineId", workflowRequest.pipeline.Id, "err", err)
//		return bean3.WorkflowTemplate{}, err
//	}
//	workflowConfigMaps, workflowSecrets, err = impl.addExistingCmCsInWorkflow(workflowRequest, false, false, workflowConfigMaps, workflowSecrets, cdPipelineLevelConfigMaps, cdPipelineLevelSecrets)
//
//	workflowTemplate.ConfigMaps = workflowConfigMaps
//	workflowTemplate.Secrets = workflowSecrets
//	workflowTemplate.NodeSelector = map[string]string{impl.ciCdConfig.CdTaintKey: impl.ciCdConfig.CdTaintValue}
//	AddNodeConstraintsFromConfig(&workflowTemplate, impl.ciCdConfig.CdWorkflowServiceAccount, impl.ciCdConfig.CdTaintKey, impl.ciCdConfig.CdTaintValue, nodeLabel, false, workflowRequest, storageConfigured, workflowConfigMaps, workflowSecrets)
//	workflowTemplate.Volumes = ExtractVolumesFromCmCs(workflowConfigMaps, workflowSecrets)
//
//}
//
//func (impl *WorkflowServiceImpl) createWorkflowTemplateForJob(workflowRequest *bean3.WorkflowRequest, workflowJson []byte, config *CiCdConfig, nodeLabel map[string]string) (bean3.WorkflowTemplate, error) {
//	storageConfigured := workflowRequest.BlobStorageConfigured
//	workflowTemplate := GetWorkflowTemplate(workflowRequest, workflowJson, config)
//	workflowConfigMaps, workflowSecrets, err := impl.appendGlobalCMCS(workflowRequest, repository2.PIPELINE_TYPE_CI, strconv.Itoa(workflowRequest.WorkflowId)+"-"+CI_WORKFLOW_NAME)
//	workflowConfigMaps, workflowSecrets, err = impl.addExistingCmCsInWorkflow(workflowRequest, true, true, workflowConfigMaps, workflowSecrets, nil, nil)
//	workflowTemplate.ConfigMaps = workflowConfigMaps
//	workflowTemplate.Secrets = workflowSecrets
//	AddNodeConstraintsFromConfig(&workflowTemplate, impl.ciCdConfig.CiWorkflowServiceAccount, impl.ciCdConfig.CiTaintKey, impl.ciCdConfig.CiTaintValue, nodeLabel, false, workflowRequest, storageConfigured, workflowConfigMaps, workflowSecrets)
//
//}

func (impl *WorkflowServiceImpl) appendGlobalCMCS(workflowRequest *WorkflowRequest, pipelineType string, globalConfigNamePrefix string) ([]bean.ConfigSecretMap, []bean.ConfigSecretMap, error) {
	var workflowConfigMaps []bean.ConfigSecretMap
	var workflowSecrets []bean.ConfigSecretMap
	if !workflowRequest.IsExtRun {
		// inject global variables only if IsExtRun is false
		globalCmCsConfigs, err := impl.globalCMCSService.FindAllActiveByPipelineType(pipelineType)
		if err != nil {
			impl.Logger.Errorw("error in getting all global cm/cs config", "err", err)
			return nil, nil, err
		}
		for i := range globalCmCsConfigs {
			globalCmCsConfigs[i].Name = strings.ToLower(globalCmCsConfigs[i].Name) + "-" + globalConfigNamePrefix
		}
		workflowConfigMaps, workflowSecrets, err = GetFromGlobalCmCsDtos(globalCmCsConfigs)
		if err != nil {
			impl.Logger.Errorw("error in creating templates for global secrets", "err", err)
			return nil, nil, err
		}
	}
	return workflowConfigMaps, workflowSecrets, nil
}

//	func (impl *WorkflowServiceImpl) GetWorkflowTemplateForCi(workflowRequest *bean3.WorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment, appLabels map[string]string) (*bean3.WorkflowTemplate, error) {
//		containerEnvVariables := []v12.EnvVar{{Name: "IMAGE_SCANNER_ENDPOINT", Value: impl.ciCdConfig.ImageScannerEndpoint}}
//		containerEnvVariables = GetContainerEnvVariables(impl.ciCdConfig, containerEnvVariables)
//		nodeLabel, err := getNodeLabel(impl.ciCdConfig, true)
//		if err != nil {
//			impl.Logger.Errorw("Error in getting nodeLabel", err)
//			return nil, err
//		}
//		pvc := appLabels[strings.ToLower(fmt.Sprintf("%s-%s", CI_NODE_PVC_PIPELINE_PREFIX, workflowRequest.PipelineName))]
//		if len(pvc) == 0 {
//			pvc = appLabels[CI_NODE_PVC_ALL_ENV]
//		}
//		if len(pvc) != 0 {
//			workflowRequest.IsPvcMounted = true
//			workflowRequest.IgnoreDockerCachePush = true
//			workflowRequest.IgnoreDockerCachePull = true
//		}
//		eventType := ciStage
//		if env != nil && env.Id != 0 {
//			workflowRequest.EnvironmentId = env.Id
//			workflowRequest.IsExtRun = true
//		}
//	}
//
//	func (impl *WorkflowServiceImpl) GetWorkflowTemplateForCd(workflowRequest *bean3.WorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment, appLabels map[string]string) (*bean3.WorkflowTemplate, error) {
//		containerEnvVariables := []v12.EnvVar{}
//		containerEnvVariables = GetContainerEnvVariables(impl.ciCdConfig, containerEnvVariables)
//		nodeLabel, err := getNodeLabel(impl.ciCdConfig, false)
//		if err != nil {
//			impl.Logger.Errorw("Error in getting nodeLabel", err)
//			return nil, err
//		}
//		if (workflowRequest.StageType == PRE && pipeline.RunPreStageInEnv) || (workflowRequest.StageType == POST && pipeline.RunPostStageInEnv) {
//			workflowRequest.IsExtRun = true
//		}
//		eventType := cdStage
//
// }
func (impl *WorkflowServiceImpl) addExistingCmCsInWorkflow(workflowRequest *WorkflowRequest, isJob bool, allowAll bool, workflowConfigMaps []bean.ConfigSecretMap, workflowSecrets []bean.ConfigSecretMap, cdPipelineLevelConfigMaps map[string]bool, cdPipelineLevelSecrets map[string]bool, namePrefix string) ([]bean.ConfigSecretMap, []bean.ConfigSecretMap, error) {

	existingConfigMap, existingSecrets, err := impl.appService.GetCmSecretNew(workflowRequest.AppId, workflowRequest.EnvironmentId, isJob)
	if err != nil {
		impl.Logger.Errorw("failed to get configmap data", "err", err)
		return nil, nil, err
	}
	impl.Logger.Debugw("existing cm", "cm", existingConfigMap, "secrets", existingSecrets)

	for _, cm := range existingConfigMap.Maps {
		// HERE we are allowing all existingSecrets in case of JOB
		if _, ok := cdPipelineLevelConfigMaps[cm.Name]; ok || allowAll {
			if !cm.External {
				cm.Name = cm.Name + "-" + namePrefix
			}
			workflowConfigMaps = append(workflowConfigMaps, cm)
		}
	}
	for _, secret := range existingSecrets.Secrets {
		// HERE we are allowing all existingSecrets in case of JOB
		if _, ok := cdPipelineLevelSecrets[secret.Name]; ok || allowAll {
			if !secret.External {
				secret.Name = secret.Name + "-" + namePrefix
			}
			workflowSecrets = append(workflowSecrets, *secret)
		}
	}
	return workflowConfigMaps, workflowSecrets, nil
}

func (impl *WorkflowServiceImpl) getConfiguredCmCs(pipeline *pipelineConfig.Pipeline, stage string) (map[string]bool, map[string]bool, error) {

	cdPipelineLevelConfigMaps := make(map[string]bool)
	cdPipelineLevelSecrets := make(map[string]bool)

	if stage == "PRE" {
		preStageConfigMapSecretsJson := pipeline.PreStageConfigMapSecretNames
		preStageConfigmapSecrets := bean2.PreStageConfigMapSecretNames{}
		err := json.Unmarshal([]byte(preStageConfigMapSecretsJson), &preStageConfigmapSecrets)
		if err != nil {
			return cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err
		}
		for _, cm := range preStageConfigmapSecrets.ConfigMaps {
			cdPipelineLevelConfigMaps[cm] = true
		}
		for _, secret := range preStageConfigmapSecrets.Secrets {
			cdPipelineLevelSecrets[secret] = true
		}
	}
	if stage == "POST" {
		postStageConfigMapSecretsJson := pipeline.PostStageConfigMapSecretNames
		postStageConfigmapSecrets := bean2.PostStageConfigMapSecretNames{}
		err := json.Unmarshal([]byte(postStageConfigMapSecretsJson), &postStageConfigmapSecrets)
		if err != nil {
			return cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err
		}
		for _, cm := range postStageConfigmapSecrets.ConfigMaps {
			cdPipelineLevelConfigMaps[cm] = true
		}
		for _, secret := range postStageConfigmapSecrets.Secrets {
			cdPipelineLevelSecrets[secret] = true
		}
	}
	return cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, nil
}

func (impl *WorkflowServiceImpl) updateBlobStorageConfig(workflowRequest *WorkflowRequest, workflowTemplate *bean3.WorkflowTemplate) {
	workflowTemplate.BlobStorageConfigured = workflowRequest.BlobStorageConfigured && (impl.ciCdConfig.UseBlobStorageConfigInCdWorkflow || !workflowRequest.IsExtRun)
	workflowTemplate.BlobStorageS3Config = workflowRequest.BlobStorageS3Config
	workflowTemplate.AzureBlobConfig = workflowRequest.AzureBlobConfig
	workflowTemplate.GcpBlobConfig = workflowRequest.GcpBlobConfig
	workflowTemplate.CloudStorageKey = workflowRequest.BlobStorageLogsKey
}

func (impl *WorkflowServiceImpl) getWorkflowExecutor(executorType pipelineConfig.WorkflowExecutorType) WorkflowExecutor {
	if executorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF {
		return impl.argoWorkflowExecutor
	} else if executorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
		return impl.systemWorkflowExecutor
	}
	impl.Logger.Warnw("workflow executor not found", "type", executorType)
	return nil
}
func (impl *WorkflowServiceImpl) GetWorkflow(name string, namespace string, isExt bool, environment *repository.Environment) (*v1alpha1.Workflow, error) {
	impl.Logger.Debug("getting wf", name)
	wfClient, err := impl.getWfClient(environment, namespace, isExt)

	if err != nil {
		return nil, err
	}

	workflow, err := wfClient.Get(context.Background(), name, v1.GetOptions{})
	return workflow, err
}

func (impl *WorkflowServiceImpl) TerminateWorkflow(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExt bool, environment *repository.Environment) error {
	impl.Logger.Debugw("terminating wf", "name", name)
	var err error
	if executorType != "" {
		workflowExecutor := impl.getWorkflowExecutor(executorType)
		err = workflowExecutor.TerminateWorkflow(name, namespace, restConfig)
	} else {
		wfClient, err := impl.getWfClient(environment, namespace, isExt)
		if err != nil {
			return err
		}
		err = util.TerminateWorkflow(context.Background(), wfClient, name)
	}
	return err
}
func (impl *WorkflowServiceImpl) getRuntimeEnvClientInstance(environment *repository.Environment) (v1alpha12.WorkflowInterface, error) {
	restConfig, err, _ := impl.k8sCommonService.GetRestConfigByClusterId(context.Background(), environment.ClusterId)
	if err != nil {
		impl.Logger.Errorw("error in getting rest config by cluster id", "err", err)
		return nil, err
	}
	wfClient, err := GetClientInstance(restConfig, environment.Namespace)
	if err != nil {
		impl.Logger.Errorw("error in getting wfClient", "err", err)
		return nil, err
	}
	return wfClient, nil
}

func (impl *WorkflowServiceImpl) getWfClient(environment *repository.Environment, namespace string, isExt bool) (v1alpha12.WorkflowInterface, error) {
	var wfClient v1alpha12.WorkflowInterface
	var err error
	if isExt {
		wfClient, err = impl.getRuntimeEnvClientInstance(environment)
		if err != nil {
			impl.Logger.Errorw("cannot build wf client", "err", err)
			return nil, err
		}
	} else {
		wfClient, err = GetClientInstance(impl.config, namespace)
		if err != nil {
			impl.Logger.Errorw("cannot build wf client", "err", err)
			return nil, err
		}
	}
	return wfClient, nil
}
