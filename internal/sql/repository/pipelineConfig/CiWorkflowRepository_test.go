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

package pipelineConfig

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCiWorkflowRepository(t *testing.T) {
	t.SkipNow()
	t.Run("FindBuildTypeAndStatusDataOfLast1Day", func(t *testing.T) {
		cfg, _ := sql.GetConfig()
		logger, err := utils.NewSugardLogger()
		con, _ := sql.NewDbConnection(cfg, logger)
		assert.Nil(t, err)
		workflowRepositoryImpl := NewCiWorkflowRepositoryImpl(con, logger)
		statusData := workflowRepositoryImpl.FindBuildTypeAndStatusDataOfLast1Day()
		for _, statusDatum := range statusData {
			fmt.Println(statusDatum.Count)
		}
	})
}
