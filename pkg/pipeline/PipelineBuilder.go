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
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/bean"
	"encoding/json"
	"fmt"
	application2 "github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type EcrConfig struct {
	EcrPrefix string `env:"ECR_REPO_NAME_PREFIX" envDefault:"test/"`
}

func GetEcrConfig() (*EcrConfig, error) {
	cfg := &EcrConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type PipelineBuilder interface {
	CreateCiPipeline(createRequest *bean.CiConfigRequest) (*bean.PipelineCreateResponse, error)
	CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error)
	CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error)
	UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error)
	DeleteApp(appId int, userId int32) error
	GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error)
	UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error)
	PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error)
	CreateCdPipelines(cdPipelines *bean.CdPipelines, ctx context.Context) (*bean.CdPipelines, error)
	GetApp(appId int) (application *bean.CreateAppDTO, err error)
	PatchCdPipelines(cdPipelines *bean.CDPatchRequest, ctx context.Context) (*bean.CdPipelines, error)
	GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error)
	GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error)
	/*	CreateCdPipelines(cdPipelines bean.CdPipelines) (*bean.CdPipelines, error)*/
	GetArtifactsByCDPipeline(cdPipelineId int, stage bean2.CdWorkflowType) (bean.CiArtifactResponse, error)
	FetchArtifactForRollback(cdPipelineId int) (bean.CiArtifactResponse, error)
	FindAppsByTeamId(teamId int) ([]AppBean, error)
	GetAppListByTeamIds(teamIds []int) ([]*TeamAppBean, error)
	FindAppsByTeamName(teamName string) ([]AppBean, error)
	FindPipelineById(cdPipelineId int) (*pipelineConfig.Pipeline, error)
	GetAppList() ([]AppBean, error)
	GetCiPipelineMin(appId int) ([]*bean.CiPipelineMin, error)

	FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error)
	GetCdPipelineById(pipelineId int) (cdPipeline *bean.CDPipelineConfigObject, err error)

	FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error)
	FindByIds(ids []*int) ([]*AppBean, error)
	GetCiPipelineById(pipelineId int) (ciPipeline *bean.CiPipeline, err error)
}

type PipelineBuilderImpl struct {
	logger                        *zap.SugaredLogger
	dbPipelineOrchestrator        DbPipelineOrchestrator
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository
	materialRepo                  pipelineConfig.MaterialRepository
	appRepo                       pipelineConfig.AppRepository
	pipelineRepository            pipelineConfig.PipelineRepository
	propertiesConfigService       PropertiesConfigService
	ciTemplateRepository          pipelineConfig.CiTemplateRepository
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	application                   application.ServiceClient
	chartRepository               chartConfig.ChartRepository
	ciArtifactRepository          repository.CiArtifactRepository
	ecrConfig                     *EcrConfig
	envConfigOverrideRepository   chartConfig.EnvConfigOverrideRepository
	environmentRepository         cluster.EnvironmentRepository
	pipelineConfigRepository      chartConfig.PipelineConfigRepository
	mergeUtil                     util.MergeUtil
	appWorkflowRepository         appWorkflow.AppWorkflowRepository
	ciConfig                      *CiConfig
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	appService                    app.AppService
	imageScanResultRepository     security.ImageScanResultRepository
}

func NewPipelineBuilderImpl(logger *zap.SugaredLogger,
	dbPipelineOrchestrator DbPipelineOrchestrator,
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository,
	materialRepo pipelineConfig.MaterialRepository,
	pipelineGroupRepo pipelineConfig.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	propertiesConfigService PropertiesConfigService,
	ciTemplateRepository pipelineConfig.CiTemplateRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	application application.ServiceClient,
	chartRepository chartConfig.ChartRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	ecrConfig *EcrConfig,
	envConfigOverrideRepository chartConfig.EnvConfigOverrideRepository,
	environmentRepository cluster.EnvironmentRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	mergeUtil util.MergeUtil,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	ciConfig *CiConfig,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	appService app.AppService,
	imageScanResultRepository security.ImageScanResultRepository,
) *PipelineBuilderImpl {
	return &PipelineBuilderImpl{
		logger:                        logger,
		dbPipelineOrchestrator:        dbPipelineOrchestrator,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
		materialRepo:                  materialRepo,
		appService:                    appService,
		appRepo:                       pipelineGroupRepo,
		pipelineRepository:            pipelineRepository,
		propertiesConfigService:       propertiesConfigService,
		ciTemplateRepository:          ciTemplateRepository,
		ciPipelineRepository:          ciPipelineRepository,
		application:                   application,
		chartRepository:               chartRepository,
		ciArtifactRepository:          ciArtifactRepository,
		ecrConfig:                     ecrConfig,
		envConfigOverrideRepository:   envConfigOverrideRepository,
		environmentRepository:         environmentRepository,
		pipelineConfigRepository:      pipelineConfigRepository,
		mergeUtil:                     mergeUtil,
		appWorkflowRepository:         appWorkflowRepository,
		ciConfig:                      ciConfig,
		cdWorkflowRepository:          cdWorkflowRepository,
		imageScanResultRepository:     imageScanResultRepository,
	}
}

func formatDate(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

func (impl PipelineBuilderImpl) CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	impl.logger.Debugw("app create request received", "req", request)
	res, err := impl.dbPipelineOrchestrator.CreateApp(request)
	if err != nil {
		impl.logger.Errorw("error in saving create app req", "req", request, "err", err)
	}
	return res, err
}

func (impl PipelineBuilderImpl) DeleteApp(appId int, userId int32) error {
	impl.logger.Debugw("app delete request received", "app", appId)
	err := impl.dbPipelineOrchestrator.DeleteApp(appId, userId)
	return err
}

