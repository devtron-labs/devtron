/*
 * Copyright (c) 2024. Devtron Inc.
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

package adapter

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	"github.com/golang/protobuf/ptypes/timestamp"
	"helm.sh/helm/v3/pkg/chart"
	"path"
)

func ParseChartGitPushRequest(installAppRequestDTO *appStoreBean.InstallAppVersionDTO, tempRefChart string) *bean.PushChartToGitRequestDTO {
	return &bean.PushChartToGitRequestDTO{
		AppName:           installAppRequestDTO.AppName,
		EnvName:           installAppRequestDTO.EnvironmentName,
		ChartAppStoreName: installAppRequestDTO.AppStoreName,
		RepoURL:           installAppRequestDTO.GitOpsRepoURL,
		TempChartRefDir:   tempRefChart,
		UserId:            installAppRequestDTO.UserId,
	}
}

func ParseChartCreateRequest(appName string, includePackageChart bool) *util.ChartCreateRequest {
	chartPath := getRefProxyChartPath()
	return &util.ChartCreateRequest{
		ChartMetaData: &chart.Metadata{
			Name:    appName,
			Version: "1.0.1",
		},
		ChartPath:           chartPath,
		IncludePackageChart: includePackageChart,
	}
}

func getRefProxyChartPath() string {
	template := appStoreBean.CHART_PROXY_TEMPLATE
	return path.Join(appStoreBean.RefChartProxyDirPath, template)
}

type UpdateVersionHistoryOperation func(installedAppVersionHistory *repository.InstalledAppVersionHistory) error

// FailedStatusUpdateOption returns an UpdateVersionHistoryOperation that updates the status of the installed app version history to failed
func FailedStatusUpdateOption(userId int32, deploymentErr error) UpdateVersionHistoryOperation {
	return func(installedAppVersionHistory *repository.InstalledAppVersionHistory) error {
		if deploymentErr == nil {
			// for failed status deploymentErr should not be nil
			return nil
		}
		installedAppVersionHistory.MarkDeploymentFailed(deploymentErr)
		installedAppVersionHistory.SetFinishedOn()
		installedAppVersionHistory.UpdateAuditLog(userId)
		return nil
	}
}

// SuccessStatusUpdateOption returns an UpdateVersionHistoryOperation that updates the status of the installed app version history to success
func SuccessStatusUpdateOption(deploymentAppType string, userId int32) UpdateVersionHistoryOperation {
	return func(installedAppVersionHistory *repository.InstalledAppVersionHistory) error {
		installedAppVersionHistory.SetStatus(cdWorkflow.WorkflowSucceeded)
		installedAppVersionHistory.SetFinishedOn()
		installedAppVersionHistory.UpdateAuditLog(userId)
		// update helm release status config if helm installed app
		// for ArgoCd installed app, we don't need to update the release status config
		if util.IsHelmApp(deploymentAppType) {
			helmInstallStatus := &appStoreBean.HelmReleaseStatusConfig{
				InstallAppVersionHistoryId: installedAppVersionHistory.Id,
				Message:                    "Release Installed",
				IsReleaseInstalled:         true,
				ErrorInInstallation:        false,
			}
			data, err := json.Marshal(helmInstallStatus)
			if err != nil {
				return err
			}
			installedAppVersionHistory.HelmReleaseStatusConfig = string(data)
		}
		return nil
	}
}

func BuildDeploymentHistory(installedAppVersionModel *repository.InstalledAppVersions, sources []string, updateHistory *repository.InstalledAppVersionHistory, emailId string) *gRPC.HelmAppDeploymentDetail {
	return &gRPC.HelmAppDeploymentDetail{
		ChartMetadata: &gRPC.ChartMetadata{
			ChartName:    installedAppVersionModel.AppStoreApplicationVersion.AppStore.Name,
			ChartVersion: installedAppVersionModel.AppStoreApplicationVersion.Version,
			Description:  installedAppVersionModel.AppStoreApplicationVersion.Description,
			Home:         installedAppVersionModel.AppStoreApplicationVersion.Home,
			Sources:      sources,
		},
		DeployedBy:   emailId,
		DockerImages: []string{installedAppVersionModel.AppStoreApplicationVersion.AppVersion},
		DeployedAt: &timestamp.Timestamp{
			Seconds: updateHistory.CreatedOn.Unix(),
			Nanos:   int32(updateHistory.CreatedOn.Nanosecond()),
		},
		Version: int32(updateHistory.Id),
		Status:  updateHistory.Status,
		Message: updateHistory.Message,
	}
}
