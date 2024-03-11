package deploymentTemplate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/xeipuuv/gojsonschema"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type DeploymentTemplateValidationService interface {
	DeploymentTemplateValidate(ctx context.Context, template interface{}, chartRefId int, scope resourceQualifiers.Scope) (bool, error)
	FlaggerCanaryEnabled(values json.RawMessage) (bool, error)
}

type DeploymentTemplateValidationServiceImpl struct {
	logger                *zap.SugaredLogger
	chartRefService       chartRef.ChartRefService
	scopedVariableManager variables.ScopedVariableManager
}

func NewDeploymentTemplateValidationServiceImpl(logger *zap.SugaredLogger,
	chartRefService chartRef.ChartRefService,
	scopedVariableManager variables.ScopedVariableManager) *DeploymentTemplateValidationServiceImpl {
	return &DeploymentTemplateValidationServiceImpl{
		logger:                logger,
		chartRefService:       chartRefService,
		scopedVariableManager: scopedVariableManager,
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
		fmt.Println("Error:", err)
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
