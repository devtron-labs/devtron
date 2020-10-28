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
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	pc "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"strings"
)

type DeploymentAction interface {
	Execute(deployment *v1.Deployment, props v1.InheritedProps, ctx context.Context) error
}

type DeploymentActionImpl struct {
	logger                   *zap.SugaredLogger
	pipelineBuilder          pipeline.PipelineBuilder
	appRepo                  pc.AppRepository
	envService               cluster.EnvironmentService
	appWorkflowRepo          appWorkflow.AppWorkflowRepository
	ciPipelineRepository     pc.CiPipelineRepository
	cdPipelineRepository     pc.PipelineRepository
	dataHolderAction         DataHolderAction
	deploymentTemplateAction DeploymentTemplateAction
}

func NewDeploymentActionImpl(pipelineBuilder pipeline.PipelineBuilder, logger *zap.SugaredLogger,
	appRepo pc.AppRepository, envService cluster.EnvironmentService, appWorkflowRepo appWorkflow.AppWorkflowRepository,
	ciPipelineRepository pc.CiPipelineRepository, cdPipelineRepository pc.PipelineRepository, dataHolderAction DataHolderAction, deploymentTemplateAction DeploymentTemplateAction) *DeploymentActionImpl {
	return &DeploymentActionImpl{
		pipelineBuilder:          pipelineBuilder,
		appRepo:                  appRepo,
		envService:               envService,
		appWorkflowRepo:          appWorkflowRepo,
		ciPipelineRepository:     ciPipelineRepository,
		cdPipelineRepository:     cdPipelineRepository,
		deploymentTemplateAction: deploymentTemplateAction,
		dataHolderAction:         dataHolderAction,
		logger:                   logger,
	}
}

var deploymentExecutor = []func(impl DeploymentActionImpl, deployment *v1.Deployment, ctx context.Context) error{executeDeploymentCreate}

func (impl DeploymentActionImpl) Execute(deployment *v1.Deployment, props v1.InheritedProps, ctx context.Context) error {
	if deployment == nil {
		return nil
	}
	err := deployment.UpdateMissingProps(props)
	if err != nil {
		return err
	}
	errs := make([]string, 0)
	for _, f := range deploymentExecutor {
		errs = util.AppendErrorString(errs, f(impl, deployment, ctx))
	}
	return util.GetErrorOrNil(errs)
}

func executeDeploymentCreate(impl DeploymentActionImpl, deployment *v1.Deployment, ctx context.Context) error {
	if deployment.Operation != v1.Create {
		return nil
	}
	if deployment.Destination.App == nil || len(*deployment.Destination.App) == 0 {
		return fmt.Errorf("app name cannot be empty in deployment pipeline creation")
	}
	if deployment.Destination.Environment == nil || len(*deployment.Destination.Environment) == 0 {
		return fmt.Errorf("environment cannot be empty in deployment pipeline creation")
	}
	if deployment.Destination.Workflow == nil || len(*deployment.Destination.Workflow) == 0 {
		return fmt.Errorf("workflow cannot be empty in deployment pipeline creation")
	}
	if deployment.Strategy.BlueGreen == nil && deployment.Strategy.Recreate == nil && deployment.Strategy.Rolling == nil && deployment.Strategy.Canary == nil {
		return fmt.Errorf("atleast one deployment strategy should be defined in deployment pipeline creation")
	}
	if len(deployment.Strategy.Default) == 0 {
		return fmt.Errorf("default cannot be empty it should have one of BLUE_GREEN, CANARY, ROLLING, RECREATE")
	}

	app, err := impl.appRepo.FindActiveByName(*deployment.Destination.App)
	if err != nil {
		return fmt.Errorf("error `%s` with appname %s while creating pipeline", err.Error(), *deployment.Destination.App)
	}
	env, err := impl.envService.FindOne(*deployment.Destination.Environment)
	if err != nil {
		return fmt.Errorf("error `%s` with environment %s while creating pipeline", err.Error(), *deployment.Destination.Environment)
	}

	workflow, err := impl.appWorkflowRepo.FindByAppId(app.Id)
	if err != nil {
		return fmt.Errorf("error `%s` finding workflow for application %s while creating deployment pipeline", err.Error(), *deployment.Destination.App)
	}

	var ciPipeline *pc.CiPipeline
	if deployment.PreviousPipeline == nil || deployment.PreviousPipeline.Build == nil || deployment.PreviousPipeline.Build.Destination == nil || deployment.PreviousPipeline.Build.Destination.Pipeline == nil {
		return fmt.Errorf("previous pipeline cannot be empty for deployment")
	}
	if deployment.PreviousPipeline != nil {
		if deployment.PreviousPipeline.Build != nil && deployment.PreviousPipeline.Build.Destination != nil && deployment.PreviousPipeline.Build.Destination.Pipeline != nil {
			previousPipeName := *deployment.PreviousPipeline.Build.Destination.Pipeline
			if !strings.HasPrefix(previousPipeName, *deployment.Destination.App+"-") {
				//change pipeline name to appName-pipelineName
				previousPipeName = fmt.Sprintf("%s-ci-%s", *deployment.Destination.App, previousPipeName)
			}
			ciPipeline, err = impl.ciPipelineRepository.FindByName(previousPipeName)
			if err != nil {
				return fmt.Errorf("error `%s` while finding previous pipeline %s in deployment pipeline creation", err.Error(), *deployment.PreviousPipeline.Build.Destination.Pipeline)
			}
			if ciPipeline.AppId != app.Id {
				return fmt.Errorf("previous pipeline `%s` should belong to same application in deployment pipeline creation", *deployment.PreviousPipeline.Build.Destination.Pipeline)
			}
		}
	}

	pipelineConfig, err := transformToDeploymentConfig(deployment, env, workflow, ciPipeline)
	if err != nil {
		return err
	}

	//TODO: pass userId
	cdPipelines := bean.CdPipelines{
		Pipelines: []*bean.CDPipelineConfigObject{pipelineConfig},
		AppId:     app.Id,
		UserId:    1,
	}

	//TODO: create secrets and configMaps deploymentTemplate
	errs := make([]string, 0)
	for i := range deployment.Secrets {
		errs = util.AppendErrorString(errs, impl.dataHolderAction.Execute(&deployment.Secrets[i], deployment.GetProps(), v1.Secret))
	}

	for i := range deployment.ConfigMaps {
		errs = util.AppendErrorString(errs, impl.dataHolderAction.Execute(&deployment.ConfigMaps[i], deployment.GetProps(), v1.ConfigMap))
	}

	_, err = impl.pipelineBuilder.CreateCdPipelines(&cdPipelines, ctx)
	if err != nil {
		return fmt.Errorf("error %s while creating deployment pipeline for app %s", err.Error(), *deployment.Destination.App)
	}

	errs = util.AppendErrorString(errs, impl.deploymentTemplateAction.Execute(deployment.Template, deployment.GetProps(), ctx))
	return util.GetErrorOrNil(errs)
}

