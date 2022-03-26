package informer

import (
	"flag"
	"os/user"
	"path/filepath"
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
	"k8s.io/client-go/tools/clientcmd"
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
}

type K8sInformerFactory interface {
	GetLatestNamespaceListGroupByCLuster() map[string]map[string]bool
	BuildInformer(clusterInfo []*bean.ClusterInfo)
	CleanNamespaceInformer(clusterName string)
}

func NewK8sInformerFactoryImpl(logger *zap.SugaredLogger, globalMapClusterNamespace map[string]map[string]bool, runtimeConfig *client.RuntimeConfig) *K8sInformerFactoryImpl {
	informerFactory := &K8sInformerFactoryImpl{
		logger:                    logger,
		globalMapClusterNamespace: globalMapClusterNamespace,
		runtimeConfig:             runtimeConfig,
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
			if impl.runtimeConfig.LocalDevMode {
				usr, err := user.Current()
				if err != nil {
					impl.logger.Errorw("Error while getting user current env details", "error", err)
				}
				kubeconfig := flag.String("build-informer", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
				flag.Parse()
				restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
				if err != nil {
					impl.logger.Errorw("Error while building config from flags", "error", err)
				}
			} else {
				clusterConfig, err := rest.InClusterConfig()
				if err != nil {
					impl.logger.Errorw("error in fetch default cluster config", "err", err, "servername", restConfig.ServerName)
					continue
				}
				restConfig = clusterConfig
			}

			impl.buildInformerAndNamespaceList(info.ClusterName, restConfig, &impl.mutex)
		} else {
			c := &rest.Config{
				Host:            info.ServerUrl,
				BearerToken:     info.BearerToken,
				TLSClientConfig: rest.TLSClientConfig{Insecure: true},
			}
			impl.buildInformerAndNamespaceList(info.ClusterName, c, &impl.mutex)
		}
	}
	return
}

func (impl *K8sInformerFactoryImpl) buildInformerAndNamespaceList(clusterName string, config *rest.Config, mutex *sync.Mutex) map[string]map[string]bool {
	allNamespaces := make(map[string]bool)
	impl.globalMapClusterNamespace[clusterName] = allNamespaces
	clusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return impl.globalMapClusterNamespace
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, time.Minute)
	nsInformer := informerFactory.Core().V1().Namespaces()
	stopper := make(chan struct{})
	nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debugw("add event handler", "cluster", clusterName, "object", obj.(metav1.Object))
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
