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

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util3 "github.com/devtron-labs/devtron/pkg/util"

	application2 "github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	. "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/commonService"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/util"
	util "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	errors2 "github.com/juju/errors"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppServiceImpl struct {
	environmentConfigRepository      chartConfig.EnvConfigOverrideRepository
	pipelineOverrideRepository       chartConfig.PipelineOverrideRepository
	mergeUtil                        *MergeUtil
	logger                           *zap.SugaredLogger
	ciArtifactRepository             repository.CiArtifactRepository
	pipelineRepository               pipelineConfig.PipelineRepository
	gitFactory                       *GitFactory
	dbMigrationConfigRepository      pipelineConfig.DbMigrationConfigRepository
	eventClient                      client.EventClient
	eventFactory                     client.EventFactory
	acdClient                        application.ServiceClient
	tokenCache                       *util3.TokenCache
	acdAuthConfig                    *util3.ACDAuthConfig
	enforcer                         casbin.Enforcer
	enforcerUtil                     rbac.EnforcerUtil
	user                             user.UserService
	appListingRepository             repository.AppListingRepository
	appRepository                    app.AppRepository
	envRepository                    repository2.EnvironmentRepository
	pipelineConfigRepository         chartConfig.PipelineConfigRepository
	configMapRepository              chartConfig.ConfigMapRepository
	chartRepository                  chartRepoRepository.ChartRepository
	appRepo                          app.AppRepository
	appLevelMetricsRepository        repository.AppLevelMetricsRepository
	envLevelMetricsRepository        repository.EnvLevelAppMetricsRepository
	ciPipelineMaterialRepository     pipelineConfig.CiPipelineMaterialRepository
	cdWorkflowRepository             pipelineConfig.CdWorkflowRepository
	commonService                    commonService.CommonService
	imageScanDeployInfoRepository    security.ImageScanDeployInfoRepository
	imageScanHistoryRepository       security.ImageScanHistoryRepository
	ArgoK8sClient                    argocdServer.ArgoK8sClient
	gitOpsRepository                 repository.GitOpsConfigRepository
	pipelineStrategyHistoryService   history2.PipelineStrategyHistoryService
	configMapHistoryService          history2.ConfigMapHistoryService
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService
	chartTemplateService             ChartTemplateService
}

type AppService interface {
	TriggerRelease(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time) (id int, err error)
	UpdateReleaseStatus(request *bean.ReleaseStatusUpdateRequest) (bool, error)
	UpdateApplicationStatusAndCheckIsHealthy(application v1alpha1.Application) (bool, error)
	TriggerCD(artifact *repository.CiArtifact, cdWorkflowId int, pipeline *pipelineConfig.Pipeline, async bool, triggeredAt time.Time) error
	GetConfigMapAndSecretJson(appId int, envId int, pipelineId int) ([]byte, error)
	UpdateCdWorkflowRunnerByACDObject(app v1alpha1.Application, cdWorkflowId int) error
	GetCmSecretNew(appId int, envId int) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error)
	MarkImageScanDeployed(appId int, envId int, imageDigest string, clusterId int) error
	GetChartRepoName(gitRepoUrl string) string
}

func NewAppService(
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	mergeUtil *MergeUtil,
	logger *zap.SugaredLogger,
	ciArtifactRepository repository.CiArtifactRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	dbMigrationConfigRepository pipelineConfig.DbMigrationConfigRepository,
	eventClient client.EventClient,
	eventFactory client.EventFactory, acdClient application.ServiceClient,
	cache *util3.TokenCache, authConfig *util3.ACDAuthConfig,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, user user.UserService,
	appListingRepository repository.AppListingRepository,
	appRepository app.AppRepository,
	envRepository repository2.EnvironmentRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository, configMapRepository chartConfig.ConfigMapRepository,
	appLevelMetricsRepository repository.AppLevelMetricsRepository, envLevelMetricsRepository repository.EnvLevelAppMetricsRepository,
	chartRepository chartRepoRepository.ChartRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository, commonService commonService.CommonService,
	imageScanDeployInfoRepository security.ImageScanDeployInfoRepository, imageScanHistoryRepository security.ImageScanHistoryRepository,
	ArgoK8sClient argocdServer.ArgoK8sClient,
	gitFactory *GitFactory, gitOpsRepository repository.GitOpsConfigRepository,
	pipelineStrategyHistoryService history2.PipelineStrategyHistoryService,
	configMapHistoryService history2.ConfigMapHistoryService,
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService,
	chartTemplateService ChartTemplateService) *AppServiceImpl {
	appServiceImpl := &AppServiceImpl{
		environmentConfigRepository:      environmentConfigRepository,
		mergeUtil:                        mergeUtil,
		pipelineOverrideRepository:       pipelineOverrideRepository,
		logger:                           logger,
		ciArtifactRepository:             ciArtifactRepository,
		pipelineRepository:               pipelineRepository,
		dbMigrationConfigRepository:      dbMigrationConfigRepository,
		eventClient:                      eventClient,
		eventFactory:                     eventFactory,
		acdClient:                        acdClient,
		tokenCache:                       cache,
		acdAuthConfig:                    authConfig,
		enforcer:                         enforcer,
		enforcerUtil:                     enforcerUtil,
		user:                             user,
		appListingRepository:             appListingRepository,
		appRepository:                    appRepository,
		envRepository:                    envRepository,
		pipelineConfigRepository:         pipelineConfigRepository,
		configMapRepository:              configMapRepository,
		chartRepository:                  chartRepository,
		appLevelMetricsRepository:        appLevelMetricsRepository,
		envLevelMetricsRepository:        envLevelMetricsRepository,
		ciPipelineMaterialRepository:     ciPipelineMaterialRepository,
		cdWorkflowRepository:             cdWorkflowRepository,
		commonService:                    commonService,
		imageScanDeployInfoRepository:    imageScanDeployInfoRepository,
		imageScanHistoryRepository:       imageScanHistoryRepository,
		ArgoK8sClient:                    ArgoK8sClient,
		gitFactory:                       gitFactory,
		gitOpsRepository:                 gitOpsRepository,
		pipelineStrategyHistoryService:   pipelineStrategyHistoryService,
		configMapHistoryService:          configMapHistoryService,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		chartTemplateService:             chartTemplateService,
	}
	return appServiceImpl
}

