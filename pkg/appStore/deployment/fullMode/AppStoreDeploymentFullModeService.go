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
	"encoding/json"
	"errors"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib-private/pubsub-lib"
	"path"
	"regexp"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/status"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	repository2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/yaml"
)

const (
	DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT = "devtron"
	CLUSTER_COMPONENT_DIR_PATH                  = "/cluster/component"
)

type AppStoreDeploymentFullModeService interface {
	AppStoreDeployOperationGIT(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, *util.ChartGitAttribute, error)
	AppStoreDeployOperationACD(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	RegisterInArgo(chartGitAttribute *util.ChartGitAttribute, ctx context.Context) error
	SyncACD(acdAppName string, ctx context.Context)
	UpdateValuesYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateRequirementYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error
	GetGitOpsRepoName(appName string, environmentName string) (string, error)
	SubscribeHelmInstallStatus() error
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
	globalEnvVariables                   *util3.GlobalEnvVariables
	installedAppRepository               repository4.InstalledAppRepository
	tokenCache                           *util2.TokenCache
	argoUserService                      argo.ArgoUserService
	gitOpsConfigRepository               repository3.GitOpsConfigRepository
	pipelineStatusTimelineService        status.PipelineStatusTimelineService
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	argoClientWrapperService             argocdServer.ArgoClientWrapperService
	pubSubClient                         *pubsub.PubSubClientServiceImpl
	installedAppRepositoryHistory        repository4.InstalledAppVersionHistoryRepository
	ACDConfig                            *argocdServer.ACDConfig
}

func NewAppStoreDeploymentFullModeServiceImpl(logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService, refChartDir appStoreBean.RefChartProxyDir,
	repositoryService repository.ServiceClient,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentRepository repository5.EnvironmentRepository,
	acdClient application2.ServiceClient,
	argoK8sClient argocdServer.ArgoK8sClient,
	gitFactory *util.GitFactory, aCDAuthConfig *util2.ACDAuthConfig,
	globalEnvVariables *util3.GlobalEnvVariables,
	installedAppRepository repository4.InstalledAppRepository, tokenCache *util2.TokenCache,
	argoUserService argo.ArgoUserService, gitOpsConfigRepository repository3.GitOpsConfigRepository,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	installedAppRepositoryHistory repository4.InstalledAppVersionHistoryRepository,
	ACDConfig *argocdServer.ACDConfig,
) *AppStoreDeploymentFullModeServiceImpl {
	appStoreDeploymentFullModeServiceImpl := &AppStoreDeploymentFullModeServiceImpl{
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
		globalEnvVariables:                   globalEnvVariables,
		installedAppRepository:               installedAppRepository,
		tokenCache:                           tokenCache,
		argoUserService:                      argoUserService,
		gitOpsConfigRepository:               gitOpsConfigRepository,
		pipelineStatusTimelineService:        pipelineStatusTimelineService,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		argoClientWrapperService:             argoClientWrapperService,
		pubSubClient:                         pubSubClient,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		ACDConfig:                            ACDConfig,
	}
	err := appStoreDeploymentFullModeServiceImpl.SubscribeHelmInstallStatus()
	if err != nil {
		return nil
	}
	return appStoreDeploymentFullModeServiceImpl
}

func (impl AppStoreDeploymentFullModeServiceImpl) AppStoreDeployOperationGIT(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, *util.ChartGitAttribute, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return installAppVersionRequest, nil, err
	}

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return installAppVersionRequest, nil, err
	}

	//STEP 1: Commit and PUSH on Gitlab
	template := appStoreBean.CHART_PROXY_TEMPLATE
	chartPath := path.Join(string(impl.refChartDir), template)
	valid, err := chartutil.IsChartDir(chartPath)
	if err != nil || !valid {
		impl.logger.Errorw("invalid base chart", "dir", chartPath, "err", err)
		return installAppVersionRequest, nil, err
	}
	chartMeta := &chart.Metadata{
		Name:    installAppVersionRequest.AppName,
		Version: "1.0.1",
	}
	_, chartGitAttr, err := impl.chartTemplateService.CreateChartProxy(chartMeta, chartPath, environment.Name, installAppVersionRequest)
	if err != nil {
		return installAppVersionRequest, nil, err
	}

	//STEP 3 - update requirements and values
	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	dependency := appStoreBean.Dependency{
		Name:    appStoreAppVersion.AppStore.Name,
		Version: appStoreAppVersion.Version,
	}
	if appStoreAppVersion.AppStore.ChartRepo != nil {
		dependency.Repository = appStoreAppVersion.AppStore.ChartRepo.Url
	}
	var dependencies []appStoreBean.Dependency
	dependencies = append(dependencies, dependency)
	requirementDependencies := &appStoreBean.Dependencies{
		Dependencies: dependencies,
	}
	requirementDependenciesByte, err := json.Marshal(requirementDependencies)
	if err != nil {
		return installAppVersionRequest, nil, err
	}
	requirementDependenciesByte, err = yaml.JSONToYAML(requirementDependenciesByte)
	if err != nil {
		return installAppVersionRequest, nil, err
	}

	gitOpsRepoName := impl.chartTemplateService.GetGitOpsRepoName(installAppVersionRequest.AppName)
	//getting username & emailId for commit author data
	userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(installAppVersionRequest.UserId)
	gitOpsConfigBitbucket, err := impl.gitOpsConfigRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			return installAppVersionRequest, nil, err
		}
	}
	requirmentYamlConfig := &util.ChartConfig{
		FileName:       appStoreBean.REQUIREMENTS_YAML_FILE,
		FileContent:    string(requirementDependenciesByte),
		ChartName:      chartMeta.Name,
		ChartLocation:  argocdAppName,
		ChartRepoName:  gitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
		UserEmailId:    userEmailId,
		UserName:       userName,
	}
	gitOpsConfig := &bean.GitOpsConfigDto{BitBucketWorkspaceId: gitOpsConfigBitbucket.BitBucketWorkspaceId}
	_, _, err = impl.gitFactory.Client.CommitValues(requirmentYamlConfig, gitOpsConfig)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return installAppVersionRequest, nil, err
	}

	//GIT PULL
	space := regexp.MustCompile(`\s+`)
	appStoreName := space.ReplaceAllString(chartMeta.Name, "-")
	clonedDir := impl.gitFactory.GitWorkingDir + "" + appStoreName
	err = impl.chartTemplateService.GitPull(clonedDir, chartGitAttr.RepoUrl, appStoreName)
	if err != nil {
		impl.logger.Errorw("error in git pull", "err", err)
		return installAppVersionRequest, nil, err
	}

	//update values yaml in chart
	ValuesOverrideByte, err := yaml.YAMLToJSON([]byte(installAppVersionRequest.ValuesOverrideYaml))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return installAppVersionRequest, nil, err
	}

	var dat map[string]interface{}
	err = json.Unmarshal(ValuesOverrideByte, &dat)

	valuesMap := make(map[string]map[string]interface{})
	valuesMap[appStoreAppVersion.AppStore.Name] = dat
	valuesByte, err := json.Marshal(valuesMap)
	if err != nil {
		impl.logger.Errorw("error in marshaling", "err", err)
		return installAppVersionRequest, nil, err
	}

	valuesYamlConfig := &util.ChartConfig{
		FileName:       appStoreBean.VALUES_YAML_FILE,
		FileContent:    string(valuesByte),
		ChartName:      chartMeta.Name,
		ChartLocation:  argocdAppName,
		ChartRepoName:  gitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
		UserEmailId:    userEmailId,
		UserName:       userName,
	}

	commitHash, _, err := impl.gitFactory.Client.CommitValues(valuesYamlConfig, gitOpsConfig)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		//update timeline status for git commit failed state
		gitCommitStatus := pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED
		gitCommitStatusDetail := fmt.Sprintf("Git commit failed - %v", err)
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
			Status:                       gitCommitStatus,
			StatusDetail:                 gitCommitStatusDetail,
			StatusTime:                   time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: installAppVersionRequest.UserId,
				CreatedOn: time.Now(),
				UpdatedBy: installAppVersionRequest.UserId,
				UpdatedOn: time.Now(),
			},
		}
		timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, true)
		if timelineErr != nil {
			impl.logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
		}
		return installAppVersionRequest, nil, err
	}
	//creating timeline for Git Commit stage
	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
		Status:                       pipelineConfig.TIMELINE_STATUS_GIT_COMMIT,
		StatusDetail:                 "Git commit done successfully.",
		StatusTime:                   time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: installAppVersionRequest.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: installAppVersionRequest.UserId,
			UpdatedOn: time.Now(),
		},
	}
	err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, true)
	if err != nil {
		impl.logger.Errorw("error in creating timeline status for git commit", "err", err, "timeline", timeline)
	}

	//sync local dir with remote
	err = impl.chartTemplateService.GitPull(clonedDir, chartGitAttr.RepoUrl, appStoreName)
	if err != nil {
		impl.logger.Errorw("error in git pull", "err", err)
		return installAppVersionRequest, nil, err
	}
	installAppVersionRequest.GitHash = commitHash
	installAppVersionRequest.ACDAppName = argocdAppName
	installAppVersionRequest.Environment = environment

	return installAppVersionRequest, chartGitAttr, nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) AppStoreDeployOperationACD(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
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
	//impl.SyncACD(installAppVersionRequest.ACDAppName, ctx)

	//STEP 7: normal refresh ACD - update for step 6 to avoid delay
	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, installAppVersionRequest.ACDAppName)
	if err != nil {
		impl.logger.Errorw("error in getting the argo application with normal refresh", "err", err)
		return nil, err
	}
	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
		Status:                       pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED,
		StatusDetail:                 "argocd sync completed.",
		StatusTime:                   time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: installAppVersionRequest.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: installAppVersionRequest.UserId,
			UpdatedOn: time.Now(),
		},
	}
	err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, true)
	if err != nil {
		impl.logger.Errorw("error in creating timeline for argocd sync", "err", err, "timeline", timeline)
	}

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
		AutoSyncEnabled: impl.ACDConfig.ArgoCDAutoSyncEnabled,
	}
	_, err := impl.ArgoK8sClient.CreateAcdApp(appreq, envModel.Cluster, argocdServer.ARGOCD_APPLICATION_TEMPLATE)
	//create
	if err != nil {
		impl.logger.Errorw("error in creating argo cd app ", "err", err)
		return err
	}
	return nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) GetGitOpsRepoName(appName string, environmentName string) (string, error) {
	gitOpsRepoName := ""
	//this method should only call in case of argo-integration and gitops configured
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return "", err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	acdAppName := fmt.Sprintf("%s-%s", appName, environmentName)
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

func (impl AppStoreDeploymentFullModeServiceImpl) UpdateValuesYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {

	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(installAppVersionRequest.AppStoreName, installAppVersionRequest.ValuesOverrideYaml)
	if err != nil {
		impl.logger.Errorw("error in getting values string", "err", err)
		return nil, err
	}

	valuesGitConfig, err := impl.appStoreDeploymentCommonService.GetGitCommitConfig(installAppVersionRequest, valuesString, appStoreBean.VALUES_YAML_FILE)
	if err != nil {
		impl.logger.Errorw("error in getting git commit config", "err", err)
	}

	commitHash, err := impl.appStoreDeploymentCommonService.CommitConfigToGit(valuesGitConfig)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return installAppVersionRequest, errors.New(pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED)
	}
	//update timeline status for git commit state
	installAppVersionRequest.GitHash = commitHash
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) UpdateRequirementYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error {

	requirementsString, err := impl.appStoreDeploymentCommonService.GetRequirementsString(appStoreAppVersion.Id)
	if err != nil {
		impl.logger.Errorw("error in getting requirements string", "err", err)
		return err
	}

	requirementsGitConfig, err := impl.appStoreDeploymentCommonService.GetGitCommitConfig(installAppVersionRequest, requirementsString, appStoreBean.REQUIREMENTS_YAML_FILE)
	if err != nil {
		impl.logger.Errorw("error in getting git commit config", "err", err)
		return err
	}

	_, err = impl.appStoreDeploymentCommonService.CommitConfigToGit(requirementsGitConfig)
	if err != nil {
		impl.logger.Errorw("error in values commit", "err", err)
		return errors.New(pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED)
	}

	return nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) SubscribeHelmInstallStatus() error {

	callback := func(msg *pubsub.PubSubMsg) {

		impl.logger.Debug("received helm install status event - HELM_INSTALL_STATUS", "data", msg.Data)
		helmInstallNatsMessage := &appStoreBean.HelmReleaseStatusConfig{}
		err := json.Unmarshal([]byte(msg.Data), helmInstallNatsMessage)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling helm install status nats message", "err", err)
			return
		}

		installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(helmInstallNatsMessage.InstallAppVersionHistoryId)
		if err != nil {
			impl.logger.Errorw("error in fetching installed app by installed app id in subscribe helm status callback", "err", err)
			return
		}
		if helmInstallNatsMessage.ErrorInInstallation {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowFailed
		} else {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
		}
		installedAppVersionHistory.HelmReleaseStatusConfig = msg.Data
		_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
		if err != nil {
			impl.logger.Errorw("error in updating helm release status data in installedAppVersionHistoryRepository", "err", err)
			return
		}
	}

	err := impl.pubSubClient.Subscribe(pubsub.HELM_CHART_INSTALL_STATUS_TOPIC, callback)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
