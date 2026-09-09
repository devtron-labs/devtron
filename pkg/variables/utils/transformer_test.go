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

package utils

import (
	"testing"

	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/stretchr/testify/assert"
)

// OSS supports only the Global category (no attribute selectors), so grouped selector keys
// (EnvNames, ...) can never trigger expansion here. These tests pin that dormancy and the
// unchanged round-trip behaviour; the expansion logic itself is exercised in enterprise.

func TestGetGroupedIdentifierForCategory_AlwaysFalseInOSS(t *testing.T) {
	for _, category := range []models.AttributeType{models.Global, "Env", "Application", "Cluster"} {
		_, _, ok := GetGroupedIdentifierForCategory(category)
		assert.False(t, ok, string(category))
	}
}

func TestManifestPayloadRoundTrip_GlobalOnly(t *testing.T) {
	original := models.ScopedVariableManifest{
		ApiVersion: "devtron.ai/v1beta1",
		Kind:       "Variable",
		Spec: []models.VariableSpec{
			{
				Name:             "VAR1",
				Notes:            "notes",
				ShortDescription: "short",
				Values: []models.VariableValueSpec{
					{Category: models.Global, Value: "value1"},
				},
			},
		},
	}
	payload, err := ManifestToPayload(original, 3)
	assert.Nil(t, err)
	assert.Equal(t, int32(3), payload.UserId)
	assert.Equal(t, original.Spec, PayloadToManifest(payload).Spec)
}

func TestManifestToPayload_SelectorsPassThroughUnchangedInOSS(t *testing.T) {
	manifest := models.ScopedVariableManifest{
		ApiVersion: "devtron.ai/v1beta1",
		Kind:       "Variable",
		Spec: []models.VariableSpec{
			{
				Name: "VAR1",
				Values: []models.VariableValueSpec{
					{
						Category: models.Global,
						Value:    "v",
						Selectors: &models.Selector{AttributeSelectors: map[models.IdentifierType]string{
							"EnvNames": "env1, env2",
						}},
					},
				},
			},
		},
	}
	payload, err := ManifestToPayload(manifest, 1)

	assert.Nil(t, err)
	attributes := payload.Variables[0].AttributeValues
	assert.Len(t, attributes, 1)
	// no expansion in OSS; the invalid key is left for payload validation to reject
	assert.Equal(t, "env1, env2", attributes[0].AttributeParams["EnvNames"])
}
