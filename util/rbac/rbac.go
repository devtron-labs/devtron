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

package rbac

import (
	"fmt"
	jwt2 "github.com/argoproj/argo-cd/util/jwt"
	"github.com/argoproj/argo-cd/util/session"
	"github.com/casbin/casbin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type Enforcer interface {
	Enforce(rvals ...interface{}) bool
	EnforceErr(rvals ...interface{}) error
	EnforceByEmail(rvals ...interface{}) bool
}

func NewEnforcerImpl(
	enforcer *casbin.Enforcer,
	sessionManager *session.SessionManager,
	logger *zap.SugaredLogger) *EnforcerImpl {
	enf := &EnforcerImpl{Enforcer: enforcer, logger: logger, SessionManager: sessionManager}
	return enf
}

// Enforcer is a wrapper around an Casbin enforcer that:
// * is backed by a kubernetes config map
// * has a predefined RBAC model
// * supports a built-in policy
// * supports a user-defined bolicy
// * supports a custom JWT claims enforce function
type EnforcerImpl struct {
	*casbin.Enforcer
	*session.SessionManager
	logger *zap.SugaredLogger
}

// Enforce is a wrapper around casbin.Enforce to additionally enforce a default role and a custom
// claims function
func (e *EnforcerImpl) Enforce(rvals ...interface{}) bool {
	return e.enforce(e.Enforcer, rvals...)
}

func (e *EnforcerImpl) EnforceByEmail(rvals ...interface{}) bool {
	return e.enforceByEmail(e.Enforcer, rvals...)
}

// EnforceErr is a convenience helper to wrap a failed enforcement with a detailed error about the request
func (e *EnforcerImpl) EnforceErr(rvals ...interface{}) error {
	if !e.Enforce(rvals...) {
		errMsg := "permission denied"
		if len(rvals) > 0 {
			rvalsStrs := make([]string, len(rvals)-1)
			for i, rval := range rvals[1:] {
				rvalsStrs[i] = fmt.Sprintf("%s", rval)
			}
			errMsg = fmt.Sprintf("%s: %s", errMsg, strings.Join(rvalsStrs, ", "))
		}
		return status.Error(codes.PermissionDenied, errMsg)
	}
	return nil
}

// enforce is a helper to additionally check a default role and invoke a custom claims enforcement function
func (e *EnforcerImpl) enforce(enf *casbin.Enforcer, rvals ...interface{}) bool {
	// check the default role
	if len(rvals) == 0 {
		return false
	}
	claims, err := e.SessionManager.VerifyToken(rvals[0].(string))
	if err != nil {
		return false
	}
	mapClaims, err := jwt2.MapClaims(claims)
	if err != nil {
		return false
	}
	email := jwt2.GetField(mapClaims, "email")
	sub := jwt2.GetField(mapClaims, "sub")
	if email == "" {
		email = sub
	}
	rvals[0] = strings.ToLower(email)
	return enf.Enforce(rvals...)
}

// enforce is a helper to additionally check a default role and invoke a custom claims enforcement function
func (e *EnforcerImpl) enforceByEmail(enf *casbin.Enforcer, rvals ...interface{}) bool {
	// check the default role
	if len(rvals) == 0 {
		return false
	}
	return enf.Enforce(rvals...)
}
