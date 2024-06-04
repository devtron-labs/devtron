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

package out

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"go.uber.org/zap"
)

type AppStoreAppsEventPublishService interface {
	PublishBulkDeployEvent(installAppVersions []*appStoreBean.InstallAppVersionDTO) map[int]error
}

type AppStoreAppsEventPublishServiceImpl struct {
	logger       *zap.SugaredLogger
	pubSubClient *pubsub.PubSubClientServiceImpl
}

func NewAppStoreAppsEventPublishServiceImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl) *AppStoreAppsEventPublishServiceImpl {
	return &AppStoreAppsEventPublishServiceImpl{
		logger:       logger,
		pubSubClient: pubSubClient,
	}
}

// PublishBulkDeployEvent take installAppVersions and published their event. Response is map of installedAppVersionId along with error in publishing if any
func (impl *AppStoreAppsEventPublishServiceImpl) PublishBulkDeployEvent(installAppVersions []*appStoreBean.InstallAppVersionDTO) map[int]error {
	responseMap := make(map[int]error, len(installAppVersions))
	for _, version := range installAppVersions {
		var publishError error
		payload := &bean.BulkDeployPayload{InstalledAppVersionId: version.InstalledAppVersionId, InstalledAppVersionHistoryId: version.InstalledAppVersionHistoryId}
		data, err := json.Marshal(payload)
		if err != nil {
			impl.logger.Errorw("error in marshaling installed app version bulk deploy event payload", "err", err, "payload", payload)
			publishError = err
		} else {
			err = impl.pubSubClient.Publish(pubsub.BULK_APPSTORE_DEPLOY_TOPIC, string(data))
			if err != nil {
				impl.logger.Errorw("err while publishing msg for app-store bulk deploy", "msg", data, "err", err)
				publishError = err
			}
		}
		responseMap[version.InstalledAppVersionId] = publishError
	}
	return responseMap
}
