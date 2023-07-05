package pipeline

import (
	"fmt"
	mocks4 "github.com/devtron-labs/devtron/internal/sql/repository/app/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/chartConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chartRepo/repository/mocks"
	mocks3 "github.com/devtron-labs/devtron/pkg/commonService/mocks"
	mocks5 "github.com/devtron-labs/devtron/pkg/pipeline/history/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestConfigMapServiceImpl_ConfigSecretEnvironmentCreate(t *testing.T) {
	t.SkipNow()
	chartRepository := mocks.NewChartRepository(t)
	sugaredLogger, _ := util.NewSugardLogger()

	repoRepository := mocks.NewChartRepoRepository(t)
	pipelineConfigRepository := mocks2.NewPipelineConfigRepository(t)
	configMapRepository := mocks2.NewConfigMapRepository(t)
	environmentConfigRepository := mocks2.NewEnvConfigOverrideRepository(t)
	commonService := mocks3.NewCommonService(t)
	appRepository := mocks4.NewAppRepository(t)
	configMapHistoryService := mocks5.NewConfigMapHistoryService(t)
	configMap := &chartConfig.ConfigMapEnvModel{
		AppId:         22,
		EnvironmentId: 5,
	}
	model := &chartConfig.ConfigMapEnvModel{
		AppId:         22,
		EnvironmentId: 5,
		Deleted:       false,
	}
	responseModel := &chartConfig.ConfigMapEnvModel{
		AppId:         22,
		EnvironmentId: 5,
		Deleted:       true,
	}
	type args struct {
		createJobEnvOverrideRequest *CreateJobEnvOverridePayload
		getByAppError               error
		getByAppResponse            *chartConfig.ConfigMapEnvModel
	}
	tests := []struct {
		name    string
		args    args
		want    *chartConfig.ConfigMapEnvModel
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "create environment override",
			args: args{
				createJobEnvOverrideRequest: &CreateJobEnvOverridePayload{
					AppId: 22,
					EnvId: 5,
				},
				getByAppError:    pg.ErrNoRows,
				getByAppResponse: nil,
			},
			want:    configMap,
			wantErr: assert.NoError,
		},
		{
			name: "create deleted override",
			args: args{
				createJobEnvOverrideRequest: &CreateJobEnvOverridePayload{
					AppId: 22,
					EnvId: 5,
				},
				getByAppError:    nil,
				getByAppResponse: responseModel,
			},
			want:    configMap,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ConfigMapServiceImpl{
				chartRepository:             chartRepository,
				logger:                      sugaredLogger,
				repoRepository:              repoRepository,
				mergeUtil:                   util.MergeUtil{},
				pipelineConfigRepository:    pipelineConfigRepository,
				configMapRepository:         configMapRepository,
				environmentConfigRepository: environmentConfigRepository,
				commonService:               commonService,
				appRepository:               appRepository,
				configMapHistoryService:     configMapHistoryService,
			}
			configMapRepository.On("GetByAppIdAndEnvIdEnvLevel", 22, 5).Return(tt.args.getByAppResponse, tt.args.getByAppError).Once()
			if tt.args.getByAppError == pg.ErrNoRows {
				configMapRepository.On("CreateEnvLevel", model).Return(model, nil).Once()
			}
			if tt.args.getByAppError == nil {
				configMapRepository.On("UpdateEnvLevel", model).Return(model, nil).Once()
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
	t.SkipNow()
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
	configMap := &chartConfig.ConfigMapEnvModel{
		AppId:         1,
		EnvironmentId: 1,
	}
	tests := []struct {
		name    string
		args    args
		want    *chartConfig.ConfigMapEnvModel
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Delete configMap",
			args: args{
				createJobEnvOverrideRequest: &CreateJobEnvOverridePayload{
					AppId: 1,
					EnvId: 1,
				},
			},
			want:    configMap,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ConfigMapServiceImpl{
				chartRepository:             chartRepository,
				logger:                      sugaredLogger,
				repoRepository:              repoRepository,
				mergeUtil:                   util.MergeUtil{},
				pipelineConfigRepository:    pipelineConfigRepository,
				configMapRepository:         configMapRepository,
				environmentConfigRepository: environmentConfigRepository,
				commonService:               commonService,
				appRepository:               appRepository,
				configMapHistoryService:     configMapHistoryService,
			}
			configMapRepository.On("GetByAppIdAndEnvIdEnvLevel", 1, 1).Return(configMap, nil)
			configMapRepository.On("UpdateEnvLevel", mock.Anything).Return(nil, nil)
			got, err := impl.ConfigSecretEnvironmentDelete(tt.args.createJobEnvOverrideRequest)
			if !tt.wantErr(t, err, fmt.Sprintf("ConfigSecretEnvironmentDelete(%v)", tt.args.createJobEnvOverrideRequest)) {
				return
			}
			assert.Equal(t, tt.want.AppId, got.AppId, "ConfigSecretEnvironmentCreate(%v)", tt.args.createJobEnvOverrideRequest)
			assert.Equal(t, tt.want.EnvironmentId, got.EnvironmentId, "ConfigSecretEnvironmentCreate(%v)", tt.args.createJobEnvOverrideRequest)
			assert.Equal(t, tt.want.Deleted, true, "ConfigSecretEnvironmentCreate(%v)", tt.args.createJobEnvOverrideRequest)
		})
	}
}

func TestConfigMapServiceImpl_ConfigSecretEnvironmentGet(t *testing.T) {
	//t.SkipNow()
	chartRepository := mocks.NewChartRepository(t)
	sugaredLogger, _ := util.NewSugardLogger()
	repoRepository := mocks.NewChartRepoRepository(t)
	pipelineConfigRepository := mocks2.NewPipelineConfigRepository(t)
	configMapRepository := mocks2.NewConfigMapRepository(t)
	environmentConfigRepository := mocks2.NewEnvConfigOverrideRepository(t)
	commonService := mocks3.NewCommonService(t)
	appRepository := mocks4.NewAppRepository(t)
	configMapHistoryService := mocks5.NewConfigMapHistoryService(t)
	configMap := []*chartConfig.ConfigMapEnvModel{
		{
			AppId:         1,
			EnvironmentId: 1,
		},
	}
	type args struct {
		appId int
	}
	tests := []struct {
		name string

		args    args
		want    []*chartConfig.ConfigMapEnvModel
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test1",
			args: args{
				appId: 1,
			},
			want:    configMap,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ConfigMapServiceImpl{
				chartRepository:             chartRepository,
				logger:                      sugaredLogger,
				repoRepository:              repoRepository,
				mergeUtil:                   util.MergeUtil{},
				pipelineConfigRepository:    pipelineConfigRepository,
				configMapRepository:         configMapRepository,
				environmentConfigRepository: environmentConfigRepository,
				commonService:               commonService,
				appRepository:               appRepository,
				configMapHistoryService:     configMapHistoryService,
			}
			configMapRepository.On("GetEnvLevelByAppId", 1).Return(configMap, nil)
			got, err := impl.ConfigSecretEnvironmentGet(tt.args.appId)
			if !tt.wantErr(t, err, fmt.Sprintf("ConfigSecretEnvironmentGet(%v)", tt.args.appId)) {
				return
			}
			assert.Equal(t, tt.want[0].EnvironmentId, got[0].EnvironmentId, "ConfigSecretEnvironmentGet(%v)", tt.args.appId)
		})
	}
}
