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

package gitops

import (
	"context"
	"encoding/json"
	"fmt"
	certificate2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/certificate"
	repocreds2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	util4 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	apiBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/client/argocdServer/certificate"
	repocreds "github.com/devtron-labs/devtron/client/argocdServer/repocreds"
	repository2 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation"
	gitOpsBean "github.com/devtron-labs/devtron/pkg/gitops/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"net/http"
	"strings"
	"time"

	cluster3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	cluster2 "github.com/devtron-labs/devtron/client/argocdServer/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/sql"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/yaml"
)

type GitOpsConfigService interface {
	ValidateAndCreateGitOpsConfig(config *apiBean.GitOpsConfigDto) (apiBean.DetailedErrorGitOpsConfigResponse, error)
	ValidateAndUpdateGitOpsConfig(config *apiBean.GitOpsConfigDto) (apiBean.DetailedErrorGitOpsConfigResponse, error)
	GitOpsValidateDryRun(config *apiBean.GitOpsConfigDto) apiBean.DetailedErrorGitOpsConfigResponse
	GetGitOpsConfigById(id int) (*apiBean.GitOpsConfigDto, error)
	GetAllGitOpsConfig() ([]*apiBean.GitOpsConfigDto, error)
	GetGitOpsConfigByProvider(provider string) (*apiBean.GitOpsConfigDto, error)
}

type GitOpsConfigServiceImpl struct {
	logger                  *zap.SugaredLogger
	gitOpsRepository        repository.GitOpsConfigRepository
	K8sUtil                 *util4.K8sServiceImpl
	aCDAuthConfig           *util3.ACDAuthConfig
	clusterService          cluster.ClusterService
	argoUserService         argo.ArgoUserService
	clusterServiceCD        cluster2.ServiceClient
	gitOpsConfigReadService config.GitOpsConfigReadService
	gitOperationService     git.GitOperationService
	gitOpsValidationService validation.GitOpsValidationService
	argoCertificateClient   certificate.Client
	argoRepoService         repository2.ServiceClient
	repocreds               repocreds.ServiceClient
}

func NewGitOpsConfigServiceImpl(Logger *zap.SugaredLogger,
	gitOpsRepository repository.GitOpsConfigRepository,
	K8sUtil *util4.K8sServiceImpl, aCDAuthConfig *util3.ACDAuthConfig,
	clusterService cluster.ClusterService,
	argoUserService argo.ArgoUserService,
	clusterServiceCD cluster2.ServiceClient,
	gitOperationService git.GitOperationService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOpsValidationService validation.GitOpsValidationService,
	argoCertificateClient certificate.Client,
	argoRepoService repository2.ServiceClient,
	repocreds repocreds.ServiceClient) *GitOpsConfigServiceImpl {
	return &GitOpsConfigServiceImpl{
		logger:                  Logger,
		gitOpsRepository:        gitOpsRepository,
		K8sUtil:                 K8sUtil,
		aCDAuthConfig:           aCDAuthConfig,
		clusterService:          clusterService,
		argoUserService:         argoUserService,
		clusterServiceCD:        clusterServiceCD,
		gitOpsConfigReadService: gitOpsConfigReadService,
		gitOperationService:     gitOperationService,
		gitOpsValidationService: gitOpsValidationService,
		argoCertificateClient:   argoCertificateClient,
		argoRepoService:         argoRepoService,
		repocreds:               repocreds,
	}
}

func (impl *GitOpsConfigServiceImpl) ValidateAndCreateGitOpsConfig(config *apiBean.GitOpsConfigDto) (apiBean.DetailedErrorGitOpsConfigResponse, error) {
	detailedErrorGitOpsConfigResponse := impl.GitOpsValidateDryRun(config)
	if len(detailedErrorGitOpsConfigResponse.StageErrorMap) == 0 {
		//create argo-cd user, if not created, here argo-cd integration has to be installed
		token := impl.argoUserService.GetOrUpdateArgoCdUserDetail()
		ctx := context.WithValue(context.Background(), "token", token)
		_, err := impl.createGitOpsConfig(ctx, config)
		if err != nil {
			impl.logger.Errorw("service err, SaveGitRepoConfig", "err", err, "payload", config)
			return detailedErrorGitOpsConfigResponse, err
		}
	}
	return detailedErrorGitOpsConfigResponse, nil
}

