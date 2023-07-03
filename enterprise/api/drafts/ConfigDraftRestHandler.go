package drafts

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type ConfigDraftRestHandler interface {
	CreateDraft(w http.ResponseWriter, r *http.Request)
	AddDraftVersion(w http.ResponseWriter, r *http.Request)
}

type ConfigDraftRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	userService        user.UserService
	enforcer           casbin.Enforcer
	enforcerUtil       rbac.EnforcerUtil
	validator          *validator.Validate
	configDraftService drafts.ConfigDraftService
}

func NewConfigDraftRestHandlerImpl(logger *zap.SugaredLogger, userService user.UserService, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, validator *validator.Validate, configDraftService drafts.ConfigDraftService) *ConfigDraftRestHandlerImpl {
	return &ConfigDraftRestHandlerImpl{
		logger:             logger,
		enforcer:           enforcer,
		enforcerUtil:       enforcerUtil,
		validator:          validator,
		configDraftService: configDraftService,
		userService:        userService,
	}
}

func (impl *ConfigDraftRestHandlerImpl) CreateDraft(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	var request drafts.ConfigDraftRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in CreateDraft", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in CreateDraft", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(request.AppId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = impl.enforcerUtil.GetEnvRBACNameByAppId(request.AppId, request.EnvId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	request.UserId = userId
	configDraftResponse, err := impl.configDraftService.CreateDraft(request)
	if err != nil {
		impl.logger.Errorw("error occurred while creating draft", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, configDraftResponse, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) AddDraftVersion(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	var request drafts.ConfigDraftVersionRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in AddDraftVersion", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in AddDraftVersion", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	draftId := request.DraftId
	configDraft, err := impl.configDraftService.GetDraftById(draftId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft", "draftId", draftId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(configDraft.AppId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = impl.enforcerUtil.GetEnvRBACNameByAppId(configDraft.AppId, configDraft.EnvId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	request.UserId = userId
	draftVersionId, err := impl.configDraftService.AddDraftVersion(request)
	if err != nil {
		impl.logger.Errorw("error occurred while adding draft version", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	response := make(map[string]string, 1)
	response["draftVersionId"] = strconv.Itoa(draftVersionId)
	common.WriteJsonResp(w, err, response, http.StatusOK)
}
