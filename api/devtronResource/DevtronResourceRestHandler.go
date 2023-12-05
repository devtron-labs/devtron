package devtronResource

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type DevtronResourceRestHandler interface {
	GetResourceList(w http.ResponseWriter, r *http.Request)
	GetResourceObject(w http.ResponseWriter, r *http.Request)
	UpdateResourceObject(w http.ResponseWriter, r *http.Request)
}

type DevtronResourceRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	enforcer               casbin.Enforcer
	validator              *validator.Validate
	devtronResourceService devtronResource.DevtronResourceService
}

func NewDevtronResourceRestHandlerImpl(logger *zap.SugaredLogger,
	enforcer casbin.Enforcer, validator *validator.Validate,
	devtronResourceService devtronResource.DevtronResourceService) *DevtronResourceRestHandlerImpl {
	return &DevtronResourceRestHandlerImpl{
		logger:                 logger,
		enforcer:               enforcer,
		validator:              validator,
		devtronResourceService: devtronResourceService,
	}
}

const (
	PathParamKind                = "kind"
	PathParamVersion             = "version"
	QueryParamId                 = "id"
	QueryParamName               = "name"
	ResourceUpdateSuccessMessage = "Resource object updated successfully."
)

func (handler *DevtronResourceRestHandlerImpl) GetResourceList(w http.ResponseWriter, r *http.Request) {
	//TODO: implement this
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
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *DevtronResourceRestHandlerImpl) UpdateResourceObject(w http.ResponseWriter, r *http.Request) {
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
	if reqBean.DevtronResourceObjectDescriptorBean == nil {
		reqBean.DevtronResourceObjectDescriptorBean = &bean.DevtronResourceObjectDescriptorBean{}
	}
	reqBean.DevtronResourceObjectDescriptorBean.Kind = kind
	reqBean.DevtronResourceObjectDescriptorBean.SubKind = subKind
	reqBean.DevtronResourceObjectDescriptorBean.Version = versionVar
	err = handler.devtronResourceService.CreateOrUpdateResourceObject(&reqBean)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateResourceObject", "err", err, "request", reqBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, err, ResourceUpdateSuccessMessage, http.StatusOK)
	return
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
