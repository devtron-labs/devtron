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
	"reflect"
)

const (
	IsSuperAdminFlag = "isSuperAdmin"
	UserId           = "userId"
	EmailId          = "emailId"
	Token            = "token"
)

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

func GetUserIdFromContext(ctx context.Context) (int32, error) {
	flag := ctx.Value(UserId)

	if flag != nil && reflect.TypeOf(flag).Kind() == reflect.Int32 {
		return flag.(int32), nil
	}
	return 0, fmt.Errorf("context not valid, userId flag not set correctly %v", flag)
}

func GetUserEmailFromContext(ctx context.Context) (string, error) {
	flag := ctx.Value(EmailId)

	if flag != nil && reflect.TypeOf(flag).Kind() == reflect.String {
		return flag.(string), nil
	}
	return "", fmt.Errorf("context not valid, emailId flag not set correctly %v", flag)
}
func GetTokenFromContext(ctx context.Context) (string, error) {
	flag := ctx.Value(Token)

	if flag != nil && reflect.TypeOf(flag).Kind() == reflect.String {
		return flag.(string), nil
	}
	return "", fmt.Errorf("context not valid, token flag not set correctly %v", flag)
}
