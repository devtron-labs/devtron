/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/devtron-labs/common-lib/async"
	informerBean "github.com/devtron-labs/common-lib/informer"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	configMap2 "github.com/devtron-labs/common-lib/utils/k8s/configMap"
	bean3 "github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/helper"
	"github.com/devtron-labs/devtron/pkg/cluster/read"
	cronUtil "github.com/devtron-labs/devtron/util/cron"
	"github.com/robfig/cron/v3"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/common-lib/utils/k8s"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	repository3 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	customErr "github.com/juju/errors"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ClusterService interface {
	Save(parent context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error)
	UpdateClusterDescription(bean *bean.ClusterBean, userId int32) error
	ValidateKubeconfig(kubeConfig string) (map[string]*bean.ValidateClusterBean, error)
	FindOne(clusterName string) (*bean.ClusterBean, error)
	FindOneActive(clusterName string) (*bean.ClusterBean, error)
	FindAll() ([]*bean.ClusterBean, error)
	FindAllExceptVirtual() ([]*bean.ClusterBean, error)
	FindAllWithoutConfig() ([]*bean.ClusterBean, error)
	FindAllActive() ([]bean.ClusterBean, error)
	DeleteFromDb(bean *clusterBean.DeleteClusterBean, userId int32) (string, error)

	FindById(id int) (*bean.ClusterBean, error)
	FindByIdWithoutConfig(id int) (*bean.ClusterBean, error)
	FindByIds(id []int) ([]bean.ClusterBean, error)
	FindByIdsWithoutConfig(ids []int) ([]*bean.ClusterBean, error)
	Update(ctx context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error)
	Delete(bean *bean.ClusterBean, userId int32) error

	FindAllForAutoComplete() ([]bean.ClusterBean, error)
	CreateGrafanaDataSource(clusterBean *bean.ClusterBean, env *repository2.Environment) (int, error)
	GetAllClusterNamespaces() map[string][]string
	FindAllNamespacesByUserIdAndClusterId(userId int32, clusterId int, isActionUserSuperAdmin bool) ([]string, error)
	FindAllForClusterByUserId(userId int32, isActionUserSuperAdmin bool) ([]bean.ClusterBean, error)
	FetchRolesFromGroup(userId int32) ([]*repository3.RoleModel, error)
	HandleErrorInClusterConnections(clusters []*bean.ClusterBean, respMap *sync.Map, clusterExistInDb bool)
	ConnectClustersInBatch(clusters []*bean.ClusterBean, clusterExistInDb bool)
	ConvertClusterBeanToCluster(clusterBean *bean.ClusterBean, userId int32) *repository.Cluster
	ConvertClusterBeanObjectToCluster(bean *bean.ClusterBean) *v1alpha1.Cluster

	GetClusterConfigByClusterId(clusterId int) (*k8s.ClusterConfig, error)
}

type ClusterServiceImpl struct {
	clusterRepository   repository.ClusterRepository
	logger              *zap.SugaredLogger
	K8sUtil             *k8s.K8sServiceImpl
	K8sInformerFactory  informer.K8sInformerFactory
	userAuthRepository  repository3.UserAuthRepository
	userRepository      repository3.UserRepository
	roleGroupRepository repository3.RoleGroupRepository
	clusterReadService  read.ClusterReadService
	asyncRunnable       *async.Runnable
}

func NewClusterServiceImpl(repository repository.ClusterRepository, logger *zap.SugaredLogger,
	K8sUtil *k8s.K8sServiceImpl, K8sInformerFactory informer.K8sInformerFactory,
	userAuthRepository repository3.UserAuthRepository, userRepository repository3.UserRepository,
	roleGroupRepository repository3.RoleGroupRepository,
	envVariables *globalUtil.EnvironmentVariables,
	cronLogger *cronUtil.CronLoggerImpl,
	clusterReadService read.ClusterReadService,
	asyncRunnable *async.Runnable) (*ClusterServiceImpl, error) {
	clusterService := &ClusterServiceImpl{
		clusterRepository:   repository,
		logger:              logger,
		K8sUtil:             K8sUtil,
		K8sInformerFactory:  K8sInformerFactory,
		userAuthRepository:  userAuthRepository,
		userRepository:      userRepository,
		roleGroupRepository: roleGroupRepository,
		clusterReadService:  clusterReadService,
		asyncRunnable:       asyncRunnable,
	}
	// initialise cron
	newCron := cron.New(cron.WithChain(cron.Recover(cronLogger)))
	newCron.Start()
	cfg := envVariables.GlobalClusterConfig
	// add function into cron
	_, err := newCron.AddFunc(fmt.Sprintf("@every %dm", cfg.ClusterStatusCronTime), clusterService.getAndUpdateClusterConnectionStatus)
	if err != nil {
		fmt.Println("error in adding cron function into cluster cron service")
		return clusterService, err
	}
	logger.Infow("cluster cron service started successfully!", "cronTime", cfg.ClusterStatusCronTime)
	go clusterService.buildInformer()
	return clusterService, nil
}

