package telemetry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	cloudProviderIdentifier "github.com/devtron-labs/common-lib/cloud-provider-identifier"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"net/http"
	"time"

	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/auth/sso"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	module2 "github.com/devtron-labs/devtron/pkg/module"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/patrickmn/go-cache"
	"github.com/posthog/posthog-go"
	"github.com/robfig/cron/v3"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

const LOGIN_COUNT_CONST = "login-count"
const SKIPPED_ONBOARDING_CONST = "SkippedOnboarding"
const ADMIN_EMAIL_ID_CONST = "admin"

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
	PosthogClient                  *PosthogClient
	moduleRepository               moduleRepo.ModuleRepository
	serverDataStore                *serverDataStore.ServerDataStore
	userAuditService               user2.UserAuditService
	helmAppClient                  gRPC.HelmAppClient
	InstalledAppRepository         repository2.InstalledAppRepository
	userAttributesRepository       repository.UserAttributesRepository
	cloudProviderIdentifierService cloudProviderIdentifier.ProviderIdentifierService
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
	PosthogClient *PosthogClient, moduleRepository moduleRepo.ModuleRepository, serverDataStore *serverDataStore.ServerDataStore, userAuditService user2.UserAuditService, helmAppClient gRPC.HelmAppClient, InstalledAppRepository repository2.InstalledAppRepository,
	cloudProviderIdentifierService cloudProviderIdentifier.ProviderIdentifierService) (*TelemetryEventClientImpl, error) {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	watcher := &TelemetryEventClientImpl{
		cron:   cron,
		logger: logger,
		client: client, clusterService: clusterService,
		K8sUtil: K8sUtil, aCDAuthConfig: aCDAuthConfig,
		userService: userService, attributeRepo: attributeRepo,
		ssoLoginService:                ssoLoginService,
		PosthogClient:                  PosthogClient,
		moduleRepository:               moduleRepository,
		serverDataStore:                serverDataStore,
		userAuditService:               userAuditService,
		helmAppClient:                  helmAppClient,
		InstalledAppRepository:         InstalledAppRepository,
		cloudProviderIdentifierService: cloudProviderIdentifierService,
	}

	watcher.HeartbeatEventForTelemetry()
	_, err := cron.AddFunc(SummaryCronExpr, watcher.SummaryEventForTelemetryEA)
	if err != nil {
		logger.Errorw("error in starting summery event", "err", err)
		return nil, err
	}

	_, err = cron.AddFunc(HeartbeatCronExpr, watcher.HeartbeatEventForTelemetry)
	if err != nil {
		logger.Errorw("error in starting heartbeat event", "err", err)
		return nil, err
	}
	return watcher, err
}

func (impl *TelemetryEventClientImpl) StopCron() {
	impl.cron.Stop()
}

type TelemetryEventEA struct {
	UCID                               string             `json:"ucid"` //unique client id
	Timestamp                          time.Time          `json:"timestamp"`
	EventMessage                       string             `json:"eventMessage,omitempty"`
	EventType                          TelemetryEventType `json:"eventType"`
	ServerVersion                      string             `json:"serverVersion,omitempty"`
	UserCount                          int                `json:"userCount,omitempty"`
	ClusterCount                       int                `json:"clusterCount,omitempty"`
	HostURL                            bool               `json:"hostURL,omitempty"`
	SSOLogin                           bool               `json:"ssoLogin,omitempty"`
	DevtronVersion                     string             `json:"devtronVersion,omitempty"`
	DevtronMode                        string             `json:"devtronMode,omitempty"`
	InstalledIntegrations              []string           `json:"installedIntegrations,omitempty"`
	InstallFailedIntegrations          []string           `json:"installFailedIntegrations,omitempty"`
	InstallTimedOutIntegrations        []string           `json:"installTimedOutIntegrations,omitempty"`
	LastLoginTime                      time.Time          `json:"LastLoginTime,omitempty"`
	InstallingIntegrations             []string           `json:"installingIntegrations,omitempty"`
	DevtronReleaseVersion              string             `json:"devtronReleaseVersion,omitempty"`
	HelmAppAccessCounter               string             `json:"HelmAppAccessCounter,omitempty"`
	HelmAppUpdateCounter               string             `json:"HelmAppUpdateCounter,omitempty"`
	ChartStoreVisitCount               string             `json:"ChartStoreVisitCount,omitempty"`
	SkippedOnboarding                  bool               `json:"SkippedOnboarding"`
	HelmChartSuccessfulDeploymentCount int                `json:"helmChartSuccessfulDeploymentCount,omitempty"`
	ExternalHelmAppClusterCount        map[int32]int      `json:"ExternalHelmAppClusterCount,omitempty"`
	ClusterProvider                    string             `json:"clusterProvider,omitempty"`
}

