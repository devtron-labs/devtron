package chartRepoRepository

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGlobalStrategyMetadataRepository(t *testing.T) {
	t.SkipNow()

	t.Run("GetByChartRefId - Rollout type chart with version < 3.7", func(t *testing.T) {
		cfg, _ := sql.GetConfig()
		logger, err := utils.NewSugardLogger()
		assert.Nil(t, err)
		dbConnection, err := sql.NewDbConnection(cfg, logger)
		assert.Nil(t, err)
		globalStrategyMetadataRepositoryImpl := NewGlobalStrategyMetadataRepositoryImpl(dbConnection, logger)
		globalStrategyMetadatas, err := globalStrategyMetadataRepositoryImpl.GetByChartRefId(1)
		assert.Nil(t, err)
		for _, globalStrategyMetadata := range globalStrategyMetadatas {
			assert.Contains(t, []DeploymentStrategy{DEPLOYMENT_STRATEGY_CANARY,
				DEPLOYMENT_STRATEGY_RECREATE,
				DEPLOYMENT_STRATEGY_ROLLING,
				DEPLOYMENT_STRATEGY_BLUE_GREEN}, globalStrategyMetadata.Name)
		}
	})

	t.Run("GetByChartRefId - Rollout type chart with version > 3.7", func(t *testing.T) {
		cfg, _ := sql.GetConfig()
		logger, err := utils.NewSugardLogger()
		assert.Nil(t, err)
		dbConnection, err := sql.NewDbConnection(cfg, logger)
		assert.Nil(t, err)
		globalStrategyMetadataRepositoryImpl := NewGlobalStrategyMetadataRepositoryImpl(dbConnection, logger)
		globalStrategyMetadatas, err := globalStrategyMetadataRepositoryImpl.GetByChartRefId(1)
		assert.Nil(t, err)
		for _, globalStrategyMetadata := range globalStrategyMetadatas {
			assert.Contains(t, []DeploymentStrategy{DEPLOYMENT_STRATEGY_ROLLING,
				DEPLOYMENT_STRATEGY_BLUE_GREEN}, globalStrategyMetadata.Name)
		}
	})

	t.Run("GetByChartRefId - Deployment type chart", func(t *testing.T) {
		cfg, _ := sql.GetConfig()
		logger, err := utils.NewSugardLogger()
		assert.Nil(t, err)
		dbConnection, err := sql.NewDbConnection(cfg, logger)
		assert.Nil(t, err)
		globalStrategyMetadataRepositoryImpl := NewGlobalStrategyMetadataRepositoryImpl(dbConnection, logger)
		globalStrategyMetadatas, err := globalStrategyMetadataRepositoryImpl.GetByChartRefId(1)
		assert.Nil(t, err)
		for _, globalStrategyMetadata := range globalStrategyMetadatas {
			assert.Contains(t, []DeploymentStrategy{DEPLOYMENT_STRATEGY_RECREATE,
				DEPLOYMENT_STRATEGY_ROLLING}, globalStrategyMetadata.Name)
		}
	})
}
