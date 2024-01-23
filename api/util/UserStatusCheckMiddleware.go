package util

import (
	"github.com/devtron-labs/devtron/pkg/auth/user"
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
		userEmail, _, err := impl.userService.GetEmailAndGroupClaimsFromToken(token)
		if err != nil {
			log.Printf("unable to fetch user by token %s", token)
		}
		//todo - changes service function
		user, err := impl.userService.GetUserByEmail(userEmail)
		if err != nil {
			log.Printf("unable to fetch user by email %s", userEmail)
		}
		log.Print(user.Id)
		//TODO - put user id into context
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
