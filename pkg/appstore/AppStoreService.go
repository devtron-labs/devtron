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
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/ghodss/yaml"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/version"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const ValidationSuccessMsg = "Configurations are validated successfully"

type DetailedErrorHelmRepoValidation struct {
	CustomErrMsg string `json:"customErrMsg"`
	ActualErrMsg string `json:"actualErrMsg"`
}
type AppStoreApplication struct {
	Id                          int                                   `json:"id"`
	Name                        string                                `json:"name"`
	ChartRepoId                 int                                   `json:"chartRepoId"`
	Active                      bool                                  `json:"active"`
	ChartGitLocation            string                                `json:"chartGitLocation"`
	CreatedOn                   time.Time                             `json:"createdOn"`
	UpdatedOn                   time.Time                             `json:"updatedOn"`
	AppStoreApplicationVersions []*AppStoreApplicationVersionResponse `json:"appStoreApplicationVersions"`
}

type AppStoreApplicationVersionResponse struct {
	Id                      int       `json:"id"`
	Version                 string    `json:"version"`
	AppVersion              string    `json:"appVersion"`
	Created                 time.Time `json:"created"`
	Deprecated              bool      `json:"deprecated"`
	Description             string    `json:"description"`
	Digest                  string    `json:"digest"`
	Icon                    string    `json:"icon"`
	Name                    string    `json:"name"`
	ChartName               string    `json:"chartName"`
	AppStoreApplicationName string    `json:"appStoreApplicationName"`
	Home                    string    `json:"home"`
	Source                  string    `json:"source"`
	ValuesYaml              string    `json:"valuesYaml"`
	ChartYaml               string    `json:"chartYaml"`
	AppStoreId              int       `json:"appStoreId"`
	Latest                  bool      `json:"latest"`
	CreatedOn               time.Time `json:"createdOn"`
	RawValues               string    `json:"rawValues"`
	Readme                  string    `json:"readme"`
	UpdatedOn               time.Time `json:"updatedOn"`
	IsChartRepoActive       bool      `json:"isChartRepoActive"`
}

type HelmRepositoriesData struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type KeyDto struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
	Url  string `json:"url,omitempty"`
}

type AcdConfigMapRepositoriesDto struct {
	Type           string  `json:"type,omitempty"`
	Name           string  `json:"name,omitempty"`
	Url            string  `json:"url,omitempty"`
	UsernameSecret *KeyDto `json:"usernameSecret,omitempty"`
	PasswordSecret *KeyDto `json:"passwordSecret,omitempty"`
	CaSecret       *KeyDto `json:"caSecret,omitempty"`
	CertSecret     *KeyDto `json:"certSecret,omitempty"`
	KeySecret      *KeyDto `json:"keySecret,omitempty"`
}

type ConfigMapDataDto struct {
	HelmRepositories string `json:"helm.repositories,omitempty"`
	Repositories     string `json:"repositories,omitempty"`
}

type AppStoreService interface {
	FindAllApps(filter *appstore.AppStoreFilter) ([]appstore.AppStoreWithVersion, error)
	FindChartDetailsById(id int) (AppStoreApplicationVersionResponse, error)
	FindChartVersionsByAppStoreId(appStoreId int) ([]AppStoreVersionsResponse, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean.AppDetailContainer, error)
	GetReadMeByAppStoreApplicationVersionId(id int) (*ReadmeRes, error)

	SearchAppStoreChartByName(chartName string) ([]*appstore.ChartRepoSearch, error)
	CreateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error)
	UpdateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error)
	GetChartRepoById(id int) (*ChartRepoDto, error)
	GetChartRepoList() ([]*ChartRepoDto, error)
	ValidateChartRepo(request *ChartRepoDto) *DetailedErrorHelmRepoValidation
	ValidateAndCreateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error, *DetailedErrorHelmRepoValidation)
	ValidateAndUpdateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error, *DetailedErrorHelmRepoValidation)
	TriggerChartSyncManual() error
}

type AppStoreVersionsResponse struct {
	Version string `json:"version"`
	Id      int    `json:"id"`
}

