package common

const (
	UnAuthenticated     = "E100"
	UnAuthorized        = "E101"
	BadRequest          = "E102"
	InternalServerError = "E103"
	ResourceNotFound    = "E104"
	UnknownError        = "E105"
	CONTENT_DISPOSITION = "Content-Disposition"
	CONTENT_TYPE        = "Content-Type"
	CONTENT_LENGTH      = "Content-Length"
	APPLICATION_JSON    = "application/json"
)

var errorMessage = map[string]string{
	UnAuthenticated: "User is not authenticated",
	UnAuthorized:    "User is not authorized to perform this action",
}

func ErrorMessage(code string) string {
	return errorMessage[code]
}
