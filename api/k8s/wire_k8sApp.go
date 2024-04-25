package k8s

import (
	"github.com/devtron-labs/devtron/api/k8s/application"
	"github.com/devtron-labs/devtron/api/k8s/capacity"
	"github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s"
	application2 "github.com/devtron-labs/devtron/pkg/k8s/application"
	capacity2 "github.com/devtron-labs/devtron/pkg/k8s/capacity"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/google/wire"
)

var K8sApplicationWireSet = wire.NewSet(
	application2.NewK8sApplicationServiceImpl,
	wire.Bind(new(application2.K8sApplicationService), new(*application2.K8sApplicationServiceImpl)),
	k8s.NewK8sCommonServiceImpl,
	wire.Bind(new(k8s.K8sCommonService), new(*k8s.K8sCommonServiceImpl)),
	application.NewK8sApplicationRouterImpl,
	wire.Bind(new(application.K8sApplicationRouter), new(*application.K8sApplicationRouterImpl)),
	application.NewK8sApplicationRestHandlerImpl,
	wire.Bind(new(application.K8sApplicationRestHandler), new(*application.K8sApplicationRestHandlerImpl)),
	clusterRepository.NewEphemeralContainersRepositoryImpl,
	wire.Bind(new(clusterRepository.EphemeralContainersRepository), new(*clusterRepository.EphemeralContainersRepositoryImpl)),
	cluster.NewEphemeralContainerServiceImpl,
	wire.Bind(new(cluster.EphemeralContainerService), new(*cluster.EphemeralContainerServiceImpl)),
	terminal.NewTerminalSessionHandlerImpl,
	wire.Bind(new(terminal.TerminalSessionHandler), new(*terminal.TerminalSessionHandlerImpl)),
	capacity.NewK8sCapacityRouterImpl,
	wire.Bind(new(capacity.K8sCapacityRouter), new(*capacity.K8sCapacityRouterImpl)),
	capacity.NewK8sCapacityRestHandlerImpl,
	wire.Bind(new(capacity.K8sCapacityRestHandler), new(*capacity.K8sCapacityRestHandlerImpl)),
	capacity2.NewK8sCapacityServiceImpl,
	wire.Bind(new(capacity2.K8sCapacityService), new(*capacity2.K8sCapacityServiceImpl)),
	informer.NewGlobalMapClusterNamespace,
	informer.NewK8sInformerFactoryImpl,
	wire.Bind(new(informer.K8sInformerFactory), new(*informer.K8sInformerFactoryImpl)),

	cluster.NewClusterCronServiceImpl,
	wire.Bind(new(cluster.ClusterCronService), new(*cluster.ClusterCronServiceImpl)),
)
