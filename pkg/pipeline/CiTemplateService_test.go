package pipeline

import (
	"fmt"
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

	t.SkipNow()
	t.Run("getCiTemplate", func(t *testing.T) {
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
		appId := 1
		ciTemplateBean, err := ciTemplateServiceImpl.FindByAppId(appId)
		assert.True(t, err == nil, err)
		assert.True(t, ciTemplateBean != nil, ciTemplateBean)
		assert.True(t, ciTemplateBean.CiTemplate != nil)
		assert.True(t, ciTemplateBean.CiTemplate.AppId == appId)
		assert.True(t, ciTemplateBean.CiBuildConfig != nil)

		ciBuildConfig := ciTemplateBean.CiBuildConfig

		//buildPackConfig := &bean.BuildPackConfig{
		//	BuilderId: "gcr.io/buildpacks/builder:v1",
		//}
		//buildPackConfig.BuilderId = "heroku/buildpacks:20"
		//ciBuildConfig.CiBuildType = bean.BUILDPACK_BUILD_TYPE
		//ciBuildConfig.BuildPackConfig = buildPackConfig

		args := make(map[string]string)
		//args["hello"] = "world"
		input := "FROM node:9\r\n\r\nWORKDIR /app\r\n\r\nRUN npm install -g contentful-cli\r\n\r\nCOPY package.json .\r\nRUN npm install\r\n\r\nCOPY . .\r\n\r\nUSER node\r\nEXPOSE 3000\r\n\r\nCMD [\"npm\", \"run\", \"start:dev\"]"
		dockerBuildConfig := &bean.DockerBuildConfig{
			DockerfilePath:    "Dockerfile",
			DockerfileContent: input,
			Args:              args,
			//TargetPlatform: "linux/amd64",
		}
		ciBuildConfig.CiBuildType = bean.MANAGED_DOCKERFILE_BUILD_TYPE
		ciBuildConfig.DockerBuildConfig = dockerBuildConfig

		err = ciTemplateServiceImpl.Update(ciTemplateBean)
		assert.True(t, err == nil, err)
	})

	t.Run("escaping char test", func(t *testing.T) {
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
