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

package appstore

import (
	"bytes"
	"context"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository4 "github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/ktrysmt/go-bitbucket"

	/* #nosec */
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	repository2 "github.com/argoproj/argo-cd/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/constants"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore/chartGroup"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"github.com/nats-io/stan.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const (
	DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT = "devtron"
	CLUSTER_COMPONENT_DIR_PATH                  = "/cluster/component"
)

type InstalledAppService interface {
	UpdateInstalledApp(ctx context.Context, installAppVersionRequest *InstallAppVersionDTO) (*InstallAppVersionDTO, error)
	GetInstalledApp(id int) (*InstallAppVersionDTO, error)
	GetInstalledAppVersion(id int) (*InstallAppVersionDTO, error)
	GetAll(filter *appstore.AppStoreFilter) (openapi.AppList, error)
	GetAllInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request, token string, appStoreId int) ([]InstalledAppsResponse, error)
	DeleteInstalledApp(ctx context.Context, installAppVersionRequest *InstallAppVersionDTO) (*InstallAppVersionDTO, error)

	DeployBulk(chartGroupInstallRequest *ChartGroupInstallRequest) (*ChartGroupInstallAppRes, error)

	CreateInstalledAppV2(installAppVersionRequest *InstallAppVersionDTO, ctx context.Context) (*InstallAppVersionDTO, error)
	AppStoreDeployOperationGIT(installAppVersionRequest *InstallAppVersionDTO) (*InstallAppVersionDTO, *util.ChartGitAttribute, error)
	AppStoreDeployOperationACD(installAppVersionRequest *InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context) (*InstallAppVersionDTO, error)
	AppStoreDeployOperationDB(installAppVersionRequest *InstallAppVersionDTO, tx *pg.Tx) (*InstallAppVersionDTO, error)
	performDeployStage(appId int) (*InstallAppVersionDTO, error)
	AppStoreDeployOperationStatusUpdate(installAppId int, status appstore.AppstoreDeploymentStatus) (bool, error)
	CheckAppExists(appNames []*AppNames) ([]*AppNames, error)

	DeployDefaultChartOnCluster(bean *cluster2.ClusterBean, userId int32) (bool, error)
	IsChartRepoActive(appStoreVersionId int) (bool, error)
	DeployDefaultComponent(chartGroupInstallRequest *ChartGroupInstallRequest) (*ChartGroupInstallAppRes, error)
}

type InstalledAppServiceImpl struct {
	chartRepository                      chartConfig.ChartRepository
	logger                               *zap.SugaredLogger
	repoRepository                       chartConfig.ChartRepoRepository
	mergeUtil                            util.MergeUtil
	pipelineConfigRepository             chartConfig.PipelineConfigRepository
	configMapRepository                  chartConfig.ConfigMapRepository
	installedAppRepository               appstore.InstalledAppRepository
	chartTemplateService                 util.ChartTemplateService
	refChartDir                          RefChartProxyDir
	repositoryService                    repository.ServiceClient
	appStoreApplicationVersionRepository appstore.AppStoreApplicationVersionRepository
	environmentRepository                repository5.EnvironmentRepository
	teamRepository                       repository4.TeamRepository
	appRepository                        app.AppRepository
	acdClient                            application2.ServiceClient
	appStoreValuesService                AppStoreValuesService
	pubsubClient                         *pubsub.PubSubClient
	tokenCache                           *util2.TokenCache
	chartGroupDeploymentRepository       chartGroup.ChartGroupDeploymentRepository
	envService                           cluster2.EnvironmentService
	clusterInstalledAppsRepository       appstore.ClusterInstalledAppsRepository
	ArgoK8sClient                        argocdServer.ArgoK8sClient
	gitFactory                           *util.GitFactory
	aCDAuthConfig                        *util2.ACDAuthConfig
	gitOpsRepository                     repository3.GitOpsConfigRepository
}