const DevtronUniqueClientIdConfigMap = "devtron-ucid"
const DevtronUniqueClientIdConfigMapKey = "UCID"
const InstallEventKey = "installEvent"
const UIEventKey = "uiEventKey"

type TelemetryEventType string

const (
	Heartbeat                    TelemetryEventType = "Heartbeat"
	InstallationStart            TelemetryEventType = "InstallationStart"
	InstallationInProgress       TelemetryEventType = "InstallationInProgress"
	InstallationInterrupt        TelemetryEventType = "InstallationInterrupt"
	InstallationSuccess          TelemetryEventType = "InstallationSuccess"
	InstallationFailure          TelemetryEventType = "InstallationFailure"
	UpgradeStart                 TelemetryEventType = "UpgradeStart"
	UpgradeInProgress            TelemetryEventType = "UpgradeInProgress"
	UpgradeInterrupt             TelemetryEventType = "UpgradeInterrupt"
	UpgradeSuccess               TelemetryEventType = "UpgradeSuccess"
	UpgradeFailure               TelemetryEventType = "UpgradeFailure"
	Summary                      TelemetryEventType = "Summary"
	InstallationApplicationError TelemetryEventType = "InstallationApplicationError"
	DashboardAccessed            TelemetryEventType = "DashboardAccessed"
	DashboardLoggedIn            TelemetryEventType = "DashboardLoggedIn"
	SIG_TERM                     TelemetryEventType = "SIG_TERM"
)