func (impl AppServiceImpl) UpdateReleaseStatus(updateStatusRequest *bean.ReleaseStatusUpdateRequest) (bool, error) {
	count, err := impl.pipelineOverrideRepository.UpdateStatusByRequestIdentifier(updateStatusRequest.RequestId, updateStatusRequest.NewStatus)
	if err != nil {
		impl.logger.Errorw("error in updating release status", "request", updateStatusRequest, "error", err)
		return false, err
	}
	return count == 1, nil
}

func (impl AppServiceImpl) UpdateApplicationStatusAndCheckIsHealthy(app v1alpha1.Application) (bool, error) {
	isHealthy := false
	repoUrl := app.Spec.Source.RepoURL
	// backward compatibility for updating application status - if unable to find app check it in charts
	chart, err := impl.chartRepository.FindByGirRepoUrl(repoUrl)
	if err != nil {
		impl.logger.Errorw("error in fetching chart", "err", err, "chart", chart.ChartName)
		return isHealthy, err
	}
	dbApp, err := impl.appRepository.FindById(chart.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching app", "err", err, "app", chart.AppId)
		return isHealthy, err
	}
	// extract environment name from argocd app
	evnName := strings.ReplaceAll(app.Name, dbApp.AppName+"-", "")
	appName := dbApp.AppName

	if dbApp.Id > 0 && dbApp.AppStore == true {
		impl.logger.Debugw("skipping application status update as this app is chart", "appName", appName)
		return isHealthy, nil
	}

	deploymentStatus, err := impl.appListingRepository.FindLastDeployedStatus(app.Name)
	if err != nil && !IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching deployment status", "err", err)
		return isHealthy, err
	}

	if string(application.Healthy) != deploymentStatus.Status {
		if deploymentStatus.Status == app.Status.Health.Status {
			impl.logger.Debug("not updating same statuses from " + deploymentStatus.Status + " to " + app.Status.Health.Status)
			return isHealthy, nil
		}

		gitHash := ""
		if app.Operation != nil && app.Operation.Sync != nil {
			gitHash = app.Operation.Sync.Revision
		} else if app.Status.OperationState != nil && app.Status.OperationState.Operation.Sync != nil {
			gitHash = app.Status.OperationState.Operation.Sync.Revision
		}
		pipelineOverride, err := impl.pipelineOverrideRepository.FindByPipelineTriggerGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error on update application status", "pipelineOverride", pipelineOverride, "CdWorkflowId", pipelineOverride.CdWorkflowId, "app", app, "err", err)
			return isHealthy, err
		}

		if pipelineOverride.Pipeline.AppId != dbApp.Id {
			impl.logger.Warnw("event received for other deleted app", "gitHash", gitHash, "pipelineOverride", pipelineOverride, "dbApp", dbApp)
			return isHealthy, nil
		}

		releaseCounter, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(pipelineOverride.PipelineId)
		if err != nil {
			impl.logger.Errorw("error on update application status", "releaseCounter", releaseCounter, "CdWorkflowId", pipelineOverride.CdWorkflowId, "app", app, "err", err)
			return isHealthy, err
		}
		if pipelineOverride.PipelineReleaseCounter == releaseCounter {
			impl.logger.Debug("inserting new app status: " + string(app.Status.Health.Status))
			newDeploymentStatus := &repository.DeploymentStatus{
				AppName:   app.Name,
				AppId:     deploymentStatus.AppId,
				EnvId:     deploymentStatus.EnvId,
				Status:    app.Status.Health.Status,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
			}
			dbConnection := impl.pipelineRepository.GetConnection()
			tx, err := dbConnection.Begin()
			if err != nil {
				impl.logger.Errorw("error on update status, db get txn failed", "CdWorkflowId", pipelineOverride.CdWorkflowId, "app", app, "err", err)
				return isHealthy, err
			}
			// Rollback tx on error.
			defer tx.Rollback()

			err = impl.appListingRepository.SaveNewDeployment(newDeploymentStatus, tx)
			if err != nil {
				impl.logger.Errorw("error on saving new deployment status for wf", "CdWorkflowId", pipelineOverride.CdWorkflowId, "app", app, "err", err)
				return isHealthy, err
			}
			err = impl.UpdateCdWorkflowRunnerByACDObject(app, pipelineOverride.CdWorkflowId)
			if err != nil {
				impl.logger.Errorw("error on update cd workflow runner", "CdWorkflowId", pipelineOverride.CdWorkflowId, "app", app, "err", err)
				return isHealthy, err
			}
			err = tx.Commit()
			if err != nil {
				impl.logger.Errorw("error on db transaction commit for", "CdWorkflowId", pipelineOverride.CdWorkflowId, "app", app, "err", err)
				return isHealthy, err
			}
			if string(application.Healthy) == newDeploymentStatus.Status {
				isHealthy = true
				go impl.WriteCDSuccessEvent(newDeploymentStatus.AppId, appName, newDeploymentStatus.EnvId, evnName, pipelineOverride)
			}
		} else {
			impl.logger.Debug("event received for older triggered revision: " + gitHash)
		}
	}
	return isHealthy, nil
}

func (impl *AppServiceImpl) WriteCDSuccessEvent(appId int, appName string, envId int, envName string, override *chartConfig.PipelineOverride) {
	event := impl.eventFactory.Build(util.Success, &override.PipelineId, appId, &envId, util.CD)
	impl.logger.Debugw("event WriteCDSuccessEvent", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, nil, override.Id, bean.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *AppServiceImpl) BuildCDSuccessPayload(appName string, environmentName string) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = appName
	payload.EnvName = environmentName
	return payload
}

