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
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type CiPipelineConfigService interface {
	//GetCiPipeline : retrieves CI pipeline configuration (CiConfigRequest) for a specific application (appId).
	// It fetches CI pipeline data, including pipeline materials, scripts, and associated configurations.
	// It returns a detailed CiConfigRequest.
	//If any errors occur during the retrieval process  CI pipeline configuration remains nil.
	//If you want less detail of ciPipeline ,Please refer GetCiPipelineMin
	GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error)
	//GetCiPipelineById : Retrieve ciPipeline for given ciPipelineId
	GetCiPipelineById(pipelineId int) (ciPipeline *bean.CiPipeline, err error)
	//GetTriggerViewCiPipeline : retrieves a detailed view of the CI pipelines configured for a specific application (appId).
	// It includes information on CI pipeline materials, scripts, configurations, and linked pipelines.
	//If any errors occur ,It returns an error along with a nil result(bean.TriggerViewCiConfig).
	GetTriggerViewCiPipeline(appId int) (*bean.TriggerViewCiConfig, error)
	//GetExternalCi : Lists externalCi for given appId
	//"external CI" refers to CI pipelines and configurations that are managed externally,
	//like by third-party services or tools.
	//It fetches information about external CI pipelines, their webhooks, payload configurations, and related roles.
	//The function constructs an array of ExternalCiConfig objects and returns it.
	// If any errors occur, the function returns an error along with a nil result([]*bean.ExternalCiConfig).
	GetExternalCi(appId int) (ciConfig []*bean.ExternalCiConfig, err error)
	//GetExternalCiById : Retrieve externalCi for given appId and externalCiId.
	//It begins by validating the provided ID and fetching the corresponding CI pipeline from the repository.
	//If the CI pipeline is found, the function constructs an ExternalCiConfig object, encapsulating essential details and returns it.
	// If any errors occur, the function returns an error along with a nil result(*bean.ExternalCiConfig).
	GetExternalCiById(appId int, externalCiId int) (ciConfig *bean.ExternalCiConfig, err error)
	//UpdateCiTemplate : handles updates to the CiTemplate based on the provided CiConfigRequest.
	//It fetches relevant Docker registry information, and updates the CI configuration.
	//The function then creates or modifies associated resources such as Docker repositories.
	//After updating the configuration, it ensures to update history .
	//Finally, the function returns the modified CiConfigRequest.
	//If an error occurs, the CiConfigRequest is nil.
	//If you want to Update CiPipeline please refer PatchCiPipeline
	UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error)
	//PatchCiPipeline :  function manages CI pipeline operations based on the provided CiPatchRequest.
	//It fetches template variables, sets specific attributes, and
	//handles following actions
	// 1. create
	//2. update source
	//3. delete pipelines
	PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error)
	//CreateCiPipeline : manages the creation of a CI pipeline based on the provided CiConfigRequest.
	// It first fetches  application data and configures Docker registry settings
	//then constructs a CI template with specified build configurations, auditing details, and related Git materials.

	CreateCiPipeline(createRequest *bean.CiConfigRequest) (*bean.PipelineCreateResponse, error)
	//GetCiPipelineMin : lists minimum detail of ciPipelines for given appId and envIds.
	//It filters and fetches CI pipelines based on the provided environment identifiers.
	//If no specific environments are provided, it retrieves all CI pipelines associated with the application.
	//If you want more details like buildConfig ,gitMaterials etc, please refer GetCiPipeline
	GetCiPipelineMin(appId int, envIds []int) ([]*bean.CiPipelineMin, error)
	//PatchRegexCiPipeline : Update CI pipeline materials based on the provided regex patch request
	PatchRegexCiPipeline(request *bean.CiRegexPatchRequest) (err error)
	//GetCiPipelineByEnvironment : lists ciPipeline for given environmentId and appIds
	GetCiPipelineByEnvironment(request resourceGroup2.ResourceGroupingRequest) ([]*bean.CiConfigRequest, error)
	//GetCiPipelineByEnvironmentMin : lists minimum detail of ciPipelines for given environmentId and appIds
	GetCiPipelineByEnvironmentMin(request resourceGroup2.ResourceGroupingRequest) ([]*bean.CiPipelineMinResponse, error)
	//GetExternalCiByEnvironment : lists externalCi for given environmentId and appIds
	GetExternalCiByEnvironment(request resourceGroup2.ResourceGroupingRequest) (ciConfig []*bean.ExternalCiConfig, err error)
	DeleteCiPipeline(request *bean.CiPatchRequest) (*bean.CiPipeline, error)
	CreateExternalCiAndAppWorkflowMapping(appId, appWorkflowId int, userId int32, tx *pg.Tx) (int, *appWorkflow.AppWorkflowMapping, error)
}

type CiPipelineConfigServiceImpl struct {
	logger                              *zap.SugaredLogger
	ciTemplateService                   CiTemplateService
	materialRepo                        pipelineConfig.MaterialRepository
	ciPipelineRepository                pipelineConfig.CiPipelineRepository
	ciConfig                            *types.CiCdConfig
	attributesService                   attributes.AttributesService
	ciWorkflowRepository                pipelineConfig.CiWorkflowRepository
	appWorkflowRepository               appWorkflow.AppWorkflowRepository
	pipelineStageService                PipelineStageService
	pipelineRepository                  pipelineConfig.PipelineRepository
	appRepo                             app2.AppRepository
	dockerArtifactStoreRepository       dockerRegistryRepository.DockerArtifactStoreRepository
	ciCdPipelineOrchestrator            CiCdPipelineOrchestrator
	ciTemplateOverrideRepository        pipelineConfig.CiTemplateOverrideRepository
	CiTemplateHistoryService            history.CiTemplateHistoryService
	securityConfig                      *SecurityConfig
	ecrConfig                           *EcrConfig
	ciPipelineMaterialRepository        pipelineConfig.CiPipelineMaterialRepository
	resourceGroupService                resourceGroup2.ResourceGroupService
	enforcerUtil                        rbac.EnforcerUtil
	customTagService                    CustomTagService
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService
	ciPipelineHistoryService            history.CiPipelineHistoryService
	cdWorkflowRepository                pipelineConfig.CdWorkflowRepository
	buildPipelineSwitchService          BuildPipelineSwitchService
}

func NewCiPipelineConfigServiceImpl(logger *zap.SugaredLogger,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator,
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository,
	materialRepo pipelineConfig.MaterialRepository,
	pipelineGroupRepo app2.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ecrConfig *EcrConfig,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	ciConfig *types.CiCdConfig,
	attributesService attributes.AttributesService,
	pipelineStageService PipelineStageService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciTemplateService CiTemplateService,
	ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository,
	CiTemplateHistoryService history.CiTemplateHistoryService,
	enforcerUtil rbac.EnforcerUtil,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	resourceGroupService resourceGroup2.ResourceGroupService,
	customTagService CustomTagService,
	ciPipelineHistoryService history.CiPipelineHistoryService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	buildPipelineSwitchService BuildPipelineSwitchService,
) *CiPipelineConfigServiceImpl {

	securityConfig := &SecurityConfig{}
	err := env.Parse(securityConfig)
	if err != nil {
		logger.Errorw("error in parsing securityConfig,setting  ForceSecurityScanning to default value", "defaultValue", securityConfig.ForceSecurityScanning, "err", err)
	}
	return &CiPipelineConfigServiceImpl{
		logger:                        logger,
		ciCdPipelineOrchestrator:      ciCdPipelineOrchestrator,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
		materialRepo:                  materialRepo,
		appRepo:                       pipelineGroupRepo,
		pipelineRepository:            pipelineRepository,
		ciPipelineRepository:          ciPipelineRepository,
		ecrConfig:                     ecrConfig,
		appWorkflowRepository:         appWorkflowRepository,
		ciConfig:                      ciConfig,
		attributesService:             attributesService,
		pipelineStageService:          pipelineStageService,
		ciPipelineMaterialRepository:  ciPipelineMaterialRepository,
		ciTemplateService:             ciTemplateService,
		ciTemplateOverrideRepository:  ciTemplateOverrideRepository,
		CiTemplateHistoryService:      CiTemplateHistoryService,
		enforcerUtil:                  enforcerUtil,
		ciWorkflowRepository:          ciWorkflowRepository,
		resourceGroupService:          resourceGroupService,
		securityConfig:                securityConfig,
		customTagService:              customTagService,
		ciPipelineHistoryService:      ciPipelineHistoryService,
		cdWorkflowRepository:          cdWorkflowRepository,
		buildPipelineSwitchService:    buildPipelineSwitchService,
	}
}

