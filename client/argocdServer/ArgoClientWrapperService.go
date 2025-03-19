/*
 * Copyright (c) 2024. Devtron Inc.
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

package argocdServer

import (
	"context"
	"encoding/json"
	"fmt"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/certificate"
	cluster2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repocreds"
	repository2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/devtron/client/argocdServer/adapter"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	certificate2 "github.com/devtron-labs/devtron/client/argocdServer/certificate"
	"github.com/devtron-labs/devtron/client/argocdServer/cluster"
	config2 "github.com/devtron-labs/devtron/client/argocdServer/config"
	bean2 "github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient/bean"
	repocreds2 "github.com/devtron-labs/devtron/client/argocdServer/repocreds"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/util/retryFunc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"strings"
	"time"
)

type ACDConfig struct {
	ArgoCDAutoSyncEnabled     bool `env:"ARGO_AUTO_SYNC_ENABLED" envDefault:"true"` // will gradually switch this flag to false in enterprise
	RegisterRepoMaxRetryCount int  `env:"ARGO_REPO_REGISTER_RETRY_COUNT" envDefault:"3"`
	RegisterRepoMaxRetryDelay int  `env:"ARGO_REPO_REGISTER_RETRY_DELAY" envDefault:"10"`
}

func (config *ACDConfig) IsManualSyncEnabled() bool {
	return config.ArgoCDAutoSyncEnabled == false
}

func (config *ACDConfig) IsAutoSyncEnabled() bool {
	return config.ArgoCDAutoSyncEnabled == true
}

func GetACDDeploymentConfig() (*ACDConfig, error) {
	cfg := &ACDConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

const (
	ErrorOperationAlreadyInProgress = "another operation is already in progress" // this string is returned from argocd
)

type ApplicationClientWrapper interface {
	ResourceTree(ctxt context.Context, query *application2.ResourcesQuery) (*v1alpha1.ApplicationTree, error)
	GetArgoClient(ctxt context.Context) (application2.ApplicationServiceClient, *grpc.ClientConn, error)
	GetApplicationResource(ctx context.Context, query *application2.ApplicationResourceRequest) (*application2.ApplicationResourceResponse, error)
	DeleteArgoApp(ctx context.Context, appName string, cascadeDelete bool) (*application2.ApplicationResponse, error)

	// GetArgoAppByName fetches an argoCd app by its name
	GetArgoAppByName(ctx context.Context, appName string) (*v1alpha1.Application, error)

	GetArgoAppByNameWithK8sClient(ctx context.Context, clusterId int, namespace, appName string) (*v1alpha1.Application, error)

	DeleteArgoAppWithK8sClient(ctx context.Context, clusterId int, namespace, appName string, cascadeDelete bool) error

	SyncArgoCDApplicationAndRefreshWithK8sClient(ctx context.Context, clusterId int, namespace, appName string) error
	// UpdateArgoCDSyncModeIfNeeded - if ARGO_AUTO_SYNC_ENABLED=true and app is in manual sync mode or vice versa update app
	UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error)

	// RegisterGitOpsRepoInArgoWithRetry - register a repository in argo-cd with retry mechanism
	RegisterGitOpsRepoInArgoWithRetry(ctx context.Context, gitOpsRepoUrl, targetRevision string, userId int32) error

	// PatchArgoCdApp performs a patch operation on an argoCd app
	PatchArgoCdApp(ctx context.Context, dto *bean.ArgoCdAppPatchReqDto) error

	// IsArgoAppPatchRequired decides weather the v1alpha1.ApplicationSource requires to be updated
	IsArgoAppPatchRequired(argoAppSpec *v1alpha1.ApplicationSource, currentGitRepoUrl, currentTargetRevision, currentChartPath string) bool

	// GetGitOpsRepoName returns the GitOps repository name, configured for the argoCd app
	GetGitOpsRepoNameForApplication(ctx context.Context, appName string) (gitOpsRepoName string, err error)

	GetGitOpsRepoURLForApplication(ctx context.Context, appName string) (gitOpsRepoURL string, err error)
}

type RepositoryClientWrapper interface {
	RegisterGitOpsRepoInArgoWithRetry(ctx context.Context, gitOpsRepoUrl, targetRevision string, userId int32) error
}

type RepoCredsClientWrapper interface {
	CreateRepoCreds(ctx context.Context, query *repocreds.RepoCredsCreateRequest) (*v1alpha1.RepoCreds, error)
	AddOrUpdateOCIRegistry(username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error
	DeleteOCIRegistry(registryURL, repo string, ociRegistryId int) error
	AddChartRepository(request bean2.ChartRepositoryAddRequest) error
	UpdateChartRepository(request bean2.ChartRepositoryUpdateRequest) error
	DeleteChartRepository(name, url string) error
}

type CertificateClientWrapper interface {
	CreateCertificate(ctx context.Context, query *certificate.RepositoryCertificateCreateRequest) (*v1alpha1.RepositoryCertificateList, error)
	DeleteCertificate(ctx context.Context, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error)
}

type ClusterClientWrapper interface {
	CreateCluster(ctx context.Context, clusterRequest *cluster2.ClusterCreateRequest) (*v1alpha1.Cluster, error)
	UpdateCluster(ctx context.Context, clusterRequest *cluster2.ClusterUpdateRequest) (*v1alpha1.Cluster, error)
}

type ArgoClientWrapperService interface {
	ClusterClientWrapper
	ApplicationClientWrapper
	RepositoryClientWrapper
	RepoCredsClientWrapper
	CertificateClientWrapper
}

type ArgoClientWrapperServiceImpl struct {
	acdApplicationClient    application.ServiceClient
	repositoryService       repository.ServiceClient
	clusterClient           cluster.ServiceClient
	repoCredsClient         repocreds2.ServiceClient
	CertificateClient       certificate2.ServiceClient
	logger                  *zap.SugaredLogger
	ACDConfig               *ACDConfig
	gitOpsConfigReadService config.GitOpsConfigReadService
	gitOperationService     git.GitOperationService
	asyncRunnable           *async.Runnable
	acdConfigGetter         config2.ArgoCDConfigGetter
	argoK8sClient           ArgoK8sClient
	*ArgoClientWrapperServiceEAImpl
}

func NewArgoClientWrapperServiceImpl(
	acdClient application.ServiceClient,
	repositoryService repository.ServiceClient,
	clusterClient cluster.ServiceClient,
	repocredsClient repocreds2.ServiceClient,
	CertificateClient certificate2.ServiceClient,
	logger *zap.SugaredLogger,
	ACDConfig *ACDConfig, gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOperationService git.GitOperationService, asyncRunnable *async.Runnable,
	acdConfigGetter config2.ArgoCDConfigGetter,
	ArgoClientWrapperServiceEAImpl *ArgoClientWrapperServiceEAImpl,
	argoK8sClient ArgoK8sClient,
) *ArgoClientWrapperServiceImpl {
	return &ArgoClientWrapperServiceImpl{
		acdApplicationClient:           acdClient,
		repositoryService:              repositoryService,
		clusterClient:                  clusterClient,
		repoCredsClient:                repocredsClient,
		CertificateClient:              CertificateClient,
		logger:                         logger,
		ACDConfig:                      ACDConfig,
		gitOpsConfigReadService:        gitOpsConfigReadService,
		gitOperationService:            gitOperationService,
		asyncRunnable:                  asyncRunnable,
		acdConfigGetter:                acdConfigGetter,
		ArgoClientWrapperServiceEAImpl: ArgoClientWrapperServiceEAImpl,
		argoK8sClient:                  argoK8sClient,
	}
}

func (impl *ArgoClientWrapperServiceImpl) ResourceTree(ctxt context.Context, query *application2.ResourcesQuery) (*v1alpha1.ApplicationTree, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, nil
	}
	return impl.acdApplicationClient.ResourceTree(ctxt, grpcConfig, query)
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoClient(ctxt context.Context) (application2.ApplicationServiceClient, *grpc.ClientConn, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, nil, err
	}
	return impl.acdApplicationClient.GetArgoClient(ctxt, grpcConfig)
}

func (impl *ArgoClientWrapperServiceImpl) GetApplicationResource(ctx context.Context, query *application2.ApplicationResourceRequest) (*application2.ApplicationResourceResponse, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, err
	}
	resource, err := impl.acdApplicationClient.GetResource(ctx, grpcConfig, query)
	if err != nil {
		impl.logger.Errorw("error in getting resource", "err", err)
		return nil, err
	}
	return resource, nil
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoApplication(ctx context.Context, query *application2.ApplicationQuery) (*v1alpha1.Application, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, err
	}
	return impl.acdApplicationClient.Get(ctx, grpcConfig, query)
}

func (impl *ArgoClientWrapperServiceImpl) DeleteArgoApp(ctx context.Context, appName string, cascadeDelete bool) (*application2.ApplicationResponse, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, err
	}
	req := &application2.ApplicationDeleteRequest{
		Name:    &appName,
		Cascade: &cascadeDelete,
	}
	return impl.acdApplicationClient.Delete(ctx, grpcConfig, req)
}

func (impl *ArgoClientWrapperServiceImpl) UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error) {
	if isArgoAppSyncModeMigrationNeeded(argoApplication, impl.ACDConfig) {

		grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
		if err != nil {
			impl.logger.Errorw("error in getting grpc config", "err", err)
			return err
		}

		syncModeUpdateRequest := createRequestForArgoCDSyncModeUpdateRequest(argoApplication, impl.ACDConfig.IsAutoSyncEnabled())
		validate := false
		_, err = impl.acdApplicationClient.Update(ctx, grpcConfig, &application2.ApplicationUpdateRequest{Application: syncModeUpdateRequest, Validate: &validate})
		if err != nil {
			impl.logger.Errorw("error in creating argo pipeline ", "name", argoApplication.Name, "err", err)
			return err
		}
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) RegisterGitOpsRepoInArgoWithRetry(ctx context.Context, gitOpsRepoUrl, targetRevision string, userId int32) error {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil
	}

	callback := func(int) error {
		return impl.createRepoInArgoCd(ctx, grpcConfig, gitOpsRepoUrl)
	}
	argoCdErr := retryFunc.Retry(callback,
		impl.isRetryableArgoRepoCreationError,
		impl.ACDConfig.RegisterRepoMaxRetryCount,
		time.Duration(impl.ACDConfig.RegisterRepoMaxRetryDelay)*time.Second,
		impl.logger)
	if argoCdErr != nil {
		impl.logger.Errorw("error in registering GitOps repository", "repoName", gitOpsRepoUrl, "err", argoCdErr)
		return impl.handleArgoRepoCreationError(ctx, argoCdErr, grpcConfig, gitOpsRepoUrl, targetRevision, userId)
	}
	impl.logger.Infow("gitOps repo registered in argo", "repoName", gitOpsRepoUrl)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) CreateRepoCreds(ctx context.Context, query *repocreds.RepoCredsCreateRequest) (*v1alpha1.RepoCreds, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, err
	}
	return impl.repoCredsClient.CreateRepoCreds(ctx, grpcConfig, query)
}

func (impl *ArgoClientWrapperServiceImpl) AddOrUpdateOCIRegistry(username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error {
	return impl.ArgoClientWrapperServiceEAImpl.AddOrUpdateOCIRegistry(username, password, uniqueId, registryUrl, repo, isPublic)
}

func (impl *ArgoClientWrapperServiceImpl) DeleteOCIRegistry(registryURL, repo string, ociRegistryId int) error {
	return impl.ArgoClientWrapperServiceEAImpl.DeleteOCIRegistry(registryURL, repo, ociRegistryId)
}

func (impl *ArgoClientWrapperServiceImpl) AddChartRepository(request bean2.ChartRepositoryAddRequest) error {
	return impl.ArgoClientWrapperServiceEAImpl.AddChartRepository(request)
}

func (impl *ArgoClientWrapperServiceImpl) UpdateChartRepository(request bean2.ChartRepositoryUpdateRequest) error {
	return impl.ArgoClientWrapperServiceEAImpl.UpdateChartRepository(request)
}

func (impl *ArgoClientWrapperServiceImpl) DeleteChartRepository(name, url string) error {
	return impl.ArgoClientWrapperServiceEAImpl.DeleteChartRepository(name, url)
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoAppByName(ctx context.Context, appName string) (*v1alpha1.Application, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, err
	}
	argoApplication, err := impl.acdApplicationClient.Get(ctx, grpcConfig, &application2.ApplicationQuery{Name: &appName})
	if err != nil {
		impl.logger.Errorw("err in getting argo app by name", "app", appName)
		return nil, err
	}
	return argoApplication, nil
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoAppByNameWithK8sClient(ctx context.Context, clusterId int, namespace, appName string) (*v1alpha1.Application, error) {
	k8sConfig, err := impl.acdConfigGetter.GetK8sConfigWithClusterIdAndNamespace(clusterId, namespace)
	if err != nil {
		impl.logger.Errorw("error in getting k8s config", "err", err)
		return nil, err
	}
	argoApplication, err := impl.argoK8sClient.GetArgoApplication(k8sConfig, appName)
	if err != nil {
		impl.logger.Errorw("err in getting argo app by name", "app", appName)
		return nil, err
	}
	application, err := GetAppObject(argoApplication)
	if err != nil {
		impl.logger.Errorw("error in getting app object", "deploymentAppName", appName, "err", err)
		return nil, err
	}
	return application, nil
}

func (impl *ArgoClientWrapperServiceImpl) DeleteArgoAppWithK8sClient(ctx context.Context, clusterId int, namespace, appName string, cascadeDelete bool) error {
	k8sConfig, err := impl.acdConfigGetter.GetK8sConfigWithClusterIdAndNamespace(clusterId, namespace)
	if err != nil {
		impl.logger.Errorw("error in getting k8s config", "err", err)
		return err
	}
	err = impl.argoK8sClient.DeleteArgoApplication(ctx, k8sConfig, appName, cascadeDelete)
	if err != nil {
		impl.logger.Errorw("err in getting argo app by name", "app", appName)
		return err
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) SyncArgoCDApplicationAndRefreshWithK8sClient(ctx context.Context, clusterId int, namespace, appName string) error {
	k8sConfig, err := impl.acdConfigGetter.GetK8sConfigWithClusterIdAndNamespace(clusterId, namespace)
	if err != nil {
		impl.logger.Errorw("error in getting k8s config", "err", err)
		return err
	}
	if impl.ACDConfig.IsManualSyncEnabled() {
		impl.logger.Debugw("syncing ArgoCd app using k8s as manual sync is enabled", "argoAppName", appName)
		err = impl.argoK8sClient.SyncArgoApplication(ctx, k8sConfig, appName)
		if err != nil {
			impl.logger.Errorw("err in syncing argo application", "app", appName, "err", err)
			return err
		}
	}
	runnableFunc := func() {
		impl.logger.Debugw("syncing ArgoCd app using k8s as manual sync is enabled", "argoAppName", appName)
		// running ArgoCd app refresh in asynchronous mode
		refreshErr := impl.argoK8sClient.RefreshApp(context.Background(), k8sConfig, appName, bean.RefreshTypeNormal)
		if refreshErr != nil {
			impl.logger.Errorw("error in refreshing argo app", "argoAppName", appName, "err", refreshErr)
		}
	}
	impl.asyncRunnable.Execute(runnableFunc)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) IsArgoAppPatchRequired(argoAppSpec *v1alpha1.ApplicationSource, currentGitRepoUrl, currentTargetRevision, currentChartPath string) bool {
	if argoAppSpec == nil {
		// if argo app spec is nil, then no need to patch
		// this means the argo app object is in corrupted state
		impl.logger.Warnw("received argo app spec is nil, skipping for patch request...")
		return false
	}
	return (len(currentGitRepoUrl) != 0 && argoAppSpec.RepoURL != currentGitRepoUrl) ||
		argoAppSpec.Path != currentChartPath ||
		argoAppSpec.TargetRevision != currentTargetRevision
}

func (impl *ArgoClientWrapperServiceImpl) PatchArgoCdApp(ctx context.Context, dto *bean.ArgoCdAppPatchReqDto) error {

	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return err
	}

	patchReq := adapter.GetArgoCdPatchReqFromDto(dto)
	reqbyte, err := json.Marshal(patchReq)
	if err != nil {
		impl.logger.Errorw("error in creating patch", "err", err)
		return err
	}
	reqString := string(reqbyte)
	_, err = impl.acdApplicationClient.Patch(ctx, grpcConfig, &application2.ApplicationPatchRequest{Patch: &reqString, Name: &dto.ArgoAppName, PatchType: &dto.PatchType})
	if err != nil {
		impl.logger.Errorw("error in patching argo pipeline ", "name", dto.ArgoAppName, "patch", reqString, "err", err)
		return err
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) GetGitOpsRepoNameForApplication(ctx context.Context, appName string) (gitOpsRepoName string, err error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return gitOpsRepoName, err
	}
	acdApplication, err := impl.acdApplicationClient.Get(ctx, grpcConfig, &application2.ApplicationQuery{Name: &appName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "acdAppName", appName, "err", err)
		return gitOpsRepoName, err
	}
	// safety checks nil pointers
	if acdApplication != nil && acdApplication.Spec.Source != nil {
		gitOpsRepoUrl := acdApplication.Spec.Source.RepoURL
		gitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(gitOpsRepoUrl)
		return gitOpsRepoName, nil
	}
	return gitOpsRepoName, fmt.Errorf("unable to get any ArgoCd application '%s'", appName)
}

func (impl *ArgoClientWrapperServiceImpl) GetGitOpsRepoURLForApplication(ctx context.Context, appName string) (gitOpsRepoName string, err error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return gitOpsRepoName, err
	}
	acdApplication, err := impl.acdApplicationClient.Get(ctx, grpcConfig, &application2.ApplicationQuery{Name: &appName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "acdAppName", appName, "err", err)
		return gitOpsRepoName, err
	}
	// safety checks nil pointers
	if acdApplication != nil && acdApplication.Spec.Source != nil {
		gitOpsRepoUrl := acdApplication.Spec.Source.RepoURL
		return gitOpsRepoUrl, nil
	}
	return "", fmt.Errorf("unable to get any ArgoCd application '%s'", appName)
}

// createRepoInArgoCd is the wrapper function to Create Repository in ArgoCd
func (impl *ArgoClientWrapperServiceImpl) createRepoInArgoCd(ctx context.Context, grpcConfig *bean.ArgoGRPCConfig, gitOpsRepoUrl string) error {
	repo := &v1alpha1.Repository{
		Repo: gitOpsRepoUrl,
	}
	repo, err := impl.repositoryService.Create(ctx, grpcConfig, &repository2.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository", "err", err)
		return err
	}
	return nil
}

// isRetryableArgoRepoCreationError returns whether to retry or not, based on the error returned from callback func
// In createRepoInArgoCd, we will retry only if the error matches to bean.ArgoRepoSyncDelayErr
func (impl *ArgoClientWrapperServiceImpl) isRetryableArgoRepoCreationError(argoCdErr error) bool {
	return strings.Contains(argoCdErr.Error(), bean.ArgoRepoSyncDelayErr)
}

// handleArgoRepoCreationError - manages the error thrown while performing createRepoInArgoCd
func (impl *ArgoClientWrapperServiceImpl) handleArgoRepoCreationError(ctx context.Context, argoCdErr error, grpcConfig *bean.ArgoGRPCConfig, gitOpsRepoUrl string, targetRevision string, userId int32) error {
	emptyRepoErrorMessages := bean.EmptyRepoErrorList
	isEmptyRepoError := false
	for _, errMsg := range emptyRepoErrorMessages {
		if strings.Contains(argoCdErr.Error(), errMsg) {
			isEmptyRepoError = true
		}
	}
	if isEmptyRepoError {
		// - found empty repository, create some file in repository
		gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(gitOpsRepoUrl)
		err := impl.gitOperationService.CreateReadmeInGitRepo(ctx, gitOpsRepoName, targetRevision, userId)
		if err != nil {
			impl.logger.Errorw("error in creating file in git repo", "err", err)
			return err
		}
	}
	// try to register with after creating readme file
	return impl.createRepoInArgoCd(ctx, grpcConfig, gitOpsRepoUrl)
}

func (impl *ArgoClientWrapperServiceImpl) CreateCluster(ctx context.Context, clusterRequest *cluster2.ClusterCreateRequest) (*v1alpha1.Cluster, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, err
	}
	cluster, err := impl.clusterClient.Create(ctx, grpcConfig, clusterRequest)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (impl *ArgoClientWrapperServiceImpl) UpdateCluster(ctx context.Context, clusterRequest *cluster2.ClusterUpdateRequest) (*v1alpha1.Cluster, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		impl.logger.Errorw("error in getting grpc config", "err", err)
		return nil, err
	}
	cluster, err := impl.clusterClient.Update(ctx, grpcConfig, clusterRequest)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (impl *ArgoClientWrapperServiceImpl) CreateCertificate(ctx context.Context, query *certificate.RepositoryCertificateCreateRequest) (*v1alpha1.RepositoryCertificateList, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		return nil, err
	}
	return impl.CertificateClient.CreateCertificate(ctx, grpcConfig, query)
}

func (impl *ArgoClientWrapperServiceImpl) DeleteCertificate(ctx context.Context, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error) {
	grpcConfig, err := impl.acdConfigGetter.GetGRPCConfig()
	if err != nil {
		return nil, err
	}
	return impl.CertificateClient.DeleteCertificate(ctx, grpcConfig, query, opts...)
}
