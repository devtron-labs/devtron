package pipeline

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	repo "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
)

func TestChartServiceImpl_DeploymentTemplateValidate(t *testing.T) {

	type fields struct {
		chartRepository           chartConfig.ChartRepository
		logger                    *zap.SugaredLogger
		repoRepository            chartConfig.ChartRepoRepository
		chartTemplateService      util.ChartTemplateService
		pipelineGroupRepository   pipelineConfig.AppRepository
		mergeUtil                 util.MergeUtil
		repositoryService         repository.ServiceClient
		refChartDir               RefChartDir
		defaultChart              DefaultChart
		chartRefRepository        chartConfig.ChartRefRepository
		envOverrideRepository     chartConfig.EnvConfigOverrideRepository
		pipelineConfigRepository  chartConfig.PipelineConfigRepository
		configMapRepository       chartConfig.ConfigMapRepository
		environmentRepository     cluster.EnvironmentRepository
		pipelineRepository        pipelineConfig.PipelineRepository
		appLevelMetricsRepository repo.AppLevelMetricsRepository
		client                    *http.Client
	}
	type args struct {
		templatejson json.RawMessage
		chartRefId   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "CPU Memory test 1",
			fields: fields{},
			args: args{
				templatejson: json.RawMessage([]byte(`    "resources": {
															  "limits": {
																"cpu": "50m",
																"memory": "50Mi"
															  },
															  "requests": {
																"cpu": "10m",
																"memory": "10Mi"
															  }
															},`)),
				chartRefId:   11,
			},
			want:    false,
			wantErr: true,
		},
		{
			name:   "CPU Memory test 2",
			fields: fields{},
			args: args{
				templatejson: json.RawMessage([]byte(`    "resources": {
															  "limits": {
																"cpu": ".05",
																"memory": "50Mi"
															  },
															  "requests": {
																"cpu": "10m",
																"memory": "10Mi"
															  }
															},`)),
				chartRefId:   12,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ChartServiceImpl{
				chartRepository:           tt.fields.chartRepository,
				logger:                    tt.fields.logger,
				repoRepository:            tt.fields.repoRepository,
				chartTemplateService:      tt.fields.chartTemplateService,
				pipelineGroupRepository:   tt.fields.pipelineGroupRepository,
				mergeUtil:                 tt.fields.mergeUtil,
				repositoryService:         tt.fields.repositoryService,
				refChartDir:               tt.fields.refChartDir,
				defaultChart:              tt.fields.defaultChart,
				chartRefRepository:        tt.fields.chartRefRepository,
				envOverrideRepository:     tt.fields.envOverrideRepository,
				pipelineConfigRepository:  tt.fields.pipelineConfigRepository,
				configMapRepository:       tt.fields.configMapRepository,
				environmentRepository:     tt.fields.environmentRepository,
				pipelineRepository:        tt.fields.pipelineRepository,
				appLevelMetricsRepository: tt.fields.appLevelMetricsRepository,
				client:                    tt.fields.client,
			}
			got, err := impl.DeploymentTemplateValidate(tt.args.templatejson, tt.args.chartRefId)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentTemplateValidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentTemplateValidate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
