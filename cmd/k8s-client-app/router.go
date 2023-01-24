package main

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/terminal"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/pkg/attributes"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	Router                   *mux.Router
	logger                   *zap.SugaredLogger
	clusterRouter            cluster.ClusterRouter
	dashboardRouter          dashboard.DashboardRouter
	k8sApplicationRouter     k8s.K8sApplicationRouter
	k8sCapacityRouter        k8s.K8sCapacityRouter
	userTerminalAccessRouter terminal.UserTerminalAccessRouter
}

func NewMuxRouter(
	logger *zap.SugaredLogger,
	clusterRouter cluster.ClusterRouter,
	dashboardRouter dashboard.DashboardRouter,
	k8sApplicationRouter k8s.K8sApplicationRouter,
	k8sCapacityRouter k8s.K8sCapacityRouter,
	userTerminalAccessRouter terminal.UserTerminalAccessRouter,
	kubeConfigFileSyncerImpl *cluster2.KubeConfigFileSyncerImpl,
) *MuxRouter {
	r := &MuxRouter{
		Router:                   mux.NewRouter(),
		logger:                   logger,
		clusterRouter:            clusterRouter,
		dashboardRouter:          dashboardRouter,
		k8sApplicationRouter:     k8sApplicationRouter,
		k8sCapacityRouter:        k8sCapacityRouter,
		userTerminalAccessRouter: userTerminalAccessRouter,
	}
	return r
}
func (r *MuxRouter) Init() {
	r.Router.StrictSlash(true)
	r.Router.Path("/health").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})
	baseRouter := r.Router.PathPrefix("/orchestrator/").Subrouter()
	baseRouter.Path("/version").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = util.GetDevtronVersion()
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})

	clusterRouter := baseRouter.PathPrefix("/cluster").Subrouter()
	r.clusterRouter.InitClusterRouter(clusterRouter)

	r.Router.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/dashboard", 301)
	})

	k8sApp := r.Router.PathPrefix("/orchestrator/k8s").Subrouter()
	r.k8sApplicationRouter.InitK8sApplicationRouter(k8sApp)

	k8sCapacityApp := r.Router.PathPrefix("/orchestrator/k8s/capacity").Subrouter()
	r.k8sCapacityRouter.InitK8sCapacityRouter(k8sCapacityApp)

	userTerminalAccessRouter := r.Router.PathPrefix("/orchestrator/user/terminal").Subrouter()
	r.userTerminalAccessRouter.InitTerminalAccessRouter(userTerminalAccessRouter)

	r.Router.PathPrefix("/orchestrator/attributes").Queries("key", "{key}").
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			vars := mux.Vars(request)
			key := vars["key"]
			if key == "DEFAULT_TERMINAL_IMAGE_LIST" {
				defaultAttrDto := &attributes.AttributesDto{
					Active: true,
					Key:    "DEFAULT_TERMINAL_IMAGE_LIST",
					Value:  "[{\"groupId\":\"latest\",\"groupRegex\":\"v1\\\\.2[4-8]\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:latest\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:latest\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:latest\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]},{\"groupId\":\"v1.22\",\"groupRegex\":\"v1\\\\.(21|22|23)\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:1.22\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:1.22\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:1.22\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]},{\"groupId\":\"v1.19\",\"groupRegex\":\"v1\\\\.(18|19|20)\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:1.19\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:1.19\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:1.19\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]},{\"groupId\":\"v1.16\",\"groupRegex\":\"v1\\\\.(15|16|17)\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:1.16\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:1.16\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:1.16\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]}]",
				}
				common.WriteJsonResp(writer, nil, defaultAttrDto, http.StatusOK)
			}
		}).Methods("GET")
}
