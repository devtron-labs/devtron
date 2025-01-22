/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/util"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/config/read"
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"net/http"
	"strconv"
	"strings"
)

// TODO: move isCi/isJob to workflowRequest

type WorkflowService interface {
	SubmitWorkflow(workflowRequest *types.WorkflowRequest) (*unstructured.UnstructuredList, string, error)
	// DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(executorType cdWorkflow.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config) (*unstructured.UnstructuredList, error)
	GetWorkflowStatus(executorType cdWorkflow.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config) (*types.WorkflowStatus, error)
	// ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	// UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	TerminateWorkflow(cancelWfDtoRequest *types.CancelWfRequestDto) error
	TerminateDanglingWorkflows(cancelWfDtoRequest *types.CancelWfRequestDto) error
}

type WorkflowServiceImpl struct {
	Logger                 *zap.SugaredLogger
	config                 *rest.Config
	ciCdConfig             *types.CiCdConfig
	configMapService       read.ConfigReadService
	appService             app.AppService
	envRepository          repository2.EnvironmentRepository
	globalCMCSService      GlobalCMCSService
	argoWorkflowExecutor   executors.ArgoWorkflowExecutor
	systemWorkflowExecutor executors.SystemWorkflowExecutor
	k8sUtil                *k8s.K8sServiceImpl
	k8sCommonService       k8s2.K8sCommonService
	infraProvider          infraProviders.InfraProvider
}

// TODO: Move to bean

