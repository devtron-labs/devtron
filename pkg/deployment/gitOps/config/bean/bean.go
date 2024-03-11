package bean

const (
	GitOpsCommitDefaultEmailId = "devtron-bot@devtron.ai"
	GitOpsCommitDefaultName    = "devtron bot"
)

// TODO: remove below object and its related methods to eliminate provider specific signature
type BitbucketProviderMetadata struct {
	BitBucketWorkspaceId string
	BitBucketProjectKey  string
}

const BITBUCKET_PROVIDER = "BITBUCKET_CLOUD"

type GitOpsConfigurationStatus struct {
	IsGitOpsConfigured    bool
	AllowCustomRepository bool
	Provider              string
}
