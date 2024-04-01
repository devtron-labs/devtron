/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	util3 "github.com/devtron-labs/devtron/pkg/auth/user/util"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/remoteConnection"
	serverConnectionBean "github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
	serverConnectionRepository "github.com/devtron-labs/devtron/pkg/remoteConnection/repository"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/common-lib-private/utils/k8s"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	errors1 "github.com/juju/errors"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

const (
	DEFAULT_CLUSTER                  = "default_cluster"
	DEFAULT_NAMESPACE                = "default"
	CLUSTER_MODIFY_EVENT_SECRET_TYPE = "cluster.request/modify"
	CLUSTER_ACTION_ADD               = "add"
	CLUSTER_ACTION_UPDATE            = "update"
	SECRET_NAME                      = "cluster-event"
	SECRET_FIELD_CLUSTER_ID          = "cluster_id"
	SECRET_FIELD_UPDATED_ON          = "updated_on"
	SECRET_FIELD_ACTION              = "action"
	TokenFilePath                    = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	SecretDataObfuscatePlaceholder   = "••••••••"
)

type ClusterService interface {
	Save(parent context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error)
	SaveVirtualCluster(bean *bean.VirtualClusterBean, userId int32) (*bean.VirtualClusterBean, error)
	UpdateClusterDescription(bean *bean.ClusterBean, userId int32) error
	ValidateKubeconfig(kubeConfig string) (map[string]*bean.ValidateClusterBean, error)
	FindOne(clusterName string) (*bean.ClusterBean, error)
	FindOneActive(clusterName string) (*bean.ClusterBean, error)
	FindAll() ([]*bean.ClusterBean, error)
	FindAllExceptVirtual() ([]*bean.ClusterBean, error)
	FindAllWithoutConfig() ([]*bean.ClusterBean, error)
	FindAllActive() ([]bean.ClusterBean, error)
	DeleteFromDb(bean *bean.ClusterBean, userId int32) error
	DeleteVirtualClusterFromDb(bean *bean.VirtualClusterBean, userId int32) error
	FindById(id int) (*bean.ClusterBean, error)
	FindByIdWithoutConfig(id int) (*bean.ClusterBean, error)
	FindByIds(id []int) ([]bean.ClusterBean, error)
	Update(ctx context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error)
	UpdateVirtualCluster(bean *bean.VirtualClusterBean, userId int32) (*bean.VirtualClusterBean, error)
	Delete(bean *bean.ClusterBean, userId int32) error

	FindAllForAutoComplete() ([]bean.ClusterBean, error)
	CreateGrafanaDataSource(clusterBean *bean.ClusterBean, env *repository.Environment) (int, error)
	GetAllClusterNamespaces() map[string][]string
	FindAllNamespacesByUserIdAndClusterId(userId int32, clusterId int, isActionUserSuperAdmin bool, token string) ([]string, error)
	FindAllForClusterByUserId(userId int32, isActionUserSuperAdmin bool, token string) ([]bean.ClusterBean, error)
	FetchRolesFromGroup(userId int32, token string) ([]*repository2.RoleModel, error)
	HandleErrorInClusterConnections(clusters []*bean.ClusterBean, respMap map[int]error, clusterExistInDb bool)
	ConnectClustersInBatch(clusters []*bean.ClusterBean, clusterExistInDb bool)
	ConvertClusterBeanObjectToCluster(bean *bean.ClusterBean) *v1alpha1.Cluster

	GetClusterConfigByClusterId(clusterId int) (*k8s2.ClusterConfig, error)
	IsPolicyConfiguredForCluster(envId, clusterId int) (bool, error)
}

type ClusterServiceImpl struct {
	clusterRepository                repository.ClusterRepository
	logger                           *zap.SugaredLogger
	K8sUtil                          *k8s.K8sUtilExtended
	K8sInformerFactory               informer.K8sInformerFactory
	userAuthRepository               repository2.UserAuthRepository
	userRepository                   repository2.UserRepository
	roleGroupRepository              repository2.RoleGroupRepository
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService
	serverConnectionService          remoteConnection.ServerConnectionService
	*ClusterRbacServiceImpl
}

