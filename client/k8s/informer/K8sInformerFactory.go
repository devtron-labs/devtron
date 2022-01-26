package informer

import (
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sync"
	"time"
)

func NewGlobalMapClusterNamespace() map[string]map[string]string {
	globalMapClusterNamespace := make(map[string]map[string]string)
	return globalMapClusterNamespace
}

type K8sInformerFactoryImpl struct {
	logger                    *zap.SugaredLogger
	clusterService            repository.ClusterRepository
	globalMapClusterNamespace map[string]map[string]string
}
type K8sInformerFactory interface {
	GetLatestNamespaceListGroupByCLuster() map[string]map[string]string
}

func NewK8sInformerFactoryImpl(logger *zap.SugaredLogger, clusterService repository.ClusterRepository, globalMapClusterNamespace map[string]map[string]string) *K8sInformerFactoryImpl {
	informerFactory := &K8sInformerFactoryImpl{
		logger:                    logger,
		clusterService:            clusterService,
		globalMapClusterNamespace: globalMapClusterNamespace,
	}
	informerFactory.buildInformer()
	return informerFactory
}

func (impl *K8sInformerFactoryImpl) GetLatestNamespaceListGroupByCLuster() map[string]map[string]string {
	copiedClusterNamespaces := make(map[string]map[string]string)
	for key, value := range impl.globalMapClusterNamespace {
		for _, namespace := range value {
			if _, ok := copiedClusterNamespaces[key]; !ok {
				allNamespaces := make(map[string]string)
				allNamespaces[namespace] =namespace
				copiedClusterNamespaces[key]=allNamespaces
			} else {
				allNamespaces := copiedClusterNamespaces[key]
				allNamespaces[namespace] =namespace
				copiedClusterNamespaces[key]=allNamespaces
			}
		}
	}
	return copiedClusterNamespaces
}

func (impl *K8sInformerFactoryImpl) buildInformer() map[string]map[string]string {
	models, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
		return impl.globalMapClusterNamespace
	}
	var mutex sync.Mutex
	for _, clusterBean := range models {
		bearerToken := clusterBean.Config["bearer_token"]
		c := &rest.Config{
			Host:            clusterBean.ServerUrl,
			BearerToken:     bearerToken,
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		}
		impl.buildInformerAndNamespaceList(clusterBean.ClusterName, c, &mutex)
	}

	//TODO - if cluster added or deleted, manage informer respectively
	return impl.globalMapClusterNamespace
}

func (impl *K8sInformerFactoryImpl) buildInformerAndNamespaceList(clusterName string, config *rest.Config, mutex *sync.Mutex) map[string]map[string]string {
	clusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return impl.globalMapClusterNamespace
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, time.Minute)
	nsInformenr := informerFactory.Core().V1().Namespaces()
	impl.logger.Debugw(clusterName, "ns informer", nsInformenr.Informer())
	nsInformenr.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debugw("add event handler", "cluster", clusterName, "object", obj.(metav1.Object))
			if mobject, ok := obj.(metav1.Object); ok {
				mutex.Lock()
				if _, ok := impl.globalMapClusterNamespace[clusterName]; !ok {
					allNamespaces := make(map[string]string)
					allNamespaces[mobject.GetName()] = mobject.GetName()
					impl.globalMapClusterNamespace[clusterName] = allNamespaces
				} else {
					allNamespaces := impl.globalMapClusterNamespace[clusterName]
					allNamespaces[mobject.GetName()] = mobject.GetName()
					impl.globalMapClusterNamespace[clusterName] = allNamespaces
				}
				mutex.Unlock()
			}
		},
		DeleteFunc: func(obj interface{}) {
			if mobject, ok := obj.(metav1.Object); ok {
				if _, ok := impl.globalMapClusterNamespace[clusterName]; ok {
					mutex.Lock()
					allNamespaces := impl.globalMapClusterNamespace[clusterName]
					delete(allNamespaces, mobject.GetName())
					impl.globalMapClusterNamespace[clusterName] = allNamespaces
					mutex.Unlock()
				}
			}
		},
	})
	informerFactory.Start(wait.NeverStop)
	return impl.globalMapClusterNamespace
}
