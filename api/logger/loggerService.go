package logger

import (
	"log"
)

func NewUserAuthService(auditLogDto *AuditLoggerDTO) {
	log.Printf("urlPath: %s, queryParams: %s, requestPayload: %s,updatedBy: %d, updatedOn: %s, apiResponseCode: %d", auditLogDto.UrlPath, auditLogDto.QueryParams, auditLogDto.RequestPayload, auditLogDto.UserID, auditLogDto.UpdatedOn, auditLogDto.ApiResponseCode)
}
