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

package gitops

import (
	"context"
	"encoding/json"
	"fmt"
	cluster3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	cluster2 "github.com/devtron-labs/devtron/client/argocdServer/cluster"
	"github.com/devtron-labs/devtron/pkg/sql"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
)

type GitOpsConfigService interface {
	ValidateAndCreateGitOpsConfig(config *bean2.GitOpsConfigDto) (DetailedErrorGitOpsConfigResponse, error)
	ValidateAndUpdateGitOpsConfig(config *bean2.GitOpsConfigDto) (DetailedErrorGitOpsConfigResponse, error)
	GitOpsValidateDryRun(config *bean2.GitOpsConfigDto) DetailedErrorGitOpsConfigResponse
	CreateGitOpsConfig(ctx context.Context, config *bean2.GitOpsConfigDto) (*bean2.GitOpsConfigDto, error)
	UpdateGitOpsConfig(config *bean2.GitOpsConfigDto) error
	GetGitOpsConfigById(id int) (*bean2.GitOpsConfigDto, error)
	GetAllGitOpsConfig() ([]*bean2.GitOpsConfigDto, error)
	GetGitOpsConfigByProvider(provider string) (*bean2.GitOpsConfigDto, error)
	GetGitOpsConfigActive() (*bean2.GitOpsConfigDto, error)
}

const (
	GitOpsSecretName      = "devtron-gitops-secret"
	DryrunRepoName        = "devtron-sample-repo-dryrun-"
	DeleteRepoStage       = "Delete Repo"
	CommitOnRestStage     = "Commit On Rest"
	PushStage             = "Push"
	CloneStage            = "Clone"
	GetRepoUrlStage       = "Get Repo Url"
	CreateRepoStage       = "Create Repo"
	CloneHttp             = "Clone Http"
	CreateReadmeStage     = "Create Readme"
	GITHUB_PROVIDER       = "GITHUB"
	GITLAB_PROVIDER       = "GITLAB"
	BITBUCKET_PROVIDER    = "BITBUCKET_CLOUD"
	AZURE_DEVOPS_PROVIDER = "AZURE_DEVOPS"
	BITBUCKET_API_HOST    = "https://api.bitbucket.org/2.0/"
)

type DetailedErrorGitOpsConfigResponse struct {
	SuccessfulStages []string          `json:"successfulStages"`
	StageErrorMap    map[string]string `json:"stageErrorMap"`
	ValidatedOn      time.Time         `json:"validatedOn"`
	DeleteRepoFailed bool              `json:"deleteRepoFailed"`
}

type GitOpsConfigServiceImpl struct {
	randSource           rand.Source
	logger               *zap.SugaredLogger
	gitOpsRepository     repository.GitOpsConfigRepository
	K8sUtil              *util.K8sUtil
	aCDAuthConfig        *util3.ACDAuthConfig
	clusterService       cluster.ClusterService
	envService           cluster.EnvironmentService
	versionService       argocdServer.VersionService
	gitFactory           *util.GitFactory
	chartTemplateService util.ChartTemplateService
	argoUserService      argo.ArgoUserService
	clusterServiceCD     cluster2.ServiceClient
}

func NewGitOpsConfigServiceImpl(Logger *zap.SugaredLogger, ciHandler pipeline.CiHandler,
	gitOpsRepository repository.GitOpsConfigRepository, K8sUtil *util.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig,
	clusterService cluster.ClusterService, envService cluster.EnvironmentService, versionService argocdServer.VersionService,
	gitFactory *util.GitFactory, chartTemplateService util.ChartTemplateService, argoUserService argo.ArgoUserService, clusterServiceCD cluster2.ServiceClient) *GitOpsConfigServiceImpl {
	return &GitOpsConfigServiceImpl{
		randSource:           rand.NewSource(time.Now().UnixNano()),
		logger:               Logger,
		gitOpsRepository:     gitOpsRepository,
		K8sUtil:              K8sUtil,
		aCDAuthConfig:        aCDAuthConfig,
		clusterService:       clusterService,
		envService:           envService,
		versionService:       versionService,
		gitFactory:           gitFactory,
		chartTemplateService: chartTemplateService,
		argoUserService:      argoUserService,
		clusterServiceCD:     clusterServiceCD,
	}
}

