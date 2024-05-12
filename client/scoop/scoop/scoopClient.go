package scoop

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	scoopClient "github.com/devtron-labs/scoop/client"
	"go.uber.org/zap"
)

type ScoopClientGetter interface {
	GetScoopClientByClusterId(clusterId int) (scoopClient.ScoopClient, error)
}

func NewScoopClientGetter(k8sApplicationService application.K8sApplicationService, logger *zap.SugaredLogger) ScoopClientGetter {
	return &ScoopClientGetterImpl{
		k8sApplicationService: k8sApplicationService,
		logger:                logger,
	}
}

type ScoopClientGetterImpl struct {
	k8sApplicationService application.K8sApplicationService
	logger                *zap.SugaredLogger
}

func (client ScoopClientGetterImpl) GetScoopClientByClusterId(clusterId int) (scoopClient.ScoopClient, error) {
	port, scoopConfig, err := client.k8sApplicationService.GetScoopPort(context.Background(), clusterId)
	if err != nil {
		client.logger.Errorw("error in initialising to scoop", "clusterId", clusterId, "scoopConfig", scoopConfig, "err", err)
		// not returning the error as we have to continue updating other scoops
		return nil, err
	}
	scoopUrl := fmt.Sprintf("http://127.0.0.1:%d", port)
	return scoopClient.NewScoopClientImpl(client.logger, scoopUrl, scoopConfig.PassKey)
}
