package logger

import (
	"bytes"
	"github.com/devtron-labs/devtron/pkg/user"
	"io"
	"log"
	"net/http"
	"time"
)

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
		// Extract the request payload.
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Restore the original request body for downstream handlers.
		r.Body = io.NopCloser(bytes.NewBuffer(payload))

		// Log the request details.
		//method := r.Method
		url := r.URL.Path

		status := w.(interface {
			Status() int
		}).Status()

		token := r.Header.Get("token")
		userId, userType, err := impl.userService.GetUserByToken(r.Context(), token)
		if err != nil {
			log.Printf("userId does not exists")
		}
		log.Printf(userType)
		NewUserAuthService(url, userId, time.Since(startTime), status, string(payload))
	})
}
