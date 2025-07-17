/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"context"
	"fmt"
	"k8s.io/utils/pointer"
	"reflect"
)

const IsSuperAdminFlag = "isSuperAdmin"
const Token = "token"
const UserId = "userId"
const UserEmailId = "emailId"
const DevtronToken = "devtronToken"

func SetSuperAdminInContext(ctx context.Context, isSuperAdmin bool) context.Context {
	ctx = context.WithValue(ctx, IsSuperAdminFlag, isSuperAdmin)
	return ctx
}

func GetIsSuperAdminFromContext(ctx context.Context) (bool, error) {
	flag := ctx.Value(IsSuperAdminFlag)

	if flag != nil && reflect.TypeOf(flag).Kind() == reflect.Bool {
		return flag.(bool), nil
	}
	return false, fmt.Errorf("context not valid, isSuperAdmin flag not set correctly %v", flag)
}

// SetTokenInContext - Set token in context
// NOTE: In OSS we don't have the token embedded in ctx already.
func SetTokenInContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, "token", token)
}

// GetDevtronTokenFromContext returns the devtronToken if found
// else return token from the context
func GetDevtronTokenFromContext(ctx context.Context) (string, error) {
	devtronToken := ctx.Value(DevtronToken)
	if devtronToken != nil && reflect.TypeOf(devtronToken).Kind() == reflect.String {
		return devtronToken.(string), nil
	}

	token := ctx.Value(Token)

	if token != nil && reflect.TypeOf(token).Kind() == reflect.String {
		return token.(string), nil
	}
	return "", fmt.Errorf("context not valid, token not set correctly %v", token)
}

type RequestCtx struct {
	token       *string
	userId      *int32
	userEmailId *string
	context.Context
}

func NewRequestCtx(ctx context.Context) *RequestCtx {
	return &RequestCtx{
		Context: ctx,
	}
}

func (r *RequestCtx) GetToken() string {
	if r.token != nil {
		return *r.token
	}
	tokenVal := r.Context.Value(Token)
	if tokenVal != nil {
		r.token = pointer.String(tokenVal.(string))
		return *r.token
	}
	return ""
}

func (r *RequestCtx) SetToken(token *string) {
	r.token = token
}

func (r *RequestCtx) GetUserId() int32 {
	if r.userId != nil {
		return *r.userId
	}
	userIdVal := r.Context.Value(UserId)
	if userIdVal != nil {
		r.userId = pointer.Int32(userIdVal.(int32))
		return *r.userId
	}
	return 0
}

func (r *RequestCtx) GetUserEmailId() string {
	if r.userEmailId != nil {
		return *r.userEmailId
	}
	userEmailIdVal := r.Context.Value(UserEmailId)
	if userEmailIdVal != nil {
		r.userEmailId = pointer.String(userEmailIdVal.(string))
		return *r.userEmailId
	}
	return ""
}

func GetEmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(UserEmailId).(string)
	return email
}

func GetTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(Token).(string)
	return token
}
