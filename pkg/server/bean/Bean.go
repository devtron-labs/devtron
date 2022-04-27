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

package serverBean

type ServerInfoDto struct {
	CurrentVersion  string `json:"currentVersion,notnull"`
	Status          string `json:"status,notnull" validate:"oneof=healthy upgrading upgradeFailed unknown timeout"`
	ReleaseName     string `json:"releaseName,notnull"`
	CanUpdateServer bool   `json:"canUpdateServer,notnull"`
}

type ServerActionRequestDto struct {
	Action  string `json:"action,notnull" validate:"oneof=upgrade"`
	Version string `json:"version,notnull"`
}

type ActionResponse struct {
	Success bool `json:"success"`
}

type ServerStatus = string
type InstallerCrdObjectStatus = string

const (
	ServerStatusHealthy       ServerStatus = "healthy"
	ServerStatusUpgrading     ServerStatus = "upgrading"
	ServerStatusUpgradeFailed ServerStatus = "upgradeFailed"
	ServerStatusUnknown       ServerStatus = "unknown"
	ServerStatusTimeout       ServerStatus = "timeout"

	InstallerCrdObjectStatusBlank      InstallerCrdObjectStatus = ""
	InstallerCrdObjectStatusDownloaded InstallerCrdObjectStatus = "Downloaded"
	InstallerCrdObjectStatusApplied    InstallerCrdObjectStatus = "Applied"
)

type HelmReleaseStatus = string

// Describe the status of a release
// NOTE: Make sure to update cmd/helm/status.go when adding or modifying any of these statuses.
const (
	// HelmReleaseStatusUnknown indicates that a release is in an uncertain state.
	HelmReleaseStatusUnknown HelmReleaseStatus = "unknown"
	// HelmReleaseStatusDeployed indicates that the release has been pushed to Kubernetes.
	HelmReleaseStatusDeployed HelmReleaseStatus = "deployed"
	// HelmReleaseStatusUninstalled indicates that a release has been uninstalled from Kubernetes.
	HelmReleaseStatusUninstalled HelmReleaseStatus = "uninstalled"
	// HelmReleaseStatusSuperseded indicates that this release object is outdated and a newer one exists.
	HelmReleaseStatusSuperseded HelmReleaseStatus = "superseded"
	// HelmReleaseStatusFailed indicates that the release was not successfully deployed.
	HelmReleaseStatusFailed HelmReleaseStatus = "failed"
	// HelmReleaseStatusUninstalling indicates that a uninstall operation is underway.
	HelmReleaseStatusUninstalling HelmReleaseStatus = "uninstalling"
	// HelmReleaseStatusPendingInstall indicates that an install operation is underway.
	HelmReleaseStatusPendingInstall HelmReleaseStatus = "pending-install"
	// HelmReleaseStatusPendingUpgrade indicates that an upgrade operation is underway.
	HelmReleaseStatusPendingUpgrade HelmReleaseStatus = "pending-upgrade"
	// HelmReleaseStatusPendingRollback indicates that an rollback operation is underway.
	HelmReleaseStatusPendingRollback HelmReleaseStatus = "pending-rollback"
)

type AppHealthStatusCode = string

const (
	AppHealthStatusProgressing AppHealthStatusCode = "Progressing"
	AppHealthStatusDegraded    AppHealthStatusCode = "Degraded"
)
