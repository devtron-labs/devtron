package scoop

import (
	"encoding/json"
	util4 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"strconv"
)

type RestHandler interface {
	HandleNotificationEvent(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	service Service
}

func NewRestHandler(service Service) *RestHandlerImpl {
	return &RestHandlerImpl{
		service: service,
	}
}

func (handler *RestHandlerImpl) HandleNotificationEvent(w http.ResponseWriter, r *http.Request) {
	// token := r.Header.Get("token")
	queryValues := r.URL.Query()
	clusterIdString := queryValues.Get("clusterId")
	if len(clusterIdString) == 0 {
		common.WriteJsonResp(w, errors.New("clusterid not present"), nil, http.StatusBadRequest)
		return
	}
	clusterId, err := strconv.Atoi(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := queryValues.Get("name")
	namespace := queryValues.Get("namespace")
	group := queryValues.Get("group")
	version := queryValues.Get("version")
	kind := queryValues.Get("kind")
	request := k8s.ResourceRequestBean{
		ClusterId: clusterId,
		K8sRequest: &util4.K8sRequestBean{
			ResourceIdentifier: util4.ResourceIdentifier{
				Name:      resourceName,
				Namespace: namespace,
				GroupVersionKind: schema.GroupVersionKind{
					Group:   group,
					Version: version,
					Kind:    kind,
				},
			},
		},
	}
	// if ok := handler.handleRbac(r, w, request, token, casbin.ActionUpdate); !ok {
	// 	common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
	// 	return
	// }
	decoder := json.NewDecoder(r.Body)
	var notification map[string]interface{}
	err = decoder.Decode(&notification)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	err = handler.service.HandleNotificationEvent(r.Context(), &request, notification)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
