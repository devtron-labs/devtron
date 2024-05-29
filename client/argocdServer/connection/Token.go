/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package connection

import "context"

type TokenAuth struct {
	token string
}

// Return value is mapped to request headers.
func (t TokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"token": t.token,
	}, nil
}

func (TokenAuth) RequireTransportSecurity() bool {
	return false
}

var tokenAuth TokenAuth

func GetTokenAuth() *TokenAuth {
	return &tokenAuth
}

func SetTokenAuth(token string) {
	tokenAuth = TokenAuth{token: token}
}