func (impl PipelineBuilderImpl) CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error) {
	res, err := impl.dbPipelineOrchestrator.CreateMaterials(request)
	if err != nil {
		impl.logger.Errorw("error in saving create materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl PipelineBuilderImpl) UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error) {
	res, err := impl.dbPipelineOrchestrator.UpdateMaterial(request)
	if err != nil {
		impl.logger.Errorw("error in updating materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl PipelineBuilderImpl) GetApp(appId int) (application *bean.CreateAppDTO, err error) {
	app, err := impl.appRepo.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app", "id", appId, "err", err)
		return nil, err
	}
	materials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching materials", "appId", appId, "err", err)
	}
	var gitMaterials []*bean.GitMaterial
	for _, material := range materials {
		gitMaterial := &bean.GitMaterial{
			Url:           material.Url,
			Name:          material.Name[strings.Index(material.Name, "-")+1:],
			Id:            material.Id,
			GitProviderId: material.GitProviderId,
			CheckoutPath:  material.CheckoutPath,
		}
		gitMaterials = append(gitMaterials, gitMaterial)
	}
	application = &bean.CreateAppDTO{
		Id:       app.Id,
		AppName:  app.AppName,
		Material: gitMaterials,
		TeamId:   app.TeamId,
	}
	return application, nil
}

/*
1. create pipelineGroup
2. save material (add credential provider support)

*/

func (impl PipelineBuilderImpl) getDefaultArtifactStore(id string) (store *repository.DockerArtifactStore, err error) {
	if id == "" {
		impl.logger.Debugw("docker repo is empty adding default repo")
		store, err = impl.dockerArtifactStoreRepository.FindActiveDefaultStore()

	} else {
		store, err = impl.dockerArtifactStoreRepository.FindOne(id)
	}
	return
}

func (impl PipelineBuilderImpl) getCiTemplateVariables(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	template, err := impl.ciTemplateRepository.FindByAppId(appId)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	if errors.IsNotFound(err) {
		impl.logger.Debugw(" no ci pipeline exists", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline exists"}
		return nil, err
	}

	gitMaterials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching git materials", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Debugw(" no git materials exists", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no git materials exists"}
		return nil, err
	}

	var materials []bean.Material
	for _, g := range gitMaterials {
		m := bean.Material{
			GitMaterialId: g.Id,
			MaterialName:  g.Name[strings.Index(g.Name, "-")+1:],
		}
		materials = append(materials, m)
	}

	dockerArgs := map[string]string{}
	if err := json.Unmarshal([]byte(template.Args), &dockerArgs); err != nil {
		impl.logger.Debugw("error in json unmarshal", "app", appId, "err", err)
		return nil, err
	}
	var beforeDockerBuild []*bean.Task
	var afterDockerBuild []*bean.Task
	if err := json.Unmarshal([]byte(template.BeforeDockerBuild), &beforeDockerBuild); err != nil {
		impl.logger.Debugw("error in BeforeDockerBuild json unmarshal", "app", appId, "err", err)
		return nil, err
	}
	if err := json.Unmarshal([]byte(template.AfterDockerBuild), &afterDockerBuild); err != nil {
		impl.logger.Debugw("error in AfterDockerBuild json unmarshal", "app", appId, "err", err)
		return nil, err
	}
	ciConfig = &bean.CiConfigRequest{
		Id:                template.Id,
		AppId:             template.AppId,
		AppName:           template.App.AppName,
		DockerRepository:  template.DockerRepository,
		DockerRegistry:    template.DockerRegistry.Id,
		DockerRegistryUrl: template.DockerRegistry.GetRegistryLocation(),
		BeforeDockerBuild: beforeDockerBuild,
		AfterDockerBuild:  afterDockerBuild,
		DockerBuildConfig: &bean.DockerBuildConfig{DockerfilePath: template.DockerfilePath, Args: dockerArgs, GitMaterialId: template.GitMaterialId},
		Version:           template.Version,
		CiTemplateName:    template.TemplateName,
		Materials:         materials,
	}
	return ciConfig, err
}
func (impl PipelineBuilderImpl) GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	ciConfig, err = impl.getCiTemplateVariables(appId)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	//TODO fill these variables
	//ciConfig.CiPipeline=
	//--------pipeline population start
	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	var ciPipelineResp []*bean.CiPipeline

	for _, pipeline := range pipelines {

		dockerArgs := make(map[string]string)
		if len(pipeline.DockerArgs) > 0 {
			err := json.Unmarshal([]byte(pipeline.DockerArgs), &dockerArgs)
			if err != nil {
				impl.logger.Warnw("error in unmarshal", "err", err)
			}
		}

		var externalCiConfig bean.ExternalCiConfig
		if pipeline.ExternalCiPipeline != nil {
			externalCiConfig = bean.ExternalCiConfig{
				Id:         pipeline.ExternalCiPipeline.Id,
				AccessKey:  pipeline.ExternalCiPipeline.AccessToken,
				WebhookUrl: impl.ciConfig.ExternalCiWebhookUrl,
				Payload:    impl.ciConfig.ExternalCiPayload,
			}
		}

		ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(pipeline.Id)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching ci scripts")
			return nil, err
		}

		var beforeDockerBuildScripts []*bean.CiScript
		var afterDockerBuildScripts []*bean.CiScript
		for _, ciScript := range ciPipelineScripts {
			ciScriptResp := &bean.CiScript{
				Id:             ciScript.Id,
				Index:          ciScript.Index,
				Name:           ciScript.Name,
				Script:         ciScript.Script,
				OutputLocation: ciScript.OutputLocation,
			}
			if ciScript.Stage == BEFORE_DOCKER_BUILD {
				beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciScriptResp)
			} else if ciScript.Stage == AFTER_DOCKER_BUILD {
				afterDockerBuildScripts = append(afterDockerBuildScripts, ciScriptResp)
			}
		}
		parentCiPipeline, err := impl.ciPipelineRepository.FindById(pipeline.ParentCiPipeline)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return nil, err
		}
		ciPipeline := &bean.CiPipeline{
			Id:                       pipeline.Id,
			Version:                  pipeline.Version,
			Name:                     pipeline.Name,
			Active:                   pipeline.Active,
			Deleted:                  pipeline.Deleted,
			DockerArgs:               dockerArgs,
			IsManual:                 pipeline.IsManual,
			IsExternal:               pipeline.IsExternal,
			ParentCiPipeline:         pipeline.ParentCiPipeline,
			ParentAppId:              parentCiPipeline.AppId,
			ExternalCiConfig:         externalCiConfig,
			BeforeDockerBuildScripts: beforeDockerBuildScripts,
			AfterDockerBuildScripts:  afterDockerBuildScripts,
			ScanEnabled:              pipeline.ScanEnabled,
		}
		for _, material := range pipeline.CiPipelineMaterials {
			ciMaterial := &bean.CiMaterial{
				Id:              material.Id,
				CheckoutPath:    material.CheckoutPath,
				Path:            material.Path,
				ScmId:           material.ScmId,
				GitMaterialId:   material.GitMaterialId,
				GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
				ScmName:         material.ScmName,
				ScmVersion:      material.ScmVersion,
				Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value},
			}
			ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
		}
		linkedCis, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipeline.Id)
		if err != nil && !util.IsErrNoRows(err) {
			return nil, err
		}
		ciPipeline.LinkedCount = len(linkedCis)
		ciPipelineResp = append(ciPipelineResp, ciPipeline)
	}
	ciConfig.CiPipelines = ciPipelineResp
	//--------pipeline population end
	return ciConfig, err
}

func (impl PipelineBuilderImpl) GetCiPipelineMin(appId int) ([]*bean.CiPipelineMin, error) {
	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows || len(pipelines) == 0 {
		impl.logger.Errorw("no ci pipeline found", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline found"}
		return nil, err
	}
	var ciPipelineResp []*bean.CiPipelineMin
	for _, pipeline := range pipelines {
		parentCiPipeline, err := impl.ciPipelineRepository.FindById(pipeline.ParentCiPipeline)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return nil, err
		}

		pipelineType := bean.PipelineType(bean.NORMAL)
		if parentCiPipeline.Id > 0 {
			pipelineType = bean.PipelineType(bean.LINKED)
		} else if pipeline.IsExternal == true {
			pipelineType = bean.PipelineType(bean.EXTERNAL)
		}

		ciPipeline := &bean.CiPipelineMin{
			Id:               pipeline.Id,
			Name:             pipeline.Name,
			ParentCiPipeline: pipeline.ParentCiPipeline,
			ParentAppId:      parentCiPipeline.AppId,
			PipelineType:     pipelineType,
			ScanEnabled:      pipeline.ScanEnabled,
		}
		ciPipelineResp = append(ciPipelineResp, ciPipeline)
	}
	return ciPipelineResp, err
}

func (impl PipelineBuilderImpl) UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {
	originalCiConf, err := impl.getCiTemplateVariables(updateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching original ciConfig for update", "appId", updateRequest.Id, "err", err)
		return nil, err
	}
	if originalCiConf.Version != updateRequest.Version {
		impl.logger.Errorw("stale version requested", "appId", updateRequest.Id, "old", originalCiConf.Version, "new", updateRequest.Version)
		return nil, fmt.Errorf("stale version of resource requested kindly refresh. requested: %s, found %s", updateRequest.Version, originalCiConf.Version)
	}
	dockerArtifaceStore, err := impl.dockerArtifactStoreRepository.FindOne(updateRequest.DockerRegistry)
	if err != nil {
		impl.logger.Errorw("error in fetching DockerRegistry  for update", "appId", updateRequest.Id, "err", err, "registry", updateRequest.DockerRegistry)
		return nil, err
	}

	originalCiConf.AfterDockerBuild = updateRequest.AfterDockerBuild
	originalCiConf.BeforeDockerBuild = updateRequest.BeforeDockerBuild
	originalCiConf.DockerBuildConfig = updateRequest.DockerBuildConfig
	originalCiConf.DockerRegistry = updateRequest.DockerRegistry
	originalCiConf.DockerRepository = updateRequest.DockerRepository
	originalCiConf.DockerRegistryUrl = dockerArtifaceStore.GetRegistryLocation()

	argByte, err := json.Marshal(originalCiConf.DockerBuildConfig.Args)
	if err != nil {
		return nil, err
	}
	afterByte, err := json.Marshal(originalCiConf.AfterDockerBuild)
	if err != nil {
		return nil, err
	}
	beforeByte, err := json.Marshal(originalCiConf.BeforeDockerBuild)
	if err != nil {
		return nil, err
	}
	ciTemplate := &pipelineConfig.CiTemplate{
		DockerfilePath:    originalCiConf.DockerBuildConfig.DockerfilePath,
		GitMaterialId:     originalCiConf.DockerBuildConfig.GitMaterialId,
		Args:              string(argByte),
		BeforeDockerBuild: string(beforeByte),
		AfterDockerBuild:  string(afterByte),
		Version:           originalCiConf.Version,
		Id:                originalCiConf.Id,
		DockerRepository:  originalCiConf.DockerRepository,
		DockerRegistryId:  originalCiConf.DockerRegistry,
	}

	err = impl.ciTemplateRepository.Update(ciTemplate)
	if err != nil {
		impl.logger.Errorw("error in updating ci template in db", "template", ciTemplate, "err", err)
		return nil, err
	}
	return originalCiConf, nil
}

