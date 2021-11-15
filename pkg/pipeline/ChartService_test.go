package pipeline

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"testing"
)

func TestChartServiceImpl_DefaultTemplateWithSavedTemplateData(t *testing.T) {
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
		appLevelMetricsRepository repository2.AppLevelMetricsRepository
		client                    *http.Client
	}
	type args struct {
		RequestChartRefId int
		templateRequest   *TemplateRequest
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
			got, err := impl.DefaultTemplateWithSavedTemplateData(tt.args.RequestChartRefId, tt.args.templateRequest)
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
