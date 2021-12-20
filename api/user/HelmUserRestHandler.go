package user

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/user"
	bean2 "github.com/devtron-labs/devtron/pkg/user/dto"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type HelmUserRestHandler interface {
	GetById(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	CreateHelmUser(w http.ResponseWriter, r *http.Request)
	UpdateHelmUser(w http.ResponseWriter, r *http.Request)

	CreateRoleGroup(w http.ResponseWriter, r *http.Request)
	UpdateRoleGroup(w http.ResponseWriter, r *http.Request)
	FetchRoleGroupById(w http.ResponseWriter, r *http.Request)
	FetchRoleGroups(w http.ResponseWriter, r *http.Request)
}

type HelmUserRestHandlerImpl struct {
	userService      user.HelmUserService
	validator        *validator.Validate
	logger           *zap.SugaredLogger
	roleGroupService user.HelmUserRoleGroupService
}

func NewHelmUserRestHandlerImpl(userService user.HelmUserService, validator *validator.Validate,
	logger *zap.SugaredLogger, roleGroupService user.HelmUserRoleGroupService) *HelmUserRestHandlerImpl {
	userAuthHandler := &HelmUserRestHandlerImpl{
		userService: userService, validator: validator,
		logger: logger, roleGroupService: roleGroupService,
	}
	return userAuthHandler
}

func (handler HelmUserRestHandlerImpl) CreateHelmUser(w http.ResponseWriter, r *http.Request) {
	var userInfo bean2.UserInfo
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userInfo)
	if err != nil {
		handler.logger.Errorw("request err, CreateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	userInfo.UserId = 1
	handler.logger.Infow("request payload, CreateUser", "payload", userInfo)

	res, err := handler.userService.CreateUser(&userInfo)
	if err != nil {
		handler.logger.Errorw("service err, CreateUser", "err", err, "payload", userInfo)
		if _, ok := err.(*util.ApiError); ok {
			common.WriteJsonResp(w, err, "User Creation Failed", http.StatusOK)
		} else {
			handler.logger.Errorw("error on creating new user", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
	return
}

func (handler HelmUserRestHandlerImpl) UpdateHelmUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var userInfo bean2.UserInfo
	err = decoder.Decode(&userInfo)
	if err != nil {
		handler.logger.Errorw("request err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	userInfo.UserId = userId
	handler.logger.Infow("request payload, UpdateUser", "payload", userInfo)

	if userInfo.EmailId == "admin" {
		userInfo.EmailId = "admin@github.com/devtron-labs"
	}
	err = handler.validator.Struct(userInfo)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if userInfo.EmailId == "admin@github.com/devtron-labs" {
		userInfo.EmailId = "admin"
	}
	res, err := handler.userService.UpdateUser(&userInfo)
	if err != nil {
		handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
	return
}

func (handler HelmUserRestHandlerImpl) GetById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, GetById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.userService.GetById(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, GetById", "err", err, "id", id)
		common.WriteJsonResp(w, err, "Failed to get by id", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HelmUserRestHandlerImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	res, err := handler.userService.GetAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAll", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HelmUserRestHandlerImpl) CreateRoleGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request bean2.HelmRoleGroupDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, CreateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	handler.logger.Infow("request payload, CreateRoleGroup", "err", err, "payload", request)

	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, CreateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := handler.roleGroupService.CreateRoleGroup(&request)
	if err != nil {
		handler.logger.Errorw("service err, CreateRoleGroup", "err", err, "payload", request)
		if _, ok := err.(*util.ApiError); ok {
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HelmUserRestHandlerImpl) UpdateRoleGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request bean2.HelmRoleGroupDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	handler.logger.Infow("request payload, UpdateRoleGroup", "err", err, "payload", request)

	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := handler.roleGroupService.UpdateRoleGroup(&request)
	if err != nil {
		handler.logger.Errorw("service err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HelmUserRestHandlerImpl) FetchRoleGroupById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchRoleGroupById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.roleGroupService.FetchRoleGroupsById(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, FetchRoleGroupById", "err", err, "id", id)
		common.WriteJsonResp(w, err, "Failed to get by id", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HelmUserRestHandlerImpl) FetchRoleGroups(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	res, err := handler.roleGroupService.FetchRoleGroups()
	if err != nil {
		handler.logger.Errorw("service err, FetchRoleGroups", "err", err)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
