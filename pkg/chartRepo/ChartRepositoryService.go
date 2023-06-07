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

package chartRepo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/version"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// secret keys
const (
	LABEL    string = "argocd.argoproj.io/secret-type"
	NAME     string = "name"
	USERNAME string = "username"
	PASSWORD string = "password"
	TYPE     string = "type"
	URL      string = "url"
	INSECRUE string = "insecure"
)

// secret values
const (
	HELM       string = "helm"
	REPOSITORY string = "repository"
)

type ChartRepositoryService interface {
	CreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error)
	UpdateData(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error)
	GetChartRepoById(id int) (*ChartRepoDto, error)
	GetChartRepoByName(name string) (*ChartRepoDto, error)
	GetChartRepoList() ([]*ChartRepoWithIsEditableDto, error)
	GetChartRepoListMin() ([]*ChartRepoDto, error)
	ValidateChartRepo(request *ChartRepoDto) *DetailedErrorHelmRepoValidation
	ValidateAndCreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation)
	ValidateAndUpdateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation)
	TriggerChartSyncManual() error
	DeleteChartRepo(request *ChartRepoDto) error
	DeleteChartSecret(secretName string) error
}

type ChartRepositoryServiceImpl struct {
	logger          *zap.SugaredLogger
	repoRepository  chartRepoRepository.ChartRepoRepository
	K8sUtil         *util.K8sUtil
	clusterService  cluster.ClusterService
	aCDAuthConfig   *util2.ACDAuthConfig
	client          *http.Client
	serverEnvConfig *serverEnvConfig.ServerEnvConfig
}

func NewChartRepositoryServiceImpl(logger *zap.SugaredLogger, repoRepository chartRepoRepository.ChartRepoRepository, K8sUtil *util.K8sUtil, clusterService cluster.ClusterService,
	aCDAuthConfig *util2.ACDAuthConfig, client *http.Client, serverEnvConfig *serverEnvConfig.ServerEnvConfig) *ChartRepositoryServiceImpl {
	return &ChartRepositoryServiceImpl{
		logger:          logger,
		repoRepository:  repoRepository,
		K8sUtil:         K8sUtil,
		clusterService:  clusterService,
		aCDAuthConfig:   aCDAuthConfig,
		client:          client,
		serverEnvConfig: serverEnvConfig,
	}
}

// Private helm charts credentials are saved as secrets
func (impl *ChartRepositoryServiceImpl) CreateSecretDataForHelmChart(request *ChartRepoDto, isPrivateChart bool) (secretData map[string]string) {
	secretData = make(map[string]string)
	secretData[NAME] = fmt.Sprintf("%s-%s", request.Name, uuid.New().String()) // making repo name unique so that "helm repo add" command in argo-repo-server doesn't give error
	secretData[TYPE] = HELM
	secretData[URL] = request.Url
	if isPrivateChart {
		secretData[USERNAME] = request.UserName
		secretData[PASSWORD] = request.Password
		isInsecureConnection := "true"
		if !request.AllowInsecureConnection {
			isInsecureConnection = "false"
		}
		secretData[INSECRUE] = isInsecureConnection
	}

	return secretData
}

func (impl *ChartRepositoryServiceImpl) CreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error) {
	//metadata.name label in Secret doesn't support uppercase hence returning if user enters uppercase letters
	if strings.ToLower(request.Name) != request.Name {
		return nil, errors.New("invalid repo name: please use lowercase")
	}
	allChartsRepos, err := impl.repoRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in getting list of all chart repos", "err", err)
		return nil, err
	}
	if len(allChartsRepos) == 0 {
		return nil, nil
	}
	for _, chart := range allChartsRepos {
		if chart.Name == request.Name {
			return nil, errors.New("repo with chart name already exists")
		}
	}
	dbConnection := impl.repoRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	chartRepo := &chartRepoRepository.ChartRepo{AuditLog: sql.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId}}
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
	chartRepo.AllowInsecureConnection = request.AllowInsecureConnection
	err = impl.repoRepository.Save(chartRepo, tx)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		return nil, err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return nil, err
	}

	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return nil, err
	}

	updateSuccess := false
	retryCount := 0

	isPrivateChart := false
	if len(chartRepo.UserName) > 0 && len(chartRepo.Password) > 0 {
		isPrivateChart = true
	}

	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		secretLabel := make(map[string]string)
		secretLabel[LABEL] = REPOSITORY
		secretData := impl.CreateSecretDataForHelmChart(request, isPrivateChart)
		_, err = impl.K8sUtil.CreateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, nil, chartRepo.Name, "", client, secretLabel, secretData)
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

