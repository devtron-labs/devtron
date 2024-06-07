/*
 * Copyright (c) 2024. Devtron Inc.
 */

package devtronResource

import "github.com/devtron-labs/devtron/pkg/devtronResource/repository"

func (impl *DevtronResourceServiceImpl) getClusterIdentifierByClusterId(clusterId int) (string, error) {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in finding cluster by cluster id", "err", err, "clusterId", clusterId)
		return "", err
	}
	return cluster.ClusterName, nil
}

func (impl *DevtronResourceServiceImpl) buildIdentifierForClusterResourceObj(object *repository.DevtronResourceObject) (string, error) {
	return impl.getClusterIdentifierByClusterId(object.OldObjectId)
}
