package devtronResource

import (
	"encoding/json"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	serviceBean "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type DevtronResourceRestHandler interface {
	GetAllDevtronResourcesList(w http.ResponseWriter, r *http.Request)
	GetResourceObjectListByKindAndVersion(w http.ResponseWriter, r *http.Request)
	GetResourceObject(w http.ResponseWriter, r *http.Request)
	CreateResourceObject(w http.ResponseWriter, r *http.Request)
	CreateOrUpdateResourceObject(w http.ResponseWriter, r *http.Request)
	PatchResourceObject(w http.ResponseWriter, r *http.Request)
	DeleteResourceObject(w http.ResponseWriter, r *http.Request)
	GetResourceDependencies(w http.ResponseWriter, r *http.Request)
	CreateOrUpdateResourceDependencies(w http.ResponseWriter, r *http.Request)
	GetSchema(w http.ResponseWriter, r *http.Request)
	UpdateSchema(w http.ResponseWriter, r *http.Request)
}

type DevtronResourceRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	enforcer               casbin.Enforcer
	enforcerUtil           rbac.EnforcerUtil
	enforcerUtilHelm       rbac.EnforcerUtilHelm
	userService            user.UserService
	validator              *validator.Validate
	devtronResourceService devtronResource.DevtronResourceService
}

func NewDevtronResourceRestHandlerImpl(logger *zap.SugaredLogger, userService user.UserService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, enforcerUtilHelm rbac.EnforcerUtilHelm, validator *validator.Validate,
	devtronResourceService devtronResource.DevtronResourceService) *DevtronResourceRestHandlerImpl {
	return &DevtronResourceRestHandlerImpl{
		logger:                 logger,
		enforcer:               enforcer,
		userService:            userService,
		enforcerUtil:           enforcerUtil,
		enforcerUtilHelm:       enforcerUtilHelm,
		validator:              validator,
		devtronResourceService: devtronResourceService,
	}
}

