package grafana

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewGrafanaHTTPReverseProxy(serverAddr string, transport http.RoundTripper) (func(writer http.ResponseWriter, request *http.Request), error) {
	target, err := url.Parse(serverAddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport

	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}, nil
}