func NewInstalledAppServiceImpl(chartRepository chartConfig.ChartRepository,
	logger *zap.SugaredLogger,
	repoRepository chartConfig.ChartRepoRepository,
	mergeUtil util.MergeUtil,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	installedAppRepository appstore.InstalledAppRepository,
	chartTemplateService util.ChartTemplateService, refChartDir RefChartProxyDir,
	repositoryService repository.ServiceClient,
	appStoreApplicationVersionRepository appstore.AppStoreApplicationVersionRepository,
	environmentRepository repository5.EnvironmentRepository, teamRepository repository4.TeamRepository,
	appRepository app.AppRepository,
	acdClient application2.ServiceClient,
	appStoreValuesService AppStoreValuesService,
	pubsubClient *pubsub.PubSubClient,
	tokenCache *util2.TokenCache,
	chartGroupDeploymentRepository chartGroup.ChartGroupDeploymentRepository,
	envService cluster2.EnvironmentService,
	clusterInstalledAppsRepository appstore.ClusterInstalledAppsRepository,
	argoK8sClient argocdServer.ArgoK8sClient,
	gitFactory *util.GitFactory, aCDAuthConfig *util2.ACDAuthConfig, gitOpsRepository repository3.GitOpsConfigRepository) (*InstalledAppServiceImpl, error) {
	impl := &InstalledAppServiceImpl{
		chartRepository:                      chartRepository,
		logger:                               logger,
		repoRepository:                       repoRepository,
		mergeUtil:                            mergeUtil,
		pipelineConfigRepository:             pipelineConfigRepository,
		configMapRepository:                  configMapRepository,
		installedAppRepository:               installedAppRepository,
		chartTemplateService:                 chartTemplateService,
		refChartDir:                          refChartDir,
		repositoryService:                    repositoryService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
		teamRepository:                       teamRepository,
		appRepository:                        appRepository,
		acdClient:                            acdClient,
		appStoreValuesService:                appStoreValuesService,
		pubsubClient:                         pubsubClient,
		tokenCache:                           tokenCache,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		envService:                           envService,
		clusterInstalledAppsRepository:       clusterInstalledAppsRepository,
		ArgoK8sClient:                        argoK8sClient,
		gitFactory:                           gitFactory,
		aCDAuthConfig:                        aCDAuthConfig,
		gitOpsRepository:                     gitOpsRepository,
	}
	err := impl.Subscribe()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

func (impl InstalledAppServiceImpl) UpdateInstalledApp(ctx context.Context, installAppVersionRequest *InstallAppVersionDTO) (*InstallAppVersionDTO, error) {

	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	installedApp, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
	if err != nil {
		return nil, err
	}
	installAppVersionRequest.EnvironmentId = installedApp.EnvironmentId
	installAppVersionRequest.AppName = installedApp.App.AppName

	environment, err := impl.environmentRepository.FindById(installedApp.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	team, err := impl.teamRepository.FindOne(installedApp.App.TeamId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	impl.logger.Debug(team.Name)
	argocdAppName := installedApp.App.AppName + "-" + environment.Name

	if installAppVersionRequest.Id == 0 {
		// upgrade chart to other repo
		_, _, err := impl.upgradeInstalledApp(ctx, installAppVersionRequest, installedApp, tx)
		if err != nil {
			impl.logger.Errorw("error while upgrade the chart", "error", err)
			return nil, err
		}
	} else {
		// update same chart or upgrade its version only
		var installedAppVersion *appstore.InstalledAppVersions
		installedAppVersionModel, err := impl.installedAppRepository.GetInstalledAppVersion(installAppVersionRequest.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching chart installed version", "error", err)
			return nil, err
		}
		if installedAppVersionModel.AppStoreApplicationVersionId != installAppVersionRequest.AppStoreVersion {
			installedAppVersionModel.Active = false
			installedAppVersionModel.UpdatedOn = time.Now()
			installedAppVersionModel.UpdatedBy = installAppVersionRequest.UserId
			_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
			if err != nil {
				impl.logger.Errorw("error while fetching from db", "error", err)
				return nil, err
			}
			appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
			if err != nil {
				impl.logger.Errorw("fetching error", "err", err)
				return nil, err
			}
			installedAppVersion = &appstore.InstalledAppVersions{
				InstalledAppId:               installAppVersionRequest.InstalledAppId,
				AppStoreApplicationVersionId: installAppVersionRequest.AppStoreVersion,
				ValuesYaml:                   installAppVersionRequest.ValuesOverrideYaml,
			}
			installedAppVersion.CreatedBy = installAppVersionRequest.UserId
			installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
			installedAppVersion.CreatedOn = time.Now()
			installedAppVersion.UpdatedOn = time.Now()
			installedAppVersion.Active = true
			installedAppVersion.ReferenceValueId = installAppVersionRequest.ReferenceValueId
			installedAppVersion.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
			_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersion, tx)
			if err != nil {
				impl.logger.Errorw("error while fetching from db", "error", err)
				return nil, err
			}
			installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion

			//update requirements yaml in chart
			err = impl.updateRequirementDependencies(environment, installedAppVersion, installAppVersionRequest, appStoreAppVersion)
			if err != nil {
				impl.logger.Errorw("error while commit required dependencies to git", "error", err)
				return nil, err
			}

		} else {
			installedAppVersion = installedAppVersionModel
		}

		//update values yaml in chart
		err = impl.updateValuesYaml(environment, installedAppVersion, installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error while commit values to git", "error", err)
			return nil, err
		}
		installAppVersionRequest.ACDAppName = argocdAppName
		installAppVersionRequest.Environment = environment

		//ACD sync operation
		impl.syncACD(installAppVersionRequest.ACDAppName, ctx)

		//DB operation
		installedAppVersion.ValuesYaml = installAppVersionRequest.ValuesOverrideYaml
		installedAppVersion.UpdatedOn = time.Now()
		installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
		installedAppVersion.ReferenceValueId = installAppVersionRequest.ReferenceValueId
		installedAppVersion.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersion, tx)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
	}

	//STEP 8: finish with return response
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return nil, err
	}
	return installAppVersionRequest, nil
}

func (impl InstalledAppServiceImpl) updateRequirementDependencies(environment *repository5.Environment, installedAppVersion *appstore.InstalledAppVersions,
	installAppVersionRequest *InstallAppVersionDTO, appStoreAppVersion *appstore.AppStoreApplicationVersion) error {

	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	dependency := Dependency{
		Name:       appStoreAppVersion.AppStore.Name,
		Version:    appStoreAppVersion.Version,
		Repository: appStoreAppVersion.AppStore.ChartRepo.Url,
	}
	var dependencies []Dependency
	dependencies = append(dependencies, dependency)
	requirementDependencies := &Dependencies{
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
	chartGitAttrForRequirement := &util.ChartConfig{
		FileName:       REQUIREMENTS_YAML_FILE,
		FileContent:    string(requirementDependenciesByte),
		ChartName:      installedAppVersion.AppStoreApplicationVersion.AppStore.Name,
		ChartLocation:  argocdAppName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
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
	_, err = impl.gitFactory.Client.CommitValues(chartGitAttrForRequirement, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return err
	}
	return nil
}

func (impl InstalledAppServiceImpl) updateValuesYaml(environment *repository5.Environment, installedAppVersion *appstore.InstalledAppVersions,
	installAppVersionRequest *InstallAppVersionDTO) error {

	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	valuesOverrideByte, err := yaml.YAMLToJSON([]byte(installAppVersionRequest.ValuesOverrideYaml))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return err
	}
	var dat map[string]interface{}
	err = json.Unmarshal(valuesOverrideByte, &dat)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return err
	}
	valuesMap := make(map[string]map[string]interface{})
	valuesMap[installedAppVersion.AppStoreApplicationVersion.AppStore.Name] = dat
	valuesByte, err := json.Marshal(valuesMap)
	if err != nil {
		impl.logger.Errorw("error in marshaling", "err", err)
		return err
	}
	valuesYaml := &util.ChartConfig{
		FileName:       VALUES_YAML_FILE,
		FileContent:    string(valuesByte),
		ChartName:      installedAppVersion.AppStoreApplicationVersion.AppStore.Name,
		ChartLocation:  argocdAppName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", installedAppVersion.AppStoreApplicationVersion.Id, environment.Id),
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
	_, err = impl.gitFactory.Client.CommitValues(valuesYaml, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return err
	}
	return nil
}

func (impl InstalledAppServiceImpl) upgradeInstalledApp(ctx context.Context, installAppVersionRequest *InstallAppVersionDTO, installedApp *appstore.InstalledApps, tx *pg.Tx) (*InstallAppVersionDTO, *appstore.InstalledAppVersions, error) {
	var installedAppVersion *appstore.InstalledAppVersions
	installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppId(installAppVersionRequest.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed version", "error", err)
		return installAppVersionRequest, installedAppVersion, err
	}
	for _, installedAppVersionModel := range installedAppVersions {
		installedAppVersionModel.Active = false
		installedAppVersionModel.UpdatedOn = time.Now()
		installedAppVersionModel.UpdatedBy = installAppVersionRequest.UserId
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
		if err != nil {
			impl.logger.Errorw("error while update installed chart", "error", err)
			return installAppVersionRequest, installedAppVersion, err
		}
	}

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return installAppVersionRequest, installedAppVersion, err
	}
	installedAppVersion = &appstore.InstalledAppVersions{
		InstalledAppId:               installAppVersionRequest.InstalledAppId,
		AppStoreApplicationVersionId: installAppVersionRequest.AppStoreVersion,
		ValuesYaml:                   installAppVersionRequest.ValuesOverrideYaml,
	}
	installedAppVersion.CreatedBy = installAppVersionRequest.UserId
	installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
	installedAppVersion.CreatedOn = time.Now()
	installedAppVersion.UpdatedOn = time.Now()
	installedAppVersion.Active = true
	installedAppVersion.ReferenceValueId = installAppVersionRequest.ReferenceValueId
	installedAppVersion.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
	_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersion, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return installAppVersionRequest, installedAppVersion, err
	}
	installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion

	//step 2 git operation pull push
	installAppVersionRequest, chartGitAttr, err := impl.AppStoreDeployOperationGIT(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return installAppVersionRequest, installedAppVersion, err
	}

	//step 3 acd operation register, sync
	installAppVersionRequest, err = impl.patchAcdApp(installAppVersionRequest, chartGitAttr, ctx)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return installAppVersionRequest, installedAppVersion, err
	}

	//step 4 db operation status triggered
	installedApp.Status = appstore.DEPLOY_SUCCESS
	_, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return installAppVersionRequest, installedAppVersion, err
	}

	return installAppVersionRequest, installedAppVersion, err
}

