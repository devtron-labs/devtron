package app

import (
	"context"
	"encoding/json"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type DevtronAppDeploymentRestHandler interface {
	CreateCdPipeline(w http.ResponseWriter, r *http.Request)
	GetCdPipelineById(w http.ResponseWriter, r *http.Request)
	PatchCdPipeline(w http.ResponseWriter, r *http.Request)
	GetCdPipelines(w http.ResponseWriter, r *http.Request)
	GetCdPipelinesForAppAndEnv(w http.ResponseWriter, r *http.Request)

	GetArtifactsByCDPipeline(w http.ResponseWriter, r *http.Request)
	GetArtifactForRollback(w http.ResponseWriter, r *http.Request)

	UpgradeForAllApps(w http.ResponseWriter, r *http.Request)

	IsReadyToTrigger(w http.ResponseWriter, r *http.Request)
	FetchCdWorkflowDetails(w http.ResponseWriter, r *http.Request)
}

type DevtronAppDeploymentConfigRestHandler interface {
	ConfigureDeploymentTemplateForApp(w http.ResponseWriter, r *http.Request)
	GetDeploymentTemplate(w http.ResponseWriter, r *http.Request)
	GetAppOverrideForDefaultTemplate(w http.ResponseWriter, r *http.Request)

	EnvConfigOverrideCreate(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideUpdate(w http.ResponseWriter, r *http.Request)
	GetEnvConfigOverride(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideReset(w http.ResponseWriter, r *http.Request)

	UpdateAppOverride(w http.ResponseWriter, r *http.Request)
	GetConfigmapSecretsForDeploymentStages(w http.ResponseWriter, r *http.Request)
	GetDeploymentPipelineStrategy(w http.ResponseWriter, r *http.Request)

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

	validate, err2 := handler.chartService.DeploymentTemplateValidate(templateRequest.ValuesOverride, chartRefId)
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
	ctx = context.WithValue(r.Context(), "token", token)
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

	ctx := context.WithValue(r.Context(), "token", token)
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
	force := v.Get("force")
	if len(force) > 0 {
		forceDelete, err = strconv.ParseBool(force)
		if err != nil {
			handler.Logger.Errorw("request err, PatchCdPipeline", "err", err, "payload", cdPipeline)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	cdPipeline.ForceDelete = forceDelete
	handler.Logger.Infow("request payload, PatchCdPipeline", "payload", cdPipeline)
	err = handler.validator.StructPartial(cdPipeline, "AppId", "Action")
	if err == nil {
		if cdPipeline.Action == bean.CD_CREATE {
			err = handler.validator.Struct(cdPipeline.Pipeline)
		} else if cdPipeline.Action == bean.CD_DELETE {
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

	ctx := context.WithValue(r.Context(), "token", token)
	createResp, err := handler.pipelineBuilder.PatchCdPipelines(&cdPipeline, ctx)
	if err != nil {
		handler.Logger.Errorw("service err, PatchCdPipeline", "err", err, "payload", cdPipeline)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
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
	var envConfigProperties pipeline.EnvironmentProperties
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
	chartRefId := envConfigProperties.ChartRefId
	validate, err2 := handler.chartService.DeploymentTemplateValidate(envConfigProperties.EnvOverrideValues, chartRefId)
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
			ctx = context.WithValue(r.Context(), "token", token)
			templateRequest := chart.TemplateRequest{
				AppId:          appId,
				ChartRefId:     envConfigProperties.ChartRefId,
				ValuesOverride: []byte("{}"),
				UserId:         userId,
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
	var envConfigProperties pipeline.EnvironmentProperties
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
	chartRefId := envConfigProperties.ChartRefId
	validate, err2 := handler.chartService.DeploymentTemplateValidate(envConfigProperties.EnvOverrideValues, chartRefId)
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
		appOverride, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartRefId)
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

func (handler PipelineConfigRestHandlerImpl) GetArtifactsByCDPipeline(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	cdPipelineId, err := strconv.Atoi(vars["cd_pipeline_id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stage := r.URL.Query().Get("stage")
	if len(stage) == 0 {
		stage = "PRE"
	}
	handler.Logger.Infow("request payload, GetArtifactsByCDPipeline", "cdPipelineId", cdPipelineId, "stage", stage)
	deploymentPipeline, err := handler.pipelineBuilder.FindPipelineById(cdPipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(deploymentPipeline.AppId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

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

	ciArtifactResponse, err := handler.pipelineBuilder.GetArtifactsByCDPipeline(cdPipelineId, bean2.WorkflowType(stage))
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var digests []string
	for _, item := range ciArtifactResponse.CiArtifacts {
		digests = append(digests, item.ImageDigest)
	}

	//FIXME: next 3 loops are same combine them
	//FIXME: already fetched above as deployment pipeline
	pipelineModel, err := handler.pipelineRepository.FindById(cdPipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	if len(digests) > 0 {
		vulnerableMap := make(map[string]bool)
		cvePolicy, severityPolicy, err := handler.policyService.GetApplicablePolicy(pipelineModel.Environment.ClusterId, pipelineModel.EnvironmentId, pipelineModel.AppId, pipelineModel.App.AppStore)
		if err != nil {
			handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		}
		for _, digest := range digests {
			if len(digest) > 0 {
				var cveStores []*security.CveStore
				imageScanResult, err := handler.scanResultRepository.FindByImageDigest(digest)
				if err != nil && err != pg.ErrNoRows {
					handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
					continue //skip for other artifact to complete
				}
				for _, item := range imageScanResult {
					cveStores = append(cveStores, &item.CveStore)
				}
				vulnerableMap[digest] = handler.policyService.HasBlockedCVE(cveStores, cvePolicy, severityPolicy)
			}
		}
		var ciArtifactsFinal []bean.CiArtifactBean
		for _, item := range ciArtifactResponse.CiArtifacts {
			if item.ScanEnabled { // skip setting for artifacts which have marked scan disabled, but here deal with same digest
				//if isVulnerable, ok := vulnerableMap[item.ImageDigest]; ok {
				item.IsVulnerable = vulnerableMap[item.ImageDigest]
				//}
			}
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

	appOverride, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartRefId)
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
	chartRefId := templateRequest.ChartRefId
	validate, err2 := handler.chartService.DeploymentTemplateValidate(templateRequest.ValuesOverride, chartRefId)
	if !validate {
		handler.Logger.Errorw("validation err, UpdateAppOverride", "err", err2, "payload", templateRequest)
		common.WriteJsonResp(w, err2, nil, http.StatusBadRequest)
		return
	}

	createResp, err := handler.chartService.UpdateAppOverride(&templateRequest)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateAppOverride", "err", err, "payload", templateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)

}
func (handler PipelineConfigRestHandlerImpl) GetArtifactForRollback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cdPipelineId, err := strconv.Atoi(vars["cd_pipeline_id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, "invalid request", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetArtifactForRollback", "cdPipelineId", cdPipelineId)
	token := r.Header.Get("token")
	deploymentPipeline, err := handler.pipelineBuilder.FindPipelineById(cdPipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(deploymentPipeline.AppId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

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

	ciArtifactResponse, err := handler.pipelineBuilder.FetchArtifactForRollback(cdPipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		common.WriteJsonResp(w, err, "unable to fetch artifacts", http.StatusInternalServerError)
		return
	}
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
	var dbMigrationConfigBean pipeline.DbMigrationConfigBean
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
	var dbMigrationConfigBean pipeline.DbMigrationConfigBean
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
	ctx = context.WithValue(r.Context(), "token", token)
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

	resp, err := handler.cdHandler.GetCdBuildHistory(appId, environmentId, pipelineId, offset, limit)
	if err != nil {
		handler.Logger.Errorw("service err, List", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
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
	handler.Logger.Infow("request payload, FetchCdWorkflowDetails", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "buildId", buildId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	resp, err := handler.cdHandler.FetchCdWorkflowDetails(appId, environmentId, pipelineId, buildId)
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

	envId, err := handler.pipelineBuilder.GetEnvironmentByCdPipelineId(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	envObject := handler.enforcerUtil.GetEnvRBACNameByCdPipelineIdAndEnvId(pipelineId, envId)
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
	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
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

	resp, err := handler.cdHandler.CancelStage(workflowRunnerId)
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
	var envConfigProperties pipeline.EnvironmentProperties
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

	newAppOverride, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartUpgradeRequest.ChartRefId)
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
	ctx = context.WithValue(r.Context(), "token", token)

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
