package telemetry

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/sso"
	"github.com/devtron-labs/devtron/pkg/user"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type TelemetryEventClientImplExtended struct {
	environmentService            cluster.EnvironmentService
	appListingRepository          repository.AppListingRepository
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	pipelineRepository            pipelineConfig.PipelineRepository
	gitHostRepository             repository.GitHostRepository
	gitProviderRepository         repository.GitProviderRepository
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository
	appRepository                 app.AppRepository
	ciWorkflowRepository          pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	*TelemetryEventClientImpl
}

func NewTelemetryEventClientImplExtended(logger *zap.SugaredLogger, client *http.Client, clusterService cluster.ClusterService,
	K8sUtil *util2.K8sUtil, aCDAuthConfig *util3.ACDAuthConfig,
	environmentService cluster.EnvironmentService, userService user.UserService,
	appListingRepository repository.AppListingRepository, PosthogClient *PosthogClient,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	gitHostRepository repository.GitHostRepository, gitProviderRepository repository.GitProviderRepository,
	attributeRepo repository.AttributesRepository, ssoLoginService sso.SSOLoginService, appRepository app.AppRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository) (*TelemetryEventClientImplExtended, error) {

	cron := cron.New(
		cron.WithChain())
	cron.Start()
	watcher := &TelemetryEventClientImplExtended{
		environmentService:            environmentService,
		appListingRepository:          appListingRepository,
		ciPipelineRepository:          ciPipelineRepository,
		pipelineRepository:            pipelineRepository,
		gitHostRepository:             gitHostRepository,
		gitProviderRepository:         gitProviderRepository,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
		appRepository:                 appRepository,
		cdWorkflowRepository:          cdWorkflowRepository,
		ciWorkflowRepository:          ciWorkflowRepository,
		TelemetryEventClientImpl: &TelemetryEventClientImpl{
			cron:            cron,
			logger:          logger,
			client:          client,
			clusterService:  clusterService,
			K8sUtil:         K8sUtil,
			aCDAuthConfig:   aCDAuthConfig,
			userService:     userService,
			attributeRepo:   attributeRepo,
			ssoLoginService: ssoLoginService,
			PosthogClient:   PosthogClient,
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
	UCID         string             `json:"ucid"` //unique client id
	Timestamp    time.Time          `json:"timestamp"`
	EventMessage string             `json:"eventMessage,omitempty"`
	EventType    TelemetryEventType `json:"eventType"`
	//Summary        *SummaryDto        `json:"summary,omitempty"`
	ProdAppCount            int    `json:"prodAppCount,omitempty"`
	NonProdAppCount         int    `json:"nonProdAppCount,omitempty"`
	UserCount               int    `json:"userCount,omitempty"`
	EnvironmentCount        int    `json:"environmentCount,omitempty"`
	ClusterCount            int    `json:"clusterCount,omitempty"`
	CiCountPerDay           int    `json:"ciCountPerDay,omitempty"`
	CdCountPerDay           int    `json:"cdCountPerDay,omitempty"`
	HelmChartCount          int    `json:"helmChartCount,omitempty"`
	SecurityScanCountPerDay int    `json:"securityScanCountPerDay,omitempty"`
	GitAccountsCount        int    `json:"gitAccountsCount,omitempty"`
	GitOpsCount             int    `json:"gitOpsCount,omitempty"`
	RegistryCount           int    `json:"registryCount,omitempty"`
	HostURL                 bool   `json:"hostURL,omitempty"`
	SSOLogin                bool   `json:"ssoLogin,omitempty"`
	CiSetup                 bool   `json:"ciSetup,omitempty"`
	CdSetup                 bool   `json:"cdSetup,omitempty"`
	AppSetup                bool   `json:"appSetup,omitempty"`
	Build                   bool   `json:"build,omitempty"`
	Deployment              bool   `json:"deployment,omitempty"`
	ServerVersion           string `json:"serverVersion,omitempty"`
	DevtronGitVersion       string `json:"devtronGitVersion,omitempty"`
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

	clusters, users, k8sServerVersion, hostURL, ssoSetup := impl.SummaryDetailsForTelemetry()
	payload := &TelemetryEventDto{UCID: ucid, Timestamp: time.Now(), EventType: Summary, DevtronVersion: "v1"}
	payload.ServerVersion = k8sServerVersion.String()

	environments, err := impl.environmentService.GetAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
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

	gitAccounts, err := impl.gitProviderRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	gitOps, err := impl.gitHostRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	containerRegistry, err := impl.dockerArtifactStoreRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	appSetup := false
	applications, err := impl.appRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
	}

	if len(applications) > 0 {
		appSetup = true
	}

	ciSetup, err := impl.ciPipelineRepository.Exists()

	cdSetup, err := impl.pipelineRepository.Exists()

	build, err := impl.ciWorkflowRepository.ExistsByStatus("Succeeded")

	deployment, err := impl.cdWorkflowRepository.ExistsByStatus("Healthy")

	devtronVersion := util.GetDevtronVersion()
	payload.ProdAppCount = prodApps
	payload.NonProdAppCount = nonProdApps
	payload.RegistryCount = len(containerRegistry)
	payload.SSOLogin = ssoSetup
	payload.GitOpsCount = len(gitOps)
	payload.UserCount = len(users)
	payload.EnvironmentCount = len(environments)
	payload.ClusterCount = len(clusters)
	payload.CiCountPerDay = len(ciPipeline)
	payload.CdCountPerDay = len(cdPipeline)
	payload.GitAccountsCount = len(gitAccounts)
	payload.GitOpsCount = len(gitOps)
	payload.HostURL = hostURL
	payload.DevtronGitVersion = devtronVersion.GitCommit
	payload.CiSetup = ciSetup
	payload.CdSetup = cdSetup
	payload.AppSetup = appSetup
	payload.Build = build
	payload.Deployment = deployment

	payload.DevtronMode = devtronVersion.ServerMode
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
