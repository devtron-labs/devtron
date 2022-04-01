package telemetry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/patrickmn/go-cache"
	"github.com/posthog/posthog-go"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"net/http"
	"time"
)

type TelemetryEventClientImpl struct {
	cron           *cron.Cron
	logger         *zap.SugaredLogger
	client         *http.Client
	clusterService cluster.ClusterService
	K8sUtil        *util2.K8sUtil
	aCDAuthConfig  *util3.ACDAuthConfig
	userService    user.UserService
	PosthogClient  *PosthogClient
}

type TelemetryEventClient interface {
	GetTelemetryMetaInfo() (*TelemetryMetaInfo, error)
	SendTelemetryInstallEventEA() (*TelemetryEventType, error)
	SendTelemetryDashboardAccessEvent() error
	SendTelemetryDashboardLoggedInEvent() error
}

func NewTelemetryEventClientImpl(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *util2.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig, userService user.UserService,
	PosthogClient *PosthogClient) (*TelemetryEventClientImpl, error) {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	watcher := &TelemetryEventClientImpl{
		cron:   cron,
		logger: logger,
		client: client, clusterService: clusterService,
		K8sUtil: K8sUtil, aCDAuthConfig: aCDAuthConfig,
		userService: userService, PosthogClient: PosthogClient,
	}

	watcher.HeartbeatEventForTelemetry()
	_, err := cron.AddFunc(SummaryCronExpr, watcher.SummaryEventForTelemetry)
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
	UCID           string             `json:"ucid"` //unique client id
	Timestamp      time.Time          `json:"timestamp"`
	EventMessage   string             `json:"eventMessage,omitempty"`
	EventType      TelemetryEventType `json:"eventType"`
	Summary        *SummaryEA         `json:"summary,omitempty"`
	ServerVersion  string             `json:"serverVersion,omitempty"`
	DevtronVersion string             `json:"devtronVersion,omitempty"`
	DevtronMode    string             `json:"devtronMode,omitempty"`
}

type SummaryEA struct {
	UserCount    int `json:"userCount,omitempty"`
	ClusterCount int `json:"clusterCount,omitempty"`
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
)

func (impl *TelemetryEventClientImpl) SummaryDetailsForTelemetry() (cluster []cluster.ClusterBean, user []bean.UserInfo, k8sServerVersion *version.Info) {
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
	return clusters, users, k8sServerVersion
}

func (impl *TelemetryEventClientImpl) SummaryEventForTelemetry() {
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	if IsOptOut {
		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
		return
	}

	clusters, users, k8sServerVersion := impl.SummaryDetailsForTelemetry()

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: Summary, DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()
	payload.DevtronMode = util.GetDevtronVersion().ServerMode

	summary := &SummaryEA{
		UserCount:    len(users),
		ClusterCount: len(clusters),
	}
	payload.Summary = summary

	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, payload marshal error", "error", err)
		return
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, payload unmarshal error", "error", err)
		return
	}

	err = impl.EnqueuePostHog(ucid, Summary, prop)
	if err != nil {
		impl.logger.Errorw("SummaryEventForTelemetry, failed to push event", "ucid", ucid, "error", err)
	}
}

func (impl *TelemetryEventClientImpl) EnqueuePostHog(ucid string, eventType TelemetryEventType, prop map[string]interface{}) error {
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
			Event:      string(eventType),
			Properties: prop,
		})
		if err != nil {
			impl.logger.Errorw("SummaryEventForTelemetry, failed to push event", "error", err)
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
	if err == nil {
		if err != nil {
			impl.logger.Warnw("HeartbeatEventForTelemetry, failed to push event", "error", err)
		} else {
			impl.logger.Debugw("HeartbeatEventForTelemetry success")
		}
	}
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

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: InstallationSuccess, DevtronVersion: "v1"}
	payload.DevtronMode = util.GetDevtronVersion().ServerMode

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

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: DashboardAccessed, DevtronVersion: "v1"}
	payload.DevtronMode = util.GetDevtronVersion().ServerMode

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

	payload := &TelemetryEventEA{UCID: ucid, Timestamp: time.Now(), EventType: DashboardLoggedIn, DevtronVersion: "v1"}
	payload.DevtronMode = util.GetDevtronVersion().ServerMode

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
