package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/generateManifest"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DefaultMinArtifactCount = 10

type DeploymentHistoryResp struct {
	CdWorkflows                []pipelineConfig.CdWorkflowWithArtifact `json:"cdWorkflows"`
	TagsEdiatable              bool                                    `json:"tagsEditable"`
	AppReleaseTagNames         []string                                `json:"appReleaseTagNames"` //unique list of tags exists in the app
	HideImageTaggingHardDelete bool                                    `json:"hideImageTaggingHardDelete"`
}
type DevtronAppDeploymentRestHandler interface {
	CreateCdPipeline(w http.ResponseWriter, r *http.Request)
	GetCdPipelineById(w http.ResponseWriter, r *http.Request)
	PatchCdPipeline(w http.ResponseWriter, r *http.Request)
	HandleChangeDeploymentRequest(w http.ResponseWriter, r *http.Request)
	HandleChangeDeploymentTypeRequest(w http.ResponseWriter, r *http.Request)
	HandleTriggerDeploymentAfterTypeChange(w http.ResponseWriter, r *http.Request)
	GetCdPipelines(w http.ResponseWriter, r *http.Request)
	GetCdPipelinesForAppAndEnv(w http.ResponseWriter, r *http.Request)

	GetArtifactsByCDPipeline(w http.ResponseWriter, r *http.Request)
	GetArtifactsForRollback(w http.ResponseWriter, r *http.Request)

	UpgradeForAllApps(w http.ResponseWriter, r *http.Request)

	IsReadyToTrigger(w http.ResponseWriter, r *http.Request)
	FetchCdWorkflowDetails(w http.ResponseWriter, r *http.Request)
	GetCdPipelinesByEnvironment(w http.ResponseWriter, r *http.Request)
	GetCdPipelinesByEnvironmentMin(w http.ResponseWriter, r *http.Request)
	PerformDeploymentApprovalAction(w http.ResponseWriter, r *http.Request)
	ChangeChartRef(w http.ResponseWriter, r *http.Request)
}

