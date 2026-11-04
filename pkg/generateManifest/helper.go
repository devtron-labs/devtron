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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devtron-labs/common-lib/utils/yaml"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/maps"
	k8sYaml "sigs.k8s.io/yaml"
)

const (
	releaseValuesYamlFile       = "release-values.yaml"
	releaseValuesYmlFile        = "release-values.yml"
	imageDescriptorTemplateFile = ".image_descriptor_template.json"
)

// readReleaseValuesJsonFromRefChart looks for release-values.yaml / release-values.yml
// at the root of the reference chart directory and returns its content converted to
// JSON. Returns an empty string (without erroring the caller) when the file is absent
// or unreadable — not every chart defines release overrides, and manifest generation
// must still proceed. This mirrors the file-discovery logic in
// ChartTemplateServiceImpl.getValues so behaviour matches the save-time path.
func (impl DeploymentTemplateServiceImpl) readReleaseValuesJsonFromRefChart(refChartPath string) string {
	data, ok := impl.readFileFromRefChart(refChartPath, releaseValuesYamlFile, releaseValuesYmlFile)
	if !ok {
		return ""
	}
	jsonBytes, err := k8sYaml.YAMLToJSON(data)
	if err != nil {
		impl.Logger.Errorw("error in converting release-values yaml to json", "refChartPath", refChartPath, "err", err)
		return ""
	}
	return string(jsonBytes)
}

// readImageDescriptorTemplateFromRefChart reads .image_descriptor_template.json from
// the reference chart directory. Used when the chart hasn't been saved to the DB yet
// and chartDto.ImageDescriptorTemplate isn't available. Returns an empty string on
// missing file / read error — RenderJson handles an empty template gracefully.
func (impl DeploymentTemplateServiceImpl) readImageDescriptorTemplateFromRefChart(refChartPath string) string {
	data, ok := impl.readFileFromRefChart(refChartPath, imageDescriptorTemplateFile)
	if !ok {
		return ""
	}
	return string(data)
}

// readFileFromRefChart scans refChartPath (non-recursive) for the first file whose
// name (case-insensitive) matches one of fileNames and returns its contents. ok=false
// means the path was empty, unreadable, or the file was absent — all logged at the
// appropriate level by the caller's context.
func (impl DeploymentTemplateServiceImpl) readFileFromRefChart(refChartPath string, fileNames ...string) (data []byte, ok bool) {
	if len(refChartPath) == 0 {
		return nil, false
	}
	entries, err := os.ReadDir(refChartPath)
	if err != nil {
		impl.Logger.Errorw("error in reading ref chart dir", "refChartPath", refChartPath, "err", err)
		return nil, false
	}
	targets := make(map[string]struct{}, len(fileNames))
	for _, name := range fileNames {
		targets[strings.ToLower(name)] = struct{}{}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if _, match := targets[strings.ToLower(entry.Name())]; !match {
			continue
		}
		path := filepath.Clean(filepath.Join(refChartPath, entry.Name()))
		contents, err := os.ReadFile(path)
		if err != nil {
			impl.Logger.Errorw("error in reading file from ref chart", "path", path, "err", err)
			return nil, false
		}
		return contents, true
	}
	return nil, false
}

// resolveImageDescriptorTemplate returns the image descriptor template for the app's
// chart. It prefers the saved chart in DB; if no chart has been saved for this
// (appId, chartRefId) pair (pg.ErrNoRows), it falls back to the reference chart's
// .image_descriptor_template.json on disk so manifest rendering works before save.
// Any other DB error is returned to the caller.
func (impl DeploymentTemplateServiceImpl) resolveImageDescriptorTemplate(appId, chartRefId int, refChartPath string) (string, error) {
	chartDto, err := impl.chartReadService.GetByAppIdAndChartRefId(appId, chartRefId)
	if err == nil {
		return chartDto.ImageDescriptorTemplate, nil
	}
	if !errors.Is(err, pg.ErrNoRows) {
		impl.Logger.Errorw("error in getting chart by appId and chartRefId", "appId", appId, "chartRefId", chartRefId, "err", err)
		return "", err
	}
	return impl.readImageDescriptorTemplateFromRefChart(refChartPath), nil
}

// mergeReleaseOverrideIntoValuesYaml merges the chart's ReleaseOverride JSON
// on top of the given values YAML. On any failure the input valuesYaml is
// returned unchanged so template rendering can still proceed.
func (impl DeploymentTemplateServiceImpl) mergeReleaseOverrideIntoValuesYaml(valuesYaml, releaseOverrideJson string) string {
	if len(releaseOverrideJson) == 0 || releaseOverrideJson == "{}" {
		return valuesYaml
	}
	valuesJsonByte, err := k8sYaml.YAMLToJSON([]byte(valuesYaml))
	if err != nil {
		impl.Logger.Errorw("error in converting values yaml to json", "err", err)
		return valuesYaml
	}
	mergedJsonBytes, err := impl.mergeUtil.JsonPatch(valuesJsonByte, []byte(releaseOverrideJson))
	if err != nil {
		impl.Logger.Errorw("error in merging releaseOverride into values yaml", "releaseOverrideJson", releaseOverrideJson, "err", err)
		return valuesYaml
	}
	mergedYamlBytes, err := k8sYaml.JSONToYAML(mergedJsonBytes)
	if err != nil {
		impl.Logger.Errorw("error in converting merged json to yaml", "err", err)
		return valuesYaml
	}
	return string(mergedYamlBytes)
}

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

func (impl DeploymentTemplateServiceImpl) constructInstallReleaseBulkReq(apps []*app.App, environment *repository.Environment, pipelineMap map[string]*pipelineConfig.Pipeline) ([]*gRPC.InstallReleaseRequest, error) {
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
	config, err := impl.helmAppReadService.GetClusterConf(clusterBean.DefaultClusterId)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}

	for _, app := range apps {
		if _, ok := appIdToInstallReleaseRequest[app.Id]; ok {
			appIdToInstallReleaseRequest[app.Id].ReleaseIdentifier = impl.getReleaseIdentifier(config, app, environment, pipelineMap)
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

func (impl DeploymentTemplateServiceImpl) getReleaseIdentifier(config *gRPC.ClusterConfig, app *app.App, env *repository.Environment, pipelineMap map[string]*pipelineConfig.Pipeline) *gRPC.ReleaseIdentifier {
	pipeline := pipelineMap[fmt.Sprintf("%d-%d", app.Id, env.Id)]
	return &gRPC.ReleaseIdentifier{
		ClusterConfig:    config,
		ReleaseName:      pipeline.DeploymentAppName,
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
