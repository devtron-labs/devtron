package logger

import (
	"log"
)

func LogRequest(auditLogDto *AuditLoggerDTO) {
	log.Printf("urlPath: %s, queryParams: %s, requestPayload: %s,updatedBy: %s, updatedOn: %s, apiResponseCode: %d", auditLogDto.UrlPath, auditLogDto.QueryParams, auditLogDto.RequestPayload, auditLogDto.UserEmail, auditLogDto.UpdatedOn, auditLogDto.ApiResponseCode)
}
