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

package telemetry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	cloudProviderIdentifier "github.com/devtron-labs/common-lib/cloud-provider-identifier"
	posthogTelemetry "github.com/devtron-labs/common-lib/telemetry"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	bean2 "github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	module2 "github.com/devtron-labs/devtron/pkg/module/bean"
	ucidService "github.com/devtron-labs/devtron/pkg/ucid"
	cron3 "github.com/devtron-labs/devtron/util/cron"
	"go.opentelemetry.io/otel"
	"net/http"
	"time"

	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/auth/sso"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/posthog/posthog-go"
	"github.com/robfig/cron/v3"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/version"
)

type TelemetryEventClientImpl struct {
	cron                           *cron.Cron
	logger                         *zap.SugaredLogger
	client                         *http.Client
	clusterService                 cluster.ClusterService
	K8sUtil                        *k8s.K8sServiceImpl
	aCDAuthConfig                  *util3.ACDAuthConfig
	userService                    user2.UserService
	attributeRepo                  repository.AttributesRepository
	ssoLoginService                sso.SSOLoginService
	posthogClient                  *posthogTelemetry.PosthogClient
	ucid                           ucidService.Service
	moduleRepository               moduleRepo.ModuleRepository
	serverDataStore                *serverDataStore.ServerDataStore
	userAuditService               user2.UserAuditService
	helmAppClient                  gRPC.HelmAppClient
	installedAppReadService        installedAppReader.InstalledAppReadServiceEA
	userAttributesRepository       repository.UserAttributesRepository
	cloudProviderIdentifierService cloudProviderIdentifier.ProviderIdentifierService
	telemetryConfig                TelemetryConfig
	globalEnvVariables             *util.GlobalEnvVariables
}

type TelemetryEventClient interface {
	GetTelemetryMetaInfo() (*TelemetryMetaInfo, error)
	SendTelemetryInstallEventEA() (*TelemetryEventType, error)
	SendTelemetryDashboardAccessEvent() error
	SendTelemetryDashboardLoggedInEvent() error
	SendGenericTelemetryEvent(eventType string, prop map[string]interface{}) error
	SendSummaryEvent(eventType string) error
}

func NewTelemetryEventClientImpl(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *k8s.K8sServiceImpl, aCDAuthConfig *util3.ACDAuthConfig, userService user2.UserService,
	attributeRepo repository.AttributesRepository, ssoLoginService sso.SSOLoginService,
	posthog *posthogTelemetry.PosthogClient, ucid ucidService.Service,
	moduleRepository moduleRepo.ModuleRepository, serverDataStore *serverDataStore.ServerDataStore,
	userAuditService user2.UserAuditService, helmAppClient gRPC.HelmAppClient,
	cloudProviderIdentifierService cloudProviderIdentifier.ProviderIdentifierService, cronLogger *cron3.CronLoggerImpl,
	installedAppReadService installedAppReader.InstalledAppReadServiceEA,
	envVariables *util.EnvironmentVariables,
	userAttributesRepository repository.UserAttributesRepository) (*TelemetryEventClientImpl, error) {
	cron := cron.New(
		cron.WithChain(cron.Recover(cronLogger)))
	cron.Start()
	watcher := &TelemetryEventClientImpl{
		cron:                           cron,
		logger:                         logger,
		client:                         client,
		clusterService:                 clusterService,
		K8sUtil:                        K8sUtil,
		aCDAuthConfig:                  aCDAuthConfig,
		userService:                    userService,
		attributeRepo:                  attributeRepo,
		ssoLoginService:                ssoLoginService,
		posthogClient:                  posthog,
		ucid:                           ucid,
		moduleRepository:               moduleRepository,
		serverDataStore:                serverDataStore,
		userAuditService:               userAuditService,
		helmAppClient:                  helmAppClient,
		installedAppReadService:        installedAppReadService,
		cloudProviderIdentifierService: cloudProviderIdentifierService,
		telemetryConfig:                TelemetryConfig{},
		globalEnvVariables:             envVariables.GlobalEnvVariables,
		userAttributesRepository:       userAttributesRepository,
	}

	watcher.HeartbeatEventForTelemetry()
	_, err := cron.AddFunc(posthogTelemetry.SummaryCronExpr, watcher.SummaryEventForTelemetryEA)
	if err != nil {
		logger.Errorw("error in starting summery event", "err", err)
		return nil, err
	}

	_, err = cron.AddFunc(posthogTelemetry.HeartbeatCronExpr, watcher.HeartbeatEventForTelemetry)
	if err != nil {
		logger.Errorw("error in starting heartbeat event", "err", err)
		return nil, err
	}
	return watcher, err
}

