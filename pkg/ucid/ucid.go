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

package ucid

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/telemetry/posthog"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Service interface {
	GetUCID() (string, bool, error)
}

type ServiceImpl struct {
	logger        *zap.SugaredLogger
	posthogClient *posthog.PosthogClient
	K8sUtil       *k8s.K8sServiceImpl
	aCDAuthConfig *util3.ACDAuthConfig
}

func NewServiceImpl(
	logger *zap.SugaredLogger,
	posthogClient *posthog.PosthogClient,
	K8sUtil *k8s.K8sServiceImpl,
	aCDAuthConfig *util3.ACDAuthConfig,
) *ServiceImpl {
	return &ServiceImpl{
		logger:        logger,
		posthogClient: posthogClient,
		K8sUtil:       K8sUtil,
		aCDAuthConfig: aCDAuthConfig,
	}
}

func (impl *ServiceImpl) GetUCID() (string, bool, error) {
	ucid, found := impl.posthogClient.GetCache().Get(DevtronUniqueClientIdConfigMapKey)
	if found {
		return ucid.(string), found, nil
	} else {
		client, err := impl.K8sUtil.GetClientForInCluster()
		if err != nil {
			impl.logger.Errorw("exception while getting unique client id", "error", err)
			return "", found, err
		}
		cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, DevtronUniqueClientIdConfigMap, client)
		if errStatus, ok := status.FromError(err); !ok || errStatus.Code() == codes.NotFound || errStatus.Code() == codes.Unknown {
			// if not found, create new cm
			cm = &v1.ConfigMap{ObjectMeta: v12.ObjectMeta{Name: DevtronUniqueClientIdConfigMap}}
			data := map[string]string{}
			data[DevtronUniqueClientIdConfigMapKey] = util.Generate(16) // generate unique random number
			data[InstallEventKey] = "1"                                 // used in operator to detect event is install or upgrade
			data[UIEventKey] = "1"
			cm.Data = data
			_, err = impl.K8sUtil.CreateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
			if err != nil {
				impl.logger.Errorw("exception while getting unique client id", "error", err)
				return "", found, err
			}
		}
		if cm == nil {
			impl.logger.Errorw("configmap not found while getting unique client id")
			return ucid.(string), found, fmt.Errorf("configmap %q not found while getting unique client id", DevtronUniqueClientIdConfigMap)
		}
		dataMap := cm.Data
		ucid = dataMap[DevtronUniqueClientIdConfigMapKey]
		impl.posthogClient.GetCache().Set(DevtronUniqueClientIdConfigMapKey, ucid, cache.DefaultExpiration)
	}
	return ucid.(string), found, nil
}