func (impl *GitOpsConfigServiceImpl) ValidateAndCreateGitOpsConfig(config *bean2.GitOpsConfigDto) (DetailedErrorGitOpsConfigResponse, error) {
	detailedErrorGitOpsConfigResponse := impl.GitOpsValidateDryRun(config)
	if len(detailedErrorGitOpsConfigResponse.StageErrorMap) == 0 {
		//create argo-cd user, if not created, here argo-cd integration has to be installed
		token := impl.argoUserService.GetOrUpdateArgoCdUserDetail()
		ctx := context.WithValue(context.Background(), "token", token)
		_, err := impl.CreateGitOpsConfig(ctx, config)
		if err != nil {
			impl.logger.Errorw("service err, SaveGitRepoConfig", "err", err, "payload", config)
			return detailedErrorGitOpsConfigResponse, err
		}
	}
	return detailedErrorGitOpsConfigResponse, nil
}
func (impl *GitOpsConfigServiceImpl) ValidateAndUpdateGitOpsConfig(config *bean2.GitOpsConfigDto) (DetailedErrorGitOpsConfigResponse, error) {
	detailedErrorGitOpsConfigResponse := impl.GitOpsValidateDryRun(config)
	if len(detailedErrorGitOpsConfigResponse.StageErrorMap) == 0 {
		err := impl.UpdateGitOpsConfig(config)
		if err != nil {
			impl.logger.Errorw("service err, UpdateGitOpsConfig", "err", err, "payload", config)
			return detailedErrorGitOpsConfigResponse, err
		}
	}
	return detailedErrorGitOpsConfigResponse, nil
}

func (impl *GitOpsConfigServiceImpl) buildGithubOrgUrl(host, orgId string) (orgUrl string, err error) {
	hostUrl, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	hostUrl.Path = path.Join(hostUrl.Path, orgId)
	return hostUrl.String(), nil
}

