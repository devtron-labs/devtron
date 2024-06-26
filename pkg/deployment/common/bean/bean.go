package bean

type DeploymentConfig struct {
	Id                 int
	AppId              int
	EnvironmentId      int
	ConfigType         string
	DeploymentAppType  string
	RepoURL            string
	RepoName           string
	ChartLocation      string
	CredentialType     string
	CredentialIdInt    int
	CredentialIdString string
	Active             bool
}

type DeploymentConfigType string

const (
	CUSTOM           DeploymentConfigType = "custom"
	SYSTEM_GENERATED DeploymentConfigType = "system_generated"
)

type DeploymentConfigCredentialType string

const (
	GitOps DeploymentConfigCredentialType = "gitOps"
)
