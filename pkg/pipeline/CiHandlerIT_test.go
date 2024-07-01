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

package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCiHandlerImpl_FetchArtifactsForCiJob(t *testing.T) {
	t.SkipNow()
	ciHandler := initCiHandler()

	t.Run("Fetch Ci Artifacts For Ci Job type", func(tt *testing.T) {
		buildId := 304 // Mocked because child workflows are only created dynamic based on number of images which are available after polling
		time.Sleep(5 * time.Second)
		_, err := ciHandler.FetchArtifactsForCiJob(buildId)
		assert.Nil(t, err)

	})
}

func initCiHandler() *CiHandlerImpl {
	sugaredLogger, _ := util.InitLogger()
	config, _ := sql.GetConfig()
	db, _ := sql.NewDbConnection(config, sugaredLogger)
	ciArtifactRepositoryImpl := repository.NewCiArtifactRepositoryImpl(db, sugaredLogger)
	ciHandlerImpl := NewCiHandlerImpl(sugaredLogger, nil, nil, nil, nil, nil, nil, ciArtifactRepositoryImpl, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return ciHandlerImpl
}
