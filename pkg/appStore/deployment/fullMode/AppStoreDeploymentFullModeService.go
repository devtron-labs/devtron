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
	"github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/remote"
	"time"

	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/status"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"go.uber.org/zap"
)

const (
	DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT = "devtron"
	CLUSTER_COMPONENT_DIR_PATH                  = "/cluster/component"
)

// ACD operation and git operation
type AppStoreDeploymentFullModeService interface {
	AppStoreDeployOperationACD(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateValuesYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateRequirementYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error
	GetGitOpsRepoName(appName string, environmentName string) (string, error)
}

type AppStoreDeploymentFullModeServiceImpl struct {
	logger                          *zap.SugaredLogger
	acdClient                       application2.ServiceClient
	ArgoK8sClient                   argocdServer.ArgoK8sClient
	aCDAuthConfig                   *util2.ACDAuthConfig
	argoUserService                 argo.ArgoUserService
	pipelineStatusTimelineService   status.PipelineStatusTimelineService
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService
	argoClientWrapperService        argocdServer.ArgoClientWrapperService
	pubSubClient                    *pubsub_lib.PubSubClientServiceImpl
	installedAppRepositoryHistory   repository4.InstalledAppVersionHistoryRepository
	ACDConfig                       *argocdServer.ACDConfig
	gitOpsConfigReadService         config.GitOpsConfigReadService
	gitOpsRemoteOperationService    remote.GitOpsRemoteOperationService
}

func NewAppStoreDeploymentFullModeServiceImpl(logger *zap.SugaredLogger,
	acdClient application2.ServiceClient,
	argoK8sClient argocdServer.ArgoK8sClient, aCDAuthConfig *util2.ACDAuthConfig,
	argoUserService argo.ArgoUserService, pipelineStatusTimelineService status.PipelineStatusTimelineService,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	pubSubClient *pubsub_lib.PubSubClientServiceImpl,
	installedAppRepositoryHistory repository4.InstalledAppVersionHistoryRepository,
	ACDConfig *argocdServer.ACDConfig,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOpsRemoteOperationService remote.GitOpsRemoteOperationService) *AppStoreDeploymentFullModeServiceImpl {
	appStoreDeploymentFullModeServiceImpl := &AppStoreDeploymentFullModeServiceImpl{
		logger:                          logger,
		acdClient:                       acdClient,
		ArgoK8sClient:                   argoK8sClient,
		aCDAuthConfig:                   aCDAuthConfig,
		argoUserService:                 argoUserService,
		pipelineStatusTimelineService:   pipelineStatusTimelineService,
		appStoreDeploymentCommonService: appStoreDeploymentCommonService,
		argoClientWrapperService:        argoClientWrapperService,
		pubSubClient:                    pubSubClient,
		installedAppRepositoryHistory:   installedAppRepositoryHistory,
		ACDConfig:                       ACDConfig,
		gitOpsConfigReadService:         gitOpsConfigReadService,
		gitOpsRemoteOperationService:    gitOpsRemoteOperationService,
	}
	err := appStoreDeploymentFullModeServiceImpl.subscribeHelmInstallStatus()
	if err != nil {
		return nil
	}
	return appStoreDeploymentFullModeServiceImpl
}

func (impl AppStoreDeploymentFullModeServiceImpl) AppStoreDeployOperationACD(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//STEP 4: registerInArgo
	err := impl.argoClientWrapperService.RegisterGitOpsRepoInArgo(ctx, chartGitAttr.RepoUrl)
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
	syncTime := time.Now()
	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, installAppVersionRequest.ACDAppName)
	if err != nil {
		impl.logger.Errorw("error in getting the argo application with normal refresh", "err", err)
		return nil, err
	}
	if !impl.ACDConfig.ArgoCDAutoSyncEnabled {
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
			Status:                       pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED,
			StatusDetail:                 "argocd sync completed.",
			StatusTime:                   syncTime,
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
	}

	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) createInArgo(chartGitAttribute *commonBean.ChartGitAttribute, ctx context.Context, envModel repository5.Environment, argocdAppName string) error {
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
		gitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(gitOpsRepoUrl)
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

	commitHash, _, err := impl.gitOpsRemoteOperationService.CommitValues(valuesGitConfig)
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

	_, _, err = impl.gitOpsRemoteOperationService.CommitValues(requirementsGitConfig)
	if err != nil {
		impl.logger.Errorw("error in values commit", "err", err)
		return errors.New(pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED)
	}

	return nil
}

func (impl AppStoreDeploymentFullModeServiceImpl) subscribeHelmInstallStatus() error {

	callback := func(msg *model.PubSubMsg) {

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

	// add required logging here
	var loggerFunc pubsub_lib.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		helmInstallNatsMessage := &appStoreBean.HelmReleaseStatusConfig{}
		err := json.Unmarshal([]byte(msg.Data), helmInstallNatsMessage)
		if err != nil {
			return "error in unmarshalling helm install status nats message", []interface{}{"err", err}
		}
		return "got nats msg for helm chart install status", []interface{}{"InstallAppVersionHistoryId", helmInstallNatsMessage.InstallAppVersionHistoryId, "ErrorInInstallation", helmInstallNatsMessage.ErrorInInstallation, "IsReleaseInstalled", helmInstallNatsMessage.IsReleaseInstalled}
	}

	err := impl.pubSubClient.Subscribe(pubsub_lib.HELM_CHART_INSTALL_STATUS_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
