package pipeline

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCiBuildConfigService(t *testing.T) {
	t.SkipNow()
	t.Run("buildTypeVsCount", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.True(t, err == nil, err)
		config, err := sql.GetConfig()
		assert.True(t, err == nil, err)
		db, err := sql.NewDbConnection(config, sugaredLogger)
		ciBuildConfigRepositoryImpl := pipelineConfig.NewCiBuildConfigRepositoryImpl(db, sugaredLogger)
		ciBuildConfigServiceImpl := NewCiBuildConfigServiceImpl(sugaredLogger, ciBuildConfigRepositoryImpl)
		countByBuildType := ciBuildConfigServiceImpl.GetCountByBuildType()
		fmt.Println(countByBuildType)
	})
}
