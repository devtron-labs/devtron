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
}

type HelmUserRestHandlerImpl struct {
	userService user.HelmUserService
	validator   *validator.Validate
	logger      *zap.SugaredLogger
}

func NewHelmUserRestHandlerImpl(userService user.HelmUserService, validator *validator.Validate,
	logger *zap.SugaredLogger) *HelmUserRestHandlerImpl {
	userAuthHandler := &HelmUserRestHandlerImpl{
		userService: userService, validator: validator,
		logger: logger,
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
