package dashboard

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//var messageRe = regexp.MustCompile(`<p>(.*)([\s\S]*?)<\/p>`)

func NewDashboardHTTPReverseProxy(serverAddr string, transport http.RoundTripper) func(writer http.ResponseWriter, request *http.Request) {
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
	//proxy.Director = func(request *http.Request) {
	//	path := request.URL.Path
	//	request.URL.Path = rewriteRequestUrl(path)
	//}
	//proxy.ModifyResponse = func(resp *http.Response) error {
	//	if resp.StatusCode == 500 {
	//		b, err := ioutil.ReadAll(resp.Body)
	//		if err != nil {
	//			return err
	//		}
	//		err = resp.Body.Close()
	//		if err != nil {
	//			return err
	//		}
	//		var message string
	//		matches := messageRe.FindSubmatch(b)
	//		if len(matches) > 1 {
	//			message = html.UnescapeString(string(matches[1]))
	//		} else {
	//			message = "Unknown error"
	//		}
	//		resp.ContentLength = 0
	//		resp.Header.Set("Content-Length", strconv.Itoa(0))
	//		resp.Header.Set(Location, fmt.Sprintf("/dashboard/login?sso_error=%s", url.QueryEscape(message)))
	//		resp.StatusCode = http.StatusSeeOther
	//		resp.Body = ioutil.NopCloser(bytes.NewReader(make([]byte, 0)))
	//		return nil
	//	}
	//	return nil
	//}
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func rewriteRequestUrl(path string) string {
	parts := strings.Split(path, "/")
	var finalParts []string
	for _, part := range parts {
		if part == "dashboard" {
			continue
		}
		finalParts = append(finalParts, part)
	}
	return strings.Join(finalParts, "/")
}