func (impl *ChartRepositoryServiceImpl) UpdateData(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error) {
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
	previousName := chartRepo.Name
	previousUrl := chartRepo.Url
	//metadata.name label in Secret doesn't support uppercase hence returning if user enters uppercase letters in repo name
	if request.Name != previousName && strings.ToLower(request.Name) != request.Name {
		return nil, errors.New("invalid repo name: please use lowercase")
	}
	chartRepo.Url = request.Url
	chartRepo.Name = request.Name
	chartRepo.AuthMode = request.AuthMode
	chartRepo.UserName = request.UserName
	chartRepo.Password = request.Password
	chartRepo.Name = request.Name
	chartRepo.AccessToken = request.AccessToken
	chartRepo.SshKey = request.SshKey
	chartRepo.Active = request.Active
	chartRepo.UpdatedBy = request.UserId
	chartRepo.UpdatedOn = time.Now()
	chartRepo.AllowInsecureConnection = request.AllowInsecureConnection
	err = impl.repoRepository.Update(chartRepo, tx)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	// modify configmap
	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		return nil, err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return nil, err
	}
	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return nil, err
	}

	isPrivateChart := false
	if len(chartRepo.UserName) > 0 && len(chartRepo.Password) > 0 {
		isPrivateChart = true
	}

	updateSuccess := false
	retryCount := 0
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		var isFoundInArgoCdCm bool
		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return nil, err
		}
		var repositories []*AcdConfigMapRepositoriesDto
		if cm != nil && cm.Data != nil {
			repoStr := cm.Data["repositories"]
			repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
			if err != nil {
				impl.logger.Errorw("error in json patch", "err", err)
				return nil, err
			}
			err = json.Unmarshal(repoByte, &repositories)
			if err != nil {
				impl.logger.Errorw("error in unmarshal", "err", err)
				return nil, err
			}
			for _, repo := range repositories {
				if repo.Name == previousName && repo.Url == previousUrl {
					//chart repo is present in argocd-cm
					isFoundInArgoCdCm = true
					break
				}
			}
		}

		if isFoundInArgoCdCm {
			var data map[string]string
			// if the repo name has been updated then, create a new repo
			if cm != nil && cm.Data != nil {
				data, err = impl.updateRepoData(cm.Data, request)
				// if the repo name has been updated then, delete the previous repo
				if err != nil {
					impl.logger.Warnw(" config map update failed", "err", err)
					continue
				}
				if previousName != request.Name {
					data, err = impl.removeRepoData(cm.Data, previousName)
				}
				if err != nil {
					impl.logger.Warnw(" config map update failed", "err", err)
					continue
				}
			}
			cm.Data = data
			_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		} else {
			secretData := impl.CreateSecretDataForHelmChart(request, isPrivateChart)
			secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, previousName, client)
			if err != nil {
				impl.logger.Errorw("error in fetching secret", "err", err)
				continue
			}
			secret.StringData = secretData
			if previousName != request.Name {
				err = impl.DeleteChartSecret(previousName)
				if err != nil {
					impl.logger.Errorw("Error in deleting secret for chart repo", "Chart Name", chartRepo.Name, "err", err)
					continue
				}
				secretLabel := make(map[string]string)
				secretLabel[LABEL] = REPOSITORY
				_, err = impl.K8sUtil.CreateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, nil, chartRepo.Name, "", client, secretLabel, secretData)
				if err != nil {
					impl.logger.Errorw("Error in creating secret for chart repo", "Chart Name", chartRepo.Name, "err", err)
				}
			} else {
				_, err = impl.K8sUtil.UpdateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, secret, client)
				if err != nil {
					impl.logger.Errorw("Error in creating secret for chart repo", "Chart Name", chartRepo.Name, "err", err)
				}
			}
			if err != nil {
				impl.logger.Warnw("secret update for chart repo failed", "err", err)
				continue
			}
		}
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