type ReadmeRes struct {
	AppStoreApplicationVersionId int    `json:"appStoreApplicationVersionId"`
	Readme                       string `json:"readme"`
}
type AppStoreServiceImpl struct {
	logger                        *zap.SugaredLogger
	appStoreRepository            appstore.AppStoreRepository
	appStoreApplicationRepository appstore.AppStoreApplicationVersionRepository
	installedAppRepository        appstore.InstalledAppRepository
	userService                   user.UserService
	repoRepository                chartConfig.ChartRepoRepository
	K8sUtil                       *util.K8sUtil
	clusterService                cluster.ClusterService
	envService                    cluster.EnvironmentService
	versionService                argocdServer.VersionService
	aCDAuthConfig                 *util2.ACDAuthConfig
	client                        *http.Client
}

func NewAppStoreServiceImpl(logger *zap.SugaredLogger, appStoreRepository appstore.AppStoreRepository,
	appStoreApplicationRepository appstore.AppStoreApplicationVersionRepository, installedAppRepository appstore.InstalledAppRepository,
	userService user.UserService, repoRepository chartConfig.ChartRepoRepository, K8sUtil *util.K8sUtil,
	clusterService cluster.ClusterService, envService cluster.EnvironmentService,
	versionService argocdServer.VersionService, aCDAuthConfig *util2.ACDAuthConfig, client *http.Client) *AppStoreServiceImpl {
	return &AppStoreServiceImpl{
		logger:                        logger,
		appStoreRepository:            appStoreRepository,
		appStoreApplicationRepository: appStoreApplicationRepository,
		installedAppRepository:        installedAppRepository,
		userService:                   userService,
		repoRepository:                repoRepository,
		K8sUtil:                       K8sUtil,
		clusterService:                clusterService,
		envService:                    envService,
		versionService:                versionService,
		aCDAuthConfig:                 aCDAuthConfig,
		client:                        client,
	}
}

func (impl *AppStoreServiceImpl) FindAllApps(filter *appstore.AppStoreFilter) ([]appstore.AppStoreWithVersion, error) {
	appStoreApplications, err := impl.appStoreApplicationRepository.FindWithFilter(filter)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	return appStoreApplications, nil
}

func (impl *AppStoreServiceImpl) FindChartDetailsById(id int) (AppStoreApplicationVersionResponse, error) {
	chartDetails, err := impl.appStoreApplicationRepository.FindById(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return AppStoreApplicationVersionResponse{}, err
	}
	appStoreApplicationVersion := AppStoreApplicationVersionResponse{
		Id:                      chartDetails.Id,
		Version:                 chartDetails.Version,
		AppVersion:              chartDetails.AppVersion,
		Created:                 chartDetails.Created,
		Deprecated:              chartDetails.Deprecated,
		Description:             chartDetails.Description,
		Digest:                  chartDetails.Digest,
		Icon:                    chartDetails.Icon,
		Name:                    chartDetails.Name,
		ChartName:               chartDetails.AppStore.ChartRepo.Name,
		AppStoreApplicationName: chartDetails.AppStore.Name,
		Home:                    chartDetails.Home,
		Source:                  chartDetails.Source,
		ValuesYaml:              chartDetails.ValuesYaml,
		ChartYaml:               chartDetails.ChartYaml,
		AppStoreId:              chartDetails.AppStoreId,
		Latest:                  chartDetails.Latest,
		CreatedOn:               chartDetails.CreatedOn,
		UpdatedOn:               chartDetails.UpdatedOn,
		RawValues:               chartDetails.RawValues,
		Readme:                  chartDetails.Readme,
		IsChartRepoActive:       chartDetails.AppStore.ChartRepo.Active,
	}
	return appStoreApplicationVersion, nil
}

func (impl *AppStoreServiceImpl) FindChartVersionsByAppStoreId(appStoreId int) ([]AppStoreVersionsResponse, error) {
	appStoreVersions, err := impl.appStoreApplicationRepository.FindVersionsByAppStoreId(appStoreId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}
	var appStoreVersionsResponse []AppStoreVersionsResponse
	for _, a := range appStoreVersions {
		res := AppStoreVersionsResponse{
			Id:      a.Id,
			Version: a.Version,
		}
		appStoreVersionsResponse = append(appStoreVersionsResponse, res)
	}
	return appStoreVersionsResponse, nil
}

