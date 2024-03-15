/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"fmt"
	"github.com/go-pg/pg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ApiError struct {
	HttpStatusCode    int         `json:"-"`
	Code              string      `json:"code,omitempty"`
	InternalMessage   string      `json:"internalMessage,omitempty"`
	UserMessage       interface{} `json:"userMessage,omitempty"`
	UserDetailMessage string      `json:"userDetailMessage,omitempty"`
}

func NewApiError() *ApiError {
	return &ApiError{}
}

func (e *ApiError) WithHttpStatusCode(httpStatusCode int) *ApiError {
	e.HttpStatusCode = httpStatusCode
	return e
}

func (e *ApiError) WithCode(code string) *ApiError {
	e.Code = code
	return e
}
func (e *ApiError) WithInternalMessage(InternalMessage string) *ApiError {
	e.InternalMessage = InternalMessage
	return e
}
func (e *ApiError) WithUserMessage(userMessage interface{}) *ApiError {
	e.UserMessage = userMessage
	return e
}

func (e *ApiError) WithUserDetailMessage(UserDetailMessage string) *ApiError {
	e.UserDetailMessage = UserDetailMessage
	return e
}

func (e *ApiError) Error() string {
	return e.InternalMessage
}

// default internal will be set
func (e *ApiError) ErrorfInternal(format string, a ...interface{}) error {
	return &ApiError{InternalMessage: fmt.Sprintf(format, a...)}
}

// default user message will be set
func (e ApiError) ErrorfUser(format string, a ...interface{}) error {
	return &ApiError{InternalMessage: fmt.Sprintf(format, a...)}
}

func IsErrNoRows(err error) bool {
	return pg.ErrNoRows == err
}

func GetGRPCErrorDetailedMessage(err error) string {
	if errStatus, ok := status.FromError(err); ok {
		return errStatus.Message()
	}
	return err.Error()
}

func GetGRPCDetailedError(err error) (codes.Code, string) {
	if errStatus, ok := status.FromError(err); ok {
		return errStatus.Code(), errStatus.Message()
	}
	return codes.Unknown, err.Error()
}
