/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package telemetry

import "time"

type TelemetryMetaInfo struct {
	Url    string `json:"url,omitempty"`
	UCID   string `json:"ucid,omitempty"`
	ApiKey string `json:"apiKey,omitempty"`
}

const (
	SkippedOnboardingConst = "SkippedOnboarding"
	AdminEmailIdConst      = "admin"
)

type TelemetryEventType string

const (
	Heartbeat                    TelemetryEventType = "Heartbeat"
	InstallationStart            TelemetryEventType = "InstallationStart"
	InstallationInProgress       TelemetryEventType = "InstallationInProgress"
	InstallationInterrupt        TelemetryEventType = "InstallationInterrupt"
	InstallationSuccess          TelemetryEventType = "InstallationSuccess"
	InstallationFailure          TelemetryEventType = "InstallationFailure"
	UpgradeStart                 TelemetryEventType = "UpgradeStart"
	UpgradeInProgress            TelemetryEventType = "UpgradeInProgress"
	UpgradeInterrupt             TelemetryEventType = "UpgradeInterrupt"
	UpgradeSuccess               TelemetryEventType = "UpgradeSuccess"
	UpgradeFailure               TelemetryEventType = "UpgradeFailure"
	Summary                      TelemetryEventType = "Summary"
	InstallationApplicationError TelemetryEventType = "InstallationApplicationError"
	DashboardAccessed            TelemetryEventType = "DashboardAccessed"
	DashboardLoggedIn            TelemetryEventType = "DashboardLoggedIn"
	SIG_TERM                     TelemetryEventType = "SIG_TERM"
)

type TelemetryEventEA struct {
	UCID                               string             `json:"ucid"` //unique client id
	Timestamp                          time.Time          `json:"timestamp"`
	EventMessage                       string             `json:"eventMessage,omitempty"`
	EventType                          TelemetryEventType `json:"eventType"`
	ServerVersion                      string             `json:"serverVersion,omitempty"`
	UserCount                          int                `json:"userCount,omitempty"`
	ClusterCount                       int                `json:"clusterCount,omitempty"`
	HostURL                            bool               `json:"hostURL,omitempty"`
	SSOLogin                           bool               `json:"ssoLogin,omitempty"`
	DevtronVersion                     string             `json:"devtronVersion,omitempty"`
	DevtronMode                        string             `json:"devtronMode,omitempty"`
	InstalledIntegrations              []string           `json:"installedIntegrations,omitempty"`
	InstallFailedIntegrations          []string           `json:"installFailedIntegrations,omitempty"`
	InstallTimedOutIntegrations        []string           `json:"installTimedOutIntegrations,omitempty"`
	LastLoginTime                      time.Time          `json:"LastLoginTime,omitempty"`
	InstallingIntegrations             []string           `json:"installingIntegrations,omitempty"`
	DevtronReleaseVersion              string             `json:"devtronReleaseVersion,omitempty"`
	HelmAppAccessCounter               string             `json:"HelmAppAccessCounter,omitempty"`
	HelmAppUpdateCounter               string             `json:"HelmAppUpdateCounter,omitempty"`
	ChartStoreVisitCount               string             `json:"ChartStoreVisitCount,omitempty"`
	SkippedOnboarding                  bool               `json:"SkippedOnboarding"`
	HelmChartSuccessfulDeploymentCount int                `json:"helmChartSuccessfulDeploymentCount,omitempty"`
	ExternalHelmAppClusterCount        map[int32]int      `json:"ExternalHelmAppClusterCount,omitempty"`
	ClusterProvider                    string             `json:"clusterProvider,omitempty"`
	// New telemetry fields
	HelmAppCount          int `json:"helmAppCount,omitempty"`
	PhysicalClusterCount  int `json:"physicalClusterCount,omitempty"`
	IsolatedClusterCount  int `json:"isolatedClusterCount,omitempty"`
	ActiveUsersLast30Days int `json:"activeUsersLast30Days,omitempty"`
}

const AppsCount int = 50

