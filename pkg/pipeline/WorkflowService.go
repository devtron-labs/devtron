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
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"strings"
)

// TODO: move isCi/isJob to workflowRequest

type WorkflowService interface {
	SubmitWorkflow(workflowRequest *types.WorkflowRequest) (*unstructured.UnstructuredList, error)
	//DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(name string, namespace string, isExt bool, environment *repository.Environment) (*v1alpha1.Workflow, error)
	//ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	//UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	TerminateWorkflow(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExt bool, environment *repository.Environment) error
}

type WorkflowServiceImpl struct {
	Logger                 *zap.SugaredLogger
	config                 *rest.Config
	ciCdConfig             *types.CiCdConfig
	appService             app.AppService
	envRepository          repository.EnvironmentRepository
	globalCMCSService      GlobalCMCSService
	argoWorkflowExecutor   executors.ArgoWorkflowExecutor
	systemWorkflowExecutor executors.SystemWorkflowExecutor
	k8sUtil                *k8s.K8sUtil
	k8sCommonService       k8s2.K8sCommonService
}

// TODO: Move to bean

func NewWorkflowServiceImpl(Logger *zap.SugaredLogger, envRepository repository.EnvironmentRepository, ciCdConfig *types.CiCdConfig,
	appService app.AppService, globalCMCSService GlobalCMCSService, argoWorkflowExecutor executors.ArgoWorkflowExecutor,
	k8sUtil *k8s.K8sUtil,
	systemWorkflowExecutor executors.SystemWorkflowExecutor, k8sCommonService k8s2.K8sCommonService) (*WorkflowServiceImpl, error) {
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
	CI_NODE_SELECTOR_APP_LABEL_KEY = "devtron.ai/node-selector"

	preCdStage  = "preCD"
	postCdStage = "postCD"
)

func (impl *WorkflowServiceImpl) SubmitWorkflow(workflowRequest *types.WorkflowRequest) (*unstructured.UnstructuredList, error) {
	workflowTemplate, err := impl.createWorkflowTemplate(workflowRequest)
	if err != nil {
		return nil, err
	}
	workflowExecutor := impl.getWorkflowExecutor(workflowRequest.WorkflowExecutor)
	if workflowExecutor == nil {
		return nil, errors.New("workflow executor not found")
	}
	createdWf, err := workflowExecutor.ExecuteWorkflow(workflowTemplate)
	return createdWf, err
}

func (impl *WorkflowServiceImpl) createWorkflowTemplate(workflowRequest *types.WorkflowRequest) (bean3.WorkflowTemplate, error) {
	workflowJson, err := workflowRequest.GetWorkflowJson(impl.ciCdConfig)
	if err != nil {
		impl.Logger.Errorw("error occurred while getting workflow json", "err", err)
		return bean3.WorkflowTemplate{}, err
	}
	workflowTemplate := workflowRequest.GetWorkflowTemplate(workflowJson, impl.ciCdConfig)
	workflowConfigMaps, workflowSecrets, err := impl.appendGlobalCMCS(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("error occurred while appending CmCs", "err", err)
		return bean3.WorkflowTemplate{}, err
	}
	workflowConfigMaps, workflowSecrets, err = impl.addExistingCmCsInWorkflow(workflowRequest, workflowConfigMaps, workflowSecrets)
	if err != nil {
		impl.Logger.Errorw("error occurred while adding existing CmCs", "err", err)
		return bean3.WorkflowTemplate{}, err
	}

	workflowTemplate.ConfigMaps = workflowConfigMaps
	workflowTemplate.Secrets = workflowSecrets
	workflowTemplate.Volumes = executors.ExtractVolumesFromCmCs(workflowConfigMaps, workflowSecrets)

	workflowRequest.AddNodeConstraintsFromConfig(&workflowTemplate, impl.ciCdConfig)
	workflowMainContainer, err := workflowRequest.GetWorkflowMainContainer(impl.ciCdConfig, workflowJson, &workflowTemplate, workflowConfigMaps, workflowSecrets)

	if err != nil {
		impl.Logger.Errorw("error occurred while getting workflow main container", "err", err)
		return bean3.WorkflowTemplate{}, err
	}

	workflowTemplate.Containers = []v12.Container{workflowMainContainer}
	impl.updateBlobStorageConfig(workflowRequest, &workflowTemplate)
	if workflowRequest.Type == bean3.CI_WORKFLOW_PIPELINE_TYPE || workflowRequest.Type == bean3.JOB_WORKFLOW_PIPELINE_TYPE {
		nodeSelector := impl.getAppLabelNodeSelector(workflowRequest)
		if nodeSelector != nil {
			workflowTemplate.NodeSelector = nodeSelector
		}
	}
	if workflowRequest.Type == bean3.CD_WORKFLOW_PIPELINE_TYPE {
		workflowTemplate.WfControllerInstanceID = impl.ciCdConfig.WfControllerInstanceID
		workflowTemplate.TerminationGracePeriod = impl.ciCdConfig.TerminationGracePeriod
	}

	clusterConfig, err := impl.getClusterConfig(workflowRequest)
	workflowTemplate.ClusterConfig = clusterConfig
	workflowTemplate.WorkflowType = workflowRequest.GetWorkflowTypeForWorkflowRequest()
	return workflowTemplate, nil
}

