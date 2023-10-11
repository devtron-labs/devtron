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
	errors3 "errors"
	"fmt"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	models2 "github.com/devtron-labs/devtron/api/helm-app/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/go-pg/pg"
	errors2 "github.com/juju/errors"
	"go.opentelemetry.io/otel"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"strings"
	"time"
)

type CdPipelineConfigService interface {
	//GetCdPipelineById : Retrieve cdPipeline for given cdPipelineId.
	//getting cdPipeline,environment and strategies ,preDeployStage, postDeployStage,appWorkflowMapping from respective repository and service layer
	//converting above data in proper bean object and then assigning to CDPipelineConfigObject
	//if any error occur , will get empty object or nil
	GetCdPipelineById(pipelineId int) (cdPipeline *bean.CDPipelineConfigObject, err error)
	CreateCdPipelines(cdPipelines *bean.CdPipelines, ctx context.Context) (*bean.CdPipelines, error)
	//PatchCdPipelines : Handle CD pipeline patch requests, making necessary changes to the configuration and returning the updated version.
	//Performs Create ,Update and Delete operation.
	PatchCdPipelines(cdPipelines *bean.CDPatchRequest, ctx context.Context) (*bean.CdPipelines, error)
	DeleteCdPipeline(pipeline *pipelineConfig.Pipeline, ctx context.Context, deleteAction int, acdDelete bool, userId int32) (*bean.AppDeleteResponseDTO, error)
	DeleteACDAppCdPipelineWithNonCascade(pipeline *pipelineConfig.Pipeline, ctx context.Context, forceDelete bool, userId int32) (err error)
	//GetTriggerViewCdPipelinesForApp :
	GetTriggerViewCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error)
	//GetCdPipelinesForApp : Retrieve cdPipeline for given appId
	GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error)
	//GetCdPipelinesForAppAndEnv : Retrieve cdPipeline for given appId and envId
	GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error)
	/*	CreateCdPipelines(cdPipelines bean.CdPipelines) (*bean.CdPipelines, error)*/
	//GetCdPipelinesByEnvironment : lists cdPipeline for given environmentId and appIds
	GetCdPipelinesByEnvironment(request resourceGroup2.ResourceGroupingRequest) (cdPipelines *bean.CdPipelines, err error)
	//GetCdPipelinesByEnvironmentMin : lists minimum detail of cdPipelines for given environmentId and appIds
	GetCdPipelinesByEnvironmentMin(request resourceGroup2.ResourceGroupingRequest) (cdPipelines []*bean.CDPipelineConfigObject, err error)
	//PerformBulkActionOnCdPipelines :
	PerformBulkActionOnCdPipelines(dto *bean.CdBulkActionRequestDto, impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun bool, userId int32) ([]*bean.CdBulkActionResponseDto, error)
	//FindPipelineById : Retrieve Pipeline object from pipelineRepository for given cdPipelineId
	FindPipelineById(cdPipelineId int) (*pipelineConfig.Pipeline, error)
	//FindAppAndEnvDetailsByPipelineId : Retrieve app and env details for given cdPipelineId
	FindAppAndEnvDetailsByPipelineId(cdPipelineId int) (*pipelineConfig.Pipeline, error)
	// RetrieveParentDetails : Retrieve the parent id and type of the parent.
	//Here ParentId refers to Parent like parent of CD can be CI , PRE-CD .
	// It first fetches the workflow details from the appWorkflow repository.
	//If the workflow is a CD pipeline, it further checks for stage configurations.
	//If the workflow is a webhook, it returns the webhook workflow type.
	//In case of error , it returns 0 for parentId and empty string for parentType
	RetrieveParentDetails(pipelineId int) (parentId int, parentType bean2.WorkflowType, err error)
	//GetEnvironmentByCdPipelineId : Retrieve environmentId for given cdPipelineId
	GetEnvironmentByCdPipelineId(pipelineId int) (int, error)
	GetBulkActionImpactedPipelines(dto *bean.CdBulkActionRequestDto) ([]*pipelineConfig.Pipeline, error) //no usage
	//IsGitOpsRequiredForCD : Determine if GitOps is required for CD based on the provided pipeline creation request
	IsGitOpsRequiredForCD(pipelineCreateRequest *bean.CdPipelines) bool
	//SetPipelineDeploymentAppType : Set pipeline deployment application(helm/argo) types based on the provided configuration
	SetPipelineDeploymentAppType(pipelineCreateRequest *bean.CdPipelines, isGitOpsConfigured bool, deploymentTypeValidationConfig map[string]bool)
	MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(appId int, envId int, acdToken string, pipeline *pipelineConfig.Pipeline) (bool, error)
	//GetEnvironmentListForAutocompleteFilter : lists environment for given configuration
	GetEnvironmentListForAutocompleteFilter(envName string, clusterIds []int, offset int, size int, emailId string, checkAuthBatch func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool), ctx context.Context) (*cluster.ResourceGroupingResponse, error)
	IsGitopsConfigured() (bool, error)
}
type DevtronAppCMCSService interface {
	//FetchConfigmapSecretsForCdStages : Delegating the request to appService for fetching cm/cs
	FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error)
	//GetDeploymentConfigMap : Retrieve deployment config values from the attributes table
	GetDeploymentConfigMap(environmentId int) (map[string]bool, error)
}
type DevtronAppStrategyService interface {
	//FetchCDPipelineStrategy : Retrieve CDPipelineStrategy for given appId
	FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error)
	//FetchDefaultCDPipelineStrategy :
	FetchDefaultCDPipelineStrategy(appId int, envId int) (PipelineStrategy, error)
}
type AppDeploymentTypeChangeManager interface {
	//ChangeDeploymentType : takes in DeploymentAppTypeChangeRequest struct and
	// deletes all the cd pipelines for that deployment type in all apps that belongs to
	// that environment and updates the db with desired deployment app type
	ChangeDeploymentType(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	//ChangePipelineDeploymentType : takes in DeploymentAppTypeChangeRequest struct and
	// deletes all the cd pipelines for that deployment type in all apps that belongs to
	// that environment and updates the db with desired deployment app type
	ChangePipelineDeploymentType(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	//TriggerDeploymentAfterTypeChange :
	TriggerDeploymentAfterTypeChange(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	//DeleteDeploymentApps : takes in a list of pipelines and delete the applications
	DeleteDeploymentApps(ctx context.Context, pipelines []*pipelineConfig.Pipeline, userId int32) *bean.DeploymentAppTypeChangeResponse
	//DeleteDeploymentAppsForEnvironment : takes in environment id and current deployment app type
	// and deletes all the cd pipelines for that deployment type in all apps that belongs to
	// that environment.
	DeleteDeploymentAppsForEnvironment(ctx context.Context, environmentId int, currentDeploymentAppType bean.DeploymentType, exclusionList []int, includeApps []int, userId int32) (*bean.DeploymentAppTypeChangeResponse, error)
}

func (impl *PipelineBuilderImpl) GetCdPipelineById(pipelineId int) (cdPipeline *bean.CDPipelineConfigObject, err error) {
	dbPipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return cdPipeline, err
	}
	environment, err := impl.environmentRepository.FindById(dbPipeline.EnvironmentId)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return cdPipeline, err
	}
	strategies, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineId(dbPipeline.Id)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching strategies", "err", err)
		return cdPipeline, err
	}
	var strategiesBean []bean.Strategy
	var deploymentTemplate chartRepoRepository.DeploymentStrategy
	for _, item := range strategies {
		strategiesBean = append(strategiesBean, bean.Strategy{
			Config:             []byte(item.Config),
			DeploymentTemplate: item.Strategy,
			Default:            item.Default,
		})

		if item.Default {
			deploymentTemplate = item.Strategy
		}
	}

	preStage := bean.CdStage{}
	if len(dbPipeline.PreStageConfig) > 0 {
		preStage.Name = "Pre-Deployment"
		preStage.Config = dbPipeline.PreStageConfig
		preStage.TriggerType = dbPipeline.PreTriggerType
	}
	postStage := bean.CdStage{}
	if len(dbPipeline.PostStageConfig) > 0 {
		postStage.Name = "Post-Deployment"
		postStage.Config = dbPipeline.PostStageConfig
		postStage.TriggerType = dbPipeline.PostTriggerType
	}

	preStageConfigmapSecrets := bean.PreStageConfigMapSecretNames{}
	postStageConfigmapSecrets := bean.PostStageConfigMapSecretNames{}

	if dbPipeline.PreStageConfigMapSecretNames != "" {
		err = json.Unmarshal([]byte(dbPipeline.PreStageConfigMapSecretNames), &preStageConfigmapSecrets)
		if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
	}
	if dbPipeline.PostStageConfigMapSecretNames != "" {
		err = json.Unmarshal([]byte(dbPipeline.PostStageConfigMapSecretNames), &postStageConfigmapSecrets)
		if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
	}
	appWorkflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCDPipelineId(pipelineId)
	if err != nil {
		return nil, err
	}
	cdPipeline = &bean.CDPipelineConfigObject{
		Id:                            dbPipeline.Id,
		Name:                          dbPipeline.Name,
		EnvironmentId:                 dbPipeline.EnvironmentId,
		EnvironmentName:               environment.Name,
		CiPipelineId:                  dbPipeline.CiPipelineId,
		DeploymentTemplate:            deploymentTemplate,
		TriggerType:                   dbPipeline.TriggerType,
		Strategies:                    strategiesBean,
		PreStage:                      preStage,
		PostStage:                     postStage,
		PreStageConfigMapSecretNames:  preStageConfigmapSecrets,
		PostStageConfigMapSecretNames: postStageConfigmapSecrets,
		RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
		RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
		CdArgoSetup:                   environment.Cluster.CdArgoSetup,
		ParentPipelineId:              appWorkflowMapping.ParentId,
		ParentPipelineType:            appWorkflowMapping.ParentType,
		DeploymentAppType:             dbPipeline.DeploymentAppType,
		DeploymentAppCreated:          dbPipeline.DeploymentAppCreated,
		IsVirtualEnvironment:          dbPipeline.Environment.IsVirtualEnvironment,
	}
	var preDeployStage *bean3.PipelineStageDto
	var postDeployStage *bean3.PipelineStageDto
	preDeployStage, postDeployStage, err = impl.pipelineStageService.GetCdPipelineStageDataDeepCopy(dbPipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in getting pre/post-CD stage data", "err", err, "cdPipelineId", dbPipeline.Id)
		return nil, err
	}
	cdPipeline.PreDeployStage = preDeployStage
	cdPipeline.PostDeployStage = postDeployStage

	return cdPipeline, err
}

