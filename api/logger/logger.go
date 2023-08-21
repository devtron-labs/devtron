package logger

import (
	"bytes"
	"github.com/devtron-labs/devtron/pkg/user"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type CapturingReadCloser struct {
	original io.ReadCloser
	captured *bytes.Buffer
}

type UserAuthImpl struct {
	userService user.UserService
}

func NewUserAuthImpl(userService user.UserService) *UserAuthImpl {
	return &UserAuthImpl{
		userService: userService,
	}
}

type UserAuth interface {
	LoggingMiddleware(next http.Handler) http.Handler
}

// LoggingMiddleware is a middleware function that logs the incoming request.
func (impl UserAuthImpl) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)

		//Log the request details.
		url := r.URL.Path

		//status := w.Header().Get("Status")
		//statusCode, err := strconv.Atoi(status)
		//if err != nil {
		//	log.Printf("error in converting status code to int")
		//}

		token := r.Header.Get("token")
		userId, userType, err := impl.userService.GetUserByToken(r.Context(), token)
		if err != nil {
			log.Printf("userId does not exists")
		}
		vars := r.URL.Query().Encode()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("error reading request body")
		}
		// Convert the request body to a string
		requestPayload := string(body)
		log.Printf(userType)
		auditLogDto := &AuditLoggerDTO{
			UrlPath:        url,
			UserID:         int(userId),
			UpdatedOn:      startTime,
			QueryParams:    vars,
			RequestPayload: requestPayload,
		}
		NewUserAuthService(auditLogDto)
	})
}
