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

package chartRepoRepository

import (
	"github.com/devtron-labs/common-lib-private/utils"
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
