package gocd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	// Version of the gocd library in the event that we change it for the user agent.
	libraryVersion = "1"
	// UserAgent to be used when calling the GoCD agent.
	userAgent = "go-gocd/" + libraryVersion
	// For the unversionned API
	apiV0 = ""
	// Version 1 of the GoCD API.
	apiV1 = "application/vnd.go.cd.v1+json"
	// Version 2 of the GoCD API.
	apiV2 = "application/vnd.go.cd.v2+json"
	// Version 3 of the GoCD API.
	apiV3 = "application/vnd.go.cd.v3+json"
	// Version 4 of the GoCD API.
	apiV4 = "application/vnd.go.cd.v4+json"
	// Version 5 of the GoCD API.
	apiV5 = "application/vnd.go.cd.v5+json"
	// Version 6 of the GoCD API.
	apiV6 = "application/vnd.go.cd.v6+json"
	// Version 7 of the GoCD API.
	apiV7 = "application/vnd.go.cd.v7+json"
	// Version 8 of the GoCD API.
	apiV8 = "application/vnd.go.cd.v8+json"
	// Version 9 of the GoCD API.
	apiV9 = "application/vnd.go.cd.v9+json"
	// Version 10 of the GoCD API.
	apiV10 = "application/vnd.go.cd.v10+json"
)

//Body Response Types
const (
	responseTypeXML  = "xml"
	responseTypeJSON = "json"
	responseTypeText = "text"
)

//Logging Environment variables
const (
	gocdLogLevel = "GOCD_LOG"
)

// StringResponse handles the unmarshaling of the single string response from DELETE requests.
type StringResponse struct {
	Message string `json:"message"`
}

// APIResponse encapsulates the net/http.Response object, a string representing the Body, and a gocd.Request object
// encapsulating the response from the API.
type APIResponse struct {
	HTTP    *http.Response
	Body    string
	Request *APIRequest
}

// APIRequest encapsulates the net/http.Request object, and a string representing the Body.
type APIRequest struct {
	HTTP *http.Request
	Body string
}

// Client struct which acts as an interface to the GoCD Server. Exposes resource service handlers.
type Client struct {
	clientMu sync.Mutex // clientMu protects the client during multi-threaded calls
	client   *http.Client

	params *ClientParameters

	Log *logrus.Logger

	Agents            *AgentsService
	PipelineGroups    *PipelineGroupsService
	Stages            *StagesService
	Jobs              *JobsService
	PipelineTemplates *PipelineTemplatesService
	Pipelines         *PipelinesService
	PipelineConfigs   *PipelineConfigsService
	Configuration     *ConfigurationService
	ConfigRepos       *ConfigRepoService
	Encryption        *EncryptionService
	Plugins           *PluginsService
	Environments      *EnvironmentsService
	Properties        *PropertiesService
	Roles             *RoleService
	ServerVersion     *ServerVersionService

	common service
	cookie string
}

// ClientParameters describe how the client interacts with the GoCD Server
type ClientParameters struct {
	BaseURL  *url.URL
	Username string
	Password string

	UserAgent string
}

// BuildPath creates an absolute URL from ClientParameters and a relative URL
func (cp *ClientParameters) BuildPath(rel *url.URL) *url.URL {
	u := cp.BaseURL.ResolveReference(rel)
	if cp.BaseURL.RawQuery != "" {
		u.RawQuery = cp.BaseURL.RawQuery
	}
	return u
}

// PaginationResponse is a struct used to handle paging through resposnes.
type PaginationResponse struct {
	Offset   int `json:"offset"`
	Total    int `json:"total"`
	PageSize int `json:"page_size"`
}

// service is a generic service encapsulating the client for talking to the GoCD server.
type service struct {
	client *Client
	log    *logrus.Logger
}

// Auth structure wrapping the Username and Password variables, which are used to get an Auth cookie header used for
// subsequent requests.
type Auth struct {
	Username string
	Password string
}

