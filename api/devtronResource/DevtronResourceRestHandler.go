package devtronResource

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type DevtronResourceRestHandler interface {
	GetResourceList(w http.ResponseWriter, r *http.Request)
	GetResourceObject(w http.ResponseWriter, r *http.Request)
	CreateOrUpdateResourceObject(w http.ResponseWriter, r *http.Request)
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

const (
	PathParamKind                    = "kind"
	PathParamVersion                 = "version"
	QueryParamIsExposed              = "onlyIsExposed"
	QueryParamId                     = "id"
	QueryParamName                   = "name"
	ResourceUpdateSuccessMessage     = "Resource object updated successfully."
	DependenciesUpdateSuccessMessage = "Resource dependencies updated successfully."
)

func (handler *DevtronResourceRestHandlerImpl) GetResourceList(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	onlyIsExposed := false
	var err error
	onlyIsExposedStr := v.Get(QueryParamIsExposed)
	if len(onlyIsExposedStr) > 0 {
		onlyIsExposed, err = strconv.ParseBool(onlyIsExposedStr)
		if err != nil {
			common.WriteJsonResp(w, fmt.Errorf("invalid parameter: onlyIsExposed"), nil, http.StatusBadRequest)
			return
		}
	}
	resp, err := handler.devtronResourceService.GetDevtronResourceList(onlyIsExposed)
	if err != nil {
		handler.logger.Errorw("service error, GetResourceList", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) GetResourceObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kindVar := vars[PathParamKind]
	versionVar := vars[PathParamVersion]
	v := r.URL.Query()
	idVar := v.Get(QueryParamId)
	id, err := strconv.Atoi(idVar)
	nameVar := v.Get(QueryParamName)
	if err != nil && len(nameVar) == 0 {
		common.WriteJsonResp(w, fmt.Errorf("invalid parameter: id, name"), nil, http.StatusBadRequest)
		return
	}
	kind, subKind, statusCode, err := resolveKindSubKindValues(kindVar)
	if err != nil {
		handler.logger.Errorw("error in resolveKindSubKindValues, GetResourceObject", "err", err, "kindVar", kindVar)
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}

	reqBeanDescriptor := &bean.DevtronResourceObjectDescriptorBean{
		Kind:        kind,
		SubKind:     subKind,
		Version:     versionVar,
		OldObjectId: id, //from FE, we are taking ids of resources entry in their own respective tables
		Name:        nameVar,
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

func (handler *DevtronResourceRestHandlerImpl) CreateOrUpdateResourceObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kindVar := vars[PathParamKind]
	versionVar := vars[PathParamVersion]
	kind, subKind, statusCode, err := resolveKindSubKindValues(kindVar)
	if err != nil {
		handler.logger.Errorw("error in resolveKindSubKindValues, GetResourceObject", "err", err, "kindVar", kindVar)
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var reqBean bean.DevtronResourceObjectBean
	err = decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.checkAuthForObject(reqBean.OldObjectId, token, kind, subKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	if reqBean.DevtronResourceObjectDescriptorBean == nil {
		reqBean.DevtronResourceObjectDescriptorBean = &bean.DevtronResourceObjectDescriptorBean{}
	}
	reqBean.DevtronResourceObjectDescriptorBean.Kind = kind
	reqBean.DevtronResourceObjectDescriptorBean.SubKind = subKind
	reqBean.DevtronResourceObjectDescriptorBean.Version = versionVar
	reqBean.UserId = userId
	err = handler.devtronResourceService.CreateOrUpdateResourceObject(&reqBean)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateResourceObject", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, ResourceUpdateSuccessMessage, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) GetResourceDependencies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kindVar := vars[PathParamKind]
	versionVar := vars[PathParamVersion]
	v := r.URL.Query()
	idVar := v.Get(QueryParamId)
	id, err := strconv.Atoi(idVar)
	nameVar := v.Get(QueryParamName)
	if err != nil && len(nameVar) == 0 {
		common.WriteJsonResp(w, fmt.Errorf("invalid parameter: id, name"), nil, http.StatusBadRequest)
		return
	}
	kind, subKind, statusCode, err := resolveKindSubKindValues(kindVar)
	if err != nil {
		handler.logger.Errorw("error in resolveKindSubKindValues, GetResourceDependencies", "err", err, "kindVar", kindVar)
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.checkAuthForDependencyGet(id, token, kind, subKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	reqBeanDescriptor := &bean.DevtronResourceObjectDescriptorBean{
		Kind:        kind,
		SubKind:     subKind,
		Version:     versionVar,
		OldObjectId: id, //from FE, we are taking ids of resources entry in their own respective tables
		Name:        nameVar,
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
	vars := mux.Vars(r)
	kindVar := vars[PathParamKind]
	versionVar := vars[PathParamVersion]
	kind, subKind, statusCode, err := resolveKindSubKindValues(kindVar)
	if err != nil {
		handler.logger.Errorw("error in resolveKindSubKindValues, GetResourceObject", "err", err, "kindVar", kindVar)
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var reqBean bean.DevtronResourceObjectBean
	err = decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isValidated := handler.checkAuthForDependencyUpdate(reqBean.OldObjectId, token, kind, subKind)
	if !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	if reqBean.DevtronResourceObjectDescriptorBean == nil {
		reqBean.DevtronResourceObjectDescriptorBean = &bean.DevtronResourceObjectDescriptorBean{}
	}
	reqBean.DevtronResourceObjectDescriptorBean.Kind = kind
	reqBean.DevtronResourceObjectDescriptorBean.SubKind = subKind
	reqBean.DevtronResourceObjectDescriptorBean.Version = versionVar
	reqBean.UserId = userId
	err = handler.devtronResourceService.CreateOrUpdateResourceDependencies(&reqBean)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateResourceDependencies", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, DependenciesUpdateSuccessMessage, http.StatusOK)
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

	reqBean := &bean.DevtronResourceBean{DevtronResourceId: resourceId}

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
	var reqBean bean.DevtronResourceSchemaRequestBean
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

	//resp := &bean.UpdateSchemaResponseBean{}
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
	case bean.DevtronResourceApplication.ToString():
		switch subKind {
		case bean.DevtronResourceDevtronApplication.ToString():
			object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
			isValidated = handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object)
		case bean.DevtronResourceHelmApplication.ToString():
			object, object2 := handler.enforcerUtilHelm.GetAppRBACNameByAppId(id)
			//checking create action because need to check admin role
			if object2 == "" {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object)
			} else {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object) ||
					handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object2)
			}
		}
	case bean.DevtronResourceCluster.ToString():
		object := handler.enforcerUtil.GetClusterNameRBACObjByClusterId(id)
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, object)
	case bean.DevtronResourceJob.ToString():
		object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
		//checking create action because need to check admin role
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionCreate, object)
	default:
		isValidated = false
	}
	return isValidated
}

func (handler *DevtronResourceRestHandlerImpl) checkAuthForDependencyGet(id int, token, kind, subKind string) bool {
	isValidated := true
	switch kind {
	case bean.DevtronResourceApplication.ToString():
		switch subKind {
		case bean.DevtronResourceDevtronApplication.ToString():
			//if user has view access in this app then they can get dependency
			object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
			isValidated = handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object)
		case bean.DevtronResourceHelmApplication.ToString():
			object, object2 := handler.enforcerUtilHelm.GetAppRBACNameByAppId(id)
			//checking create action because need to check admin role
			if object2 == "" {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object)
			} else {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object) ||
					handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object2)
			}
		}
	case bean.DevtronResourceCluster.ToString():
		object := handler.enforcerUtil.GetClusterNameRBACObjByClusterId(id)
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, object)
	case bean.DevtronResourceJob.ToString():
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
	case bean.DevtronResourceApplication.ToString():
		switch subKind {
		case bean.DevtronResourceDevtronApplication.ToString():
			//if user has build and deploy access to any env in this app then they can update dependency
			//so checking app trigger access
			object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
			isValidated = handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object)
		case bean.DevtronResourceHelmApplication.ToString():
			object, object2 := handler.enforcerUtilHelm.GetAppRBACNameByAppId(id)
			//checking create action because need to check admin role
			if object2 == "" {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object)
			} else {
				isValidated = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object) ||
					handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionCreate, object2)
			}
		}
	case bean.DevtronResourceCluster.ToString():
		object := handler.enforcerUtil.GetClusterNameRBACObjByClusterId(id)
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, object)
	case bean.DevtronResourceJob.ToString():
		object := handler.enforcerUtil.GetAppRBACNameByAppId(id)
		//checking create action because need to check admin role
		isValidated = handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionCreate, object)
	default:
		isValidated = false
	}
	return isValidated
}

func resolveKindSubKindValues(kindVar string) (kind, subKind string, statusCode int, err error) {
	kindSplits := strings.Split(kindVar, "/")
	if len(kindSplits) == 1 {
		kind = kindSplits[0]
	} else if len(kindSplits) == 2 {
		kind = kindSplits[0]
		subKind = kindSplits[1]
	} else {
		err = fmt.Errorf("invalid parameter: kind")
		statusCode = http.StatusBadRequest
	}
	return kind, subKind, statusCode, err
}
