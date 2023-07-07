package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitopsOrHelmOption(t *testing.T) {

	t.Run("DeploymentAppSetterFunctionIfGitOpsConfiguredExternalUse", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		pipelineBuilderService := NewPipelineBuilderImpl(sugaredLogger, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, util.MergeUtil{Logger: sugaredLogger}, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, &DeploymentServiceTypeConfig{IsInternalUse: false}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		pipelineCreateRequest := &bean.CdPipelines{
			Pipelines: []*bean.CDPipelineConfigObject{
				{
					Id:                            0,
					EnvironmentId:                 1,
					EnvironmentName:               "",
					CiPipelineId:                  1,
					TriggerType:                   "AUTOMATIC",
					Name:                          "cd-1-vo8q",
					Strategies:                    nil,
					Namespace:                     "devtron-demo",
					AppWorkflowId:                 1,
					DeploymentTemplate:            "",
					PreStage:                      bean.CdStage{},
					PostStage:                     bean.CdStage{},
					PreStageConfigMapSecretNames:  bean.PreStageConfigMapSecretNames{},
					PostStageConfigMapSecretNames: bean.PostStageConfigMapSecretNames{},
					RunPreStageInEnv:              false,
					RunPostStageInEnv:             false,
					CdArgoSetup:                   false,
					ParentPipelineId:              1,
					ParentPipelineType:            "CI_PIPELINE",
					DeploymentAppType:             "",
				},
			},
			AppId:  1,
			UserId: 0,
		}
		isGitOpsConfigured := true
		pipelineBuilderService.SetPipelineDeploymentAppType(pipelineCreateRequest, isGitOpsConfigured)

		for _, pipeline := range pipelineCreateRequest.Pipelines {
			assert.Equal(t, pipeline.DeploymentAppType, "argo_cd")
		}

	})

	t.Run("DeploymentAppSetterFunctionIfGitOpsNotConfiguredExternalUse", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		pipelineBuilderService := NewPipelineBuilderImpl(sugaredLogger, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, util.MergeUtil{Logger: sugaredLogger}, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, &DeploymentServiceTypeConfig{IsInternalUse: false}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		pipelineCreateRequest := &bean.CdPipelines{
			Pipelines: []*bean.CDPipelineConfigObject{
				{
					Id:                            0,
					EnvironmentId:                 1,
					EnvironmentName:               "",
					CiPipelineId:                  1,
					TriggerType:                   "AUTOMATIC",
					Name:                          "cd-1-vo8q",
					Strategies:                    nil,
					Namespace:                     "devtron-demo",
					AppWorkflowId:                 1,
					DeploymentTemplate:            "",
					PreStage:                      bean.CdStage{},
					PostStage:                     bean.CdStage{},
					PreStageConfigMapSecretNames:  bean.PreStageConfigMapSecretNames{},
					PostStageConfigMapSecretNames: bean.PostStageConfigMapSecretNames{},
					RunPreStageInEnv:              false,
					RunPostStageInEnv:             false,
					CdArgoSetup:                   false,
					ParentPipelineId:              1,
					ParentPipelineType:            "CI_PIPELINE",
					DeploymentAppType:             "",
				},
			},
			AppId:  1,
			UserId: 0,
		}
		isGitOpsConfigured := false
		pipelineBuilderService.SetPipelineDeploymentAppType(pipelineCreateRequest, isGitOpsConfigured)

		for _, pipeline := range pipelineCreateRequest.Pipelines {
			assert.Equal(t, pipeline.DeploymentAppType, "helm")
		}

	})

	t.Run("DeploymentAppSetterFunctionInternalUse", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		pipelineBuilderService := NewPipelineBuilderImpl(sugaredLogger, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, util.MergeUtil{Logger: sugaredLogger}, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, &DeploymentServiceTypeConfig{IsInternalUse: true}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		pipelineCreateRequestHelm := &bean.CdPipelines{
			Pipelines: []*bean.CDPipelineConfigObject{
				&bean.CDPipelineConfigObject{
					Id:                            0,
					EnvironmentId:                 1,
					EnvironmentName:               "",
					CiPipelineId:                  1,
					TriggerType:                   "AUTOMATIC",
					Name:                          "cd-1-vo8q",
					Strategies:                    nil,
					Namespace:                     "devtron-demo",
					AppWorkflowId:                 1,
					DeploymentTemplate:            "",
					PreStage:                      bean.CdStage{},
					PostStage:                     bean.CdStage{},
					PreStageConfigMapSecretNames:  bean.PreStageConfigMapSecretNames{},
					PostStageConfigMapSecretNames: bean.PostStageConfigMapSecretNames{},
					RunPreStageInEnv:              false,
					RunPostStageInEnv:             false,
					CdArgoSetup:                   false,
					ParentPipelineId:              1,
					ParentPipelineType:            "CI_PIPELINE",
					DeploymentAppType:             "helm",
				},
			},
			AppId:  1,
			UserId: 0,
		}
		isGitOpsConfigured := true
		pipelineBuilderService.SetPipelineDeploymentAppType(pipelineCreateRequestHelm, isGitOpsConfigured)

		for _, pipeline := range pipelineCreateRequestHelm.Pipelines {
			assert.Equal(t, pipeline.DeploymentAppType, "helm")
		}

		pipelineCreateRequestGitOps := &bean.CdPipelines{
			Pipelines: []*bean.CDPipelineConfigObject{
				&bean.CDPipelineConfigObject{
					Id:                            0,
					EnvironmentId:                 1,
					EnvironmentName:               "",
					CiPipelineId:                  1,
					TriggerType:                   "AUTOMATIC",
					Name:                          "cd-1-vo8q",
					Strategies:                    nil,
					Namespace:                     "devtron-demo",
					AppWorkflowId:                 1,
					DeploymentTemplate:            "",
					PreStage:                      bean.CdStage{},
					PostStage:                     bean.CdStage{},
					PreStageConfigMapSecretNames:  bean.PreStageConfigMapSecretNames{},
					PostStageConfigMapSecretNames: bean.PostStageConfigMapSecretNames{},
					RunPreStageInEnv:              false,
					RunPostStageInEnv:             false,
					CdArgoSetup:                   false,
					ParentPipelineId:              1,
					ParentPipelineType:            "CI_PIPELINE",
					DeploymentAppType:             "argo_cd",
				},
			},
			AppId:  1,
			UserId: 0,
		}
		pipelineBuilderService.SetPipelineDeploymentAppType(pipelineCreateRequestGitOps, isGitOpsConfigured)

		for _, pipeline := range pipelineCreateRequestGitOps.Pipelines {
			assert.Equal(t, pipeline.DeploymentAppType, "argo_cd")
		}

	})

	t.Run("TestValidateCDPipelineRequestExternalUse", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		mockedPipelineRepository := mocks.NewPipelineRepository(t)

		mockedPipelineRepository.On("FindActiveByAppIdAndEnvironmentId", 1, 1).Return(nil, nil)

		pipelineBuilderService := NewPipelineBuilderImpl(sugaredLogger, nil, nil, nil, nil,
			mockedPipelineRepository, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, util.MergeUtil{Logger: sugaredLogger}, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, &DeploymentServiceTypeConfig{IsInternalUse: false}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		pipelineCreateRequest := &bean.CdPipelines{
			Pipelines: []*bean.CDPipelineConfigObject{
				&bean.CDPipelineConfigObject{
					Id:                            0,
					EnvironmentId:                 1,
					EnvironmentName:               "",
					CiPipelineId:                  1,
					TriggerType:                   "AUTOMATIC",
					Name:                          "cd-1-vo8q",
					Strategies:                    nil,
					Namespace:                     "devtron-demo",
					AppWorkflowId:                 1,
					DeploymentTemplate:            "",
					PreStage:                      bean.CdStage{},
					PostStage:                     bean.CdStage{},
					PreStageConfigMapSecretNames:  bean.PreStageConfigMapSecretNames{},
					PostStageConfigMapSecretNames: bean.PostStageConfigMapSecretNames{},
					RunPreStageInEnv:              false,
					RunPostStageInEnv:             false,
					CdArgoSetup:                   false,
					ParentPipelineId:              1,
					ParentPipelineType:            "CI_PIPELINE",
					DeploymentAppType:             "",
				},
			},
			AppId:  1,
			UserId: 0,
		}
		isGitOpsConfigured := false
		isGitOpsRequired := pipelineBuilderService.IsGitOpsRequiredForCD(pipelineCreateRequest)

		_, err = pipelineBuilderService.ValidateCDPipelineRequest(pipelineCreateRequest, isGitOpsConfigured, isGitOpsRequired)

		assert.Nil(t, err)

	})

	t.Run("TestValidateCDPipelineRequestInternalUse", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		mockedPipelineRepository := mocks.NewPipelineRepository(t)

		//mockedPipelineRepository.On("FindActiveByAppIdAndEnvironmentId", 1, 1).Return(nil, nil)

		pipelineBuilderService := NewPipelineBuilderImpl(sugaredLogger, nil, nil, nil, nil,
			mockedPipelineRepository, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, util.MergeUtil{Logger: sugaredLogger}, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, &DeploymentServiceTypeConfig{IsInternalUse: true}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		pipelineCreateRequest := &bean.CdPipelines{
			Pipelines: []*bean.CDPipelineConfigObject{
				&bean.CDPipelineConfigObject{
					Id:                            0,
					EnvironmentId:                 1,
					EnvironmentName:               "",
					CiPipelineId:                  1,
					TriggerType:                   "AUTOMATIC",
					Name:                          "cd-1-vo8q",
					Strategies:                    nil,
					Namespace:                     "devtron-demo",
					AppWorkflowId:                 1,
					DeploymentTemplate:            "",
					PreStage:                      bean.CdStage{},
					PostStage:                     bean.CdStage{},
					PreStageConfigMapSecretNames:  bean.PreStageConfigMapSecretNames{},
					PostStageConfigMapSecretNames: bean.PostStageConfigMapSecretNames{},
					RunPreStageInEnv:              false,
					RunPostStageInEnv:             false,
					CdArgoSetup:                   false,
					ParentPipelineId:              1,
					ParentPipelineType:            "CI_PIPELINE",
					DeploymentAppType:             "argo_cd",
				},
			},
			AppId:  1,
			UserId: 0,
		}
		isGitOpsConfigured := false
		isGitOpsRequired := pipelineBuilderService.IsGitOpsRequiredForCD(pipelineCreateRequest)

		InvalidRequest, err := pipelineBuilderService.ValidateCDPipelineRequest(pipelineCreateRequest, isGitOpsConfigured, isGitOpsRequired)

		if err != nil {
			assert.Equal(t, InvalidRequest, false)
		}

	})

}
