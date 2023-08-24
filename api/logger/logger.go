package logger

import (
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/pkg/user"
	"io"
	"log"
	"net/http"
	"time"
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
		// Call the next handler in the chain.
		d := middleware.NewDelegator(w, nil)
		next.ServeHTTP(d, r)

		token := r.Header.Get("token")
		userEmail, err := impl.userService.GetEmailFromToken(token)
		if err != nil {
			log.Printf("AUDIT_LOG: user does not exists")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("AUDIT_LOG: error reading request body for urlPath: %s queryParams: %s userEmail: %s", r.URL.Path, r.URL.Query().Encode(), userEmail)
		}
		auditLogDto := &AuditLoggerDTO{
			UrlPath:         r.URL.Path,
			UserEmail:       userEmail,
			UpdatedOn:       time.Now(),
			QueryParams:     r.URL.Query().Encode(),
			RequestPayload:  body,
			ApiResponseCode: d.Status(),
		}
		LogRequest(auditLogDto)
	})
}

func LogRequest(auditLogDto *AuditLoggerDTO) {
	log.Printf("AUDIT_LOG: urlPath: %s, queryParams: %s, requestPayload: %s,updatedBy: %s, updatedOn: %s, apiResponseCode: %d", auditLogDto.UrlPath, auditLogDto.QueryParams, auditLogDto.RequestPayload, auditLogDto.UserEmail, auditLogDto.UpdatedOn, auditLogDto.ApiResponseCode)
}
