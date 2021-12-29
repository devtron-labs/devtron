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
	pc "github.com/devtron-labs/devtron/internal/sql/repository/app"
	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
)

type DeploymentTemplateAction interface {
	Execute(template *v1.DeploymentTemplate, props v1.InheritedProps, ctx context.Context) error
}

type DeploymentTemplateActionImpl struct {
	appRepo      pc.AppRepository
	chartService pipeline.ChartService
	logger       *zap.SugaredLogger
}

func NewDeploymentTemplateActionImpl(logger *zap.SugaredLogger, appRepo pc.AppRepository, chartService pipeline.ChartService) *DeploymentTemplateActionImpl {
	return &DeploymentTemplateActionImpl{
		logger:       logger,
		appRepo:      appRepo,
		chartService: chartService,
	}
}

var deploymentTemplateExecutor = []func(impl DeploymentTemplateActionImpl, template *v1.DeploymentTemplate, ctx context.Context) error{executeDeploymentTemplateCreate}

func (impl DeploymentTemplateActionImpl) Execute(template *v1.DeploymentTemplate, props v1.InheritedProps, ctx context.Context) error {
	if template == nil {
		return nil
	}
	err := template.UpdateMissingProps(props)
	if err != nil {
		impl.logger.Errorw("update err", "err", err)
		return fmt.Errorf("unable to update props for deploymentTemplate")
	}
	errs := make([]string, 0)
	for _, f := range deploymentTemplateExecutor {
		errs = util.AppendErrorString(errs, f(impl, template, ctx))
	}
	return util.GetErrorOrNil(errs)
}

func executeDeploymentTemplateCreate(impl DeploymentTemplateActionImpl, template *v1.DeploymentTemplate, ctx context.Context) error {
	if template.Operation != v1.Create {
		return nil
	}
	if template.Destination.App == nil || len(*template.Destination.App) == 0 {
		return fmt.Errorf("app cannot be empty for template creation")
	}
	app, err := impl.appRepo.FindActiveByName(*template.Destination.App)
	if err != nil {
		impl.logger.Errorw("fetch err", "err", err)
		return err
	}
	//TODO: set userId
	valueOverride, err := json.Marshal(template.ValuesOverride)
	if err != nil {
		impl.logger.Errorw("marshal err", "err", err)
		return fmt.Errorf("invalid values for deployment template")
	}
	dTemplate := pipeline.TemplateRequest{
		Id:                  0,
		AppId:               app.Id,
		ValuesOverride:      valueOverride,
		DefaultAppOverride:  nil,
		IsAppMetricsEnabled: template.IsAppMetricsEnabled,
		UserId:              1,
	}
	_, err = impl.chartService.Create(dTemplate, ctx)
	if err != nil {
		impl.logger.Errorw("create err", "err", err)
		return err
	}
	return nil
}
