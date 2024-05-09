package client

import (
	"fmt"
	pubsub_lib "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"testing"
)

func TestSendEventsOnNats(t *testing.T) {
	logger, err := util.NewSugardLogger()
	//nats, err := pubsub_lib.NewNatsClient(logger)
	//mockPubsubClient := NewPubSubClientServiceImpl(logger)
	mockPubsubClient := pubsub_lib.NewPubSubClientServiceImpl(logger)
	client := util.NewHttpClient()
	config := sql.Config{
		Port:     "5434",
		User:     "postgres",
		Password: "shared-devtron-pg",
		Database: "orchestrator_14",
		Addr:     "localhost:8080",
	}
	db, err := sql.NewDbConnection(&config, logger)
	impl := &EventRESTClientImpl{
		logger:               logger,
		pubsubClient:         mockPubsubClient,
		client:               client,
		config:               &EventClientConfig{DestinationURL: "localhost:3000/notify", NotificationMedium: PUB_SUB},
		ciPipelineRepository: pipelineConfig.NewCiPipelineRepositoryImpl(db, logger),
		pipelineRepository:   pipelineConfig.NewPipelineRepositoryImpl(db, logger),
		attributesRepository: repository.NewAttributesRepositoryImpl(db),
	}
	//xpectedTopic := "NOTIFICATION_EVENT_TOPIC"
	expectedMsg := "'{\"eventTypeId\":1,\"pipelineId\":123,\"payload\":{\"key\":\"value\"},\"eventTime\":\"2024-05-09T12:00:00Z\",\"appId\":456,\"envId\":789,\"teamId\":101}'"

	err = impl.sendEventsOnNats([]byte(expectedMsg))
	fmt.Println(err)

}