func (impl InstalledAppServiceImpl) GetInstalledApp(id int) (*InstallAppVersionDTO, error) {

	app, err := impl.installedAppRepository.GetInstalledApp(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	chartTemplate, err := impl.chartAdaptor2(app)
	return chartTemplate, err
}

func (impl InstalledAppServiceImpl) GetInstalledAppVersion(id int) (*InstallAppVersionDTO, error) {
	app, err := impl.installedAppRepository.GetInstalledAppVersion(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	installAppVersion := &InstallAppVersionDTO{
		InstalledAppId:     app.InstalledAppId,
		AppName:            app.InstalledApp.App.AppName,
		AppId:              app.InstalledApp.App.Id,
		Id:                 app.Id,
		TeamId:             app.InstalledApp.App.TeamId,
		EnvironmentId:      app.InstalledApp.EnvironmentId,
		ValuesOverrideYaml: app.ValuesYaml,
		Readme:             app.AppStoreApplicationVersion.Readme,
		ReferenceValueKind: app.ReferenceValueKind,
		ReferenceValueId:   app.ReferenceValueId,
		AppStoreVersion:    app.AppStoreApplicationVersionId, //check viki
		Status:             app.InstalledApp.Status,
		AppStoreId:         app.AppStoreApplicationVersion.AppStoreId,
		AppStoreName:       app.AppStoreApplicationVersion.AppStore.Name,
		Deprecated:         app.AppStoreApplicationVersion.Deprecated,
	}
	return installAppVersion, err
}

func (impl InstalledAppServiceImpl) GetAll(filter *appstore.AppStoreFilter) (openapi.AppList, error) {
	applicationType := "DEVTRON-CHART-STORE"
	var clusterIdsConverted []int32
	for _, clusterId := range filter.ClusterIds {
		clusterIdsConverted = append(clusterIdsConverted, int32(clusterId))
	}
	installedAppsResponse := openapi.AppList{
		ApplicationType: &applicationType,
		ClusterIds:      &clusterIdsConverted,
	}
	installedApps, err := impl.installedAppRepository.GetAllInstalledApps(filter)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return installedAppsResponse, err
	}
	var helmAppsResponse []openapi.HelmApp
	for _, a := range installedApps {
		appLocal := a // copied data from here because value is passed as reference
		if appLocal.TeamId == 0 {
			//skipping entries for empty projectId
			continue
		}
		appId := strconv.Itoa(appLocal.Id)
		projectId := int32(appLocal.TeamId)
		envId := int32(appLocal.EnvironmentId)
		clusterId := int32(appLocal.ClusterId)
		environmentDetails := openapi.AppEnvironmentDetail{
			EnvironmentName: &appLocal.EnvironmentName,
			EnvironmentId:   &envId,
			Namespace:       &appLocal.Namespace,
			ClusterName:     &appLocal.ClusterName,
			ClusterId:       &clusterId,
		}
		helmAppResp := openapi.HelmApp{
			AppName:           &appLocal.AppName,
			ChartName:         &appLocal.AppStoreApplicationName,
			AppId:             &appId,
			ProjectId:         &projectId,
			EnvironmentDetail: &environmentDetails,
			ChartAvatar:       &appLocal.Icon,
			LastDeployedAt:    &appLocal.UpdatedOn,
		}
		helmAppsResponse = append(helmAppsResponse, helmAppResp)
	}
	installedAppsResponse.HelmApps = &helmAppsResponse
	return installedAppsResponse, nil
}

// TODO: Test ACD to get status
func (impl InstalledAppServiceImpl) GetAllInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request, token string, appStoreId int) ([]InstalledAppsResponse, error) {
	installedApps, err := impl.installedAppRepository.GetAllIntalledAppsByAppStoreId(appStoreId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}
	var installedAppsEnvResponse []InstalledAppsResponse
	for _, a := range installedApps {
		status, err := impl.getACDStatus(a, w, r, token)
		if apiErr, ok := err.(*util.ApiError); ok {
			if apiErr.Code == constants.AppDetailResourceTreeNotFound {
				status = "Not Found"
			}
		} else if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
		installedAppRes := InstalledAppsResponse{
			EnvironmentName:              a.EnvironmentName,
			AppName:                      a.AppName,
			DeployedAt:                   a.UpdatedOn,
			DeployedBy:                   a.EmailId,
			Status:                       status,
			AppStoreApplicationVersionId: a.AppStoreApplicationVersionId,
			InstalledAppVersionId:        a.InstalledAppVersionId,
			InstalledAppsId:              a.InstalledAppId,
			EnvironmentId:                a.EnvironmentId,
		}
		installedAppsEnvResponse = append(installedAppsEnvResponse, installedAppRes)
	}
	return installedAppsEnvResponse, nil
}

func (impl InstalledAppServiceImpl) getACDStatus(a appstore.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error) {
	if len(a.AppName) > 0 && len(a.EnvironmentName) > 0 {
		acdAppName := a.AppName + "-" + a.EnvironmentName
		query := &application.ResourcesQuery{
			ApplicationName: &acdAppName,
		}
		ctx, cancel := context.WithCancel(r.Context())
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}
		ctx = context.WithValue(ctx, "token", token)
		defer cancel()
		impl.logger.Debugf("Getting status for app %s in env %s", a.AppName, a.EnvironmentName)
		start := time.Now()
		resp, err := impl.acdClient.ResourceTree(ctx, query)
		elapsed := time.Since(start)
		impl.logger.Debugf("Time elapsed %s in fetching application %s for environment %s", elapsed, a.AppName, a.EnvironmentName)
		if err != nil {
			impl.logger.Errorw("error fetching resource tree", "error", err)
			err = &util.ApiError{
				Code:            constants.AppDetailResourceTreeNotFound,
				InternalMessage: "app detail fetched, failed to get resource tree from acd",
				UserMessage:     "app detail fetched, failed to get resource tree from acd",
			}
			return "", err

		}
		return resp.Status, nil
	}
	return "", errors.New("invalid app name or env name")
}