func (impl *AppStoreServiceImpl) FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean.AppDetailContainer, error) {
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Error(err)
		return bean.AppDetailContainer{}, err
	}
	deploymentContainer := bean.DeploymentDetailContainer{
		InstalledAppId:                installedAppVerison.InstalledApp.Id,
		AppId:                         installedAppVerison.InstalledApp.App.Id,
		AppStoreInstalledAppVersionId: installedAppVerison.Id,
		EnvironmentId:                 installedAppVerison.InstalledApp.EnvironmentId,
		AppName:                       installedAppVerison.InstalledApp.App.AppName,
		AppStoreChartName:             installedAppVerison.AppStoreApplicationVersion.AppStore.ChartRepo.Name,
		AppStoreChartId:               installedAppVerison.AppStoreApplicationVersion.AppStore.Id,
		AppStoreAppName:               installedAppVerison.AppStoreApplicationVersion.Name,
		AppStoreAppVersion:            installedAppVerison.AppStoreApplicationVersion.Version,
		EnvironmentName:               installedAppVerison.InstalledApp.Environment.Name,
		LastDeployedTime:              installedAppVerison.UpdatedOn.Format(bean2.LayoutRFC3339),
		Namespace:                     installedAppVerison.InstalledApp.Environment.Namespace,
		Deprecated:                    installedAppVerison.AppStoreApplicationVersion.Deprecated,
	}
	userInfo, err := impl.userService.GetByIdIncludeDeleted(installedAppVerison.AuditLog.UpdatedBy)
	if err != nil {
		impl.logger.Errorw("error fetching user info", "err", err)
		return bean.AppDetailContainer{}, err
	}
	deploymentContainer.LastDeployedBy = userInfo.EmailId
	appDetail := bean.AppDetailContainer{
		DeploymentDetailContainer: deploymentContainer,
	}
	return appDetail, nil
}

func (impl *AppStoreServiceImpl) GetReadMeByAppStoreApplicationVersionId(id int) (*ReadmeRes, error) {
	appVersion, err := impl.appStoreApplicationRepository.GetReadMeById(id)
	if err != nil {
		return nil, err
	}
	readme := &ReadmeRes{
		AppStoreApplicationVersionId: appVersion.Id,
		Readme:                       appVersion.Readme,
	}
	return readme, nil
}

func (impl *AppStoreServiceImpl) SearchAppStoreChartByName(chartName string) ([]*appstore.ChartRepoSearch, error) {
	appStoreApplications, err := impl.appStoreApplicationRepository.SearchAppStoreChartByName(chartName)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	return appStoreApplications, nil
}

func (impl *AppStoreServiceImpl) CreateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error) {
	dbConnection := impl.repoRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	chartRepo := &chartConfig.ChartRepo{AuditLog: sql.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId}}
	chartRepo.Name = request.Name
	chartRepo.Url = request.Url
	chartRepo.AuthMode = request.AuthMode
	chartRepo.UserName = request.UserName
	chartRepo.Password = request.Password
	chartRepo.Active = request.Active
	chartRepo.AccessToken = request.AccessToken
	chartRepo.SshKey = request.SshKey
	chartRepo.Active = true
	chartRepo.Default = false
	chartRepo.External = true
	err = impl.repoRepository.Save(chartRepo, tx)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	clusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		return nil, err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return nil, err
	}

	apiVersion, err := impl.versionService.GetVersion()
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	apiMinorVersion, err := strconv.Atoi(apiVersion[3:4])
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return nil, err
	}

	updateSuccess := false
	retryCount := 0
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return nil, err
		}
		data := impl.updateData(cm.Data, request, apiMinorVersion)
		cm.Data = data
		_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			continue
		}
		if err == nil {
			updateSuccess = true
		}
	}
	if !updateSuccess {
		return nil, fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return chartRepo, nil
}

func (impl *AppStoreServiceImpl) UpdateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error) {
	dbConnection := impl.repoRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	chartRepo, err := impl.repoRepository.FindById(request.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	chartRepo.Url = request.Url
	chartRepo.AuthMode = request.AuthMode
	chartRepo.UserName = request.UserName
	chartRepo.Password = request.Password
	chartRepo.Active = request.Active
	chartRepo.AccessToken = request.AccessToken
	chartRepo.SshKey = request.SshKey
	chartRepo.Active = request.Active
	chartRepo.UpdatedBy = request.UserId
	chartRepo.UpdatedOn = time.Now()
	err = impl.repoRepository.Update(chartRepo, tx)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	// modify configmap
	clusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		return nil, err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return nil, err
	}

	apiVersion, err := impl.versionService.GetVersion()
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	apiMinorVersion, err := strconv.Atoi(apiVersion[3:4])
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return nil, err
	}
	updateSuccess := false
	retryCount := 0
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return nil, err
		}
		data := impl.updateData(cm.Data, request, apiMinorVersion)
		cm.Data = data
		_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			impl.logger.Warnw(" config map failed", "err", err)
			continue
		}
		if err == nil {
			impl.logger.Warnw(" config map apply succeeded", "on retryCount", retryCount)
			updateSuccess = true
		}
	}
	if !updateSuccess {
		return nil, fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return chartRepo, nil
}

