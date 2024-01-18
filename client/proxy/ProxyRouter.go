package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type ProxyRouter interface {
	InitProxyRouter(router *mux.Router)
}

type ProxyConnection struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type Config struct {
	ProxyServiceConfig string `env:"PROXY_SERVICE_CONFIG" envDefault:""`
}

func GetProxyConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

type ProxyRouterImpl struct {
	logger *zap.SugaredLogger
	proxy  map[string]func(writer http.ResponseWriter, request *http.Request)
}

func NewProxyRouterImpl(logger *zap.SugaredLogger, proxyCfg *Config, enforcer casbin.Enforcer) *ProxyRouterImpl {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   120 * time.Second,
				KeepAlive: 120 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	proxyConnection := make(map[string]ProxyConnection)
	err := json.Unmarshal([]byte(proxyCfg.ProxyServiceConfig), &proxyConnection)
	if err != nil {
		logger.Warnw("bad env value for PROXY_SERVICE_CONFIG", "err", err)
	}

	proxy := make(map[string]func(writer http.ResponseWriter, request *http.Request))
	for s, connection := range proxyConnection {
		proxy[s] = NewHTTPReverseProxy(fmt.Sprintf("http://%s:%s", connection.Host, connection.Port), client.Transport, enforcer)
	}

	router := &ProxyRouterImpl{
		proxy:  proxy,
		logger: logger,
	}
	return router
}

func (router ProxyRouterImpl) InitProxyRouter(ProxyRouter *mux.Router) {

	ProxyRouter.PathPrefix("/kubelink").HandlerFunc(router.proxy["kubelink"])
	ProxyRouter.PathPrefix("/gitsensor").HandlerFunc(router.proxy["gitsensor"])
	ProxyRouter.PathPrefix("/kubewatch").HandlerFunc(router.proxy["kubewatch"])
	ProxyRouter.PathPrefix("/image-scanner").HandlerFunc(router.proxy["image-scanner"])
}