func (impl PipelineBuilderImpl) CreateCiPipeline(createRequest *bean.CiConfigRequest) (*bean.PipelineCreateResponse, error) {
	impl.logger.Debugw("pipeline create request received", "req", createRequest)

	//-----------fetch data
	app, err := impl.appRepo.FindById(createRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline group", "groupId", createRequest.AppId, "err", err)
		return nil, err
	}
	//--ecr config
	createRequest.AppName = app.AppName
	store, err := impl.getDefaultArtifactStore(createRequest.DockerRegistry)
	if err != nil {
		impl.logger.Errorw("error in fetching docker store ", "id", createRequest.DockerRepository, "err", err)
		return nil, err
	}
	createRequest.DockerRegistryUrl = store.GetRegistryLocation()
	createRequest.DockerRegistry = store.Id

	if createRequest.DockerRepository == "" {
		repo := impl.ecrConfig.EcrPrefix + app.AppName
		impl.logger.Debugw("repo is empty creating ecr repo ", "repo", repo)
		err := util.CreateEcrRepo(repo, createRequest.DockerRepository, store.AWSRegion, store.AWSAccessKeyId, store.AWSSecretAccessKey)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				impl.logger.Warnw("repo already exists , skipping", "repo", repo)
			} else {
				impl.logger.Errorw("error in creating repo", "repo", repo, "err", err)
				return nil, err
			}
		}
		createRequest.DockerRepository = repo
	}
	//--ecr config	end
	//-- template config start

	argByte, err := json.Marshal(createRequest.DockerBuildConfig.Args)
	if err != nil {
		return nil, err
	}
	afterByte, err := json.Marshal(createRequest.AfterDockerBuild)
	if err != nil {
		return nil, err
	}
	beforeByte, err := json.Marshal(createRequest.BeforeDockerBuild)
	if err != nil {
		return nil, err
	}

	ciTemplate := &pipelineConfig.CiTemplate{
		DockerRegistryId:  createRequest.DockerRegistry,
		DockerRepository:  createRequest.DockerRepository,
		GitMaterialId:     createRequest.DockerBuildConfig.GitMaterialId,
		DockerfilePath:    createRequest.DockerBuildConfig.DockerfilePath,
		Args:              string(argByte),
		Active:            true,
		TemplateName:      createRequest.CiTemplateName,
		Version:           createRequest.Version,
		AppId:             createRequest.AppId,
		AfterDockerBuild:  string(afterByte),
		BeforeDockerBuild: string(beforeByte),
		AuditLog:          models.AuditLog{CreatedOn: time.Now(), UpdatedOn: time.Now(), CreatedBy: createRequest.UserId, UpdatedBy: createRequest.UserId},
	}

	err = impl.ciTemplateRepository.Save(ciTemplate)
	if err != nil {
		impl.logger.Errorw("error in saving ci template in db ", "template", ciTemplate, "err", err)
		//TODO delete template from gocd otherwise dangling+ no create in future
		return nil, err
	}
	//-- template config end
	createRequest.Id = ciTemplate.Id
	createRequest.CiTemplateName = ciTemplate.TemplateName
	if len(createRequest.CiPipelines) > 0 {
		conf, err := impl.addpipelineToTemplate(createRequest)
		if err != nil {
			impl.logger.Errorw("error in pipeline creation ", "err", err)
			return nil, err
		}
		impl.logger.Debugw("pipeline created ", "detail", conf)
	}
	createRes := &bean.PipelineCreateResponse{AppName: app.AppName, AppId: createRequest.AppId} //FIXME
	return createRes, nil
}

func (impl PipelineBuilderImpl) getGitMaterialsForApp(appId int) ([]*bean.GitMaterial, error) {
	materials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching materials for app", "appId", appId, "err", err)
		return nil, err
	}
	var gitMaterials []*bean.GitMaterial

	for _, material := range materials {
		gitUrl := material.Url
		if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD ||
			material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
			u, err := url.Parse(gitUrl)
			if err != nil {
				return nil, err
			}
			var password string
			userName := material.GitProvider.UserName
			if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
				password = material.GitProvider.Password

			} else if material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
				password = material.GitProvider.AccessToken
				if len(userName) == 0 {
					userName = "devtron-boat"
				}
			}
			if len(userName) == 0 || len(password) == 0 {
				return nil, util.ApiError{}.ErrorfUser("invalid git credentials config")
			}
			u.User = url.UserPassword(userName, password)
			gitUrl = u.String()
		}
		gitMaterial := &bean.GitMaterial{
			Id:            material.Id,
			Url:           gitUrl,
			GitProviderId: material.GitProviderId,
			Name:          material.Name[strings.Index(material.Name, "-")+1:],
			CheckoutPath:  material.CheckoutPath,
		}
		gitMaterials = append(gitMaterials, gitMaterial)
	}
	return gitMaterials, nil
}

func (impl PipelineBuilderImpl) addpipelineToTemplate(createRequest *bean.CiConfigRequest) (resp *bean.CiConfigRequest, err error) {

	//single ci in same wf validation
	workflowMapping, err := impl.appWorkflowRepository.FindWFCIMappingByWorkflowId(createRequest.AppWorkflowId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching workflow mapping for ci validation", "err", err)
		return nil, err
	}
	if len(workflowMapping) > 0 {
		return nil, &util.ApiError{
			InternalMessage:   "pipeline already exists",
			UserDetailMessage: fmt.Sprintf("pipeline already exists in workflow"),
			UserMessage:       fmt.Sprintf("pipeline already exists in workflow")}
	}

	//pipeline name validation
	var pipelineNames []string
	for _, pipeline := range createRequest.CiPipelines {
		pipelineNames = append(pipelineNames, pipeline.Name)
	}
	/*
	if pipelineNames != nil && len(pipelineNames) > 0 {
		found, err := impl.ciPipelineRepository.PipelineExistsByName(pipelineNames)
		if err != nil {
			impl.logger.Errorw("err in duplicate check for ci pipeline", "app", createRequest.AppName, "names", pipelineNames, "err", err)
			return nil, err
		} else if found != nil && len(found) > 0 {
			impl.logger.Warnw("duplicate pipelins ", "app", createRequest.AppName, "duplicates", found)
			//return nil,  errors.AlreadyExistsf("pipelines exists %s", found)
			return nil, &util.ApiError{
				HttpStatusCode:    409,
				Code:              "0409",
				InternalMessage:   "pipeline already exists",
				UserDetailMessage: fmt.Sprintf("pipeline already exists %s", found),
				UserMessage:       fmt.Sprintf("pipeline already exists %s", found)}
		}
	}
	*/
	if err != nil {
		impl.logger.Errorw("error in creating pipeline group", "err", err)
		return nil, err
	}
	createRequest, err = impl.dbPipelineOrchestrator.CreateCiConf(createRequest, createRequest.Id)
	if err != nil {
		return nil, err
	}
	return createRequest, err
}

func (impl PipelineBuilderImpl) PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error) {
	ciConfig, err = impl.getCiTemplateVariables(request.AppId)
	if err != nil {
		impl.logger.Errorw("err in fetching template for pipeline patch, ", "err", err, "appId", request.AppId)
		return nil, err
	}
	ciConfig.AppWorkflowId = request.AppWorkflowId
	ciConfig.UserId = request.UserId
	if request.CiPipeline != nil {
		ciConfig.ScanEnabled = request.CiPipeline.ScanEnabled
	}
	switch request.Action {
	case bean.CREATE:
		impl.logger.Debugw("create patch request")
		ciConfig.CiPipelines = []*bean.CiPipeline{request.CiPipeline} //request.CiPipeline
		res, err := impl.addpipelineToTemplate(ciConfig)
		if err != nil {
			impl.logger.Errorw("error in adding pipeline to template", "ciConf", ciConfig, "err", err)
			return nil, err
		}
		return res, nil
	case bean.UPDATE_SOURCE:
		return impl.patchCiPipelineUpdateSource(ciConfig, request.CiPipeline)
	case bean.DELETE:
		pipeline, err := impl.deletePipeline(request)
		if err != nil {
			return nil, err
		}
		ciConfig.CiPipelines = []*bean.CiPipeline{pipeline}
		return ciConfig, nil
	default:
		impl.logger.Errorw("unsupported operation ", "op", request.Action)
		return nil, fmt.Errorf("unsupported operation %s", request.Action)
	}

}