func (impl *PipelineBuilderImpl) CreateCdPipelines(pipelineCreateRequest *bean.CdPipelines, ctx context.Context) (*bean.CdPipelines, error) {

	//Validation for checking deployment App type
	isGitOpsConfigured, err := impl.IsGitopsConfigured()

	for _, pipeline := range pipelineCreateRequest.Pipelines {
		// if no deployment app type sent from user then we'll not validate
		deploymentConfig, err := impl.GetDeploymentConfigMap(pipeline.EnvironmentId)
		if err != nil {
			return nil, err
		}
		impl.SetPipelineDeploymentAppType(pipelineCreateRequest, isGitOpsConfigured, deploymentConfig)
		if err := impl.validateDeploymentAppType(pipeline, deploymentConfig); err != nil {
			impl.logger.Errorw("validation error in creating pipeline", "name", pipeline.Name, "err", err)
			return nil, err
		}
	}

	isGitOpsRequiredForCD := impl.IsGitOpsRequiredForCD(pipelineCreateRequest)
	app, err := impl.appRepo.FindById(pipelineCreateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("app not found", "err", err, "appId", pipelineCreateRequest.AppId)
		return nil, err
	}
	_, err = impl.ValidateCDPipelineRequest(pipelineCreateRequest, isGitOpsConfigured, isGitOpsRequiredForCD)
	if err != nil {
		return nil, err
	}

	// TODO: creating git repo for all apps irrespective of acd or helm
	if isGitOpsConfigured && isGitOpsRequiredForCD {

		gitopsRepoName, chartGitAttr, err := impl.appService.CreateGitopsRepo(app, pipelineCreateRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in creating git repo", "err", err)
			return nil, err
		}

		err = impl.RegisterInACD(gitopsRepoName, chartGitAttr, pipelineCreateRequest.UserId, ctx)
		if err != nil {
			impl.logger.Errorw("error in registering app in acd", "err", err)
			return nil, err
		}

		err = impl.updateGitRepoUrlInCharts(pipelineCreateRequest.AppId, chartGitAttr, pipelineCreateRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in updating git repo url in charts", "err", err)
			return nil, err
		}
	}

	for _, pipeline := range pipelineCreateRequest.Pipelines {

		id, err := impl.createCdPipeline(ctx, app, pipeline, pipelineCreateRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in creating pipeline", "name", pipeline.Name, "err", err)
			return nil, err
		}
		pipeline.Id = id

		//creating pipeline_stage entry here after tx commit due to FK issue
		if pipeline.PreDeployStage != nil && len(pipeline.PreDeployStage.Steps) > 0 {
			err = impl.pipelineStageService.CreatePipelineStage(pipeline.PreDeployStage, repository5.PIPELINE_STAGE_TYPE_PRE_CD, id, pipelineCreateRequest.UserId)
			if err != nil {
				impl.logger.Errorw("error in creating pre-cd stage", "err", err, "preCdStage", pipeline.PreDeployStage, "pipelineId", id)
				return nil, err
			}
		}
		if pipeline.PostDeployStage != nil && len(pipeline.PostDeployStage.Steps) > 0 {
			err = impl.pipelineStageService.CreatePipelineStage(pipeline.PostDeployStage, repository5.PIPELINE_STAGE_TYPE_POST_CD, id, pipelineCreateRequest.UserId)
			if err != nil {
				impl.logger.Errorw("error in creating post-cd stage", "err", err, "postCdStage", pipeline.PostDeployStage, "pipelineId", id)
				return nil, err
			}
		}

	}

	return pipelineCreateRequest, nil
}

func (impl *PipelineBuilderImpl) PatchCdPipelines(cdPipelines *bean.CDPatchRequest, ctx context.Context) (*bean.CdPipelines, error) {
	pipelineRequest := &bean.CdPipelines{
		UserId:    cdPipelines.UserId,
		AppId:     cdPipelines.AppId,
		Pipelines: []*bean.CDPipelineConfigObject{cdPipelines.Pipeline},
	}
	deleteAction := bean.CASCADE_DELETE
	if cdPipelines.ForceDelete {
		deleteAction = bean.FORCE_DELETE
	} else if cdPipelines.NonCascadeDelete {
		deleteAction = bean.NON_CASCADE_DELETE
	}
	switch cdPipelines.Action {
	case bean.CD_CREATE:
		return impl.CreateCdPipelines(pipelineRequest, ctx)
	case bean.CD_UPDATE:
		err := impl.updateCdPipeline(ctx, cdPipelines.Pipeline, cdPipelines.UserId)
		return pipelineRequest, err
	case bean.CD_DELETE:
		pipeline, err := impl.pipelineRepository.FindById(cdPipelines.Pipeline.Id)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "id", cdPipelines.Pipeline.Id)
			return pipelineRequest, err
		}
		deleteResponse, err := impl.DeleteCdPipeline(pipeline, ctx, deleteAction, false, cdPipelines.UserId)
		pipelineRequest.AppDeleteResponse = deleteResponse
		return pipelineRequest, err
	case bean.CD_DELETE_PARTIAL:
		pipeline, err := impl.pipelineRepository.FindById(cdPipelines.Pipeline.Id)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "id", cdPipelines.Pipeline.Id)
			return pipelineRequest, err
		}
		deleteResponse, err := impl.DeleteCdPipelinePartial(pipeline, ctx, deleteAction, cdPipelines.UserId)
		pipelineRequest.AppDeleteResponse = deleteResponse
		return pipelineRequest, err
	default:
		return nil, &util.ApiError{Code: "404", HttpStatusCode: 404, UserMessage: "operation not supported"}
	}
}

