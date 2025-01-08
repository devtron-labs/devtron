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

package deployment

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/argocdServer"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type InstalledAppArgoCdService interface {
	// GetAcdAppGitOpsRepoName will return the Git Repo used in the ACD app object
	GetAcdAppGitOpsRepoName(appName string, environmentName string) (string, error)
	// DeleteACDAppObject will delete the app object from ACD
	DeleteACDAppObject(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error
	// CheckIfArgoAppExists will return- isFound: if Argo app object exists; err: if any err found
	CheckIfArgoAppExists(acdAppName string) (isFound bool, err error)
	// UpdateAndSyncACDApps this will update chart info in acd app if required in case of mono repo migration and will refresh argo app
	UpdateAndSyncACDApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, isMonoRepoMigrationRequired bool, ctx context.Context, tx *pg.Tx) error
	DeleteACD(acdAppName string, ctx context.Context, isNonCascade bool) error
	GetAcdAppGitOpsRepoURL(appName string, environmentName string) (string, error)
}

func (impl *FullModeDeploymentServiceImpl) GetAcdAppGitOpsRepoName(appName string, environmentName string) (string, error) {
	//this method should only call in case of argo-integration and gitops configured
	acdAppName := util2.BuildDeployedAppName(appName, environmentName)
	return impl.argoClientWrapperService.GetGitOpsRepoNameForApplication(context.Background(), acdAppName)
}

func (impl *FullModeDeploymentServiceImpl) DeleteACDAppObject(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	acdAppName := util2.BuildDeployedAppName(appName, environmentName)
	var err error
	err = impl.DeleteACD(acdAppName, ctx, installAppVersionRequest.NonCascadeDelete)
	if err != nil {
		impl.Logger.Errorw("error in deleting ACD ", "name", acdAppName, "err", err)
		if installAppVersionRequest.ForceDelete {
			impl.Logger.Warnw("error while deletion of app in acd, continue to delete in db as this operation is force delete", "error", err)
		} else {
			//statusError, _ := err.(*errors2.StatusError)
			if !installAppVersionRequest.NonCascadeDelete && strings.Contains(err.Error(), "code = NotFound") {
				err = &util.ApiError{
					UserMessage:     "Could not delete as application not found in argocd",
					InternalMessage: err.Error(),
				}
			} else {
				err = &util.ApiError{
					UserMessage:     "Could not delete application",
					InternalMessage: err.Error(),
				}
			}
			return err
		}
	}
	return nil
}

func (impl *FullModeDeploymentServiceImpl) CheckIfArgoAppExists(acdAppName string) (isFound bool, err error) {
	_, acdAppGetErr := impl.argoClientWrapperService.GetArgoAppByName(context.Background(), acdAppName)
	isFound = acdAppGetErr == nil
	return isFound, nil
}

func isArgoCdGitOpsRepoUrlOutOfSync(argoApplication *v1alpha1.Application, gitOpsRepoURLInDb string) bool {
	if argoApplication != nil && argoApplication.Spec.Source != nil {
		return argoApplication.Spec.Source.RepoURL != gitOpsRepoURLInDb
	}
	return false
}

func (impl *FullModeDeploymentServiceImpl) UpdateAndSyncACDApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, isMonoRepoMigrationRequired bool, ctx context.Context, tx *pg.Tx) error {
	acdAppName := installAppVersionRequest.ACDAppName
	argoApplication, err := impl.argoClientWrapperService.GetArgoAppByName(ctx, acdAppName)
	if err != nil {
		impl.Logger.Errorw("Service err:UpdateAndSyncACDApps - error in acd app by name", "acdAppName", acdAppName, "err", err)
		return err
	}
	//if either monorepo case is true or there is diff. in git-ops repo url registered with argo-cd and git-ops repo url saved in db,
	//then sync argo with git-ops repo url from db because we have already pushed changes to that repo
	isArgoRepoUrlOutOfSync := isArgoCdGitOpsRepoUrlOutOfSync(argoApplication, installAppVersionRequest.GitOpsRepoURL)
	if isMonoRepoMigrationRequired || isArgoRepoUrlOutOfSync {
		// update repo details on ArgoCD as repo is changed
		err := impl.UpgradeDeployment(installAppVersionRequest, ChartGitAttribute, 0, ctx)
		if err != nil {
			return err
		}
	}

	err = impl.argoClientWrapperService.UpdateArgoCDSyncModeIfNeeded(ctx, argoApplication)
	if err != nil {
		impl.Logger.Errorw("error in updating argocd sync mode", "err", err)
		return err
	}
	syncTime := time.Now()
	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, acdAppName)
	if err != nil {
		impl.Logger.Errorw("error in getting argocd application with normal refresh", "err", err, "argoAppName", installAppVersionRequest.ACDAppName)
		clientErrCode, errMsg := util.GetClientDetailedError(err)
		if clientErrCode.IsFailedPreconditionCode() {
			return &util.ApiError{HttpStatusCode: http.StatusPreconditionFailed, Code: strconv.Itoa(http.StatusPreconditionFailed), InternalMessage: errMsg, UserMessage: errMsg}
		}
		return err
	}
	if impl.acdConfig.IsManualSyncEnabled() {
		err = impl.SaveTimelineForHelmApps(installAppVersionRequest, timelineStatus.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED, timelineStatus.TIMELINE_DESCRIPTION_ARGOCD_SYNC_COMPLETED, syncTime, tx)
		if err != nil {
			impl.Logger.Errorw("error in saving timeline for acd helm apps", "err", err)
			return err
		}
	}
	return nil
}

