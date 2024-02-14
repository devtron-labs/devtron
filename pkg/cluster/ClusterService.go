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
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/common-lib/utils/k8s"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	repository3 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	errors1 "github.com/juju/errors"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
)

type PrometheusAuth struct {
	UserName      string `json:"userName,omitempty"`
	Password      string `json:"password,omitempty"`
	TlsClientCert string `json:"tlsClientCert,omitempty"`
	TlsClientKey  string `json:"tlsClientKey,omitempty"`
	IsAnonymous   bool   `json:"isAnonymous"`
}

type ClusterBean struct {
	Id                      int                        `json:"id" validate:"number"`
	ClusterName             string                     `json:"cluster_name,omitempty" validate:"required"`
	Description             string                     `json:"description"`
	ServerUrl               string                     `json:"server_url,omitempty" validate:"url,required"`
	PrometheusUrl           string                     `json:"prometheus_url,omitempty" validate:"validate-non-empty-url"`
	Active                  bool                       `json:"active"`
	Config                  map[string]string          `json:"config,omitempty"`
	PrometheusAuth          *PrometheusAuth            `json:"prometheusAuth,omitempty"`
	DefaultClusterComponent []*DefaultClusterComponent `json:"defaultClusterComponent"`
	AgentInstallationStage  int                        `json:"agentInstallationStage,notnull"` // -1=external, 0=not triggered, 1=progressing, 2=success, 3=fails
	K8sVersion              string                     `json:"k8sVersion"`
	HasConfigOrUrlChanged   bool                       `json:"-"`
	UserName                string                     `json:"userName,omitempty"`
	InsecureSkipTLSVerify   bool                       `json:"insecureSkipTlsVerify"`
	ErrorInConnecting       string                     `json:"errorInConnecting"`
	IsCdArgoSetup           bool                       `json:"isCdArgoSetup"`
	IsVirtualCluster        bool                       `json:"isVirtualCluster"`
	isClusterNameEmpty      bool                       `json:"-"`
	ClusterUpdated          bool                       `json:"clusterUpdated"`
}

func GetClusterBean(model repository.Cluster) ClusterBean {
	bean := ClusterBean{}
	bean.Id = model.Id
	bean.ClusterName = model.ClusterName
	//bean.Note = model.Note
	bean.ServerUrl = model.ServerUrl
	bean.PrometheusUrl = model.PrometheusEndpoint
	bean.AgentInstallationStage = model.AgentInstallationStage
	bean.Active = model.Active
	bean.Config = model.Config
	bean.K8sVersion = model.K8sVersion
	bean.InsecureSkipTLSVerify = model.InsecureSkipTlsVerify
	bean.IsVirtualCluster = model.IsVirtualCluster
	bean.ErrorInConnecting = model.ErrorInConnecting
	bean.PrometheusAuth = &PrometheusAuth{
		UserName:      model.PUserName,
		Password:      model.PPassword,
		TlsClientCert: model.PTlsClientCert,
		TlsClientKey:  model.PTlsClientKey,
	}
	return bean
}

func (bean ClusterBean) GetClusterConfig() (*k8s.ClusterConfig, error) {
	host := bean.ServerUrl
	configMap := bean.Config
	bearerToken := configMap[k8s.BearerToken]
	clusterCfg := &k8s.ClusterConfig{Host: host, BearerToken: bearerToken}
	clusterCfg.InsecureSkipTLSVerify = bean.InsecureSkipTLSVerify
	if bean.InsecureSkipTLSVerify == false {
		clusterCfg.KeyData = configMap[k8s.TlsKey]
		clusterCfg.CertData = configMap[k8s.CertData]
		clusterCfg.CAData = configMap[k8s.CertificateAuthorityData]
	}
	return clusterCfg, nil
}

type UserInfo struct {
	UserName          string            `json:"userName,omitempty"`
	Config            map[string]string `json:"config,omitempty"`
	ErrorInConnecting string            `json:"errorInConnecting"`
}

type ValidateClusterBean struct {
	UserInfos map[string]*UserInfo `json:"userInfos,omitempty""`
	*ClusterBean
}

type UserClusterBeanMapping struct {
	Mapping map[string]*ClusterBean `json:"mapping"`
}