func NewWorkflowServiceImpl(Logger *zap.SugaredLogger, envRepository repository2.EnvironmentRepository, ciCdConfig *types.CiCdConfig,
	appService app.AppService, globalCMCSService GlobalCMCSService, argoWorkflowExecutor executors.ArgoWorkflowExecutor,
	k8sUtil *k8s.K8sServiceImpl,
	systemWorkflowExecutor executors.SystemWorkflowExecutor, k8sCommonService k8s2.K8sCommonService,
	infraProvider infraProviders.InfraProvider) (*WorkflowServiceImpl, error) {
	commonWorkflowService := &WorkflowServiceImpl{
		Logger:                 Logger,
		ciCdConfig:             ciCdConfig,
		appService:             appService,
		envRepository:          envRepository,
		globalCMCSService:      globalCMCSService,
		argoWorkflowExecutor:   argoWorkflowExecutor,
		k8sUtil:                k8sUtil,
		systemWorkflowExecutor: systemWorkflowExecutor,
		k8sCommonService:       k8sCommonService,
		infraProvider:          infraProvider,
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

func (impl *WorkflowServiceImpl) SubmitWorkflow(workflowRequest *types.WorkflowRequest) (*unstructured.UnstructuredList, string, error) {
	workflowTemplate, err := impl.createWorkflowTemplate(workflowRequest)
	if err != nil {
		return nil, "", err
	}
	workflowExecutor := impl.getWorkflowExecutor(workflowRequest.WorkflowExecutor)
	if workflowExecutor == nil {
		return nil, "", errors.New("workflow executor not found")
	}
	createdWf, err := workflowExecutor.ExecuteWorkflow(workflowTemplate)
	jobHelmPackagePath := "" // due to ENT
	return createdWf, jobHelmPackagePath, err
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

	workflowRequest.AddNodeConstraintsFromConfig(&workflowTemplate, impl.ciCdConfig)
	infraConfiguration := &v1.InfraConfig{}
	shouldAddExistingCmCsInWorkflow := impl.shouldAddExistingCmCsInWorkflow(workflowRequest)
	if workflowRequest.Type == bean3.CI_WORKFLOW_PIPELINE_TYPE || workflowRequest.Type == bean3.JOB_WORKFLOW_PIPELINE_TYPE {
		nodeSelector := impl.getAppLabelNodeSelector(workflowRequest)
		if nodeSelector != nil {
			workflowTemplate.NodeSelector = nodeSelector
		}
		infraGetterRequest := infraGetters.NewInfraRequest(workflowRequest.Scope).
			WithAppId(workflowRequest.AppId).
			WithEnvId(workflowRequest.EnvironmentId).
			WithPlatform(v1.RUNNER_PLATFORM)
		infraGetter, _ := impl.infraProvider.GetInfraProvider(workflowRequest.Type)
		infraConfigurations, err := infraGetter.GetConfigurationsByScopeAndTargetPlatforms(infraGetterRequest)
		if err != nil {
			impl.Logger.Errorw("error occurred while getting infra config", "infraGetterRequest", infraGetterRequest, "err", err)
			return bean3.WorkflowTemplate{}, err
		}
		impl.Logger.Debugw("infra config for workflow", "infraConfigurations", infraConfigurations, "infraGetterRequest", infraGetterRequest)
		infraConfiguration = infraConfigurations[v1.RUNNER_PLATFORM]
		infraConfigMaps, infraSecrets, err := impl.prepareCmCsForWorkflowTemplate(workflowRequest, infraConfiguration.ConfigMaps, infraConfiguration.Secrets)
		if err != nil {
			impl.Logger.Errorw("error occurred while preparing build infra cm/ cs for workflow template", "err", err)
			return bean3.WorkflowTemplate{}, err
		}
		workflowConfigMaps = append(workflowConfigMaps, infraConfigMaps...)
		workflowSecrets = append(workflowSecrets, infraSecrets...)
		workflowRequest.AddInfraConfigurations(&workflowTemplate, infraConfiguration)
		err = infraGetter.SaveInfraConfigHistorySnapshot(workflowRequest.WorkflowId, workflowRequest.TriggeredBy, infraConfigurations)
		if err != nil {
			impl.Logger.Errorw("error occurred while saving infra config history snapshot", "err", err, "infraConfigurations", infraConfigurations, "workflowId", workflowRequest.WorkflowId)
		}
	} else {
		if shouldAddExistingCmCsInWorkflow {
			workflowConfigMaps, workflowSecrets, err = impl.addExistingCmCsInWorkflowForCDStage(workflowRequest, workflowConfigMaps, workflowSecrets)
			if err != nil {
				impl.Logger.Errorw("error occurred while adding existing cm/ cs", "err", err)
				return bean3.WorkflowTemplate{}, err
			}
			cdStageConfigMaps, cdStageSecrets, err := impl.prepareCmCsForWorkflowTemplate(workflowRequest, workflowConfigMaps, workflowSecrets)
			if err != nil {
				impl.Logger.Errorw("error occurred while preparing cd stage cm/ cs for workflow template", "err", err)
				return bean3.WorkflowTemplate{}, err
			}
			workflowConfigMaps = append(workflowConfigMaps, cdStageConfigMaps...)
			workflowSecrets = append(workflowSecrets, cdStageSecrets...)
		}
	}
	// internally inducing BlobStorageCmName and BlobStorageSecretName for getting logs, caches and artifacts from
	// in-cluster configured blob storage, if USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW = false and isExt = true
	if shouldAddExistingCmCsInWorkflow && workflowRequest.UseExternalClusterBlob {
		workflowConfigMaps, workflowSecrets = impl.addExtBlobStorageCmCsInResponse(workflowConfigMaps, workflowSecrets)
	}
	workflowTemplate.ConfigMaps = workflowConfigMaps
	workflowTemplate.Secrets = workflowSecrets
	workflowTemplate.Volumes = executors.ExtractVolumesFromCmCs(workflowConfigMaps, workflowSecrets)

	workflowMainContainer, err := workflowRequest.GetWorkflowMainContainer(impl.ciCdConfig, infraConfiguration, workflowJson, &workflowTemplate, workflowConfigMaps, workflowSecrets)
	if err != nil {
		impl.Logger.Errorw("error occurred while getting workflow main container", "err", err)
		return bean3.WorkflowTemplate{}, err
	}
	workflowTemplate.Containers = []v12.Container{workflowMainContainer}
	impl.updateBlobStorageConfig(workflowRequest, &workflowTemplate)
	if workflowRequest.Type == bean3.CD_WORKFLOW_PIPELINE_TYPE {
		workflowTemplate.WfControllerInstanceID = impl.ciCdConfig.WfControllerInstanceID
	}
	workflowTemplate.TerminationGracePeriod = impl.ciCdConfig.TerminationGracePeriod

	clusterConfig, err := impl.getClusterConfig(workflowRequest)
	workflowTemplate.ClusterConfig = clusterConfig
	workflowTemplate.WorkflowType = workflowRequest.GetWorkflowTypeForWorkflowRequest()
	return workflowTemplate, nil
}

func (impl *WorkflowServiceImpl) shouldAddExistingCmCsInWorkflow(workflowRequest *types.WorkflowRequest) bool {
	// CmCs are not added for CI_JOB if IgnoreCmCsInCiJob is true
	if workflowRequest.CiPipelineType == string(bean2.CI_JOB) && impl.ciCdConfig.IgnoreCmCsInCiJob {
		return false
	}
	return true
}

func (impl *WorkflowServiceImpl) getClusterConfig(workflowRequest *types.WorkflowRequest) (*rest.Config, error) {
	env := workflowRequest.Env
	if workflowRequest.IsExtRun {
		configMap := env.Cluster.Config
		bearerToken := configMap[commonBean.BearerToken]
		clusterConfig := &k8s.ClusterConfig{
			ClusterName:           env.Cluster.ClusterName,
			BearerToken:           bearerToken,
			Host:                  env.Cluster.ServerUrl,
			InsecureSkipTLSVerify: env.Cluster.InsecureSkipTlsVerify,
		}
		if !env.Cluster.InsecureSkipTlsVerify {
			clusterConfig.KeyData = configMap[commonBean.TlsKey]
			clusterConfig.CertData = configMap[commonBean.CertData]
			clusterConfig.CAData = configMap[commonBean.CertificateAuthorityData]
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
		globalCmCsConfigs, err := impl.globalCMCSService.FindAllActiveByPipelineType(workflowRequest.GetPipelineTypeForGlobalCMCS())
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

func (impl *WorkflowServiceImpl) addExistingCmCsInWorkflowForCDStage(workflowRequest *types.WorkflowRequest, workflowConfigMaps []bean.ConfigSecretMap, workflowSecrets []bean.ConfigSecretMap) ([]bean.ConfigSecretMap, []bean.ConfigSecretMap, error) {
	existingConfigMap, existingSecrets, err := impl.configMapService.GetCmCsForPrePostStageTrigger(workflowRequest.Scope, workflowRequest.AppId, workflowRequest.EnvironmentId, false)
	if err != nil {
		impl.Logger.Errorw("failed to get configmap data", "err", err)
		return nil, nil, err
	}
	impl.Logger.Debugw("existing cm", "cm", existingConfigMap, "secrets", existingSecrets)
	return workflowConfigMaps, workflowSecrets, nil
}

func (impl *WorkflowServiceImpl) prepareCmCsForWorkflowTemplate(workflowRequest *types.WorkflowRequest, workflowConfigMaps []bean.ConfigSecretMap, workflowSecrets []bean.ConfigSecretMap) ([]bean.ConfigSecretMap, []bean.ConfigSecretMap, error) {
	modifiedWorkflowConfigMaps := make([]bean.ConfigSecretMap, 0)
	modifiedWorkflowSecrets := make([]bean.ConfigSecretMap, 0)
	pipelineLevelConfigMaps, pipelineLevelSecrets, err := workflowRequest.GetConfiguredCmCs()
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching pipeline configured cm and cs", "pipelineId", workflowRequest.Pipeline.Id, "err", err)
		return nil, nil, err
	}
	allowAll := workflowRequest.IsDevtronJob() || workflowRequest.IsDevtronCI()
	namePrefix := workflowRequest.GetExistingCmCsNamePrefix()
	for _, cm := range workflowConfigMaps {
		// HERE we are allowing all existingSecrets in case of JOB/ BUILD Infra
		if _, ok := pipelineLevelConfigMaps[cm.Name]; ok || allowAll {
			if !cm.External {
				cm.Name = fmt.Sprintf("%s-cm-%s", cm.Name, namePrefix)
			}
			modifiedWorkflowConfigMaps = append(modifiedWorkflowConfigMaps, cm)
		}
	}
	for _, secret := range workflowSecrets {
		// HERE we are allowing all existingSecrets in case of JOB/ BUILD Infra
		if _, ok := pipelineLevelSecrets[secret.Name]; ok || allowAll {
			if !secret.External {
				secret.Name = fmt.Sprintf("%s-cs-%s", secret.Name, namePrefix)
			}
			modifiedWorkflowSecrets = append(modifiedWorkflowSecrets, secret)
		}
	}
	return modifiedWorkflowConfigMaps, modifiedWorkflowSecrets, nil
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
	if val, ok := workflowRequest.AppLabels[CI_NODE_SELECTOR_APP_LABEL_KEY]; ok && !(workflowRequest.IsDevtronJob() && workflowRequest.IsExtRun) {
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

func (impl *WorkflowServiceImpl) getWorkflowExecutor(executorType cdWorkflow.WorkflowExecutorType) executors.WorkflowExecutor {
	if executorType == "" || executorType == cdWorkflow.WORKFLOW_EXECUTOR_TYPE_AWF {
		return impl.argoWorkflowExecutor
	} else if executorType == cdWorkflow.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
		return impl.systemWorkflowExecutor
	}
	impl.Logger.Warnw("workflow executor not found", "type", executorType)
	return nil
}
func (impl *WorkflowServiceImpl) GetWorkflow(executorType cdWorkflow.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config) (*unstructured.UnstructuredList, error) {
	impl.Logger.Debug("getting wf", name)
	workflowExecutor := impl.getWorkflowExecutor(executorType)
	if workflowExecutor == nil {
		return nil, errors.New("workflow executor not found")
	}
	if restConfig == nil {
		restConfig = impl.config
	}
	return workflowExecutor.GetWorkflow(name, namespace, restConfig)
}

func (impl *WorkflowServiceImpl) GetWorkflowStatus(executorType cdWorkflow.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config) (*types.WorkflowStatus, error) {
	impl.Logger.Debug("getting wf", name)
	workflowExecutor := impl.getWorkflowExecutor(executorType)
	if workflowExecutor == nil {
		return nil, errors.New("workflow executor not found")
	}
	if restConfig == nil {
		restConfig = impl.config
	}
	wfStatus, err := workflowExecutor.GetWorkflowStatus(name, namespace, restConfig)
	return wfStatus, err
}

func (impl *WorkflowServiceImpl) TerminateWorkflow(cancelWfDtoRequest *types.CancelWfRequestDto) error {
	impl.Logger.Debugw("terminating wf", "name", cancelWfDtoRequest.WorkflowName)
	var err error
	if cancelWfDtoRequest.ExecutorType != "" {
		workflowExecutor := impl.getWorkflowExecutor(cancelWfDtoRequest.ExecutorType)
		if workflowExecutor == nil {
			return errors.New("workflow executor not found")
		}
		if cancelWfDtoRequest.RestConfig == nil {
			cancelWfDtoRequest.RestConfig = impl.config
		}
		err = workflowExecutor.TerminateWorkflow(cancelWfDtoRequest.WorkflowName, cancelWfDtoRequest.Namespace, cancelWfDtoRequest.RestConfig)
	} else {
		wfClient, err := impl.getWfClient(cancelWfDtoRequest.Environment, cancelWfDtoRequest.Namespace, cancelWfDtoRequest.IsExt)
		if err != nil {
			return err
		}
		err = util.TerminateWorkflow(context.Background(), wfClient, cancelWfDtoRequest.WorkflowName)
	}
	return err
}

func (impl *WorkflowServiceImpl) TerminateDanglingWorkflows(cancelWfDtoRequest *types.CancelWfRequestDto) error {
	impl.Logger.Debugw("terminating dangling wf", "name", cancelWfDtoRequest.WorkflowName)
	var err error
	workflowExecutor := impl.getWorkflowExecutor(cancelWfDtoRequest.ExecutorType)
	if workflowExecutor == nil {
		return &utils.ApiError{HttpStatusCode: http.StatusNotFound, Code: strconv.Itoa(http.StatusNotFound), InternalMessage: "workflow executor not found", UserMessage: "workflow executor not found"}
	}
	if cancelWfDtoRequest.RestConfig == nil {
		cancelWfDtoRequest.RestConfig = impl.config
	}
	err = workflowExecutor.TerminateDanglingWorkflow(cancelWfDtoRequest.WorkflowGenerateName, cancelWfDtoRequest.Namespace, cancelWfDtoRequest.RestConfig)
	return err
}

func (impl *WorkflowServiceImpl) getRuntimeEnvClientInstance(environment *repository2.Environment) (v1alpha12.WorkflowInterface, error) {
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

func (impl *WorkflowServiceImpl) getWfClient(environment *repository2.Environment, namespace string, isExt bool) (v1alpha12.WorkflowInterface, error) {
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
