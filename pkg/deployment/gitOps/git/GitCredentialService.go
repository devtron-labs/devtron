package git

import (
	"strings"

	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	"go.uber.org/zap"
)

type GitCredentialService interface {
	GetCredentials(cfg *gitOps.GitOpsConfigDto) Credential
}

type Credential struct {
	Username string
	Token    string
}

type GitCredentialServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewGitCredentialServiceImpl(logger *zap.SugaredLogger) *GitCredentialServiceImpl {
	return &GitCredentialServiceImpl{
		logger: logger,
	}
}

func (impl *GitCredentialServiceImpl) GetCredentials(cfg *gitOps.GitOpsConfigDto) Credential {

	return Credential{
		Username: getUsername(cfg),
		Token:    cfg.Token,
	}
}

// GitOpsRepoGitUsername returns the username to use for git-over-HTTPS auth against the GitOps repo.
// This is the single source of truth for git credentials consumed by Devtron's own git operations,
// the ArgoCD credential template + secret, and the FluxCD git secret. Bitbucket Cloud token modes
// use a fixed git username that differs from the stored/REST username (the account email):
//
//	access token -> x-token-auth   |   API token -> x-bitbucket-api-token-auth
func getUsername(dto *gitOps.GitOpsConfigDto) string {
	if dto == nil {
		return ""
	}
	if strings.ToUpper(dto.Provider) == bean.BITBUCKET_PROVIDER {
		switch dto.AuthMode {
		case gitOps.ACCESS_TOKEN:
			return bean.BITBUCKET_ACCESS_TOKEN_USERNAME
		case gitOps.API_TOKEN:
			return bean.BITBUCKET_API_TOKEN_USERNAME
		}
	}
	return dto.Username
}

// BitbucketCloudAuthMode derives the effective Bitbucket Cloud auth mode. An explicit token mode
// (set via the API, or stored from a prior save) always wins; otherwise it is inferred from the
// username. Keeping the inference here means GitOpsRepoGitUsername does NOT depend on
// ResolveBitbucketCloudAuthMode having run upstream — a legacy config stored as PASSWORD with an
// email/x-token-auth username still resolves correctly on read paths (e.g. ArgoCD/FluxCD).
func BitbucketCloudAuthMode(username string, current gitOps.AuthMode) gitOps.AuthMode {
	if !current.IsPassword() {
		return current
	}
	switch {
	case username == bean.BITBUCKET_ACCESS_TOKEN_USERNAME: // TODO this is assuming that the frontend sets
		return gitOps.ACCESS_TOKEN
	case strings.Contains(username, "@"):
		return gitOps.API_TOKEN
	}
	return current
}
