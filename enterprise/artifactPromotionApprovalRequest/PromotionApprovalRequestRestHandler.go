package artifactPromotionApprovalRequest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/artifactPromotionApprovalRequest"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
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
	promotionApprovalRequestService            artifactPromotionApprovalRequest.ArtifactPromotionApprovalService
	logger                                     *zap.SugaredLogger
	userService                                user.UserService
	enforcer                                   casbin.Enforcer
	validator                                  *validator.Validate
	userCommonService                          user.UserCommonService
	enforcerUtil                               rbac.EnforcerUtil
	environmentRepository                      repository.EnvironmentRepository
	artifactPromotionApprovalRequestRepository artifactPromotionApprovalRequest.ArtifactPromotionApprovalRequestRepository
}

func NewArtifactPromotionApprovalServiceImpl(
	promotionApprovalRequestService artifactPromotionApprovalRequest.ArtifactPromotionApprovalService,
	logger *zap.SugaredLogger,
	userService user.UserService,
	validator *validator.Validate,
	userCommonService user.UserCommonService,
	enforcerUtil rbac.EnforcerUtil,
	environmentRepository repository.EnvironmentRepository,
	artifactPromotionApprovalRequestRepository artifactPromotionApprovalRequest.ArtifactPromotionApprovalRequestRepository,
) *PromotionApprovalRequestRestHandlerImpl {
	return &PromotionApprovalRequestRestHandlerImpl{
		promotionApprovalRequestService: promotionApprovalRequestService,
		logger:                          logger,
		userService:                     userService,
		validator:                       validator,
		userCommonService:               userCommonService,
		enforcerUtil:                    enforcerUtil,
		environmentRepository:           environmentRepository,
		artifactPromotionApprovalRequestRepository: artifactPromotionApprovalRequestRepository,
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

	envAuthorized := make(map[string]bool)

	switch promotionRequest.Action {
	case artifactPromotionApprovalRequest.ACTION_PROMOTE:

		appName := promotionRequest.AppName
		appRbacObject := handler.enforcerUtil.GetAppRBACName(appName)
		ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appRbacObject)
		if !ok {
			common.WriteJsonResp(w, err, nil, http.StatusForbidden)
			return
		}

		environmentNames := promotionRequest.EnvironmentNames
		envRbacObjectMap := handler.enforcerUtil.GetEnvRBACByAppNameAndEnvNames(appName, environmentNames)
		envObjectArr := make([]string, 0)
		for _, obj := range envObjectArr {
			envObjectArr = append(envObjectArr, obj)
		}
		results := handler.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObjectArr)
		for _, env := range environmentNames {
			rbacObject := envRbacObjectMap[env]
			isAuthorised = results[rbacObject]
			envAuthorized[env] = isAuthorised
		}

	case artifactPromotionApprovalRequest.ACTION_APPROVE:
		appName := promotionRequest.AppName
		environmentNames := promotionRequest.EnvironmentNames
		teamEnvRbacObjectMap := handler.enforcerUtil.GetTeamEnvRbacObjByAppAndEnvNames(appName, environmentNames)
		teamEnvObjectArr := make([]string, 0)
		for _, obj := range teamEnvObjectArr {
			teamEnvObjectArr = append(teamEnvObjectArr, obj)
		}
		results := handler.enforcer.EnforceInBatch(token, casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, teamEnvObjectArr)
		for _, env := range promotionRequest.EnvironmentNames {
			rbacObject := teamEnvRbacObjectMap[env]
			isAuthorised = results[rbacObject]
			envAuthorized[env] = isAuthorised
		}

	}

	_, err = handler.promotionApprovalRequestService.HandleArtifactPromotionRequest(&promotionRequest, envAuthorized)
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

	artifactPromotionDao, err := handler.artifactPromotionApprovalRequestRepository.FindById(promotionRequestId)
	if err == pg.ErrNoRows {
		handler.logger.Errorw("promotion request for given id does not exist", "promotionRequestId", promotionRequestId, "err", err)
		common.WriteJsonResp(w, errors.New("promotion request for given id does not exist"), nil, http.StatusNotFound)
		return
	}
	if err != nil {
		handler.logger.Errorw("error in fetching artifact promotion request by id", "artifactPromotionRequestId", promotionRequestId, "err", err)
		return
	}

	// rbac block starts from here
	object := handler.enforcerUtil.GetAppAndEnvObjectByPipelineIds([]int{artifactPromotionDao.DestinationPipelineId})
	if len(object) == 0 {
		common.WriteJsonResp(w, errors.New("unable to perform authorization check for given resource"), nil, http.StatusInternalServerError)
	}

	rbacObjects := object[artifactPromotionDao.DestinationPipelineId]
	appRbacObj := rbacObjects[0]
	envRbacObj := rbacObjects[1]

	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, appRbacObj); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, envRbacObj); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	_, err = handler.promotionApprovalRequestService.GetByPromotionRequestId(artifactPromotionDao)
	if err != nil {
		handler.logger.Errorw("error in getting data for promotion request id", "promotionRequestId", promotionRequestId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