func (impl *TelemetryEventClientImpl) GetCloudProvider() (string, error) {
	// assumption: the IMDS server will be reachable on startup
	if len(impl.telemetryConfig.cloudProvider) == 0 {
		provider, err := impl.cloudProviderIdentifierService.IdentifyProvider()
		if err != nil {
			impl.logger.Errorw("exception while getting cluster provider", "error", err)
			return "", err
		}
		impl.telemetryConfig.cloudProvider = provider
	}
	return impl.telemetryConfig.cloudProvider, nil
}

func (impl *TelemetryEventClientImpl) StopCron() {
	impl.cron.Stop()
}

func (impl *TelemetryEventClientImpl) SummaryDetailsForTelemetry() (cluster []bean3.ClusterBean, user []bean.UserInfo,
	k8sServerVersion *version.Info, hostURL bool, ssoSetup bool, HelmAppAccessCount string, ChartStoreVisitCount string,
	SkippedOnboarding bool, HelmAppUpdateCounter string, helmChartSuccessfulDeploymentCount int, ExternalHelmAppClusterCount map[int32]int) {

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}
	k8sServerVersion, err = discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	users, err := impl.userService.GetAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	clusters, err := impl.clusterService.FindAllActive()

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	hostURL = false

	attribute, err := impl.attributeRepo.FindByKey(bean2.HostUrlKey)
	if err == nil && attribute.Id > 0 {
		hostURL = true
	}

	attribute, err = impl.attributeRepo.FindByKey("HelmAppAccessCounter")

	if err == nil {
		HelmAppAccessCount = attribute.Value
	}

	attribute, err = impl.attributeRepo.FindByKey("ChartStoreVisitCount")

	if err == nil {
		ChartStoreVisitCount = attribute.Value
	}

	attribute, err = impl.attributeRepo.FindByKey("HelmAppUpdateCounter")

	if err == nil {
		HelmAppUpdateCounter = attribute.Value
	}

	helmChartSuccessfulDeploymentCount, err = impl.installedAppReadService.GetDeploymentSuccessfulStatusCountForTelemetry()

	//externalHelmCount := make(map[int32]int)
	ExternalHelmAppClusterCount = make(map[int32]int)

	for _, clusterDetail := range clusters {
		req := &gRPC.AppListRequest{}
		config := &gRPC.ClusterConfig{
			ApiServerUrl:          clusterDetail.ServerUrl,
			Token:                 clusterDetail.Config[commonBean.BearerToken],
			ClusterId:             int32(clusterDetail.Id),
			ClusterName:           clusterDetail.ClusterName,
			InsecureSkipTLSVerify: clusterDetail.InsecureSkipTLSVerify,
		}

		if clusterDetail.InsecureSkipTLSVerify == false {
			config.KeyData = clusterDetail.Config[commonBean.TlsKey]
			config.CertData = clusterDetail.Config[commonBean.CertData]
			config.CaData = clusterDetail.Config[commonBean.CertificateAuthorityData]
		}
		req.Clusters = append(req.Clusters, config)
		applicationStream, err := impl.helmAppClient.ListApplication(context.Background(), req)
		if err == nil {
			clusterList, err1 := applicationStream.Recv()
			if err1 != nil {
				impl.logger.Errorw("error in list helm applications streams recv", "err", err)
			}
			if err1 != nil && clusterList != nil && !clusterList.Errored {
				ExternalHelmAppClusterCount[clusterList.ClusterId] = len(clusterList.DeployedAppDetail)
			}
		} else {
			impl.logger.Errorw("error while fetching list application from kubelink", "err", err)
		}
	}

	//getting userData from emailId
	userData, err := impl.userAttributesRepository.GetUserDataByEmailId(AdminEmailIdConst)

	SkippedOnboardingValue := gjson.Get(userData, SkippedOnboardingConst).Str

	if SkippedOnboardingValue == "true" {
		SkippedOnboarding = true
	} else {
		SkippedOnboarding = false
	}

	ssoSetup = false

	ssoConfig, err := impl.ssoLoginService.GetAll()
	if err == nil && len(ssoConfig) > 0 {
		ssoSetup = true
	}

	return clusters, users, k8sServerVersion, hostURL, ssoSetup, HelmAppAccessCount, ChartStoreVisitCount, SkippedOnboarding, HelmAppUpdateCounter, helmChartSuccessfulDeploymentCount, ExternalHelmAppClusterCount
}

