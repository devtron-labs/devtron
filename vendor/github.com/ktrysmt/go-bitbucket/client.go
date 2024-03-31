package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
	"golang.org/x/oauth2/clientcredentials"
)

const DEFAULT_PAGE_LENGTH = 10
const DEFAULT_LIMIT_PAGES = 0
const DEFAULT_MAX_DEPTH = 1
const DEFAULT_BITBUCKET_API_BASE_URL = "https://api.bitbucket.org/2.0"

func apiBaseUrl() (*url.URL, error) {
	ev := os.Getenv("BITBUCKET_API_BASE_URL")
	if ev == "" {
		ev = DEFAULT_BITBUCKET_API_BASE_URL
	}

	return url.Parse(ev)
}

type Client struct {
	Auth         *auth
	Users        users
	User         user
	Teams        teams
	Repositories *Repositories
	Workspaces   *Workspace
	Pagelen      int
	MaxDepth     int
	// LimitPages limits the number of pages for a request
	//	default value as 0 -- disable limits
	LimitPages int
	// DisableAutoPaging allows you to disable the default behavior of automatically requesting
	// all the pages for a paginated response.
	DisableAutoPaging bool
	apiBaseURL        *url.URL

	HttpClient *http.Client
}

type auth struct {
	appID, secret  string
	user, password string
	token          oauth2.Token
	bearerToken    string
}

type Response struct {
	Size     int           `json:"size"`
	Page     int           `json:"page"`
	Pagelen  int           `json:"pagelen"`
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Values   []interface{} `json:"values"`
}

// Uses the Client Credentials Grant oauth2 flow to authenticate to Bitbucket
func NewOAuthClientCredentials(i, s string) *Client {
	a := &auth{appID: i, secret: s}
	ctx := context.Background()
	conf := &clientcredentials.Config{
		ClientID:     i,
		ClientSecret: s,
		TokenURL:     bitbucket.Endpoint.TokenURL,
	}

	tok, err := conf.Token(ctx)
	if err != nil {
		log.Fatal(err)
	}
	a.token = *tok
	return injectClient(a)

}

func NewOAuth(i, s string) *Client {
	a := &auth{appID: i, secret: s}
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     i,
		ClientSecret: s,
		Endpoint:     bitbucket.Endpoint,
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog:\n%v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	fmt.Printf("Enter the code in the return URL: ")
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}
	a.token = *tok
	return injectClient(a)
}

// NewOAuthWithCode finishes the OAuth handshake with a given code
// and returns a *Client
func NewOAuthWithCode(i, s, c string) (*Client, string) {
	a := &auth{appID: i, secret: s}
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     i,
		ClientSecret: s,
		Endpoint:     bitbucket.Endpoint,
	}

	tok, err := conf.Exchange(ctx, c)
	if err != nil {
		log.Fatal(err)
	}
	a.token = *tok
	return injectClient(a), tok.AccessToken
}

// NewOAuthWithRefreshToken obtains a new access token with a given refresh token
// and returns a *Client
func NewOAuthWithRefreshToken(i, s, rt string) (*Client, string) {
	a := &auth{appID: i, secret: s}
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     i,
		ClientSecret: s,
		Endpoint:     bitbucket.Endpoint,
	}

	tokenSource := conf.TokenSource(ctx, &oauth2.Token{
		RefreshToken: rt,
	})
	tok, err := tokenSource.Token()
	if err != nil {
		log.Fatal(err)
	}
	a.token = *tok
	return injectClient(a), tok.AccessToken
}

func NewOAuthbearerToken(t string) *Client {
	a := &auth{bearerToken: t}
	return injectClient(a)
}

func NewBasicAuth(u, p string) *Client {
	a := &auth{user: u, password: p}
	return injectClient(a)
}

func injectClient(a *auth) *Client {
	bitbucketUrl, err := apiBaseUrl()
	if err != nil {
		log.Fatalf("invalid bitbucket url")
	}
	c := &Client{Auth: a, Pagelen: DEFAULT_PAGE_LENGTH, MaxDepth: DEFAULT_MAX_DEPTH,
		apiBaseURL: bitbucketUrl, LimitPages: DEFAULT_LIMIT_PAGES}
	c.Repositories = &Repositories{
		c:                  c,
		PullRequests:       &PullRequests{c: c},
		Pipelines:          &Pipelines{c: c},
		Repository:         &Repository{c: c},
		Issues:             &Issues{c: c},
		Commits:            &Commits{c: c},
		Diff:               &Diff{c: c},
		BranchRestrictions: &BranchRestrictions{c: c},
		Webhooks:           &Webhooks{c: c},
		Downloads:          &Downloads{c: c},
		DeployKeys:         &DeployKeys{c: c},
	}
	c.Users = &Users{c: c}
	c.User = &User{c: c}
	c.Teams = &Teams{c: c}
	c.Workspaces = &Workspace{c: c, Repositories: c.Repositories, Permissions: &Permission{c: c}}
	c.HttpClient = new(http.Client)
	return c
}

func (c *Client) GetOAuthToken() oauth2.Token {
	return c.Auth.token
}

func (c *Client) GetApiBaseURL() string {
	return fmt.Sprintf("%s%s", c.GetApiHostnameURL(), c.apiBaseURL.Path)
}

func (c *Client) GetApiHostnameURL() string {
	return fmt.Sprintf("%s://%s", c.apiBaseURL.Scheme, c.apiBaseURL.Host)
}

func (c *Client) SetApiBaseURL(urlStr url.URL) {
	c.apiBaseURL = &urlStr
}

