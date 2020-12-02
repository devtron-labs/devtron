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
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
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

type RepositoriesData struct {
	UsernameSecret *KeyDto `json:"usernameSecret,omitempty"`
	PasswordSecret *KeyDto `json:"passwordSecret,omitempty"`
	CaSecret       *KeyDto `json:"caSecret,omitempty"`
	CertSecret     *KeyDto `json:"certSecret,omitempty"`
	KeySecret      *KeyDto `json:"keySecret,omitempty"`
	Url            string  `json:"url,omitempty"`
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
const ChartRepoConfigMapNamespace string = "default"

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
}

func NewAppStoreServiceImpl(logger *zap.SugaredLogger, appStoreRepository appstore.AppStoreRepository,
	appStoreApplicationRepository appstore.AppStoreApplicationVersionRepository, installedAppRepository appstore.InstalledAppRepository,
	userService user.UserService, repoRepository chartConfig.ChartRepoRepository, K8sUtil *util.K8sUtil,
	clusterService cluster.ClusterService, envService cluster.EnvironmentService) *AppStoreServiceImpl {
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
	/*err := impl.repoRepository.Save(chartRepo)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}*/

	//TODO - config map update - patch
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
	} else {
		//return chartRepo, nil
	}
	impl.logger.Info(cm.Data)
	data := impl.updateData(cm.Data, request)
	impl.logger.Info(data)
	_, err = impl.K8sUtil.PatchConfigMap(ChartRepoConfigMapNamespace, cfg, ChartRepoConfigMap, data)
	if err != nil {
		return nil, err
	}
	return chartRepo, nil
}

func (impl *AppStoreServiceImpl) updateData(data map[string]string, request *ChartRepoDto) map[string]interface{} {

	helmRepoStr := data["helm.repositories"]
	helmRepoByte, err := yaml.YAMLToJSON([]byte(helmRepoStr))
	if err != nil {
		panic(err)
	}
	//helmRepositories := make([]map[string]interface{}, 0)
	var helmRepositories []*HelmRepositoriesData
	err = json.Unmarshal(helmRepoByte, &helmRepositories)
	if err != nil {
		panic(err)
	}
	found := false
	for _, item := range helmRepositories {
		a := fmt.Sprintf("%s/%s", request.Name, item.Name)
		fmt.Println(a, found)
		if request.Name == item.Name {
			item.Url = request.Url
			found = true
		}
	}

	// add new only if not found
	if !found {
		helmRepository := &HelmRepositoriesData{Name: "devtron-charts-1001", Url: "https://devtron-charts.s3.us-east-2.amazonaws.com/charts-1001",
		}
		helmRepositories = append(helmRepositories, helmRepository)
	}

	rb, err := json.Marshal(helmRepositories)
	if err != nil {
		panic(err)
	}
	helmRepositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		panic(err)
	}

	var repositories []*RepositoriesData
	repoStr := data["repositories"]
	repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(repoByte, &repositories)
	if err != nil {
		panic(err)
	}
	found = false
	for _, item := range repositories {
		if request.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
			passwordSecret := item.PasswordSecret
			usernameSecret := item.UsernameSecret
			if passwordSecret.Name == request.Password {
				found = true
			}
			if usernameSecret.Name == request.UserName {
				found = true
			}
		} else if request.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {

		} else if request.AuthMode == repository.AUTH_MODE_SSH {

		}
	}

	if !found {
		usernameSecret := &KeyDto{Name: "my-secret-123000", Key: "username"}
		passwordSecret := &KeyDto{Name: "my-secret-1230000", Key: "password"}
		repository := &RepositoriesData{PasswordSecret: passwordSecret, UsernameSecret: usernameSecret,}
		repositories = append(repositories, repository)
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
	mergedData["helm.repositories"] = string(helmRepositoriesYamlByte)
	mergedData["repositories"] = string(repositoriesYamlByte)

	newDataFinal := map[string]interface{}{}
	newDataFinal["data"] = mergedData
	return newDataFinal
}

func (impl *AppStoreServiceImpl) UpdateChartRepo(request *ChartRepoDto) (*chartConfig.ChartRepo, error) {
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
	err = impl.repoRepository.Update(chartRepo)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	//TODO - config map update - patch
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
	} else {
		//return chartRepo, nil
	}
	impl.logger.Info(cm.Data)
	data := impl.updateData(cm.Data, request)
	impl.logger.Info(data)
	_, err = impl.K8sUtil.PatchConfigMap(ChartRepoConfigMapNamespace, cfg, ChartRepoConfigMap, data)
	if err != nil {
		return nil, err
	}

	if request.Active == false {
		// TODO - handle existing charts and charts group
	}

	return chartRepo, nil
}

func (impl *AppStoreServiceImpl) updateDataSample(data map[string]string, request *ChartRepoDto) map[string]interface{} {
	var helmRepositories []map[string]interface{}
	helmRepository := map[string]interface{}{}
	helmRepository["name"] = "devtron-charts-1001"
	helmRepository["url"] = "https://devtron-charts.s3.us-east-2.amazonaws.com/charts-1001"
	helmRepositories = append(helmRepositories, helmRepository)

	usernameSecret := map[string]interface{}{}
	usernameSecret["name"] = "my-secret-123"
	usernameSecret["key"] = "username"

	passwordSecret := map[string]interface{}{}
	passwordSecret["name"] = "my-secret-123"
	passwordSecret["key"] = "password"

	repository := map[string]interface{}{}
	repository["passwordSecret"] = passwordSecret
	repository["usernameSecret"] = usernameSecret

	var repositories []map[string]interface{}
	repositories = append(repositories, repository)

	rb, err := json.Marshal(repositories)
	if err != nil {
		panic(err)
	}
	yamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		panic(err)
	}

	newData := map[string]interface{}{}
	newData["repositories"] = string(yamlByte)
	newData["helm.repositories"] = helmRepositories

	newDataFinal := map[string]interface{}{}
	newDataFinal["data"] = newData
	return newDataFinal
}
