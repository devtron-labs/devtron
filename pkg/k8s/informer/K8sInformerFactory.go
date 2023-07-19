package informer

import (
	"github.com/devtron-labs/devtron/util/k8s"
	"sync"
	"time"

	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/api/bean"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func NewGlobalMapClusterNamespace() map[string]map[string]bool {
	globalMapClusterNamespace := make(map[string]map[string]bool)
	return globalMapClusterNamespace
}

type K8sInformerFactoryImpl struct {
	logger                    *zap.SugaredLogger
	globalMapClusterNamespace map[string]map[string]bool // {"cluster1":{"ns1":true","ns2":true"}}
	mutex                     sync.Mutex
	informerStopper           map[string]chan struct{}
	runtimeConfig             *client.RuntimeConfig
	k8sUtil                   *k8s.K8sUtil
}

type K8sInformerFactory interface {
	GetLatestNamespaceListGroupByCLuster() map[string]map[string]bool
	BuildInformer(clusterInfo []*bean.ClusterInfo)
	CleanNamespaceInformer(clusterName string)
}

func NewK8sInformerFactoryImpl(logger *zap.SugaredLogger, globalMapClusterNamespace map[string]map[string]bool, runtimeConfig *client.RuntimeConfig, k8sUtil *k8s.K8sUtil) *K8sInformerFactoryImpl {
	informerFactory := &K8sInformerFactoryImpl{
		logger:                    logger,
		globalMapClusterNamespace: globalMapClusterNamespace,
		runtimeConfig:             runtimeConfig,
		k8sUtil:                   k8sUtil,
	}
	informerFactory.informerStopper = make(map[string]chan struct{})
	return informerFactory
}

func (impl *K8sInformerFactoryImpl) GetLatestNamespaceListGroupByCLuster() map[string]map[string]bool {
	copiedClusterNamespaces := make(map[string]map[string]bool)
	for key, value := range impl.globalMapClusterNamespace {
		for namespace, v := range value {
			if _, ok := copiedClusterNamespaces[key]; !ok {
				allNamespaces := make(map[string]bool)
				allNamespaces[namespace] = v
				copiedClusterNamespaces[key] = allNamespaces
			} else {
				allNamespaces := copiedClusterNamespaces[key]
				allNamespaces[namespace] = v
				copiedClusterNamespaces[key] = allNamespaces
			}
		}
	}
	return copiedClusterNamespaces
}

func (impl *K8sInformerFactoryImpl) BuildInformer(clusterInfo []*bean.ClusterInfo) {
	var restConfig *rest.Config
	for _, info := range clusterInfo {
		if info.ClusterName == "default_cluster" {
			clusterConfig, err := impl.k8sUtil.GetInClusterConfig(bool(impl.runtimeConfig.LocalDevMode))
			if err != nil {
				impl.logger.Errorw("error in fetch default cluster config", "err", err, "servername", restConfig.ServerName)
				continue
			}
			restConfig = clusterConfig
			impl.buildInformerAndNamespaceList(info.ClusterName, restConfig, &impl.mutex)
		} else {
			clusterConfig := &k8s.ClusterConfig{
				ClusterName:           info.ClusterName,
				BearerToken:           info.BearerToken,
				Host:                  info.ServerUrl,
				InsecureSkipTLSVerify: info.InsecureSkipTLSVerify,
				KeyData:               info.KeyData,
				CertData:              info.CertData,
				CAData:                info.CAData,
			}
			c, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
			if err != nil {
				impl.logger.Errorw("error in getting rest config from cluster config", "err", err, "clusterName", info.ClusterName)
				return
			}
			impl.buildInformerAndNamespaceList(info.ClusterName, c, &impl.mutex)
		}
	}
	return
}

func (impl *K8sInformerFactoryImpl) buildInformerAndNamespaceList(clusterName string, config *rest.Config, mutex *sync.Mutex) map[string]map[string]bool {
	allNamespaces := make(map[string]bool)
	impl.globalMapClusterNamespace[clusterName] = allNamespaces
	httpClient, err := k8s.OverrideK8sHttpClientWithTracer(config)
	if err != nil {
		return impl.globalMapClusterNamespace
	}
	clusterClient, err := kubernetes.NewForConfigAndClient(config, httpClient)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return impl.globalMapClusterNamespace
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, time.Minute)
	nsInformer := informerFactory.Core().V1().Namespaces()
	stopper := make(chan struct{})
	nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if mobject, ok := obj.(metav1.Object); ok {
				mutex.Lock()
				defer mutex.Unlock()
				if _, ok := impl.globalMapClusterNamespace[clusterName]; !ok {
					allNamespaces := make(map[string]bool)
					allNamespaces[mobject.GetName()] = true
					impl.globalMapClusterNamespace[clusterName] = allNamespaces
				} else {
					allNamespaces := impl.globalMapClusterNamespace[clusterName]
					allNamespaces[mobject.GetName()] = true
					impl.globalMapClusterNamespace[clusterName] = allNamespaces
				}
				//mutex.Unlock()
			}
		},
		DeleteFunc: func(obj interface{}) {
			if object, ok := obj.(metav1.Object); ok {
				mutex.Lock()
				defer mutex.Unlock()
				if _, ok := impl.globalMapClusterNamespace[clusterName]; ok {
					allNamespaces := impl.globalMapClusterNamespace[clusterName]
					delete(allNamespaces, object.GetName())
					impl.globalMapClusterNamespace[clusterName] = allNamespaces
					//mutex.Unlock()
				}
			}
		},
	})
	informerFactory.Start(stopper)
	impl.informerStopper[clusterName] = stopper
	return impl.globalMapClusterNamespace
}

func (impl *K8sInformerFactoryImpl) CleanNamespaceInformer(clusterName string) {
	stopper := impl.informerStopper[clusterName]
	if stopper != nil {
		close(stopper)
		delete(impl.informerStopper, clusterName)
	}
	return
}
