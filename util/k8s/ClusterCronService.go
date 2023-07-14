package k8s

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type ClusterCronService interface {
}

type ClusterCronServiceImpl struct {
	logger         *zap.SugaredLogger
	clusterService cluster.ClusterService
}

type ClusterStatusConfig struct {
	ClusterStatusCronTime int `env:"CLUSTER_STATUS_CRON_TIME" envDefault:"15"`
}

func NewClusterCronServiceImpl(logger *zap.SugaredLogger, clusterService cluster.ClusterService) (*ClusterCronServiceImpl, error) {
	clusterCronServiceImpl := &ClusterCronServiceImpl{
		logger:         logger,
		clusterService: clusterService,
	}
	// initialise cron
	newCron := cron.New(cron.WithChain())
	newCron.Start()
	cfg := &ClusterStatusConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server cluster status config: " + err.Error())
	}
	// add function into cron
	_, err = newCron.AddFunc(fmt.Sprintf("@every %dm", cfg.ClusterStatusCronTime), clusterCronServiceImpl.GetAndUpdateClusterConnectionStatus)
	if err != nil {
		fmt.Println("error in adding cron function into cluster cron service")
		return clusterCronServiceImpl, err
	}
	return clusterCronServiceImpl, nil
}

func (impl *ClusterCronServiceImpl) GetAndUpdateClusterConnectionStatus() {
	impl.logger.Debug("starting cluster connection status fetch thread")
	defer impl.logger.Debug("stopped cluster connection status fetch thread")

	//getting all clusters
	clusters, err := impl.clusterService.FindAll()
	if err != nil {
		impl.logger.Errorw("error in getting all clusters", "err", err)
		return
	}
	impl.clusterService.ConnectClustersInBatch(clusters, true)
}