func (impl *GitOpsConfigServiceImpl) CreateGitOpsConfig(ctx context.Context, request *bean2.GitOpsConfigDto) (*bean2.GitOpsConfigDto, error) {
	impl.logger.Debugw("gitops create request", "req", request)
	dbConnection := impl.gitOpsRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	existingModel, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in creating new gitops config", "error", err)
		return nil, err
	}
	if existingModel != nil && existingModel.Id > 0 {
		existingModel.Active = false
		existingModel.UpdatedOn = time.Now()
		existingModel.UpdatedBy = request.UserId
		err = impl.gitOpsRepository.UpdateGitOpsConfig(existingModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new gitops config", "error", err)
			return nil, err
		}
	}
	model := &repository.GitOpsConfig{
		Provider:             strings.ToUpper(request.Provider),
		Username:             request.Username,
		Token:                request.Token,
		GitHubOrgId:          request.GitHubOrgId,
		GitLabGroupId:        request.GitLabGroupId,
		Host:                 request.Host,
		Active:               true,
		AzureProject:         request.AzureProjectName,
		BitBucketWorkspaceId: request.BitBucketWorkspaceId,
		BitBucketProjectKey:  request.BitBucketProjectKey,
		AuditLog:             sql.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}
	model, err = impl.gitOpsRepository.CreateGitOpsConfig(model, tx)
	if err != nil {
		impl.logger.Errorw("error in saving gitops config", "data", model, "err", err)
		err = &util.ApiError{
			InternalMessage: "gitops config failed to create in db",
			UserMessage:     "gitops config failed to create in db",
		}
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
	secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, GitOpsSecretName, client)
	statusError, _ := err.(*errors.StatusError)
	if err != nil && statusError.Status().Code != http.StatusNotFound {
		impl.logger.Errorw("secret not found", "err", err)
		return nil, err
	}
	data := make(map[string][]byte)
	data["username"] = []byte(request.Username)
	data["password"] = []byte(request.Token)
	if secret == nil {
		secret, err = impl.K8sUtil.CreateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, data, GitOpsSecretName, client)
		if err != nil {
			impl.logger.Errorw("err on creating secret", "err", err)
			return nil, err
		}
	} else {
		secret.Data = data
		secret, err = impl.K8sUtil.UpdateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, secret, client)
		if err != nil {
			operationComplete := false
			retryCount := 0
			for !operationComplete && retryCount < 3 {
				retryCount = retryCount + 1
				secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, GitOpsSecretName, client)
				if err != nil {
					impl.logger.Errorw("secret not found", "err", err)
					return nil, err
				}
				secret.Data = data
				secret, err = impl.K8sUtil.UpdateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, secret, client)
				if err != nil {
					continue
				}
				if err == nil {
					operationComplete = true
				}
			}

		}
	}
	if strings.ToUpper(request.Provider) == GITHUB_PROVIDER {
		orgUrl, err := impl.buildGithubOrgUrl(request.Host, request.GitHubOrgId)
		if err != nil {
			return nil, err
		}
		request.Host = orgUrl
	}
	if strings.ToUpper(request.Provider) == GITLAB_PROVIDER {
		groupName, err := impl.gitFactory.GetGitLabGroupPath(request)
		if err != nil {
			return nil, err
		}
		slashSuffixPresent := strings.HasSuffix(request.Host, "/")
		if slashSuffixPresent {
			request.Host += groupName
		} else {
			request.Host = fmt.Sprintf(request.Host+"/%s", groupName)
		}
	}
	if strings.ToUpper(request.Provider) == BITBUCKET_PROVIDER {
		request.Host = util.BITBUCKET_CLONE_BASE_URL + request.BitBucketWorkspaceId
	}
	operationComplete := false
	retryCount := 0
	for !operationComplete && retryCount < 3 {
		retryCount = retryCount + 1

		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return nil, err
		}
		updatedData := impl.updateData(cm.Data, request, GitOpsSecretName, existingModel.Host)
		data := cm.Data
		if data == nil {
			data = make(map[string]string, 0)
		}
		data["repository.credentials"] = updatedData["repository.credentials"]
		cm.Data = data
		_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			continue
		}
		if err == nil {
			operationComplete = true
		}
	}
	if !operationComplete {
		return nil, fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}

	// if git-ops config is created/saved successfully (just before transaction commit) and this was first git-ops config, then upsert clusters in acd
	isGitOpsConfigured, err := impl.gitOpsRepository.IsGitOpsConfigured()
	if err != nil {
		return nil, err
	}
	if !isGitOpsConfigured {
		clusters, err := impl.clusterService.FindAllActive()
		if err != nil {
			impl.logger.Errorw("Error while fetching all the clusters", "err", err)
			return nil, err
		}
		for _, cluster := range clusters {
			cl := &v1alpha1.Cluster{
				Name:   cluster.ClusterName,
				Server: cluster.ServerUrl,
				Config: v1alpha1.ClusterConfig{
					BearerToken: cluster.Config["bearer_token"],
					TLSClientConfig: v1alpha1.TLSClientConfig{
						Insecure: true,
					},
				},
			}
			_, err = impl.clusterServiceCD.Create(ctx, &cluster3.ClusterCreateRequest{Upsert: true, Cluster: cl})
			if err != nil {
				impl.logger.Errorw("Error while upserting cluster in acd", "clusterName", cluster.ClusterName, "err", err)
				return nil, err
			}
		}
	}

	// now commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	err = impl.gitFactory.Reload()
	if err != nil {
		return nil, err
	}
	request.Id = model.Id
	return request, nil
}

