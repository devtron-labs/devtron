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

package appStoreDeploymentFullMode

import (
	"context"
	"github.com/devtron-labs/devtron/client/argocdServer"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"

	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"time"

	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	repository2 "github.com/argoproj/argo-cd/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const (
	DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT = "devtron"
	CLUSTER_COMPONENT_DIR_PATH                  = "/cluster/component"
)

type AppStoreDeploymentFullModeService interface {
	AppStoreDeployOperationGIT(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, *util.ChartGitAttribute, error)
	AppStoreDeployOperationACD(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	RegisterInArgo(chartGitAttribute *util.ChartGitAttribute, ctx context.Context) error
	SyncACD(acdAppName string, ctx context.Context)
	UpdateValuesYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateRequirementYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error
	GetGitOpsRepoName(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (string, error)
}

type AppStoreDeploymentFullModeServiceImpl struct {
	logger                               *zap.SugaredLogger
	chartTemplateService                 util.ChartTemplateService
	refChartDir                          appStoreBean.RefChartProxyDir
	repositoryService                    repository.ServiceClient
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentRepository                repository5.EnvironmentRepository
	acdClient                            application2.ServiceClient
	ArgoK8sClient                        argocdServer.ArgoK8sClient
	gitFactory                           *util.GitFactory
	aCDAuthConfig                        *util2.ACDAuthConfig
	gitOpsRepository                     repository3.GitOpsConfigRepository
	globalEnvVariables                   *util3.GlobalEnvVariables
	installedAppRepository               repository4.InstalledAppRepository
	tokenCache                           *util2.TokenCache
}

func NewAppStoreDeploymentFullModeServiceImpl(logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService, refChartDir appStoreBean.RefChartProxyDir,
	repositoryService repository.ServiceClient,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentRepository repository5.EnvironmentRepository,
	acdClient application2.ServiceClient,
	argoK8sClient argocdServer.ArgoK8sClient,
	gitFactory *util.GitFactory, aCDAuthConfig *util2.ACDAuthConfig,
	gitOpsRepository repository3.GitOpsConfigRepository, globalEnvVariables *util3.GlobalEnvVariables,
	installedAppRepository repository4.InstalledAppRepository, tokenCache *util2.TokenCache) *AppStoreDeploymentFullModeServiceImpl {
	return &AppStoreDeploymentFullModeServiceImpl{
		logger:                               logger,
		chartTemplateService:                 chartTemplateService,
		refChartDir:                          refChartDir,
		repositoryService:                    repositoryService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
		acdClient:                            acdClient,
		ArgoK8sClient:                        argoK8sClient,
		gitFactory:                           gitFactory,
		aCDAuthConfig:                        aCDAuthConfig,
		gitOpsRepository:                     gitOpsRepository,
		globalEnvVariables:                   globalEnvVariables,
		installedAppRepository:               installedAppRepository,
		tokenCache:                           tokenCache,
	}
}

func (impl AppStoreDeploymentFullModeServiceImpl) AppStoreDeployOperationGIT(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, *util.ChartGitAttribute, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, nil, err
	}

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, nil, err
	}

	//STEP 1: Commit and PUSH on Gitlab
	template := appStoreBean.CHART_PROXY_TEMPLATE
	chartPath := path.Join(string(impl.refChartDir), template)
	valid, err := chartutil.IsChartDir(chartPath)
	if err != nil || !valid {
		impl.logger.Errorw("invalid base chart", "dir", chartPath, "err", err)
		return nil, nil, err
	}
	chartMeta := &chart.Metadata{
		Name:    appStoreAppVersion.AppStore.Name,
		Version: "1.0.1",
	}
	_, chartGitAttr, err := impl.chartTemplateService.CreateChartProxy(chartMeta, chartPath, template, appStoreAppVersion.Version, environment.Name, installAppVersionRequest)
	if err != nil {
		return nil, nil, err
	}

	//STEP 3 - update requirements and values

	//update requirements yaml in chart
	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	dependency := appStoreBean.Dependency{
		Name:       appStoreAppVersion.AppStore.Name,
		Version:    appStoreAppVersion.Version,
		Repository: appStoreAppVersion.AppStore.ChartRepo.Url,
	}
	var dependencies []appStoreBean.Dependency
	dependencies = append(dependencies, dependency)
	requirementDependencies := &appStoreBean.Dependencies{
		Dependencies: dependencies,
	}
	requirementDependenciesByte, err := json.Marshal(requirementDependencies)
	if err != nil {
		return nil, nil, err
	}
	requirementDependenciesByte, err = yaml.JSONToYAML(requirementDependenciesByte)
	if err != nil {
		return nil, nil, err
	}

	gitOpsRepoName := impl.chartTemplateService.GetGitOpsRepoName(chartMeta.Name)
	requirmentYamlConfig := &util.ChartConfig{
		FileName:       appStoreBean.REQUIREMENTS_YAML_FILE,
		FileContent:    string(requirementDependenciesByte),
		ChartName:      chartMeta.Name,
		ChartLocation:  argocdAppName,
		ChartRepoName:  gitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
	}
	gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			return nil, nil, err
		}
	}
	_, err = impl.gitFactory.Client.CommitValues(requirmentYamlConfig, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return nil, nil, err
	}

	//GIT PULL
	space := regexp.MustCompile(`\s+`)
	appStoreName := space.ReplaceAllString(chartMeta.Name, "-")
	clonedDir := impl.gitFactory.GitWorkingDir + "" + appStoreName
	err = impl.chartTemplateService.GitPull(clonedDir, chartGitAttr.RepoUrl, appStoreName)
	if err != nil {
		impl.logger.Errorw("error in git pull", "err", err)
		return nil, nil, err
	}

	//update values yaml in chart
	ValuesOverrideByte, err := yaml.YAMLToJSON([]byte(installAppVersionRequest.ValuesOverrideYaml))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, nil, err
	}

	var dat map[string]interface{}
	err = json.Unmarshal(ValuesOverrideByte, &dat)

	valuesMap := make(map[string]map[string]interface{})
	valuesMap[chartMeta.Name] = dat
	valuesByte, err := json.Marshal(valuesMap)
	if err != nil {
		impl.logger.Errorw("error in marshaling", "err", err)
		return nil, nil, err
	}

	valuesYamlConfig := &util.ChartConfig{
		FileName:       appStoreBean.VALUES_YAML_FILE,
		FileContent:    string(valuesByte),
		ChartName:      chartMeta.Name,
		ChartLocation:  argocdAppName,
		ChartRepoName:  gitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
	}
	commitHash, err := impl.gitFactory.Client.CommitValues(valuesYamlConfig, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return nil, nil, err
	}
	//sync local dir with remote
	err = impl.chartTemplateService.GitPull(clonedDir, chartGitAttr.RepoUrl, appStoreName)
	if err != nil {
		impl.logger.Errorw("error in git pull", "err", err)
		return nil, nil, err
	}
	installAppVersionRequest.GitHash = commitHash
	installAppVersionRequest.ACDAppName = argocdAppName
	installAppVersionRequest.Environment = environment
	return installAppVersionRequest, chartGitAttr, nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) AppStoreDeployOperationACD(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//STEP 4: registerInArgo
	err := impl.RegisterInArgo(chartGitAttr, ctx)
	if err != nil {
		impl.logger.Errorw("error in argo registry", "err", err)
		return nil, err
	}
	//STEP 5: createInArgo
	err = impl.createInArgo(chartGitAttr, ctx, *installAppVersionRequest.Environment, installAppVersionRequest.ACDAppName)
	if err != nil {
		impl.logger.Errorw("error in create in argo", "err", err)
		return nil, err
	}
	//STEP 6: Force Sync ACD - works like trigger deployment
	impl.SyncACD(installAppVersionRequest.ACDAppName, ctx)

	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) RegisterInArgo(chartGitAttribute *util.ChartGitAttribute, ctx context.Context) error {
	repo := &v1alpha1.Repository{
		Repo: chartGitAttribute.RepoUrl,
	}
	repo, err := impl.repositoryService.Create(ctx, &repository2.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository ", "err", err)
	}
	impl.logger.Debugw("repo registered in argo", "name", chartGitAttribute.RepoUrl)
	return err
}