func (impl *CiPipelineConfigServiceImpl) getCiTemplateVariablesByAppIds(appIds []int) (map[int]*bean.CiConfigRequest, error) {
	ciConfigMap := make(map[int]*bean.CiConfigRequest)
	ciTemplateMap, err := impl.ciTemplateService.FindByAppIds(appIds)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appIds", appIds, "err", err)
		return nil, err
	}
	if errors.IsNotFound(err) {
		impl.logger.Debugw("no ci pipeline exists", "appIds", appIds, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline exists"}
		return nil, err
	}
	gitMaterialsMap := make(map[int][]*pipelineConfig.GitMaterial)
	allGitMaterials, err := impl.materialRepo.FindByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching git materials", "appIds", appIds, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Debugw(" no git materials exists", "appIds", appIds, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no git materials exists"}
		return nil, err
	}
	for _, gitMaterial := range allGitMaterials {
		gitMaterialsMap[gitMaterial.AppId] = append(gitMaterialsMap[gitMaterial.AppId], gitMaterial)
	}
	for _, ciTemplate := range ciTemplateMap {
		template := ciTemplate.CiTemplate
		var materials []bean.Material
		gitMaterials := gitMaterialsMap[ciTemplate.CiTemplate.AppId]
		for _, g := range gitMaterials {
			m := bean.Material{
				GitMaterialId: g.Id,
				MaterialName:  g.Name[strings.Index(g.Name, "-")+1:],
			}
			materials = append(materials, m)
		}

		var regHost string
		dockerRegistry := template.DockerRegistry
		if dockerRegistry != nil {
			regHost, err = dockerRegistry.GetRegistryLocation()
			if err != nil {
				impl.logger.Errorw("invalid reg url", "err", err)
				return nil, err
			}
		}
		ciConfig := &bean.CiConfigRequest{
			Id:                template.Id,
			AppId:             template.AppId,
			AppName:           template.App.AppName,
			DockerRepository:  template.DockerRepository,
			DockerRegistryUrl: regHost,
			CiBuildConfig:     ciTemplate.CiBuildConfig,
			Version:           template.Version,
			CiTemplateName:    template.TemplateName,
			Materials:         materials,
			//UpdatedOn:         template.UpdatedOn,
			//UpdatedBy:         template.UpdatedBy,
			//CreatedBy:         template.CreatedBy,
			//CreatedOn:         template.CreatedOn,
		}
		if dockerRegistry != nil {
			ciConfig.DockerRegistry = dockerRegistry.Id
		}
		ciConfigMap[template.AppId] = ciConfig
	}
	return ciConfigMap, err
}

func (impl *CiPipelineConfigServiceImpl) getDefaultArtifactStore(id string) (store *dockerRegistryRepository.DockerArtifactStore, err error) {
	if id == "" {
		impl.logger.Debugw("docker repo is empty adding default repo")
		store, err = impl.dockerArtifactStoreRepository.FindActiveDefaultStore()

	} else {
		store, err = impl.dockerArtifactStoreRepository.FindOne(id)
	}
	return
}

func (impl *CiPipelineConfigServiceImpl) patchCiPipelineUpdateSource(baseCiConfig *bean.CiConfigRequest, modifiedCiPipeline *bean.CiPipeline) (ciConfig *bean.CiConfigRequest, err error) {

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
		modifiedCiPipeline, err = impl.ciCdPipelineOrchestrator.PatchMaterialValue(modifiedCiPipeline, baseCiConfig.UserId, pipeline)
		if err != nil {
			return nil, err
		}
		baseCiConfig.CiPipelines = append(baseCiConfig.CiPipelines, modifiedCiPipeline)
		return baseCiConfig, err
	}

}

func (impl *CiPipelineConfigServiceImpl) buildResponses() []bean.ResponseSchemaObject {
	responseSchemaObjects := make([]bean.ResponseSchemaObject, 0)
	schema := make(map[string]interface{})
	schema["code"] = &bean.SchemaObject{Description: "http status code", DataType: "integer", Example: "200,400,401", Optional: false}
	schema["result"] = &bean.SchemaObject{Description: "api response", DataType: "string", Example: "url", Optional: true}
	schema["status"] = &bean.SchemaObject{Description: "api response status", DataType: "string", Example: "url", Optional: true}

	error := make(map[string]interface{})
	error["code"] = &bean.SchemaObject{Description: "http status code", DataType: "integer", Example: "200,400,401", Optional: true}
	error["userMessage"] = &bean.SchemaObject{Description: "api error user message", DataType: "string", Example: "message", Optional: true}
	schema["error"] = &bean.SchemaObject{Description: "api error", DataType: "object", Example: "{}", Optional: true, Child: error}
	description200 := bean.ResponseDescriptionSchemaObject{
		Description: "success http api response",
		ExampleValue: bean.ExampleValueDto{
			Code:   200,
			Result: "api response result",
		},
		Schema: schema,
	}
	response200 := bean.ResponseSchemaObject{
		Description: description200,
		Code:        "200",
	}
	badReq := bean.ErrorDto{
		Code:        400,
		UserMessage: "Bad request",
	}
	description400 := bean.ResponseDescriptionSchemaObject{
		Description: "bad http request api response",
		ExampleValue: bean.ExampleValueDto{
			Code:   400,
			Errors: []bean.ErrorDto{badReq},
		},
		Schema: schema,
	}

	response400 := bean.ResponseSchemaObject{
		Description: description400,
		Code:        "400",
	}
	description401 := bean.ResponseDescriptionSchemaObject{
		Description: "unauthorized http api response",
		ExampleValue: bean.ExampleValueDto{
			Code:   401,
			Result: "Unauthorized",
		},
		Schema: schema,
	}
	response401 := bean.ResponseSchemaObject{
		Description: description401,
		Code:        "401",
	}
	responseSchemaObjects = append(responseSchemaObjects, response200)
	responseSchemaObjects = append(responseSchemaObjects, response400)
	responseSchemaObjects = append(responseSchemaObjects, response401)
	return responseSchemaObjects
}

func (impl *CiPipelineConfigServiceImpl) buildPayloadOption() []bean.PayloadOptionObject {
	payloadOption := make([]bean.PayloadOptionObject, 0)
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "dockerImage",
		PayloadKey: []string{"dockerImage"},
		Label:      "Container image tag",
		Mandatory:  true,
	})

	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "commitHash",
		PayloadKey: []string{"ciProjectDetails.commitHash"},
		Label:      "Commit hash",
		Mandatory:  false,
	})
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "message",
		PayloadKey: []string{"ciProjectDetails.message"},
		Label:      "Commit message",
		Mandatory:  false,
	})
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "author",
		PayloadKey: []string{"ciProjectDetails.author"},
		Label:      "Author",
		Mandatory:  false,
	})
	payloadOption = append(payloadOption, bean.PayloadOptionObject{
		Key:        "commitTime",
		PayloadKey: []string{"ciProjectDetails.commitTime"},
		Label:      "Date & time of commit",
		Mandatory:  false,
	})
	return payloadOption
}

func (impl *CiPipelineConfigServiceImpl) buildExternalCiWebhookSchema() map[string]interface{} {
	schema := make(map[string]interface{})
	schema["dockerImage"] = &bean.SchemaObject{Description: "docker image created for your application (Eg. docker/test:latest)", DataType: "String", Example: "test-docker-repo/test:b150cc81-5-20", Optional: false}

	ciProjectDetails := make([]map[string]interface{}, 0)
	ciProjectDetail := make(map[string]interface{})
	ciProjectDetail["commitHash"] = &bean.SchemaObject{Description: "Hash of git commit used to build the image (Eg. 4bd84gba5ebdd6b2ad52ede782)", DataType: "String", Example: "dg46f67559dbsdfdfdfdsfba47901caf47f8b7e", Optional: true}
	ciProjectDetail["commitTime"] = &bean.SchemaObject{Description: "Time at which the code was committed to git (Eg. 2022-11-12T12:12:00)", DataType: "String", Example: "2022-11-12T12:12:00", Optional: true}
	ciProjectDetail["message"] = &bean.SchemaObject{Description: "Message provided during code commit (Eg. This is a sample commit message)", DataType: "String", Example: "commit message", Optional: true}
	ciProjectDetail["author"] = &bean.SchemaObject{Description: "Name or email id of the user who has done git commit (Eg. John Doe, johndoe@company.com)", DataType: "String", Example: "Devtron User", Optional: true}
	ciProjectDetails = append(ciProjectDetails, ciProjectDetail)

	schema["ciProjectDetails"] = &bean.SchemaObject{Description: "Git commit details used to build the image", DataType: "Array", Example: "[{}]", Optional: true, Child: ciProjectDetails}
	return schema
}

