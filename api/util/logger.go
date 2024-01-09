package util

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/pkg/auth/user"
)

type AuditLoggerDTO struct {
	UrlPath         string    `json:"urlPath"`
	UserEmail       string    `json:"userEmail"`
	UpdatedOn       time.Time `json:"updatedOn"`
	QueryParams     string    `json:"queryParams"`
	ApiResponseCode int       `json:"apiResponseCode"`
	RequestPayload  []byte    `json:"requestPayload"`
}

type LoggingMiddlewareImpl struct {
	userService user.UserService
}

func NewLoggingMiddlewareImpl(userService user.UserService) *LoggingMiddlewareImpl {
	return &LoggingMiddlewareImpl{
		userService: userService,
	}
}

type LoggingMiddleware interface {
	LoggingMiddleware(next http.Handler) http.Handler
}

// LoggingMiddleware is a middleware function that logs the incoming request.
func (impl LoggingMiddlewareImpl) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := middleware.NewDelegator(w, nil)

		token := r.Header.Get("token")
		userEmail, _, err := impl.userService.GetEmailAndGroupClaimsFromToken(token)
		if err != nil {
			log.Printf("AUDIT_LOG: user does not exists")
		}

		// Read the request body into a buffer
		var bodyBuffer bytes.Buffer
		_, err = io.Copy(&bodyBuffer, r.Body)
		if err != nil {
			log.Printf("AUDIT_LOG: error reading request body for urlPath: %s queryParams: %s userEmail: %s", r.URL.Path, r.URL.Query().Encode(), userEmail)
		}

		// Restore the request body for downstream handlers
		r.Body = io.NopCloser(&bodyBuffer)

		auditLogDto := &AuditLoggerDTO{
			UrlPath:        r.URL.Path,
			UserEmail:      userEmail,
			UpdatedOn:      time.Now(),
			QueryParams:    r.URL.Query().Encode(),
			RequestPayload: bodyBuffer.Bytes(),
		}
		// Call the next handler in the chain.
		next.ServeHTTP(d, r)

		auditLogDto.ApiResponseCode = d.Status()
		LogRequest(auditLogDto)
	})
}

func LogRequest(auditLogDto *AuditLoggerDTO) {
	log.Printf("AUDIT_LOG: urlPath: %s, queryParams: %s,updatedBy: %s, updatedOn: %s, apiResponseCode: %d,requestPayload: %s", auditLogDto.UrlPath, auditLogDto.QueryParams, auditLogDto.UserEmail, auditLogDto.UpdatedOn, auditLogDto.ApiResponseCode, auditLogDto.RequestPayload)
}
