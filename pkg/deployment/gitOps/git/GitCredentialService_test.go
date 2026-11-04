package git

import (
	"testing"

	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
)

func TestGitCredentialService(t *testing.T) {
	impl := &GitCredentialServiceImpl{}

	tests := []struct {
		name         string
		cfg          *apiGitOpsBean.GitOpsConfigDto
		wantUsername string
		wantToken    string
		wantSSHKey   string
	}{
		// Non-Bitbucket providers
		{
			name:         "github password returns configured username",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: "GITHUB", Username: "devtron-user", AuthMode: apiGitOpsBean.PASSWORD, Token: "ghp_tok"},
			wantUsername: "devtron-user",
			wantToken:    "ghp_tok",
		},
		{
			name:         "non-bitbucket with email username is NOT rewritten",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: "AZURE_DEVOPS", Username: "user@org.com", AuthMode: apiGitOpsBean.PASSWORD},
			wantUsername: "user@org.com",
		},

		// Bitbucket Cloud
		{
			name:         "bitbucket access token mode -> x-token-auth",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: bean.BITBUCKET_PROVIDER, Username: "user@org.com", AuthMode: apiGitOpsBean.ACCESS_TOKEN, Token: "atok"},
			wantUsername: bean.BITBUCKET_ACCESS_TOKEN_USERNAME,
			wantToken:    "atok",
		},
		{
			name:         "bitbucket api token mode -> x-bitbucket-api-token-auth",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: bean.BITBUCKET_PROVIDER, Username: "user@org.com", AuthMode: apiGitOpsBean.API_TOKEN, Token: "apitok"},
			wantUsername: bean.BITBUCKET_API_TOKEN_USERNAME,
			wantToken:    "apitok",
		},

		// auth mode inferred from username
		{
			name:         "bitbucket password mode with x-token-auth username infers access token",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: bean.BITBUCKET_PROVIDER, Username: bean.BITBUCKET_ACCESS_TOKEN_USERNAME, AuthMode: apiGitOpsBean.PASSWORD},
			wantUsername: bean.BITBUCKET_ACCESS_TOKEN_USERNAME,
		},
		{
			name:         "bitbucket password mode with email username infers api token",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: bean.BITBUCKET_PROVIDER, Username: "user", AuthMode: apiGitOpsBean.PASSWORD},
			wantUsername: "user",
		},
		{
			name:         "bitbucket empty auth mode with email username infers api token (backward compat)",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: bean.BITBUCKET_PROVIDER, Username: "user", AuthMode: ""},
			wantUsername: "user",
		},
		{
			name:         "bitbucket password mode with plain username stays app-password username",
			cfg:          &apiGitOpsBean.GitOpsConfigDto{Provider: bean.BITBUCKET_PROVIDER, Username: "plainuser", AuthMode: apiGitOpsBean.PASSWORD},
			wantUsername: "plainuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := impl.GetCredentials(tt.cfg)
			if got.Username != tt.wantUsername {
				t.Errorf("Username = %q, want %q", got.Username, tt.wantUsername)
			}
			if got.Token != tt.wantToken {
				t.Errorf("Token = %q, want %q", got.Token, tt.wantToken)
			}
			if got.SSHKey != tt.wantSSHKey {
				t.Errorf("SSHKey = %q, want %q", got.SSHKey, tt.wantSSHKey)
			}
		})
	}
}
