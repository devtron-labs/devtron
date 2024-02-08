package gitUtil

import "strings"

func GetGitRepoNameFromGitRepoUrl(gitRepoUrl string) string {
	gitRepoUrl = gitRepoUrl[strings.LastIndex(gitRepoUrl, "/")+1:]
	return strings.TrimSuffix(gitRepoUrl, ".git")
}
