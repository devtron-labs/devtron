package argocdServer

import (
	"context"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	repository2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type ACDConfig struct {
	ArgoCDAutoSyncEnabled bool `env:"ARGO_AUTO_SYNC_ENABLED" envDefault:"true"` //will gradually switch this flag to false in enterprise
}

func GetACDDeploymentConfig() (*ACDConfig, error) {
	cfg := &ACDConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

type ArgoClientWrapperService interface {

	//GetArgoAppWithNormalRefresh - refresh app at argocd side
	GetArgoAppWithNormalRefresh(context context.Context, argoAppName string) error

	//SyncArgoCDApplicationIfNeededAndRefresh - if ARGO_AUTO_SYNC_ENABLED=true, app will be refreshed to initiate refresh at argoCD side or else it will be synced and refreshed
	SyncArgoCDApplicationIfNeededAndRefresh(context context.Context, argoAppName string) error

	// UpdateArgoCDSyncModeIfNeeded - if ARGO_AUTO_SYNC_ENABLED=true and app is in manual sync mode or vice versa update app
	UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error)

	//RegisterGitOpsRepoInArgo - register a repository in argo-cd with retry mechanism
	RegisterGitOpsRepoInArgo(ctx context.Context, gitOpsRepoUrl string, userId int32) error
}

type ArgoClientWrapperServiceImpl struct {
	logger                  *zap.SugaredLogger
	acdClient               application.ServiceClient
	ACDConfig               *ACDConfig
	repositoryService       repository.ServiceClient
	gitOpsConfigReadService config.GitOpsConfigReadService
	gitOperationService     git.GitOperationService
}

func NewArgoClientWrapperServiceImpl(logger *zap.SugaredLogger, acdClient application.ServiceClient,
	ACDConfig *ACDConfig, repositoryService repository.ServiceClient, gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOperationService git.GitOperationService) *ArgoClientWrapperServiceImpl {
	return &ArgoClientWrapperServiceImpl{
		logger:                  logger,
		acdClient:               acdClient,
		ACDConfig:               ACDConfig,
		repositoryService:       repositoryService,
		gitOpsConfigReadService: gitOpsConfigReadService,
		gitOperationService:     gitOperationService,
	}
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoAppWithNormalRefresh(context context.Context, argoAppName string) error {
	refreshType := bean.RefreshTypeNormal
	impl.logger.Debugw("trying to normal refresh application through get ", "argoAppName", argoAppName)
	_, err := impl.acdClient.Get(context, &application2.ApplicationQuery{Name: &argoAppName, Refresh: &refreshType})
	if err != nil {
		impl.logger.Errorw("cannot get application with refresh", "app", argoAppName)
		return err
	}
	impl.logger.Debugw("done getting the application with refresh with no error", "argoAppName", argoAppName)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) SyncArgoCDApplicationIfNeededAndRefresh(context context.Context, argoAppName string) error {
	impl.logger.Info("argocd manual sync for app started", "argoAppName", argoAppName)
	if !impl.ACDConfig.ArgoCDAutoSyncEnabled {
		impl.logger.Debugw("syncing argocd app as manual sync is enabled", "argoAppName", argoAppName)
		revision := "master"
		pruneResources := true
		_, syncErr := impl.acdClient.Sync(context, &application2.ApplicationSyncRequest{Name: &argoAppName, Revision: &revision, Prune: &pruneResources})
		if syncErr != nil {
			impl.logger.Errorw("cannot get application with refresh", "app", argoAppName)
			return syncErr
		}
		impl.logger.Debugw("argocd sync completed", "argoAppName", argoAppName)
	}
	refreshErr := impl.GetArgoAppWithNormalRefresh(context, argoAppName)
	if refreshErr != nil {
		impl.logger.Errorw("error in refreshing argo app", "err", refreshErr)
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error) {
	if impl.isArgoAppSyncModeMigrationNeeded(argoApplication) {
		syncModeUpdateRequest := impl.CreateRequestForArgoCDSyncModeUpdateRequest(argoApplication)
		validate := false
		_, err = impl.acdClient.Update(ctx, &application2.ApplicationUpdateRequest{Application: syncModeUpdateRequest, Validate: &validate})
		if err != nil {
			impl.logger.Errorw("error in creating argo pipeline ", "name", argoApplication.Name, "err", err)
			return err
		}
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) isArgoAppSyncModeMigrationNeeded(argoApplication *v1alpha1.Application) bool {
	if !impl.ACDConfig.ArgoCDAutoSyncEnabled && argoApplication.Spec.SyncPolicy.Automated != nil {
		return true
	}
	if impl.ACDConfig.ArgoCDAutoSyncEnabled && argoApplication.Spec.SyncPolicy.Automated == nil {
		return true
	}
	return false
}

func (impl *ArgoClientWrapperServiceImpl) CreateRequestForArgoCDSyncModeUpdateRequest(argoApplication *v1alpha1.Application) *v1alpha1.Application {
	// set automated field in update request
	var automated *v1alpha1.SyncPolicyAutomated
	if impl.ACDConfig.ArgoCDAutoSyncEnabled {
		automated = &v1alpha1.SyncPolicyAutomated{
			Prune: true,
		}
	}
	return &v1alpha1.Application{
		ObjectMeta: v1.ObjectMeta{
			Name:      argoApplication.Name,
			Namespace: DevtronInstalationNs,
		},
		Spec: v1alpha1.ApplicationSpec{
			Destination: argoApplication.Spec.Destination,
			Source:      argoApplication.Spec.Source,
			SyncPolicy: &v1alpha1.SyncPolicy{
				Automated:   automated,
				SyncOptions: argoApplication.Spec.SyncPolicy.SyncOptions,
				Retry:       argoApplication.Spec.SyncPolicy.Retry,
			}}}
}

func (impl *ArgoClientWrapperServiceImpl) RegisterGitOpsRepoInArgo(ctx context.Context, gitOpsRepoUrl string, userId int32) error {

	retryCount := 0
	// label to register git repository in ArgoCd
	// ArgoCd requires approx 80 to 120 sec after the last commit to allow create-repository action
	// hence this operation needed to be perform with retry
CreateArgoRepositoryWithRetry:
	repo := &v1alpha1.Repository{
		Repo: gitOpsRepoUrl,
	}
	repo, err := impl.repositoryService.Create(ctx, &repository2.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository", "err", err)
		retryCount, err = impl.handleArgoRepoCreationError(retryCount, gitOpsRepoUrl, userId, err)
		if err != nil {
			impl.logger.Errorw("error in RegisterGitOpsRepoInArgo with retry operation", "err", err)
			return err
		}
		impl.logger.Errorw("retrying RegisterGitOpsRepoInArgo operation", "retry count", retryCount)
		time.Sleep(10 * time.Second)
		goto CreateArgoRepositoryWithRetry
	}
	impl.logger.Infow("gitOps repo registered in argo", "name", gitOpsRepoUrl)
	return err
}

func (impl *ArgoClientWrapperServiceImpl) handleArgoRepoCreationError(retryCount int, repoUrl string, userId int32, argoCdErr error) (int, error) {
	// retry limit exhausted
	if retryCount >= bean.RegisterRepoMaxRetryCount {
		return 0, argoCdErr
	}
	// This error occurs inconsistently; ArgoCD requires 80-120s after last commit for create repository operation
	argoRepoSyncDelayErrMessage := "Unable to resolve 'HEAD' to a commit SHA"
	isSyncDelayError := strings.Contains(argoCdErr.Error(), argoRepoSyncDelayErrMessage)

	// ArgoCD can't register empty repo and throws these error message in such cases
	emptyRepoErrorMessages := []string{"failed to get index: 404 Not Found", "remote repository is empty"}
	isEmptyRepoError := false
	for _, errMsg := range emptyRepoErrorMessages {
		if strings.Contains(argoCdErr.Error(), errMsg) {
			isEmptyRepoError = true
		}
	}
	// unknown error handling
	if !isSyncDelayError && !isEmptyRepoError {
		return 0, argoCdErr
	}
	if isEmptyRepoError {
		// - found empty repository, create some file in repository
		gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(repoUrl)
		err := impl.gitOperationService.CreateReadmeInGitRepo(gitOpsRepoName, userId)
		if err != nil {
			impl.logger.Errorw("error in creating file in git repo", "err", err)
			return 0, err
		}
	}
	return retryCount + 1, nil
}
