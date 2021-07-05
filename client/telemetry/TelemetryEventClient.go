package telemetry

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/patrickmn/go-cache"
	"github.com/posthog/posthog-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/robfig/cron.v3"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"
)

type TelemetryEventClientImpl struct {
	cron                 *cron.Cron
	logger               *zap.SugaredLogger
	client               *http.Client
	clusterService       cluster.ClusterService
	K8sUtil              *util2.K8sUtil
	aCDAuthConfig        *user.ACDAuthConfig
	environmentService   cluster.EnvironmentService
	userService          user.UserService
	appListingRepository repository.AppListingRepository
	PosthogClient        *PosthogClient
	ciPipelineRepository pipelineConfig.CiPipelineRepository
	pipelineRepository   pipelineConfig.PipelineRepository
	posthogConfig        *PosthogConfig
}

type TelemetryEventClient interface {
	GetTelemetryMetaInfo() (*TelemetryMetaInfo, error)
}

func NewTelemetryEventClientImpl(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *util2.K8sUtil, aCDAuthConfig *user.ACDAuthConfig,
	environmentService cluster.EnvironmentService, userService user.UserService,
	appListingRepository repository.AppListingRepository, PosthogClient *PosthogClient,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	posthogConfig *PosthogConfig) (*TelemetryEventClientImpl, error) {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	watcher := &TelemetryEventClientImpl{
		cron:   cron,
		logger: logger,
		client: client, clusterService: clusterService,
		K8sUtil: K8sUtil, aCDAuthConfig: aCDAuthConfig,
		environmentService: environmentService, userService: userService,
		appListingRepository: appListingRepository, PosthogClient: PosthogClient,
		ciPipelineRepository: ciPipelineRepository, pipelineRepository: pipelineRepository,
		posthogConfig: posthogConfig,
	}

	watcher.HeartbeatEventForTelemetry()
	_, err := cron.AddFunc(watcher.posthogConfig.SummaryCronExpr, watcher.SummaryEventForTelemetry)
	if err != nil {
		fmt.Println("error in starting summery event", "err", err)
		return nil, err
	}

	_, err = cron.AddFunc(watcher.posthogConfig.HeartbeatCronExpr, watcher.HeartbeatEventForTelemetry)
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
	EventMessage   string             `json:"eventMessage,omitempty"`
	EventType      TelemetryEventType `json:"eventType"`
	Summary        *SummaryDto        `json:"summary,omitempty"`
	ServerVersion  string             `json:"serverVersion,omitempty"`
	DevtronVersion string             `json:"devtronVersion,omitempty"`
}

type SummaryDto struct {
	ProdAppCount            int `json:"prodAppCount,omitempty"`
	NonProdAppCount         int `json:"nonProdAppCount,omitempty"`
	UserCount               int `json:"userCount,omitempty"`
	EnvironmentCount        int `json:"environmentCount,omitempty"`
	ClusterCount            int `json:"clusterCount,omitempty"`
	CiCountPerDay           int `json:"ciCountPerDay,omitempty"`
	CdCountPerDay           int `json:"cdCountPerDay,omitempty"`
	HelmChartCount          int `json:"helmChartCount,omitempty"`
	SecurityScanCountPerDay int `json:"securityScanCountPerDay,omitempty"`
}

const DevtronUniqueClientIdConfigMap = "devtron-ucid"
const DevtronUniqueClientIdConfigMapKey = "UCID"
const InstallEventKey = "installEvent"

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
)

func (impl *TelemetryEventClientImpl) SummaryEventForTelemetry() {
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}
	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}
	k8sServerVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}
	payload := &TelemetryEventDto{UCID: ucid, Timestamp: time.Now(), EventType: Summary, DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()
	clusters, err := impl.clusterService.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	environments, err := impl.environmentService.GetAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	users, err := impl.userService.GetAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	prodApps, err := impl.appListingRepository.FindAppCount(true)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	nonProdApps, err := impl.appListingRepository.FindAppCount(false)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	ciPipeline, err := impl.ciPipelineRepository.FindAllPipelineInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	cdPipeline, err := impl.pipelineRepository.FindAllPipelineInLast24Hour()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	summery := &SummaryDto{
		ProdAppCount:     prodApps,
		NonProdAppCount:  nonProdApps,
		UserCount:        len(users),
		EnvironmentCount: len(environments),
		ClusterCount:     len(clusters),
		CiCountPerDay:    len(ciPipeline),
		CdCountPerDay:    len(cdPipeline),
	}
	payload.Summary = summery
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

	if impl.PosthogClient.Client != nil {
		err = impl.PosthogClient.Client.Enqueue(posthog.Capture{
			DistinctId: ucid,
			Event:      string(Summary),
			Properties: prop,
		})
		if err != nil {
			impl.logger.Errorw("SummaryEventForTelemetry, failed to push event", "error", err)
		}
	}
}

func (impl *TelemetryEventClientImpl) HeartbeatEventForTelemetry() {
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
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
	if impl.PosthogClient.Client != nil {
		err = impl.PosthogClient.Client.Enqueue(posthog.Capture{
			DistinctId: ucid,
			Event:      string(Heartbeat),
			Properties: prop,
		})
		if err != nil {
			impl.logger.Errorw("HeartbeatEventForTelemetry, failed to push event", "error", err)
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
		Url:    impl.posthogConfig.PosthogEndpoint,
		UCID:   ucid,
		ApiKey: impl.PosthogClient.cfg.PosthogEncodedApiKey,
	}
	return data, err
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
	}
	return ucid.(string), nil
}