func (impl *PipelineBuilderImpl) DeleteCdPipeline(pipeline *pipelineConfig.Pipeline, ctx context.Context, deleteAction int, deleteFromAcd bool, userId int32) (*bean.AppDeleteResponseDTO, error) {
	cascadeDelete := true
	forceDelete := false
	deleteResponse := &bean.AppDeleteResponseDTO{
		DeleteInitiated:  false,
		ClusterReachable: true,
	}
	if deleteAction == bean.FORCE_DELETE {
		forceDelete = true
		cascadeDelete = false
	} else if deleteAction == bean.NON_CASCADE_DELETE {
		cascadeDelete = false
	}
	// updating cluster reachable flag
	clusterBean, err := impl.clusterRepository.FindById(pipeline.Environment.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster details", "err", err, "clusterId", pipeline.Environment.ClusterId)
	}
	deleteResponse.ClusterName = clusterBean.ClusterName
	if len(clusterBean.ErrorInConnecting) > 0 {
		deleteResponse.ClusterReachable = false
	}
	//getting children CD pipeline details
	childNodes, err := impl.appWorkflowRepository.FindWFCDMappingByParentCDPipelineId(pipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting children cd details", "err", err)
		return deleteResponse, err
	} else if len(childNodes) > 0 {
		impl.logger.Debugw("cannot delete cd pipeline, contains children cd")
		return deleteResponse, fmt.Errorf("Please delete children CD pipelines before deleting this pipeline.")
	}
	//getting deployment group for this pipeline
	deploymentGroupNames, err := impl.deploymentGroupRepository.GetNamesByAppIdAndEnvId(pipeline.EnvironmentId, pipeline.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting deployment group names by appId and envId", "err", err)
		return deleteResponse, err
	} else if len(deploymentGroupNames) > 0 {
		groupNamesByte, err := json.Marshal(deploymentGroupNames)
		if err != nil {
			impl.logger.Errorw("error in marshaling deployment group names", "err", err, "deploymentGroupNames", deploymentGroupNames)
		}
		impl.logger.Debugw("cannot delete cd pipeline, is being used in deployment group")
		return deleteResponse, fmt.Errorf("Please remove this CD pipeline from deployment groups : %s", string(groupNamesByte))
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return deleteResponse, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	if err = impl.ciCdPipelineOrchestrator.DeleteCdPipeline(pipeline.Id, userId, tx); err != nil {
		impl.logger.Errorw("err in deleting pipeline from db", "id", pipeline, "err", err)
		return deleteResponse, err
	}
	// delete entry in app_status table
	err = impl.appStatusRepository.Delete(tx, pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting app_status from db", "appId", pipeline.AppId, "envId", pipeline.EnvironmentId, "err", err)
		return deleteResponse, err
	}
	//delete app workflow mapping
	appWorkflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCDPipelineId(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "err", err)
		return deleteResponse, err
	}
	if appWorkflowMapping.ParentType == appWorkflow.WEBHOOK {
		childNodes, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiId(appWorkflowMapping.ParentId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching external ci", "err", err)
			return deleteResponse, err
		}
		noOtherChildNodes := true
		for _, childNode := range childNodes {
			if appWorkflowMapping.Id != childNode.Id {
				noOtherChildNodes = false
			}
		}
		if noOtherChildNodes {
			externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(appWorkflowMapping.ParentId)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return deleteResponse, err
			}
			externalCiPipeline.Active = false
			externalCiPipeline.UpdatedOn = time.Now()
			externalCiPipeline.UpdatedBy = userId
			_, err = impl.ciPipelineRepository.UpdateExternalCi(externalCiPipeline, tx)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return deleteResponse, err
			}

			appWorkflow, err := impl.appWorkflowRepository.FindById(appWorkflowMapping.AppWorkflowId)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return deleteResponse, err
			}
			err = impl.appWorkflowRepository.DeleteAppWorkflow(appWorkflow, tx)
			if err != nil {
				impl.logger.Errorw("error in deleting workflow mapping", "err", err)
				return deleteResponse, err
			}
		}
	}
	appWorkflowMapping.UpdatedBy = userId
	appWorkflowMapping.UpdatedOn = time.Now()
	err = impl.appWorkflowRepository.DeleteAppWorkflowMapping(appWorkflowMapping, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "err", err)
		return deleteResponse, err
	}

	if pipeline.PreStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.PRE_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating pre cd script entry", "err", err, "pipeline", pipeline)
			return deleteResponse, err
		}
	}
	if pipeline.PostStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.POST_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
			return deleteResponse, err
		}
	}
	cdPipelinePluginDeleteReq, err := impl.GetCdPipelineById(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in getting cdPipeline by id", "err", err, "id", pipeline.Id)
		return deleteResponse, err
	}
	if cdPipelinePluginDeleteReq.PreDeployStage != nil && cdPipelinePluginDeleteReq.PreDeployStage.Id > 0 {
		//deleting pre-stage
		err = impl.pipelineStageService.DeletePipelineStage(cdPipelinePluginDeleteReq.PreDeployStage, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting pre-CD stage", "err", err, "preDeployStage", cdPipelinePluginDeleteReq.PreDeployStage)
			return deleteResponse, err
		}
	}
	if cdPipelinePluginDeleteReq.PostDeployStage != nil && cdPipelinePluginDeleteReq.PostDeployStage.Id > 0 {
		//deleting post-stage
		err = impl.pipelineStageService.DeletePipelineStage(cdPipelinePluginDeleteReq.PostDeployStage, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting post-CD stage", "err", err, "postDeployStage", cdPipelinePluginDeleteReq.PostDeployStage)
			return deleteResponse, err
		}
	}
	//delete app from argo cd, if created
	if pipeline.DeploymentAppCreated == true {
		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		if util.IsAcdApp(pipeline.DeploymentAppType) {
			if !deleteResponse.ClusterReachable {
				impl.logger.Errorw("cluster connection error", "err", clusterBean.ErrorInConnecting)
				if cascadeDelete {
					return deleteResponse, nil
				}
			}
			impl.logger.Debugw("acd app is already deleted for this pipeline", "pipeline", pipeline)
			if deleteFromAcd {
				req := &application2.ApplicationDeleteRequest{
					Name:    &deploymentAppName,
					Cascade: &cascadeDelete,
				}
				if _, err := impl.application.Delete(ctx, req); err != nil {
					impl.logger.Errorw("err in deleting pipeline on argocd", "id", pipeline, "err", err)

					if forceDelete {
						impl.logger.Warnw("error while deletion of app in acd, continue to delete in db as this operation is force delete", "error", err)
					} else {
						//statusError, _ := err.(*errors2.StatusError)
						if cascadeDelete && strings.Contains(err.Error(), "code = NotFound") {
							err = &util.ApiError{
								UserMessage:     "Could not delete as application not found in argocd",
								InternalMessage: err.Error(),
							}
						} else {
							err = &util.ApiError{
								UserMessage:     "Could not delete application",
								InternalMessage: err.Error(),
							}
						}
						return deleteResponse, err
					}
				}
				impl.logger.Infow("app deleted from argocd", "id", pipeline.Id, "pipelineName", pipeline.Name, "app", deploymentAppName)
			}
		} else if util.IsHelmApp(pipeline.DeploymentAppType) {
			appIdentifier := &client.AppIdentifier{
				ClusterId:   pipeline.Environment.ClusterId,
				ReleaseName: deploymentAppName,
				Namespace:   pipeline.Environment.Namespace,
			}
			deleteResourceResponse, err := impl.helmAppService.DeleteApplication(ctx, appIdentifier)
			if forceDelete || errors3.As(err, &models2.NamespaceNotExistError{}) {
				impl.logger.Warnw("error while deletion of helm application, ignore error and delete from db since force delete req", "error", err, "pipelineId", pipeline.Id)
			} else {
				if err != nil {
					impl.logger.Errorw("error in deleting helm application", "error", err, "appIdentifier", appIdentifier)
					return deleteResponse, err
				}
				if deleteResourceResponse == nil || !deleteResourceResponse.GetSuccess() {
					return deleteResponse, errors2.New("delete application response unsuccessful")
				}
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing db transaction", "err", err)
		return deleteResponse, err
	}
	deleteResponse.DeleteInitiated = true
	return deleteResponse, nil
}

func (impl *PipelineBuilderImpl) DeleteACDAppCdPipelineWithNonCascade(pipeline *pipelineConfig.Pipeline, ctx context.Context, forceDelete bool, userId int32) error {
	if forceDelete {
		_, err := impl.DeleteCdPipeline(pipeline, ctx, bean.FORCE_DELETE, false, userId)
		return err
	}
	//delete app from argo cd with non-cascade, if created
	if pipeline.DeploymentAppCreated && util.IsAcdApp(pipeline.DeploymentAppType) {
		appDetails, err := impl.appRepo.FindById(pipeline.AppId)
		deploymentAppName := fmt.Sprintf("%s-%s", appDetails.AppName, pipeline.Environment.Name)
		impl.logger.Debugw("acd app is already deleted for this pipeline", "pipeline", pipeline)
		cascadeDelete := false
		req := &application2.ApplicationDeleteRequest{
			Name:    &deploymentAppName,
			Cascade: &cascadeDelete,
		}
		if _, err = impl.application.Delete(ctx, req); err != nil {
			impl.logger.Errorw("err in deleting pipeline on argocd", "id", pipeline, "err", err)
			//statusError, _ := err.(*errors2.StatusError)
			if !strings.Contains(err.Error(), "code = NotFound") {
				err = &util.ApiError{
					UserMessage:     "Could not delete application",
					InternalMessage: err.Error(),
				}
				return err
			}
		}

	}
	return nil
}

func (impl *PipelineBuilderImpl) GetTriggerViewCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	triggerViewCdPipelinesResp, err := impl.ciCdPipelineOrchestrator.GetCdPipelinesForApp(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching triggerViewCdPipelinesResp by appId", "err", err, "appId", appId)
		return triggerViewCdPipelinesResp, err
	}
	var dbPipelineIds []int
	for _, dbPipeline := range triggerViewCdPipelinesResp.Pipelines {
		dbPipelineIds = append(dbPipelineIds, dbPipeline.Id)
	}

	//construct strategiesMapping to get all strategies against pipelineId
	strategiesMapping, err := impl.getStrategiesMapping(dbPipelineIds)
	if err != nil {
		return triggerViewCdPipelinesResp, err
	}
	for _, dbPipeline := range triggerViewCdPipelinesResp.Pipelines {
		var strategies []*chartConfig.PipelineStrategy
		var deploymentTemplate chartRepoRepository.DeploymentStrategy
		if len(strategiesMapping[dbPipeline.Id]) != 0 {
			strategies = strategiesMapping[dbPipeline.Id]
		}
		for _, item := range strategies {
			if item.Default {
				deploymentTemplate = item.Strategy
			}
		}
		dbPipeline.DeploymentTemplate = deploymentTemplate
	}

	return triggerViewCdPipelinesResp, err
}

