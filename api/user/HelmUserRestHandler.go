package user

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type HelmUserRestHandler interface {
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
	var userInfo bean.UserInfo
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userInfo)
	if err != nil {
		handler.logger.Errorw("request err, CreateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	userInfo.UserId = 1
	handler.logger.Infow("request payload, CreateUser", "payload", userInfo)

	res, err := handler.userService.CreateHelmUser(&userInfo)
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
	common.WriteJsonResp(w, nil, "res", http.StatusOK)
	return
}
