package bitbucket

import (
	"encoding/json"
	"os"

	"github.com/k0kubun/pp"
)

type Webhooks struct {
	c *Client
}

func (r *Webhooks) buildWebhooksBody(ro *WebhooksOptions) string {

	body := map[string]interface{}{}

	if ro.Description != "" {
		body["description"] = ro.Description
	}
	if ro.Url != "" {
		body["url"] = ro.Url
	}
	if ro.Active == true || ro.Active == false {
		body["active"] = ro.Active
	}

	body["events"] = ro.Events

	data, err := json.Marshal(body)
	if err != nil {
		pp.Println(err)
		os.Exit(9)
	}

	return string(data)
}

func (r *Webhooks) Gets(ro *WebhooksOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/", ro.Owner, ro.RepoSlug)
	return r.c.execute("GET", urlStr, "")
}

func (r *Webhooks) Create(ro *WebhooksOptions) (interface{}, error) {
	data := r.buildWebhooksBody(ro)
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks", ro.Owner, ro.RepoSlug)
	return r.c.execute("POST", urlStr, data)
}

func (r *Webhooks) Get(ro *WebhooksOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/%s", ro.Owner, ro.RepoSlug, ro.Uuid)
	return r.c.execute("GET", urlStr, "")
}

func (r *Webhooks) Update(ro *WebhooksOptions) (interface{}, error) {
	data := r.buildWebhooksBody(ro)
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/%s", ro.Owner, ro.RepoSlug, ro.Uuid)
	return r.c.execute("PUT", urlStr, data)
}

func (r *Webhooks) Delete(ro *WebhooksOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/%s", ro.Owner, ro.RepoSlug, ro.Uuid)
	return r.c.execute("DELETE", urlStr, "")
}

//
