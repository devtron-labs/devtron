package k8s

import (
	application2 "github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/client/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/google/wire"
)

var K8sApplicationWireSet = wire.NewSet(
	NewK8sApplicationRouterImpl,
	wire.Bind(new(K8sApplicationRouter), new(*K8sApplicationRouterImpl)),
	NewK8sApplicationRestHandlerImpl,
	wire.Bind(new(K8sApplicationRestHandler), new(*K8sApplicationRestHandlerImpl)),
	NewK8sApplicationServiceImpl,
	wire.Bind(new(K8sApplicationService), new(*K8sApplicationServiceImpl)),
	application2.NewK8sClientServiceImpl,
	wire.Bind(new(application2.K8sClientService), new(*application2.K8sClientServiceImpl)),
	terminal.NewTerminalSessionHandlerImpl,
	wire.Bind(new(terminal.TerminalSessionHandler), new(*terminal.TerminalSessionHandlerImpl)),
	NewK8sCapacityRouterImpl,
	wire.Bind(new(K8sCapacityRouter), new(*K8sCapacityRouterImpl)),
	NewK8sCapacityRestHandlerImpl,
	wire.Bind(new(K8sCapacityRestHandler), new(*K8sCapacityRestHandlerImpl)),
	NewK8sCapacityServiceImpl,
	wire.Bind(new(K8sCapacityService), new(*K8sCapacityServiceImpl)),
	informer.NewGlobalMapClusterNamespace,
	informer.NewK8sInformerFactoryImpl,
	wire.Bind(new(informer.K8sInformerFactory), new(*informer.K8sInformerFactoryImpl)),
	NewClusterCronServiceImpl,
	wire.Bind(new(ClusterCronService), new(*ClusterCronServiceImpl)),
)
