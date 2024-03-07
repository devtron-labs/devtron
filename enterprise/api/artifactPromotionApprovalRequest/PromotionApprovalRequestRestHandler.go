package artifactPromotionApprovalRequest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	artifactPromotion2 "github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	repository2 "github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/strings/slices"
	"net/http"
)

type RestHandler interface {
	HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request)
	FetchAwaitingApprovalEnvListForArtifact(w http.ResponseWriter, r *http.Request)
	FetchEnvironmentsList(w http.ResponseWriter, r *http.Request)
}

type MaterialRestHandler interface {
	GetArtifactsForPromotion(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	promotionApprovalRequestService            artifactPromotion2.ApprovalRequestService
	logger                                     *zap.SugaredLogger
	userService                                user.UserService
	enforcer                                   casbin.Enforcer
	validator                                  *validator.Validate
	userCommonService                          user.UserCommonService
	enforcerUtil                               rbac.EnforcerUtil
	environmentRepository                      repository.EnvironmentRepository
	artifactPromotionApprovalRequestRepository repository2.RequestRepository
	appArtifactManager                         pipeline.AppArtifactManager
	CiArtifactRepository                       repository3.CiArtifactRepository
}

func NewRestHandlerImpl(
	promotionApprovalRequestService artifactPromotion2.ApprovalRequestService,
	logger *zap.SugaredLogger,
	userService user.UserService,
	validator *validator.Validate,
	userCommonService user.UserCommonService,
	enforcerUtil rbac.EnforcerUtil,
	environmentRepository repository.EnvironmentRepository,
	artifactPromotionApprovalRequestRepository repository2.RequestRepository,
	appArtifactManager pipeline.AppArtifactManager,
	enforcer casbin.Enforcer,
) *RestHandlerImpl {
	return &RestHandlerImpl{
		promotionApprovalRequestService: promotionApprovalRequestService,
		logger:                          logger,
		userService:                     userService,
		validator:                       validator,
		userCommonService:               userCommonService,
		enforcerUtil:                    enforcerUtil,
		environmentRepository:           environmentRepository,
		artifactPromotionApprovalRequestRepository: artifactPromotionApprovalRequestRepository,
		appArtifactManager:                         appArtifactManager,
		enforcer:                                   enforcer,
	}
}

func (handler *RestHandlerImpl) HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request) {
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
	var promotionRequest bean.ArtifactPromotionRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&promotionRequest)
	if err != nil {
		handler.logger.Errorw("err in decoding request in promotionRequest", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	promotionRequest.UserId = userId

	authorizedEnvironments := make(map[string]bool)

	switch promotionRequest.Action {
	case constants.ACTION_PROMOTE:

		authorizedEnvironments = handler.promoteActionRbac(token, promotionRequest.AppName, promotionRequest.EnvironmentNames)

	case constants.ACTION_APPROVE:

		authorizedEnvironments = handler.approveActionRbac(token, promotionRequest.AppName, promotionRequest.EnvironmentNames)

	case constants.ACTION_CANCEL:
		// get this info from service layer
		artifactPromotionDao, err := handler.promotionApprovalRequestService.GetPromotionRequestById(promotionRequest.PromotionRequestId)
		if err != nil {
			handler.logger.Errorw("error in fetching promotion request by id", "promotionRequestId", promotionRequest.PromotionRequestId, "err", err)
			common.WriteJsonResp(w, errors.New("error in fetching promotion request "), nil, http.StatusInternalServerError)
			return
		}
		appRbacObj, envRbacObj := handler.getAppAndEnvObjectByCdPipelineId(artifactPromotionDao.DestinationPipelineId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appRbacObj); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envRbacObj); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}

	resp, err := handler.promotionApprovalRequestService.HandleArtifactPromotionRequest(&promotionRequest, authorizedEnvironments)
	if err != nil {
		handler.logger.Errorw("error in handling promotion artifact request", "promotionRequest", promotionRequest, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *RestHandlerImpl) getAppAndEnvObjectByCdPipelineId(cdPipelineId int) (string, string) {
	object := handler.enforcerUtil.GetAppAndEnvObjectByPipelineIds([]int{cdPipelineId})
	rbacObjects := object[cdPipelineId]
	return rbacObjects[0], rbacObjects[1]
}
func (handler *RestHandlerImpl) FetchAwaitingApprovalEnvListForArtifact(w http.ResponseWriter, r *http.Request) {

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

	artifactId := 0
	artifactId, err = common.ExtractIntQueryParam(w, r, "artifactId", &artifactId)
	if err != nil {
		handler.logger.Errorw("error in extracting artifactId from query param", "artifactIdStr", queryParams.Get("artifactId"), "err", err)
		common.WriteJsonResp(w, errors.New("artifactId should be an integer value"), nil, http.StatusBadRequest)
		return
	}

	environmentName := queryParams.Get("environmentName")

	environmentApprovalMetadata, err := handler.promotionApprovalRequestService.FetchApprovalAllowedEnvList(artifactId, environmentName, userId, token, handler.CheckImagePromoterAuth)
	if err != nil {
		handler.logger.Errorw("error in fetching environments with pending approval for artifact", "artifactId", artifactId, "err", err)
		return
	}

	common.WriteJsonResp(w, nil, environmentApprovalMetadata, http.StatusOK)
	return

}

func (handler *RestHandlerImpl) GetArtifactsForPromotion(w http.ResponseWriter, r *http.Request) {

	artifactPromotionMaterialRequest, err := handler.parsePromotionMaterialRequest(w, r)
	if err != nil {
		return
	}

	validRequest := handler.validatePromotionMaterialRequest(w, artifactPromotionMaterialRequest)
	if !validRequest {
		return
	}

	if artifactPromotionMaterialRequest.Resource == string(constants.SOURCE_TYPE_CI) || artifactPromotionMaterialRequest.Resource == string(constants.SOURCE_TYPE_CD) {
		// check if user has trigger access for any one env for this app
		if hasTriggerAccess := handler.checkTriggerAccessForAnyEnv(artifactPromotionMaterialRequest.Token,
			artifactPromotionMaterialRequest.AppId); !hasTriggerAccess {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	} else if artifactPromotionMaterialRequest.Resource == string(constants.PROMOTION_APPROVAL_PENDING_NODE) && !artifactPromotionMaterialRequest.PendingForCurrentUser {
		// check if either user has trigger access or artifact promoter access for this env
		appRbacObj := handler.enforcerUtil.GetAppRBACNameByAppId(artifactPromotionMaterialRequest.AppId)
		env, err := handler.environmentRepository.FindByName(artifactPromotionMaterialRequest.ResourceName)
		if err != nil {
			handler.logger.Errorw("env not found for given envName", "envName", env.Name, "err", err)
			common.WriteJsonResp(w, err, "invalid environment name in request", http.StatusBadRequest)
			return
		}
		envObjectMap, _ := handler.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId([]int{env.Id}, artifactPromotionMaterialRequest.AppId)
		if ok := handler.enforcer.Enforce(artifactPromotionMaterialRequest.Token, casbin.ResourceEnvironment, casbin.ActionGet, envObjectMap[env.Id]); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}

		triggerAccess := handler.enforcer.Enforce(artifactPromotionMaterialRequest.Token, casbin.ResourceApplications, casbin.ActionTrigger, appRbacObj) &&
			handler.enforcer.Enforce(artifactPromotionMaterialRequest.Token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObjectMap[env.Id])

		teamRbac := handler.enforcerUtil.GetTeamEnvRBACNameByAppId(artifactPromotionMaterialRequest.AppId, env.Id)

		approverAccess := handler.enforcer.Enforce(artifactPromotionMaterialRequest.Token, casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, teamRbac)

		if !triggerAccess && !approverAccess {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}

	artifactPromotionMaterialResponse, err := handler.appArtifactManager.FetchMaterialForArtifactPromotion(*artifactPromotionMaterialRequest, handler.CheckImagePromoterAuth)
	if err != nil {
		handler.logger.Errorw("error in fetching artifacts for given promotion request parameters", "err", err)
		common.WriteJsonResp(w, errors.New("error in fetching artifacts response for given request parameters"), nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, artifactPromotionMaterialResponse, http.StatusOK)
}

func (handler *RestHandlerImpl) parsePromotionMaterialRequest(w http.ResponseWriter, r *http.Request) (*bean2.ArtifactPromotionMaterialRequest, error) {

	userId, err := handler.userService.GetLoggedInUser(r)
	if err != nil || userId == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return nil, err
	}

	token := r.Header.Get("token")

	queryParams := r.URL.Query()
	resource := queryParams.Get("resource")
	resourceName := queryParams.Get("resourceName")

	pendingForCurrentUser, err := common.ExtractBooleanQueryParam(w, r, "pendingForCurrentUser", false)
	if err != nil {
		handler.logger.Errorw("error in parsing pendingForCurrentUser from string to bool", "pendingForCurrentUser", queryParams.Get("pendingForCurrentUser"))
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, err
	}

	workflowId := 0
	workflowId, err = common.ExtractIntQueryParam(w, r, "workflowId", &workflowId)
	if err != nil {
		handler.logger.Errorw("error in parsing workflowId from string to int", "workflowId", queryParams.Get("workflowId"))
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, err
	}

	appId := 0
	appId, err = common.ExtractIntQueryParam(w, r, "appId", &appId)
	if err != nil {
		handler.logger.Errorw("error in parsing appId from string to int", "workflowId", queryParams.Get("appId"))
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, err
	}

	offsetDefault := 0
	offset, err := common.ExtractIntQueryParam(w, r, "offset", &offsetDefault)
	if err != nil {
		handler.logger.Errorw("error in parsing offset from string to int", "workflowId", queryParams.Get("offset"))
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, err
	}

	limitDefault := 20
	limit, err := common.ExtractIntQueryParam(w, r, "size", &limitDefault)
	if err != nil {
		handler.logger.Errorw("error in parsing limit from string to int", "workflowId", queryParams.Get("size"))
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, err
	}

	searchQueryParam := r.URL.Query().Get("search") // image search string

	artifactPromotionMaterialRequest := &bean2.ArtifactPromotionMaterialRequest{
		Resource:              resource,
		ResourceName:          resourceName,
		AppId:                 appId,
		WorkflowId:            workflowId,
		PendingForCurrentUser: pendingForCurrentUser,
		Limit:                 limit,
		Offset:                offset,
		ImageSearchRegex:      searchQueryParam,
		UserId:                userId,
		Token:                 token,
	}

	return artifactPromotionMaterialRequest, nil
}

func (handler *RestHandlerImpl) validatePromotionMaterialRequest(w http.ResponseWriter, request *bean2.ArtifactPromotionMaterialRequest) bool {

	if len(request.Resource) == 0 {

		handler.logger.Errorw("resource is a mandatory field")
		common.WriteJsonResp(w, errors.New("resource is a mandatory field"), nil, http.StatusBadRequest)
		return false

	} else if len(request.Resource) > 0 {

		if slices.Contains([]string{string(constants.SOURCE_TYPE_CI), string(constants.SOURCE_TYPE_CD), string(constants.SOURCE_TYPE_WEBHOOK)}, request.Resource) {
			if len(request.ResourceName) == 0 || request.AppId == 0 {
				common.WriteJsonResp(w, errors.New(fmt.Sprintf("resourceName/appId is required field for resource = %s ", request.Resource)), nil, http.StatusBadRequest)
				return false
			}

		} else if request.Resource == string(constants.PROMOTION_APPROVAL_PENDING_NODE) {
			if !request.PendingForCurrentUser {
				if len(request.Resource) == 0 || request.AppId == 0 {
					common.WriteJsonResp(w, errors.New(fmt.Sprintf("resourceName/appId is required field for resource = %s if pendingForCurrentUser is false", request.Resource)), nil, http.StatusBadRequest)
					return false
				}
			} else {
				if request.WorkflowId == 0 {
					common.WriteJsonResp(w, errors.New("workflowId is required field if pendingForCurrentUser is true"), nil, http.StatusBadRequest)
					return false
				}
			}
		} else {
			common.WriteJsonResp(w, errors.New(fmt.Sprintf("invalid resource name - %s ", request.Resource)), nil, http.StatusBadRequest)
			return false
		}
	}
	return true
}

func (handler *RestHandlerImpl) checkTriggerAccessForAnyEnv(token string, appId int) bool {

	appObj := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appObj); !ok {
		return false
	}

	envObjects := handler.enforcerUtil.GetEnvRBACArrayByAppId(appId)
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObjects)
	for _, isAuthorised := range results {
		if isAuthorised {
			return true
		}
	}
	return false
}

func (handler *RestHandlerImpl) CheckImagePromoterAuth(token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, object); !ok {
		return false
	}
	return true
}