// DeleteChartRepo update the active state from DB and modify the argo-cm with repo URL to null in case of public chart and delete secret in case of private chart
func (impl *ChartRepositoryServiceImpl) DeleteChartRepo(request *ChartRepoDto) error {
	dbConnection := impl.repoRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in establishing db connection, DeleteChartRepo", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	chartRepo, err := impl.repoRepository.FindById(request.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in finding chart repo by id", "err", err, "id", request.Id)
		return err
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
	err = impl.repoRepository.MarkChartRepoDeleted(chartRepo, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting chart repo", "err", err)
		return err
	}

	// modify configmap
	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		return err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return err
	}
	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return err
	}
	updateSuccess := false
	retryCount := 0
	//request.Url = ""

	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		var isFoundInArgoCdCm bool
		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return err
		}
		var repositories []*AcdConfigMapRepositoriesDto
		if cm != nil && cm.Data != nil {
			repoStr := cm.Data["repositories"]
			repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
			if err != nil {
				impl.logger.Errorw("error in json patch", "err", err)
				return err
			}
			err = json.Unmarshal(repoByte, &repositories)
			if err != nil {
				impl.logger.Errorw("error in unmarshal", "err", err)
				return err
			}
			for _, repo := range repositories {
				if repo.Name == request.Name && repo.Url == request.Url {
					//chart repo is present in argocd-cm
					isFoundInArgoCdCm = true
					break
				}
			}
		}

		if isFoundInArgoCdCm {
			var data map[string]string

			if cm != nil && cm.Data != nil {
				data, err = impl.removeRepoData(cm.Data, request.Name)
				if err != nil {
					impl.logger.Warnw(" config map update failed", "err", err)
					continue
				}
			}
			cm.Data = data
			_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		} else {
			err = impl.DeleteChartSecret(chartRepo.Name)
			if err != nil {
				impl.logger.Errorw("Error in deleting secret for chart repo", "Chart Name", chartRepo.Name, "err", err)
			}
		}
		if err != nil {
			impl.logger.Warnw(" error in deleting config/secret failed", "err", err)
			continue
		}
		if err == nil {
			impl.logger.Warnw(" config map apply succeeded", "on retryCount", retryCount)
			updateSuccess = true
		}
	}
	if !updateSuccess {
		return fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in tx commit, DeleteChartRepo", "err", err)
		return err
	}
	return nil
}

func (impl *ChartRepositoryServiceImpl) GetChartRepoById(id int) (*ChartRepoDto, error) {
	model, err := impl.repoRepository.FindById(id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	chartRepo := impl.convertFromDbResponse(model)
	return chartRepo, nil
}

func (impl *ChartRepositoryServiceImpl) GetChartRepoByName(name string) (*ChartRepoDto, error) {
	model, err := impl.repoRepository.FindByName(name)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	chartRepo := impl.convertFromDbResponse(model)
	return chartRepo, nil
}

func (impl *ChartRepositoryServiceImpl) convertFromDbResponse(model *chartRepoRepository.ChartRepo) *ChartRepoDto {
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
	chartRepo.AllowInsecureConnection = model.AllowInsecureConnection
	return chartRepo
}

func (impl *ChartRepositoryServiceImpl) GetChartRepoList() ([]*ChartRepoWithIsEditableDto, error) {
	var chartRepos []*ChartRepoWithIsEditableDto
	models, err := impl.repoRepository.FindAllWithDeploymentCount()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, model := range models {
		chartRepo := &ChartRepoWithIsEditableDto{}
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
		chartRepo.IsEditable = true
		if model.ActiveDeploymentCount > 0 {
			chartRepo.IsEditable = false
		}
		chartRepo.AllowInsecureConnection = model.AllowInsecureConnection
		chartRepos = append(chartRepos, chartRepo)
	}
	return chartRepos, nil
}

func (impl *ChartRepositoryServiceImpl) GetChartRepoListMin() ([]*ChartRepoDto, error) {
	var chartRepos []*ChartRepoDto
	models, err := impl.repoRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, model := range models {
		chartRepo := &ChartRepoDto{
			Id:                      model.Id,
			Name:                    model.Name,
			Url:                     model.Url,
			AuthMode:                model.AuthMode,
			Password:                model.Password,
			UserName:                model.UserName,
			SshKey:                  model.SshKey,
			AccessToken:             model.AccessToken,
			Default:                 model.Default,
			Active:                  model.Active,
			AllowInsecureConnection: model.AllowInsecureConnection,
		}
		chartRepos = append(chartRepos, chartRepo)
	}
	return chartRepos, nil
}

func (impl *ChartRepositoryServiceImpl) ValidateChartRepo(request *ChartRepoDto) *DetailedErrorHelmRepoValidation {
	var detailedErrorHelmRepoValidation DetailedErrorHelmRepoValidation
	helmRepoConfig := &repo.Entry{
		Name:     request.Name,
		URL:      request.Url,
		Username: request.UserName,
		Password: request.Password,
	}
	helmRepo, err, customMsg := impl.newChartRepository(helmRepoConfig, getter.All(environment.EnvSettings{}))
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

func (impl *ChartRepositoryServiceImpl) ValidateAndCreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation) {
	validationResult := impl.ValidateChartRepo(request)
	if validationResult.CustomErrMsg != ValidationSuccessMsg {
		return nil, nil, validationResult
	}
	chartRepo, err := impl.CreateChartRepo(request)
	if err != nil {
		return nil, err, validationResult
	}
	if chartRepo == nil && err == nil {
		//case when no entry for chart in chart_repo table and no need to trigger chart sync
		return nil, nil, validationResult
	}

	// Trigger chart sync job, ignore error
	err = impl.TriggerChartSyncManual()
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually ", "err", err)
	}

	return chartRepo, err, validationResult
}

