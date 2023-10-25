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

package argocdServer

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