type DevtronAppDeploymentConfigRestHandler interface {
	ConfigureDeploymentTemplateForApp(w http.ResponseWriter, r *http.Request)
	GetDeploymentTemplate(w http.ResponseWriter, r *http.Request)
	GetDefaultDeploymentTemplate(w http.ResponseWriter, r *http.Request)
	GetAppOverrideForDefaultTemplate(w http.ResponseWriter, r *http.Request)
	GetTemplateComparisonMetadata(w http.ResponseWriter, r *http.Request)
	GetDeploymentTemplateData(w http.ResponseWriter, r *http.Request)

	EnvConfigOverrideCreate(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideUpdate(w http.ResponseWriter, r *http.Request)
	GetEnvConfigOverride(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideReset(w http.ResponseWriter, r *http.Request)

	UpdateAppOverride(w http.ResponseWriter, r *http.Request)
	GetConfigmapSecretsForDeploymentStages(w http.ResponseWriter, r *http.Request)
	GetDeploymentPipelineStrategy(w http.ResponseWriter, r *http.Request)
	GetDefaultDeploymentPipelineStrategy(w http.ResponseWriter, r *http.Request)

	AppMetricsEnableDisable(w http.ResponseWriter, r *http.Request)
	EnvMetricsEnableDisable(w http.ResponseWriter, r *http.Request)

	EnvConfigOverrideCreateNamespace(w http.ResponseWriter, r *http.Request)
}

type DevtronAppPrePostDeploymentRestHandler interface {
	GetMigrationConfig(w http.ResponseWriter, r *http.Request)
	CreateMigrationConfig(w http.ResponseWriter, r *http.Request)
	UpdateMigrationConfig(w http.ResponseWriter, r *http.Request)
	GetStageStatus(w http.ResponseWriter, r *http.Request)
	GetPrePostDeploymentLogs(w http.ResponseWriter, r *http.Request)
	// CancelStage Cancel Pre/Post ArgoWorkflow execution
	CancelStage(w http.ResponseWriter, r *http.Request)
}

type DevtronAppDeploymentHistoryRestHandler interface {
	ListDeploymentHistory(w http.ResponseWriter, r *http.Request)
	DownloadArtifacts(w http.ResponseWriter, r *http.Request)
}

func (handler PipelineConfigRestHandlerImpl) ConfigureDeploymentTemplateForApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var templateRequest chart.TemplateRequest
	err = decoder.Decode(&templateRequest)
	templateRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, ConfigureDeploymentTemplateForApp", "err", err, "payload", templateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId := templateRequest.ChartRefId
	//VARIABLE_RESOLVE
	scope := resourceQualifiers.Scope{
		AppId: templateRequest.AppId,
	}
	validate, err2 := handler.chartService.DeploymentTemplateValidate(r.Context(), templateRequest.ValuesOverride, chartRefId, scope)
	if !validate {
		common.WriteJsonResp(w, err2, nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Infow("request payload, ConfigureDeploymentTemplateForApp", "payload", templateRequest)
	err = handler.validator.Struct(templateRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, ConfigureDeploymentTemplateForApp", "err", err, "payload", templateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(templateRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx = context.WithValue(r.Context(), "token", acdToken)
	createResp, err := handler.chartService.Create(templateRequest, ctx)
	if err != nil {
		handler.Logger.Errorw("service err, ConfigureDeploymentTemplateForApp", "err", err, "payload", templateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateCdPipeline(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var cdPipeline bean.CdPipelines
	err = decoder.Decode(&cdPipeline)
	cdPipeline.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateCdPipeline", "err", err, "payload", cdPipeline)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CreateCdPipeline", "payload", cdPipeline)
	userUploaded, err := handler.chartService.CheckCustomChartByAppId(cdPipeline.AppId)
	if !userUploaded {
		err = handler.validator.Struct(cdPipeline)
		if err != nil {
			handler.Logger.Errorw("validation err, CreateCdPipeline", "err", err, "payload", cdPipeline)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	handler.Logger.Debugw("pipeline create request ", "req", cdPipeline)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(cdPipeline.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if app.AppType == helper.Job {
		common.WriteJsonResp(w, fmt.Errorf("cannot create cd-pipeline for job"), "cannot create cd-pipeline for job", http.StatusBadRequest)
		return
	}
	//RBAC
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	for _, deploymentPipeline := range cdPipeline.Pipelines {
		object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, deploymentPipeline.EnvironmentId)
		handler.Logger.Debugw("Triggered Request By:", "object", object)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	//RBAC
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx := context.WithValue(r.Context(), "token", acdToken)
	createResp, err := handler.pipelineBuilder.CreateCdPipelines(&cdPipeline, ctx)

	if err != nil {
		handler.Logger.Errorw("service err, CreateCdPipeline", "err", err, "payload", cdPipeline)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) PatchCdPipeline(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var cdPipeline bean.CDPatchRequest
	err = decoder.Decode(&cdPipeline)
	cdPipeline.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, PatchCdPipeline", "err", err, "payload", cdPipeline)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	v := r.URL.Query()
	forceDelete := false
	cascadeDelete := true
	force := v.Get("force")
	cascade := v.Get("cascade")
	if len(force) > 0 && len(cascade) > 0 {
		handler.Logger.Errorw("request err, PatchCdPipeline", "err", fmt.Errorf("cannot perform both cascade and force delete"), "payload", cdPipeline)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if len(force) > 0 {
		forceDelete, err = strconv.ParseBool(force)
		if err != nil {
			handler.Logger.Errorw("request err, PatchCdPipeline", "err", err, "payload", cdPipeline)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	} else if len(cascade) > 0 {
		cascadeDelete, err = strconv.ParseBool(cascade)
		if err != nil {
			handler.Logger.Errorw("request err, PatchCdPipeline", "err", err, "payload", cdPipeline)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	cdPipeline.ForceDelete = forceDelete
	cdPipeline.NonCascadeDelete = !cascadeDelete
	handler.Logger.Infow("request payload, PatchCdPipeline", "payload", cdPipeline)
	err = handler.validator.StructPartial(cdPipeline, "AppId", "Action")
	if err == nil {
		if cdPipeline.Action == bean.CD_CREATE {
			err = handler.validator.Struct(cdPipeline.Pipeline)
		} else if cdPipeline.Action == bean.CD_DELETE {
			err = handler.validator.Var(cdPipeline.Pipeline.Id, "gt=0")
		} else if cdPipeline.Action == bean.CD_DELETE_PARTIAL {
			err = handler.validator.Var(cdPipeline.Pipeline.Id, "gt=0")
		}
	}
	if err != nil {
		handler.Logger.Errorw("validation err, PatchCdPipeline", "err", err, "payload", cdPipeline)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(cdPipeline.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	object := handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(cdPipeline.AppId, cdPipeline.Pipeline.Id)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx := context.WithValue(r.Context(), "token", acdToken)
	createResp, err := handler.pipelineBuilder.PatchCdPipelines(&cdPipeline, ctx)
	if err != nil {
		handler.Logger.Errorw("service err, PatchCdPipeline", "err", err, "payload", cdPipeline)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

// HandleChangeDeploymentRequest changes the deployment app type for all pipelines in all apps for a given environment.
func (handler PipelineConfigRestHandlerImpl) HandleChangeDeploymentRequest(w http.ResponseWriter, r *http.Request) {

	// Auth check
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Retrieving and parsing request body
	decoder := json.NewDecoder(r.Body)
	var deploymentAppTypeChangeRequest *bean.DeploymentAppTypeChangeRequest
	err = decoder.Decode(&deploymentAppTypeChangeRequest)
	if err != nil {
		handler.Logger.Errorw("request err, HandleChangeDeploymentRequest", "err", err, "payload",
			deploymentAppTypeChangeRequest)

		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	deploymentAppTypeChangeRequest.UserId = userId

	// Validate incoming request
	err = handler.validator.Struct(deploymentAppTypeChangeRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, HandleChangeDeploymentRequest", "err", err, "payload",
			deploymentAppTypeChangeRequest)

		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Only super-admin access
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// Retrieve argocd token
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx := context.WithValue(r.Context(), "token", acdToken)

	resp, err := handler.pipelineBuilder.ChangeDeploymentType(ctx, deploymentAppTypeChangeRequest)

	if err != nil {
		nErr := errors.New("failed to change deployment type with error msg: " + err.Error())
		handler.Logger.Errorw(err.Error(),
			"payload", deploymentAppTypeChangeRequest,
			"err", err)

		common.WriteJsonResp(w, nErr, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
	return
}

func (handler PipelineConfigRestHandlerImpl) HandleChangeDeploymentTypeRequest(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var deploymentTypeChangeRequest *bean.DeploymentAppTypeChangeRequest
	err = decoder.Decode(&deploymentTypeChangeRequest)
	if err != nil {
		handler.Logger.Errorw("request err, HandleChangeDeploymentTypeRequest", "err", err, "payload",
			deploymentTypeChangeRequest)

		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	deploymentTypeChangeRequest.UserId = userId

	err = handler.validator.Struct(deploymentTypeChangeRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, HandleChangeDeploymentTypeRequest", "err", err, "payload",
			deploymentTypeChangeRequest)

		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx := context.WithValue(r.Context(), "token", acdToken)

	resp, err := handler.pipelineBuilder.ChangePipelineDeploymentType(ctx, deploymentTypeChangeRequest)

	if err != nil {
		handler.Logger.Errorw(err.Error(), "payload", deploymentTypeChangeRequest, "err", err)

		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
	return
}

func (handler PipelineConfigRestHandlerImpl) HandleTriggerDeploymentAfterTypeChange(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var deploymentAppTriggerRequest *bean.DeploymentAppTypeChangeRequest
	err = decoder.Decode(&deploymentAppTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("request err, HandleChangeDeploymentTypeRequest", "err", err, "payload",
			deploymentAppTriggerRequest)

		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	deploymentAppTriggerRequest.UserId = userId

	err = handler.validator.Struct(deploymentAppTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, HandleChangeDeploymentTypeRequest", "err", err, "payload",
			deploymentAppTriggerRequest)

		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")

	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()

	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	ctx := context.WithValue(r.Context(), "token", acdToken)

	resp, err := handler.pipelineBuilder.TriggerDeploymentAfterTypeChange(ctx, deploymentAppTriggerRequest)

	if err != nil {
		handler.Logger.Errorw(err.Error(),
			"payload", deploymentAppTriggerRequest,
			"err", err)

		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
	return
}

func (handler PipelineConfigRestHandlerImpl) ChangeChartRef(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request chart.ChartRefChangeRequest
	err = decoder.Decode(&request)
	if err != nil || request.EnvId == 0 || request.TargetChartRefId == 0 || request.AppId == 0 {
		handler.Logger.Errorw("request err, ChangeChartRef", "err", err, "payload", request)
		common.WriteJsonResp(w, err, request, http.StatusBadRequest)
		return
	}
	envConfigProperties, err := handler.propertiesConfigService.GetLatestEnvironmentProperties(request.AppId, request.EnvId)
	if err != nil || envConfigProperties == nil {
		handler.Logger.Errorw("env properties not found, ChangeChartRef", "err", err, "payload", request)
		common.WriteJsonResp(w, err, "env properties not found", http.StatusNotFound)
		return
	}
	if !envConfigProperties.IsOverride {
		handler.Logger.Errorw("isOverride is not true, ChangeChartRef", "err", err, "payload", request)
		common.WriteJsonResp(w, err, "specific environment is not overriden", http.StatusUnprocessableEntity)
		return
	}
	compatible, oldChartType, newChartType := handler.chartService.ChartRefIdsCompatible(envConfigProperties.ChartRefId, request.TargetChartRefId)
	if !compatible {
		common.WriteJsonResp(w, fmt.Errorf("charts not compatible"), "chart not compatible", http.StatusUnprocessableEntity)
		return
	}

	envConfigProperties.EnvOverrideValues, err = handler.chartService.PatchEnvOverrides(envConfigProperties.EnvOverrideValues, oldChartType, newChartType)
	if err != nil {
		common.WriteJsonResp(w, err, "error in patching env override", http.StatusInternalServerError)
		return
	}

	if newChartType == chart.RolloutChartType {
		enabled, err := handler.chartService.FlaggerCanaryEnabled(envConfigProperties.EnvOverrideValues)
		if err != nil || enabled {
			handler.Logger.Errorw("rollout charts do not support flaggerCanary, ChangeChartRef", "err", err, "payload", request)
			common.WriteJsonResp(w, err, "rollout charts do not support flaggerCanary, ChangeChartRef", http.StatusBadRequest)
			return
		}
	}

	envMetrics, err := handler.propertiesConfigService.FindEnvLevelAppMetricsByAppIdAndEnvId(request.AppId, request.EnvId)
	if err != nil {
		handler.Logger.Errorw("could not find envMetrics for, ChangeChartRef", "err", err, "payload", request)
		common.WriteJsonResp(w, err, "env metric could not be fetched", http.StatusBadRequest)
		return
	}
	envConfigProperties.ChartRefId = request.TargetChartRefId
	envConfigProperties.UserId = userId
	envConfigProperties.EnvironmentId = request.EnvId
	envConfigProperties.AppMetrics = envMetrics.AppMetrics

	token := r.Header.Get("token")
	handler.Logger.Infow("request payload, EnvConfigOverrideCreate", "payload", request)
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(request.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(request.AppId, request.EnvId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	//VARIABLE_RESOLVE
	scope := resourceQualifiers.Scope{
		AppId:     request.AppId,
		EnvId:     request.EnvId,
		ClusterId: envConfigProperties.ClusterId,
	}
	validate, err2 := handler.chartService.DeploymentTemplateValidate(r.Context(), envConfigProperties.EnvOverrideValues, envConfigProperties.ChartRefId, scope)
	if !validate {
		handler.Logger.Errorw("validation err, UpdateAppOverride", "err", err2, "payload", request)
		common.WriteJsonResp(w, err2, "validation err, UpdateAppOverrid", http.StatusBadRequest)
		return
	}
	envConfigPropertiesOld, err := handler.propertiesConfigService.FetchEnvProperties(request.AppId, request.EnvId, request.TargetChartRefId)
	if err == nil {
		envConfigProperties.Id = envConfigPropertiesOld.Id
		createResp, err := handler.propertiesConfigService.UpdateEnvironmentProperties(request.AppId, envConfigProperties, userId)
		if err != nil {
			handler.Logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
			common.WriteJsonResp(w, err, createResp, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, err, createResp, http.StatusOK)
		return
	}
	createResp, err := handler.propertiesConfigService.CreateEnvironmentProperties(request.AppId, envConfigProperties)

	if err != nil {
		if err.Error() == bean2.NOCHARTEXIST {
			ctx, cancel := context.WithCancel(r.Context())
			if cn, ok := w.(http.CloseNotifier); ok {
				go func(done <-chan struct{}, closed <-chan bool) {
					select {
					case <-done:
					case <-closed:
						cancel()
					}
				}(ctx.Done(), cn.CloseNotify())
			}
			acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
			if err != nil {
				handler.Logger.Errorw("error in getting acd token", "err", err)
				common.WriteJsonResp(w, err, "error in getting acd token", http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), "token", acdToken)
			appMetrics := false
			if envConfigProperties.AppMetrics != nil {
				appMetrics = *envMetrics.AppMetrics
			}
			templateRequest := chart.TemplateRequest{
				AppId:               request.AppId,
				ChartRefId:          request.TargetChartRefId,
				ValuesOverride:      []byte("{}"),
				UserId:              userId,
				IsAppMetricsEnabled: appMetrics,
			}

			_, err = handler.chartService.CreateChartFromEnvOverride(templateRequest, ctx)
			if err != nil {
				handler.Logger.Errorw("service err, CreateChartFromEnvOverride", "err", err, "payload", request)
				common.WriteJsonResp(w, err, "could not create chart from env override", http.StatusInternalServerError)
				return
			}
			createResp, err = handler.propertiesConfigService.CreateEnvironmentProperties(request.AppId, envConfigProperties)
			if err != nil {
				handler.Logger.Errorw("service err, CreateEnvironmentProperties", "err", err, "payload", request)
				common.WriteJsonResp(w, err, "could not create env properties", http.StatusInternalServerError)
				return
			}
			common.WriteJsonResp(w, err, createResp, http.StatusOK)
			return
		} else {
			handler.Logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", request)
			common.WriteJsonResp(w, err, "service err, EnvConfigOverrideCreate", http.StatusInternalServerError)
			return
		}
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideCreate(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var envConfigProperties bean3.EnvironmentProperties
	err = decoder.Decode(&envConfigProperties)
	if err != nil {
		handler.Logger.Errorw("request err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envConfigProperties.UserId = userId
	envConfigProperties.EnvironmentId = environmentId
	handler.Logger.Infow("request payload, EnvConfigOverrideCreate", "payload", envConfigProperties)

	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, environmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	protectionEnabled := handler.resourceProtectionService.ResourceProtectionEnabled(appId, environmentId)
	if protectionEnabled {
		handler.Logger.Errorw("resource protection enabled", "appId", appId, "envId", environmentId)
		common.WriteJsonResp(w, fmt.Errorf("resource protection enabled"), "resource protection enabled", http.StatusLocked)
		return
	}

	chartRefId := envConfigProperties.ChartRefId
	//VARIABLE_RESOLVE
	scope := resourceQualifiers.Scope{
		AppId:     appId,
		EnvId:     environmentId,
		ClusterId: envConfigProperties.ClusterId,
	}
	validate, err2 := handler.chartService.DeploymentTemplateValidate(r.Context(), envConfigProperties.EnvOverrideValues, chartRefId, scope)
	if !validate {
		handler.Logger.Errorw("validation err, UpdateAppOverride", "err", err2, "payload", envConfigProperties)
		common.WriteJsonResp(w, err2, nil, http.StatusBadRequest)
		return
	}

	createResp, err := handler.propertiesConfigService.CreateEnvironmentProperties(appId, &envConfigProperties)
	if err != nil {
		if err.Error() == bean2.NOCHARTEXIST {
			ctx, cancel := context.WithCancel(r.Context())
			if cn, ok := w.(http.CloseNotifier); ok {
				go func(done <-chan struct{}, closed <-chan bool) {
					select {
					case <-done:
					case <-closed:
						cancel()
					}
				}(ctx.Done(), cn.CloseNotify())
			}
			acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
			if err != nil {
				handler.Logger.Errorw("error in getting acd token", "err", err)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), "token", acdToken)
			appMetrics := false
			if envConfigProperties.AppMetrics != nil {
				appMetrics = *envConfigProperties.AppMetrics
			}
			templateRequest := chart.TemplateRequest{
				AppId:               appId,
				ChartRefId:          envConfigProperties.ChartRefId,
				ValuesOverride:      []byte("{}"),
				UserId:              userId,
				IsAppMetricsEnabled: appMetrics,
			}

			_, err = handler.chartService.CreateChartFromEnvOverride(templateRequest, ctx)
			if err != nil {
				handler.Logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
			createResp, err = handler.propertiesConfigService.CreateEnvironmentProperties(appId, &envConfigProperties)
			if err != nil {
				handler.Logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
		} else {
			handler.Logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	//userId := getLoggedInUser(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var envConfigProperties bean3.EnvironmentProperties
	err = decoder.Decode(&envConfigProperties)
	envConfigProperties.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvConfigOverrideUpdate", "payload", envConfigProperties)
	err = handler.validator.Struct(envConfigProperties)
	if err != nil {
		handler.Logger.Errorw("validation err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	envConfigOverride, err := handler.propertiesConfigService.GetAppIdByChartEnvId(envConfigProperties.Id)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	appId := envConfigOverride.Chart.AppId
	envId := envConfigOverride.TargetEnvironment
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	protectionEnabled := handler.resourceProtectionService.ResourceProtectionEnabled(appId, envId)
	if protectionEnabled {
		handler.Logger.Errorw("resource protection enabled", "appId", appId, "envId", envId)
		common.WriteJsonResp(w, fmt.Errorf("resource protection enabled"), "resource protection enabled", http.StatusLocked)
		return
	}
	chartRefId := envConfigProperties.ChartRefId
	//VARIABLE_RESOLVE
	scope := resourceQualifiers.Scope{
		AppId:     appId,
		EnvId:     envId,
		ClusterId: envConfigProperties.ClusterId,
	}
	validate, err2 := handler.chartService.DeploymentTemplateValidate(r.Context(), envConfigProperties.EnvOverrideValues, chartRefId, scope)
	if !validate {
		handler.Logger.Errorw("validation err, UpdateAppOverride", "err", err2, "payload", envConfigProperties)
		common.WriteJsonResp(w, err2, nil, http.StatusBadRequest)
		return
	}

	createResp, err := handler.propertiesConfigService.UpdateEnvironmentProperties(appId, &envConfigProperties, userId)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetEnvConfigOverride(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetEnvConfigOverride", "err", err, "payload", appId, environmentId, chartRefId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetEnvConfigOverride", "payload", appId, environmentId, chartRefId)
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	env, err := handler.propertiesConfigService.GetEnvironmentProperties(appId, environmentId, chartRefId)
	if err != nil {
		handler.Logger.Errorw("service err, GetEnvConfigOverride", "err", err, "payload", appId, environmentId, chartRefId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	schema, readme, err := handler.chartService.GetSchemaAndReadmeForTemplateByChartRefId(chartRefId)
	if err != nil {
		handler.Logger.Errorw("err in getting schema and readme, GetEnvConfigOverride", "err", err, "appId", appId, "chartRefId", chartRefId)
	}
	env.Schema = schema
	env.Readme = string(readme)
	common.WriteJsonResp(w, err, env, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetTemplateComparisonMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := r.Header.Get("token")
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	resp, err := handler.deploymentTemplateService.FetchDeploymentsWithChartRefs(appId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchDeploymentsWithChartRefs", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetDeploymentTemplateData(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var request generateManifest.DeploymentTemplateRequest
	err := decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, GetDeploymentTemplate by API", "err", err, "GetYaluesAndManifest", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACNameByAppId(request.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		handler.Logger.Errorw("request err, userId", "err", err, "payload", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isSuperAdmin, _ := handler.userAuthService.IsSuperAdmin(int(userId))

	//RBAC enforcer Ends

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	ctx = util2.SetSuperAdminInContext(ctx, isSuperAdmin)
	defer cancel()
	//TODO fix
	resp, err := handler.deploymentTemplateService.GetDeploymentTemplate(ctx, request)
	if err != nil {
		handler.Logger.Errorw("service err, GetEnvConfigOverride", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetDeploymentTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		handler.Logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetDeploymentTemplate", "appId", appId, "chartRefId", chartRefId)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	appConfigResponse := make(map[string]interface{})
	appConfigResponse["globalConfig"] = nil

	err = handler.chartService.CheckChartExists(chartRefId)
	if err != nil {
		handler.Logger.Errorw("refChartDir Not Found err, JsonSchemaExtractFromFile", err)
		common.WriteJsonResp(w, err, nil, http.StatusForbidden)
		return
	}

	schema, readme, err := handler.chartService.GetSchemaAndReadmeForTemplateByChartRefId(chartRefId)
	if err != nil {
		handler.Logger.Errorw("err in getting schema and readme, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
	}

	template, err := handler.chartService.FindLatestChartForAppByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		handler.Logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	if pg.ErrNoRows == err {
		appOverride, _, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartRefId)
		if err != nil {
			handler.Logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		appOverride["schema"] = json.RawMessage(schema)
		appOverride["readme"] = string(readme)
		mapB, err := json.Marshal(appOverride)
		if err != nil {
			handler.Logger.Errorw("marshal err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			return
		}
		appConfigResponse["globalConfig"] = json.RawMessage(mapB)
	} else {
		if template.ChartRefId != chartRefId {
			templateRequested, err := handler.chartService.GetByAppIdAndChartRefId(appId, chartRefId)
			if err != nil && err != pg.ErrNoRows {
				handler.Logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}

			if pg.ErrNoRows == err {
				template.ChartRefId = chartRefId
				template.Id = 0
				template.Latest = false
			} else {
				template.ChartRefId = templateRequested.ChartRefId
				template.Id = templateRequested.Id
				template.ChartRepositoryId = templateRequested.ChartRepositoryId
				template.RefChartTemplate = templateRequested.RefChartTemplate
				template.RefChartTemplateVersion = templateRequested.RefChartTemplateVersion
				template.Latest = templateRequested.Latest
			}
		}
		template.Schema = schema
		template.Readme = string(readme)
		bytes, err := json.Marshal(template)
		if err != nil {
			handler.Logger.Errorw("marshal err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			return
		}
		appOverride := json.RawMessage(bytes)
		appConfigResponse["globalConfig"] = appOverride
	}

	common.WriteJsonResp(w, nil, appConfigResponse, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetDefaultDeploymentTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Error("error in getting appId path param, GetDefaultDeploymentTemplate", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		handler.Logger.Error("error in getting chartRefId path param, GetDefaultDeploymentTemplate", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	obj := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, obj); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "unauthorized user", http.StatusForbidden)
		return
	}
	defaultTemplate, _, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartRefId)
	if err != nil {
		handler.Logger.Errorw("error in getting default deployment template, GetDefaultDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, defaultTemplate, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelines(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelines", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCdPipelines", "appId", appId)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelines", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	ciConf, err := handler.pipelineBuilder.GetCdPipelinesForApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelines", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelinesForAppAndEnv(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelinesForAppAndEnv", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelinesForAppAndEnv", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCdPipelinesForAppAndEnv", "appId", appId, "envId", envId)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelinesForAppAndEnv", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac

	cdPipelines, err := handler.pipelineBuilder.GetCdPipelinesForAppAndEnv(appId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelinesForAppAndEnv", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, cdPipelines, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) PerformDeploymentApprovalAction(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var approvalActionRequest bean.UserApprovalActionRequest
	err = decoder.Decode(&approvalActionRequest)

	appId := approvalActionRequest.AppId
	pipelineId := approvalActionRequest.PipelineId

	approvalActionType := approvalActionRequest.ActionType

	if approvalActionType == bean.APPROVAL_APPROVE_ACTION {
		pipelineInfo, err := handler.pipelineBuilder.FindPipelineById(pipelineId)
		if err != nil {
			handler.Logger.Errorw("error occurred while fetching pipeline details", "pipelineId", pipelineId, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		allowed := handler.userAuthService.CheckForApproverAccess(pipelineInfo.App.AppName, pipelineInfo.Environment.EnvironmentIdentifier, userId)
		if !allowed {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	} else {
		//rbac block starts from here
		token := r.Header.Get("token")
		object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		object = handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(appId, pipelineId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		//rback block ends here
	}

	err = handler.cdHandler.PerformDeploymentApprovalAction(userId, approvalActionRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)

}

func (handler PipelineConfigRestHandlerImpl) GetArtifactsByCDPipeline(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	cdPipelineId, err := strconv.Atoi(vars["cd_pipeline_id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stage := r.URL.Query().Get("stage")
	if len(stage) == 0 {
		stage = pipeline.WorklowTypePre
	}
	searchString := ""
	search := r.URL.Query().Get("search")
	if len(search) != 0 {
		searchString = search
	}

	offset := 0
	limit := 10
	offsetQueryParam := r.URL.Query().Get("offset")
	if offsetQueryParam != "" {
		offset, err = strconv.Atoi(offsetQueryParam)
		if err != nil || offset < 0 {
			handler.Logger.Errorw("request err, GetArtifactsForRollback", "err", err, "offsetQueryParam", offsetQueryParam)
			common.WriteJsonResp(w, err, "invalid offset", http.StatusBadRequest)
			return
		}
	}

	sizeQueryParam := r.URL.Query().Get("size")
	if sizeQueryParam != "" {
		limit, err = strconv.Atoi(sizeQueryParam)
		if err != nil {
			handler.Logger.Errorw("request err, GetArtifactsForRollback", "err", err, "sizeQueryParam", sizeQueryParam)
			common.WriteJsonResp(w, err, "invalid size", http.StatusBadRequest)
			return
		}
	}

	isApprovalNode := false

	if stage == pipeline.WorkflowApprovalNode {
		isApprovalNode = true
		stage = pipeline.WorklowTypeDeploy
	}

	if !(stage == string(bean2.CD_WORKFLOW_TYPE_PRE) || stage == string(bean2.CD_WORKFLOW_TYPE_POST) || stage == string(bean2.CD_WORKFLOW_TYPE_DEPLOY)) {
		common.WriteJsonResp(w, fmt.Errorf("invalid stage param"), nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Infow("request payload, GetArtifactsByCDPipeline", "cdPipelineId", cdPipelineId, "stage", stage)

	pipeline, err := handler.pipelineBuilder.FindPipelineById(cdPipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACName(pipeline.App.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac for edit tags access
	triggerAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object)
	//rbac
	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(pipeline.App.AppName, pipeline.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac block ends here
	var ciArtifactResponse *bean.CiArtifactResponse

	handler.pipelineRestHandlerEnvConfig.UseArtifactListApiV2 = true

	pendingApprovalParam := r.URL.Query().Get("resource")
	if isApprovalNode && pendingApprovalParam == "PENDING_APPROVAL" {
		artifactsListFilterOptions := &bean2.ArtifactsListFilterOptions{
			Limit:        limit,
			Offset:       offset,
			SearchString: searchString,
		}
		ciArtifactResponse, err = handler.pipelineBuilder.FetchApprovalPendingArtifacts(pipeline, artifactsListFilterOptions)
	} else {
		if handler.pipelineRestHandlerEnvConfig.UseArtifactListApiV2 {
			artifactsListFilterOptions := &bean2.ArtifactsListFilterOptions{
				Limit:        limit,
				Offset:       offset,
				SearchString: searchString,
			}
			ciArtifactResponse, err = handler.pipelineBuilder.RetrieveArtifactsByCDPipelineV2(pipeline, bean2.WorkflowType(stage), artifactsListFilterOptions, isApprovalNode)
		} else {
			ciArtifactResponse, err = handler.pipelineBuilder.RetrieveArtifactsByCDPipeline(pipeline, bean2.WorkflowType(stage), searchString, limit, isApprovalNode)
		}
	}
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	if isApprovalNode {
		// fetch users with approval access to this app and Env
		approvalUsersByEnv, err := handler.userAuthService.GetApprovalUsersByEnv(pipeline.App.AppName, pipeline.Environment.EnvironmentIdentifier)
		if err != nil {
			handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		ciArtifactResponse.ApprovalUsers = approvalUsersByEnv
	}

	appTags, err := handler.imageTaggingService.GetUniqueTagsByAppId(pipeline.AppId)
	if err != nil {
		handler.Logger.Errorw("service err, GetTagsByAppId", "err", err, "appId", pipeline.AppId)
		common.WriteJsonResp(w, err, ciArtifactResponse, http.StatusInternalServerError)
		return
	}

	ciArtifactResponse.AppReleaseTagNames = appTags

	prodEnvExists, err := handler.imageTaggingService.GetProdEnvByCdPipelineId(pipeline.Id)
	ciArtifactResponse.TagsEditable = prodEnvExists && triggerAccess
	ciArtifactResponse.HideImageTaggingHardDelete = handler.imageTaggingService.GetImageTaggingServiceConfig().HideImageTaggingHardDelete
	if err != nil {
		handler.Logger.Errorw("service err, GetProdEnvByCdPipelineId", "err", err, "cdPipelineId", pipeline.Id)
		common.WriteJsonResp(w, err, ciArtifactResponse, http.StatusInternalServerError)
		return
	}

	var digests []string
	for _, item := range ciArtifactResponse.CiArtifacts {
		if len(item.ImageDigest) > 0 {
			digests = append(digests, item.ImageDigest)
		}
	}

	ciArtifactResponse.RequestedUserId = userId
	ciArtifactResponse.IsVirtualCluster = pipeline.Environment.IsVirtualEnvironment

	if len(digests) > 0 {
		//vulnerableMap := make(map[string]bool)
		cvePolicy, severityPolicy, err := handler.policyService.GetApplicablePolicy(pipeline.Environment.ClusterId,
			pipeline.EnvironmentId,
			pipeline.AppId,
			pipeline.App.AppType == helper.ChartStoreApp)

		if err != nil {
			handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		}

		// get image scan results from DB for given digests
		imageScanResults, err := handler.scanResultRepository.FindByImageDigests(digests)
		// ignore error
		if err != nil && err != pg.ErrNoRows {
			handler.Logger.Errorw("service err, FindByImageDigests", "err", err, "cdPipelineId", cdPipelineId, "stage", stage, "digests", digests)
		}

		// build digest vs cve-stores
		digestVsCveStores := make(map[string][]*security.CveStore)
		for _, result := range imageScanResults {
			imageHash := result.ImageScanExecutionHistory.ImageHash

			// For an imageHash, append all cveStores
			if val, ok := digestVsCveStores[imageHash]; !ok {

				// configuring size as len of ImageScanExecutionResult assuming all the
				//scan results could belong to a single hash
				cveStores := make([]*security.CveStore, 0, len(imageScanResults))
				cveStores = append(cveStores, &result.CveStore)
				digestVsCveStores[imageHash] = cveStores

			} else {
				// append to existing one
				digestVsCveStores[imageHash] = append(val, &result.CveStore)
			}
		}

		var ciArtifactsFinal []bean.CiArtifactBean
		for _, item := range ciArtifactResponse.CiArtifacts {

			// ignore cve check if scan is not enabled
			if !item.ScanEnabled {
				ciArtifactsFinal = append(ciArtifactsFinal, item)
				continue
			}

			cveStores, _ := digestVsCveStores[item.ImageDigest]
			item.IsVulnerable = handler.policyService.HasBlockedCVE(cveStores, cvePolicy, severityPolicy)
			ciArtifactsFinal = append(ciArtifactsFinal, item)
		}
		ciArtifactResponse.CiArtifacts = ciArtifactsFinal
	}

	common.WriteJsonResp(w, err, ciArtifactResponse, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetAppOverrideForDefaultTemplate(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetAppOverrideForDefaultTemplate", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetAppOverrideForDefaultTemplate", "err", err, "chartRefId", chartRefId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	appOverride, _, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartRefId)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, appOverride, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpdateAppOverride(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	var templateRequest chart.TemplateRequest
	err = decoder.Decode(&templateRequest)
	templateRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateAppOverride", "err", err, "payload", templateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.validator.Struct(templateRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateAppOverride", "err", err, "payload", templateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, UpdateAppOverride", "payload", templateRequest)

	token := r.Header.Get("token")
	ctx := r.Context()
	_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineBuilder.GetApp")
	app, err := handler.pipelineBuilder.GetApp(templateRequest.AppId)
	span.End()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	chartRefId := templateRequest.ChartRefId
	//VARIABLE_RESOLVE
	scope := resourceQualifiers.Scope{
		AppId: templateRequest.AppId,
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartService.DeploymentTemplateValidate")
	validate, err2 := handler.chartService.DeploymentTemplateValidate(ctx, templateRequest.ValuesOverride, chartRefId, scope)
	span.End()
	if !validate {
		handler.Logger.Errorw("validation err, UpdateAppOverride", "err", err2, "payload", templateRequest)
		common.WriteJsonResp(w, err2, nil, http.StatusBadRequest)
		return
	}

	protectionEnabled := handler.resourceProtectionService.ResourceProtectionEnabled(templateRequest.AppId, -1)
	if protectionEnabled {
		handler.Logger.Errorw("resource protection enabled", "appId", templateRequest.AppId, "envId", -1)
		common.WriteJsonResp(w, fmt.Errorf("resource protection enabled"), "resource protection enabled", http.StatusLocked)
		return
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "chartService.UpdateAppOverride")
	createResp, err := handler.chartService.UpdateAppOverride(ctx, &templateRequest)
	span.End()
	if err != nil {
		handler.Logger.Errorw("service err, UpdateAppOverride", "err", err, "payload", templateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)

}
func (handler PipelineConfigRestHandlerImpl) GetArtifactsForRollback(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	cdPipelineId, err := strconv.Atoi(vars["cd_pipeline_id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetArtifactsForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, "invalid request", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetArtifactsForRollback", "cdPipelineId", cdPipelineId)
	token := r.Header.Get("token")
	deploymentPipeline, err := handler.pipelineBuilder.FindPipelineById(cdPipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(deploymentPipeline.AppId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	offsetQueryParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetQueryParam)
	if offsetQueryParam == "" || err != nil {
		handler.Logger.Errorw("request err, GetArtifactsForRollback", "err", err, "offsetQueryParam", offsetQueryParam)
		common.WriteJsonResp(w, err, "invalid offset", http.StatusBadRequest)
		return
	}
	sizeQueryParam := r.URL.Query().Get("size")
	limit, err := strconv.Atoi(sizeQueryParam)
	if sizeQueryParam == "" || err != nil {
		handler.Logger.Errorw("request err, GetArtifactsForRollback", "err", err, "sizeQueryParam", sizeQueryParam)
		common.WriteJsonResp(w, err, "invalid size", http.StatusBadRequest)
		return
	}
	searchString := r.URL.Query().Get("search")

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, deploymentPipeline.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac block ends here
	//rbac for edit tags access
	var ciArtifactResponse bean.CiArtifactResponse
	triggerAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object)
	if handler.pipelineRestHandlerEnvConfig.UseArtifactListApiV2 {
		ciArtifactResponse, err = handler.pipelineBuilder.FetchArtifactForRollbackV2(cdPipelineId, app.Id, offset, limit, searchString, app, deploymentPipeline)
	} else {
		ciArtifactResponse, err = handler.pipelineBuilder.FetchArtifactForRollback(cdPipelineId, app.Id, offset, limit, searchString, app, deploymentPipeline)
	}

	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, "unable to fetch artifacts", http.StatusInternalServerError)
		return
	}
	appTags, err := handler.imageTaggingService.GetUniqueTagsByAppId(app.Id)
	if err != nil {
		handler.Logger.Errorw("service err, GetTagsByAppId", "err", err, "appId", app.Id)
		common.WriteJsonResp(w, err, ciArtifactResponse, http.StatusInternalServerError)
		return
	}

	ciArtifactResponse.AppReleaseTagNames = appTags

	prodEnvExists, err := handler.imageTaggingService.GetProdEnvByCdPipelineId(cdPipelineId)
	ciArtifactResponse.TagsEditable = prodEnvExists && triggerAccess
	ciArtifactResponse.HideImageTaggingHardDelete = handler.imageTaggingService.GetImageTaggingServiceConfig().HideImageTaggingHardDelete
	if err != nil {
		handler.Logger.Errorw("service err, GetProdEnvByCdPipelineId", "err", err, "cdPipelineId", app.Id)
		common.WriteJsonResp(w, err, ciArtifactResponse, http.StatusInternalServerError)
		return
	}
	ciArtifactResponse.RequestedUserId = userId
	common.WriteJsonResp(w, err, ciArtifactResponse, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetMigrationConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetMigrationConfig", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetMigrationConfig", "pipelineId", pipelineId)
	token := r.Header.Get("token")
	deploymentPipeline, err := handler.pipelineBuilder.FindPipelineById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(deploymentPipeline.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciConf, err := handler.dbMigrationService.GetByPipelineId(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetMigrationConfig", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateMigrationConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var dbMigrationConfigBean types.DbMigrationConfigBean
	err = decoder.Decode(&dbMigrationConfigBean)

	dbMigrationConfigBean.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CreateMigrationConfig", "payload", dbMigrationConfigBean)
	token := r.Header.Get("token")
	deploymentPipeline, err := handler.pipelineBuilder.FindPipelineById(dbMigrationConfigBean.PipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(deploymentPipeline.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.dbMigrationService.Save(&dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("service err, CreateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}
func (handler PipelineConfigRestHandlerImpl) UpdateMigrationConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var dbMigrationConfigBean types.DbMigrationConfigBean
	err = decoder.Decode(&dbMigrationConfigBean)
	dbMigrationConfigBean.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, UpdateMigrationConfig", "payload", dbMigrationConfigBean)
	token := r.Header.Get("token")
	deploymentPipeline, err := handler.pipelineBuilder.FindPipelineById(dbMigrationConfigBean.PipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(deploymentPipeline.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.dbMigrationService.Update(&dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideReset(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvConfigOverrideReset", "appId", appId, "environmentId", environmentId)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideReset", "err", err, "appId", appId, "environmentId", environmentId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionDelete, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, environmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionDelete, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	protectionEnabled := handler.resourceProtectionService.ResourceProtectionEnabled(appId, environmentId)
	if protectionEnabled {
		handler.Logger.Errorw("resource protection enabled", "appId", appId, "envId", environmentId)
		common.WriteJsonResp(w, fmt.Errorf("resource protection enabled"), "resource protection enabled", http.StatusLocked)
		return
	}
	isSuccess, err := handler.propertiesConfigService.ResetEnvironmentProperties(id)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideReset", "err", err, "appId", appId, "environmentId", environmentId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, isSuccess, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) AppMetricsEnableDisable(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	var appMetricEnableDisableRequest chart.AppMetricEnableDisableRequest
	err = decoder.Decode(&appMetricEnableDisableRequest)
	appMetricEnableDisableRequest.AppId = appId
	appMetricEnableDisableRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, AppMetricsEnableDisable", "err", err, "appId", appId, "payload", appMetricEnableDisableRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, AppMetricsEnableDisable", "err", err, "appId", appId, "payload", appMetricEnableDisableRequest)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.chartService.AppMetricsEnableDisable(appMetricEnableDisableRequest)
	if err != nil {
		handler.Logger.Errorw("service err, AppMetricsEnableDisable", "err", err, "appId", appId, "payload", appMetricEnableDisableRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvMetricsEnableDisable(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var appMetricEnableDisableRequest chart.AppMetricEnableDisableRequest
	err = decoder.Decode(&appMetricEnableDisableRequest)
	appMetricEnableDisableRequest.UserId = userId
	appMetricEnableDisableRequest.AppId = appId
	appMetricEnableDisableRequest.EnvironmentId = environmentId
	if err != nil {
		handler.Logger.Errorw("request err, EnvMetricsEnableDisable", "err", err, "appId", appId, "environmentId", environmentId, "payload", appMetricEnableDisableRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvMetricsEnableDisable", "err", err, "appId", appId, "environmentId", environmentId, "payload", appMetricEnableDisableRequest)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, appMetricEnableDisableRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.propertiesConfigService.EnvMetricsEnableDisable(&appMetricEnableDisableRequest)
	if err != nil {
		handler.Logger.Errorw("service err, EnvMetricsEnableDisable", "err", err, "appId", appId, "environmentId", environmentId, "payload", appMetricEnableDisableRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) ListDeploymentHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	offsetQueryParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetQueryParam)
	if offsetQueryParam == "" || err != nil {
		handler.Logger.Errorw("request err, ListDeploymentHistory", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
		common.WriteJsonResp(w, err, "invalid offset", http.StatusBadRequest)
		return
	}
	sizeQueryParam := r.URL.Query().Get("size")
	limit, err := strconv.Atoi(sizeQueryParam)
	if sizeQueryParam == "" || err != nil {
		handler.Logger.Errorw("request err, ListDeploymentHistory", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "sizeQueryParam", sizeQueryParam)
		common.WriteJsonResp(w, err, "invalid size", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, ListDeploymentHistory", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK
	resp := DeploymentHistoryResp{}
	wfs, err := handler.cdHandler.GetCdBuildHistory(appId, environmentId, pipelineId, offset, limit)
	resp.CdWorkflows = wfs
	if err != nil {
		handler.Logger.Errorw("service err, List", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}

	appTags, err := handler.imageTaggingService.GetUniqueTagsByAppId(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetTagsByAppId", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	resp.AppReleaseTagNames = appTags

	prodEnvExists, err := handler.imageTaggingService.GetProdEnvByCdPipelineId(pipelineId)
	resp.TagsEdiatable = prodEnvExists
	resp.HideImageTaggingHardDelete = handler.imageTaggingService.GetImageTaggingServiceConfig().HideImageTaggingHardDelete
	if err != nil {
		handler.Logger.Errorw("service err, GetProdEnvFromParentAndLinkedWorkflow", "err", err, "cdPipelineId", pipelineId)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetPrePostDeploymentLogs(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetPrePostDeploymentLogs", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "workflowId", workflowId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	logsReader, cleanUp, err := handler.cdHandler.GetRunningWorkflowLogs(environmentId, pipelineId, workflowId)
	if err != nil {
		handler.Logger.Errorw("service err, GetPrePostDeploymentLogs", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "workflowId", workflowId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	lastSeenMsgId := -1
	lastEventId := r.Header.Get("Last-Event-ID")
	if len(lastEventId) > 0 {
		lastSeenMsgId, err = strconv.Atoi(lastEventId)
		if err != nil {
			handler.Logger.Errorw("request err, GetPrePostDeploymentLogs", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "workflowId", workflowId, "lastEventId", lastEventId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	defer cancel()
	defer cleanUp()
	handler.streamOutput(w, logsReader, lastSeenMsgId)
}

func (handler PipelineConfigRestHandlerImpl) FetchCdWorkflowDetails(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowRunnerId"])
	if err != nil || buildId == 0 {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	showAppliedFiltersStr := r.URL.Query().Get("SHOW_APPLIED_FILTERS")
	showAppliedFilters := showAppliedFiltersStr == "true"
	handler.Logger.Infow("request payload, FetchCdWorkflowDetails", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "buildId", buildId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	resp, err := handler.cdHandler.FetchCdWorkflowDetails(appId, environmentId, pipelineId, buildId, showAppliedFilters)
	if err != nil {
		handler.Logger.Errorw("service err, FetchCdWorkflowDetails", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no workflow found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) DownloadArtifacts(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowRunnerId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, DownloadArtifacts", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(appId, pipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	file, err := handler.cdHandler.DownloadCdWorkflowArtifacts(pipelineId, buildId)
	defer file.Close()

	if err != nil {
		handler.Logger.Errorw("service err, DownloadArtifacts", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no workflow found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Itoa(buildId)+".zip")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))
	_, err = io.Copy(w, file)
	if err != nil {
		handler.Logger.Errorw("service err, DownloadArtifacts", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId)
	}
}

func (handler PipelineConfigRestHandlerImpl) GetStageStatus(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetStageStatus", "err", err, "appId", appId, "pipelineId", pipelineId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(appId, pipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	resp, err := handler.cdHandler.FetchCdPrePostStageStatus(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetStageStatus", "err", err, "appId", appId, "pipelineId", pipelineId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no status found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetConfigmapSecretsForDeploymentStages(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetConfigmapSecretsForDeploymentStages", "err", err, "pipelineId", pipelineId)
	deploymentPipeline, err := handler.pipelineBuilder.FindPipelineById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//FIXME: add RBAC
	resp, err := handler.pipelineBuilder.FetchConfigmapSecretsForCdStages(deploymentPipeline.AppId, deploymentPipeline.EnvironmentId, pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetConfigmapSecretsForDeploymentStages", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelineById(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	version := "v1"
	if strings.Contains(r.URL.Path, "v2") {
		version = "v2"
	}
	handler.Logger.Infow("request payload, GetCdPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	envObject := handler.enforcerUtil.GetEnvRBACNameByCdPipelineIdAndEnvId(pipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, envObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	ciConf, err := handler.pipelineBuilder.GetCdPipelineById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	cdResp, err := pipeline.CreatePreAndPostStageResponse(ciConf, version)
	if err != nil {
		handler.Logger.Errorw("service err, CheckForVersionAndCreatePreAndPostStagePayload", "err", err, "appId", appId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, cdResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CancelStage(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	workflowRunnerId, err := strconv.Atoi(vars["workflowRunnerId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	cdPipeline, err := handler.pipelineRepository.FindById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	handler.Logger.Infow("request payload, CancelStage", "pipelineId", pipelineId, "workflowRunnerId", workflowRunnerId)

	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(cdPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.cdHandler.CancelStage(workflowRunnerId, userId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelStage", "err", err, "pipelineId", pipelineId, "workflowRunnerId", workflowRunnerId)
		if util.IsErrNoRows(err) {
			common.WriteJsonResp(w, err, nil, http.StatusNotFound)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetDeploymentPipelineStrategy(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetDeploymentPipelineStrategy", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	result, err := handler.pipelineBuilder.FetchCDPipelineStrategy(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetDeploymentPipelineStrategy", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}
func (handler PipelineConfigRestHandlerImpl) GetDefaultDeploymentPipelineStrategy(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetDefaultDeploymentPipelineStrategy", "appId", appId, "envId", envId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	result, err := handler.pipelineBuilder.FetchDefaultCDPipelineStrategy(appId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, GetDefaultDeploymentPipelineStrategy", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideCreateNamespace(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var envConfigProperties bean3.EnvironmentProperties
	err = decoder.Decode(&envConfigProperties)
	envConfigProperties.UserId = userId
	envConfigProperties.EnvironmentId = environmentId
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvConfigOverrideCreateNamespace", "appId", appId, "environmentId", environmentId, "payload", envConfigProperties)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, environmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.propertiesConfigService.CreateEnvironmentPropertiesWithNamespace(appId, &envConfigProperties)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideCreateNamespace", "err", err, "appId", appId, "environmentId", environmentId, "payload", envConfigProperties)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) IsReadyToTrigger(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, IsReadyToTrigger", "appId", appId, "envId", envId, "pipelineId", pipelineId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, strings.ToLower(object)); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	result, err := handler.chartService.IsReadyToTrigger(appId, envId, pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, IsReadyToTrigger", "err", err, "appId", appId, "envId", envId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpgradeForAllApps(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var chartUpgradeRequest chart.ChartUpgradeRequest
	err = decoder.Decode(&chartUpgradeRequest)
	if err != nil {
		handler.Logger.Errorw("request err, UpgradeForAllApps", "err", err, "payload", chartUpgradeRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartUpgradeRequest.ChartRefId = chartRefId
	chartUpgradeRequest.UserId = userId
	handler.Logger.Infow("request payload, UpgradeForAllApps", "payload", chartUpgradeRequest)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, "*/*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, "*/*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	newAppOverride, _, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartUpgradeRequest.ChartRefId)
	if err != nil {
		handler.Logger.Errorw("service err, UpgradeForAllApps", "err", err, "payload", chartUpgradeRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx = context.WithValue(r.Context(), "token", acdToken)

	var appIds []int
	if chartUpgradeRequest.All || len(chartUpgradeRequest.AppIds) == 0 {
		apps, err := handler.pipelineBuilder.GetAppList()
		if err != nil {
			handler.Logger.Errorw("service err, UpgradeForAllApps", "err", err, "payload", chartUpgradeRequest)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		for _, app := range apps {
			appIds = append(appIds, app.Id)
		}
	} else {
		appIds = chartUpgradeRequest.AppIds
	}
	response := make(map[string][]map[string]string)
	var failedIds []map[string]string
	for _, appId := range appIds {
		appResponse := make(map[string]string)
		template, err := handler.chartService.GetByAppIdAndChartRefId(appId, chartRefId)
		if err != nil && pg.ErrNoRows != err {
			handler.Logger.Errorw("err in checking weather exist or not, skip for upgrade", "err", err, "payload", chartUpgradeRequest)
			appResponse["appId"] = strconv.Itoa(appId)
			appResponse["message"] = "err in checking weather exist or not, skip for upgrade"
			failedIds = append(failedIds, appResponse)
			continue
		}
		if template != nil && template.Id > 0 {
			handler.Logger.Warnw("this ref chart already configured for this app, skip for upgrade", "payload", chartUpgradeRequest)
			appResponse["appId"] = strconv.Itoa(appId)
			appResponse["message"] = "this ref chart already configured for this app, skip for upgrade"
			failedIds = append(failedIds, appResponse)
			continue
		}
		flag, err := handler.chartService.UpgradeForApp(appId, chartRefId, newAppOverride, userId, ctx)
		if err != nil {
			handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "payload", chartUpgradeRequest)
			appResponse["appId"] = strconv.Itoa(appId)
			appResponse["message"] = err.Error()
			failedIds = append(failedIds, appResponse)
		} else if flag == false {
			handler.Logger.Debugw("unable to upgrade for app", "appId", appId, "payload", chartUpgradeRequest)
			appResponse["appId"] = strconv.Itoa(appId)
			appResponse["message"] = "no error found, but failed to upgrade"
			failedIds = append(failedIds, appResponse)
		}

	}
	response["failed"] = failedIds
	common.WriteJsonResp(w, err, response, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelinesByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelines", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	appIdsString := v.Get("appIds")
	var appIds []int
	if len(appIdsString) > 0 {
		appIdsSlices := strings.Split(appIdsString, ",")
		for _, appId := range appIdsSlices {
			id, err := strconv.Atoi(appId)
			if err != nil {
				common.WriteJsonResp(w, err, "please provide valid appIds", http.StatusBadRequest)
				return
			}
			appIds = append(appIds, id)
		}
	}
	var appGroupId int
	appGroupIdStr := v.Get("appGroupId")
	if len(appGroupIdStr) > 0 {
		appGroupId, err = strconv.Atoi(appGroupIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid appGroupId", http.StatusBadRequest)
			return
		}
	}

	request := resourceGroup2.ResourceGroupingRequest{
		ParentResourceId:  envId,
		ResourceGroupId:   appGroupId,
		ResourceGroupType: resourceGroup2.APP_GROUP,
		ResourceIds:       appIds,
		EmailId:           userEmailId,
		CheckAuthBatch:    handler.checkAuthBatch,
		UserId:            userId,
		Ctx:               r.Context(),
	}
	_, span := otel.Tracer("orchestrator").Start(r.Context(), "cdHandler.FetchCdPipelinesForResourceGrouping")
	results, err := handler.pipelineBuilder.GetCdPipelinesByEnvironment(request)
	span.End()
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelines", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, results, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelinesByEnvironmentMin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelines", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	appIdsString := v.Get("appIds")
	var appIds []int
	if len(appIdsString) > 0 {
		appIdsSlices := strings.Split(appIdsString, ",")
		for _, appId := range appIdsSlices {
			id, err := strconv.Atoi(appId)
			if err != nil {
				common.WriteJsonResp(w, err, "please provide valid appIds", http.StatusBadRequest)
				return
			}
			appIds = append(appIds, id)
		}
	}
	var appGroupId int
	appGroupIdStr := v.Get("appGroupId")
	if len(appGroupIdStr) > 0 {
		appGroupId, err = strconv.Atoi(appGroupIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid appGroupId", http.StatusBadRequest)
			return
		}
	}

	request := resourceGroup2.ResourceGroupingRequest{
		ParentResourceId:  envId,
		ResourceGroupId:   appGroupId,
		ResourceGroupType: resourceGroup2.APP_GROUP,
		ResourceIds:       appIds,
		EmailId:           userEmailId,
		CheckAuthBatch:    handler.checkAuthBatch,
		UserId:            userId,
		Ctx:               r.Context(),
	}

	_, span := otel.Tracer("orchestrator").Start(r.Context(), "cdHandler.FetchCdPipelinesForResourceGrouping")
	results, err := handler.pipelineBuilder.GetCdPipelinesByEnvironmentMin(request)
	span.End()
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelines", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, results, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) checkAuthBatch(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool) {
	var appResult map[string]bool
	var envResult map[string]bool
	if len(appObject) > 0 {
		appResult = handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceApplications, casbin.ActionGet, appObject)
	}
	if len(envObject) > 0 {
		envResult = handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceEnvironment, casbin.ActionGet, envObject)
	}
	return appResult, envResult
}