type Kubeconfig struct {
	Config string `json:"config"`
}

type DefaultClusterComponent struct {
	ComponentName  string `json:"name"`
	AppId          int    `json:"appId"`
	InstalledAppId int    `json:"installedAppId,omitempty"`
	EnvId          int    `json:"envId"`
	EnvName        string `json:"envName"`
	Status         string `json:"status"`
}

type ClusterService interface {
	Save(parent context.Context, bean *ClusterBean, userId int32) (*ClusterBean, error)
	UpdateClusterDescription(bean *ClusterBean, userId int32) error
	ValidateKubeconfig(kubeConfig string) (map[string]*ValidateClusterBean, error)
	FindOne(clusterName string) (*ClusterBean, error)
	FindOneActive(clusterName string) (*ClusterBean, error)
	FindAll() ([]*ClusterBean, error)
	FindAllExceptVirtual() ([]*ClusterBean, error)
	FindAllWithoutConfig() ([]*ClusterBean, error)
	FindAllActive() ([]ClusterBean, error)
	DeleteFromDb(bean *ClusterBean, userId int32) error

	FindById(id int) (*ClusterBean, error)
	FindByIdWithoutConfig(id int) (*ClusterBean, error)
	FindByIds(id []int) ([]ClusterBean, error)
	Update(ctx context.Context, bean *ClusterBean, userId int32) (*ClusterBean, error)
	Delete(bean *ClusterBean, userId int32) error

	FindAllForAutoComplete() ([]ClusterBean, error)
	CreateGrafanaDataSource(clusterBean *ClusterBean, env *repository.Environment) (int, error)
	GetAllClusterNamespaces() map[string][]string
	FindAllNamespacesByUserIdAndClusterId(userId int32, clusterId int, isActionUserSuperAdmin bool) ([]string, error)
	FindAllForClusterByUserId(userId int32, isActionUserSuperAdmin bool) ([]ClusterBean, error)
	FetchRolesFromGroup(userId int32) ([]*repository3.RoleModel, error)
	HandleErrorInClusterConnections(clusters []*ClusterBean, respMap map[int]error, clusterExistInDb bool)
	ConnectClustersInBatch(clusters []*ClusterBean, clusterExistInDb bool)
	ConvertClusterBeanToCluster(clusterBean *ClusterBean, userId int32) *repository.Cluster
	ConvertClusterBeanObjectToCluster(bean *ClusterBean) *v1alpha1.Cluster

	GetClusterConfigByClusterId(clusterId int) (*k8s.ClusterConfig, error)
	IsClusterReachable(clusterId int) (bool, error)
}

type ClusterServiceImpl struct {
	clusterRepository   repository.ClusterRepository
	logger              *zap.SugaredLogger
	K8sUtil             *k8s.K8sServiceImpl
	K8sInformerFactory  informer.K8sInformerFactory
	userAuthRepository  repository3.UserAuthRepository
	userRepository      repository3.UserRepository
	roleGroupRepository repository3.RoleGroupRepository
	*ClusterRbacServiceImpl
}

func NewClusterServiceImpl(repository repository.ClusterRepository, logger *zap.SugaredLogger,
	K8sUtil *k8s.K8sServiceImpl, K8sInformerFactory informer.K8sInformerFactory,
	userAuthRepository repository3.UserAuthRepository, userRepository repository3.UserRepository,
	roleGroupRepository repository3.RoleGroupRepository) *ClusterServiceImpl {
	clusterService := &ClusterServiceImpl{
		clusterRepository:   repository,
		logger:              logger,
		K8sUtil:             K8sUtil,
		K8sInformerFactory:  K8sInformerFactory,
		userAuthRepository:  userAuthRepository,
		userRepository:      userRepository,
		roleGroupRepository: roleGroupRepository,
		ClusterRbacServiceImpl: &ClusterRbacServiceImpl{
			logger: logger,
		},
	}
	go clusterService.buildInformer()
	return clusterService
}

