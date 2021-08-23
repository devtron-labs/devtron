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
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	pc "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type BuildAction interface {
	Execute(build *v1.Build, props v1.InheritedProps) error
}

type BuildActionImpl struct {
	logger               *zap.SugaredLogger
	pipelineBuilder      pipeline.PipelineBuilder
	appRepo              pc.AppRepository
	appWorkflowRepo      appWorkflow.AppWorkflowRepository
	ciPipelineRepository pc.CiPipelineRepository
	materialRepo         pc.MaterialRepository
}

func NewBuildActionImpl(pipelineBuilder pipeline.PipelineBuilder, logger *zap.SugaredLogger,
	appRepo pc.AppRepository, appWorkflowRepo appWorkflow.AppWorkflowRepository,
	ciPipelineRepository pc.CiPipelineRepository, materialRepo pc.MaterialRepository) *BuildActionImpl {
	return &BuildActionImpl{
		pipelineBuilder:      pipelineBuilder,
		appRepo:              appRepo,
		appWorkflowRepo:      appWorkflowRepo,
		ciPipelineRepository: ciPipelineRepository,
		materialRepo:         materialRepo,
		logger:               logger,
	}
}

var buildExecutor = []func(impl BuildActionImpl, build *v1.Build) error{executeBuildCreate}

func (impl BuildActionImpl) Execute(build *v1.Build, props v1.InheritedProps) error {
	if build == nil {
		return nil
	}
	err := build.UpdateMissingProps(props)
	if err != nil {
		return err
	}
	errs := make([]string, 0)
	for _, f := range buildExecutor {
		errs = util.AppendErrorString(errs, f(impl, build))
	}
	return util.GetErrorOrNil(errs)
}

