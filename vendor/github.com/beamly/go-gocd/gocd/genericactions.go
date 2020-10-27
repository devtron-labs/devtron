package gocd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// APIClientRequest helper struct to reduce amount of code.
type APIClientRequest struct {
	Method       string
	Path         string
	APIVersion   string
	RequestBody  interface{}
	ResponseType string
	ResponseBody interface{}
	Headers      map[string]string
}

// Handles any call to HEAD by returning whether or not we got a 2xx code.
func (c *Client) genericHeadAction(ctx context.Context, path string, apiversion string) (bool, *APIResponse, error) {
	_, resp, err := c.httpAction(ctx, &APIClientRequest{
		Method:       "HEAD",
		Path:         path,
		APIVersion:   apiversion,
		ResponseType: responseTypeJSON,
	})

	exists := resp.HTTP.StatusCode >= 300 || resp.HTTP.StatusCode < 200

	return exists, resp, err

}

func (c *Client) patchAction(ctx context.Context, r *APIClientRequest) (interface{}, *APIResponse, error) {
	r.Method = "PATCH"
	return c.httpAction(ctx, r)
}

func (c *Client) getAction(ctx context.Context, r *APIClientRequest) (interface{}, *APIResponse, error) {
	r.Method = "GET"
	return c.httpAction(ctx, r)
}

func (c *Client) postAction(ctx context.Context, r *APIClientRequest) (interface{}, *APIResponse, error) {
	r.Method = "POST"
	return c.httpAction(ctx, r)
}

func (c *Client) putAction(ctx context.Context, r *APIClientRequest) (interface{}, *APIResponse, error) {
	r.Method = "PUT"
	return c.httpAction(ctx, r)
}

// Returns a message from the DELETE action on the provided HTTP resource.
func (c *Client) deleteAction(ctx context.Context, path string, apiversion string) (string, *APIResponse, error) {
	a := StringResponse{}
	_, resp, err := c.httpAction(ctx, &APIClientRequest{
		Method:       "DELETE",
		Path:         path,
		APIVersion:   apiversion,
		ResponseType: responseTypeJSON,
		ResponseBody: &a,
	})

	return a.Message, resp, err
}

func (c *Client) httpAction(ctx context.Context, r *APIClientRequest) (responseBody interface{}, resp *APIResponse, err error) {

	var req *APIRequest
	var requestBodyProvided, hasEmptyResponseType, hasJSONResponseType bool

	requestBodyProvided = r.RequestBody != nil
	hasEmptyResponseType = r.ResponseType == ""
	if !hasEmptyResponseType {
		hasJSONResponseType = r.ResponseType == responseTypeJSON
	}

	c.Log.WithFields(logrus.Fields{
		"Method": r.Method,
		"Path":   r.Path,
	}).Debug("Requesting Endpoint")

	if hasEmptyResponseType {
		r.ResponseType = responseTypeJSON
	}

	versionAction(r.RequestBody, func(ver Versioned) {
		if r.Headers == nil {
			r.Headers = map[string]string{}
		}
		r.Headers["If-Match"] = fmt.Sprintf(`"%s"`, ver.GetVersion())
	})

	// Build the request
	if req, err = c.NewRequest(r.Method, r.Path, r.RequestBody, r.APIVersion); err != nil {
		return false, nil, err
	}
	if requestBodyProvided {
		c.Log.WithField("RequestBody", req.Body).Debug("Sending Request Body")
	}

	if len(r.Headers) > 0 {
		for key, value := range r.Headers {
			req.HTTP.Header.Set(key, value)
		}
	}

	c.Log.WithFields(headerLogFields(req.HTTP.Header)).Debug("Request Header")

	if resp, err = c.Do(ctx, req, r.ResponseBody, r.ResponseType); err != nil {
		return r.ResponseBody, resp, err
	}

	versionAction(r.ResponseBody, func(ver Versioned) {
		parseVersions(resp.HTTP, ver)
	})

	if hasJSONResponseType {
		b, _ := json.Marshal(r.ResponseBody)
		c.Log.WithFields(headerLogFields(resp.HTTP.Header)).Debug("Response Headers")
		c.Log.WithFields(logrus.Fields{
			"Protocol": resp.HTTP.Proto,
			"Status":   resp.HTTP.Status,
			"Body":     string(b),
		}).Debug("Response")
	}

	return r.ResponseBody, resp, err
}

func versionAction(versioned interface{}, verFunc func(versionsed Versioned)) {
	if ver, isVersioned := versioned.(Versioned); isVersioned {
		verFunc(ver)
	}
}

func headerLogFields(headers http.Header) logrus.Fields {
	fields := logrus.Fields{}
	for header, values := range headers {
		for _, value := range values {
			fields[header] = value
		}
	}
	return fields
}

func parseVersions(response *http.Response, versioned Versioned) {
	etag := response.Header.Get("Etag")
	versioned.SetVersion(
		strings.Replace(etag, "\"", "", -1),
	)
}
