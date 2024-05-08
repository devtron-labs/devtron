package proxy

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const Dashboard = "dashboard"
const Proxy = "proxy"

func NewDashboardHTTPReverseProxy(serverAddr string, transport http.RoundTripper) func(writer http.ResponseWriter, request *http.Request) {
	proxy := GetProxyServer(serverAddr, transport, Dashboard, "", NewNoopActivityLogger())
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func GetProxyServer(serverAddr string, transport http.RoundTripper, pathToExclude string, basePathToInclude string, activityLogger RequestActivityLogger) *httputil.ReverseProxy {
	proxy := GetProxyServerWithPathTrimFunc(serverAddr, transport, pathToExclude, basePathToInclude, activityLogger, nil)
	return proxy
}

func GetProxyServerWithPathTrimFunc(serverAddr string, transport http.RoundTripper, pathToExclude string, basePathToInclude string, activityLogger RequestActivityLogger, pathTrimFunc func(string) string) *httputil.ReverseProxy {
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
		if pathTrimFunc == nil {
			request.URL.Path = rewriteRequestUrl(basePathToInclude, path, pathToExclude)
		} else {
			request.URL.Path = pathTrimFunc(request.URL.Path)
		}
		activityLogger.LogActivity()
	}
	return proxy
}

type RequestActivityLogger interface {
	LogActivity()
}

func NewNoopActivityLogger() RequestActivityLogger { return NoopActivityLogger{} }

type NoopActivityLogger struct{}

func (logger NoopActivityLogger) LogActivity() {}

func rewriteRequestUrl(basePathToInclude string, path string, pathToExclude string) string {
	parts := strings.Split(path, "/")
	var finalParts []string
	if len(basePathToInclude) > 0 {
		finalParts = append(finalParts, basePathToInclude)
	}
	for _, part := range parts {
		if part == pathToExclude {
			continue
		}
		finalParts = append(finalParts, part)
	}
	return strings.Join(finalParts, "/")
}

func NewProxyTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   120 * time.Second,
			KeepAlive: 120 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func NewHTTPReverseProxy(connection ProxyConnection, transport http.RoundTripper, enforcer casbin.Enforcer) func(writer http.ResponseWriter, request *http.Request) {
	serverAddr := fmt.Sprintf("http://%s:%s", connection.Host, connection.Port)
	proxy := GetProxyServer(serverAddr, transport, Proxy, "", NewNoopActivityLogger())
	return func(w http.ResponseWriter, r *http.Request) {

		if len(connection.PassKey) > 0 {
			r.Header.Add("X-PASS-KEY", connection.PassKey)
		}
		token := r.Header.Get("token")
		if ok := enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, nil, "Unauthorized User", http.StatusForbidden)
			return
		}
		proxy.ServeHTTP(w, r)
	}
}
