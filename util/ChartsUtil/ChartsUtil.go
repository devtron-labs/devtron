package ChartsUtil

import bean2 "github.com/devtron-labs/devtron/api/bean"

func IsGitOpsRepoNotConfigured(gitRepoUrl string) bool {
	return len(gitRepoUrl) == 0 || gitRepoUrl == bean2.GIT_REPO_NOT_CONFIGURED
}