func (c *Client) executeRaw(method string, urlStr string, text string) (io.ReadCloser, error) {
	body := strings.NewReader(text)

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if text != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	c.authenticateRequest(req)
	return c.doRawRequest(req, false)
}

func (c *Client) execute(method string, urlStr string, text string) (interface{}, error) {
	body := strings.NewReader(text)
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if text != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	c.authenticateRequest(req)
	result, err := c.doRequest(req, false)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) executePaginated(method string, urlStr string, text string, page *int) (interface{}, error) {
	if c.Pagelen != DEFAULT_PAGE_LENGTH {
		urlObj, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		q := urlObj.Query()
		q.Set("pagelen", strconv.Itoa(c.Pagelen))
		urlObj.RawQuery = q.Encode()
		urlStr = urlObj.String()
	}

	body := strings.NewReader(text)
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if text != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	c.authenticateRequest(req)
	result, err := c.doPaginatedRequest(req, page, false)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) executeFileUpload(method string, urlStr string, filePath string, fileName string, fieldname string, params map[string]string) (interface{}, error) {
	fileReader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var fw io.Writer
	if fw, err = w.CreateFormFile(fieldname, fileName); err != nil {
		return nil, err
	}

	if _, err = io.Copy(fw, fileReader); err != nil {
		return nil, err
	}

	for key, value := range params {
		err = w.WriteField(key, value)
		if err != nil {
			return nil, err
		}
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest(method, urlStr, &b)
	if err != nil {
		return nil, err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	c.authenticateRequest(req)
	return c.doRequest(req, true)

}

func (c *Client) authenticateRequest(req *http.Request) {
	if c.Auth.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Auth.bearerToken)
	}

	if c.Auth.user != "" && c.Auth.password != "" {
		req.SetBasicAuth(c.Auth.user, c.Auth.password)
	} else if c.Auth.token.Valid() {
		c.Auth.token.SetAuthHeader(req)
	}
}

func (c *Client) doRequest(req *http.Request, emptyResponse bool) (interface{}, error) {
	resBody, err := c.doRawRequest(req, emptyResponse)
	if err != nil {
		return nil, err
	}
	if emptyResponse || resBody == nil {
		return nil, nil
	}

	defer resBody.Close()

	responseBytes, err := ioutil.ReadAll(resBody)
	if err != nil {
		return resBody, err
	}

	var result interface{}
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		return responseBytes, err
	}
	return result, nil
}

func (c *Client) doPaginatedRequest(req *http.Request, page *int, emptyResponse bool) (interface{}, error) {
	disableAutoPaging := c.DisableAutoPaging
	curPage := 1
	if page != nil {
		disableAutoPaging = true
		curPage = *page
		q := req.URL.Query()
		q.Set("page", strconv.Itoa(curPage))
		req.URL.RawQuery = q.Encode()
	}
	// q.Encode() does not encode "~".
	req.URL.RawQuery = strings.ReplaceAll(req.URL.RawQuery, "~", "%7E")

	resBody, err := c.doRawRequest(req, emptyResponse)
	if err != nil {
		return nil, err
	}
	if emptyResponse || resBody == nil {
		return nil, nil
	}

	defer resBody.Close()

	responseBytes, err := ioutil.ReadAll(resBody)
	if err != nil {
		return resBody, err
	}

	responsePaginated := &Response{}
	err = json.Unmarshal(responseBytes, responsePaginated)
	if err == nil && len(responsePaginated.Values) > 0 {
		values := responsePaginated.Values
		for {
			if disableAutoPaging || responsePaginated.Next == "" ||
				(curPage >= c.LimitPages && c.LimitPages != 0) {
				break
			}
			curPage++
			newReq, err := http.NewRequest(req.Method, responsePaginated.Next, nil)
			if err != nil {
				return resBody, err
			}
			c.authenticateRequest(newReq)
			resp, err := c.doRawRequest(newReq, false)
			if err != nil {
				return resBody, err
			}

			responsePaginated = &Response{}
			json.NewDecoder(resp).Decode(responsePaginated)
			values = append(values, responsePaginated.Values...)
		}
		responsePaginated.Values = values
		responseBytes, err = json.Marshal(responsePaginated)
		if err != nil {
			return resBody, err
		}
	}

	var result interface{}
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		return resBody, err
	}
	return result, nil
}

func (c *Client) doRawRequest(req *http.Request, emptyResponse bool) (io.ReadCloser, error) {
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if unexpectedHttpStatusCode(resp.StatusCode) {
		defer resp.Body.Close()

		out := &UnexpectedResponseStatusError{Status: resp.Status}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			out.Body = []byte(fmt.Sprintf("could not read the response body: %v", err))
		} else {
			out.Body = body
		}

		return nil, out
	}

	if emptyResponse || resp.StatusCode == http.StatusNoContent {
		resp.Body.Close()
		return nil, nil
	}

	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	return resp.Body, nil
}

func unexpectedHttpStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusOK,
		http.StatusCreated,
		http.StatusNoContent,
		http.StatusAccepted:
		return false
	default:
		return true
	}
}

func (c *Client) requestUrl(template string, args ...interface{}) string {

	if len(args) == 1 && args[0] == "" {
		return c.GetApiBaseURL() + template
	}
	return c.GetApiBaseURL() + fmt.Sprintf(template, args...)
}

func (c *Client) addMaxDepthParam(params *url.Values, customMaxDepth *int) {
	maxDepth := c.MaxDepth
	if customMaxDepth != nil && *customMaxDepth > 0 {
		maxDepth = *customMaxDepth
	}

	if maxDepth != DEFAULT_MAX_DEPTH {
		params.Set("max_depth", strconv.Itoa(maxDepth))
	}
}
