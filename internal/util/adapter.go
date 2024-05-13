package util

func GetApiErrorAdapter(httpStatusCode int, code, userMessage, internalMessage string) *ApiError {
	return &ApiError{
		HttpStatusCode:  httpStatusCode,
		Code:            code,
		UserMessage:     userMessage,
		InternalMessage: internalMessage,
	}
}
