package imageDigestPolicy

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
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
	imageDigestPolicyService imageDigestPolicy.ImageDigestQualifierMappingService
}

func NewImageDigestPolicyRestHandlerImpl(logger *zap.SugaredLogger,
	userAuthService user.UserService,
	enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
	imageDigestPolicyService imageDigestPolicy.ImageDigestQualifierMappingService) *ImageDigestPolicyRestHandlerImpl {
	return &ImageDigestPolicyRestHandlerImpl{
		logger:                   logger,
		userAuthService:          userAuthService,
		enforcerUtil:             enforcerUtil,
		enforcer:                 enforcer,
		validator:                validator,
		imageDigestPolicyService: imageDigestPolicyService,
	}
}

func (handler ImageDigestPolicyRestHandlerImpl) GetAllImageDigestPolicies(w http.ResponseWriter, r *http.Request) {
	res, err := handler.imageDigestPolicyService.GetAllConfiguredPolicies()
	if err != nil {
		handler.logger.Errorw("error in getting active resource filters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler ImageDigestPolicyRestHandlerImpl) SaveOrUpdateImageDigestPolicy(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := &imageDigestPolicy.PolicyRequest{}
	err := decoder.Decode(req)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "request", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	policyRequest, err := handler.imageDigestPolicyService.CreateOrUpdatePolicyForCluster(req)
	if err != nil {
		handler.logger.Errorw("service err, imageDigestPolicyService", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, policyRequest, http.StatusOK)
}
