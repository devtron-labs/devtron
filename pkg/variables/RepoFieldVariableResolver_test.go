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

package variables

import (
	"errors"
	"testing"

	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/stretchr/testify/assert"
)

// fakeScopedVariableService resolves variables from a canned name->value map, and records
// the exact scope it was last called with so tests can assert env/cluster context is stripped.
type fakeScopedVariableService struct {
	ScopedVariableService
	definedVariables map[string]string
	err              error
	lastScope        resourceQualifiers.Scope
}

func (f *fakeScopedVariableService) GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, unmaskSensitiveData bool) ([]*models.ScopedVariableData, error) {
	f.lastScope = scope
	if f.err != nil {
		return nil, f.err
	}
	scopedVariableData := make([]*models.ScopedVariableData, 0)
	for _, name := range varNames {
		if value, ok := f.definedVariables[name]; ok {
			scopedVariableData = append(scopedVariableData, &models.ScopedVariableData{
				VariableName:  name,
				VariableValue: &models.VariableValue{Value: value},
			})
		}
	}
	return scopedVariableData, nil
}

func newResolverForTest(t *testing.T, svc ScopedVariableService) *RepoFieldVariableResolverImpl {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	return NewRepoFieldVariableResolverImpl(logger, svc)
}

func TestResolveRepoFieldValue(t *testing.T) {
	svc := &fakeScopedVariableService{
		definedVariables: map[string]string{
			"FULL_GIT_URL": "https://github.com/org/repo.git",
		},
	}
	resolver := newResolverForTest(t, svc)
	scope := resourceQualifiers.Scope{AppId: 1}

	t.Run("plain literal passes through unchanged", func(t *testing.T) {
		resolved, varName, err := resolver.ResolveRepoFieldValue(scope, "https://github.com/org/literal.git")
		assert.Nil(t, err)
		assert.Equal(t, "https://github.com/org/literal.git", resolved)
		assert.Empty(t, varName)
	})

	t.Run("full match resolves to defined value", func(t *testing.T) {
		resolved, varName, err := resolver.ResolveRepoFieldValue(scope, "@{{FULL_GIT_URL}}")
		assert.Nil(t, err)
		assert.Equal(t, "https://github.com/org/repo.git", resolved)
		assert.Equal(t, "FULL_GIT_URL", varName)
	})

	t.Run("undefined variable is rejected", func(t *testing.T) {
		_, _, err := resolver.ResolveRepoFieldValue(scope, "@{{NOT_DEFINED}}")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("partial match with surrounding text is rejected", func(t *testing.T) {
		_, _, err := resolver.ResolveRepoFieldValue(scope, "https://github.com/@{{FULL_GIT_URL}}/repo.git")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "single full variable reference")
	})

	t.Run("multiple tokens in one field is rejected", func(t *testing.T) {
		_, _, err := resolver.ResolveRepoFieldValue(scope, "@{{A}}@{{FULL_GIT_URL}}")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "single full variable reference")
	})

	t.Run("fixed system variable is rejected", func(t *testing.T) {
		_, _, err := resolver.ResolveRepoFieldValue(scope, "@{{DEVTRON_APP_NAME}}")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "system variable")
	})

	t.Run("underlying resolution error is propagated", func(t *testing.T) {
		failingSvc := &fakeScopedVariableService{err: errors.New("db unavailable")}
		failingResolver := newResolverForTest(t, failingSvc)
		_, _, err := failingResolver.ResolveRepoFieldValue(scope, "@{{FULL_GIT_URL}}")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "db unavailable")
	})
}

func TestIsFullVariableReference(t *testing.T) {
	assert.True(t, IsFullVariableReference("@{{FULL_GIT_URL}}"))
	assert.False(t, IsFullVariableReference("https://github.com/@{{GIT_ORG}}/repo.git"))
	assert.False(t, IsFullVariableReference("https://github.com/org/repo.git"))
	assert.False(t, IsFullVariableReference("@{{A}}@{{B}}"))
}

func TestExtractVariableName(t *testing.T) {
	name, ok := ExtractVariableName("@{{FULL_GIT_URL}}")
	assert.True(t, ok)
	assert.Equal(t, "FULL_GIT_URL", name)

	_, ok = ExtractVariableName("https://github.com/org/repo.git")
	assert.False(t, ok)

	_, ok = ExtractVariableName("https://github.com/@{{GIT_ORG}}/repo.git")
	assert.False(t, ok)
}