func (impl *PipelineBuilderImpl) GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	cdPipelines, err = impl.ciCdPipelineOrchestrator.GetCdPipelinesForApp(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd Pipelines for appId", "err", err, "appId", appId)
		return nil, err
	}
	var envIds []*int
	var dbPipelineIds []int
	for _, dbPipeline := range cdPipelines.Pipelines {
		envIds = append(envIds, &dbPipeline.EnvironmentId)
		dbPipelineIds = append(dbPipelineIds, dbPipeline.Id)
	}
	if len(envIds) == 0 || len(dbPipelineIds) == 0 {
		return cdPipelines, nil
	}
	envMapping := make(map[int]*repository2.Environment)
	appWorkflowMapping := make(map[int]*appWorkflow.AppWorkflowMapping)

	envs, err := impl.environmentRepository.FindByIds(envIds)
	if err != nil && errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching environments", "err", err)
		return cdPipelines, err
	}
	//creating map for envId and respective env
	for _, env := range envs {
		envMapping[env.Id] = env
	}
	strategiesMapping, err := impl.getStrategiesMapping(dbPipelineIds)
	if err != nil {
		return cdPipelines, err
	}
	appWorkflowMappings, err := impl.appWorkflowRepository.FindByCDPipelineIds(dbPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching app workflow mappings by pipelineIds", "err", err)
		return nil, err
	}
	for _, appWorkflow := range appWorkflowMappings {
		appWorkflowMapping[appWorkflow.ComponentId] = appWorkflow
	}

	var pipelines []*bean.CDPipelineConfigObject
	for _, dbPipeline := range cdPipelines.Pipelines {
		environment := &repository2.Environment{}
		var strategies []*chartConfig.PipelineStrategy
		appToWorkflowMapping := &appWorkflow.AppWorkflowMapping{}

		if envMapping[dbPipeline.EnvironmentId] != nil {
			environment = envMapping[dbPipeline.EnvironmentId]
		}
		if len(strategiesMapping[dbPipeline.Id]) != 0 {
			strategies = strategiesMapping[dbPipeline.Id]
		}
		if appWorkflowMapping[dbPipeline.Id] != nil {
			appToWorkflowMapping = appWorkflowMapping[dbPipeline.Id]
		}
		var strategiesBean []bean.Strategy
		var deploymentTemplate chartRepoRepository.DeploymentStrategy
		for _, item := range strategies {
			strategiesBean = append(strategiesBean, bean.Strategy{
				Config:             []byte(item.Config),
				DeploymentTemplate: item.Strategy,
				Default:            item.Default,
			})

			if item.Default {
				deploymentTemplate = item.Strategy
			}
		}
		pipeline := &bean.CDPipelineConfigObject{
			Id:                            dbPipeline.Id,
			Name:                          dbPipeline.Name,
			EnvironmentId:                 dbPipeline.EnvironmentId,
			EnvironmentName:               environment.Name,
			Description:                   environment.Description,
			CiPipelineId:                  dbPipeline.CiPipelineId,
			DeploymentTemplate:            deploymentTemplate,
			TriggerType:                   dbPipeline.TriggerType,
			Strategies:                    strategiesBean,
			PreStage:                      dbPipeline.PreStage,
			PostStage:                     dbPipeline.PostStage,
			PreStageConfigMapSecretNames:  dbPipeline.PreStageConfigMapSecretNames,
			PostStageConfigMapSecretNames: dbPipeline.PostStageConfigMapSecretNames,
			RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
			DeploymentAppType:             dbPipeline.DeploymentAppType,
			ParentPipelineType:            appToWorkflowMapping.ParentType,
			ParentPipelineId:              appToWorkflowMapping.ParentId,
			DeploymentAppDeleteRequest:    dbPipeline.DeploymentAppDeleteRequest,
			IsVirtualEnvironment:          dbPipeline.IsVirtualEnvironment,
			PreDeployStage:                dbPipeline.PreDeployStage,
			PostDeployStage:               dbPipeline.PostDeployStage,
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines.Pipelines = pipelines
	return cdPipelines, err
}

func (impl *PipelineBuilderImpl) GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error) {
	return impl.ciCdPipelineOrchestrator.GetCdPipelinesForAppAndEnv(appId, envId)
}

