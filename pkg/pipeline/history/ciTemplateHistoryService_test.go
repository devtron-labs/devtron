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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean/CiPipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCiTemplateHistoryService(t *testing.T) {

	t.Run("SaveHistory", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		mockedCiTemplateHistoryRepository := mocks.NewCiTemplateHistoryRepository(t)
		CiTemplateHistoryServiceImpl := NewCiTemplateHistoryServiceImpl(mockedCiTemplateHistoryRepository, sugaredLogger)
		var DockerRegistryId *string = nil
		mockedCiTemplateHistoryObject := repository.CiTemplateHistory{
			CiTemplateId:       28,
			AppId:              38,
			DockerRegistryId:   DockerRegistryId,
			DockerRepository:   "prakash1001/sams-repository-3",
			DockerfilePath:     "",
			Args:               "",
			TargetPlatform:     "",
			BeforeDockerBuild:  "",
			AfterDockerBuild:   "",
			TemplateName:       "",
			Version:            "",
			Active:             true,
			GitMaterialId:      22,
			DockerBuildOptions: "",
			CiBuildConfigId:    20,
			BuildMetaDataType:  "self-dockerfile-build",
			BuildMetadata:      "{\"dockerfileContent\":\"\"}",
			Trigger:            "update",
			AuditLog:           sql.AuditLog{},
		}
		//dockerBuildOptions := map[string]string{}
		//dockerBuildOptions["dockerfileRelativePath"] = "Dockerfile"
		//dockerBuildOptions["dockerfileContent"] = ""
		//dockerBuildOptions["dockerfileRelativePath"] = "Dockerfile"

		ciServiceObject := bean2.CiTemplateBean{
			CiTemplate: &pipelineConfig.CiTemplate{
				Id:                 28,
				AppId:              38,
				DockerRegistryId:   DockerRegistryId,
				DockerRepository:   "prakash1001/sams-repository-3",
				DockerfilePath:     "",
				Args:               "",
				TargetPlatform:     "",
				BeforeDockerBuild:  "",
				AfterDockerBuild:   "",
				TemplateName:       "",
				Version:            "",
				Active:             true,
				GitMaterialId:      22,
				DockerBuildOptions: "",
				CiBuildConfigId:    20,
				AuditLog:           sql.AuditLog{},
				App:                nil,
				DockerRegistry:     nil,
				GitMaterial:        nil,
				CiBuildConfig:      nil,
			},
			CiTemplateOverride: nil,
			CiBuildConfig: &CiPipeline.CiBuildConfigBean{
				Id:                20,
				GitMaterialId:     22,
				CiBuildType:       "self-dockerfile-build",
				DockerBuildConfig: &CiPipeline.DockerBuildConfig{DockerfileContent: ""},
				BuildPackConfig:   nil,
			},
			UserId: 0,
		}

		mockedCiTemplateHistoryRepository.On("Save", &mockedCiTemplateHistoryObject).Return(
			nil,
		)

		err = CiTemplateHistoryServiceImpl.SaveHistory(&ciServiceObject, "update")

		assert.Nil(t, err)

	})

}
