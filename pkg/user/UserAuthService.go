/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package user

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casbin/casbin"
	"github.com/devtron-labs/authenticator/middleware"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/util/session"
	"github.com/caarlos0/env"
	"github.com/coreos/go-oidc"
	"github.com/devtron-labs/devtron/api/bean"
	session2 "github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type UserAuthService interface {
	HandleLogin(username string, password string) (string, error)
	HandleDexCallback(w http.ResponseWriter, r *http.Request)
	HandleRefresh(w http.ResponseWriter, r *http.Request)

	CreateRole(roleData *bean.RoleData) (bool, error)
	AuthVerification(r *http.Request) (bool, error)
	DeleteRoles(entityType string, entityName string, tx *pg.Tx) error
}

type UserAuthServiceImpl struct {
	userAuthRepository repository2.UserAuthRepository
	//sessionClient is being used for argocd username-password login proxy
	sessionClient  session2.ServiceClient
	logger         *zap.SugaredLogger
	userRepository repository2.UserRepository
	sessionManager *middleware.SessionManager
}

var (
	cStore         *sessions.CookieStore
	dexOauthConfig *oauth2.Config
	//googleOauthConfig *oauth2.Config
	oauthStateString     = randToken()
	idTokenVerifier      *oidc.IDTokenVerifier
	jwtKey               = randKey()
	CookieExpirationTime int
	JwtExpirationTime    int
)

type User struct {
	email  string
	groups []string
}

var Claims struct {
	Email    string   `json:"email"`
	Verified bool     `json:"email_verified"`
	Groups   []string `json:"groups"`
	Token    string   `json:"token"`
	Roles    []string `json:"roles"`
	jwt.StandardClaims
}

type DexConfig struct {
	RedirectURL          string `env:"DEX_RURL" envDefault:"http://127.0.0.1:8080/callback"`
	ClientID             string `env:"DEX_CID" envDefault:"example-app"`
	ClientSecret         string `env:"DEX_SECRET" `
	DexURL               string `env:"DEX_URL" `
	DexJwtKey            string `env:"DEX_JWTKEY" `
	CStoreKey            string `env:"DEX_CSTOREKEY"`
	CookieExpirationTime int    `env:"CExpirationTime" envDefault:"600"`
	JwtExpirationTime    int    `env:"JwtExpirationTime" envDefault:"120"`
}

type WebhookToken struct {
	WebhookToken string `env:"WEBHOOK_TOKEN" envDefault:""`
}

func NewUserAuthServiceImpl(userAuthRepository repository2.UserAuthRepository, sessionManager *middleware.SessionManager,
	client session2.ServiceClient, logger *zap.SugaredLogger, userRepository repository2.UserRepository,
) *UserAuthServiceImpl {
	serviceImpl := &UserAuthServiceImpl{
		userAuthRepository: userAuthRepository,
		sessionManager:     sessionManager,
		sessionClient:      client,
		logger:             logger,
		userRepository:     userRepository,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func GetConfig() (*DexConfig, error) {
	cfg := &DexConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func GetWebhookToken() (*WebhookToken, error) {
	cfg := &WebhookToken{}
	err := env.Parse(cfg)
	return cfg, err
}

/* #nosec */
func randToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		util.GetLogger().Error(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

/* #nosec */
func randKey() []byte {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		util.GetLogger().Error(err)
	}
	return b
}

// authorize verifies a bearer token and pulls user information form the claims.
func authorize(ctx context.Context, bearerToken string) (*User, error) {
	idToken, err := idTokenVerifier.Verify(ctx, bearerToken)
	if err != nil {
		return nil, fmt.Errorf("could not verify bearer token: %v", err)
	}
	if err := idToken.Claims(&Claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %v", err)
	}
	if !Claims.Verified {
		return nil, fmt.Errorf("email (%q) in returned claims was not verified", Claims.Email)
	}
	return &User{Claims.Email, Claims.Groups}, nil
}

func (impl UserAuthServiceImpl) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	session, _ := cStore.Get(r, "JWT_TOKEN")
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, dexOauthConfig.AuthCodeURL(oauthStateString), http.StatusFound)
	} else {
		jwtToken := session.Values["jwtToken"].(string)
		claims := &Claims

		// Parse the JWT string and store the result in `claims`.
		// Note that we are passing the key in this method as well. This method will return an error
		// if the token is invalid (if it has expired according to the expiry time we set on sign in),
		// or if the signature does not match
		tkn, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if !tkn.Valid {
			session.Options = &sessions.Options{
				MaxAge: -1,
			}
			writeResponse(http.StatusUnauthorized, "TOKEN EXPIRED", w, errors.New("token expired"))
			return
		}
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				writeResponse(http.StatusUnauthorized, "SignatureInvalid", w, errors.New("SignatureInvalid"))
				return
			}
			writeResponse(http.StatusBadRequest, "StatusBadRequest", w, errors.New("StatusBadRequest"))
			return
		}
		bearerToken := claims.Token
		user, err := authorize(context.Background(), bearerToken)
		if err != nil {
			fmt.Print("Exception :", err)
		}
		fmt.Print(user)

		// We ensure that a new token is not issued until enough time has elapsed
		// In this case, a new token will only be issued if the old token is within
		// 30 seconds of expiry. Otherwise, return a bad request status
		/*if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
			w.WriteHeader(http.StatusBadRequest)
			return
		}*/

		dbUser, err := impl.userRepository.FetchUserDetailByEmail(Claims.Email)
		if err != nil {
			impl.logger.Errorw("Exception while fetching user from db", "err", err)
		}
		if dbUser.Id > 0 {
			// Do nothing, User already exist in our db. (unique check by email id)
		} else {
			// TODO - need to handle case
		}

		// Now, create a new token for the current use, with a renewed expiration time
		expirationTime := time.Now().Add(time.Duration(JwtExpirationTime) * time.Second)
		// Create the JWT claims, which includes the username and expiry time
		claims.ExpiresAt = expirationTime.Unix()

		claims.Roles = dbUser.Roles
		// Declare the token with the algorithm used for signing, and the claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// Create the JWT string
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			writeResponse(http.StatusInternalServerError, "StatusInternalServerError", w, errors.New("unauthorized"))
			return
		}

		// Set some session values.
		session.Values["jwtToken"] = tokenString
		session.Values["authenticated"] = true
		session.Options = &sessions.Options{
			MaxAge: CookieExpirationTime,
		}
		// Save it before we write to the response/return from the handler.
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (impl UserAuthServiceImpl) HandleLogin(username string, password string) (string, error) {
	return impl.sessionClient.Create(context.Background(), username, password)
}

