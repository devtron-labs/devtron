package logger

import (
	"bytes"
	"github.com/devtron-labs/devtron/pkg/user"
	"io"
	"log"
	"net/http"
	"time"
)

type CapturingReadCloser struct {
	original io.ReadCloser
	captured *bytes.Buffer
}

func NewCapturingReadCloser(original io.ReadCloser) *CapturingReadCloser {
	return &CapturingReadCloser{
		original: original,
		captured: new(bytes.Buffer),
	}
}

func (crc *CapturingReadCloser) Read(p []byte) (n int, err error) {
	n, err = crc.original.Read(p)
	if n > 0 {
		crc.captured.Write(p[:n])
	}
	return
}

func (crc *CapturingReadCloser) Close() error {
	return crc.original.Close()
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

		// Capture the request payload while preserving the original request body.
		capturingBody := NewCapturingReadCloser(r.Body)
		r.Body = capturingBody

		//Log the request details.
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
		NewUserAuthService(url, userId, time.Since(startTime), status, capturingBody.captured.String())
	})
}