func (impl PipelineBuilderImpl) GetCdPipelinesByEnvironment(request resourceGroup2.ResourceGroupingRequest) (cdPipelines *bean.CdPipelines, err error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.authorizationCdPipelinesForResourceGrouping")
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	cdPipelines, err = impl.ciCdPipelineOrchestrator.GetCdPipelinesForEnv(request.ParentResourceId, request.ResourceIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return cdPipelines, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines.Pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return cdPipelines, err
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipeline(cdPipelines.Pipelines)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	//authorization block ends here
	span.End()
	var pipelines []*bean.CDPipelineConfigObject
	authorizedPipelines := make(map[int]*bean.CDPipelineConfigObject)
	for _, dbPipeline := range cdPipelines.Pipelines {
		appObject := objects[dbPipeline.Id][0]
		envObject := objects[dbPipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pipelineIds = append(pipelineIds, dbPipeline.Id)
		authorizedPipelines[dbPipeline.Id] = dbPipeline
	}

	pipelineDeploymentTemplate := make(map[int]chartRepoRepository.DeploymentStrategy)
	pipelineWorkflowMapping := make(map[int]*appWorkflow.AppWorkflowMapping)
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no authorized pipeline found"}
		return cdPipelines, err
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.GetAllStrategyByPipelineIds")
	strategies, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineIds(pipelineIds)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching strategies", "err", err)
		return cdPipelines, err
	}
	for _, item := range strategies {
		if item.Default {
			pipelineDeploymentTemplate[item.PipelineId] = item.Strategy
		}
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.FindByCDPipelineIds")
	appWorkflowMappings, err := impl.appWorkflowRepository.FindByCDPipelineIds(pipelineIds)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching workflows", "err", err)
		return nil, err
	}
	for _, item := range appWorkflowMappings {
		pipelineWorkflowMapping[item.ComponentId] = item
	}

	for _, dbPipeline := range authorizedPipelines {
		pipeline := &bean.CDPipelineConfigObject{
			Id:                            dbPipeline.Id,
			Name:                          dbPipeline.Name,
			EnvironmentId:                 dbPipeline.EnvironmentId,
			EnvironmentName:               dbPipeline.EnvironmentName,
			CiPipelineId:                  dbPipeline.CiPipelineId,
			DeploymentTemplate:            pipelineDeploymentTemplate[dbPipeline.Id],
			TriggerType:                   dbPipeline.TriggerType,
			PreStage:                      dbPipeline.PreStage,
			PostStage:                     dbPipeline.PostStage,
			PreStageConfigMapSecretNames:  dbPipeline.PreStageConfigMapSecretNames,
			PostStageConfigMapSecretNames: dbPipeline.PostStageConfigMapSecretNames,
			RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
			DeploymentAppType:             dbPipeline.DeploymentAppType,
			ParentPipelineType:            pipelineWorkflowMapping[dbPipeline.Id].ParentType,
			ParentPipelineId:              pipelineWorkflowMapping[dbPipeline.Id].ParentId,
			AppName:                       dbPipeline.AppName,
			AppId:                         dbPipeline.AppId,
			IsVirtualEnvironment:          dbPipeline.IsVirtualEnvironment,
			PreDeployStage:                dbPipeline.PreDeployStage,
			PostDeployStage:               dbPipeline.PostDeployStage,
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines.Pipelines = pipelines
	return cdPipelines, err
}

func (impl PipelineBuilderImpl) GetCdPipelinesByEnvironmentMin(request resourceGroup2.ResourceGroupingRequest) (cdPipelines []*bean.CDPipelineConfigObject, err error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "cdHandler.authorizationCdPipelinesForResourceGrouping")
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return cdPipelines, err
		}
		//override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	var pipelines []*pipelineConfig.Pipeline
	if len(request.ResourceIds) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return cdPipelines, err
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(pipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	//authorization block ends here
	span.End()
	for _, dbPipeline := range pipelines {
		appObject := objects[dbPipeline.Id][0]
		envObject := objects[dbPipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pcObject := &bean.CDPipelineConfigObject{
			AppId:                dbPipeline.AppId,
			AppName:              dbPipeline.App.AppName,
			EnvironmentId:        dbPipeline.EnvironmentId,
			Id:                   dbPipeline.Id,
			DeploymentAppType:    dbPipeline.DeploymentAppType,
			IsVirtualEnvironment: dbPipeline.Environment.IsVirtualEnvironment,
		}
		cdPipelines = append(cdPipelines, pcObject)
	}
	return cdPipelines, err
}

func (impl *PipelineBuilderImpl) PerformBulkActionOnCdPipelines(dto *bean.CdBulkActionRequestDto, impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun bool, userId int32) ([]*bean.CdBulkActionResponseDto, error) {
	switch dto.Action {
	case bean.CD_BULK_DELETE:
		deleteAction := bean.CASCADE_DELETE
		if dto.ForceDelete {
			deleteAction = bean.FORCE_DELETE
		} else if !dto.CascadeDelete {
			deleteAction = bean.NON_CASCADE_DELETE
		}
		bulkDeleteResp := impl.BulkDeleteCdPipelines(impactedPipelines, ctx, dryRun, deleteAction, userId)
		return bulkDeleteResp, nil
	default:
		return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "this action is not supported"}
	}
}

func (impl *PipelineBuilderImpl) FindPipelineById(cdPipelineId int) (*pipelineConfig.Pipeline, error) {
	return impl.pipelineRepository.FindById(cdPipelineId)
}

func (impl *PipelineBuilderImpl) FindAppAndEnvDetailsByPipelineId(cdPipelineId int) (*pipelineConfig.Pipeline, error) {
	return impl.pipelineRepository.FindAppAndEnvDetailsByPipelineId(cdPipelineId)
}

func (impl *PipelineBuilderImpl) RetrieveParentDetails(pipelineId int) (parentId int, parentType bean2.WorkflowType, err error) {

	workflow, err := impl.appWorkflowRepository.GetParentDetailsByPipelineId(pipelineId)
	if err != nil {
		impl.logger.Errorw("failed to get parent component details",
			"componentId", pipelineId,
			"err", err)
		return 0, "", err
	}

	if workflow.ParentType == appWorkflow.CDPIPELINE {
		// workflow is of type CD, check for stage
		parentPipeline, err := impl.pipelineRepository.GetPostStageConfigById(workflow.ParentId)
		if err != nil {
			impl.logger.Errorw("failed to get the post_stage_config_yaml",
				"cdPipelineId", workflow.ParentId,
				"err", err)
			return 0, "", err
		}

		if len(parentPipeline.PostStageConfig) == 0 {
			return workflow.ParentId, bean2.CD_WORKFLOW_TYPE_DEPLOY, nil
		}
		return workflow.ParentId, bean2.CD_WORKFLOW_TYPE_POST, nil

	} else if workflow.ParentType == appWorkflow.WEBHOOK {
		// For webhook type
		return workflow.ParentId, bean2.WEBHOOK_WORKFLOW_TYPE, nil
	}

	return workflow.ParentId, bean2.CI_WORKFLOW_TYPE, nil
}

func (impl *PipelineBuilderImpl) GetEnvironmentByCdPipelineId(pipelineId int) (int, error) {
	dbPipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil || dbPipeline == nil {
		impl.logger.Errorw("error in fetching pipeline", "err", err)
		return 0, err
	}
	return dbPipeline.EnvironmentId, err
}

func (impl *PipelineBuilderImpl) GetBulkActionImpactedPipelines(dto *bean.CdBulkActionRequestDto) ([]*pipelineConfig.Pipeline, error) {
	if len(dto.EnvIds) == 0 || (len(dto.AppIds) == 0 && len(dto.ProjectIds) == 0) {
		//invalid payload, envIds are must and either of appIds or projectIds are must
		return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "invalid payload, can not get pipelines for this filter"}
	}
	var pipelineIdsByAppLevel []int
	var pipelineIdsByProjectLevel []int
	var err error
	if len(dto.AppIds) > 0 && len(dto.EnvIds) > 0 {
		//getting pipeline IDs for app level deletion request
		pipelineIdsByAppLevel, err = impl.pipelineRepository.FindIdsByAppIdsAndEnvironmentIds(dto.AppIds, dto.EnvIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting cd pipelines by appIds and envIds", "err", err)
			return nil, err
		}
	}
	if len(dto.ProjectIds) > 0 && len(dto.EnvIds) > 0 {
		//getting pipeline IDs for project level deletion request
		pipelineIdsByProjectLevel, err = impl.pipelineRepository.FindIdsByProjectIdsAndEnvironmentIds(dto.ProjectIds, dto.EnvIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting cd pipelines by projectIds and envIds", "err", err)
			return nil, err
		}
	}
	var pipelineIdsMerged []int
	//it might be possible that pipelineIdsByAppLevel & pipelineIdsByProjectLevel have some same values
	//we are still appending them to save operation cost of checking same ids as we will get pipelines from
	//in clause which gives correct results even if some values are repeating
	pipelineIdsMerged = append(pipelineIdsMerged, pipelineIdsByAppLevel...)
	pipelineIdsMerged = append(pipelineIdsMerged, pipelineIdsByProjectLevel...)
	var pipelines []*pipelineConfig.Pipeline
	if len(pipelineIdsMerged) > 0 {
		pipelines, err = impl.pipelineRepository.FindByIdsIn(pipelineIdsMerged)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipelines by ids", "err", err, "ids", pipelineIdsMerged)
			return nil, err
		}
	}
	return pipelines, nil
}