func transformToDeploymentConfig(deployment *v1.Deployment, env *cluster.EnvironmentBean, workflow []*appWorkflow.AppWorkflow, ciPipeline *pc.CiPipeline) (pipelineConfig *bean.CDPipelineConfigObject, err error) {
	pipelineConfig = &bean.CDPipelineConfigObject{}
	var pipelineName string
	if deployment.Destination.Pipeline == nil || len(*deployment.Destination.Pipeline) == 0 {
		n := uuid.NewV4()
		pipelineName = fmt.Sprintf("%s", n)
	} else {
		pipelineName = *deployment.Destination.Pipeline
	}
	if deployment.Trigger == nil || *deployment.Trigger == v1.Automatic {
		pipelineConfig.TriggerType = pc.TRIGGER_TYPE_AUTOMATIC
	} else {
		pipelineConfig.TriggerType = pc.TRIGGER_TYPE_MANUAL
	}
	workflowId := 0
	for _, wf := range workflow {
		if wf.Name == *deployment.Destination.Workflow {
			workflowId = wf.Id
		}
	}
	if workflowId == 0 {
		return nil, fmt.Errorf("incorrect workflow name %s", *deployment.Destination.Workflow)
	}

	strategies, err2 := transformStrategy(deployment)
	if err2 != nil {
		return nil, err2
	}
	pipelineConfig.Name = pipelineName
	pipelineConfig.EnvironmentId = env.Id
	pipelineConfig.EnvironmentName = env.Environment
	pipelineConfig.Namespace = env.Namespace
	pipelineConfig.AppWorkflowId = workflowId
	pipelineConfig.Strategies = strategies
	pipelineConfig.CiPipelineId = ciPipeline.Id

	pipelineConfig.RunPreStageInEnv = deployment.RunPreStageInEnv
	pipelineConfig.RunPostStageInEnv = deployment.RunPostStageInEnv

	if deployment.PreDeployment != nil {
		pipelineConfig.PreStageConfigMapSecretNames = bean.PreStageConfigMapSecretNames{
			ConfigMaps: deployment.PreDeployment.ConfigMaps,
			Secrets:    deployment.PreDeployment.Secrets,
		}
		preCDConf, err := transformStages(deployment.PreDeployment.Stages, "pre")
		if err != nil {
			return nil, fmt.Errorf("invalid precd configuration")
		}

		pipelineConfig.PreStage = bean.CdStage{
			Config: string(preCDConf),
		}
		if deployment.PreDeployment.Trigger == nil || *deployment.PreDeployment.Trigger == v1.Automatic {
			pipelineConfig.PreStage.TriggerType = pc.TRIGGER_TYPE_AUTOMATIC
		} else {
			pipelineConfig.PreStage.TriggerType = pc.TRIGGER_TYPE_MANUAL
		}
	}

	if deployment.PostDeployment != nil {
		pipelineConfig.PostStageConfigMapSecretNames = bean.PostStageConfigMapSecretNames{
			ConfigMaps: deployment.PostDeployment.ConfigMaps,
			Secrets:    deployment.PostDeployment.Secrets,
		}
		postCDConf, err := transformStages(deployment.PostDeployment.Stages, "post")
		if err != nil {
			return nil, fmt.Errorf("invalid postcd configuration")
		}
		pipelineConfig.PostStage = bean.CdStage{
			Config: string(postCDConf),
		}
		if deployment.PostDeployment.Trigger == nil || *deployment.PostDeployment.Trigger == v1.Automatic {
			pipelineConfig.PostStage.TriggerType = pc.TRIGGER_TYPE_AUTOMATIC
		} else {
			pipelineConfig.PostStage.TriggerType = pc.TRIGGER_TYPE_MANUAL
		}
	}

	return pipelineConfig, err
}

