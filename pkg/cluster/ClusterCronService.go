/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
