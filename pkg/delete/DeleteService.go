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

package delete

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	bean4 "github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/chartRepo"
	"github.com/devtron-labs/devtron/pkg/cluster"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/team"
	bean3 "github.com/devtron-labs/devtron/pkg/team/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	http2 "net/http"
)

type DeleteService interface {
	DeleteCluster(deleteRequest *bean2.ClusterBean, userId int32) error
	DeleteEnvironment(deleteRequest *bean.EnvironmentBean, userId int32) error
	DeleteTeam(deleteRequest *bean3.TeamRequest) error
	DeleteChartRepo(deleteRequest *chartRepo.ChartRepoDto) error
	DeleteDockerRegistryConfig(deleteRequest *types.DockerArtifactStoreBean) error
	CanDeleteChartRegistryPullConfig(storeId string) bool
	DeleteClusterConfigMap(deleteRequest *bean2.ClusterBean, err error) error
}

type DeleteServiceImpl struct {
	logger                   *zap.SugaredLogger
	teamService              team.TeamService
	clusterService           cluster.ClusterService
	environmentService       environment.EnvironmentService
	chartRepositoryService   chartRepo.ChartRepositoryService
	installedAppRepository   repository.InstalledAppRepository
	dockerRegistryConfig     pipeline.DockerRegistryConfig
	dockerRegistryRepository dockerRegistryRepository.DockerArtifactStoreRepository
	K8sUtil                  k8s.K8sService
	k8sInformerFactory       informer.K8sInformerFactory
}

func NewDeleteServiceImpl(logger *zap.SugaredLogger,
	teamService team.TeamService,
	clusterService cluster.ClusterService,
	environmentService environment.EnvironmentService,
	chartRepositoryService chartRepo.ChartRepositoryService,
	installedAppRepository repository.InstalledAppRepository,
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	dockerRegistryRepository dockerRegistryRepository.DockerArtifactStoreRepository,
	k8sInformerFactory informer.K8sInformerFactory,
	K8sUtil k8s.K8sService,
) *DeleteServiceImpl {
	return &DeleteServiceImpl{
		logger:                   logger,
		teamService:              teamService,
		clusterService:           clusterService,
		environmentService:       environmentService,
		chartRepositoryService:   chartRepositoryService,
		installedAppRepository:   installedAppRepository,
		dockerRegistryConfig:     dockerRegistryConfig,
		dockerRegistryRepository: dockerRegistryRepository,
		K8sUtil:                  K8sUtil,
		k8sInformerFactory:       k8sInformerFactory,
	}
}

func (impl DeleteServiceImpl) DeleteCluster(deleteRequest *bean2.ClusterBean, userId int32) error {
	clusterName, err := impl.clusterService.DeleteFromDb(deleteRequest, userId)
	if err != nil {
		impl.logger.Errorw("error im deleting cluster", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	// deleting a cluster config map created at time of cluster creation/updation so that informer in kubelink and kubewatch can delete the cluster from cache
	err = impl.DeleteClusterConfigMap(deleteRequest, err)
	if err != nil {
		impl.logger.Errorw("error in deleting cluster cm", "clusterId", deleteRequest.Id, "error", err)
		return err
	}
	impl.k8sInformerFactory.DeleteClusterFromCache(clusterName)
	return nil
}

func (impl DeleteServiceImpl) DeleteClusterConfigMap(deleteRequest *bean2.ClusterBean, err error) error {
	// kubelink informers are listening this secret, deleting this secret will inform kubelink that this cluster is deleted
	k8sClient, err := impl.K8sUtil.GetCoreV1ClientInCluster()
	if err != nil {
		impl.logger.Errorw("error in getting in cluster k8s client", "err", err, "clusterName", deleteRequest.ClusterName)
		return nil
	}
	cmName := cluster.ParseCmNameForK8sInformerOnClusterEvent(deleteRequest.Id)
	err = impl.K8sUtil.DeleteConfigMap(bean4.DevtronCDNamespae, cmName, k8sClient)
	if k8sError.IsNotFound(err) {
		// when cm not found in devtroncd ns then delete the secret in default ns(secret name would be the same as cm name)
		err1 := impl.K8sUtil.DeleteSecret(bean2.DefaultNamespace, cmName, k8sClient)
		if err1 != nil {
			impl.logger.Errorw("error in deleting cluster secret in default ns ", "secretName", cmName, "err", err1)
			return err1
		}
	} else if err != nil {
		impl.logger.Errorw("error in deleting cluster config map in devtroncd ns ", "cmName", cmName, "err", err)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteEnvironment(deleteRequest *bean.EnvironmentBean, userId int32) error {
	err := impl.environmentService.Delete(deleteRequest, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting environment", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}
func (impl DeleteServiceImpl) DeleteTeam(deleteRequest *bean3.TeamRequest) error {
	err := impl.teamService.Delete(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting team", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteChartRepo(deleteRequest *chartRepo.ChartRepoDto) error {

	deployedCharts, err := impl.installedAppRepository.GetAllInstalledAppsByChartRepoId(deleteRequest.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting repo", "deleteRequest", deployedCharts)
		return err
	}
	if len(deployedCharts) > 0 {
		impl.logger.Errorw("err in deleting repo, found charts deployed using this repo", "deleteRequest", deployedCharts)
		return fmt.Errorf("cannot delete repo, found charts deployed in this repo")
	}
	err = impl.chartRepositoryService.DeleteChartRepo(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting chart repo", "err", err, "deleteRequest", deleteRequest)
		return err
	}

	return nil
}

func (impl DeleteServiceImpl) DeleteDockerRegistryConfig(deleteRequest *types.DockerArtifactStoreBean) error {
	deploymentCount, err := impl.dockerRegistryRepository.FindDeploymentCount(deleteRequest.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting docker registry", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	if deploymentCount > 0 {
		impl.logger.Errorw("err in deleting docker registry, found chart deployments using registry", "dockerRegistry", deleteRequest.Id, "err", err)
		return &util.ApiError{
			HttpStatusCode:  http2.StatusUnprocessableEntity,
			InternalMessage: " Please update all related docker config before deleting this registry",
			UserMessage:     "err in deleting docker registry, found chart deployments using registry",
		}
	}
	err = impl.dockerRegistryConfig.DeleteReg(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting docker registry", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) CanDeleteChartRegistryPullConfig(storeId string) bool {
	//finding if docker reg chart is used in any deployment, if yes then will not delete
	deploymentCount, err := impl.dockerRegistryRepository.FindDeploymentCount(storeId)
	if err != nil {
		impl.logger.Errorw("error in fetching registry chart deployment docker registry", "dockerRegistry", storeId, "err", err)
		return false
	}
	if deploymentCount > 0 {
		return false
	}
	return true
}
