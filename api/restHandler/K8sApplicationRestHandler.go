package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"go.uber.org/zap"
	"net/http"
)

type K8sApplicationRestHandler interface {
	GetResource(w http.ResponseWriter, r *http.Request)
	UpdateResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
	ListEvents(w http.ResponseWriter, r *http.Request)
}

type K8sApplicationRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	k8sApplication application.K8sApplicationService
}

func NewK8sApplicationRestHandlerImpl(logger *zap.SugaredLogger, k8sApplication application.K8sApplicationService) *K8sApplicationRestHandlerImpl {
	return &K8sApplicationRestHandlerImpl{
		logger:         logger,
		k8sApplication: k8sApplication,
	}
}

func (impl K8sApplicationRestHandlerImpl) GetResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resource, err := impl.k8sApplication.GetResource(&request)
	if err!=nil{
		common.WriteJsonResp(w,err,resource,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,resource,http.StatusOK)
	return
}

func (impl K8sApplicationRestHandlerImpl) UpdateResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resource, err := impl.k8sApplication.UpdateResource(&request)
	if err!=nil{
		common.WriteJsonResp(w,err,resource,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,resource,http.StatusOK)
	return
}

func (impl K8sApplicationRestHandlerImpl) DeleteResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resource, err := impl.k8sApplication.DeleteResource(&request)
	if err!=nil{
		common.WriteJsonResp(w,err,resource,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,resource,http.StatusOK)
	return
}

func (impl K8sApplicationRestHandlerImpl) ListEvents(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	events, err := impl.k8sApplication.ListEvents(&request)
	if err!=nil{
		common.WriteJsonResp(w,err,events,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,events,http.StatusOK)
	return
}
