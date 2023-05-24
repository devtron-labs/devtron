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

package batch

import (
	"context"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

var (
	LoggerMock = zap.SugaredLogger{}
)

//--------------
type AppRepositoryMock struct{}

func (repo AppRepositoryMock) Save(pipelineGroup *app.App) error {
	panic("implement me")
}

func (repo AppRepositoryMock) Update(app *app.App) error {
	panic("implement me")
}

func (repo AppRepositoryMock) FindActiveByName(appName string) (*app.App, error) {
	return &app.App{Id: 1}, nil
	//panic("implement me")
}

func (repo AppRepositoryMock) AppExists(appName string) (bool, error) {
	panic("implement me")
}

func (repo AppRepositoryMock) FindById(id int) (*app.App, error) {
	panic("implement me")
}

func (repo AppRepositoryMock) FindAppsByTeamId(teamId int) ([]app.App, error) {
	panic("implement me")
}

func (repo AppRepositoryMock) FindAppsByTeamName(teamName string) ([]app.App, error) {
	panic("implement me")
}

func (repo AppRepositoryMock) FindAll() ([]app.App, error) {
	panic("implement me")
}

func (repo AppRepositoryMock) FindAppsByEnvironmentId(environmentId int) ([]app.App, error) {
	panic("implement me")
}

//--------------
type ConfigMapServiceMock struct{}

func (impl ConfigMapServiceMock) CMGlobalAddUpdate(configMapRequest *pipeline.ConfigDataRequest) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CMGlobalFetch(appId int) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CMEnvironmentAddUpdate(configMapRequest *pipeline.ConfigDataRequest) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CMEnvironmentFetch(appId int, envId int) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

// ---------------------------------------------------------------------------------------------

func (impl ConfigMapServiceMock) CSGlobalAddUpdate(configMapRequest *pipeline.ConfigDataRequest) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSGlobalFetch(appId int) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSEnvironmentAddUpdate(configMapRequest *pipeline.ConfigDataRequest) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSEnvironmentFetch(appId int, envId int) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CMGlobalDelete(name string, id int, userId int32) (bool, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CMEnvironmentDelete(name string, id int, userId int32) (bool, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSGlobalDelete(name string, id int, userId int32) (bool, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSEnvironmentDelete(name string, id int, userId int32) (bool, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CMGlobalDeleteByAppId(name string, appId int, userId int32) (bool, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CMEnvironmentDeleteByAppIdAndEnvId(name string, appId int, envId int, userId int32) (bool, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSGlobalDeleteByAppId(name string, appId int, userId int32) (bool, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSEnvironmentDeleteByAppIdAndEnvId(name string, appId int, envId int, userId int32) (bool, error) {
	panic("implement me")
}

////

func (impl ConfigMapServiceMock) CSGlobalFetchForEdit(name string, id int, userId int32) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

func (impl ConfigMapServiceMock) CSEnvironmentFetchForEdit(name string, id int, appId int, envId int, userId int32) (*pipeline.ConfigDataRequest, error) {
	panic("implement me")
}

type EnvironmentServiceMock struct{}

func (impl EnvironmentServiceMock) Create(mappings *cluster.EnvironmentBean, userId int32) (*cluster.EnvironmentBean, error) {
	panic("implement me")
}

func (impl EnvironmentServiceMock) FindOne(environment string) (*cluster.EnvironmentBean, error) {
	return &cluster.EnvironmentBean{Id: 1}, nil
	//panic("implement me")
}

func (impl EnvironmentServiceMock) GetAll() ([]cluster.EnvironmentBean, error) {
	panic("implement me")
}

func (impl EnvironmentServiceMock) GetAllActive() ([]cluster.EnvironmentBean, error) {
	panic("implement me")
}

func (impl EnvironmentServiceMock) FindById(id int) (*cluster.EnvironmentBean, error) {
	panic("implement me")
}

func (impl EnvironmentServiceMock) getClusterConfig(cluster *cluster.ClusterBean) (*util.ClusterConfig, error) {
	panic("implement me")
}

func (impl EnvironmentServiceMock) Update(mappings *cluster.EnvironmentBean, userId int32) (*cluster.EnvironmentBean, error) {
	panic("implement me")
}

func (impl EnvironmentServiceMock) FindClusterByEnvId(id int) (*cluster.ClusterBean, error) {
	panic("implement me")
}

func (impl EnvironmentServiceMock) GetEnvironmentListForAutocomplete() ([]cluster.EnvironmentBean, error) {
	panic("implement me")
}

//--------------
type PipelineBuilderMock struct{}

func (impl PipelineBuilderMock) CreateCiPipeline(createRequest *bean.CiConfigRequest) (*bean.PipelineCreateResponse, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) DeleteApp(appId int, userId int32) error {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) CreateCdPipelines(cdPipelines *bean.CdPipelines, ctx context.Context) (*bean.CdPipelines, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetApp(appId int) (application *bean.CreateAppDTO, err error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) PatchCdPipelines(cdPipelines *bean.CDPatchRequest, ctx context.Context) (*bean.CdPipelines, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetArtifactsByCDPipeline(cdPipelineId int, stage bean2.CdWorkflowType) (bean.CiArtifactResponse, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) FetchArtifactForRollback(cdPipelineId int) (bean.CiArtifactResponse, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) FindAppsByTeamId(teamId int) ([]pipeline.AppBean, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) FindAppsByTeamName(teamName string) ([]pipeline.AppBean, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) FindPipelineById(cdPipelineId int) (*pipelineConfig.Pipeline, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetAppListForAutocomplete() ([]pipeline.AppBean, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetCiPipelineMin(appId int) ([]*bean.CiPipeline, error) {
	panic("implement me")
}

func (impl PipelineBuilderMock) FetchCDPipelineStrategy(appId int) (pipeline.PipelineStrategiesResponse, error) {
	panic("implement me")
}
func (impl PipelineBuilderMock) GetCdPipelineById(pipelineId int) (cdPipeline *bean.CDPipelineConfigObject, err error) {
	panic("implement me")
}

func (impl PipelineBuilderMock) FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (pipeline.ConfigMapSecretsResponse, error) {
	panic("implement me")
}

//--------------
type AppWorkflowRepositoryMock struct{}

func (impl AppWorkflowRepositoryMock) SaveAppWorkflow(wf *appWorkflow.AppWorkflow) (*appWorkflow.AppWorkflow, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) UpdateAppWorkflow(wf *appWorkflow.AppWorkflow) (*appWorkflow.AppWorkflow, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindByIdAndAppId(id int, appId int) (*appWorkflow.AppWorkflow, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindByAppId(appId int) (appWorkflow []*appWorkflow.AppWorkflow, err error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) DeleteAppWorkflow(appWorkflow *appWorkflow.AppWorkflow) error {
	panic("implement me")
}

func (impl AppWorkflowRepositoryMock) SaveAppWorkflowMapping(wf *appWorkflow.AppWorkflowMapping) (*appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindByWorkflowId(workflowId int) ([]appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}

func (impl AppWorkflowRepositoryMock) FindByComponent(id int, componentType string) ([]appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}

func (impl AppWorkflowRepositoryMock) FindByNameAndAppId(name string, appId int) (*appWorkflow.AppWorkflow, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindWFCIMappingByWorkflowId(workflowId int) ([]appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindWFAllMappingByWorkflowId(workflowId int) ([]appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindWFCIMappingByCIPipelineId(cdPipelineId int) ([]appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindWFCDMappingByCIPipelineId(cdPipelineId int) ([]appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindWFCDMappingByCDPipelineId(cdPipelineId int) ([]appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) DeleteAppWorkflowMapping(appWorkflow *appWorkflow.AppWorkflowMapping) error {
	panic("implement me")
}
func (impl AppWorkflowRepositoryMock) FindWFCDMappingByCIPipelineIds(ciPipelineIds []int) ([]*appWorkflow.AppWorkflowMapping, error) {
	panic("implement me")
}

//--------
type CiPipelineRepositoryMock struct{}

func (impl CiPipelineRepositoryMock) Save(pipeline *pipelineConfig.CiPipeline) error {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) SaveExternalCi(pipeline *pipelineConfig.ExternalCiPipeline) (*pipelineConfig.ExternalCiPipeline, error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) UpdateExternalCi(pipeline *pipelineConfig.ExternalCiPipeline) (*pipelineConfig.ExternalCiPipeline, int, error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FindExternalCiByCiPipelineId(ciPipelineId int) (*pipelineConfig.ExternalCiPipeline, error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FindCiScriptsByCiPipelineId(ciPipelineId int) ([]*pipelineConfig.CiPipelineScript, error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) SaveCiPipelineScript(ciPipelineScript *pipelineConfig.CiPipelineScript) error {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) UpdateCiPipelineScript(script *pipelineConfig.CiPipelineScript) error {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FindByAppId(appId int) (pipelines []*pipelineConfig.CiPipeline, err error) {
	panic("implement me")
}

//find non deleted pipeline
func (impl CiPipelineRepositoryMock) FindById(id int) (pipeline *pipelineConfig.CiPipeline, err error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FindByCiAndAppDetailsById(pipelineId int) (pipeline *pipelineConfig.CiPipeline, err error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FindByIdsIn(ids []int) ([]*pipelineConfig.CiPipeline, error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) Update(pipeline *pipelineConfig.CiPipeline) error {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) PipelineExistsByName(names []string) (found []string, err error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FindByName(pipelineName string) (pipeline *pipelineConfig.CiPipeline, err error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FindByParentCiPipelineId(parentCiPipelineId int) ([]*pipelineConfig.CiPipeline, error) {
	panic("implement me")
}

func (impl CiPipelineRepositoryMock) FetchParentCiPipelinesForDG() ([]*pipelineConfig.CiPipelinesMap, error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FetchCiPipelinesForDG(parentId int, childCiPipelineIds []int) (*pipelineConfig.CiPipeline, int, error) {
	panic("implement me")
}
func (impl CiPipelineRepositoryMock) FinDByParentCiPipelineAndAppId(parentCiPipeline int, appIds []int) ([]*pipelineConfig.CiPipeline, error) {
	panic("implement me")
}

//------
type PipelineRepositoryMock struct{}

func (impl PipelineRepositoryMock) Save(pipeline []*pipelineConfig.Pipeline) error {
	panic("implement me")
}
func (impl PipelineRepositoryMock) Update(pipeline *pipelineConfig.Pipeline) error {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindActiveByAppId(appId int) (pipelines []*pipelineConfig.Pipeline, err error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) Delete(id int) error {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindByName(pipelineName string) (pipeline *pipelineConfig.Pipeline, err error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) PipelineExists(pipelineName string) (bool, error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindById(id int) (pipeline *pipelineConfig.Pipeline, err error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindByIdsIn(ids []int) ([]*pipelineConfig.Pipeline, error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindByCiPipelineIdsIn(ciPipelineIds []int) ([]*pipelineConfig.Pipeline, error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindAutomaticByCiPipelineId(ciPipelineId int) (pipelines []*pipelineConfig.Pipeline, err error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) GetByEnvOverrideId(envOverrideId int) ([]pipelineConfig.Pipeline, error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindActiveByAppIdAndEnvironmentId(appId int, environmentId int) (pipelines []*pipelineConfig.Pipeline, err error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) UndoDelete(id int) error {
	panic("implement me")
}
func (impl PipelineRepositoryMock) UniqueAppEnvironmentPipelines() ([]*pipelineConfig.Pipeline, error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindByCiPipelineId(ciPipelineId int) (pipelines []*pipelineConfig.Pipeline, err error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindByPipelineTriggerGitHash(gitHash string) (pipeline *pipelineConfig.Pipeline, err error) {
	panic("implement me")
}
func (impl PipelineRepositoryMock) FindByIdsInAndEnvironment(ids []int, environmentId int) ([]*pipelineConfig.Pipeline, error) {
	panic("implement me")
}

//--------
type MaterialRepositoryMock struct{}

func (impl MaterialRepositoryMock) MaterialExists(url string) (bool, error) {
	panic("implement me")
}
func (impl MaterialRepositoryMock) SaveMaterial(material *pipelineConfig.GitMaterial) error {
	panic("implement me")
}
func (impl MaterialRepositoryMock) UpdateMaterial(material *pipelineConfig.GitMaterial) error {
	panic("implement me")
}
func (impl MaterialRepositoryMock) Update(materials []*pipelineConfig.GitMaterial) error {
	panic("implement me")
}
func (impl MaterialRepositoryMock) FindByAppId(appId int) ([]*pipelineConfig.GitMaterial, error) {
	panic("implement me")
}
func (impl MaterialRepositoryMock) FindById(Id int) (*pipelineConfig.GitMaterial, error) {
	panic("implement me")
}
func (impl MaterialRepositoryMock) UpdateMaterialScmId(material *pipelineConfig.GitMaterial) error {
	panic("implement me")
}
