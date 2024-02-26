package argocdServer

import (
	"context"
	"encoding/json"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	repository2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/client/argocdServer/adapter"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// GetArgoAppWithNormalRefresh - refresh app at argocd side
	GetArgoAppWithNormalRefresh(context context.Context, argoAppName string) error

	// SyncArgoCDApplicationIfNeededAndRefresh - if ARGO_AUTO_SYNC_ENABLED=true, app will be refreshed to initiate refresh at argoCD side or else it will be synced and refreshed
	SyncArgoCDApplicationIfNeededAndRefresh(context context.Context, argoAppName string) error

	// UpdateArgoCDSyncModeIfNeeded - if ARGO_AUTO_SYNC_ENABLED=true and app is in manual sync mode or vice versa update app
	UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error)

	// RegisterGitOpsRepoInArgo - register a repository in argo-cd
	RegisterGitOpsRepoInArgo(ctx context.Context, repoUrl string) (err error)

	// GetArgoAppByName fetches an argoCd app by its name
	GetArgoAppByName(ctx context.Context, appName string) (*v1alpha1.Application, error)

	// PatchArgoCdApp performs a patch operation on an argoCd app
	PatchArgoCdApp(ctx context.Context, dto *bean.ArgoCdAppPatchReqDto) error
}

type ArgoClientWrapperServiceImpl struct {
	logger            *zap.SugaredLogger
	acdClient         application.ServiceClient
	ACDConfig         *ACDConfig
	repositoryService repository.ServiceClient
}

func NewArgoClientWrapperServiceImpl(logger *zap.SugaredLogger, acdClient application.ServiceClient,
	ACDConfig *ACDConfig, repositoryService repository.ServiceClient) *ArgoClientWrapperServiceImpl {
	return &ArgoClientWrapperServiceImpl{
		logger:            logger,
		acdClient:         acdClient,
		ACDConfig:         ACDConfig,
		repositoryService: repositoryService,
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

func (impl *ArgoClientWrapperServiceImpl) RegisterGitOpsRepoInArgo(ctx context.Context, repoUrl string) (err error) {
	repo := &v1alpha1.Repository{
		Repo: repoUrl,
	}
	repo, err = impl.repositoryService.Create(ctx, &repository2.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository ", "err", err)
	}
	impl.logger.Infow("gitOps repo registered in argo", "name", repoUrl)
	return err
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoAppByName(ctx context.Context, appName string) (*v1alpha1.Application, error) {
	argoApplication, err := impl.acdClient.Get(ctx, &application2.ApplicationQuery{Name: &appName})
	if err != nil {
		impl.logger.Errorw("err in getting argo app by name", "app", appName)
		return nil, err
	}
	return argoApplication, nil
}

func (impl *ArgoClientWrapperServiceImpl) PatchArgoCdApp(ctx context.Context, dto *bean.ArgoCdAppPatchReqDto) error {
	patchReq := adapter.GetArgoCdPatchReqFromDto(dto)
	reqbyte, err := json.Marshal(patchReq)
	if err != nil {
		impl.logger.Errorw("error in creating patch", "err", err)
		return err
	}
	reqString := string(reqbyte)
	_, err = impl.acdClient.Patch(ctx, &application2.ApplicationPatchRequest{Patch: &reqString, Name: &dto.ArgoAppName, PatchType: &dto.PatchType})
	if err != nil {
		impl.logger.Errorw("error in patching argo pipeline ", "name", dto.ArgoAppName, "patch", reqString, "err", err)
		return err
	}
	return nil
}
