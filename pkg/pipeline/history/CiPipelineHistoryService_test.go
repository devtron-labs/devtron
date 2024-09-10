/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package history

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean/CiPipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCiPipelineHistoryService(t *testing.T) {
	t.SkipNow()
	t.Run("SaveHistory", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		mockedCiPipelineHistoryRepository := mocks.NewCiPipelineHistoryRepository(t)
		mockedCiPipelineRepository := mocks2.NewCiPipelineRepository(t)

		CiPipelineHistoryServiceImpl := NewCiPipelineHistoryServiceImpl(mockedCiPipelineHistoryRepository, sugaredLogger, mockedCiPipelineRepository)

		PipelineObject := pipelineConfig.CiPipeline{
			Id:                       5,
			AppId:                    2,
			App:                      nil,
			CiTemplateId:             3,
			DockerArgs:               "",
			Name:                     "",
			Version:                  "",
			Active:                   false,
			Deleted:                  false,
			IsManual:                 false,
			IsExternal:               false,
			ParentCiPipeline:         0,
			ScanEnabled:              false,
			IsDockerConfigOverridden: true,
			AuditLog:                 sql.AuditLog{},
		}

		PipelineMaterialsObject := []*pipelineConfig.CiPipelineMaterial{&pipelineConfig.CiPipelineMaterial{
			Id:            0,
			GitMaterialId: 22,
			CiPipelineId:  5,
			Path:          "",
			CheckoutPath:  "",
			Type:          "",
			Value:         "",
			ScmId:         "",
			ScmName:       "",
			ScmVersion:    "",
			Active:        false,
			Regex:         "",
			GitTag:        "",
			AuditLog:      sql.AuditLog{}},
		}

		CiTemplateObject := bean2.CiTemplateBean{
			CiTemplateOverride: &pipelineConfig.CiTemplateOverride{
				Id:               0,
				CiPipelineId:     5,
				DockerRegistryId: "prakash",
				DockerRepository: "prakash1001/sams-repository-3",
				DockerfilePath:   "",
				GitMaterialId:    22,
				Active:           true,
				CiBuildConfigId:  20,
				AuditLog:         sql.AuditLog{},
				GitMaterial:      nil,
				DockerRegistry:   nil,
				CiBuildConfig:    nil,
			},
			CiBuildConfig: &CiPipeline.CiBuildConfigBean{
				Id:                20,
				GitMaterialId:     22,
				CiBuildType:       "self-dockerfile-build",
				DockerBuildConfig: &CiPipeline.DockerBuildConfig{DockerfileContent: ""},
				BuildPackConfig:   nil,
			},
			UserId: 0,
		}

		MockedCiPipelineMaterialJson, _ := json.Marshal(PipelineMaterialsObject)

		MockedCiTemplateOverrideHistory, _ := json.Marshal(
			repository.CiPipelineTemplateOverrideHistoryDTO{
				DockerRegistryId:      "prakash",
				DockerRepository:      "prakash1001/sams-repository-3",
				DockerfilePath:        "",
				Active:                true,
				CiBuildConfigId:       20,
				BuildMetaDataType:     "self-dockerfile-build",
				BuildMetadata:         "{\"dockerfileContent\":\"\"}",
				IsCiTemplateOverriden: true,
				AuditLog:              sql.AuditLog{},
			},
		)

		mockedCiPipelineHistoryObject := repository.CiPipelineHistory{
			Id:                        0,
			CiPipelineId:              5,
			CiTemplateOverrideHistory: string(MockedCiTemplateOverrideHistory),
			CiPipelineMaterialHistory: string(MockedCiPipelineMaterialJson),
			Trigger:                   "update",
			ScanEnabled:               false,
			Manual:                    false,
		}
		mockedCiPipelineObject := repository.CiEnvMappingHistory{
			Id:            0,
			CiPipelineId:  5,
			EnvironmentId: 1,
		}
		CiEnvMapping := &pipelineConfig.CiEnvMapping{
			Id:            1,
			EnvironmentId: 1,
			CiPipelineId:  5,
			Deleted:       false,
		}

		mockedCiPipelineHistoryRepository.On("Save", &mockedCiPipelineHistoryObject).Return(nil).Once()
		mockedCiPipelineRepository.On("FindCiEnvMappingByCiPipelineId", 5).Return(CiEnvMapping, nil).Once()
		mockedCiPipelineHistoryRepository.On("SaveCiEnvMappingHistory", &mockedCiPipelineObject).Return(nil).Once()

		err = CiPipelineHistoryServiceImpl.SaveHistory(&PipelineObject, PipelineMaterialsObject, &CiTemplateObject, "update")

		assert.Nil(t, err)

	})

}
