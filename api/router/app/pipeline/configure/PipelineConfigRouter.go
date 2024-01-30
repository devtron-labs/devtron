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

package configure

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline/configure"
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline/webhook"
	"github.com/gorilla/mux"
)

type PipelineConfigRouter interface {
	InitPipelineConfigRouter(configRouter *mux.Router)
}
type PipelineConfigRouterImpl struct {
	restHandler            configure.PipelineConfigRestHandler
	webhookDataRestHandler webhook.WebhookDataRestHandler
}

func NewPipelineRouterImpl(restHandler configure.PipelineConfigRestHandler,
	webhookDataRestHandler webhook.WebhookDataRestHandler) *PipelineConfigRouterImpl {
	return &PipelineConfigRouterImpl{
		restHandler:            restHandler,
		webhookDataRestHandler: webhookDataRestHandler,
	}
}

func (router PipelineConfigRouterImpl) InitPipelineConfigRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(router.restHandler.CreateApp).Methods("POST")
	configRouter.Path("/{appId}").HandlerFunc(router.restHandler.DeleteApp).Methods("DELETE")
	configRouter.Path("/delete/{appId}/{envId}/non-cascade").HandlerFunc(router.restHandler.DeleteACDAppWithNonCascade).Methods("DELETE")
	configRouter.Path("/material").HandlerFunc(router.restHandler.CreateMaterial).Methods("POST")
	configRouter.Path("/material").HandlerFunc(router.restHandler.UpdateMaterial).Methods("PUT")
	configRouter.Path("/material/delete").HandlerFunc(router.restHandler.DeleteMaterial).Methods("DELETE")
	configRouter.Path("/get/{appId}").HandlerFunc(router.restHandler.GetApp).Methods("GET")

	//Deprecated
	configRouter.Path("/template/{appId}/default/{chartRefId}").HandlerFunc(router.restHandler.GetAppOverrideForDefaultTemplate).Methods("GET")

	configRouter.Path("/template").HandlerFunc(router.restHandler.ConfigureDeploymentTemplateForApp).Methods("POST")
	configRouter.Path("/template/{appId}/{chartRefId}").HandlerFunc(router.restHandler.GetDeploymentTemplate).Methods("GET")
	configRouter.Path("/template/default/{appId}/{chartRefId}").HandlerFunc(router.restHandler.GetDefaultDeploymentTemplate).Methods("GET")
	configRouter.Path("/template/update").HandlerFunc(router.restHandler.UpdateAppOverride).Methods("POST")
	configRouter.Path("/template/list").Queries("appId", "{appId}").Queries("envId", "{envId}").HandlerFunc(router.restHandler.GetTemplateComparisonMetadata).Methods("GET")
	configRouter.Path("/template/data").HandlerFunc(router.restHandler.GetDeploymentTemplateData).Methods("POST")

	configRouter.Path("/cd-pipeline").HandlerFunc(router.restHandler.CreateCdPipeline).Methods("POST")
	configRouter.Path("/cd-pipeline/patch").HandlerFunc(router.restHandler.PatchCdPipeline).Methods("POST")
	configRouter.Path("/cd-pipeline/patch/deployment").HandlerFunc(router.restHandler.HandleChangeDeploymentRequest).Methods("POST")
	configRouter.Path("/cd-pipeline/patch/deployment/type").HandlerFunc(router.restHandler.HandleChangeDeploymentTypeRequest).Methods("POST")
	configRouter.Path("/cd-pipeline/patch/deployment/trigger").HandlerFunc(router.restHandler.HandleTriggerDeploymentAfterTypeChange).Methods("POST")
	configRouter.Path("/cd-pipeline/{appId}").HandlerFunc(router.restHandler.GetCdPipelines).Methods("GET")
	configRouter.Path("/cd-pipeline/{appId}/env/{envId}").HandlerFunc(router.restHandler.GetCdPipelinesForAppAndEnv).Methods("GET")
	//save environment specific override
	configRouter.Path("/env/{appId}/{environmentId}").HandlerFunc(router.restHandler.EnvConfigOverrideCreate).Methods("POST")
	configRouter.Path("/env/patch").HandlerFunc(router.restHandler.ChangeChartRef).Methods("PATCH")
	configRouter.Path("/env").HandlerFunc(router.restHandler.EnvConfigOverrideUpdate).Methods("PUT")
	configRouter.Path("/env/{appId}/{environmentId}/{chartRefId}").HandlerFunc(router.restHandler.GetEnvConfigOverride).Methods("GET")

	configRouter.Path("/ci-pipeline").HandlerFunc(router.restHandler.CreateCiConfig).Methods("POST")
	configRouter.Path("/ci-pipeline/{appId}").HandlerFunc(router.restHandler.GetCiPipeline).Methods("GET")
	configRouter.Path("/external-ci/{appId}").HandlerFunc(router.restHandler.GetExternalCi).Methods("GET")
	configRouter.Path("/external-ci/{appId}/{externalCiId}").HandlerFunc(router.restHandler.GetExternalCiById).Methods("GET")
	configRouter.Path("/ci-pipeline/template/patch").HandlerFunc(router.restHandler.UpdateCiTemplate).Methods("POST")
	configRouter.Path("/ci-pipeline/patch").HandlerFunc(router.restHandler.PatchCiPipelines).Methods("POST")
	configRouter.Path("/ci-pipeline/patch-source").HandlerFunc(router.restHandler.PatchCiMaterialSourceWithAppIdAndEnvironmentId).Methods("PATCH")
	configRouter.Path("/ci-pipeline/bulk/branch-update").HandlerFunc(router.restHandler.PatchCiMaterialSourceWithAppIdsAndEnvironmentId).Methods("PUT")
	configRouter.Path("/ci-pipeline/patch/regex").HandlerFunc(router.restHandler.UpdateBranchCiPipelinesWithRegex).Methods("POST")

	configRouter.Path("/cd-pipeline/{cd_pipeline_id}/material").HandlerFunc(router.restHandler.GetArtifactsByCDPipeline).Methods("GET")
	configRouter.Path("/cd-pipeline/{cd_pipeline_id}/material/rollback").HandlerFunc(router.restHandler.GetArtifactsForRollback).Methods("GET")

	configRouter.Path("/team/by-id/{teamId}").HandlerFunc(router.restHandler.FindAppsByTeamId).Methods("GET")
	configRouter.Path("/team/by-name/{teamName}").HandlerFunc(router.restHandler.FindAppsByTeamName).Methods("GET")

	configRouter.Path("/ci-pipeline/trigger").HandlerFunc(router.restHandler.TriggerCiPipeline).Methods("POST")

	configRouter.Path("/{appId}/ci-pipeline/min").HandlerFunc(router.restHandler.GetCiPipelineMin).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/material").HandlerFunc(router.restHandler.FetchMaterials).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/material/{gitMaterialId}").HandlerFunc(router.restHandler.FetchMaterialsByMaterialId).Methods("GET")
	configRouter.Path("/ci-pipeline/refresh-material/{gitMaterialId}").HandlerFunc(router.restHandler.RefreshMaterials).Methods("GET")

	configRouter.Path("/{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}").HandlerFunc(router.restHandler.FetchWorkflowDetails).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/workflow/{workflowId}/ci-job/artifacts").HandlerFunc(router.restHandler.GetArtifactsForCiJob).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/artifacts/{workflowId}").HandlerFunc(router.restHandler.DownloadCiWorkflowArtifacts).Methods("GET")

	configRouter.Path("/ci-pipeline/{pipelineId}/git-changes/{ciMaterialId}").HandlerFunc(router.restHandler.FetchChanges).Methods("GET")

	configRouter.Path("/ci-pipeline/{pipelineId}/workflow/{workflowId}/logs/old").HandlerFunc(router.restHandler.GetHistoricBuildLogs).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/workflow/{workflowId}/logs").HandlerFunc(router.restHandler.GetBuildLogs).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/workflows").HandlerFunc(router.restHandler.GetBuildHistory).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/workflow/{workflowId}").HandlerFunc(router.restHandler.CancelWorkflow).Methods("DELETE")
	configRouter.Path("/cd-pipeline/{pipelineId}/workflowRunner/{workflowRunnerId}").HandlerFunc(router.restHandler.CancelStage).Methods("DELETE")

	configRouter.Path("/cd-pipeline/defaultStrategy/{appId}/{envId}").HandlerFunc(router.restHandler.GetDefaultDeploymentPipelineStrategy).Methods("GET")
	configRouter.Path("/cd-pipeline/{appId}/{envId}/{pipelineId}").HandlerFunc(router.restHandler.IsReadyToTrigger).Methods("GET")
	configRouter.Path("/cd-pipeline/strategies/{appId}").HandlerFunc(router.restHandler.GetDeploymentPipelineStrategy).Methods("GET")

	configRouter.Path("/upgrade/all/{chartRefId}").HandlerFunc(router.restHandler.UpgradeForAllApps).Methods("POST")

	configRouter.Path("/env/reset/{appId}/{environmentId}/{id}").HandlerFunc(router.restHandler.EnvConfigOverrideReset).Methods("DELETE")
	configRouter.Path("/env/namespace/{appId}/{environmentId}").HandlerFunc(router.restHandler.EnvConfigOverrideCreateNamespace).Methods("POST")

	configRouter.Path("/cd-pipeline/workflow/history/{appId}/{environmentId}/{pipelineId}").HandlerFunc(router.restHandler.ListDeploymentHistory).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/logs/{appId}/{environmentId}/{pipelineId}/{workflowId}").HandlerFunc(router.restHandler.GetPrePostDeploymentLogs).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/trigger-info/{appId}/{environmentId}/{pipelineId}/{workflowRunnerId}").HandlerFunc(router.restHandler.FetchCdWorkflowDetails).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/download/{appId}/{environmentId}/{pipelineId}/{workflowRunnerId}").HandlerFunc(router.restHandler.DownloadArtifacts).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/status/{appId}/{environmentId}/{pipelineId}").HandlerFunc(router.restHandler.GetStageStatus).Methods("GET")

	configRouter.Path("/cd-pipeline/{appId}/{pipelineId}").HandlerFunc(router.restHandler.GetCdPipelineById).Methods("GET")
	configRouter.Path("/v2/cd-pipeline/{appId}/{pipelineId}").HandlerFunc(router.restHandler.GetCdPipelineById).Methods("GET")

	configRouter.Path("/cd/configmap-secrets/{pipelineId}").HandlerFunc(router.restHandler.GetConfigmapSecretsForDeploymentStages).Methods("GET")

	configRouter.Path("/workflow/status/{appId}").HandlerFunc(router.restHandler.FetchAppWorkflowStatusForTriggerView).Methods("GET")
	configRouter.Path("/workflow/status/{appId}/{version}").HandlerFunc(router.restHandler.FetchAppWorkflowStatusForTriggerView).Methods("GET")

	configRouter.Path("/material-info/{envId}/{ciArtifactId}").HandlerFunc(router.restHandler.FetchMaterialInfo).Methods("GET")
	configRouter.Path("/ci-pipeline/webhook-payload/{pipelineMaterialId}").HandlerFunc(router.webhookDataRestHandler.GetWebhookPayloadDataForPipelineMaterialId).Methods("GET")
	configRouter.Path("/ci-pipeline/webhook-payload/{pipelineMaterialId}/{parsedDataId}").HandlerFunc(router.webhookDataRestHandler.GetWebhookPayloadFilterDataForPipelineMaterialId).Methods("GET")
	configRouter.Path("/ci-pipeline/{appId}/{pipelineId}").HandlerFunc(router.restHandler.GetCIPipelineById).Methods("GET")

	configRouter.Path("/pipeline/suggest/{type}/{appId}").
		HandlerFunc(router.restHandler.PipelineNameSuggestion).Methods("GET")

	configRouter.Path("/commit-info/{ciPipelineMaterialId}/{gitHash}").HandlerFunc(router.restHandler.GetCommitMetadataForPipelineMaterial).Methods("GET")

	configRouter.Path("/image-tagging/{ciPipelineId}/{artifactId}").HandlerFunc(router.restHandler.CreateUpdateImageTagging).Methods("POST")
	configRouter.Path("/image-tagging/{ciPipelineId}/{artifactId}").HandlerFunc(router.restHandler.GetImageTaggingData).Methods("GET")
}
