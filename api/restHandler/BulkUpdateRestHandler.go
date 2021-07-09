package restHandler

import (
	"encoding/json"
	"fmt"
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
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type BulkUpdateRestHandler interface {
	FindBulkUpdateReadme(w http.ResponseWriter, r *http.Request)
	GetAppNameDeploymentTemplate(w http.ResponseWriter, r *http.Request)
	BulkUpdateDeploymentTemplate(w http.ResponseWriter, r *http.Request)
}
type BulkUpdateRestHandlerImpl struct {
	pipelineBuilder         pipeline.PipelineBuilder
	ciPipelineRepository    pipelineConfig.CiPipelineRepository
	ciHandler               pipeline.CiHandler
	logger                  *zap.SugaredLogger
	bulkUpdateService       pipeline.BulkUpdateService
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

func NewBulkUpdateRestHandlerImpl(pipelineBuilder pipeline.PipelineBuilder, logger *zap.SugaredLogger,
	bulkUpdateService pipeline.BulkUpdateService,
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
		logger:                  logger,
		bulkUpdateService:       bulkUpdateService,
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

func (handler BulkUpdateRestHandlerImpl) FindBulkUpdateReadme(w http.ResponseWriter, r *http.Request) {
	var operation string
	vars := mux.Vars(r)
	apiVersion := vars["apiVersion"]
	kind := vars["kind"]
	operation = fmt.Sprintf("%s/%s", apiVersion, kind)
	response, err := handler.bulkUpdateService.FindBulkUpdateReadme(operation)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//auth free, only login required
	var responseArr []pipeline.BulkUpdateSeeExampleResponse
	responseArr = append(responseArr, response)
	writeJsonResp(w, nil, responseArr, http.StatusOK)
}
func (handler BulkUpdateRestHandlerImpl) GetAppNameDeploymentTemplate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var script pipeline.BulkUpdateScript
	err := decoder.Decode(&script)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(script)
	if err != nil {
		handler.logger.Errorw("validation err, Script", "err", err, "BulkUpdateScript", script)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	impactedApps, err := handler.bulkUpdateService.GetBulkAppName(script.Spec)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	appResourceObjects, envResourceObjects := handler.enforcerUtil.GetRbacObjectsForAllAppsAndEnvironments()
	for _, impactedApp := range impactedApps {
		resourceName := appResourceObjects[impactedApp.AppId]
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
			writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		if impactedApp.EnvId > 0 {
			key := fmt.Sprintf("%d-%d", impactedApp.EnvId, impactedApp.AppId)
			envResourceName := envResourceObjects[key]
			if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, envResourceName); !ok {
				writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
				return
			}
		}
	}

	writeJsonResp(w, err, impactedApps, http.StatusOK)
}

func (handler BulkUpdateRestHandlerImpl) BulkUpdateDeploymentTemplate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var script pipeline.BulkUpdateScript
	err := decoder.Decode(&script)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(script)
	if err != nil {
		handler.logger.Errorw("validation err, Script", "err", err, "BulkUpdateScript", script)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	impactedApps, err := handler.bulkUpdateService.GetBulkAppName(script.Spec)
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	rbacObjects := handler.enforcerUtil.GetRbacObjectsForAllApps()
	for _, impactedApp := range impactedApps {
		resourceName := rbacObjects[impactedApp.AppId]
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, resourceName); !ok {
			writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		if impactedApp.EnvId > 0 {
			resourceName := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(impactedApp.AppName, impactedApp.EnvId)
			if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionUpdate, resourceName); !ok {
				writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
				return
			}
		}
	}

	response := handler.bulkUpdateService.BulkUpdateDeploymentTemplate(script.Spec)
	writeJsonResp(w, nil, response, http.StatusOK)
}
