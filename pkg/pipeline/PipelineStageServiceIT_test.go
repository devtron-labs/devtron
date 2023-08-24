package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

const CiPipelineStageCreateReqJson = `{"appId":1,"appWorkflowId":0,"action":0,"ciPipelineRequest":{"active":true,"ciMaterial":[{"gitMaterialId":1,"id":0,"source":{"type":"SOURCE_TYPE_BRANCH_FIXED","value":"main","regex":""}}],"dockerArgs":{},"externalCiConfig":{},"id":0,"isExternal":false,"isManual":false,"name":"ci-1-xyze","linkedCount":0,"scanEnabled":false,"preBuildStage":{"id":0,"steps":[{"id":1,"index":1,"name":"Task 1","description":"chbsdkhbc","stepType":"INLINE","directoryPath":"","inlineStepDetail":{"scriptType":"CONTAINER_IMAGE","script":"echo \"ifudsbvnv\"","conditionDetails":[],"inputVariables":[{"id":1,"name":"Hello","value":"","format":"STRING","description":"jnsdvbdvbsd","defaultValue":"","variableType":"GLOBAL","refVariableStepIndex":0,"refVariableName":"WORKING_DIRECTORY","refVariableStage":""}],"outputVariables":null,"commandArgsMap":[{"command":"echo","args":["\"HOSTNAME\"","\"PORT\""]}],"portMap":[{"portOnLocal":8080,"portOnContainer":9090}],"mountCodeToContainer":true,"mountDirectoryFromHost":true,"mountCodeToContainerPath":"/sourcecode","mountPathMap":[{"filePathOnDisk":"./test","filePathOnContainer":"./test_container"}],"containerImagePath":"python:latest","isMountCustomScript":true,"storeScriptAt":"./directory/script"},"outputDirectoryPath":["./test1"]},{"id":2,"index":2,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","stepType":"REF_PLUGIN","directoryPath":"","pluginRefStepDetail":{"id":0,"pluginId":1,"conditionDetails":[{"id":0,"conditionOnVariable":"RelativePathToScript","conditionOperator":"==","conditionType":"TRIGGER","conditionalValue":"svfsv"},{"id":0,"conditionOnVariable":"PrometheusApiKey","conditionOperator":"==","conditionType":"TRIGGER","conditionalValue":"dfbeavafsv"}],"inputVariables":[{"id":1,"name":"RelativePathToScript","format":"STRING","description":"checkout path + script path along with script name","isExposed":true,"allowEmptyValue":false,"defaultValue":"/./script.js","variableType":"NEW","variableStepIndexInPlugin":1,"value":"dethdt","refVariableName":"","refVariableStage":""},{"id":2,"name":"PrometheusUsername","format":"STRING","description":"username of prometheus account","isExposed":true,"allowEmptyValue":true,"defaultValue":"","variableType":"NEW","variableStepIndexInPlugin":1,"value":"ghmfnbd","refVariableName":"","refVariableStage":""},{"id":3,"name":"PrometheusApiKey","format":"STRING","description":"api key of prometheus account","isExposed":true,"allowEmptyValue":true,"defaultValue":"","variableType":"NEW","variableStepIndexInPlugin":1,"value":"afegs","refVariableName":"","refVariableStage":""},{"id":4,"name":"PrometheusRemoteWriteEndpoint","format":"STRING","description":"remote write endpoint of prometheus account","isExposed":true,"allowEmptyValue":true,"defaultValue":"","variableType":"NEW","variableStepIndexInPlugin":1,"value":"aef","refVariableName":"","refVariableStage":""},{"id":5,"name":"OutputType","format":"STRING","description":"output type - LOG or PROMETHEUS","isExposed":true,"allowEmptyValue":false,"defaultValue":"LOG","variableType":"NEW","variableStepIndexInPlugin":1,"value":"fdgn","refVariableName":"","refVariableStage":""}]}},{"id":3,"index":3,"name":"Task 3","description":"sfdbvf","stepType":"INLINE","directoryPath":"","inlineStepDetail":{"scriptType":"SHELL","script":"#!/bin/sh \nset -eo pipefail \n#set -v  ## uncomment this to debug the script \n","conditionDetails":[{"id":0,"conditionOnVariable":"Hello","conditionOperator":"==","conditionType":"PASS","conditionalValue":"aedfrwgwr"},{"id":0,"conditionOnVariable":"Hello","conditionOperator":"!=","conditionType":"PASS","conditionalValue":"tegegr"}],"inputVariables":[],"outputVariables":[{"id":1,"name":"Hello","value":"","format":"STRING","description":"dsuihvsuvhbdv","defaultValue":"","variableType":"NEW","refVariableStepIndex":0,"refVariableName":""}],"commandArgsMap":[{"command":"","args":[]}],"portMap":[],"mountCodeToContainer":false,"mountDirectoryFromHost":false},"outputDirectoryPath":["./test2"]}]},"postBuildStage":{},"dockerConfigOverride":{}}}`
const CiPipelineStageUpdateReqJson = `{"appId":1,"appWorkflowId":3,"action":1,"ciPipelineRequest":{"isManual":false,"dockerArgs":{},"isExternal":false,"parentCiPipeline":0,"parentAppId":0,"appId":1,"externalCiConfig":{"id":0,"webhookUrl":"","payload":"","accessKey":"","payloadOption":null,"schema":null,"responses":null,"projectId":0,"projectName":"","environmentId":"","environmentName":"","environmentIdentifier":"","appId":0,"appName":"","role":""},"ciMaterial":[{"gitMaterialId":1,"id":3,"source":{"type":"SOURCE_TYPE_BRANCH_FIXED","value":"main","regex":""}}],"name":"ci-1-unov","id":3,"active":true,"linkedCount":0,"scanEnabled":false,"appWorkflowId":3,"preBuildStage":{"id":5,"type":"PRE_CI","steps":[{"id":9,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","index":1,"stepType":"REF_PLUGIN","outputDirectoryPath":null,"inlineStepDetail":null,"pluginRefStepDetail":{"pluginId":1,"inputVariables":[{"id":44,"name":"RelativePathToScript","format":"STRING","description":"checkout path + script path along with script name","isExposed":true,"defaultValue":"/./script.js","value":"sfds","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":45,"name":"PrometheusUsername","format":"STRING","description":"username of prometheus account","isExposed":true,"allowEmptyValue":true,"value":"sdf","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":46,"name":"PrometheusApiKey","format":"STRING","description":"api key of prometheus account","isExposed":true,"allowEmptyValue":true,"value":"hter","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":47,"name":"PrometheusRemoteWriteEndpoint","format":"STRING","description":"remote write endpoint of prometheus account","isExposed":true,"allowEmptyValue":true,"value":"ewq","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":48,"name":"OutputType","format":"STRING","description":"output type - LOG or PROMETHEUS","isExposed":true,"defaultValue":"LOG","value":"erg","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""}],"outputVariables":null,"conditionDetails":null},"triggerIfParentStageFail":false},{"id":10,"name":"Task 3","description":"","index":3,"stepType":"INLINE","outputDirectoryPath":["./asap"],"inlineStepDetail":{"scriptType":"CONTAINER_IMAGE","script":null,"storeScriptAt":null,"mountDirectoryFromHost":false,"containerImagePath":"alpine:latest","commandArgsMap":[{"command":"echo","args":["HOSTNAME"]}],"inputVariables":null,"outputVariables":null,"conditionDetails":null,"portMap":[],"mountCodeToContainerPath":null,"mountPathMap":null},"pluginRefStepDetail":null,"triggerIfParentStageFail":false},{"id":8,"name":"Task 1","description":"","index":2,"stepType":"INLINE","outputDirectoryPath":["./test"],"inlineStepDetail":{"scriptType":"SHELL","script":"#!/bin/sh \nset -eo pipefail \necho \"Hello from inside pre-build stage\"\n#set -v  ## uncomment this to debug the script \n","storeScriptAt":"","mountDirectoryFromHost":false,"commandArgsMap":[{"command":"","args":null}],"inputVariables":null,"outputVariables":null,"conditionDetails":null},"pluginRefStepDetail":null,"triggerIfParentStageFail":false}]},"postBuildStage":{},"isDockerConfigOverridden":false,"dockerConfigOverride":{}}}`
const CiPipelineStageDeleteReqJson = `{"appId":1,"appWorkflowId":8,"action":2,"ciPipelineRequest":{"isManual":false,"dockerArgs":{},"isExternal":false,"parentCiPipeline":0,"parentAppId":0,"appId":1,"externalCiConfig":{"id":0,"webhookUrl":"","payload":"","accessKey":"","payloadOption":null,"schema":null,"responses":null,"projectId":0,"projectName":"","environmentId":"","environmentName":"","environmentIdentifier":"","appId":0,"appName":"","role":""},"ciMaterial":[{"gitMaterialId":1,"id":8,"source":{"type":"SOURCE_TYPE_BRANCH_FIXED","value":"main","regex":""}}],"name":"ci-1-unjn","id":8,"active":true,"linkedCount":0,"scanEnabled":false,"appWorkflowId":8,"preBuildStage":{"id":7,"type":"PRE_CI","steps":[{"id":11,"name":"Task 1","description":"","index":1,"stepType":"INLINE","outputDirectoryPath":["./test"],"inlineStepDetail":{"scriptType":"SHELL","script":"#!/bin/sh \nset -eo pipefail \necho \"Prakash\"\n#set -v  ## uncomment this to debug the script \n","storeScriptAt":"","mountDirectoryFromHost":false,"commandArgsMap":[{"command":"","args":null}],"inputVariables":null,"outputVariables":null,"conditionDetails":null},"pluginRefStepDetail":null,"triggerIfParentStageFail":false},{"id":12,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","index":2,"stepType":"REF_PLUGIN","outputDirectoryPath":null,"inlineStepDetail":null,"pluginRefStepDetail":{"pluginId":1,"inputVariables":[{"id":49,"name":"RelativePathToScript","format":"STRING","description":"checkout path + script path along with script name","isExposed":true,"defaultValue":"/./script.js","value":"sfds","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":50,"name":"PrometheusUsername","format":"STRING","description":"username of prometheus account","isExposed":true,"allowEmptyValue":true,"value":"sdf","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":51,"name":"PrometheusApiKey","format":"STRING","description":"api key of prometheus account","isExposed":true,"allowEmptyValue":true,"value":"hter","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":52,"name":"PrometheusRemoteWriteEndpoint","format":"STRING","description":"remote write endpoint of prometheus account","isExposed":true,"allowEmptyValue":true,"value":"ewq","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""},{"id":53,"name":"OutputType","format":"STRING","description":"output type - LOG or PROMETHEUS","isExposed":true,"defaultValue":"LOG","value":"erg","variableType":"NEW","variableStepIndexInPlugin":1,"refVariableStage":""}],"outputVariables":null,"conditionDetails":null},"triggerIfParentStageFail":false},{"id":13,"name":"Task 3","description":"","index":3,"stepType":"INLINE","outputDirectoryPath":["./asap"],"inlineStepDetail":{"scriptType":"CONTAINER_IMAGE","script":"","storeScriptAt":"","mountDirectoryFromHost":false,"containerImagePath":"alpine:latest","commandArgsMap":[{"command":"echo","args":["HOSTNAME"]}],"inputVariables":null,"outputVariables":null,"conditionDetails":null,"portMap":[]},"pluginRefStepDetail":null,"triggerIfParentStageFail":false}]},"isDockerConfigOverridden":false,"dockerConfigOverride":{},"postBuildStage":{}}}`

