package k8s

import (
	"github.com/devtron-labs/devtron/client/k8s/informer"
	"github.com/google/wire"
)

var K8sWireSet = wire.NewSet(
	informer.NewGlobalMapClusterNamespace,
	informer.NewK8sInformerFactoryImpl,
	wire.Bind(new(informer.K8sInformerFactory), new(*informer.K8sInformerFactoryImpl)),
)