func (handler *RestHandlerImpl) FetchEnvironmentsList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if err != nil || userId == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")

	queryParams := r.URL.Query()

	workflowId := 0
	workflowId, err = common.ExtractIntQueryParam(w, r, "workflowId", &workflowId)
	if err != nil {
		handler.logger.Errorw("error in extracting workflowId from query param", "workflowIdStr", queryParams.Get("workflowId"), "err", err)
		common.WriteJsonResp(w, errors.New("workflowId should be an integer value"), nil, http.StatusBadRequest)
		return
	}

	artifactId := 0
	artifactId, err = common.ExtractIntQueryParam(w, r, "artifactId", &artifactId)
	if err != nil {
		handler.logger.Errorw("error in extracting artifactId from query param", "artifactIdStr", queryParams.Get("artifactId"), "err", err)
		common.WriteJsonResp(w, errors.New("artifactId should be an integer value"), nil, http.StatusBadRequest)
		return
	}

	resp, err := handler.promotionApprovalRequestService.FetchWorkflowPromoteNodeList(token, workflowId, artifactId, handler.promoteActionRbac)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}

func (handler *RestHandlerImpl) promoteActionRbac(token, appName string, envNames []string) map[string]bool {
	appRbacObject := handler.enforcerUtil.GetAppRBACName(appName)
	ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appRbacObject)
	if !ok {
		return nil
	}

	envRbacObjectMap := handler.enforcerUtil.GetEnvRBACByAppNameAndEnvNames(appName, envNames)
	envObjectArr := make([]string, 0)
	for _, obj := range envRbacObjectMap {
		envObjectArr = append(envObjectArr, obj)
	}
	authorizedEnvironments := make(map[string]bool)
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObjectArr)
	for _, env := range envNames {
		rbacObject := envRbacObjectMap[env]
		authorizedEnvironments[env] = results[rbacObject]
	}
	return authorizedEnvironments
}

func (handler *RestHandlerImpl) approveActionRbac(token, appName string, environmentNames []string) map[string]bool {
	authorizedEnvironments := make(map[string]bool)
	teamEnvRbacObjectMap := handler.enforcerUtil.GetTeamEnvRbacObjByAppAndEnvNames(appName, environmentNames)
	teamEnvObjectArr := make([]string, 0)
	for _, obj := range teamEnvRbacObjectMap {
		teamEnvObjectArr = append(teamEnvObjectArr, obj)
	}
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, teamEnvObjectArr)
	for _, env := range environmentNames {
		rbacObject := teamEnvRbacObjectMap[env]
		isAuthorised := results[rbacObject]
		authorizedEnvironments[env] = isAuthorised
	}
	return authorizedEnvironments
}
