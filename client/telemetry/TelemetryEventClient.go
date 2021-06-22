package telemetry

import (
	"encoding/json"
	"fmt"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/posthog/posthog-go"
	"go.uber.org/zap"
	"gopkg.in/robfig/cron.v3"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
	"time"
)

type TelemetryEventClientImpl struct {
	cron                 *cron.Cron
	logger               *zap.SugaredLogger
	client               *http.Client
	clusterService       cluster.ClusterService
	K8sUtil              *util2.K8sUtil
	aCDAuthConfig        *user.ACDAuthConfig
	config               *client.EventClientConfig
	environmentService   cluster.EnvironmentService
	userService          user.UserService
	appListingRepository repository.AppListingRepository
	PosthogClient        *PosthogClient
	ciPipelineRepository pipelineConfig.CiPipelineRepository
	pipelineRepository   pipelineConfig.PipelineRepository
}

type TelemetryEventClient interface {
}

func NewTelemetryEventClientImpl(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *util2.K8sUtil, aCDAuthConfig *user.ACDAuthConfig, config *client.EventClientConfig,
	environmentService cluster.EnvironmentService, userService user.UserService,
	appListingRepository repository.AppListingRepository, PosthogClient *PosthogClient,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository) (*TelemetryEventClientImpl, error) {
	cronLogger := &CronLoggerImpl{logger: logger}
	cron := cron.New(
		cron.WithChain(
			cron.SkipIfStillRunning(cronLogger),
			cron.Recover(cronLogger)))
	cron.Start()
	watcher := &TelemetryEventClientImpl{
		cron:   cron,
		logger: logger,
		client: client, clusterService: clusterService,
		K8sUtil: K8sUtil, aCDAuthConfig: aCDAuthConfig, config: config,
		environmentService: environmentService, userService: userService,
		appListingRepository: appListingRepository, PosthogClient: PosthogClient,
		ciPipelineRepository: ciPipelineRepository, pipelineRepository: pipelineRepository,
	}
	watcher.HeartbeatEventForTelemetry()
	_, err := cron.AddFunc(fmt.Sprintf("@every %dm", 5), watcher.SummeryEventForTelemetry)
	if err != nil {
		fmt.Println("error in starting summery event", "err", err)
		return nil, err
	}

	_, err = cron.AddFunc(fmt.Sprintf("@every %dm", 2), watcher.HeartbeatEventForTelemetry)
	if err != nil {
		fmt.Println("error in starting heartbeat event", "err", err)
		return nil, err
	}
	return watcher, err
}

func (impl *TelemetryEventClientImpl) StopCron() {
	impl.cron.Stop()
}

type TelemetryEventDto struct {
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

const DevtronUniqueClientIdConfigMap = "devtron-ucid"
const DevtronUniqueClientIdConfigMapKey = "UCID"

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

func (impl *TelemetryEventClientImpl) SummeryEventForTelemetry() {
	clusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	client, err := impl.K8sUtil.GetClientForIncluster(cfg)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}
	cm, err := impl.K8sUtil.GetConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, DevtronUniqueClientIdConfigMap, client)
	if err != nil && strings.Contains(err.Error(), "not found") {
		// if not found, create new cm
		//cm = &v12.ConfigMap{ObjectMeta: v13.ObjectMeta{Name: "devtron-upid"}}
		cm = &v1.ConfigMap{ObjectMeta: v12.ObjectMeta{Name: DevtronUniqueClientIdConfigMap}}
		data := map[string]string{}
		data[DevtronUniqueClientIdConfigMapKey] = util.Generate(16) // generate unique random number
		cm.Data = data
		_, err = impl.K8sUtil.CreateConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			return
		}
	}
	if cm == nil {
		impl.logger.Errorw("cm found nil inside telemetry summery event", "cm", cm)
		return
	}
	dataMap := cm.Data
	ucid := dataMap[DevtronUniqueClientIdConfigMapKey]

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}
	payload := &TelemetryEventDto{UCID: ucid, Timestamp: time.Now(), EventType: Summery, DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()
	clusters, err := impl.clusterService.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	environments, err := impl.environmentService.GetAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	users, err := impl.userService.GetAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	prodApps, err := impl.appListingRepository.FindAppCount(true)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	nonProdApps, err := impl.appListingRepository.FindAppCount(false)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	ciPipeline, err := impl.ciPipelineRepository.FindAllPipelineInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	cdPipeline, err := impl.pipelineRepository.FindAllPipelineInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	summery := &SummeryDto{
		ProdAppCount:     prodApps,
		NonProdAppCount:  nonProdApps,
		UserCount:        len(users),
		EnvironmentCount: len(environments),
		ClusterCount:     len(clusters),
		CiCountPerDay:    len(ciPipeline),
		CdCountPerDay:    len(cdPipeline),
	}
	payload.Summery = summery
	reqBody, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("SummeryEventForTelemetry, payload marshal error", "error", err)
		return
	}
	prop := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &prop)
	if err != nil {
		impl.logger.Errorw("SummeryEventForTelemetry, payload unmarshal error", "error", err)
		return
	}
	impl.PosthogClient.Client.Enqueue(posthog.Capture{
		DistinctId: ucid,
		Event:      fmt.Sprintf("summary event sent from devtron"),
		Properties: prop,
	})
}

func (impl *TelemetryEventClientImpl) HeartbeatEventForTelemetry() {
	clusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}

	client, err := impl.K8sUtil.GetClientForIncluster(cfg)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}
	cm, err := impl.K8sUtil.GetConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, DevtronUniqueClientIdConfigMap, client)
	if err != nil && strings.Contains(err.Error(), "not found") {
		// if not found, create new cm
		//cm = &v12.ConfigMap{ObjectMeta: v13.ObjectMeta{Name: "devtron-upid"}}
		cm = &v1.ConfigMap{ObjectMeta: v12.ObjectMeta{Name: DevtronUniqueClientIdConfigMap}}
		data := map[string]string{}
		data[DevtronUniqueClientIdConfigMapKey] = util.Generate(16) // generate unique random number
		cm.Data = data
		_, err = impl.K8sUtil.CreateConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
			return
		}
	}
	if cm == nil {
		impl.logger.Errorw("configmap found nil for telemetry heartbeat event", "cm", cm)
		return
	}
	dataMap := cm.Data
	ucid := dataMap[DevtronUniqueClientIdConfigMapKey]

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClient(cfg)
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
		return
	}
	payload := &TelemetryEventDto{UCID: ucid, Timestamp: time.Now(), EventType: Heartbeat, DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()
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
	impl.PosthogClient.Client.Enqueue(posthog.Capture{
		DistinctId: ucid,
		Event:      fmt.Sprintf("heartbeat event sent from devtron"),
		Properties: prop,
	})
}

type CronLoggerImpl struct {
	logger *zap.SugaredLogger
}

func (impl *CronLoggerImpl) Info(msg string, keysAndValues ...interface{}) {
	impl.logger.Infow(msg, keysAndValues...)
}

func (impl *CronLoggerImpl) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"err", err}, keysAndValues...)
	impl.logger.Errorw(msg, keysAndValues...)
}