func (impl *CiPipelineConfigServiceImpl) getCiTemplateVariables(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	ciTemplateBean, err := impl.ciTemplateService.FindByAppId(appId)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	if errors.IsNotFound(err) {
		impl.logger.Debugw("no ci pipeline exists", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline exists"}
		return nil, err
	}
	template := ciTemplateBean.CiTemplate

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

	var regHost string
	var templateDockerRegistryId string
	dockerRegistry := template.DockerRegistry
	if dockerRegistry != nil {
		regHost, err = dockerRegistry.GetRegistryLocation()
		if err != nil {
			impl.logger.Errorw("invalid reg url", "err", err)
			return nil, err
		}
		templateDockerRegistryId = dockerRegistry.Id
	}
	ciConfig = &bean.CiConfigRequest{
		Id:                template.Id,
		AppId:             template.AppId,
		AppName:           template.App.AppName,
		DockerRepository:  template.DockerRepository,
		DockerRegistryUrl: regHost,
		CiBuildConfig:     ciTemplateBean.CiBuildConfig,
		Version:           template.Version,
		CiTemplateName:    template.TemplateName,
		Materials:         materials,
		UpdatedOn:         template.UpdatedOn,
		UpdatedBy:         template.UpdatedBy,
		CreatedBy:         template.CreatedBy,
		CreatedOn:         template.CreatedOn,
		CiGitMaterialId:   template.GitMaterialId,
		DockerRegistry:    templateDockerRegistryId,
	}
	if dockerRegistry != nil {
		ciConfig.DockerRegistry = dockerRegistry.Id
	}
	return ciConfig, err
}

func (impl *CiPipelineConfigServiceImpl) GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	ciConfig, err = impl.getCiTemplateVariables(appId)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	//TODO fill these variables
	//ciCdConfig.CiPipeline=
	//--------pipeline population start
	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}

	if impl.ciConfig.ExternalCiWebhookUrl == "" {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, types.ExternalCiWebhookPath)
		}
	}
	//map of ciPipelineId and their templateOverrideConfig
	ciOverrideTemplateMap := make(map[int]*bean3.CiTemplateBean)
	ciTemplateBeanOverrides, err := impl.ciTemplateService.FindTemplateOverrideByAppId(appId)
	if err != nil {
		return nil, err
	}

	for _, templateBeanOverride := range ciTemplateBeanOverrides {
		ciTemplateOverride := templateBeanOverride.CiTemplateOverride
		ciOverrideTemplateMap[ciTemplateOverride.CiPipelineId] = templateBeanOverride
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
			IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
			PipelineType:             bean.PipelineType(pipeline.PipelineType),
		}
		ciEnvMapping, err := impl.ciPipelineRepository.FindCiEnvMappingByCiPipelineId(pipeline.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching ciEnvMapping", "ciPipelineId ", pipeline.Id, "err", err)
			return nil, err
		}
		customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(bean3.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id))
		if err != nil && err != pg.ErrNoRows {
			return nil, err
		}
		if customTag.Id != 0 {
			ciPipeline.CustomTagObject = &bean.CustomTagData{
				TagPattern: customTag.TagPattern,
				CounterX:   customTag.AutoIncreasingNumber,
				Enabled:    customTag.Enabled,
			}
			ciPipeline.EnableCustomTag = customTag.Enabled
		}
		if ciEnvMapping.Id > 0 {
			ciPipeline.EnvironmentId = ciEnvMapping.EnvironmentId
		}

		lastTriggeredWorkflowEnv, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflow(pipeline.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching lasTriggeredWorkflowEnv", "ciPipelineId ", pipeline.Id, "err", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			ciPipeline.LastTriggeredEnvId = -1
		} else {
			ciPipeline.LastTriggeredEnvId = lastTriggeredWorkflowEnv.EnvironmentId
		}

		if ciTemplateBean, ok := ciOverrideTemplateMap[pipeline.Id]; ok {
			templateOverride := ciTemplateBean.CiTemplateOverride
			ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
				DockerRegistry:   templateOverride.DockerRegistryId,
				DockerRepository: templateOverride.DockerRepository,
				CiBuildConfig:    ciTemplateBean.CiBuildConfig,
			}
		}
		for _, material := range pipeline.CiPipelineMaterials {
			// ignore those materials which have inactive git material
			if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
				continue
			}
			ciMaterial := &bean.CiMaterial{
				Id:              material.Id,
				CheckoutPath:    material.CheckoutPath,
				Path:            material.Path,
				ScmId:           material.ScmId,
				GitMaterialId:   material.GitMaterialId,
				GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
				ScmName:         material.ScmName,
				ScmVersion:      material.ScmVersion,
				IsRegex:         material.Regex != "",
				Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
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

func (impl *CiPipelineConfigServiceImpl) GetCiPipelineById(pipelineId int) (ciPipeline *bean.CiPipeline, err error) {
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

	if impl.ciConfig.ExternalCiWebhookUrl == "" {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			impl.logger.Errorw("there is no external ci webhook url configured", "ci pipeline", pipeline)
			return nil, err
		}
		if hostUrl != nil {
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, types.ExternalCiWebhookPath)
		}
	}

	var externalCiConfig bean.ExternalCiConfig

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
		AppId:                    pipeline.AppId,
		ParentCiPipeline:         pipeline.ParentCiPipeline,
		ParentAppId:              parentCiPipeline.AppId,
		ExternalCiConfig:         externalCiConfig,
		BeforeDockerBuildScripts: beforeDockerBuildScripts,
		AfterDockerBuildScripts:  afterDockerBuildScripts,
		ScanEnabled:              pipeline.ScanEnabled,
		IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
		PipelineType:             bean.PipelineType(pipeline.PipelineType),
	}
	customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(bean3.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id))
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if customTag.Id != 0 {
		ciPipeline.CustomTagObject = &bean.CustomTagData{
			TagPattern: customTag.TagPattern,
			CounterX:   customTag.AutoIncreasingNumber,
		}
		ciPipeline.EnableCustomTag = customTag.Enabled
	}
	ciEnvMapping, err := impl.ciPipelineRepository.FindCiEnvMappingByCiPipelineId(pipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching ci env mapping", "pipelineId", pipelineId, "err", err)
		return nil, err
	}
	if ciEnvMapping.Id > 0 {
		ciPipeline.EnvironmentId = ciEnvMapping.EnvironmentId
	}
	if !ciPipeline.IsExternal && ciPipeline.IsDockerConfigOverridden {
		ciTemplateBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(ciPipeline.Id)
		if err != nil {
			return nil, err
		}
		templateOverride := ciTemplateBean.CiTemplateOverride
		ciBuildConfig := ciTemplateBean.CiBuildConfig
		ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
			DockerRegistry:   templateOverride.DockerRegistryId,
			DockerRepository: templateOverride.DockerRepository,
			CiBuildConfig:    ciBuildConfig,
			//DockerBuildConfig: &bean.DockerBuildConfig{
			//	GitMaterialId:  templateOverride.GitMaterialId,
			//	DockerfilePath: templateOverride.DockerfilePath,
			//},
		}
	}
	for _, material := range pipeline.CiPipelineMaterials {
		if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
			continue
		}
		ciMaterial := &bean.CiMaterial{
			Id:              material.Id,
			CheckoutPath:    material.CheckoutPath,
			Path:            material.Path,
			ScmId:           material.ScmId,
			GitMaterialId:   material.GitMaterialId,
			GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
			ScmName:         material.ScmName,
			ScmVersion:      material.ScmVersion,
			IsRegex:         material.Regex != "",
			Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
		}
		ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
	}

	linkedCis, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	ciPipeline.LinkedCount = len(linkedCis)

	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCIMappingByCIPipelineId(ciPipeline.Id)
	for _, mapping := range appWorkflowMappings {
		//there will be only one active entry in db always
		ciPipeline.AppWorkflowId = mapping.AppWorkflowId
	}

	//getting pre stage and post stage details
	preStageDetail, postStageDetail, err := impl.pipelineStageService.GetCiPipelineStageData(ciPipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in getting pre & post stage detail by ciPipelineId", "err", err, "ciPipelineId", ciPipeline.Id)
		return nil, err
	}
	ciPipeline.PreBuildStage = preStageDetail
	ciPipeline.PostBuildStage = postStageDetail
	return ciPipeline, err
}

func (impl *CiPipelineConfigServiceImpl) GetTriggerViewCiPipeline(appId int) (*bean.TriggerViewCiConfig, error) {

	triggerViewCiConfig := &bean.TriggerViewCiConfig{}

	ciConfig, err := impl.getCiTemplateVariables(appId)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}

	triggerViewCiConfig.CiGitMaterialId = ciConfig.CiBuildConfig.GitMaterialId

	// fetch pipelines
	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}

	ciOverrideTemplateMap := make(map[int]*bean3.CiTemplateBean)
	ciTemplateBeanOverrides, err := impl.ciTemplateService.FindTemplateOverrideByAppId(appId)
	if err != nil {
		return nil, err
	}

	for _, templateBeanOverride := range ciTemplateBeanOverrides {
		ciTemplateOverride := templateBeanOverride.CiTemplateOverride
		ciOverrideTemplateMap[ciTemplateOverride.CiPipelineId] = templateBeanOverride
	}

	var ciPipelineResp []*bean.CiPipeline
	for _, pipeline := range pipelines {
		isLinkedCiPipeline := pipeline.IsExternal
		ciPipeline := &bean.CiPipeline{
			Id:                       pipeline.Id,
			Version:                  pipeline.Version,
			Name:                     pipeline.Name,
			Active:                   pipeline.Active,
			Deleted:                  pipeline.Deleted,
			IsManual:                 pipeline.IsManual,
			IsExternal:               isLinkedCiPipeline,
			ParentCiPipeline:         pipeline.ParentCiPipeline,
			ScanEnabled:              pipeline.ScanEnabled,
			IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
			PipelineType:             bean.PipelineType(pipeline.PipelineType),
		}
		if ciTemplateBean, ok := ciOverrideTemplateMap[pipeline.Id]; ok {
			templateOverride := ciTemplateBean.CiTemplateOverride
			ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
				DockerRegistry:   templateOverride.DockerRegistryId,
				DockerRepository: templateOverride.DockerRepository,
				CiBuildConfig:    ciTemplateBean.CiBuildConfig,
			}
		}
		for _, material := range pipeline.CiPipelineMaterials {
			// ignore those materials which have inactive git material
			if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
				continue
			}
			ciMaterial := &bean.CiMaterial{
				Id:              material.Id,
				CheckoutPath:    material.CheckoutPath,
				Path:            material.Path,
				ScmId:           material.ScmId,
				GitMaterialId:   material.GitMaterialId,
				GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
				ScmName:         material.ScmName,
				ScmVersion:      material.ScmVersion,
				IsRegex:         material.Regex != "",
				Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
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
	triggerViewCiConfig.CiPipelines = ciPipelineResp
	triggerViewCiConfig.Materials = ciConfig.Materials

	return triggerViewCiConfig, nil
}

