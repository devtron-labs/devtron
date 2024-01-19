package proxy

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const Dashboard = "dashboard"
const Proxy = "proxy"

func NewDashboardHTTPReverseProxy(serverAddr string, transport http.RoundTripper) func(writer http.ResponseWriter, request *http.Request) {
	proxy := GetProxyServer(serverAddr, transport, Dashboard)
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func GetProxyServer(serverAddr string, transport http.RoundTripper, pathToExclude string) *httputil.ReverseProxy {
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
		request.URL.Path = rewriteRequestUrl(path, pathToExclude)
		fmt.Printf("%s\n", request.URL.Path)
	}
	return proxy
}

func rewriteRequestUrl(path string, pathToExclude string) string {
	parts := strings.Split(path, "/")
	var finalParts []string
	for _, part := range parts {
		if part == pathToExclude {
			continue
		}
		finalParts = append(finalParts, part)
	}
	return strings.Join(finalParts, "/")
}

func NewHTTPReverseProxy(serverAddr string, transport http.RoundTripper, enforcer casbin.Enforcer) func(writer http.ResponseWriter, request *http.Request) {
	proxy := GetProxyServer(serverAddr, transport, Proxy)
	return func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("token")
		if ok := enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, nil, "Unauthorized User", http.StatusForbidden)
			return
		}
		proxy.ServeHTTP(w, r)
	}
}