func (impl *PipelineBuilderImpl) IsGitOpsRequiredForCD(pipelineCreateRequest *bean.CdPipelines) bool {

	// if deploymentAppType is not coming in request than hasAtLeastOneGitOps will be false

	haveAtLeastOneGitOps := false
	for _, pipeline := range pipelineCreateRequest.Pipelines {
		if pipeline.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
			haveAtLeastOneGitOps = true
		}
	}
	return haveAtLeastOneGitOps
}

func (impl *PipelineBuilderImpl) SetPipelineDeploymentAppType(pipelineCreateRequest *bean.CdPipelines, isGitOpsConfigured bool, deploymentTypeValidationConfig map[string]bool) {
	for _, pipeline := range pipelineCreateRequest.Pipelines {
		// by default both deployment app type are allowed
		AllowedDeploymentAppTypes := map[string]bool{
			util.PIPELINE_DEPLOYMENT_TYPE_ACD:  true,
			util.PIPELINE_DEPLOYMENT_TYPE_HELM: true,
		}
		for k, v := range deploymentTypeValidationConfig {
			// rewriting allowed deployment types based on config provided by user
			AllowedDeploymentAppTypes[k] = v
		}
		if !impl.deploymentConfig.IsInternalUse {
			if isGitOpsConfigured && AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_ACD] {
				pipeline.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
			} else if AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_HELM] {
				pipeline.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
			}
		}
		if pipeline.DeploymentAppType == "" {
			if isGitOpsConfigured && AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_ACD] {
				pipeline.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
			} else if AllowedDeploymentAppTypes[util.PIPELINE_DEPLOYMENT_TYPE_HELM] {
				pipeline.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
			}
		}
	}
}

func (impl *PipelineBuilderImpl) MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(appId int, envId int, acdToken string, pipeline *pipelineConfig.Pipeline) (bool, error) {

	acdAppFound := false
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	acdAppName := pipeline.DeploymentAppName
	_, err := impl.application.Get(ctx, &application2.ApplicationQuery{Name: &acdAppName})
	if err == nil {
		// acd app is not yet deleted so return
		acdAppFound = true
		return acdAppFound, err
	}
	impl.logger.Warnw("app not found in argo, deleting from db ", "err", err)
	//make call to delete it from pipeline DB because it's ACD counterpart is deleted
	_, err = impl.DeleteCdPipeline(pipeline, context.Background(), bean.FORCE_DELETE, false, 1)
	if err != nil {
		impl.logger.Errorw("error in deleting cd pipeline", "err", err)
		return acdAppFound, err
	}
	return acdAppFound, nil
}

func (impl PipelineBuilderImpl) GetEnvironmentListForAutocompleteFilter(envName string, clusterIds []int, offset int, size int, emailId string, checkAuthBatch func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool), ctx context.Context) (*cluster.ResourceGroupingResponse, error) {
	result := &cluster.ResourceGroupingResponse{}
	var models []*repository2.Environment
	var beans []cluster.EnvironmentBean
	var err error
	if len(envName) > 0 && len(clusterIds) > 0 {
		models, err = impl.environmentRepository.FindByEnvNameAndClusterIds(envName, clusterIds)
	} else if len(clusterIds) > 0 {
		models, err = impl.environmentRepository.FindByClusterIdsWithFilter(clusterIds)
	} else if len(envName) > 0 {
		models, err = impl.environmentRepository.FindByEnvName(envName)
	} else {
		models, err = impl.environmentRepository.FindAllActiveWithFilter()
	}
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching environment", "err", err)
		return result, err
	}
	var envIds []int
	for _, model := range models {
		envIds = append(envIds, model.Id)
	}
	if len(envIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching environment found"}
		return nil, err
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineBuilder.FindActiveByEnvIds")
	cdPipelines, err := impl.pipelineRepository.FindActiveByEnvIds(envIds)
	span.End()
	if err != nil && err != pg.ErrNoRows {
		return result, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return nil, err
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineBuilder.GetAppAndEnvObjectByPipelineIds")
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	span.End()
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineBuilder.checkAuthBatch")
	appResults, envResults := checkAuthBatch(emailId, appObjectArr, envObjectArr)
	span.End()
	//authorization block ends here

	pipelinesMap := make(map[int][]*pipelineConfig.Pipeline)
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		pipelinesMap[pipeline.EnvironmentId] = append(pipelinesMap[pipeline.EnvironmentId], pipeline)
	}
	for _, model := range models {
		environment := cluster.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			Namespace:             model.Namespace,
			CdArgoSetup:           model.Cluster.CdArgoSetup,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			ClusterName:           model.Cluster.ClusterName,
			IsVirtualEnvironment:  model.IsVirtualEnvironment,
		}

		//authorization block starts here
		appCount := 0
		envPipelines := pipelinesMap[model.Id]
		if _, ok := pipelinesMap[model.Id]; ok {
			appCount = len(envPipelines)
		}
		environment.AppCount = appCount
		beans = append(beans, environment)
	}

	envCount := len(beans)
	// Apply pagination
	if size > 0 {
		if offset+size <= len(beans) {
			beans = beans[offset : offset+size]
		} else {
			beans = beans[offset:]
		}
	}
	result.EnvList = beans
	result.EnvCount = envCount
	return result, nil
}

func (impl *PipelineBuilderImpl) IsGitopsConfigured() (bool, error) {

	isGitOpsConfigured := false
	gitOpsConfig, err := impl.gitOpsRepository.GetGitOpsConfigActive()

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GetGitOpsConfigActive, error while getting", "err", err)
		return false, err
	}
	if gitOpsConfig != nil && gitOpsConfig.Id > 0 {
		isGitOpsConfigured = true
	}

	return isGitOpsConfigured, nil

}

func (impl *PipelineBuilderImpl) FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error) {
	configMapSecrets, err := impl.appService.GetConfigMapAndSecretJson(appId, envId, cdPipelineId)
	if err != nil {
		impl.logger.Errorw("error while fetching config secrets ", "err", err)
		return ConfigMapSecretsResponse{}, err
	}
	existingConfigMapSecrets := ConfigMapSecretsResponse{}
	err = json.Unmarshal([]byte(configMapSecrets), &existingConfigMapSecrets)
	if err != nil {
		impl.logger.Error(err)
		return ConfigMapSecretsResponse{}, err
	}
	return existingConfigMapSecrets, nil
}

func (impl *PipelineBuilderImpl) GetDeploymentConfigMap(environmentId int) (map[string]bool, error) {
	var deploymentConfig map[string]map[string]bool
	var deploymentConfigEnv map[string]bool
	deploymentConfigValues, err := impl.attributesRepository.FindByKey(attributes.ENFORCE_DEPLOYMENT_TYPE_CONFIG)
	if err == pg.ErrNoRows {
		return deploymentConfigEnv, nil
	}
	//if empty config received(doesn't exist in table) which can't be parsed
	if deploymentConfigValues.Value != "" {
		if err := json.Unmarshal([]byte(deploymentConfigValues.Value), &deploymentConfig); err != nil {
			rerr := &util.ApiError{
				HttpStatusCode:  http.StatusInternalServerError,
				InternalMessage: err.Error(),
				UserMessage:     "Failed to fetch deployment config values from the attributes table",
			}
			return deploymentConfigEnv, rerr
		}
		deploymentConfigEnv, _ = deploymentConfig[fmt.Sprintf("%d", environmentId)]
	}
	return deploymentConfigEnv, nil
}

