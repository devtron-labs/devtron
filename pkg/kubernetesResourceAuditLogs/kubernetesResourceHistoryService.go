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

package kubernetesResourceAuditLogs

import (
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/helm-app/service/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"time"
)

const (
	delete string = "delete"
	helm   string = "helm"
	GitOps string = "argo_cd"
)

type K8sResourceHistoryService interface {
	SaveArgoCdAppsResourceDeleteHistory(query *application.ApplicationResourceDeleteRequest, appId int, envId int, userId int32) error
	SaveHelmAppsResourceHistory(appIdentifier *bean.AppIdentifier, k8sRequestBean *k8s.K8sRequestBean, userId int32, actionType string) error
}

type K8sResourceHistoryServiceImpl struct {
	appRepository                app.AppRepository
	K8sResourceHistoryRepository repository.K8sResourceHistoryRepository
	logger                       *zap.SugaredLogger
	envRepository                repository2.EnvironmentRepository
}

func Newk8sResourceHistoryServiceImpl(K8sResourceHistoryRepository repository.K8sResourceHistoryRepository,
	logger *zap.SugaredLogger, appRepository app.AppRepository, envRepository repository2.EnvironmentRepository) *K8sResourceHistoryServiceImpl {
	return &K8sResourceHistoryServiceImpl{
		K8sResourceHistoryRepository: K8sResourceHistoryRepository,
		logger:                       logger,
		appRepository:                appRepository,
		envRepository:                envRepository,
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
		ActionType:        delete,
		DeploymentAppType: GitOps,
	}

	err := impl.K8sResourceHistoryRepository.SaveK8sResourceHistory(&k8sResourceHistory)

	if err != nil {
		return err
	}

	return nil

}

func (impl K8sResourceHistoryServiceImpl) SaveHelmAppsResourceHistory(appIdentifier *bean.AppIdentifier, k8sRequestBean *k8s.K8sRequestBean, userId int32, actionType string) error {

	app, err := impl.appRepository.FindActiveByName(appIdentifier.ReleaseName)

	env, err := impl.envRepository.FindOneByNamespaceAndClusterId(appIdentifier.Namespace, appIdentifier.ClusterId)

	k8sResourceHistory := repository.K8sResourceHistory{
		AppId:        app.Id,
		AppName:      appIdentifier.ReleaseName,
		EnvId:        env.Id,
		Namespace:    appIdentifier.Namespace,
		ResourceName: k8sRequestBean.ResourceIdentifier.Name,
		Kind:         k8sRequestBean.ResourceIdentifier.GroupVersionKind.Kind,
		Group:        k8sRequestBean.ResourceIdentifier.GroupVersionKind.Group,
		ForceDelete:  false,
		AuditLog: sql.AuditLog{
			UpdatedBy: userId,
			UpdatedOn: time.Now(),
		},
		ActionType:        actionType,
		DeploymentAppType: helm,
	}

	err = impl.K8sResourceHistoryRepository.SaveK8sResourceHistory(&k8sResourceHistory)

	return err

}
