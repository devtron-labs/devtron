package pipeline

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	mocks4 "github.com/devtron-labs/devtron/internal/sql/repository/app/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/chartConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/chartRepo/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/commonService"
	mocks3 "github.com/devtron-labs/devtron/pkg/commonService/mocks"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	mocks5 "github.com/devtron-labs/devtron/pkg/pipeline/history/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestConfigMapServiceImpl_ConfigSecretEnvironmentCreate(t *testing.T) {
	chartRepository := mocks.NewChartRepository(t)
	sugaredLogger, _ := util.NewSugardLogger()

	repoRepository := mocks.NewChartRepoRepository(t)
	pipelineConfigRepository := mocks2.NewPipelineConfigRepository(t)
	configMapRepository := mocks2.NewConfigMapRepository(t)
	environmentConfigRepository := mocks2.NewEnvConfigOverrideRepository(t)
	commonService := mocks3.NewCommonService(t)
	appRepository := mocks4.NewAppRepository(t)
	configMapHistoryService := mocks5.NewConfigMapHistoryService(t)
	type args struct {
		createJobEnvOverrideRequest *CreateJobEnvOverridePayload
	}
	tests := []struct {
		name string
		//fields  fields
		args    args
		want    *chartConfig.ConfigMapEnvModel
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{
			name: "create environment override",
			args: args{createJobEnvOverrideRequest: &CreateJobEnvOverridePayload{
				AppId:  20,
				EnvId:  4,
				UserId: 1,
			}},
			want: &chartConfig.ConfigMapEnvModel{
				AppId:         20,
				EnvironmentId: 4,
				Deleted:       false,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ConfigMapServiceImpl{
				chartRepository: chartRepository,
				logger:          sugaredLogger,
				repoRepository:  repoRepository,
				//mergeUtil:                   tt.fields.mergeUtil,
				pipelineConfigRepository:    pipelineConfigRepository,
				configMapRepository:         configMapRepository,
				environmentConfigRepository: environmentConfigRepository,
				commonService:               commonService,
				appRepository:               appRepository,
				configMapHistoryService:     configMapHistoryService,
			}
			got, err := impl.ConfigSecretEnvironmentCreate(tt.args.createJobEnvOverrideRequest)
			if !tt.wantErr(t, err, fmt.Sprintf("ConfigSecretEnvironmentCreate(%v)", tt.args.createJobEnvOverrideRequest)) {
				return
			}
			assert.Equal(t, tt.want.AppId, got.AppId, "ConfigSecretEnvironmentCreate(%v)", tt.args.createJobEnvOverrideRequest)
			assert.Equal(t, tt.want.EnvironmentId, got.EnvironmentId, "ConfigSecretEnvironmentCreate(%v)", tt.args.createJobEnvOverrideRequest)
			assert.Equal(t, tt.want.Deleted, false, "ConfigSecretEnvironmentCreate(%v)", tt.args.createJobEnvOverrideRequest)
		})
	}
}

func TestConfigMapServiceImpl_ConfigSecretEnvironmentDelete(t *testing.T) {
	type fields struct {
		chartRepository             chartRepoRepository.ChartRepository
		logger                      *zap.SugaredLogger
		repoRepository              chartRepoRepository.ChartRepoRepository
		mergeUtil                   util.MergeUtil
		pipelineConfigRepository    chartConfig.PipelineConfigRepository
		configMapRepository         chartConfig.ConfigMapRepository
		environmentConfigRepository chartConfig.EnvConfigOverrideRepository
		commonService               commonService.CommonService
		appRepository               app.AppRepository
		configMapHistoryService     history.ConfigMapHistoryService
	}
	type args struct {
		createJobEnvOverrideRequest *CreateJobEnvOverridePayload
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *chartConfig.ConfigMapEnvModel
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ConfigMapServiceImpl{
				chartRepository:             tt.fields.chartRepository,
				logger:                      tt.fields.logger,
				repoRepository:              tt.fields.repoRepository,
				mergeUtil:                   tt.fields.mergeUtil,
				pipelineConfigRepository:    tt.fields.pipelineConfigRepository,
				configMapRepository:         tt.fields.configMapRepository,
				environmentConfigRepository: tt.fields.environmentConfigRepository,
				commonService:               tt.fields.commonService,
				appRepository:               tt.fields.appRepository,
				configMapHistoryService:     tt.fields.configMapHistoryService,
			}
			got, err := impl.ConfigSecretEnvironmentDelete(tt.args.createJobEnvOverrideRequest)
			if !tt.wantErr(t, err, fmt.Sprintf("ConfigSecretEnvironmentDelete(%v)", tt.args.createJobEnvOverrideRequest)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ConfigSecretEnvironmentDelete(%v)", tt.args.createJobEnvOverrideRequest)
		})
	}
}

