package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/go-pg/pg"
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
}

func (impl *FullModeDeploymentServiceImpl) GetAcdAppGitOpsRepoName(appName string, environmentName string) (string, error) {
	gitOpsRepoName := ""
	//this method should only call in case of argo-integration and gitops configured
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.Logger.Errorw("error in getting acd token", "err", err)
		return "", err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	acdAppName := fmt.Sprintf("%s-%s", appName, environmentName)
	acdApplication, err := impl.acdClient.Get(ctx, &application.ApplicationQuery{Name: &acdAppName})
	if err != nil {
		impl.Logger.Errorw("no argo app exists", "acdAppName", acdAppName, "err", err)
		return "", err
	}
	if acdApplication != nil {
		gitOpsRepoUrl := acdApplication.Spec.Source.RepoURL
		gitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(gitOpsRepoUrl)
	}
	return gitOpsRepoName, nil
}

func (impl *FullModeDeploymentServiceImpl) DeleteACDAppObject(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	acdAppName := appName + "-" + environmentName
	var err error
	err = impl.deleteACD(acdAppName, ctx, installAppVersionRequest.NonCascadeDelete)
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
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.Logger.Errorw("error in getting acd token", "err", err)
		return isFound, fmt.Errorf("error in getting acd token")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)

	_, acdAppGetErr := impl.acdClient.Get(ctx, &application.ApplicationQuery{Name: &acdAppName})
	isFound = acdAppGetErr == nil
	return isFound, nil
}

func (impl *FullModeDeploymentServiceImpl) UpdateAndSyncACDApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, isMonoRepoMigrationRequired bool, ctx context.Context, tx *pg.Tx) error {
	if isMonoRepoMigrationRequired {
		// update repo details on ArgoCD as repo is changed
		err := impl.UpgradeDeployment(installAppVersionRequest, ChartGitAttribute, 0, ctx)
		if err != nil {
			return err
		}
	}
	acdAppName := installAppVersionRequest.ACDAppName
	argoApplication, err := impl.acdClient.Get(ctx, &application.ApplicationQuery{Name: &acdAppName})
	if err != nil {
		impl.Logger.Errorw("Service err:UpdateAndSyncACDApps - error in acd app by name", "acdAppName", acdAppName, "err", err)
		return err
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
		return err
	}
	if !impl.acdConfig.ArgoCDAutoSyncEnabled {
		err = impl.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED, "argocd sync completed", syncTime, tx)
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

func (impl *FullModeDeploymentServiceImpl) deleteACD(acdAppName string, ctx context.Context, isNonCascade bool) error {
	req := new(application.ApplicationDeleteRequest)
	req.Name = &acdAppName
	cascadeDelete := !isNonCascade
	req.Cascade = &cascadeDelete
	if ctx == nil {
		impl.Logger.Errorw("err in delete ACD for AppStore, ctx is NULL", "acdAppName", acdAppName)
		return fmt.Errorf("context is null")
	}
	if _, err := impl.acdClient.Delete(ctx, req); err != nil {
		impl.Logger.Errorw("err in delete ACD for AppStore", "acdAppName", acdAppName, "err", err)
		return err
	}
	return nil
}

func (impl *FullModeDeploymentServiceImpl) createInArgo(chartGitAttribute *commonBean.ChartGitAttribute, envModel bean.EnvironmentBean, argocdAppName string) error {
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
	_, err := impl.argoK8sClient.CreateAcdApp(appReq, argocdServer.ARGOCD_APPLICATION_TEMPLATE)
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
	err := impl.argoClientWrapperService.RegisterGitOpsRepoInArgo(ctx, chartGitAttr.RepoUrl)
	if err != nil {
		impl.Logger.Errorw("error in argo registry", "err", err)
		return err
	}
	// update acd app
	patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: &v1alpha1.ApplicationSource{Path: chartGitAttr.ChartLocation, RepoURL: chartGitAttr.RepoUrl, TargetRevision: "master"}}}
	reqbyte, err := json.Marshal(patchReq)
	if err != nil {
		impl.Logger.Errorw("error in creating patch", "err", err)
	}
	reqString := string(reqbyte)
	patchType := "merge"
	_, err = impl.acdClient.Patch(ctx, &application.ApplicationPatchRequest{Patch: &reqString, Name: &installAppVersionRequest.ACDAppName, PatchType: &patchType})
	if err != nil {
		impl.Logger.Errorw("error in creating argo app ", "name", installAppVersionRequest.ACDAppName, "patch", string(reqbyte), "err", err)
		return err
	}
	return nil
}