func (impl *AppStoreServiceImpl) GetChartRepoById(id int) (*ChartRepoDto, error) {
	chartRepo := &ChartRepoDto{}
	model, err := impl.repoRepository.FindById(id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	chartRepo.Id = model.Id
	chartRepo.Name = model.Name
	chartRepo.Url = model.Url
	chartRepo.AuthMode = model.AuthMode
	chartRepo.Password = model.Password
	chartRepo.UserName = model.UserName
	chartRepo.SshKey = model.SshKey
	chartRepo.AccessToken = model.AccessToken
	chartRepo.Default = model.Default
	chartRepo.Active = model.Active
	return chartRepo, nil
}

func (impl *AppStoreServiceImpl) GetChartRepoList() ([]*ChartRepoDto, error) {
	var chartRepos []*ChartRepoDto
	models, err := impl.repoRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, model := range models {
		chartRepo := &ChartRepoDto{}
		chartRepo.Id = model.Id
		chartRepo.Name = model.Name
		chartRepo.Url = model.Url
		chartRepo.AuthMode = model.AuthMode
		chartRepo.Password = model.Password
		chartRepo.UserName = model.UserName
		chartRepo.SshKey = model.SshKey
		chartRepo.AccessToken = model.AccessToken
		chartRepo.Default = model.Default
		chartRepo.Active = model.Active
		chartRepos = append(chartRepos, chartRepo)
	}
	return chartRepos, nil
}

func (impl *AppStoreServiceImpl) updateData(data map[string]string, request *ChartRepoDto, apiMinorVersion int) map[string]string {
	helmRepoStr := data["helm.repositories"]
	helmRepoByte, err := yaml.YAMLToJSON([]byte(helmRepoStr))
	if err != nil {
		panic(err)
	}
	var helmRepositories []*AcdConfigMapRepositoriesDto
	err = json.Unmarshal(helmRepoByte, &helmRepositories)
	if err != nil {
		panic(err)
	}
	if apiMinorVersion < 3 {
		found := false
		for _, item := range helmRepositories {
			//if request chart repo found, than update its values
			if item.Name == request.Name {
				if request.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
					usernameSecret := &KeyDto{Name: request.UserName, Key: "username"}
					passwordSecret := &KeyDto{Name: request.Password, Key: "password"}
					item.PasswordSecret = passwordSecret
					item.UsernameSecret = usernameSecret
				} else if request.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
					// TODO - is it access token or ca cert nd secret
				} else if request.AuthMode == repository.AUTH_MODE_SSH {
					keySecret := &KeyDto{Name: request.SshKey, Key: "key"}
					item.KeySecret = keySecret
				}
				item.Url = request.Url
				found = true
			}
		}

		// if request chart repo not found, add new one
		if !found {
			repoData := impl.createRepoElement(apiMinorVersion, request)
			helmRepositories = append(helmRepositories, repoData)
		}
	}
	rb, err := json.Marshal(helmRepositories)
	if err != nil {
		panic(err)
	}
	helmRepositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		panic(err)
	}

	//SETUP for repositories
	var repositories []*AcdConfigMapRepositoriesDto
	repoStr := data["repositories"]
	repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(repoByte, &repositories)
	if err != nil {
		panic(err)
	}
	if apiMinorVersion >= 3 {
		found := false
		for _, item := range repositories {
			//if request chart repo found, than update its values
			if item.Name == request.Name {
				if request.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
					usernameSecret := &KeyDto{Name: request.UserName, Key: "username"}
					passwordSecret := &KeyDto{Name: request.Password, Key: "password"}
					item.PasswordSecret = passwordSecret
					item.UsernameSecret = usernameSecret
				} else if request.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
					// TODO - is it access token or ca cert nd secret
				} else if request.AuthMode == repository.AUTH_MODE_SSH {
					keySecret := &KeyDto{Name: request.SshKey, Key: "key"}
					item.KeySecret = keySecret
				}
				item.Url = request.Url
				found = true
			}
		}

		// if request chart repo not found, add new one
		if !found {
			repoData := impl.createRepoElement(apiMinorVersion, request)
			repositories = append(repositories, repoData)
		}
	}
	rb, err = json.Marshal(repositories)
	if err != nil {
		panic(err)
	}
	repositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		panic(err)
	}

	if len(helmRepositoriesYamlByte) > 0 {
		data["helm.repositories"] = string(helmRepositoriesYamlByte)
	}
	if len(repositoriesYamlByte) > 0 {
		data["repositories"] = string(repositoriesYamlByte)
	}
	//dex config copy as it is
	dexConfigStr := data["dex.config"]
	data["dex.config"] = string([]byte(dexConfigStr))
	return data
}