func (impl *TelemetryEventClientImpl) SummaryDetailsForTelemetry() (cluster []cluster.ClusterBean, user []bean.UserInfo,
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

	attribute, err := impl.attributeRepo.FindByKey(attributes.HostUrlKey)
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

	helmChartSuccessfulDeploymentCount, err = impl.InstalledAppRepository.GetDeploymentSuccessfulStatusCountForTelemetry()

	//externalHelmCount := make(map[int32]int)
	ExternalHelmAppClusterCount = make(map[int32]int)

	for _, clusterDetail := range clusters {
		req := &gRPC.AppListRequest{}
		config := &gRPC.ClusterConfig{
			ApiServerUrl:          clusterDetail.ServerUrl,
			Token:                 clusterDetail.Config[k8s.BearerToken],
			ClusterId:             int32(clusterDetail.Id),
			ClusterName:           clusterDetail.ClusterName,
			InsecureSkipTLSVerify: clusterDetail.InsecureSkipTLSVerify,
		}
		if clusterDetail.InsecureSkipTLSVerify == false {
			config.KeyData = clusterDetail.Config[k8s.TlsKey]
			config.CertData = clusterDetail.Config[k8s.CertData]
			config.CaData = clusterDetail.Config[k8s.CertificateAuthorityData]
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
	userData, err := impl.userAttributesRepository.GetUserDataByEmailId(ADMIN_EMAIL_ID_CONST)

	SkippedOnboardingValue := gjson.Get(userData, SKIPPED_ONBOARDING_CONST).Str

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
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return err
	}

	if IsOptOut {
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

	provider, err := impl.cloudProviderIdentifierService.IdentifyProvider()
	if err != nil {
		impl.logger.Errorw("exception while getting cluster provider", "error", err)
		return err
	}
	payload.ClusterProvider = provider

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
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry generic event", "err", err)
		return nil
	}

	if IsOptOut {
		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
		return nil
	}

	return impl.EnqueueGenericPostHogEvent(ucid, eventType, prop)
}

func (impl *TelemetryEventClientImpl) EnqueueGenericPostHogEvent(ucid string, eventType string, prop map[string]interface{}) error {
	if impl.PosthogClient.Client == nil {
		impl.logger.Warn("no posthog client found, creating new")
		client, err := impl.retryPosthogClient(PosthogApiKey, PosthogEndpoint)
		if err == nil {
			impl.PosthogClient.Client = client
		}
	}
	if impl.PosthogClient.Client != nil {
		err := impl.PosthogClient.Client.Enqueue(posthog.Capture{
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
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}
	if IsOptOut {
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
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception while getting unique client id", "error", err)
		return nil, err
	}
	data := &TelemetryMetaInfo{
		Url:    PosthogEndpoint,
		UCID:   ucid,
		ApiKey: PosthogEncodedApiKey,
	}
	return data, err
}

func (impl *TelemetryEventClientImpl) SendTelemetryInstallEventEA() (*TelemetryEventType, error) {
	ucid, err := impl.getUCID()
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

	provider, err := impl.cloudProviderIdentifierService.IdentifyProvider()
	if err != nil {
		impl.logger.Errorw("exception while getting cluster provider", "error", err)
		return nil, err
	}
	payload.ClusterProvider = provider

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
	cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, DevtronUniqueClientIdConfigMap, client)
	datamap := cm.Data

	installEventValue, installEventKeyExists := datamap[InstallEventKey]

	if installEventKeyExists == false || installEventValue == "1" {
		err = impl.EnqueuePostHog(ucid, InstallationSuccess, prop)
		if err == nil {
			datamap[InstallEventKey] = "2"
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
	ucid, err := impl.getUCID()
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

	provider, err := impl.cloudProviderIdentifierService.IdentifyProvider()
	if err != nil {
		impl.logger.Errorw("exception while getting cluster provider", "error", err)
		return err
	}
	payload.ClusterProvider = provider

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
	cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, DevtronUniqueClientIdConfigMap, client)
	datamap := cm.Data

	accessEventValue, installEventKeyExists := datamap[UIEventKey]

	if installEventKeyExists == false || accessEventValue <= "1" {
		err = impl.EnqueuePostHog(ucid, DashboardAccessed, prop)
		if err == nil {
			datamap[UIEventKey] = "2"
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
	ucid, err := impl.getUCID()
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

	provider, err := impl.cloudProviderIdentifierService.IdentifyProvider()
	if err != nil {
		impl.logger.Errorw("exception while getting cluster provider", "error", err)
		return err
	}
	payload.ClusterProvider = provider

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
	cm, err := impl.K8sUtil.GetConfigMap(impl.aCDAuthConfig.ACDConfigMapNamespace, DevtronUniqueClientIdConfigMap, client)
	datamap := cm.Data

	accessEventValue, installEventKeyExists := datamap[UIEventKey]

	if installEventKeyExists == false || accessEventValue != "3" {
		err = impl.EnqueuePostHog(ucid, DashboardLoggedIn, prop)
		if err == nil {
			datamap[UIEventKey] = "3"
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

type TelemetryMetaInfo struct {
	Url    string `json:"url,omitempty"`
	UCID   string `json:"ucid,omitempty"`
	ApiKey string `json:"apiKey,omitempty"`
}

func (impl *TelemetryEventClientImpl) getUCID() (string, error) {
	ucid, found := impl.PosthogClient.cache.Get(DevtronUniqueClientIdConfigMapKey)
	if found {
		return ucid.(string), nil
	} else {
		client, err := impl.K8sUtil.GetClientForInCluster()
		if err != nil {
			impl.logger.Errorw("exception while getting unique client id", "error", err)
			return "", err
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
				return "", err
			}
		}
		dataMap := cm.Data
		ucid = dataMap[DevtronUniqueClientIdConfigMapKey]
		impl.PosthogClient.cache.Set(DevtronUniqueClientIdConfigMapKey, ucid, cache.DefaultExpiration)
		if cm == nil {
			impl.logger.Errorw("configmap not found while getting unique client id", "cm", cm)
			return ucid.(string), err
		}
		flag, err := impl.checkForOptOut(ucid.(string))
		if err != nil {
			impl.logger.Errorw("error sending event to posthog, failed check for opt-out", "err", err)
			return "", err
		}
		IsOptOut = flag
	}
	return ucid.(string), nil
}

func (impl *TelemetryEventClientImpl) checkForOptOut(UCID string) (bool, error) {
	decodedUrl, err := base64.StdEncoding.DecodeString(TelemetryOptOutApiBaseUrl)
	if err != nil {
		impl.logger.Errorw("check opt-out list failed, decode error", "err", err)
		return false, err
	}
	encodedUrl := string(decodedUrl)
	url := fmt.Sprintf("%s/%s", encodedUrl, UCID)

	response, err := util.HttpRequest(url)
	if err != nil {
		impl.logger.Errorw("check opt-out list failed, rest api error", "err", err)
		return false, err
	}
	flag := response["result"].(bool)
	return flag, err
}

func (impl *TelemetryEventClientImpl) retryPosthogClient(PosthogApiKey string, PosthogEndpoint string) (posthog.Client, error) {
	client, err := posthog.NewWithConfig(PosthogApiKey, posthog.Config{Endpoint: PosthogEndpoint})
	//defer client.Close()
	if err != nil {
		impl.logger.Errorw("exception caught while creating posthog client", "err", err)
	}
	return client, err
}

// returns installedIntegrations, installFailedIntegrations, installTimedOutIntegrations, installingIntegrations
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
