package artifactPromotionPolicy

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	artifactPromotion2 "github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/read"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type RestHandler interface {
	CreatePolicy(w http.ResponseWriter, r *http.Request)
	UpdatePolicy(w http.ResponseWriter, r *http.Request)
	DeletePolicy(w http.ResponseWriter, r *http.Request)
	GetPolicyByName(w http.ResponseWriter, r *http.Request)
	GetPoliciesMetadata(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	artifactPromotionReadService read.ArtifactPromotionDataReadService
	promotionPolicyCUDService    artifactPromotion2.PromotionPolicyCUDService
	userService                  user.UserService
	enforcer                     casbin.Enforcer
	validator                    *validator.Validate
	logger                       *zap.SugaredLogger
}

func NewArtifactPromotionPolicyRestHandlerImpl(
	artifactPromotionReadService read.ArtifactPromotionDataReadService,
	promotionPolicyCUDService artifactPromotion2.PromotionPolicyCUDService,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
	logger *zap.SugaredLogger) *RestHandlerImpl {
	return &RestHandlerImpl{
		artifactPromotionReadService: artifactPromotionReadService,
		promotionPolicyCUDService:    promotionPolicyCUDService,
		userService:                  userService,
		enforcer:                     enforcer,
		validator:                    validator,
		logger:                       logger,
	}
}

func (handler RestHandlerImpl) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return

	}

	promotionPolicy := &bean.PromotionPolicy{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(promotionPolicy)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.validator.Struct(promotionPolicy)
	if err != nil {
		handler.logger.Errorw("error in validating the request payload", "err", err, "payload", promotionPolicy)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.promotionPolicyCUDService.CreatePolicy(userId, promotionPolicy)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (handler RestHandlerImpl) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	policyName := vars["name"]
	promotionPolicy := &bean.PromotionPolicy{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(promotionPolicy)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.validator.Struct(promotionPolicy)
	if err != nil {
		handler.logger.Errorw("error in validating the request payload", "err", err, "payload", promotionPolicy)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.promotionPolicyCUDService.UpdatePolicy(userId, policyName, promotionPolicy)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)

}

func (handler RestHandlerImpl) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return

	}

	vars := mux.Vars(r)
	policyName := vars["name"]
	err = handler.promotionPolicyCUDService.DeletePolicy(userId, policyName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusNoContent)
}

func (handler RestHandlerImpl) GetPoliciesMetadata(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return

	}
	queryParams := r.URL.Query()
	sortBy := queryParams.Get("sortBy")
	sortOrder := queryParams.Get("sortOrder")
	search := queryParams.Get("search")

	if sortBy == "" {
		sortBy = "policyName"
	}

	if sortOrder == "" {
		sortOrder = "ASC"
	}

	listFilter := bean.PromotionPolicyMetaRequest{
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	policies, err := handler.artifactPromotionReadService.GetPoliciesMetadata(listFilter)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, policies, http.StatusOK)

}

func (handler RestHandlerImpl) GetPolicyByName(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	policyName := vars["name"]
	policy, err := handler.artifactPromotionReadService.GetPromotionPolicyByName(policyName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, policy, http.StatusOK)
}
