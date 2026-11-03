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
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/variables"
	variableEntity "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopVariableEntityMappingService struct {
	variables.VariableEntityMappingService
}

func (noopVariableEntityMappingService) UpdateVariablesForEntity(variableNames []string, entity variableEntity.Entity, userId int32, tx *pg.Tx) error {
	return nil
}

// captures every UpdateVariablesForEntity call, so a test can assert what Save/Update tracked
type capturingVariableEntityMappingService struct {
	variables.VariableEntityMappingService
	calls []trackedVariableMapping
}

type trackedVariableMapping struct {
	variableNames []string
	entity        variableEntity.Entity
}

func (c *capturingVariableEntityMappingService) UpdateVariablesForEntity(variableNames []string, entity variableEntity.Entity, userId int32, tx *pg.Tx) error {
	c.calls = append(c.calls, trackedVariableMapping{variableNames: variableNames, entity: entity})
	return nil
}

type noopCiTemplateRepository struct {
	pipelineConfig.CiTemplateRepository
}

func (noopCiTemplateRepository) Save(*pipelineConfig.CiTemplate) error   { return nil }
func (noopCiTemplateRepository) Update(*pipelineConfig.CiTemplate) error { return nil }

type noopCiTemplateOverrideRepository struct {
	pipelineConfig.CiTemplateOverrideRepository
}

func (noopCiTemplateOverrideRepository) Save(o *pipelineConfig.CiTemplateOverride) (*pipelineConfig.CiTemplateOverride, error) {
	return o, nil
}
func (noopCiTemplateOverrideRepository) Update(o *pipelineConfig.CiTemplateOverride) (*pipelineConfig.CiTemplateOverride, error) {
	return o, nil
}

// captures what would be written to ci_build_config
type capturingCiBuildConfigRepository struct {
	pipelineConfig.CiBuildConfigRepository
	saved   []*pipelineConfig.CiBuildConfig
	updated []*pipelineConfig.CiBuildConfig
}

func (r *capturingCiBuildConfigRepository) Save(c *pipelineConfig.CiBuildConfig) error {
	c.Id = 100 + len(r.saved)
	r.saved = append(r.saved, c)
	return nil
}

func (r *capturingCiBuildConfigRepository) Update(c *pipelineConfig.CiBuildConfig) error {
	r.updated = append(r.updated, c)
	return nil
}

// Update must track (or clear) the scoped variable a Container Repository field references,
// so deletion protection sees it. Covers: a reference, a changed reference, and a literal
// value (no variable, mapping cleared).
func TestCiTemplateService_TracksContainerRepositoryVariableMapping(t *testing.T) {
	logger, err := util.NewSugardLogger()
	require.NoError(t, err)
	mapping := &capturingVariableEntityMappingService{}
	svc := NewCiTemplateServiceImpl(logger, NewCiBuildConfigServiceImpl(logger, &capturingCiBuildConfigRepository{}),
		noopCiTemplateRepository{}, noopCiTemplateOverrideRepository{}, mapping)

	update := func(dockerRepository string) {
		tb := &bean2.CiTemplateBean{
			CiTemplate:    &pipelineConfig.CiTemplate{Id: 42, DockerRepository: dockerRepository},
			CiBuildConfig: &bean2.CiBuildConfigBean{},
			UserId:        1,
		}
		require.NoError(t, svc.Update(tb))
	}

	update("@{{A}}")
	require.Len(t, mapping.calls, 1)
	assert.Equal(t, []string{"A"}, mapping.calls[0].variableNames)
	assert.Equal(t, variableEntity.EntityTypeContainerRepositoryAppLevel, mapping.calls[0].entity.EntityType)
	assert.Equal(t, 42, mapping.calls[0].entity.EntityId)

	update("@{{B}}")
	require.Len(t, mapping.calls, 2)
	assert.Equal(t, []string{"B"}, mapping.calls[1].variableNames)

	update("my-repo")
	require.Len(t, mapping.calls, 3)
	assert.Empty(t, mapping.calls[2].variableNames)
}