func (impl *CiPipelineConfigServiceImpl) GetExternalCi(appId int) (ciConfig []*bean.ExternalCiConfig, err error) {
	externalCiPipelines, err := impl.ciPipelineRepository.FindExternalCiByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}

	hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
	if err != nil {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}
	if hostUrl != nil {
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, types.ExternalCiWebhookPath)
	}

	externalCiConfigs := make([]*bean.ExternalCiConfig, 0)

	var externalCiPipelineIds []int
	appWorkflowMappingsMap := make(map[int][]*appWorkflow.AppWorkflowMapping)

	for _, externalCiPipeline := range externalCiPipelines {
		externalCiPipelineIds = append(externalCiPipelineIds, externalCiPipeline.Id)
	}
	if len(externalCiPipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no external ci pipeline found"}
		return externalCiConfigs, err
	}
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiIdByIdsIn(externalCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("Error in fetching app workflow mapping for CD pipeline by external CI ID", "err", err)
		return nil, err
	}

	for _, appWorkflowMapping := range appWorkflowMappings {
		appWorkflowMappingsMap[appWorkflowMapping.ParentId] = append(appWorkflowMappingsMap[appWorkflowMapping.ParentId], appWorkflowMapping)
	}

	for _, externalCiPipeline := range externalCiPipelines {
		externalCiConfig := &bean.ExternalCiConfig{
			Id:         externalCiPipeline.Id,
			WebhookUrl: fmt.Sprintf("%s/%d", impl.ciConfig.ExternalCiWebhookUrl, externalCiPipeline.Id),
			Payload:    impl.ciConfig.ExternalCiPayload,
			AccessKey:  "",
		}

		if _, ok := appWorkflowMappingsMap[externalCiPipeline.Id]; !ok {
			impl.logger.Errorw("unable to find app workflow cd mapping corresponding to external ci pipeline id")
			return nil, errors.New("unable to find app workflow cd mapping corresponding to external ci pipeline id")
		}

		var appWorkflowComponentIds []int
		var appIds []int

		CDPipelineMap := make(map[int]*pipelineConfig.Pipeline)
		appIdMap := make(map[int]*app2.App)

		for _, appWorkflowMappings := range appWorkflowMappings {
			appWorkflowComponentIds = append(appWorkflowComponentIds, appWorkflowMappings.ComponentId)
		}
		if len(appWorkflowComponentIds) == 0 {
			continue
		}
		cdPipelines, err := impl.pipelineRepository.FindAppAndEnvironmentAndProjectByPipelineIds(appWorkflowComponentIds)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
			return nil, err
		}
		for _, pipeline := range cdPipelines {
			CDPipelineMap[pipeline.Id] = pipeline
			appIds = append(appIds, pipeline.AppId)
		}
		if len(appIds) == 0 {
			continue
		}
		apps, err := impl.appRepo.FindAppAndProjectByIdsIn(appIds)
		for _, app := range apps {
			appIdMap[app.Id] = app
		}

		roleData := make(map[string]interface{})
		for _, appWorkflowMapping := range appWorkflowMappings {
			if _, ok := CDPipelineMap[appWorkflowMapping.ComponentId]; !ok {
				impl.logger.Errorw("error in getting cd pipeline data for workflow", "app workflow id", appWorkflowMapping.ComponentId, "err", err)
				return nil, errors.New("error in getting cd pipeline data for workflow")
			}
			cdPipeline := CDPipelineMap[appWorkflowMapping.ComponentId]

			if _, ok := roleData[teamIdKey]; !ok {
				if _, ok := appIdMap[cdPipeline.AppId]; !ok {
					impl.logger.Errorw("error in getting app data for pipeline", "app id", cdPipeline.AppId)
					return nil, errors.New("error in getting app data for pipeline")
				}
				app := appIdMap[cdPipeline.AppId]
				roleData[teamIdKey] = app.TeamId
				roleData[teamNameKey] = app.Team.Name
				roleData[appIdKey] = cdPipeline.AppId
				roleData[appNameKey] = cdPipeline.App.AppName
			}
			if _, ok := roleData[environmentNameKey]; !ok {
				roleData[environmentNameKey] = cdPipeline.Environment.Name
			} else {
				roleData[environmentNameKey] = fmt.Sprintf("%s,%s", roleData[environmentNameKey], cdPipeline.Environment.Name)
			}
			if _, ok := roleData[environmentIdentifierKey]; !ok {
				roleData[environmentIdentifierKey] = cdPipeline.Environment.EnvironmentIdentifier
			} else {
				roleData[environmentIdentifierKey] = fmt.Sprintf("%s,%s", roleData[environmentIdentifierKey], cdPipeline.Environment.EnvironmentIdentifier)
			}
		}

		externalCiConfig.ExternalCiConfigRole = bean.ExternalCiConfigRole{
			ProjectId:             roleData[teamIdKey].(int),
			ProjectName:           roleData[teamNameKey].(string),
			AppId:                 roleData[appIdKey].(int),
			AppName:               roleData[appNameKey].(string),
			EnvironmentName:       roleData[environmentNameKey].(string),
			EnvironmentIdentifier: roleData[environmentIdentifierKey].(string),
			Role:                  "Build and deploy",
		}
		externalCiConfigs = append(externalCiConfigs, externalCiConfig)
	}
	//--------pipeline population end
	return externalCiConfigs, err
}

func (impl *CiPipelineConfigServiceImpl) GetExternalCiById(appId int, externalCiId int) (ciConfig *bean.ExternalCiConfig, err error) {

	externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(externalCiId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}

	if externalCiPipeline.Id == 0 {
		impl.logger.Errorw("invalid external ci id", "externalCiId", externalCiId, "err", err)
		return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "invalid external ci id"}
	}

	hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
	if err != nil {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}
	if hostUrl != nil {
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, types.ExternalCiWebhookPath)
	}

	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiId(externalCiPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
		return nil, err
	}

	roleData := make(map[string]interface{})
	for _, appWorkflowMapping := range appWorkflowMappings {
		cdPipeline, err := impl.pipelineRepository.FindById(appWorkflowMapping.ComponentId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
			return nil, err
		}
		if _, ok := roleData[teamIdKey]; !ok {
			app, err := impl.appRepo.FindAppAndProjectByAppId(cdPipeline.AppId)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("error in fetching external ci", "appId", appId, "err", err)
				return nil, err
			}
			roleData[teamIdKey] = app.TeamId
			roleData[teamNameKey] = app.Team.Name
			roleData[appIdKey] = cdPipeline.AppId
			roleData[appNameKey] = cdPipeline.App.AppName
		}
		if _, ok := roleData[environmentNameKey]; !ok {
			roleData[environmentNameKey] = cdPipeline.Environment.Name
		} else {
			roleData[environmentNameKey] = fmt.Sprintf("%s,%s", roleData[environmentNameKey], cdPipeline.Environment.Name)
		}
		if _, ok := roleData[environmentIdentifierKey]; !ok {
			roleData[environmentIdentifierKey] = cdPipeline.Environment.EnvironmentIdentifier
		} else {
			roleData[environmentIdentifierKey] = fmt.Sprintf("%s,%s", roleData[environmentIdentifierKey], cdPipeline.Environment.EnvironmentIdentifier)
		}
	}

	externalCiConfig := &bean.ExternalCiConfig{
		Id:         externalCiPipeline.Id,
		WebhookUrl: fmt.Sprintf("%s/%d", impl.ciConfig.ExternalCiWebhookUrl, externalCiId),
		Payload:    impl.ciConfig.ExternalCiPayload,
		AccessKey:  "",
	}
	externalCiConfig.ExternalCiConfigRole = bean.ExternalCiConfigRole{
		ProjectId:             roleData[teamIdKey].(int),
		ProjectName:           roleData[teamNameKey].(string),
		AppId:                 roleData[appIdKey].(int),
		AppName:               roleData[appNameKey].(string),
		EnvironmentName:       roleData[environmentNameKey].(string),
		EnvironmentIdentifier: roleData[environmentIdentifierKey].(string),
		Role:                  "Build and deploy",
	}
	externalCiConfig.Schema = impl.buildExternalCiWebhookSchema()
	externalCiConfig.PayloadOption = impl.buildPayloadOption()
	externalCiConfig.Responses = impl.buildResponses()
	//--------pipeline population end
	return externalCiConfig, err
}