func (impl *AppStoreServiceImpl) createRepoElement(apiMinorVersion int, request *ChartRepoDto) *AcdConfigMapRepositoriesDto {
	repoData := &AcdConfigMapRepositoriesDto{}
	if request.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
		usernameSecret := &KeyDto{Name: request.UserName, Key: "username"}
		passwordSecret := &KeyDto{Name: request.Password, Key: "password"}
		repoData.PasswordSecret = passwordSecret
		repoData.UsernameSecret = usernameSecret
	} else if request.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
		// TODO - is it access token or ca cert nd secret
	} else if request.AuthMode == repository.AUTH_MODE_SSH {
		keySecret := &KeyDto{Name: request.SshKey, Key: "key"}
		repoData.KeySecret = keySecret
	}
	repoData.Url = request.Url
	repoData.Name = request.Name
	if apiMinorVersion >= 3 {
		repoData.Type = "helm"
	}
	return repoData
}
func (impl *AppStoreServiceImpl) ValidateChartRepo(request *ChartRepoDto) *DetailedErrorHelmRepoValidation {
	var detailedErrorHelmRepoValidation DetailedErrorHelmRepoValidation
	helmRepoConfig := &repo.Entry{
		Name:     request.Name,
		URL:      request.Url,
		Username: request.UserName,
		Password: request.Password,
	}
	helmRepo, err, customMsg := impl.NewChartRepository(helmRepoConfig, getter.All(environment.EnvSettings{}))
	if err != nil {
		impl.logger.Errorw("failed to create chart repo for validating", "url", request.Url, "err", err)
		detailedErrorHelmRepoValidation.ActualErrMsg = err.Error()
		detailedErrorHelmRepoValidation.CustomErrMsg = customMsg
		return &detailedErrorHelmRepoValidation
	}
	parsedURL, _ := url.Parse(helmRepo.Config.URL)
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/index.yaml"
	indexURL := parsedURL.String()
	if t, ok := helmRepo.Client.(*getter.HttpGetter); ok {
		t.SetCredentials(helmRepo.Config.Username, helmRepo.Config.Password)
	}
	resp, err, statusCode := impl.get(indexURL, helmRepo)
	if statusCode == 401 || statusCode == 403 {
		impl.logger.Errorw("authentication or authorization error in request for getting index file", "statusCode", statusCode, "url", request.Url, "err", err)
		detailedErrorHelmRepoValidation.ActualErrMsg = err.Error()
		detailedErrorHelmRepoValidation.CustomErrMsg = fmt.Sprintf("Invalid authentication credentials. Please verify.")
		return &detailedErrorHelmRepoValidation
	} else if statusCode == 404 {
		impl.logger.Errorw("error in getting index file : not found", "url", request.Url, "err", err)
		detailedErrorHelmRepoValidation.ActualErrMsg = err.Error()
		detailedErrorHelmRepoValidation.CustomErrMsg = fmt.Sprintf("Could not find an index.yaml file in the repo directory. Please try another chart repo.")
		return &detailedErrorHelmRepoValidation
	} else if statusCode < 200 || statusCode > 299 {
		impl.logger.Errorw("error in getting index file", "url", request.Url, "err", err)
		detailedErrorHelmRepoValidation.ActualErrMsg = err.Error()
		detailedErrorHelmRepoValidation.CustomErrMsg = fmt.Sprintf("Could not validate the repo. Please try again.")
		return &detailedErrorHelmRepoValidation
	} else {
		_, err = ioutil.ReadAll(resp)
		if err != nil {
			impl.logger.Errorw("error in reading index file")
			detailedErrorHelmRepoValidation.ActualErrMsg = err.Error()
			detailedErrorHelmRepoValidation.CustomErrMsg = fmt.Sprintf("Devtron was unable to read the index.yaml file in the repo directory. Please try another chart repo.")
			return &detailedErrorHelmRepoValidation
		}
	}
	detailedErrorHelmRepoValidation.CustomErrMsg = ValidationSuccessMsg
	return &detailedErrorHelmRepoValidation
}
func (impl *AppStoreServiceImpl) ValidateAndCreateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error, *DetailedErrorHelmRepoValidation) {
	validationResult := impl.ValidateChartRepo(request)
	if validationResult.CustomErrMsg != ValidationSuccessMsg {
		return nil, nil, validationResult
	}
	chartRepo, err := impl.CreateChartRepo(request)
	if err != nil {
		return nil, err, validationResult
	}

	// Trigger chart sync job, ignore error
	err = impl.TriggerChartSyncManual()
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually", "err", err)
	}

	return chartRepo, err, validationResult
}
func (impl *AppStoreServiceImpl) ValidateAndUpdateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error, *DetailedErrorHelmRepoValidation) {
	validationResult := impl.ValidateChartRepo(request)
	if validationResult.CustomErrMsg != ValidationSuccessMsg {
		return nil, nil, validationResult
	}
	chartRepo, err := impl.UpdateChartRepo(request)
	if err != nil {
		return nil, err, validationResult
	}

	// Trigger chart sync job, ignore error
	err = impl.TriggerChartSyncManual()
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually", "err", err)
	}

	return chartRepo, err, validationResult
}