func (impl *GitOpsConfigServiceImpl) ValidateAndUpdateGitOpsConfig(config *apiBean.GitOpsConfigDto) (apiBean.DetailedErrorGitOpsConfigResponse, error) {
	isTokenEmpty := config.Token == ""
	isTlsDetailsEmpty := config.EnableTLSVerification &&
		(config.TLSConfig == nil ||
			(config.TLSConfig != nil && (len(config.TLSConfig.CaData) == 0 || len(config.TLSConfig.TLSCertData) == 0 || len(config.TLSConfig.TLSKeyData) == 0)))

	if isTokenEmpty || isTlsDetailsEmpty {
		model, err := impl.gitOpsRepository.GetGitOpsConfigById(config.Id)
		if err != nil {
			impl.logger.Errorw("No matching entry found for update.", "id", config.Id)
			err = &util.ApiError{
				InternalMessage: "gitops config update failed, does not exist",
				UserMessage:     "gitops config update failed, does not exist",
			}
			return apiBean.DetailedErrorGitOpsConfigResponse{}, err
		}
		if isTokenEmpty {
			config.Token = model.Token
		}
		if isTlsDetailsEmpty {
			caData := model.CaCert
			tlsCert := model.TlsCert
			tlsKey := model.TlsKey

			if config.TLSConfig != nil {
				if len(config.TLSConfig.CaData) > 0 {
					caData = config.TLSConfig.CaData
				}
				if len(config.TLSConfig.TLSCertData) > 0 {
					tlsCert = config.TLSConfig.TLSCertData
				}
				if len(config.TLSConfig.TLSKeyData) > 0 {
					tlsKey = config.TLSConfig.TLSKeyData
				}
			}

			config.TLSConfig = &bean.TLSConfig{
				CaData:      caData,
				TLSCertData: tlsCert,
				TLSKeyData:  tlsKey,
			}
		}
	}
	detailedErrorGitOpsConfigResponse := impl.GitOpsValidateDryRun(config)
	if len(detailedErrorGitOpsConfigResponse.StageErrorMap) == 0 {
		err := impl.updateGitOpsConfig(config)
		if err != nil {
			impl.logger.Errorw("service err, updateGitOpsConfig", "err", err, "payload", config)
			return detailedErrorGitOpsConfigResponse, err
		}
	}
	return detailedErrorGitOpsConfigResponse, nil
}