func (impl *GitOpsConfigServiceImpl) UpdateGitOpsConfig(request *bean2.GitOpsConfigDto) error {
	impl.logger.Debugw("gitops config update request", "req", request)
	dbConnection := impl.gitOpsRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	model, err := impl.gitOpsRepository.GetGitOpsConfigById(request.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for update.", "id", request.Id)
		err = &util.ApiError{
			InternalMessage: "gitops config update failed, does not exist",
			UserMessage:     "gitops config update failed, does not exist",
		}
		return err
	}

	existingModel, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in creating new gitops config", "error", err)
		return err
	}
	if request.Active {
		if existingModel != nil && existingModel.Id > 0 && existingModel.Id != model.Id {
			existingModel.Active = false
			existingModel.UpdatedOn = time.Now()
			existingModel.UpdatedBy = request.UserId
			err = impl.gitOpsRepository.UpdateGitOpsConfig(existingModel, tx)
			if err != nil {
				impl.logger.Errorw("error in creating new gitops config", "error", err)
				return err
			}
		}
	} else {
		if existingModel == nil || existingModel.Id == 0 {
			return fmt.Errorf("no active config found, please ensure atleast on gitops config active")
		}
	}

	model.Provider = strings.ToUpper(request.Provider)
	model.Username = request.Username
	model.Token = request.Token
	model.GitLabGroupId = request.GitLabGroupId
	model.GitHubOrgId = request.GitHubOrgId
	model.Host = request.Host
	model.Active = request.Active
	model.AzureProject = request.AzureProjectName
	model.BitBucketWorkspaceId = request.BitBucketWorkspaceId
	model.BitBucketProjectKey = request.BitBucketProjectKey
	err = impl.gitOpsRepository.UpdateGitOpsConfig(model, tx)
	if err != nil {
		impl.logger.Errorw("error in updating team", "data", model, "err", err)
		err = &util.ApiError{
			InternalMessage: "gitops config failed to update in db",
			UserMessage:     "gitops config failed to update in db",
		}
		return err
	}
	request.Id = model.Id

	clusterBean, err := impl.clusterService.FindOne(cluster.DefaultClusterName)
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

	secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, GitOpsSecretName, client)
	statusError, _ := err.(*errors.StatusError)
	if err != nil && statusError.Status().Code != http.StatusNotFound {
		impl.logger.Errorw("secret not found", "err", err)
		return err
	}
	data := make(map[string][]byte)
	data["username"] = []byte(request.Username)
	data["password"] = []byte(request.Token)
	if secret == nil {
		secret, err = impl.K8sUtil.CreateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, data, GitOpsSecretName, client)
		if err != nil {
			impl.logger.Errorw("err on creating secret", "err", err)
			return err
		}
	} else {
		secret.Data = data
		secret, err = impl.K8sUtil.UpdateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, secret, client)
		if err != nil {
			operationComplete := false
			retryCount := 0
			for !operationComplete && retryCount < 3 {
				retryCount = retryCount + 1
				secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, GitOpsSecretName, client)
				if err != nil {
					impl.logger.Errorw("secret not found", "err", err)
					return err
				}
				secret.Data = data
				secret, err = impl.K8sUtil.UpdateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, secret, client)
				if err != nil {
					continue
				}
				if err == nil {
					operationComplete = true
				}
			}

		}
	}
	if strings.ToUpper(request.Provider) == GITHUB_PROVIDER {
		orgUrl, err := impl.buildGithubOrgUrl(request.Host, request.GitHubOrgId)
		if err != nil {
			return err
		}
		request.Host = orgUrl
	}
	if strings.ToUpper(request.Provider) == GITLAB_PROVIDER {
		groupName, err := impl.gitFactory.GetGitLabGroupPath(request)
		if err != nil {
			return err
		}
		slashSuffixPresent := strings.HasSuffix(request.Host, "/")
		if slashSuffixPresent {
			request.Host += groupName
		} else {
			request.Host = fmt.Sprintf(request.Host+"/%s", groupName)
		}
	}
	if strings.ToUpper(request.Provider) == BITBUCKET_PROVIDER {
		request.Host = util.BITBUCKET_CLONE_BASE_URL + request.BitBucketWorkspaceId
	}
	operationComplete := false
	retryCount := 0
	for !operationComplete && retryCount < 3 {
		retryCount = retryCount + 1

		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return err
		}
		updatedData := impl.updateData(cm.Data, request, GitOpsSecretName, existingModel.Host)
		data := cm.Data
		data["repository.credentials"] = updatedData["repository.credentials"]
		cm.Data = data
		_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			continue
		}
		if err == nil {
			operationComplete = true
		}
	}
	if !operationComplete {
		return fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	err = impl.gitFactory.Reload()
	if err != nil {
		return err
	}
	return nil
}