//converts db object to bean
func (impl InstalledAppServiceImpl) chartAdaptor(chart *appstore.InstalledAppVersions) (*InstallAppVersionDTO, error) {

	return &InstallAppVersionDTO{
		InstalledAppId:     chart.InstalledAppId,
		Id:                 chart.Id,
		AppStoreVersion:    chart.AppStoreApplicationVersionId,
		ValuesOverrideYaml: chart.ValuesYaml,
	}, nil
}

//converts db object to bean
func (impl InstalledAppServiceImpl) chartAdaptor2(chart *appstore.InstalledApps) (*InstallAppVersionDTO, error) {

	return &InstallAppVersionDTO{
		EnvironmentId: chart.EnvironmentId,
		Id:            chart.Id,
		AppId:         chart.AppId,
	}, nil
}

func (impl InstalledAppServiceImpl) registerInArgo(chartGitAttribute *util.ChartGitAttribute, ctx context.Context) error {
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

func (impl InstalledAppServiceImpl) createInArgo(chartGitAttribute *util.ChartGitAttribute, ctx context.Context, envModel repository5.Environment, argocdAppName string) error {
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

func (impl InstalledAppServiceImpl) CheckAppExists(appNames []*AppNames) ([]*AppNames, error) {
	if len(appNames) == 0 {
		return nil, nil
	}
	var names []string
	for _, appName := range appNames {
		names = append(names, appName.Name)
	}

	apps, err := impl.appRepository.CheckAppExists(names)
	if err != nil {
		return nil, err
	}
	existingApps := make(map[string]bool)
	for _, app := range apps {
		existingApps[app.AppName] = true
	}
	for _, appName := range appNames {
		if _, ok := existingApps[appName.Name]; ok {
			appName.Exists = true
			appName.SuggestedName = strings.ToLower(randomdata.SillyName())
		}
	}
	return appNames, nil
}

func (impl InstalledAppServiceImpl) createAppForAppStore(createRequest *bean.CreateAppDTO, tx *pg.Tx) (*bean.CreateAppDTO, error) {
	app1, err := impl.appRepository.FindActiveByName(createRequest.AppName)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if app1 != nil && app1.Id > 0 {
		impl.logger.Infow(" app already exists", "name", createRequest.AppName)
		err = &util.ApiError{
			Code:            constants.AppAlreadyExists.Code,
			InternalMessage: "app already exists",
			UserMessage:     fmt.Sprintf("app already exists with name %s", createRequest.AppName),
		}
		return nil, err
	}
	pg := &app.App{
		Active:   true,
		AppName:  createRequest.AppName,
		TeamId:   createRequest.TeamId,
		AppStore: true,
		AuditLog: sql.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err = impl.appRepository.SaveWithTxn(pg, tx)
	if err != nil {
		impl.logger.Errorw("error in saving entity ", "entity", pg)
		return nil, err
	}

	// if found more than 1 application, soft delete all except first item
	apps, err := impl.appRepository.FindActiveListByName(createRequest.AppName)
	if err != nil {
		return nil, err
	}
	appLen := len(apps)
	if appLen > 1 {
		firstElement := apps[0]
		if firstElement.Id != pg.Id {
			pg.Active = false
			err = impl.appRepository.UpdateWithTxn(pg, tx)
			if err != nil {
				impl.logger.Errorw("error in saving entity ", "entity", pg)
				return nil, err
			}
			err = &util.ApiError{
				Code:            constants.AppAlreadyExists.Code,
				InternalMessage: "app already exists",
				UserMessage:     fmt.Sprintf("app already exists with name %s", createRequest.AppName),
			}
			return nil, err
		}
	}

	createRequest.Id = pg.Id
	return createRequest, nil
}

func (impl InstalledAppServiceImpl) syncACD(acdAppName string, ctx context.Context) {
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

func (impl InstalledAppServiceImpl) deleteACD(acdAppName string, ctx context.Context) error {
	req := new(application.ApplicationDeleteRequest)
	req.Name = &acdAppName
	if ctx == nil {
		impl.logger.Errorw("err in delete ACD for AppStore, ctx is NULL", "acdAppName", acdAppName)
		return fmt.Errorf("context is null")
	}
	if _, err := impl.acdClient.Delete(ctx, req); err != nil {
		impl.logger.Errorw("err in delete ACD for AppStore", "acdAppName", acdAppName, "err", err)
		return err
	}
	return nil
}

func (impl InstalledAppServiceImpl) DeleteInstalledApp(ctx context.Context, installAppVersionRequest *InstallAppVersionDTO) (*InstallAppVersionDTO, error) {

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	app, err := impl.appRepository.FindById(installAppVersionRequest.AppId)
	if err != nil {
		return nil, err
	}
	app.Active = false
	app.UpdatedBy = installAppVersionRequest.UserId
	app.UpdatedOn = time.Now()
	err = impl.appRepository.UpdateWithTxn(app, tx)
	if err != nil {
		impl.logger.Errorw("error in update entity ", "entity", app)
		return nil, err
	}

	model, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "id", installAppVersionRequest.InstalledAppId, "err", err)
		return nil, err
	}
	model.Active = false
	model.UpdatedBy = installAppVersionRequest.UserId
	model.UpdatedOn = time.Now()
	_, err = impl.installedAppRepository.UpdateInstalledApp(model, tx)
	if err != nil {
		impl.logger.Errorw("error while creating install app", "error", err)
		return nil, err
	}
	models, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppId(installAppVersionRequest.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching install app versions", "error", err)
		return nil, err
	}
	for _, item := range models {
		item.Active = false
		item.UpdatedBy = installAppVersionRequest.UserId
		item.UpdatedOn = time.Now()
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(item, tx)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
	}

	acdAppName := app.AppName + "-" + environment.Name
	err = impl.deleteACD(acdAppName, ctx)
	if err != nil {
		impl.logger.Errorw("error in deleting ACD ", "name", acdAppName, "err", err)
		if installAppVersionRequest.ForceDelete {
			impl.logger.Warnw("error while deletion of app in acd, continue to delete in db as this operation is force delete", "error", err)
		} else {
			//statusError, _ := err.(*errors2.StatusError)
			if strings.Contains(err.Error(), "code = NotFound") {
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
			return nil, err
		}
	}
	deployment, err := impl.chartGroupDeploymentRepository.FindByInstalledAppId(model.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching chartGroupMapping", "id", model.Id, "err", err)
		return nil, err
	} else if err == pg.ErrNoRows {
		impl.logger.Infow("not a chart group deployment skipping chartGroupMapping delete", "id", model.Id)
	} else {
		deployment.Deleted = true
		deployment.UpdatedOn = time.Now()
		deployment.UpdatedBy = installAppVersionRequest.UserId
		_, err := impl.chartGroupDeploymentRepository.Update(deployment, tx)
		if err != nil {
			impl.logger.Errorw("error in mapping delete", "err", err)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commit db transaction on delete", "err", err)
		return nil, err
	}
	return installAppVersionRequest, nil
}

func (impl InstalledAppServiceImpl) DeployBulk(chartGroupInstallRequest *ChartGroupInstallRequest) (*ChartGroupInstallAppRes, error) {
	impl.logger.Debugw("bulk app install request", "req", chartGroupInstallRequest)
	//save in db
	// raise nats event

	var installAppVersionDTOList []*InstallAppVersionDTO
	for _, chartGroupInstall := range chartGroupInstallRequest.ChartGroupInstallChartRequest {
		installAppVersionDTO, err := impl.requestBuilderForBulkDeployment(chartGroupInstall, chartGroupInstallRequest.ProjectId, chartGroupInstallRequest.UserId)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error in request builder", "err", err)
			return nil, err
		}
		installAppVersionDTOList = append(installAppVersionDTOList, installAppVersionDTO)
	}
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	var installAppVersions []*InstallAppVersionDTO
	// Rollback tx on error.
	defer tx.Rollback()
	for _, installAppVersionDTO := range installAppVersionDTOList {
		installAppVersionDTO, err = impl.AppStoreDeployOperationDB(installAppVersionDTO, tx)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error while app store deploy db operation", "err", err)
			return nil, err
		}
		installAppVersions = append(installAppVersions, installAppVersionDTO)
	}
	if chartGroupInstallRequest.ChartGroupId > 0 {
		groupINstallationId, err := impl.getInstallationId(installAppVersions)
		if err != nil {
			return nil, err
		}
		for _, installAppVersionDTO := range installAppVersions {
			chartGroupEntry := impl.createChartGroupEntryObject(installAppVersionDTO, chartGroupInstallRequest.ChartGroupId, groupINstallationId)
			err := impl.chartGroupDeploymentRepository.Save(tx, chartGroupEntry)
			if err != nil {
				impl.logger.Errorw("DeployBulk, error in creating ChartGroupEntryObject", "err", err)
				return nil, err
			}
		}
	}
	//commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("DeployBulk, error in tx commit", "err", err)
		return nil, err
	}
	//nats event
	impl.triggerDeploymentEvent(installAppVersions)
	return &ChartGroupInstallAppRes{}, nil
}

//generate unique installation ID using APPID
func (impl InstalledAppServiceImpl) getInstallationId(installAppVersions []*InstallAppVersionDTO) (string, error) {
	var buffer bytes.Buffer
	for _, installAppVersionDTO := range installAppVersions {
		if installAppVersionDTO.AppId == 0 {
			return "", fmt.Errorf("app ID not present")
		}
		buffer.WriteString(
			strconv.Itoa(installAppVersionDTO.AppId))
	}
	/* #nosec */
	h := sha1.New()
	_, err := h.Write([]byte(buffer.String()))
	if err != nil {
		return "", err
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs), nil
}

func (impl InstalledAppServiceImpl) createChartGroupEntryObject(installAppVersionDTO *InstallAppVersionDTO, chartGroupId int, groupINstallationId string) *chartGroup.ChartGroupDeployment {
	return &chartGroup.ChartGroupDeployment{
		ChartGroupId:        chartGroupId,
		ChartGroupEntryId:   installAppVersionDTO.ChartGroupEntryId,
		InstalledAppId:      installAppVersionDTO.InstalledAppId,
		Deleted:             false,
		GroupInstallationId: groupINstallationId,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: installAppVersionDTO.UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: installAppVersionDTO.UserId,
		},
	}
}

