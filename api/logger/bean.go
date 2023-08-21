package logger

import "time"

type AuditLoggerDTO struct {
	UrlPath         string    `json:"urlPath"`
	UserID          int       `json:"userID"`
	UpdatedOn       time.Time `json:"updatedOn"`
	QueryParams     string    `json:"queryParams"`
	ApiResponseCode int       `json:"apiResponseCode"`
	RequestPayload  string    `json:"requestPayload"`
}
