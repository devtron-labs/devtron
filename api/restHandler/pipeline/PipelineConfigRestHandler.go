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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"io"
	"net/http"
	"strconv"
	"strings"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
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
)

type DevtronAppAutoCompleteRestHandler interface {
	GitListAutocomplete(w http.ResponseWriter, r *http.Request)
	DockerListAutocomplete(w http.ResponseWriter, r *http.Request)
	TeamListAutocomplete(w http.ResponseWriter, r *http.Request)
	EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request)
	GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request)
}

type DevtronAppRestHandler interface {
	CreateApp(w http.ResponseWriter, r *http.Request)
	DeleteApp(w http.ResponseWriter, r *http.Request)
	GetApp(w http.ResponseWriter, r *http.Request)

	FindAppsByTeamId(w http.ResponseWriter, r *http.Request)
	FindAppsByTeamName(w http.ResponseWriter, r *http.Request)
	GetAppListByTeamIds(w http.ResponseWriter, r *http.Request)
}

type DevtronAppWorkflowRestHandler interface {
	FetchWorkflowDetails(w http.ResponseWriter, r *http.Request)
	FetchAppWorkflowStatusForTriggerView(w http.ResponseWriter, r *http.Request)
}

type PipelineConfigRestHandler interface {
	DevtronAppAutoCompleteRestHandler
	DevtronAppRestHandler
	DevtronAppWorkflowRestHandler
	DevtronAppBuildRestHandler
	DevtronAppBuildMaterialRestHandler
	DevtronAppBuildHistoryRestHandler
	DevtronAppDeploymentRestHandler
	DevtronAppDeploymentHistoryRestHandler
	DevtronAppPrePostDeploymentRestHandler
	DevtronAppDeploymentConfigRestHandler

	EnvConfigOverrideCreateNamespace(w http.ResponseWriter, r *http.Request)
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
	cdHandler               pipeline.CdHandler
	appCloneService         appClone.AppCloneService
	materialRepository      pipelineConfig.MaterialRepository
	policyService           security2.PolicyService
	scanResultRepository    security.ImageScanResultRepository
	gitProviderRepo         repository.GitProviderRepository
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
	cdHandler pipeline.CdHandler,
	appCloneService appClone.AppCloneService,
	appWorkflowService appWorkflow.AppWorkflowService,
	materialRepository pipelineConfig.MaterialRepository, policyService security2.PolicyService,
	scanResultRepository security.ImageScanResultRepository, gitProviderRepo repository.GitProviderRepository) *PipelineConfigRestHandlerImpl {
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
		cdHandler:               cdHandler,
		appCloneService:         appCloneService,
		appWorkflowService:      appWorkflowService,
		materialRepository:      materialRepository,
		policyService:           policyService,
		scanResultRepository:    scanResultRepository,
		gitProviderRepo:         gitProviderRepo,
	}
}

const (
	devtron          = "DEVTRON"
	SSH_URL_PREFIX   = "git@"
	HTTPS_URL_PREFIX = "https://"
)