func (impl *ClusterServiceImpl) ConvertClusterBeanToCluster(clusterBean *bean.ClusterBean, userId int32) *repository.Cluster {

	model := &repository.Cluster{}

	model.ClusterName = clusterBean.ClusterName
	//model.Note = clusterBean.Note
	model.Active = true
	model.ServerUrl = clusterBean.ServerUrl
	model.Config = clusterBean.Config
	model.PrometheusEndpoint = clusterBean.PrometheusUrl
	model.InsecureSkipTlsVerify = clusterBean.InsecureSkipTLSVerify
	model.IsProd = clusterBean.IsProd

	if clusterBean.PrometheusAuth != nil {
		model.PUserName = clusterBean.PrometheusAuth.UserName
		model.PPassword = clusterBean.PrometheusAuth.Password
		model.PTlsClientCert = clusterBean.PrometheusAuth.TlsClientCert
		model.PTlsClientKey = clusterBean.PrometheusAuth.TlsClientKey
	}

	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()

	return model
}

// getAndUpdateClusterConnectionStatus is a cron function to update the connection status of all clusters
func (impl *ClusterServiceImpl) getAndUpdateClusterConnectionStatus() {
	impl.logger.Info("starting cluster connection status fetch thread")
	startTime := time.Now()
	defer func() {
		impl.logger.Debugw("cluster connection status fetch thread completed", "timeTaken", time.Since(startTime))
	}()

	//getting all clusters
	clusters, err := impl.FindAll()
	if err != nil {
		impl.logger.Errorw("error in getting all clusters", "err", err)
		return
	}
	impl.ConnectClustersInBatch(clusters, true)
}