func executeBuildCreate(impl BuildActionImpl, build *v1.Build) error {
	if build.Operation != v1.Create {
		return nil
	}
	if build.Destination.App == nil || len(*build.Destination.App) == 0 {
		return fmt.Errorf("app name cannot be empty in build pipeline creation")
	}
	if build.Destination.Workflow == nil || len(*build.Destination.Workflow) == 0 {
		return fmt.Errorf("workflow cannot be empty in build pipeline creation")
	}
	app, err := impl.appRepo.FindActiveByName(*build.Destination.App)
	if err != nil {
		return fmt.Errorf("error %s with appname %s while creating pipeline", err.Error(), *build.Destination.App)
	}

	workflow, err := impl.appWorkflowRepo.FindByAppId(app.Id)
	if err != nil {
		return fmt.Errorf("error %s finding workflow for application %s while creating deployment pipeline", err.Error(), *build.Destination.App)
	}
	workflowId := 0
	for _, wf := range workflow {
		if *build.Destination.Workflow == wf.Name {
			workflowId = wf.Id
		}
	}
	if workflowId == 0 {
		return fmt.Errorf("unable to find workflow %s for build", *build.Destination.Workflow)
	}
	dockerArgs := make(map[string]string, len(build.DockerArguments))
	for k, v := range build.DockerArguments {
		dockerArgs[k] = v.(string)
	}

	//If source.pipeline is present in create then it means its an external ci
	var parentCiPipeline *pc.CiPipeline
	var parentCiPipelineId int
	var parentAppId int
	if build.Source != nil {
		if build.Source.Pipeline != nil {
			parentCiPipeline, err = impl.ciPipelineRepository.FindByName(*build.Source.Pipeline)
			if err != nil {
				return fmt.Errorf("parent pipeline %s not found for ci pipeline creation", *build.Source.Pipeline)
			}
			parentCiPipelineId = parentCiPipeline.Id
			parentAppId = parentCiPipeline.AppId
		}
	}

	var pipelineName string
	if build.Destination.Pipeline == nil || len(*build.Destination.Pipeline) == 0 {
		n := uuid.NewV4()
		pipelineName = fmt.Sprintf("%s", n)
	} else {
		pipelineName = *build.Destination.Pipeline
	}

	//populate BeforeDockerBuildScripts
	beforeDockerBuildScripts := transformScripts(build.PreBuild)
	if !allIndexUnique(beforeDockerBuildScripts) {
		return fmt.Errorf("preBuild script doesnt have all unique index")
	}

	//populate AfterDockerBuildScripts
	afterDockerBuildScripts := transformScripts(build.PostBuild)
	if !allIndexUnique(afterDockerBuildScripts) {
		return fmt.Errorf("postBuild script doesnt have all unique index")
	}

	//externalCIConfig
	externalCiConfig := bean.ExternalCiConfig{}
	if build.WebHookUrl != nil && len(*build.WebHookUrl) != 0 {
		externalCiConfig.WebhookUrl = *build.WebHookUrl
	}
	if build.AccessKey != nil && len(*build.AccessKey) != 0 {
		externalCiConfig.AccessKey = *build.AccessKey
	}
	if build.Payload != nil && len(*build.Payload) != 0 {
		externalCiConfig.Payload = *build.Payload
	}

	gitMaterials, err := impl.materialRepo.FindByAppId(app.Id)

	if err != nil {
		return err
	}
	if len(gitMaterials) == 0 {
		return fmt.Errorf("git material not configured for the build")
	}

	ciMaterial := make([]*bean.CiMaterial, 0)
	for _, material := range build.BuildMaterials {
		stc := bean.SourceTypeConfig{
			Value: material.Source.Value,
		}
		if material.Source.Type == v1.BranchFixed {
			stc.Type = pc.SOURCE_TYPE_BRANCH_FIXED
		} else if material.Source.Type == v1.BranchRegex {
			stc.Type = pc.SOURCE_TYPE_BRANCH_REGEX
		} else if material.Source.Type == v1.TagAny {
			stc.Type = pc.SOURCE_TYPE_TAG_ANY
		} else if material.Source.Type == v1.Webhook {
			stc.Type = pc.SOURCE_TYPE_WEBHOOK
		}

		cm := bean.CiMaterial{
			Source: &stc,
		}

		for _, gm := range gitMaterials {
			if gm.Url == material.GitMaterialUrl {
				cm.GitMaterialId = gm.Id
				cm.GitMaterialName = gm.Name
			}
		}
		if cm.GitMaterialId == 0 {
			return fmt.Errorf("git url `%s` is not configured for this application %s", material.GitMaterialUrl, *build.Destination.App)
		}
		ciMaterial = append(ciMaterial, &cm)
	}

	if len(ciMaterial) == 0 {
		return fmt.Errorf("git material not configured for build")
	}

	ciPipeline := bean.CiPipeline{
		IsManual:                 build.Trigger == v1.Manual,
		DockerArgs:               dockerArgs,
		IsExternal:               parentCiPipeline != nil || build.WebHookUrl != nil,
		ParentCiPipeline:         parentCiPipelineId,
		ParentAppId:              parentAppId,
		ExternalCiConfig:         externalCiConfig,
		CiMaterial:               ciMaterial,
		Name:                     pipelineName,
		Active:                   true,
		Deleted:                  false,
		BeforeDockerBuildScripts: beforeDockerBuildScripts,
		AfterDockerBuildScripts:  afterDockerBuildScripts,
		LinkedCount:              0,
	}
	//TODO: add userId
	ciRequest := bean.CiPatchRequest{
		CiPipeline:    &ciPipeline,
		AppId:         app.Id,
		Action:        0,
		AppWorkflowId: workflowId,
		UserId:        1,
	}
	_, err = impl.pipelineBuilder.PatchCiPipeline(&ciRequest)
	if err != nil {
		return fmt.Errorf("unable to create ci pipeline error %s", err.Error())
	}
	return nil
}

func transformScripts(task *v1.Task) []*bean.CiScript {
	prePostDockerBuildScripts := make([]*bean.CiScript, 0)
	if task != nil {
		for i, stage := range task.Stages {
			index := 0
			if stage.Position != nil {
				index = int(*stage.Position)
			} else {
				index = i + 1
			}
			if stage.Script == nil {
				continue
			}
			outputLocation := ""
			if stage.OutputLocation != nil {
				outputLocation = *stage.OutputLocation
			}
			script := bean.CiScript{
				Id:             0,
				Index:          index,
				Name:           stage.Name,
				Script:         *stage.Script,
				OutputLocation: outputLocation,
			}
			prePostDockerBuildScripts = append(prePostDockerBuildScripts, &script)
		}
	}
	return prePostDockerBuildScripts
}

func allIndexUnique(ciScripts []*bean.CiScript) bool {
	m := make(map[int]bool, len(ciScripts))
	for _, script := range ciScripts {
		m[script.Index] = true
	}
	return len(m) == len(ciScripts)
}