func (impl PipelineBuilderImpl) deletePipeline(request *bean.CiPatchRequest) (*bean.CiPipeline, error) {

	//wf validation
	workflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCIPipelineId(request.CiPipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching workflow mapping for ci validation", "err", err)
		return nil, err
	}
	if len(workflowMapping) > 0 {
		return nil, &util.ApiError{
			InternalMessage:   "cd pipeline exists for this CI",
			UserDetailMessage: fmt.Sprintf("cd pipeline exists for this CI"),
			UserMessage:       fmt.Sprintf("cd pipeline exists for this CI")}
	}

	pipeline, err := impl.ciPipelineRepository.FindById(request.CiPipeline.Id)
	if err != nil {
		impl.logger.Errorw("pipeline fetch err", "id", request.CiPipeline.Id, "err", err)
	}
	if pipeline.AppId != request.AppId {
		return nil, fmt.Errorf("invalid appid: %d pipelineId: %d mapping", request.AppId, request.CiPipeline.Id)
	}

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.dbPipelineOrchestrator.DeleteCiPipeline(pipeline, request.UserId, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting pipeline db")
		return nil, err
	}

	//delete app workflow mapping
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCIMappingByCIPipelineId(pipeline.Id)
	for _, mapping := range appWorkflowMappings {
		err := impl.appWorkflowRepository.DeleteAppWorkflowMapping(mapping, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting workflow mapping", "err", err)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	request.CiPipeline.Deleted = true
	request.CiPipeline.Name = pipeline.Name
	return request.CiPipeline, nil
	//delete pipeline
	//delete scm

}

func (impl PipelineBuilderImpl) patchCiPipelineUpdateSource(baseCiConfig *bean.CiConfigRequest, modifiedCiPipeline *bean.CiPipeline) (ciConfig *bean.CiConfigRequest, err error) {

	pipeline, err := impl.ciPipelineRepository.FindById(modifiedCiPipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline", "id", modifiedCiPipeline.Id, "err", err)
		return nil, err
	}

	cannotUpdate := false
	for _, material := range pipeline.CiPipelineMaterials {
		if material.ScmId != "" {
			cannotUpdate = true
		}
	}

	if cannotUpdate {
		//scm plugin material change scm object
		//material.ScmName
		return nil, fmt.Errorf("update of plugin scm material not supported")
	} else {
		modifiedCiPipeline.ScanEnabled = baseCiConfig.ScanEnabled
		modifiedCiPipeline, err = impl.dbPipelineOrchestrator.PatchMaterialValue(modifiedCiPipeline, baseCiConfig.UserId)
		if err != nil {
			return nil, err
		}
		baseCiConfig.CiPipelines = append(baseCiConfig.CiPipelines, modifiedCiPipeline)
		return baseCiConfig, err
	}

}

func (impl PipelineBuilderImpl) CreateCdPipelines(cdPipelines *bean.CdPipelines, ctx context.Context) (*bean.CdPipelines, error) {
	app, err := impl.appRepo.FindById(cdPipelines.AppId)
	if err != nil {
		impl.logger.Errorw("app not found", "err", err, "appId", cdPipelines.AppId)
		return nil, err
	}

	envPipelineMap := make(map[int]string)
	for _, pipeline := range cdPipelines.Pipelines {
		if envPipelineMap[pipeline.EnvironmentId] != "" {
			err = &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
				UserMessage:     "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
			}
			return nil, err
		}
		envPipelineMap[pipeline.EnvironmentId] = pipeline.Name

		existingCdPipelinesForEnv, pErr := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(cdPipelines.AppId, pipeline.EnvironmentId)
		if pErr != nil && !util.IsErrNoRows(pErr) {
			impl.logger.Errorw("error in fetching cd pipelines ", "err", pErr, "appId", cdPipelines.AppId)
			return nil, pErr
		}
		if len(existingCdPipelinesForEnv) > 0 {
			err = &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
				UserMessage:     "cd-pipelines already exist for this app and env, cannot create multiple cd-pipelines",
			}
			return nil, err
		}

		if len(pipeline.PreStage.Config) > 0 && !strings.Contains(pipeline.PreStage.Config, "beforeStages") {
			err = &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "invalid yaml config, must include - beforeStages",
				UserMessage:     "invalid yaml config, must include - beforeStages",
			}
			return nil, err
		}
		if len(pipeline.PostStage.Config) > 0 && !strings.Contains(pipeline.PostStage.Config, "afterStages") {
			err = &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "invalid yaml config, must include - afterStages",
				UserMessage:     "invalid yaml config, must include - afterStages",
			}
			return nil, err
		}
	}

	for _, pipeline := range cdPipelines.Pipelines {
		id, err := impl.createCdPipeline(app, pipeline, cdPipelines.UserId, ctx)
		if err != nil {
			impl.logger.Errorw("error in creating pipeline", "name", pipeline.Name, "err", err)
			return nil, err
		}
		pipeline.Id = id
	}

	return cdPipelines, nil
}

func (impl PipelineBuilderImpl) PatchCdPipelines(cdPipelines *bean.CDPatchRequest, ctx context.Context) (*bean.CdPipelines, error) {
	pipelineRequest := &bean.CdPipelines{
		UserId:    cdPipelines.UserId,
		AppId:     cdPipelines.AppId,
		Pipelines: []*bean.CDPipelineConfigObject{cdPipelines.Pipeline},
	}
	switch cdPipelines.Action {
	case bean.CD_CREATE:
		return impl.CreateCdPipelines(pipelineRequest, ctx)
	case bean.CD_UPDATE:
		err := impl.updateCdPipeline(cdPipelines.Pipeline, cdPipelines.UserId, ctx)
		return pipelineRequest, err
	case bean.CD_DELETE:
		err := impl.deleteCdPipeline(cdPipelines.Pipeline.Id, cdPipelines.UserId, ctx)
		return pipelineRequest, err
	default:
		return nil, &util.ApiError{Code: "404", HttpStatusCode: 404, UserMessage: "operation not supported"}
	}
}

func (impl PipelineBuilderImpl) deleteCdPipeline(pipelineId int, userId int32, ctx context.Context) (err error) {
	pipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.logger.Errorw("err in fetching pipeline", "id", pipelineId, "err", err)
		return err
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	if err = impl.dbPipelineOrchestrator.DeleteCdPipeline(pipelineId, tx); err != nil {
		impl.logger.Errorw("err in deleting pipeline from db", "id", pipeline, "err", err)
		return err
	}

	//delete app workflow mapping
	appWorkflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCDPipelineId(pipelineId)
	for _, mapping := range appWorkflowMapping {
		err := impl.appWorkflowRepository.DeleteAppWorkflowMapping(mapping, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting workflow mapping", "err", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && !util.IsErrNoRows(err) {
		if err1 := impl.pipelineRepository.UndoDelete(pipelineId); err1 != nil {
			impl.logger.Errorw("unable to revert pipeline delete, this might lead inconsistency ", "pipeline", pipelineId, "err", err1)
		}
		return err
	} else if len(pipelines) == 0 {
		envModel, err := impl.environmentRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			return err
		}
		argoAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, envModel.Name)
		req := &application2.ApplicationDeleteRequest{
			Name: &argoAppName,
		}
		if _, err := impl.application.Delete(ctx, req); err != nil {
			impl.logger.Errorw("err in deleting pipeline on argocd", "id", pipeline, "err", err)
			if err1 := impl.pipelineRepository.UndoDelete(pipelineId); err1 != nil {
				impl.logger.Errorw("unable to revert pipeline delete, this might lead inconsistency ", "pipeline", pipelineId, "err", err1)
			}
			return err
		}
		impl.logger.Infow("app deleted from argocd", "id", pipelineId, "pipelineName", pipeline.Name, "app", argoAppName)
		return nil
	} else {
		impl.logger.Infow("pipelins for environment exists not deleting argo app", "pipelines", pipelines)
		return nil
	}
}

type DeploymentType struct {
	Deployment Deployment `json:"deployment"`
}

type Deployment struct {
	Strategy Strategy `json:"strategy"`
}

type Strategy struct {
	BlueGreen *BlueGreen `json:"blueGreen,omitempty"`
	Rolling   *Rolling   `json:"rolling,omitempty"`
	Canary    *Canary    `json:"canary,omitempty"`
	Recreate  *Recreate  `json:"recreate,omitempty"`
}

type BlueGreen struct {
	AutoPromotionSeconds  int  `json:"autoPromotionSeconds"`
	ScaleDownDelaySeconds int  `json:"scaleDownDelaySeconds"`
	PreviewReplicaCount   int  `json:"previewReplicaCount"`
	AutoPromotionEnabled  bool `json:"autoPromotionEnabled"`
}

type Canary struct {
	MaxSurge       string       `json:"maxSurge,omitempty"`
	MaxUnavailable int          `json:"maxUnavailable,omitempty"`
	Steps          []CanaryStep `json:"steps,omitempty"`
}

type CanaryStep struct {
	// SetWeight sets what percentage of the newRS should receive
	SetWeight *int32 `json:"setWeight,omitempty"`
	// Pause freezes the rollout by setting spec.Paused to true.
	// A Rollout will resume when spec.Paused is reset to false.
	// +optional
	Pause *RolloutPause `json:"pause,omitempty"`
}

type RolloutPause struct {
	// Duration the amount of time to wait before moving to the next step.
	// +optional
	Duration *int32 `json:"duration,omitempty"`
}
type Recreate struct {
}

