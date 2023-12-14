package argocdServer

import (
	"context"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
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

	//GetArgoAppWithNormalRefresh - refresh app at argocd side
	GetArgoAppWithNormalRefresh(context context.Context, argoAppName string) error

	//SyncArgoCDApplicationIfNeededAndRefresh - if ARGO_AUTO_SYNC_ENABLED=true, app will be refreshed to initiate refresh at argoCD side or else it will be synced and refreshed
	SyncArgoCDApplicationIfNeededAndRefresh(context context.Context, argoAppName string) error

	// UpdateArgoCDSyncModeIfNeeded - if ARGO_AUTO_SYNC_ENABLED=true and app is in manual sync mode or vice versa update app
	UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error)
}

type ArgoClientWrapperServiceImpl struct {
	logger    *zap.SugaredLogger
	acdClient application.ServiceClient
	ACDConfig *ACDConfig
}

func NewArgoClientWrapperServiceImpl(logger *zap.SugaredLogger,
	acdClient application.ServiceClient,
	ACDConfig *ACDConfig,
) *ArgoClientWrapperServiceImpl {
	return &ArgoClientWrapperServiceImpl{
		logger:    logger,
		acdClient: acdClient,
		ACDConfig: ACDConfig,
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
		revision := "master"
		pruneResources := true
		_, syncErr := impl.acdClient.Sync(context, &application2.ApplicationSyncRequest{Name: &argoAppName, Revision: &revision, Prune: &pruneResources})
		if syncErr != nil {
			impl.logger.Errorw("cannot get application with refresh", "app", argoAppName)
			return syncErr
		}
	}
	refreshErr := impl.GetArgoAppWithNormalRefresh(context, argoAppName)
	if refreshErr != nil {
		impl.logger.Errorw("error in refreshing argo app", "err", refreshErr)
	}
	impl.logger.Debugw("argo app sync completed", "argoAppName", argoAppName)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) UpdateArgoCDSyncModeIfNeeded(ctx context.Context, argoApplication *v1alpha1.Application) (err error) {
	if impl.IsArgoAppSyncModeMigrationNeeded(argoApplication) {
		syncModeUpdateRequest := impl.GetArgocdAppSyncModeUpdateRequest(argoApplication)
		validate := false
		_, err = impl.acdClient.Update(ctx, &application2.ApplicationUpdateRequest{Application: syncModeUpdateRequest, Validate: &validate})
		if err != nil {
			impl.logger.Errorw("error in creating argo pipeline ", "name", argoApplication.Name, "err", err)
			return err
		}
	}
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) IsArgoAppSyncModeMigrationNeeded(argoApplication *v1alpha1.Application) bool {
	if !impl.ACDConfig.ArgoCDAutoSyncEnabled && argoApplication.Spec.SyncPolicy.Automated != nil {
		return true
	}
	if impl.ACDConfig.ArgoCDAutoSyncEnabled && argoApplication.Spec.SyncPolicy.Automated == nil {
		return true
	}
	return false
}

func (impl *ArgoClientWrapperServiceImpl) GetArgocdAppSyncModeUpdateRequest(argoApplication *v1alpha1.Application) *v1alpha1.Application {
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
