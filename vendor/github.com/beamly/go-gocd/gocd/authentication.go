package gocd

import "context"

// Login sends basic auth to the GoCD Server and sets an auth cookie in the client to enable cookie based auth
// for future requests.
func (c *Client) Login(ctx context.Context) (err error) {
	var req *APIRequest
	var resp *APIResponse

	req, err = c.NewRequest("GET", "api/agents", nil, apiV2)
	if err != nil {
		return
	}
	req.HTTP.SetBasicAuth(c.params.Username, c.params.Password)

	resp, err = c.Do(ctx, req, nil, responseTypeJSON)
	if err == nil {
		c.cookie = resp.HTTP.Header["Set-Cookie"][0]
	}
	return
}
