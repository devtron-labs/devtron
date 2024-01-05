package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type ProxyRouter interface {
	InitProxyRouter(router *mux.Router)
}

type ProxyRouterImpl struct {
	logger *zap.SugaredLogger
	proxy  map[string]func(writer http.ResponseWriter, request *http.Request)
}

func NewProxyRouterImpl(logger *zap.SugaredLogger, proxyCfg *Config, userService user.UserService) *ProxyRouterImpl {
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
		proxy[s] = NewHTTPReverseProxy(fmt.Sprintf("http://%s:%s", connection.Host, connection.Port), client.Transport, userService)
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
}

var ProxyWireSet = wire.NewSet(
	GetProxyConfig,
	NewProxyRouterImpl,
	wire.Bind(new(ProxyRouter), new(*ProxyRouterImpl)),
)

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

func NewHTTPReverseProxy(serverAddr string, transport http.RoundTripper, userService user.UserService) func(writer http.ResponseWriter, request *http.Request) {
	target, err := url.Parse(serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.Director = func(request *http.Request) {
		path := request.URL.Path
		request.URL.Host = target.Host
		request.URL.Scheme = target.Scheme
		request.URL.Path = rewriteRequestUrl(path)
		fmt.Printf("%s\n", request.URL.Path)
	}
	return func(w http.ResponseWriter, r *http.Request) {

		userId, err := userService.GetLoggedInUser(r)
		if userId == 0 || err != nil {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
			return
		}
		isActionUserSuperAdmin, err := userService.IsSuperAdmin(int(userId))
		if err != nil {
			common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
			return
		}
		if isActionUserSuperAdmin {
			proxy.ServeHTTP(w, r)
		}

	}
}

func rewriteRequestUrl(path string) string {
	parts := strings.Split(path, "/")
	var finalParts []string
	for _, part := range parts {
		if part == "proxy" {
			continue
		}
		finalParts = append(finalParts, part)
	}
	return strings.Join(finalParts, "/")
}
