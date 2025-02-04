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

package argoApplication

import (
	"context"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/helper"
	"github.com/devtron-labs/devtron/pkg/argoApplication/read/config"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common/read"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s/bean"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"slices"
)

type ArgoApplicationService interface {
	ListApplications(clusterIds []int) ([]*bean.ArgoApplicationListDto, error)
	HibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)

	//FUll mode
	// ResourceTree	returns the status for all Apps deployed via ArgoCd
	ResourceTree(ctx context.Context, query *application2.ResourcesQuery) (*argoApplication.ResourceTreeResponse, error)
}

type ArgoApplicationServiceImpl struct {
	logger                       *zap.SugaredLogger
	clusterRepository            clusterRepository.ClusterRepository
	k8sUtil                      *k8s.K8sServiceImpl
	helmAppClient                gRPC.HelmAppClient
	helmAppService               service.HelmAppService
	k8sApplicationService        application.K8sApplicationService
	argoApplicationConfigService config.ArgoApplicationConfigService
	deploymentConfigReadService  read.DeploymentConfigReadService
}

func NewArgoApplicationServiceImpl(logger *zap.SugaredLogger,
	clusterRepository clusterRepository.ClusterRepository,
	k8sUtil *k8s.K8sServiceImpl,
	helmAppClient gRPC.HelmAppClient,
	helmAppService service.HelmAppService,
	k8sApplicationService application.K8sApplicationService,
	argoApplicationConfigService config.ArgoApplicationConfigService,
	deploymentConfigReadService read.DeploymentConfigReadService) *ArgoApplicationServiceImpl {
	return &ArgoApplicationServiceImpl{
		logger:                       logger,
		clusterRepository:            clusterRepository,
		k8sUtil:                      k8sUtil,
		helmAppService:               helmAppService,
		helmAppClient:                helmAppClient,
		k8sApplicationService:        k8sApplicationService,
		argoApplicationConfigService: argoApplicationConfigService,
		deploymentConfigReadService:  deploymentConfigReadService,
	}

}

func (impl *ArgoApplicationServiceImpl) ListApplications(clusterIds []int) ([]*bean.ArgoApplicationListDto, error) {
	var clusters []clusterRepository.Cluster
	var err error
	if len(clusterIds) > 0 {
		// getting cluster details by ids
		clusters, err = impl.clusterRepository.FindByIds(clusterIds)
		if err != nil {
			impl.logger.Errorw("error in getting clusters by ids", "err", err, "clusterIds", clusterIds)
			return nil, err
		}
	} else {
		clusters, err = impl.clusterRepository.FindAllActive()
		if err != nil {
			impl.logger.Errorw("error in getting all active clusters", "err", err)
			return nil, err
		}
	}

	listReq := &k8s2.ResourceRequestBean{
		K8sRequest: &k8s.K8sRequestBean{
			ResourceIdentifier: k8s.ResourceIdentifier{
				Namespace:        bean.AllNamespaces,
				GroupVersionKind: bean.GvkForArgoApplication,
			},
		},
	}
	// TODO: make goroutine and channel for optimization
	appListFinal := make([]*bean.ArgoApplicationListDto, 0)
	for _, cluster := range clusters {
		clusterObj := cluster
		if clusterObj.IsVirtualCluster || len(clusterObj.ErrorInConnecting) != 0 {
			continue
		}
		clusterBean := adapter.GetClusterBean(clusterObj)
		clusterConfig := clusterBean.GetClusterConfig()
		restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
		if err != nil {
			impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterObj.Id)
			return nil, err
		}
		resp, err := impl.k8sApplicationService.GetResourceListWithRestConfig(context.Background(), "", listReq, nil, restConfig, clusterObj.ClusterName)
		if err != nil {
			if errStatus, ok := err.(*errors.StatusError); ok {
				if errStatus.Status().Code == 404 {
					// no argo apps found, not sending error
					impl.logger.Warnw("error in getting external argo app list, no apps found", "err", err, "clusterId", clusterObj.Id)
					continue
				}
			}
			impl.logger.Errorw("error in getting resource list", "err", err)
			return nil, err
		}
		appLists := getApplicationListDtos(resp, clusterObj.ClusterName, clusterObj.Id)
		appListFinal = append(appListFinal, appLists...)
	}
	appListClusterIds := sliceUtil.NewSliceFromFuncExec(appListFinal, func(app *bean.ArgoApplicationListDto) int {
		return app.ClusterId
	})
	allDevtronManagedArgoAppNames, err := impl.deploymentConfigReadService.GetAllArgoAppNamesByCluster(appListClusterIds)
	if err != nil {
		impl.logger.Errorw("error in getting all argo app names by cluster", "err", err, "clusterIds", appListClusterIds)
		return nil, err
	}
	filteredAppList := make([]*bean.ArgoApplicationListDto, 0)
	filteredAppList = sliceUtil.Filter(filteredAppList, appListFinal, func(app *bean.ArgoApplicationListDto) bool {
		return !slices.Contains(allDevtronManagedArgoAppNames, app.Name)
	})
	return filteredAppList, nil
}

