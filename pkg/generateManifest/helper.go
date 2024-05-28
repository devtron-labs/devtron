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

package generateManifest

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/yaml"
	"github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
)

func (impl DeploymentTemplateServiceImpl) constructRotatePodResponse(templateChartResponse []*gRPC.TemplateChartResponse, appNameToId map[string]int, appIdToUserApprovalConfig map[int]string, environment *repository.Environment) (*RestartPodResponse, error) {
	appIdToResourceIdentifier := make(map[int]*ResourceIdentifierResponse)
	for _, tcResp := range templateChartResponse {
		manifests, err := yamlUtil.SplitYAMLs([]byte(tcResp.GeneratedManifest))
		if err != nil {
			return nil, err
		}
		appName := tcResp.AppName
		resourceMeta := make([]*ResourceMetadata, 0)
		for _, manifest := range manifests {
			gvk := manifest.GroupVersionKind()
			name := manifest.GetName()
			labels := manifest.GetLabels()
			switch gvk.Kind {
			case string(Deployment), string(StatefulSet), string(DemonSet), string(Rollout):
				if labels != nil && labels[LabelReleaseKey] != "" {
					resourceMeta = append(resourceMeta, &ResourceMetadata{
						Name:             labels[LabelReleaseKey],
						GroupVersionKind: gvk,
					})
				} else {
					resourceMeta = append(resourceMeta, &ResourceMetadata{
						Name:             name,
						GroupVersionKind: gvk,
					})
				}

			}
		}
		appIdToResourceIdentifier[appNameToId[tcResp.AppName]] = &ResourceIdentifierResponse{
			ResourceMetaData: resourceMeta,
			AppName:          appName,
		}
	}
	appIdToName := make(map[int]string)

	for name, id := range appNameToId {
		appIdToName[id] = name
	}
	for id, approvalConfig := range appIdToUserApprovalConfig {
		appIdToResourceIdentifier[id] = &ResourceIdentifierResponse{AppName: appIdToName[id], ErrorResponse: approvalConfig}
	}
	for name, id := range appNameToId {
		if _, ok := appIdToResourceIdentifier[id]; !ok {
			appIdToResourceIdentifier[id] = &ResourceIdentifierResponse{AppName: name}
		}
	}

	podResp := &RestartPodResponse{
		EnvironmentId: environment.Id,
		Namespace:     environment.Namespace,
		RestartPodMap: appIdToResourceIdentifier,
	}
	return podResp, nil
}

func (impl DeploymentTemplateServiceImpl) constructInstallReleaseBulkReq(apps []*app.App, environment *repository.Environment) ([]*gRPC.InstallReleaseRequest, map[int]string, error) {
	appIdToInstallReleaseRequest := make(map[int]*gRPC.InstallReleaseRequest)
	installReleaseRequests := make([]*gRPC.InstallReleaseRequest, 0)
	var applicationIds []int
	for _, application := range apps {
		applicationIds = append(applicationIds, application.Id)
	}
	appIdToUserApprovalConfig, err := impl.setValuesYaml(applicationIds, environment.Id, appIdToInstallReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in setting values yaml", "appIds", applicationIds, "err", err)
		return nil, nil, err
	}
	applicationIds = []int{}
	for appId, _ := range appIdToInstallReleaseRequest {
		applicationIds = append(applicationIds, appId)
	}
	if appIdToInstallReleaseRequest == nil || len(appIdToInstallReleaseRequest) == 0 {
		return nil, appIdToUserApprovalConfig, err
	}
	err = impl.setChartContent(applicationIds, appIdToInstallReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in setting chart content", "appIds", applicationIds, "err", err)
		return nil, nil, err
	}
	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.Logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, nil, err
	}
	config, err := impl.helmAppService.GetClusterConf(bean.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, nil, err
	}
	for _, app := range apps {
		if _, ok := appIdToInstallReleaseRequest[app.Id]; ok {
			appIdToInstallReleaseRequest[app.Id].ReleaseIdentifier = impl.getReleaseIdentifier(config, app, environment)
			appIdToInstallReleaseRequest[app.Id].K8SVersion = k8sServerVersion.String()
			appIdToInstallReleaseRequest[app.Id].AppName = app.AppName
			appIdToInstallReleaseRequest[app.Id].ChartRepository = ChartRepository
		}
	}

	for _, req := range appIdToInstallReleaseRequest {
		installReleaseRequests = append(installReleaseRequests, req)
	}
	return installReleaseRequests, appIdToUserApprovalConfig, nil
}

func (impl DeploymentTemplateServiceImpl) setChartContent(appIds []int, appIdToInstallReleaseRequest map[int]*gRPC.InstallReleaseRequest) error {
	charts, err := impl.chartRepository.FindLatestChartByAppIds(appIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching chart", "err", err, "appIds", appIds)
		return err
	}
	appIdToChartRefId := make(map[int]int)
	var chartRefIds []int
	for _, chart := range charts {
		appIdToChartRefId[chart.AppId] = chart.ChartRefId
		chartRefIds = append(chartRefIds, chart.ChartRefId)
	}
	chartRefIdToBytes, err := impl.chartRefService.GetChartBytesInBulk(chartRefIds, false)
	if err != nil {
		impl.Logger.Errorw("error in fetching chartRefBean", "err", err, "chartRefIds", chartRefIds)
		return err
	}
	for appId, chartRefId := range appIdToChartRefId {
		if bytes, ok := chartRefIdToBytes[chartRefId]; ok {
			if _, ok := appIdToInstallReleaseRequest[appId]; ok {
				appIdToInstallReleaseRequest[appId].ChartContent = &gRPC.ChartContent{Content: bytes}
			}
		}
	}
	return err
}

func (impl DeploymentTemplateServiceImpl) getReleaseIdentifier(config *gRPC.ClusterConfig, app *app.App, env *repository.Environment) *gRPC.ReleaseIdentifier {
	return &gRPC.ReleaseIdentifier{
		ClusterConfig:    config,
		ReleaseName:      fmt.Sprintf("%s-%s", app.AppName, env.Name),
		ReleaseNamespace: env.Namespace,
	}
}

func (impl DeploymentTemplateServiceImpl) setValuesYaml(appIds []int, envId int, appIdToInstallReleaseRequest map[int]*gRPC.InstallReleaseRequest) (map[int]string, error) {
	appIdToUserApprovalConfig := make(map[int]string)
	pipelineOverrides, err := impl.pipelineOverrideRepository.GetLatestReleaseForAppIds(appIds, envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelineOverrides for appIds", "appIds", appIds, "err", err)
		return nil, err
	}
	for _, pco := range pipelineOverrides {
		if len(pco.UserApprovalConfig) > 0 {
			appIdToUserApprovalConfig[pco.AppId] = pco.UserApprovalConfig
		} else {
			appIdToInstallReleaseRequest[pco.AppId] = &gRPC.InstallReleaseRequest{ValuesYaml: pco.MergedValuesYaml}
		}
	}
	return appIdToUserApprovalConfig, err
}
