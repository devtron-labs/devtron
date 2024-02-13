package FullMode

import (
	"context"
	"errors"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/client/argocdServer"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStatus2 "github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/chartGroup"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/deployment"
	util2 "github.com/devtron-labs/devtron/pkg/appStore/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/k8s"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
)

type InstalledAppDeploymentTypeChangeService interface {
	// MigrateDeploymentType migrates the deployment type of installed app and then trigger in loop
	MigrateDeploymentType(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	// TriggerAfterMigration triggers all the installed apps for which the deployment types were migrated via MigrateDeploymentType
	TriggerAfterMigration(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
}

type InstalledAppDeploymentTypeChangeServiceImpl struct {
	logger                        *zap.SugaredLogger
	installedAppRepository        repository2.InstalledAppRepository
	appRepository                 app.AppRepository
	userService                   user.UserService
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository
	appStatusService              appStatus.AppStatusService
	pubSubClient                  *pubsub.PubSubClientServiceImpl
	appStatusRepository           appStatus2.AppStatusRepository
	clusterRepository             repository5.ClusterRepository
	gitOpsConfigReadService       config.GitOpsConfigReadService
	environmentRepository         repository5.EnvironmentRepository
	acdClient                     application2.ServiceClient
	k8sCommonService              k8s.K8sCommonService
	k8sUtil                       *k8s2.K8sServiceImpl
	fullModeDeploymentService     deployment.FullModeDeploymentService
	eaModeDeploymentService       EAMode.EAModeDeploymentService
	argoClientWrapperService      argocdServer.ArgoClientWrapperService
	chartGroupService             chartGroup.ChartGroupService
	helmAppService                client.HelmAppService
}

func NewInstalledAppDeploymentTypeChangeServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository2.InstalledAppRepository,
	appRepository app.AppRepository,
	userService user.UserService,
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository,
	appStatusService appStatus.AppStatusService,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	appStatusRepository appStatus2.AppStatusRepository,
	clusterRepository repository5.ClusterRepository,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	environmentRepository repository5.EnvironmentRepository,
	acdClient application2.ServiceClient, k8sCommonService k8s.K8sCommonService,
	k8sUtil *k8s2.K8sServiceImpl, fullModeDeploymentService deployment.FullModeDeploymentService,
	eaModeDeploymentService EAMode.EAModeDeploymentService,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	chartGroupService chartGroup.ChartGroupService, helmAppService client.HelmAppService) *InstalledAppDeploymentTypeChangeServiceImpl {
	return &InstalledAppDeploymentTypeChangeServiceImpl{
		logger:                        logger,
		installedAppRepository:        installedAppRepository,
		appRepository:                 appRepository,
		userService:                   userService,
		installedAppRepositoryHistory: installedAppRepositoryHistory,
		appStatusService:              appStatusService,
		pubSubClient:                  pubSubClient,
		appStatusRepository:           appStatusRepository,
		clusterRepository:             clusterRepository,
		gitOpsConfigReadService:       gitOpsConfigReadService,
		environmentRepository:         environmentRepository,
		acdClient:                     acdClient,
		k8sCommonService:              k8sCommonService,
		k8sUtil:                       k8sUtil,
		fullModeDeploymentService:     fullModeDeploymentService,
		eaModeDeploymentService:       eaModeDeploymentService,
		argoClientWrapperService:      argoClientWrapperService,
		chartGroupService:             chartGroupService,
		helmAppService:                helmAppService,
	}
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) MigrateDeploymentType(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {
	response := &bean.DeploymentAppTypeChangeResponse{
		EnvId:                 request.EnvId,
		DesiredDeploymentType: request.DesiredDeploymentType,
	}
	var deleteDeploymentType bean.DeploymentType

	if request.DesiredDeploymentType == bean.ArgoCd {
		deleteDeploymentType = bean.Helm
	} else {
		deleteDeploymentType = bean.ArgoCd
	}
	//if cluster unreachable return with error, this is done to handle the case when cluster is unreachable and
	//delete req sent to argo cd the app deletion is stuck in deleting state
	isClusterReachable, err := impl.isClusterReachable(request.EnvId)
	if err != nil {
		return response, err
	}
	if !isClusterReachable {
		return response, &util.ApiError{HttpStatusCode: http.StatusNotFound, InternalMessage: "cluster unreachable", UserMessage: "cluster unreachable"}
	}

	installedApps, err := impl.installedAppRepository.GetActiveInstalledAppByEnvIdAndDeploymentType(request.EnvId,
		deleteDeploymentType, util2.ConvertIntArrayToStringArray(request.ExcludeApps), util2.ConvertIntArrayToStringArray(request.IncludeApps))
	if err != nil {
		impl.logger.Errorw("error in fetching installed apps by env id and deployment type", "endId", request.EnvId, "deleteDeploymentType", deleteDeploymentType)
		return response, err
	}
	var installedAppIds []int
	for _, item := range installedApps {
		installedAppIds = append(installedAppIds, item.Id)
	}

	if len(installedAppIds) == 0 {
		return response, nil
	}
	if request.DesiredDeploymentType == bean.Helm {
		//before deleting the installed app we'll first annotate CRD's manifest created by argo-cd with helm supported
		//annotations so that helm install doesn't throw crd already exist error while migrating from argo-cd to helm.
		for _, installedApp := range installedApps {
			err = impl.AnnotateCRDsIfExist(ctx, installedApp.App.AppName, installedApp.Environment.Name, installedApp.Environment.Namespace, installedApp.Environment.ClusterId)
			if err != nil {
				impl.logger.Errorw("error in annotating CRDs in manifest for argo-cd deployed installed apps", "installedAppId", installedApp.Id, "appId", installedApp.AppId)
				return response, err
			}
		}
	}
	envBean, err := impl.environmentRepository.FindById(request.EnvId)
	if err != nil {
		impl.logger.Errorw("error in registering acd app", "err", err)
		return response, err
	}
	deleteResponse, err := impl.deleteInstalledApps(ctx, installedApps, request.UserId, envBean.Cluster)
	if err != nil {
		return response, err
	}
	response.SuccessfulPipelines = deleteResponse.SuccessfulPipelines
	response.FailedPipelines = deleteResponse.FailedPipelines

	//instead of failed pipelines, mark successful pipelines
	var successInstalledAppIds []int
	for _, item := range response.SuccessfulPipelines {
		successInstalledAppIds = append(successInstalledAppIds, item.InstalledAppId)
	}
	err = impl.installedAppRepository.UpdateDeploymentAppTypeInInstalledApp(request.DesiredDeploymentType, successInstalledAppIds, request.UserId)
	if err != nil {
		impl.logger.Errorw("failed to update deployment app type for successfully deleted installed apps in db",
			"envId", request.EnvId,
			"successfully deleted installedApp ids", successInstalledAppIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, err
	}

	return response, nil
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) AnnotateCRDsIfExist(ctx context.Context, appName, envName, namespace string, clusterId int) error {
	deploymentAppName := fmt.Sprintf("%s-%s", appName, envName)
	query := &application.ResourcesQuery{
		ApplicationName: &deploymentAppName,
	}
	resp, err := impl.acdClient.ResourceTree(ctx, query)
	if err != nil {
		impl.logger.Errorw("error in fetching resource tree", "err", err)
		err = &util.ApiError{
			HttpStatusCode:  http.StatusNotFound,
			Code:            constants.AppDetailResourceTreeNotFound,
			InternalMessage: err.Error(),
			UserMessage:     "failed to get resource tree from acd",
		}
		return err
	}
	crdsList := make([]v1alpha1.ResourceNode, 0)
	for _, node := range resp.ApplicationTree.Nodes {
		if node.ResourceRef.Kind == kube.CustomResourceDefinitionKind {
			crdsList = append(crdsList, node)
		}
	}
	restConfig, err, _ := impl.k8sCommonService.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterId)
		return err
	}
	for _, crd := range crdsList {
		gvk := schema.GroupVersionKind{
			Group:   crd.ResourceRef.Group,
			Version: crd.ResourceRef.Version,
			Kind:    crd.ResourceRef.Kind,
		}
		//fetch annotation and labels keys if not exist then do something to add those labels with these keys
		helmAnnotation := fmt.Sprintf(bean.HelmReleaseMetadataAnnotation, appName, namespace)
		_, err = impl.k8sUtil.PatchResourceRequest(ctx, restConfig, types.StrategicMergePatchType, helmAnnotation, crd.ResourceRef.Name, "", gvk)
		if err != nil {
			impl.logger.Errorw("error in patching release-name annotation in manifest", "err", err, "appName", appName)
			return err
		}
	}
	return nil
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) deleteInstalledApps(ctx context.Context, installedApps []*repository2.InstalledApps, userId int32, cluster *repository5.Cluster) (*bean.DeploymentAppTypeChangeResponse, error) {
	successfullyDeletedApps := make([]*bean.DeploymentChangeStatus, 0)
	failedToDeleteApps := make([]*bean.DeploymentChangeStatus, 0)

	isGitOpsConfigured, gitOpsConfigErr := impl.gitOpsConfigReadService.IsGitOpsConfigured()

	for _, installedApp := range installedApps {

		var isValid bool
		// check if installed app info like app name and environment is empty or not
		if failedToDeleteApps, isValid = impl.isInstalledAppInfoValid(installedApp, failedToDeleteApps); !isValid {
			continue
		}

		var healthChkErr error
		// check health of the app if it is argo-cd deployment type
		if _, healthChkErr = impl.handleNotDeployedAppsIfArgoDeploymentType(installedApp, failedToDeleteApps); healthChkErr != nil {
			// cannot delete unhealthy app
			continue
		}

		deploymentAppName := fmt.Sprintf("%s-%s", installedApp.App.AppName, installedApp.Environment.Name)
		var err error

		// delete request
		if installedApp.DeploymentAppType == bean.ArgoCd {
			err = impl.fullModeDeploymentService.DeleteACD(deploymentAppName, ctx, false)
		} else {
			// For converting from Helm to ArgoCD, GitOps should be configured
			if gitOpsConfigErr != nil || !isGitOpsConfigured {
				err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: "200", UserMessage: errors.New("GitOps not configured or unable to fetch GitOps configuration")}
			} else {
				// Register app in ACD
				var acdRegisterErr, repoNameUpdateErr, createInArgoErr error
				installAppVersionRequest := &appStoreBean.InstallAppVersionDTO{
					AppName:        installedApp.App.AppName,
					GitOpsRepoName: installedApp.GitOpsRepoName,
					UserId:         userId,
				}
				repoUrl, _, createGitRepoErr := impl.fullModeDeploymentService.CreateGitOpsRepo(installAppVersionRequest)
				if createGitRepoErr != nil {
					impl.logger.Errorw("error in creating git repo", "err", err)
				}
				if createGitRepoErr == nil {
					chartGitAttr := &bean2.ChartGitAttribute{RepoUrl: repoUrl, ChartLocation: deploymentAppName}
					acdRegisterErr = impl.argoClientWrapperService.RegisterGitOpsRepoInArgo(ctx, chartGitAttr.RepoUrl)
					if acdRegisterErr != nil {
						impl.logger.Errorw("error in registering acd app", "err", err)
					}
					if acdRegisterErr == nil {
						installedApp.Environment.Cluster = cluster
						createInArgoErr = impl.fullModeDeploymentService.CreateInArgo(chartGitAttr, installedApp.Environment, deploymentAppName)
						if createInArgoErr != nil {
							impl.logger.Errorw("error in creating acd app", "err", err)
						}
					}

				}
				if createGitRepoErr != nil {
					err = createGitRepoErr
				} else if acdRegisterErr != nil {
					err = acdRegisterErr
				} else if createInArgoErr != nil {
					err = createInArgoErr
				} else if repoNameUpdateErr != nil {
					err = repoNameUpdateErr
				}
			}
			if err != nil {
				impl.logger.Errorw("error registering app on ACD with error: "+err.Error(),
					"deploymentAppName", deploymentAppName,
					"envId", installedApp.EnvironmentId,
					"appId", installedApp.AppId,
					"err", err)

				// deletion failed, append to the list of failed to delete installed apps
				failedToDeleteApps = impl.handleFailedInstalledAppChange(installedApp, failedToDeleteApps, appStoreBean.FAILED_TO_REGISTER_IN_ACD_ERROR+err.Error())
				continue
			}
			installAppVersionRequest := &appStoreBean.InstallAppVersionDTO{
				ClusterId: installedApp.Environment.ClusterId,
				AppName:   installedApp.App.AppName,
				Namespace: installedApp.Environment.Namespace,
			}
			err = impl.eaModeDeploymentService.DeleteInstalledApp(ctx, "", "", installAppVersionRequest, nil, nil)
		}

		if err != nil {
			impl.logger.Errorw("error deleting app on "+installedApp.DeploymentAppType,
				"deployment app name", deploymentAppName,
				"err", err)

			// deletion failed, append to the list of failed pipelines
			failedToDeleteApps = impl.handleFailedInstalledAppChange(installedApp, failedToDeleteApps, appStoreBean.FAILED_TO_DELETE_APP_PREFIX_ERROR+err.Error())
			continue
		}

		// deletion successful, append to the list of successful pipelines
		successfullyDeletedApps = appendToDeploymentChangeStatusList(successfullyDeletedApps, installedApp, "", bean.INITIATED)

	}
	return &bean.DeploymentAppTypeChangeResponse{
		SuccessfulPipelines: successfullyDeletedApps,
		FailedPipelines:     failedToDeleteApps,
	}, nil
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) isClusterReachable(envId int) (bool, error) {
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.logger.Errorw("error in finding env from envId", "envId", envId)
		return false, err
	}
	if len(env.Cluster.ErrorInConnecting) > 0 {
		return false, nil
	}
	return true, nil

}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) isInstalledAppInfoValid(installedApp *repository2.InstalledApps,
	failedToDeleteApps []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, bool) {

	if len(installedApp.App.AppName) == 0 || len(installedApp.Environment.Name) == 0 {
		impl.logger.Errorw("app name or environment name is not present", "installed app id", installedApp.Id)

		failedToDeleteApps = impl.handleFailedInstalledAppChange(installedApp, failedToDeleteApps, appStoreBean.COULD_NOT_FETCH_APP_NAME_AND_ENV_NAME_ERR)

		return failedToDeleteApps, false
	}
	return failedToDeleteApps, true
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) handleNotDeployedAppsIfArgoDeploymentType(installedApp *repository2.InstalledApps,
	failedToDeleteApps []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, error) {

	if installedApp.DeploymentAppType == string(bean.ArgoCd) {
		// check if app status is Healthy
		status, err := impl.appStatusRepository.Get(installedApp.AppId, installedApp.EnvironmentId)

		// case: missing status row in db
		if len(status.Status) == 0 {
			return failedToDeleteApps, nil
		}

		// cannot delete the app from argo-cd if app status is Progressing
		if err != nil {
			healthCheckErr := errors.New("unable to fetch app status")
			impl.logger.Errorw(healthCheckErr.Error(), "appId", installedApp.AppId, "environmentId", installedApp.EnvironmentId, "err", err)
			failedToDeleteApps = impl.handleFailedInstalledAppChange(installedApp, failedToDeleteApps, healthCheckErr.Error())
			return failedToDeleteApps, healthCheckErr
		}
		return failedToDeleteApps, nil
	}
	return failedToDeleteApps, nil
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) handleFailedInstalledAppChange(installedApp *repository2.InstalledApps,
	failedPipelines []*bean.DeploymentChangeStatus, err string) []*bean.DeploymentChangeStatus {

	return appendToDeploymentChangeStatusList(failedPipelines, installedApp, err, bean.Failed)
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) TriggerAfterMigration(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {
	response := &bean.DeploymentAppTypeChangeResponse{
		EnvId:                 request.EnvId,
		DesiredDeploymentType: request.DesiredDeploymentType,
	}
	var err error

	installedApps, err := impl.installedAppRepository.GetActiveInstalledAppByEnvIdAndDeploymentType(request.EnvId, request.DesiredDeploymentType,
		util2.ConvertIntArrayToStringArray(request.ExcludeApps), util2.ConvertIntArrayToStringArray(request.IncludeApps))

	if err != nil {
		impl.logger.Errorw("Error fetching installed apps",
			"environmentId", request.EnvId,
			"desiredDeploymentAppType", request.DesiredDeploymentType,
			"err", err)
		return response, err
	}

	var installedAppIds []int
	for _, item := range installedApps {
		installedAppIds = append(installedAppIds, item.Id)
	}

	if len(installedAppIds) == 0 {
		return response, nil
	}

	deleteResponse := impl.fetchDeletedInstalledApp(ctx, installedApps)

	response.SuccessfulPipelines = deleteResponse.SuccessfulPipelines
	response.FailedPipelines = deleteResponse.FailedPipelines

	successfulInstalledAppIds := make([]int, 0, len(response.SuccessfulPipelines))
	for _, item := range response.SuccessfulPipelines {
		successfulInstalledAppIds = append(successfulInstalledAppIds, item.InstalledAppId)
	}

	successInstalledApps, err := impl.installedAppRepository.FindInstalledAppByIds(successfulInstalledAppIds)
	if err != nil {
		impl.logger.Errorw("failed to fetch installed app details",
			"ids", successfulInstalledAppIds,
			"err", err)

		return response, nil
	}
	var installedAppVersionDTOList []*appStoreBean.InstallAppVersionDTO
	for _, installedApp := range successInstalledApps {
		installedAppVersion, err := impl.installedAppRepository.GetActiveInstalledAppVersionByInstalledAppId(installedApp.Id)
		if err != nil {
			impl.logger.Errorw("error in getting installedAppVersion from installedAppId",
				"installedAppId", installedApp.Id,
				"err", err)
			return nil, err
		}
		installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetLatestInstalledAppVersionHistory(installedAppVersion.Id)
		if err != nil {
			impl.logger.Errorw("error in getting installedAppVersionHistory from installedAppVersionId",
				"installedAppVersionId", installedAppVersion.Id,
				"err", err)
			return nil, err
		}
		installedAppVersionDTOList = append(installedAppVersionDTOList, &appStoreBean.InstallAppVersionDTO{
			InstalledAppVersionId:        installedAppVersion.Id,
			InstalledAppVersionHistoryId: installedAppVersionHistory.Id,
		})
	}

	impl.chartGroupService.TriggerDeploymentEvent(installedAppVersionDTOList)

	return response, nil
}

func (impl *InstalledAppDeploymentTypeChangeServiceImpl) fetchDeletedInstalledApp(ctx context.Context,
	installedApps []*repository2.InstalledApps) *bean.DeploymentAppTypeChangeResponse {

	successfulInstalledApps := make([]*bean.DeploymentChangeStatus, 0)
	failedInstalledApps := make([]*bean.DeploymentChangeStatus, 0)

	for _, installedApp := range installedApps {

		deploymentAppName := fmt.Sprintf("%s-%s", installedApp.App.AppName, installedApp.Environment.Name)
		var err error
		if installedApp.DeploymentAppType == bean.ArgoCd {
			appIdentifier := &client.AppIdentifier{
				ClusterId:   installedApp.Environment.ClusterId,
				ReleaseName: deploymentAppName,
				Namespace:   installedApp.Environment.Namespace,
			}
			_, err = impl.helmAppService.GetApplicationDetail(ctx, appIdentifier)
		} else {
			req := &application.ApplicationQuery{
				Name: &deploymentAppName,
			}
			_, err = impl.acdClient.Get(ctx, req)
		}
		if err != nil {
			impl.logger.Errorw("error in getting application detail", "err", err, "deploymentAppName", deploymentAppName)
		}

		if err != nil && util2.CheckAppReleaseNotExist(err) {
			successfulInstalledApps = appendToDeploymentChangeStatusList(successfulInstalledApps, installedApp, "", bean.Success)
		} else {
			failedInstalledApps = appendToDeploymentChangeStatusList(failedInstalledApps, installedApp, appStoreBean.APP_NOT_DELETED_YET_ERROR, bean.NOT_YET_DELETED)
		}
	}

	return &bean.DeploymentAppTypeChangeResponse{
		SuccessfulPipelines: successfulInstalledApps,
		FailedPipelines:     failedInstalledApps,
	}
}

func appendToDeploymentChangeStatusList(installedApps []*bean.DeploymentChangeStatus,
	installedApp *repository2.InstalledApps, error string, status bean.Status) []*bean.DeploymentChangeStatus {

	return append(installedApps, &bean.DeploymentChangeStatus{
		InstalledAppId: installedApp.Id,
		AppId:          installedApp.AppId,
		AppName:        installedApp.App.AppName,
		EnvId:          installedApp.EnvironmentId,
		EnvName:        installedApp.Environment.Name,
		Error:          error,
		Status:         status,
	})
}
