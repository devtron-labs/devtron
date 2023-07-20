package pipeline

import (
	"fmt"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	app2 "github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

var (
	ciCdPipelineOrchestrator *CiCdPipelineOrchestratorImpl
)

func TestCiCdPipelineOrchestratorImpl_CreateCiConf(t *testing.T) {
	t.SkipNow()
	InitClusterNoteService()
	type args struct {
		createRequest *bean.CiConfigRequest
		templateId    int
	}
	tests := []struct {
		name    string
		args    args
		want    *bean.CiConfigRequest
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "CreateCiConf success",
			args: args{
				createRequest: &bean.CiConfigRequest{},
				templateId:    123,
			},
			want:    &bean.CiConfigRequest{},
			wantErr: assert.NoError,
		},
		{
			name: "CreateCiConf success with payload",
			args: args{
				createRequest: &bean.CiConfigRequest{
					Id:    12,
					AppId: 20,
				},
				templateId: 123,
			},
			want: &bean.CiConfigRequest{
				Id:    12,
				AppId: 20,
			},
			wantErr: assert.NoError,
		},
		{
			name: "CreateCiConf success with job payload",
			args: args{
				createRequest: &bean.CiConfigRequest{
					Id:    0,
					AppId: 21,
					IsJob: true,
					CiPipelines: []*bean.CiPipeline{{
						IsExternal: true,
						IsManual:   false,
						AppId:      21,
						CiMaterial: []*bean.CiMaterial{
							{
								Source: &bean.SourceTypeConfig{
									Type:  "SOURCE_TYPE_BRANCH_FIXED",
									Value: "main",
									Regex: "",
								},
								GitMaterialId:   13,
								Id:              38,
								GitMaterialName: "devtron-test",
								IsRegex:         false,
							},
						},
						EnvironmentId: 3,
					}},
					UserId: 1,
				},
				templateId: 13,
			},
			want: &bean.CiConfigRequest{
				Id:          0,
				AppId:       21,
				CiPipelines: []*bean.CiPipeline{{EnvironmentId: 3}},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ciCdPipelineOrchestrator.CreateCiConf(tt.args.createRequest, tt.args.templateId)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateCiConf(%v, %v)", tt.args.createRequest, tt.args.templateId)) {
				return
			}
			assert.Equal(t, tt.want.AppId, got.AppId, "CreateCiConf(%v, %v)", tt.args.createRequest, tt.args.templateId)
			assert.Equal(t, tt.want.Id, got.Id, "CreateCiConf(%v, %v)", tt.args.createRequest, tt.args.templateId)
			assert.Equal(t, tt.want.CiPipelines[0].EnvironmentId, got.CiPipelines[0].EnvironmentId, "CreateCiConf(%v, %v)", tt.args.createRequest, tt.args.templateId)
		})
	}
}

func InitClusterNoteService() {
	if ciCdPipelineOrchestrator != nil {
		return
	}
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
	}
	conn, err := getDbConn()
	if err != nil {
		log.Fatalf("error in db connection initialization %s, %s", "err", err)
	}

	appRepository := app.NewAppRepositoryImpl(conn, logger)
	materialRepository := pipelineConfig.NewMaterialRepositoryImpl(conn)
	pipelineRepository := pipelineConfig.NewPipelineRepositoryImpl(conn, logger)
	ciPipelineRepository := pipelineConfig.NewCiPipelineRepositoryImpl(conn, logger)
	ciPipelineHistoryRepository := repository4.NewCiPipelineHistoryRepositoryImpl(conn, logger)
	ciPipelineMaterialRepository := pipelineConfig.NewCiPipelineMaterialRepositoryImpl(conn, logger)
	GitSensorClient, err := gitSensor.NewGitSensorClient(logger, &gitSensor.ClientConfig{})
	ciConfig := &CiConfig{}
	appWorkflowRepository := appWorkflow.NewAppWorkflowRepositoryImpl(logger, conn)
	envRepository := repository3.NewEnvironmentRepositoryImpl(conn, logger, nil)
	attributesService := attributes.NewAttributesServiceImpl(logger, nil)
	appListingRepositoryQueryBuilder := helper.NewAppListingRepositoryQueryBuilder(logger)
	appListingRepository := repository.NewAppListingRepositoryImpl(logger, conn, appListingRepositoryQueryBuilder, envRepository)
	appLabelsService := app2.NewAppCrudOperationServiceImpl(nil, logger, nil, nil, nil)
	userAuthService := user.NewUserAuthServiceImpl(nil, nil, nil, nil, nil, nil, nil)
	prePostCdScriptHistoryService := history.NewPrePostCdScriptHistoryServiceImpl(logger, nil, nil, nil)
	prePostCiScriptHistoryService := history.NewPrePostCiScriptHistoryServiceImpl(logger, nil)
	pipelineStageService := NewPipelineStageService(logger, nil, nil, nil)
	ciTemplateOverrideRepository := pipelineConfig.NewCiTemplateOverrideRepositoryImpl(conn, logger)
	ciTemplateService := *NewCiTemplateServiceImpl(logger, nil, nil, nil)
	gitMaterialHistoryService := history.NewGitMaterialHistoryServiceImpl(nil, logger)
	ciPipelineHistoryService := history.NewCiPipelineHistoryServiceImpl(ciPipelineHistoryRepository, logger, ciPipelineRepository)
	dockerArtifactStoreRepository := repository2.NewDockerArtifactStoreRepositoryImpl(conn)
	configMapRepository := chartConfig.NewConfigMapRepositoryImpl(logger, conn)
	configMapService := NewConfigMapServiceImpl(nil, nil, nil, util.MergeUtil{}, nil, configMapRepository, nil, nil, appRepository, nil, envRepository)
	ciCdPipelineOrchestrator = NewCiCdPipelineOrchestrator(appRepository, logger, materialRepository, pipelineRepository, ciPipelineRepository, ciPipelineMaterialRepository, GitSensorClient, ciConfig, appWorkflowRepository, envRepository, attributesService, appListingRepository, appLabelsService, userAuthService, prePostCdScriptHistoryService, prePostCiScriptHistoryService, pipelineStageService, ciTemplateOverrideRepository, gitMaterialHistoryService, ciPipelineHistoryService, ciTemplateService, dockerArtifactStoreRepository, configMapService)
}

