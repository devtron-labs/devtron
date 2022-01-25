package informer

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

var (
//GlobalMapClusterNamespace map[string]map[string]string
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
	return impl.globalMapClusterNamespace
}

func (impl *K8sInformerFactoryImpl) buildInformer() map[string]map[string]string {
	models, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
		return impl.globalMapClusterNamespace
	}
	for _, clusterBean := range models {
		bearerToken := clusterBean.Config["bearer_token"]
		c := &rest.Config{
			Host:            clusterBean.ServerUrl,
			BearerToken:     bearerToken,
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		}
		impl.buildInformerAndNamespaceList(clusterBean.ClusterName, clusterBean.Id, c)
	}

	//TODO - if cluster added or deleted, manage informer respectively
	return impl.globalMapClusterNamespace
}

func (impl *K8sInformerFactoryImpl) buildInformerAndNamespaceList(clusterName string, clusterId int, config *rest.Config) map[string]map[string]string {
	clusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return impl.globalMapClusterNamespace
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, time.Minute)
	nsInformenr := informerFactory.Core().V1().Namespaces()
	impl.logger.Infow(clusterName, "ns informer", nsInformenr.Informer())
	nsInformenr.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Debugw("add event handler", "cluster", clusterName, "object", obj.(metav1.Object))
			if mobject, ok := obj.(metav1.Object); ok {
				clusterKey := fmt.Sprintf("%s__%d", clusterName, clusterId)
				if _, ok := impl.globalMapClusterNamespace[clusterKey]; !ok {
					allNamespaces := make(map[string]string)
					allNamespaces[mobject.GetName()] = mobject.GetName()
					impl.globalMapClusterNamespace[clusterKey] = allNamespaces
				} else {
					allNamespaces := impl.globalMapClusterNamespace[clusterKey]
					allNamespaces[mobject.GetName()] = mobject.GetName()
					impl.globalMapClusterNamespace[clusterKey] = allNamespaces
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if mobject, ok := obj.(metav1.Object); ok {
				clusterKey := fmt.Sprintf("%s__%d", clusterName, clusterId)
				if _, ok := impl.globalMapClusterNamespace[clusterKey]; ok {
					allNamespaces := impl.globalMapClusterNamespace[clusterName]
					delete(allNamespaces, mobject.GetName())
					impl.globalMapClusterNamespace[clusterKey] = allNamespaces
				}
			}
		},
	})
	informerFactory.Start(wait.NeverStop)
	return impl.globalMapClusterNamespace
}
