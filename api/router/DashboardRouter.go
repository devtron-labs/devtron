package router

import (
	"fmt"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type DashboardRouter interface {
	initDashboardRouter(router *mux.Router)
}

type DashboardRouterImpl struct {
	logger         *zap.SugaredLogger
	dashboardProxy func(writer http.ResponseWriter, request *http.Request)
}

func NewDashboardRouterImpl(logger *zap.SugaredLogger, dashboardCfg *dashboard.Config) *DashboardRouterImpl {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	dashboardProxy := dashboard.NewDashboardHTTPReverseProxy(fmt.Sprintf("http://%s:%s", dashboardCfg.Host, dashboardCfg.Port), client.Transport)
	router := &DashboardRouterImpl{
		dashboardProxy: dashboardProxy,
		logger:         logger,
	}
	return router
}

func (router DashboardRouterImpl) initDashboardRouter(dashboardRouter *mux.Router) {
	dashboardRouter.PathPrefix("").HandlerFunc(router.dashboardProxy)
}