package scopedVariable

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
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
	paylaod := repository.Payload{}
	err = decoder.Decode(&paylaod)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "payload", paylaod)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	paylaod.UserId = userId
	// validate request
	err = handler.validator.Struct(paylaod)
	if err != nil {
		handler.logger.Errorw("struct validation err in CreateVariables", "err", err, "request", paylaod)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = validateVariableScopeRequest(paylaod)
	if err != nil {
		handler.logger.Errorw("custom validation err in CreateVariables", "err", err, "request", paylaod)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
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
	err = handler.scopedVariableService.CreateVariables(paylaod)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
func (handler *ScopedVariableRestHandlerImpl) GetScopedVariables(w http.ResponseWriter, r *http.Request) {
	appIdQueryParam := r.URL.Query().Get("appId")
	var appId int
	var err error
	if appIdQueryParam != "" {
		appId, err = strconv.Atoi(appIdQueryParam)
		if err != nil {
			common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
			return
		}
	}
	queryParams := r.URL.Query()
	scopeParam := queryParams.Get("scope")

	var scope variables.Scope
	if err := json.Unmarshal([]byte(scopeParam), &scope); err != nil {
		http.Error(w, "Invalid JSON format for 'scope' parameter", http.StatusBadRequest)
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
	if scope.AppId == 0 && scope.EnvId == 0 && scope.ClusterId == 0 {
		return
	}
	var scopedVariableData []*variables.ScopedVariableData

	scopedVariableData, err = handler.scopedVariableService.GetScopedVariables(scope, nil)
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
	var paylaod *repository.Payload
	paylaod, err = handler.scopedVariableService.GetJsonForVariables()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	common.WriteJsonResp(w, nil, paylaod, http.StatusOK)
}

func validateVariableScopeRequest(payload repository.Payload) error {
	for _, variable := range payload.Variables {
		for _, attributeValue := range variable.AttributeValues {
			for key, _ := range attributeValue.AttributeParams {
				match := false
				for _, identifier := range repository.IdentifiersList {
					if identifier == key {
						match = true
					}
				}
				if !match {
					return fmt.Errorf("invalid identifier key %s for variable %s", key, variable.Definition.VarName)
				}
			}
		}
	}
	return nil
}
