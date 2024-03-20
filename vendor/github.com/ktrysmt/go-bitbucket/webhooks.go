package bitbucket

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type Webhooks struct {
	c *Client
}

type Webhook struct {
	Owner       string   `json:"owner"`
	RepoSlug    string   `json:"repo_slug"`
	Uuid        string   `json:"uuid"`
	Description string   `json:"description"`
	Url         string   `json:"url"`
	Active      bool     `json:"active"`
	Events      []string `json:"events"` // EX: {'repo:push','issue:created',..} REF: https://bit.ly/3FjRHHu
}

func decodeWebhook(response interface{}) (*Webhook, error) {
	respMap := response.(map[string]interface{})

	if respMap["type"] == "error" {
		return nil, DecodeError(respMap)
	}

	var webhook = new(Webhook)
	err := mapstructure.Decode(respMap, webhook)
	if err != nil {
		return nil, err
	}

	return webhook, nil
}

func decodeWebhooks(response interface{}) ([]Webhook, error) {
	webhooks := make([]Webhook, 0)
	resMap := response.(map[string]interface{})
	for _, v := range resMap["values"].([]interface{}) {
		wh, err := decodeWebhook(v)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, *wh)
	}
	return webhooks, nil
}

func (r *Webhooks) buildWebhooksBody(ro *WebhooksOptions) (string, error) {
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
		return "", err
	}

	return string(data), nil
}

func (r *Webhooks) List(ro *WebhooksOptions) ([]Webhook, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/", ro.Owner, ro.RepoSlug)
	res, err := r.c.executePaginated("GET", urlStr, "", nil)
	if err != nil {
		return nil, err
	}
	return decodeWebhooks(res)
}

// Deprecate Gets for List call
func (r *Webhooks) Gets(ro *WebhooksOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/", ro.Owner, ro.RepoSlug)
	return r.c.executePaginated("GET", urlStr, "", nil)
}

func (r *Webhooks) Create(ro *WebhooksOptions) (*Webhook, error) {
	data, err := r.buildWebhooksBody(ro)
	if err != nil {
		return nil, err
	}
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks", ro.Owner, ro.RepoSlug)
	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeWebhook(response)
}

func (r *Webhooks) Get(ro *WebhooksOptions) (*Webhook, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/%s", ro.Owner, ro.RepoSlug, ro.Uuid)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeWebhook(response)
}

func (r *Webhooks) Update(ro *WebhooksOptions) (*Webhook, error) {
	data, err := r.buildWebhooksBody(ro)
	if err != nil {
		return nil, err
	}
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/%s", ro.Owner, ro.RepoSlug, ro.Uuid)
	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeWebhook(response)
}

func (r *Webhooks) Delete(ro *WebhooksOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/hooks/%s", ro.Owner, ro.RepoSlug, ro.Uuid)
	return r.c.execute("DELETE", urlStr, "")
}
