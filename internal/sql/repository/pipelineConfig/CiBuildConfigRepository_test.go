package pipelineConfig

import (
	"fmt"
	"github.com/devtron-labs/common-lib-private/utils"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCiBuildConfigRepository(t *testing.T) {
	t.SkipNow()
	t.Run("GetCountByBuildType", func(t *testing.T) {
		cfg, _ := sql.GetConfig()
		logger, err := utils.NewSugardLogger()
		con, _ := sql.NewDbConnection(cfg, logger)
		assert.Nil(t, err)
		configRepositoryImpl := NewCiBuildConfigRepositoryImpl(con, logger)
		countByBuildType, err := configRepositoryImpl.GetCountByBuildType()
		assert.Nil(t, err)
		for buildType, count := range countByBuildType {
			fmt.Println("type:", buildType, ", count:", count)
		}
	})
}
