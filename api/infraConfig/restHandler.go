package infraConfig

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
)

const InvalidProfileRequest = "requested profile doesn't exist"

type InfraConfigRestHandler interface {
	UpdateInfraProfile(w http.ResponseWriter, r *http.Request)
	GetProfile(w http.ResponseWriter, r *http.Request)
	DeleteProfile(w http.ResponseWriter, r *http.Request)
	CreateProfile(w http.ResponseWriter, r *http.Request)
	GetProfileList(w http.ResponseWriter, r *http.Request)
}

type InfraConfigRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	infraProfileService service.InfraConfigService
	userService         user.UserService
	enforcer            casbin.Enforcer
	enforcerUtil        rbac.EnforcerUtil
}

func NewInfraConfigRestHandlerImpl(logger *zap.SugaredLogger, infraProfileService service.InfraConfigService, userService user.UserService, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil) *InfraConfigRestHandlerImpl {
	return &InfraConfigRestHandlerImpl{
		logger:              logger,
		infraProfileService: infraProfileService,
		userService:         userService,
		enforcer:            enforcer,
		enforcerUtil:        enforcerUtil,
	}
}

func (handler *InfraConfigRestHandlerImpl) CreateProfile(w http.ResponseWriter, r *http.Request) {

}

func (handler *InfraConfigRestHandlerImpl) UpdateInfraProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileName := vars["name"]
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(service.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	payload := &infraConfig.ProfileBean{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(payload)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.infraProfileService.UpdateProfile(userId, profileName, payload)
	if err != nil {
		handler.logger.Errorw("error in updating profile and configurations", "profileName", profileName, "payLoad", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) GetProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileName := vars["name"]
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(service.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}

	defaultProfile, err := handler.infraProfileService.GetDefaultProfile()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	profile, err := handler.infraProfileService.GetProfileByName(profileName)

	resp := infraConfig.ProfileResponse{
		Profile: *profile,
	}
	resp.DefaultConfigurations = defaultProfile.Configurations
	resp.ConfigurationUnits = handler.infraProfileService.GetConfigurationUnits()
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) GetProfileList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// vars := mux.Vars(r)
	// profileNameLike := vars["profileNameLike"]

}

func (handler *InfraConfigRestHandlerImpl) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileName := vars["name"]
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(service.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	err = handler.infraProfileService.DeleteProfile(profileName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