var ciPipelineId int
var cdPipelineId int

var pipelineStageReq = &bean.PipelineStageDto{
	Id:          0,
	Name:        "",
	Description: "",
	Type:        repository.PIPELINE_STAGE_TYPE_PRE_CD,
	Steps: []*bean.PipelineStageStepDto{
		{
			Id:                  1,
			Name:                "Task-1",
			Description:         "jhbjhbjvjgvj",
			Index:               1,
			StepType:            repository.PIPELINE_STEP_TYPE_INLINE,
			OutputDirectoryPath: []string{"./test"},
			InlineStepDetail: &bean.InlineStepDetailDto{
				OutputVariables: []*bean.StepVariableDto{
					{
						Id:                        1,
						Name:                      "Hello",
						Format:                    "STRING",
						Description:               "dsuihvsuvhbdv",
						IsExposed:                 false,
						AllowEmptyValue:           false,
						DefaultValue:              "",
						Value:                     "",
						ValueType:                 "NEW",
						PreviousStepIndex:         0,
						ReferenceVariableName:     "",
						VariableStepIndexInPlugin: 0,
						ReferenceVariableStage:    "",
					},
				},
				ScriptType: repository2.SCRIPT_TYPE_SHELL,
				Script:     "#!/bin/sh \nset -eo pipefail \necho \"Prakash\"\n#set -v  ## uncomment this to debug the script \n",
			},
		},
		{
			Id:          2,
			Name:        "K6-LoadTesting",
			Description: "K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.",
			Index:       2,
			StepType:    repository.PIPELINE_STEP_TYPE_REF_PLUGIN,
			RefPluginStepDetail: &bean.RefPluginStepDetailDto{
				PluginId: 1,
				ConditionDetails: []*bean.ConditionDetailDto{
					{
						Id:                  0,
						ConditionOnVariable: "RelativePathToScript",
						ConditionType:       "TRIGGER",
						ConditionalOperator: "==",
						ConditionalValue:    "svfsv",
					},
					{
						Id:                  0,
						ConditionOnVariable: "PrometheusApiKey",
						ConditionType:       "TRIGGER",
						ConditionalOperator: "!=",
						ConditionalValue:    "dfbeavafsv",
					},
				},
				InputVariables: []*bean.StepVariableDto{
					{
						Id:                        1,
						Name:                      "RelativePathToScript",
						Format:                    "STRING",
						Description:               "checkout path + script path along with script name",
						IsExposed:                 true,
						AllowEmptyValue:           false,
						DefaultValue:              "/./script.js",
						Value:                     "sfds",
						ValueType:                 "NEW",
						PreviousStepIndex:         0,
						ReferenceVariableName:     "",
						VariableStepIndexInPlugin: 1,
						ReferenceVariableStage:    "",
					},
					{
						Id:                        2,
						Name:                      "PrometheusUsername",
						Format:                    "STRING",
						Description:               "username of prometheus accoun",
						IsExposed:                 true,
						AllowEmptyValue:           true,
						DefaultValue:              "",
						Value:                     "sdf",
						ValueType:                 "NEW",
						PreviousStepIndex:         0,
						ReferenceVariableName:     "",
						VariableStepIndexInPlugin: 1,
						ReferenceVariableStage:    "",
					},
					{
						Id:                        3,
						Name:                      "PrometheusApiKey",
						Format:                    "STRING",
						Description:               "api key of prometheus account",
						IsExposed:                 true,
						AllowEmptyValue:           true,
						DefaultValue:              "",
						Value:                     "gwrsd",
						ValueType:                 "NEW",
						PreviousStepIndex:         0,
						ReferenceVariableName:     "",
						VariableStepIndexInPlugin: 1,
						ReferenceVariableStage:    "",
					},
					{
						Id:                        4,
						Name:                      "PrometheusRemoteWriteEndpoint",
						Format:                    "STRING",
						Description:               "remote write endpoint of prometheus account",
						IsExposed:                 true,
						AllowEmptyValue:           true,
						DefaultValue:              "",
						Value:                     "ewq",
						ValueType:                 "NEW",
						PreviousStepIndex:         0,
						ReferenceVariableName:     "",
						VariableStepIndexInPlugin: 1,
						ReferenceVariableStage:    "",
					},
					{
						Id:                        5,
						Name:                      "OutputType",
						Format:                    "STRING",
						Description:               "output type - LOG or PROMETHEUS",
						IsExposed:                 true,
						AllowEmptyValue:           false,
						DefaultValue:              "LOG",
						Value:                     "Log",
						ValueType:                 "NEW",
						PreviousStepIndex:         0,
						ReferenceVariableName:     "",
						VariableStepIndexInPlugin: 1,
						ReferenceVariableStage:    "",
					},
				},
			},
		},
		{
			Id:                  3,
			Name:                "Task 3",
			Description:         "",
			Index:               3,
			StepType:            repository.PIPELINE_STEP_TYPE_INLINE,
			OutputDirectoryPath: []string{"./asasp"},
			InlineStepDetail: &bean.InlineStepDetailDto{
				ScriptType:               repository2.SCRIPT_TYPE_CONTAINER_IMAGE,
				Script:                   "#!/bin/sh \nset -eo pipefail \necho \"Hello from inside Container Image\"\n#set -v  ## uncomment this to debug the script \n",
				MountCodeToContainer:     true,
				MountCodeToContainerPath: "/sourcecode",
				MountDirectoryFromHost:   true,
				ContainerImagePath:       "alpine:latest",
				CommandArgsMap: []*bean.CommandArgsMap{
					{
						Command: "echo",
						Args:    []string{"HOSTNAME", "PORT"},
					},
				},
				InputVariables: []*bean.StepVariableDto{
					{
						Id:                        1,
						Name:                      "Hello",
						Format:                    "STRING",
						Description:               "jnsdvbdvbsd",
						IsExposed:                 false,
						AllowEmptyValue:           false,
						DefaultValue:              "",
						Value:                     "",
						ValueType:                 "GLOBAL",
						PreviousStepIndex:         0,
						ReferenceVariableName:     "WORKING_DIRECTORY",
						VariableStepIndexInPlugin: 0,
						ReferenceVariableStage:    "",
					},
				},
			},
		},
	},
	TriggerType: pipelineConfig.TRIGGER_TYPE_AUTOMATIC,
}

