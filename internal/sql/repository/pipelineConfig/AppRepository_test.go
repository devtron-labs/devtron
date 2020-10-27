/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pipelineConfig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getPipelineGroupRepo() *AppRepositoryImpl {
	return nil
	//return NewAppRepositoryImpl(models.NewDbConnection())
}

func TestPipelineGroupRepositoryImpl_FindActiveByName(t *testing.T) {
	pg, err := getPipelineGroupRepo().FindActiveByName("ke")
	assert.NoError(t, err)
	assert.NotNil(t, pg)
}

func TestPipelineGroupRepositoryImpl_ActivePipelineExists(t *testing.T) {
	exists, err := getPipelineGroupRepo().AppExists("ke")
	assert.NoError(t, err)
	assert.True(t, exists)
}
