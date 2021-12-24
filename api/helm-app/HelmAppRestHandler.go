package client

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type HelmAppRestHandler interface {
	ListApplications(w http.ResponseWriter, r *http.Request)
	GetApplicationDetail(w http.ResponseWriter, r *http.Request)
	GetResource(w http.ResponseWriter, r *http.Request)
	UpdateResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
	ListEvents(w http.ResponseWriter, r *http.Request)
}
type HelmAppRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	helmAppService HelmAppService
}

func NewHelmAppRestHandlerImpl(logger *zap.SugaredLogger,
	helmAppService HelmAppService) *HelmAppRestHandlerImpl {
	return &HelmAppRestHandlerImpl{
		logger:         logger,
		helmAppService: helmAppService,
	}
}

func (handler *HelmAppRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["clusterIds"]
	clusterIdSlices := strings.Split(clusterIdString, ",")
	var clusterIds []int
	for _, is := range clusterIdSlices {
		if len(is) == 0 {
			continue
		}
		j, err := strconv.Atoi(is)
		if err != nil {
			handler.logger.Errorw("request err, CreateUser", "err", err, "payload", clusterIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		clusterIds = append(clusterIds, j)
	}
	handler.helmAppService.ListHelmApplications(clusterIds, w)
}

func (handler *HelmAppRestHandlerImpl) GetApplicationDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appdetail, err := handler.helmAppService.GetApplicationDetail(context.Background(), appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, appdetail, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) GetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//TODO : add rbac
	resource, err := handler.helmAppService.GetResource(appIdentifier, &request)
	if err != nil {
		handler.logger.Errorw("error in getting resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) UpdateResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//TODO : add rbac
	resource, err := handler.helmAppService.UpdateResource(appIdentifier, &request)
	if err != nil {
		handler.logger.Errorw("error in updating resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) DeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//TODO : add rbac
	resource, err := handler.helmAppService.DeleteResource(appIdentifier, &request)
	if err != nil {
		handler.logger.Errorw("error in deleting resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) ListEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request application.K8sRequestBean
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//TODO : add rbac
	events, err := handler.helmAppService.ListEvents(appIdentifier, &request)
	if err != nil {
		handler.logger.Errorw("error in getting events list", "err", err)
		common.WriteJsonResp(w, err, events, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, events, http.StatusOK)
}
