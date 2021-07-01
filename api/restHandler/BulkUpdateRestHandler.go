package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/pkg/appClone"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	request "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	security2 "github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type BulkUpdateRestHandler interface {
	GetExampleInputBulkUpdate(w http.ResponseWriter, r *http.Request)
	GetAppNameDeploymentTemplate(w http.ResponseWriter, r *http.Request)
	BulkUpdateDeploymentTemplate(w http.ResponseWriter, r *http.Request)
}
type BulkUpdateRestHandlerImpl struct {
	pipelineBuilder         pipeline.PipelineBuilder
	ciPipelineRepository    pipelineConfig.CiPipelineRepository
	ciHandler               pipeline.CiHandler
	Logger                  *zap.SugaredLogger
	chartService            pipeline.ChartService
	propertiesConfigService pipeline.PropertiesConfigService
	dbMigrationService      pipeline.DbMigrationService
	application             application.ServiceClient
	userAuthService         user.UserService
	validator               *validator.Validate
	teamService             team.TeamService
	enforcer                rbac.Enforcer
	gitSensorClient         gitSensor.GitSensorClient
	pipelineRepository      pipelineConfig.PipelineRepository
	appWorkflowService      appWorkflow.AppWorkflowService
	enforcerUtil            rbac.EnforcerUtil
	envService              request.EnvironmentService
	gitRegistryConfig       pipeline.GitRegistryConfig
	dockerRegistryConfig    pipeline.DockerRegistryConfig
	cdHandelr               pipeline.CdHandler
	appCloneService         appClone.AppCloneService
	materialRepository      pipelineConfig.MaterialRepository
	policyService           security2.PolicyService
	scanResultRepository    security.ImageScanResultRepository
}

func NewBulkUpdateRestHandlerImpl(pipelineBuilder pipeline.PipelineBuilder, Logger *zap.SugaredLogger,
	chartService pipeline.ChartService,
	propertiesConfigService pipeline.PropertiesConfigService,
	dbMigrationService pipeline.DbMigrationService,
	application application.ServiceClient,
	userAuthService user.UserService,
	teamService team.TeamService,
	enforcer rbac.Enforcer,
	ciHandler pipeline.CiHandler,
	validator *validator.Validate,
	gitSensorClient gitSensor.GitSensorClient,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	enforcerUtil rbac.EnforcerUtil, envService request.EnvironmentService,
	gitRegistryConfig pipeline.GitRegistryConfig, dockerRegistryConfig pipeline.DockerRegistryConfig,
	cdHandelr pipeline.CdHandler,
	appCloneService appClone.AppCloneService,
	appWorkflowService appWorkflow.AppWorkflowService,
	materialRepository pipelineConfig.MaterialRepository, policyService security2.PolicyService,
	scanResultRepository security.ImageScanResultRepository) *BulkUpdateRestHandlerImpl {
	return &BulkUpdateRestHandlerImpl{
		pipelineBuilder:         pipelineBuilder,
		Logger:                  Logger,
		chartService:            chartService,
		propertiesConfigService: propertiesConfigService,
		dbMigrationService:      dbMigrationService,
		application:             application,
		userAuthService:         userAuthService,
		validator:               validator,
		teamService:             teamService,
		enforcer:                enforcer,
		ciHandler:               ciHandler,
		gitSensorClient:         gitSensorClient,
		ciPipelineRepository:    ciPipelineRepository,
		pipelineRepository:      pipelineRepository,
		enforcerUtil:            enforcerUtil,
		envService:              envService,
		gitRegistryConfig:       gitRegistryConfig,
		dockerRegistryConfig:    dockerRegistryConfig,
		cdHandelr:               cdHandelr,
		appCloneService:         appCloneService,
		appWorkflowService:      appWorkflowService,
		materialRepository:      materialRepository,
		policyService:           policyService,
		scanResultRepository:    scanResultRepository,
	}
}

func (handler BulkUpdateRestHandlerImpl) GetExampleInputBulkUpdate(w http.ResponseWriter, r *http.Request) {
	response, err := handler.chartService.GetBulkUpdateInput("deployment-template")
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	var responseArr []pipeline.BulkUpdateGet
	responseArr = append(responseArr, response)
	writeJsonResp(w, nil, responseArr, http.StatusOK)
}
func (handler BulkUpdateRestHandlerImpl) GetAppNameDeploymentTemplate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var bulkUpdatePayload pipeline.BulkUpdatePayload
	err := decoder.Decode(&bulkUpdatePayload)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	AppName, err := handler.chartService.GetBulkAppName(bulkUpdatePayload.Payload)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, AppName, http.StatusOK)
}
func (handler BulkUpdateRestHandlerImpl) BulkUpdateDeploymentTemplate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var bulkUpdatePayload pipeline.BulkUpdatePayload
	err := decoder.Decode(&bulkUpdatePayload)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	response, err := handler.chartService.BulkUpdateDeploymentTemplate(bulkUpdatePayload.Payload)
	if err != nil {
		writeJsonResp(w, err, response, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, response, http.StatusOK)
}