func (impl *CiPipelineConfigServiceImpl) UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {
	originalCiConf, err := impl.getCiTemplateVariables(updateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching original ciCdConfig for update", "appId", updateRequest.Id, "err", err)
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
	regHost, err := dockerArtifaceStore.GetRegistryLocation()
	if err != nil {
		impl.logger.Errorw("invalid reg url", "err", err)
		return nil, err
	}

	var repo string
	if updateRequest.DockerRepository != "" {
		repo = updateRequest.DockerRepository
	} else {
		repo = originalCiConf.DockerRepository
	}

	if dockerArtifaceStore.RegistryType == dockerRegistryRepository.REGISTRYTYPE_ECR {
		err := impl.ciCdPipelineOrchestrator.CreateEcrRepo(repo, dockerArtifaceStore.AWSRegion, dockerArtifaceStore.AWSAccessKeyId, dockerArtifaceStore.AWSSecretAccessKey)
		if err != nil {
			impl.logger.Errorw("ecr repo creation failed while updating ci template", "repo", repo, "err", err)
			return nil, err
		}
	}

	originalCiConf.AfterDockerBuild = updateRequest.AfterDockerBuild
	originalCiConf.BeforeDockerBuild = updateRequest.BeforeDockerBuild
	//originalCiConf.CiBuildConfigBean = updateRequest.CiBuildConfigBean
	originalCiConf.DockerRegistry = updateRequest.DockerRegistry
	originalCiConf.DockerRepository = updateRequest.DockerRepository
	originalCiConf.DockerRegistryUrl = regHost

	//argByte, err := json.Marshal(originalCiConf.DockerBuildConfig.Args)
	//if err != nil {
	//	return nil, err
	//}
	afterByte, err := json.Marshal(originalCiConf.AfterDockerBuild)
	if err != nil {
		return nil, err
	}
	beforeByte, err := json.Marshal(originalCiConf.BeforeDockerBuild)
	if err != nil {
		return nil, err
	}
	//buildOptionsByte, err := json.Marshal(originalCiConf.DockerBuildConfig.DockerBuildOptions)
	//if err != nil {
	//	impl.logger.Errorw("error in marshaling dockerBuildOptions", "err", err)
	//	return nil, err
	//}
	ciBuildConfig := updateRequest.CiBuildConfig
	originalCiBuildConfig := originalCiConf.CiBuildConfig
	ciTemplate := &pipelineConfig.CiTemplate{
		//DockerfilePath:    originalCiConf.DockerBuildConfig.DockerfilePath,
		GitMaterialId:             ciBuildConfig.GitMaterialId,
		BuildContextGitMaterialId: ciBuildConfig.BuildContextGitMaterialId,
		//Args:              string(argByte),
		//TargetPlatform:    originalCiConf.DockerBuildConfig.TargetPlatform,
		AppId:             originalCiConf.AppId,
		BeforeDockerBuild: string(beforeByte),
		AfterDockerBuild:  string(afterByte),
		Version:           originalCiConf.Version,
		Id:                originalCiConf.Id,
		DockerRepository:  originalCiConf.DockerRepository,
		DockerRegistryId:  &originalCiConf.DockerRegistry,
		Active:            true,
		AuditLog: sql.AuditLog{
			CreatedOn: originalCiConf.CreatedOn,
			CreatedBy: originalCiConf.CreatedBy,
			UpdatedOn: time.Now(),
			UpdatedBy: updateRequest.UserId,
		},
	}

	ciBuildConfig.Id = originalCiBuildConfig.Id
	ciTemplateBean := &bean3.CiTemplateBean{
		CiTemplate:    ciTemplate,
		CiBuildConfig: ciBuildConfig,
		UserId:        updateRequest.UserId,
	}
	err = impl.ciTemplateService.Update(ciTemplateBean)
	if err != nil {
		return nil, err
	}

	originalCiConf.CiBuildConfig = ciBuildConfig
	//TODO: below update code is a hack for ci_job and should be reviewed

	// updating ci_template_override for ci_pipeline type = CI_JOB because for this pipeling ci_template and ci_template_override are kept same as
	pipelines, err := impl.ciPipelineRepository.FindByAppId(originalCiConf.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in finding pipeline for app")
	}
	ciPipelineIds := make([]int, 0)
	ciPipelineIdsMap := make(map[int]*pipelineConfig.CiPipeline)
	for _, p := range pipelines {
		ciPipelineIds = append(ciPipelineIds, p.Id)
		ciPipelineIdsMap[p.Id] = p
	}
	var ciTemplateOverrides []*pipelineConfig.CiTemplateOverride
	if len(ciPipelineIds) > 0 {
		ciTemplateOverrides, err = impl.ciTemplateOverrideRepository.FindByCiPipelineIds(ciPipelineIds)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching ci tempalate by pipeline ids", "err", err, "ciPipelineIds", ciPipelineIds)
		}
	}
	for _, ciTemplateOverride := range ciTemplateOverrides {
		if _, ok := ciPipelineIdsMap[ciTemplateOverride.CiPipelineId]; ok {
			if ciPipelineIdsMap[ciTemplateOverride.CiPipelineId].PipelineType == string(bean.CI_JOB) {
				ciTemplateOverride.DockerRepository = updateRequest.DockerRepository
				ciTemplateOverride.DockerRegistryId = updateRequest.DockerRegistry
				_, err = impl.ciTemplateOverrideRepository.Update(ciTemplateOverride)
				if err != nil {
					impl.logger.Errorw("error in updating ci template for ci_job", "err", err)
				}
			}
		}
	}
	// update completed for ci_pipeline_type = ci_job

	err = impl.CiTemplateHistoryService.SaveHistory(ciTemplateBean, "update")

	if err != nil {
		impl.logger.Errorw("error in saving update history for ci template", "error", err)
	}

	return originalCiConf, nil
}

