package adapter

import (
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
)

func ConvertGitOpsConfigToGitConfig(dto *bean2.GitOpsConfigDto) *bean.GitConfig {
	return &bean.GitConfig{
		GitlabGroupId:        dto.GitLabGroupId,
		GitToken:             dto.Token,
		GitUserName:          dto.Username,
		GithubOrganization:   dto.GitHubOrgId,
		GitProvider:          dto.Provider,
		GitHost:              dto.Host,
		AzureToken:           dto.Token,
		AzureProject:         dto.AzureProjectName,
		BitbucketWorkspaceId: dto.BitBucketWorkspaceId,
		BitbucketProjectKey:  dto.BitBucketProjectKey,
	}
}