func getDbConnAndLoggerService(t *testing.T) (*zap.SugaredLogger, *pg.DB) {
	cfg, _ := sql.GetConfig()
	logger, err := utils.NewSugardLogger()
	assert.Nil(t, err)
	dbConnection, err := sql.NewDbConnection(cfg, logger)
	assert.Nil(t, err)

	return logger, dbConnection
}

func getPipelineStageServiceImpl(t *testing.T) *PipelineStageServiceImpl {
	logger, dbConnection := getDbConnAndLoggerService(t)

	pipelineStageRepoImpl := repository.NewPipelineStageRepository(logger, dbConnection)
	pipelineRepoImpl := pipelineConfig.NewPipelineRepositoryImpl(dbConnection, logger)

	pipelineStageServiceImpl := NewPipelineStageService(logger, pipelineStageRepoImpl, nil, pipelineRepoImpl)
	return pipelineStageServiceImpl
}

func getCiPatchReq(action string) *bean2.CiPatchRequest {
	ciPatchReq := &bean2.CiPatchRequest{}
	if action == "create" {
		json.Unmarshal([]byte(CiPipelineStageCreateReqJson), ciPatchReq)
	} else if action == "update" {
		json.Unmarshal([]byte(CiPipelineStageUpdateReqJson), ciPatchReq)
	} else {
		json.Unmarshal([]byte(CiPipelineStageDeleteReqJson), ciPatchReq)
	}

	return ciPatchReq
}

