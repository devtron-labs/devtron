package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/argoAppStatus"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
)

type ArgoAppStatusRestHandler interface {
	GetAllDevtronAppStatuses(w http.ResponseWriter, r *http.Request)
	GetAllInstalledAppStatuses(w http.ResponseWriter, r *http.Request)
}

type ArgoAppStatusRestHandlerImpl struct {
	logger               *zap.SugaredLogger
	userAuthService      user.UserService
	argoAppStatusService argoAppStatus.ArgoAppStatusService
	enforcer             casbin.Enforcer
	enforcerUtil         rbac.EnforcerUtil
}

func NewArgoAppStatusRestHandlerImpl(logger *zap.SugaredLogger, argoAppStatusService argoAppStatus.ArgoAppStatusService, userAuthService user.UserService, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil) *ArgoAppStatusRestHandlerImpl {
	return &ArgoAppStatusRestHandlerImpl{
		logger:               logger,
		userAuthService:      userAuthService,
		argoAppStatusService: argoAppStatusService,
		enforcer:             enforcer,
		enforcerUtil:         enforcerUtil,
	}
}

func (handler *ArgoAppStatusRestHandlerImpl) GetAllDevtronAppStatuses(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	handler.logger.Debugw("request by user", "userId", userId)
	if userId == 0 || err != nil {
		return
	}
	var requests []argoAppStatus.AppStatusRequestResponseDto
	err = decoder.Decode(&requests)
	if err != nil {
		handler.logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//Filter out Apps with atleast view Access and find statuses for them
	//TODO : Filter out Accessible requests
	responses, err := handler.argoAppStatusService.GetAllDevtronAppStatuses(requests)
	if err != nil {
		handler.logger.Errorw("error in fetching app statuses for argo configured environments", "err", err)
		apiError := &util.ApiError{
			InternalMessage: "error occurred while fetching app status for argo-configured" + " err : " + err.Error(),
			UserMessage:     "error in fetching app statuses",
		}
		common.WriteJsonResp(w, apiError, nil, http.StatusInternalServerError)
	}

	//TODO : Filter out Accessible responses
	//accessibleDevtronAppStatuses := make([]argoAppStatus.AppStatusRequestResponseDto, 0)
	common.WriteJsonResp(w, nil, responses, http.StatusOK)

}

func (handler *ArgoAppStatusRestHandlerImpl) GetAllInstalledAppStatuses(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	handler.logger.Debugw("request by user", "userId", userId)
	if userId == 0 || err != nil {
		return
	}
	var requests []argoAppStatus.AppStatusRequestResponseDto
	err = decoder.Decode(&requests)
	if err != nil {
		handler.logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//Filter out Apps with atleast view Access and find statuses for them
	//TODO : Filter out Accessible requests
	responses, err := handler.argoAppStatusService.GetAllInstalledAppStatuses(requests)
	if err != nil {
		handler.logger.Errorw("error in fetching app statuses for argo configured environments", "err", err)
		apiError := &util.ApiError{
			InternalMessage: "error occurred while fetching app status for argo-configured",
			UserMessage:     "error in fetching app statuses",
		}
		common.WriteJsonResp(w, apiError, nil, http.StatusInternalServerError)
	}

	//TODO : Filter out Accessible responses
	//accessibleDevtronAppStatuses := make([]argoAppStatus.AppStatusRequestResponseDto, 0)
	common.WriteJsonResp(w, nil, responses, http.StatusOK)
}
