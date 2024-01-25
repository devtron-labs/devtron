package util

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/response"
	"log"
	"net/http"
)

type UserStatusCheckMiddleware interface {
	UserStatusCheckMiddleware(next http.Handler) http.Handler
}

type UserStatusCheckMiddlewareImpl struct {
	userService user.UserService
}

func NewUserStatusCheckMiddlewareImpl(userService user.UserService) *UserStatusCheckMiddlewareImpl {
	return &UserStatusCheckMiddlewareImpl{
		userService: userService,
	}
}

func (impl UserStatusCheckMiddlewareImpl) UserStatusCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")
		emailId, _, err := impl.userService.GetEmailAndGroupClaimsFromToken(token)
		if err != nil {
			log.Printf("unable to fetch user by token %s", token)
		}
		userId, isInactive, err := impl.userService.GetUserWithTimeoutWindowConfiguration(emailId)
		if err != nil {
			log.Printf("unable to fetch user by email %s", emailId)
			// todo - correct status code
			response.WriteResponse(http.StatusUnauthorized, "UN-AUTHENTICATED", w, fmt.Errorf("unauthenticated"))
			return
		}
		if isInactive {
			response.WriteResponse(http.StatusUnauthorized, "UN-AUTHENTICATED", w, fmt.Errorf("unauthenticated"))
			return
		}
		//TODO - put user id into context
		context.WithValue(r.Context(), "userId", userId)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
