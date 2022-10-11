package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	pipelineMocks "github.com/devtron-labs/devtron/pkg/pipeline/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"os"
	"testing"
)

func TestCiTemplateService(t *testing.T) {

	t.Run("FetchTemplateWithBuildpackConfiguration", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		builderId := "sample-builder"
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
		assert.Equal(t, template.Id, templateDbEntity.Id)
		assert.NotNil(t, templateBean.CiBuildConfig)
		assert.Equal(t, bean.BUILDPACK_BUILD_TYPE, templateBean.CiBuildConfig.CiBuildType)
		assert.Equal(t, builderId, templateBean.CiBuildConfig.BuildPackConfig.BuilderId)
	})

	t.Run("FetchTemplateWithOldDockerData", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		mockCiTemplateRepository := mocks.NewCiTemplateRepository(t)
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
		assert.Equal(t, template.Id, templateDbEntity.Id)
		assert.Equal(t, template.TargetPlatform, templateDbEntity.TargetPlatform)
		assert.Nil(t, template.CiBuildConfig)
		ciBuildConfig := templateBean.CiBuildConfig
		assert.Equal(t, bean.SELF_DOCKERFILE_BUILD_TYPE, ciBuildConfig.CiBuildType)
		assert.NotNil(t, ciBuildConfig.DockerBuildConfig)
		assert.Equal(t, templateDbEntity.TargetPlatform, ciBuildConfig.DockerBuildConfig.TargetPlatform)
		args := ciBuildConfig.DockerBuildConfig.Args
		assert.NotEmpty(t, args)
		assert.Equal(t, argsValue, args[argsKey])
		dockerBuildOptions := ciBuildConfig.DockerBuildConfig.DockerBuildOptions
		assert.NotEmpty(t, dockerBuildOptions)
		assert.Equal(t, buildOptionsValue, dockerBuildOptions[buildOptionsKey])
	})

	t.Run("FetchTemplateWithManagedDockerData", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		mockCiTemplateRepository := mocks.NewCiTemplateRepository(t)
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
		assert.Nil(t, err)
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
		assert.Equal(t, templateDbEntity.Id, ciTemplateEntity.Id)
		assert.Equal(t, ciBuildConfigId, ciBuildConfig.Id)
		assert.Equal(t, managedDockerfileBuildType, ciBuildConfig.CiBuildType)
		assert.Equal(t, gitMaterialId, ciBuildConfig.GitMaterialId)
		assert.Nil(t, ciBuildConfig.BuildPackConfig)
		dockerBuildConfig := ciBuildConfig.DockerBuildConfig
		assert.NotNil(t, dockerBuildConfig)
		assert.Nil(t, ciBuildConfig.BuildPackConfig)
		assert.Equal(t, dockerfileContent, dockerBuildConfig.DockerfileContent)
		assert.Equal(t, targetPlatform, dockerBuildConfig.TargetPlatform)
	})

	t.Run("ciTemplateOverrideWithOldDockerData", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		mockedCiTemplateOverrideRepository := mocks.NewCiTemplateOverrideRepository(t)
		appId := 1
		mockedTemplateOverrides := []*pipelineConfig.CiTemplateOverride{{
			Id:             1,
			CiPipelineId:   2,
			GitMaterialId:  3,
			DockerfilePath: "Dockerfile",
		}, {
			Id:             2,
			CiPipelineId:   3,
			GitMaterialId:  3,
			DockerfilePath: "Dockerfile_ea",
		}}
		mockedCiTemplateOverrideRepository.On("FindByAppId", appId).Return(mockedTemplateOverrides, nil)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, nil, nil, mockedCiTemplateOverrideRepository)
		templateBeans, err := ciTemplateServiceImpl.FindTemplateOverrideByAppId(appId)
		assert.Nil(t, err)
		assert.Equal(t, len(mockedTemplateOverrides), len(templateBeans))
		for index, _ := range templateBeans {
			templateBean := templateBeans[index]
			mockedTemplateOverride := mockedTemplateOverrides[index]
			templateOverride := templateBean.CiTemplateOverride
			ciBuildConfig := templateBean.CiBuildConfig
			assert.Equal(t, mockedTemplateOverride.Id, templateOverride.Id)
			assert.Equal(t, mockedTemplateOverride.GitMaterialId, templateOverride.GitMaterialId)
			assert.Equal(t, mockedTemplateOverride.CiPipelineId, templateOverride.CiPipelineId)
			assert.Equal(t, bean.SELF_DOCKERFILE_BUILD_TYPE, ciBuildConfig.CiBuildType)
			assert.Nil(t, ciBuildConfig.BuildPackConfig)
			assert.Empty(t, ciBuildConfig.DockerBuildConfig.DockerBuildOptions, "docker build options not supported in pipeline override")
			assert.Equal(t, mockedTemplateOverride.DockerfilePath, ciBuildConfig.DockerBuildConfig.DockerfilePath)
		}
	})

	t.Run("ciTemplateOverrideWithBuildPackData", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		mockedCiTemplateOverrideRepository := mocks.NewCiTemplateOverrideRepository(t)
		appId := 1
		builderId1 := "sample-builder-1"
		mockedTemplateOverrides := []*pipelineConfig.CiTemplateOverride{{
			Id:            1,
			CiPipelineId:  2,
			GitMaterialId: 3,
			CiBuildConfig: &pipelineConfig.CiBuildConfig{
				Type:          string(bean.BUILDPACK_BUILD_TYPE),
				BuildMetadata: "{\"BuilderId\":\"" + builderId1 + "\"}",
			},
		}, {
			Id:            2,
			CiPipelineId:  3,
			GitMaterialId: 3,
			CiBuildConfig: &pipelineConfig.CiBuildConfig{
				Type:          string(bean.BUILDPACK_BUILD_TYPE),
				BuildMetadata: "{\"BuilderId\":\"" + builderId1 + "\"}",
			},
		}}
		mockedCiTemplateOverrideRepository.On("FindByAppId", appId).Return(mockedTemplateOverrides, nil)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, nil, nil, mockedCiTemplateOverrideRepository)
		templateBeans, err := ciTemplateServiceImpl.FindTemplateOverrideByAppId(appId)
		assert.Nil(t, err)
		assert.Equal(t, len(mockedTemplateOverrides), len(templateBeans))
		for index, _ := range templateBeans {
			templateBean := templateBeans[index]
			mockedTemplateOverride := mockedTemplateOverrides[index]
			templateOverride := templateBean.CiTemplateOverride
			ciBuildConfig := templateBean.CiBuildConfig
			assert.Equal(t, mockedTemplateOverride.Id, templateOverride.Id)
			assert.Equal(t, mockedTemplateOverride.GitMaterialId, templateOverride.GitMaterialId)
			assert.Equal(t, mockedTemplateOverride.CiPipelineId, templateOverride.CiPipelineId)
			assert.Equal(t, bean.BUILDPACK_BUILD_TYPE, ciBuildConfig.CiBuildType)
			assert.Nil(t, ciBuildConfig.DockerBuildConfig)
			assert.NotNil(t, ciBuildConfig.BuildPackConfig)
			assert.Equal(t, builderId1, ciBuildConfig.BuildPackConfig.BuilderId)
		}
	})

	t.Run("templateOverrideWithManagedDockerfileAndBuildpack", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		mockedCiTemplateOverrideRepository := mocks.NewCiTemplateOverrideRepository(t)
		appId := 1
		dockerfileContent := "FROM node:9\r\n\r\nWORKDIR /app\r\n\r\nRUN npm install -g contentful-cli\r\n\r\nCOPY package.json .\r\nRUN npm install\r\n\r\nCOPY . .\r\n\r\nUSER node\r\nEXPOSE 3000\r\n\r\nCMD [\"npm\", \"run\", \"start:dev\"]"
		targetPlatform := "linux/amd64"
		builderId := "sample-builder"
		buildConfigMetadata := &bean.DockerBuildConfig{
			DockerfileContent: dockerfileContent,
			TargetPlatform:    targetPlatform,
		}
		buildMetadata, err := json.Marshal(buildConfigMetadata)
		assert.Nil(t, err)
		mockedTemplateOverrides := []*pipelineConfig.CiTemplateOverride{{
			Id:            1,
			CiPipelineId:  2,
			GitMaterialId: 3,
			CiBuildConfig: &pipelineConfig.CiBuildConfig{
				Type:          string(bean.MANAGED_DOCKERFILE_BUILD_TYPE),
				BuildMetadata: string(buildMetadata),
			},
		}, {
			Id:            2,
			CiPipelineId:  3,
			GitMaterialId: 3,
			CiBuildConfig: &pipelineConfig.CiBuildConfig{
				Type:          string(bean.BUILDPACK_BUILD_TYPE),
				BuildMetadata: "{\"BuilderId\":\"" + builderId + "\"}",
			},
		}}
		mockedCiTemplateOverrideRepository.On("FindByAppId", appId).Return(mockedTemplateOverrides, nil)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, nil, nil, mockedCiTemplateOverrideRepository)
		templateBeans, err := ciTemplateServiceImpl.FindTemplateOverrideByAppId(appId)
		assert.Nil(t, err)
		assert.Equal(t, len(mockedTemplateOverrides), len(templateBeans))
		for index, _ := range templateBeans {
			templateBean := templateBeans[index]
			mockedTemplateOverride := mockedTemplateOverrides[index]
			templateOverride := templateBean.CiTemplateOverride
			ciBuildConfig := templateBean.CiBuildConfig
			assert.Equal(t, mockedTemplateOverride.Id, templateOverride.Id)
			assert.Equal(t, mockedTemplateOverride.GitMaterialId, templateOverride.GitMaterialId)
			assert.Equal(t, mockedTemplateOverride.CiPipelineId, templateOverride.CiPipelineId)
			if ciBuildConfig.CiBuildType == bean.MANAGED_DOCKERFILE_BUILD_TYPE {
				assert.Equal(t, bean.MANAGED_DOCKERFILE_BUILD_TYPE, ciBuildConfig.CiBuildType)
				assert.Nil(t, ciBuildConfig.BuildPackConfig)
				assert.NotNil(t, ciBuildConfig.DockerBuildConfig)
				assert.Equal(t, dockerfileContent, ciBuildConfig.DockerBuildConfig.DockerfileContent)
			} else if ciBuildConfig.CiBuildType == bean.BUILDPACK_BUILD_TYPE {
				assert.Equal(t, bean.BUILDPACK_BUILD_TYPE, ciBuildConfig.CiBuildType)
				assert.Nil(t, ciBuildConfig.DockerBuildConfig)
				assert.NotNil(t, ciBuildConfig.BuildPackConfig)
				assert.Equal(t, builderId, ciBuildConfig.BuildPackConfig.BuilderId)
			}
		}
	})

	t.Run("UpdateTemplateOverrideWithBuildConfig", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		mockedCiTemplateOverrideRepository := mocks.NewCiTemplateOverrideRepository(t)
		mockedBuildConfigService := pipelineMocks.NewCiBuildConfigService(t)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, mockedBuildConfigService, nil, mockedCiTemplateOverrideRepository)
		mockedCiTemplateBean := &bean.CiTemplateBean{}
		materialId := 3
		mockedTemplateOverrideId := 1
		mockedCiBuildConfigId := 5
		mockedTemplateOverride := &pipelineConfig.CiTemplateOverride{Id: mockedTemplateOverrideId, CiPipelineId: 2, GitMaterialId: materialId}
		mockedCiTemplateBean.CiTemplateOverride = mockedTemplateOverride
		dockerBuildOptions := map[string]string{}
		dockerBuildOptions["volume"] = "abcd:defg"
		mockedCiTemplateBean.CiBuildConfig = &bean.CiBuildConfigBean{
			Id:                mockedCiBuildConfigId,
			GitMaterialId:     materialId,
			CiBuildType:       bean.SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &bean.DockerBuildConfig{DockerfilePath: "Dockerfile", TargetPlatform: "linux/amd64", DockerBuildOptions: dockerBuildOptions},
		}
		mockedUserId := int32(4)
		mockedCiTemplateBean.UserId = mockedUserId
		mockedCiTemplateOverrideRepository.On("Update", mock.AnythingOfType("*pipelineConfig.CiTemplateOverride")).
			Return(func(templateOverride *pipelineConfig.CiTemplateOverride) *pipelineConfig.CiTemplateOverride {
				assert.Equal(t, mockedCiBuildConfigId, templateOverride.CiBuildConfigId)
				return nil
			}, nil)
		mockedBuildConfigService.On("UpdateOrSave", mock.AnythingOfType("int"), mock.AnythingOfType("int"),
			mock.AnythingOfType("*bean.CiBuildConfigBean"), mock.AnythingOfType("int32")).
			Return(
				func(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean, userId int32) *bean.CiBuildConfigBean {
					assert.Equal(t, 0, templateId)
					assert.Equal(t, mockedTemplateOverrideId, overrideTemplateId)
					assert.Equal(t, mockedUserId, userId)
					mockedBuildConfigBean := mockedCiTemplateBean.CiBuildConfig
					assert.Equal(t, mockedBuildConfigBean, ciBuildConfig)
					return ciBuildConfig
				},
				nil,
			)
		err = ciTemplateServiceImpl.Update(mockedCiTemplateBean)
		assert.Nil(t, err)
		assert.Equal(t, mockedCiBuildConfigId, mockedTemplateOverride.CiBuildConfigId)
	})

	t.Run("UpdateTemplateWithBuildConfig", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		mockedCiTemplateRepository := mocks.NewCiTemplateRepository(t)
		mockedBuildConfigService := pipelineMocks.NewCiBuildConfigService(t)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, mockedBuildConfigService, mockedCiTemplateRepository, nil)
		mockedCiTemplateBean := &bean.CiTemplateBean{}
		materialId := 3
		mockedTemplateId := 1
		mockedCiBuildConfigId := 5
		mockedTemplate := &pipelineConfig.CiTemplate{Id: mockedTemplateId, GitMaterialId: materialId}
		mockedCiTemplateBean.CiTemplate = mockedTemplate
		dockerBuildOptions := map[string]string{}
		dockerBuildOptions["volume"] = "abcd:defg"
		mockedCiTemplateBean.CiBuildConfig = &bean.CiBuildConfigBean{
			Id:                mockedCiBuildConfigId,
			GitMaterialId:     materialId,
			CiBuildType:       bean.SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &bean.DockerBuildConfig{DockerfilePath: "Dockerfile", TargetPlatform: "linux/amd64", DockerBuildOptions: dockerBuildOptions},
		}
		mockedUserId := int32(4)
		mockedCiTemplateBean.UserId = mockedUserId
		mockedCiTemplateRepository.On("Update", mock.AnythingOfType("*pipelineConfig.CiTemplate")).
			Return(func(template *pipelineConfig.CiTemplate) error {
				assert.Equal(t, mockedCiBuildConfigId, template.CiBuildConfigId)
				return nil
			})
		mockedBuildConfigService.On("UpdateOrSave", mock.AnythingOfType("int"), mock.AnythingOfType("int"),
			mock.AnythingOfType("*bean.CiBuildConfigBean"), mock.AnythingOfType("int32")).
			Return(
				func(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean, userId int32) *bean.CiBuildConfigBean {
					assert.Equal(t, 0, overrideTemplateId)
					assert.Equal(t, mockedTemplateId, templateId)
					assert.Equal(t, mockedUserId, userId)
					assert.Equal(t, mockedCiTemplateBean.CiBuildConfig, ciBuildConfig)
					return ciBuildConfig
				},
				nil,
			)
		err = ciTemplateServiceImpl.Update(mockedCiTemplateBean)
		assert.Nil(t, err)
		assert.Equal(t, mockedCiBuildConfigId, mockedTemplate.CiBuildConfigId)
	})

	t.Run("SaveTemplateAndBuildConfig", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		mockedCiTemplateRepository := mocks.NewCiTemplateRepository(t)
		mockedBuildConfigService := pipelineMocks.NewCiBuildConfigService(t)
		ciTemplateServiceImpl := NewCiTemplateServiceImpl(sugaredLogger, mockedBuildConfigService, mockedCiTemplateRepository, nil)
		mockedCiTemplateBean := &bean.CiTemplateBean{}
		materialId := 3
		mockedTemplateId := 7
		mockedCiBuildConfigId := 6
		mockedTemplate := &pipelineConfig.CiTemplate{Id: 0, GitMaterialId: materialId}
		mockedCiTemplateBean.CiTemplate = mockedTemplate
		dockerBuildOptions := map[string]string{}
		dockerBuildOptions["volume"] = "abcd:defg"
		mockedCiTemplateBean.CiBuildConfig = &bean.CiBuildConfigBean{
			Id:                0,
			GitMaterialId:     materialId,
			CiBuildType:       bean.SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &bean.DockerBuildConfig{DockerfilePath: "Dockerfile", TargetPlatform: "linux/amd64", DockerBuildOptions: dockerBuildOptions},
		}
		mockedUserId := int32(4)
		mockedCiTemplateBean.UserId = mockedUserId
		mockedBuildConfigService.On("Save", mock.AnythingOfType("int"), mock.AnythingOfType("int"),
			mock.AnythingOfType("*bean.CiBuildConfigBean"), mock.AnythingOfType("int32")).
			Return(
				func(templateId int, overrideTemplateId int, ciBuildConfig *bean.CiBuildConfigBean, userId int32) error {
					assert.Equal(t, 0, overrideTemplateId)
					assert.Equal(t, mockedTemplate.Id, templateId)
					assert.Equal(t, mockedUserId, userId)
					assert.Equal(t, mockedCiTemplateBean.CiBuildConfig, ciBuildConfig)
					ciBuildConfig.Id = mockedCiBuildConfigId
					return nil
				},
			)
		mockedCiTemplateRepository.On("Save", mock.AnythingOfType("*pipelineConfig.CiTemplate")).
			Return(func(template *pipelineConfig.CiTemplate) error {
				assert.Equal(t, mockedCiBuildConfigId, template.CiBuildConfigId)
				template.Id = mockedTemplateId
				return nil
			})
		err = ciTemplateServiceImpl.Save(mockedCiTemplateBean)
		assert.Nil(t, err)
		assert.Equal(t, mockedTemplateId, mockedTemplate.Id)
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

	t.Run("fetch ci pipeline details", func(t *testing.T) {
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
		appId := 7
		templateBeans, _ := ciTemplateServiceImpl.FindTemplateOverrideByAppId(appId)
		for _, templateBean := range templateBeans {
			buildConfig := templateBean.CiBuildConfig
			fmt.Println(buildConfig)
		}
		ciTemplateBean, _ := ciTemplateServiceImpl.FindByAppId(appId)
		fmt.Println(ciTemplateBean.CiBuildConfig)
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
