package pipeline

import (
	"github.com/devtron-labs/devtron/pkg/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"reflect"
	"strings"
	"testing"
)

func TestStageServiceUtil_CreatePreAndPostStageResponse(t *testing.T) {
	type args struct {
		cdPipelineConfigObject *bean.CDPipelineConfigObject
		version                string
	}
	tests := []struct {
		name    string
		payload args
		want    *bean.CDPipelineConfigObject
		wantErr bool
	}{
		{name: "Test1_v1_preDeploy_nil_preStage_present", payload: args{
			cdPipelineConfigObject: &bean.CDPipelineConfigObject{
				Id:                 4,
				EnvironmentId:      4,
				EnvironmentName:    "env3",
				Description:        "",
				CiPipelineId:       4,
				TriggerType:        "AUTOMATIC",
				Name:               "cd-1-e7l1",
				Namespace:          "ns3",
				AppWorkflowId:      4,
				DeploymentTemplate: "ROLLING",
				PreStage: bean.CdStage{
					TriggerType: "AUTOMATIC",
					Name:        "Pre-Deployment",
					Config:      "version: 0.0.1\ncdPipelineConf:\n  - beforeStages:\n      - name: test-1\n        script: |\n          date > test.report\n          echo 'hello from inside pre-cd'\n        outputLocation: ./test.report\n      - name: test-2\n        script: |\n          date > test2.report\n        outputLocation: ./test2.report",
				},
				PostStage: bean.CdStage{
					TriggerType: "AUTOMATIC",
					Name:        "Post-Deployment",
					Config:      "version: 0.0.1\ncdPipelineConf:\n  - afterStages:\n      - name: test-1\n        script: |\n          date > test.report\n          echo 'hello from inside post-cd'\n        outputLocation: ./test.report\n      - name: test-2\n        script: |\n          date > test2.report\n        outputLocation: ./test2.report",
				},
				ParentPipelineId:   4,
				ParentPipelineType: "CI_PIPELINE",
				DeploymentAppType:  "helm",
				PreDeployStage:     nil,
				PostDeployStage:    nil,
			},
			version: "v1",
		}, want: &bean.CDPipelineConfigObject{
			Id:                 4,
			EnvironmentId:      4,
			EnvironmentName:    "env3",
			Description:        "",
			CiPipelineId:       4,
			TriggerType:        "AUTOMATIC",
			Name:               "cd-1-e7l1",
			Namespace:          "ns3",
			AppWorkflowId:      4,
			DeploymentTemplate: "ROLLING",
			PreStage: bean.CdStage{
				TriggerType: "AUTOMATIC",
				Name:        "Pre-Deployment",
				Config:      "version: 0.0.1\ncdPipelineConf:\n  - beforeStages:\n      - name: test-1\n        script: |\n          date > test.report\n          echo 'hello from inside pre-cd'\n        outputLocation: ./test.report\n      - name: test-2\n        script: |\n          date > test2.report\n        outputLocation: ./test2.report",
			},
			PostStage: bean.CdStage{
				TriggerType: "AUTOMATIC",
				Name:        "Post-Deployment",
				Config:      "version: 0.0.1\ncdPipelineConf:\n  - afterStages:\n      - name: test-1\n        script: |\n          date > test.report\n          echo 'hello from inside post-cd'\n        outputLocation: ./test.report\n      - name: test-2\n        script: |\n          date > test2.report\n        outputLocation: ./test2.report",
			},
			ParentPipelineId:   4,
			ParentPipelineType: "CI_PIPELINE",
			DeploymentAppType:  "helm",
			PreDeployStage:     nil,
			PostDeployStage:    nil,
		}, wantErr: false},

		//it will convert preDeploy into preStageYaml
		{name: "Test2_v1_preDeploy_present_preStage_nil", payload: args{
			cdPipelineConfigObject: &bean.CDPipelineConfigObject{
				Id:                 4,
				EnvironmentId:      4,
				EnvironmentName:    "env3",
				Description:        "",
				CiPipelineId:       4,
				TriggerType:        "AUTOMATIC",
				Name:               "cd-1-e7l1",
				Namespace:          "ns3",
				AppWorkflowId:      4,
				DeploymentTemplate: "ROLLING",
				PreStage: bean.CdStage{
					TriggerType: "AUTOMATIC",
					Name:        "",
					Config:      "",
				},
				PostStage: bean.CdStage{
					TriggerType: "AUTOMATIC",
					Name:        "",
					Config:      "",
				},
				ParentPipelineId:   4,
				ParentPipelineType: "CI_PIPELINE",
				DeploymentAppType:  "helm",
				PreDeployStage: &bean2.PipelineStageDto{
					Id:          0,
					Name:        "Pre-Deployment",
					Description: "",
					Type:        "PRE_CD",
					Steps: []*bean2.PipelineStageStepDto{
						&bean2.PipelineStageStepDto{
							Id:                  0,
							Name:                "test-1",
							Description:         "",
							Index:               0,
							StepType:            "INLINE",
							OutputDirectoryPath: []string{"./test.report"},
							InlineStepDetail: &bean2.InlineStepDetailDto{
								ScriptType: "SHELL",
								Script:     "date > test.report\n",
							},
						},
						&bean2.PipelineStageStepDto{
							Id:                  0,
							Name:                "test-2",
							Description:         "",
							Index:               0,
							StepType:            "INLINE",
							OutputDirectoryPath: []string{"./test2.report"},
							InlineStepDetail: &bean2.InlineStepDetailDto{
								ScriptType: "SHELL",
								Script:     "date > test2.report\n",
							},
						},
					},
					TriggerType: "AUTOMATIC",
				},
				PostDeployStage: &bean2.PipelineStageDto{
					Id:          0,
					Name:        "Post-Deployment",
					Description: "",
					Type:        "POST_CD",
					Steps: []*bean2.PipelineStageStepDto{
						&bean2.PipelineStageStepDto{
							Id:                  0,
							Name:                "test-1",
							Description:         "",
							Index:               0,
							StepType:            "INLINE",
							OutputDirectoryPath: []string{"./test.report"},
							InlineStepDetail: &bean2.InlineStepDetailDto{
								ScriptType: "SHELL",
								Script:     "date > test.report\n",
							},
						},
						&bean2.PipelineStageStepDto{
							Id:                  0,
							Name:                "test-2",
							Description:         "",
							Index:               0,
							StepType:            "INLINE",
							OutputDirectoryPath: []string{"./test2.report"},
							InlineStepDetail: &bean2.InlineStepDetailDto{
								ScriptType: "SHELL",
								Script:     "date > test2.report\n",
							},
						},
					},
					TriggerType: "AUTOMATIC",
				},
			},
			version: "v1",
		}, want: &bean.CDPipelineConfigObject{
			Id:                 4,
			EnvironmentId:      4,
			EnvironmentName:    "env3",
			Description:        "",
			CiPipelineId:       4,
			TriggerType:        "AUTOMATIC",
			Name:               "cd-1-e7l1",
			Namespace:          "ns3",
			AppWorkflowId:      4,
			DeploymentTemplate: "ROLLING",
			PreStage: bean.CdStage{
				TriggerType: "AUTOMATIC",
				Name:        "Pre-Deployment",
				Config:      "version: \"\"\ncdPipelineConf:\n- beforeStages:\n  - id: 0\n    index: 0\n    name: test-1\n    script: |\n      date > test.report\n    outputLocation: ./test.report\n    runstatus: false\n  - id: 0\n    index: 0\n    name: test-2\n    script: |\n      date > test2.report\n    outputLocation: ./test2.report\n    runstatus: false\n  afterStages: []\n",
			},
			PostStage: bean.CdStage{
				TriggerType: "AUTOMATIC",
				Name:        "Post-Deployment",
				Config:      "version: \"\"\ncdPipelineConf:\n- beforeStages: []\n  afterStages:\n  - id: 0\n    index: 0\n    name: test-1\n    script: |\n      date > test.report\n    outputLocation: ./test.report\n    runstatus: false\n  - id: 0\n    index: 0\n    name: test-2\n    script: |\n      date > test2.report\n    outputLocation: ./test2.report\n    runstatus: false\n",
			},
			ParentPipelineId:   4,
			ParentPipelineType: "CI_PIPELINE",
			DeploymentAppType:  "helm",
			PreDeployStage:     nil,
			PostDeployStage:    nil,
		}, wantErr: false},

		//v2 :- will convert preStageYaml(if present) to preDeployStageSteps
		{name: "Test3_v2_preDeploy_nil_preStage_present", payload: args{
			cdPipelineConfigObject: &bean.CDPipelineConfigObject{
				Id:                 4,
				EnvironmentId:      4,
				EnvironmentName:    "env3",
				Description:        "",
				CiPipelineId:       4,
				TriggerType:        "AUTOMATIC",
				Name:               "cd-1-e7l1",
				Namespace:          "ns3",
				AppWorkflowId:      4,
				DeploymentTemplate: "ROLLING",
				PreStage: bean.CdStage{
					TriggerType: "AUTOMATIC",
					Name:        "Pre-Deployment",
					Config:      "version: \"\"\ncdPipelineConf:\n- beforeStages:\n  - id: 0\n    index: 0\n    name: test-1\n    script: |\n      date > test.report\n    outputLocation: ./test.report\n    runstatus: false\n  - id: 0\n    index: 0\n    name: test-2\n    script: |\n      date > test2.report\n    outputLocation: ./test2.report\n    runstatus: false\n  afterStages: []\n",
				},
				PostStage: bean.CdStage{
					TriggerType: "AUTOMATIC",
					Name:        "Post-Deployment",
					Config:      "version: \"\"\ncdPipelineConf:\n- beforeStages: []\n  afterStages:\n  - id: 0\n    index: 0\n    name: test-1\n    script: |\n      date > test.report\n    outputLocation: ./test.report\n    runstatus: false\n  - id: 0\n    index: 0\n    name: test-2\n    script: |\n      date > test2.report\n    outputLocation: ./test2.report\n    runstatus: false\n",
				},
				ParentPipelineId:   4,
				ParentPipelineType: "CI_PIPELINE",
				DeploymentAppType:  "helm",
				PreDeployStage:     nil,
				PostDeployStage:    nil,
			},
			version: "v2",
		}, want: &bean.CDPipelineConfigObject{
			Id:                 4,
			EnvironmentId:      4,
			EnvironmentName:    "env3",
			Description:        "",
			CiPipelineId:       4,
			TriggerType:        "AUTOMATIC",
			Name:               "cd-1-e7l1",
			Namespace:          "ns3",
			AppWorkflowId:      4,
			DeploymentTemplate: "ROLLING",
			PreStage: bean.CdStage{
				TriggerType: "AUTOMATIC",
				Name:        "Pre-Deployment",
				Config:      "version: \"\"\ncdPipelineConf:\n- beforeStages:\n  - id: 0\n    index: 0\n    name: test-1\n    script: |\n      date > test.report\n    outputLocation: ./test.report\n    runstatus: false\n  - id: 0\n    index: 0\n    name: test-2\n    script: |\n      date > test2.report\n    outputLocation: ./test2.report\n    runstatus: false\n  afterStages: []\n",
			},
			PostStage: bean.CdStage{
				TriggerType: "AUTOMATIC",
				Name:        "Post-Deployment",
				Config:      "version: \"\"\ncdPipelineConf:\n- beforeStages: []\n  afterStages:\n  - id: 0\n    index: 0\n    name: test-1\n    script: |\n      date > test.report\n    outputLocation: ./test.report\n    runstatus: false\n  - id: 0\n    index: 0\n    name: test-2\n    script: |\n      date > test2.report\n    outputLocation: ./test2.report\n    runstatus: false\n",
			},
			ParentPipelineId:   4,
			ParentPipelineType: "CI_PIPELINE",
			DeploymentAppType:  "helm",
			PreDeployStage: &bean2.PipelineStageDto{
				Id:          0,
				Name:        "",
				Description: "",
				Type:        "PRE_CD",
				Steps: []*bean2.PipelineStageStepDto{
					{
						Id:                  0,
						Name:                "test-1",
						Description:         "",
						Index:               1,
						StepType:            "INLINE",
						OutputDirectoryPath: []string{"./test.report"},
						InlineStepDetail: &bean2.InlineStepDetailDto{
							ScriptType: "SHELL",
							Script:     "date > test.report\n",
						},
					},
					{
						Id:                  0,
						Name:                "test-2",
						Description:         "",
						Index:               2,
						StepType:            "INLINE",
						OutputDirectoryPath: []string{"./test2.report"},
						InlineStepDetail: &bean2.InlineStepDetailDto{
							ScriptType: "SHELL",
							Script:     "date > test2.report\n",
						},
					},
				},
				TriggerType: "AUTOMATIC",
			},
			PostDeployStage: &bean2.PipelineStageDto{
				Id:          0,
				Name:        "",
				Description: "",
				Type:        "POST_CD",
				Steps: []*bean2.PipelineStageStepDto{
					{
						Id:                  0,
						Name:                "test-1",
						Description:         "",
						Index:               1,
						StepType:            "INLINE",
						OutputDirectoryPath: []string{"./test.report"},
						InlineStepDetail: &bean2.InlineStepDetailDto{
							ScriptType: "SHELL",
							Script:     "date > test.report\n",
						},
					},
					{
						Id:                  0,
						Name:                "test-2",
						Description:         "",
						Index:               2,
						StepType:            "INLINE",
						OutputDirectoryPath: []string{"./test2.report"},
						InlineStepDetail: &bean2.InlineStepDetailDto{
							ScriptType: "SHELL",
							Script:     "date > test2.report\n",
						},
					},
				},
				TriggerType: "AUTOMATIC",
			},
		}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreatePreAndPostStageResponse(tt.payload.cdPipelineConfigObject, tt.payload.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePreAndPostStageResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(tt.name, "v2") {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreatePreAndPostStageResponse() got = %v, want %v", got, tt.want)
				}
			} else {
				if !reflect.DeepEqual(*got.PreDeployStage, *tt.want.PreDeployStage) {
					t.Errorf("error in DeepEqual of PreDeployStage CreatePreAndPostStageResponse() got = %v, want %v", got, tt.want)
				}
				if !reflect.DeepEqual(*got.PostDeployStage, *tt.want.PostDeployStage) {
					t.Errorf("error in DeepEqual of PreDeployStage CreatePreAndPostStageResponse() got = %v, want %v", got, tt.want)
				}
			}

		})
	}
}
