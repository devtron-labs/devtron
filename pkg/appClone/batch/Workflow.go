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
	"fmt"
	pc "github.com/devtron-labs/devtron/internal/sql/repository/app"
	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowAction interface {
	Execute(workflow *v1.Workflow, props v1.InheritedProps, ctx context.Context) error
}

type WorkflowActionImpl struct {
	logger           *zap.SugaredLogger
	appRepo          pc.AppRepository
	workflowService  appWorkflow.AppWorkflowService
	buildAction      BuildAction
	deploymentAction DeploymentAction
}

func NewWorkflowActionImpl(logger *zap.SugaredLogger,
	appRepo pc.AppRepository, workflowService appWorkflow.AppWorkflowService, buildAction BuildAction, deploymentAction DeploymentAction) *WorkflowActionImpl {
	return &WorkflowActionImpl{
		appRepo:          appRepo,
		workflowService:  workflowService,
		buildAction:      buildAction,
		deploymentAction: deploymentAction,
		logger:           logger,
	}
}

var workflowExecutor = []func(impl WorkflowActionImpl, workflow *v1.Workflow, ctx context.Context) error{executeWorkflowCreate}

func (impl WorkflowActionImpl) Execute(workflow *v1.Workflow, props v1.InheritedProps, ctx context.Context) error {
	err := workflow.UpdateMissingProps(props)
	if err != nil {
		return err
	}
	errs := make([]string, 0)
	for _, f := range workflowExecutor {
		errs = util.AppendErrorString(errs, f(impl, workflow, ctx))
	}
	return util.GetErrorOrNil(errs)
}

func executeWorkflowCreate(impl WorkflowActionImpl, workflow *v1.Workflow, ctx context.Context) error {
	if workflow.Operation != v1.Create {
		return nil
	}
	if workflow.Destination == nil {
		return fmt.Errorf("destination empty in workflow creation")
	}
	if workflow.Destination.App == nil || len(*workflow.Destination.App) == 0 {
		return fmt.Errorf("app name cannot be empty in workflow creation")
	}
	if workflow.Destination.Workflow == nil || len(*workflow.Destination.Workflow) == 0 {
		return fmt.Errorf("workflow cannot be empty in workflow creation")
	}

	app, err := impl.appRepo.FindActiveByName(*workflow.Destination.App)
	if err != nil {
		return fmt.Errorf("error '%s' for app with name `%s` in workflow creation", err.Error(), *workflow.Destination.App)
	}
	//TODO: add unique check for workflow
	_, err = impl.workflowService.FindAppWorkflowByName(*workflow.Destination.Workflow, app.Id)
	if err == nil {
		return fmt.Errorf("error workflow `%s` exists for app with name `%s` in workflow creation", *workflow.Destination.Workflow, *workflow.Destination.App)
	}
	if pg.ErrNoRows != err {
		return fmt.Errorf("error '%s' workflow `%s` exists for app with name `%s` in workflow creation", err.Error(), *workflow.Destination.Workflow, *workflow.Destination.App)
	}
	//TODO: update userId
	workflowReq := appWorkflow.AppWorkflowDto{
		Name:   *workflow.Destination.Workflow,
		AppId:  app.Id,
		UserId: 1,
	}
	_, err = impl.workflowService.CreateAppWorkflow(workflowReq)
	if err != nil {
		return err
	}
	if workflow.Pipelines == nil {
		return nil
	}

	//first create build pipelines as deployment pipelines will be connected to these
	for _, pipeline := range *workflow.Pipelines {
		if pipeline.Build != nil {
			err = impl.buildAction.Execute(pipeline.Build, workflow.GetProps())
			if err != nil {
				return err
			}
		}
	}

	for _, pipeline := range *workflow.Pipelines {
		if pipeline.Deployment != nil {
			err = impl.deploymentAction.Execute(pipeline.Deployment, workflow.GetProps(), ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
