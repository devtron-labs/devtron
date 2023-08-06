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
	GetDraftVersionMetadata(w http.ResponseWriter, r *http.Request)
	GetDraftComments(w http.ResponseWriter, r *http.Request)
	GetAppDrafts(w http.ResponseWriter, r *http.Request)
	GetDraftById(w http.ResponseWriter, r *http.Request)
	ApproveDraft(w http.ResponseWriter, r *http.Request)
	DeleteUserComment(w http.ResponseWriter, r *http.Request)
	UpdateDraftState(w http.ResponseWriter, r *http.Request)
	GetDraftsCount(w http.ResponseWriter, r *http.Request)
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
	enforced := impl.enforceForAppAndEnv(request.AppId, request.EnvId, token, casbin.ActionCreate)
	if !enforced {
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
	configDraft, err := impl.configDraftService.GetDraftById(draftId, userId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft", "draftId", draftId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")
	enforced := impl.enforceForAppAndEnv(configDraft.AppId, configDraft.EnvId, token, casbin.ActionCreate)
	if !enforced {
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

func (impl *ConfigDraftRestHandlerImpl) GetDraftVersionMetadata(w http.ResponseWriter, r *http.Request) {

	draftId, _, errorOccurred, _ := impl.enforceDraftRequest(w, r, casbin.ActionGet, true)
	if errorOccurred {
		return
	}

	draftVersionMetadata, err := impl.configDraftService.GetDraftVersionMetadata(draftId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft version metadata", "err", err, "draftId", draftId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, draftVersionMetadata, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) GetDraftComments(w http.ResponseWriter, r *http.Request) {
	draftId, _, errorOccurred, _ := impl.enforceDraftRequest(w, r, casbin.ActionGet, true)
	if errorOccurred {
		return
	}
	draftComments, err := impl.configDraftService.GetDraftComments(draftId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft comments", "err", err, "draftId", draftId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, draftComments, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) GetAppDrafts(w http.ResponseWriter, r *http.Request) {

	// need to send Approver's data, need to send encrypted secret data
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	appId, err := common.ExtractIntQueryParam(w, r, "appId")
	if err != nil {
		return
	}
	envId, err := common.ExtractIntQueryParam(w, r, "envId")
	if err != nil {
		return
	}
	resourceType, err := common.ExtractIntQueryParam(w, r, "resourceType")
	if err != nil {
		return
	}
	token := r.Header.Get("token")
	enforced := impl.enforceForAppAndEnv(appId, envId, token, casbin.ActionGet)
	if !enforced {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	configDrafts, err := impl.configDraftService.GetDrafts(appId, envId, drafts.DraftResourceType(resourceType), userId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft comments", "err", err, "appId", appId, "envId", envId, "resourceType", resourceType)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, configDrafts, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) GetDraftsCount(w http.ResponseWriter, r *http.Request) {

	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	appId, err := common.ExtractIntQueryParam(w, r, "appId")
	if err != nil {
		return
	}
	envIds, err := common.ExtractIntArrayQueryParam(w, r, "envIds")
	if err != nil {
		return
	}

	//RBAC Starts
	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC Ends

	configDrafts, err := impl.configDraftService.GetDraftsCount(appId, envIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft count", "err", err, "appId", appId, "envIds", envIds)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, configDrafts, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) GetDraftById(w http.ResponseWriter, r *http.Request) {
	// if user has admin access then its fine
	_, _, notAnAppAdmin, draftResponse := impl.enforceDraftRequest(w, r, casbin.ActionUpdate, false)
	if draftResponse == nil {
		return
	}
	if draftResponse.Resource == drafts.CSDraftResource && notAnAppAdmin {
		token := r.Header.Get("token")
		if notAnApprover := impl.checkForApproverAccess(w, draftResponse.EnvId, draftResponse.AppId, token, true); notAnApprover {
			// not an admin and config approver, protecting secret data
			encryptedCSData := impl.configDraftService.EncryptCSData(draftResponse.Data)
			draftResponse.Data = encryptedCSData
		}
	}
	common.WriteJsonResp(w, nil, draftResponse, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) checkForApproverAccess(w http.ResponseWriter, envId int, appId int, token string, logError bool) bool {
	var object string
	if envId > 0 {
		object = impl.enforcerUtil.GetTeamEnvRBACNameByAppId(appId, envId)
	} else {
		object = impl.enforcerUtil.GetTeamNoEnvRBACNameByAppId(appId)
	}
	if ok := impl.enforcer.Enforce(token, casbin.ResourceConfig, casbin.ActionApprove, object); !ok {
		if logError {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		}
		return true
	}
	return false
}

func (impl *ConfigDraftRestHandlerImpl) enforceDraftRequest(w http.ResponseWriter, r *http.Request, action string, writeErrorMsg bool) (int, int32, bool, *drafts.ConfigDraftResponse) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return 0, 0, true, nil
	}

	draftId, err := common.ExtractIntPathParam(w, r, "draftId")
	if err != nil {
		return 0, 0, true, nil
	}

	draftResponse, notAllowed := impl.enforceForDraftId(w, r, draftId, userId, action, writeErrorMsg)
	return draftId, userId, notAllowed, draftResponse
}

func (impl *ConfigDraftRestHandlerImpl) enforceForDraftId(w http.ResponseWriter, r *http.Request, draftId int, userId int32, action string, writeErrorMsg bool) (*drafts.ConfigDraftResponse, bool) {
	configDraft, err := impl.configDraftService.GetDraftById(draftId, userId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft", "draftId", draftId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, true
	}

	token := r.Header.Get("token")
	enforced := impl.enforceForAppAndEnv(configDraft.AppId, configDraft.EnvId, token, action)
	if !enforced {
		if writeErrorMsg {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		}
		return configDraft, true
	}
	return configDraft, false
}

func (impl *ConfigDraftRestHandlerImpl) enforceForAppAndEnv(appId int, envId int, token string, action string) bool {
	object := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, action, object); !ok {
		return false
	}
	if envId != -1 {
		object = impl.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, action, object); !ok {
			return false
		}
	}
	return true
}

func (impl *ConfigDraftRestHandlerImpl) ApproveDraft(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	draftId, err := common.ExtractIntPathParam(w, r, "draftId")
	if err != nil {
		return
	}
	draftVersionId, err := common.ExtractIntPathParam(w, r, "draftVersionId")
	if err != nil {
		return
	}
	draftResponse, err := impl.configDraftService.GetDraftById(draftId, 0)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching draft", "err", err, "draftVersionId", draftVersionId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")
	envId := draftResponse.EnvId
	appId := draftResponse.AppId
	if notAnApprover := impl.checkForApproverAccess(w, envId, appId, token, true); notAnApprover {
		return
	}
	err = impl.configDraftService.ApproveDraft(draftId, draftVersionId, userId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) DeleteUserComment(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	draftId, err := common.ExtractIntQueryParam(w, r, "draftId")
	if err != nil {
		return
	}
	draftCommentId, err := common.ExtractIntQueryParam(w, r, "draftCommentId")
	if err != nil {
		return
	}

	_, notAllowed := impl.enforceForDraftId(w, r, draftId, userId, casbin.ActionDelete, true)
	if notAllowed {
		return
	}
	err = impl.configDraftService.DeleteComment(draftId, draftCommentId, userId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl *ConfigDraftRestHandlerImpl) UpdateDraftState(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	draftId, err := common.ExtractIntQueryParam(w, r, "draftId")
	if err != nil {
		return
	}
	draftVersionId, err := common.ExtractIntQueryParam(w, r, "draftVersionId")
	if err != nil {
		return
	}
	state, err := common.ExtractIntQueryParam(w, r, "state")
	if err != nil {
		return
	}
	_, notAllowed := impl.enforceForDraftId(w, r, draftId, userId, casbin.ActionUpdate, true)
	if notAllowed {
		return
	}
	draftVersion, err := impl.configDraftService.UpdateDraftState(draftId, draftVersionId, drafts.DraftState(state), userId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, draftVersion, http.StatusOK)
}




