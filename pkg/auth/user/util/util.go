/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package util

import (
	"context"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"strings"
)

const (
	ApiTokenPrefix = "API-TOKEN:"
)

func CheckValidationForRoleGroupCreation(name string) bool {
	if strings.Contains(name, ",") {
		return false
	}
	return true
}

func CheckIfAdminOrApiToken(email string) bool {
	if email == "admin" || CheckIfApiToken(email) {
		return true
	}
	return false
}

func CheckIfApiToken(email string) bool {
	return strings.HasPrefix(email, ApiTokenPrefix)
}

func GetUserMetadata(ctx context.Context, userId int32, isSuperAdmin bool) *bean.UserMetadata {
	// Get user email from context
	userEmail := util2.GetEmailFromContext(ctx)

	// Create and return the UserMetadata object
	return &bean.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
}
