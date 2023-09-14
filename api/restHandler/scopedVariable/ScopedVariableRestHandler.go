package scopedVariable

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type ScopedVariableRestHandler interface {
	CreateVariables(w http.ResponseWriter, r *http.Request)
	GetScopedVariables(w http.ResponseWriter, r *http.Request)
	GetJsonForVariables(w http.ResponseWriter, r *http.Request)
}

type ScopedVariableRestHandlerImpl struct {
	logger                *zap.SugaredLogger
	userAuthService       user.UserService
	validator             *validator.Validate
	pipelineBuilder       pipeline.PipelineBuilder
	enforcerUtil          rbac.EnforcerUtil
	enforcer              casbin.Enforcer
	scopedVariableService variables.ScopedVariableService
}
type JsonResponse struct {
	Manifest   *models.ScopedVariableManifest `json:"manifest"`
	JsonSchema string                         `json:"jsonSchema"`
}

func NewScopedVariableRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService, validator *validator.Validate, pipelineBuilder pipeline.PipelineBuilder, enforcerUtil rbac.EnforcerUtil, enforcer casbin.Enforcer, scopedVariableService variables.ScopedVariableService) *ScopedVariableRestHandlerImpl {
	return &ScopedVariableRestHandlerImpl{
		logger:                logger,
		userAuthService:       userAuthService,
		validator:             validator,
		pipelineBuilder:       pipelineBuilder,
		enforcerUtil:          enforcerUtil,
		enforcer:              enforcer,
		scopedVariableService: scopedVariableService,
	}
}
func (handler *ScopedVariableRestHandlerImpl) CreateVariables(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	request := models.VariableRequest{}
	decoder.UseNumber()
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId

	// validate request
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("struct validation err in CreateVariables", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	payload := utils.ManifestToPayload(request.Manifest, userId)

	// not logging bean object as it contains sensitive data
	handler.logger.Infow("request payload received for variables")

	// RBAC enforcer applying
	isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = handler.scopedVariableService.CreateVariables(payload)
	if err != nil {
		if errors.As(err, &models.ValidationError{}) {
			common.WriteJsonResp(w, err, nil, http.StatusNotAcceptable)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
func (handler *ScopedVariableRestHandlerImpl) GetScopedVariables(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))

	if err != nil {
		handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		return
	}

	appIdQueryParam := r.URL.Query().Get("appId")
	var appId int
	appId, err = strconv.Atoi(appIdQueryParam)
	if err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}

	var scope models.Scope
	scopeQueryParam := r.URL.Query().Get("scope")
	if scopeQueryParam != "" {
		if err := json.Unmarshal([]byte(scopeQueryParam), &scope); err != nil {
			http.Error(w, "Invalid JSON format for 'scope' parameter", http.StatusBadRequest)
			return
		}
	}
	if scope.AppId != 0 && scope.AppId != appId {
		http.Error(w, "scope.AppId provided in scope is not equal to appId", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	var app *bean.CreateAppDTO
	app, err = handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetScopedVariables", "err", err, "payload", scope.AppId, scope.EnvId, scope.ClusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, GetScopedVariables", "payload", scope.AppId, scope.EnvId, scope.ClusterId)
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	var scopedVariableData []*models.ScopedVariableData

	scopedVariableData, err = handler.scopedVariableService.GetScopedVariables(scope, nil, isSuperAdmin)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, nil, scopedVariableData, http.StatusOK)

}
func (handler *ScopedVariableRestHandlerImpl) GetJsonForVariables(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// not logging bean object as it contains sensitive data
	handler.logger.Infow("request payload received for variables")

	// RBAC enforcer applying
	isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	//var payload *repository.Payload

	payload, err := handler.scopedVariableService.GetJsonForVariables()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	schema, err := util.GetSchemaFromType(models.ScopedVariableManifest{})
	if err != nil {
		common.WriteJsonResp(w, err, "schema cannot be generated for manifest type", http.StatusInternalServerError)
		return
	}

	jsonResponse := JsonResponse{
		JsonSchema: schema,
	}

	if payload != nil {
		manifest := utils.PayloadToManifest(*payload)
		jsonResponse.Manifest = &manifest
	}
	common.WriteJsonResp(w, nil, jsonResponse, http.StatusOK)
}