func (impl *ChartRepositoryServiceImpl) ValidateAndUpdateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation) {
	validationResult := impl.ValidateChartRepo(request)
	if validationResult.CustomErrMsg != ValidationSuccessMsg {
		return nil, nil, validationResult
	}
	chartRepo, err := impl.UpdateData(request)
	if err != nil {
		return nil, err, validationResult
	}

	// Trigger chart sync job, ignore error
	err = impl.TriggerChartSyncManual()
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually", "err", err)
	}

	return chartRepo, nil, validationResult
}

func (impl *ChartRepositoryServiceImpl) TriggerChartSyncManual() error {
	defaultClusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		impl.logger.Errorw("defaultClusterBean err, TriggerChartSyncManual", "err", err)
		return err
	}

	defaultClusterConfig, err := impl.clusterService.GetClusterConfig(defaultClusterBean)
	if err != nil {
		impl.logger.Errorw("defaultClusterConfig err, TriggerChartSyncManual", "err", err)
		return err
	}

	manualAppSyncJobByteArr := manualAppSyncJobByteArr(impl.serverEnvConfig.AppSyncImage, impl.serverEnvConfig.AppSyncJobResourcesObj)

	err = impl.K8sUtil.DeleteAndCreateJob(manualAppSyncJobByteArr, impl.aCDAuthConfig.ACDConfigMapNamespace, defaultClusterConfig)
	if err != nil {
		impl.logger.Errorw("DeleteAndCreateJob err, TriggerChartSyncManual", "err", err)
		return err
	}

	return nil
}

func (impl *ChartRepositoryServiceImpl) get(href string, chartRepository *repo.ChartRepository) (*bytes.Buffer, error, int) {
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

// NewChartRepository constructs ChartRepository
func (impl *ChartRepositoryServiceImpl) newChartRepository(cfg *repo.Entry, getters getter.Providers) (*repo.ChartRepository, error, string) {
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

func (impl *ChartRepositoryServiceImpl) createRepoElement(request *ChartRepoDto) *AcdConfigMapRepositoriesDto {
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
	repoData.Type = "helm"

	return repoData
}

// updateRepoData update the request field in the argo-cm
func (impl *ChartRepositoryServiceImpl) updateRepoData(data map[string]string, request *ChartRepoDto) (map[string]string, error) {
	helmRepoStr := data["helm.repositories"]
	helmRepoByte, err := yaml.YAMLToJSON([]byte(helmRepoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	var helmRepositories []*AcdConfigMapRepositoriesDto
	err = json.Unmarshal(helmRepoByte, &helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	rb, err := json.Marshal(helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	helmRepositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
	}

	//SETUP for repositories
	var repositories []*AcdConfigMapRepositoriesDto
	repoStr := data["repositories"]
	repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	err = json.Unmarshal(repoByte, &repositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	found := false
	for _, item := range repositories {
		//if request chart repo found, then update its values
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
		repoData := impl.createRepoElement(request)
		repositories = append(repositories, repoData)
	}

	rb, err = json.Marshal(repositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	repositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
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
	return data, nil
}

// removeRepoData delete the request field from the argo-cm
func (impl *ChartRepositoryServiceImpl) removeRepoData(data map[string]string, name string) (map[string]string, error) {
	helmRepoStr := data["helm.repositories"]
	helmRepoByte, err := yaml.YAMLToJSON([]byte(helmRepoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	var helmRepositories []*AcdConfigMapRepositoriesDto
	err = json.Unmarshal(helmRepoByte, &helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	rb, err := json.Marshal(helmRepositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	helmRepositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
	}

	//SETUP for repositories
	var repositories []*AcdConfigMapRepositoriesDto
	repoStr := data["repositories"]
	repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
	if err != nil {
		impl.logger.Errorw("error in json patch", "err", err)
		return nil, err
	}
	err = json.Unmarshal(repoByte, &repositories)
	if err != nil {
		impl.logger.Errorw("error in unmarshal", "err", err)
		return nil, err
	}

	found := false
	for index, item := range repositories {
		//if request chart repo found, then delete its values
		if item.Name == name {
			repositories = append(repositories[:index], repositories[index+1:]...)
			found = true
			break
		}
	}

	// if request chart repo not found, add new one
	if !found {
		impl.logger.Errorw("Repo not found", "err", err)
		return nil, fmt.Errorf("Repo not found in config-map")
	}

	rb, err = json.Marshal(repositories)
	if err != nil {
		impl.logger.Errorw("error in marshal", "err", err)
		return nil, err
	}
	repositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		impl.logger.Errorw("error in yaml patch", "err", err)
		return nil, err
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
	return data, nil
}

func (impl *ChartRepositoryServiceImpl) DeleteChartSecret(secretName string) error {
	clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
	if err != nil {
		return err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return err
	}
	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return err
	}
	err = impl.K8sUtil.DeleteSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, secretName, client)
	return err
}