func (impl *TelemetryEventClientImpl) SummaryEventForTelemetryEA() {
	err := impl.SendSummaryEvent(string(Summary))
	if err != nil {
		impl.logger.Errorw("error occurred in SummaryEventForTelemetryEA", "err", err)
	}
}

func (impl *TelemetryEventClientImpl) SendSummaryEvent(eventType string) error {
	impl.logger.Infow("sending summary event", "eventType", eventType)
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return err
	}

	if posthogTelemetry.IsOptOut {
		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
		return err
	}

	// build integrations data
	installedIntegrations, installFailedIntegrations, installTimedOutIntegrations, installingIntegrations, err := impl.buildIntegrationsList()
	if err != nil {
		return err
	}

	clusters, users, k8sServerVersion, hostURL, ssoSetup, HelmAppAccessCount, ChartStoreVisitCount, SkippedOnboarding, HelmAppUpdateCounter, helmChartSuccessfulDeploymentCount, ExternalHelmAppClusterCount := impl.SummaryDetailsForTelemetry()

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: TelemetryEventType(eventType), DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()
	payload.DevtronMode = util.GetDevtronVersion().ServerMode
	payload.HostURL = hostURL
	payload.SSOLogin = ssoSetup
	payload.UserCount = len(users)
	payload.ClusterCount = len(clusters)
	payload.InstalledIntegrations = installedIntegrations
	payload.InstallFailedIntegrations = installFailedIntegrations
	payload.InstallTimedOutIntegrations = installTimedOutIntegrations
	payload.InstallingIntegrations = installingIntegrations
	payload.DevtronReleaseVersion = impl.serverDataStore.CurrentVersion
	payload.HelmAppAccessCounter = HelmAppAccessCount
	payload.ChartStoreVisitCount = ChartStoreVisitCount
	payload.SkippedOnboarding = SkippedOnboarding
	payload.HelmAppUpdateCounter = HelmAppUpdateCounter
	payload.HelmChartSuccessfulDeploymentCount = helmChartSuccessfulDeploymentCount
	payload.ExternalHelmAppClusterCount = ExternalHelmAppClusterCount

	payload.ClusterProvider, err = impl.GetCloudProvider()
	if err != nil {
		impl.logger.Errorw("error while getting cluster provider", "error", err)
		return err
	}

	latestUser, err := impl.userAuditService.GetLatestUser()
	if err == nil {
		loginTime := latestUser.UpdatedOn
		if loginTime.IsZero() {
			loginTime = latestUser.CreatedOn
		}
		payload.LastLoginTime = loginTime
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, payload marshal error", "error", err)
		return err
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, payload unmarshal error", "error", err)
		return err
	}

	err = impl.EnqueuePostHog(ucid, TelemetryEventType(eventType), prop)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, failed to push event", "ucid", ucid, "error", err)
		return err
	}
	return nil
}

func (impl *TelemetryEventClientImpl) EnqueuePostHog(ucid string, eventType TelemetryEventType, prop map[string]interface{}) error {
	return impl.EnqueueGenericPostHogEvent(ucid, string(eventType), prop)
}

func (impl *TelemetryEventClientImpl) SendGenericTelemetryEvent(eventType string, prop map[string]interface{}) error {
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry generic event", "err", err)
		return nil
	}

	if posthogTelemetry.IsOptOut {
		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
		return nil
	}

	return impl.EnqueueGenericPostHogEvent(ucid, eventType, prop)
}

func (impl *TelemetryEventClientImpl) EnqueueGenericPostHogEvent(ucid string, eventType string, prop map[string]interface{}) error {
	if impl.posthogClient.Client == nil {
		impl.logger.Warn("no posthog client found, creating new")
		client, err := impl.retryPosthogClient(posthogTelemetry.PosthogApiKey, posthogTelemetry.PosthogEndpoint)
		if err == nil {
			impl.posthogClient.Client = client
		}
	}
	if impl.posthogClient.Client != nil && !impl.globalEnvVariables.IsAirGapEnvironment {
		err := impl.posthogClient.Client.Enqueue(posthog.Capture{
			DistinctId: ucid,
			Event:      eventType,
			Properties: prop,
		})
		if err != nil {
			impl.logger.Errorw("EnqueueGenericPostHogEvent, failed to push event", "error", err)
			return err
		}
	}
	return nil
}