func (impl *ClusterServiceImpl) ConvertClusterBeanToCluster(clusterBean *ClusterBean, userId int32) *repository.Cluster {

	model := &repository.Cluster{}

	model.ClusterName = clusterBean.ClusterName
	//model.Note = clusterBean.Note
	model.Active = true
	model.ServerUrl = clusterBean.ServerUrl
	model.Config = clusterBean.Config
	model.PrometheusEndpoint = clusterBean.PrometheusUrl
	model.InsecureSkipTlsVerify = clusterBean.InsecureSkipTLSVerify

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

func (impl *ClusterServiceImpl) Save(parent context.Context, bean *ClusterBean, userId int32) (*ClusterBean, error) {
	//validating config

	err := impl.CheckIfConfigIsValid(bean)

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

	model := impl.ConvertClusterBeanToCluster(bean, userId)

	cfg, err := bean.GetClusterConfig()
	if err != nil {
		return nil, err
	}
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

// UpdateClusterDescription is new api service logic to only update description, this should be done in cluster update operation only
// but not supported currently as per product
func (impl *ClusterServiceImpl) UpdateClusterDescription(bean *ClusterBean, userId int32) error {
	//updating description as other fields are not supported yet
	err := impl.clusterRepository.SetDescription(bean.Id, bean.Description, userId)
	if err != nil {
		impl.logger.Errorw("error in setting cluster description", "err", err, "clusterId", bean.Id)
		return err
	}
	return nil
}

func (impl *ClusterServiceImpl) FindOne(clusterName string) (*ClusterBean, error) {
	model, err := impl.clusterRepository.FindOne(clusterName)
	if err != nil {
		return nil, err
	}
	bean := GetClusterBean(*model)
	return &bean, nil
}

func (impl *ClusterServiceImpl) FindOneActive(clusterName string) (*ClusterBean, error) {
	model, err := impl.clusterRepository.FindOneActive(clusterName)
	if err != nil {
		return nil, err
	}
	bean := GetClusterBean(*model)
	return &bean, nil

}

func (impl *ClusterServiceImpl) FindAllWithoutConfig() ([]*ClusterBean, error) {
	models, err := impl.FindAll()
	if err != nil {
		return nil, err
	}
	for _, model := range models {
		model.Config = map[string]string{k8s.BearerToken: ""}
	}
	return models, nil
}

func (impl *ClusterServiceImpl) FindAll() ([]*ClusterBean, error) {
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		return nil, err
	}
	var beans []*ClusterBean
	for _, model := range models {
		bean := GetClusterBean(model)
		beans = append(beans, &bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FindAllExceptVirtual() ([]*ClusterBean, error) {
	models, err := impl.clusterRepository.FindAllActiveExceptVirtual()
	if err != nil {
		return nil, err
	}
	var beans []*ClusterBean
	for _, model := range models {
		bean := GetClusterBean(model)
		beans = append(beans, &bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FindAllActive() ([]ClusterBean, error) {
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		return nil, err
	}
	var beans []ClusterBean
	for _, model := range models {
		bean := GetClusterBean(model)
		beans = append(beans, bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) FindById(id int) (*ClusterBean, error) {
	model, err := impl.clusterRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	bean := GetClusterBean(*model)
	return &bean, nil
}

func (impl *ClusterServiceImpl) FindByIdWithoutConfig(id int) (*ClusterBean, error) {
	model, err := impl.FindById(id)
	if err != nil {
		return nil, err
	}
	//empty bearer token as it will be hidden for user
	model.Config = map[string]string{k8s.BearerToken: ""}
	return model, nil
}

func (impl *ClusterServiceImpl) FindByIds(ids []int) ([]ClusterBean, error) {
	models, err := impl.clusterRepository.FindByIds(ids)
	if err != nil {
		return nil, err
	}
	var beans []ClusterBean

	for _, model := range models {
		bean := GetClusterBean(model)
		beans = append(beans, bean)
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) Update(ctx context.Context, bean *ClusterBean, userId int32) (*ClusterBean, error) {
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
	dbConfigBearerToken := model.Config[k8s.BearerToken]
	requestConfigBearerToken := bean.Config[k8s.BearerToken]
	if len(requestConfigBearerToken) == 0 {
		bean.Config[k8s.BearerToken] = model.Config[k8s.BearerToken]
	}

	dbConfigTlsKey := model.Config[k8s.TlsKey]
	requestConfigTlsKey := bean.Config[k8s.TlsKey]
	if len(requestConfigTlsKey) == 0 {
		bean.Config[k8s.TlsKey] = model.Config[k8s.TlsKey]
	}

	dbConfigCertData := model.Config[k8s.CertData]
	requestConfigCertData := bean.Config[k8s.CertData]
	if len(requestConfigCertData) == 0 {
		bean.Config[k8s.CertData] = model.Config[k8s.CertData]
	}

	dbConfigCAData := model.Config[k8s.CertificateAuthorityData]
	requestConfigCAData := bean.Config[k8s.CertificateAuthorityData]
	if len(requestConfigCAData) == 0 {
		bean.Config[k8s.CertificateAuthorityData] = model.Config[k8s.CertificateAuthorityData]
	}

	if bean.ServerUrl != model.ServerUrl || bean.InsecureSkipTLSVerify != model.InsecureSkipTlsVerify || dbConfigBearerToken != requestConfigBearerToken || dbConfigTlsKey != requestConfigTlsKey || dbConfigCertData != requestConfigCertData || dbConfigCAData != requestConfigCAData {
		if bean.ClusterName == DEFAULT_CLUSTER {
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
		cfg, err := bean.GetClusterConfig()
		if err != nil {
			return nil, err
		}
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

func (impl *ClusterServiceImpl) SyncNsInformer(bean *ClusterBean) {
	requestConfig := bean.Config[k8s.BearerToken]
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
		clusterInfo.KeyData = bean.Config[k8s.TlsKey]
		clusterInfo.CertData = bean.Config[k8s.CertData]
		clusterInfo.CAData = bean.Config[k8s.CertificateAuthorityData]
	}
	impl.K8sInformerFactory.BuildInformer([]*bean2.ClusterInfo{clusterInfo})
}

func (impl *ClusterServiceImpl) Delete(bean *ClusterBean, userId int32) error {
	model, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		return err
	}
	return impl.clusterRepository.Delete(model)
}

func (impl *ClusterServiceImpl) FindAllForAutoComplete() ([]ClusterBean, error) {
	model, err := impl.clusterRepository.FindAll()
	if err != nil {
		return nil, err
	}
	var beans []ClusterBean
	for _, m := range model {
		beans = append(beans, ClusterBean{
			Id:                m.Id,
			ClusterName:       m.ClusterName,
			ErrorInConnecting: m.ErrorInConnecting,
			IsCdArgoSetup:     m.CdArgoSetup,
			IsVirtualCluster:  m.IsVirtualCluster,
		})
	}
	return beans, nil
}

func (impl *ClusterServiceImpl) CreateGrafanaDataSource(clusterBean *ClusterBean, env *repository.Environment) (int, error) {
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
		if !model.IsVirtualCluster {
			bearerToken := model.Config[k8s.BearerToken]
			clusterInfo = append(clusterInfo, &bean2.ClusterInfo{
				ClusterId:             model.Id,
				ClusterName:           model.ClusterName,
				BearerToken:           bearerToken,
				ServerUrl:             model.ServerUrl,
				InsecureSkipTLSVerify: model.InsecureSkipTlsVerify,
				KeyData:               model.Config[k8s.TlsKey],
				CertData:              model.Config[k8s.CertData],
				CAData:                model.Config[k8s.CertificateAuthorityData],
			})
		}
	}
	impl.K8sInformerFactory.BuildInformer(clusterInfo)
}

func (impl ClusterServiceImpl) DeleteFromDb(bean *ClusterBean, userId int32) error {
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

func (impl ClusterServiceImpl) CheckIfConfigIsValid(cluster *ClusterBean) error {
	clusterConfig, err := cluster.GetClusterConfig()
	if err != nil {
		impl.logger.Errorw("error in getting cluster config ", "err", "err", "clusterId", cluster.Id)
		return err
	}
	response, err := impl.K8sUtil.DiscoveryClientGetLiveZCall(clusterConfig)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			return fmt.Errorf("Incorrect server url : %v", err)
		} else if statusError, ok := err.(*errors.StatusError); ok {
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

func (impl *ClusterServiceImpl) FindAllNamespacesByUserIdAndClusterId(userId int32, clusterId int, isActionUserSuperAdmin bool) ([]string, error) {
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

func (impl *ClusterServiceImpl) FindAllForClusterByUserId(userId int32, isActionUserSuperAdmin bool) ([]ClusterBean, error) {
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
	var beans []ClusterBean
	for _, model := range models {
		if _, ok := allowedClustersMap[model.ClusterName]; ok {
			beans = append(beans, ClusterBean{
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

func (impl *ClusterServiceImpl) ConnectClustersInBatch(clusters []*ClusterBean, clusterExistInDb bool) {
	var wg sync.WaitGroup
	respMap := make(map[int]error)
	mutex := &sync.Mutex{}

	for idx, cluster := range clusters {
		wg.Add(1)
		go func(idx int, cluster *ClusterBean) {
			defer wg.Done()
			clusterConfig, err := cluster.GetClusterConfig()
			if err != nil {
				impl.logger.Errorw("error in getting cluster config", "err", err, "clusterId", cluster.Id)
				return
			}
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

func (impl *ClusterServiceImpl) HandleErrorInClusterConnections(clusters []*ClusterBean, respMap map[int]error, clusterExistInDb bool) {
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

func (impl *ClusterServiceImpl) ValidateKubeconfig(kubeConfig string) (map[string]*ValidateClusterBean, error) {

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
	ValidateObjects := make(map[string]*ValidateClusterBean)

	var clusterBeansWithNoValidationErrors []*ClusterBean
	var clusterBeanObjects []*ClusterBean
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

	userInfosMap := map[string]*UserInfo{}
	for _, ctx := range kubeConfigObject.Contexts {
		clusterBeanObject := &ClusterBean{}
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
		Config[k8s.BearerToken] = userInfoObj.Token

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
				Config[k8s.TlsKey] = string(userInfoObj.ClientKeyData)
				Config[k8s.CertData] = string(userInfoObj.ClientCertificateData)
				Config[k8s.CertificateAuthorityData] = string(clusterObj.CertificateAuthorityData)
			}
		}

		userInfo := UserInfo{
			UserName:          userName,
			Config:            Config,
			ErrorInConnecting: clusterBeanObject.ErrorInConnecting,
		}

		userInfosMap[userInfo.UserName] = &userInfo
		validateObject := &ValidateClusterBean{}
		if _, ok := ValidateObjects[clusterBeanObject.ClusterName]; !ok {
			validateObject.UserInfos = make(map[string]*UserInfo)
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
	response, err := impl.K8sUtil.GetLiveZCall(k8s.LiveZ, k8sClientSet)
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

func (impl ClusterServiceImpl) ConvertClusterBeanObjectToCluster(bean *ClusterBean) *v1alpha1.Cluster {
	configMap := bean.Config
	serverUrl := bean.ServerUrl
	bearerToken := ""
	if configMap[k8s.BearerToken] != "" {
		bearerToken = configMap[k8s.BearerToken]
	}
	tlsConfig := v1alpha1.TLSClientConfig{
		Insecure: bean.InsecureSkipTLSVerify,
	}

	if !bean.InsecureSkipTLSVerify {
		tlsConfig.KeyData = []byte(bean.Config[k8s.TlsKey])
		tlsConfig.CertData = []byte(bean.Config[k8s.CertData])
		tlsConfig.CAData = []byte(bean.Config[k8s.CertificateAuthorityData])
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

func (impl ClusterServiceImpl) GetClusterConfigByClusterId(clusterId int) (*k8s.ClusterConfig, error) {
	clusterBean, err := impl.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting clusterBean by cluster id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	rq := *clusterBean
	clusterConfig, err := rq.GetClusterConfig()
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "err", err, "clusterId", clusterBean.Id)
		return nil, err
	}
	return clusterConfig, nil
}

func (impl ClusterServiceImpl) IsClusterReachable(clusterId int) (bool, error) {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in finding cluster from clusterId", "envId", clusterId)
		return false, err
	}
	if len(cluster.ErrorInConnecting) > 0 {
		return false, nil
	}
	return true, nil

}
