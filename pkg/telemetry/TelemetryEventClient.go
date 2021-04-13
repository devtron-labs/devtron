/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/jasonlvhit/gocron"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
	"time"
)

type TelemetryEventClient interface {
}

type TelemetryEventClientImpl struct {
	logger               *zap.SugaredLogger
	client               *http.Client
	clusterService       cluster.ClusterService
	K8sUtil              *util2.K8sUtil
	aCDAuthConfig        *user.ACDAuthConfig
	config               *client.EventClientConfig
	environmentService   cluster.EnvironmentService
	userService          user.UserService
	appListingRepository repository.AppListingRepository
}

func NewTelemetryEventClientImpl(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *util2.K8sUtil, aCDAuthConfig *user.ACDAuthConfig, config *client.EventClientConfig,
	environmentService cluster.EnvironmentService, userService user.UserService,
	appListingRepository repository.AppListingRepository) *TelemetryEventClientImpl {
	TelemetryEventClientImpl := &TelemetryEventClientImpl{
		logger: logger, client: client, clusterService: clusterService,
		K8sUtil: K8sUtil, aCDAuthConfig: aCDAuthConfig, config: config,
		environmentService: environmentService, userService: userService,
		appListingRepository: appListingRepository,
	}
	TelemetryEventClientImpl.WriteEventToTelemetryUserAnalyticsOnStartup()
	gocron.Every(1).Minute().Do(TelemetryEventClientImpl.WriteEventToTelemetryUserAnalytics)
	<-gocron.Start()

	return TelemetryEventClientImpl
}

func (impl *TelemetryEventClientImpl) WriteEventToTelemetryUserAnalyticsOnStartup() {
	impl.logger.Info(">>>>>>>>>>  startup event ")

	clusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		return
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return
	}

	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return
	}
	cm, err := impl.K8sUtil.GetConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, "devtron-upid", client)
	if err != nil && strings.Contains(err.Error(), "not found") {
		// if not found, create new cm
		//cm = &v12.ConfigMap{ObjectMeta: v13.ObjectMeta{Name: "devtron-upid"}}
		cm = &v1.ConfigMap{ObjectMeta: v12.ObjectMeta{Name: "devtron-upid"}}
		data := map[string]string{}
		data["UPID"] = util.Generate(10) // generate unique random number
		cm.Data = data
		_, err = impl.K8sUtil.CreateConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			return
		}
	}
	if cm == nil {
		return
	}
	dataMap := cm.Data
	upid := dataMap["UPID"]

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
	if err != nil {
		return
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return
	}
	payload := &TelemetryUserAnalyticsDto{Timestamp: time.Now(), EventType: "STARTUP", DevtronVersion: "v1"}
	payload.UPID = upid
	payload.ServerVersion = k8sServerVersion.String()

	clusters, err := impl.clusterService.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	environments, err := impl.environmentService.GetAllActive()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	users, err := impl.userService.GetAll()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	prodApps, err := impl.appListingRepository.FindAppCount(true)
	if err != nil && err != pg.ErrNoRows {
		return
	}

	nonProdApps, err := impl.appListingRepository.FindAppCount(false)
	if err != nil && err != pg.ErrNoRows {
		return
	}

	payload.Clusters = len(clusters)
	payload.Environments = len(environments)
	payload.Users = len(users)
	payload.NoOfProdApps = prodApps
	payload.NoOfNonProdApps = nonProdApps

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("WriteEventToTelemetryUserAnalyticsOnStartup, payload marshal error", "error", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/devtron/telemetry/event", impl.config.TelemetryUserAnalyticsUrl), bytes.NewBuffer(reqBody))
	if err != nil {
		impl.logger.Errorw("error while WriteEventToTelemetryUserAnalyticsOnStartup", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error while WriteEventToTelemetryUserAnalyticsOnStartup request ", "err", err)
		return
	}

}

func (impl *TelemetryEventClientImpl) WriteEventToTelemetryUserAnalytics() {
	impl.logger.Debug(">>>>>>>>>>  normal event ")

	clusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		return
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return
	}

	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return
	}
	cm, err := impl.K8sUtil.GetConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, "devtron-upid", client)
	if err != nil {
		return
	}
	dataMap := cm.Data
	upid := dataMap["UPID"]

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
	if err != nil {
		return
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return
	}
	payload := &TelemetryUserAnalyticsDto{Timestamp: time.Now(), EventType: "NORMAL", DevtronVersion: "v1"}
	payload.UPID = upid
	payload.ServerVersion = k8sServerVersion.String()

	clusters, err := impl.clusterService.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	environments, err := impl.environmentService.GetAllActive()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	users, err := impl.userService.GetAll()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	prodApps, err := impl.appListingRepository.FindAppCount(true)
	if err != nil && err != pg.ErrNoRows {
		return
	}

	nonProdApps, err := impl.appListingRepository.FindAppCount(false)
	if err != nil && err != pg.ErrNoRows {
		return
	}

	payload.Clusters = len(clusters)
	payload.Environments = len(environments)
	payload.Users = len(users)
	payload.NoOfProdApps = prodApps
	payload.NoOfNonProdApps = nonProdApps

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("WriteEventToTelemetryUserAnalytics, payload marshal error", "error", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/devtron/telemetry/event", impl.config.TelemetryUserAnalyticsUrl), bytes.NewBuffer(reqBody))
	if err != nil {
		impl.logger.Errorw("error while WriteEventToTelemetryUserAnalytics", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error while WriteEventToTelemetryUserAnalytics request ", "err", err)
		return
	}
}

type TelemetryUserAnalyticsDto struct {
	UPID            string    `json:"upid"`
	Timestamp       time.Time `json:"timestamp"`
	EventType       string    `json:"eventType"` //startup,normal,frequency
	ServerVersion   string    `json:"serverVersion"`
	DevtronVersion  string    `json:"devtronVersion"`
	Clusters        int       `json:"clusters"`
	Environments    int       `json:"environments"`
	NoOfProdApps    int       `json:"prodApps"`
	NoOfNonProdApps int       `json:"nonProdApps"`
	Users           int       `json:"users"`
}
