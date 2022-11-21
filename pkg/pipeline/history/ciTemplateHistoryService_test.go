package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
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

		mockedCiTemplateHistoryObject := repository.CiTemplateHistory{
			CiTemplateId:       28,
			AppId:              38,
			DockerRegistryId:   "prakash",
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
				DockerRegistryId:   "prakash",
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
			CiBuildConfig: &bean2.CiBuildConfigBean{
				Id:                20,
				GitMaterialId:     22,
				CiBuildType:       "self-dockerfile-build",
				DockerBuildConfig: &bean2.DockerBuildConfig{DockerfileContent: ""},
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
