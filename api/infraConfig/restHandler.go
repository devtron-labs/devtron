package infraConfig

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

const InvalidIdentifierType = "identifier %s is not valid"

type InfraConfigRestHandler interface {
	UpdateInfraProfile(w http.ResponseWriter, r *http.Request)
	GetProfile(w http.ResponseWriter, r *http.Request)
	DeleteProfile(w http.ResponseWriter, r *http.Request)
	CreateProfile(w http.ResponseWriter, r *http.Request)
	GetProfileList(w http.ResponseWriter, r *http.Request)
	GetIdentifierList(w http.ResponseWriter, r *http.Request)
	ApplyProfileToIdentifiers(w http.ResponseWriter, r *http.Request)
}

type InfraConfigRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	infraProfileService infraConfig.InfraConfigService
	userService         user.UserService
	enforcer            casbin.Enforcer
	enforcerUtil        rbac.EnforcerUtil
}

func NewInfraConfigRestHandlerImpl(logger *zap.SugaredLogger, infraProfileService infraConfig.InfraConfigService, userService user.UserService, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil) *InfraConfigRestHandlerImpl {
	return &InfraConfigRestHandlerImpl{
		logger:              logger,
		infraProfileService: infraProfileService,
		userService:         userService,
		enforcer:            enforcer,
		enforcerUtil:        enforcerUtil,
	}
}

func (handler *InfraConfigRestHandlerImpl) CreateProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
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

	err = handler.infraProfileService.CreateProfile(userId, payload)
	if err != nil {
		handler.logger.Errorw("error in creating profile and configurations", "payLoad", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
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
	profileName := strings.ToLower(vars["name"])
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(infraConfig.InvalidProfileName), nil, http.StatusBadRequest)
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
	payload.Name = strings.ToLower(payload.Name)
	err = handler.infraProfileService.UpdateProfile(userId, profileName, payload)
	if err != nil {
		handler.logger.Errorw("error in updating profile and configurations", "profileName", profileName, "payLoad", payload, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
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
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileName := strings.ToLower(vars["name"])
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(infraConfig.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}

	var profile *infraConfig.ProfileBean
	if profileName != infraConfig.DEFAULT_PROFILE_NAME {
		profile, err = handler.infraProfileService.GetProfileByName(profileName)
		if err != nil {
			statusCode := http.StatusBadRequest
			if errors.Is(err, pg.ErrNoRows) {
				err = errors.New(fmt.Sprintf("profile %s not found", profileName))
				statusCode = http.StatusNotFound
			}
			common.WriteJsonResp(w, err, nil, statusCode)
			return
		}
	}

	defaultProfile, err := handler.infraProfileService.GetDefaultProfile(false)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if profileName == infraConfig.DEFAULT_PROFILE_NAME {
		profile = defaultProfile
	}
	resp := infraConfig.ProfileResponse{
		Profile: *profile,
	}
	resp.ConfigurationUnits = handler.infraProfileService.GetConfigurationUnits()
	resp.DefaultConfigurations = defaultProfile.Configurations
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) GetProfileList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileNameLike := vars["search"]
	profilesResponse, err := handler.infraProfileService.GetProfileList(profileNameLike)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	common.WriteJsonResp(w, nil, profilesResponse, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileName := strings.ToLower(vars["name"])
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(infraConfig.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	err = handler.infraProfileService.DeleteProfile(profileName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) GetIdentifierList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	identifierType := vars["identifierType"]
	if identifierType != string(infraConfig.APPLICATION) {
		common.WriteJsonResp(w, errors.New(fmt.Sprintf(InvalidIdentifierType, identifierType)), nil, http.StatusBadRequest)
		return
	}
	identifierNameLike := vars["search"]
	sortOrder := vars["sort"]
	if sortOrder == "" {
		// set to default asc order
		sortOrder = "ASC"
	}
	if !(sortOrder == "ASC" || sortOrder == "DESC") {
		common.WriteJsonResp(w, errors.New("sort order can only be ASC or DESC"), nil, http.StatusBadRequest)
		return
	}
	sizeStr, ok := vars["size"]
	size := 20
	if ok {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 0 {
			common.WriteJsonResp(w, errors.Wrap(err, "invalid size"), nil, http.StatusBadRequest)
			return
		}
	}
	offsetStr, ok := vars["offset"]
	offset := 0
	if ok {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			common.WriteJsonResp(w, errors.Wrap(err, "invalid offset"), nil, http.StatusBadRequest)
			return
		}
	}

	profileName := strings.ToLower(vars["profileName"])
	listFilter := infraConfig.IdentifierListFilter{
		Limit:              size,
		Offset:             offset,
		SortOrder:          sortOrder,
		ProfileName:        profileName,
		IdentifierNameLike: identifierNameLike,
		IdentifierType:     infraConfig.APPLICATION,
	}

	res, err := handler.infraProfileService.GetIdentifierList(&listFilter)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, pg.ErrNoRows) {
			statusCode = http.StatusNotFound
		}
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) ApplyProfileToIdentifiers(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	identifierType := vars["identifierType"]
	if identifierType != string(infraConfig.APPLICATION) {
		common.WriteJsonResp(w, errors.New(fmt.Sprintf(InvalidIdentifierType, identifierType)), nil, http.StatusBadRequest)
		return
	}
	var request infraConfig.InfraProfileApplyRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.infraProfileService.ApplyProfileToIdentifiers(userId, request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
