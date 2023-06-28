package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"gopkg.in/yaml.v2"
	"strings"
)

type TaskYaml struct {
	Version          string             `yaml:"version"`
	CdPipelineConfig []CdPipelineConfig `yaml:"cdPipelineConf"`
}
type CdPipelineConfig struct {
	BeforeTasks []*Task `yaml:"beforeStages"`
	AfterTasks  []*Task `yaml:"afterStages"`
}
type Task struct {
	Id             int    `json:"id,omitempty"`
	Index          int    `json:"index,omitempty"`
	Name           string `json:"name" yaml:"name"`
	Script         string `json:"script" yaml:"script"`
	OutputLocation string `json:"outputLocation" yaml:"outputLocation"` // file/dir
	RunStatus      bool   `json:"-,omitempty"`                          // task run was attempted or not
}

func ConvertStageYamlScriptsToPipelineStageSteps(cdPipeline *bean2.CDPipelineConfigObject) (*bean2.CDPipelineConfigObject, error) {
	if cdPipeline.PreDeployStage == nil && len(cdPipeline.PreStage.Config) > 0 {
		preDeployStageConverted, err := StageYamlToPipelineStageAdapter(cdPipeline.PreStage.Config, repository2.PIPELINE_STAGE_TYPE_PRE_CD, cdPipeline.PreStage.TriggerType)
		if err != nil {
			return nil, err
		}
		cdPipeline.PreDeployStage = preDeployStageConverted
	}
	if cdPipeline.PostDeployStage == nil && len(cdPipeline.PostStage.Config) > 0 {
		postDeployStageConverted, err := StageYamlToPipelineStageAdapter(cdPipeline.PostStage.Config, repository2.PIPELINE_STAGE_TYPE_POST_CD, cdPipeline.PostStage.TriggerType)
		if err != nil {
			return nil, err
		}
		cdPipeline.PostDeployStage = postDeployStageConverted
	}
	return cdPipeline, nil

}

func checkForOtherParamsInInlineStepDetail(inlineStepDetail *bean.InlineStepDetailDto) bool {
	if len(inlineStepDetail.InputVariables) > 0 || len(inlineStepDetail.OutputVariables) > 0 ||
		len(inlineStepDetail.ConditionDetails) > 0 {
		return false
	}
	return true
}

func StageStepsToCdStageAdapter(deployStage *bean.PipelineStageDto) (*bean2.CdStage, error) {
	cdStage := &bean2.CdStage{
		Name:        deployStage.Name,
		TriggerType: deployStage.TriggerType,
	}
	cdPipelineConfig := make([]CdPipelineConfig, 0)
	beforeTasks := make([]*Task, 0)
	afterTasks := make([]*Task, 0)
	for _, step := range deployStage.Steps {
		if step.InlineStepDetail != nil && checkForOtherParamsInInlineStepDetail(step.InlineStepDetail) {
			if deployStage.Type == repository2.PIPELINE_STAGE_TYPE_PRE_CD {
				beforeTask := &Task{
					Name:           step.Name,
					Script:         step.InlineStepDetail.Script,
					OutputLocation: strings.Join(step.OutputDirectoryPath, ","),
				}
				beforeTasks = append(beforeTasks, beforeTask)

			}
			if deployStage.Type == repository2.PIPELINE_STAGE_TYPE_POST_CD {
				afterTask := &Task{
					Name:           step.Name,
					Script:         step.InlineStepDetail.Script,
					OutputLocation: strings.Join(step.OutputDirectoryPath, ","),
				}
				afterTasks = append(afterTasks, afterTask)
			}

		} else {
			return nil, nil
		}

	}
	cdPipelineConfig = append(cdPipelineConfig, CdPipelineConfig{
		BeforeTasks: beforeTasks,
		AfterTasks:  afterTasks,
	})
	taskYaml := TaskYaml{
		Version:          "",
		CdPipelineConfig: cdPipelineConfig,
	}
	stageConfig, err := yaml.Marshal(taskYaml)
	if err != nil {
		return nil, err
	}
	cdStage.Config = string(stageConfig)
	return cdStage, nil
}