func (impl UserAuthServiceImpl) HandleDexCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	session, _ := cStore.Get(r, "JWT_TOKEN")
	fmt.Print(state)
	// Verify state.

	oauth2Token, err := dexOauthConfig.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		// handle error
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		// handle missing token
	}

	// Parse and verify ID Token payload.
	idToken, err := idTokenVerifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		// handle error
	}

	if err := idToken.Claims(&Claims); err != nil {
		// handle error
	}

	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return
	}
	// Rollback tx on error.
	defer tx.Rollback()

	dbUser, err := impl.userRepository.FetchUserDetailByEmail(Claims.Email)
	if err != nil {
		impl.logger.Errorw("Exception while fetching user from db", "err", err)
	}
	if dbUser.Id > 0 {
		// Do nothing, User already exist in our db. (unique check by email id)
	} else {
		//create new user in our db on d basis of info got from google api or hex. assign a basic role
		model := &repository2.UserModel{
			EmailId:     Claims.Email,
			AccessToken: rawIDToken,
		}
		_, err := impl.userRepository.CreateUser(model, tx)
		if err != nil {
			log.Println(err)
		}
		dbUser, err = impl.userRepository.FetchUserDetailByEmail(Claims.Email)
	}
	err = tx.Commit()
	if err != nil {
		return
	}

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(time.Duration(JwtExpirationTime) * time.Second)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims
	claims.Email = dbUser.EmailId
	claims.Verified = dbUser.Exist
	claims.ExpiresAt = expirationTime.Unix()
	claims.Token = rawIDToken
	claims.Roles = dbUser.Roles
	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set some session values.
	session.Values["jwtToken"] = tokenString
	session.Values["authenticated"] = true
	session.Options = &sessions.Options{
		MaxAge: CookieExpirationTime,
	}
	// Save it before we write to the response/return from the handler.
	session.Save(r, w)
	fmt.Print()

	http.Redirect(w, r, "/", http.StatusFound)
}

// Authorizer is a middleware for authorization
func Authorizer(e *casbin.Enforcer, sessionManager *session.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			//var users []string
			cookie, _ := r.Cookie("argocd.token")
			token := ""
			if cookie != nil {
				token = cookie.Value
				r.Header.Set("token", token)
			}
			if token == "" && cookie == nil {
				token = r.Header.Get("token")
				//if cookie == nil && len(token) != 0 {
				//	http.SetCookie(w, &http.Cookie{Name: "argocd.token", Value: token, Path: "/"})
				//}
			}
			//users = append(users, "anonymous")
			authEnabled := true
			pass := false
			config := auth.GetConfig()
			authEnabled = config.AuthEnabled

			if token != "" && authEnabled && !WhitelistChecker(r.URL.Path) {
				_, err := sessionManager.VerifyToken(token)
				if err != nil {
					log.Printf("Error verifying token: %+v\n", err)
					http.SetCookie(w, &http.Cookie{Name: "argocd.token", Value: token, Path: "/", MaxAge: -1})
					writeResponse(http.StatusUnauthorized, "Unauthorized", w, err)
					return
				}
				pass = true
				//TODO - we also can set user info in session (to avoid fetch it for all create n active)
			}
			if pass {
				next.ServeHTTP(w, r)
			} else if WhitelistChecker(r.URL.Path) {
				if r.URL.Path == "/app/ci-pipeline/github-webhook/trigger" {
					apiKey := r.Header.Get("api-key")
					t, err := GetWebhookToken()
					if err != nil || t.WebhookToken == "" || t.WebhookToken != apiKey {
						writeResponse(http.StatusUnauthorized, "UN-AUTHENTICATED", w, errors.New("unauthenticated"))
						return
					}
				}
				next.ServeHTTP(w, r)
			} else if token == "" {
				writeResponse(http.StatusUnauthorized, "UN-AUTHENTICATED", w, errors.New("unauthenticated"))
				return
			} else {
				writeResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
				return
			}

		}

		return http.HandlerFunc(fn)
	}
}