func (impl *TelemetryEventClientImpl) HeartbeatEventForTelemetry() {
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}
	if posthogTelemetry.IsOptOut {
		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
		return
	}

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}

	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}
	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: Heartbeat, DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()
	payload.DevtronMode = util.GetDevtronVersion().ServerMode

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("HeartbeatEventForTelemetry, payload marshal error", "error", err)
		return
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("HeartbeatEventForTelemetry, payload unmarshal error", "error", err)
		return
	}

	err = impl.EnqueuePostHog(ucid, Heartbeat, prop)
	if err != nil {
		impl.logger.Warnw("HeartbeatEventForTelemetry, failed to push event", "error", err)
		return
	}
	impl.logger.Debugw("HeartbeatEventForTelemetry success")
	return
}

func (impl *TelemetryEventClientImpl) GetTelemetryMetaInfo() (*TelemetryMetaInfo, error) {
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return nil, err
	}
	data := &TelemetryMetaInfo{
		Url:    posthogTelemetry.PosthogEndpoint,
		UCID:   ucid,
		ApiKey: posthogTelemetry.PosthogEncodedApiKey,
	}
	return data, err
}

func (impl *TelemetryEventClientImpl) SendTelemetryInstallEventEA() (*TelemetryEventType, error) {
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return nil, err
	}

	client, err := impl.K8sUtil.GetClientForInCluster()
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return nil, err
	}

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return nil, err
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return nil, err
	}

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: InstallationSuccess, DevtronVersion: "v1"}
	payload.DevtronMode = util.GetDevtronVersion().ServerMode
	payload.ServerVersion = k8sServerVersion.String()

	payload.ClusterProvider, err = impl.GetCloudProvider()
	if err != nil {
		impl.logger.Errorw("error while getting cluster provider", "error", err)
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("Installation EventForTelemetry EA Mode, payload marshal error", "error", err)
		return nil, nil
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("Installation EventForTelemetry EA Mode, payload unmarshal error", "error", err)
		return nil, nil
	}
	cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, ucidService.DevtronUniqueClientIdConfigMap, client)
	if err != nil {
		impl.logger.Errorw("Installation EventForTelemetry EA Mode, failed to get DevtronUniqueClientIdConfigMap", "error", err)
		return nil, err
	}
	datamap := cm.Data

	installEventValue, installEventKeyExists := datamap[ucidService.InstallEventKey]

	if installEventKeyExists == false || installEventValue == "1" {
		err = impl.EnqueuePostHog(ucid, InstallationSuccess, prop)
		if err == nil {
			datamap[ucidService.InstallEventKey] = "2"
			cm.Data = datamap
			_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
			if err != nil {
				impl.logger.Warnw("config map update failed for install event", "err", err)
			} else {
				impl.logger.Debugw("config map apply succeeded for install event")
			}
		}
	}
	return nil, nil
}

func (impl *TelemetryEventClientImpl) SendTelemetryDashboardAccessEvent() error {
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return err
	}

	client, err := impl.K8sUtil.GetClientForInCluster()
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return err
	}

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return err
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return err
	}

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: DashboardAccessed, DevtronVersion: "v1"}
	payload.DevtronMode = util.GetDevtronVersion().ServerMode
	payload.ServerVersion = k8sServerVersion.String()

	payload.ClusterProvider, err = impl.GetCloudProvider()
	if err != nil {
		impl.logger.Errorw("error while getting cluster provider", "error", err)
		return err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("DashboardAccessed EventForTelemetry, payload marshal error", "error", err)
		return err
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("DashboardAccessed EventForTelemetry, payload unmarshal error", "error", err)
		return err
	}
	cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, ucidService.DevtronUniqueClientIdConfigMap, client)
	if err != nil {
		impl.logger.Errorw("DashboardAccessed EventForTelemetry,failed to get DevtronUniqueClientIdConfigMap", "error", err)
		return err
	}
	datamap := cm.Data

	accessEventValue, installEventKeyExists := datamap[ucidService.UIEventKey]

	if installEventKeyExists == false || accessEventValue <= "1" {
		err = impl.EnqueuePostHog(ucid, DashboardAccessed, prop)
		if err == nil {
			datamap[ucidService.UIEventKey] = "2"
			cm.Data = datamap
			_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
			if err != nil {
				impl.logger.Warnw("config map update failed for install event", "err", err)
			} else {
				impl.logger.Debugw("config map apply succeeded for install event")
			}
		}
	}
	return nil
}