// UpgradeDeployment this will update chart info in acd app, needed when repo for an app is changed
func (impl *FullModeDeploymentServiceImpl) UpgradeDeployment(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, installedAppVersionHistoryId int, ctx context.Context) error {
	if !util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
		return nil
	}
	err := impl.patchAcdApp(ctx, installAppVersionRequest, ChartGitAttribute)
	if err != nil {
		impl.Logger.Errorw("error in acd patch request", "err", err)
	}
	return err
}

func (impl *FullModeDeploymentServiceImpl) DeleteACD(acdAppName string, ctx context.Context, isNonCascade bool) error {
	cascadeDelete := !isNonCascade
	if ctx == nil {
		impl.Logger.Errorw("err in delete ACD for AppStore, ctx is NULL", "acdAppName", acdAppName)
		return fmt.Errorf("context is null")
	}
	if _, err := impl.argoClientWrapperService.DeleteArgoApp(ctx, acdAppName, cascadeDelete); err != nil {
		impl.Logger.Errorw("err in delete ACD for AppStore", "acdAppName", acdAppName, "err", err)
		return err
	}
	return nil
}

func (impl *FullModeDeploymentServiceImpl) createInArgo(ctx context.Context, chartGitAttribute *commonBean.ChartGitAttribute, envModel bean.EnvironmentBean, argocdAppName string) error {
	appNamespace := envModel.Namespace
	if appNamespace == "" {
		appNamespace = cluster2.DEFAULT_NAMESPACE
	}
	appReq := &argocdServer.AppTemplate{
		ApplicationName: argocdAppName,
		Namespace:       impl.aCDAuthConfig.ACDConfigMapNamespace,
		TargetNamespace: appNamespace,
		TargetServer:    envModel.ClusterServerUrl,
		Project:         "default",
		ValuesFile:      fmt.Sprintf("values.yaml"),
		RepoPath:        chartGitAttribute.ChartLocation,
		RepoUrl:         chartGitAttribute.RepoUrl,
		AutoSyncEnabled: impl.acdConfig.ArgoCDAutoSyncEnabled,
	}
	_, err := impl.argoK8sClient.CreateAcdApp(ctx, appReq, argocdServer.ARGOCD_APPLICATION_TEMPLATE)
	//create
	if err != nil {
		impl.Logger.Errorw("error in creating argo cd app ", "err", err)
		return err
	}
	return nil
}

func (impl *FullModeDeploymentServiceImpl) patchAcdApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//registerInArgo
	err := impl.argoClientWrapperService.RegisterGitOpsRepoInArgoWithRetry(ctx, chartGitAttr.RepoUrl, installAppVersionRequest.UserId)
	if err != nil {
		impl.Logger.Errorw("error in argo registry", "err", err)
		return err
	}
	// update acd app

	patchReq := &argoApplication.ArgoCdAppPatchReqDto{
		ArgoAppName:    installAppVersionRequest.ACDAppName,
		ChartLocation:  chartGitAttr.ChartLocation,
		GitRepoUrl:     chartGitAttr.RepoUrl,
		TargetRevision: "master",
		PatchType:      "merge",
	}
	err = impl.argoClientWrapperService.PatchArgoCdApp(ctx, patchReq)
	if err != nil {
		impl.Logger.Errorw("error in patching acd app ", "appName", installAppVersionRequest.ACDAppName, "err", err)
		return err
	}
	return nil
}

func (impl *FullModeDeploymentServiceImpl) GetAcdAppGitOpsRepoURL(appName string, environmentName string) (string, error) {
	ctx := context.Background()
	acdAppName := util2.BuildDeployedAppName(appName, environmentName)
	return impl.argoClientWrapperService.GetGitOpsRepoURLForApplication(ctx, acdAppName)
}
