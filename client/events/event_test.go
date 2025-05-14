/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	mockPubsubClient, err := pubsub_lib.NewPubSubClientServiceImpl(logger)
	client := util.NewHttpClient()
	config := sql.Config{}
	db, err := sql.NewDbConnection(&config, logger)
	trans := sql.NewTransactionUtilImpl(db)
	impl := &EventRESTClientImpl{
		logger:               logger,
		pubsubClient:         mockPubsubClient,
		client:               client,
		config:               &EventClientConfig{DestinationURL: "localhost:3000/notify", NotificationMedium: PUB_SUB},
		ciPipelineRepository: pipelineConfig.NewCiPipelineRepositoryImpl(db, logger, trans),
		pipelineRepository:   pipelineConfig.NewPipelineRepositoryImpl(db, logger),
		attributesRepository: repository.NewAttributesRepositoryImpl(db),
	}
	//xpectedTopic := "NOTIFICATION_EVENT_TOPIC"
	expectedMsg := "'{\"eventTypeId\":1,\"pipelineId\":123,\"payload\":{\"key\":\"value\"},\"eventTime\":\"2024-05-09T12:00:00Z\",\"appId\":456,\"envId\":789,\"teamId\":101}'"

	err = impl.sendEventsOnNats([]byte(expectedMsg))
	fmt.Println(err)

}
