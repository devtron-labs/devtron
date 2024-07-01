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

package commandManager

import (
	"context"
	"time"
)

type GitContext struct {
	context.Context
	auth *BasicAuth
}

func (gitCtx GitContext) WithCredentials(auth *BasicAuth) GitContext {
	gitCtx.auth = auth
	return gitCtx
}

func BuildGitContext(ctx context.Context) GitContext {
	return GitContext{
		Context: ctx,
	}
}

func (gitCtx GitContext) WithTimeout(timeoutSeconds int) (GitContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(gitCtx.Context, time.Duration(timeoutSeconds)*time.Second)
	gitCtx.Context = ctx
	return gitCtx, cancel
}

// BasicAuth represent a HTTP basic auth
type BasicAuth struct {
	Username, Password string
}