func NewClusterServiceImpl(repository repository.ClusterRepository, logger *zap.SugaredLogger,
	K8sUtil *k8s.K8sUtilExtended, K8sInformerFactory informer.K8sInformerFactory,
	userAuthRepository repository2.UserAuthRepository, userRepository repository2.UserRepository,
	roleGroupRepository repository2.RoleGroupRepository,
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService,
	userService user.UserService, serverConnectionService remoteConnection.ServerConnectionService) *ClusterServiceImpl {
	clusterService := &ClusterServiceImpl{
		clusterRepository:       repository,
		logger:                  logger,
		K8sUtil:                 K8sUtil,
		K8sInformerFactory:      K8sInformerFactory,
		userAuthRepository:      userAuthRepository,
		userRepository:          userRepository,
		roleGroupRepository:     roleGroupRepository,
		serverConnectionService: serverConnectionService,
		ClusterRbacServiceImpl: &ClusterRbacServiceImpl{
			logger:      logger,
			userService: userService,
		},
		globalAuthorisationConfigService: globalAuthorisationConfigService,
	}
	go clusterService.buildInformer()
	return clusterService
}

func IsProxyOrSSHConfigured(bean *bean.ClusterBean) bool {
	return bean.ServerConnectionConfig != nil &&
		(bean.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodProxy ||
			bean.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodSSH)
}

