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
