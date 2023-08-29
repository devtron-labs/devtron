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
	"golang.org/x/exp/slices"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"regexp"
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
	Payload    *repository.Payload `json:"payload"`
	JsonSchema string              `json:"jsonSchema"`
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
	decoder.UseNumber()
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
		return
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
	var scope repository.Scope
	scopeQueryParam := r.URL.Query().Get("scope")
	if scopeQueryParam != "" {
		if err := json.Unmarshal([]byte(scopeQueryParam), &scope); err != nil {
			http.Error(w, "Invalid JSON format for 'scope' parameter", http.StatusBadRequest)
			return
		}

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
	//if scope.AppId == 0 && scope.EnvId == 0 && scope.ClusterId == 0 {
	//	http.Error(w, "scope is empty", http.StatusBadRequest)
	//	return
	//}
	var scopedVariableData []*variables.ScopedVariableData

	scopedVariableData, err = handler.scopedVariableService.GetScopedVariables(scope, nil)
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
	var paylaod *repository.Payload
	var jsonSchema string

	paylaod, jsonSchema, err = handler.scopedVariableService.GetJsonForVariables()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	jsonResponse := JsonResponse{
		Payload:    paylaod,
		JsonSchema: jsonSchema,
	}
	common.WriteJsonResp(w, nil, jsonResponse, http.StatusOK)
}
func getIdentifierType(attribute repository.AttributeType) []repository.IdentifierType {
	switch attribute {
	case repository.ApplicationEnv:
		return []repository.IdentifierType{repository.ApplicationName, repository.EnvName}
	case repository.Application:
		return []repository.IdentifierType{repository.ApplicationName}
	case repository.Env:
		return []repository.IdentifierType{repository.EnvName}
	case repository.Cluster:
		return []repository.IdentifierType{repository.ClusterName}
	default:
		return nil
	}
}
func validateVariableScopeRequest(payload repository.Payload) error {
	variableNamesList := make([]string, 0)
	for _, variable := range payload.Variables {
		if slices.Contains(variableNamesList, variable.Definition.VarName) {
			return fmt.Errorf("duplicate variable name")
		}
		exp := `^[a-zA-Z0-9_-]{1,64}$`
		rExp := regexp.MustCompile(exp)
		if !rExp.MatchString(variable.Definition.VarName) {
			return fmt.Errorf("invalid variable name")
		}
		if variable.Definition.VarName[0] == '_' ||
			variable.Definition.VarName[0] == '-' ||
			variable.Definition.VarName[len(variable.Definition.VarName)-1] == '_' ||
			variable.Definition.VarName[len(variable.Definition.VarName)-1] == '-' {
			return fmt.Errorf("invalid variable name")
		}
		variableNamesList = append(variableNamesList, variable.Definition.VarName)
		uniqueVariableMap := make(map[string]interface{})
		for _, attributeValue := range variable.AttributeValues {
			validIdentifierTypeList := getIdentifierType(attributeValue.AttributeType)
			if len(validIdentifierTypeList) != len(attributeValue.AttributeParams) {
				return fmt.Errorf("length of AttributeParams is not valid")
			}
			for key, _ := range attributeValue.AttributeParams {
				if !slices.Contains(validIdentifierTypeList, key) {
					return fmt.Errorf("invalid IdentifierType %s for validIdentifierTypeList %s", key, validIdentifierTypeList)
				}
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
			identifierString := fmt.Sprintf("%s-%s", variable.Definition.VarName, string(attributeValue.AttributeType))
			for _, key := range validIdentifierTypeList {
				identifierString = fmt.Sprintf("%s-%s", identifierString, attributeValue.AttributeParams[key])
			}
			if _, ok := uniqueVariableMap[identifierString]; ok {
				return fmt.Errorf("duplicate AttributeParams found for AttributeType %v", attributeValue.AttributeType)
			}
			uniqueVariableMap[identifierString] = attributeValue.VariableValue.Value
		}
	}
	return nil
}
