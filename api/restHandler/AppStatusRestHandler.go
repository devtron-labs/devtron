package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type AppStatusRestHandler interface {
	GetAllDevtronAppStatuses(w http.ResponseWriter, r *http.Request)
}

type AppStatusRestHandlerImpl struct {
	logger               *zap.SugaredLogger
	userAuthService      user.UserService
	argoAppStatusService appStatus.AppStatusService
}

func NewAppStatusRestHandlerImpl(logger *zap.SugaredLogger, argoAppStatusService appStatus.AppStatusService, userAuthService user.UserService) *AppStatusRestHandlerImpl {
	return &AppStatusRestHandlerImpl{
		logger:               logger,
		userAuthService:      userAuthService,
		argoAppStatusService: argoAppStatusService,
	}
}

func (handler *AppStatusRestHandlerImpl) GetAllDevtronAppStatuses(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	handler.logger.Debugw("request by user", "userId", userId)
	if userId == 0 || err != nil {
		return
	}
	userInfo, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(userInfo.EmailId)
	var requests []appStatus.AppStatusRequestResponseDto
	err = decoder.Decode(&requests)
	if err != nil {
		handler.logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	var responses []appStatus.AppStatusRequestResponseDto
	var allDevtronAppRequests, allInstalledAppRequests int

	for _, request := range requests {
		if request.AppId > 0 {
			allDevtronAppRequests += 1
		}
		if request.InstalledAppId > 0 {
			allInstalledAppRequests += 1
		}
	}
	if allDevtronAppRequests == len(requests) {
		responses, err = handler.argoAppStatusService.GetAllDevtronAppStatuses(requests, userEmailId)
	} else if allInstalledAppRequests == len(requests) {
		responses, err = handler.argoAppStatusService.GetAllInstalledAppStatuses(requests, userEmailId)
	} else {
		handler.logger.Errorw("Invalid request payload", "payload", requests)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if err != nil {
		handler.logger.Errorw("error in fetching app statuses for argo configured environments", "err", err)
		apiError := &util.ApiError{
			InternalMessage: "error occurred while fetching app status for argo-configured" + " err : " + err.Error(),
			UserMessage:     "error in fetching app statuses",
		}
		common.WriteJsonResp(w, apiError, nil, http.StatusInternalServerError)
	}

	common.WriteJsonResp(w, nil, responses, http.StatusOK)

}
