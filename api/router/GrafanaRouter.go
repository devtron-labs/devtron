package router

import (
	"fmt"
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type GrafanaRouter interface {
	initGrafanaRouter(router *mux.Router)
}

type GrafanaRouterImpl struct {
	logger       *zap.SugaredLogger
	grafanaProxy func(writer http.ResponseWriter, request *http.Request)
}

func NewGrafanaRouterImpl(logger *zap.SugaredLogger, grafanaCfg *grafana.Config) (*GrafanaRouterImpl, error) {
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
	grafanaProxy, err := grafana.NewGrafanaHTTPReverseProxy(fmt.Sprintf("http://%s:%s", grafanaCfg.Host, grafanaCfg.Port), client.Transport)
	if err != nil {
		return nil, err
	}
	router := &GrafanaRouterImpl{
		grafanaProxy: grafanaProxy,
		logger:       logger,
	}
	return router, nil
}

func (router GrafanaRouterImpl) initGrafanaRouter(grafanaRouter *mux.Router) {
	grafanaRouter.PathPrefix("").HandlerFunc(router.grafanaProxy)
}
