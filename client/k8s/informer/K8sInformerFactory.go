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
	GlobalMapClusterNamespace map[string]map[string]string
)

type K8sInformerFactoryImpl struct {
	logger         *zap.SugaredLogger
	clusterService repository.ClusterRepository
}
type K8sInformerFactory interface {
	GetLatestNamespaceListGroupByCLuster() map[string]map[string]string
}

func NewK8sInformerFactoryImpl(logger *zap.SugaredLogger, clusterService repository.ClusterRepository) *K8sInformerFactoryImpl {
	GlobalMapClusterNamespace := make(map[string]map[string]string)
	informerFactory := &K8sInformerFactoryImpl{
		logger:         logger,
		clusterService: clusterService,
	}
	informerFactory.logger.Info(GlobalMapClusterNamespace)
	informerFactory.buildInformer()
	informerFactory.logger.Infow(">>>>>", "data", informerFactory.GetLatestNamespaceListGroupByCLuster())
	return informerFactory
}

func (impl *K8sInformerFactoryImpl) GetLatestNamespaceListGroupByCLuster() map[string]map[string]string {
	return GlobalMapClusterNamespace
}

func (impl *K8sInformerFactoryImpl) buildInformer() {
	models, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
	}
	for _, clusterBean := range models {
		bearerToken := clusterBean.Config["bearer_token"]
		c := &rest.Config{
			Host:        clusterBean.ServerUrl,
			BearerToken: bearerToken,
		}
		impl.buildInformerAndNamespaceList(clusterBean.ClusterName, clusterBean.Id, c)
	}

	//TODO - if cluster added or deleted, manage informer respectively
	return
}

func (impl *K8sInformerFactoryImpl) buildInformerAndNamespaceList(clusterName string, clusterId int, config *rest.Config) {
	clusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return
	}
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, time.Minute*time.Duration(5))
	nsInformenr := informerFactory.Core().V1().Namespaces()
	impl.logger.Infow(clusterName, "ns informer", nsInformenr.Informer())
	nsInformenr.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			impl.logger.Infow("add event handler", "cluster", clusterName, "object", obj.(metav1.Object))
			if mobject, ok := obj.(metav1.Object); ok {
				clusterKey := fmt.Sprintf("%s__%d", clusterName, clusterId)
				if _, ok := GlobalMapClusterNamespace[clusterKey]; !ok {
					allNamespaces := make(map[string]string)
					allNamespaces[mobject.GetName()] = mobject.GetName()
					GlobalMapClusterNamespace[clusterKey] = allNamespaces
				} else {
					allNamespaces := GlobalMapClusterNamespace[clusterKey]
					allNamespaces[mobject.GetName()] = mobject.GetName()
					GlobalMapClusterNamespace[clusterKey] = allNamespaces
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if mobject, ok := obj.(metav1.Object); ok {
				clusterKey := fmt.Sprintf("%s__%d", clusterName, clusterId)
				if _, ok := GlobalMapClusterNamespace[clusterKey]; ok {
					allNamespaces := GlobalMapClusterNamespace[clusterName]
					delete(allNamespaces, mobject.GetName())
					GlobalMapClusterNamespace[clusterKey] = allNamespaces
				}
			}
		},
	})
	informerFactory.Start(wait.NeverStop)
	return
}