func (impl *PipelineBuilderImpl) FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error) {
	pipelineStrategiesResponse := PipelineStrategiesResponse{}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorf("invalid state", "err", err, "appId", appId)
		return pipelineStrategiesResponse, err
	}
	if chart.Id == 0 {
		return pipelineStrategiesResponse, fmt.Errorf("no chart configured")
	}

	//get global strategy for this chart
	globalStrategies, err := impl.globalStrategyMetadataChartRefMappingRepository.GetByChartRefId(chart.ChartRefId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting global strategies", "err", err)
		return pipelineStrategiesResponse, err
	} else if err == pg.ErrNoRows {
		impl.logger.Infow("no strategies configured for chart", "chartRefId", chart.ChartRefId)
		return pipelineStrategiesResponse, nil
	}
	pipelineOverride := chart.PipelineOverride
	for _, globalStrategy := range globalStrategies {
		pipelineStrategyJson, err := impl.filterDeploymentTemplate(globalStrategy.GlobalStrategyMetadata.Key, pipelineOverride)
		if err != nil {
			return pipelineStrategiesResponse, err
		}
		pipelineStrategy := PipelineStrategy{
			DeploymentTemplate: globalStrategy.GlobalStrategyMetadata.Name,
			Config:             []byte(pipelineStrategyJson),
		}
		pipelineStrategy.Default = globalStrategy.Default
		pipelineStrategiesResponse.PipelineStrategy = append(pipelineStrategiesResponse.PipelineStrategy, pipelineStrategy)
	}
	return pipelineStrategiesResponse, nil
}

func (impl *PipelineBuilderImpl) FetchDefaultCDPipelineStrategy(appId int, envId int) (PipelineStrategy, error) {
	pipelineStrategy := PipelineStrategy{}
	cdPipelines, err := impl.ciCdPipelineOrchestrator.GetCdPipelinesForAppAndEnv(appId, envId)
	if err != nil || (cdPipelines.Pipelines) == nil || len(cdPipelines.Pipelines) == 0 {
		return pipelineStrategy, err
	}
	cdPipelineId := cdPipelines.Pipelines[0].Id

	cdPipeline, err := impl.GetCdPipelineById(cdPipelineId)
	if err != nil {
		return pipelineStrategy, nil
	}
	pipelineStrategy.DeploymentTemplate = cdPipeline.DeploymentTemplate
	pipelineStrategy.Default = true
	return pipelineStrategy, nil
}