func (impl InstalledAppServiceImpl) performDeployStage(installedAppVersionId int) (*InstallAppVersionDTO, error) {
	ctx, err := impl.tokenCache.BuildACDSynchContext()
	if err != nil {
		return nil, err
	}
	/*installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersion(installedAppVersionId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}*/

	installedAppVersion, err := impl.GetInstalledAppVersion(installedAppVersionId)
	if err != nil {
		return nil, err
	}
	chartGitAttr := &util.ChartGitAttribute{}
	if installedAppVersion.Status == appstore.DEPLOY_INIT ||
		installedAppVersion.Status == appstore.ENQUEUED ||
		installedAppVersion.Status == appstore.QUE_ERROR ||
		installedAppVersion.Status == appstore.GIT_ERROR {
		//step 2 git operation pull push
		installAppVersionRequest, chartGitAttrDB, err := impl.AppStoreDeployOperationGIT(installedAppVersion)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			_, err = impl.AppStoreDeployOperationStatusUpdate(installAppVersionRequest.InstalledAppId, appstore.GIT_ERROR)
			if err != nil {
				impl.logger.Errorw(" error", "err", err)
				return nil, err
			}
			return nil, err
		}
		impl.logger.Infow("GIT SUCCESSFUL", "chartGitAttrDB", chartGitAttrDB)
		_, err = impl.AppStoreDeployOperationStatusUpdate(installAppVersionRequest.InstalledAppId, appstore.GIT_SUCCESS)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			return nil, err
		}
		chartGitAttr.RepoUrl = chartGitAttrDB.RepoUrl
		chartGitAttr.ChartLocation = chartGitAttrDB.ChartLocation
	} else {
		impl.logger.Infow("DB and GIT operation already done for this app and env, proceed for further step", "installedAppId", installedAppVersion.InstalledAppId, "existing status", installedAppVersion.Status)
		environment, err := impl.environmentRepository.FindById(installedAppVersion.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("fetching error", "err", err)
			return nil, err
		}
		gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
		if err != nil {
			if err == pg.ErrNoRows {
				gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
				gitOpsConfigBitbucket.BitBucketProjectKey = ""
			} else {
				return nil, err
			}
		}
		bitbucketRepoOptions := &bitbucket.RepositoryOptions{
			Owner:    gitOpsConfigBitbucket.BitBucketWorkspaceId,
			Project:  gitOpsConfigBitbucket.BitBucketProjectKey,
			RepoSlug: installedAppVersion.AppStoreName,
		}
		repoUrl, err := impl.gitFactory.Client.GetRepoUrl(installedAppVersion.AppStoreName, bitbucketRepoOptions)
		if err != nil {
			//will allow to continue to persist status on next operation
			impl.logger.Errorw("fetching error", "err", err)
		}
		chartGitAttr.RepoUrl = repoUrl
		chartGitAttr.ChartLocation = fmt.Sprintf("%s-%s", installedAppVersion.AppName, environment.Name)
		installedAppVersion.ACDAppName = fmt.Sprintf("%s-%s", installedAppVersion.AppName, environment.Name)
		installedAppVersion.Environment = environment
	}

	if installedAppVersion.Status == appstore.DEPLOY_INIT ||
		installedAppVersion.Status == appstore.ENQUEUED ||
		installedAppVersion.Status == appstore.QUE_ERROR ||
		installedAppVersion.Status == appstore.GIT_ERROR ||
		installedAppVersion.Status == appstore.GIT_SUCCESS ||
		installedAppVersion.Status == appstore.ACD_ERROR {
		//step 3 acd operation register, sync
		_, err = impl.AppStoreDeployOperationACD(installedAppVersion, chartGitAttr, ctx)
		if err != nil {
			impl.logger.Errorw(" error", "chartGitAttr", chartGitAttr, "err", err)
			_, err = impl.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appstore.ACD_ERROR)
			if err != nil {
				impl.logger.Errorw(" error", "err", err)
				return nil, err
			}
			return nil, err
		}
		impl.logger.Infow("ACD SUCCESSFUL", "chartGitAttr", chartGitAttr)
		_, err = impl.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appstore.ACD_SUCCESS)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Infow("DB and GIT and ACD operation already done for this app and env. process has been completed", "installedAppId", installedAppVersion.InstalledAppId, "existing status", installedAppVersion.Status)
	}
	//step 4 db operation status triggered
	_, err = impl.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appstore.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}
	return installedAppVersion, nil
}