func TestConfigMapServiceImpl_ConfigSecretEnvironmentGet(t *testing.T) {
	type fields struct {
		chartRepository             chartRepoRepository.ChartRepository
		logger                      *zap.SugaredLogger
		repoRepository              chartRepoRepository.ChartRepoRepository
		mergeUtil                   util.MergeUtil
		pipelineConfigRepository    chartConfig.PipelineConfigRepository
		configMapRepository         chartConfig.ConfigMapRepository
		environmentConfigRepository chartConfig.EnvConfigOverrideRepository
		commonService               commonService.CommonService
		appRepository               app.AppRepository
		configMapHistoryService     history.ConfigMapHistoryService
	}
	type args struct {
		appId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*chartConfig.ConfigMapEnvModel
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ConfigMapServiceImpl{
				chartRepository:             tt.fields.chartRepository,
				logger:                      tt.fields.logger,
				repoRepository:              tt.fields.repoRepository,
				mergeUtil:                   tt.fields.mergeUtil,
				pipelineConfigRepository:    tt.fields.pipelineConfigRepository,
				configMapRepository:         tt.fields.configMapRepository,
				environmentConfigRepository: tt.fields.environmentConfigRepository,
				commonService:               tt.fields.commonService,
				appRepository:               tt.fields.appRepository,
				configMapHistoryService:     tt.fields.configMapHistoryService,
			}
			got, err := impl.ConfigSecretEnvironmentGet(tt.args.appId)
			if !tt.wantErr(t, err, fmt.Sprintf("ConfigSecretEnvironmentGet(%v)", tt.args.appId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ConfigSecretEnvironmentGet(%v)", tt.args.appId)
		})
	}
}

func TestNewConfigMapServiceImpl(t *testing.T) {
	type args struct {
		chartRepository             chartRepoRepository.ChartRepository
		logger                      *zap.SugaredLogger
		repoRepository              chartRepoRepository.ChartRepoRepository
		mergeUtil                   util.MergeUtil
		pipelineConfigRepository    chartConfig.PipelineConfigRepository
		configMapRepository         chartConfig.ConfigMapRepository
		environmentConfigRepository chartConfig.EnvConfigOverrideRepository
		commonService               commonService.CommonService
		appRepository               app.AppRepository
		configMapHistoryService     history.ConfigMapHistoryService
	}
	tests := []struct {
		name string
		args args
		want *ConfigMapServiceImpl
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewConfigMapServiceImpl(tt.args.chartRepository, tt.args.logger, tt.args.repoRepository, tt.args.mergeUtil, tt.args.pipelineConfigRepository, tt.args.configMapRepository, tt.args.environmentConfigRepository, tt.args.commonService, tt.args.appRepository, tt.args.configMapHistoryService), "NewConfigMapServiceImpl(%v, %v, %v, %v, %v, %v, %v, %v, %v, %v)", tt.args.chartRepository, tt.args.logger, tt.args.repoRepository, tt.args.mergeUtil, tt.args.pipelineConfigRepository, tt.args.configMapRepository, tt.args.environmentConfigRepository, tt.args.commonService, tt.args.appRepository, tt.args.configMapHistoryService)
		})
	}
}
