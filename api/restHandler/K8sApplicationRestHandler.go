package restHandler

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	//vars := mux.Vars(r)
	//name := vars["name"]
	v := r.URL.Query()
	nameSpace := v.Get("namespace")
	version := v.Get("version")
	group := v.Get("group")
	kind := v.Get("kind")
	resourceName := v.Get("resourceName")
	token := r.Header.Get("token")
	request := &application.GetRequest{
		Name:         resourceName,
		Namespace: nameSpace,
		GroupVersionKind: schema.GroupVersionKind{
			Kind: kind,
			Group: group,
			Version: version,
		},
	}
	resource, err := impl.k8sApplication.GetResource(token, request)
	if err!=nil{
		common.WriteJsonResp(w,err,resource,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,resource,http.StatusOK)
	return
}

func (impl K8sApplicationRestHandlerImpl) UpdateResource(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	nameSpace := v.Get("namespace")
	version := v.Get("version")
	group := v.Get("group")
	kind := v.Get("kind")
	resourceName := v.Get("resourceName")
	token := r.Header.Get("token")
	//TODO : confirm patch & patchType placement(header/url param)
	request := &application.UpdateRequest{
		Name:         resourceName,
		Namespace: nameSpace,
		GroupVersionKind: schema.GroupVersionKind{
			Kind: kind,
			Group: group,
			Version: version,
		},
	}
	resource, err := impl.k8sApplication.UpdateResource(token, request)
	if err!=nil{
		common.WriteJsonResp(w,err,resource,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,resource,http.StatusOK)
	return
}

func (impl K8sApplicationRestHandlerImpl) DeleteResource(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	nameSpace := v.Get("namespace")
	version := v.Get("version")
	group := v.Get("group")
	kind := v.Get("kind")
	resourceName := v.Get("resourceName")
	token := r.Header.Get("token")
	//TODO : confirm force bool value
	request := &application.DeleteRequest{
		Name:         resourceName,
		Namespace: nameSpace,
		GroupVersionKind: schema.GroupVersionKind{
			Kind: kind,
			Group: group,
			Version: version,
		},
	}
	resource, err := impl.k8sApplication.DeleteResource(token, request)
	if err!=nil{
		common.WriteJsonResp(w,err,resource,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,resource,http.StatusOK)
	return
}

func (impl K8sApplicationRestHandlerImpl) ListEvents(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	nameSpace := v.Get("namespace")
	version := v.Get("version")
	group := v.Get("group")
	kind := v.Get("kind")
	resourceName := v.Get("resourceName")
	token := r.Header.Get("token")
	request := &application.GetRequest{
		Name:         resourceName,
		Namespace: nameSpace,
		GroupVersionKind: schema.GroupVersionKind{
			Kind: kind,
			Group: group,
			Version: version,
		},
	}
	events, err := impl.k8sApplication.ListEvents(token, request)
	if err!=nil{
		common.WriteJsonResp(w,err,events,http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w,nil,events,http.StatusOK)
	return
}
