package telemetry

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/posthog/posthog-go"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type TelemetryEventClientImplExtended struct {
	environmentService   cluster.EnvironmentService
	appListingRepository repository.AppListingRepository
	ciPipelineRepository pipelineConfig.CiPipelineRepository
	pipelineRepository   pipelineConfig.PipelineRepository
	*TelemetryEventClientImpl
}

//type TelemetryEventClientExtended interface {
//	GetTelemetryMetaInfo() (*TelemetryMetaInfo, error)
//}

func (impl *TelemetryEventClientImplExtended) NewTelemetryEventClientImplExtended(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *util2.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig,
	environmentService cluster.EnvironmentService, userService user.UserService,
	appListingRepository repository.AppListingRepository, PosthogClient *PosthogClient,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository) (*TelemetryEventClientImplExtended, error) {

	cron := cron.New(
		cron.WithChain())
	cron.Start()
	watcher := &TelemetryEventClientImplExtended{
		environmentService:   environmentService,
		appListingRepository: appListingRepository,
		ciPipelineRepository: ciPipelineRepository, pipelineRepository: pipelineRepository,
		TelemetryEventClientImpl: &TelemetryEventClientImpl{
			cron:           cron,
			logger:         logger,
			client:         client,
			clusterService: clusterService,
			K8sUtil:        K8sUtil,
			aCDAuthConfig:  aCDAuthConfig,
			userService:    userService,
			PosthogClient:  PosthogClient,
		},
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
	ProdAppCount            int    `json:"prodAppCount,omitempty"`
	NonProdAppCount         int    `json:"nonProdAppCount,omitempty"`
	UserCount               int    `json:"userCount,omitempty"`
	EnvironmentCount        int    `json:"environmentCount,omitempty"`
	ClusterCount            int    `json:"clusterCount,omitempty"`
	CiCountPerDay           int    `json:"ciCountPerDay,omitempty"`
	CdCountPerDay           int    `json:"cdCountPerDay,omitempty"`
	HelmChartCount          int    `json:"helmChartCount,omitempty"`
	SecurityScanCountPerDay int    `json:"securityScanCountPerDay,omitempty"`
	DevtronVersion          string `json:"devtronVersion,omitempty"`
	DevtronMode             string `json:"devtronMode,omitempty"`
}

func (impl *TelemetryEventClientImplExtended) SummaryEventForTelemetry() {
	ucid, err := impl.getUCID()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	if IsOptOut {
		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
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

	devtronVersion := util.GetDevtronVersion()

	//if devtronVersion.ServerMode == "FULL" {
	//	summary := &SummaryDto{
	//		ProdAppCount:     prodApps,
	//		NonProdAppCount:  nonProdApps,
	//		UserCount:        len(users),
	//		EnvironmentCount: len(environments),
	//		ClusterCount:     len(clusters),
	//		CiCountPerDay:    len(ciPipeline),
	//		CdCountPerDay:    len(cdPipeline),
	//		DevtronVersion:   devtronVersion.GitCommit,
	//		DevtronMode:      devtronVersion.ServerMode,
	//	}
	//	payload.Summary = summary
	//}else{
	//	summary := &SummaryEA{
	//		UserCount:      len(users),
	//		ClusterCount:   len(clusters),
	//		DevtronVersion: devtronVersion.GitCommit,
	//		DevtronMode:    devtronVersion.ServerMode,
	//	}
	//}
	summary := &SummaryDto{
		ProdAppCount:     prodApps,
		NonProdAppCount:  nonProdApps,
		UserCount:        len(users),
		EnvironmentCount: len(environments),
		ClusterCount:     len(clusters),
		CiCountPerDay:    len(ciPipeline),
		CdCountPerDay:    len(cdPipeline),
		DevtronVersion:   devtronVersion.GitCommit,
		DevtronMode:      devtronVersion.ServerMode,
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

	if impl.PosthogClient.Client == nil {
		impl.logger.Warn("no posthog client found, creating new")
		client, err := impl.retryPosthogClient(PosthogApiKey, PosthogEndpoint)
		if err == nil {
			impl.PosthogClient.Client = client
		}
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

//func (impl *TelemetryEventClientImpl) HeartbeatEventForTelemetry() {
//	ucid, err := impl.getUCID()
//	if err != nil {
//		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
//		return
//	}
//	if IsOptOut {
//		impl.logger.Warnw("client is opt-out for telemetry, there will be no events capture", "ucid", ucid)
//		return
//	}
//
//	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
//	if err != nil {
//		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
//		return
//	}
//	k8sServerVersion, err := discoveryClient.ServerVersion()
//	if err != nil {
//		impl.logger.Errorw("exception caught inside telemetry heartbeat event", "err", err)
//		return
//	}
//	payload := &TelemetryEventDto{UCID: ucid, Timestamp: time.Now(), EventType: Heartbeat, DevtronVersion: "v1"}
//	payload.ServerVersion = k8sServerVersion.String()
//	reqBody, err := json.Marshal(payload)
//	if err != nil {
//		impl.logger.Errorw("HeartbeatEventForTelemetry, payload marshal error", "error", err)
//		return
//	}
//	prop := make(map[string]interface{})
//	err = json.Unmarshal(reqBody, &prop)
//	if err != nil {
//		impl.logger.Errorw("HeartbeatEventForTelemetry, payload unmarshal error", "error", err)
//		return
//	}
//
//	if impl.PosthogClient.Client == nil {
//		impl.logger.Warn("no posthog client found, creating new")
//		client, err := impl.retryPosthogClient(PosthogApiKey, PosthogEndpoint)
//		if err == nil {
//			impl.PosthogClient.Client = client
//		}
//	}
//	if impl.PosthogClient.Client != nil {
//		err = impl.PosthogClient.Client.Enqueue(posthog.Capture{
//			DistinctId: ucid,
//			Event:      string(Heartbeat),
//			Properties: prop,
//		})
//		if err != nil {
//			impl.logger.Errorw("HeartbeatEventForTelemetry, failed to push event", "error", err)
//		}
//	}
//}
