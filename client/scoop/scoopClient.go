/*
 * Copyright (c) 2024. Devtron Inc.
 */

package scoop

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	scoopClient "github.com/devtron-labs/scoop/client"
	"go.uber.org/zap"
)

type ScoopClientGetter interface {
	GetScoopClientByClusterId(clusterId int) (scoopClient.ScoopClient, error)
}

func NewScoopClientGetter(k8sApplicationService application.K8sApplicationService,
	environmentService cluster.EnvironmentService,
	logger *zap.SugaredLogger) *ScoopClientGetterImpl {

	scoopGetter := &ScoopClientGetterImpl{
		k8sApplicationService: k8sApplicationService,
		environmentService:    environmentService,
		logger:                logger,
	}

	// this is done this way because of cyclic imports with cluster->this->k8sApplicationService->helm-app-service->cluster
	// also we don't want to expose this function in hyperian mode
	environmentService.SetScoopClientGetter(scoopGetter.GetScoopClientByClusterId)
	return scoopGetter
}

type ScoopClientGetterImpl struct {
	k8sApplicationService application.K8sApplicationService
	environmentService    cluster.EnvironmentService
	logger                *zap.SugaredLogger
}

func (client ScoopClientGetterImpl) GetScoopClientByClusterId(clusterId int) (scoopClient.ScoopClient, error) {
	port, scoopConfig, err := client.k8sApplicationService.GetScoopPort(context.Background(), clusterId)
	if err != nil {
		client.logger.Errorw("error in initialising to scoop", "clusterId", clusterId, "scoopConfig", scoopConfig, "err", err)
		return nil, err
	}
	scoopUrl := fmt.Sprintf("http://127.0.0.1:%d", port)
	return scoopClient.NewScoopClientImpl(client.logger, scoopUrl, scoopConfig.PassKey)
}
