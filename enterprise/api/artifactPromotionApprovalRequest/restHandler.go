package artifactPromotionApprovalRequest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	artifactPromotion2 "github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/strings/slices"
	"net/http"
)

const unAuthorisedUser = "unauthorized user"

type RestHandler interface {
	HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request)
	FetchAwaitingApprovalEnvListForArtifact(w http.ResponseWriter, r *http.Request)
	FetchWorkflowPromoteNodeList(w http.ResponseWriter, r *http.Request)
}

type MaterialRestHandler interface {
	GetArtifactsForPromotion(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	promotionApprovalRequestService artifactPromotion2.ApprovalRequestService
	logger                          *zap.SugaredLogger
	userService                     user.UserService
	appService                      app.AppService
	enforcer                        casbin.Enforcer
	validator                       *validator.Validate
	userCommonService               user.UserCommonService
	enforcerUtil                    rbac.EnforcerUtil
	environmentService              cluster.EnvironmentService
	appArtifactManager              pipeline.AppArtifactManager
}

func NewRestHandlerImpl(
	promotionApprovalRequestService artifactPromotion2.ApprovalRequestService,
	logger *zap.SugaredLogger,
	userService user.UserService,
	appService app.AppService,
	validator *validator.Validate,
	userCommonService user.UserCommonService,
	enforcerUtil rbac.EnforcerUtil,
	environmentService cluster.EnvironmentService,
	appArtifactManager pipeline.AppArtifactManager,
	enforcer casbin.Enforcer,
) *RestHandlerImpl {
	return &RestHandlerImpl{
		promotionApprovalRequestService: promotionApprovalRequestService,
		logger:                          logger,
		userService:                     userService,
		appService:                      appService,
		validator:                       validator,
		userCommonService:               userCommonService,
		enforcerUtil:                    enforcerUtil,
		environmentService:              environmentService,
		appArtifactManager:              appArtifactManager,
		enforcer:                        enforcer,
	}
}

func (handler *RestHandlerImpl) HandleArtifactPromotionRequest(w http.ResponseWriter, r *http.Request) {
	ctx := util.NewRequestCtx(r.Context())
	isAuthorised, err := handler.userService.IsUserAdminOrManagerForAnyApp(ctx.GetUserId(), ctx.GetToken())
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	promotionRequest := &bean.ArtifactPromotionRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(promotionRequest)
	if err != nil {
		handler.logger.Errorw("err in decoding request in promotionRequest", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = promotionRequest.ValidateRequest()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	app, err := handler.appService.FindAppById(promotionRequest.AppId)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			common.WriteJsonResp(w, errors.New("app not found"), nil, http.StatusConflict)
			return
		}
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	promotionRequest.AppName = app.AppName
	authorizedEnvironments := make(map[string]bool)

	switch promotionRequest.Action {
	case constants.ACTION_PROMOTE:

		authorizedEnvironments = handler.promoteActionRbac(ctx.GetToken(), promotionRequest.AppName, promotionRequest.EnvironmentNames)

	case constants.ACTION_APPROVE:

		authorizedEnvironments = handler.approveActionRbac(ctx.GetToken(), promotionRequest.AppName, promotionRequest.EnvironmentNames)

	case constants.ACTION_CANCEL:
		// get this info from service layer
		if ok := handler.cancelActionRbac(ctx.GetToken(), w, promotionRequest.PromotionRequestId); !ok {
			return
		}
	}

	resp, err := handler.promotionApprovalRequestService.HandleArtifactPromotionRequest(ctx, promotionRequest, authorizedEnvironments)
	if err != nil {
		handler.logger.Errorw("error in handling promotion artifact request", "promotionRequest", promotionRequest, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	for i, envMetadata := range resp {
		resp[i].PromotionValidationState = envMetadata.PromotionValidationMessage.GetValidationState()
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *RestHandlerImpl) FetchWorkflowPromoteNodeList(w http.ResponseWriter, r *http.Request) {

	ctx := util.NewRequestCtx(r.Context())
	workflowId := 0
	workflowId, err := common.ExtractIntQueryParam(w, r, "workflowId", &workflowId)
	if err != nil {
		handler.logger.Errorw("error in extracting workflowId from query param", "err", err)
		return
	}

	artifactId := 0
	artifactId, err = common.ExtractIntQueryParam(w, r, "artifactId", &artifactId)
	if err != nil {
		handler.logger.Errorw("error in extracting artifactId from query param", "err", err)
		return
	}

	resp, err := handler.promotionApprovalRequestService.FetchWorkflowPromoteNodeList(ctx, workflowId, artifactId, handler.promoteActionRbac)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	for i, envMetadata := range resp.Environments {
		resp.Environments[i].PromotionValidationState = envMetadata.PromotionValidationMessage.GetValidationState()
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}

func (handler *RestHandlerImpl) FetchAwaitingApprovalEnvListForArtifact(w http.ResponseWriter, r *http.Request) {
	ctx := util.NewRequestCtx(r.Context())

	isAuthorised, err := handler.userService.IsUserAdminOrManagerForAnyApp(ctx.GetUserId(), ctx.GetToken())
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
		return
	}

	environmentName := queryParams.Get("environmentName")

	environmentApprovalMetadata, err := handler.promotionApprovalRequestService.FetchApprovalAllowedEnvList(ctx, artifactId, environmentName, handler.enforcerUtil.CheckImagePromoterBulkAuth)
	if err != nil {
		handler.logger.Errorw("error in fetching environments with pending approval for artifact", "artifactId", artifactId, "err", err)
		return
	}
	common.WriteJsonResp(w, nil, environmentApprovalMetadata, http.StatusOK)
	return

}

func (handler *RestHandlerImpl) GetArtifactsForPromotion(w http.ResponseWriter, r *http.Request) {

	ctx := util.NewRequestCtx(r.Context())

	artifactPromotionMaterialRequest, err := handler.parsePromotionMaterialRequest(w, r)
	if err != nil {
		return
	}

	validRequest := handler.validatePromotionMaterialRequest(w, artifactPromotionMaterialRequest)
	if !validRequest {
		return
	}

	// todo: ayush
	if artifactPromotionMaterialRequest.Resource == string(constants.SOURCE_TYPE_CI) || artifactPromotionMaterialRequest.Resource == string(constants.SOURCE_TYPE_CD) {
		// check if user has trigger access for any one env for this app
		if hasTriggerAccess := handler.checkTriggerAccessForAnyEnv(ctx.GetToken(), artifactPromotionMaterialRequest.AppId); !hasTriggerAccess {
			common.WriteJsonResp(w, fmt.Errorf(unAuthorisedUser), unAuthorisedUser, http.StatusForbidden)
			return
		}
	} else if artifactPromotionMaterialRequest.Resource == string(constants.PROMOTION_APPROVAL_PENDING_NODE) && !artifactPromotionMaterialRequest.PendingForCurrentUser {
		// check if either user has trigger access or artifact promoter access for this env
		appRbacObj := handler.enforcerUtil.GetAppRBACNameByAppId(artifactPromotionMaterialRequest.AppId)
		env, err := handler.environmentService.FindOne(artifactPromotionMaterialRequest.ResourceName)
		if err != nil {
			handler.logger.Errorw("env not found for given envName", "envName", artifactPromotionMaterialRequest.ResourceName, "err", err)
			common.WriteJsonResp(w, err, "invalid environment name in request", http.StatusBadRequest)
			return
		}
		envObjectMap, _ := handler.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId([]int{env.Id}, artifactPromotionMaterialRequest.AppId)
		if ok := handler.enforcer.Enforce(ctx.GetToken(), casbin.ResourceEnvironment, casbin.ActionGet, envObjectMap[env.Id]); !ok {
			common.WriteJsonResp(w, err, unAuthorisedUser, http.StatusForbidden)
			return
		}

		triggerAccess := handler.enforcer.Enforce(ctx.GetToken(), casbin.ResourceApplications, casbin.ActionTrigger, appRbacObj) &&
			handler.enforcer.Enforce(ctx.GetToken(), casbin.ResourceEnvironment, casbin.ActionTrigger, envObjectMap[env.Id])

		teamRbac := handler.enforcerUtil.GetTeamEnvRBACNameByAppId(artifactPromotionMaterialRequest.AppId, env.Id)

		approverAccess := handler.enforcer.Enforce(ctx.GetToken(), casbin.ResourceApprovalPolicy, casbin.ActionArtifactPromote, teamRbac)

		if !triggerAccess && !approverAccess {
			common.WriteJsonResp(w, err, unAuthorisedUser, http.StatusForbidden)
			return
		}
	}

	artifactPromotionMaterialResponse, err := handler.appArtifactManager.FetchMaterialForArtifactPromotion(ctx, *artifactPromotionMaterialRequest, handler.enforcerUtil.CheckImagePromoterBulkAuth)
	if err != nil {
		handler.logger.Errorw("error in fetching artifacts for given promotion request parameters", "err", err)
		common.WriteJsonResp(w, errors.New("error in fetching artifacts response for given request parameters"), nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, artifactPromotionMaterialResponse, http.StatusOK)
}

func (handler *RestHandlerImpl) getAppAndEnvObjectByCdPipelineId(cdPipelineId int) (string, string) {
	object := handler.enforcerUtil.GetAppAndEnvObjectByPipelineIds([]int{cdPipelineId})
	rbacObjects := object[cdPipelineId]
	return rbacObjects[0], rbacObjects[1]
}

func (handler *RestHandlerImpl) parsePromotionMaterialRequest(w http.ResponseWriter, r *http.Request) (*bean2.ArtifactPromotionMaterialRequest, error) {

	queryParams := r.URL.Query()
	resource := queryParams.Get("resource")
	resourceName := queryParams.Get("resourceName")

	pendingForCurrentUser, err := common.ExtractBooleanQueryParam(w, r, "pendingForCurrentUser", false)
	if err != nil {
		handler.logger.Errorw("error in parsing pendingForCurrentUser from string to bool", "pendingForCurrentUser", queryParams.Get("pendingForCurrentUser"))
		return nil, err
	}

	workflowId := 0
	workflowId, err = common.ExtractIntQueryParam(w, r, "workflowId", &workflowId)
	if err != nil {
		handler.logger.Errorw("error in parsing workflowId from string to int", "workflowId", queryParams.Get("workflowId"))
		return nil, err
	}

	appId := 0
	appId, err = common.ExtractIntQueryParam(w, r, "appId", &appId)
	if err != nil {
		handler.logger.Errorw("error in parsing appId from string to int", "workflowId", queryParams.Get("appId"))
		return nil, err
	}

	offsetDefault := 0
	offset, err := common.ExtractIntQueryParam(w, r, "offset", &offsetDefault)
	if err != nil {
		handler.logger.Errorw("error in parsing offset from string to int", "workflowId", queryParams.Get("offset"))
		return nil, err
	}

	limitDefault := 20
	limit, err := common.ExtractIntQueryParam(w, r, "size", &limitDefault)
	if err != nil {
		handler.logger.Errorw("error in parsing limit from string to int", "workflowId", queryParams.Get("size"))
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

func (handler *RestHandlerImpl) cancelActionRbac(token string, w http.ResponseWriter, promotionRequestId int) bool {
	artifactPromotionDao, err := handler.promotionApprovalRequestService.GetPromotionRequestById(promotionRequestId)
	if err != nil {
		handler.logger.Errorw("error in fetching promotion request by id", "promotionRequestId", promotionRequestId, "err", err)
		common.WriteJsonResp(w, errors.New("error in fetching promotion request "), nil, http.StatusInternalServerError)
		return false
	}
	appRbacObj, envRbacObj := handler.getAppAndEnvObjectByCdPipelineId(artifactPromotionDao.DestinationPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appRbacObj); !ok {
		common.WriteJsonResp(w, fmt.Errorf(unAuthorisedUser), unAuthorisedUser, http.StatusForbidden)
		return false
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envRbacObj); !ok {
		common.WriteJsonResp(w, err, unAuthorisedUser, http.StatusForbidden)
		return false
	}
	return true
}