func (impl *CiPipelineConfigServiceImpl) handlePipelineCreate(request *bean.CiPatchRequest, ciConfig *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {

	pipelineExists, err := impl.ciPipelineRepository.CheckIfPipelineExistsByNameAndAppId(request.CiPipeline.Name, request.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching pipeline by name, FindByName", "err", err, "patch cipipeline name", request.CiPipeline.Name)
		return nil, err
	}

	if pipelineExists {
		impl.logger.Errorw("pipeline name already exist", "err", err, "patch cipipeline name", request.CiPipeline.Name)
		return nil, fmt.Errorf("pipeline name already exist")
	}

	if request.IsSwitchCiPipelineRequest() {
		impl.logger.Debugw("handling switch ci pipeline", "switchFromCiPipelineId", request.SwitchFromCiPipelineId, "switchFromExternalCiPipelineId", request.SwitchFromExternalCiPipelineId)
		return impl.buildPipelineSwitchService.SwitchToCiPipelineExceptExternal(request, ciConfig)
	}

	ciConfig.CiPipelines = []*bean.CiPipeline{request.CiPipeline} //request.CiPipeline
	res, err := impl.ciCdPipelineOrchestrator.AddPipelineToTemplate(ciConfig, false)
	if err != nil {
		impl.logger.Errorw("error in adding pipeline to template", "ciConf", ciConfig, "err", err)
		return nil, err
	}
	return res, nil

}

func (impl *CiPipelineConfigServiceImpl) PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error) {
	ciConfig, err = impl.getCiTemplateVariables(request.AppId)
	if err != nil {
		impl.logger.Errorw("err in fetching template for pipeline patch, ", "err", err, "appId", request.AppId)
		return nil, err
	}
	if request.CiPipeline.PipelineType == bean.CI_JOB {
		request.CiPipeline.IsDockerConfigOverridden = true
		request.CiPipeline.DockerConfigOverride = bean.DockerConfigOverride{
			DockerRegistry:   ciConfig.DockerRegistry,
			DockerRepository: ciConfig.DockerRepository,
			CiBuildConfig: &bean3.CiBuildConfigBean{
				Id:                        0,
				GitMaterialId:             request.CiPipeline.CiMaterial[0].GitMaterialId,
				BuildContextGitMaterialId: request.CiPipeline.CiMaterial[0].GitMaterialId,
				UseRootBuildContext:       false,
				CiBuildType:               bean3.SKIP_BUILD_TYPE,
				DockerBuildConfig:         nil,
				BuildPackConfig:           nil,
			},
		}
	}
	ciConfig.AppWorkflowId = request.AppWorkflowId
	ciConfig.UserId = request.UserId
	if request.CiPipeline != nil {
		//setting ScanEnabled value from env variable,
		request.CiPipeline.ScanEnabled = request.CiPipeline.ScanEnabled || impl.securityConfig.ForceSecurityScanning
		ciConfig.ScanEnabled = request.CiPipeline.ScanEnabled
	}

	ciConfig.IsJob = request.IsJob
	// Check for clone job to not create env override again
	ciConfig.IsCloneJob = request.IsCloneJob
	switch request.Action {
	case bean.CREATE:
		res, err := impl.handlePipelineCreate(request, ciConfig)
		if err != nil {
			impl.logger.Errorw("error in creating ci pipeline", "err", err, "request", request, "ciConfig", ciConfig)
		}
		return res, err
	case bean.UPDATE_SOURCE:
		return impl.patchCiPipelineUpdateSource(ciConfig, request.CiPipeline)
	case bean.DELETE:
		pipeline, err := impl.DeleteCiPipeline(request)
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

func (impl *CiPipelineConfigServiceImpl) CreateCiPipeline(createRequest *bean.CiConfigRequest) (*bean.PipelineCreateResponse, error) {
	impl.logger.Debugw("pipeline create request received", "req", createRequest)

	//-----------fetch data
	app, err := impl.appRepo.FindById(createRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline group", "groupId", createRequest.AppId, "err", err)
		return nil, err
	}
	//--ecr config
	createRequest.AppName = app.AppName
	if !createRequest.IsJob {
		store, err := impl.getDefaultArtifactStore(createRequest.DockerRegistry)
		if err != nil {
			impl.logger.Errorw("error in fetching docker store ", "id", createRequest.DockerRepository, "err", err)
			return nil, err
		}

		regHost, err := store.GetRegistryLocation()
		if err != nil {
			impl.logger.Errorw("invalid reg url", "err", err)
			return nil, err
		}
		createRequest.DockerRegistryUrl = regHost
		createRequest.DockerRegistry = store.Id

		var repo string
		if createRequest.DockerRepository != "" {
			repo = createRequest.DockerRepository
		} else {
			repo = impl.ecrConfig.EcrPrefix + app.AppName
		}

		if store.RegistryType == dockerRegistryRepository.REGISTRYTYPE_ECR {
			err := impl.ciCdPipelineOrchestrator.CreateEcrRepo(repo, store.AWSRegion, store.AWSAccessKeyId, store.AWSSecretAccessKey)
			if err != nil {
				impl.logger.Errorw("ecr repo creation failed while creating ci pipeline", "repo", repo, "err", err)
				return nil, err
			}
		}
		createRequest.DockerRepository = repo
	}
	//--ecr config	end
	//-- template config start

	//argByte, err := json.Marshal(createRequest.DockerBuildConfig.Args)
	//if err != nil {
	//	return nil, err
	//}
	afterByte, err := json.Marshal(createRequest.AfterDockerBuild)
	if err != nil {
		return nil, err
	}
	beforeByte, err := json.Marshal(createRequest.BeforeDockerBuild)
	if err != nil {
		return nil, err
	}
	buildConfig := createRequest.CiBuildConfig
	ciTemplate := &pipelineConfig.CiTemplate{
		//DockerRegistryId: createRequest.DockerRegistry,
		//DockerRepository: createRequest.DockerRepository,
		GitMaterialId:             buildConfig.GitMaterialId,
		BuildContextGitMaterialId: buildConfig.BuildContextGitMaterialId,
		//DockerfilePath:    createRequest.DockerBuildConfig.DockerfilePath,
		//Args:              string(argByte),
		//TargetPlatform:    createRequest.DockerBuildConfig.TargetPlatform,
		Active:            true,
		TemplateName:      createRequest.CiTemplateName,
		Version:           createRequest.Version,
		AppId:             createRequest.AppId,
		AfterDockerBuild:  string(afterByte),
		BeforeDockerBuild: string(beforeByte),
		AuditLog:          sql.AuditLog{CreatedOn: time.Now(), UpdatedOn: time.Now(), CreatedBy: createRequest.UserId, UpdatedBy: createRequest.UserId},
	}
	if !createRequest.IsJob {
		ciTemplate.DockerRegistryId = &createRequest.DockerRegistry
		ciTemplate.DockerRepository = createRequest.DockerRepository
	}

	ciTemplateBean := &bean3.CiTemplateBean{
		CiTemplate:    ciTemplate,
		CiBuildConfig: createRequest.CiBuildConfig,
	}
	err = impl.ciTemplateService.Save(ciTemplateBean)
	if err != nil {
		return nil, err
	}

	//-- template config end

	err = impl.CiTemplateHistoryService.SaveHistory(ciTemplateBean, "add")

	if err != nil {
		impl.logger.Errorw("error in saving audit logs of ci Template", "error", err)
	}

	createRequest.Id = ciTemplate.Id
	createRequest.CiTemplateName = ciTemplate.TemplateName
	if len(createRequest.CiPipelines) > 0 {
		conf, err := impl.ciCdPipelineOrchestrator.AddPipelineToTemplate(createRequest, false)
		if err != nil {
			impl.logger.Errorw("error in pipeline creation ", "err", err)
			return nil, err
		}
		impl.logger.Debugw("pipeline created ", "detail", conf)
	}
	createRes := &bean.PipelineCreateResponse{AppName: app.AppName, AppId: createRequest.AppId} //FIXME
	return createRes, nil
}

func (impl *CiPipelineConfigServiceImpl) GetCiPipelineMin(appId int, envIds []int) ([]*bean.CiPipelineMin, error) {
	pipelines := make([]*pipelineConfig.CiPipeline, 0)
	var err error
	if len(envIds) > 0 {
		//filter ci_pipelines based on env list
		pipelines, err = impl.ciPipelineRepository.FindCiPipelineByAppIdAndEnvIds(appId, envIds)
	} else {
		pipelines, err = impl.ciPipelineRepository.FindByAppId(appId)
	}

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows || len(pipelines) == 0 {
		impl.logger.Errorw("no ci pipeline found", "appId", appId, "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 404, UserMessage: "no ci pipeline found"}
		return nil, err
	}
	parentCiPipelines, linkedCiPipelineIds, err := impl.ciPipelineRepository.FindParentCiPipelineMapByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	pipelineParentCiMap := make(map[int]*pipelineConfig.CiPipeline)
	for index, item := range parentCiPipelines {
		pipelineParentCiMap[linkedCiPipelineIds[index]] = item
	}

	var ciPipelineResp []*bean.CiPipelineMin
	for _, pipeline := range pipelines {
		parentCiPipeline := pipelineConfig.CiPipeline{}
		pipelineType := bean.NORMAL

		if pipelineParentCiMap[pipeline.Id] != nil {
			parentCiPipeline = *pipelineParentCiMap[pipeline.Id]
			pipelineType = bean.LINKED
		} else if pipeline.IsExternal == true {
			pipelineType = bean.EXTERNAL
		} else if pipeline.PipelineType == string(bean.CI_JOB) {
			pipelineType = bean.CI_JOB
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

func (impl *CiPipelineConfigServiceImpl) PatchRegexCiPipeline(request *bean.CiRegexPatchRequest) (err error) {
	var materials []*pipelineConfig.CiPipelineMaterial
	for _, material := range request.CiPipelineMaterial {
		materialDbObject, err := impl.ciPipelineMaterialRepository.GetById(material.Id)
		if err != nil {
			impl.logger.Errorw("err in fetching material, ", "err", err)
			return err
		}
		if materialDbObject.Regex != "" {
			if !impl.ciCdPipelineOrchestrator.CheckStringMatchRegex(materialDbObject.Regex, material.Value) {
				impl.logger.Errorw("not matching given regex, ", "err", err)
				return errors.New("not matching given regex")
			}
		}
		pipelineMaterial := &pipelineConfig.CiPipelineMaterial{
			Id:            material.Id,
			Value:         material.Value,
			CiPipelineId:  materialDbObject.CiPipelineId,
			Type:          pipelineConfig.SourceType(material.Type),
			Active:        true,
			GitMaterialId: materialDbObject.GitMaterialId,
			Regex:         materialDbObject.Regex,
			AuditLog:      sql.AuditLog{UpdatedBy: request.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now(), CreatedBy: request.UserId},
		}
		materials = append(materials, pipelineMaterial)
	}

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.ciPipelineMaterialRepository.Update(tx, materials...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = impl.ciCdPipelineOrchestrator.AddPipelineMaterialInGitSensor(materials)
	if err != nil {
		impl.logger.Errorf("error in saving pipelineMaterials in git sensor", "materials", materials, "err", err)
		return err
	}
	return nil
}

func (impl *CiPipelineConfigServiceImpl) GetCiPipelineByEnvironment(request resourceGroup2.ResourceGroupingRequest) ([]*bean.CiConfigRequest, error) {
	ciPipelinesConfigByApps := make([]*bean.CiConfigRequest, 0)
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.ResourceGroupingCiPipelinesAuthorization")
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	if len(request.ResourceIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}

	var appIds []int
	ciPipelineIds := make([]int, 0)
	cdPipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		cdPipelineIds = append(cdPipelineIds, pipeline.Id)
	}

	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	ciPipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
	}
	appResults, _ := request.CheckAuthBatch(request.EmailId, appObjectArr, []string{})
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id]
		if !appResults[appObject[0]] {
			//if user unauthorized, skip items
			continue
		}
		appIds = append(appIds, pipeline.AppId)
		ciPipelineIds = append(ciPipelineIds, pipeline.CiPipelineId)
	}
	//authorization block ends here
	span.End()
	if len(appIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching app found"}
		return nil, err
	}
	if impl.ciConfig.ExternalCiWebhookUrl == "" {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			return nil, err
		}
		if hostUrl != nil {
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, types.ExternalCiWebhookPath)
		}
	}

	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.GetCiTemplateVariables")
	defer span.End()
	ciPipelinesConfigMap, err := impl.getCiTemplateVariablesByAppIds(appIds)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appIds", appIds, "err", err)
		return nil, err
	}

	ciPipelineByApp := make(map[int][]*pipelineConfig.CiPipeline)
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindByAppIds")
	ciPipelines, err := impl.ciPipelineRepository.FindByAppIds(appIds)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "appIds", appIds, "err", err)
		return nil, err
	}
	parentCiPipelineIds := make([]int, 0)
	for _, ciPipeline := range ciPipelines {
		ciPipelineByApp[ciPipeline.AppId] = append(ciPipelineByApp[ciPipeline.AppId], ciPipeline)
		if ciPipeline.ParentCiPipeline > 0 && ciPipeline.IsExternal {
			parentCiPipelineIds = append(parentCiPipelineIds, ciPipeline.ParentCiPipeline)
		}
	}
	pipelineIdVsAppId, err := impl.ciPipelineRepository.FindAppIdsForCiPipelineIds(parentCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching appIds for pipelineIds", "parentCiPipelineIds", parentCiPipelineIds, "err", err)
		return nil, err
	}

	if len(ciPipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching ci pipeline found"}
		return nil, err
	}
	linkedCiPipelinesMap := make(map[int][]*pipelineConfig.CiPipeline)
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindByParentCiPipelineIds")
	linkedCiPipelines, err := impl.ciPipelineRepository.FindByParentCiPipelineIds(ciPipelineIds)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, linkedCiPipeline := range linkedCiPipelines {
		linkedCiPipelinesMap[linkedCiPipeline.ParentCiPipeline] = append(linkedCiPipelinesMap[linkedCiPipeline.Id], linkedCiPipeline)
	}

	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindTemplateOverrideByCiPipelineIds")
	ciTemplateBeanOverrides, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineIds(ciPipelineIds)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching templates override", "appIds", appIds, "err", err)
		return nil, err
	}
	ciOverrideTemplateMap := make(map[int]*bean3.CiTemplateBean)
	for _, templateBeanOverride := range ciTemplateBeanOverrides {
		ciTemplateOverride := templateBeanOverride.CiTemplateOverride
		ciOverrideTemplateMap[ciTemplateOverride.CiPipelineId] = templateBeanOverride
	}

	var externalCiConfig bean.ExternalCiConfig
	//var parentCiPipelineIds []int
	for appId, ciPipelinesConfigByApp := range ciPipelinesConfigMap {
		var ciPipelineResp []*bean.CiPipeline

		ciPipelines := ciPipelineByApp[appId]
		for _, pipeline := range ciPipelines {
			dockerArgs := make(map[string]string)
			if len(pipeline.DockerArgs) > 0 {
				err := json.Unmarshal([]byte(pipeline.DockerArgs), &dockerArgs)
				if err != nil {
					impl.logger.Warnw("error in unmarshal", "err", err)
				}
			}
			parentCiPipelineId := pipeline.ParentCiPipeline
			ciPipeline := &bean.CiPipeline{
				Id:                       pipeline.Id,
				Version:                  pipeline.Version,
				Name:                     pipeline.Name,
				Active:                   pipeline.Active,
				Deleted:                  pipeline.Deleted,
				DockerArgs:               dockerArgs,
				IsManual:                 pipeline.IsManual,
				IsExternal:               pipeline.IsExternal,
				ParentCiPipeline:         parentCiPipelineId,
				ExternalCiConfig:         externalCiConfig,
				ScanEnabled:              pipeline.ScanEnabled,
				IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
				PipelineType:             bean.PipelineType(pipeline.PipelineType),
			}
			parentPipelineAppId, ok := pipelineIdVsAppId[parentCiPipelineId]
			if ok {
				ciPipeline.ParentAppId = parentPipelineAppId
			}
			if ciTemplateBean, ok := ciOverrideTemplateMap[pipeline.Id]; ok {
				templateOverride := ciTemplateBean.CiTemplateOverride
				ciPipeline.DockerConfigOverride = bean.DockerConfigOverride{
					DockerRegistry:   templateOverride.DockerRegistryId,
					DockerRepository: templateOverride.DockerRepository,
					CiBuildConfig:    ciTemplateBean.CiBuildConfig,
				}
			}

			//this will build ci materials for each ci pipeline
			for _, material := range pipeline.CiPipelineMaterials {
				// ignore those materials which have inactive git material
				if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
					continue
				}
				ciMaterial := &bean.CiMaterial{
					Id:              material.Id,
					CheckoutPath:    material.CheckoutPath,
					Path:            material.Path,
					ScmId:           material.ScmId,
					GitMaterialId:   material.GitMaterialId,
					GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
					ScmName:         material.ScmName,
					ScmVersion:      material.ScmVersion,
					IsRegex:         material.Regex != "",
					Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
				}
				ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
			}

			//this will count the length of child ci pipelines, of each ci pipeline
			linkedCi := linkedCiPipelinesMap[pipeline.Id]
			ciPipeline.LinkedCount = len(linkedCi)
			ciPipelineResp = append(ciPipelineResp, ciPipeline)

			//this will use for fetch the parent ci pipeline app id, of each ci pipeline
			//parentCiPipelineIds = append(parentCiPipelineIds, pipeline.ParentCiPipeline)
		}
		ciPipelinesConfigByApp.CiPipelines = ciPipelineResp
		ciPipelinesConfigByApp.CiGitMaterialId = ciPipelinesConfigByApp.CiBuildConfig.GitMaterialId
		ciPipelinesConfigByApps = append(ciPipelinesConfigByApps, ciPipelinesConfigByApp)
	}

	return ciPipelinesConfigByApps, err
}