func (impl InstalledAppServiceImpl) requestBuilderForBulkDeployment(installRequest *ChartGroupInstallChartRequest, projectId int, userId int32) (*InstallAppVersionDTO, error) {
	valYaml := installRequest.ValuesOverrideYaml
	if valYaml == "" {
		valVersion, err := impl.appStoreValuesService.FindValuesByIdAndKind(installRequest.ReferenceValueId, installRequest.ReferenceValueKind)
		if err != nil {
			return nil, err
		}
		valYaml = valVersion.Values
	}
	req := &InstallAppVersionDTO{
		AppName:                 installRequest.AppName,
		TeamId:                  projectId,
		EnvironmentId:           installRequest.EnvironmentId,
		AppStoreVersion:         installRequest.AppStoreVersion,
		ValuesOverrideYaml:      valYaml,
		UserId:                  userId,
		ReferenceValueId:        installRequest.ReferenceValueId,
		ReferenceValueKind:      installRequest.ReferenceValueKind,
		ChartGroupEntryId:       installRequest.ChartGroupEntryId,
		DefaultClusterComponent: installRequest.DefaultClusterComponent,
	}
	return req, nil
}

func (impl InstalledAppServiceImpl) CreateInstalledAppV2(installAppVersionRequest *InstallAppVersionDTO, ctx context.Context) (*InstallAppVersionDTO, error) {

	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//step 1 db operation initiated
	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}

	//step 2 git operation pull push
	installAppVersionRequest, chartGitAttr, err := impl.AppStoreDeployOperationGIT(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}

	//step 3 acd operation register, sync
	installAppVersionRequest, err = impl.AppStoreDeployOperationACD(installAppVersionRequest, chartGitAttr, ctx)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}

	// tx commit here because next operation will be process after this commit.
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//step 4 db operation status update to deploy success
	_, err = impl.AppStoreDeployOperationStatusUpdate(installAppVersionRequest.InstalledAppId, appstore.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}

	return installAppVersionRequest, nil
}

func (impl InstalledAppServiceImpl) AppStoreDeployOperationGIT(installAppVersionRequest *InstallAppVersionDTO) (*InstallAppVersionDTO, *util.ChartGitAttribute, error) {
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
	template := CHART_PROXY_TEMPLATE
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
	_, chartGitAttr, err := impl.chartTemplateService.CreateChartProxy(chartMeta, chartPath, template, appStoreAppVersion.Version, environment.Name, installAppVersionRequest.AppName)
	if err != nil {
		return nil, nil, err
	}

	//STEP 3 - update requirements and values

	//update requirements yaml in chart
	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	dependency := Dependency{
		Name:       appStoreAppVersion.AppStore.Name,
		Version:    appStoreAppVersion.Version,
		Repository: appStoreAppVersion.AppStore.ChartRepo.Url,
	}
	var dependencies []Dependency
	dependencies = append(dependencies, dependency)
	requirementDependencies := &Dependencies{
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
	chartGitAttrForRequirement := &util.ChartConfig{
		FileName:       REQUIREMENTS_YAML_FILE,
		FileContent:    string(requirementDependenciesByte),
		ChartName:      chartMeta.Name,
		ChartLocation:  argocdAppName,
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
	_, err = impl.gitFactory.Client.CommitValues(chartGitAttrForRequirement, gitOpsConfigBitbucket.BitBucketWorkspaceId)
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

	valuesYaml := &util.ChartConfig{
		FileName:       VALUES_YAML_FILE,
		FileContent:    string(valuesByte),
		ChartName:      chartMeta.Name,
		ChartLocation:  argocdAppName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
	}
	_, err = impl.gitFactory.Client.CommitValues(valuesYaml, gitOpsConfigBitbucket.BitBucketWorkspaceId)
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
	installAppVersionRequest.ACDAppName = argocdAppName
	installAppVersionRequest.Environment = environment
	return installAppVersionRequest, chartGitAttr, nil
}

func (impl InstalledAppServiceImpl) patchAcdApp(installAppVersionRequest *InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context) (*InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//STEP 4: registerInArgo
	err := impl.registerInArgo(chartGitAttr, ctx)
	if err != nil {
		impl.logger.Errorw("error in argo registry", "err", err)
		return nil, err
	}
	// update acd app
	patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: v1alpha1.ApplicationSource{Path: chartGitAttr.ChartLocation, RepoURL: chartGitAttr.RepoUrl}}}
	reqbyte, err := json.Marshal(patchReq)
	if err != nil {
		impl.logger.Errorw("error in creating patch", "err", err)
	}
	_, err = impl.acdClient.Patch(ctx, &application.ApplicationPatchRequest{Patch: string(reqbyte), Name: &installAppVersionRequest.ACDAppName, PatchType: "merge"})
	if err != nil {
		impl.logger.Errorw("error in creating argo app ", "name", installAppVersionRequest.ACDAppName, "patch", string(reqbyte), "err", err)
		return nil, err
	}
	impl.syncACD(installAppVersionRequest.ACDAppName, ctx)
	return installAppVersionRequest, nil
}

