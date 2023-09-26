package resourceFilter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type ResourceFilterRestHandler interface {
	ListFilters(w http.ResponseWriter, r *http.Request)
	GetFilterById(w http.ResponseWriter, r *http.Request)
	UpdateFilter(w http.ResponseWriter, r *http.Request)
	CreateFilter(w http.ResponseWriter, r *http.Request)
	DeleteFilter(w http.ResponseWriter, r *http.Request)
}

type ResourceFilterRestHandlerImpl struct {
	logger                *zap.SugaredLogger
	userAuthService       user.UserService
	enforcerUtil          rbac.EnforcerUtil
	enforcer              casbin.Enforcer
	resourceFilterService resourceFilter.ResourceFilterService
}

func NewResourceFilterRestHandlerImpl(logger *zap.SugaredLogger,
	userAuthService user.UserService,
	enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer,
	resourceFilterService resourceFilter.ResourceFilterService) *ResourceFilterRestHandlerImpl {
	return &ResourceFilterRestHandlerImpl{
		logger:                logger,
		userAuthService:       userAuthService,
		enforcerUtil:          enforcerUtil,
		enforcer:              enforcer,
		resourceFilterService: resourceFilterService,
	}
}

func (handler *ResourceFilterRestHandlerImpl) ListFilters(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	authorised, err := handler.applyAuth(userId)
	if err != nil || !authorised {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	res, err := handler.resourceFilterService.ListFilters()
	if err != nil {
		handler.logger.Errorw("error in getting active resource filters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *ResourceFilterRestHandlerImpl) GetFilterById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	authorised, err := handler.applyAuth(userId)
	if err != nil || !authorised {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	filterId, err := strconv.Atoi(vars["Id"])
	if err != nil {
		common.WriteJsonResp(w, errors.New(fmt.Sprintf("invalid param Id '%s'", vars["Id"])), nil, http.StatusBadRequest)
		return
	}
	res, err := handler.resourceFilterService.GetFilterById(filterId)
	if err != nil {
		handler.logger.Errorw("error in getting  resource filter", "err", err, "filterId", filterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *ResourceFilterRestHandlerImpl) CreateFilter(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	authorised, err := handler.applyAuth(userId)
	if err != nil || !authorised {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	req := &resourceFilter.FilterRequestResponseBean{}
	err = decoder.Decode(req)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "request", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.resourceFilterService.CreateFilter(userId, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		handler.logger.Errorw("error in creating resource filters", "err", err)
		if err.Error() == resourceFilter.AppAndEnvSelectorRequiredMessage {
			statusCode = http.StatusBadRequest
		} else if err.Error() == resourceFilter.InvalidExpressions {
			err = nil
			statusCode = http.StatusPreconditionFailed
		}
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *ResourceFilterRestHandlerImpl) UpdateFilter(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	authorised, err := handler.applyAuth(userId)
	if err != nil || !authorised {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	filterId, err := strconv.Atoi(vars["Id"])
	if err != nil {
		common.WriteJsonResp(w, errors.New(fmt.Sprintf("invalid param Id '%s'", vars["Id"])), nil, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	req := &resourceFilter.FilterRequestResponseBean{}
	err = decoder.Decode(req)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "request", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	req.Id = filterId
	err = handler.resourceFilterService.UpdateFilter(userId, req)
	if err != nil {
		handler.logger.Errorw("error in updating active resource filters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *ResourceFilterRestHandlerImpl) DeleteFilter(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	authorised, err := handler.applyAuth(userId)
	if err != nil || !authorised {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	filterId, err := strconv.Atoi(vars["Id"])
	if err != nil {
		handler.logger.Errorw("error in getting active resource filters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}

	err = handler.resourceFilterService.DeleteFilter(userId, filterId)
	if err != nil {
		handler.logger.Errorw("error in deleting resource filters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *ResourceFilterRestHandlerImpl) applyAuth(userId int32) (authorised bool, err error) {

	isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))

	if err != nil {
		handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		return false, err
	}

	if !isSuperAdmin {
		return false, nil
	}

	return true, nil
}
