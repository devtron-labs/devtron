package api

import "github.com/devtron-labs/devtron/pkg/sourceController/internal/util"

type Response struct {
	Code   int              `json:"code,omitempty"`
	Status string           `json:"status,omitempty"`
	Result interface{}      `json:"result,omitempty"`
	Errors []*util.ApiError `json:"errors,omitempty"`
}