func transformStages(deploymentStages []v1.Stage, preOrPost string) ([]byte, error) {
	stages := make([]map[string]string, 0)
	for _, stage := range deploymentStages {
		if len(stage.Name) == 0 || stage.Script == nil || len(*stage.Script) == 0 {
			continue
		}
		m := make(map[string]string, 0)
		m["name"] = stage.Name
		m["script"] = *stage.Script
		if stage.OutputLocation != nil {
			m["outputLocation"] = *stage.OutputLocation
		}
		stages = append(stages, m)
	}
	bs := make(map[string]interface{}, 0)
	if preOrPost == "post" {
		bs["afterStages"] = stages
	} else {
		bs["beforeStages"] = stages
	}
	bss := []map[string]interface{}{bs}
	postCD := make(map[string]interface{}, 0)
	postCD["version"] = "0.0.1"
	postCD["cdPipelineConf"] = bss
	postCDConf, err := yaml.Marshal(postCD)
	return postCDConf, err
}

func transformStrategy(deployment *v1.Deployment) ([]bean.Strategy, error) {
	strategies := make([]bean.Strategy, 0)
	m := make(map[string]interface{}, 0)
	strategy := make(map[string]interface{}, 0)
	m["deployment"] = strategy
	if deployment.Strategy.BlueGreen != nil {
		blueGreen := make(map[string]interface{}, 0)
		strategy["strategy"] = blueGreen
		blueGreen["blueGreen"] = *deployment.Strategy.BlueGreen
		config, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("unable to parse blueGreen strategy")
		}
		strategy := bean.Strategy{
			DeploymentTemplate: "BLUE-GREEN",
			Config:             config,
			Default:            deployment.Strategy.Default == "BLUE-GREEN",
		}
		strategies = append(strategies, strategy)
	}
	if deployment.Strategy.Recreate != nil {
		recreate := make(map[string]interface{}, 0)
		strategy["strategy"] = recreate
		recreate["recreate"] = *deployment.Strategy.Recreate
		config, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("unable to parse recreate strategy")
		}
		strategy := bean.Strategy{
			DeploymentTemplate: "RECREATE",
			Config:             config,
			Default:            deployment.Strategy.Default == "RECREATE",
		}
		strategies = append(strategies, strategy)
	}
	if deployment.Strategy.Canary != nil {
		canary := make(map[string]interface{}, 0)
		strategy["strategy"] = canary
		canary["canary"] = *deployment.Strategy.Canary
		config, err := json.Marshal(m)
		//config, err := json.Marshal(*deployment.Strategy.Canary)
		if err != nil {
			return nil, fmt.Errorf("unable to parse canary strategy")
		}
		strategy := bean.Strategy{
			DeploymentTemplate: "CANARY",
			Config:             config,
			Default:            deployment.Strategy.Default == "CANARY",
		}
		strategies = append(strategies, strategy)
	}
	if deployment.Strategy.Rolling != nil {
		rolling := make(map[string]interface{}, 0)
		strategy["strategy"] = rolling
		rolling["rolling"] = *deployment.Strategy.Rolling
		config, err := json.Marshal(m)
		//config, err := json.Marshal(*deployment.Strategy.Rolling)
		if err != nil {
			return nil, fmt.Errorf("unable to parse rolling strategy")
		}
		strategy := bean.Strategy{
			DeploymentTemplate: "ROLLING",
			Config:             config,
			Default:            deployment.Strategy.Default == "ROLLING",
		}
		strategies = append(strategies, strategy)
	}
	return strategies, nil
}
