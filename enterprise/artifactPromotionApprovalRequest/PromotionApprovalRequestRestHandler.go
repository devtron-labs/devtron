package artifactPromotionApprovalRequest

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/artifactPromotionApprovalRequest"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type PromotionApprovalRequestRestHandler interface {
	HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request)
	GetByPromotionRequestId(w http.ResponseWriter, r *http.Request)
}

type PromotionApprovalRequestRestHandlerImpl struct {
	promotionApprovalRequestService artifactPromotionApprovalRequest.ArtifactPromotionApprovalService
	logger                          *zap.SugaredLogger
	userService                     user.UserService
	enforcer                        casbin.Enforcer
	validator                       *validator.Validate
	userCommonService               user.UserCommonService
	enforcerUtil                    rbac.EnforcerUtil
}

func NewArtifactPromotionApprovalServiceImpl(
	promotionApprovalRequestService artifactPromotionApprovalRequest.ArtifactPromotionApprovalService,
	logger *zap.SugaredLogger,
	userService user.UserService,
	validator *validator.Validate,
	userCommonService user.UserCommonService,
	enforcerUtil rbac.EnforcerUtil,
) *PromotionApprovalRequestRestHandlerImpl {
	return &PromotionApprovalRequestRestHandlerImpl{
		promotionApprovalRequestService: promotionApprovalRequestService,
		logger:                          logger,
		userService:                     userService,
		validator:                       validator,
		userCommonService:               userCommonService,
		enforcerUtil:                    enforcerUtil,
	}
}

func (handler PromotionApprovalRequestRestHandlerImpl) HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if err != nil || userId == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isAuthorised, err := handler.userService.IsUserAdminOrManagerForAnyApp(userId, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	var promotionRequest artifactPromotionApprovalRequest.ArtifactPromotionRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&promotionRequest)
	if err != nil {
		handler.logger.Errorw("err in decoding request in promotionRequest", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	_, err = handler.promotionApprovalRequestService.HandleArtifactPromotionRequest(&promotionRequest)
	if err != nil {
		handler.logger.Errorw("error in handling promotion artifact request", "promotionRequest", promotionRequest, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler PromotionApprovalRequestRestHandlerImpl) GetByPromotionRequestId(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if err != nil || userId == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isAuthorised, err := handler.userService.IsUserAdminOrManagerForAnyApp(userId, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	queryParams := r.URL.Query()
	promotionRequestIdStr := queryParams.Get("promotionRequestId")
	promotionRequestId, err := strconv.Atoi(promotionRequestIdStr)
	if err != nil {
		handler.logger.Errorw("error in parsing promotionRequestId from string to int", "promotionRequestId", promotionRequestId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	_, err = handler.promotionApprovalRequestService.GetByPromotionRequestId(promotionRequestId)
	if err != nil {
		handler.logger.Errorw("error in getting data for promotion request id", "promotionRequestId", promotionRequestId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
