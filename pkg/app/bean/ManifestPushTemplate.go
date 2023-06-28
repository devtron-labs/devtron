package bean

import "time"

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
}

type ManifestPushResponse struct {
	CommitHash string
	CommitTime time.Time
	Error      error
}

type ContainerRegistryConfig struct {
	RegistryUrl string
	Username    string
	Password    string
	Insecure    bool
	AwsRegion   string
	AccessKey   string
	SecretKey   string
}

type HelmRepositoryConfig struct {
	RepositoryName        string
	ContainerRegistryName string
}

type GitRepositoryConfig struct {
	repositoryName string
}
