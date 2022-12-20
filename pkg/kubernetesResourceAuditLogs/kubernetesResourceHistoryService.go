package kubernetesResourceAuditLogs

import (
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	client "github.com/devtron-labs/devtron/api/helm-app"
	application2 "github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"time"
)

type K8sResourceHistoryService interface {
	SaveArgoCdAppsResourceDeleteHistory(query *application.ApplicationResourceDeleteRequest, appId int, envId int, userId int32) error
	SaveHelmAppsResourceHistory(appIdentifier *client.AppIdentifier, k8sRequestBean *application2.K8sRequestBean, userId int32, actionType string) error
}

type K8sResourceHistoryServiceImpl struct {
	K8sResourceHistoryRepository repository.K8sResourceHistoryRepository
	logger                       *zap.SugaredLogger
}

func Newk8sResourceHistoryServiceImpl(K8sResourceHistoryRepository repository.K8sResourceHistoryRepository,
	logger *zap.SugaredLogger) *K8sResourceHistoryServiceImpl {
	return &K8sResourceHistoryServiceImpl{
		K8sResourceHistoryRepository: K8sResourceHistoryRepository,
		logger:                       logger,
	}
}

func (impl K8sResourceHistoryServiceImpl) SaveArgoCdAppsResourceDeleteHistory(query *application.ApplicationResourceDeleteRequest, appId int, envId int, userId int32) error {

	k8sResourceHistory := repository.K8sResourceHistory{
		AppId:        appId,
		AppName:      *query.Name,
		EnvId:        envId,
		Namespace:    *query.Namespace,
		ResourceName: *query.ResourceName,
		Kind:         *query.Kind,
		Group:        *query.Group,
		ForceDelete:  *query.Force,
		AuditLog: sql.AuditLog{
			UpdatedBy: userId,
			UpdatedOn: time.Now(),
		},
		ActionType: "delete",
	}

	err := impl.K8sResourceHistoryRepository.SaveK8sResourceHistory(&k8sResourceHistory)

	if err != nil {
		return err
	}

	return nil

}

func (impl K8sResourceHistoryServiceImpl) SaveHelmAppsResourceHistory(appIdentifier *client.AppIdentifier, k8sRequestBean *application2.K8sRequestBean, userId int32, actionType string) error {

	k8sResourceHistory := repository.K8sResourceHistory{
		AppId:        0,
		AppName:      appIdentifier.ReleaseName,
		EnvId:        0,
		Namespace:    appIdentifier.Namespace,
		ResourceName: k8sRequestBean.ResourceIdentifier.Name,
		Kind:         k8sRequestBean.ResourceIdentifier.GroupVersionKind.Kind,
		Group:        k8sRequestBean.ResourceIdentifier.GroupVersionKind.Group,
		ForceDelete:  false,
		AuditLog: sql.AuditLog{
			UpdatedBy: userId,
			UpdatedOn: time.Now(),
		},
		ActionType: actionType,
	}

	err := impl.K8sResourceHistoryRepository.SaveK8sResourceHistory(&k8sResourceHistory)

	return err

}