func (impl *CiPipelineConfigServiceImpl) GetCiPipelineByEnvironmentMin(request resourceGroup2.ResourceGroupingRequest) ([]*bean.CiPipelineMinResponse, error) {
	results := make([]*bean.CiPipelineMinResponse, 0)
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return results, err
		}
		//override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	if len(request.ResourceIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return results, err
	}
	foundAppIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		foundAppIds = append(foundAppIds, pipeline.AppId)
	}
	if len(foundAppIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return nil, err
	}
	ciPipelines, err := impl.ciPipelineRepository.FindByAppIds(foundAppIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching ci pipeline", "err", err)
		return nil, err
	}
	ciPipelineByApp := make(map[int]*pipelineConfig.CiPipeline)
	parentCiPipelineIds := make([]int, 0)
	for _, ciPipeline := range ciPipelines {
		ciPipelineByApp[ciPipeline.Id] = ciPipeline
		if ciPipeline.ParentCiPipeline > 0 && ciPipeline.IsExternal {
			parentCiPipelineIds = append(parentCiPipelineIds, ciPipeline.ParentCiPipeline)
		}
	}
	pipelineIdVsAppId, err := impl.ciPipelineRepository.FindAppIdsForCiPipelineIds(parentCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching appIds for pipelineIds", "parentCiPipelineIds", parentCiPipelineIds, "err", err)
		return nil, err
	}

	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
	}
	appResults, _ := request.CheckAuthBatch(request.EmailId, appObjectArr, []string{})
	authorizedIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id]
		if !appResults[appObject[0]] {
			//if user unauthorized, skip items
			continue
		}
		if pipeline.CiPipelineId == 0 {
			//skip for external ci
			continue
		}
		ciPipeline := ciPipelineByApp[pipeline.CiPipelineId]
		parentAppId := pipelineIdVsAppId[ciPipeline.ParentCiPipeline]
		result := &bean.CiPipelineMinResponse{
			Id:               pipeline.CiPipelineId,
			AppId:            pipeline.AppId,
			AppName:          pipeline.App.AppName,
			ParentCiPipeline: ciPipeline.ParentCiPipeline,
			ParentAppId:      parentAppId,
			PipelineType:     ciPipeline.PipelineType,
		}
		results = append(results, result)
		authorizedIds = append(authorizedIds, pipeline.CiPipelineId)
	}
	//authorization block ends here

	return results, err
}

