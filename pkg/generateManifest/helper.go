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
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/yaml"
	"github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/maps"
)

func (impl DeploymentTemplateServiceImpl) constructRotatePodResponse(templateChartResponse []*gRPC.TemplateChartResponse, appNameToId map[string]int, environment *repository.Environment) (*RestartPodResponse, error) {
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
			switch gvk.Kind {
			case string(Deployment), string(StatefulSet), string(DemonSet), string(Rollout):
				resourceMeta = append(resourceMeta, &ResourceMetadata{
					Name:             name,
					GroupVersionKind: gvk,
				})
			}
		}

		appIdToResourceIdentifier[appNameToId[tcResp.AppName]] = &ResourceIdentifierResponse{
			ResourceMetaData: resourceMeta,
			AppName:          appName,
		}
	}
	podResp := &RestartPodResponse{
		EnvironmentId: environment.Id,
		Namespace:     environment.Namespace,
		RestartPodMap: appIdToResourceIdentifier,
	}
	for name, id := range appNameToId {
		if _, ok := appIdToResourceIdentifier[id]; !ok {
			appIdToResourceIdentifier[id] = &ResourceIdentifierResponse{AppName: name}
		}
	}
	return podResp, nil
}

func (impl DeploymentTemplateServiceImpl) constructInstallReleaseBulkReq(apps []*app.App, environment *repository.Environment) ([]*gRPC.InstallReleaseRequest, error) {
	appIdToInstallReleaseRequest := make(map[int]*gRPC.InstallReleaseRequest)
	installReleaseRequests := make([]*gRPC.InstallReleaseRequest, 0)
	var applicationIds []int
	for _, application := range apps {
		applicationIds = append(applicationIds, application.Id)
	}
	err := impl.setValuesYaml(applicationIds, environment.Id, appIdToInstallReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in setting values yaml", "appIds", applicationIds, "err", err)
		return nil, err
	}
	applicationIds = []int{}
	for appId, _ := range appIdToInstallReleaseRequest {
		applicationIds = append(applicationIds, appId)
	}
	if appIdToInstallReleaseRequest == nil || len(appIdToInstallReleaseRequest) == 0 {
		return nil, err
	}

	for _, app := range apps {
		if _, ok := appIdToInstallReleaseRequest[app.Id]; ok {
			appIdToInstallReleaseRequest[app.Id].AppName = app.AppName
			appIdToInstallReleaseRequest[app.Id].ChartRepository = ChartRepository
		}
	}

	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.Logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}
	config, err := impl.helmAppService.GetClusterConf(bean.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}
	for _, app := range apps {
		if _, ok := appIdToInstallReleaseRequest[app.Id]; ok {
			appIdToInstallReleaseRequest[app.Id].ReleaseIdentifier = impl.getReleaseIdentifier(config, app, environment)
			appIdToInstallReleaseRequest[app.Id].K8SVersion = k8sServerVersion.String()
		}
	}

	for _, req := range appIdToInstallReleaseRequest {
		installReleaseRequests = append(installReleaseRequests, req)
	}
	return installReleaseRequests, nil
}

func (impl DeploymentTemplateServiceImpl) setChartContent(ctx context.Context, installReleaseRequests []*gRPC.InstallReleaseRequest, appNameToId map[string]int) error {
	appIdToInstallReleaseRequest := make(map[int]*gRPC.InstallReleaseRequest)
	requestAppNameToId := make(map[int]string)
	for _, req := range installReleaseRequests {
		appId := appNameToId[req.AppName]
		requestAppNameToId[appId] = req.AppName
		appIdToInstallReleaseRequest[appId] = req
	}

	_, span := otel.Tracer("orchestrator").Start(ctx, "chartRefService.GetChartBytesForApps")
	appIdToBytes, err := impl.chartRefService.GetChartBytesForApps(ctx, requestAppNameToId)
	span.End()
	if err != nil {
		impl.Logger.Errorw("error in fetching chartRefBean", "err", err, "appNames", maps.Keys(appNameToId))
		return err
	}
	for appId, _ := range appIdToInstallReleaseRequest {
		appIdToInstallReleaseRequest[appId].ChartContent = &gRPC.ChartContent{Content: appIdToBytes[appId]}
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

func (impl DeploymentTemplateServiceImpl) setValuesYaml(appIds []int, envId int, appIdToInstallReleaseRequest map[int]*gRPC.InstallReleaseRequest) error {
	pipelineOverrides, err := impl.pipelineOverrideRepository.GetLatestReleaseForAppIds(appIds, envId)
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelineOverrides for appIds", "appIds", appIds, "err", err)
		return err
	}
	for _, pco := range pipelineOverrides {
		appIdToInstallReleaseRequest[pco.AppId] = &gRPC.InstallReleaseRequest{ValuesYaml: pco.MergedValuesYaml}
	}
	return err
}
