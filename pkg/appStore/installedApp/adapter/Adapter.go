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
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
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
