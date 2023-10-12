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
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	errors2 "github.com/juju/errors"
	"go.opentelemetry.io/otel"
	"sort"
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
}
type CiMaterialConfigService interface {
	//CreateMaterialsForApp : Delegating the request to ciCdPipelineOrchestrator for Material creation
	CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error)
	//UpdateMaterialsForApp : Delegating the request to ciCdPipelineOrchestrator for updating Material
	UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error)
	DeleteMaterial(request *bean.UpdateMaterialDTO) error
	//PatchCiMaterialSource : Delegating the request to ciCdPipelineOrchestrator for updating source
	PatchCiMaterialSource(ciPipeline *bean.CiMaterialPatchRequest, userId int32) (*bean.CiMaterialPatchRequest, error)
	//BulkPatchCiMaterialSource : Delegating the request to ciCdPipelineOrchestrator for bulk updating source
	BulkPatchCiMaterialSource(ciPipelines *bean.CiMaterialBulkPatchRequest, userId int32, token string, checkAppSpecificAccess func(token, action string, appId int) (bool, error)) (*bean.CiMaterialBulkPatchResponse, error)
	//GetMaterialsForAppId : Retrieve material for given appId
	GetMaterialsForAppId(appId int) []*bean.GitMaterial
}
type AppArtifactManager interface {
	//RetrieveArtifactsByCDPipeline : RetrieveArtifactsByCDPipeline returns all the artifacts for the cd pipeline (pre / deploy / post)
	RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean2.WorkflowType, searchString string, count int, isApprovalNode bool) (*bean.CiArtifactResponse, error)

	//FetchArtifactForRollback :
	FetchArtifactForRollback(cdPipelineId, appId, offset, limit int, app *bean.CreateAppDTO, pipeline *pipelineConfig.Pipeline) (bean.CiArtifactResponse, error)
}

