package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
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

		buildPackConfig := &bean.BuildPackConfig{
			BuilderId: "gcr.io/buildpacks/builder:v1",
		}
		//buildPackConfig.BuilderId = "heroku/buildpacks:20"
		ciBuildConfig := ciTemplateBean.CiBuildConfig
		ciBuildConfig.CiBuildType = bean.BUILDPACK_BUILD_TYPE
		ciBuildConfig.BuildPackConfig = buildPackConfig
		err = ciTemplateServiceImpl.Update(ciTemplateBean)
		assert.True(t, err == nil, err)
	})

}
