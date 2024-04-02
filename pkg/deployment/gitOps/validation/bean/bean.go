package bean

type ExtraValidationStageType int

type ValidateCustomGitRepoURLRequest struct {
	GitRepoURL     string
	AppName        string
	UserId         int32
	GitOpsProvider string
}

const (
	DryrunRepoName    = "devtron-sample-repo-dryrun-"
	DeleteRepoStage   = "Delete Repo"
	CommitOnRestStage = "Commit On Rest"
	PushStage         = "Push"
	CloneStage        = "Clone"
	GetRepoUrlStage   = "Get Repo RedirectionUrl"
	CreateRepoStage   = "Create Repo"
	CloneHttp         = "Clone Http"
	CloneSSH          = "Clone Ssh"
	CreateReadmeStage = "Create Readme"
)