func (impl *WorkflowServiceImpl) getClusterConfig(workflowRequest *types.WorkflowRequest) (*rest.Config, error) {
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

func (impl *WorkflowServiceImpl) appendGlobalCMCS(workflowRequest *types.WorkflowRequest) ([]bean.ConfigSecretMap, []bean.ConfigSecretMap, error) {
	var workflowConfigMaps []bean.ConfigSecretMap
	var workflowSecrets []bean.ConfigSecretMap
	if !workflowRequest.IsExtRun {
		// inject global variables only if IsExtRun is false
		globalCmCsConfigs, err := impl.globalCMCSService.FindAllActiveByPipelineType(workflowRequest.GetEventTypeForWorkflowRequest())
		if err != nil {
			impl.Logger.Errorw("error in getting all global cm/cs config", "err", err)
			return nil, nil, err
		}
		for i := range globalCmCsConfigs {
			globalCmCsConfigs[i].Name = strings.ToLower(globalCmCsConfigs[i].Name) + "-" + workflowRequest.GetGlobalCmCsNamePrefix()
		}
		workflowConfigMaps, workflowSecrets, err = executors.GetFromGlobalCmCsDtos(globalCmCsConfigs)
		if err != nil {
			impl.Logger.Errorw("error in creating templates for global secrets", "err", err)
			return nil, nil, err
		}
	}
	return workflowConfigMaps, workflowSecrets, nil
}

func (impl *WorkflowServiceImpl) addExistingCmCsInWorkflow(workflowRequest *types.WorkflowRequest, workflowConfigMaps []bean.ConfigSecretMap, workflowSecrets []bean.ConfigSecretMap) ([]bean.ConfigSecretMap, []bean.ConfigSecretMap, error) {

	pipelineLevelConfigMaps, pipelineLevelSecrets, err := workflowRequest.GetConfiguredCmCs()
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching pipeline configured cm and cs", "pipelineId", workflowRequest.Pipeline.Id, "err", err)
		return nil, nil, err
	}
	isJob := workflowRequest.CheckForJob()
	allowAll := isJob
	namePrefix := workflowRequest.GetExistingCmCsNamePrefix()
	existingConfigMap, existingSecrets, err := impl.appService.GetCmSecretNew(workflowRequest.AppId, workflowRequest.EnvironmentId, isJob)
	if err != nil {
		impl.Logger.Errorw("failed to get configmap data", "err", err)
		return nil, nil, err
	}
	impl.Logger.Debugw("existing cm", "cm", existingConfigMap, "secrets", existingSecrets)

	for _, cm := range existingConfigMap.Maps {
		// HERE we are allowing all existingSecrets in case of JOB
		if _, ok := pipelineLevelConfigMaps[cm.Name]; ok || allowAll {
			if !cm.External {
				cm.Name = cm.Name + "-" + namePrefix
			}
			workflowConfigMaps = append(workflowConfigMaps, cm)
		}
	}
	for _, secret := range existingSecrets.Secrets {
		// HERE we are allowing all existingSecrets in case of JOB
		if _, ok := pipelineLevelSecrets[secret.Name]; ok || allowAll {
			if !secret.External {
				secret.Name = secret.Name + "-" + namePrefix
			}
			workflowSecrets = append(workflowSecrets, *secret)
		}
	}

	//internally inducing BlobStorageCmName and BlobStorageSecretName for getting logs, caches and artifacts from
	//in-cluster configured blob storage, if USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW = false and isExt = true
	if workflowRequest.UseExternalClusterBlob {
		workflowConfigMaps, workflowSecrets = impl.addExtBlobStorageCmCsInResponse(workflowConfigMaps, workflowSecrets)
	}

	return workflowConfigMaps, workflowSecrets, nil
}
func (impl *WorkflowServiceImpl) addExtBlobStorageCmCsInResponse(workflowConfigMaps []bean.ConfigSecretMap, workflowSecrets []bean.ConfigSecretMap) ([]bean.ConfigSecretMap, []bean.ConfigSecretMap) {
	blobDetailsConfigMap := bean.ConfigSecretMap{
		Name:     impl.ciCdConfig.ExtBlobStorageCmName,
		Type:     "environment",
		External: true,
	}
	workflowConfigMaps = append(workflowConfigMaps, blobDetailsConfigMap)

	blobDetailsSecret := bean.ConfigSecretMap{
		Name:     impl.ciCdConfig.ExtBlobStorageSecretName,
		Type:     "environment",
		External: true,
	}
	workflowSecrets = append(workflowSecrets, blobDetailsSecret)
	return workflowConfigMaps, workflowSecrets
}

