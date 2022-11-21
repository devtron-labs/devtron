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

	t.Run("Save", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		mockedCiTemplateHistoryRepository := mocks.NewCiTemplateHistoryRepository(t)

		CiTemplateHistoryServiceImpl := NewCiTemplateHistoryServiceImpl(mockedCiTemplateHistoryRepository, sugaredLogger)

		mockedCiTemplateHistoryObject := repository.CiTemplateHistory{
			Id:                 1,
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
			BuildMetadata:      "{\"dockerfileRelativePath\":\"Dockerfile\",\"dockerfileContent\":\"# Build Stage\\n# First pull Golang image\\nFROM golang:latest as builder \\n \\n RUN mkdir /app\\nADD . ./app\\nWORKDIR /app\\nCOPY . ./\\nRUN go build -o main .\\nEXPOSE 8080\\nCMD [\\\"/app/main\\\"]\\n\\n# Set environment variable\\n# ENV APP_NAME sams\\n# ENV CMD_PATH main.go\\n \\n# Copy application data into image\\n# COPY . ./\\n# WORKDIR $GOPATH/src/$APP_NAME\\n \\n# # Budild application\\n# RUN CGO_ENABLED=0 go build -v -o /$APP_NAME $GOPATH/src/Package/$CMD_PATH\\n \\n# # Run Stage\\n# FROM alpine:3.14\\n \\n# # Set environment variable\\n# ENV APP_NAME sample-dockerize-app\\n \\n# # Copy only required data into this image\\n# COPY --from=build-env /$APP_NAME .\\n \\n# # Expose application port\\n# EXPOSE 8081\\n \\n# # Start app\\n# CMD ./$APP_NAME\",\"targetPlatform\":\"linux/amd64,linux/arm64\",\"language\":\"Go\"}",
			Trigger:            "update",
			AuditLog:           sql.AuditLog{},
		}
		dockerBuildOptions := map[string]string{}
		dockerBuildOptions["volume"] = "abcd:defg"

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
				Active:             false,
				GitMaterialId:      0,
				DockerBuildOptions: "",
				CiBuildConfigId:    0,
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
				DockerBuildConfig: &bean2.DockerBuildConfig{DockerfilePath: "Dockerfile", TargetPlatform: "linux/amd64", DockerBuildOptions: dockerBuildOptions},
				BuildPackConfig:   nil,
			},
			UserId: 0,
		}

		mockedCiTemplateHistoryRepository.On("Save", mockedCiTemplateHistoryObject).Return(nil)

		err = CiTemplateHistoryServiceImpl.SaveHistory(&ciServiceObject, "update")

		assert.Nil(t, err)

	})

}