func TestPatchCiMaterialSourceWhenOldPipelineIsNotFoundItShouldReturnError(t *testing.T) {
	//ctrl := gomock.NewController(t)
	pipeline := &bean.CiPipeline{Id: 1}
	userId := int32(10)

	mockedCiPipelineRepository := mocks.NewCiPipelineRepository(t)
	mockedCiPipelineRepository.On("FindById", pipeline.Id).Return(nil, fmt.Errorf("dsfsdf"))
	//mockedCiPipelineMaterialRepository := &mocks.MockCiPipelineMaterialRepository{}
	//mockedGitSensor := &mock_gitSensor.MockClient{}
	impl := CiCdPipelineOrchestratorImpl{
		ciPipelineRepository: mockedCiPipelineRepository,
		//ciPipelineMaterialRepository: mockedCiPipelineMaterialRepository,
		//GitSensorClient:              mockedGitSensor,
	}
	res, err := impl.PatchCiMaterialSource(pipeline, userId)
	assert.Error(t, err)
	assert.Nil(t, res)
}

//	func TestPatchCiMaterialSourceWhenOldPipelineExistsAndSaveUpdatedMaterialFailsItShouldReturnError(t *testing.T) {
//		//ctrl := gomock.NewController(t)
//		userId := int32(10)
//		oldPipeline := &bean.CiPipeline{
//			ParentAppId: 0,
//			AppId:       4,
//			CiMaterial: []*bean.CiMaterial{
//				{
//					Source: &bean.SourceTypeConfig{
//						Type:  "SOURCE_TYPE_BRANCH_FIXED",
//						Value: "main",
//					},
//					Id:      0,
//					IsRegex: false,
//				},
//			},
//			Id:     1,
//			Active: false,
//		}
//
//		newPipeline := &bean.CiPipeline{
//			ParentAppId: 0,
//			AppId:       4,
//			CiMaterial: []*bean.CiMaterial{
//				{
//					Source: &bean.SourceTypeConfig{
//						Type:  "SOURCE_TYPE_BRANCH_FIXED",
//						Value: "main",
//					},
//					Id:      1,
//					IsRegex: false,
//				},
//			},
//			Id:     0,
//			Active: false,
//		}
//		mockedCiPipelineRepository := mocks.NewCiPipelineRepository(t)
//		mockedCiPipelineRepository.On("FindById", newPipeline.Id).Return(oldPipeline, nil)
//		//mockedCiPipelineMaterialRepository := &mocks.MockCiPipelineMaterialRepository{}
//		//mockedGitSensor := &mock_gitSensor.MockClient{}
//		impl := CiCdPipelineOrchestratorImpl{
//			ciPipelineRepository: mockedCiPipelineRepository,
//			//ciPipelineMaterialRepository: mockedCiPipelineMaterialRepository,
//			//GitSensorClient:              mockedGitSensor,
//		}
//		res, err := impl.PatchCiMaterialSource(pipeline, userId)
//		assert.Error(t, err)
//		assert.Nil(t, res)
//	}
func Test_mapCiMaterialToPipelineMaterial(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now()
	type args struct {
		ciPipeline  *bean.CiPipeline
		userId      int32
		oldPipeline *pipelineConfig.CiPipeline
	}
	tests := []struct {
		name string
		args args
		want []*pipelineConfig.CiPipelineMaterial
	}{
		{
			name: "It should return []*pipelineConfig.CiPipelineMaterial with only source changed",
			args: args{
				ciPipeline: &bean.CiPipeline{
					AppId: 0,
					CiMaterial: []*bean.CiMaterial{
						{
							Source: &bean.SourceTypeConfig{
								Type:  "QWERTY",
								Value: "masterrrrr",
								Regex: "A@$%DS",
							},
							Id:            2,
							GitMaterialId: 2,
						},
					},
					Id: 2,
				},
				userId: 1,
				oldPipeline: &pipelineConfig.CiPipeline{
					Id:                       1,
					AppId:                    2,
					App:                      nil,
					CiTemplateId:             3,
					DockerArgs:               "abc",
					Name:                     "def",
					Version:                  "234",
					Active:                   true,
					Deleted:                  false,
					IsManual:                 true,
					IsExternal:               false,
					ParentCiPipeline:         0,
					ScanEnabled:              false,
					IsDockerConfigOverridden: false,
					AuditLog: sql.AuditLog{
						CreatedOn: t1,
						CreatedBy: 2,
						UpdatedOn: t2,
						UpdatedBy: 1,
					},
					CiPipelineMaterials: []*pipelineConfig.CiPipelineMaterial{
						{
							Id:            2,
							GitMaterialId: 2,
							CiPipelineId:  2,
							Path:          "",
							CheckoutPath:  "",
							Type:          "ABC",
							Value:         "DEF",
							ScmId:         "",
							ScmName:       "",
							ScmVersion:    "",
							Active:        false,
							Regex:         "A$%",
							GitTag:        "",
							CiPipeline:    nil,
							GitMaterial:   nil,
							AuditLog: sql.AuditLog{
								CreatedOn: t1,
								CreatedBy: 2,
								UpdatedOn: t1,
								UpdatedBy: 2,
							},
						},
						{
							Id:            3,
							GitMaterialId: 2,
							CiPipelineId:  2,
							Path:          "",
							CheckoutPath:  "",
							Type:          "ABC123",
							Value:         "main",
							ScmId:         "",
							ScmName:       "",
							ScmVersion:    "",
							Active:        false,
							Regex:         "A$%a3e2",
							GitTag:        "sdf",
							CiPipeline:    nil,
							GitMaterial:   nil,
							AuditLog: sql.AuditLog{
								CreatedOn: t1,
								CreatedBy: 2,
								UpdatedOn: t1,
								UpdatedBy: 2,
							},
						},
					},
					CiTemplate: nil,
				},
			},
			want: []*pipelineConfig.CiPipelineMaterial{
				{
					Id:            2,
					GitMaterialId: 2,
					CiPipelineId:  2,
					Path:          "",
					CheckoutPath:  "",
					Type:          "QWERTY",
					Value:         "masterrrrr",
					Regex:         "A@$%DS",
					ScmId:         "",
					ScmName:       "",
					ScmVersion:    "",
					Active:        false,
					GitTag:        "",
					CiPipeline:    nil,
					GitMaterial:   nil,
					AuditLog: sql.AuditLog{
						CreatedOn: t1,
						CreatedBy: 2,
						UpdatedOn: t2,
						UpdatedBy: 1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := mapCiMaterialToPipelineMaterial(tt.args.ciPipeline, tt.args.userId, tt.args.oldPipeline)
			assert.Equal(t, len(res), 1)
			assert.Equal(t, tt.want[0].Id, res[0].Id)
			assert.Equal(t, tt.want[0].GitMaterialId, res[0].GitMaterialId)
			assert.Equal(t, tt.want[0].CiPipelineId, res[0].CiPipelineId)
			assert.Equal(t, tt.want[0].Type, res[0].Type)
			assert.Equal(t, tt.want[0].Value, res[0].Value)
			assert.Equal(t, tt.want[0].Regex, res[0].Regex)
			assert.Equal(t, tt.want[0].AuditLog.CreatedBy, res[0].AuditLog.CreatedBy)
			assert.Equal(t, tt.want[0].AuditLog.CreatedOn, res[0].AuditLog.CreatedOn)
			assert.Equal(t, tt.want[0].AuditLog.UpdatedBy, res[0].AuditLog.UpdatedBy)
		})
	}
}
