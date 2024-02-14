package argoApplication

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/argoApplication"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type ArgoApplicationRestHandler interface {
	ListApplications(w http.ResponseWriter, r *http.Request)
	GetApplicationDetail(w http.ResponseWriter, r *http.Request)
}

type ArgoApplicationRestHandlerImpl struct {
	argoApplicationService argoApplication.ArgoApplicationService
	logger                 *zap.SugaredLogger
}

func NewArgoApplicationRestHandlerImpl(argoApplicationService argoApplication.ArgoApplicationService,
	logger *zap.SugaredLogger) *ArgoApplicationRestHandlerImpl {
	return &ArgoApplicationRestHandlerImpl{
		argoApplicationService: argoApplicationService,
		logger:                 logger,
	}

}

func (handler *ArgoApplicationRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	clusterIdString := v.Get("clusterIds")
	var clusterIds []int
	if clusterIdString != "" {
		clusterIdSlices := strings.Split(clusterIdString, ",")
		for _, clusterId := range clusterIdSlices {
			id, err := strconv.Atoi(clusterId)
			if err != nil {
				handler.logger.Errorw("error in converting clusterId", "err", err, "clusterIdString", clusterIdString)
				common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
				return
			}
			clusterIds = append(clusterIds, id)
		}
	}
	resp, err := handler.argoApplicationService.ListApplications(clusterIds)
	if err != nil {
		handler.logger.Errorw("error in listing all argo applications", "err", err, "clusterIds", clusterIds)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *ArgoApplicationRestHandlerImpl) GetApplicationDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	v := r.URL.Query()
	resourceName := v.Get("name")
	namespace := v.Get("namespace")
	clusterIdString := v.Get("clusterId")

	var clusterId int
	if clusterIdString != "" {
		clusterId, err = strconv.Atoi(clusterIdString)
		if err != nil {
			handler.logger.Errorw("error in converting clusterId", "err", err, "clusterIdString", clusterIdString)
			common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
			return
		}
	}
	resp, err := handler.argoApplicationService.GetAppDetail(resourceName, namespace, clusterId)
	if err != nil {
		handler.logger.Errorw("error in listing all argo applications", "err", err, "resourceName", resourceName, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