// step-1: save data in DB
// step-3: add ca cert if present to list of trusted certificates on argoCD using certificate.Client service
// step-3: add repository credentials in secret declared using env variable GITOPS_SECRET_NAME
// step-4 add repository URL in argocd-cm, argocd-cm will have reference to secret created in step-3 for credentials
// steps-5 upsert cluster in acd
func (impl *GitOpsConfigServiceImpl) createGitOpsConfig(ctx context.Context, request *apiBean.GitOpsConfigDto) (*apiBean.GitOpsConfigDto, error) {
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
		Provider:              strings.ToUpper(request.Provider),
		Username:              request.Username,
		Token:                 request.Token,
		GitLabGroupId:         request.GitLabGroupId,
		GitHubOrgId:           request.GitHubOrgId,
		AzureProject:          request.AzureProjectName,
		Host:                  request.Host,
		Active:                true,
		AllowCustomRepository: request.AllowCustomRepository,
		BitBucketWorkspaceId:  request.BitBucketWorkspaceId,
		BitBucketProjectKey:   request.BitBucketProjectKey,
		EnableTLSVerification: request.EnableTLSVerification,
		AuditLog:              sql.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}

	if request.EnableTLSVerification {
		if len(request.TLSConfig.CaData) > 0 {
			model.CaCert = request.TLSConfig.CaData
		}
		if len(request.TLSConfig.TLSCertData) > 0 && len(request.TLSConfig.TLSKeyData) > 0 {
			model.TlsKey = request.TLSConfig.TLSKeyData
			model.TlsCert = request.TLSConfig.TLSCertData
		}

		if !request.IsCADataPresent {
			model.CaCert = ""
		}
		if !request.IsTLSCertDataPresent {
			model.TlsCert = ""
		}
		if !request.IsTLSKeyDataPresent {
			model.TlsKey = ""
		}

		if (len(model.TlsKey) > 0 && len(model.TlsCert) == 0) || (len(model.TlsKey) == 0 && len(model.TlsCert) > 0) {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				InternalMessage: "failed to update gitOps config in db",
				UserMessage:     "failed to update gitOps config in db",
			}
		}
		if len(model.CaCert) == 0 && len(model.TlsKey) == 0 && len(model.TlsCert) == 0 {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				InternalMessage: "failed to update gitOps config in db",
				UserMessage:     "failed to update gitOps config in db",
			}
		}

	}
	if model.EnableTLSVerification && request.TLSConfig == nil {
		request.TLSConfig = &bean.TLSConfig{
			CaData:      model.CaCert,
			TLSCertData: model.TlsCert,
			TLSKeyData:  model.TlsKey,
		}
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

	if model.EnableTLSVerification {

		err = impl.gitOperationService.UpdateGitHostUrlByProvider(request)
		if err != nil {
			return nil, err
		}
		_, err = impl.repocreds.CreateRepoCreds(ctx, &repocreds2.RepoCredsCreateRequest{
			Creds: &v1alpha1.RepoCreds{
				URL:               request.Host,
				Username:          model.Username,
				Password:          model.Token,
				TLSClientCertData: model.TlsCert,
				TLSClientCertKey:  model.TlsKey,
			},
			Upsert: true,
		})
		if err != nil {
			impl.logger.Errorw("error in saving repo credential template to argocd", "err", err)
			return nil, err
		}

		err = impl.addCACertInArgoIfPresent(ctx, model)
		if err != nil {
			impl.logger.Errorw("error in adding ca cert to argo", "err", err)
			return nil, err
		}

	} else {

		clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
		if err != nil {
			return nil, err
		}
		cfg := clusterBean.GetClusterConfig()

		client, err := impl.K8sUtil.GetCoreV1Client(cfg)
		if err != nil {
			return nil, err
		}
		secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.GitOpsSecretName, client)
		statusError, _ := err.(*errors.StatusError)
		if err != nil && statusError.Status().Code != http.StatusNotFound {
			impl.logger.Errorw("secret not found", "err", err)
			return nil, err
		}
		data := make(map[string][]byte)
		data[gitOpsBean.USERNAME] = []byte(request.Username)
		data[gitOpsBean.PASSWORD] = []byte(request.Token)

		if secret == nil {
			secret, err = impl.K8sUtil.CreateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, data, impl.aCDAuthConfig.GitOpsSecretName, "", client, nil, nil)
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
					secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.GitOpsSecretName, client)
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
			err = impl.gitOperationService.UpdateGitHostUrlByProvider(request)
			if err != nil {
				return nil, err
			}
			operationComplete := false
			retryCount := 0
			for !operationComplete && retryCount < 3 {
				retryCount = retryCount + 1

				cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
				if err != nil {
					return nil, err
				}
				currentHost := request.Host
				updatedData := impl.updateData(cm.Data, request, impl.aCDAuthConfig.GitOpsSecretName, currentHost)
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
		}
	}

	// if git-ops config is created/saved successfully (just before transaction commit) and this was first git-ops config, then upsert clusters in acd
	gitOpsConfigurationStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		return nil, err
	}
	if !gitOpsConfigurationStatus.IsGitOpsConfigured {
		clusters, err := impl.clusterService.FindAllActive()
		if err != nil {
			impl.logger.Errorw("Error while fetching all the clusters", "err", err)
			return nil, err
		}
		for _, cluster := range clusters {
			cl := impl.clusterService.ConvertClusterBeanObjectToCluster(&cluster)
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

	err = impl.gitOperationService.ReloadGitOpsProvider()
	if err != nil {
		return nil, err
	}
	request.Id = model.Id
	return request, nil
}

func (impl *GitOpsConfigServiceImpl) addCACertInArgoIfPresent(ctx context.Context, model *repository.GitOpsConfig) error {
	if len(model.CaCert) > 0 {
		host, err := util2.GetHost(model.Host)
		if err != nil {
			impl.logger.Errorw("invalid gitOps host", "host", host, "err", err)
			return err
		}
		_, err = impl.argoCertificateClient.CreateCertificate(ctx, &certificate2.RepositoryCertificateCreateRequest{
			Certificates: &v1alpha1.RepositoryCertificateList{
				Items: []v1alpha1.RepositoryCertificate{{
					ServerName: host,
					CertData:   []byte(model.CaCert),
					CertType:   "https",
				}},
			},
			Upsert: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl *GitOpsConfigServiceImpl) updateGitOpsConfig(request *apiBean.GitOpsConfigDto) error {
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
	model.AllowCustomRepository = request.AllowCustomRepository
	model.EnableTLSVerification = request.EnableTLSVerification
	model.UpdatedBy = request.UserId
	model.UpdatedOn = time.Now()

	if request.EnableTLSVerification {
		if len(request.TLSConfig.CaData) > 0 {
			model.CaCert = request.TLSConfig.CaData
		}
		if len(request.TLSConfig.TLSCertData) > 0 && len(request.TLSConfig.TLSKeyData) > 0 {
			model.TlsKey = request.TLSConfig.TLSKeyData
			model.TlsCert = request.TLSConfig.TLSCertData
		}

		if !request.IsCADataPresent {
			model.CaCert = ""
		}
		if !request.IsTLSCertDataPresent {
			model.TlsCert = ""
		}
		if !request.IsTLSKeyDataPresent {
			model.TlsKey = ""
		}

		if (len(model.TlsKey) > 0 && len(model.TlsCert) == 0) || (len(model.TlsKey) == 0 && len(model.TlsCert) > 0) {
			return &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				InternalMessage: "failed to update gitOps config in db",
				UserMessage:     "failed to update gitOps config in db",
			}
		}
		if len(model.CaCert) == 0 && len(model.TlsKey) == 0 && len(model.TlsCert) == 0 {
			return &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				InternalMessage: "failed to update gitOps config in db",
				UserMessage:     "failed to update gitOps config in db",
			}
		}
	} else {
		model.TlsKey = ""
		model.TlsCert = ""
		model.CaCert = ""
	}
	if model.EnableTLSVerification && request.TLSConfig == nil {
		request.TLSConfig = &bean.TLSConfig{
			CaData:      model.CaCert,
			TLSCertData: model.TlsCert,
			TLSKeyData:  model.TlsKey,
		}
	}

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

	if model.EnableTLSVerification {

		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
			return err
		}
		ctx := context.WithValue(context.Background(), "token", acdToken)

		err = impl.gitOperationService.UpdateGitHostUrlByProvider(request)
		if err != nil {
			return err
		}

		_, err = impl.repocreds.CreateRepoCreds(ctx, &repocreds2.RepoCredsCreateRequest{
			Creds: &v1alpha1.RepoCreds{
				URL:               request.Host,
				Username:          model.Username,
				Password:          model.Token,
				TLSClientCertData: model.TlsCert,
				TLSClientCertKey:  model.TlsKey,
			},
			Upsert: true,
		})
		if err != nil {
			impl.logger.Errorw("error in saving repo credential template to argocd", "err", err)
			return err
		}

		err = impl.addCACertInArgoIfPresent(ctx, model)
		if err != nil {
			impl.logger.Errorw("error in adding ca cert to argo", "err", err)
			return err
		}

	} else {
		clusterBean, err := impl.clusterService.FindOne(cluster.DEFAULT_CLUSTER)
		if err != nil {
			return err
		}
		cfg := clusterBean.GetClusterConfig()
		if err != nil {
			return err
		}

		client, err := impl.K8sUtil.GetCoreV1Client(cfg)
		if err != nil {
			return err
		}

		secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.GitOpsSecretName, client)
		statusError, _ := err.(*errors.StatusError)
		if err != nil && statusError.Status().Code != http.StatusNotFound {
			impl.logger.Errorw("secret not found", "err", err)
			return err
		}
		data := make(map[string][]byte)
		data[gitOpsBean.USERNAME] = []byte(request.Username)
		data[gitOpsBean.PASSWORD] = []byte(request.Token)

		if secret == nil {
			secret, err = impl.K8sUtil.CreateSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, data, impl.aCDAuthConfig.GitOpsSecretName, "", client, nil, nil)
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
					secret, err := impl.K8sUtil.GetSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.GitOpsSecretName, client)
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
		err = impl.gitOperationService.UpdateGitHostUrlByProvider(request)
		if err != nil {
			return err
		}
		operationComplete := false
		retryCount := 0
		for !operationComplete && retryCount < 3 {
			retryCount = retryCount + 1

			cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
			if err != nil {
				return err
			}
			currentHost := request.Host
			updatedData := impl.updateData(cm.Data, request, impl.aCDAuthConfig.GitOpsSecretName, currentHost)
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

	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	err = impl.gitOperationService.ReloadGitOpsProvider()
	if err != nil {
		return err
	}
	return nil
}

func (impl *GitOpsConfigServiceImpl) GetGitOpsConfigById(id int) (*apiBean.GitOpsConfigDto, error) {
	model, err := impl.gitOpsRepository.GetGitOpsConfigById(id)
	if err != nil {
		impl.logger.Errorw("GetGitOpsConfigById, error while get by id", "err", err, "id", id)
		return nil, err
	}
	config := &apiBean.GitOpsConfigDto{
		Id:                    model.Id,
		Provider:              model.Provider,
		GitHubOrgId:           model.GitHubOrgId,
		GitLabGroupId:         model.GitLabGroupId,
		Username:              model.Username,
		Token:                 model.Token,
		Host:                  model.Host,
		Active:                model.Active,
		UserId:                model.CreatedBy,
		AzureProjectName:      model.AzureProject,
		BitBucketWorkspaceId:  model.BitBucketWorkspaceId,
		BitBucketProjectKey:   model.BitBucketProjectKey,
		AllowCustomRepository: model.AllowCustomRepository,
		EnableTLSVerification: model.EnableTLSVerification,
		TLSConfig: &bean.TLSConfig{ // sending empty values as they are hidden in FE
			CaData:      "",
			TLSCertData: "",
			TLSKeyData:  "",
		},
		IsCADataPresent:      len(model.CaCert) > 0,
		IsTLSCertDataPresent: len(model.TlsCert) > 0,
		IsTLSKeyDataPresent:  len(model.TlsKey) > 0,
	}
	return config, err
}

func (impl *GitOpsConfigServiceImpl) GetAllGitOpsConfig() ([]*apiBean.GitOpsConfigDto, error) {
	models, err := impl.gitOpsRepository.GetAllGitOpsConfig()
	if err != nil {
		impl.logger.Errorw("GetAllGitOpsConfig, error while fetch all", "err", err)
		return nil, err
	}
	configs := make([]*apiBean.GitOpsConfigDto, 0)
	for _, model := range models {
		config := &apiBean.GitOpsConfigDto{
			Id:                    model.Id,
			Provider:              model.Provider,
			GitHubOrgId:           model.GitHubOrgId,
			GitLabGroupId:         model.GitLabGroupId,
			Username:              model.Username,
			Token:                 "",
			Host:                  model.Host,
			Active:                model.Active,
			UserId:                model.CreatedBy,
			AzureProjectName:      model.AzureProject,
			BitBucketWorkspaceId:  model.BitBucketWorkspaceId,
			BitBucketProjectKey:   model.BitBucketProjectKey,
			AllowCustomRepository: model.AllowCustomRepository,
			EnableTLSVerification: model.EnableTLSVerification,
			TLSConfig: &bean.TLSConfig{ // sending empty values as they are hidden in FE
				CaData:      "",
				TLSCertData: "",
				TLSKeyData:  "",
			},
			IsCADataPresent:      len(model.CaCert) > 0,
			IsTLSCertDataPresent: len(model.TlsCert) > 0,
			IsTLSKeyDataPresent:  len(model.TlsKey) > 0,
		}
		configs = append(configs, config)
	}
	return configs, err
}

func (impl *GitOpsConfigServiceImpl) GetGitOpsConfigByProvider(provider string) (*apiBean.GitOpsConfigDto, error) {
	model, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(provider)
	if err != nil {
		impl.logger.Errorw("GetGitOpsConfigByProvider, error while get by name", "err", err, "provider", provider)
		return nil, err
	}
	config := &apiBean.GitOpsConfigDto{
		Id:                    model.Id,
		Provider:              model.Provider,
		GitHubOrgId:           model.GitHubOrgId,
		GitLabGroupId:         model.GitLabGroupId,
		Username:              model.Username,
		Token:                 model.Token,
		Host:                  model.Host,
		Active:                model.Active,
		UserId:                model.CreatedBy,
		AzureProjectName:      model.AzureProject,
		BitBucketWorkspaceId:  model.BitBucketWorkspaceId,
		BitBucketProjectKey:   model.BitBucketProjectKey,
		AllowCustomRepository: model.AllowCustomRepository,
		EnableTLSVerification: model.EnableTLSVerification,
		TLSConfig: &bean.TLSConfig{ // sending empty values as they are hidden in FE
			CaData:      "",
			TLSCertData: "",
			TLSKeyData:  "",
		},
		IsCADataPresent:      len(model.CaCert) > 0,
		IsTLSCertDataPresent: len(model.TlsCert) > 0,
		IsTLSKeyDataPresent:  len(model.TlsKey) > 0,
	}

	return config, err
}

func (impl *GitOpsConfigServiceImpl) GitOpsValidateDryRun(config *apiBean.GitOpsConfigDto) apiBean.DetailedErrorGitOpsConfigResponse {

	isTokenEmpty := config.Token == ""
	isTlsDetailsEmpty := config.EnableTLSVerification && (len(config.TLSConfig.CaData) == 0 && len(config.TLSConfig.TLSCertData) == 0 && len(config.TLSConfig.TLSKeyData) == 0)

	if isTokenEmpty || isTlsDetailsEmpty {
		model, err := impl.gitOpsRepository.GetGitOpsConfigById(config.Id)
		if err != nil {
			impl.logger.Errorw("No matching entry found for update.", "id", config.Id)
			err = &util.ApiError{
				InternalMessage: "gitops config update failed, does not exist",
				UserMessage:     "gitops config update failed, does not exist",
			}
			return apiBean.DetailedErrorGitOpsConfigResponse{}
		}
		if isTokenEmpty {
			config.Token = model.Token
		}
		if isTlsDetailsEmpty {
			caData := model.CaCert
			tlsCert := model.TlsCert
			tlsKey := model.TlsKey

			if config.TLSConfig != nil {
				if len(config.TLSConfig.CaData) > 0 {
					caData = config.TLSConfig.CaData
				}
				if len(config.TLSConfig.TLSCertData) > 0 {
					tlsCert = config.TLSConfig.TLSCertData
				}
				if len(config.TLSConfig.TLSKeyData) > 0 {
					tlsKey = config.TLSConfig.TLSKeyData
				}
			}
			config.TLSConfig = &bean.TLSConfig{
				CaData:      caData,
				TLSCertData: tlsCert,
				TLSKeyData:  tlsKey,
			}
		}
	}

	return impl.gitOpsValidationService.GitOpsValidateDryRun(config)
}

func (impl *GitOpsConfigServiceImpl) updateData(data map[string]string, request *apiBean.GitOpsConfigDto, secretName string, currentHost string) map[string]string {
	var newRepositories []*gitOpsBean.RepositoryCredentialsDto
	var existingRepositories []*gitOpsBean.RepositoryCredentialsDto
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
		if item.Url != currentHost {
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

func (impl *GitOpsConfigServiceImpl) createRepoElement(secretName string, request *apiBean.GitOpsConfigDto) *gitOpsBean.RepositoryCredentialsDto {
	repoData := &gitOpsBean.RepositoryCredentialsDto{}
	repoData.Url = request.Host

	usernameSecret := &gitOpsBean.KeyDto{Name: secretName, Key: gitOpsBean.USERNAME}
	passwordSecret := &gitOpsBean.KeyDto{Name: secretName, Key: gitOpsBean.PASSWORD}

	repoData.PasswordSecret = passwordSecret
	repoData.UsernameSecret = usernameSecret
	return repoData
}