func (impl *PipelineBuilderImpl) ChangeDeploymentType(ctx context.Context,
	request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {

	var response *bean.DeploymentAppTypeChangeResponse
	var deleteDeploymentType bean.DeploymentType
	var err error

	if request.DesiredDeploymentType == bean.ArgoCd {
		deleteDeploymentType = bean.Helm
	} else {
		deleteDeploymentType = bean.ArgoCd
	}

	// Force delete apps
	response, err = impl.DeleteDeploymentAppsForEnvironment(ctx,
		request.EnvId, deleteDeploymentType, request.ExcludeApps, request.IncludeApps, request.UserId)

	if err != nil {
		return nil, err
	}

	// Updating the env id and desired deployment app type received from request in the response
	response.EnvId = request.EnvId
	response.DesiredDeploymentType = request.DesiredDeploymentType
	response.TriggeredPipelines = make([]*bean.CdPipelineTrigger, 0)

	// Update the deployment app type to Helm and toggle deployment_app_created to false in db
	var cdPipelineIds []int
	for _, item := range response.SuccessfulPipelines {
		cdPipelineIds = append(cdPipelineIds, item.Id)
	}

	// If nothing to update in db
	if len(cdPipelineIds) == 0 {
		return response, nil
	}

	// Update in db
	err = impl.pipelineRepository.UpdateCdPipelineDeploymentAppInFilter(string(request.DesiredDeploymentType),
		cdPipelineIds, request.UserId, false, true)

	if err != nil {
		impl.logger.Errorw("failed to update deployment app type in db",
			"pipeline ids", cdPipelineIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, nil
	}

	if !request.AutoTriggerDeployment {
		return response, nil
	}

	// Bulk trigger all the successfully changed pipelines (async)
	bulkTriggerRequest := make([]*BulkTriggerRequest, 0)

	pipelineIds := make([]int, 0, len(response.SuccessfulPipelines))
	for _, item := range response.SuccessfulPipelines {
		pipelineIds = append(pipelineIds, item.Id)
	}

	// Get all pipelines
	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("failed to fetch pipeline details",
			"ids", pipelineIds,
			"err", err)

		return response, nil
	}

	for _, pipeline := range pipelines {

		artifactDetails, err := impl.RetrieveArtifactsByCDPipeline(pipeline, "DEPLOY")

		if err != nil {
			impl.logger.Errorw("failed to fetch artifact details for cd pipeline",
				"pipelineId", pipeline.Id,
				"appId", pipeline.AppId,
				"envId", pipeline.EnvironmentId,
				"err", err)

			return response, nil
		}

		if artifactDetails.LatestWfArtifactId == 0 || artifactDetails.LatestWfArtifactStatus == "" {
			continue
		}

		bulkTriggerRequest = append(bulkTriggerRequest, &BulkTriggerRequest{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
		response.TriggeredPipelines = append(response.TriggeredPipelines, &bean.CdPipelineTrigger{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
	}

	// pg panics if empty slice is passed as an argument
	if len(bulkTriggerRequest) == 0 {
		return response, nil
	}

	// Trigger
	_, err = impl.workflowDagExecutor.TriggerBulkDeploymentAsync(bulkTriggerRequest, request.UserId)

	if err != nil {
		impl.logger.Errorw("failed to bulk trigger cd pipelines with error: "+err.Error(),
			"err", err)
	}
	return response, nil
}

func (impl *PipelineBuilderImpl) ChangePipelineDeploymentType(ctx context.Context,
	request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {

	response := &bean.DeploymentAppTypeChangeResponse{
		EnvId:                 request.EnvId,
		DesiredDeploymentType: request.DesiredDeploymentType,
		TriggeredPipelines:    make([]*bean.CdPipelineTrigger, 0),
	}

	var deleteDeploymentType bean.DeploymentType

	if request.DesiredDeploymentType == bean.ArgoCd {
		deleteDeploymentType = bean.Helm
	} else {
		deleteDeploymentType = bean.ArgoCd
	}

	pipelines, err := impl.pipelineRepository.FindActiveByEnvIdAndDeploymentType(request.EnvId,
		string(deleteDeploymentType), request.ExcludeApps, request.IncludeApps)

	if err != nil {
		impl.logger.Errorw("Error fetching cd pipelines",
			"environmentId", request.EnvId,
			"currentDeploymentAppType", string(deleteDeploymentType),
			"err", err)
		return response, err
	}

	var pipelineIds []int
	for _, item := range pipelines {
		pipelineIds = append(pipelineIds, item.Id)
	}

	if len(pipelineIds) == 0 {
		return response, nil
	}

	err = impl.pipelineRepository.UpdateCdPipelineDeploymentAppInFilter(string(request.DesiredDeploymentType),
		pipelineIds, request.UserId, false, true)

	if err != nil {
		impl.logger.Errorw("failed to update deployment app type in db",
			"pipeline ids", pipelineIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, nil
	}
	deleteResponse := impl.DeleteDeploymentApps(ctx, pipelines, request.UserId)

	response.SuccessfulPipelines = deleteResponse.SuccessfulPipelines
	response.FailedPipelines = deleteResponse.FailedPipelines

	var cdPipelineIds []int
	for _, item := range response.FailedPipelines {
		cdPipelineIds = append(cdPipelineIds, item.Id)
	}

	if len(cdPipelineIds) == 0 {
		return response, nil
	}

	err = impl.pipelineRepository.UpdateCdPipelineDeploymentAppInFilter(string(deleteDeploymentType),
		cdPipelineIds, request.UserId, true, false)

	if err != nil {
		impl.logger.Errorw("failed to update deployment app type in db",
			"pipeline ids", cdPipelineIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, nil
	}

	return response, nil
}

func (impl *PipelineBuilderImpl) TriggerDeploymentAfterTypeChange(ctx context.Context,
	request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {

	response := &bean.DeploymentAppTypeChangeResponse{
		EnvId:                 request.EnvId,
		DesiredDeploymentType: request.DesiredDeploymentType,
		TriggeredPipelines:    make([]*bean.CdPipelineTrigger, 0),
	}
	var err error

	cdPipelines, err := impl.pipelineRepository.FindActiveByEnvIdAndDeploymentType(request.EnvId,
		string(request.DesiredDeploymentType), request.ExcludeApps, request.IncludeApps)

	if err != nil {
		impl.logger.Errorw("Error fetching cd pipelines",
			"environmentId", request.EnvId,
			"desiredDeploymentAppType", string(request.DesiredDeploymentType),
			"err", err)
		return response, err
	}

	var cdPipelineIds []int
	for _, item := range cdPipelines {
		cdPipelineIds = append(cdPipelineIds, item.Id)
	}

	if len(cdPipelineIds) == 0 {
		return response, nil
	}

	deleteResponse := impl.FetchDeletedApp(ctx, cdPipelines)

	response.SuccessfulPipelines = deleteResponse.SuccessfulPipelines
	response.FailedPipelines = deleteResponse.FailedPipelines

	var successPipelines []int
	for _, item := range response.SuccessfulPipelines {
		successPipelines = append(successPipelines, item.Id)
	}

	bulkTriggerRequest := make([]*BulkTriggerRequest, 0)

	pipelineIds := make([]int, 0, len(response.SuccessfulPipelines))
	for _, item := range response.SuccessfulPipelines {
		pipelineIds = append(pipelineIds, item.Id)
	}

	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("failed to fetch pipeline details",
			"ids", pipelineIds,
			"err", err)

		return response, nil
	}

	for _, pipeline := range pipelines {

		artifactDetails, err := impl.RetrieveArtifactsByCDPipeline(pipeline, "DEPLOY")

		if err != nil {
			impl.logger.Errorw("failed to fetch artifact details for cd pipeline",
				"pipelineId", pipeline.Id,
				"appId", pipeline.AppId,
				"envId", pipeline.EnvironmentId,
				"err", err)

			return response, nil
		}

		if artifactDetails.LatestWfArtifactId == 0 || artifactDetails.LatestWfArtifactStatus == "" {
			continue
		}

		bulkTriggerRequest = append(bulkTriggerRequest, &BulkTriggerRequest{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
		response.TriggeredPipelines = append(response.TriggeredPipelines, &bean.CdPipelineTrigger{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
	}

	if len(bulkTriggerRequest) == 0 {
		return response, nil
	}

	_, err = impl.workflowDagExecutor.TriggerBulkDeploymentAsync(bulkTriggerRequest, request.UserId)

	if err != nil {
		impl.logger.Errorw("failed to bulk trigger cd pipelines with error: "+err.Error(),
			"err", err)
	}

	err = impl.pipelineRepository.UpdateCdPipelineAfterDeployment(string(request.DesiredDeploymentType),
		successPipelines, request.UserId, false)

	if err != nil {
		impl.logger.Errorw("failed to update cd pipelines with error: : "+err.Error(),
			"err", err)
	}

	return response, nil
}

func (impl *PipelineBuilderImpl) DeleteDeploymentApps(ctx context.Context,
	pipelines []*pipelineConfig.Pipeline, userId int32) *bean.DeploymentAppTypeChangeResponse {

	successfulPipelines := make([]*bean.DeploymentChangeStatus, 0)
	failedPipelines := make([]*bean.DeploymentChangeStatus, 0)

	isGitOpsConfigured, gitOpsConfigErr := impl.IsGitopsConfigured()

	// Iterate over all the pipelines in the environment for given deployment app type
	for _, pipeline := range pipelines {

		var isValid bool
		// check if pipeline info like app name and environment is empty or not
		if failedPipelines, isValid = impl.isPipelineInfoValid(pipeline, failedPipelines); !isValid {
			continue
		}

		var healthChkErr error
		// check health of the app if it is argocd deployment type
		if _, healthChkErr = impl.handleNotDeployedAppsIfArgoDeploymentType(pipeline, failedPipelines); healthChkErr != nil {

			// cannot delete unhealthy app
			continue
		}

		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		var err error

		// delete request
		if pipeline.DeploymentAppType == bean.ArgoCd {
			err = impl.deleteArgoCdApp(ctx, pipeline, deploymentAppName, true)

		} else {

			// For converting from Helm to ArgoCD, GitOps should be configured
			if gitOpsConfigErr != nil || !isGitOpsConfigured {
				err = errors2.New("GitOps not configured or unable to fetch GitOps configuration")

			} else {
				// Register app in ACD
				var AcdRegisterErr, RepoURLUpdateErr error
				gitopsRepoName, chartGitAttr, createGitRepoErr := impl.appService.CreateGitopsRepo(&app.App{Id: pipeline.AppId, AppName: pipeline.App.AppName}, userId)
				if createGitRepoErr != nil {
					impl.logger.Errorw("error increating git repo", "err", err)
				}
				if createGitRepoErr == nil {
					AcdRegisterErr = impl.RegisterInACD(gitopsRepoName,
						chartGitAttr,
						userId,
						ctx)
					if AcdRegisterErr != nil {
						impl.logger.Errorw("error in registering acd app", "err", err)
					}
					if AcdRegisterErr == nil {
						RepoURLUpdateErr = impl.chartTemplateService.UpdateGitRepoUrlInCharts(pipeline.AppId, chartGitAttr, userId)
						if RepoURLUpdateErr != nil {
							impl.logger.Errorw("error in updating git repo url in charts", "err", err)
						}
					}
				}
				if createGitRepoErr != nil {
					err = createGitRepoErr
				} else if AcdRegisterErr != nil {
					err = AcdRegisterErr
				} else if RepoURLUpdateErr != nil {
					err = RepoURLUpdateErr
				}
			}
			if err != nil {
				impl.logger.Errorw("error registering app on ACD with error: "+err.Error(),
					"deploymentAppName", deploymentAppName,
					"envId", pipeline.EnvironmentId,
					"appId", pipeline.AppId,
					"err", err)

				// deletion failed, append to the list of failed pipelines
				failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
					"failed to register app on ACD with error: "+err.Error())

				continue
			}
			err = impl.deleteHelmApp(ctx, pipeline)
		}

		if err != nil {
			impl.logger.Errorw("error deleting app on "+pipeline.DeploymentAppType,
				"deployment app name", deploymentAppName,
				"err", err)

			// deletion failed, append to the list of failed pipelines
			failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
				"error deleting app with error: "+err.Error())

			continue
		}

		// deletion successful, append to the list of successful pipelines
		successfulPipelines = impl.appendToDeploymentChangeStatusList(
			successfulPipelines,
			pipeline,
			"",
			bean.INITIATED)
	}

	return &bean.DeploymentAppTypeChangeResponse{
		SuccessfulPipelines: successfulPipelines,
		FailedPipelines:     failedPipelines,
	}
}

func (impl *PipelineBuilderImpl) DeleteDeploymentAppsForEnvironment(ctx context.Context, environmentId int,
	currentDeploymentAppType bean.DeploymentType, exclusionList []int, includeApps []int, userId int32) (*bean.DeploymentAppTypeChangeResponse, error) {

	// fetch active pipelines from database for the given environment id and current deployment app type
	pipelines, err := impl.pipelineRepository.FindActiveByEnvIdAndDeploymentType(environmentId,
		string(currentDeploymentAppType), exclusionList, includeApps)

	if err != nil {
		impl.logger.Errorw("Error fetching cd pipelines",
			"environmentId", environmentId,
			"currentDeploymentAppType", currentDeploymentAppType,
			"err", err)

		return &bean.DeploymentAppTypeChangeResponse{
			EnvId:               environmentId,
			SuccessfulPipelines: []*bean.DeploymentChangeStatus{},
			FailedPipelines:     []*bean.DeploymentChangeStatus{},
		}, err
	}

	// Currently deleting apps only in argocd is supported
	return impl.DeleteDeploymentApps(ctx, pipelines, userId), nil
}