func (impl InstalledAppServiceImpl) AppStoreDeployOperationACD(installAppVersionRequest *InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context) (*InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//STEP 4: registerInArgo
	err := impl.registerInArgo(chartGitAttr, ctx)
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
	impl.syncACD(installAppVersionRequest.ACDAppName, ctx)

	return installAppVersionRequest, nil
}

func (impl InstalledAppServiceImpl) AppStoreDeployOperationDB(installAppVersionRequest *InstallAppVersionDTO, tx *pg.Tx) (*InstallAppVersionDTO, error) {

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	appCreateRequest := &bean.CreateAppDTO{
		Id:      installAppVersionRequest.AppId,
		AppName: installAppVersionRequest.AppName,
		TeamId:  installAppVersionRequest.TeamId,
		UserId:  installAppVersionRequest.UserId,
	}

	appCreateRequest, err = impl.createAppForAppStore(appCreateRequest, tx)
	if err != nil {
		impl.logger.Errorw("error while creating app", "error", err)
		return nil, err
	}
	installAppVersionRequest.AppId = appCreateRequest.Id

	installedAppModel := &appstore.InstalledApps{
		AppId:         appCreateRequest.Id,
		EnvironmentId: environment.Id,
		Status:        appstore.DEPLOY_INIT,
	}
	installedAppModel.CreatedBy = installAppVersionRequest.UserId
	installedAppModel.UpdatedBy = installAppVersionRequest.UserId
	installedAppModel.CreatedOn = time.Now()
	installedAppModel.UpdatedOn = time.Now()
	installedAppModel.Active = true
	installedApp, err := impl.installedAppRepository.CreateInstalledApp(installedAppModel, tx)
	if err != nil {
		impl.logger.Errorw("error while creating install app", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppId = installedApp.Id

	installedAppVersions := &appstore.InstalledAppVersions{
		InstalledAppId:               installAppVersionRequest.InstalledAppId,
		AppStoreApplicationVersionId: appStoreAppVersion.Id,
		ValuesYaml:                   installAppVersionRequest.ValuesOverrideYaml,
		//Values:                       "{}",
	}
	installedAppVersions.CreatedBy = installAppVersionRequest.UserId
	installedAppVersions.UpdatedBy = installAppVersionRequest.UserId
	installedAppVersions.CreatedOn = time.Now()
	installedAppVersions.UpdatedOn = time.Now()
	installedAppVersions.Active = true
	installedAppVersions.ReferenceValueId = installAppVersionRequest.ReferenceValueId
	installedAppVersions.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
	_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersions, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppVersionId = installedAppVersions.Id

	if installAppVersionRequest.DefaultClusterComponent {
		clusterInstalledAppsModel := &appstore.ClusterInstalledApps{
			ClusterId:      environment.ClusterId,
			InstalledAppId: installAppVersionRequest.InstalledAppId,
		}
		clusterInstalledAppsModel.CreatedBy = installAppVersionRequest.UserId
		clusterInstalledAppsModel.UpdatedBy = installAppVersionRequest.UserId
		clusterInstalledAppsModel.CreatedOn = time.Now()
		clusterInstalledAppsModel.UpdatedOn = time.Now()
		err = impl.clusterInstalledAppsRepository.Save(clusterInstalledAppsModel, tx)
		if err != nil {
			impl.logger.Errorw("error while creating cluster install app", "error", err)
			return nil, err
		}
	}
	return installAppVersionRequest, nil
}

func (impl InstalledAppServiceImpl) AppStoreDeployOperationStatusUpdate(installAppId int, status appstore.AppstoreDeploymentStatus) (bool, error) {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	installedApp, err := impl.installedAppRepository.GetInstalledApp(installAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	installedApp.Status = status
	_, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while commit db transaction to db", "error", err)
		return false, err
	}
	return true, nil
}

//------------ nats config

func (impl *InstalledAppServiceImpl) triggerDeploymentEvent(installAppVersions []*InstallAppVersionDTO) {

	for _, versions := range installAppVersions {
		var status appstore.AppstoreDeploymentStatus
		payload := &DeployPayload{InstalledAppVersionId: versions.InstalledAppVersionId}
		data, err := json.Marshal(payload)
		if err != nil {
			status = appstore.QUE_ERROR
		} else {
			err := impl.pubsubClient.Conn.Publish(BULK_APPSTORE_DEPLOY_TOPIC, data)
			if err != nil {
				impl.logger.Errorw("err while publishing msg for app-store bulk deploy", "msg", data, "err", err)
				status = appstore.QUE_ERROR
			} else {
				status = appstore.ENQUEUED
			}

		}
		if versions.Status == appstore.DEPLOY_INIT || versions.Status == appstore.QUE_ERROR || versions.Status == appstore.ENQUEUED {
			impl.logger.Debugw("status for bulk app-store deploy", "status", status)
			_, err = impl.AppStoreDeployOperationStatusUpdate(payload.InstalledAppVersionId, status)
			if err != nil {
				impl.logger.Errorw("error while bulk app-store deploy status update", "err", err)
			}
		}
	}
}

func (impl *InstalledAppServiceImpl) Subscribe() error {
	_, err := impl.pubsubClient.Conn.QueueSubscribe(BULK_APPSTORE_DEPLOY_TOPIC, BULK_APPSTORE_DEPLOY_GROUP, func(msg *stan.Msg) {
		impl.logger.Debug("cd stage event received")
		defer msg.Ack()
		deployPayload := &DeployPayload{}
		err := json.Unmarshal([]byte(string(msg.Data)), &deployPayload)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}
		impl.logger.Debugw("deployPayload:", "deployPayload", deployPayload)
		_, err = impl.performDeployStage(deployPayload.InstalledAppVersionId)
		if err != nil {
			impl.logger.Errorw("error in performing deploy stage", "deployPayload", deployPayload, "err", err)
		}
	}, stan.DurableName(BULK_APPSTORE_DEPLOY_DURABLE), stan.StartWithLastReceived(), stan.AckWait(time.Duration(200)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(3))
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *InstalledAppServiceImpl) DeployDefaultChartOnCluster(bean *cluster2.ClusterBean, userId int32) (bool, error) {
	// STEP 1 - create environment with name "devton"
	impl.logger.Infow("STEP 1", "create environment for cluster component", bean)
	envName := fmt.Sprintf("%s-%s", bean.ClusterName, DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT)
	env, err := impl.envService.FindOne(envName)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	if err == pg.ErrNoRows {
		env = &cluster2.EnvironmentBean{
			Environment: envName,
			ClusterId:   bean.Id,
			Namespace:   envName,
			Default:     false,
			Active:      true,
		}
		_, err := impl.envService.Create(env, userId)
		if err != nil {
			impl.logger.Errorw("DeployDefaultChartOnCluster, error in creating environment", "data", env, "err", err)
			return false, err
		}
	}

	// STEP 2 - create project with name "devtron"
	impl.logger.Info("STEP 2", "create project for cluster components")
	t, err := impl.teamRepository.FindByTeamName(DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	if err == pg.ErrNoRows {
		t := &repository4.Team{
			Name:     DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT,
			Active:   true,
			AuditLog: sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		err = impl.teamRepository.Save(t)
		if err != nil {
			impl.logger.Errorw("DeployDefaultChartOnCluster, error in creating team", "data", t, "err", err)
			return false, err
		}
	}

	// STEP 3- read the input data from env variables
	impl.logger.Info("STEP 3", "read the input data from env variables")
	charts := &ChartComponents{}
	var chartComponents []*ChartComponent
	if _, err := os.Stat(CLUSTER_COMPONENT_DIR_PATH); os.IsNotExist(err) {
		impl.logger.Infow("default cluster component directory error", "cluster", bean.ClusterName, "err", err)
		return false, nil
	} else {
		fileInfo, err := ioutil.ReadDir(CLUSTER_COMPONENT_DIR_PATH)
		if err != nil {
			impl.logger.Errorw("DeployDefaultChartOnCluster, err while reading directory", "err", err)
			return false, err
		}
		for _, file := range fileInfo {
			impl.logger.Infow("file", "name", file.Name())
			if strings.Contains(file.Name(), ".yaml") {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", CLUSTER_COMPONENT_DIR_PATH, file.Name()))
				if err != nil {
					impl.logger.Errorw("DeployDefaultChartOnCluster, error on reading file", "err", err)
					return false, err
				}
				chartComponent := &ChartComponent{
					Name:   strings.ReplaceAll(file.Name(), ".yaml", ""),
					Values: string(content),
				}
				chartComponents = append(chartComponents, chartComponent)
			}
		}

		if len(chartComponents) > 0 {
			charts.ChartComponent = chartComponents
			impl.logger.Info("STEP 4 - prepare a bulk request")
			// STEP 4 - prepare a bulk request (unique names need to apply for deploying chart)
			// STEP 4.1 - fetch chart for required name(actual chart name (app-store)) with default values
			// STEP 4.2 - update all the required charts, override values.yaml with env variables.
			chartGroupInstallRequest := &ChartGroupInstallRequest{}
			chartGroupInstallRequest.ProjectId = t.Id
			chartGroupInstallRequest.UserId = userId
			var chartGroupInstallChartRequests []*ChartGroupInstallChartRequest
			for _, item := range charts.ChartComponent {
				appStore, err := impl.appStoreApplicationVersionRepository.FindByAppStoreName(item.Name)
				if err != nil {
					impl.logger.Errorw("DeployDefaultChartOnCluster, error in getting app store", "data", t, "err", err)
					return false, err
				}
				chartGroupInstallChartRequest := &ChartGroupInstallChartRequest{
					AppName:                 fmt.Sprintf("%s-%s-%s", bean.ClusterName, env.Environment, item.Name),
					EnvironmentId:           env.Id,
					ValuesOverrideYaml:      item.Values,
					AppStoreVersion:         appStore.AppStoreApplicationVersionId,
					ReferenceValueId:        appStore.AppStoreApplicationVersionId,
					ReferenceValueKind:      REFERENCE_TYPE_DEFAULT,
					DefaultClusterComponent: true,
				}
				chartGroupInstallChartRequests = append(chartGroupInstallChartRequests, chartGroupInstallChartRequest)
			}
			chartGroupInstallRequest.ChartGroupInstallChartRequest = chartGroupInstallChartRequests

			impl.logger.Info("STEP 5 - deploy bulk initiated")
			// STEP 5 - deploy
			_, err = impl.DeployDefaultComponent(chartGroupInstallRequest)
			if err != nil {
				impl.logger.Errorw("DeployDefaultChartOnCluster, error on bulk deploy", "err", err)
				return false, err
			}
		}
	}
	return true, nil
}

type ChartComponents struct {
	ChartComponent []*ChartComponent `json:"charts"`
}
type ChartComponent struct {
	Name   string `json:"name"`
	Values string `json:"values"`
}

func (impl *InstalledAppServiceImpl) IsChartRepoActive(appStoreVersionId int) (bool, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(appStoreVersionId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return false, err
	}
	return appStoreAppVersion.AppStore.ChartRepo.Active, nil
}

func (impl InstalledAppServiceImpl) DeployDefaultComponent(chartGroupInstallRequest *ChartGroupInstallRequest) (*ChartGroupInstallAppRes, error) {
	impl.logger.Debugw("bulk app install request", "req", chartGroupInstallRequest)
	//save in db
	// raise nats event

	var installAppVersionDTOList []*InstallAppVersionDTO
	for _, chartGroupInstall := range chartGroupInstallRequest.ChartGroupInstallChartRequest {
		installAppVersionDTO, err := impl.requestBuilderForBulkDeployment(chartGroupInstall, chartGroupInstallRequest.ProjectId, chartGroupInstallRequest.UserId)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error in request builder", "err", err)
			return nil, err
		}
		installAppVersionDTOList = append(installAppVersionDTOList, installAppVersionDTO)
	}
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	var installAppVersions []*InstallAppVersionDTO
	// Rollback tx on error.
	defer tx.Rollback()
	for _, installAppVersionDTO := range installAppVersionDTOList {
		installAppVersionDTO, err = impl.AppStoreDeployOperationDB(installAppVersionDTO, tx)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error while app store deploy db operation", "err", err)
			return nil, err
		}
		installAppVersions = append(installAppVersions, installAppVersionDTO)
	}
	if chartGroupInstallRequest.ChartGroupId > 0 {
		groupINstallationId, err := impl.getInstallationId(installAppVersions)
		if err != nil {
			return nil, err
		}
		for _, installAppVersionDTO := range installAppVersions {
			chartGroupEntry := impl.createChartGroupEntryObject(installAppVersionDTO, chartGroupInstallRequest.ChartGroupId, groupINstallationId)
			err := impl.chartGroupDeploymentRepository.Save(tx, chartGroupEntry)
			if err != nil {
				impl.logger.Errorw("DeployBulk, error in creating ChartGroupEntryObject", "err", err)
				return nil, err
			}
		}
	}
	//commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("DeployBulk, error in tx commit", "err", err)
		return nil, err
	}
	//nats event

	for _, versions := range installAppVersions {
		_, err := impl.performDeployStage(versions.InstalledAppVersionId)
		if err != nil {
			impl.logger.Errorw("error in performing deploy stage", "deployPayload", versions, "err", err)
			_, err = impl.AppStoreDeployOperationStatusUpdate(versions.InstalledAppVersionId, appstore.QUE_ERROR)
			if err != nil {
				impl.logger.Errorw("error while bulk app-store deploy status update", "err", err)
			}
		}
	}

	return &ChartGroupInstallAppRes{}, nil
}
