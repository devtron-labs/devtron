package dashboard

import (
	"fmt"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type DashboardRouter interface {
	InitDashboardRouter(router *mux.Router)
}

type DashboardRouterImpl struct {
	logger         *zap.SugaredLogger
	dashboardProxy func(writer http.ResponseWriter, request *http.Request)
}

func NewDashboardRouterImpl(logger *zap.SugaredLogger, dashboardCfg *Config) *DashboardRouterImpl {
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
	dashboardProxy := NewDashboardHTTPReverseProxy(fmt.Sprintf("http://%s:%s", dashboardCfg.Host, dashboardCfg.Port), client.Transport)
	router := &DashboardRouterImpl{
		dashboardProxy: dashboardProxy,
		logger:         logger,
	}
	return router
}

func (router DashboardRouterImpl) InitDashboardRouter(dashboardRouter *mux.Router) {
	dashboardRouter.PathPrefix("").HandlerFunc(router.dashboardProxy)
}

var DashboardWireSet = wire.NewSet(
	GetConfig,
	NewDashboardRouterImpl,
	wire.Bind(new(DashboardRouter), new(*DashboardRouterImpl)),
)
