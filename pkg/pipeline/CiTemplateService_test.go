package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestCiTemplateService(t *testing.T) {

	//t.SkipNow()
	t.Run("TemplateWithBuildpackConfiguration", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		builderId := "sample-builder"
		//mockCiTemplateRepository := &CiTemplateRepositoryMock{}
		ciTemplateRepositoryMocked := mocks.NewCiTemplateRepository(t)
		templateDbEntity := &pipelineConfig.CiTemplate{
			Id: 1,
			CiBuildConfig: &pipelineConfig.CiBuildConfig{
				Type:          string(bean.BUILDPACK_BUILD_TYPE),
				BuildMetadata: "{\"BuilderId\":\"" + builderId + "\"}",
			},
		}
		ciTemplateRepositoryMocked.On("FindByAppId", 1).Return(templateDbEntity, nil)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, nil, ciTemplateRepositoryMocked, nil)
		templateBean, err := ciTemplateServiceImpl.FindByAppId(1)
		template := templateBean.CiTemplate
		assert.True(t, templateDbEntity.Id == template.Id)
		assert.True(t, templateBean.CiBuildConfig != nil)
		assert.True(t, templateBean.CiBuildConfig.CiBuildType == bean.BUILDPACK_BUILD_TYPE)
		assert.True(t, templateBean.CiBuildConfig.BuildPackConfig.BuilderId == builderId)
	})

	t.Run("TemplateWithOldDockerData", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		mockCiTemplateRepository := &CiTemplateRepositoryMock{}
		argsKey := "hello"
		argsValue := "world"
		buildOptionsKey := "volume"
		buildOptionsValue := "hello1"
		templateDbEntity := &pipelineConfig.CiTemplate{
			Id:                 1,
			TargetPlatform:     "linux/amd64",
			DockerBuildOptions: "{\"" + buildOptionsKey + "\":\"" + buildOptionsValue + "\"}",
			Args:               "{\"" + argsKey + "\":\"" + argsValue + "\"}",
		}
		mockCiTemplateRepository.On("FindByAppId", 1).Return(templateDbEntity, nil)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, nil, mockCiTemplateRepository, nil)
		templateBean, err := ciTemplateServiceImpl.FindByAppId(1)
		template := templateBean.CiTemplate
		assert.True(t, templateDbEntity.Id == template.Id)
		assert.True(t, templateDbEntity.TargetPlatform == template.TargetPlatform)
		assert.True(t, template.CiBuildConfig == nil)
		ciBuildConfig := templateBean.CiBuildConfig
		assert.True(t, ciBuildConfig.CiBuildType == bean.SELF_DOCKERFILE_BUILD_TYPE)
		assert.True(t, ciBuildConfig.DockerBuildConfig != nil)
		assert.True(t, ciBuildConfig.DockerBuildConfig.TargetPlatform == templateDbEntity.TargetPlatform)
		args := ciBuildConfig.DockerBuildConfig.Args
		assert.True(t, args != nil)
		assert.True(t, args[argsKey] == argsValue)
		dockerBuildOptions := ciBuildConfig.DockerBuildConfig.DockerBuildOptions
		assert.True(t, dockerBuildOptions != nil)
		assert.True(t, dockerBuildOptions[buildOptionsKey] == buildOptionsValue, dockerBuildOptions)
	})

	t.Run("TemplateWithManagedDockerData", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		mockCiTemplateRepository := &CiTemplateRepositoryMock{}
		targetPlatform := "linux/amd64"
		dockerfileContent := "FROM node:9\r\n\r\nWORKDIR /app\r\n\r\nRUN npm install -g contentful-cli\r\n\r\nCOPY package.json .\r\nRUN npm install\r\n\r\nCOPY . .\r\n\r\nUSER node\r\nEXPOSE 3000\r\n\r\nCMD [\"npm\", \"run\", \"start:dev\"]"
		notPlatform := "linux/arm64"
		gitMaterialId := 2
		ciBuildConfigId := 3
		managedDockerfileBuildType := bean.MANAGED_DOCKERFILE_BUILD_TYPE
		buildConfigMetadata := &bean.DockerBuildConfig{
			DockerfileContent: dockerfileContent,
			TargetPlatform:    targetPlatform,
		}
		buildMetadata, err := json.Marshal(buildConfigMetadata)
		assert.True(t, err == nil, err)
		templateDbEntity := &pipelineConfig.CiTemplate{
			Id:             1,
			GitMaterialId:  gitMaterialId,
			TargetPlatform: notPlatform,
			CiBuildConfig: &pipelineConfig.CiBuildConfig{
				Id:            ciBuildConfigId,
				Type:          string(managedDockerfileBuildType),
				BuildMetadata: string(buildMetadata),
			},
		}
		mockCiTemplateRepository.On("FindByAppId", 1).Return(templateDbEntity, nil)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, nil, mockCiTemplateRepository, nil)
		templateBean, err := ciTemplateServiceImpl.FindByAppId(1)
		ciTemplateEntity := templateBean.CiTemplate
		ciBuildConfig := templateBean.CiBuildConfig
		assert.True(t, ciTemplateEntity.Id == templateDbEntity.Id)
		assert.True(t, ciBuildConfig.Id == ciBuildConfigId)
		assert.True(t, ciBuildConfig.CiBuildType == managedDockerfileBuildType)
		assert.True(t, ciBuildConfig.GitMaterialId == gitMaterialId)
		assert.True(t, ciBuildConfig.BuildPackConfig == nil)
		dockerBuildConfig := ciBuildConfig.DockerBuildConfig
		assert.True(t, dockerBuildConfig != nil)
		assert.True(t, dockerBuildConfig.DockerfileContent == dockerfileContent)
		assert.True(t, dockerBuildConfig.TargetPlatform == targetPlatform)
	})

	t.Run("ciTemplateOverride", func(t *testing.T) {

	})

	t.Run("getCiTemplate", func(t *testing.T) {
		t.SkipNow()
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		config, err := sql.GetConfig()
		assert.True(t, err == nil, err)
		db, err := sql.NewDbConnection(config, sugaredLogger)
		ciBuildConfigRepositoryImpl := pipelineConfig.NewCiBuildConfigRepositoryImpl(db, sugaredLogger)
		ciBuildConfigServiceImpl := NewCiBuildConfigServiceImpl(sugaredLogger, ciBuildConfigRepositoryImpl)
		ciTemplateRepositoryImpl := pipelineConfig.NewCiTemplateRepositoryImpl(db, sugaredLogger)
		ciTemplateOverrideRepositoryImpl := pipelineConfig.NewCiTemplateOverrideRepositoryImpl(db, sugaredLogger)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, ciBuildConfigServiceImpl, ciTemplateRepositoryImpl, ciTemplateOverrideRepositoryImpl)
		_, err = ciTemplateServiceImpl.FindTemplateOverrideByCiPipelineId(1)
		appId := 1
		ciTemplateBean, err := ciTemplateServiceImpl.FindByAppId(appId)
		assert.True(t, err == nil, err)
		assert.True(t, ciTemplateBean != nil, ciTemplateBean)
		assert.True(t, ciTemplateBean.CiTemplate != nil)
		assert.True(t, ciTemplateBean.CiTemplate.AppId == appId)
		assert.True(t, ciTemplateBean.CiBuildConfig != nil)

		ciBuildConfig := ciTemplateBean.CiBuildConfig

		buildPackConfig := &bean.BuildPackConfig{
			BuilderId: "gcr.io/buildpacks/builder:v1",
		}
		//buildPackConfig.BuilderId = "heroku/buildpacks:20"
		ciBuildConfig.CiBuildType = bean.BUILDPACK_BUILD_TYPE
		ciBuildConfig.BuildPackConfig = buildPackConfig

		//args := make(map[string]string)
		////args["hello"] = "world"
		//input := "FROM node:9\r\n\r\nWORKDIR /app\r\n\r\nRUN npm install -g contentful-cli\r\n\r\nCOPY package.json .\r\nRUN npm install\r\n\r\nCOPY . .\r\n\r\nUSER node\r\nEXPOSE 3000\r\n\r\nCMD [\"npm\", \"run\", \"start:dev\"]"
		//dockerBuildConfig := &bean.DockerBuildConfig{
		//	DockerfilePath:    "Dockerfile",
		//	DockerfileContent: input,
		//	Args:              args,
		//	//TargetPlatform: "linux/amd64",
		//}
		//ciBuildConfig.CiBuildType = bean.MANAGED_DOCKERFILE_BUILD_TYPE
		//ciBuildConfig.DockerBuildConfig = dockerBuildConfig

		err = ciTemplateServiceImpl.Update(ciTemplateBean)
		assert.True(t, err == nil, err)
	})

	t.Run("escaping char test", func(t *testing.T) {
		t.SkipNow()
		input := "FROM debian\r\nRUN export node_version=\"0.10\" \\\r\n&& apt-get update && apt-get -y install nodejs=\"$node_verion\"\r\nCOPY package.json usr/src/app\r\nRUN cd /usr/src/app \\\r\n&& npm install node-static\r\n\r\nEXPOSE 80000\r\nCMD [\"npm\", \"start\"]"
		//fmt.Println(input)
		output := fmt.Sprint(input)
		fmt.Println(output)
		f, err := os.Create("/tmp/devtroncd/data.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err2 := f.WriteString(input)

		if err2 != nil {
			log.Fatal(err2)
		}
		fmt.Println("done")
	})

}
