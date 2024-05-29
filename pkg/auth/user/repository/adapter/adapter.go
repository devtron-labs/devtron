/*
 * Copyright (c) 2024. Devtron Inc.
 */

package adapter

import "github.com/devtron-labs/devtron/pkg/auth/user/repository"

func GetUserModelBasicAdapter(emailId, accessToken, userType string) *repository.UserModel {
	model := &repository.UserModel{
		EmailId:     emailId,
		AccessToken: accessToken,
		UserType:    userType,
	}
	return model
}