func (impl *PipelineBuilderImpl) GetCiPipeline(appId int) (ciConfig *bean.CiConfigRequest, err error) {
	ciConfig, err = impl.getCiTemplateVariables(appId)
	if err != nil {
		impl.logger.Debugw("error in fetching ci pipeline", "appId", appId, "err", err)
		return nil, err
	}
	app, err := impl.appRepo.FindActiveById(appId)
	if err != nil {
		impl.logger.Debugw("error in fetching app details", "appId", appId, "err", err)
		return nil, err
	}
	isJob := app.AppType == helper.Job
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
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
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
		branchesForCheckingBlockageState := make([]string, 0, len(pipeline.CiPipelineMaterials))
		for _, material := range pipeline.CiPipelineMaterials {
			// ignore those materials which have inactive git material
			if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
				continue
			}
			isRegex := material.Regex != ""
			if !(isRegex && len(material.Value) == 0) { //add branches for all cases except if type regex and branch is not set
				branchesForCheckingBlockageState = append(branchesForCheckingBlockageState, material.Value)
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
			if !isJob {
				err = impl.setCiPipelineBlockageState(ciPipeline, branchesForCheckingBlockageState, false)
				if err != nil {
					impl.logger.Errorw("error in getting blockage state for ci pipeline", "err", err, "ciPipelineId", ciPipeline.Id)
					return nil, err
				}
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

func (impl *PipelineBuilderImpl) GetCiPipelineById(pipelineId int) (ciPipeline *bean.CiPipeline, err error) {
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
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
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
	branchesForCheckingBlockageState := make([]string, 0, len(pipeline.CiPipelineMaterials))
	for _, material := range pipeline.CiPipelineMaterials {
		if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
			continue
		}
		isRegex := material.Regex != ""
		if !(isRegex && len(material.Value) == 0) { //add branches for all cases except if type regex and branch is not set
			branchesForCheckingBlockageState = append(branchesForCheckingBlockageState, material.Value)
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

	appDetails := pipeline.App
	isJob := appDetails != nil && appDetails.AppType == helper.Job
	if !isJob {
		err = impl.setCiPipelineBlockageState(ciPipeline, branchesForCheckingBlockageState, false)
		if err != nil {
			impl.logger.Errorw("error in getting blockage state for ci pipeline", "err", err, "ciPipelineId", ciPipeline.Id)
			return nil, err
		}
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

func (impl *PipelineBuilderImpl) GetTriggerViewCiPipeline(appId int) (*bean.TriggerViewCiConfig, error) {

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

		branchesForCheckingBlockageState := make([]string, 0, len(pipeline.CiPipelineMaterials))
		for _, material := range pipeline.CiPipelineMaterials {
			// ignore those materials which have inactive git material
			if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
				continue
			}
			isRegex := material.Regex != ""
			if !(isRegex && len(material.Value) == 0) { //add branches for all cases except if type regex and branch is not set
				branchesForCheckingBlockageState = append(branchesForCheckingBlockageState, material.Value)
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
				IsRegex:         isRegex,
				Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value, Regex: material.Regex},
			}
			ciPipeline.CiMaterial = append(ciPipeline.CiMaterial, ciMaterial)
		}
		linkedCis, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipeline.Id)
		if err != nil && !util.IsErrNoRows(err) {
			return nil, err
		}
		ciPipeline.LinkedCount = len(linkedCis)
		err = impl.setCiPipelineBlockageState(ciPipeline, branchesForCheckingBlockageState, true)
		if err != nil {
			impl.logger.Errorw("error in getting blockage state for ci pipeline", "err", err, "ciPipelineId", ciPipeline.Id)
			return nil, err
		}
		ciPipelineResp = append(ciPipelineResp, ciPipeline)
	}
	triggerViewCiConfig.CiPipelines = ciPipelineResp
	triggerViewCiConfig.Materials = ciConfig.Materials

	return triggerViewCiConfig, nil
}

func (impl *PipelineBuilderImpl) GetExternalCi(appId int) (ciConfig []*bean.ExternalCiConfig, err error) {
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
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
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

func (impl *PipelineBuilderImpl) GetExternalCiById(appId int, externalCiId int) (ciConfig *bean.ExternalCiConfig, err error) {

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
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
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

func (impl PipelineBuilderImpl) UpdateCiTemplate(updateRequest *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {
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

	err = impl.CiTemplateHistoryService.SaveHistory(ciTemplateBean, "update")

	if err != nil {
		impl.logger.Errorw("error in saving update history for ci template", "error", err)
	}

	return originalCiConf, nil
}

func (impl *PipelineBuilderImpl) PatchCiPipeline(request *bean.CiPatchRequest) (ciConfig *bean.CiConfigRequest, err error) {
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
		impl.logger.Debugw("create patch request")
		ciConfig.CiPipelines = []*bean.CiPipeline{request.CiPipeline} //request.CiPipeline

		res, err := impl.addpipelineToTemplate(ciConfig)
		if err != nil {
			impl.logger.Errorw("error in adding pipeline to template", "ciConf", ciConfig, "err", err)
			return nil, err
		}
		return res, nil
	case bean.UPDATE_SOURCE:
		return impl.updateCiPipelineSourceValue(ciConfig, request.CiPipeline)
	case bean.DELETE:
		pipeline, err := impl.DeleteCiPipeline(request)
		if err != nil {
			return nil, err
		}
		ciConfig.CiPipelines = []*bean.CiPipeline{pipeline}
		return ciConfig, nil
	case bean.UPDATE_PIPELINE:
		return impl.patchCiPipelineUpdateSource(ciConfig, request.CiPipeline)
	default:
		impl.logger.Errorw("unsupported operation ", "op", request.Action)
		return nil, fmt.Errorf("unsupported operation %s", request.Action)
	}

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

func (impl *PipelineBuilderImpl) GetCiPipelineMin(appId int, envIds []int) ([]*bean.CiPipelineMin, error) {
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
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no ci pipeline found"}
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

func (impl *PipelineBuilderImpl) PatchRegexCiPipeline(request *bean.CiRegexPatchRequest) (err error) {
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

func (impl PipelineBuilderImpl) GetCiPipelineByEnvironment(request resourceGroup2.ResourceGroupingRequest) ([]*bean.CiConfigRequest, error) {
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
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
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
			branchesForCheckingBlockageState := make([]string, 0, len(pipeline.CiPipelineMaterials))

			//this will build ci materials for each ci pipeline
			for _, material := range pipeline.CiPipelineMaterials {
				// ignore those materials which have inactive git material
				if material == nil || material.GitMaterial == nil || !material.GitMaterial.Active {
					continue
				}
				isRegex := material.Regex != ""
				if !(isRegex && len(material.Value) == 0) { //add branches for all cases except if type regex and branch is not set
					branchesForCheckingBlockageState = append(branchesForCheckingBlockageState, material.Value)
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

			err = impl.setCiPipelineBlockageState(ciPipeline, branchesForCheckingBlockageState, true)
			if err != nil {
				impl.logger.Errorw("error in getting blockage state for ci pipeline", "err", err, "ciPipelineId", ciPipeline.Id)
				return nil, err
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

func (impl PipelineBuilderImpl) GetCiPipelineByEnvironmentMin(request resourceGroup2.ResourceGroupingRequest) ([]*bean.CiPipelineMinResponse, error) {
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

func (impl PipelineBuilderImpl) GetExternalCiByEnvironment(request resourceGroup2.ResourceGroupingRequest) (ciConfig []*bean.ExternalCiConfig, err error) {
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
		impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
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

func (impl *PipelineBuilderImpl) DeleteCiPipeline(request *bean.CiPatchRequest) (*bean.CiPipeline, error) {
	ciPipelineId := request.CiPipeline.Id
	//wf validation
	workflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCIPipelineId(ciPipelineId)
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

func (impl *PipelineBuilderImpl) CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error) {
	res, err := impl.ciCdPipelineOrchestrator.CreateMaterials(request)
	if err != nil {
		impl.logger.Errorw("error in saving create materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl *PipelineBuilderImpl) UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error) {
	res, err := impl.ciCdPipelineOrchestrator.UpdateMaterial(request)
	if err != nil {
		impl.logger.Errorw("error in updating materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl *PipelineBuilderImpl) DeleteMaterial(request *bean.UpdateMaterialDTO) error {
	//finding ci pipelines for this app; if found any, will not delete git material
	pipelines, err := impl.ciPipelineRepository.FindByAppId(request.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting git material", "gitMaterial", request.Material, "err", err)
		return err
	}
	if len(pipelines) > 0 {
		//pipelines are present, in this case we will check if this material is used in docker config
		//if it is used, then we won't delete
		ciTemplateBean, err := impl.ciTemplateService.FindByAppId(request.AppId)
		if err != nil && err == errors2.NotFoundf(err.Error()) {
			impl.logger.Errorw("err in getting docker registry", "appId", request.AppId, "err", err)
			return err
		}
		if ciTemplateBean != nil {
			ciTemplate := ciTemplateBean.CiTemplate
			if ciTemplate != nil && ciTemplate.GitMaterialId == request.Material.Id {
				return fmt.Errorf("cannot delete git material, is being used in docker config")
			}
		}
	}
	existingMaterial, err := impl.materialRepo.FindById(request.Material.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "gitMaterial", request.Material)
		return err
	}
	existingMaterial.UpdatedOn = time.Now()
	existingMaterial.UpdatedBy = request.UserId

	err = impl.materialRepo.MarkMaterialDeleted(existingMaterial)
	if err != nil {
		impl.logger.Errorw("error in deleting git material", "gitMaterial", existingMaterial)
		return err
	}

	err = impl.gitMaterialHistoryService.MarkMaterialDeletedAndCreateHistory(existingMaterial)

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	var materials []*pipelineConfig.CiPipelineMaterial
	for _, pipeline := range pipelines {
		materialDbObject, err := impl.ciPipelineMaterialRepository.GetByPipelineIdAndGitMaterialId(pipeline.Id, request.Material.Id)
		if err != nil {
			return err
		}
		if len(materialDbObject) == 0 {
			continue
		}
		materialDbObject[0].Active = false
		materials = append(materials, materialDbObject[0])
	}

	if len(materials) == 0 {
		return nil
	}

	err = impl.ciPipelineMaterialRepository.Update(tx, materials...)
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (impl *PipelineBuilderImpl) PatchCiMaterialSource(ciPipeline *bean.CiMaterialPatchRequest, userId int32) (*bean.CiMaterialPatchRequest, error) {
	return impl.ciCdPipelineOrchestrator.PatchCiMaterialSource(ciPipeline, userId)
}

func (impl *PipelineBuilderImpl) BulkPatchCiMaterialSource(ciPipelines *bean.CiMaterialBulkPatchRequest, userId int32, token string, checkAppSpecificAccess func(token, action string, appId int) (bool, error)) (*bean.CiMaterialBulkPatchResponse, error) {
	response := &bean.CiMaterialBulkPatchResponse{}
	var ciPipelineMaterials []*pipelineConfig.CiPipelineMaterial
	for _, appId := range ciPipelines.AppIds {
		ciPipeline := &bean.CiMaterialValuePatchRequest{
			AppId:         appId,
			EnvironmentId: ciPipelines.EnvironmentId,
		}
		ciPipelineMaterial, err := impl.ciCdPipelineOrchestrator.PatchCiMaterialSourceValue(ciPipeline, userId, ciPipelines.Value, token, checkAppSpecificAccess)

		if err == nil {
			ciPipelineMaterial.Type = pipelineConfig.SOURCE_TYPE_BRANCH_FIXED
			ciPipelineMaterials = append(ciPipelineMaterials, ciPipelineMaterial)
		}
		response.Apps = append(response.Apps, bean.CiMaterialPatchResponse{
			AppId:   appId,
			Status:  getPatchStatus(err),
			Message: getPatchMessage(err),
		})
	}
	if len(ciPipelineMaterials) == 0 {
		return response, nil
	}
	if err := impl.ciCdPipelineOrchestrator.UpdateCiPipelineMaterials(ciPipelineMaterials); err != nil {
		return nil, err
	}
	return response, nil
}

func (impl *PipelineBuilderImpl) GetMaterialsForAppId(appId int) []*bean.GitMaterial {
	materials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching materials", "appId", appId, "err", err)
	}

	ciTemplateBean, err := impl.ciTemplateService.FindByAppId(appId)
	if err != nil && err != errors2.NotFoundf(err.Error()) {
		impl.logger.Errorw("err in getting ci-template", "appId", appId, "err", err)
	}

	var gitMaterials []*bean.GitMaterial
	for _, material := range materials {
		gitMaterial := &bean.GitMaterial{
			Url:             material.Url,
			Name:            material.Name[strings.Index(material.Name, "-")+1:],
			Id:              material.Id,
			GitProviderId:   material.GitProviderId,
			CheckoutPath:    material.CheckoutPath,
			FetchSubmodules: material.FetchSubmodules,
			FilterPattern:   material.FilterPattern,
		}
		//check if git material is deletable or not
		if ciTemplateBean != nil {
			ciTemplate := ciTemplateBean.CiTemplate
			if ciTemplate != nil && (ciTemplate.GitMaterialId == material.Id || ciTemplate.BuildContextGitMaterialId == material.Id) {
				gitMaterial.IsUsedInCiConfig = true
			}
		}
		gitMaterials = append(gitMaterials, gitMaterial)
	}
	return gitMaterials
}

func (impl *PipelineBuilderImpl) RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean2.WorkflowType, searchString string, count int, isApprovalNode bool) (*bean.CiArtifactResponse, error) {

	// retrieve parent details
	parentId, parentType, err := impl.RetrieveParentDetails(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("failed to retrieve parent details",
			"cdPipelineId", pipeline.Id,
			"err", err)
		return nil, err
	}

	parentCdId := 0
	if parentType == bean2.CD_WORKFLOW_TYPE_POST || (parentType == bean2.CD_WORKFLOW_TYPE_DEPLOY && stage != bean2.CD_WORKFLOW_TYPE_POST) {
		// parentCdId is being set to store the artifact currently deployed on parent cd (if applicable).
		// Parent component is CD only if parent type is POST/DEPLOY
		parentCdId = parentId
	}

	if stage == bean2.CD_WORKFLOW_TYPE_DEPLOY && len(pipeline.PreStageConfig) > 0 {
		// Parent type will be PRE for DEPLOY stage
		parentId = pipeline.Id
		parentType = bean2.CD_WORKFLOW_TYPE_PRE
	}
	if stage == bean2.CD_WORKFLOW_TYPE_POST {
		// Parent type will be DEPLOY for POST stage
		parentId = pipeline.Id
		parentType = bean2.CD_WORKFLOW_TYPE_DEPLOY
	}

	// Build artifacts for cd stages
	var ciArtifacts []bean.CiArtifactBean
	ciArtifactsResponse := &bean.CiArtifactResponse{}

	artifactMap := make(map[int]int)
	limit := count

	ciArtifacts, artifactMap, latestWfArtifactId, latestWfArtifactStatus, err := impl.
		BuildArtifactsForCdStage(pipeline.Id, stage, ciArtifacts, artifactMap, false, searchString, limit, parentCdId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for child cd stage", "err", err, "stage", stage)
		return nil, err
	}

	ciArtifacts, err = impl.BuildArtifactsForParentStage(pipeline.Id, parentId, parentType, ciArtifacts, artifactMap, searchString, limit, parentCdId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for cd", "err", err, "parentStage", parentType, "stage", stage)
		return nil, err
	}

	//sorting ci artifacts on the basis of creation time
	if ciArtifacts != nil {
		sort.SliceStable(ciArtifacts, func(i, j int) bool {
			return ciArtifacts[i].Id > ciArtifacts[j].Id
		})
	}

	artifactIds := make([]int, 0, len(ciArtifacts))
	for _, artifact := range ciArtifacts {
		artifactIds = append(artifactIds, artifact.Id)
	}

	artifacts, err := impl.ciArtifactRepository.GetArtifactParentCiAndWorkflowDetailsByIdsInDesc(artifactIds)
	if err != nil {
		return ciArtifactsResponse, err
	}
	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(pipeline.AppId)
	if err != nil {
		impl.logger.Errorw("error in getting image tagging data with appId", "err", err, "appId", pipeline.AppId)
		return ciArtifactsResponse, err
	}

	imageCommentsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error in getting GetImageCommentsDataMapByArtifactIds", "err", err, "appId", pipeline.AppId, "artifactIds", artifactIds)
		return ciArtifactsResponse, err
	}

	environment := pipeline.Environment
	scope := resourceQualifiers.Scope{AppId: pipeline.AppId, ProjectId: pipeline.App.TeamId, EnvId: pipeline.EnvironmentId, ClusterId: environment.ClusterId, IsProdEnv: environment.Default}
	filters, err := impl.resourceFilterService.GetFiltersByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "pipelineId", pipeline.Id, "err", err)
		return ciArtifactsResponse, err
	}

	for i, artifact := range artifacts {
		imageTaggingResp := imageTagsDataMap[ciArtifacts[i].Id]
		if imageTaggingResp != nil {
			ciArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[ciArtifacts[i].Id]; imageCommentResp != nil {
			ciArtifacts[i].ImageComment = imageCommentResp
		}

		releaseTags := make([]string, 0, len(imageTaggingResp))
		for _, imageTag := range imageTaggingResp {
			releaseTags = append(releaseTags, imageTag.TagName)
		}

		params := impl.celService.GetParamsFromArtifact(ciArtifacts[i].Image, releaseTags)
		metadata := resourceFilter.ExpressionMetadata{
			Params: params,
		}
		filterState, _, err := impl.resourceFilterService.CheckForResource(filters, metadata)
		if err != nil {
			return ciArtifactsResponse, err
		}
		ciArtifacts[i].FilterState = filterState

		if artifact.ExternalCiPipelineId != 0 {
			// if external webhook continue
			continue
		}

		var ciWorkflow *pipelineConfig.CiWorkflow
		if artifact.ParentCiArtifact != 0 {
			ciWorkflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowGitTriggersByArtifactId(artifact.ParentCiArtifact)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}

		} else {
			ciWorkflow, err = impl.ciWorkflowRepository.FindCiWorkflowGitTriggersById(*artifact.WorkflowId)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}
		}
		ciArtifacts[i].TriggeredBy = ciWorkflow.TriggeredBy
		ciArtifacts[i].CiConfigureSourceType = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceType
		ciArtifacts[i].CiConfigureSourceValue = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceValue
	}
	ciArtifactsResponse.ResourceFilters = filters
	ciArtifactsResponse.CdPipelineId = pipeline.Id
	ciArtifactsResponse.LatestWfArtifactId = latestWfArtifactId
	ciArtifactsResponse.LatestWfArtifactStatus = latestWfArtifactStatus
	if ciArtifacts == nil {
		ciArtifacts = []bean.CiArtifactBean{}
	}
	ciArtifactsResponse.CiArtifacts = ciArtifacts

	if pipeline.ApprovalNodeConfigured() && stage == bean2.CD_WORKFLOW_TYPE_DEPLOY { // for now, we are checking artifacts for deploy stage only
		ciArtifactsFinal, approvalConfig, err := impl.overrideArtifactsWithUserApprovalData(pipeline, ciArtifactsResponse.CiArtifacts, isApprovalNode, latestWfArtifactId)
		if err != nil {
			return ciArtifactsResponse, err
		}
		ciArtifactsResponse.UserApprovalConfig = &approvalConfig
		ciArtifactsResponse.CiArtifacts = ciArtifactsFinal
	}
	return ciArtifactsResponse, nil
}

func (impl *PipelineBuilderImpl) FetchArtifactForRollback(cdPipelineId, appId, offset, limit int, app *bean.CreateAppDTO, deploymentPipeline *pipelineConfig.Pipeline) (bean.CiArtifactResponse, error) {
	var deployedCiArtifacts []bean.CiArtifactBean
	var deployedCiArtifactsResponse bean.CiArtifactResponse
	var pipeline *pipelineConfig.Pipeline

	cdWfrs, err := impl.cdWorkflowRepository.FetchArtifactsByCdPipelineId(cdPipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY, offset, limit)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for rollback by cdPipelineId", "err", err, "cdPipelineId", cdPipelineId)
		return deployedCiArtifactsResponse, err
	}
	var ids []int32
	for _, item := range cdWfrs {
		ids = append(ids, item.TriggeredBy)
		if pipeline == nil && item.CdWorkflow != nil {
			pipeline = item.CdWorkflow.Pipeline
		}
	}
	userEmails := make(map[int32]string)
	users, err := impl.userService.GetByIds(ids)
	if err != nil {
		impl.logger.Errorw("unable to fetch users by ids", "err", err, "ids", ids)
	}
	for _, item := range users {
		userEmails[item.Id] = item.EmailId
	}

	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting image tagging data with appId", "err", err, "appId", appId)
		return deployedCiArtifactsResponse, err
	}
	artifactIds := make([]int, 0)

	for _, cdWfr := range cdWfrs {
		ciArtifact := &repository.CiArtifact{}
		if cdWfr.CdWorkflow != nil && cdWfr.CdWorkflow.CiArtifact != nil {
			ciArtifact = cdWfr.CdWorkflow.CiArtifact
		}
		if ciArtifact == nil {
			continue
		}
		mInfo, err := parseMaterialInfo([]byte(ciArtifact.MaterialInfo), ciArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("error in parsing ciArtifact material info", "err", err, "ciArtifact", ciArtifact)
		}
		userEmail := userEmails[cdWfr.TriggeredBy]
		deployedCiArtifacts = append(deployedCiArtifacts, bean.CiArtifactBean{
			Id:           ciArtifact.Id,
			Image:        ciArtifact.Image,
			MaterialInfo: mInfo,
			DeployedTime: formatDate(cdWfr.StartedOn, bean.LayoutRFC3339),
			WfrId:        cdWfr.Id,
			DeployedBy:   userEmail,
		})
		artifactIds = append(artifactIds, ciArtifact.Id)
	}
	imageCommentsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error in getting GetImageCommentsDataMapByArtifactIds", "err", err, "appId", appId, "artifactIds", artifactIds)
		return deployedCiArtifactsResponse, err
	}

	scope := resourceQualifiers.Scope{AppId: app.Id, EnvId: deploymentPipeline.EnvironmentId, ClusterId: deploymentPipeline.Environment.ClusterId, ProjectId: app.TeamId, IsProdEnv: deploymentPipeline.Environment.Default}
	impl.logger.Infow("scope for rollback deployment ", "scope", scope)
	filters, err := impl.resourceFilterService.GetFiltersByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "pipelineId", pipeline.Id, "err", err)
		return deployedCiArtifactsResponse, err
	}

	for i, _ := range deployedCiArtifacts {
		imageTaggingResp := imageTagsDataMap[deployedCiArtifacts[i].Id]
		if imageTaggingResp != nil {
			deployedCiArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[deployedCiArtifacts[i].Id]; imageCommentResp != nil {
			deployedCiArtifacts[i].ImageComment = imageCommentResp
		}
		releaseTags := make([]string, 0, len(imageTaggingResp))
		for _, imageTag := range imageTaggingResp {
			releaseTags = append(releaseTags, imageTag.TagName)
		}

		params := impl.celService.GetParamsFromArtifact(deployedCiArtifacts[i].Image, releaseTags)
		metadata := resourceFilter.ExpressionMetadata{
			Params: params,
		}
		filterState, _, err := impl.resourceFilterService.CheckForResource(filters, metadata)
		if err != nil {
			return deployedCiArtifactsResponse, err
		}
		deployedCiArtifacts[i].FilterState = filterState
	}
	deployedCiArtifactsResponse.ResourceFilters = filters
	deployedCiArtifactsResponse.CdPipelineId = cdPipelineId
	if deployedCiArtifacts == nil {
		deployedCiArtifacts = []bean.CiArtifactBean{}
	}
	if pipeline != nil && pipeline.ApprovalNodeConfigured() {
		deployedCiArtifacts, _, err = impl.overrideArtifactsWithUserApprovalData(pipeline, deployedCiArtifacts, false, 0)
		if err != nil {
			return deployedCiArtifactsResponse, err
		}
	}
	deployedCiArtifactsResponse.CiArtifacts = deployedCiArtifacts

	return deployedCiArtifactsResponse, nil
}

func (impl *PipelineBuilderImpl) setCiPipelineBlockageState(ciPipeline *bean.CiPipeline, branchesForCheckingBlockageState []string, toOnlyGetBlockedStatePolicies bool) error {
	isOffendingMandatoryPlugin, isCiTriggerBlocked, blockageState, err :=
		impl.globalPolicyService.GetBlockageStateForACIPipelineTrigger(ciPipeline.Id, ciPipeline.ParentCiPipeline, branchesForCheckingBlockageState, toOnlyGetBlockedStatePolicies)
	if err != nil {
		impl.logger.Errorw("error in getting blockage state for ci pipeline", "err", err, "ciPipelineId", ciPipeline.Id)
		return err
	}
	if toOnlyGetBlockedStatePolicies {
		ciPipeline.IsCITriggerBlocked = &isCiTriggerBlocked
	} else {
		ciPipeline.IsOffendingMandatoryPlugin = &isOffendingMandatoryPlugin
	}
	ciPipeline.CiBlockState = blockageState
	return nil
}

func (impl PipelineBuilderImpl) updateCiPipelineSourceValue(baseCiConfig *bean.CiConfigRequest, modifiedCiPipeline *bean.CiPipeline) (ciConfig *bean.CiConfigRequest, err error) {
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
		return nil, fmt.Errorf("update of plugin scm material not supported")
	} else {
		dbConnection := impl.pipelineRepository.GetConnection()
		tx, err := dbConnection.Begin()
		if err != nil {
			return nil, err
		}
		// Rollback tx on error.
		defer tx.Rollback()
		var materialsUpdate []*pipelineConfig.CiPipelineMaterial
		for _, material := range modifiedCiPipeline.CiMaterial {
			pipelineMaterial := &pipelineConfig.CiPipelineMaterial{
				Id:       material.Id,
				Value:    material.Source.Value,
				Active:   true,
				AuditLog: sql.AuditLog{UpdatedBy: baseCiConfig.UserId, UpdatedOn: time.Now()},
			}
			if material.Id == 0 {
				continue
			} else {
				materialsUpdate = append(materialsUpdate, pipelineMaterial)
			}
		}
		if len(materialsUpdate) > 0 {
			//using update not null
			err = impl.ciPipelineMaterialRepository.UpdateNotNull(tx, materialsUpdate...)
			if err != nil {
				return nil, err
			}
		}
		err = tx.Commit()
		if err != nil {
			impl.logger.Errorw("error in committing transaction", "err", err)
			return nil, err
		}
		if !modifiedCiPipeline.IsExternal {
			err = impl.ciCdPipelineOrchestrator.AddPipelineMaterialInGitSensor(materialsUpdate)
			if err != nil {
				impl.logger.Errorw("error in saving pipelineMaterials in git sensor", "materials", materialsUpdate, "err", err)
				return nil, err
			}
		}
		modifiedCiPipeline.ScanEnabled = baseCiConfig.ScanEnabled
		baseCiConfig.CiPipelines = append(baseCiConfig.CiPipelines, modifiedCiPipeline)
		return baseCiConfig, nil
	}

}

func (impl PipelineBuilderImpl) overrideArtifactsWithUserApprovalData(pipeline *pipelineConfig.Pipeline, inputArtifacts []bean.CiArtifactBean, isApprovalNode bool, latestArtifactId int) ([]bean.CiArtifactBean, pipelineConfig.UserApprovalConfig, error) {
	impl.logger.Infow("approval node configured", "pipelineId", pipeline.Id, "isApproval", isApprovalNode)
	ciArtifactsFinal := make([]bean.CiArtifactBean, 0, len(inputArtifacts))
	artifactIds := make([]int, 0, len(inputArtifacts))
	cdPipelineId := pipeline.Id
	approvalConfig, err := pipeline.GetApprovalConfig()
	if err != nil {
		impl.logger.Errorw("failed to unmarshal userApprovalConfig", "err", err, "cdPipelineId", cdPipelineId, "approvalConfig", approvalConfig)
		return ciArtifactsFinal, approvalConfig, err
	}

	for _, item := range inputArtifacts {
		artifactIds = append(artifactIds, item.Id)
	}

	var userApprovalMetadata map[int]*pipelineConfig.UserApprovalMetadata
	requiredApprovals := approvalConfig.RequiredCount
	userApprovalMetadata, err = impl.workflowDagExecutor.FetchApprovalDataForArtifacts(artifactIds, cdPipelineId, requiredApprovals) // it will fetch all the request data with nil cd_wfr_rnr_id
	if err != nil {
		impl.logger.Errorw("error occurred while fetching approval data for artifacts", "cdPipelineId", cdPipelineId, "artifactIds", artifactIds, "err", err)
		return ciArtifactsFinal, approvalConfig, err
	}
	for _, artifact := range inputArtifacts {
		approvalRuntimeState := pipelineConfig.InitApprovalState
		approvalMetadataForArtifact, ok := userApprovalMetadata[artifact.Id]
		if ok { // either approved or requested
			approvalRuntimeState = approvalMetadataForArtifact.ApprovalRuntimeState
			artifact.UserApprovalMetadata = approvalMetadataForArtifact
		} else if artifact.Deployed {
			approvalRuntimeState = pipelineConfig.ConsumedApprovalState
		}

		allowed := false
		if isApprovalNode { // return all the artifacts with state in init, requested or consumed
			allowed = approvalRuntimeState == pipelineConfig.InitApprovalState || approvalRuntimeState == pipelineConfig.RequestedApprovalState || approvalRuntimeState == pipelineConfig.ConsumedApprovalState
		} else { // return only approved state artifacts
			allowed = approvalRuntimeState == pipelineConfig.ApprovedApprovalState || artifact.Latest || artifact.Id == latestArtifactId
		}
		if allowed {
			ciArtifactsFinal = append(ciArtifactsFinal, artifact)
		}
	}
	return ciArtifactsFinal, approvalConfig, nil
}
