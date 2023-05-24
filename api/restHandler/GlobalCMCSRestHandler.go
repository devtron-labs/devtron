package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type GlobalCMCSRestHandler interface {
	CreateGlobalCMCSConfig(w http.ResponseWriter, r *http.Request)
	UpdateGlobalCMCSDataById(w http.ResponseWriter, r *http.Request)
	GetGlobalCMCSDataByConfigTypeAndName(w http.ResponseWriter, r *http.Request)
	GetAllGlobalCMCSData(w http.ResponseWriter, r *http.Request)
	DeleteByID(w http.ResponseWriter, r *http.Request)
}

type GlobalCMCSRestHandlerImpl struct {
	logger            *zap.SugaredLogger
	userAuthService   user.UserService
	validator         *validator.Validate
	enforcer          casbin.Enforcer
	globalCMCSService pipeline.GlobalCMCSService
}

func NewGlobalCMCSRestHandlerImpl(
	logger *zap.SugaredLogger,
	userAuthService user.UserService,
	validator *validator.Validate,
	enforcer casbin.Enforcer,
	globalCMCSService pipeline.GlobalCMCSService) *GlobalCMCSRestHandlerImpl {
	return &GlobalCMCSRestHandlerImpl{
		logger:            logger,
		userAuthService:   userAuthService,
		validator:         validator,
		enforcer:          enforcer,
		globalCMCSService: globalCMCSService,
	}
}

func (handler *GlobalCMCSRestHandlerImpl) CreateGlobalCMCSConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.GlobalCMCSDto
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, CreateGlobalCMCSConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Infow("request payload, CreateGlobalCMCSConfig", "err", err, "payload", bean)
	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, CreateGlobalCMCSConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := handler.globalCMCSService.Create(&bean)
	if err != nil {
		handler.logger.Errorw("service err, CreateGlobalCMCSConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *GlobalCMCSRestHandlerImpl) UpdateGlobalCMCSDataById(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.GlobalCMCSDataUpdateDto
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, CreateGlobalCMCSConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Infow("request payload, CreateGlobalCMCSConfig", "err", err, "payload", bean)
	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, CreateGlobalCMCSConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := handler.globalCMCSService.UpdateDataById(&bean)
	if err != nil {
		handler.logger.Errorw("service err, CreateGlobalCMCSConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *GlobalCMCSRestHandlerImpl) GetGlobalCMCSDataByConfigTypeAndName(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	configName := vars["config_name"]
	configType := vars["config_type"]

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := handler.globalCMCSService.GetGlobalCMCSDataByConfigTypeAndName(configName, configType)
	if err != nil {
		handler.logger.Errorw("service err, CreateGlobalCMCSConfig", "err", err, "payload", "configName", configName, "configType", configType)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *GlobalCMCSRestHandlerImpl) GetAllGlobalCMCSData(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	res, err := handler.globalCMCSService.FindAllActive()
	if err != nil {
		handler.logger.Errorw("service err, CreateGlobalCMCSConfig", "err", err, "payload")
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
	//RBAC enforcer Ends

}

func (handler *GlobalCMCSRestHandlerImpl) DeleteByID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.WriteJsonResp(w, err, "error in parsing id for delete config reuest", http.StatusBadRequest)
		return
	}

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	err = handler.globalCMCSService.DeleteById(id)
	if err != nil {
		handler.logger.Errorw("service err, CreateGlobalCMCSConfig", "err", err, "payload")
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "Successfully deleted", http.StatusOK)
}