func (impl AppStoreDeploymentFullModeServiceImpl) SyncACD(acdAppName string, ctx context.Context) {
	req := new(application.ApplicationSyncRequest)
	req.Name = &acdAppName
	if ctx == nil {
		impl.logger.Errorw("err in syncing ACD for AppStore, ctx is NULL", "acdAppName", acdAppName)
		return
	}
	if _, err := impl.acdClient.Sync(ctx, req); err != nil {
		impl.logger.Errorw("err in syncing ACD for AppStore", "acdAppName", acdAppName, "err", err)
	}
}

func (impl AppStoreDeploymentFullModeServiceImpl) createInArgo(chartGitAttribute *util.ChartGitAttribute, ctx context.Context, envModel repository5.Environment, argocdAppName string) error {
	appNamespace := envModel.Namespace
	if appNamespace == "" {
		appNamespace = "default"
	}

	appreq := &argocdServer.AppTemplate{
		ApplicationName: argocdAppName,
		Namespace:       impl.aCDAuthConfig.ACDConfigMapNamespace,
		TargetNamespace: appNamespace,
		TargetServer:    envModel.Cluster.ServerUrl,
		Project:         "default",
		ValuesFile:      fmt.Sprintf("values.yaml"),
		RepoPath:        chartGitAttribute.ChartLocation,
		RepoUrl:         chartGitAttribute.RepoUrl,
	}
	_, err := impl.ArgoK8sClient.CreateAcdApp(appreq, envModel.Cluster)
	//create
	if err != nil {
		impl.logger.Errorw("error in creating argo cd app ", "err", err)
		return err
	}

	return nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) GetGitOpsRepoName(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (string, error) {
	gitOpsRepoName := ""
	ctx, err := impl.tokenCache.BuildACDSynchContext()
	if err != nil {
		impl.logger.Errorw("error in creating acd sync context", "err", err)
		return "", err
	}
	acdAppName := fmt.Sprintf("%s-%s", installAppVersionRequest.AppName, installAppVersionRequest.EnvironmentName)
	application, err := impl.acdClient.Get(ctx, &application.ApplicationQuery{Name: &acdAppName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "acdAppName", acdAppName, "err", err)
		return "", err
	}
	if application != nil {
		gitOpsRepoUrl := application.Spec.Source.RepoURL
		gitOpsRepoName = impl.chartTemplateService.GetGitOpsRepoNameFromUrl(gitOpsRepoUrl)
	}
	return gitOpsRepoName, nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) UpdateValuesYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {
	acdAppName := fmt.Sprintf("%s-%s", installAppVersionRequest.AppName, installAppVersionRequest.EnvironmentName)
	if len(installAppVersionRequest.GitOpsRepoName) == 0 {
		gitOpsRepoName, err := impl.GetGitOpsRepoName(installAppVersionRequest)
		if err != nil {
			return installAppVersionRequest, err
		}
		installAppVersionRequest.GitOpsRepoName = gitOpsRepoName
	}
	valuesOverrideByte, err := yaml.YAMLToJSON([]byte(installAppVersionRequest.ValuesOverrideYaml))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return installAppVersionRequest, err
	}
	var dat map[string]interface{}
	err = json.Unmarshal(valuesOverrideByte, &dat)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return installAppVersionRequest, err
	}
	valuesMap := make(map[string]map[string]interface{})
	valuesMap[installAppVersionRequest.AppStoreName] = dat
	valuesByte, err := json.Marshal(valuesMap)
	if err != nil {
		impl.logger.Errorw("error in marshaling", "err", err)
		return installAppVersionRequest, err
	}
	valuesConfig := &util.ChartConfig{
		FileName:       appStoreBean.VALUES_YAML_FILE,
		FileContent:    string(valuesByte),
		ChartName:      installAppVersionRequest.AppStoreName,
		ChartLocation:  acdAppName,
		ChartRepoName:  installAppVersionRequest.GitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", installAppVersionRequest.AppStoreVersion, installAppVersionRequest.EnvironmentId),
	}
	gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			impl.logger.Errorw("error in fetching gitOps bitbucket config", "err", err)
			return installAppVersionRequest, err
		}
	}
	commitHash, err := impl.gitFactory.Client.CommitValues(valuesConfig, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return installAppVersionRequest, err
	}
	installAppVersionRequest.GitHash = commitHash
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) UpdateRequirementYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error {
	acdAppName := fmt.Sprintf("%s-%s", installAppVersionRequest.AppName, installAppVersionRequest.EnvironmentName)
	dependency := appStoreBean.Dependency{
		Name:       appStoreAppVersion.AppStore.Name,
		Version:    appStoreAppVersion.Version,
		Repository: appStoreAppVersion.AppStore.ChartRepo.Url,
	}
	var dependencies []appStoreBean.Dependency
	dependencies = append(dependencies, dependency)
	requirementDependencies := &appStoreBean.Dependencies{
		Dependencies: dependencies,
	}
	requirementDependenciesByte, err := json.Marshal(requirementDependencies)
	if err != nil {
		return err
	}
	requirementDependenciesByte, err = yaml.JSONToYAML(requirementDependenciesByte)
	if err != nil {
		return err
	}
	requirmentYamlConfig := &util.ChartConfig{
		FileName:       appStoreBean.REQUIREMENTS_YAML_FILE,
		FileContent:    string(requirementDependenciesByte),
		ChartName:      appStoreAppVersion.AppStore.Name,
		ChartLocation:  acdAppName,
		ChartRepoName:  installAppVersionRequest.GitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, installAppVersionRequest.EnvironmentId),
	}
	gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			impl.logger.Errorw("error in fetching gitOps bitbucket config", "err", err)
			return err
		}
	}
	_, err = impl.gitFactory.Client.CommitValues(requirmentYamlConfig, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return err
	}
	return nil
}