func WhitelistChecker(url string) bool {
	urls := []string{
		"/health",
		"/metrics",
		"/orchestrator/webhook/ci/gocd/artifact",
		"/orchestrator/webhook/ext-ci/",
		"/orchestrator/auth/login",
		"/orchestrator/auth/callback",
		"/orchestrator/api/v1/session",
		"/orchestrator/app/ci-pipeline/github-webhook/trigger",
		"/orchestrator/webhook/msg/nats",
		"/orchestrator/devtron/auth/verify",
		"/orchestrator/security/policy/verify/webhook",
		"/orchestrator/sso/list",
		"/",
	}
	for _, a := range urls {
		if a == url {
			return true
		}
	}
	prefixUrls := []string{
		"/orchestrator/webhook/ext-ci/",
		"/orchestrator/api/vi/pod/exec/ws",
		"/orchestrator/k8s/pod/exec/sockjs/ws",
		"/orchestrator/api/dex",
		"/orchestrator/auth/callback",
		"/orchestrator/auth/login",
		"/dashboard",
		"/orchestrator/webhook/git",
	}
	for _, a := range prefixUrls {
		if strings.Contains(url, a) {
			return true
		}
	}
	return false
}

func writeResponse(status int, message string, w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	type Response struct {
		Code   int              `json:"code,omitempty"`
		Status string           `json:"status,omitempty"`
		Result interface{}      `json:"result,omitempty"`
		Errors []*util.ApiError `json:"errors,omitempty"`
	}
	response := Response{}
	response.Code = status
	response.Result = message
	b, err := json.Marshal(response)
	if err != nil {
		b = []byte("OK")
		util.GetLogger().Errorw("Unexpected error in apiError", "err", err)
	}
	_, err = w.Write(b)
	if err != nil {
		util.GetLogger().Errorw("error", "err", err)
	}
}

func (impl UserAuthServiceImpl) CreateRole(roleData *bean.RoleData) (bool, error) {
	roleModel := &repository2.RoleModel{
		Role:        roleData.Role,
		Team:        roleData.Team,
		EntityName:  roleData.EntityName,
		Environment: roleData.Environment,
		Action:      roleData.Action,
	}
	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	roleModel, err = impl.userAuthRepository.CreateRole(roleModel, tx)
	if err != nil || roleModel == nil {
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (impl UserAuthServiceImpl) AuthVerification(r *http.Request) (bool, error) {
	token := r.Header.Get("token")
	if token == "" {
		impl.logger.Infow("no token provided", "token", token)
		err := &util.ApiError{
			HttpStatusCode:  http.StatusUnauthorized,
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "no token provided",
		}
		return false, err
	}

	_, err := impl.sessionManager.VerifyToken(token)
	if err != nil {
		impl.logger.Errorw("failed to verify token", "error", err)
		err := &util.ApiError{
			HttpStatusCode:  http.StatusUnauthorized,
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "failed to verify token",
			UserMessage:     fmt.Sprintf("token verification failed while getting logged in user: %s", token),
		}
		return false, err
	}

	//TODO - extends for other purpose
	return true, nil
}
func (impl UserAuthServiceImpl) DeleteRoles(entityType string, entityName string, tx *pg.Tx) (err error) {
	var roleModels []*repository2.RoleModel
	switch entityType {
	case repository2.PROJECT_TYPE:
		roleModels, err = impl.userAuthRepository.GetRolesForProject(entityName)
	case repository2.ENV_TYPE:
		roleModels, err = impl.userAuthRepository.GetRolesForEnvironment(entityName)
	case repository2.APP_TYPE:
		roleModels, err = impl.userAuthRepository.GetRolesForApp(entityName)
	case repository2.CHART_GROUP_TYPE:
		roleModels, err = impl.userAuthRepository.GetRolesForChartGroup(entityName)
	}
	if err != nil {
		impl.logger.Errorw(fmt.Sprintf("error in getting roles by %s", entityType), "err", err, "name", entityName)
		return err
	}

	// deleting policies in casbin for deleted roles
	var casbinDeleteFailed []bool
	for _, roleModel := range roleModels {
		success := casbin2.RemovePoliciesByRoles(roleModel.Role)
		if !success {
			impl.logger.Errorw("error in deleting casbin policy for role", "role", roleModel.Role)
			casbinDeleteFailed = append(casbinDeleteFailed, success)
		}
	}
	//deleting roles
	if len(roleModels) > 0 {
		err = impl.userAuthRepository.DeleteRoles(roleModels, tx)
		if err != nil {
			impl.logger.Errorw(fmt.Sprintf("error in deleting roles for %s:%s", entityType, entityName), "err", err, "name", entityName)
			return err
		}
	}
	return nil
}