func (impl *GitOpsConfigServiceImpl) GetGitOpsConfigById(id int) (*bean2.GitOpsConfigDto, error) {
	model, err := impl.gitOpsRepository.GetGitOpsConfigById(id)
	if err != nil {
		impl.logger.Errorw("GetGitOpsConfigById, error while get by id", "err", err, "id", id)
		return nil, err
	}
	config := &bean2.GitOpsConfigDto{
		Id:                   model.Id,
		Provider:             model.Provider,
		GitHubOrgId:          model.GitHubOrgId,
		GitLabGroupId:        model.GitLabGroupId,
		Username:             model.Username,
		Token:                model.Token,
		Host:                 model.Host,
		Active:               model.Active,
		UserId:               model.CreatedBy,
		AzureProjectName:     model.AzureProject,
		BitBucketWorkspaceId: model.BitBucketWorkspaceId,
		BitBucketProjectKey:  model.BitBucketProjectKey,
	}

	return config, err
}

func (impl *GitOpsConfigServiceImpl) GetAllGitOpsConfig() ([]*bean2.GitOpsConfigDto, error) {
	models, err := impl.gitOpsRepository.GetAllGitOpsConfig()
	if err != nil {
		impl.logger.Errorw("GetAllGitOpsConfig, error while fetch all", "err", err)
		return nil, err
	}
	configs := make([]*bean2.GitOpsConfigDto, 0)
	for _, model := range models {
		config := &bean2.GitOpsConfigDto{
			Id:                   model.Id,
			Provider:             model.Provider,
			GitHubOrgId:          model.GitHubOrgId,
			GitLabGroupId:        model.GitLabGroupId,
			Username:             model.Username,
			Token:                model.Token,
			Host:                 model.Host,
			Active:               model.Active,
			UserId:               model.CreatedBy,
			AzureProjectName:     model.AzureProject,
			BitBucketWorkspaceId: model.BitBucketWorkspaceId,
			BitBucketProjectKey:  model.BitBucketProjectKey,
		}
		configs = append(configs, config)
	}
	return configs, err
}

func (impl *GitOpsConfigServiceImpl) GetGitOpsConfigByProvider(provider string) (*bean2.GitOpsConfigDto, error) {
	model, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(provider)
	if err != nil {
		impl.logger.Errorw("GetGitOpsConfigByProvider, error while get by name", "err", err, "provider", provider)
		return nil, err
	}
	config := &bean2.GitOpsConfigDto{
		Id:                   model.Id,
		Provider:             model.Provider,
		GitHubOrgId:          model.GitHubOrgId,
		GitLabGroupId:        model.GitLabGroupId,
		Username:             model.Username,
		Token:                model.Token,
		Host:                 model.Host,
		Active:               model.Active,
		UserId:               model.CreatedBy,
		AzureProjectName:     model.AzureProject,
		BitBucketWorkspaceId: model.BitBucketWorkspaceId,
		BitBucketProjectKey:  model.BitBucketProjectKey,
	}

	return config, err
}

func (impl *GitOpsConfigServiceImpl) updateData(data map[string]string, request *bean2.GitOpsConfigDto, secretName string, existingHost string) map[string]string {
	var newRepositories []*RepositoryCredentialsDto
	var existingRepositories []*RepositoryCredentialsDto
	repoStr := data["repository.credentials"]
	if len(repoStr) > 0 {
		repoByte, err := yaml.YAMLToJSON([]byte(repoStr))
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(repoByte, &existingRepositories)
		if err != nil {
			panic(err)
		}
	}

	for _, item := range existingRepositories {
		if item.Url != existingHost {
			newRepositories = append(newRepositories, item)
		}
	}
	repoData := impl.createRepoElement(secretName, request)
	newRepositories = append(newRepositories, repoData)

	rb, err := json.Marshal(newRepositories)
	if err != nil {
		panic(err)
	}
	repositoriesYamlByte, err := yaml.JSONToYAML(rb)
	if err != nil {
		panic(err)
	}
	repositoryCredentials := map[string]string{}
	if len(repositoriesYamlByte) > 0 {
		repositoryCredentials["repository.credentials"] = string(repositoriesYamlByte)
	}
	return repositoryCredentials
}

