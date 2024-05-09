package utils

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

// Create a Resty Client
var client = resty.New().
	SetHeader("Accept", "application/json")

func CallGetApi[T any](api string, query, headers map[string]string, result *T) error {

	resp, err := client.R().
		SetQueryParams(query).
		SetResult(result).
		EnableTrace().
		SetHeaders(headers).
		Get(api)

	if resp.StatusCode() != 200 {
		err = fmt.Errorf("API %s call failed with reason %s", api, string(resp.Body()))
	}
	return err
}

func CallPostApi[T, R any](api string, query, headers map[string]string, request R, response *T) error {

	resp, err := client.R().
		SetBody(request).
		SetResult(response).
		SetQueryParams(query).
		SetHeaders(headers).
		EnableTrace().
		Post(api)

	if resp.StatusCode() != 200 {
		err = fmt.Errorf("API %s call failed with reason %s", api, string(resp.Body()))
	}
	return err
}

type Response struct {
	Code   int         `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Errors interface{} `json:"errors,omitempty"`
}
