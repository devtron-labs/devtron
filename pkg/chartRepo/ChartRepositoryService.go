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
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/sql"
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
	"strings"
	"time"
)

type ChartRepositoryService interface {
	CreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error)
	UpdateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error)
	GetChartRepoById(id int) (*ChartRepoDto, error)
	GetChartRepoByName(name string) (*ChartRepoDto, error)
	GetChartRepoList() ([]*ChartRepoDto, error)
	ValidateChartRepo(request *ChartRepoDto) *DetailedErrorHelmRepoValidation
	ValidateAndCreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation)
	ValidateAndUpdateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation)
	TriggerChartSyncManual() error
	DeleteChartRepo(request *ChartRepoDto) error
}

type ChartRepositoryServiceImpl struct {
	logger         *zap.SugaredLogger
	repoRepository chartRepoRepository.ChartRepoRepository
	K8sUtil        *util.K8sUtil
	clusterService cluster.ClusterService
	aCDAuthConfig  *util2.ACDAuthConfig
	client         *http.Client
}

func NewChartRepositoryServiceImpl(logger *zap.SugaredLogger, repoRepository chartRepoRepository.ChartRepoRepository, K8sUtil *util.K8sUtil, clusterService cluster.ClusterService,
	aCDAuthConfig *util2.ACDAuthConfig, client *http.Client) *ChartRepositoryServiceImpl {
	return &ChartRepositoryServiceImpl{
		logger:         logger,
		repoRepository: repoRepository,
		K8sUtil:        K8sUtil,
		clusterService: clusterService,
		aCDAuthConfig:  aCDAuthConfig,
		client:         client,
	}
}

func (impl *ChartRepositoryServiceImpl) CreateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error) {
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
	err = impl.repoRepository.Save(chartRepo, tx)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	clusterBean, err := impl.clusterService.FindOne(cluster.DefaultClusterName)
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
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return nil, err
		}
		data := impl.updateData(cm.Data, request)
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

func (impl *ChartRepositoryServiceImpl) UpdateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error) {
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
	clusterBean, err := impl.clusterService.FindOne(cluster.DefaultClusterName)
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
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return nil, err
		}
		data := impl.updateData(cm.Data, request)
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
	return chartRepo
}

func (impl *ChartRepositoryServiceImpl) GetChartRepoList() ([]*ChartRepoDto, error) {
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

	// Trigger chart sync job, ignore error
	err = impl.TriggerChartSyncManual()
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually", "err", err)
	}

	return chartRepo, err, validationResult
}

func (impl *ChartRepositoryServiceImpl) ValidateAndUpdateChartRepo(request *ChartRepoDto) (*chartRepoRepository.ChartRepo, error, *DetailedErrorHelmRepoValidation) {
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

func (impl *ChartRepositoryServiceImpl) TriggerChartSyncManual() error {
	defaultClusterBean, err := impl.clusterService.FindOne(cluster.DefaultClusterName)
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

func (impl *ChartRepositoryServiceImpl) updateData(data map[string]string, request *ChartRepoDto) map[string]string {
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
		repoData := impl.createRepoElement(request)
		repositories = append(repositories, repoData)
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
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in tx commit, DeleteChartRepo", "err", err)
		return err
	}
	return nil
}
