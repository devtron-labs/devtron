package imageDigestPolicy

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ImageDigestPolicyRestHandler interface {
	GetAllImageDigestPolicies(w http.ResponseWriter, r *http.Request)
	SaveOrUpdateImageDigestPolicy(w http.ResponseWriter, r *http.Request)
}

type ImageDigestPolicyRestHandlerImpl struct {
	logger                   *zap.SugaredLogger
	userAuthService          user.UserService
	enforcerUtil             rbac.EnforcerUtil
	enforcer                 casbin.Enforcer
	validator                *validator.Validate
	imageDigestPolicyService imageDigestPolicy.ImageDigestPolicyService
}

func NewImageDigestPolicyRestHandlerImpl(logger *zap.SugaredLogger,
	userAuthService user.UserService,
	enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
	imageDigestPolicyService imageDigestPolicy.ImageDigestPolicyService) (*ImageDigestPolicyRestHandlerImpl, error) {
	err := validator.RegisterValidation("validate-image-digest-policy-type", imageDigestPolicy.ValidateImageDigestPolicyType)
	if err != nil {
		logger.Errorw("error in registering image digest policy validation function", "err", err)
		return nil, err
	}
	return &ImageDigestPolicyRestHandlerImpl{
		logger:                   logger,
		userAuthService:          userAuthService,
		enforcerUtil:             enforcerUtil,
		enforcer:                 enforcer,
		validator:                validator,
		imageDigestPolicyService: imageDigestPolicyService,
	}, nil
}

func (handler ImageDigestPolicyRestHandlerImpl) GetAllImageDigestPolicies(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	if !isSuperAdmin {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusUnauthorized)
		return
	}

	res, err := handler.imageDigestPolicyService.GetAllPoliciesConfiguredForClusterOrEnv()
	if err != nil {
		handler.logger.Errorw("error in getting image digest policies", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler ImageDigestPolicyRestHandlerImpl) SaveOrUpdateImageDigestPolicy(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	if !isSuperAdmin {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	req := &imageDigestPolicy.PolicyBean{}
	err := decoder.Decode(req)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "request", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.validator.Struct(req)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "request", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusInternalServerError)
		return
	}
	req.UserId = userId

	policyRequest, err := handler.imageDigestPolicyService.CreateOrUpdatePolicyForCluster(req)
	if err != nil {
		handler.logger.Errorw("service err, imageDigestPolicyService", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}

	common.WriteJsonResp(w, nil, policyRequest, http.StatusOK)
}
