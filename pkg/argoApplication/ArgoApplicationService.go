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
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/helper"
	"github.com/devtron-labs/devtron/pkg/argoApplication/read"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/util/argo"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
)

type ArgoApplicationService interface {
	ListApplications(clusterIds []int) ([]*bean.ArgoApplicationListDto, error)
	HibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
}

type ArgoApplicationServiceImpl struct {
	logger                *zap.SugaredLogger
	clusterRepository     clusterRepository.ClusterRepository
	k8sUtil               *k8s.K8sServiceImpl
	argoUserService       argo.ArgoUserService
	helmAppClient         gRPC.HelmAppClient
	helmAppService        service.HelmAppService
	k8sApplicationService application.K8sApplicationService
	readService           read.ArgoApplicationReadService
}

func NewArgoApplicationServiceImpl(logger *zap.SugaredLogger,
	clusterRepository clusterRepository.ClusterRepository,
	k8sUtil *k8s.K8sServiceImpl,
	argoUserService argo.ArgoUserService, helmAppClient gRPC.HelmAppClient,
	helmAppService service.HelmAppService,
	k8sApplicationService application.K8sApplicationService,
	readService read.ArgoApplicationReadService) *ArgoApplicationServiceImpl {
	return &ArgoApplicationServiceImpl{
		logger:                logger,
		clusterRepository:     clusterRepository,
		k8sUtil:               k8sUtil,
		argoUserService:       argoUserService,
		helmAppService:        helmAppService,
		helmAppClient:         helmAppClient,
		k8sApplicationService: k8sApplicationService,
		readService:           readService,
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
		clusterBean := cluster2.GetClusterBean(clusterObj)
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
	return appListFinal, nil
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
				ClusterId:    clusterId,
				ClusterName:  clusterName,
				Name:         rowData[k8sCommonBean.K8sResourceColumnDefinitionName].(string),
				SyncStatus:   rowData[k8sCommonBean.K8sResourceColumnDefinitionSyncStatus].(string),
				HealthStatus: rowData[k8sCommonBean.K8sResourceColumnDefinitionHealthStatus].(string),
				Namespace:    rowData[k8sCommonBean.K8sClusterResourceNamespaceKey].(string),
			}
			appLists[i] = appListDto
		}
	}
	return appLists
}

func (impl *ArgoApplicationServiceImpl) HibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	_, clusterBean, _, err := impl.readService.GetClusterConfigFromAllClusters(app.ClusterId)
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
	_, clusterBean, _, err := impl.readService.GetClusterConfigFromAllClusters(app.ClusterId)
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