type EnvironmentOverride struct {
	Enabled   bool        `json:"enabled"`
	EnvValues []*KeyValue `json:"envValues"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (conf *EnvironmentOverride) appendEnvironmentVariable(key, value string) {
	item := &KeyValue{Key: key, Value: value}
	conf.EnvValues = append(conf.EnvValues, item)
}

func (impl *AppServiceImpl) TriggerCD(artifact *repository.CiArtifact, cdWorkflowId int, pipeline *pipelineConfig.Pipeline, async bool, triggeredAt time.Time) error {
	impl.logger.Debugw("automatic pipeline trigger attempt async", "artifactId", artifact.Id)

	return impl.triggerReleaseAsync(artifact, cdWorkflowId, pipeline, triggeredAt)
}

func (impl *AppServiceImpl) triggerReleaseAsync(artifact *repository.CiArtifact, cdWorkflowId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	err := impl.validateAndTrigger(pipeline, artifact, cdWorkflowId, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in trigger for pipeline", "pipelineId", strconv.Itoa(pipeline.Id))
	}
	impl.logger.Debugw("trigger attempted for all pipeline ", "artifactId", artifact.Id)
	return err
}

func (impl AppServiceImpl) validateAndTrigger(p *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId int, triggeredAt time.Time) error {
	object := impl.enforcerUtil.GetAppRBACNameByAppId(p.AppId)
	envApp := strings.Split(object, "/")
	if len(envApp) != 2 {
		impl.logger.Error("invalid req, app and env not found from rbac")
		return errors.New("invalid req, app and env not found from rbac")
	}
	err := impl.releasePipeline(p, artifact, cdWorkflowId, triggeredAt)
	return err
}

func (impl AppServiceImpl) releasePipeline(pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId int, triggeredAt time.Time) error {
	impl.logger.Debugw("triggering release for ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id)
	//Iterate for each even if there is error in one
	request := &bean.ValuesOverrideRequest{
		PipelineId:   pipeline.Id,
		UserId:       artifact.CreatedBy,
		CiArtifactId: artifact.Id,
		AppId:        pipeline.AppId,
		CdWorkflowId: cdWorkflowId,
		ForceTrigger: true,
	}

	ctx, err := impl.buildACDSynchContext()
	if err != nil {
		impl.logger.Errorw("error in creating acd synch context", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
		return err
	}

	id, err := impl.TriggerRelease(request, ctx, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in auto  cd pipeline trigger", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
	} else {
		impl.logger.Infow("pipeline successfully triggered ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id, "releaseId", id)
	}
	return err
}

func (impl AppServiceImpl) buildACDSynchContext() (acdContext context.Context, err error) {
	return impl.tokenCache.BuildACDSynchContext()
	/*token, found := impl.tokenCache.Cache.Get("token")
	if !found {
		token, err := impl.userAuthService.HandleLogin(impl.acdAuthConfig.ACDUsername, impl.acdAuthConfig.ACDPassword)
		if err != nil {
			impl.logger.Errorw("error while acd login", "err", err)
			return nil, err
		}
		impl.tokenCache.Cache.Set("token", token, cache.NoExpiration)
	}
	token, _ = impl.tokenCache.Cache.Get("token")
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", token)
	return ctx, nil*/
}

func (impl AppServiceImpl) getDbMigrationOverride(overrideRequest *bean.ValuesOverrideRequest, artifact *repository.CiArtifact, isRollback bool) (overrideJson []byte, err error) {
	if isRollback {
		return nil, fmt.Errorf("rollback not supported ye")
	}
	notConfigured := false
	config, err := impl.dbMigrationConfigRepository.FindByPipelineId(overrideRequest.PipelineId)
	if err != nil && !IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching pipeline override config", "req", overrideRequest, "err", err)
		return nil, err
	} else if IsErrNoRows(err) {
		notConfigured = true
	}
	envVal := &EnvironmentOverride{}
	if notConfigured {
		impl.logger.Warnw("no active db migration found", "pipeline", overrideRequest.PipelineId)
		envVal.Enabled = false
	} else {
		materialInfos, err := artifact.ParseMaterialInfo()
		if err != nil {
			return nil, err
		}

		hash, ok := materialInfos[config.GitMaterial.Url]
		if !ok {
			impl.logger.Errorf("wrong url map ", "map", materialInfos, "url", config.GitMaterial.Url)
			return nil, fmt.Errorf("configured url not found in material %s", config.GitMaterial.Url)
		}

		envVal.Enabled = true
		if config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_USERNAME_PASSWORD &&
			config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_ACCESS_TOKEN &&
			config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_ANONYMOUS {
			return nil, fmt.Errorf("auth mode %s not supported for migration", config.GitMaterial.GitProvider.AuthMode)
		}
		envVal.appendEnvironmentVariable("GIT_REPO_URL", config.GitMaterial.Url)
		envVal.appendEnvironmentVariable("GIT_USER", config.GitMaterial.GitProvider.UserName)
		var password string
		if config.GitMaterial.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
			password = config.GitMaterial.GitProvider.Password
		} else {
			password = config.GitMaterial.GitProvider.AccessToken
		}
		envVal.appendEnvironmentVariable("GIT_AUTH_TOKEN", password)
		// parse git-tag not required
		//envVal.appendEnvironmentVariable("GIT_TAG", "")
		envVal.appendEnvironmentVariable("GIT_HASH", hash)
		envVal.appendEnvironmentVariable("SCRIPT_LOCATION", config.ScriptSource)
		envVal.appendEnvironmentVariable("DB_TYPE", string(config.DbConfig.Type))
		envVal.appendEnvironmentVariable("DB_USER_NAME", config.DbConfig.UserName)
		envVal.appendEnvironmentVariable("DB_PASSWORD", config.DbConfig.Password)
		envVal.appendEnvironmentVariable("DB_HOST", config.DbConfig.Host)
		envVal.appendEnvironmentVariable("DB_PORT", config.DbConfig.Port)
		envVal.appendEnvironmentVariable("DB_NAME", config.DbConfig.DbName)
		//Will be used for rollback don't delete it
		//envVal.appendEnvironmentVariable("MIGRATE_TO_VERSION", strconv.Itoa(overrideRequest.TargetDbVersion))
	}
	dbMigrationConfig := map[string]interface{}{"dbMigrationConfig": envVal}
	confByte, err := json.Marshal(dbMigrationConfig)
	if err != nil {
		return nil, err
	}
	return confByte, nil
}

func (impl AppServiceImpl) TriggerRelease(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time) (id int, err error) {
	if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	pipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	if err != nil {
		impl.logger.Errorf("invalid req", "err", err, "req", overrideRequest)
		return 0, err
	}
	envOverride, err := impl.environmentConfigRepository.ActiveEnvConfigOverride(overrideRequest.AppId, pipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorf("invalid state", "err", err, "req", overrideRequest)
		return 0, err
	}

	if envOverride.Id == 0 {
		chart, err := impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
		if err != nil {
			impl.logger.Errorf("invalid state", "err", err, "req", overrideRequest)
			return 0, err
		}
		envOverride, err = impl.environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId(overrideRequest.AppId, pipeline.EnvironmentId, chart.ChartRefId)
		if err != nil && !errors2.IsNotFound(err) {
			impl.logger.Errorf("invalid state", "err", err, "req", overrideRequest)
			return 0, err
		}

		//creating new env override config
		if errors2.IsNotFound(err) || envOverride == nil {
			environment, err := impl.envRepository.FindById(pipeline.EnvironmentId)
			if err != nil && !IsErrNoRows(err) {
				return 0, err
			}
			envOverride = &chartConfig.EnvConfigOverride{
				Active:            true,
				ManualReviewed:    true,
				Status:            models.CHARTSTATUS_SUCCESS,
				TargetEnvironment: pipeline.EnvironmentId,
				ChartId:           chart.Id,
				AuditLog:          sql.AuditLog{UpdatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId},
				Namespace:         environment.Namespace,
				IsOverride:        false,
				EnvOverrideValues: "{}",
				Latest:            false,
			}
			err = impl.environmentConfigRepository.Save(envOverride)
			if err != nil {
				impl.logger.Errorw("error in creating envconfig", "data", envOverride, "error", err)
				return 0, err
			}
		}

		env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
		if err != nil {
			impl.logger.Errorw("unable to find env", "err", err)
			return 0, err
		}
		envOverride.Environment = env
		envOverride.Chart = chart
	} else if envOverride.Id > 0 && !envOverride.IsOverride {
		chart, err := impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
		if err != nil {
			impl.logger.Errorf("invalid state", "err", err, "req", overrideRequest)
			return 0, err
		}
		envOverride.Chart = chart
	}

	artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
	if err != nil {
		return 0, err
	}
	materialInfoMap, mErr := artifact.ParseMaterialInfo()
	if mErr != nil {
		impl.logger.Errorw("material info map error", mErr)
		return 0, err
	}

	//FIXME: how to determine rollback
	//we can't depend on ciArtifact ID because CI pipeline can be manually triggered in any order regardless of sourcecode status
	dbMigrationOverride, err := impl.getDbMigrationOverride(overrideRequest, artifact, false)
	if err != nil {
		impl.logger.Errorw("error in fetching db migration config", "req", overrideRequest, "err", err)
		return 0, err
	}

	//fetch pipeline config from strategy table, if pipeline is automatic fetch always default, else depends on request
	var strategy *chartConfig.PipelineStrategy
	//forceTrigger true if CD triggered Auto, triggered occurred from CI
	if overrideRequest.ForceTrigger {
		strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
	} else {
		var deploymentTemplate pipelineConfig.DeploymentTemplate
		if overrideRequest.DeploymentTemplate == "ROLLING" {
			deploymentTemplate = pipelineConfig.DEPLOYMENT_TEMPLATE_ROLLING
		} else if overrideRequest.DeploymentTemplate == "BLUE-GREEN" {
			deploymentTemplate = pipelineConfig.DEPLOYMENT_TEMPLATE_BLUE_GREEN
		} else if overrideRequest.DeploymentTemplate == "CANARY" {
			deploymentTemplate = pipelineConfig.DEPLOYMENT_TEMPLATE_CANARY
		} else if overrideRequest.DeploymentTemplate == "RECREATE" {
			deploymentTemplate = pipelineConfig.DEPLOYMENT_TEMPLATE_RECREATE
		}

		if len(deploymentTemplate) > 0 {
			strategy, err = impl.pipelineConfigRepository.FindByStrategyAndPipelineId(deploymentTemplate, overrideRequest.PipelineId)
		} else {
			strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
		}
	}
	if err != nil && errors2.IsNotFound(err) == false {
		impl.logger.Errorf("invalid state", "err", err, "req", strategy)
		return 0, err
	}

	valid, err := impl.validateVersionForStrategy(envOverride, strategy)
	if err != nil || !valid {
		impl.logger.Errorw("error in validating pipeline strategy ", "strategy", strategy.Strategy, "err", err)
		return 0, err
	}

	chartVersion := envOverride.Chart.ChartVersion
	configMapJson, err := impl.getConfigMapAndSecretJsonV2(overrideRequest.AppId, envOverride.TargetEnvironment, overrideRequest.PipelineId, chartVersion)
	if err != nil {
		impl.logger.Errorw("error in fetching config map n secret ", "err", err)
		configMapJson = nil
	}

	releaseId, pipelineOverrideId, saveErr := impl.mergeAndSave(envOverride, overrideRequest, dbMigrationOverride, artifact, pipeline, configMapJson, strategy, ctx, triggeredAt)
	if releaseId != 0 {
		flag, err := impl.updateArgoPipeline(overrideRequest.AppId, pipeline.Name, envOverride, ctx)
		if err != nil {
			impl.logger.Errorw("error in updating argocd  app ", "err", err)
			return 0, err
		}
		if flag {
			impl.logger.Debug("argocd successfully updated")
		} else {
			impl.logger.Debug("argocd failed to update, ignoring it")
		}

		deploymentStatus := &repository.DeploymentStatus{
			AppName:   pipeline.App.AppName + "-" + envOverride.Environment.Name,
			AppId:     pipeline.AppId,
			EnvId:     pipeline.EnvironmentId,
			Status:    repository.NewDeployment,
			CreatedOn: triggeredAt,
			UpdatedOn: triggeredAt,
		}
		dbConnection := impl.pipelineRepository.GetConnection()
		tx, err := dbConnection.Begin()
		if err != nil {
			return 0, err
		}
		// Rollback tx on error.
		defer tx.Rollback()
		err = impl.appListingRepository.SaveNewDeployment(deploymentStatus, tx)
		if err != nil {
			impl.logger.Errorw("error in saving new deployment history", "req", overrideRequest, "err", err)
			return 0, err
		}
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
		impl.synchCD(pipeline, ctx, overrideRequest, envOverride)
		go impl.WriteCDTriggerEvent(overrideRequest, pipeline, envOverride, materialInfoMap, artifact, releaseId, pipelineOverrideId)
		if artifact.ScanEnabled {
			_ = impl.MarkImageScanDeployed(overrideRequest.AppId, envOverride.TargetEnvironment, artifact.ImageDigest, pipeline.Environment.ClusterId)
		}
	}
	middleware.CdTriggerCounter.WithLabelValues(strconv.Itoa(pipeline.AppId), strconv.Itoa(pipeline.EnvironmentId), strconv.Itoa(pipeline.Id)).Inc()
	return releaseId, saveErr
}

func (impl AppServiceImpl) MarkImageScanDeployed(appId int, envId int, imageDigest string, clusterId int) error {
	impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "imageDigest", imageDigest)
	executionHistory, err := impl.imageScanHistoryRepository.FindByImageDigest(imageDigest)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching execution history", "err", err)
		return err
	}
	if executionHistory == nil || executionHistory.Id == 0 {
		impl.logger.Errorw("no execution history found for digest", "digest", imageDigest)
		return fmt.Errorf("no execution history found for digest - %s", imageDigest)
	}
	impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "executionHistory", executionHistory)
	var ids []int
	ids = append(ids, executionHistory.Id)

	ot, err := impl.imageScanDeployInfoRepository.FindByTypeMetaAndTypeId(appId, security.ScanObjectType_APP) //todo insure this touple unique in db
	if err != nil && err != pg.ErrNoRows {
		return err
	} else if err == pg.ErrNoRows {
		imageScanDeployInfo := &security.ImageScanDeployInfo{
			ImageScanExecutionHistoryId: ids,
			ScanObjectMetaId:            appId,
			ObjectType:                  security.ScanObjectType_APP,
			EnvId:                       envId,
			ClusterId:                   clusterId,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: 1,
				UpdatedOn: time.Now(),
				UpdatedBy: 1,
			},
		}
		impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "imageScanDeployInfo", imageScanDeployInfo)
		err = impl.imageScanDeployInfoRepository.Save(imageScanDeployInfo)
		if err != nil {
			impl.logger.Errorw("error in creating deploy info", "err", err)
		}
	} else {
		impl.logger.Debugw("pt", "ot", ot)
	}
	return err
}

func (impl AppServiceImpl) validateVersionForStrategy(envOverride *chartConfig.EnvConfigOverride, strategy *chartConfig.PipelineStrategy) (bool, error) {
	chartVersion := envOverride.Chart.ChartVersion
	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return false, err
	}

	if (chartMajorVersion <= 3 && chartMinorVersion < 2) &&
		(strategy.Strategy == pipelineConfig.DEPLOYMENT_TEMPLATE_CANARY || strategy.Strategy == pipelineConfig.DEPLOYMENT_TEMPLATE_RECREATE) {
		err = &ApiError{
			Code:            "422",
			InternalMessage: "incompatible chart for selected cd strategy:" + string(strategy.Strategy),
			UserMessage:     "incompatible chart for selected cd strategy:" + string(strategy.Strategy),
		}
		return false, err
	}
	return true, nil
}

//FIXME tmp workaround
func (impl AppServiceImpl) GetCmSecretNew(appId int, envId int) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error) {
	var configMapJson string
	var secretDataJson string
	var configMapJsonApp string
	var secretDataJsonApp string
	var configMapJsonEnv string
	var secretDataJsonEnv string
	//var configMapJsonPipeline string
	//var secretDataJsonPipeline string

	configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		return nil, nil, err
	}
	if configMapA != nil && configMapA.Id > 0 {
		configMapJsonApp = configMapA.ConfigMapData
		secretDataJsonApp = configMapA.SecretData
	}

	configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		return nil, nil, err
	}
	if configMapE != nil && configMapE.Id > 0 {
		configMapJsonEnv = configMapE.ConfigMapData
		secretDataJsonEnv = configMapE.SecretData
	}

	configMapJson, err = impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
	if err != nil {
		return nil, nil, err
	}

	chart, err := impl.commonService.FetchLatestChart(appId, envId)
	if err != nil {
		return nil, nil, err
	}
	chartVersion := chart.ChartVersion
	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, nil, err
	}
	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion)
	if err != nil {
		return nil, nil, err
	}
	configResponse := bean.ConfigMapJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(configMapJson), &configResponse)
		if err != nil {
			return nil, nil, err
		}
	}
	secretResponse := bean.ConfigSecretJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
		if err != nil {
			return nil, nil, err
		}
	}
	return &configResponse, &secretResponse, nil
}

//depricated
//TODO remove this method
func (impl AppServiceImpl) GetConfigMapAndSecretJson(appId int, envId int, pipelineId int) ([]byte, error) {
	var configMapJson string
	var secretDataJson string
	merged := []byte("{}")
	configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		return []byte("{}"), err
	}
	if configMapA != nil && configMapA.Id > 0 {
		configMapJson = configMapA.ConfigMapData
		secretDataJson = configMapA.SecretData
		if configMapJson == "" {
			configMapJson = "{}"
		}
		if secretDataJson == "" {
			secretDataJson = "{}"
		}
		config, err := impl.mergeUtil.JsonPatch([]byte(configMapJson), []byte(secretDataJson))
		if err != nil {
			return []byte("{}"), err
		}
		merged, err = impl.mergeUtil.JsonPatch(merged, config)
		if err != nil {
			return []byte("{}"), err
		}
	}

	configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		return []byte("{}"), err
	}
	if configMapE != nil && configMapE.Id > 0 {
		configMapJson = configMapE.ConfigMapData
		secretDataJson = configMapE.SecretData
		if configMapJson == "" {
			configMapJson = "{}"
		}
		if secretDataJson == "" {
			secretDataJson = "{}"
		}
		config, err := impl.mergeUtil.JsonPatch([]byte(configMapJson), []byte(secretDataJson))
		if err != nil {
			return []byte("{}"), err
		}
		merged, err = impl.mergeUtil.JsonPatch(merged, config)
		if err != nil {
			return []byte("{}"), err
		}
	}

	return merged, nil
}

func (impl AppServiceImpl) getConfigMapAndSecretJsonV2(appId int, envId int, pipelineId int, chartVersion string) ([]byte, error) {

	var configMapJson string
	var secretDataJson string
	var configMapJsonApp string
	var secretDataJsonApp string
	var configMapJsonEnv string
	var secretDataJsonEnv string
	//var configMapJsonPipeline string
	//var secretDataJsonPipeline string

	merged := []byte("{}")
	configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
	if err != nil && pg.ErrNoRows != err {
		return []byte("{}"), err
	}
	if configMapA != nil && configMapA.Id > 0 {
		configMapJsonApp = configMapA.ConfigMapData
		secretDataJsonApp = configMapA.SecretData
	}

	configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
	if err != nil && pg.ErrNoRows != err {
		return []byte("{}"), err
	}
	if configMapE != nil && configMapE.Id > 0 {
		configMapJsonEnv = configMapE.ConfigMapData
		secretDataJsonEnv = configMapE.SecretData
	}

	configMapJson, err = impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
	if err != nil {
		return []byte("{}"), err
	}
	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return []byte("{}"), err
	}
	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion)
	if err != nil {
		return []byte("{}"), err
	}
	configResponseR := bean.ConfigMapRootJson{}
	configResponse := bean.ConfigMapJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(configMapJson), &configResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	configResponseR.ConfigMapJson = configResponse
	secretResponseR := bean.ConfigSecretRootJson{}
	secretResponse := bean.ConfigSecretJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	secretResponseR.ConfigSecretJson = secretResponse

	configMapByte, err := json.Marshal(configResponseR)
	if err != nil {
		return []byte("{}"), err
	}
	secretDataByte, err := json.Marshal(secretResponseR)
	if err != nil {
		return []byte("{}"), err
	}

	merged, err = impl.mergeUtil.JsonPatch(configMapByte, secretDataByte)
	if err != nil {
		return []byte("{}"), err
	}
	return merged, nil
}

func (impl AppServiceImpl) synchCD(pipeline *pipelineConfig.Pipeline, ctx context.Context,
	overrideRequest *bean.ValuesOverrideRequest, envOverride *chartConfig.EnvConfigOverride) {
	req := new(application2.ApplicationSyncRequest)
	pipelineName := pipeline.App.AppName + "-" + envOverride.Environment.Name
	req.Name = &pipelineName
	req.Prune = true
	if ctx == nil {
		impl.logger.Errorw("err in syncing ACD, ctx is NULL", "pipelineId", overrideRequest.PipelineId)
		return
	}
	if _, err := impl.acdClient.Sync(ctx, req); err != nil {
		impl.logger.Errorw("err in syncing ACD", "pipelineId", overrideRequest.PipelineId, "err", err)
	}
}

func (impl *AppServiceImpl) WriteCDTriggerEvent(overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline,
	envOverride *chartConfig.EnvConfigOverride, materialInfoMap map[string]string, artifact *repository.CiArtifact, releaseId, pipelineOverrideId int) {
	event := impl.eventFactory.Build(util.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util.CD)
	impl.logger.Debugw("event WriteCDTriggerEvent", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, nil, pipelineOverrideId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	deploymentEvent := DeploymentEvent{
		ApplicationId:      pipeline.AppId,
		EnvironmentId:      pipeline.EnvironmentId, //check for production Environment
		ReleaseId:          releaseId,
		PipelineOverrideId: pipelineOverrideId,
		TriggerTime:        time.Now(),
		CiArtifactId:       overrideRequest.CiArtifactId,
	}
	ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(artifact.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in ")
	}
	for _, ciPipelineMaterial := range ciPipelineMaterials {
		hash := materialInfoMap[ciPipelineMaterial.GitMaterial.Url]
		pipelineMaterialInfo := &PipelineMaterialInfo{PipelineMaterialId: ciPipelineMaterial.Id, CommitHash: hash}
		deploymentEvent.PipelineMaterials = append(deploymentEvent.PipelineMaterials, pipelineMaterialInfo)
	}
	impl.logger.Infow("triggering deployment event", "event", deploymentEvent)
	err = impl.eventClient.WriteNatsEvent(util2.CD_SUCCESS, deploymentEvent)
	if err != nil {
		impl.logger.Errorw("error in writing cd trigger event", "err", err)
	}
}

type DeploymentEvent struct {
	ApplicationId      int
	EnvironmentId      int
	ReleaseId          int
	PipelineOverrideId int
	TriggerTime        time.Time
	PipelineMaterials  []*PipelineMaterialInfo
	CiArtifactId       int
}
type PipelineMaterialInfo struct {
	PipelineMaterialId int
	CommitHash         string
}

func buildCDTriggerEvent(impl *AppServiceImpl, overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline,
	envOverride *chartConfig.EnvConfigOverride, materialInfo map[string]string, artifact *repository.CiArtifact) client.Event {
	event := impl.eventFactory.Build(util.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util.CD)
	return event
}

func (impl *AppServiceImpl) BuildPayload(overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline,
	envOverride *chartConfig.EnvConfigOverride, materialInfo map[string]string, artifact *repository.CiArtifact) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = pipeline.App.AppName
	payload.PipelineName = pipeline.Name
	payload.EnvName = envOverride.Environment.Name

	var revision string
	for _, v := range materialInfo {
		revision = v
		break
	}
	payload.Source = url.PathEscape(revision)
	payload.DockerImageUrl = artifact.Image
	return payload
}

type ReleaseAttributes struct {
	Name           string
	Tag            string
	PipelineName   string
	ReleaseVersion string
	DeploymentType string
	App            string
	Env            string
	AppMetrics     *bool
}

func (impl AppServiceImpl) getReleaseOverride(envOverride *chartConfig.EnvConfigOverride,
	overrideRequest *bean.ValuesOverrideRequest,
	artifact *repository.CiArtifact,
	pipeline *pipelineConfig.Pipeline,
	pipelineOverride *chartConfig.PipelineOverride, strategy *chartConfig.PipelineStrategy) (releaseOverride string, err error) {

	artifactImage := artifact.Image
	imageTag := strings.Split(artifactImage, ":")

	appId := strconv.Itoa(pipeline.App.Id)
	envId := strconv.Itoa(pipeline.EnvironmentId)

	var appMetrics *bool
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(pipeline.AppId)
	if err != nil && !IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return "", &ApiError{InternalMessage: "unable to fetch app level metrics flag"}
	}
	appMetrics = &appLevelMetrics.AppMetrics

	envLevelMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil && !IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return "", &ApiError{InternalMessage: "unable to fetch env level metrics flag"}
	}
	if envLevelMetrics.Id != 0 && envLevelMetrics.AppMetrics != nil {
		appMetrics = envLevelMetrics.AppMetrics
	}

	releaseAttribute := ReleaseAttributes{
		Name:           imageTag[0],
		Tag:            imageTag[1],
		PipelineName:   pipeline.Name,
		ReleaseVersion: strconv.Itoa(pipelineOverride.PipelineReleaseCounter),
		DeploymentType: string(strategy.Strategy),
		App:            appId,
		Env:            envId,
		AppMetrics:     appMetrics,
	}
	override, err := util2.Tprintf(envOverride.Chart.ImageDescriptorTemplate, releaseAttribute)
	if err != nil {
		return "", &ApiError{InternalMessage: "unable to render ImageDescriptorTemplate"}
	}
	if overrideRequest.AdditionalOverride != nil {
		userOverride, err := overrideRequest.AdditionalOverride.MarshalJSON()
		if err != nil {
			return "", err
		}
		data, err := impl.mergeUtil.JsonPatch(userOverride, []byte(override))
		if err != nil {
			return "", err
		}
		override = string(data)
	}
	return override, nil
}

func (impl AppServiceImpl) GetChartRepoName(gitRepoUrl string) string {
	gitRepoUrl = gitRepoUrl[strings.LastIndex(gitRepoUrl, "/")+1:]
	chartRepoName := strings.ReplaceAll(gitRepoUrl, ".git", "")
	return chartRepoName
}

func (impl AppServiceImpl) mergeAndSave(envOverride *chartConfig.EnvConfigOverride,
	overrideRequest *bean.ValuesOverrideRequest,
	dbMigrationOverride []byte,
	artifact *repository.CiArtifact,
	pipeline *pipelineConfig.Pipeline, configMapJson []byte, strategy *chartConfig.PipelineStrategy, ctx context.Context,
	triggeredAt time.Time) (releaseId int, overrideId int, err error) {

	//register release , obtain release id TODO: populate releaseId to template
	override, err := impl.savePipelineOverride(overrideRequest, envOverride.Id, triggeredAt)
	if err != nil {
		return 0, 0, err
	}
	//TODO: check status and apply lock
	overrideJson, err := impl.getReleaseOverride(envOverride, overrideRequest, artifact, pipeline, override, strategy)
	if err != nil {
		return 0, 0, err
	}

	//merge three values on the fly
	//ordering is important here
	//global < environment < db< release
	var merged []byte
	if !envOverride.IsOverride {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.Chart.GlobalOverride))
		if err != nil {
			return 0, 0, err
		}
	} else {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.EnvOverrideValues))
		if err != nil {
			return 0, 0, err
		}
	}

	//pipeline override here comes from pipeline strategy table
	if strategy != nil && len(strategy.Config) > 0 {
		merged, err = impl.mergeUtil.JsonPatch(merged, []byte(strategy.Config))
		if err != nil {
			return 0, 0, err
		}
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, dbMigrationOverride)
	if err != nil {
		return 0, 0, err
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, []byte(overrideJson))
	if err != nil {
		return 0, 0, err
	}

	if configMapJson != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, configMapJson)
		if err != nil {
			return 0, 0, err
		}
	}

	appName := fmt.Sprintf("%s-%s", pipeline.App.AppName, envOverride.Environment.Name)
	merged = impl.hpaCheckBeforeTrigger(ctx, appName, envOverride.Namespace, merged, pipeline.AppId)

	chartRepoName := impl.GetChartRepoName(envOverride.Chart.GitRepoUrl)
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(overrideRequest.UserId)
	chartGitAttr := &ChartConfig{
		FileName:       fmt.Sprintf("_%d-values.yaml", envOverride.TargetEnvironment),
		FileContent:    string(merged),
		ChartName:      envOverride.Chart.ChartName,
		ChartLocation:  envOverride.Chart.ChartLocation,
		ChartRepoName:  chartRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", override.Id, envOverride.TargetEnvironment),
		UserName:       userName,
		UserEmailId:    userEmailId,
	}
	gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			return 0, 0, err
		}
	}
	commitHash, err := impl.gitFactory.Client.CommitValues(chartGitAttr, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return 0, 0, err
	}
	pipelineOverride := &chartConfig.PipelineOverride{
		Id:                     override.Id,
		GitHash:                commitHash,
		EnvConfigOverrideId:    envOverride.Id,
		PipelineOverrideValues: overrideJson,
		PipelineId:             overrideRequest.PipelineId,
		CiArtifactId:           overrideRequest.CiArtifactId,
		PipelineMergedValues:   string(merged),
		AuditLog:               sql.AuditLog{UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
	}
	err = impl.pipelineOverrideRepository.Update(pipelineOverride)
	if err != nil {
		return 0, 0, err
	}
	err = impl.CreateHistoriesForDeploymentTrigger(pipeline, strategy, envOverride, overrideJson, triggeredAt, pipelineOverride.UpdatedBy)
	if err != nil {
		impl.logger.Errorw("error in creating history entries for deployment trigger", "err", err)
		return 0, 0, err
	}
	return override.PipelineReleaseCounter, override.Id, nil
}

func (impl AppServiceImpl) savePipelineOverride(overrideRequest *bean.ValuesOverrideRequest, envOverrideId int, triggeredAt time.Time) (override *chartConfig.PipelineOverride, err error) {
	currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(overrideRequest.PipelineId)
	if err != nil {
		return nil, err
	}
	po := &chartConfig.PipelineOverride{
		EnvConfigOverrideId:    envOverrideId,
		Status:                 models.CHARTSTATUS_NEW,
		PipelineId:             overrideRequest.PipelineId,
		CiArtifactId:           overrideRequest.CiArtifactId,
		PipelineReleaseCounter: currentReleaseNo + 1,
		CdWorkflowId:           overrideRequest.CdWorkflowId,
		AuditLog:               sql.AuditLog{CreatedBy: overrideRequest.UserId, CreatedOn: triggeredAt, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
		DeploymentType:         overrideRequest.DeploymentType,
	}

	err = impl.pipelineOverrideRepository.Save(po)
	if err != nil {
		return nil, err
	}
	err = impl.checkAndFixDuplicateReleaseNo(po)
	if err != nil {
		impl.logger.Errorw("error in checking release no duplicacy", "pipeline", po, "err", err)
		return nil, err
	}
	return po, nil
}

func (impl AppServiceImpl) checkAndFixDuplicateReleaseNo(override *chartConfig.PipelineOverride) error {

	uniqueVerified := false
	retryCount := 0

	for !uniqueVerified && retryCount < 5 {
		retryCount = retryCount + 1
		overrides, err := impl.pipelineOverrideRepository.GetByPipelineIdAndReleaseNo(override.PipelineId, override.PipelineReleaseCounter)
		if err != nil {
			return err
		}
		if overrides[0].Id == override.Id {
			uniqueVerified = true
		} else {
			//duplicate might be due to concurrency, lets fix it
			currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(override.PipelineId)
			if err != nil {
				return err
			}
			override.PipelineReleaseCounter = currentReleaseNo + 1
			err = impl.pipelineOverrideRepository.Save(override)
			if err != nil {
				return err
			}
		}
	}
	if !uniqueVerified {
		return fmt.Errorf("duplicate verification retry count exide max overrideId: %d ,count: %d", override.Id, retryCount)
	}
	return nil
}

func (impl AppServiceImpl) updateArgoPipeline(appId int, pipelineName string, envOverride *chartConfig.EnvConfigOverride, ctx context.Context) (bool, error) {
	//repo has been registered while helm create
	if ctx == nil {
		impl.logger.Errorw("err in syncing ACD, ctx is NULL", "pipelineName", pipelineName)
		return false, nil
	}
	app, err := impl.appRepository.FindById(appId)
	if err != nil {
		impl.logger.Errorw("no app found ", "err", err)
		return false, err
	}
	envModel, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
	if err != nil {
		return false, err
	}
	argoAppName := fmt.Sprintf("%s-%s", app.AppName, envModel.Name)
	application, err := impl.acdClient.Get(ctx, &application2.ApplicationQuery{Name: &argoAppName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "app", argoAppName, "pipeline", pipelineName)
		return false, err
	}
	//if status, ok:=status.FromError(err);ok{
	appStatus, _ := status.FromError(err)

	if appStatus.Code() == codes.OK {
		impl.logger.Debugw("argo app exists", "app", argoAppName, "pipeline", pipelineName)

		if application.Spec.Source.Path != envOverride.Chart.ChartLocation {
			patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: v1alpha1.ApplicationSource{Path: envOverride.Chart.ChartLocation, RepoURL: envOverride.Chart.GitRepoUrl}}}
			reqbyte, err := json.Marshal(patchReq)
			if err != nil {
				impl.logger.Errorw("error in creating patch", "err", err)
			}
			_, err = impl.acdClient.Patch(ctx, &application2.ApplicationPatchRequest{Patch: string(reqbyte), Name: &argoAppName, PatchType: "merge"})
			if err != nil {
				impl.logger.Errorw("error in creating argo pipeline ", "name", pipelineName, "patch", string(reqbyte), "err", err)
				return false, err
			}
			impl.logger.Debugw("pipeline update req ", "res", patchReq)
		} else {
			impl.logger.Debug("pipeline no need to update ")
		}
		return true, nil
	} else if appStatus.Code() == codes.NotFound {
		impl.logger.Errorw("argo app not found", "app", argoAppName, "pipeline", pipelineName)
		return false, nil
	} else {
		impl.logger.Errorw("err in checking application on gocd", "err", err, "pipeline", pipelineName)
		return false, err
	}
}

func (impl *AppServiceImpl) UpdateCdWorkflowRunnerByACDObject(app v1alpha1.Application, cdWorkflowId int) error {
	cdWorkflow, err := impl.cdWorkflowRepository.FindById(cdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner, fetch failed for cdwf", "cdWorkflow", cdWorkflow, "app", app, "err", err)
		return err
	}
	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(cdWorkflow.Id, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner, fetch failed for runner type", "wfr", wfr, "app", app, "err", err)
		return err
	}
	wfr.Status = app.Status.Health.Status
	wfr.FinishedOn = time.Now()
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&wfr)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner", "wfr", wfr, "app", app, "err", err)
		return err
	}
	return nil
}

func (impl *AppServiceImpl) hpaCheckBeforeTrigger(ctx context.Context, appName string, namespace string, merged []byte, appId int) []byte {
	templateMap := make(map[string]interface{})
	err := json.Unmarshal(merged, &templateMap)
	if err != nil {
		return merged
	}
	if _, ok := templateMap[autoscaling.ServiceName]; ok {
		as := templateMap[autoscaling.ServiceName]
		asd := as.(map[string]interface{})
		isEnable := false
		if _, ok := asd["enabled"]; ok {
			isEnable = asd["enabled"].(bool)
		}
		if isEnable {
			reqReplicaCount := templateMap["replicaCount"].(float64)
			reqMaxReplicas := asd["MaxReplicas"].(float64)
			reqMinReplicas := asd["MinReplicas"].(float64)
			query := &application2.ApplicationResourceRequest{
				Name:         &appName,
				Version:      "",
				Group:        autoscaling.ServiceName,
				Kind:         "HorizontalPodAutoscaler",
				ResourceName: fmt.Sprintf("%s-%s", appName, "hpa"),
				Namespace:    namespace,
			}
			recv, err := impl.acdClient.GetResource(ctx, query)
			impl.logger.Debugw("resource manifest get replica count", "response", recv)
			if err != nil {
				impl.logger.Errorw("ACD Get Resource API Failed", "err", err)
				middleware.AcdGetResourceCounter.WithLabelValues(strconv.Itoa(appId), namespace, appName).Inc()
				return merged
			}
			if recv != nil && len(recv.Manifest) > 0 {
				resourceManifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(recv.Manifest), &resourceManifest)
				if err != nil {
					impl.logger.Errorw("unmarshal failed for hpa check", "err", err)
					return merged
				}
				status := resourceManifest["status"]
				statusMap := status.(map[string]interface{})
				currentReplicaCount := statusMap["currentReplicas"].(float64)

				if currentReplicaCount <= reqMaxReplicas && currentReplicaCount >= reqMinReplicas {
					reqReplicaCount = currentReplicaCount
				} else if currentReplicaCount > reqMaxReplicas {
					reqReplicaCount = reqMaxReplicas
				} else if currentReplicaCount < reqMinReplicas {
					reqReplicaCount = reqMinReplicas
				}
				templateMap["replicaCount"] = reqReplicaCount
				merged, err = json.Marshal(&templateMap)
				if err != nil {
					impl.logger.Errorw("marshaling failed for hpa check", "err", err)
					return merged
				}
			}
		}
	}

	return merged
}

func (impl *AppServiceImpl) CreateHistoriesForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, strategy *chartConfig.PipelineStrategy, envOverride *chartConfig.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) error {
	//creating history for deployment template
	err := impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryForDeploymentTrigger(pipeline, envOverride, renderedImageTemplate, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating deployment template history for deployment trigger", "err", err)
		return err
	}
	err = impl.configMapHistoryService.CreateCMCSHistoryForDeploymentTrigger(pipeline, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating CM/CS history for deployment trigger", "err", err)
		return err
	}
	err = impl.pipelineStrategyHistoryService.CreateStrategyHistoryForDeploymentTrigger(strategy, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating strategy history for deployment trigger", "err", err)
		return err
	}
	return nil
}
