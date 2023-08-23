package logger

import "time"

type AuditLoggerDTO struct {
	UrlPath         string    `json:"urlPath"`
	UserEmail       string    `json:"userEmail"`
	UpdatedOn       time.Time `json:"updatedOn"`
	QueryParams     string    `json:"queryParams"`
	ApiResponseCode int       `json:"apiResponseCode"`
	RequestPayload  []byte    `json:"requestPayload"`
}