type Rolling struct {
	MaxSurge       string `json:"maxSurge"`
	MaxUnavailable int    `json:"maxUnavailable"`
}

func (impl PipelineBuilderImpl) createCdPipeline(app *pipelineConfig.App, pipeline *bean.CDPipelineConfigObject, userID int32, ctx context.Context) (pipelineRes int, err error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(app.Id)
	if err != nil {
		return 0, err
	}
	envOverride, err := impl.propertiesConfigService.CreateIfRequired(chart, pipeline.EnvironmentId, userID, false, models.CHARTSTATUS_NEW, false, pipeline.Namespace)
	if err != nil {
		return 0, err
	}
	/*exists, err := impl.dbPipelineOrchestrator.PipelineExists(pipeline.Name)
	if err != nil {
		impl.logger.Errorw("error in pipeline name duplicate check", "name", pipeline.Name, "err", err)
	}
	if exists {
		return 0, errors.AlreadyExistsf("pipeline already exists name:%s", pipeline.Name)
	}*/

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return 0, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//new pipeline
	impl.logger.Debugw("new pipeline found", "pipeline", pipeline)
	name, err := impl.createArgoPipelineIfRequired(app, pipeline, envOverride, ctx)
	if err != nil {
		return 0, err
	}
	impl.logger.Debugw("argocd application created", "name", name)

	// Get pipeline override based on Deployment strategy
	//TODO mark as created in our db
	pipelineId, err := impl.dbPipelineOrchestrator.CreateCDPipelines(pipeline, app.Id, userID, tx)
	if err != nil {
		impl.logger.Errorw("error in ")
		return 0, err
	}

	//adding ci pipeline to workflow
	appWorkflowModel, err := impl.appWorkflowRepository.FindByIdAndAppId(pipeline.AppWorkflowId, app.Id)
	if err != nil && err != pg.ErrNoRows {
		return 0, err
	}
	if appWorkflowModel.Id > 0 {
		appWorkflowMap := &appWorkflow.AppWorkflowMapping{
			AppWorkflowId: appWorkflowModel.Id,
			ParentId:      pipeline.CiPipelineId,
			ComponentId:   pipelineId,
			Type:          "CD_PIPELINE",
			Active:        true,
			ParentType:    "CI_PIPELINE",
			AuditLog:      models.AuditLog{CreatedBy: userID, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userID},
		}
		_, err = impl.appWorkflowRepository.SaveAppWorkflowMapping(appWorkflowMap, tx)
		if err != nil {
			return 0, err
		}
	}

	// strategies for pipeline ids, there is only one is default
	defaultCount := 0
	for _, item := range pipeline.Strategies {
		if item.Default {
			defaultCount = defaultCount + 1
			if defaultCount > 1 {
				impl.logger.Warnw("already have one strategy is default in this pipeline, skip this", "strategy", item.DeploymentTemplate)
				continue
			}
		}
		strategy := &chartConfig.PipelineStrategy{
			PipelineId: pipelineId,
			Strategy:   item.DeploymentTemplate,
			Config:     string(item.Config),
			Default:    item.Default,
			Deleted:    false,
			AuditLog:   models.AuditLog{UpdatedBy: userID, CreatedBy: userID, UpdatedOn: time.Now(), CreatedOn: time.Now()},
		}
		err = impl.pipelineConfigRepository.Save(strategy, tx)
		if err != nil {
			impl.logger.Errorw("error in saving strategy", "strategy", item.DeploymentTemplate)
			return pipelineId, fmt.Errorf("pipeline created but failed to add strategy")
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	impl.logger.Debugw("pipeline created with GitMaterialId ", "id", pipelineId, "pipeline", pipeline)
	return pipelineId, nil
}

func (impl PipelineBuilderImpl) updateCdPipeline(pipeline *bean.CDPipelineConfigObject, userID int32, ctx context.Context) (err error) {

	if len(pipeline.PreStage.Config) > 0 && !strings.Contains(pipeline.PreStage.Config, "beforeStages") {
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "invalid yaml config, must include - beforeStages",
			UserMessage:     "invalid yaml config, must include - beforeStages",
		}
		return err
	}
	if len(pipeline.PostStage.Config) > 0 && !strings.Contains(pipeline.PostStage.Config, "afterStages") {
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "invalid yaml config, must include - afterStages",
			UserMessage:     "invalid yaml config, must include - afterStages",
		}
		return err
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	err = impl.dbPipelineOrchestrator.UpdateCDPipeline(pipeline, userID, tx)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline")
		return err
	}

	// strategies for pipeline ids, there is only one is default
	existingStrategies, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineId(pipeline.Id)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in getting pipeline strategies", "err", err)
		return err
	}
	for _, oldItem := range existingStrategies {
		notFound := true
		for _, newItem := range pipeline.Strategies {
			if newItem.DeploymentTemplate == oldItem.Strategy {
				notFound = false
			}
		}

		if notFound {
			//delete from db
			err := impl.pipelineConfigRepository.Delete(oldItem, tx)
			if err != nil {
				impl.logger.Errorw("error in delete pipeline strategies", "err", err)
				return fmt.Errorf("error in delete pipeline strategies")
			}
		}
	}

	defaultCount := 0
	for _, item := range pipeline.Strategies {
		if item.Default {
			defaultCount = defaultCount + 1
			if defaultCount > 1 {
				impl.logger.Warnw("already have one strategy is default in this pipeline, skip this", "strategy", item.DeploymentTemplate)
				continue
			}
		}
		strategy, err := impl.pipelineConfigRepository.FindByStrategyAndPipelineId(item.DeploymentTemplate, pipeline.Id)
		if err != nil && pg.ErrNoRows != err {
			impl.logger.Errorw("error in getting strategy", "err", err)
			return err
		}
		if strategy.Id > 0 {
			strategy.Config = string(item.Config)
			strategy.Default = item.Default
			strategy.UpdatedBy = userID
			strategy.UpdatedOn = time.Now()
			err = impl.pipelineConfigRepository.Update(strategy, tx)
			if err != nil {
				impl.logger.Errorw("error in updating strategy", "strategy", item.DeploymentTemplate)
				return fmt.Errorf("pipeline updated but failed to update one strategy")
			}
		} else {
			strategy := &chartConfig.PipelineStrategy{
				PipelineId: pipeline.Id,
				Strategy:   item.DeploymentTemplate,
				Config:     string(item.Config),
				Default:    item.Default,
				Deleted:    false,
				AuditLog:   models.AuditLog{UpdatedBy: userID, CreatedBy: userID, UpdatedOn: time.Now(), CreatedOn: time.Now()},
			}
			err = impl.pipelineConfigRepository.Save(strategy, tx)
			if err != nil {
				impl.logger.Errorw("error in saving strategy", "strategy", item.DeploymentTemplate)
				return fmt.Errorf("pipeline created but failed to add strategy")
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl PipelineBuilderImpl) filterDeploymentTemplate(deploymentTemplate pipelineConfig.DeploymentTemplate, pipelineOverride string) (string, error) {
	var deploymentType DeploymentType
	err := json.Unmarshal([]byte(pipelineOverride), &deploymentType)
	if err != nil {
		impl.logger.Errorw("err", err)
		return "", err
	}
	if pipelineConfig.DEPLOYMENT_TEMPLATE_BLUE_GREEN == deploymentTemplate {
		newDeploymentType := DeploymentType{
			Deployment: Deployment{
				Strategy: Strategy{
					BlueGreen: deploymentType.Deployment.Strategy.BlueGreen,
				},
			},
		}
		pipelineOverrideBytes, err := json.Marshal(newDeploymentType)
		if err != nil {
			impl.logger.Errorw("err", err)
			return "", err
		}
		pipelineOverride = string(pipelineOverrideBytes)
	} else if pipelineConfig.DEPLOYMENT_TEMPLATE_ROLLING == deploymentTemplate {
		newDeploymentType := DeploymentType{
			Deployment: Deployment{
				Strategy: Strategy{
					Rolling: deploymentType.Deployment.Strategy.Rolling,
				},
			},
		}
		pipelineOverrideBytes, err := json.Marshal(newDeploymentType)
		if err != nil {
			impl.logger.Errorw("err", err)
			return "", err
		}
		pipelineOverride = string(pipelineOverrideBytes)
	} else if pipelineConfig.DEPLOYMENT_TEMPLATE_CANARY == deploymentTemplate {
		newDeploymentType := DeploymentType{
			Deployment: Deployment{
				Strategy: Strategy{
					Canary: deploymentType.Deployment.Strategy.Canary,
				},
			},
		}
		pipelineOverrideBytes, err := json.Marshal(newDeploymentType)
		if err != nil {
			impl.logger.Errorw("err", err)
			return "", err
		}
		pipelineOverride = string(pipelineOverrideBytes)
	} else if pipelineConfig.DEPLOYMENT_TEMPLATE_RECREATE == deploymentTemplate {
		newDeploymentType := DeploymentType{
			Deployment: Deployment{
				Strategy: Strategy{
					Recreate: deploymentType.Deployment.Strategy.Recreate,
				},
			},
		}
		pipelineOverrideBytes, err := json.Marshal(newDeploymentType)
		if err != nil {
			impl.logger.Errorw("err", err)
			return "", err
		}
		pipelineOverride = string(pipelineOverrideBytes)
	}
	return pipelineOverride, nil
}

func (impl PipelineBuilderImpl) createArgoPipelineIfRequired(app *pipelineConfig.App, pipeline *bean.CDPipelineConfigObject, envConfigOverride *chartConfig.EnvConfigOverride, ctx context.Context) (string, error) {
	//repo has been registered while helm create
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(app.Id)
	if err != nil {
		impl.logger.Errorw("no chart found ", "app", app.Id)
		return "", err
	}
	envModel, err := impl.environmentRepository.FindById(envConfigOverride.TargetEnvironment)
	if err != nil {
		return "", err
	}
	argoAppName := fmt.Sprintf("%s-%s", app.AppName, envModel.Name)
	_, err = impl.application.Get(ctx, &application2.ApplicationQuery{Name: &argoAppName})
	//if status, ok:=status.FromError(err);ok{
	appStatus, _ := status.FromError(err)

	if appStatus.Code() == codes.OK {
		impl.logger.Infow("argo app already exists", "app", argoAppName, "pipeline", pipeline.Name)
		return argoAppName, nil
	} else if appStatus.Code() == codes.NotFound {
		//create
		appNamespace := envConfigOverride.Namespace
		if len(appNamespace) == 0 {
			appNamespace = "default"
		}

		gocdApplication := v1alpha1.Application{
			ObjectMeta: v1.ObjectMeta{Name: argoAppName},
			Spec: v1alpha1.ApplicationSpec{
				Destination: v1alpha1.ApplicationDestination{Server: envModel.Cluster.ServerUrl, Namespace: appNamespace},
				Source: v1alpha1.ApplicationSource{
					Path:           chart.ChartLocation,
					RepoURL:        chart.GitRepoUrl,
					TargetRevision: "HEAD",
					Helm: &v1alpha1.ApplicationSourceHelm{
						ValueFiles: []string{getValuesFileForEnv(pipeline.EnvironmentId)},
					},
				},
				Project:    "default",
				SyncPolicy: &v1alpha1.SyncPolicy{Automated: &v1alpha1.SyncPolicyAutomated{Prune: true}},
			},
		}
		upsert := true
		create := &application2.ApplicationCreateRequest{
			Application: gocdApplication,
			Upsert:      &upsert,
		}
		impl.logger.Debugw("ass create req", "req", create)
		appRes, err := impl.application.Create(ctx, create)
		if err != nil {
			impl.logger.Errorw("error in creating argo pipeline ", "err", err, "name", pipeline.Name)
			return "", err
		}
		impl.logger.Debugw("pipeline create res ", "res", appRes)
		return argoAppName, nil
	} else {
		impl.logger.Errorw("err in checking application on gocd", "err", err, "pipeline", pipeline.Name)
		return "", err
	}
}

func getValuesFileForEnv(environmentId int) string {
	return fmt.Sprintf("_%d-values.yaml", environmentId) //-{envId}-values.yaml
}
func (impl PipelineBuilderImpl) GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	cdPipelines, err = impl.dbPipelineOrchestrator.GetCdPipelinesForApp(appId)
	var pipelines []*bean.CDPipelineConfigObject
	for _, dbPipeline := range cdPipelines.Pipelines {
		environment, err := impl.environmentRepository.FindById(dbPipeline.EnvironmentId)
		if err != nil && errors.IsNotFound(err) {
			impl.logger.Errorw("error in fetching pipeline", "err", err)
			return cdPipelines, err
		}
		strategies, err := impl.pipelineConfigRepository.GetAllStrategyByPipelineId(dbPipeline.Id)
		if err != nil && errors.IsNotFound(err) {
			impl.logger.Errorw("error in fetching strategies", "err", err)
			return cdPipelines, err
		}
		var strategiesBean []bean.Strategy
		var deploymentTemplate pipelineConfig.DeploymentTemplate
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
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines.Pipelines = pipelines
	return cdPipelines, err
}

func (impl PipelineBuilderImpl) GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error) {
	return impl.dbPipelineOrchestrator.GetCdPipelinesForAppAndEnv(appId, envId)
}

type ConfigMapSecretsResponse struct {
	Maps    []bean2.Map `json:"maps"`
	Secrets []bean2.Map `json:"secrets"`
}

func (impl PipelineBuilderImpl) FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error) {
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

func (impl PipelineBuilderImpl) GetArtifactsByCDPipeline(cdPipelineId int, stage bean2.CdWorkflowType) (bean.CiArtifactResponse, error) {
	var ciArtifacts []bean.CiArtifactBean
	var ciArtifactsResponse bean.CiArtifactResponse
	if stage == bean2.CD_WORKFLOW_TYPE_PRE {
		artifacts, err := impl.ciArtifactRepository.GetArtifactsByCDPipeline(cdPipelineId)
		if err != nil {
			return ciArtifactsResponse, err
		}

		for _, artifact := range artifacts {
			mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
			if err != nil {
				mInfo = []byte("[]")
				impl.logger.Errorw("Error", err)
			}

			ciArtifacts = append(ciArtifacts, bean.CiArtifactBean{
				Id:           artifact.Id,
				Image:        artifact.Image,
				ImageDigest:  artifact.ImageDigest,
				MaterialInfo: mInfo,
				DeployedTime: formatDate(artifact.DeployedTime, bean.LayoutRFC3339),
				Deployed:     artifact.Deployed,
				Latest:       artifact.Latest,
			})
		}

		ciArtifactsResponse.CdPipelineId = cdPipelineId
		if ciArtifacts == nil {
			ciArtifacts = []bean.CiArtifactBean{}
		}
		ciArtifactsResponse.CiArtifacts = ciArtifacts
	} else if stage == bean2.CD_WORKFLOW_TYPE_DEPLOY {
		pipeline, err := impl.pipelineRepository.FindById(cdPipelineId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("Error", err)
		}
		var artifacts []repository.CiArtifact
		if len(pipeline.PreStageConfig) > 0 {
			artifacts, err = impl.ciArtifactRepository.GetArtifactsByCDPipelineAndRunnerType(cdPipelineId, bean2.CD_WORKFLOW_TYPE_PRE)
			if err != nil {
				return ciArtifactsResponse, err
			}
		} else {
			artifacts, err = impl.ciArtifactRepository.GetArtifactsByCDPipeline(cdPipelineId)
			if err != nil {
				return ciArtifactsResponse, err
			}
		}
		latestFound := false
		artifactMap := make(map[int]int)
		for _, artifact := range artifacts {
			artifactMap[artifact.Id] = artifact.Id
			mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
			if err != nil {
				mInfo = []byte("[]")
				impl.logger.Errorw("Error", "err", err)
			}

			ciArtifacts = append(ciArtifacts, bean.CiArtifactBean{
				Id:           artifact.Id,
				Image:        artifact.Image,
				ImageDigest:  artifact.ImageDigest,
				MaterialInfo: mInfo,
				DeployedTime: formatDate(artifact.DeployedTime, bean.LayoutRFC3339),
				Deployed:     artifact.Deployed,
				Latest:       artifact.Latest,
			})
			if artifact.Latest == true {
				latestFound = true
			}
		}

		//start adding deployed items
		latestCiArtifactId, err := impl.ciArtifactRepository.GetLatest(cdPipelineId)
		if err != nil {
			return ciArtifactsResponse, err
		}
		wfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(cdPipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY, 10)
		if err != nil {
			return ciArtifactsResponse, err
		}
		for _, wfr := range wfrList {
			if _, ok := artifactMap[wfr.CdWorkflow.CiArtifact.Id]; !ok {
				mInfo, err := parseMaterialInfo([]byte(wfr.CdWorkflow.CiArtifact.MaterialInfo), wfr.CdWorkflow.CiArtifact.DataSource)
				if err != nil {
					mInfo = []byte("[]")
					impl.logger.Errorw("Error", "err", err)
				}

				deployed := false
				latest := false
				if wfr.Status == application.Healthy || wfr.Status == application.Degraded {
					deployed = true
				}
				if latestFound == false && (latestCiArtifactId == wfr.CdWorkflow.CiArtifactId) {
					latest = true
					latestFound = true
				}
				ciArtifacts = append(ciArtifacts, bean.CiArtifactBean{
					Id:           wfr.CdWorkflow.CiArtifact.Id,
					Image:        wfr.CdWorkflow.CiArtifact.Image,
					ImageDigest:  wfr.CdWorkflow.CiArtifact.ImageDigest,
					MaterialInfo: mInfo,
					DeployedTime: formatDate(wfr.StartedOn, bean.LayoutRFC3339),
					Deployed:     deployed,
					Latest:       latest,
				})
				artifactMap[wfr.CdWorkflow.CiArtifact.Id] = wfr.CdWorkflow.CiArtifact.Id
			}
		}
		//end

		ciArtifactsResponse.CdPipelineId = cdPipelineId
		if ciArtifacts == nil {
			ciArtifacts = []bean.CiArtifactBean{}
		}
		ciArtifactsResponse.CiArtifacts = ciArtifacts
	} else if stage == bean2.CD_WORKFLOW_TYPE_POST {
		artifactMap := make(map[int]int)
		latestFound := false
		latestCiArtifactId, err := impl.ciArtifactRepository.GetLatest(cdPipelineId)
		if err != nil {
			return ciArtifactsResponse, err
		}
		wfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(cdPipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY, 10)
		if err != nil {
			return ciArtifactsResponse, err
		}
		for _, wfr := range wfrList {
			if _, ok := artifactMap[wfr.CdWorkflow.CiArtifact.Id]; !ok {
				mInfo, err := parseMaterialInfo([]byte(wfr.CdWorkflow.CiArtifact.MaterialInfo), wfr.CdWorkflow.CiArtifact.DataSource)
				if err != nil {
					mInfo = []byte("[]")
					impl.logger.Errorw("Error", "err", err)
				}
				deployed := false
				latest := false
				if wfr.Status == application.Healthy || wfr.Status == application.Degraded {
					deployed = true
				}
				if latestFound == false && (latestCiArtifactId == wfr.CdWorkflow.CiArtifactId) {
					latest = true
					latestFound = true
				}
				ciArtifacts = append(ciArtifacts, bean.CiArtifactBean{
					Id:           wfr.CdWorkflow.CiArtifact.Id,
					Image:        wfr.CdWorkflow.CiArtifact.Image,
					MaterialInfo: mInfo,
					DeployedTime: formatDate(wfr.StartedOn, bean.LayoutRFC3339),
					Deployed:     deployed,
					Latest:       latest,
				})
				artifactMap[wfr.CdWorkflow.CiArtifact.Id] = wfr.CdWorkflow.CiArtifact.Id
			}
		}

		ciArtifactsResponse.CdPipelineId = cdPipelineId
		if ciArtifacts == nil {
			ciArtifacts = []bean.CiArtifactBean{}
		}
		ciArtifactsResponse.CiArtifacts = ciArtifacts
	}

	if len(ciArtifactsResponse.CiArtifacts) > 0 {
		var ids []int
		for _, item := range ciArtifactsResponse.CiArtifacts {
			ids = append(ids, item.Id)
		}
		artifacts, err := impl.ciArtifactRepository.GetByIds(ids)
		if err != nil {
			return ciArtifactsResponse, err
		}
		artifactMap := make(map[int]*repository.CiArtifact)
		for _, artifact := range artifacts {
			artifactMap[artifact.Id] = artifact
		}

		var ciArtifactsFinal []bean.CiArtifactBean
		for _, item := range ciArtifactsResponse.CiArtifacts {
			artifact := artifactMap[item.Id]
			item.Scanned = artifact.Scanned
			item.ScanEnabled = artifact.ScanEnabled
			ciArtifactsFinal = append(ciArtifactsFinal, item)
		}
		ciArtifactsResponse.CiArtifacts = ciArtifactsFinal
	}
	return ciArtifactsResponse, nil
}

func (impl PipelineBuilderImpl) FetchArtifactForRollback(cdPipelineId int) (bean.CiArtifactResponse, error) {
	var ciArtifacts []bean.CiArtifactBean
	var ciArtifactsResponse bean.CiArtifactResponse
	artifacts, err := impl.ciArtifactRepository.FetchArtifactForRollback(cdPipelineId)
	if err != nil {
		return ciArtifactsResponse, err
	}

	for _, artifact := range artifacts {
		mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error", "err", err)
		}
		ciArtifacts = append(ciArtifacts, bean.CiArtifactBean{
			Id:           artifact.Id,
			Image:        artifact.Image,
			MaterialInfo: mInfo,
			//ImageDigest: artifact.ImageDigest,
			//DataSource:   artifact.DataSource,
			DeployedTime: formatDate(artifact.CreatedOn, bean.LayoutRFC3339),
		})
	}

	ciArtifactsResponse.CdPipelineId = cdPipelineId
	if ciArtifacts == nil {
		ciArtifacts = []bean.CiArtifactBean{}
	}
	ciArtifactsResponse.CiArtifacts = ciArtifacts

	return ciArtifactsResponse, nil
}

