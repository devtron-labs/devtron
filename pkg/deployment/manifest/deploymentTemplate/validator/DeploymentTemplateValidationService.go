/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/chart/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	pipelineBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/xeipuuv/gojsonschema"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
)

type DeploymentTemplateValidationService interface {
	DeploymentTemplateValidate(ctx context.Context, template interface{}, chartRefId int, scope resourceQualifiers.Scope) (bool, error)
	FlaggerCanaryEnabled(values json.RawMessage) (bool, error)
	ValidateChangeChartRefRequest(ctx context.Context, envConfigProperties *pipelineBean.EnvironmentProperties, request *bean3.ChartRefChangeRequest) (*pipelineBean.EnvironmentProperties, bool, error)
	DeploymentTemplateValidationServiceEnt
}

type DeploymentTemplateValidationServiceImpl struct {
	logger                    *zap.SugaredLogger
	chartRefService           chartRef.ChartRefService
	scopedVariableManager     variables.ScopedVariableManager
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService
	*DeploymentTemplateValidationServiceEntImpl
}

func NewDeploymentTemplateValidationServiceImpl(logger *zap.SugaredLogger,
	chartRefService chartRef.ChartRefService,
	scopedVariableManager variables.ScopedVariableManager,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	deploymentTemplateValidationServiceEntImpl *DeploymentTemplateValidationServiceEntImpl,
) *DeploymentTemplateValidationServiceImpl {
	return &DeploymentTemplateValidationServiceImpl{
		logger:                    logger,
		chartRefService:           chartRefService,
		scopedVariableManager:     scopedVariableManager,
		deployedAppMetricsService: deployedAppMetricsService,
		DeploymentTemplateValidationServiceEntImpl: deploymentTemplateValidationServiceEntImpl,
	}
}

func (impl *DeploymentTemplateValidationServiceImpl) DeploymentTemplateValidate(ctx context.Context, template interface{}, chartRefId int, scope resourceQualifiers.Scope) (bool, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "JsonSchemaExtractFromFile")
	schemajson, version, err := impl.chartRefService.JsonSchemaExtractFromFile(chartRefId)
	span.End()
	if err != nil {
		impl.logger.Errorw("Json Schema not found err, FindJsonSchema", "err", err)
		return true, nil
	}

	templateBytes := template.(json.RawMessage)
	templatejsonstring, _, err := impl.scopedVariableManager.ExtractVariablesAndResolveTemplate(scope, string(templateBytes), parsers.JsonVariableTemplate, true, false)
	if err != nil {
		return false, err
	}
	var templatejson interface{}
	err = json.Unmarshal([]byte(templatejsonstring), &templatejson)
	if err != nil {
		impl.logger.Errorw("json Schema parsing error", "err", err)
		return false, err
	}

	schemaLoader := gojsonschema.NewGoLoader(schemajson)
	documentLoader := gojsonschema.NewGoLoader(templatejson)
	marshalTemplatejson, err := json.Marshal(templatejson)
	if err != nil {
		impl.logger.Errorw("json template marshal err, DeploymentTemplateValidate", "err", err)
		return false, err
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "gojsonschema.Validate")
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	span.End()
	if err != nil {
		impl.logger.Errorw("result validate err, DeploymentTemplateValidate", "err", err)
		return false, err
	}
	if result.Valid() {
		var dat map[string]interface{}
		if err := json.Unmarshal(marshalTemplatejson, &dat); err != nil {
			impl.logger.Errorw("json template unmarshal err, DeploymentTemplateValidate", "err", err)
			return false, err
		}
		_, err := util2.CompareLimitsRequests(dat, version)
		if err != nil {
			impl.logger.Errorw("LimitRequestCompare err, DeploymentTemplateValidate", "err", err)
			return false, err
		}
		_, err = util2.AutoScale(dat)
		if err != nil {
			impl.logger.Errorw("LimitRequestCompare err, DeploymentTemplateValidate", "err", err)
			return false, err
		}
		return true, nil
	} else {
		var stringerror string
		for _, err := range result.Errors() {
			impl.logger.Errorw("result err, DeploymentTemplateValidate", "err", err.Details())
			if err.Details()["format"] == bean.Cpu {
				stringerror = stringerror + err.Field() + ": Format should be like " + bean.CpuPattern + "\n"
			} else if err.Details()["format"] == bean.Memory {
				stringerror = stringerror + err.Field() + ": Format should be like " + bean.MemoryPattern + "\n"
			} else {
				stringerror = stringerror + err.String() + "\n"
			}
		}
		return false, errors.New(stringerror)
	}
}

