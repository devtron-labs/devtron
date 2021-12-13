/*
 * Copyright (c) 2021 Devtron Labs
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
 * Some of the code has been taken from argocd, for them argocd licensing terms apply
 */

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/authenticator/client"

	passwordutil "github.com/devtron-labs/authenticator/password"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
	"time"
)

type LoginService struct {
	sessionManager *SessionManager
	k8sClient      *client.K8sClient
}

func NewUserLogin(sessionManager *SessionManager, k8sClient *client.K8sClient) *LoginService {
	return &LoginService{
		sessionManager: sessionManager,
		k8sClient:      k8sClient,
	}
}
func (impl LoginService) Create(ctxt context.Context, username string, password string) (string, error) {
	if username == "" || password == "" {
		return "", status.Errorf(codes.Unauthenticated, "no credentials supplied")
	}
	err := impl.verifyUsernamePassword(username, password)
	if err != nil {
		return "", err
	}
	uniqueId, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	impl.sessionManager.GetUserSessionDuration().Seconds()
	jwtToken, err := impl.sessionManager.Create(
		fmt.Sprintf("%s", username),
		int64(impl.sessionManager.GetUserSessionDuration().Seconds()),
		uniqueId.String())
	return jwtToken, err
}

// VerifyUsernamePassword verifies if a username/password combo is correct
func (impl LoginService) verifyUsernamePassword(username string, password string) error {
	if password == "" {
		return status.Errorf(codes.Unauthenticated, blankPasswordError)
	}
	// Enforce maximum length of username on local accounts
	if len(username) > maxUsernameLength {
		return status.Errorf(codes.InvalidArgument, usernameTooLongError, maxUsernameLength)
	}
	account, err := impl.GetAccount(username)
	if err != nil {
		if errStatus, ok := status.FromError(err); ok && errStatus.Code() == codes.NotFound {
			err = InvalidLoginErr
		}
		return err
	}
	valid, _ := passwordutil.VerifyPassword(password, account.PasswordHash)
	if !valid {
		return InvalidLoginErr
	}
	if !account.Enabled {
		return status.Errorf(codes.Unauthenticated, accountDisabled, username)
	}
	if !account.HasCapability(AccountCapabilityLogin) {
		return status.Errorf(codes.Unauthenticated, userDoesNotHaveCapability, username, AccountCapabilityLogin)
	}
	return nil
}

type Account struct {
	PasswordHash  string
	PasswordMtime *time.Time
	Enabled       bool
	Capabilities  []AccountCapability
	Tokens        []Token
}
type AccountCapability string

const (
	AccountCapabilityLogin = "login"
)

type Token struct {
	ID        string `json:"id"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp,omitempty"`
}

// FormatPasswordMtime return the formatted password modify time or empty string of password modify time is nil.
func (a *Account) FormatPasswordMtime() string {
	if a.PasswordMtime == nil {
		return ""
	}
	return a.PasswordMtime.Format(time.RFC3339)
}

// FormatCapabilities returns comma separate list of user capabilities.
func (a *Account) FormatCapabilities() string {
	var items []string
	for i := range a.Capabilities {
		items = append(items, string(a.Capabilities[i]))
	}
	return strings.Join(items, ",")
}

// TokenIndex return an index of a token with the given identifier or -1 if token not found.
func (a *Account) TokenIndex(id string) int {
	for i := range a.Tokens {
		if a.Tokens[i].ID == id {
			return i
		}
	}
	return -1
}

// HasCapability return true if the account has the specified capability.
func (a *Account) HasCapability(capability AccountCapability) bool {
	for _, c := range a.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

func (impl LoginService) GetAccount(name string) (*Account, error) {
	if name != "admin" {
		return nil, fmt.Errorf("no account supported: %s", name)
	}
	secret, cm, err := impl.k8sClient.GetArgoConfig()
	if err != nil {
		return nil, err
	}
	account, err := parseAdminAccount(secret, cm)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func parseAdminAccount(secret *v1.Secret, cm *v1.ConfigMap) (*Account, error) {
	adminAccount := &Account{Enabled: true, Capabilities: []AccountCapability{AccountCapabilityLogin}}
	if adminPasswordHash, ok := secret.Data[client.SettingAdminPasswordHashKey]; ok {
		adminAccount.PasswordHash = string(adminPasswordHash)
	}
	if adminPasswordMtimeBytes, ok := secret.Data[client.SettingAdminPasswordMtimeKey]; ok {
		if mTime, err := time.Parse(time.RFC3339, string(adminPasswordMtimeBytes)); err == nil {
			adminAccount.PasswordMtime = &mTime
		}
	}

	adminAccount.Tokens = make([]Token, 0)
	if tokensStr, ok := secret.Data[client.SettingAdminTokensKey]; ok && string(tokensStr) != "" {
		if err := json.Unmarshal(tokensStr, &adminAccount.Tokens); err != nil {
			return nil, err
		}
	}
	if enabledStr, ok := cm.Data[client.SettingAdminEnabledKey]; ok {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			adminAccount.Enabled = enabled
		} else {
			log.Warnf("ConfigMap has invalid key %s: %v", client.SettingAdminTokensKey, err)
		}
	}

	return adminAccount, nil
}
