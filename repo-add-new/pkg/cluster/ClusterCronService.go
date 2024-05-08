package cluster

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	cron2 "github.com/devtron-labs/devtron/util/cron"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type ClusterCronService interface {
}

type ClusterCronServiceImpl struct {
	logger         *zap.SugaredLogger
	clusterService ClusterService
}

type ClusterStatusConfig struct {
	ClusterStatusCronTime int `env:"CLUSTER_STATUS_CRON_TIME" envDefault:"15"`
}

func NewClusterCronServiceImpl(logger *zap.SugaredLogger, clusterService ClusterService, cronLogger *cron2.CronLoggerImpl) (*ClusterCronServiceImpl, error) {
	clusterCronServiceImpl := &ClusterCronServiceImpl{
		logger:         logger,
		clusterService: clusterService,
	}
	// initialise cron
	newCron := cron.New(cron.WithChain(cron.Recover(cronLogger)))
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
	clusters, err := impl.clusterService.FindAllExceptVirtual()
	if err != nil {
		impl.logger.Errorw("error in getting all clusters", "err", err)
		return
	}
	impl.clusterService.ConnectClustersInBatch(clusters, true)
}
