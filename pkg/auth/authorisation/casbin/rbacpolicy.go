/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package casbin

const (
	ResourceCluster               = "cluster"
	ResourceGlobalEnvironment     = "global-environment"
	ResourceEnvironment           = "environment"
	ResourceGit                   = "git"
	ResourceDocker                = "docker"
	ResourceMigrate               = "migrate"
	ResourceUser                  = "user"
	ResourceNotification          = "notification"
	ResourceTemplate              = "template"
	ResourceTerminal              = "terminal"
	ResourceCiPipelineSourceValue = "ci-pipeline/source-value"
	ResourceConfig                = "config"
	// todo: make this resource as artifact
	ResourceArtifact = "artifact"

	ResourceProjects     = "projects"
	ResourceApplications = "applications"
	ResourceDockerAuto   = "docker-auto"
	ResourceGitAuto      = "git-auto"

	ResourceAutocomplete = "autocomplete"
	ResourceChartGroup   = "chart-group"

	ResourceTeam          = "team"
	ResourceAdmin         = "admin"
	ResourceGlobal        = "global-resource"
	ResourceHelmApp       = "helm-app"
	ActionGet             = "get"
	ActionCreate          = "create"
	ActionUpdate          = "update"
	ActionDelete          = "delete"
	ActionSync            = "sync"
	ActionTrigger         = "trigger"
	ActionNotify          = "notify"
	ActionExec            = "exec"
	ActionAllPlaceHolder  = "all"
	ActionApprove         = "approve"
	ActionArtifactPromote = "promote"

	// ResourceJobs ,ResourceJobsEnv ,ResourceWorkflow these three resources are being used in jobs for rbac.
	ResourceJobs     = "jobs"
	ResourceJobsEnv  = "jobenv"
	ResourceWorkflow = "workflow"

	ClusterResourceRegex         = "%s/%s"    // {cluster}/{namespace}
	ClusterObjectRegex           = "%s/%s/%s" // {groupName}/{kindName}/{objectName}
	ClusterEmptyGroupPlaceholder = "k8sempty"

	ResourceObjectIgnorePlaceholder = ":ignore"
)
