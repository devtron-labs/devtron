/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package casbin

const (
	ResourceCluster           = "cluster"
	ResourceGlobalEnvironment = "global-environment"
	ResourceEnvironment       = "environment"
	ResourceGit               = "git"
	ResourceDocker            = "docker"
	ResourceMigrate           = "migrate"
	ResourceUser              = "user"
	ResourceNotification      = "notification"
	ResourceTemplate          = "template"

	ResourceProjects     = "projects"
	ResourceApplications = "applications"
	ResourceDockerAuto   = "docker-auto"
	ResourceGitAuto      = "git-auto"

	ResourceAutocomplete = "autocomplete"
	ResourceChartGroup   = "chart-group"

	ResourceTeam   = "team"
	ResourceAdmin  = "admin"
	ResourceGlobal = "global-resource"

	ActionGet     = "get"
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionSync    = "sync"
	ActionTrigger = "trigger"
	ActionNotify  = "notify"
)
