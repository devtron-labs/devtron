package restHandler

import (
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/pkg/appClone"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	security2 "github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"reflect"
	"testing"
	"encoding/json"

)

func TestPipelineConfigRestHandlerImpl_GetDeploymentTemplate(t *testing.T) {
	type fields struct {
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
		envService              cluster.EnvironmentService
		gitRegistryConfig       pipeline.GitRegistryConfig
		dockerRegistryConfig    pipeline.DockerRegistryConfig
		cdHandelr               pipeline.CdHandler
		appCloneService         appClone.AppCloneService
		materialRepository      pipelineConfig.MaterialRepository
		policyService           security2.PolicyService
		scanResultRepository    security.ImageScanResultRepository
		gitProviderRepo         repository.GitProviderRepository
	}
	type args struct {
	w                 http.ResponseWriter
	r                 *http.Request
	RequestChartRefId int
	templateRequest   *pipeline.TemplateRequest
}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := PipelineConfigRestHandlerImpl{
				pipelineBuilder:         tt.fields.pipelineBuilder,
				ciPipelineRepository:    tt.fields.ciPipelineRepository,
				ciHandler:               tt.fields.ciHandler,
				Logger:                  tt.fields.Logger,
				chartService:            tt.fields.chartService,
				propertiesConfigService: tt.fields.propertiesConfigService,
				dbMigrationService:      tt.fields.dbMigrationService,
				application:             tt.fields.application,
				userAuthService:         tt.fields.userAuthService,
				validator:               tt.fields.validator,
				teamService:             tt.fields.teamService,
				enforcer:                tt.fields.enforcer,
				gitSensorClient:         tt.fields.gitSensorClient,
				pipelineRepository:      tt.fields.pipelineRepository,
				appWorkflowService:      tt.fields.appWorkflowService,
				enforcerUtil:            tt.fields.enforcerUtil,
				envService:              tt.fields.envService,
				gitRegistryConfig:       tt.fields.gitRegistryConfig,
				dockerRegistryConfig:    tt.fields.dockerRegistryConfig,
				cdHandelr:               tt.fields.cdHandelr,
				appCloneService:         tt.fields.appCloneService,
				materialRepository:      tt.fields.materialRepository,
				policyService:           tt.fields.policyService,
				scanResultRepository:    tt.fields.scanResultRepository,
				gitProviderRepo:         tt.fields.gitProviderRepo,
			}
			got, err := handler.chartService.DefaultTemplateWithSavedTemplateData(tt.args.RequestChartRefId, tt.args.templateRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultTemplateWithSavedTemplateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultTemplateWithSavedTemplateData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
