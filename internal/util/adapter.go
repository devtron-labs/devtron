/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import "net/http"

func GetApiErrorAdapter(httpStatusCode int, code, userMessage, internalMessage string) *ApiError {
	return &ApiError{
		HttpStatusCode:  httpStatusCode,
		Code:            code,
		UserMessage:     userMessage,
		InternalMessage: internalMessage,
	}
}
func GetUnProcessableError() *ApiError {
	return &ApiError{
		HttpStatusCode:  http.StatusUnprocessableEntity,
		Code:            "422",
		UserMessage:     "UnprocessableEntity",
		InternalMessage: "UnprocessableEntity",
	}
}
