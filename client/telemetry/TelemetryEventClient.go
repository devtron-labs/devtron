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
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
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
	//TelemetryEventClientImpl.WriteEventToTelemetryUserAnalyticsOnStartup()
	//gocron.Every(1).Minute().Do(TelemetryEventClientImpl.WriteEventToTelemetryUserAnalytics)
	//<-gocron.Start()

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
		data["UCID"] = util.Generate(16) // generate unique random number
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
	ucid := dataMap["UCID"]

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
	if err != nil {
		return
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return
	}
	payload := &TelemetryUserAnalyticsDto{Timestamp: time.Now(), EventType: Summery, DevtronVersion: "v1"}
	payload.UCID = ucid
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

	summery := &SummeryDto{
		ProdAppCount:            prodApps,
		NonProdAppCount:         nonProdApps,
		UserCount:               len(users),
		EnvironmentCount:        len(environments),
		ClusterCount:            len(clusters),
		CiCountPerDay:           0,
		CdCountPerDay:           0,
		HelmChartCount:          0,
		SecurityScanCountPerDay: 0,
	}
	payload.Summery = summery
	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("WriteEventToTelemetryUserAnalyticsOnStartup, payload marshal error", "error", err)
		return
	}
	fmt.Print(reqBody)
}

type TelemetryUserAnalyticsDto struct {
	UCID           string             `json:"ucid"` //unique client id
	Timestamp      time.Time          `json:"timestamp"`
	EventMessage   string             `json:"eventMessage"`
	EventType      TelemetryEventType `json:"eventType"`
	Summery        *SummeryDto        `json:"summery"`
	ServerVersion  string             `json:"serverVersion"`
	DevtronVersion string             `json:"devtronVersion"`
}

type SummeryDto struct {
	ProdAppCount            int `json:"prodAppCount"`
	NonProdAppCount         int `json:"nonProdAppCount"`
	UserCount               int `json:"userCount"`
	EnvironmentCount        int `json:"environmentCount"`
	ClusterCount            int `json:"nonProdApps"`
	CiCountPerDay           int `json:"ciCountPerDay"`
	CdCountPerDay           int `json:"cdCountPerDay"`
	HelmChartCount          int `json:"helmChartCount"`
	SecurityScanCountPerDay int `json:"securityScanCountPerDay"`
}

const DevtronUniqueClientId = "devtron-ucid"

type TelemetryEventType int

const (
	Heartbeat TelemetryEventType = iota
	InstallationStart
	InstallationSuccess
	InstallationFailure
	UpgradeSuccess
	UpgradeFailure
	Summery
)

func (d TelemetryEventType) String() string {
	return [...]string{"Heartbeat", "InstallationStart", "InstallationSuccess", "InstallationFailure", "UpgradeSuccess", "UpgradeFailure", "Summery"}[d]
}

func (impl *TelemetryEventClientImpl) HeartbeatEventForTelemetry() {
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
		data["UCID"] = util.Generate(16) // generate unique random number
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
	ucid := dataMap["UCID"]

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
	if err != nil {
		return
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return
	}
	payload := &TelemetryUserAnalyticsDto{Timestamp: time.Now(), EventType: Heartbeat, DevtronVersion: "v1"}
	payload.UCID = ucid
	payload.ServerVersion = k8sServerVersion.String()
	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("HeartbeatEventForTelemetry, payload marshal error", "error", err)
		return
	}
	fmt.Print(reqBody)
}