func (impl *ClusterServiceImpl) Save(parent context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	//validating config

	err := impl.CheckIfConfigIsValid(bean)

	if err != nil {
		if len(err.Error()) > 2000 {
			err = k8sError.NewBadRequest("unable to connect to cluster")
		}
		return nil, err
	}

	existingModel, err := impl.clusterRepository.FindOne(bean.ClusterName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error(err)
		return nil, err
	}
	if existingModel.Id > 0 {
		impl.logger.Errorw("error on fetching cluster, duplicate", "name", bean.ClusterName)
		return nil, fmt.Errorf("cluster already exists")
	}

	model := impl.ConvertClusterBeanToCluster(bean, userId)

	cfg := bean.GetClusterConfig()
	client, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
	if err != nil {
		return nil, err
	}
	k8sServerVersion, err := client.ServerVersion()
	if err != nil {
		return nil, err
	}
	model.K8sVersion = k8sServerVersion.String()
	err = impl.clusterRepository.Save(model)
	if err != nil {
		impl.logger.Errorw("error in saving cluster in db", "err", err)
		err = &util.ApiError{
			Code:            constants.ClusterCreateDBFailed,
			InternalMessage: "cluster creation failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
	}
	bean.Id = model.Id

	//on successful creation of new cluster, update informer cache for namespace group by cluster
	//here sync for ea mode only
	if globalUtil.IsBaseStack() {
		impl.SyncNsInformer(bean)
	}
	impl.logger.Info("saving secret for cluster informer")
	cmData, labels := helper.CreateClusterModifyEventData(bean.Id, informerBean.ClusterActionAdd)
	if err = impl.upsertClusterConfigMap(bean, cmData, labels); err != nil {
		impl.logger.Errorw("error upserting cluster secret", "cmData", cmData, "err", err)
		return bean, nil
	}
	return bean, nil
}

// UpdateClusterDescription is new api service logic to only update description, this should be done in cluster update operation only
// but not supported currently as per product
func (impl *ClusterServiceImpl) UpdateClusterDescription(bean *bean.ClusterBean, userId int32) error {
	//updating description as other fields are not supported yet
	err := impl.clusterRepository.SetDescription(bean.Id, bean.Description, userId)
	if err != nil {
		impl.logger.Errorw("error in setting cluster description", "err", err, "clusterId", bean.Id)
		return err
	}
	return nil
}

func (impl *ClusterServiceImpl) FindOne(clusterName string) (*bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindOne(clusterName)
	if err != nil {
		return nil, err
	}
	bean := adapter.GetClusterBean(*model)
	return &bean, nil
}

func (impl *ClusterServiceImpl) FindOneActive(clusterName string) (*bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindOneActive(clusterName)
	if err != nil {
		return nil, err
	}
	bean := adapter.GetClusterBean(*model)
	return &bean, nil
}

func (impl *ClusterServiceImpl) FindAllWithoutConfig() ([]*bean.ClusterBean, error) {
	models, err := impl.FindAll()
	if err != nil {
		return nil, err
	}
	for _, model := range models {
		model.Config = map[string]string{commonBean.BearerToken: ""}
	}
	return models, nil
}

func (impl *ClusterServiceImpl) FindAll() ([]*bean.ClusterBean, error) {
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		return nil, err
	}
	var beans []*bean.ClusterBean
	for _, model := range models {
		bean := adapter.GetClusterBean(model)
		bean.SetClusterStatus()
		beans = append(beans, &bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FindAllExceptVirtual() ([]*bean.ClusterBean, error) {
	models, err := impl.clusterRepository.FindAllActiveExceptVirtual()
	if err != nil {
		return nil, err
	}
	var beans []*bean.ClusterBean
	for _, model := range models {
		bean := adapter.GetClusterBean(model)
		beans = append(beans, &bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FindAllActive() ([]bean.ClusterBean, error) {
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		return nil, err
	}
	var beans []bean.ClusterBean
	for _, model := range models {
		bean := adapter.GetClusterBean(model)
		beans = append(beans, bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FindById(id int) (*bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	bean := adapter.GetClusterBean(*model)
	return &bean, nil
}

func (impl *ClusterServiceImpl) FindByIdWithoutConfig(id int) (*bean.ClusterBean, error) {
	model, err := impl.clusterReadService.FindById(id)
	if err != nil {
		return nil, err
	}
	//empty bearer token as it will be hidden for user
	model.Config = map[string]string{commonBean.BearerToken: ""}
	return model, nil
}

func (impl *ClusterServiceImpl) FindByIds(ids []int) ([]bean.ClusterBean, error) {
	models, err := impl.clusterRepository.FindByIds(ids)
	if err != nil {
		return nil, err
	}
	var beans []bean.ClusterBean

	for _, model := range models {
		bean := adapter.GetClusterBean(model)
		beans = append(beans, bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FindByIdsWithoutConfig(ids []int) ([]*bean.ClusterBean, error) {
	models, err := impl.clusterRepository.FindByIds(ids)
	if err != nil {
		return nil, err
	}
	var beans []*bean.ClusterBean
	for _, model := range models {
		bean := adapter.GetClusterBean(model)
		//empty bearer token as it will be hidden for user
		bean.Config = map[string]string{commonBean.BearerToken: ""}
		beans = append(beans, &bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) Update(ctx context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	existingModel, err := impl.clusterRepository.FindOne(bean.ClusterName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error(err)
		return nil, err
	}
	if existingModel.Id > 0 && model.Id != existingModel.Id {
		impl.logger.Errorw("error on fetching cluster, duplicate", "name", bean.ClusterName)
		return nil, fmt.Errorf("cluster already exists")
	}

	// check whether config modified or not, if yes create informer with updated config
	dbConfigBearerToken := model.Config[commonBean.BearerToken]
	requestConfigBearerToken := bean.Config[commonBean.BearerToken]
	if len(requestConfigBearerToken) == 0 {
		bean.Config[commonBean.BearerToken] = model.Config[commonBean.BearerToken]
	}

	dbConfigTlsKey := model.Config[commonBean.TlsKey]
	requestConfigTlsKey := bean.Config[commonBean.TlsKey]
	if len(requestConfigTlsKey) == 0 {
		bean.Config[commonBean.TlsKey] = model.Config[commonBean.TlsKey]
	}

	dbConfigCertData := model.Config[commonBean.CertData]
	requestConfigCertData := bean.Config[commonBean.CertData]
	if len(requestConfigCertData) == 0 {
		bean.Config[commonBean.CertData] = model.Config[commonBean.CertData]
	}

	dbConfigCAData := model.Config[commonBean.CertificateAuthorityData]
	requestConfigCAData := bean.Config[commonBean.CertificateAuthorityData]
	if len(requestConfigCAData) == 0 {
		bean.Config[commonBean.CertificateAuthorityData] = model.Config[commonBean.CertificateAuthorityData]
	}

	if bean.ServerUrl != model.ServerUrl || bean.InsecureSkipTLSVerify != model.InsecureSkipTlsVerify || dbConfigBearerToken != requestConfigBearerToken || dbConfigTlsKey != requestConfigTlsKey || dbConfigCertData != requestConfigCertData || dbConfigCAData != requestConfigCAData {
		if bean.ClusterName == clusterBean.DefaultCluster {
			impl.logger.Errorw("default_cluster is reserved by the system and cannot be updated, default_cluster", "name", bean.ClusterName)
			return nil, fmt.Errorf("default_cluster is reserved by the system and cannot be updated")
		}
		bean.HasConfigOrUrlChanged = true
		//validating config
		err := impl.CheckIfConfigIsValid(bean)
		if err != nil {
			return nil, err
		}
	}
	model.ClusterName = bean.ClusterName
	model.ServerUrl = bean.ServerUrl
	model.InsecureSkipTlsVerify = bean.InsecureSkipTLSVerify
	model.IsProd = bean.IsProd
	model.PrometheusEndpoint = bean.PrometheusUrl

	if bean.PrometheusAuth != nil {
		if bean.PrometheusAuth.UserName != "" || bean.PrometheusAuth.IsAnonymous {
			model.PUserName = bean.PrometheusAuth.UserName
		}
		if bean.PrometheusAuth.Password != "" || bean.PrometheusAuth.IsAnonymous {
			model.PPassword = bean.PrometheusAuth.Password
		}
		if bean.PrometheusAuth.TlsClientCert != "" {
			model.PTlsClientCert = bean.PrometheusAuth.TlsClientCert
		}
		if bean.PrometheusAuth.TlsClientKey != "" {
			model.PTlsClientKey = bean.PrometheusAuth.TlsClientKey
		}
	}
	model.ErrorInConnecting = "" //setting empty because config to be updated is already validated
	model.Active = bean.Active
	model.Config = bean.Config
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()

	if model.K8sVersion == "" {
		cfg := bean.GetClusterConfig()
		client, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
		if err != nil {
			return nil, err
		}
		k8sServerVersion, err := client.ServerVersion()
		if err != nil {
			return nil, err
		}
		model.K8sVersion = k8sServerVersion.String()
	}
	err = impl.clusterRepository.Update(model)
	if err != nil {
		err = &util.ApiError{
			Code:            constants.ClusterUpdateDBFailed,
			InternalMessage: "cluster update failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
		return bean, err
	}
	bean.Id = model.Id

	//here sync for ea mode only
	if bean.HasConfigOrUrlChanged && globalUtil.IsBaseStack() {
		impl.SyncNsInformer(bean)
	}
	impl.logger.Infow("saving secret for cluster informer")
	if bean.HasConfigOrUrlChanged {
		cmData, labels := helper.CreateClusterModifyEventData(bean.Id, informerBean.ClusterActionUpdate)
		if err = impl.upsertClusterConfigMap(bean, cmData, labels); err != nil {
			impl.logger.Errorw("error upserting cluster secret", "cmData", cmData, "err", err)
			// TODO Asutosh: why error is not propagated ??
			return bean, nil
		}
	}
	return bean, nil
}

func (impl *ClusterServiceImpl) upsertClusterConfigMap(bean *bean.ClusterBean, data, labels map[string]string) error {
	k8sClient, err := impl.K8sUtil.GetCoreV1ClientInCluster()
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "err", err)
		return err
	}
	// below cm will act as an event for informer running on a secret object in kubelink and kubewatch
	cmName := ParseCmNameForK8sInformerOnClusterEvent(bean.Id)
	configMap, err := impl.K8sUtil.GetConfigMap(bean3.DevtronCDNamespae, cmName, k8sClient)
	if err != nil && !k8sError.IsNotFound(err) {
		impl.logger.Errorw("error in getting cluster config map", "cmName", cmName, "err", err)
		return err
	} else if k8sError.IsNotFound(err) {
		_, err = impl.K8sUtil.CreateConfigMapObject(cmName, bean3.DevtronCDNamespae, k8sClient, configMap2.WithData(data), configMap2.WithLabels(labels))
		if err != nil {
			impl.logger.Errorw("error in creating cm object for informer", "cmName", cmName, "err", err)
			return err
		}
	} else {
		configMap.Labels = labels
		configMap.Data = data
		configMap, err = impl.K8sUtil.UpdateConfigMap(bean3.DevtronCDNamespae, configMap, k8sClient)
		if err != nil {
			impl.logger.Errorw("error in updating cm for informers", "cmName", cmName, "err", err)
			return err
		}
	}
	return nil
}

func (impl *ClusterServiceImpl) SyncNsInformer(bean *bean.ClusterBean) {
	requestConfig := bean.Config[commonBean.BearerToken]
	//before creating new informer for cluster, close existing one
	impl.K8sInformerFactory.CleanNamespaceInformer(bean.ClusterName)
	//create new informer for cluster with new config
	clusterInfo := &bean2.ClusterInfo{
		ClusterId:             bean.Id,
		ClusterName:           bean.ClusterName,
		BearerToken:           requestConfig,
		ServerUrl:             bean.ServerUrl,
		InsecureSkipTLSVerify: bean.InsecureSkipTLSVerify,
	}
	if !bean.InsecureSkipTLSVerify {
		clusterInfo.KeyData = bean.Config[commonBean.TlsKey]
		clusterInfo.CertData = bean.Config[commonBean.CertData]
		clusterInfo.CAData = bean.Config[commonBean.CertificateAuthorityData]
	}
	impl.K8sInformerFactory.BuildInformer([]*bean2.ClusterInfo{clusterInfo})
}

func (impl *ClusterServiceImpl) Delete(bean *bean.ClusterBean, userId int32) error {
	model, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		return err
	}
	return impl.clusterRepository.Delete(model)
}

func (impl *ClusterServiceImpl) FindAllForAutoComplete() ([]bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindAll()
	if err != nil {
		return nil, err
	}
	var beans []bean.ClusterBean
	for _, m := range model {
		beans = append(beans, bean.ClusterBean{
			Id:                m.Id,
			ClusterName:       m.ClusterName,
			ErrorInConnecting: m.ErrorInConnecting,
			IsCdArgoSetup:     m.CdArgoSetup,
			IsVirtualCluster:  m.IsVirtualCluster,
			IsProd:            m.IsProd,
		})
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) CreateGrafanaDataSource(clusterBean *bean.ClusterBean, env *repository2.Environment) (int, error) {
	impl.logger.Errorw("CreateGrafanaDataSource not implemented in ClusterServiceImpl")
	return 0, fmt.Errorf("method not implemented")
}

func (impl *ClusterServiceImpl) buildInformer() {
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
		return
	}
	var clusterInfo []*bean2.ClusterInfo
	for _, model := range models {
		if !model.IsVirtualCluster {
			bearerToken := model.Config[commonBean.BearerToken]
			clusterInfo = append(clusterInfo, &bean2.ClusterInfo{
				ClusterId:             model.Id,
				ClusterName:           model.ClusterName,
				BearerToken:           bearerToken,
				ServerUrl:             model.ServerUrl,
				InsecureSkipTLSVerify: model.InsecureSkipTlsVerify,
				KeyData:               model.Config[commonBean.TlsKey],
				CertData:              model.Config[commonBean.CertData],
				CAData:                model.Config[commonBean.CertificateAuthorityData],
			})
		}
	}
	impl.K8sInformerFactory.BuildInformer(clusterInfo)
}

func (impl *ClusterServiceImpl) DeleteFromDb(bean *clusterBean.DeleteClusterBean, userId int32) (string, error) {
	existingCluster, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", bean.Id)
		return "", err
	}
	deleteReq := existingCluster
	deleteReq.UpdatedOn = time.Now()
	deleteReq.UpdatedBy = userId
	err = impl.clusterRepository.MarkClusterDeleted(deleteReq)
	if err != nil {
		impl.logger.Errorw("error in deleting cluster", "id", bean.Id, "err", err)
		return "", err
	}
	return existingCluster.ClusterName, nil
}

func (impl *ClusterServiceImpl) CheckIfConfigIsValid(cluster *bean.ClusterBean) error {
	clusterConfig := cluster.GetClusterConfig()
	response, err := impl.K8sUtil.DiscoveryClientGetLiveZCall(clusterConfig)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			return fmt.Errorf("Incorrect server url : %v", err)
		} else if statusError, ok := err.(*k8sError.StatusError); ok {
			if statusError != nil {
				errReason := statusError.ErrStatus.Reason
				var errMsg string
				if errReason == v1.StatusReasonUnauthorized {
					errMsg = "token seems invalid or does not have sufficient permissions"
				} else {
					errMsg = statusError.ErrStatus.Message
				}
				return fmt.Errorf("%s : %s", errReason, errMsg)
			} else {
				return fmt.Errorf("Validation failed : %v", err)
			}
		} else {
			return fmt.Errorf("Validation failed : %v", err)
		}
	} else if err == nil && string(response) != "ok" {
		return fmt.Errorf("Validation failed with response : %s", string(response))
	}
	return nil
}

func (impl *ClusterServiceImpl) GetAllClusterNamespaces() map[string][]string {
	result := make(map[string][]string)
	namespaceListGroupByCLuster := impl.K8sInformerFactory.GetLatestNamespaceListGroupByCLuster()
	for clusterName, namespaces := range namespaceListGroupByCLuster {
		copiedNamespaces := result[clusterName]
		for namespace, value := range namespaces {
			if value {
				copiedNamespaces = append(copiedNamespaces, namespace)
			}
		}
		result[clusterName] = copiedNamespaces
	}
	return result
}

const (
	// Cluster connectivity error constants
	ErrClusterNotReachable = "cluster is not reachable"
)

func (impl *ClusterServiceImpl) FindAllNamespacesByUserIdAndClusterId(userId int32, clusterId int, isActionUserSuperAdmin bool) ([]string, error) {
	result := make([]string, 0)
	clusterBean, err := impl.clusterReadService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("failed to find cluster for id", "error", err, "clusterId", clusterId)
		return nil, err
	}

	// Check if cluster has connection errors
	if len(clusterBean.ErrorInConnecting) > 0 {
		impl.logger.Errorw("cluster is not reachable", "clusterId", clusterId, "clusterName", clusterBean.ClusterName, "error", clusterBean.ErrorInConnecting)
		return nil, fmt.Errorf("%s: %s", ErrClusterNotReachable, clusterBean.ErrorInConnecting)
	}

	namespaceListGroupByCLuster := impl.K8sInformerFactory.GetLatestNamespaceListGroupByCLuster()
	namespaces := namespaceListGroupByCLuster[clusterBean.ClusterName]
	if len(namespaces) == 0 {
		// TODO: Verify if this is a valid scenario, if yes, then handle it
		// ideally, all clusters should have at least one `default` namespace
		// this is a fallback scenario, for handling the namespace informer failure at start up...
		impl.logger.Warnw("no namespaces found for cluster", "clusterName", clusterBean.ClusterName)
		impl.SyncNsInformer(clusterBean)
	}
	if isActionUserSuperAdmin {
		for namespace, value := range namespaces {
			if value {
				result = append(result, namespace)
			}
		}
	} else {
		roles, err := impl.FetchRolesFromGroup(userId)
		if err != nil {
			impl.logger.Errorw("error on fetching user roles for cluster list", "err", err)
			return nil, err
		}
		allowedAll := false
		allowedNamespaceMap := make(map[string]bool)
		for _, role := range roles {
			if clusterBean.ClusterName == role.Cluster {
				allowedNamespaceMap[role.Namespace] = true
				if role.Namespace == "" {
					allowedAll = true
				}
			}
		}

		//adding final namespace list
		for namespace, value := range namespaces {
			if _, ok := allowedNamespaceMap[namespace]; ok || allowedAll {
				if value {
					result = append(result, namespace)
				}
			}
		}
	}
	return result, nil
}

func (impl *ClusterServiceImpl) FindAllForClusterByUserId(userId int32, isActionUserSuperAdmin bool) ([]bean.ClusterBean, error) {
	if isActionUserSuperAdmin {
		return impl.FindAllForAutoComplete()
	}
	allowedClustersMap := make(map[string]bool)
	roles, err := impl.FetchRolesFromGroup(userId)
	if err != nil {
		impl.logger.Errorw("error while fetching user roles from db", "error", err)
		return nil, err
	}
	for _, role := range roles {
		allowedClustersMap[role.Cluster] = true
	}

	models, err := impl.clusterRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error on fetching clusters", "err", err)
		return nil, err
	}
	var beans []bean.ClusterBean
	for _, model := range models {
		if _, ok := allowedClustersMap[model.ClusterName]; ok {
			beans = append(beans, bean.ClusterBean{
				Id:                model.Id,
				ClusterName:       model.ClusterName,
				ErrorInConnecting: model.ErrorInConnecting,
			})
		}
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FetchRolesFromGroup(userId int32) ([]*repository3.RoleModel, error) {
	user, err := impl.userRepository.GetByIdIncludeDeleted(userId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	groups, err := casbin2.GetRolesForUser(user.EmailId)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "id", user.Id)
		return nil, err
	}
	roleEntity := "cluster"
	roles, err := impl.userAuthRepository.GetRolesByUserIdAndEntityType(userId, roleEntity)
	if err != nil {
		impl.logger.Errorw("error on fetching user roles for cluster list", "err", err)
		return nil, err
	}
	rolesFromGroup, err := impl.roleGroupRepository.GetRolesByGroupNamesAndEntity(groups, roleEntity)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting roles by group names", "err", err)
		return nil, err
	}
	if len(rolesFromGroup) > 0 {
		roles = append(roles, rolesFromGroup...)
	}
	return roles, nil
}

func (impl *ClusterServiceImpl) updateConnectionStatusForVirtualCluster(respMap *sync.Map, clusterId int, clusterName string) {
	connErr := fmt.Errorf("Get virtual cluster '%s' error: connection not setup for isolated clusters", clusterName)
	respMap.Store(clusterId, connErr)
}

func (impl *ClusterServiceImpl) ConnectClustersInBatch(clusters []*bean.ClusterBean, clusterExistInDb bool) {
	var wg sync.WaitGroup
	respMap := &sync.Map{}
	for idx := range clusters {
		cluster := clusters[idx]
		if cluster.IsVirtualCluster {
			impl.updateConnectionStatusForVirtualCluster(respMap, cluster.Id, cluster.ClusterName)
			continue
		}
		wg.Add(1)
		runnableFunc := func(idx int, cluster *bean.ClusterBean) {
			defer wg.Done()
			clusterConfig := cluster.GetClusterConfig()
			_, _, k8sClientSet, err := impl.K8sUtil.GetK8sConfigAndClients(clusterConfig)
			if err != nil {
				respMap.Store(cluster.Id, err)
				return
			}

			id := cluster.Id
			if !clusterExistInDb {
				id = idx
			}
			impl.GetAndUpdateConnectionStatusForOneCluster(k8sClientSet, id, respMap)
		}
		impl.asyncRunnable.Execute(func() { runnableFunc(idx, cluster) })
	}

	wg.Wait()
	impl.HandleErrorInClusterConnections(clusters, respMap, clusterExistInDb)
}

func (impl *ClusterServiceImpl) HandleErrorInClusterConnections(clusters []*bean.ClusterBean, respMap *sync.Map, clusterExistInDb bool) {
	respMap.Range(func(key, value any) bool {
		defer func() {
			// defer to handle panic on type assertion
			if r := recover(); r != nil {
				impl.logger.Errorw("error in handling error in cluster connections", "key", key, "value", value, "err", r)
			}
		}()
		id := key.(int)
		var err error
		if connectionError, ok := value.(error); ok {
			err = connectionError
		}
		errorInConnecting := ""
		if err != nil {
			errorInConnecting = err.Error()
			// limiting error message to 2000 characters. Can be changed if needed.
			if len(errorInConnecting) > 2000 {
				errorInConnecting = "unable to connect to cluster"
			}
		}
		//updating cluster connection status
		if clusterExistInDb {
			//id is clusterId if clusterExistInDb
			errInUpdating := impl.clusterRepository.UpdateClusterConnectionStatus(id, errorInConnecting)
			if errInUpdating != nil {
				impl.logger.Errorw("error in updating cluster connection status", "err", err, "clusterId", id, "errorInConnecting", errorInConnecting)
			}
		} else {
			//id is index of the cluster in clusters array
			clusters[id].ErrorInConnecting = errorInConnecting
		}
		return true
	})
}

func (impl *ClusterServiceImpl) ValidateKubeconfig(kubeConfig string) (map[string]*bean.ValidateClusterBean, error) {

	kubeConfigObject := api.Config{}

	gvk := &schema.GroupVersionKind{}

	var kubeConfigDataMap map[string]interface{}
	err := json.Unmarshal([]byte(kubeConfig), &kubeConfigDataMap)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling kubeConfig")
		return nil, customErr.New("invalid kubeConfig found , " + err.Error())
	}

	if kubeConfigDataMap["apiVersion"] == nil {
		impl.logger.Errorw("api version missing from kubeConfig")
		return nil, customErr.New("api version missing from kubeConfig")
	}
	if kubeConfigDataMap["kind"] == nil {
		impl.logger.Errorw("kind missing from kubeConfig")
		return nil, customErr.New("kind missing from kubeConfig")
	}

	gvk.Version = kubeConfigDataMap["apiVersion"].(string)
	gvk.Kind = kubeConfigDataMap["kind"].(string)

	_, _, err = latest.Codec.Decode([]byte(kubeConfig), gvk, &kubeConfigObject)
	if err != nil {
		impl.logger.Errorw("error in decoding kubeConfig")
		return nil, err
	}

	//var clusterBeanObjects []*ClusterBean
	ValidateObjects := make(map[string]*bean.ValidateClusterBean)

	var clusterBeansWithNoValidationErrors []*bean.ClusterBean
	var clusterBeanObjects []*bean.ClusterBean
	if err != nil {
		return ValidateObjects, err
	}

	clusterList, err := impl.clusterRepository.FindActiveClusters()
	if err != nil {
		return nil, err
	}
	clusterListMap := make(map[string]bool)
	clusterListMapWithId := make(map[string]int)
	for _, c := range clusterList {
		clusterListMap[c.ClusterName] = c.Active
		clusterListMapWithId[c.ClusterName] = c.Id
	}

	userInfosMap := map[string]*bean.UserInfo{}
	for _, ctx := range kubeConfigObject.Contexts {
		clusterBeanObject := &bean.ClusterBean{}
		clusterName := ctx.Cluster
		userName := ctx.AuthInfo
		clusterObj := kubeConfigObject.Clusters[clusterName]
		userInfoObj := kubeConfigObject.AuthInfos[userName]

		if clusterObj == nil {
			continue
		}

		if clusterName != "" {
			clusterBeanObject.ClusterName = clusterName
		} else {
			clusterBeanObject.ErrorInConnecting = "cluster name missing from kubeconfig"
		}

		if clusterBeanObject.ClusterName == clusterBean.DefaultCluster {
			clusterBeanObject.ErrorInConnecting = "default_cluster is reserved by the system and cannot be updated"
		}

		if (clusterObj == nil || clusterObj.Server == "") && (clusterBeanObject.ErrorInConnecting == "") {
			clusterBeanObject.ErrorInConnecting = "server url missing from the kubeconfig"
		} else {
			clusterBeanObject.ServerUrl = clusterObj.Server
		}

		if gvk.Version == "" {
			clusterBeanObject.ErrorInConnecting = "api version missing from the contexts in kubeconfig"
		} else {
			clusterBeanObject.K8sVersion = gvk.Version
		}

		Config := make(map[string]string)

		if (userInfoObj == nil || userInfoObj.Token == "" && clusterObj.InsecureSkipTLSVerify) && (clusterBeanObject.ErrorInConnecting == "") {
			clusterBeanObject.ErrorInConnecting = "token missing from the kubeconfig"
		}
		Config[commonBean.BearerToken] = userInfoObj.Token

		if clusterObj != nil {
			clusterBeanObject.InsecureSkipTLSVerify = clusterObj.InsecureSkipTLSVerify
		}

		if (clusterObj != nil) && !clusterObj.InsecureSkipTLSVerify && (clusterBeanObject.ErrorInConnecting == "") {
			missingFieldsStr := ""
			if string(userInfoObj.ClientKeyData) == "" {
				missingFieldsStr += "client-key-data" + ", "
			}
			if string(clusterObj.CertificateAuthorityData) == "" {
				missingFieldsStr += "certificate-authority-data" + ", "
			}
			if string(userInfoObj.ClientCertificateData) == "" {
				missingFieldsStr += "client-certificate-data" + ", "
			}
			if len(missingFieldsStr) > 0 {
				missingFieldsStr = missingFieldsStr[:len(missingFieldsStr)-2]
				clusterBeanObject.ErrorInConnecting = fmt.Sprintf("Missing fields against user: %s", missingFieldsStr)
			} else {
				Config[commonBean.TlsKey] = string(userInfoObj.ClientKeyData)
				Config[commonBean.CertData] = string(userInfoObj.ClientCertificateData)
				Config[commonBean.CertificateAuthorityData] = string(clusterObj.CertificateAuthorityData)
			}
		}

		userInfo := bean.UserInfo{
			UserName:          userName,
			Config:            Config,
			ErrorInConnecting: clusterBeanObject.ErrorInConnecting,
		}

		userInfosMap[userInfo.UserName] = &userInfo
		validateObject := &bean.ValidateClusterBean{}
		if _, ok := ValidateObjects[clusterBeanObject.ClusterName]; !ok {
			validateObject.UserInfos = make(map[string]*bean.UserInfo)
			validateObject.ClusterBean = clusterBeanObject
			ValidateObjects[clusterBeanObject.ClusterName] = validateObject
		}
		clusterBeanObject.UserName = userName
		ValidateObjects[clusterBeanObject.ClusterName].UserInfos[userName] = &userInfo
		if clusterBeanObject.ErrorInConnecting == "" || clusterBeanObject.ErrorInConnecting == "cluster already exists" {
			clusterBeanObject.Config = Config
			clusterBeansWithNoValidationErrors = append(clusterBeansWithNoValidationErrors, clusterBeanObject)
		}
		clusterBeanObjects = append(clusterBeanObjects, clusterBeanObject)
	}

	if clusterBeansWithNoValidationErrors != nil {
		impl.ConnectClustersInBatch(clusterBeansWithNoValidationErrors, false)
	}

	for _, clusterBeanObject := range clusterBeanObjects {
		if _, ok := clusterListMap[clusterBeanObject.ClusterName]; ok && clusterBeanObject.ErrorInConnecting == "" {
			clusterBeanObject.ErrorInConnecting = "cluster-already-exists"
			clusterBeanObject.Id = clusterListMapWithId[clusterBeanObject.ClusterName]
			ValidateObjects[clusterBeanObject.ClusterName].Id = clusterBeanObject.Id
		}
		if clusterBeanObject.ErrorInConnecting != "" {
			ValidateObjects[clusterBeanObject.ClusterName].UserInfos[clusterBeanObject.UserName].ErrorInConnecting = clusterBeanObject.ErrorInConnecting
		}
	}
	for _, clusterBeanObject := range clusterBeanObjects {
		clusterBeanObject.Config = nil
		clusterBeanObject.ErrorInConnecting = ""
	}

	if len(ValidateObjects) == 0 {
		impl.logger.Errorw("No valid cluster object provided in kubeconfig for context", "context", kubeConfig)
		return nil, customErr.New("No valid cluster object provided in kubeconfig for context")
	} else {
		return ValidateObjects, nil
	}

}

func (impl *ClusterServiceImpl) GetAndUpdateConnectionStatusForOneCluster(k8sClientSet *kubernetes.Clientset, clusterId int, respMap *sync.Map) {
	response, err := impl.K8sUtil.GetLiveZCall(commonBean.LiveZ, k8sClientSet)
	log.Println("received response for cluster livez status", "response", string(response), "err", err, "clusterId", clusterId)

	if err != nil {
		if _, ok := err.(*url.Error); ok {
			err = fmt.Errorf("Incorrect server url : %v", err)
		} else if statusError, ok := err.(*k8sError.StatusError); ok {
			if statusError != nil {
				errReason := statusError.ErrStatus.Reason
				var errMsg string
				if errReason == v1.StatusReasonUnauthorized {
					errMsg = "token seems invalid or does not have sufficient permissions"
				} else {
					errMsg = statusError.ErrStatus.Message
				}
				err = fmt.Errorf("%s : %s", errReason, errMsg)
			} else {
				err = fmt.Errorf("Validation failed : %v", err)
			}
		} else {
			err = fmt.Errorf("Validation failed : %v", err)
		}
	} else if err == nil && string(response) != "ok" {
		err = fmt.Errorf("Validation failed with response : %s", string(response))
	}

	respMap.Store(clusterId, err)
}

func (impl *ClusterServiceImpl) ConvertClusterBeanObjectToCluster(bean *bean.ClusterBean) *v1alpha1.Cluster {
	configMap := bean.Config
	serverUrl := bean.ServerUrl
	bearerToken := ""
	if configMap[commonBean.BearerToken] != "" {
		bearerToken = configMap[commonBean.BearerToken]
	}
	tlsConfig := v1alpha1.TLSClientConfig{
		Insecure: bean.InsecureSkipTLSVerify,
	}

	if !bean.InsecureSkipTLSVerify {
		tlsConfig.KeyData = []byte(bean.Config[commonBean.TlsKey])
		tlsConfig.CertData = []byte(bean.Config[commonBean.CertData])
		tlsConfig.CAData = []byte(bean.Config[commonBean.CertificateAuthorityData])
	}
	cdClusterConfig := v1alpha1.ClusterConfig{
		BearerToken:     bearerToken,
		TLSClientConfig: tlsConfig,
	}

	cl := &v1alpha1.Cluster{
		Name:   bean.ClusterName,
		Server: serverUrl,
		Config: cdClusterConfig,
	}
	return cl
}

func (impl *ClusterServiceImpl) GetClusterConfigByClusterId(clusterId int) (*k8s.ClusterConfig, error) {
	clusterBean, err := impl.clusterReadService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting clusterBean by cluster id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	rq := *clusterBean
	clusterConfig := rq.GetClusterConfig()
	return clusterConfig, nil
}
