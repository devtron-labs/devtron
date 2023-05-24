package bitbucket

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type DeployKeys struct {
	c *Client
}

type DeployKey struct {
	Id      int    `json:"id"`
	Label   string `json:"label"`
	Key     string `json:"key"`
	Comment string `json:"comment"`
}

func decodeDeployKey(response interface{}) (*DeployKey, error) {
	respMap := response.(map[string]interface{})

	if respMap["type"] == "error" {
		return nil, DecodeError(respMap)
	}

	var deployKey = new(DeployKey)
	err := mapstructure.Decode(respMap, deployKey)
	if err != nil {
		return nil, err
	}

	return deployKey, nil
}

func buildDeployKeysBody(opt *DeployKeyOptions) (string, error) {
	body := map[string]interface{}{}
	body["label"] = opt.Label
	body["key"] = opt.Key

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (dk *DeployKeys) Create(opt *DeployKeyOptions) (*DeployKey, error) {
	data, err := buildDeployKeysBody(opt)
	if err != nil {
		return nil, err
	}
	urlStr := dk.c.requestUrl("/repositories/%s/%s/deploy-keys", opt.Owner, opt.RepoSlug)
	response, err := dk.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeDeployKey(response)
}

func (dk *DeployKeys) Get(opt *DeployKeyOptions) (*DeployKey, error) {
	urlStr := dk.c.requestUrl("/repositories/%s/%s/deploy-keys/%d", opt.Owner, opt.RepoSlug, opt.Id)
	response, err := dk.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeDeployKey(response)
}

func (dk *DeployKeys) Delete(opt *DeployKeyOptions) (interface{}, error) {
	urlStr := dk.c.requestUrl("/repositories/%s/%s/deploy-keys/%d", opt.Owner, opt.RepoSlug, opt.Id)
	return dk.c.execute("DELETE", urlStr, "")
}
