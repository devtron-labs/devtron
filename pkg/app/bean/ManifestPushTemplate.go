package bean

import "time"

const WORKFLOW_EXIST_ERROR = "workflow with this name already exist in this app"
const Workflows = "workflows"

type ManifestPushTemplate struct {
	WorkflowRunnerId            int
	AppId                       int
	ChartRefId                  int
	EnvironmentId               int
	EnvironmentName             string
	UserId                      int32
	PipelineOverrideId          int
	AppName                     string
	TargetEnvironmentName       int
	ChartReferenceTemplate      string
	ChartName                   string
	ChartVersion                string
	ChartLocation               string
	RepoUrl                     string
	IsCustomGitRepository       bool
	GitOpsRepoMigrationRequired bool
	BuiltChartPath              string
	BuiltChartBytes             *[]byte
	MergedValues                string
}

type ManifestPushResponse struct {
	OverRiddenRepoUrl string
	CommitHash        string
	CommitTime        time.Time
	Error             error
}

type HelmRepositoryConfig struct {
	repositoryName        string
	containerRegistryName string
}

type GitRepositoryConfig struct {
	repositoryName string
}