func (handler PipelineConfigRestHandlerImpl) DeleteApp(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, delete app", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, delete app", "appId", appId)
	wfs, err := handler.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		handler.Logger.Errorw("could not fetch wfs", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if len(wfs) != 0 {
		handler.Logger.Info("cannot delete app with workflow's")
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "cannot delete app having workflow's"}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	resourceObject := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, resourceObject); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	err = handler.pipelineBuilder.DeleteApp(appId, userId)
	if err != nil {
		handler.Logger.Errorw("service error, delete app", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateApp(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var createRequest bean.CreateAppDTO
	err = decoder.Decode(&createRequest)
	createRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateApp", "err", err, "CreateApp", createRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Infow("request payload, CreateApp", "CreateApp", createRequest)
	err = handler.validator.Struct(createRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateApp", "err", err, "CreateApp", createRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	project, err := handler.teamService.FetchOne(createRequest.TeamId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// with admin roles, you have to access for all the apps of the project to create new app. (admin or manager with specific app permission can't create app.)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, fmt.Sprintf("%s/%s", strings.ToLower(project.Name), "*")); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
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
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) ValidateGitMaterialUrl(gitProviderId int, url string) (bool, error) {
	gitProvider, err := handler.gitProviderRepo.FindOne(strconv.Itoa(gitProviderId))
	if err != nil {
		return false, err
	}
	if gitProvider.AuthMode == repository.AUTH_MODE_SSH {
		hasPrefixResult := strings.HasPrefix(url, SSH_URL_PREFIX)
		return hasPrefixResult, nil
	}
	hasPrefixResult := strings.HasPrefix(url, HTTPS_URL_PREFIX)
	return hasPrefixResult, nil
}

func (handler PipelineConfigRestHandlerImpl) GetApp(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, get app", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, get app", "appId", appId)
	ciConf, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, get app", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac implementation starts here
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac implementation ends here

	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FindAppsByTeamId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamId, err := strconv.Atoi(vars["teamId"])
	if err != nil {
		handler.Logger.Errorw("request err, FindAppsByTeamId", "err", err, "teamId", teamId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FindAppsByTeamId", "teamId", teamId)
	project, err := handler.pipelineBuilder.FindAppsByTeamId(teamId)
	if err != nil {
		handler.Logger.Errorw("service err, FindAppsByTeamId", "err", err, "teamId", teamId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, project, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FindAppsByTeamName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["teamName"]
	handler.Logger.Infow("request payload, FindAppsByTeamName", "teamName", teamName)
	project, err := handler.pipelineBuilder.FindAppsByTeamName(teamName)
	if err != nil {
		handler.Logger.Errorw("service err, FindAppsByTeamName", "err", err, "teamName", teamName)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, project, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchWorkflowDetails(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
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
	buildId, err := strconv.Atoi(vars["workflowId"])
	if err != nil || buildId == 0 {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchWorkflowDetails", "appId", appId, "pipelineId", pipelineId, "buildId", buildId, "buildId", buildId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.FetchWorkflowDetails(appId, pipelineId, buildId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchWorkflowDetails", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId, "buildId", buildId)
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
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

func (handler PipelineConfigRestHandlerImpl) CancelWorkflow(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		handler.Logger.Errorw("request err, CancelWorkflow", "err", err, "workflowId", workflowId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.Logger.Errorw("request err, CancelWorkflow", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CancelWorkflow", "workflowId", workflowId, "pipelineId", pipelineId)

	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelWorkflow", "err", err, "workflowId", workflowId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.ciHandler.CancelBuild(workflowId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelWorkflow", "err", err, "workflowId", workflowId, "pipelineId", pipelineId)
		if util.IsErrNoRows(err) {
			common.WriteJsonResp(w, err, nil, http.StatusNotFound)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

// FetchChanges FIXME check if deprecated
func (handler PipelineConfigRestHandlerImpl) FetchChanges(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	ciMaterialId, err := strconv.Atoi(vars["ciMaterialId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchChanges", "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("request err, FetchChanges", "err", err, "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	changeRequest := &gitSensor.FetchScmChangesRequest{
		PipelineMaterialId: ciMaterialId,
	}
	changes, err := handler.gitSensorClient.FetchChanges(changeRequest)
	if err != nil {
		handler.Logger.Errorw("service err, FetchChanges", "err", err, "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, changes.Commits, http.StatusCreated)
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
	buf, err2 := json.Marshal(response)
	if err2 != nil {
		handler.Logger.Errorw("marshal err, handleForwardResponseStreamError", "err", err2, "response", response)
	}
	if _, err3 := w.Write(buf); err3 != nil {
		handler.Logger.Errorw("Failed to notify error to client, handleForwardResponseStreamError", "err", err3, "response", response)
		return
	}
}

func (handler PipelineConfigRestHandlerImpl) GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
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
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		teamId, err := strconv.Atoi(teamId)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			apps, err = handler.pipelineBuilder.FindAppsByTeamId(teamId)
			if err != nil {
				handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
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
	common.WriteJsonResp(w, err, accessedApps, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetAppListByTeamIds(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	//vars := mux.Vars(r)
	v := r.URL.Query()
	params := v.Get("teamIds")
	if len(params) == 0 {
		common.WriteJsonResp(w, err, "StatusBadRequest", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetAppListByTeamIds", "payload", params)
	var teamIds []int
	teamIdList := strings.Split(params, ",")
	for _, item := range teamIdList {
		teamId, err := strconv.Atoi(item)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		teamIds = append(teamIds, teamId)
	}
	projectWiseApps, err := handler.pipelineBuilder.GetAppListByTeamIds(teamIds)
	if err != nil {
		handler.Logger.Errorw("service err, GetAppListByTeamIds", "err", err, "payload", params)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
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
	common.WriteJsonResp(w, err, projectWiseApps, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvironmentListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	result, err := handler.envService.GetEnvironmentListForAutocomplete()
	if err != nil {
		handler.Logger.Errorw("service err, EnvironmentListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GitListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GitListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	res, err := handler.gitRegistryConfig.GetAll()
	if err != nil {
		handler.Logger.Errorw("service err, GitListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) DockerListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, DockerListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	res, err := handler.dockerRegistryConfig.ListAllActive()
	if err != nil {
		handler.Logger.Errorw("service err, DockerListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) TeamListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, TeamListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	result, err := handler.teamService.FetchForAutocomplete()
	if err != nil {
		handler.Logger.Errorw("service err, TeamListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, strings.ToLower(object)); !ok {
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
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
	var chartUpgradeRequest pipeline.ChartUpgradeRequest
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, "*/*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, "*/*"); !ok {
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, environmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, object); !ok {
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

func (handler PipelineConfigRestHandlerImpl) FetchAppWorkflowStatusForTriggerView(w http.ResponseWriter, r *http.Request) {
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
	handler.Logger.Infow("request payload, FetchAppWorkflowStatusForTriggerView", "err", err, "appId", appId)
	//RBAC CHECK
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC CHECK

	triggerWorkflowStatus := pipelineConfig.TriggerWorkflowStatus{}
	ciWorkflowStatus, err := handler.ciHandler.FetchCiStatusForTriggerView(appId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppWorkflowStatusForTriggerView", "err", err, "appId", appId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no workflow found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}

	cdWorkflowStatus, err := handler.cdHandler.FetchAppWorkflowStatusForTriggerView(appId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppWorkflowStatusForTriggerView", "err", err, "appId", appId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no status found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	triggerWorkflowStatus.CiWorkflowStatus = ciWorkflowStatus
	triggerWorkflowStatus.CdWorkflowStatus = cdWorkflowStatus
	common.WriteJsonResp(w, err, triggerWorkflowStatus, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) PipelineNameSuggestion(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pType := vars["type"]
	handler.Logger.Infow("request payload, PipelineNameSuggestion", "err", err, "appId", appId)
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Infow("service error, GetCIPipelineById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	suggestedName := fmt.Sprintf("%s-%d-%s", pType, appId, util2.Generate(4))
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	common.WriteJsonResp(w, err, suggestedName, http.StatusOK)
}
