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

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/restHandler/app"
	"github.com/gorilla/mux"
)

type PipelineConfigRouter interface {
	initPipelineConfigRouter(configRouter *mux.Router)
}
type PipelineConfigRouterImpl struct {
	restHandler                app.PipelineConfigRestHandler
	appWorkflowRestHandler     restHandler.AppWorkflowRestHandler
	webhookDataRestHandler     restHandler.WebhookDataRestHandler
	pipelineHistoryRestHandler restHandler.PipelineHistoryRestHandler
}

func NewPipelineRouterImpl(restHandler app.PipelineConfigRestHandler,
	appWorkflowRestHandler restHandler.AppWorkflowRestHandler,
	webhookDataRestHandler restHandler.WebhookDataRestHandler,
	pipelineHistoryRestHandler restHandler.PipelineHistoryRestHandler) *PipelineConfigRouterImpl {
	return &PipelineConfigRouterImpl{
		restHandler:                restHandler,
		appWorkflowRestHandler:     appWorkflowRestHandler,
		webhookDataRestHandler:     webhookDataRestHandler,
		pipelineHistoryRestHandler: pipelineHistoryRestHandler,
	}

}

func (router PipelineConfigRouterImpl) initPipelineConfigRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(router.restHandler.CreateApp).Methods("POST")
	configRouter.Path("/{appId}").HandlerFunc(router.restHandler.DeleteApp).Methods("DELETE")
	configRouter.Path("/material").HandlerFunc(router.restHandler.CreateMaterial).Methods("POST")
	configRouter.Path("/material").HandlerFunc(router.restHandler.UpdateMaterial).Methods("PUT")
	configRouter.Path("/material/delete").HandlerFunc(router.restHandler.DeleteMaterial).Methods("DELETE")
	configRouter.Path("/get/{appId}").HandlerFunc(router.restHandler.GetApp).Methods("GET")
	configRouter.Path("/autocomplete").HandlerFunc(router.restHandler.GetAppListForAutocomplete).Methods("GET")
	configRouter.Path("/min").HandlerFunc(router.restHandler.GetAppListByTeamIds).Methods("GET")

	//Deprecated
	configRouter.Path("/template/{appId}/default/{chartRefId}").HandlerFunc(router.restHandler.GetAppOverrideForDefaultTemplate).Methods("GET")

	configRouter.Path("/template").HandlerFunc(router.restHandler.ConfigureDeploymentTemplateForApp).Methods("POST")
	configRouter.Path("/template/{appId}/{chartRefId}").HandlerFunc(router.restHandler.GetDeploymentTemplate).Methods("GET")
	configRouter.Path("/template/update").HandlerFunc(router.restHandler.UpdateAppOverride).Methods("POST")

	configRouter.Path("/cd-pipeline").HandlerFunc(router.restHandler.CreateCdPipeline).Methods("POST")
	configRouter.Path("/cd-pipeline/patch").HandlerFunc(router.restHandler.PatchCdPipeline).Methods("POST")
	configRouter.Path("/cd-pipeline/{appId}").HandlerFunc(router.restHandler.GetCdPipelines).Methods("GET")
	configRouter.Path("/cd-pipeline/{appId}/env/{envId}").HandlerFunc(router.restHandler.GetCdPipelinesForAppAndEnv).Methods("GET")

	//save environment specific override
	configRouter.Path("/env/{appId}/{environmentId}").HandlerFunc(router.restHandler.EnvConfigOverrideCreate).Methods("POST")
	configRouter.Path("/env").HandlerFunc(router.restHandler.EnvConfigOverrideUpdate).Methods("PUT")
	configRouter.Path("/env/{appId}/{environmentId}/{chartRefId}").HandlerFunc(router.restHandler.GetEnvConfigOverride).Methods("GET")

	configRouter.Path("/ci-pipeline").HandlerFunc(router.restHandler.CreateCiConfig).Methods("POST")
	configRouter.Path("/ci-pipeline/{appId}").HandlerFunc(router.restHandler.GetCiPipeline).Methods("GET")
	configRouter.Path("/ci-pipeline/template/patch").HandlerFunc(router.restHandler.UpdateCiTemplate).Methods("POST")
	configRouter.Path("/ci-pipeline/patch").HandlerFunc(router.restHandler.PatchCiPipelines).Methods("POST")

	configRouter.Path("/cd-pipeline/{cd_pipeline_id}/material").HandlerFunc(router.restHandler.GetArtifactsByCDPipeline).Methods("GET")
	configRouter.Path("/cd-pipeline/{cd_pipeline_id}/material/rollback").HandlerFunc(router.restHandler.GetArtifactForRollback).Methods("GET")

	configRouter.Path("/migrate/db").HandlerFunc(router.restHandler.CreateMigrationConfig).Methods("POST")
	configRouter.Path("/migrate/db/update").HandlerFunc(router.restHandler.UpdateMigrationConfig).Methods("POST")
	configRouter.Path("/migrate/db/{pipelineId}").HandlerFunc(router.restHandler.GetMigrationConfig).Methods("GET")

	configRouter.Path("/team/by-id/{teamId}").HandlerFunc(router.restHandler.FindAppsByTeamId).Methods("GET")
	configRouter.Path("/team/by-name/{teamName}").HandlerFunc(router.restHandler.FindAppsByTeamName).Methods("GET")

	configRouter.Path("/ci-pipeline/trigger").HandlerFunc(router.restHandler.TriggerCiPipeline).Methods("POST")
	configRouter.Path("/{appId}/ci-pipeline/min").HandlerFunc(router.restHandler.GetCiPipelineMin).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/material").HandlerFunc(router.restHandler.FetchMaterials).Methods("GET")
	configRouter.Path("/ci-pipeline/refresh-material/{gitMaterialId}").HandlerFunc(router.restHandler.RefreshMaterials).Methods("GET")

	configRouter.Path("/{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}").HandlerFunc(router.restHandler.FetchWorkflowDetails).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/artifacts/{workflowId}").HandlerFunc(router.restHandler.DownloadCiWorkflowArtifacts).Methods("GET")

	configRouter.Path("/ci-pipeline/{pipelineId}/git-changes/{ciMaterialId}").HandlerFunc(router.restHandler.FetchChanges).Methods("GET")

	configRouter.Path("/ci-pipeline/{pipelineId}/workflow/{workflowId}/logs/old").HandlerFunc(router.restHandler.GetHistoricBuildLogs).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/workflow/{workflowId}/logs").HandlerFunc(router.restHandler.GetBuildLogs).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/workflows").HandlerFunc(router.restHandler.GetBuildHistory).Methods("GET")
	configRouter.Path("/ci-pipeline/{pipelineId}/workflow/{workflowId}").HandlerFunc(router.restHandler.CancelWorkflow).Methods("DELETE")
	configRouter.Path("/cd-pipeline/{pipelineId}/workflowRunner/{workflowRunnerId}").HandlerFunc(router.restHandler.CancelStage).Methods("DELETE")

	configRouter.Path("/{appId}/autocomplete/environment").HandlerFunc(router.restHandler.EnvironmentListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/git").HandlerFunc(router.restHandler.GitListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/docker").HandlerFunc(router.restHandler.DockerListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/team").HandlerFunc(router.restHandler.TeamListAutocomplete).Methods("GET")

	configRouter.Path("/cd-pipeline/{appId}/{envId}/{pipelineId}").HandlerFunc(router.restHandler.IsReadyToTrigger).Methods("GET")
	configRouter.Path("/cd-pipeline/strategies/{appId}").HandlerFunc(router.restHandler.GetDeploymentPipelineStrategy).Methods("GET")

	configRouter.Path("/upgrade/all/{chartRefId}").HandlerFunc(router.restHandler.UpgradeForAllApps).Methods("POST")

	configRouter.Path("/env/reset/{appId}/{environmentId}/{id}").HandlerFunc(router.restHandler.EnvConfigOverrideReset).Methods("DELETE")
	configRouter.Path("/env/namespace/{appId}/{environmentId}").HandlerFunc(router.restHandler.EnvConfigOverrideCreateNamespace).Methods("POST")

	configRouter.Path("/template/metrics/{appId}").HandlerFunc(router.restHandler.AppMetricsEnableDisable).Methods("POST")
	configRouter.Path("/env/metrics/{appId}/{environmentId}").HandlerFunc(router.restHandler.EnvMetricsEnableDisable).Methods("POST")

	configRouter.Path("/app-wf").
		HandlerFunc(router.appWorkflowRestHandler.CreateAppWorkflow).Methods("POST")

	configRouter.Path("/app-wf/{app-id}").
		HandlerFunc(router.appWorkflowRestHandler.FindAppWorkflow).Methods("GET")

	configRouter.Path("/app-wf/{app-id}/{app-wf-id}").
		HandlerFunc(router.appWorkflowRestHandler.DeleteAppWorkflow).Methods("DELETE")

	configRouter.Path("/cd-pipeline/workflow/history/{appId}/{environmentId}/{pipelineId}").HandlerFunc(router.restHandler.ListDeploymentHistory).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/logs/{appId}/{environmentId}/{pipelineId}/{workflowId}").HandlerFunc(router.restHandler.GetPrePostDeploymentLogs).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/trigger-info/{appId}/{environmentId}/{pipelineId}/{workflowRunnerId}").HandlerFunc(router.restHandler.FetchCdWorkflowDetails).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/download/{appId}/{environmentId}/{pipelineId}/{workflowRunnerId}").HandlerFunc(router.restHandler.DownloadArtifacts).Methods("GET")
	configRouter.Path("/cd-pipeline/workflow/status/{appId}/{environmentId}/{pipelineId}").HandlerFunc(router.restHandler.GetStageStatus).Methods("GET")

	configRouter.Path("/cd-pipeline/{appId}/{pipelineId}").HandlerFunc(router.restHandler.GetCdPipelineById).Methods("GET")
	configRouter.Path("/cd/configmap-secrets/{pipelineId}").HandlerFunc(router.restHandler.GetConfigmapSecretsForDeploymentStages).Methods("GET")

	configRouter.Path("/workflow/status/{appId}").HandlerFunc(router.restHandler.FetchAppWorkflowStatusForTriggerView).Methods("GET")

	configRouter.Path("/material-info/{appId}/{ciArtifactId}").HandlerFunc(router.restHandler.FetchMaterialInfo).Methods("GET")
	configRouter.Path("/ci-pipeline/webhook-payload/{pipelineMaterialId}").HandlerFunc(router.webhookDataRestHandler.GetWebhookPayloadDataForPipelineMaterialId).Methods("GET")
	configRouter.Path("/ci-pipeline/webhook-payload/{pipelineMaterialId}/{parsedDataId}").HandlerFunc(router.webhookDataRestHandler.GetWebhookPayloadFilterDataForPipelineMaterialId).Methods("GET")
	configRouter.Path("/ci-pipeline/{appId}/{pipelineId}").HandlerFunc(router.restHandler.GetCIPipelineById).Methods("GET")

	configRouter.Path("/pipeline/suggest/{type}/{appId}").
		HandlerFunc(router.restHandler.PipelineNameSuggestion).Methods("GET")

	configRouter.Path("/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}").
		HandlerFunc(router.pipelineHistoryRestHandler.FetchDeployedConfigurationsForWorkflow).
		Methods("GET")

	configRouter.Path("/history/deployed-component/list/{appId}/{pipelineId}").
		HandlerFunc(router.pipelineHistoryRestHandler.FetchDeployedHistoryComponentList).
		Methods("GET")

	configRouter.Path("/history/deployed-component/detail/{appId}/{pipelineId}/{id}").
		HandlerFunc(router.pipelineHistoryRestHandler.FetchDeployedHistoryComponentDetail).
		Methods("GET")

	configRouter.Path("/commit-info/{ciPipelineMaterialId}/{gitHash}").HandlerFunc(router.restHandler.GetCommitMetadataForPipelineMaterial).Methods("GET")
}