func setupSuite(t *testing.T) func(t *testing.T) {
	logger, dbConnection := getDbConnAndLoggerService(t)
	tx, _ := dbConnection.Begin()
	defer tx.Rollback()
	ciPipelineRepoImpl := pipelineConfig.NewCiPipelineRepositoryImpl(dbConnection, logger)
	pipelineRepoImpl := pipelineConfig.NewPipelineRepositoryImpl(dbConnection, logger)
	appStatusRepoImpl := appStatus.NewAppStatusRepositoryImpl(dbConnection, logger)
	environmentRepositoryImpl := repository3.NewEnvironmentRepositoryImpl(dbConnection, logger, appStatusRepoImpl)
	//create ci-pipeline entry in db
	var userId int32 = 1
	appId := 1
	ciTemplateId := 1
	ciPipelineName := "ci-1-ersfjkajsdnceusdc"
	ciPipeline := &pipelineConfig.CiPipeline{
		AppId:        appId,
		CiTemplateId: ciTemplateId,
		Name:         ciPipelineName,
		Active:       true,
		Deleted:      false,
		AuditLog:     sql.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err := ciPipelineRepoImpl.Save(ciPipeline, tx)
	assert.Nil(t, err)
	ciPipelineId = ciPipeline.Id
	//create a random env

	charset := "abcdefghijklmnopqrstuvwxyz"
	// Getting random character
	randomEnvName := charset[rand.Intn(len(charset))]
	randomNsName := charset[rand.Intn(len(charset))]
	envIdentifier := string(randomEnvName) + "__" + string(randomNsName)
	model := &repository3.Environment{
		Name:                  string(randomEnvName),
		ClusterId:             1,
		Active:                true,
		Namespace:             string(randomNsName),
		Default:               false,
		EnvironmentIdentifier: envIdentifier,
	}
	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	err = environmentRepositoryImpl.Create(model)
	assert.Nil(t, err)
	envId := model.Id

	randomAppName := charset[rand.Intn(len(charset))]
	//create pipeline entry in db
	pipeline := &pipelineConfig.Pipeline{
		EnvironmentId:        envId,
		AppId:                appId,
		Name:                 "cd-2-dakjsnbcdbcvudvb",
		Deleted:              false,
		CiPipelineId:         ciPipelineId,
		TriggerType:          "AUTOMATIC",
		PreTriggerType:       "AUTOMATIC",
		PostTriggerType:      "AUTOMATIC",
		RunPreStageInEnv:     false,
		RunPostStageInEnv:    false,
		DeploymentAppCreated: false,
		DeploymentAppType:    "helm",
		DeploymentAppName:    fmt.Sprintf("%s-%s", string(randomAppName), string(randomEnvName)),
		AuditLog:             sql.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err = pipelineRepoImpl.Save([]*pipelineConfig.Pipeline{pipeline}, tx)
	assert.Nil(t, err)
	cdPipelineId = pipeline.Id
	_ = tx.Commit()
	// Return a function to teardown the test
	return func(t *testing.T) {
		tx, _ = dbConnection.Begin()
		err = pipelineRepoImpl.Delete(cdPipelineId, userId, tx)
		assert.Nil(t, err)
		err = environmentRepositoryImpl.MarkEnvironmentDeleted(model, tx)
		assert.Nil(t, err)
		p := &pipelineConfig.CiPipeline{
			Id:       ciPipelineId,
			Deleted:  true,
			AuditLog: sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
		}
		err := ciPipelineRepoImpl.Update(p, tx)
		assert.Nil(t, err)
		_ = tx.Commit()
	}
}

func TestPipelineStageService_CreatePipelineStage(t *testing.T) {
	t.SkipNow()
	t.Run("Create Stage With Valid Pre CD Payload", func(t *testing.T) {
		pipelineStageServiceImpl := getPipelineStageServiceImpl(t)
		err := pipelineStageServiceImpl.CreatePipelineStage(pipelineStageReq, repository.PIPELINE_STAGE_TYPE_PRE_CD, cdPipelineId, 1)
		assert.Nil(t, err)
	})

	t.Run("Create Stage With Valid Post CD Payload", func(t *testing.T) {
		pipelineStageServiceImpl := getPipelineStageServiceImpl(t)
		pipelineStageReq.Type = repository.PIPELINE_STAGE_TYPE_POST_CD
		err := pipelineStageServiceImpl.CreatePipelineStage(pipelineStageReq, repository.PIPELINE_STAGE_TYPE_POST_CD, cdPipelineId, 1)
		assert.Nil(t, err)
	})

	t.Run("Create Stage With Invalid CD Type", func(t *testing.T) {
		pipelineStageServiceImpl := getPipelineStageServiceImpl(t)
		pipelineStageReq.Type = repository.PIPELINE_STAGE_TYPE_POST_CD
		err := pipelineStageServiceImpl.CreatePipelineStage(pipelineStageReq, "RANDOM_TYPE", cdPipelineId, 1)
		assert.Equal(t, err.Error(), "unknown stage type")
	})

	t.Run("Create Stage With Valid Pre CI Payload", func(t *testing.T) {
		ciPatchReq := getCiPatchReq("create")
		pipelineStageServiceImpl := getPipelineStageServiceImpl(t)
		err := pipelineStageServiceImpl.CreatePipelineStage(ciPatchReq.CiPipeline.PreBuildStage, repository.PIPELINE_STAGE_TYPE_PRE_CI, ciPipelineId, 1)
		assert.Nil(t, err)
	})

}

func TestPipelineStageService_GetCdPipelineStageDataDeepCopy(t *testing.T) {
	t.SkipNow()
	pipelineStageServiceImpl := getPipelineStageServiceImpl(t)
	logger, dbConnection := getDbConnAndLoggerService(t)
	pipelineStageRepoImpl := repository.NewPipelineStageRepository(logger, dbConnection)

	//preparing want payload for preDeployStage
	stage, err := pipelineStageRepoImpl.GetCdStageByCdPipelineIdAndStageType(cdPipelineId, repository.PIPELINE_STAGE_TYPE_PRE_CD)
	assert.Nil(t, err)
	pipelineStageReq.Id = stage.Id
	steps, err := pipelineStageRepoImpl.GetAllStepsByStageId(stage.Id)
	assert.Nil(t, err)

	for i, step := range steps {
		pipelineStageReq.Steps[i].Id = step.Id
	}

	tests := []struct {
		name         string
		cdPipelineId int
		want         *bean.PipelineStageDto
		wantErr      bool
	}{
		{
			name:         "get_cd_pipeline_stage_data_with_valid_cd_pipeline_id",
			cdPipelineId: cdPipelineId,
			want:         pipelineStageReq,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//testing of preDeployStage also covers postDeployStage
			gotPreDeployStage, _, gotErr := pipelineStageServiceImpl.GetCdPipelineStageDataDeepCopy(tt.cdPipelineId)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetCdPipelineStageDataDeepCopy error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotPreDeployStage.Id != tt.want.Id && gotPreDeployStage.Type != tt.want.Type &&
				gotPreDeployStage.TriggerType != tt.want.TriggerType {
				t.Errorf("GetCdPipelineStageDataDeepCopy() got = %+v, want %+v", gotPreDeployStage, tt.want)
			}
			for i, step := range gotPreDeployStage.Steps {
				d, _ := json.Marshal(gotPreDeployStage)
				fmt.Printf("%s\n", d)
				stepFailureCondition := step.Id != tt.want.Steps[i].Id && step.Name != tt.want.Steps[i].Name && step.Description != tt.want.Steps[i].Description &&
					step.Index != tt.want.Steps[i].Index && step.StepType != tt.want.Steps[i].StepType &&
					!reflect.DeepEqual(step.OutputDirectoryPath, tt.want.Steps[i].OutputDirectoryPath) && step.TriggerIfParentStageFail != tt.want.Steps[i].TriggerIfParentStageFail
				if stepFailureCondition {
					t.Errorf("GetCdPipelineStageDataDeepCopy() got = %+v, want %+v", *step, *tt.want.Steps[i])
				}
				if step.InlineStepDetail != nil {
					if step.InlineStepDetail.Script != tt.want.Steps[i].InlineStepDetail.Script &&
						step.InlineStepDetail.ScriptType != tt.want.Steps[i].InlineStepDetail.ScriptType {
						t.Errorf("GetCdPipelineStageDataDeepCopy() got = %+v, want %+v", *step, *tt.want.Steps[i])
					}
				}

				if step.RefPluginStepDetail != nil {
					if step.RefPluginStepDetail.PluginId != tt.want.Steps[i].RefPluginStepDetail.PluginId {
						t.Errorf("GetCdPipelineStageDataDeepCopy() got = %+v, want %+v", *step, *tt.want.Steps[i])
					}
				}

			}

		})
	}
}

func TestPipelineStageService_UpdatePipelineStage(t *testing.T) {
	t.SkipNow()
	tests := []struct {
		name    string
		payload *bean.PipelineStageDto
		want    error
		wantErr bool
	}{
		{
			name:    "update_pipeline_with_valid_pre-cd_payload",
			payload: pipelineStageReq,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "update_pipeline_with_valid_post-cd_payload",
			payload: pipelineStageReq,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "update_pipeline_with_valid_pre-cd_payload_with_changed_index_of_steps",
			payload: pipelineStageReq,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "update_pipeline_with_valid_pre-ci_payload",
			payload: getCiPatchReq("update").CiPipeline.PreBuildStage,
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		stageType := repository.PIPELINE_STAGE_TYPE_PRE_CD
		pipelineId := cdPipelineId
		t.Run(tt.name, func(t *testing.T) {
			pipelineStageServiceImpl := getPipelineStageServiceImpl(t)
			if strings.Contains(tt.name, "post-cd") {
				stageType = repository.PIPELINE_STAGE_TYPE_POST_CD
			} else if strings.Contains(tt.name, "pre-ci") {
				stageType = repository.PIPELINE_STAGE_TYPE_PRE_CI
				pipelineId = ciPipelineId
			} else if strings.Contains(tt.name, "post-ci") {
				stageType = repository.PIPELINE_STAGE_TYPE_POST_CI
				pipelineId = ciPipelineId
			}
			tt.payload.Type = stageType
			if tt.name == "update_pipeline_with_valid_pre-cd_payload_with_changed_index_of_steps" {
				tt.payload.Steps[0].Index = 2
				tt.payload.Steps[1].Index = 1
			}

			err := pipelineStageServiceImpl.UpdatePipelineStage(tt.payload, tt.payload.Type, pipelineId, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePipelineStage error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}

func TestPipelineStageService_DeletePipelineStage(t *testing.T) {
	t.SkipNow()
	tests := []struct {
		name    string
		payload *bean.PipelineStageDto
		want    error
		wantErr bool
	}{
		{
			name:    "delete_pipeline_with_valid_pre-cd_payload",
			payload: pipelineStageReq,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "delete_pipeline_with_valid_post-cd_payload",
			payload: pipelineStageReq,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "delete_pipeline_with_valid_pre-ci_payload",
			payload: getCiPatchReq("delete").CiPipeline.PreBuildStage,
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipelineStageServiceImpl := getPipelineStageServiceImpl(t)

			logger, dbConnection := getDbConnAndLoggerService(t)
			pipelineStageRepoImpl := repository.NewPipelineStageRepository(logger, dbConnection)
			tx, _ := dbConnection.Begin()

			stage, _ := pipelineStageRepoImpl.GetCdStageByCdPipelineIdAndStageType(cdPipelineId, repository.PIPELINE_STAGE_TYPE_PRE_CD)
			tt.payload.Id = stage.Id

			if strings.Contains(tt.name, "post-cd") {
				stage, _ = pipelineStageRepoImpl.GetCdStageByCdPipelineIdAndStageType(cdPipelineId, repository.PIPELINE_STAGE_TYPE_POST_CD)
				tt.payload.Id = stage.Id
			} else if strings.Contains(tt.name, "pre-ci") {
				stage, _ = pipelineStageRepoImpl.GetCiStageByCiPipelineIdAndStageType(ciPipelineId, repository.PIPELINE_STAGE_TYPE_PRE_CI)
				tt.payload.Id = stage.Id
			} else if strings.Contains(tt.name, "post-ci") {
				stage, _ = pipelineStageRepoImpl.GetCiStageByCiPipelineIdAndStageType(ciPipelineId, repository.PIPELINE_STAGE_TYPE_POST_CI)
				tt.payload.Id = stage.Id
			}

			err := pipelineStageServiceImpl.DeletePipelineStage(tt.payload, 1, tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePipelineStage error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}

//func TestMain(m *testing.M) {
//	var t *testing.T
//	tearDownSuite := setupSuite(t)
//	code := m.Run()
//	tearDownSuite(t)
//	os.Exit(code)
//}