func parseMaterialInfo(materialInfo json.RawMessage, source string) (json.RawMessage, error) {
	if source != "GOCD" && source != "CI-RUNNER" && source != "EXTERNAL" {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	var ciMaterials []repository.CiMaterialInfo
	err := json.Unmarshal(materialInfo, &ciMaterials)
	if err != nil {
		println("material info", materialInfo)
		println("unmarshal error for material info", "err", err)
	}
	var scmMapList []map[string]string

	for _, material := range ciMaterials {
		scmMap := map[string]string{}
		var url string
		if material.Material.Type == "git" {
			url = material.Material.GitConfiguration.URL
		} else if material.Material.Type == "scm" {
			url = material.Material.ScmConfiguration.URL
		} else {
			return nil, fmt.Errorf("unknown material type:%s ", material.Material.Type)
		}
		if material.Modifications != nil && len(material.Modifications) > 0 {
			revision := material.Modifications[0].Revision
			url = strings.TrimSpace(url)

			scmMap["url"] = url
			scmMap["revision"] = revision
			scmMap["modifiedTime"] = material.Modifications[0].ModifiedTime
			scmMap["author"] = material.Modifications[0].Author
			scmMap["message"] = material.Modifications[0].Message
			scmMap["tag"] = material.Modifications[0].Tag
		}
		scmMapList = append(scmMapList, scmMap)
	}
	mInfo, err := json.Marshal(scmMapList)
	return mInfo, err
}

func (impl PipelineBuilderImpl) FindAppsByTeamId(teamId int) ([]AppBean, error) {
	var appsRes []AppBean
	apps, err := impl.appRepo.FindAppsByTeamId(teamId)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) FindAppsByTeamName(teamName string) ([]AppBean, error) {
	var appsRes []AppBean
	apps, err := impl.appRepo.FindAppsByTeamName(teamName)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) FindPipelineById(cdPipelineId int) (*pipelineConfig.Pipeline, error) {
	return impl.pipelineRepository.FindById(cdPipelineId)
}

type TeamAppBean struct {
	ProjectId   int        `json:"projectId"`
	ProjectName string     `json:"projectName"`
	AppList     []*AppBean `json:"appList"`
}

type AppBean struct {
	Id     int    `json:"id"`
	Name   string `json:"name,notnull"`
	TeamId int    `json:"teamId,omitempty"`
}

func (impl PipelineBuilderImpl) GetAppListByTeamIds(teamIds []int) ([]*TeamAppBean, error) {
	var appsRes []*TeamAppBean
	teamMap := make(map[int]*TeamAppBean)
	if len(teamIds) == 0 {
		return appsRes, nil
	}
	apps, err := impl.appRepo.FindAppsByTeamIds(teamIds)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		if _, ok := teamMap[app.TeamId]; ok {
			teamMap[app.TeamId].AppList = append(teamMap[app.TeamId].AppList, &AppBean{Id: app.Id, Name: app.AppName,})
		} else {

			teamMap[app.TeamId] = &TeamAppBean{ProjectId: app.Team.Id, ProjectName: app.Team.Name,}
			teamMap[app.TeamId].AppList = append(teamMap[app.TeamId].AppList, &AppBean{Id: app.Id, Name: app.AppName,})
		}
	}

	for _, v := range teamMap {
		if len(v.AppList) == 0 {
			v.AppList = make([]*AppBean, 0)
		}
		appsRes = append(appsRes, v)
	}

	if len(appsRes) == 0 {
		appsRes = make([]*TeamAppBean, 0)
	}

	return appsRes, err
}

