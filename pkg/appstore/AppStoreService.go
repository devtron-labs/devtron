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
	"encoding/json"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/ghodss/yaml"
	"go.uber.org/zap"
	"strconv"
	"time"
)

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
	HelmRepositories *AcdConfigMapRepositoriesDto `json:"helm.repositories,omitempty"`
	Repositories     *AcdConfigMapRepositoriesDto `json:"repositories,omitempty"`
}

type AppStoreService interface {
	FindAllApps() ([]appstore.AppStoreWithVersion, error)
	FindChartDetailsById(id int) (AppStoreApplicationVersionResponse, error)
	FindChartVersionsByAppStoreId(appStoreId int) ([]AppStoreVersionsResponse, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean.AppDetailContainer, error)
	GetReadMeByAppStoreApplicationVersionId(id int) (*ReadmeRes, error)

	SearchAppStoreChartByName(chartName string) ([]*appstore.ChartRepoSearch, error)
	CreateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error)
	UpdateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error)
}

const ChartRepoConfigMap string = "argocd-test-cm"
const ChartRepoConfigMapNamespace string = "devtroncd"

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
}

func NewAppStoreServiceImpl(logger *zap.SugaredLogger, appStoreRepository appstore.AppStoreRepository,
	appStoreApplicationRepository appstore.AppStoreApplicationVersionRepository, installedAppRepository appstore.InstalledAppRepository,
	userService user.UserService, repoRepository chartConfig.ChartRepoRepository, K8sUtil *util.K8sUtil,
	clusterService cluster.ClusterService, envService cluster.EnvironmentService,
	versionService argocdServer.VersionService) *AppStoreServiceImpl {
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
	}
}

func (impl *AppStoreServiceImpl) FindAllApps() ([]appstore.AppStoreWithVersion, error) {
	appStoreApplications, err := impl.appStoreApplicationRepository.FindAll()
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
	userInfo, err := impl.userService.GetById(installedAppVerison.AuditLog.UpdatedBy)
	if err != nil {
		impl.logger.Error(err)
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

	chartRepo := &chartConfig.ChartRepo{AuditLog: models.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},}
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
	cfg, err := impl.envService.GetClusterConfig(clusterBean)
	if err != nil {
		return nil, err
	}
	cm, err := impl.K8sUtil.GetConfigMap(ChartRepoConfigMapNamespace, ChartRepoConfigMap, cfg)
	if err != nil {
		return nil, err
	}
	data := impl.updateData(cm.Data, request)
	_, err = impl.K8sUtil.PatchConfigMap(ChartRepoConfigMapNamespace, cfg, ChartRepoConfigMap, data)
	if err != nil {
		return nil, err
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
	cfg, err := impl.envService.GetClusterConfig(clusterBean)
	if err != nil {
		return nil, err
	}
	cm, err := impl.K8sUtil.GetConfigMap(ChartRepoConfigMapNamespace, ChartRepoConfigMap, cfg)
	if err != nil {
		return nil, err
	}
	data := impl.updateData(cm.Data, request)
	_, err = impl.K8sUtil.PatchConfigMap(ChartRepoConfigMapNamespace, cfg, ChartRepoConfigMap, data)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return chartRepo, nil
}

func (impl *AppStoreServiceImpl) updateData(data map[string]string, request *ChartRepoDto) map[string]interface{} {
	apiVersion, err := impl.versionService.GetVersion()
	apiMinorVersion, err := strconv.Atoi(apiVersion[3:4])
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil
	}

	helmRepoStr := data["helm.repositories"]
	helmRepoByte, err := yaml.YAMLToJSON([]byte(helmRepoStr))
	if err != nil {
		panic(err)
	}
	//helmRepositories := make([]map[string]interface{}, 0)
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

	mergedData := map[string]interface{}{}
	if apiMinorVersion >= 3 {
		mergedData["repositories"] = string(repositoriesYamlByte)
		return mergedData
	} else {
		mergedData["helm.repositories"] = string(helmRepositoriesYamlByte)
		return mergedData
	}
	//mergedData["helm.repositories"] = string(helmRepositoriesYamlByte)
	//mergedData["repositories"] = string(repositoriesYamlByte)
	newDataFinal := map[string]interface{}{}
	newDataFinal["data"] = mergedData
	return newDataFinal
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