func CreatePreAndPostStageResponse(cdPipeline *bean2.CDPipelineConfigObject, version string) (*bean2.CDPipelineConfigObject, error) {
	var err error
	cdRespMigrated := cdPipeline
	if version == "v2" {
		//in v2, users will be expecting the pre-stage and post-stage in step format
		cdRespMigrated, err = ConvertStageYamlScriptsToPipelineStageSteps(cdPipeline)
		if err != nil {
			return nil, err
		}
		cdRespMigrated.PreStage = bean2.CdStage{}
		cdRespMigrated.PostStage = bean2.CdStage{}

	} else if version == "v1" {
		//in v1, users will be expecting pre-stage and post-stage in yaml format
		if cdPipeline.PreDeployStage != nil {
			//it means that user is trying to access migrated pre-stage stage steps in v1,
			//in that case convert the stage steps into yaml form and send response
			convertedPreCdStage, err := StageStepsToCdStageAdapter(cdPipeline.PreDeployStage)
			if err != nil {
				return nil, err
			}
			cdRespMigrated.PreStage = *convertedPreCdStage
			cdRespMigrated.PreDeployStage = nil
		} else if len(cdPipeline.PreStage.Config) > 0 {
			//set pre stage
			preStage := cdPipeline.PreStage
			cdRespMigrated.PreStage = bean2.CdStage{
				TriggerType: preStage.TriggerType,
				Name:        preStage.Name,
				Config:      preStage.Config,
			}
		} else {
			//users haven't configured pre-cd stage or post-cd stage
		}

		if cdPipeline.PostDeployStage != nil {
			//it means that user is trying to access migrated post-stage stage steps in v1,
			//in that case convert the stage steps into yaml form and send response
			convertedPostCdStage, err := StageStepsToCdStageAdapter(cdPipeline.PostDeployStage)
			if err != nil {
				return nil, err
			}
			cdRespMigrated.PostStage = *convertedPostCdStage
			cdRespMigrated.PostDeployStage = nil
		} else if len(cdPipeline.PostStage.Config) > 0 {
			//set post stage
			postStage := cdPipeline.PostStage
			cdRespMigrated.PostStage = bean2.CdStage{
				TriggerType: postStage.TriggerType,
				Name:        postStage.Name,
				Config:      postStage.Config,
			}
		} else {
			//users haven't configured post-cd stage
		}
	}
	return cdRespMigrated, nil
}

func StageYamlToPipelineStageAdapter(stageConfig string, stageType repository2.PipelineStageType, triggerType pipelineConfig.TriggerType) (*bean.PipelineStageDto, error) {
	//sample stageConfig:= "version: 0.0.1\ncdPipelineConf:\n  - afterStages:\n      - name: test-1\n        script: |\n          date > test.report\n          echo 'hello'\n        outputLocation: ./test.report\n      - name: test-2\n        script: |\n          date > test2.report\n        outputLocation: ./test2.report"

	pipelineStageDto := &bean.PipelineStageDto{}
	var err error
	taskYamlObject, err := ToTaskYaml([]byte(stageConfig))
	if err != nil {
		return nil, err
	}
	for _, task := range taskYamlObject.CdPipelineConfig {
		if len(task.BeforeTasks) > 0 {
			beforeStepIndex := 1
			var beforeStepDtos []*bean.PipelineStageStepDto
			for _, beforeTask := range task.BeforeTasks {
				inlineStepDetail := &bean.InlineStepDetailDto{
					ScriptType: repository.SCRIPT_TYPE_SHELL,
					Script:     beforeTask.Script,
				}
				//index really matters as the task order on the UI is decided by the index field
				stepData := &bean.PipelineStageStepDto{
					Id:                  0,
					Name:                beforeTask.Name,
					Description:         "",
					Index:               beforeStepIndex,
					StepType:            repository2.PIPELINE_STEP_TYPE_INLINE,
					OutputDirectoryPath: []string{beforeTask.OutputLocation},
					InlineStepDetail:    inlineStepDetail,
					RefPluginStepDetail: nil,
				}
				beforeStepDtos = append(beforeStepDtos, stepData)
				beforeStepIndex++
			}
			pipelineStageDto.Steps = beforeStepDtos
			pipelineStageDto.Type = stageType
			pipelineStageDto.Id = 0
			if triggerType != pipelineConfig.TRIGGER_TYPE_AUTOMATIC && triggerType != pipelineConfig.TRIGGER_TYPE_MANUAL {
				pipelineStageDto.TriggerType = pipelineConfig.TRIGGER_TYPE_MANUAL
			} else {
				pipelineStageDto.TriggerType = triggerType
			}

		}

		if len(task.AfterTasks) > 0 {
			afterStepIndex := 1
			var afterStepDtos []*bean.PipelineStageStepDto
			for _, afterTask := range task.AfterTasks {
				inlineStepDetail := &bean.InlineStepDetailDto{
					ScriptType: repository.SCRIPT_TYPE_SHELL,
					Script:     afterTask.Script,
				}
				stepData := &bean.PipelineStageStepDto{
					Id:                  0,
					Name:                afterTask.Name,
					Description:         "",
					Index:               afterStepIndex,
					StepType:            repository2.PIPELINE_STEP_TYPE_INLINE,
					OutputDirectoryPath: []string{afterTask.OutputLocation},
					InlineStepDetail:    inlineStepDetail,
					RefPluginStepDetail: nil,
				}
				afterStepDtos = append(afterStepDtos, stepData)
				afterStepIndex++
			}
			pipelineStageDto.Steps = afterStepDtos
			pipelineStageDto.Type = stageType
			pipelineStageDto.Id = 0
			if triggerType != pipelineConfig.TRIGGER_TYPE_AUTOMATIC && triggerType != pipelineConfig.TRIGGER_TYPE_MANUAL {
				pipelineStageDto.TriggerType = pipelineConfig.TRIGGER_TYPE_MANUAL
			} else {
				pipelineStageDto.TriggerType = triggerType
			}
		}
	}

	return pipelineStageDto, nil
}

func ToTaskYaml(yamlFile []byte) (*TaskYaml, error) {
	taskYaml := &TaskYaml{}
	err := yaml.Unmarshal(yamlFile, taskYaml)
	return taskYaml, err
}