func (impl PipelineBuilderImpl) GetAppList() ([]AppBean, error) {
	var appsRes []AppBean
	apps, err := impl.appRepo.FindAll()
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) updateArgoPipeline(appId int, pipelineName string, envId int, ctx context.Context) (bool, error) {
	//repo has been registered while helm create
	app, err := impl.GetApp(appId)
	if err != nil {
		impl.logger.Errorw("no app found ", "err", err)
		return false, err
	}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(app.Id)
	if err != nil {
		impl.logger.Errorw("no chart found ", "app", app.Id)
		return false, err
	}
	envModel, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return false, err
	}
	argoAppName := fmt.Sprintf("%s-%s", app.AppName, envModel.Name)
	application, err := impl.application.Get(ctx, &application2.ApplicationQuery{Name: &argoAppName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "app", argoAppName, "pipeline", pipelineName)
		return false, err
	}
	//if status, ok:=status.FromError(err);ok{
	appStatus, _ := status.FromError(err)

	if appStatus.Code() == codes.OK {
		impl.logger.Infow("argo app exists", "app", argoAppName, "pipeline", pipelineName)

		if application.Spec.Source.Path != chart.ChartLocation {
			application.Spec.Source.Path = chart.ChartLocation
			updateReq := &application2.ApplicationUpdateRequest{
				Application: application,
			}
			appRes, err := impl.application.Update(ctx, updateReq)
			if err != nil {
				impl.logger.Errorw("error in creating argo pipeline ", "err", err, "name", pipelineName)
				return false, err
			}
			impl.logger.Debugw("pipeline update req ", "res", appRes)
		} else {
			impl.logger.Debug("pipeline no need to update ")
		}
		return true, nil
	} else if appStatus.Code() == codes.NotFound {
		impl.logger.Infow("argo app not found", "app", argoAppName, "pipeline", pipelineName)
		return false, nil
	} else {
		impl.logger.Errorw("err in checking application on gocd", "err", err, "pipeline", pipelineName)
		return false, err
	}
}