func (impl *CiPipelineConfigServiceImpl) GetExternalCiByEnvironment(request resourceGroup2.ResourceGroupingRequest) (ciConfig []*bean.ExternalCiConfig, err error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.authorizationExternalCiForResourceGrouping")
	externalCiConfigs := make([]*bean.ExternalCiConfig, 0)
	var cdPipelines []*pipelineConfig.Pipeline
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	if len(request.ResourceIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}

	var appIds []int
	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
	}
	appResults, _ := request.CheckAuthBatch(request.EmailId, appObjectArr, []string{})
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id]
		if !appResults[appObject[0]] {
			//if user unauthorized, skip items
			continue
		}
		//add only those who have external ci
		if pipeline.CiPipelineId == 0 {
			appIds = append(appIds, pipeline.AppId)
		}
	}

	//authorization block ends here
	span.End()

	if len(appIds) == 0 {
		impl.logger.Warnw("there is no app id found for fetching external ci pipelines", "request", request)
		return externalCiConfigs, nil
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindExternalCiByAppIds")
	externalCiPipelines, err := impl.ciPipelineRepository.FindExternalCiByAppIds(appIds)
	span.End()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "request", request, "err", err)
		return nil, err
	}
	hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
	if err != nil {
		impl.logger.Errorw("error in fetching external ci", "request", request, "err", err)
		return nil, err
	}
	if hostUrl != nil {
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, types.ExternalCiWebhookPath)
	}

	var externalCiPipelineIds []int
	appWorkflowMappingsMap := make(map[int][]*appWorkflow.AppWorkflowMapping)

	for _, externalCiPipeline := range externalCiPipelines {
		externalCiPipelineIds = append(externalCiPipelineIds, externalCiPipeline.Id)
	}
	if len(externalCiPipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no external ci pipeline found"}
		return externalCiConfigs, err
	}
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiIdByIdsIn(externalCiPipelineIds)
	if err != nil {
		impl.logger.Errorw("Error in fetching app workflow mapping for CD pipeline by external CI ID", "err", err)
		return nil, err
	}

	CDPipelineMap := make(map[int]*pipelineConfig.Pipeline)
	appIdMap := make(map[int]*app2.App)
	var componentIds []int
	for _, appWorkflowMapping := range appWorkflowMappings {
		appWorkflowMappingsMap[appWorkflowMapping.ParentId] = append(appWorkflowMappingsMap[appWorkflowMapping.ParentId], appWorkflowMapping)
		componentIds = append(componentIds, appWorkflowMapping.ComponentId)
	}
	if len(componentIds) == 0 {
		return nil, err
	}
	cdPipelines, err = impl.pipelineRepository.FindAppAndEnvironmentAndProjectByPipelineIds(componentIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching external ci", "request", request, "err", err)
		return nil, err
	}
	for _, pipeline := range cdPipelines {
		CDPipelineMap[pipeline.Id] = pipeline
		appIds = append(appIds, pipeline.AppId)
	}
	if len(appIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching apps found"}
		return nil, err
	}
	apps, err := impl.appRepo.FindAppAndProjectByIdsIn(appIds)
	for _, app := range apps {
		appIdMap[app.Id] = app
	}

	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "ciHandler.FindAppAndEnvironmentAndProjectByPipelineIds")
	for _, externalCiPipeline := range externalCiPipelines {
		externalCiConfig := &bean.ExternalCiConfig{
			Id:         externalCiPipeline.Id,
			WebhookUrl: fmt.Sprintf("%s/%d", impl.ciConfig.ExternalCiWebhookUrl, externalCiPipeline.Id),
			Payload:    impl.ciConfig.ExternalCiPayload,
			AccessKey:  "",
		}

		if _, ok := appWorkflowMappingsMap[externalCiPipeline.Id]; !ok {
			return nil, errors.New("Error in fetching app workflow mapping for cd pipeline by parent id")
		}
		appWorkflowMappings := appWorkflowMappingsMap[externalCiPipeline.Id]
		roleData := make(map[string]interface{})
		for _, appWorkflowMapping := range appWorkflowMappings {
			if _, ok := CDPipelineMap[appWorkflowMapping.ComponentId]; !ok {
				impl.logger.Errorw("error in getting cd pipeline data for workflow", "app workflow id", appWorkflowMapping.ComponentId, "err", err)
				return nil, errors.New("error in getting cd pipeline data for workflow")
			}
			cdPipeline := CDPipelineMap[appWorkflowMapping.ComponentId]
			if _, ok := roleData[teamIdKey]; !ok {
				if _, ok := appIdMap[cdPipeline.AppId]; !ok {
					impl.logger.Errorw("error in getting app data for pipeline", "app id", cdPipeline.AppId, "err", err)
					return nil, errors.New("error in getting app data for pipeline")
				}
				app := appIdMap[cdPipeline.AppId]
				roleData[teamIdKey] = app.TeamId
				roleData[teamNameKey] = app.Team.Name
				roleData[appIdKey] = cdPipeline.AppId
				roleData[appNameKey] = cdPipeline.App.AppName
			}
			if _, ok := roleData[environmentNameKey]; !ok {
				roleData[environmentNameKey] = cdPipeline.Environment.Name
			} else {
				roleData[environmentNameKey] = fmt.Sprintf("%s,%s", roleData[environmentNameKey], cdPipeline.Environment.Name)
			}
			if _, ok := roleData[environmentIdentifierKey]; !ok {
				roleData[environmentIdentifierKey] = cdPipeline.Environment.EnvironmentIdentifier
			} else {
				roleData[environmentIdentifierKey] = fmt.Sprintf("%s,%s", roleData[environmentIdentifierKey], cdPipeline.Environment.EnvironmentIdentifier)
			}
		}

		externalCiConfig.ExternalCiConfigRole = bean.ExternalCiConfigRole{
			ProjectId:             roleData[teamIdKey].(int),
			ProjectName:           roleData[teamNameKey].(string),
			AppId:                 roleData[appIdKey].(int),
			AppName:               roleData[appNameKey].(string),
			EnvironmentName:       roleData[environmentNameKey].(string),
			EnvironmentIdentifier: roleData[environmentIdentifierKey].(string),
			Role:                  "Build and deploy",
		}
		externalCiConfigs = append(externalCiConfigs, externalCiConfig)
	}
	span.End()
	//--------pipeline population end
	return externalCiConfigs, err
}

func (impl *CiPipelineConfigServiceImpl) DeleteCiPipeline(request *bean.CiPatchRequest) (*bean.CiPipeline, error) {
	ciPipelineId := request.CiPipeline.Id
	//wf validation
	workflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCIPipelineId(ciPipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching workflow mapping for ci validation", "err", err)
		return nil, err
	}
	if len(workflowMapping) > 0 {
		return nil, &util.ApiError{
			InternalMessage:   "Please delete deployment pipelines for this workflow first and try again.",
			UserDetailMessage: fmt.Sprintf("Please delete deployment pipelines for this workflow first and try again."),
			UserMessage:       fmt.Sprintf("Please delete deployment pipelines for this workflow first and try again.")}
	}

	pipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("pipeline fetch err", "id", ciPipelineId, "err", err)
		return nil, err
	}
	appId := request.AppId
	if pipeline.AppId != appId {
		return nil, fmt.Errorf("invalid appid: %d pipelineId: %d mapping", appId, ciPipelineId)
	}

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.ciCdPipelineOrchestrator.DeleteCiPipeline(pipeline, request, tx)
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
	if request.CiPipeline.PreBuildStage != nil && request.CiPipeline.PreBuildStage.Id > 0 {
		//deleting pre stage
		err = impl.pipelineStageService.DeletePipelineStage(request.CiPipeline.PreBuildStage, request.UserId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting pre stage", "err", err, "preBuildStage", request.CiPipeline.PreBuildStage)
			return nil, err
		}
	}
	if request.CiPipeline.PostBuildStage != nil && request.CiPipeline.PostBuildStage.Id > 0 {
		//deleting post stage
		err = impl.pipelineStageService.DeletePipelineStage(request.CiPipeline.PostBuildStage, request.UserId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting post stage", "err", err, "postBuildStage", request.CiPipeline.PostBuildStage)
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

func (impl *CiPipelineConfigServiceImpl) CreateExternalCiAndAppWorkflowMapping(appId, appWorkflowId int, userId int32, tx *pg.Tx) (int, *appWorkflow.AppWorkflowMapping, error) {
	externalCiPipeline := &pipelineConfig.ExternalCiPipeline{
		AppId:       appId,
		AccessToken: "",
		Active:      true,
		AuditLog:    sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
	}
	externalCiPipeline, err := impl.ciPipelineRepository.SaveExternalCi(externalCiPipeline, tx)
	if err != nil {
		impl.logger.Errorw("error in saving external ci", "appId", appId, "err", err)
		return 0, nil, err
	}
	appWorkflowMap := &appWorkflow.AppWorkflowMapping{
		AppWorkflowId: appWorkflowId,
		ComponentId:   externalCiPipeline.Id,
		Type:          "WEBHOOK",
		Active:        true,
		AuditLog:      sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
	}
	appWorkflowMap, err = impl.appWorkflowRepository.SaveAppWorkflowMapping(appWorkflowMap, tx)
	if err != nil {
		impl.logger.Errorw("error in saving app workflow mapping for external ci", "appId", appId, "appWorkflowId", appWorkflowId, "externalCiPipelineId", externalCiPipeline.Id, "err", err)
		return 0, nil, err
	}
	return externalCiPipeline.Id, appWorkflowMap, nil
}
