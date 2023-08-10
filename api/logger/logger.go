package logger

import (
	"github.com/devtron-labs/devtron/pkg/user"
	"log"
	"net/http"
	"time"
)

type UserAuth struct {
	userService user.UserService
}

// LoggingMiddleware is a middleware function that logs the incoming request.
func (impl UserAuth) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)

		// Log the request details.
		method := r.Method
		url := r.URL.Path

		status := w.(interface {
			Status() int
		}).Status()

		token := r.Header.Get("token")
		userId, userType, err := impl.userService.GetUserByToken(r.Context(), token)
		if err != nil {
			log.Printf("userId does not exists")
		}
		log.Printf("[%s] %d %d %s %s %d", method, status, userId, userType, url, time.Since(startTime))

	})
}