// HasAuth checks whether or not we have the required Username/Password variables provided.
func (c *Configuration) HasAuth() bool {
	return (c.Username != "") && (c.Password != "")
}

// Client returns a client which allows us to interact with the GoCD Server.
func (c *Configuration) Client() *Client {
	return NewClient(c, nil)
}

// NewClient creates a new client based on the provided configuration payload, and optionally a custom httpClient to
// allow overriding of http client structures.
func NewClient(cfg *Configuration, httpClient *http.Client) *Client {

	httpClient = generateHTTPClient(cfg, httpClient)

	baseURL, _ := url.Parse(cfg.Server)

	c := &Client{
		client: httpClient,
		params: &ClientParameters{
			BaseURL:   baseURL,
			UserAgent: userAgent,
			Username:  cfg.Username,
			Password:  cfg.Password,
		},
		Log: logrus.New(),
	}

	c.common.client = c
	c.common.log = c.Log

	attachServices(c)

	SetupLogging(c.Log)

	return c
}

// generateHTTPClient taking into account ssl, and existing httpClient
func generateHTTPClient(cfg *Configuration, httpClient *http.Client) *http.Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
		if strings.HasPrefix(cfg.Server, "https") && cfg.SkipSslCheck {
			httpClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.SkipSslCheck},
			}
		}
	}
	return httpClient
}

// attachServices to the client to give access to the difference API resources.
// codebeat:disable[ABC]
func attachServices(c *Client) {
	c.Agents = (*AgentsService)(&c.common)
	c.PipelineGroups = (*PipelineGroupsService)(&c.common)
	c.Stages = (*StagesService)(&c.common)
	c.Jobs = (*JobsService)(&c.common)
	c.PipelineTemplates = (*PipelineTemplatesService)(&c.common)
	c.Pipelines = (*PipelinesService)(&c.common)
	c.PipelineConfigs = (*PipelineConfigsService)(&c.common)
	c.Configuration = (*ConfigurationService)(&c.common)
	c.ConfigRepos = (*ConfigRepoService)(&c.common)
	c.Encryption = (*EncryptionService)(&c.common)
	c.Plugins = (*PluginsService)(&c.common)
	c.Environments = (*EnvironmentsService)(&c.common)
	c.Properties = (*PropertiesService)(&c.common)
	c.Roles = (*RoleService)(&c.common)
	c.ServerVersion = (*ServerVersionService)(&c.common)
}

// codebeat:enable[ABC]

// BaseURL creates a URL from the ClientParameters BaseURL
func (c *Client) BaseURL() *url.URL {
	return c.params.BaseURL
}

// Lock the client until release
func (c *Client) Lock() {
	c.clientMu.Lock()
}

// Unlock the client after a lock action
func (c *Client) Unlock() {
	c.clientMu.Unlock()
}

// NewRequest creates an HTTP requests to the GoCD API endpoints.
func (c *Client) NewRequest(method, urlStr string, body interface{}, apiVersion string) (req *APIRequest, err error) {
	var rel *url.URL
	var buf io.ReadWriter
	req = &APIRequest{}

	// I'm not sure how to get this method to return an error intentionally for testing. For testing purposes, I've
	// added a switch so that the error handling in dependent methods can be tested.
	if os.Getenv("GOCD_RAISE_ERROR_NEW_REQUEST") == "yes" {
		return req, errors.New("Mock Testing Error")
	}

	// Some calls
	if strings.HasPrefix(urlStr, "/") {
		urlStr = urlStr[1:]
	} else {
		urlStr = "api/" + urlStr
	}
	if rel, err = url.Parse(urlStr); err != nil {
		return req, err
	}

	u := c.params.BuildPath(rel)

	if body != nil {
		buf = new(bytes.Buffer)

		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		err := enc.Encode(body)

		if err != nil {
			return nil, err
		}
		bdy, _ := ioutil.ReadAll(buf)
		req.Body = string(bdy)

		buf = new(bytes.Buffer)
		enc = json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		enc.Encode(body)
	}

	if req.HTTP, err = http.NewRequest(method, u.String(), buf); err != nil {
		return req, err
	}

	if body != nil {
		req.HTTP.Header.Set("Content-Type", "application/json")
	}
	if apiVersion != "" {
		req.HTTP.Header.Set("Accept", apiVersion)
	}
	req.HTTP.Header.Set("User-Agent", c.params.UserAgent)

	if c.cookie == "" {
		if c.params.Username != "" && c.params.Password != "" {
			req.HTTP.SetBasicAuth(c.params.Username, c.params.Password)
		}
	} else {
		req.HTTP.Header.Set("Cookie", c.cookie)
	}

	return
}