func (impl *DeploymentTemplateValidationServiceImpl) FlaggerCanaryEnabled(values json.RawMessage) (bool, error) {
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(values, &jsonMap); err != nil {
		return false, err
	}

	flaggerCanary, found := jsonMap[bean.FlaggerCanary]
	if !found {
		return false, nil
	}
	var flaggerCanaryUnmarshalled map[string]json.RawMessage
	if err := json.Unmarshal(flaggerCanary, &flaggerCanaryUnmarshalled); err != nil {
		return false, err
	}
	enabled, found := flaggerCanaryUnmarshalled[bean.EnabledFlag]
	if !found {
		return true, fmt.Errorf("flagger canary enabled field must be set and be equal to false")
	}
	return string(enabled) == bean.TrueFlag, nil
}

func (impl *DeploymentTemplateValidationServiceImpl) ValidateChangeChartRefRequest(ctx context.Context, envConfigProperties *pipelineBean.EnvironmentProperties, request *bean3.ChartRefChangeRequest) (*pipelineBean.EnvironmentProperties, bool, error) {
	compatible, chartChangeType := impl.chartRefService.ChartRefIdsCompatible(envConfigProperties.ChartRefId, request.TargetChartRefId)
	if !compatible {
		errMsg := fmt.Sprintf("chart not compatible for appId: %d, envId: %d", request.AppId, request.EnvId)
		return envConfigProperties, false, util.NewApiError(http.StatusUnprocessableEntity, errMsg, errMsg)
	}
	if !chartChangeType.IsFlaggerCanarySupported() {
		enabled, err := impl.FlaggerCanaryEnabled(envConfigProperties.EnvOverrideValues)
		if err != nil {
			impl.logger.Errorw("error in checking flaggerCanary, ChangeChartRef", "err", err, "payload", envConfigProperties.EnvOverrideValues)
			return envConfigProperties, false, err
		} else if enabled {
			impl.logger.Errorw("rollout charts do not support flaggerCanary, ChangeChartRef", "templateValues", envConfigProperties.EnvOverrideValues)
			errMsg := fmt.Sprintf("%q charts do not support flaggerCanary", chartChangeType.NewChartType)
			return envConfigProperties, false, util.NewApiError(http.StatusUnprocessableEntity, errMsg, errMsg)
		}
	}
	var err error
	envConfigProperties.EnvOverrideValues, err = impl.chartRefService.PerformChartSpecificPatchForSwitch(envConfigProperties.EnvOverrideValues, chartChangeType)
	if err != nil {
		impl.logger.Errorw("error in chart specific patch operations, ValidateChangeChartRefRequest", "err", err, "payload", request)
		return envConfigProperties, false, err
	}
	envMetrics, err := impl.deployedAppMetricsService.GetMetricsFlagByAppIdAndEnvId(request.AppId, request.EnvId)
	if err != nil {
		impl.logger.Errorw("could not find envMetrics for, ChangeChartRef", "err", err, "payload", request)
		errMsg := fmt.Sprintf("could not find envMetrics for appId: %d, envId: %d", request.AppId, request.EnvId)
		return envConfigProperties, false, util.NewApiError(http.StatusBadRequest, errMsg, err.Error())
	}
	envConfigProperties.ChartRefId = request.TargetChartRefId
	envConfigProperties.EnvironmentId = request.EnvId
	envConfigProperties.AppMetrics = &envMetrics
	// VARIABLE_RESOLVE
	scope := resourceQualifiers.Scope{
		AppId:     request.AppId,
		EnvId:     request.EnvId,
		ClusterId: envConfigProperties.ClusterId,
	}
	validate, err2 := impl.DeploymentTemplateValidate(ctx, envConfigProperties.EnvOverrideValues, envConfigProperties.ChartRefId, scope)
	if !validate {
		impl.logger.Errorw("validation err, UpdateAppOverride", "err", err2, "payload", request)
		errMsg := fmt.Sprintf("template schema validation error for appId: %d, envId: %d", request.AppId, request.EnvId)
		return envConfigProperties, envMetrics, util.NewApiError(http.StatusBadRequest, errMsg, err2.Error())
	}
	return envConfigProperties, envMetrics, nil
}