func (handler *DevtronResourceRestHandlerImpl) GetAllDevtronResourcesList(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	onlyIsExposed := false
	var err error
	onlyIsExposedStr := v.Get(apiBean.QueryParamIsExposed)
	if len(onlyIsExposedStr) > 0 {
		onlyIsExposed, err = strconv.ParseBool(onlyIsExposedStr)
		if err != nil {
			common.WriteJsonResp(w, fmt.Errorf("invalid parameter: onlyIsExposed"), nil, http.StatusBadRequest)
			return
		}
	}
	resp, err := handler.devtronResourceService.GetDevtronResourceList(onlyIsExposed)
	if err != nil {
		handler.logger.Errorw("service error, GetAllDevtronResourcesList", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) GetResourceObjectListByKindAndVersion(w http.ResponseWriter, r *http.Request) {
	kind, subKind, versionVar, caughtError := getKindSubKindVersion(w, r)
	if caughtError {
		return
	}
	v := r.URL.Query()
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	queryParams := apiBean.GetResourceListQueryParams{}
	err := decoder.Decode(&queryParams, v)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	resp, err := handler.devtronResourceService.ListResourceObjectByKindAndVersion(kind, subKind, versionVar, queryParams.IsLite, queryParams.FetchChild)
	if err != nil {
		handler.logger.Errorw("service error, GetAllDevtronResourcesList", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) GetResourceObject(w http.ResponseWriter, r *http.Request) {
	reqBeanDescriptor, caughtError := getDescriptorBeanObj(w, r)
	if caughtError {
		return
	}
	resp, err := handler.devtronResourceService.GetResourceObject(reqBeanDescriptor)
	if err != nil {
		handler.logger.Errorw("service error, GetResourceObject", "err", err, "request", reqBeanDescriptor)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func getKindSubKindVersion(w http.ResponseWriter, r *http.Request) (kind string, subKind string, version string, caughtError bool) {
	vars := mux.Vars(r)
	kindVar := vars[apiBean.PathParamKind]
	versionVar := vars[apiBean.PathParamVersion]
	kind, subKind, statusCode, err := resolveKindSubKindValues(kindVar)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		caughtError = true
	}
	return kind, subKind, versionVar, caughtError
}

func getDescriptorBeanObj(w http.ResponseWriter, r *http.Request) (reqBeanDescriptor *serviceBean.DevtronResourceObjectDescriptorBean, caughtError bool) {
	kind, subKind, versionVar, caughtError := getKindSubKindVersion(w, r)
	if caughtError {
		return
	}
	v := r.URL.Query()
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	queryParams := apiBean.GetResourceQueryParams{}
	err := decoder.Decode(&queryParams, v)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if queryParams.Id == 0 && len(queryParams.Identifier) == 0 {
		common.WriteJsonResp(w, fmt.Errorf("invalid parameter: id or identifier"), nil, http.StatusBadRequest)
		caughtError = true
		return nil, caughtError
	}
	reqBeanDescriptor = &serviceBean.DevtronResourceObjectDescriptorBean{
		Kind:         kind,
		SubKind:      subKind,
		Version:      versionVar,
		OldObjectId:  queryParams.Id,
		Identifier:   queryParams.Identifier,
		UIComponents: queryParams.Component,
	}
	return reqBeanDescriptor, caughtError
}

func (handler *DevtronResourceRestHandlerImpl) CreateResourceObject(w http.ResponseWriter, r *http.Request) {
	ctx := util.NewRequestCtx(r.Context())
	kind, subKind, version, caughtError := getKindSubKindVersion(w, r)
	if caughtError {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var reqBean serviceBean.DevtronResourceObjectBean
	err := decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if reqBean.DevtronResourceObjectDescriptorBean == nil {
		reqBean.DevtronResourceObjectDescriptorBean = &serviceBean.DevtronResourceObjectDescriptorBean{}
	}
	reqBean.Kind = kind
	reqBean.SubKind = subKind
	reqBean.Version = version

	token := r.Header.Get("token")
	// RBAC enforcer applying
	isValidated := handler.checkAuthForObject(reqBean.OldObjectId, token, kind, subKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends
	reqBean.UserId = ctx.GetUserId()
	err = handler.devtronResourceService.CreateResourceObject(ctx, &reqBean)
	if err != nil {
		handler.logger.Errorw("service error, CreateResourceObject", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, apiBean.ResourceCreateSuccessMessage, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) CreateOrUpdateResourceObject(w http.ResponseWriter, r *http.Request) {
	ctx := util.NewRequestCtx(r.Context())
	kind, subKind, versionVar, caughtError := getKindSubKindVersion(w, r)
	if caughtError {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var reqBean serviceBean.DevtronResourceObjectBean
	err := decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.checkAuthForObject(reqBean.OldObjectId, token, kind, subKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	if reqBean.DevtronResourceObjectDescriptorBean == nil {
		reqBean.DevtronResourceObjectDescriptorBean = &serviceBean.DevtronResourceObjectDescriptorBean{}
	}
	reqBean.Kind = kind
	reqBean.SubKind = subKind
	reqBean.Version = versionVar
	reqBean.UserId = ctx.GetUserId()
	err = handler.devtronResourceService.CreateOrUpdateResourceObject(ctx, &reqBean)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateResourceObject", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, apiBean.ResourceUpdateSuccessMessage, http.StatusOK)
	return
}
func (handler *DevtronResourceRestHandlerImpl) PatchResourceObject(w http.ResponseWriter, r *http.Request) {
	ctx := util.NewRequestCtx(r.Context())
	kind, subKind, version, caughtError := getKindSubKindVersion(w, r)
	if caughtError {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var reqBean serviceBean.DevtronResourceObjectBean
	err := decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// struct tags validation
	err = handler.validator.Struct(reqBean)
	if err != nil {
		handler.logger.Errorw("validation err, PatchResourceObject", "err", err, "payload", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	// RBAC enforcer applying
	isValidated := handler.checkAuthForObject(reqBean.OldObjectId, token, kind, subKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends
	if reqBean.DevtronResourceObjectDescriptorBean == nil {
		reqBean.DevtronResourceObjectDescriptorBean = &serviceBean.DevtronResourceObjectDescriptorBean{}
	} else {
		reqBean.DevtronResourceObjectDescriptorBean.Kind = kind
		reqBean.DevtronResourceObjectDescriptorBean.SubKind = subKind
		reqBean.DevtronResourceObjectDescriptorBean.Version = version
	}
	reqBean.UserId = ctx.GetUserId()
	resp, err := handler.devtronResourceService.PatchResourceObject(ctx, &reqBean)
	if err != nil {
		handler.logger.Errorw("service error, CreateResourceObject", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) DeleteResourceObject(w http.ResponseWriter, r *http.Request) {
	ctx := util.NewRequestCtx(r.Context())
	reqBeanDescriptor, caughtError := getDescriptorBeanObj(w, r)
	if caughtError {
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	resp, err := handler.devtronResourceService.DeleteResourceObject(ctx, reqBeanDescriptor)
	if err != nil {
		handler.logger.Errorw("service error, GetResourceObject", "err", err, "request", reqBeanDescriptor)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) GetResourceDependencies(w http.ResponseWriter, r *http.Request) {
	reqBeanDescriptor, caughtError := getDescriptorBeanObj(w, r)
	if caughtError {
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.checkAuthForDependencyGet(reqBeanDescriptor.OldObjectId, token, reqBeanDescriptor.Kind, reqBeanDescriptor.SubKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	resp, err := handler.devtronResourceService.GetResourceDependencies(reqBeanDescriptor)
	if err != nil {
		handler.logger.Errorw("service error, GetResourceDependencies", "err", err, "request", reqBeanDescriptor)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) CreateOrUpdateResourceDependencies(w http.ResponseWriter, r *http.Request) {
	ctx := util.NewRequestCtx(r.Context())
	kind, subKind, versionVar, caughtError := getKindSubKindVersion(w, r)
	if caughtError {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var reqBean serviceBean.DevtronResourceObjectBean
	err := decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.checkAuthForDependencyUpdate(reqBean.OldObjectId, token, kind, subKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	if reqBean.DevtronResourceObjectDescriptorBean == nil {
		reqBean.DevtronResourceObjectDescriptorBean = &serviceBean.DevtronResourceObjectDescriptorBean{}
	}
	reqBean.Kind = kind
	reqBean.SubKind = subKind
	reqBean.Version = versionVar
	reqBean.UserId = ctx.GetUserId()
	err = handler.devtronResourceService.CreateOrUpdateResourceDependencies(r.Context(), &reqBean)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateResourceDependencies", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, apiBean.DependenciesUpdateSuccessMessage, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) GetSchema(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	//RBAC block starts from here
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC block ends here

	vars := r.URL.Query()
	resourceIdString := vars.Get("resourceId")
	resourceId, err := strconv.Atoi(resourceIdString)
	if err != nil {
		handler.logger.Errorw("error in converting string to integer", "err", err, "resourceIdString", resourceIdString)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	reqBean := &serviceBean.DevtronResourceBean{DevtronResourceId: resourceId}

	resp, err := handler.devtronResourceService.GetSchema(reqBean)
	if err != nil {
		handler.logger.Errorw("service error, GetSchema", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) UpdateSchema(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	//RBAC block starts from here
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC block ends here

	vars := r.URL.Query()
	dryRunString := vars.Get("dryRun")
	dryRun, err := strconv.ParseBool(dryRunString)

	decoder := json.NewDecoder(r.Body)
	var reqBean serviceBean.DevtronResourceSchemaRequestBean
	err = decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	reqBean.UserId = int(userId)

	err = handler.validator.Struct(reqBean)
	if err != nil {
		handler.logger.Errorw("validate err, CreateChartGroup", "err", err, "payload", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//resp := &serviceBean.UpdateSchemaResponseBean{}
	resp, err := handler.devtronResourceService.UpdateSchema(&reqBean, dryRun)
	if err != nil {
		handler.logger.Errorw("service error, GetSchema", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) checkAuthForObject(id int, token, kind, subKind string) bool {
	isValidated := true
	switch kind {
	case serviceBean.DevtronResourceApplication.ToString():
		switch subKind {
		case serviceBean.DevtronResourceDevtronApplication.ToString():
			object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
			isValidated = handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object)
		case serviceBean.DevtronResourceHelmApplication.ToString():
			object, object2 := handler.enforcerUtilHelm.GetAppRBACNameByAppId(id)
			//checking create action because need to check admin role
			if object2 == "" {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object)
			} else {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object) ||
					handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object2)
			}
		}
	case serviceBean.DevtronResourceCluster.ToString():
		object := handler.enforcerUtil.GetClusterNameRBACObjByClusterId(id)
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, object)
	case serviceBean.DevtronResourceJob.ToString():
		object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
		//checking create action because need to check admin role
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionCreate, object)
	case serviceBean.DevtronResourceReleaseTrack.ToString():
	case serviceBean.DevtronResourceRelease.ToString():
		// checking for super admin access
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	default:
		isValidated = false
	}
	return isValidated
}

func (handler *DevtronResourceRestHandlerImpl) checkAuthForDependencyGet(id int, token, kind, subKind string) bool {
	isValidated := true
	switch kind {
	case serviceBean.DevtronResourceApplication.ToString():
		switch subKind {
		case serviceBean.DevtronResourceDevtronApplication.ToString():
			//if user has view access in this app then they can get dependency
			object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
			isValidated = handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object)
		case serviceBean.DevtronResourceHelmApplication.ToString():
			object, object2 := handler.enforcerUtilHelm.GetAppRBACNameByAppId(id)
			//checking create action because need to check admin role
			if object2 == "" {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object)
			} else {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object) ||
					handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object2)
			}
		}
	case serviceBean.DevtronResourceCluster.ToString():
		object := handler.enforcerUtil.GetClusterNameRBACObjByClusterId(id)
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, object)
	case serviceBean.DevtronResourceJob.ToString():
		object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
		//checking create action because need to check admin role
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionGet, object)
	default:
		isValidated = false
	}
	return isValidated
}

func (handler *DevtronResourceRestHandlerImpl) checkAuthForDependencyUpdate(id int, token, kind, subKind string) bool {
	isValidated := true
	switch kind {
	case serviceBean.DevtronResourceApplication.ToString():
		switch subKind {
		case serviceBean.DevtronResourceDevtronApplication.ToString():
			//if user has build and deploy access to any env in this app then they can update dependency
			//so checking app trigger access
			object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
			isValidated = handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object)
		case serviceBean.DevtronResourceHelmApplication.ToString():
			object, object2 := handler.enforcerUtilHelm.GetAppRBACNameByAppId(id)
			//checking create action because need to check admin role
			if object2 == "" {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object)
			} else {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object) ||
					handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object2)
			}
		}
	case serviceBean.DevtronResourceCluster.ToString():
		object := handler.enforcerUtil.GetClusterNameRBACObjByClusterId(id)
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, object)
	case serviceBean.DevtronResourceJob.ToString():
		object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
		//checking create action because need to check admin role
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionCreate, object)
	default:
		isValidated = false
	}
	return isValidated
}

func resolveKindSubKindValues(kindVar string) (kind, subKind string, statusCode int, err error) {
	kind, subKind, err = devtronResource.GetKindAndSubKindFrom(kindVar)
	if err != nil {
		err = fmt.Errorf("invalid parameter: kind")
		statusCode = http.StatusBadRequest
	}
	return kind, subKind, statusCode, err
}