// Do takes an HTTP request and resposne the response from the GoCD API endpoint.
func (c *Client) Do(ctx context.Context, req *APIRequest, v interface{}, responseType string) (*APIResponse, error) {
	var err error
	var resp *http.Response

	req.HTTP = req.HTTP.WithContext(ctx)

	if resp, err = c.client.Do(req.HTTP); err != nil {
		return nil, err
	}

	r := &APIResponse{
		Request: req,
		HTTP:    resp,
	}

	if v != nil {
		if r.Body, err = readDoResponseBody(v, &r.HTTP.Body, responseType); err != nil {
			return nil, err
		}
	}

	if err = CheckResponse(r); err != nil {
		return r, err
	}

	return r, err
}

// getAPIVersion is a wrapper around ServerVersion.GetAPIVersion that starts by making sure ServerVersionService.Get has
// been called. Note that it also adds the /api/ in front of the provided endpoint
func (c *Client) getAPIVersion(ctx context.Context, endpoint string) (apiVersion string, err error) {
	v, _, err := c.ServerVersion.Get(ctx)
	if err != nil {
		return "", err
	}
	return v.GetAPIVersion(fmt.Sprintf("/api/%s", endpoint))
}

func readDoResponseBody(v interface{}, bodyReader *io.ReadCloser, responseType string) (body string, err error) {
	var bodyBytes []byte

	if w, ok := v.(io.Writer); ok {
		_, err := io.Copy(w, *bodyReader)
		return "", err
	}

	bodyBytes, err = ioutil.ReadAll(*bodyReader)
	if responseType == responseTypeText {
		body = string(bodyBytes)
		v = &body
	} else if responseType == responseTypeXML {
		err = xml.Unmarshal(bodyBytes, v)
	} else {
		err = json.Unmarshal(bodyBytes, v)
	}

	body = string(bodyBytes)

	if err == io.EOF {
		err = nil // ignore EOF errors caused by empty response body
	}
	return

}

// CheckResponse asserts that the http response status code was 2xx.
func CheckResponse(response *APIResponse) (err error) {
	if response.HTTP.StatusCode < 200 || response.HTTP.StatusCode >= 400 {

		errorParts := []string{
			fmt.Sprintf("Received HTTP Status '%s'", response.HTTP.Status),
		}
		if message := createErrorResponseMessage(response.Body); message != "" {
			errorParts = append(errorParts, message)
		}

		err = errors.New(strings.Join(errorParts, ": "))
	}
	return
}

func createErrorResponseMessage(body string) (resp string) {
	reqBody := make(map[string]interface{})
	resBody := make(map[string]interface{})

	json.Unmarshal([]byte(body), &reqBody)

	if message, hasMessage := reqBody["message"]; hasMessage {
		resBody["message"] = message
	}

	if data, hasData := reqBody["data"]; hasData {
		if data, isData := data.(map[string]interface{}); isData {
			if err, hasErrors := data["errors"]; hasErrors {
				resBody["errors"] = err
			}
		}
	}

	if len(resBody) > 0 {
		b, _ := json.MarshalIndent(resBody, "", "  ")
		resp = string(b)
	}

	return

}

// String returns a pointer to the string value passed in. Allows `omitempty` to function in json building
func String(v string) *string {
	return &v
}

// Int returns a pointer to the int value passed in. Allows `omitempty` to function in json building
func Int(v int) *int {
	return &v
}
