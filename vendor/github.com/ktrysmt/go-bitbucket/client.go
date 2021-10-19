package bitbucket

import (
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bytes"
	"io"
	"mime/multipart"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
	"golang.org/x/oauth2/clientcredentials"
)

const DEFAULT_PAGE_LENGTH = 10
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
	Pagelen      uint64
	MaxDepth     uint64
	apiBaseURL   *url.URL

	HttpClient *http.Client
}

type auth struct {
	appID, secret  string
	user, password string
	token          oauth2.Token
	bearerToken    string
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
	c := &Client{Auth: a, Pagelen: DEFAULT_PAGE_LENGTH, MaxDepth: DEFAULT_MAX_DEPTH, apiBaseURL: bitbucketUrl}
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
	}
	c.Users = &Users{c: c}
	c.User = &User{c: c}
	c.Teams = &Teams{c: c}
	c.Workspaces = &Workspace{c: c, Repositories: c.Repositories, Permissions: &Permission{c: c}}
	c.HttpClient = new(http.Client)
	return c
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
	// Use pagination if changed from default value
	const DEC_RADIX = 10
	if strings.Contains(urlStr, "/repositories/") {
		if c.Pagelen != DEFAULT_PAGE_LENGTH {
			urlObj, err := url.Parse(urlStr)
			if err != nil {
				return nil, err
			}
			q := urlObj.Query()
			q.Set("pagelen", strconv.FormatUint(c.Pagelen, DEC_RADIX))
			urlObj.RawQuery = q.Encode()
			urlStr = urlObj.String()
		}

		if c.MaxDepth != DEFAULT_MAX_DEPTH {
			urlObj, err := url.Parse(urlStr)
			if err != nil {
				return nil, err
			}
			q := urlObj.Query()
			q.Set("max_depth", strconv.FormatUint(c.MaxDepth, DEC_RADIX))
			urlObj.RawQuery = q.Encode()
			urlStr = urlObj.String()
		}
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
	result, err := c.doRequest(req, false)
	if err != nil {
		return nil, err
	}

	//autopaginate.
	resultMap, isMap := result.(map[string]interface{})
	if isMap {
		nextIn := resultMap["next"]
		valuesIn := resultMap["values"]
		if nextIn != nil && valuesIn != nil {
			nextUrl := nextIn.(string)
			if nextUrl != "" {
				valuesSlice := valuesIn.([]interface{})
				if valuesSlice != nil {
					nextResult, err := c.execute(method, nextUrl, text)
					if err != nil {
						return nil, err
					}
					nextResultMap, isNextMap := nextResult.(map[string]interface{})
					if !isNextMap {
						return nil, fmt.Errorf("next page result is not map, it's %T", nextResult)
					}
					nextValuesIn := nextResultMap["values"]
					if nextValuesIn == nil {
						return nil, fmt.Errorf("next page result has no values")
					}
					nextValuesSlice, isSlice := nextValuesIn.([]interface{})
					if !isSlice {
						return nil, fmt.Errorf("next page result 'values' is not slice")
					}
					valuesSlice = append(valuesSlice, nextValuesSlice...)
					resultMap["values"] = valuesSlice
					delete(resultMap, "page")
					delete(resultMap, "pagelen")
					delete(resultMap, "max_depth")
					delete(resultMap, "size")
					result = resultMap
				}
			}
		}
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
	return
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

	var result interface{}
	if err := json.NewDecoder(resBody).Decode(&result); err != nil {
		log.Println("Could not unmarshal JSON payload, returning raw response")
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
		resp.Body.Close()
		return nil, fmt.Errorf(resp.Status)
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
	case http.StatusOK:
		return false
	case http.StatusCreated:
		return false
	case http.StatusNoContent:
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
