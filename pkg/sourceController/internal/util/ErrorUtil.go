package util

import (
	"fmt"
	"github.com/go-pg/pg"
)

type ApiError struct {
	HttpStatusCode    int         `json:"-"`
	Code              string      `json:"code,omitempty"`
	InternalMessage   string      `json:"internalMessage,omitempty"`
	UserMessage       interface{} `json:"userMessage,omitempty"`
	UserDetailMessage string      `json:"userDetailMessage,omitempty"`
}

func (e *ApiError) Error() string {
	return e.InternalMessage
}

// default internal will be set
func (e *ApiError) ErrorfInternal(format string, a ...interface{}) error {
	return &ApiError{InternalMessage: fmt.Sprintf(format, a...)}
}

//default user message will be set
func (e ApiError) ErrorfUser(format string, a ...interface{}) error {
	return &ApiError{InternalMessage: fmt.Sprintf(format, a...)}
}

func IsErrNoRows(err error) bool {
	return pg.ErrNoRows == err
}