func (impl *GitOpsConfigServiceImpl) createRepoElement(secretName string, request *bean2.GitOpsConfigDto) *RepositoryCredentialsDto {
	repoData := &RepositoryCredentialsDto{}
	usernameSecret := &KeyDto{Name: secretName, Key: "username"}
	passwordSecret := &KeyDto{Name: secretName, Key: "password"}
	repoData.PasswordSecret = passwordSecret
	repoData.UsernameSecret = usernameSecret
	repoData.Url = request.Host
	return repoData
}

type RepositoryCredentialsDto struct {
	Url            string  `json:"url,omitempty"`
	UsernameSecret *KeyDto `json:"usernameSecret,omitempty"`
	PasswordSecret *KeyDto `json:"passwordSecret,omitempty"`
}

type KeyDto struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
}

func (impl *GitOpsConfigServiceImpl) GetGitOpsConfigActive() (*bean2.GitOpsConfigDto, error) {
	model, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil {
		impl.logger.Errorw("GetGitOpsConfigActive, error while getting error", "err", err)
		return nil, err
	}
	config := &bean2.GitOpsConfigDto{
		Id:                   model.Id,
		Provider:             model.Provider,
		GitHubOrgId:          model.GitHubOrgId,
		GitLabGroupId:        model.GitLabGroupId,
		Active:               model.Active,
		UserId:               model.CreatedBy,
		AzureProjectName:     model.AzureProject,
		BitBucketWorkspaceId: model.BitBucketWorkspaceId,
		BitBucketProjectKey:  model.BitBucketProjectKey,
	}
	return config, err
}

func (impl *GitOpsConfigServiceImpl) GitOpsValidateDryRun(config *bean2.GitOpsConfigDto) DetailedErrorGitOpsConfigResponse {
	detailedErrorGitOpsConfigActions := util.DetailedErrorGitOpsConfigActions{}
	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)
	/*if strings.ToUpper(config.Provider) == GITHUB_PROVIDER {
		config.Host = GITHUB_HOST
	}*/
	if strings.ToUpper(config.Provider) == BITBUCKET_PROVIDER {
		config.Host = util.BITBUCKET_CLONE_BASE_URL
		config.BitBucketProjectKey = strings.ToUpper(config.BitBucketProjectKey)
	}
	client, gitService, err := impl.gitFactory.NewClientForValidation(config)
	if err != nil {
		impl.logger.Errorw("error in creating new client for validation")
		detailedErrorGitOpsConfigActions.StageErrorMap[fmt.Sprintf("error in connecting with %s", strings.ToUpper(config.Provider))] = impl.extractErrorMessageByProvider(err, config.Provider)
		detailedErrorGitOpsConfigActions.ValidatedOn = time.Now()
		detailedErrorGitOpsConfigResponse := impl.convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions)
		return detailedErrorGitOpsConfigResponse
	}
	appName := DryrunRepoName + util2.Generate(6)
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(config.UserId)
	repoUrl, _, detailedErrorCreateRepo := client.CreateRepository(appName, "sample dry-run repo", config.BitBucketWorkspaceId, config.BitBucketProjectKey, userName, userEmailId)

	detailedErrorGitOpsConfigActions.StageErrorMap = detailedErrorCreateRepo.StageErrorMap
	detailedErrorGitOpsConfigActions.SuccessfulStages = detailedErrorCreateRepo.SuccessfulStages

	for stage, stageErr := range detailedErrorGitOpsConfigActions.StageErrorMap {
		if stage == CreateRepoStage || stage == GetRepoUrlStage {
			_, ok := detailedErrorGitOpsConfigActions.StageErrorMap[GetRepoUrlStage]
			if ok {
				detailedErrorGitOpsConfigActions.StageErrorMap[fmt.Sprintf("error in connecting with %s", strings.ToUpper(config.Provider))] = impl.extractErrorMessageByProvider(stageErr, config.Provider)
				delete(detailedErrorGitOpsConfigActions.StageErrorMap, GetRepoUrlStage)
			} else {
				detailedErrorGitOpsConfigActions.StageErrorMap[CreateRepoStage] = impl.extractErrorMessageByProvider(stageErr, config.Provider)
			}
			detailedErrorGitOpsConfigActions.ValidatedOn = time.Now()
			detailedErrorGitOpsConfigResponse := impl.convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions)
			return detailedErrorGitOpsConfigResponse
		} else if stage == CloneHttp || stage == CreateReadmeStage {
			detailedErrorGitOpsConfigActions.StageErrorMap[stage] = impl.extractErrorMessageByProvider(stageErr, config.Provider)
		}
	}
	chartDir := fmt.Sprintf("%s-%s", appName, impl.getDir())
	clonedDir := gitService.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		clonedDir, err = gitService.Clone(repoUrl, chartDir)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
			detailedErrorGitOpsConfigActions.StageErrorMap[CloneStage] = err
		} else {
			detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneStage)
		}
	}

	commit, err := gitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in commit and pushing git", "err", err)
		if commit == "" {
			detailedErrorGitOpsConfigActions.StageErrorMap[CommitOnRestStage] = err
		} else {
			detailedErrorGitOpsConfigActions.StageErrorMap[PushStage] = err
		}
	} else {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CommitOnRestStage, PushStage)
	}
	repoOptions := &bitbucket.RepositoryOptions{
		Owner:     config.BitBucketWorkspaceId,
		RepoSlug:  appName,
		IsPrivate: "true",
		Project:   config.BitBucketProjectKey,
	}
	err = client.DeleteRepository(appName, config.Username, config.GitHubOrgId, config.AzureProjectName, repoOptions)
	if err != nil {
		impl.logger.Errorw("error in deleting repo", "err", err)
		//here below the assignment of delete is removed for making this stage optional, and it's failure not preventing it from saving/updating gitOps config
		//detailedErrorGitOpsConfigActions.StageErrorMap[DeleteRepoStage] = impl.extractErrorMessageByProvider(err, config.Provider)
		detailedErrorGitOpsConfigActions.DeleteRepoFailed = true
	} else {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, DeleteRepoStage)
	}
	detailedErrorGitOpsConfigActions.ValidatedOn = time.Now()
	defer impl.cleanDir(clonedDir)
	detailedErrorGitOpsConfigResponse := impl.convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions)
	return detailedErrorGitOpsConfigResponse
}

