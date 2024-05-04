package scoop

import (
	"context"
	"fmt"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/k8s"
	util2 "github.com/devtron-labs/devtron/util"
	util5 "github.com/devtron-labs/devtron/util/event"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type Service interface {
	HandleNotificationEvent(ctx context.Context, request *k8s.ResourceRequestBean, notification map[string]interface{}) error
}

type ServiceImpl struct {
	logger         *zap.SugaredLogger
	clusterService cluster.ClusterService
	eventClient    client2.EventClient
	eventFactory   client2.EventFactory
}

func NewServiceImpl(logger *zap.SugaredLogger, clusterService cluster.ClusterService, eventClient client2.EventClient, eventFactory client2.EventFactory) *ServiceImpl {
	return &ServiceImpl{
		logger:         logger,
		clusterService: clusterService,
		eventClient:    eventClient,
		eventFactory:   eventFactory,
	}
}

func (impl ServiceImpl) HandleNotificationEvent(ctx context.Context, request *k8s.ResourceRequestBean, notification map[string]interface{}) error {
	cluster, err := impl.clusterService.FindById(request.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in finding cluster information by id", "clusterId", request.ClusterId, "err", err)
		return err
	}
	notification["cluster"] = cluster.ClusterName
	notification["eventTypeId"] = util5.ScoopNotification
	emailIfs, _ := notification["emails"].([]interface{})
	emails := util2.Map(emailIfs, func(emailIf interface{}) string {
		return emailIf.(string)
	})
	notification["correlationId"] = fmt.Sprintf("%s", uuid.NewV4())
	notification["payload"] = &client2.Payload{
		Providers: impl.eventFactory.BuildScoopNotificationEventProviders(emails),
	}

	_, err = impl.eventClient.SendAnyEvent(notification)
	if err != nil {
		impl.logger.Errorw("error in sending scoop event notification", "clusterId", request.ClusterId, "notification", notification, "err", err)
	}
	return err
}
