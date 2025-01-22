package argocdServer

import (
	"context"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/certificate"
	cluster2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	config2 "github.com/devtron-labs/devtron/client/argocdServer/config"
	"github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient"
	bean2 "github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient/bean"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ArgoClientWrapperServiceEAImpl struct {
	logger             *zap.SugaredLogger
	repoCredsK8sClient repoCredsK8sClient.RepositoryCredsK8sClient
	acdConfigGetter    config2.ArgoCDConfigGetter
}

func NewArgoClientWrapperServiceEAImpl(
	logger *zap.SugaredLogger,
	repoCredsK8sClient repoCredsK8sClient.RepositoryCredsK8sClient,
	acdConfigGetter config2.ArgoCDConfigGetter,
) *ArgoClientWrapperServiceEAImpl {
	return &ArgoClientWrapperServiceEAImpl{
		logger:             logger,
		repoCredsK8sClient: repoCredsK8sClient,
		acdConfigGetter:    acdConfigGetter,
	}
}

func (impl *ArgoClientWrapperServiceEAImpl) ResourceTree(ctxt context.Context, query *application2.ResourcesQuery) (*v1alpha1.ApplicationTree, error) {
	impl.logger.Info("not implemented")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) GetArgoClient(ctxt context.Context) (application2.ApplicationServiceClient, *grpc.ClientConn, error) {
	impl.logger.Info("not implemented")
	return nil, nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) GetApplicationResource(ctx context.Context, query *application2.ApplicationResourceRequest) (*application2.ApplicationResourceResponse, error) {
	impl.logger.Info("not implemented")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) GetArgoApplication(ctx context.Context, query *application2.ApplicationQuery) (*v1alpha1.Application, error) {
	impl.logger.Info("not implemented")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) DeleteArgoApp(ctx context.Context, appName string, cascadeDelete bool) (*application2.ApplicationResponse, error) {
	impl.logger.Info("not implemented")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) SyncArgoCDApplicationIfNeededAndRefresh(ctx context.Context, argoAppName string) error {
	impl.logger.Info("not implemented")
	return nil
}

func (impl *ArgoClientWrapperServiceEAImpl) UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error) {
	impl.logger.Info("not implemented")
	return nil
}

func (impl *ArgoClientWrapperServiceEAImpl) RegisterGitOpsRepoInArgoWithRetry(ctx context.Context, gitOpsRepoUrl string, userId int32) error {
	impl.logger.Info("not implemented")
	return nil
}

func (impl *ArgoClientWrapperServiceEAImpl) CreateRepoCreds(ctx context.Context, query *repocreds.RepoCredsCreateRequest) (*v1alpha1.RepoCreds, error) {
	impl.logger.Info("not implemented")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) AddOrUpdateOCIRegistry(username, password string, uniqueId int, registryUrl, repo string, isPublic bool) error {
	argoK8sConfig, err := impl.acdConfigGetter.GetK8sConfig()
	if err != nil {
		impl.logger.Errorw("error in getting k8s config", "err", err)
	}
	return impl.repoCredsK8sClient.AddOrUpdateOCIRegistry(argoK8sConfig, username, password, uniqueId, registryUrl, repo, isPublic)
}

func (impl *ArgoClientWrapperServiceEAImpl) DeleteOCIRegistry(registryURL, repo string, ociRegistryId int) error {
	argoK8sConfig, err := impl.acdConfigGetter.GetK8sConfig()
	if err != nil {
		impl.logger.Errorw("error in getting k8s config", "err", err)
	}
	return impl.repoCredsK8sClient.DeleteOCIRegistry(argoK8sConfig, registryURL, repo, ociRegistryId)
}

func (impl *ArgoClientWrapperServiceEAImpl) AddChartRepository(request bean2.ChartRepositoryAddRequest) error {
	argoK8sConfig, err := impl.acdConfigGetter.GetK8sConfig()
	if err != nil {
		impl.logger.Errorw("error in getting k8s config", "err", err)
	}
	return impl.repoCredsK8sClient.AddChartRepository(argoK8sConfig, request)
}

func (impl *ArgoClientWrapperServiceEAImpl) UpdateChartRepository(request bean2.ChartRepositoryUpdateRequest) error {
	argoK8sConfig, err := impl.acdConfigGetter.GetK8sConfig()
	if err != nil {
		return err
	}
	return impl.repoCredsK8sClient.UpdateChartRepository(argoK8sConfig, request)
}

func (impl *ArgoClientWrapperServiceEAImpl) DeleteChartRepository(name, url string) error {
	argoK8sConfig, err := impl.acdConfigGetter.GetK8sConfig()
	if err != nil {
		return err
	}
	return impl.repoCredsK8sClient.DeleteChartRepository(argoK8sConfig, name, url)
}

func (impl *ArgoClientWrapperServiceEAImpl) GetArgoAppByName(ctx context.Context, appName string) (*v1alpha1.Application, error) {
	impl.logger.Info("not implemented for EA mode")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) GetArgoAppByNameWithK8s(ctx context.Context, clusterId int, namespace, appName string) (map[string]interface{}, error) {
	impl.logger.Info("not implemented for EA mode")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) IsArgoAppPatchRequired(argoAppSpec *v1alpha1.ApplicationSource, currentGitRepoUrl, currentChartPath string) bool {
	impl.logger.Info("not implemented for EA mode")
	return false
}

func (impl *ArgoClientWrapperServiceEAImpl) PatchArgoCdApp(ctx context.Context, dto *bean.ArgoCdAppPatchReqDto) error {
	impl.logger.Info("not implemented for EA mode")
	return nil
}

func (impl *ArgoClientWrapperServiceEAImpl) GetGitOpsRepoNameForApplication(ctx context.Context, appName string) (gitOpsRepoName string, err error) {
	impl.logger.Info("not implemented for EA mode")
	return "", nil
}

func (impl *ArgoClientWrapperServiceEAImpl) GetGitOpsRepoURLForApplication(ctx context.Context, appName string) (gitOpsRepoName string, err error) {
	impl.logger.Info("not implemented for EA mode")
	return "", nil
}

func (impl *ArgoClientWrapperServiceEAImpl) CreateCluster(ctx context.Context, clusterRequest *cluster2.ClusterCreateRequest) (*v1alpha1.Cluster, error) {
	impl.logger.Info("not implemented for EA mode")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) UpdateCluster(ctx context.Context, clusterRequest *cluster2.ClusterUpdateRequest) (*v1alpha1.Cluster, error) {
	impl.logger.Info("not implemented for EA mode")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) CreateCertificate(ctx context.Context, query *certificate.RepositoryCertificateCreateRequest) (*v1alpha1.RepositoryCertificateList, error) {
	impl.logger.Info("not implemented for EA mode")
	return nil, nil
}

func (impl *ArgoClientWrapperServiceEAImpl) DeleteCertificate(ctx context.Context, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error) {
	impl.logger.Info("not implemented for EA mode")
	return nil, nil
}