func (impl *GitOpsConfigServiceImpl) cleanDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		impl.logger.Warnw("error in deleting dir ", "dir", dir)
	}
}

func (impl *GitOpsConfigServiceImpl) getDir() string {
	/* #nosec */
	r1 := rand.New(impl.randSource).Int63()
	return strconv.FormatInt(r1, 10)
}

func (impl *GitOpsConfigServiceImpl) extractErrorMessageByProvider(err error, provider string) error {
	if provider == GITLAB_PROVIDER {
		errorResponse, ok := err.(*gitlab.ErrorResponse)
		if ok {
			errorMessage := fmt.Errorf("%s", errorResponse.Message)
			return errorMessage
		}
	} else if provider == AZURE_DEVOPS_PROVIDER {
		if errorResponse, ok := err.(azuredevops.WrappedError); ok {
			errorMessage := fmt.Errorf("%s", *errorResponse.Message)
			return errorMessage
		} else if errorResponse, ok := err.(*azuredevops.WrappedError); ok {
			errorMessage := fmt.Errorf("%s", *errorResponse.Message)
			return errorMessage
		}
	}
	return err
}

func (impl *GitOpsConfigServiceImpl) convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions util.DetailedErrorGitOpsConfigActions) (detailedErrorResponse DetailedErrorGitOpsConfigResponse) {
	detailedErrorResponse.StageErrorMap = make(map[string]string)
	detailedErrorResponse.SuccessfulStages = detailedErrorGitOpsConfigActions.SuccessfulStages
	for stage, err := range detailedErrorGitOpsConfigActions.StageErrorMap {
		detailedErrorResponse.StageErrorMap[stage] = err.Error()
	}
	detailedErrorResponse.DeleteRepoFailed = detailedErrorGitOpsConfigActions.DeleteRepoFailed
	return detailedErrorResponse
}
