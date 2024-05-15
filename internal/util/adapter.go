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
func GetNotFoundError() *ApiError {
	return &ApiError{
		HttpStatusCode:  http.StatusNotFound,
		Code:            "400",
		UserMessage:     "Not Found",
		InternalMessage: "Not Found",
	}
}
