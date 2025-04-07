/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package chartRepo

import (
	"bytes"
	"errors"
	"fmt"
	util3 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/read"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/version"
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
	TriggerChartSyncManual(chartProviderConfig *ChartProviderConfig) error
	DeleteChartRepo(request *ChartRepoDto) error
}

type ChartRepositoryServiceImpl struct {
	logger                   *zap.SugaredLogger
	repoRepository           chartRepoRepository.ChartRepoRepository
	K8sUtil                  *util3.K8sServiceImpl
	aCDAuthConfig            *util2.ACDAuthConfig
	client                   *http.Client
	serverEnvConfig          *serverEnvConfig.ServerEnvConfig
	argoClientWrapperService argocdServer.ArgoClientWrapperService
	clusterReadService       read.ClusterReadService
}

func NewChartRepositoryServiceImpl(logger *zap.SugaredLogger, repoRepository chartRepoRepository.ChartRepoRepository, K8sUtil *util3.K8sServiceImpl,
	aCDAuthConfig *util2.ACDAuthConfig, client *http.Client, serverEnvConfig *serverEnvConfig.ServerEnvConfig,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	clusterReadService read.ClusterReadService) *ChartRepositoryServiceImpl {
	return &ChartRepositoryServiceImpl{
		logger:                   logger,
		repoRepository:           repoRepository,
		K8sUtil:                  K8sUtil,
		aCDAuthConfig:            aCDAuthConfig,
		client:                   client,
		serverEnvConfig:          serverEnvConfig,
		argoClientWrapperService: argoClientWrapperService,
		clusterReadService:       clusterReadService,
	}
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
		impl.logger.Errorw("error in saving chart repo in DB", "err", err)
		return nil, err
	}

	isPrivateChart := false
	if len(chartRepo.UserName) > 0 && len(chartRepo.Password) > 0 {
		isPrivateChart = true
	}

	argoServerAddRequest := bean.ChartRepositoryAddRequest{
		Name:                    chartRepo.Name,
		Username:                chartRepo.UserName,
		Password:                chartRepo.Password,
		URL:                     chartRepo.Url,
		AllowInsecureConnection: chartRepo.AllowInsecureConnection,
		IsPrivateChart:          isPrivateChart,
	}

	err = impl.argoClientWrapperService.AddChartRepository(argoServerAddRequest)
	if err != nil {
		impl.logger.Errorw("error in adding chart repository to argocd server", "name", request.Name, "err", err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return chartRepo, nil
}

func (impl *ChartRepositoryServiceImpl) getCountOfDeployedCharts(chartRepoId int) (int, error) {
	activeDeploymentCount, err := impl.repoRepository.FindDeploymentCountByChartRepoId(chartRepoId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment count, CheckDeploymentCount", "chartRepoId", chartRepoId, "err", err)
		return 0, err
	}
	return activeDeploymentCount, nil
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

	deployedChartCount, err := impl.getCountOfDeployedCharts(request.Id)
	if err != nil {
		impl.logger.Errorw("error in getting charts deployed via chart repo", "chartRepoId", request.Id, "err", err)
		return nil, err
	}

	if deployedChartCount > 0 && (request.Name != previousName || request.Url != previousUrl) {
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "cannot update, found charts deployed using this repo"}
		return nil, err
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

	isPrivateChart := false
	if len(chartRepo.UserName) > 0 && len(chartRepo.Password) > 0 {
		isPrivateChart = true
	}

	argoServerUpdateRequest := bean.ChartRepositoryUpdateRequest{
		PreviousName:            previousName,
		PreviousURL:             previousUrl,
		Name:                    chartRepo.Name,
		AuthMode:                string(chartRepo.AuthMode),
		Username:                chartRepo.UserName,
		Password:                chartRepo.Password,
		SSHKey:                  chartRepo.SshKey,
		URL:                     chartRepo.Url,
		AllowInsecureConnection: chartRepo.AllowInsecureConnection,
		IsPrivateChart:          isPrivateChart,
	}
	// modify configmap
	err = impl.argoClientWrapperService.UpdateChartRepository(argoServerUpdateRequest)
	if err != nil {
		return nil, err
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
	err = impl.argoClientWrapperService.DeleteChartRepository(request.Name, request.Url)
	if err != nil {
		impl.logger.Errorw("error in deleting chart repository from argocd", "name", request.Name, "err", err)
		return err
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
	if len(request.Name) < 3 || strings.Contains(request.Name, " ") {
		impl.logger.Errorw("name should not contain white spaces and should contain min 3 chars")
		detailedErrorHelmRepoValidation.CustomErrMsg = fmt.Sprintf("name should not contain white spaces and should have min 3 chars")
		detailedErrorHelmRepoValidation.ActualErrMsg = fmt.Sprintf("name should not contain white spaces and should have min 3 chars")
		return &detailedErrorHelmRepoValidation
	}
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
	chartProviderConfig := &ChartProviderConfig{
		ChartProviderId: strconv.Itoa(chartRepo.Id),
		IsOCIRegistry:   false,
	}
	err = impl.TriggerChartSyncManual(chartProviderConfig)
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
	chartProviderConfig := &ChartProviderConfig{
		ChartProviderId: strconv.Itoa(chartRepo.Id),
		IsOCIRegistry:   false,
	}
	err = impl.TriggerChartSyncManual(chartProviderConfig)
	if err != nil {
		impl.logger.Errorw("Error in triggering chart sync job manually", "err", err)
	}

	return chartRepo, nil, validationResult
}

func (impl *ChartRepositoryServiceImpl) TriggerChartSyncManual(chartProviderConfig *ChartProviderConfig) error {
	defaultClusterBean, err := impl.clusterReadService.FindOne(bean2.DEFAULT_CLUSTER)
	if err != nil {
		impl.logger.Errorw("defaultClusterBean err, TriggerChartSyncManual", "err", err)
		return err
	}

	defaultClusterConfig := defaultClusterBean.GetClusterConfig()

	manualAppSyncJobByteArr := manualAppSyncJobByteArr(impl.serverEnvConfig.AppSyncImage, impl.serverEnvConfig.AppSyncJobResourcesObj, impl.serverEnvConfig.AppSyncServiceAccount, chartProviderConfig, impl.serverEnvConfig.ParallelismLimitForTagProcessing, impl.serverEnvConfig.AppSyncJobShutDownWaitDuration)
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
