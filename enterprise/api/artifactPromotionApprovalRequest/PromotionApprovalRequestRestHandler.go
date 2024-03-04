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
	repository2 "github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/strings/slices"
	"net/http"
	"strconv"
)

type RestHandler interface {
	HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request)
	GetByPromotionRequestId(w http.ResponseWriter, r *http.Request)
	FetchAwaitingApprovalEnvListForArtifact(w http.ResponseWriter, r *http.Request)
	FetchEnvironmentsList(w http.ResponseWriter, r *http.Request)
}

type MaterialRestHandler interface {
	GetArtifactsForPromotion(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	promotionApprovalRequestService            artifactPromotion2.ArtifactPromotionApprovalService
	logger                                     *zap.SugaredLogger
	userService                                user.UserService
	enforcer                                   casbin.Enforcer
	validator                                  *validator.Validate
	userCommonService                          user.UserCommonService
	enforcerUtil                               rbac.EnforcerUtil
	environmentRepository                      repository.EnvironmentRepository
	artifactPromotionApprovalRequestRepository repository2.ArtifactPromotionApprovalRequestRepository
	appArtifactManager                         pipeline.AppArtifactManager
	CiArtifactRepository                       repository3.CiArtifactRepository
}

func NewRestHandlerImpl(
	promotionApprovalRequestService artifactPromotion2.ArtifactPromotionApprovalService,
	logger *zap.SugaredLogger,
	userService user.UserService,
	validator *validator.Validate,
	userCommonService user.UserCommonService,
	enforcerUtil rbac.EnforcerUtil,
	environmentRepository repository.EnvironmentRepository,
	artifactPromotionApprovalRequestRepository repository2.ArtifactPromotionApprovalRequestRepository,
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

func (handler RestHandlerImpl) HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request) {
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
	promotionRequest.UserId = 1

	authorizedEnvironments := make(map[string]bool)

	switch promotionRequest.Action {
	case bean.ACTION_PROMOTE:

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
			authorizedEnvironments[env] = true
		}

	case bean.ACTION_APPROVE:
		appName := promotionRequest.AppName
		environmentNames := promotionRequest.EnvironmentNames
		teamEnvRbacObjectMap := handler.enforcerUtil.GetTeamEnvRbacObjByAppAndEnvNames(appName, environmentNames)
		teamEnvObjectArr := make([]string, 0)
		for _, obj := range teamEnvObjectArr {
			teamEnvObjectArr = append(teamEnvObjectArr, obj)
		}
		results := handler.enforcer.EnforceInBatch(token, casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, teamEnvObjectArr)
		for _, env := range environmentNames {
			rbacObject := teamEnvRbacObjectMap[env]
			isAuthorised = results[rbacObject]
			authorizedEnvironments[env] = true
		}

	case bean.ACTION_CANCEL:
		artifactPromotionDao, err := handler.artifactPromotionApprovalRequestRepository.FindById(promotionRequest.PromotionRequestId)
		if err == pg.ErrNoRows {
			handler.logger.Errorw("promotion request for given id does not exist", "promotionRequestId", promotionRequest.PromotionRequestId, "err", err)
			common.WriteJsonResp(w, errors.New("promotion request for given id does not exist"), nil, http.StatusNotFound)
			return
		}
		if err != nil {
			handler.logger.Errorw("error in fetching artifact promotion request by id", "artifactPromotionRequestId", promotionRequest.PromotionRequestId, "err", err)
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

func (handler RestHandlerImpl) GetByPromotionRequestId(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	promotionRequestId, err := strconv.Atoi(vars["promotionRequestId"])
	if err != nil {
		handler.logger.Errorw("error in parsing promotionRequestId from string to int", "promotionRequestId", vars["promotionRequestId"])
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
	appRbacObj, envRbacObj := handler.getAppAndEnvObjectByCdPipelineId(artifactPromotionDao.DestinationPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, appRbacObj); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, envRbacObj); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	resp, err := handler.promotionApprovalRequestService.GetByPromotionRequestId(artifactPromotionDao)
	if err != nil {
		handler.logger.Errorw("error in getting data for promotion request id", "promotionRequestId", promotionRequestId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler RestHandlerImpl) getAppAndEnvObjectByCdPipelineId(cdPipelineId int) (string, string) {
	object := handler.enforcerUtil.GetAppAndEnvObjectByPipelineIds([]int{cdPipelineId})
	rbacObjects := object[cdPipelineId]
	return rbacObjects[0], rbacObjects[1]
}
func (handler RestHandlerImpl) FetchAwaitingApprovalEnvListForArtifact(w http.ResponseWriter, r *http.Request) {

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
	artifactId, err := strconv.Atoi(queryParams.Get("artifactId"))
	if err != nil {
		handler.logger.Errorw("error in parsing artifactId from string to int", "artifactId", queryParams.Get("artifactId"))
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	environmentApprovalMetadata, err := handler.promotionApprovalRequestService.FetchApprovalAllowedEnvList(artifactId, userId, token, handler.CheckImagePromoterAuth)
	if err != nil {
		handler.logger.Errorw("error in fetching environments with pending approval for artifact", "artifactId", artifactId, "err", err)
		return
	}

	common.WriteJsonResp(w, nil, environmentApprovalMetadata, http.StatusOK)
	return

}

func (handler RestHandlerImpl) GetArtifactsForPromotion(w http.ResponseWriter, r *http.Request) {

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

	resource := queryParams.Get("resource")
	resourceName := queryParams.Get("resourceName")

	pendingForCurrentUserQueryParam := queryParams.Get("pendingForCurrentUser")
	var pendingForCurrentUser bool
	if len(pendingForCurrentUserQueryParam) > 0 {
		pendingForCurrentUser, err = strconv.ParseBool(pendingForCurrentUserQueryParam)
		if err != nil {
			handler.logger.Errorw("error in parsing pendingForCurrentUser from string to bool", "pendingForCurrentUser", queryParams.Get("pendingForCurrentUser"))
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	workflowIdQueryParam := queryParams.Get("workflowId")
	var workflowId int
	if len(workflowIdQueryParam) > 0 {
		workflowId, err = strconv.Atoi(workflowIdQueryParam)
		if err != nil {
			handler.logger.Errorw("error in parsing workflowId from string to int", "workflowId", queryParams.Get("workflowId"))
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	appIDQueryParam := queryParams.Get("appId")
	var appId int
	if len(appIDQueryParam) > 0 {
		appId, err = strconv.Atoi(appIDQueryParam)
		if err != nil {
			handler.logger.Errorw("error in parsing appId from string to int", "appId", queryParams.Get("appId"))
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	if len(resource) == 0 && !pendingForCurrentUser {
		handler.logger.Errorw("resource is a mandatory field")
		common.WriteJsonResp(w, errors.New("resource is a mandatory field"), nil, http.StatusBadRequest)
		return
	} else if len(resource) > 0 && !pendingForCurrentUser {
		if !slices.Contains([]string{string(bean.SOURCE_TYPE_CI), string(bean.SOURCE_TYPE_CD), string(bean.SOURCE_TYPE_WEBHOOK), string(bean.PROMOTION_APPROVAL_PENDING_NODE)}, resource) {
			common.WriteJsonResp(w, errors.New(fmt.Sprintf("invalid resource name - %s ", resource)), nil, http.StatusBadRequest)
			return
		}
		if len(resourceName) == 0 || appId == 0 {
			common.WriteJsonResp(w, errors.New(fmt.Sprintf("resourceName/appId is required field for resource = %s ", resource)), nil, http.StatusBadRequest)
			return
		}
	} else if pendingForCurrentUser {
		if workflowId == 0 {
			common.WriteJsonResp(w, errors.New("workflowId is required field if pendingForCurrentUser is true"), nil, http.StatusBadRequest)
			return
		}
	}

	offset := 0
	limit := 10
	offsetQueryParam := queryParams.Get("offset")
	if offsetQueryParam != "" {
		offset, err = strconv.Atoi(offsetQueryParam)
		if err != nil || offset < 0 {
			handler.logger.Errorw("error in parsing ", "err", err, "offsetQueryParam", offsetQueryParam)
			common.WriteJsonResp(w, err, "invalid offset", http.StatusBadRequest)
			return
		}
	}

	sizeQueryParam := r.URL.Query().Get("size")
	if sizeQueryParam != "" {
		limit, err = strconv.Atoi(sizeQueryParam)
		if err != nil {
			handler.logger.Errorw("request err, GetArtifactsForRollback", "err", err, "sizeQueryParam", sizeQueryParam)
			common.WriteJsonResp(w, err, "invalid size", http.StatusBadRequest)
			return
		}
	}

	searchQueryParam := r.URL.Query().Get("search") // image search string

	if resource == string(bean.SOURCE_TYPE_CI) || resource == string(bean.SOURCE_TYPE_CD) {
		// check if user has trigger access for any one env for this app
		if isAuthorised := handler.checkTriggerAccessForAnyEnv(token, appId); !isAuthorised {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}

	} else if resource == string(bean.PROMOTION_APPROVAL_PENDING_NODE) {
		// check if either user has trigger access or artifact promoter access for this env
		appRbacObj := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
		env, err := handler.environmentRepository.FindByName(resourceName)
		if err != nil {
			handler.logger.Errorw("env not found for given envName", "envName", env.Name, "err", err)
			common.WriteJsonResp(w, err, "invalid environment name in request", http.StatusBadRequest)
			return
		}
		envObjectMap, _ := handler.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId([]int{env.Id}, appId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, envObjectMap[env.Id]); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}

		triggerAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appRbacObj) &&
			handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObjectMap[env.Id])

		teamRbac := handler.enforcerUtil.GetTeamEnvRBACNameByAppId(appId, env.Id)

		approverAccess := handler.enforcer.Enforce(token, casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, teamRbac)

		if !triggerAccess && !approverAccess {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}

	artifactPromotionMaterialRequest := bean2.ArtifactPromotionMaterialRequest{
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
	artifactPromotionMaterialRequest.Resource = resource
	artifactPromotionMaterialRequest.ResourceName = resourceName
	artifactPromotionMaterialRequest.AppId = appId
	artifactPromotionMaterialRequest.WorkflowId = workflowId
	artifactPromotionMaterialRequest.Offset = offset
	artifactPromotionMaterialRequest.Limit = limit

	artifactPromotionMaterialRequest.PendingForCurrentUser = pendingForCurrentUser
	artifactPromotionMaterialRequest.UserId = userId
	artifactPromotionMaterialRequest.Token = token

	artifactPromotionMaterialResponse, err := handler.appArtifactManager.FetchMaterialForArtifactPromotion(artifactPromotionMaterialRequest, handler.CheckImagePromoterAuth)
	if err != nil {
		handler.logger.Errorw("error in fetching artifacts for given promotion request parameters", "artifactPromotionRequest", artifactPromotionMaterialRequest, "err", err)
		common.WriteJsonResp(w, errors.New("error in fetching artifacts response for given request parameters"), nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, artifactPromotionMaterialResponse, http.StatusOK)
}

func (handler RestHandlerImpl) checkTriggerAccessForAnyEnv(token string, appId int) bool {

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

func (handler RestHandlerImpl) CheckImagePromoterAuth(token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, object); !ok {
		return false
	}
	return true
}

func (handler RestHandlerImpl) FetchEnvironmentsList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if err != nil || userId == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")

	queryParams := r.URL.Query()
	workflowIdStr := queryParams.Get("workflowId")
	workflowId := 0
	if workflowIdStr == "" {
		common.WriteJsonResp(w, err, "workflowId cannot be empty", http.StatusBadRequest)
		return
	}
	workflowId, err = strconv.Atoi(workflowIdStr)
	if err != nil {
		handler.logger.Errorw("error in extracting workflowId from query param", "workflowIdStr", workflowIdStr, "err", err)
		common.WriteJsonResp(w, errors.New("workflowId should be an integer value"), nil, http.StatusBadRequest)
		return
	}
	artifactIdStr := queryParams.Get("artifactId")
	artifactId := 0
	if artifactIdStr != "" {
		artifactId, err = strconv.Atoi(artifactIdStr)
		if err != nil {
			handler.logger.Errorw("error in extracting artifactId from query param", "artifactIdStr", artifactIdStr, "err", err)
			common.WriteJsonResp(w, errors.New("artifactId should be an integer value"), nil, http.StatusBadRequest)
			return
		}
	}

	wfMeta, err := handler.promotionApprovalRequestService.GetAppAndEnvsMapByWorkflowId(workflowId)
	if err != nil {
		handler.logger.Errorw("error in finding app and env details using workflowId", "workflowId", workflowId, "err", err)
		common.WriteJsonResp(w, errors.New("error in finding application and environment details for the workflow"), nil, http.StatusInternalServerError)
		return
	}

	appRbacObject := handler.enforcerUtil.GetAppRBACName(wfMeta.AppName)
	ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appRbacObject)
	if !ok {
		common.WriteJsonResp(w, err, nil, http.StatusForbidden)
		return
	}

	environmentNames := make([]string, 0, len(wfMeta.EnvMap))
	for envName, _ := range wfMeta.EnvMap {
		environmentNames = append(environmentNames, envName)
	}
	envRbacObjectMap := handler.enforcerUtil.GetEnvRBACByAppNameAndEnvNames(wfMeta.AppName, environmentNames)
	envObjectArr := make([]string, 0)
	for _, obj := range envObjectArr {
		envObjectArr = append(envObjectArr, obj)
	}
	authorizedEnvironments := make(map[string]bool)
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObjectArr)
	for _, env := range environmentNames {
		rbacObject := envRbacObjectMap[env]
		authorizedEnvironments[env] = results[rbacObject]
	}

	resp, err := handler.promotionApprovalRequestService.FetchEnvironmentsList(wfMeta.EnvMap, wfMeta.AppId, wfMeta.AppName, authorizedEnvironments, artifactId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resp.CiSource = wfMeta.CiSourceData
	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}
