package bean

import "time"

const WORKFLOW_EXIST_ERROR = "workflow with this name already exist in this app"
const Workflows = "workflows"

type ManifestPushTemplate struct {
	WorkflowRunnerId        int
	AppId                   int
	ChartRefId              int
	EnvironmentId           int
	UserId                  int32
	PipelineOverrideId      int
	AppName                 string
	TargetEnvironmentName   int
	ChartReferenceTemplate  string
	ChartName               string
	ChartVersion            string
	ChartLocation           string
	RepoUrl                 string
	RepoName                string
	BuiltChartPath          string
	BuiltChartBytes         *[]byte
	MergedValues            string
	ContainerRegistryConfig *ContainerRegistryConfig
	StorageType             string
}

type ManifestPushResponse struct {
	CommitHash string
	CommitTime time.Time
	Error      error
}

type ContainerRegistryConfig struct {
	RegistryUrl  string
	Username     string
	Password     string
	Insecure     bool
	AwsRegion    string
	AccessKey    string
	SecretKey    string
	RegistryType string
	IsPublic     bool
	RepoName     string
}

type HelmRepositoryConfig struct {
	RepositoryName        string
	ContainerRegistryName string
}

type GitRepositoryConfig struct {
	repositoryName string
}