type TelemetryEventDto struct {
	UCID                                 string             `json:"ucid"` //unique client id
	Timestamp                            time.Time          `json:"timestamp"`
	EventMessage                         string             `json:"eventMessage,omitempty"`
	EventType                            TelemetryEventType `json:"eventType"`
	ProdAppCount                         int                `json:"prodAppCount,omitempty"`
	NonProdAppCount                      int                `json:"nonProdAppCount,omitempty"`
	UserCount                            int                `json:"userCount,omitempty"`
	EnvironmentCount                     int                `json:"environmentCount,omitempty"`
	ClusterCount                         int                `json:"clusterCount,omitempty"`
	CiCreatedPerDay                      int                `json:"ciCreatedPerDay"`
	CdCreatedPerDay                      int                `json:"cdCreatedPerDay"`
	CiDeletedPerDay                      int                `json:"ciDeletedPerDay"`
	CdDeletedPerDay                      int                `json:"cdDeletedPerDay"`
	CiTriggeredPerDay                    int                `json:"ciTriggeredPerDay"`
	CdTriggeredPerDay                    int                `json:"cdTriggeredPerDay"`
	HelmChartCount                       int                `json:"helmChartCount,omitempty"`
	SecurityScanCountPerDay              int                `json:"securityScanCountPerDay,omitempty"`
	GitAccountsCount                     int                `json:"gitAccountsCount,omitempty"`
	GitOpsCount                          int                `json:"gitOpsCount,omitempty"`
	RegistryCount                        int                `json:"registryCount,omitempty"`
	HostURL                              bool               `json:"hostURL,omitempty"`
	SSOLogin                             bool               `json:"ssoLogin,omitempty"`
	AppCount                             int                `json:"appCount,omitempty"`
	AppsWithGitRepoConfigured            int                `json:"appsWithGitRepoConfigured,omitempty"`
	AppsWithDockerConfigured             int                `json:"appsWithDockerConfigured,omitempty"`
	AppsWithDeploymentTemplateConfigured int                `json:"appsWithDeploymentTemplateConfigured,omitempty"`
	AppsWithCiPipelineConfigured         int                `json:"appsWithCiPipelineConfigured,omitempty"`
	AppsWithCdPipelineConfigured         int                `json:"appsWithCdPipelineConfigured,omitempty"`
	Build                                bool               `json:"build,omitempty"`
	Deployment                           bool               `json:"deployment,omitempty"`
	ServerVersion                        string             `json:"serverVersion,omitempty"`
	DevtronGitVersion                    string             `json:"devtronGitVersion,omitempty"`
	DevtronVersion                       string             `json:"devtronVersion,omitempty"`
	DevtronMode                          string             `json:"devtronMode,omitempty"`
	InstalledIntegrations                []string           `json:"installedIntegrations,omitempty"`
	InstallFailedIntegrations            []string           `json:"installFailedIntegrations,omitempty"`
	InstallTimedOutIntegrations          []string           `json:"installTimedOutIntegrations,omitempty"`
	InstallingIntegrations               []string           `json:"installingIntegrations,omitempty"`
	DevtronReleaseVersion                string             `json:"devtronReleaseVersion,omitempty"`
	LastLoginTime                        time.Time          `json:"LastLoginTime,omitempty"`
	SelfDockerfileCount                  int                `json:"selfDockerfileCount"`
	ManagedDockerfileCount               int                `json:"managedDockerfileCount"`
	BuildPackCount                       int                `json:"buildPackCount"`
	SelfDockerfileSuccessCount           int                `json:"selfDockerfileSuccessCount"`
	SelfDockerfileFailureCount           int                `json:"selfDockerfileFailureCount"`
	ManagedDockerfileSuccessCount        int                `json:"managedDockerfileSuccessCount"`
	ManagedDockerfileFailureCount        int                `json:"managedDockerfileFailureCount"`
	BuildPackSuccessCount                int                `json:"buildPackSuccessCount"`
	BuildPackFailureCount                int                `json:"buildPackFailureCount"`
	HelmAppAccessCounter                 string             `json:"HelmAppAccessCounter,omitempty"`
	ChartStoreVisitCount                 string             `json:"ChartStoreVisitCount,omitempty"`
	SkippedOnboarding                    bool               `json:"SkippedOnboarding"`
	HelmAppUpdateCounter                 string             `json:"HelmAppUpdateCounter,omitempty"`
	HelmChartSuccessfulDeploymentCount   int                `json:"helmChartSuccessfulDeploymentCount,omitempty"`
	ExternalHelmAppClusterCount          map[int32]int      `json:"ExternalHelmAppClusterCount"`
	ClusterProvider                      string             `json:"clusterProvider,omitempty"`
	// New telemetry fields
	HelmAppCount                int `json:"helmAppCount,omitempty"`
	DevtronAppCount             int `json:"devtronAppCount,omitempty"`
	JobCount                    int `json:"jobCount,omitempty"`
	JobPipelineCount            int `json:"jobPipelineCount,omitempty"`
	JobPipelineTriggeredLast24h int `json:"jobPipelineTriggeredLast24h,omitempty"`
	JobPipelineSucceededLast24h int `json:"jobPipelineSucceededLast24h,omitempty"`
	UserCreatedPluginCount      int `json:"userCreatedPluginCount,omitempty"`
	PolicyCount                 int `json:"policyCount,omitempty"`
	AppliedPolicyRowCount       int `json:"appliedPolicyRowCount,omitempty"`
	PhysicalClusterCount        int `json:"physicalClusterCount,omitempty"`
	IsolatedClusterCount        int `json:"isolatedClusterCount,omitempty"`
	ActiveUsersLast30Days       int `json:"activeUsersLast30Days,omitempty"`
	GitOpsPipelineCount         int `json:"gitOpsPipelineCount,omitempty"`
	HelmPipelineCount           int `json:"helmPipelineCount,omitempty"`
}