func getApplicationListDtos(resp *k8s.ClusterResourceListMap, clusterName string, clusterId int) []*bean.ArgoApplicationListDto {
	appLists := make([]*bean.ArgoApplicationListDto, 0)
	if resp != nil {
		appLists = make([]*bean.ArgoApplicationListDto, len(resp.Data))
		for i, rowData := range resp.Data {
			if rowData == nil {
				continue
			}
			appListDto := &bean.ArgoApplicationListDto{
				ClusterId:   clusterId,
				ClusterName: clusterName,
			}
			if rowData[k8sCommonBean.K8sClusterResourceNameKey] != nil {
				if nameStr, ok := rowData[k8sCommonBean.K8sClusterResourceNameKey].(string); ok {
					appListDto.Name = nameStr
				}
			}
			if rowData[k8sCommonBean.K8sResourceColumnDefinitionSyncStatus] != nil {
				if syncStatusStr, ok := rowData[k8sCommonBean.K8sResourceColumnDefinitionSyncStatus].(string); ok {
					appListDto.SyncStatus = syncStatusStr
				}
			}
			if rowData[k8sCommonBean.K8sResourceColumnDefinitionHealthStatus] != nil {
				if healthStatusStr, ok := rowData[k8sCommonBean.K8sResourceColumnDefinitionHealthStatus].(string); ok {
					appListDto.HealthStatus = healthStatusStr
				}
			}
			if rowData[k8sCommonBean.K8sClusterResourceNamespaceKey] != nil {
				if namespaceStr, ok := rowData[k8sCommonBean.K8sClusterResourceNamespaceKey].(string); ok {
					appListDto.Namespace = namespaceStr
				}
			}
			appLists[i] = appListDto
		}
	}
	return appLists
}

func (impl *ArgoApplicationServiceImpl) HibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	_, clusterBean, _, err := impl.argoApplicationConfigService.GetClusterConfigFromAllClusters(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("HibernateArgoApplication", "error in getting the cluster config", err, "clusterId", app.ClusterId, "appName", app.AppName)
		return nil, err
	}
	conf := helper.ConvertClusterBeanToGrpcConfig(clusterBean)

	req := service.HibernateReqAdaptor(hibernateRequest)
	req.ClusterConfig = conf
	res, err := impl.helmAppClient.Hibernate(ctx, req)
	if err != nil {
		impl.logger.Errorw("HibernateArgoApplication", "error in hibernating the requested resource", err, "clusterId", app.ClusterId, "appName", app.AppName)
		return nil, err
	}
	response := service.HibernateResponseAdaptor(res.Status)
	return response, nil
}

func (impl *ArgoApplicationServiceImpl) UnHibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	_, clusterBean, _, err := impl.argoApplicationConfigService.GetClusterConfigFromAllClusters(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("HibernateArgoApplication", "error in getting the cluster config", err, "clusterId", app.ClusterId, "appName", app.AppName)
		return nil, err
	}
	conf := helper.ConvertClusterBeanToGrpcConfig(clusterBean)

	req := service.HibernateReqAdaptor(hibernateRequest)
	req.ClusterConfig = conf
	res, err := impl.helmAppClient.UnHibernate(ctx, req)
	if err != nil {
		impl.logger.Errorw("UnHibernateArgoApplication", "error in unHibernating the requested resources", err, "clusterId", app.ClusterId, "appName", app.AppName)
		return nil, err
	}
	response := service.HibernateResponseAdaptor(res.Status)
	return response, nil
}

func (impl *ArgoApplicationServiceImpl) ResourceTree(ctx context.Context, query *application2.ResourcesQuery) (*argoApplication.ResourceTreeResponse, error) {
	return nil, util2.DefaultApiError().WithHttpStatusCode(http.StatusNotFound).WithInternalMessage(util.NotSupportedErr).WithUserMessage(util.NotSupportedErr)
}