func (impl *WorkflowServiceImpl) updateBlobStorageConfig(workflowRequest *types.WorkflowRequest, workflowTemplate *bean3.WorkflowTemplate) {
	workflowTemplate.BlobStorageConfigured = workflowRequest.BlobStorageConfigured && (workflowRequest.CheckBlobStorageConfig(impl.ciCdConfig) || !workflowRequest.IsExtRun)
	workflowTemplate.BlobStorageS3Config = workflowRequest.BlobStorageS3Config
	workflowTemplate.AzureBlobConfig = workflowRequest.AzureBlobConfig
	workflowTemplate.GcpBlobConfig = workflowRequest.GcpBlobConfig
	workflowTemplate.CloudStorageKey = workflowRequest.BlobStorageLogsKey
}

func (impl *WorkflowServiceImpl) getAppLabelNodeSelector(workflowRequest *types.WorkflowRequest) map[string]string {
	// node selector
	if val, ok := workflowRequest.AppLabels[CI_NODE_SELECTOR_APP_LABEL_KEY]; ok && !(workflowRequest.CheckForJob() && workflowRequest.IsExtRun) {
		var nodeSelectors map[string]string
		// Unmarshal or Decode the JSON to the interface.
		err := json.Unmarshal([]byte(val), &nodeSelectors)
		if err != nil {
			impl.Logger.Errorw("err in unmarshalling nodeSelectors", "err", err, "val", val)
			return nil
		}
		return nodeSelectors
	}
	return nil
}

func (impl *WorkflowServiceImpl) getWorkflowExecutor(executorType pipelineConfig.WorkflowExecutorType) executors.WorkflowExecutor {
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
		if restConfig == nil {
			restConfig = impl.config
		}
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
	wfClient, err := executors.GetClientInstance(restConfig, environment.Namespace)
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
		wfClient, err = executors.GetClientInstance(impl.config, namespace)
		if err != nil {
			impl.Logger.Errorw("cannot build wf client", "err", err)
			return nil, err
		}
	}
	return wfClient, nil
}
