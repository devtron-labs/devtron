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
	UserStatusCheckInDb(token string) (bool, int32, error)
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
		//token := r.Header.Get("token")
		token := ""
		cookie, _ := r.Cookie("argocd.token")
		if cookie != nil {
			token = cookie.Value
			r.Header.Set("token", token)
			log.Print("token fetching from cookie")
		}
		if token == "" && cookie == nil {
			token = r.Header.Get("token")
			log.Print("token fetching from header")
		}
		emailId, _, err := impl.userService.GetEmailAndGroupClaimsFromToken(token)
		if err != nil {
			log.Print("unable to fetch user by token")
		}
		userId, isInactive, err := impl.userService.GetUserWithTimeoutWindowConfiguration(emailId)
		if err != nil {
			log.Printf("unable to fetch user by email : %s", emailId)
			// todo - correct status code
			//response.WriteResponse(http.StatusUnauthorized, "UN-AUTHENTICATED", w, fmt.Errorf("unauthenticated"))
			//return
		}
		if isInactive && err == nil {
			response.WriteResponse(http.StatusUnauthorized, "UN-AUTHENTICATED", w, fmt.Errorf("unauthenticated"))
			return
		}
		//TODO - put user id into context
		context.WithValue(r.Context(), "userId", userId)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (impl UserStatusCheckMiddlewareImpl) UserStatusCheckInDb(token string) (bool, int32, error) {
	emailId, _, err := impl.userService.GetEmailAndGroupClaimsFromToken(token)
	if err != nil {
		log.Printf("unable to fetch user by token")
		return false, 0, err
	}
	userId, isInactive, err := impl.userService.GetUserWithTimeoutWindowConfiguration(emailId)
	if err != nil {
		log.Printf("unable to fetch user by email : %s", emailId)
		return isInactive, userId, err
	}
	return isInactive, userId, nil
}