func (impl *TelemetryEventClientImpl) SendTelemetryDashboardLoggedInEvent() error {
	ucid, err := impl.getUCIDAndCheckIsOptedOut(context.Background())
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return err
	}

	client, err := impl.K8sUtil.GetClientForInCluster()
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return err
	}

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return err
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return err
	}

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: DashboardLoggedIn, DevtronVersion: "v1"}
	payload.DevtronMode = util.GetDevtronVersion().ServerMode
	payload.ServerVersion = k8sServerVersion.String()

	payload.ClusterProvider, err = impl.GetCloudProvider()
	if err != nil {
		impl.logger.Errorw("error while getting cluster provider", "error", err)
		return err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("DashboardLoggedIn EventForTelemetry, payload marshal error", "error", err)
		return err
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("DashboardLoggedIn EventForTelemetry, payload unmarshal error", "error", err)
		return err
	}
	cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, ucidService.DevtronUniqueClientIdConfigMap, client)
	if err != nil {
		impl.logger.Errorw("DashboardLoggedIn EventForTelemetry,failed to get DevtronUniqueClientIdConfigMap", "error", err)
		return err
	}
	datamap := cm.Data

	accessEventValue, installEventKeyExists := datamap[ucidService.UIEventKey]

	if installEventKeyExists == false || accessEventValue != "3" {
		err = impl.EnqueuePostHog(ucid, DashboardLoggedIn, prop)
		if err == nil {
			datamap[ucidService.UIEventKey] = "3"
			cm.Data = datamap
			_, err = impl.K8sUtil.UpdateConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
			if err != nil {
				impl.logger.Warnw("config map update failed for install event", "err", err)
			} else {
				impl.logger.Debugw("config map apply succeeded for install event")
			}
		}
	}
	return nil
}

func (impl *TelemetryEventClientImpl) getUCIDAndCheckIsOptedOut(ctx context.Context) (string, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "TelemetryEventClientImpl.getUCIDAndCheckIsOptedOut")
	defer span.End()
	ucid, found, err := impl.ucid.GetUCIDWithCache(impl.posthogClient.GetCache())
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return "", err
	}
	if !found {
		flag, err := impl.checkForOptOut(newCtx, ucid)
		if err != nil {
			impl.logger.Errorw("error sending event to posthog, failed check for opt-out", "err", err)
			return "", err
		}
		posthogTelemetry.IsOptOut = flag
	}
	return ucid, nil
}

func (impl *TelemetryEventClientImpl) checkForOptOut(ctx context.Context, UCID string) (bool, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "TelemetryEventClientImpl.checkForOptOut")
	defer span.End()
	decodedUrl, err := base64.StdEncoding.DecodeString(posthogTelemetry.TelemetryOptOutApiBaseUrl)
	if err != nil {
		impl.logger.Errorw("check opt-out list failed, decode error", "err", err)
		return false, err
	}
	encodedUrl := string(decodedUrl)
	url := fmt.Sprintf("%s/%s", encodedUrl, UCID)

	response, err := util.HttpRequest(newCtx, url)
	if err != nil {
		// this should be non-blocking call and should not fail the request for ucid getting
		impl.logger.Errorw("check opt-out list failed, rest api error", "ucid", UCID, "err", err)
	}
	flag := response["result"].(bool)
	return flag, nil
}

func (impl *TelemetryEventClientImpl) retryPosthogClient(PosthogApiKey string, PosthogEndpoint string) (posthog.Client, error) {
	client, err := posthog.NewWithConfig(PosthogApiKey, posthog.Config{Endpoint: PosthogEndpoint})
	//defer client.Close()
	if err != nil {
		impl.logger.Errorw("exception caught while creating posthog client", "err", err)
	}
	return client, err
}

// buildIntegrationsList - returns installedIntegrations, installFailedIntegrations, installTimedOutIntegrations, installingIntegrations
func (impl *TelemetryEventClientImpl) buildIntegrationsList() ([]string, []string, []string, []string, error) {
	impl.logger.Info("building integrations list for telemetry")

	modules, err := impl.moduleRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting integrations list", "err", err)
		return nil, nil, nil, nil, err
	}

	var installedIntegrations []string
	var installFailedIntegrations []string
	var installTimedOutIntegrations []string
	var installingIntegrations []string

	for _, module := range modules {
		integrationName := module.Name
		switch module.Status {
		case module2.ModuleStatusInstalled:
			installedIntegrations = append(installedIntegrations, integrationName)
		case module2.ModuleStatusInstallFailed:
			installFailedIntegrations = append(installFailedIntegrations, integrationName)
		case module2.ModuleStatusTimeout:
			installTimedOutIntegrations = append(installTimedOutIntegrations, integrationName)
		case module2.ModuleStatusInstalling:
			installingIntegrations = append(installingIntegrations, integrationName)
		}
	}

	return installedIntegrations, installFailedIntegrations, installTimedOutIntegrations, installingIntegrations, nil

}