func (impl *ClusterServiceImpl) Save(parent context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	bean = adapter.ConvertClusterBeanToNewClusterBean(bean) // bean is converted according to new struct
	//validating config
	k8sServerVersion, err := impl.CheckIfConfigIsValidAndGetServerVersion(bean)
	if err != nil {
		if len(err.Error()) > 2000 {
			err = errors.NewBadRequest("unable to connect to cluster")
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

	//initiate DB transaction
	dbConnection := impl.clusterRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model := adapter.ConvertClusterBeanToCluster(bean, userId)
	model.K8sVersion = k8sServerVersion.String()

	// save clusterConnectionConfig
	err = impl.serverConnectionService.CreateOrUpdateServerConnectionConfig(bean.ServerConnectionConfig, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in saving clusterConnectionConfig in db", "err", err)
		err = &util.ApiError{
			Code:            constants.ClusterCreateDBFailed,
			InternalMessage: "cluster creation failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
	}
	model.ServerConnectionConfigId = bean.ServerConnectionConfig.RemoteConnectionConfigId
	// save cluster
	err = impl.clusterRepository.Save(model, tx)
	if err != nil {
		impl.logger.Errorw("error in saving cluster in db", "err", err)
		err = &util.ApiError{
			Code:            constants.ClusterCreateDBFailed,
			InternalMessage: "cluster creation failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
	}
	bean.Id = model.Id

	// now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	//on successful creation of new cluster, update informer cache for namespace group by cluster
	//here sync for ea mode only
	if util2.IsBaseStack() {
		impl.SyncNsInformer(bean)
	}
	impl.logger.Info("saving secret for cluster informer")
	k8sClient, err := impl.K8sUtil.GetCoreV1ClientInCluster()
	if err != nil {
		impl.logger.Errorw("error in getting k8s Client in cluster", "err", err, "clusterName", bean.ClusterName)
		return bean, nil
	}
	//creating cluster secret, this secret will be read informer in kubelink to know that a new cluster has been added
	secretName := fmt.Sprintf("%s-%v", SECRET_NAME, bean.Id)

	data := make(map[string][]byte)
	data[SECRET_FIELD_CLUSTER_ID] = []byte(fmt.Sprintf("%v", bean.Id))
	data[SECRET_FIELD_ACTION] = []byte(CLUSTER_ACTION_ADD)
	data[SECRET_FIELD_UPDATED_ON] = []byte(time.Now().String()) // this field will ensure that informer detects change as other fields can be constant even if cluster config changes
	_, err = impl.K8sUtil.CreateSecret(DEFAULT_NAMESPACE, data, secretName, CLUSTER_MODIFY_EVENT_SECRET_TYPE, k8sClient, nil, nil)
	if err != nil {
		impl.logger.Errorw("error in updating secret for informers")
		return bean, nil
	}

	return bean, err
}

func (impl *ClusterServiceImpl) SaveVirtualCluster(bean *bean.VirtualClusterBean, userId int32) (*bean.VirtualClusterBean, error) {

	existingModel, err := impl.clusterRepository.FindOne(bean.ClusterName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error(err)
		return nil, err
	}
	if existingModel.Id > 0 {
		impl.logger.Errorw("error on fetching cluster, duplicate", "name", bean.ClusterName)
		return nil, fmt.Errorf("cluster already exists")
	}
	model := &repository.Cluster{
		ClusterName:      bean.ClusterName,
		Active:           true,
		IsVirtualCluster: true,
	}
	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()

	// initiate DB transaction
	dbConnection := impl.clusterRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.clusterRepository.Save(model, tx)
	if err != nil {
		impl.logger.Errorw("error in saving cluster in db")
		return nil, err
	}

	// now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, err
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
		model.Config = map[string]string{k8s2.BearerToken: ""}
		if model.ServerConnectionConfig != nil && model.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodSSH &&
			model.ServerConnectionConfig.SSHTunnelConfig != nil {
			if len(model.ServerConnectionConfig.SSHTunnelConfig.SSHPassword) > 0 {
				model.ServerConnectionConfig.SSHTunnelConfig.SSHPassword = SecretDataObfuscatePlaceholder
			}
			if len(model.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey) > 0 {
				model.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey = SecretDataObfuscatePlaceholder
			}
		}
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
	model, err := impl.FindById(id)
	if err != nil {
		return nil, err
	}
	//empty bearer token as it will be hidden for user
	model.Config = map[string]string{k8s2.BearerToken: ""}
	if model.ServerConnectionConfig != nil && model.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodSSH &&
		model.ServerConnectionConfig.SSHTunnelConfig != nil {
		if len(model.ServerConnectionConfig.SSHTunnelConfig.SSHPassword) > 0 {
			model.ServerConnectionConfig.SSHTunnelConfig.SSHPassword = SecretDataObfuscatePlaceholder
		}
		if len(model.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey) > 0 {
			model.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey = SecretDataObfuscatePlaceholder
		}
	}
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

func (impl *ClusterServiceImpl) Update(ctx context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	bean = adapter.ConvertClusterBeanToNewClusterBean(bean)
	model, err := impl.clusterRepository.FindById(bean.Id)
	model = adapter.ConvertClusterToNewCluster(model)
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
	dbConfigBearerToken := model.Config[k8s2.BearerToken]
	requestConfigBearerToken := bean.Config[k8s2.BearerToken]
	if len(requestConfigBearerToken) == 0 {
		bean.Config[k8s2.BearerToken] = model.Config[k8s2.BearerToken]
	}

	if bean.ServerConnectionConfig != nil && bean.ServerConnectionConfig.ConnectionMethod == serverConnectionBean.RemoteConnectionMethodSSH &&
		bean.ServerConnectionConfig.SSHTunnelConfig != nil {
		if bean.ServerConnectionConfig.SSHTunnelConfig.SSHPassword == SecretDataObfuscatePlaceholder {
			bean.ServerConnectionConfig.SSHTunnelConfig.SSHPassword = model.ServerConnectionConfig.SSHPassword
		}
		if bean.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey == SecretDataObfuscatePlaceholder {
			bean.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey = model.ServerConnectionConfig.SSHPassword
		}
	}

	dbConfigTlsKey := model.Config[k8s2.TlsKey]
	requestConfigTlsKey := bean.Config[k8s2.TlsKey]
	if len(requestConfigTlsKey) == 0 {
		bean.Config[k8s2.TlsKey] = model.Config[k8s2.TlsKey]
	}

	dbConfigCertData := model.Config[k8s2.CertData]
	requestConfigCertData := bean.Config[k8s2.CertData]
	if len(requestConfigCertData) == 0 {
		bean.Config[k8s2.CertData] = model.Config[k8s2.CertData]
	}

	dbConfigCAData := model.Config[k8s2.CertificateAuthorityData]
	requestConfigCAData := bean.Config[k8s2.CertificateAuthorityData]
	if len(requestConfigCAData) == 0 {
		bean.Config[k8s2.CertificateAuthorityData] = model.Config[k8s2.CertificateAuthorityData]
	}
	//below we are checking if any configuration change has been made or not that will impact the connection with the cluster
	//if any such change is made then only we will check if the given config is valid or not by connecting to the cluster
	if bean.ServerUrl != model.ServerUrl ||
		bean.InsecureSkipTLSVerify != model.InsecureSkipTlsVerify || dbConfigBearerToken != requestConfigBearerToken ||
		dbConfigTlsKey != requestConfigTlsKey || dbConfigCertData != requestConfigCertData ||
		dbConfigCAData != requestConfigCAData || checkIfConnectionConfigHasChanged(bean, model) {
		if bean.ClusterName == DEFAULT_CLUSTER {
			impl.logger.Errorw("default_cluster is reserved by the system and cannot be updated, default_cluster", "name", bean.ClusterName)
			return nil, fmt.Errorf("default_cluster is reserved by the system and cannot be updated")
		}
		bean.HasConfigOrUrlChanged = true
		//validating config
		k8sServerVersion, err := impl.CheckIfConfigIsValidAndGetServerVersion(bean)
		if err != nil {
			return nil, err
		}
		model.K8sVersion = k8sServerVersion.String()
	}
	model.ClusterName = bean.ClusterName
	model.ServerUrl = bean.ServerUrl
	model.InsecureSkipTlsVerify = bean.InsecureSkipTLSVerify
	model.PrometheusEndpoint = bean.PrometheusUrl
	if bean.ServerConnectionConfig != nil {
		model.ServerConnectionConfig = &serverConnectionRepository.RemoteConnectionConfig{
			Id:               bean.ServerConnectionConfig.RemoteConnectionConfigId,
			ConnectionMethod: bean.ServerConnectionConfig.ConnectionMethod,
		}
		if bean.ServerConnectionConfig.ProxyConfig != nil {
			model.ServerConnectionConfig.ProxyUrl = bean.ServerConnectionConfig.ProxyConfig.ProxyUrl
		} else if bean.ServerConnectionConfig.SSHTunnelConfig != nil {
			model.ServerConnectionConfig.SSHServerAddress = bean.ServerConnectionConfig.SSHTunnelConfig.SSHServerAddress
			model.ServerConnectionConfig.SSHUsername = bean.ServerConnectionConfig.SSHTunnelConfig.SSHUsername
			model.ServerConnectionConfig.SSHPassword = bean.ServerConnectionConfig.SSHTunnelConfig.SSHPassword
			model.ServerConnectionConfig.SSHAuthKey = bean.ServerConnectionConfig.SSHTunnelConfig.SSHAuthKey
		} else {
			return nil, err
		}
	}
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
	// initiate DB transaction
	dbConnection := impl.clusterRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.serverConnectionService.CreateOrUpdateServerConnectionConfig(bean.ServerConnectionConfig, userId, tx)
	if err != nil {
		err = &util.ApiError{
			HttpStatusCode:  http.StatusInternalServerError,
			Code:            constants.ClusterUpdateDBFailed,
			InternalMessage: "clusterConnectionConfig update failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
		return bean, err
	}
	model.ServerConnectionConfigId = bean.ServerConnectionConfig.RemoteConnectionConfigId
	err = impl.clusterRepository.Update(model, tx)
	if err != nil {
		err = &util.ApiError{
			Code:            constants.ClusterUpdateDBFailed,
			HttpStatusCode:  http.StatusInternalServerError,
			InternalMessage: "cluster update failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
		return bean, err
	}
	bean.Id = model.Id

	// now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	//here sync for ea mode only
	if bean.HasConfigOrUrlChanged && util2.IsBaseStack() {
		impl.SyncNsInformer(bean)
	}
	impl.logger.Infow("saving secret for cluster informer")
	k8sClient, err := impl.K8sUtil.GetCoreV1ClientInCluster()
	if err != nil {
		return bean, nil
	}
	// below secret will act as an event for informer running on secret object in kubelink
	if bean.HasConfigOrUrlChanged {
		secretName := fmt.Sprintf("%s-%v", SECRET_NAME, bean.Id)
		secret, err := impl.K8sUtil.GetSecret(DEFAULT_NAMESPACE, secretName, k8sClient)
		statusError, _ := err.(*errors.StatusError)
		if err != nil && statusError.Status().Code != http.StatusNotFound {
			impl.logger.Errorw("secret not found", "err", err)
			return bean, nil
		}
		data := make(map[string][]byte)
		data[SECRET_FIELD_CLUSTER_ID] = []byte(fmt.Sprintf("%v", bean.Id))
		data[SECRET_FIELD_ACTION] = []byte(CLUSTER_ACTION_UPDATE)
		data[SECRET_FIELD_UPDATED_ON] = []byte(time.Now().String()) // this field will ensure that informer detects change as other fields can be constant even if cluster config changes
		if secret == nil {
			_, err = impl.K8sUtil.CreateSecret(DEFAULT_NAMESPACE, data, secretName, CLUSTER_MODIFY_EVENT_SECRET_TYPE, k8sClient, nil, nil)
			if err != nil {
				impl.logger.Errorw("error in creating secret for informers")
			}
		} else {
			secret.Data = data
			secret, err = impl.K8sUtil.UpdateSecret(DEFAULT_NAMESPACE, secret, k8sClient)
			if err != nil {
				impl.logger.Errorw("error in updating secret for informers")
			}
		}
	}
	return bean, nil
}

func checkIfConnectionConfigHasChanged(bean *bean.ClusterBean, model *repository.Cluster) bool {
	beanConnectionConfig := bean.ServerConnectionConfig
	modelConnectionConfig := model.ServerConnectionConfig
	hasConnectionConfigChanged := beanConnectionConfig != nil && modelConnectionConfig != nil &&
		(beanConnectionConfig.ConnectionMethod != modelConnectionConfig.ConnectionMethod || (beanConnectionConfig.ProxyConfig != nil && beanConnectionConfig.ProxyConfig.ProxyUrl != modelConnectionConfig.ProxyUrl) ||
			(beanConnectionConfig.SSHTunnelConfig != nil && (beanConnectionConfig.SSHTunnelConfig.SSHServerAddress != modelConnectionConfig.SSHServerAddress ||
				beanConnectionConfig.SSHTunnelConfig.SSHUsername != modelConnectionConfig.SSHUsername || beanConnectionConfig.SSHTunnelConfig.SSHPassword != modelConnectionConfig.SSHPassword ||
				beanConnectionConfig.SSHTunnelConfig.SSHAuthKey != modelConnectionConfig.SSHAuthKey)))
	return hasConnectionConfigChanged
}

func (impl *ClusterServiceImpl) UpdateVirtualCluster(bean *bean.VirtualClusterBean, userId int32) (*bean.VirtualClusterBean, error) {
	existingModel, err := impl.clusterRepository.FindOne(bean.ClusterName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error(err)
		return nil, err
	}
	if existingModel.Id > 0 && bean.Id != existingModel.Id {
		impl.logger.Errorw("error on fetching cluster, duplicate", "name", bean.ClusterName)
		return nil, fmt.Errorf("cluster already exists")
	}
	model, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	model.ClusterName = bean.ClusterName

	// initiate DB transaction
	dbConnection := impl.clusterRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.clusterRepository.Update(model, tx)
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()
	if err != nil {
		impl.logger.Errorw("error in updating cluster", "err", err)
		return bean, err
	}

	// now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return nil, err
}

func (impl *ClusterServiceImpl) SyncNsInformer(bean *bean.ClusterBean) {
	requestConfig := bean.Config[k8s2.BearerToken]
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
		clusterInfo.KeyData = bean.Config[k8s2.TlsKey]
		clusterInfo.CertData = bean.Config[k8s2.CertData]
		clusterInfo.CAData = bean.Config[k8s2.CertificateAuthorityData]
	}
	beanConnectionConfig := bean.ServerConnectionConfig
	if bean.ServerConnectionConfig != nil {
		connectionConfig := &serverConnectionBean.RemoteConnectionConfigBean{
			RemoteConnectionConfigId: beanConnectionConfig.RemoteConnectionConfigId,
			ConnectionMethod:         beanConnectionConfig.ConnectionMethod,
		}
		if beanConnectionConfig.ProxyConfig != nil {
			connectionConfig.ProxyConfig = &serverConnectionBean.ProxyConfig{
				ProxyUrl: beanConnectionConfig.ProxyConfig.ProxyUrl,
			}
		}
		if beanConnectionConfig.SSHTunnelConfig != nil {
			connectionConfig.SSHTunnelConfig = &serverConnectionBean.SSHTunnelConfig{
				SSHServerAddress: beanConnectionConfig.SSHTunnelConfig.SSHServerAddress,
				SSHUsername:      beanConnectionConfig.SSHTunnelConfig.SSHUsername,
				SSHPassword:      beanConnectionConfig.SSHTunnelConfig.SSHPassword,
				SSHAuthKey:       beanConnectionConfig.SSHTunnelConfig.SSHAuthKey,
			}
		}
		clusterInfo.ServerConnectionConfig = connectionConfig
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
		m = *adapter.ConvertClusterToNewCluster(&m)
		beans = append(beans, bean.ClusterBean{
			Id:                m.Id,
			ClusterName:       m.ClusterName,
			ErrorInConnecting: m.ErrorInConnecting,
			IsCdArgoSetup:     m.CdArgoSetup,
			IsVirtualCluster:  m.IsVirtualCluster,
		})
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) CreateGrafanaDataSource(clusterBean *bean.ClusterBean, env *repository.Environment) (int, error) {
	impl.logger.Errorw("CreateGrafanaDataSource not inplementd in ClusterServiceImpl")
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
		model = *adapter.ConvertClusterToNewCluster(&model)
		if !model.IsVirtualCluster {
			bearerToken := model.Config[k8s2.BearerToken]
			connectionConfig := &serverConnectionBean.RemoteConnectionConfigBean{
				RemoteConnectionConfigId: model.ServerConnectionConfigId,
			}
			if model.ServerConnectionConfig != nil {
				connectionConfig.ConnectionMethod = model.ServerConnectionConfig.ConnectionMethod
				connectionConfig.ProxyConfig = &serverConnectionBean.ProxyConfig{
					ProxyUrl: model.ServerConnectionConfig.ProxyUrl,
				}
				connectionConfig.SSHTunnelConfig = &serverConnectionBean.SSHTunnelConfig{
					SSHServerAddress: model.ServerConnectionConfig.SSHServerAddress,
					SSHUsername:      model.ServerConnectionConfig.SSHUsername,
					SSHPassword:      model.ServerConnectionConfig.SSHPassword,
					SSHAuthKey:       model.ServerConnectionConfig.SSHAuthKey,
				}
			}
			clusterInfo = append(clusterInfo, &bean2.ClusterInfo{
				ClusterId:              model.Id,
				ClusterName:            model.ClusterName,
				BearerToken:            bearerToken,
				ServerUrl:              model.ServerUrl,
				InsecureSkipTLSVerify:  model.InsecureSkipTlsVerify,
				KeyData:                model.Config[k8s2.TlsKey],
				CertData:               model.Config[k8s2.CertData],
				CAData:                 model.Config[k8s2.CertificateAuthorityData],
				ServerConnectionConfig: connectionConfig,
			})
		}
	}
	impl.K8sInformerFactory.BuildInformer(clusterInfo)
}

func (impl ClusterServiceImpl) DeleteFromDb(bean *bean.ClusterBean, userId int32) error {
	existingCluster, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", bean.Id)
		return err
	}
	deleteReq := existingCluster
	deleteReq.UpdatedOn = time.Now()
	deleteReq.UpdatedBy = userId
	err = impl.clusterRepository.MarkClusterDeleted(deleteReq)
	if err != nil {
		impl.logger.Errorw("error in deleting cluster", "id", bean.Id, "err", err)
		return err
	}
	k8sClient, err := impl.K8sUtil.GetCoreV1ClientInCluster()
	if err != nil {
		impl.logger.Errorw("error in getting in cluster k8s client", "err", err, "clusterName", bean.ClusterName)
		return nil
	}
	secretName := fmt.Sprintf("%s-%v", SECRET_NAME, bean.Id)
	err = impl.K8sUtil.DeleteSecret(DEFAULT_NAMESPACE, secretName, k8sClient)
	impl.logger.Errorw("error in deleting secret", "error", err)
	return nil
}

func (impl ClusterServiceImpl) DeleteVirtualClusterFromDb(bean *bean.VirtualClusterBean, userId int32) error {
	existingCluster, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", bean.Id)
		return err
	}
	existingCluster.UpdatedBy = userId
	existingCluster.UpdatedOn = time.Now()
	err = impl.clusterRepository.MarkClusterDeleted(existingCluster)
	if err != nil {
		impl.logger.Errorw("error in deleting virtual cluster", "err", err)
		return err
	}
	return nil
}

func (impl ClusterServiceImpl) CheckIfConfigIsValidAndGetServerVersion(cluster *bean.ClusterBean) (*version.Info, error) {
	clusterConfig := cluster.GetClusterConfig()
	//setting flag toConnectForVerification True
	clusterConfig.ToConnectForClusterVerification = true
	defer impl.K8sUtil.CleanupForClusterUsedForVerification(clusterConfig)
	response, err := impl.K8sUtil.DiscoveryClientGetLiveZCall(clusterConfig)
	impl.logger.Debugw("DiscoveryClientGetLiveZCall call completed", "response", response, "err", err)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			return nil, fmt.Errorf("Incorrect server url : %v", err)
		} else if statusError, ok := err.(*errors.StatusError); ok {
			if statusError != nil {
				errReason := statusError.ErrStatus.Reason
				var errMsg string
				if errReason == v1.StatusReasonUnauthorized {
					errMsg = "token seems invalid or does not have sufficient permissions"
				} else {
					errMsg = statusError.ErrStatus.Message
				}
				return nil, fmt.Errorf("%s : %s", errReason, errMsg)
			} else {
				return nil, fmt.Errorf("Validation failed : %v", err)
			}
		} else {
			return nil, fmt.Errorf("Validation failed : %v", err)
		}
	} else if err == nil && string(response) != "ok" {
		return nil, fmt.Errorf("Validation failed with response : %s", string(response))
	}
	client, err := impl.K8sUtil.GetK8sDiscoveryClient(clusterConfig)
	if err != nil {
		return nil, err
	}
	k8sServerVersion, err := client.ServerVersion()
	if err != nil {
		return nil, err
	}
	return k8sServerVersion, nil
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

func (impl *ClusterServiceImpl) FindAllNamespacesByUserIdAndClusterId(userId int32, clusterId int, isActionUserSuperAdmin bool, token string) ([]string, error) {
	result := make([]string, 0)
	clusterBean, err := impl.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("failed to find cluster for id", "error", err, "clusterId", clusterId)
		return nil, err
	}
	namespaceListGroupByCLuster := impl.K8sInformerFactory.GetLatestNamespaceListGroupByCLuster()
	namespaces := namespaceListGroupByCLuster[clusterBean.ClusterName]

	if isActionUserSuperAdmin {
		for namespace, value := range namespaces {
			if value {
				result = append(result, namespace)
			}
		}
	} else {
		roles, err := impl.FetchRolesFromGroup(userId, token)
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

func (impl *ClusterServiceImpl) FindAllForClusterByUserId(userId int32, isActionUserSuperAdmin bool, token string) ([]bean.ClusterBean, error) {
	if isActionUserSuperAdmin {
		return impl.FindAllForAutoComplete()
	}
	allowedClustersMap := make(map[string]bool)
	roles, err := impl.FetchRolesFromGroup(userId, token)
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

func (impl *ClusterServiceImpl) FetchRolesFromGroup(userId int32, token string) ([]*repository2.RoleModel, error) {
	user, err := impl.userRepository.GetByIdIncludeDeleted(userId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	isDevtronSystemActive := impl.globalAuthorisationConfigService.IsDevtronSystemManagedConfigActive()
	var groups []string
	if isDevtronSystemActive || util3.CheckIfAdminOrApiToken(user.EmailId) {
		groupsCasbin, err := casbin2.GetRolesForUser(user.EmailId)
		if err != nil {
			impl.logger.Errorw("No Roles Found for user", "id", user.Id)
			return nil, err
		}
		groups = append(groups, groupsCasbin...)
	}

	if isGroupClaimsActive {
		_, groupClaims, err := impl.ClusterRbacServiceImpl.userService.GetEmailAndGroupClaimsFromToken(token)
		if err != nil {
			impl.logger.Errorw("error in GetEmailAndGroupClaimsFromToken", "err", err)
			return nil, err
		}
		groupsCasbinNames := util3.GetGroupCasbinName(groupClaims)

		groups = append(groups, groupsCasbinNames...)
	}

	roleEntity := "cluster"
	roles, err := impl.userAuthRepository.GetRolesByUserIdAndEntityType(userId, roleEntity)
	if err != nil {
		impl.logger.Errorw("error on fetching user roles for cluster list", "err", err)
		return nil, err
	}
	if len(groups) > 0 {
		rolesFromGroup, err := impl.roleGroupRepository.GetRolesByGroupNamesAndEntity(groups, roleEntity)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting roles by group names", "err", err)
			return nil, err
		}
		if len(rolesFromGroup) > 0 {
			roles = append(roles, rolesFromGroup...)
		}
	}
	return roles, nil
}

func (impl *ClusterServiceImpl) ConnectClustersInBatch(clusters []*bean.ClusterBean, clusterExistInDb bool) {
	var wg sync.WaitGroup
	respMap := make(map[int]error)
	mutex := &sync.Mutex{}

	for idx, cluster := range clusters {
		wg.Add(1)
		go func(idx int, cluster *bean.ClusterBean) {
			defer wg.Done()
			clusterConfig := cluster.GetClusterConfig()
			_, _, k8sClientSet, err := impl.K8sUtil.GetK8sConfigAndClients(clusterConfig)
			if err != nil {
				mutex.Lock()
				respMap[cluster.Id] = err
				mutex.Unlock()
				return
			}

			id := cluster.Id
			if !clusterExistInDb {
				id = idx
			}
			impl.GetAndUpdateConnectionStatusForOneCluster(k8sClientSet, id, respMap, mutex)
		}(idx, cluster)
	}

	wg.Wait()
	impl.HandleErrorInClusterConnections(clusters, respMap, clusterExistInDb)
}

func (impl *ClusterServiceImpl) HandleErrorInClusterConnections(clusters []*bean.ClusterBean, respMap map[int]error, clusterExistInDb bool) {
	for id, err := range respMap {
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
	}
}

func (impl *ClusterServiceImpl) ValidateKubeconfig(kubeConfig string) (map[string]*bean.ValidateClusterBean, error) {

	kubeConfigObject := api.Config{}

	gvk := &schema.GroupVersionKind{}

	var kubeConfigDataMap map[string]interface{}
	err := json.Unmarshal([]byte(kubeConfig), &kubeConfigDataMap)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling kubeConfig")
		return nil, errors1.New("invalid kubeConfig found , " + err.Error())
	}

	if kubeConfigDataMap["apiVersion"] == nil {
		impl.logger.Errorw("api version missing from kubeConfig")
		return nil, errors1.New("api version missing from kubeConfig")
	}
	if kubeConfigDataMap["kind"] == nil {
		impl.logger.Errorw("kind missing from kubeConfig")
		return nil, errors1.New("kind missing from kubeConfig")
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

		if clusterBeanObject.ClusterName == DEFAULT_CLUSTER {
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
		Config[k8s2.BearerToken] = userInfoObj.Token

		if clusterObj != nil {
			clusterBeanObject.InsecureSkipTLSVerify = clusterObj.InsecureSkipTLSVerify
			clusterBeanObject.ServerConnectionConfig = &serverConnectionBean.RemoteConnectionConfigBean{
				ProxyConfig: &serverConnectionBean.ProxyConfig{
					ProxyUrl: clusterObj.ProxyURL,
				},
			}
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
				Config[k8s2.TlsKey] = string(userInfoObj.ClientKeyData)
				Config[k8s2.CertData] = string(userInfoObj.ClientCertificateData)
				Config[k8s2.CertificateAuthorityData] = string(clusterObj.CertificateAuthorityData)
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
		return nil, errors1.New("No valid cluster object provided in kubeconfig for context")
	} else {
		return ValidateObjects, nil
	}

}

func (impl *ClusterServiceImpl) GetAndUpdateConnectionStatusForOneCluster(k8sClientSet *kubernetes.Clientset, clusterId int, respMap map[int]error, mutex *sync.Mutex) {
	response, err := impl.K8sUtil.GetLiveZCall(k8s2.LiveZ, k8sClientSet)
	log.Println("received response for cluster livez status", "response", string(response), "err", err, "clusterId", clusterId)

	if err != nil {
		if _, ok := err.(*url.Error); ok {
			err = fmt.Errorf("Incorrect server url : %v", err)
		} else if statusError, ok := err.(*errors.StatusError); ok {
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
	mutex.Lock()
	respMap[clusterId] = err
	mutex.Unlock()
}

func (impl ClusterServiceImpl) ConvertClusterBeanObjectToCluster(bean *bean.ClusterBean) *v1alpha1.Cluster {
	configMap := bean.Config
	serverUrl := bean.ServerUrl
	bearerToken := ""
	if configMap[k8s2.BearerToken] != "" {
		bearerToken = configMap[k8s2.BearerToken]
	}
	tlsConfig := v1alpha1.TLSClientConfig{
		Insecure: bean.InsecureSkipTLSVerify,
	}

	if !bean.InsecureSkipTLSVerify {
		tlsConfig.KeyData = []byte(bean.Config[k8s2.TlsKey])
		tlsConfig.CertData = []byte(bean.Config[k8s2.CertData])
		tlsConfig.CAData = []byte(bean.Config[k8s2.CertificateAuthorityData])
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

func (impl ClusterServiceImpl) GetClusterConfigByClusterId(clusterId int) (*k8s2.ClusterConfig, error) {
	clusterBean, err := impl.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting clusterBean by cluster id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	rq := *clusterBean
	clusterConfig := rq.GetClusterConfig()
	return clusterConfig, nil
}

func (impl ClusterServiceImpl) IsPolicyConfiguredForCluster(envId, clusterId int) (bool, error) {
	// this implementation is used in hyperion mode, so IsPolicyConfiguredForCluster is always false
	return false, nil
}