// NewChartRepository constructs ChartRepository
func (impl *AppStoreServiceImpl) NewChartRepository(cfg *repo.Entry, getters getter.Providers) (*repo.ChartRepository, error, string) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err, fmt.Sprintf("Invalid chart URL format: %s. Please provide a valid URL.", cfg.URL)
	}

	getterConstructor, err := getters.ByScheme(u.Scheme)
	if err != nil {
		return nil, err, fmt.Sprintf("Protocol \"%s\" is not supported. Supported protocols are http/https.", u.Scheme)
	}
	client, err := getterConstructor(cfg.URL, cfg.CertFile, cfg.KeyFile, cfg.CAFile)
	if err != nil {
		return nil, err, fmt.Sprintf("Unable to construct URL for the protocol \"%s\"", u.Scheme)
	}

	return &repo.ChartRepository{
		Config:    cfg,
		IndexFile: repo.NewIndexFile(),
		Client:    client,
	}, nil, fmt.Sprintf("")
}

func (impl *AppStoreServiceImpl) get(href string, chartRepository *repo.ChartRepository) (*bytes.Buffer, error, int) {
	buf := bytes.NewBuffer(nil)

	// Set a helm specific user agent so that a repo server and metrics can
	// separate helm calls from other tools interacting with repos.
	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return buf, err, http.StatusBadRequest
	}
	req.Header.Set("User-Agent", "Helm/"+strings.TrimPrefix(version.GetVersion(), "v"))

	if chartRepository.Config.Username != "" && chartRepository.Config.Password != "" {
		req.SetBasicAuth(chartRepository.Config.Username, chartRepository.Config.Password)
	}

	resp, err := impl.client.Do(req)
	if err != nil {
		return buf, err, http.StatusInternalServerError
	}
	if resp.StatusCode != 200 {
		return buf, fmt.Errorf("Failed to fetch %s : %s", href, resp.Status), resp.StatusCode
	}
	_, err = io.Copy(buf, resp.Body)
	resp.Body.Close()
	return buf, err, http.StatusOK
}

func (impl *AppStoreServiceImpl) TriggerChartSyncManual() error {
	defaultClusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		impl.logger.Errorw("defaultClusterBean err, TriggerChartSyncManual", "err", err)
		return err
	}

	defaultClusterConfig, err := impl.clusterService.GetClusterConfig(defaultClusterBean)
	if err != nil {
		impl.logger.Errorw("defaultClusterConfig err, TriggerChartSyncManual", "err", err)
		return err
	}

	manualAppSyncJobByteArr := manualAppSyncJobByteArr()

	err = impl.K8sUtil.DeleteAndCreateJob(manualAppSyncJobByteArr, argocdServer.DevtronInstalationNs, defaultClusterConfig)
	if err != nil {
		impl.logger.Errorw("DeleteAndCreateJob err, TriggerChartSyncManual", "err", err)
		return err
	}

	return nil
}
