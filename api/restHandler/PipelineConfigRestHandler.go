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

package restHandler

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appClone"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/bean"
	request "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	security2 "github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type PipelineConfigRestHandler interface {
	CreateCiConfig(w http.ResponseWriter, r *http.Request)
	ConfigureDeploymentTemplateForApp(w http.ResponseWriter, r *http.Request)
	CreateApp(w http.ResponseWriter, r *http.Request)
	DeleteApp(w http.ResponseWriter, r *http.Request)
	CreateMaterial(w http.ResponseWriter, r *http.Request)
	UpdateMaterial(w http.ResponseWriter, r *http.Request)
	CreateCdPipeline(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideCreate(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideUpdate(w http.ResponseWriter, r *http.Request)
	GetEnvConfOverride(w http.ResponseWriter, r *http.Request)
	GetCiPipeline(w http.ResponseWriter, r *http.Request)
	UpdateCiTemplate(w http.ResponseWriter, r *http.Request)
	PatchCiPipelines(w http.ResponseWriter, r *http.Request)
	GetDeploymentTemplate(w http.ResponseWriter, r *http.Request)
	GetApp(w http.ResponseWriter, r *http.Request)
	PatchCdPipeline(w http.ResponseWriter, r *http.Request)
	GetCdPipelines(w http.ResponseWriter, r *http.Request)
	GetCdPipelinesForAppAndEnv(w http.ResponseWriter, r *http.Request)
	GetArtifactsByCDPipeline(w http.ResponseWriter, r *http.Request)
	FetchArtifactForRollback(w http.ResponseWriter, r *http.Request)
	GetAppOverrideForDefaultTemplate(w http.ResponseWriter, r *http.Request)
	UpdateAppOverride(w http.ResponseWriter, r *http.Request)

	GetMigrationConfig(w http.ResponseWriter, r *http.Request)
	CreateMigrationConfig(w http.ResponseWriter, r *http.Request)
	UpdateMigrationConfig(w http.ResponseWriter, r *http.Request)

	FindAppsByTeamId(w http.ResponseWriter, r *http.Request)
	FindAppsByTeamName(w http.ResponseWriter, r *http.Request)

	TriggerCiPipeline(w http.ResponseWriter, r *http.Request)
	TriggerCiPipelineFromGitWebhook(w http.ResponseWriter, r *http.Request)
	FetchMaterials(w http.ResponseWriter, r *http.Request)
	FetchWorkflowDetails(w http.ResponseWriter, r *http.Request)
	GetCiPipelineMin(w http.ResponseWriter, r *http.Request)

	FetchChanges(w http.ResponseWriter, r *http.Request)
	GetHistoricBuildLogs(w http.ResponseWriter, r *http.Request)
	GetBuildLogs(w http.ResponseWriter, r *http.Request)
	GetBuildHistory(w http.ResponseWriter, r *http.Request)
	HandleWorkflowWebhook(w http.ResponseWriter, r *http.Request)
	CancelWorkflow(w http.ResponseWriter, r *http.Request)
	CancelStage(w http.ResponseWriter, r *http.Request)
	DownloadCiWorkflowArtifacts(w http.ResponseWriter, r *http.Request)

	GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request)
	GetAppListByTeamIds(w http.ResponseWriter, r *http.Request)
	EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request)
	GitListAutocomplete(w http.ResponseWriter, r *http.Request)
	DockerListAutocomplete(w http.ResponseWriter, r *http.Request)
	TeamListAutocomplete(w http.ResponseWriter, r *http.Request)

	IsReadyToTrigger(w http.ResponseWriter, r *http.Request)
	FetchCDPipelineStrategy(w http.ResponseWriter, r *http.Request)

	UpgradeForAllApps(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideReset(w http.ResponseWriter, r *http.Request)
	EnvConfigOverrideCreateNamespace(w http.ResponseWriter, r *http.Request)

	AppMetricsEnableDisable(w http.ResponseWriter, r *http.Request)
	EnvMetricsEnableDisable(w http.ResponseWriter, r *http.Request)

	GetCdBuildHistory(w http.ResponseWriter, r *http.Request)
	GetCdBuildLogs(w http.ResponseWriter, r *http.Request)
	FetchCdWorkflowDetails(w http.ResponseWriter, r *http.Request)
	DownloadCdWorkflowArtifacts(w http.ResponseWriter, r *http.Request)
	FetchCdPrePostStageStatus(w http.ResponseWriter, r *http.Request)

	GetCdPipelineById(w http.ResponseWriter, r *http.Request)
	FetchConfigmapSecretsForCdStages(w http.ResponseWriter, r *http.Request)

	FetchAppWorkflowStatusForTriggerView(w http.ResponseWriter, r *http.Request)

	RefreshMaterials(w http.ResponseWriter, r *http.Request)
	FetchMaterialInfo(w http.ResponseWriter, r *http.Request)
	GetCIPipelineById(w http.ResponseWriter, r *http.Request)
	PipelineNameSuggestion(w http.ResponseWriter, r *http.Request)
}

type PipelineConfigRestHandlerImpl struct {
	pipelineBuilder         pipeline.PipelineBuilder
	ciPipelineRepository    pipelineConfig.CiPipelineRepository
	ciHandler               pipeline.CiHandler
	Logger                  *zap.SugaredLogger
	chartService            pipeline.ChartService
	propertiesConfigService pipeline.PropertiesConfigService
	dbMigrationService      pipeline.DbMigrationService
	application             application.ServiceClient
	userAuthService         user.UserService
	validator               *validator.Validate
	teamService             team.TeamService
	enforcer                rbac.Enforcer
	gitSensorClient         gitSensor.GitSensorClient
	pipelineRepository      pipelineConfig.PipelineRepository
	appWorkflowService      appWorkflow.AppWorkflowService
	enforcerUtil            rbac.EnforcerUtil
	envService              request.EnvironmentService
	gitRegistryConfig       pipeline.GitRegistryConfig
	dockerRegistryConfig    pipeline.DockerRegistryConfig
	cdHandelr               pipeline.CdHandler
	appCloneService         appClone.AppCloneService
	materialRepository      pipelineConfig.MaterialRepository
	policyService           security2.PolicyService
	scanResultRepository    security.ImageScanResultRepository
}

func NewPipelineRestHandlerImpl(pipelineBuilder pipeline.PipelineBuilder, Logger *zap.SugaredLogger,
	chartService pipeline.ChartService,
	propertiesConfigService pipeline.PropertiesConfigService,
	dbMigrationService pipeline.DbMigrationService,
	application application.ServiceClient,
	userAuthService user.UserService,
	teamService team.TeamService,
	enforcer rbac.Enforcer,
	ciHandler pipeline.CiHandler,
	validator *validator.Validate,
	gitSensorClient gitSensor.GitSensorClient,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	enforcerUtil rbac.EnforcerUtil, envService request.EnvironmentService,
	gitRegistryConfig pipeline.GitRegistryConfig, dockerRegistryConfig pipeline.DockerRegistryConfig,
	cdHandelr pipeline.CdHandler,
	appCloneService appClone.AppCloneService,
	appWorkflowService appWorkflow.AppWorkflowService,
	materialRepository pipelineConfig.MaterialRepository, policyService security2.PolicyService,
	scanResultRepository security.ImageScanResultRepository) *PipelineConfigRestHandlerImpl {
	return &PipelineConfigRestHandlerImpl{
		pipelineBuilder:         pipelineBuilder,
		Logger:                  Logger,
		chartService:            chartService,
		propertiesConfigService: propertiesConfigService,
		dbMigrationService:      dbMigrationService,
		application:             application,
		userAuthService:         userAuthService,
		validator:               validator,
		teamService:             teamService,
		enforcer:                enforcer,
		ciHandler:               ciHandler,
		gitSensorClient:         gitSensorClient,
		ciPipelineRepository:    ciPipelineRepository,
		pipelineRepository:      pipelineRepository,
		enforcerUtil:            enforcerUtil,
		envService:              envService,
		gitRegistryConfig:       gitRegistryConfig,
		dockerRegistryConfig:    dockerRegistryConfig,
		cdHandelr:               cdHandelr,
		appCloneService:         appCloneService,
		appWorkflowService:      appWorkflowService,
		materialRepository:      materialRepository,
		policyService:           policyService,
		scanResultRepository:    scanResultRepository,
	}
}

const devtron = "DEVTRON"

func (handler PipelineConfigRestHandlerImpl) DeleteApp(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, delete app", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, delete app", "appId", appId)
	wfs, err := handler.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		handler.Logger.Errorw("could not fetch wfs", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if len(wfs) != 0 {
		handler.Logger.Info("cannot delete app with workflow's")
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "cannot delete app having workflow's"}
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	team, err := handler.teamService.FindTeamByAppId(appId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, fmt.Sprintf("%s/%s", strings.ToLower(team.Name), "*")); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	err = handler.pipelineBuilder.DeleteApp(appId, userId)
	if err != nil {
		handler.Logger.Errorw("service error, delete app", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, nil, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateApp(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var createRequest bean.CreateAppDTO
	err = decoder.Decode(&createRequest)
	createRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateApp", "err", err, "CreateApp", createRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Infow("request payload, CreateApp", "CreateApp", createRequest)
	err = handler.validator.Struct(createRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateApp", "err", err, "CreateApp", createRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	team, err := handler.teamService.FetchOne(createRequest.TeamId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, fmt.Sprintf("%s/%s", strings.ToLower(team.Name), "*")); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	var createResp *bean.CreateAppDTO
	err = nil
	if createRequest.TemplateId == 0 {
		createResp, err = handler.pipelineBuilder.CreateApp(&createRequest)
	} else {
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
		createResp, err = handler.appCloneService.CloneApp(&createRequest, ctx)
	}
	if err != nil {
		handler.Logger.Errorw("service err, CreateApp", "err", err, "CreateApp", createRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var createMaterialDto bean.CreateMaterialDTO
	err = decoder.Decode(&createMaterialDto)
	createMaterialDto.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateMaterial", "err", err, "CreateMaterial", createMaterialDto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CreateMaterial", "CreateMaterial", createMaterialDto)
	err = handler.validator.Struct(createMaterialDto)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateMaterial", "err", err, "CreateMaterial", createMaterialDto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	team, err := handler.teamService.FindTeamByAppId(createMaterialDto.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, fmt.Sprintf("%s/%s", strings.ToLower(team.Name), "*")); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	createResp, err := handler.pipelineBuilder.CreateMaterialsForApp(&createMaterialDto)
	if err != nil {
		handler.Logger.Errorw("service err, CreateMaterial", "err", err, "CreateMaterial", createMaterialDto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var updateMaterialDto bean.UpdateMaterialDTO
	err = decoder.Decode(&updateMaterialDto)
	updateMaterialDto.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, UpdateMaterial", "UpdateMaterial", updateMaterialDto)
	err = handler.validator.Struct(updateMaterialDto)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	team, err := handler.teamService.FindTeamByAppId(updateMaterialDto.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, fmt.Sprintf("%s/%s", strings.ToLower(team.Name), "*")); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	createResp, err := handler.pipelineBuilder.UpdateMaterialsForApp(&updateMaterialDto)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetApp(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, get app", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, get app", "appId", appId)
	ciConf, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, get app", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rback implementation starts here
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback implementation ends here

	writeJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateCiConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var createRequest bean.CiConfigRequest
	err = decoder.Decode(&createRequest)
	createRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, create ci config", "err", err, "create request", createRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, create ci config", "create request", createRequest)
	err = handler.validator.Struct(createRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, create ci config", "err", err, "create request", createRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(createRequest.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.pipelineBuilder.CreateCiPipeline(&createRequest)
	if err != nil {
		handler.Logger.Errorw("service err, create", "err", err, "create request", createRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpdateCiTemplate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var configRequest bean.CiConfigRequest
	err = decoder.Decode(&configRequest)
	configRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateCiTemplate", "err", err, "UpdateCiTemplate", configRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, update ci template", "UpdateCiTemplate", configRequest)
	err = handler.validator.Struct(configRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateCiTemplate", "err", err, "UpdateCiTemplate", configRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(configRequest.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.pipelineBuilder.UpdateCiTemplate(&configRequest)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "UpdateCiTemplate", configRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) PatchCiPipelines(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var patchRequest bean.CiPatchRequest
	err = decoder.Decode(&patchRequest)
	patchRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, PatchCiPipelines", "err", err, "PatchCiPipelines", patchRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, PatchCiPipelines", "PatchCiPipelines", patchRequest)
	err = handler.validator.Struct(patchRequest)
	if err != nil {
		handler.Logger.Errorw("validation err", "err", err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Debugw("update request ", "req", patchRequest)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(patchRequest.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.pipelineBuilder.PatchCiPipeline(&patchRequest)
	if err != nil {
		handler.Logger.Errorw("service err, PatchCiPipelines", "err", err, "PatchCiPipelines", patchRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) ConfigureDeploymentTemplateForApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var templateRequest pipeline.TemplateRequest
	err = decoder.Decode(&templateRequest)
	templateRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, ConfigureDeploymentTemplateForApp", "err", err, "payload", templateRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, ConfigureDeploymentTemplateForApp", "payload", templateRequest)
	err = handler.validator.Struct(templateRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, ConfigureDeploymentTemplateForApp", "err", err, "payload", templateRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(templateRequest.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
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
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateCdPipeline(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var cdPipeline bean.CdPipelines
	err = decoder.Decode(&cdPipeline)
	cdPipeline.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateCdPipeline", "err", err, "payload", cdPipeline)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CreateCdPipeline", "payload", cdPipeline)
	err = handler.validator.Struct(cdPipeline)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateCdPipeline", "err", err, "payload", cdPipeline)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Debugw("pipeline create request ", "req", cdPipeline)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(cdPipeline.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	for _, pline := range cdPipeline.Pipelines {
		object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, pline.EnvironmentId)
		handler.Logger.Debugw("Triggered Request By:", "object", object)
		if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, object); !ok {
			writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	//RBAC

	ctx := context.WithValue(r.Context(), "token", token)
	createResp, err := handler.pipelineBuilder.CreateCdPipelines(&cdPipeline, ctx)
	if err != nil {
		handler.Logger.Errorw("service err, CreateCdPipeline", "err", err, "payload", cdPipeline)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) PatchCdPipeline(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var cdPipeline bean.CDPatchRequest
	err = decoder.Decode(&cdPipeline)
	cdPipeline.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, PatchCdPipeline", "err", err, "payload", cdPipeline)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, PatchCdPipeline", "payload", cdPipeline)
	err = handler.validator.StructPartial(cdPipeline, "AppId", "Action")
	if err == nil {
		if cdPipeline.Action == bean.CdPatchAction(bean.CD_CREATE) {
			err = handler.validator.Struct(cdPipeline.Pipeline)
		} else if cdPipeline.Action == bean.CdPatchAction(bean.CD_DELETE) {
			err = handler.validator.Var(cdPipeline.Pipeline.Id, "gt=0")
		}
	}
	if err != nil {
		handler.Logger.Errorw("validation err, PatchCdPipeline", "err", err, "payload", cdPipeline)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(cdPipeline.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	object := handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(cdPipeline.AppId, cdPipeline.Pipeline.Id)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	ctx := context.WithValue(r.Context(), "token", token)
	createResp, err := handler.pipelineBuilder.PatchCdPipelines(&cdPipeline, ctx)
	if err != nil {
		handler.Logger.Errorw("service err, PatchCdPipeline", "err", err, "payload", cdPipeline)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideCreate(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var envConfigProperties pipeline.EnvironmentProperties
	err = decoder.Decode(&envConfigProperties)
	if err != nil {
		handler.Logger.Errorw("request err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envConfigProperties.UserId = userId
	envConfigProperties.EnvironmentId = environmentId
	handler.Logger.Infow("request payload, EnvConfigOverrideCreate", "payload", envConfigProperties)

	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, environmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
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
			templateRequest := pipeline.TemplateRequest{
				AppId:          appId,
				ChartRefId:     envConfigProperties.ChartRefId,
				ValuesOverride: []byte("{}"),
				UserId:         userId,
			}

			_, err = handler.chartService.CreateChartFromEnvOverride(templateRequest, ctx)
			if err != nil {
				handler.Logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
			createResp, err = handler.propertiesConfigService.CreateEnvironmentProperties(appId, &envConfigProperties)
			if err != nil {
				handler.Logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
		} else {
			handler.Logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", envConfigProperties)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	//userId := getLoggedInUser(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var envConfigProperties pipeline.EnvironmentProperties
	err = decoder.Decode(&envConfigProperties)
	envConfigProperties.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvConfigOverrideUpdate", "payload", envConfigProperties)
	err = handler.validator.Struct(envConfigProperties)
	if err != nil {
		handler.Logger.Errorw("validation err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	envConfigOverride, err := handler.propertiesConfigService.GetAppIdByChartEnvId(envConfigProperties.Id)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	appId := envConfigOverride.Chart.AppId
	envId := envConfigOverride.TargetEnvironment
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.propertiesConfigService.UpdateEnvironmentProperties(appId, &envConfigProperties, userId)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetEnvConfOverride(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetEnvConfOverride", "err", err, "payload", appId, environmentId, chartRefId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetEnvConfOverride", "payload", appId, environmentId, chartRefId)
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	env, err := handler.propertiesConfigService.GetEnvironmentProperties(appId, environmentId, chartRefId)
	if err != nil {
		handler.Logger.Errorw("service err, GetEnvConfOverride", "err", err, "payload", appId, environmentId, chartRefId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, env, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCiPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipeline", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciConf, err := handler.pipelineBuilder.GetCiPipeline(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipeline", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if ciConf == nil || ciConf.Id == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no data found"}
	}
	writeJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetDeploymentTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		handler.Logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetDeploymentTemplate", "appId", appId, "chartRefId", chartRefId)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	appConfigResponse := map[string]json.RawMessage{}
	appConfigResponse["globalConfig"] = nil

	template, err := handler.chartService.FindLatestChartForAppByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		handler.Logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if pg.ErrNoRows == err {
		appOverride, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartRefId)
		if err != nil {
			handler.Logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		mapB, _ := json.Marshal(appOverride)
		if err != nil {
			handler.Logger.Errorw("marshal err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			return
		}
		appConfigResponse["globalConfig"] = mapB
	} else {
		if template.ChartRefId != chartRefId {
			templateRequested, err := handler.chartService.GetByAppIdAndChartRefId(appId, chartRefId)
			if err != nil && err != pg.ErrNoRows {
				handler.Logger.Errorw("service err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
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

		bytes, err := json.Marshal(template)
		if err != nil {
			handler.Logger.Errorw("marshal err, GetDeploymentTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
			return
		}
		appOverride := json.RawMessage(bytes)
		appConfigResponse["globalConfig"] = appOverride
	}

	writeJsonResp(w, nil, appConfigResponse, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelines(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelines", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCdPipelines", "appId", appId)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelines", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	ciConf, err := handler.pipelineBuilder.GetCdPipelinesForApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelines", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelinesForAppAndEnv(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelinesForAppAndEnv", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelinesForAppAndEnv", "err", err, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCdPipelinesForAppAndEnv", "appId", appId, "envId", envId)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelinesForAppAndEnv", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac

	cdPipelines, err := handler.pipelineBuilder.GetCdPipelinesForAppAndEnv(appId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelinesForAppAndEnv", "err", err, "appId", appId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, cdPipelines, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetArtifactsByCDPipeline(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	cdPipelineId, err := strconv.Atoi(vars["cd_pipeline_id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stage := r.URL.Query().Get("stage")
	if len(stage) == 0 {
		stage = "PRE"
	}
	handler.Logger.Infow("request payload, GetArtifactsByCDPipeline", "cdPipelineId", cdPipelineId, "stage", stage)
	pipeline, err := handler.pipelineBuilder.FindPipelineById(cdPipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(pipeline.AppId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, pipeline.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here

	ciArtifactResponse, err := handler.pipelineBuilder.GetArtifactsByCDPipeline(cdPipelineId, bean2.CdWorkflowType(stage))
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var digests []string
	for _, item := range ciArtifactResponse.CiArtifacts {
		digests = append(digests, item.ImageDigest)
	}

	pipelineModel, err := handler.pipelineRepository.FindById(cdPipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	if len(digests) > 0 {
		vulnerableMap := make(map[string]bool)
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
				blockCveList, err := handler.policyService.GetBlockedCVEList(cveStores, pipelineModel.Environment.ClusterId, pipelineModel.EnvironmentId, pipelineModel.AppId, pipelineModel.App.AppStore)
				if err != nil {
					handler.Logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", cdPipelineId, "stage", stage)
				}
				if blockCveList != nil && len(blockCveList) > 0 {
					vulnerableMap[digest] = true
				}
			}
		}
		var ciArtifactsFinal []bean.CiArtifactBean
		for _, item := range ciArtifactResponse.CiArtifacts {
			if item.ScanEnabled { // skip setting for artifacts which have marked scan disabled, but here deal with same digest
				if _, ok := vulnerableMap[item.ImageDigest]; ok {
					item.IsVulnerable = true
				}
			}
			ciArtifactsFinal = append(ciArtifactsFinal, item)
		}
		ciArtifactResponse.CiArtifacts = ciArtifactsFinal
	}

	writeJsonResp(w, err, ciArtifactResponse, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetAppOverrideForDefaultTemplate(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetAppOverrideForDefaultTemplate", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetAppOverrideForDefaultTemplate", "err", err, "chartRefId", chartRefId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	appOverride, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartRefId)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "appId", appId, "chartRefId", chartRefId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, appOverride, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpdateAppOverride(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var templateRequest pipeline.TemplateRequest
	err = decoder.Decode(&templateRequest)
	templateRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateAppOverride", "err", err, "payload", templateRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(templateRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateAppOverride", "err", err, "payload", templateRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, UpdateAppOverride", "payload", templateRequest)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(templateRequest.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	createResp, err := handler.chartService.UpdateAppOverride(&templateRequest)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateAppOverride", "err", err, "payload", templateRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}
func (handler PipelineConfigRestHandlerImpl) FetchArtifactForRollback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cdPipelineId, err := strconv.Atoi(vars["cd_pipeline_id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		writeJsonResp(w, err, "invalid request", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchArtifactForRollback", "cdPipelineId", cdPipelineId)
	token := r.Header.Get("token")
	pipeline, err := handler.pipelineBuilder.FindPipelineById(cdPipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(pipeline.AppId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, pipeline.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here

	ciArtifactResponse, err := handler.pipelineBuilder.FetchArtifactForRollback(cdPipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchArtifactForRollback", "err", err, "cdPipelineId", cdPipelineId)
		writeJsonResp(w, err, "unable to fetch artifacts", http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, ciArtifactResponse, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetMigrationConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetMigrationConfig", "err", err, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetMigrationConfig", "pipelineId", pipelineId)
	token := r.Header.Get("token")
	pipeline, err := handler.pipelineBuilder.FindPipelineById(pipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(pipeline.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciConf, err := handler.dbMigrationService.GetByPipelineId(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetMigrationConfig", "err", err, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateMigrationConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var dbMigrationConfigBean pipeline.DbMigrationConfigBean
	err = decoder.Decode(&dbMigrationConfigBean)

	dbMigrationConfigBean.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CreateMigrationConfig", "payload", dbMigrationConfigBean)
	token := r.Header.Get("token")
	pipeline, err := handler.pipelineBuilder.FindPipelineById(dbMigrationConfigBean.PipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(pipeline.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.dbMigrationService.Save(&dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("service err, CreateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}
func (handler PipelineConfigRestHandlerImpl) UpdateMigrationConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var dbMigrationConfigBean pipeline.DbMigrationConfigBean
	err = decoder.Decode(&dbMigrationConfigBean)
	dbMigrationConfigBean.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, UpdateMigrationConfig", "payload", dbMigrationConfigBean)
	token := r.Header.Get("token")
	pipeline, err := handler.pipelineBuilder.FindPipelineById(dbMigrationConfigBean.PipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(pipeline.AppId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.dbMigrationService.Update(&dbMigrationConfigBean)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateMigrationConfig", "err", err, "payload", dbMigrationConfigBean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FindAppsByTeamId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamId, err := strconv.Atoi(vars["teamId"])
	if err != nil {
		handler.Logger.Errorw("request err, FindAppsByTeamId", "err", err, "teamId", teamId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FindAppsByTeamId", "teamId", teamId)
	team, err := handler.pipelineBuilder.FindAppsByTeamId(teamId)
	if err != nil {
		handler.Logger.Errorw("service err, FindAppsByTeamId", "err", err, "teamId", teamId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, team, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FindAppsByTeamName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["teamName"]
	handler.Logger.Infow("request payload, FindAppsByTeamName", "teamName", teamName)
	team, err := handler.pipelineBuilder.FindAppsByTeamName(teamName)
	if err != nil {
		handler.Logger.Errorw("service err, FindAppsByTeamName", "err", err, "teamName", teamName)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, team, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) TriggerCiPipelineFromGitWebhook(w http.ResponseWriter, r *http.Request) {
	var ciGitTriggerRequest bean.CiGitWebhookTriggerRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&ciGitTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("request err, TriggerCiPipelineFromGitWebhook", "err", err, "payload", ciGitTriggerRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, TriggerCiPipelineFromGitWebhook", "payload", ciGitTriggerRequest)
	err = handler.ciHandler.HandleGitCiTrigger(ciGitTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("service err, TriggerCiPipelineFromGitWebhook", "err", err, "payload", ciGitTriggerRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	writeJsonResp(w, err, nil, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) TriggerCiPipeline(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var ciTriggerRequest bean.CiTriggerRequest
	err = decoder.Decode(&ciTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("request err, TriggerCiPipeline", "err", err, "payload", ciTriggerRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !handler.validForMultiMaterial(ciTriggerRequest) {
		handler.Logger.Errorw("invalid req, commit hash not present for multi-git", "payload", ciTriggerRequest)
		writeJsonResp(w, errors.New("invalid req, commit hash not present for multi-git"),
			nil, http.StatusBadRequest)
	}
	ciTriggerRequest.TriggeredBy = userId
	handler.Logger.Infow("request payload, TriggerCiPipeline", "payload", ciTriggerRequest)

	//RBAC CHECK CD PIPELINE - FOR USER
	pipelines, err := handler.pipelineRepository.FindAutomaticByCiPipelineId(ciTriggerRequest.PipelineId)
	var authorizedPipelines []pipelineConfig.Pipeline
	var unauthorizedPipelines []pipelineConfig.Pipeline
	//fetching user only for getting token
	user, err := handler.userAuthService.GetById(ciTriggerRequest.TriggeredBy)
	if err != nil {
		handler.Logger.Errorw("service err, TriggerCiPipeline", "err", err, "payload", ciTriggerRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := user.AccessToken
	for _, p := range pipelines {
		pass := 0
		object := handler.enforcerUtil.GetAppRBACNameByAppId(p.AppId)
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
			handler.Logger.Debug(fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		} else {
			pass = 1
		}
		object = handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(p.AppId, p.Id)
		if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionTrigger, object); !ok {
			handler.Logger.Debug(fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		} else {
			pass = 2
		}
		if pass == 2 {
			authorizedPipelines = append(authorizedPipelines, *p)
		} else {
			unauthorizedPipelines = append(unauthorizedPipelines, *p)
		}
	}
	resMessage := "allowed for all pipelines"
	response := make(map[string]string)
	if len(unauthorizedPipelines) > 0 {
		resMessage = "not authorized for few pipelines, will not effected"
	}
	//RBAC CHECK CD PIPELINE - FOR USER

	resp, err := handler.ciHandler.HandleCIManual(ciTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("service err, TriggerCiPipeline", "err", err, "payload", ciTriggerRequest)
		writeJsonResp(w, err, response, http.StatusInternalServerError)
	}
	response["apiResponse"] = strconv.Itoa(resp)
	response["authStatus"] = resMessage

	writeJsonResp(w, err, response, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchMaterials(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchMaterials", "pipelineId", pipelineId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.FetchMaterialsByPipelineId(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchMaterials", "err", err, "pipelineId", pipelineId)
		writeJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) RefreshMaterials(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	gitMaterialId, err := strconv.Atoi(vars["gitMaterialId"])
	if err != nil {
		handler.Logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, RefreshMaterials", "gitMaterialId", gitMaterialId)
	material, err := handler.materialRepository.FindById(gitMaterialId)
	if err != nil {
		handler.Logger.Errorw("service err, RefreshMaterials", "err", err, "gitMaterialId", gitMaterialId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(material.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.ciHandler.RefreshMaterialByCiPipelineMaterialId(material.Id)
	if err != nil {
		handler.Logger.Errorw("service err, RefreshMaterials", "err", err, "gitMaterialId", gitMaterialId)
		writeJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCiPipelineMin(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//RBAC
	handler.Logger.Infow("request payload, GetCiPipelineMin", "appId", appId)
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	ciPipelines, err := handler.pipelineBuilder.GetCiPipelineMin(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipelineMin", "err", err, "appId", appId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no data found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	writeJsonResp(w, err, ciPipelines, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchWorkflowDetails(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowId"])
	if err != nil || buildId == 0 {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchWorkflowDetails", "appId", appId, "pipelineId", pipelineId, "buildId", buildId, "buildId", buildId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.FetchWorkflowDetails(appId, pipelineId, buildId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchWorkflowDetails", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no workflow found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) DownloadCiWorkflowArtifacts(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, DownloadCiWorkflowArtifacts", "pipelineId", pipelineId, "buildId", buildId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	file, err := handler.ciHandler.DownloadCiWorkflowArtifacts(pipelineId, buildId)
	defer file.Close()
	if err != nil {
		handler.Logger.Errorw("service err, DownloadCiWorkflowArtifacts", "err", err, "pipelineId", pipelineId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no workflow found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Itoa(buildId)+".zip")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))
	_, err = io.Copy(w, file)
	if err != nil {
		handler.Logger.Errorw("service err, DownloadCiWorkflowArtifacts", "err", err, "pipelineId", pipelineId, "buildId", buildId)
	}
}

func (handler PipelineConfigRestHandlerImpl) CancelStage(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	workflowRunnerId, err := strconv.Atoi(vars["workflowRunnerId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	cdPipeline, err := handler.pipelineRepository.FindById(pipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	handler.Logger.Infow("request payload, CancelStage", "pipelineId", pipelineId, "workflowRunnerId", workflowRunnerId)

	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(cdPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.cdHandelr.CancelStage(workflowRunnerId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelStage", "err", err, "pipelineId", pipelineId, "workflowRunnerId", workflowRunnerId)
		if util.IsErrNoRows(err) {
			writeJsonResp(w, err, nil, http.StatusNotFound)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CancelWorkflow(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		handler.Logger.Errorw("request err, CancelWorkflow", "err", err, "workflowId", workflowId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.Logger.Errorw("request err, CancelWorkflow", "err", err, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CancelWorkflow", "workflowId", workflowId, "pipelineId", pipelineId)

	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelWorkflow", "err", err, "workflowId", workflowId, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.ciHandler.CancelBuild(workflowId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelWorkflow", "err", err, "workflowId", workflowId, "pipelineId", pipelineId)
		if util.IsErrNoRows(err) {
			writeJsonResp(w, err, nil, http.StatusNotFound)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

//FIXME check if deprecated
func (handler PipelineConfigRestHandlerImpl) FetchChanges(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	ciMaterialId, err := strconv.Atoi(vars["ciMaterialId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchChanges", "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("request err, FetchChanges", "err", err, "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	changeRequest := &gitSensor.FetchScmChangesRequest{
		PipelineMaterialId: ciMaterialId,
	}
	changes, err := handler.gitSensorClient.FetchChanges(changeRequest)
	if err != nil {
		handler.Logger.Errorw("service err, FetchChanges", "err", err, "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, changes.Commits, http.StatusCreated)
}

func (handler PipelineConfigRestHandlerImpl) GetHistoricBuildLogs(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.Logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		handler.Logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetHistoricBuildLogs", "pipelineId", pipelineId, "workflowId", workflowId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetHistoricBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.GetHistoricBuildLogs(pipelineId, workflowId, nil)
	if err != nil {
		handler.Logger.Errorw("service err, GetHistoricBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId)
		writeJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetBuildHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	offsetQueryParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetQueryParam)
	if offsetQueryParam == "" || err != nil {
		writeJsonResp(w, err, "invalid offset", http.StatusBadRequest)
		return
	}
	sizeQueryParam := r.URL.Query().Get("size")
	limit, err := strconv.Atoi(sizeQueryParam)
	if sizeQueryParam == "" || err != nil {
		writeJsonResp(w, err, "invalid size", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetBuildHistory", "pipelineId", pipelineId, "offset", offset)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetBuildHistory", "err", err, "pipelineId", pipelineId, "offset", offset)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.ciHandler.GetBuildHistory(pipelineId, offset, limit)
	if err != nil {
		handler.Logger.Errorw("service err, GetBuildHistory", "err", err, "pipelineId", pipelineId, "offset", offset)
		writeJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) HandleWorkflowWebhook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var wfUpdateReq v1alpha1.WorkflowStatus
	err := decoder.Decode(&wfUpdateReq)
	if err != nil {
		handler.Logger.Errorw("request err, HandleWorkflowWebhook", "err", err, "payload", wfUpdateReq)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, HandleWorkflowWebhook", "payload", wfUpdateReq)
	resp, err := handler.ciHandler.UpdateWorkflow(wfUpdateReq)
	if err != nil {
		handler.Logger.Errorw("service err, HandleWorkflowWebhook", "err", err, "payload", wfUpdateReq)
		writeJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetBuildLogs", "pipelineId", pipelineId, "workflowId", workflowId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	lastSeenMsgId := -1
	lastEventId := r.Header.Get("Last-Event-ID")
	if len(lastEventId) > 0 {
		lastSeenMsgId, err = strconv.Atoi(lastEventId)
		if err != nil {
			handler.Logger.Errorw("request err, GetBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId, "lastEventId", lastEventId)
			writeJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	logsReader, cleanUp, err := handler.ciHandler.GetRunningWorkflowLogs(pipelineId, workflowId)
	if err != nil {
		handler.Logger.Errorw("service err, GetBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId, "lastEventId", lastEventId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
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
	defer cancel()
	defer cleanUp()
	handler.streamOutput(w, logsReader, lastSeenMsgId)
}

func (handler *PipelineConfigRestHandlerImpl) streamOutput(w http.ResponseWriter, reader *bufio.Reader, lastSeenMsgId int) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "unexpected server doesnt support streaming", http.StatusInternalServerError)
	}

	// Important to make it work in browsers
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	//var wroteHeader bool
	startOfStream := []byte("START_OF_STREAM")
	endOfStreamEvent := []byte("END_OF_STREAM")
	reconnectEvent := []byte("RECONNECT_STREAM")
	unexpectedEndOfStreamEvent := []byte("UNEXPECTED_END_OF_STREAM")
	streamStarted := false
	msgCounter := 0
	if lastSeenMsgId == -1 {
		handler.sendData(startOfStream, w, msgCounter)
		handler.sendEvent(startOfStream, w)
		f.Flush()
	} else {
		handler.sendEvent(reconnectEvent, w)
		f.Flush()
	}

	for {
		data, err := reader.ReadBytes('\n')
		if err == io.EOF {
			if streamStarted {
				handler.sendData(endOfStreamEvent, w, msgCounter)
				handler.sendEvent(endOfStreamEvent, w)
				f.Flush()
				return
			}
			return
		}
		if err != nil {
			//TODO handle error
			handler.sendData(unexpectedEndOfStreamEvent, w, msgCounter)
			handler.sendEvent(unexpectedEndOfStreamEvent, w)
			f.Flush()
			return
		}
		msgCounter = msgCounter + 1
		//skip for seen msg
		if msgCounter <= lastSeenMsgId {
			continue
		}
		if strings.Contains(string(data), devtron) {
			continue
		}

		var res []byte
		res = append(res, "id:"...)
		res = append(res, fmt.Sprintf("%d\n", msgCounter)...)
		res = append(res, "data:"...)
		res = append(res, data...)
		res = append(res, '\n')

		if _, err = w.Write(res); err != nil {
			//TODO handle error
			handler.Logger.Errorw("Failed to send response chunk, streamOutput", "err", err)
			handler.sendData(unexpectedEndOfStreamEvent, w, msgCounter)
			handler.sendEvent(unexpectedEndOfStreamEvent, w)
			f.Flush()
			return
		}
		streamStarted = true
		f.Flush()
	}
}

func (handler *PipelineConfigRestHandlerImpl) sendEvent(event []byte, w http.ResponseWriter) {
	var res []byte
	res = append(res, "event:"...)
	res = append(res, event...)
	res = append(res, '\n')
	res = append(res, "data:"...)
	res = append(res, '\n', '\n')

	if _, err := w.Write(res); err != nil {
		handler.Logger.Debugf("Failed to send response chunk: %v", err)
		return
	}

}
func (handler *PipelineConfigRestHandlerImpl) sendData(event []byte, w http.ResponseWriter, msgId int) {
	var res []byte
	res = append(res, "id:"...)
	res = append(res, fmt.Sprintf("%d\n", msgId)...)
	res = append(res, "data:"...)
	res = append(res, event...)
	res = append(res, '\n', '\n')
	if _, err := w.Write(res); err != nil {
		handler.Logger.Errorw("Failed to send response chunk, sendData", "err", err)
		return
	}
}

func (handler *PipelineConfigRestHandlerImpl) handleForwardResponseStreamError(wroteHeader bool, w http.ResponseWriter, err error) {
	code := "000"
	if !wroteHeader {
		s, ok := status.FromError(err)
		if !ok {
			s = status.New(codes.Unknown, err.Error())
		}
		w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
		code = fmt.Sprint(s.Code())
	}
	response := bean2.Response{}
	apiErr := bean2.ApiError{}
	apiErr.Code = code // 000=unknown
	apiErr.InternalMessage = err.Error()
	response.Errors = []bean2.ApiError{apiErr}
	buf, merr := json.Marshal(response)
	if merr != nil {
		handler.Logger.Errorw("marshal err, handleForwardResponseStreamError", "err", merr, "response", response)
	}
	if _, werr := w.Write(buf); werr != nil {
		handler.Logger.Errorw("Failed to notify error to client, handleForwardResponseStreamError", "err", werr, "response", response)
		return
	}
}

func (handler PipelineConfigRestHandlerImpl) GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	teamId := v.Get("teamId")
	handler.Logger.Infow("request payload, GetAppListForAutocomplete", "teamId", teamId)
	var apps []pipeline.AppBean
	if len(teamId) == 0 {
		apps, err = handler.pipelineBuilder.GetAppList()
		if err != nil {
			handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		teamId, err := strconv.Atoi(teamId)
		if err != nil {
			writeJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			apps, err = handler.pipelineBuilder.FindAppsByTeamId(teamId)
			if err != nil {
				handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
				writeJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
		}
	}

	token := r.Header.Get("token")
	var accessedApps []pipeline.AppBean
	// RBAC
	objects := handler.enforcerUtil.GetRbacObjectsForAllApps()
	for _, app := range apps {
		object := objects[app.Id]
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); ok {
			accessedApps = append(accessedApps, app)
		}
	}
	// RBAC
	if len(accessedApps) == 0 {
		accessedApps = make([]pipeline.AppBean, 0)
	}
	writeJsonResp(w, err, accessedApps, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetAppListByTeamIds(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	//vars := mux.Vars(r)
	v := r.URL.Query()
	params := v.Get("teamIds")
	if len(params) == 0 {
		writeJsonResp(w, err, "StatusBadRequest", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetAppListByTeamIds", "payload", params)
	var teamIds []int
	teamIdList := strings.Split(params, ",")
	for _, item := range teamIdList {
		teamId, err := strconv.Atoi(item)
		if err != nil {
			writeJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		teamIds = append(teamIds, teamId)
	}
	projectWiseApps, err := handler.pipelineBuilder.GetAppListByTeamIds(teamIds)
	if err != nil {
		handler.Logger.Errorw("service err, GetAppListByTeamIds", "err", err, "payload", params)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")
	// RBAC
	for _, project := range projectWiseApps {
		var accessedApps []*pipeline.AppBean
		for _, app := range project.AppList {
			object := fmt.Sprintf("%s/%s", project.ProjectName, app.Name)
			if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); ok {
				accessedApps = append(accessedApps, app)
			}
		}
		if len(accessedApps) == 0 {
			accessedApps = make([]*pipeline.AppBean, 0)
		}
		project.AppList = accessedApps
	}
	// RBAC
	writeJsonResp(w, err, projectWiseApps, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) validForMultiMaterial(ciTriggerRequest bean.CiTriggerRequest) bool {
	if len(ciTriggerRequest.CiPipelineMaterial) > 1 {
		for _, m := range ciTriggerRequest.CiPipelineMaterial {
			if m.GitCommit.Commit == "" {
				return false
			}
		}
	}
	return true
}

func (handler PipelineConfigRestHandlerImpl) EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvironmentListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	result, err := handler.envService.GetEnvironmentListForAutocomplete()
	if err != nil {
		handler.Logger.Errorw("service err, EnvironmentListAutocomplete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GitListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GitListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	res, err := handler.gitRegistryConfig.GetAll()
	if err != nil {
		handler.Logger.Errorw("service err, GitListAutocomplete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) DockerListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, DockerListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	res, err := handler.dockerRegistryConfig.ListAllActive()
	if err != nil {
		handler.Logger.Errorw("service err, DockerListAutocomplete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) TeamListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, TeamListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	result, err := handler.teamService.FetchForAutocomplete()
	if err != nil {
		handler.Logger.Errorw("service err, TeamListAutocomplete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) IsReadyToTrigger(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, IsReadyToTrigger", "appId", appId, "envId", envId, "pipelineId", pipelineId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, strings.ToLower(object)); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	result, err := handler.chartService.IsReadyToTrigger(appId, envId, pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, IsReadyToTrigger", "err", err, "appId", appId, "envId", envId, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchCDPipelineStrategy(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchCDPipelineStrategy", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	result, err := handler.pipelineBuilder.FetchCDPipelineStrategy(appId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchCDPipelineStrategy", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpgradeForAllApps(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	chartRefId, err := strconv.Atoi(vars["chartRefId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var chartUpgradeRequest pipeline.ChartUpgradeRequest
	err = decoder.Decode(&chartUpgradeRequest)
	if err != nil {
		handler.Logger.Errorw("request err, UpgradeForAllApps", "err", err, "payload", chartUpgradeRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	chartUpgradeRequest.ChartRefId = chartRefId
	chartUpgradeRequest.UserId = userId
	handler.Logger.Infow("request payload, UpgradeForAllApps", "payload", chartUpgradeRequest)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, "*/*"); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, "*/*"); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	newAppOverride, err := handler.chartService.GetAppOverrideForDefaultTemplate(chartUpgradeRequest.ChartRefId)
	if err != nil {
		handler.Logger.Errorw("service err, UpgradeForAllApps", "err", err, "payload", chartUpgradeRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
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
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
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
	writeJsonResp(w, err, response, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideReset(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvConfigOverrideReset", "appId", appId, "environmentId", environmentId)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideReset", "err", err, "appId", appId, "environmentId", environmentId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, environmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	isSuccess, err := handler.propertiesConfigService.ResetEnvironmentProperties(id)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideReset", "err", err, "appId", appId, "environmentId", environmentId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, isSuccess, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvConfigOverrideCreateNamespace(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var envConfigProperties pipeline.EnvironmentProperties
	err = decoder.Decode(&envConfigProperties)
	envConfigProperties.UserId = userId
	envConfigProperties.EnvironmentId = environmentId
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvConfigOverrideCreateNamespace", "appId", appId, "environmentId", environmentId, "payload", envConfigProperties)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, environmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.propertiesConfigService.CreateEnvironmentPropertiesWithNamespace(appId, &envConfigProperties)
	if err != nil {
		handler.Logger.Errorw("service err, EnvConfigOverrideCreateNamespace", "err", err, "appId", appId, "environmentId", environmentId, "payload", envConfigProperties)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) AppMetricsEnableDisable(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	var request pipeline.AppMetricEnableDisableRequest
	err = decoder.Decode(&request)
	request.AppId = appId
	request.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, AppMetricsEnableDisable", "err", err, "appId", appId, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, AppMetricsEnableDisable", "err", err, "appId", appId, "payload", request)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
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
	createResp, err := handler.chartService.AppMetricsEnableDisable(request)
	if err != nil {
		handler.Logger.Errorw("service err, AppMetricsEnableDisable", "err", err, "appId", appId, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvMetricsEnableDisable(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request pipeline.AppMetricEnableDisableRequest
	err = decoder.Decode(&request)
	request.UserId = userId
	request.AppId = appId
	request.EnvironmentId = environmentId
	if err != nil {
		handler.Logger.Errorw("request err, EnvMetricsEnableDisable", "err", err, "appId", appId, "environmentId", environmentId, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvMetricsEnableDisable", "err", err, "appId", appId, "environmentId", environmentId, "payload", request)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, request.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.propertiesConfigService.EnvMetricsEnableDisable(&request)
	if err != nil {
		handler.Logger.Errorw("service err, EnvMetricsEnableDisable", "err", err, "appId", appId, "environmentId", environmentId, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, createResp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetCdBuildHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	offsetQueryParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetQueryParam)
	if offsetQueryParam == "" || err != nil {
		handler.Logger.Errorw("request err, GetCdBuildHistory", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
		writeJsonResp(w, err, "invalid offset", http.StatusBadRequest)
		return
	}
	sizeQueryParam := r.URL.Query().Get("size")
	limit, err := strconv.Atoi(sizeQueryParam)
	if sizeQueryParam == "" || err != nil {
		handler.Logger.Errorw("request err, GetCdBuildHistory", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "sizeQueryParam", sizeQueryParam)
		writeJsonResp(w, err, "invalid size", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCdBuildHistory", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	resp, err := handler.cdHandelr.GetCdBuildHistory(appId, environmentId, pipelineId, offset, limit)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdBuildHistory", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
		writeJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetCdBuildLogs(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCdBuildLogs", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "workflowId", workflowId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	logsReader, cleanUp, err := handler.cdHandelr.GetRunningWorkflowLogs(environmentId, pipelineId, workflowId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdBuildLogs", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "workflowId", workflowId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	lastSeenMsgId := -1
	lastEventId := r.Header.Get("Last-Event-ID")
	if len(lastEventId) > 0 {
		lastSeenMsgId, err = strconv.Atoi(lastEventId)
		if err != nil {
			handler.Logger.Errorw("request err, GetCdBuildLogs", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "workflowId", workflowId, "lastEventId", lastEventId)
			writeJsonResp(w, err, nil, http.StatusBadRequest)
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
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowRunnerId"])
	if err != nil || buildId == 0 {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchCdWorkflowDetails", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "buildId", buildId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	resp, err := handler.cdHandelr.FetchCdWorkflowDetails(appId, environmentId, pipelineId, buildId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchCdWorkflowDetails", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no workflow found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) DownloadCdWorkflowArtifacts(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowRunnerId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, DownloadCdWorkflowArtifacts", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(appId, pipelineId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	file, err := handler.cdHandelr.DownloadCdWorkflowArtifacts(pipelineId, buildId)
	defer file.Close()

	if err != nil {
		handler.Logger.Errorw("service err, DownloadCdWorkflowArtifacts", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no workflow found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Itoa(buildId)+".zip")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))
	_, err = io.Copy(w, file)
	if err != nil {
		handler.Logger.Errorw("service err, DownloadCdWorkflowArtifacts", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId)
	}
}

func (handler PipelineConfigRestHandlerImpl) FetchCdPrePostStageStatus(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchCdPrePostStageStatus", "err", err, "appId", appId, "pipelineId", pipelineId)

	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(appId, pipelineId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	resp, err := handler.cdHandelr.FetchCdPrePostStageStatus(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchCdPrePostStageStatus", "err", err, "appId", appId, "pipelineId", pipelineId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no status found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchConfigmapSecretsForCdStages(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchConfigmapSecretsForCdStages", "err", err, "pipelineId", pipelineId)
	pipeline, err := handler.pipelineBuilder.FindPipelineById(pipelineId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	resp, err := handler.pipelineBuilder.FetchConfigmapSecretsForCdStages(pipeline.AppId, pipeline.EnvironmentId, pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchConfigmapSecretsForCdStages", "err", err, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCdPipelineById(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCdPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciConf, err := handler.pipelineBuilder.GetCdPipelineById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCdPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchAppWorkflowStatusForTriggerView(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchAppWorkflowStatusForTriggerView", "err", err, "appId", appId)
	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	triggerWorkflowStatus := pipelineConfig.TriggerWorkflowStatus{}
	ciWorkflowStatus, err := handler.ciHandler.FetchCiStatusForTriggerView(appId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppWorkflowStatusForTriggerView", "err", err, "appId", appId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no workflow found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}

	cdWorkflowStatus, err := handler.cdHandelr.FetchAppWorkflowStatusForTriggerView(appId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppWorkflowStatusForTriggerView", "err", err, "appId", appId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no status found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	triggerWorkflowStatus.CiWorkflowStatus = ciWorkflowStatus
	triggerWorkflowStatus.CdWorkflowStatus = cdWorkflowStatus
	writeJsonResp(w, err, triggerWorkflowStatus, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchMaterialInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	ciArtifactId, err := strconv.Atoi(vars["ciArtifactId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchMaterialInfo", "err", err, "ciArtifactId", ciArtifactId)
	resp, err := handler.ciHandler.FetchMaterialInfoByArtifactId(ciArtifactId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchMaterialInfo", "err", err, "ciArtifactId", ciArtifactId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no material info found"}
			writeJsonResp(w, err, nil, http.StatusOK)
		} else {
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(resp.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	writeJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCIPipelineById(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetCIPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Infow("service error, GetCIPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciPipeline, err := handler.pipelineBuilder.GetCiPipelineById(pipelineId)
	if err != nil {
		handler.Logger.Infow("service error, GetCIPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, ciPipeline, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) PipelineNameSuggestion(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pType := vars["type"]
	handler.Logger.Infow("request payload, PipelineNameSuggestion", "err", err, "appId", appId)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Infow("service error, GetCIPipelineById", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	suggestedName := fmt.Sprintf("%s-%d-%s", pType, appId, util2.Generate(4))
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	writeJsonResp(w, err, suggestedName, http.StatusOK)
}