func (impl PipelineBuilderImpl) FetchCDPipelineStrategy(appId int) (PipelineStrategiesResponse, error) {
	pipelineStrategiesResponse := PipelineStrategiesResponse{}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorf("invalid state", "err", err, "appId", appId)
		return pipelineStrategiesResponse, err
	}
	if chart.Id == 0 {
		return pipelineStrategiesResponse, fmt.Errorf("no chart configured")
	}
	pipelineOverride := chart.PipelineOverride
	rollingConfig, err := impl.filterDeploymentTemplate("ROLLING", pipelineOverride)
	if err != nil {
		return pipelineStrategiesResponse, err
	}
	pipelineStrategiesResponse.PipelineStrategy = append(pipelineStrategiesResponse.PipelineStrategy, PipelineStrategy{
		DeploymentTemplate: "ROLLING",
		Config:             []byte(rollingConfig),
		Default:            true,
	})
	bgConfig, err := impl.filterDeploymentTemplate("BLUE-GREEN", pipelineOverride)
	if err != nil {
		return pipelineStrategiesResponse, err
	}
	pipelineStrategiesResponse.PipelineStrategy = append(pipelineStrategiesResponse.PipelineStrategy, PipelineStrategy{
		DeploymentTemplate: "BLUE-GREEN",
		Config:             []byte(bgConfig),
		Default:            false,
	})

	chartVersion := chart.ChartVersion
	chartMajorVersion, err := strconv.Atoi(chartVersion[:1])
	if err != nil {
		impl.logger.Errorw("err", err)
		return pipelineStrategiesResponse, err
	}
	chartMinorVersion, err := strconv.Atoi(chartVersion[2:3])
	if err != nil {
		impl.logger.Errorw("err", err)
		return pipelineStrategiesResponse, err
	}
	if chartMajorVersion <= 3 && chartMinorVersion < 2 {
		return pipelineStrategiesResponse, nil
	}

	canaryConfig, err := impl.filterDeploymentTemplate("CANARY", pipelineOverride)
	if err != nil {
		return pipelineStrategiesResponse, err
	}
	pipelineStrategiesResponse.PipelineStrategy = append(pipelineStrategiesResponse.PipelineStrategy, PipelineStrategy{
		DeploymentTemplate: "CANARY",
		Config:             []byte(canaryConfig),
		Default:            false,
	})

	recreateConfig, err := impl.filterDeploymentTemplate("RECREATE", pipelineOverride)
	if err != nil {
		return pipelineStrategiesResponse, err
	}
	pipelineStrategiesResponse.PipelineStrategy = append(pipelineStrategiesResponse.PipelineStrategy, PipelineStrategy{
		DeploymentTemplate: "RECREATE",
		Config:             []byte(recreateConfig),
		Default:            false,
	})

	return pipelineStrategiesResponse, nil
}

type PipelineStrategiesResponse struct {
	PipelineStrategy []PipelineStrategy `json:"pipelineStrategy"`
}
type PipelineStrategy struct {
	DeploymentTemplate pipelineConfig.DeploymentTemplate `json:"deploymentTemplate,omitempty" validate:"oneof=BLUE-GREEN ROLLING"` //
	Config             json.RawMessage                   `json:"config"`
	Default            bool                              `json:"default"`
}

func (impl PipelineBuilderImpl) GetCdPipelineById(pipelineId int) (cdPipeline *bean.CDPipelineConfigObject, err error) {
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
	var deploymentTemplate pipelineConfig.DeploymentTemplate
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
	}

	return cdPipeline, err
}

func (impl PipelineBuilderImpl) FindByIds(ids []*int) ([]*AppBean, error) {
	var appsRes []*AppBean
	apps, err := impl.appRepo.FindByIds(ids)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, &AppBean{Id: app.Id, Name: app.AppName, TeamId: app.TeamId})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) GetCiPipelineById(pipelineId int) (ciPipeline *bean.CiPipeline, err error) {
	pipeline, err := impl.ciPipelineRepository.FindById(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "pipelineId", pipelineId, "err", err)
		return nil, err
	}
	dockerArgs := make(map[string]string)
	if len(pipeline.DockerArgs) > 0 {
		err := json.Unmarshal([]byte(pipeline.DockerArgs), &dockerArgs)
		if err != nil {
			impl.logger.Warnw("error in unmarshal", "err", err)
		}
	}

	var externalCiConfig bean.ExternalCiConfig
	if pipeline.ExternalCiPipeline != nil {
		externalCiConfig = bean.ExternalCiConfig{
			Id:         pipeline.ExternalCiPipeline.Id,
			AccessKey:  pipeline.ExternalCiPipeline.AccessToken,
			WebhookUrl: impl.ciConfig.ExternalCiWebhookUrl,
			Payload:    impl.ciConfig.ExternalCiPayload,
		}
	}

	ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(pipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci scripts")
		return nil, err
	}

	var beforeDockerBuildScripts []*bean.CiScript
	var afterDockerBuildScripts []*bean.CiScript
	for _, ciScript := range ciPipelineScripts {
		ciScriptResp := &bean.CiScript{
			Id:             ciScript.Id,
			Index:          ciScript.Index,
			Name:           ciScript.Name,
			Script:         ciScript.Script,
			OutputLocation: ciScript.OutputLocation,
		}
		if ciScript.Stage == BEFORE_DOCKER_BUILD {
			beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciScriptResp)
		} else if ciScript.Stage == AFTER_DOCKER_BUILD {
			afterDockerBuildScripts = append(afterDockerBuildScripts, ciScriptResp)
		}
	}
	parentCiPipeline, err := impl.ciPipelineRepository.FindById(pipeline.ParentCiPipeline)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	ciPipeline = &bean.CiPipeline{
		Id:                       pipeline.Id,
		Version:                  pipeline.Version,
		Name:                     pipeline.Name,
		Active:                   pipeline.Active,
		Deleted:                  pipeline.Deleted,
		DockerArgs:               dockerArgs,
		IsManual:                 pipeline.IsManual,
		IsExternal:               pipeline.IsExternal,
		ParentCiPipeline:         pipeline.ParentCiPipeline,
		ParentAppId:              parentCiPipeline.AppId,
		ExternalCiConfig:         externalCiConfig,
		BeforeDockerBuildScripts: beforeDockerBuildScripts,
		AfterDockerBuildScripts:  afterDockerBuildScripts,
		ScanEnabled:              pipeline.ScanEnabled,
	}
	for _, material := range pipeline.CiPipelineMaterials {
		ciMaterial := &bean.CiMaterial{
			Id:              material.Id,
			CheckoutPath:    material.CheckoutPath,
			Path:            material.Path,
			ScmId:           material.ScmId,
			GitMaterialId:   material.GitMaterialId,
			GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
			ScmName:         material.ScmName,
			ScmVersion:      material.ScmVersion,
			Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value},
		}
		ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
	}
	linkedCis, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	ciPipeline.LinkedCount = len(linkedCis)
	return ciPipeline, err
}
