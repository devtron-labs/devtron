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
	"github.com/devtron-labs/devtron/pkg/variables/models"
)

func ManifestToPayload(manifest models.ScopedVariableManifest, userId int32) models.Payload {

	variableList := make([]*models.Variables, 0)

	for _, spec := range manifest.Spec {
		attributes := make([]models.AttributeValue, 0)
		for _, value := range spec.Values {
			attribute := models.AttributeValue{
				VariableValue: models.VariableValue{Value: value.Value},
				AttributeType: models.Global,
			}

			if value.Selectors != nil && value.Selectors.AttributeSelectors != nil {
				attribute.AttributeParams = value.Selectors.AttributeSelectors
			}
			attributes = append(attributes, attribute)
		}
		variable := models.Variables{
			Definition: models.Definition{
				VarName:          spec.Name,
				DataType:         models.PRIMITIVE_TYPE,
				VarType:          models.PUBLIC,
				Description:      spec.Notes,
				ShortDescription: spec.ShortDescription,
			},
			AttributeValues: attributes,
		}
		if spec.IsSensitive {
			variable.Definition.VarType = models.PRIVATE
		}
		variableList = append(variableList, &variable)
	}
	payload := models.Payload{
		Variables: variableList,
		UserId:    userId,
	}
	return payload
}

func PayloadToManifest(payload models.Payload) models.ScopedVariableManifest {
	manifest := models.ScopedVariableManifest{
		ApiVersion: "devtron.ai/v1beta1",
		Kind:       "Variable",
		Spec:       make([]models.VariableSpec, 0),
	}
	for _, variable := range payload.Variables {
		spec := models.VariableSpec{
			Name:             variable.Definition.VarName,
			Notes:            variable.Definition.Description,
			ShortDescription: variable.Definition.ShortDescription,
			Values:           make([]models.VariableValueSpec, 0),
			IsSensitive:      variable.Definition.VarType.IsTypeSensitive(),
		}
		for _, attribute := range variable.AttributeValues {
			valueSpec := models.VariableValueSpec{
				Value:    attribute.VariableValue.Value,
				Category: attribute.AttributeType,
			}
			if attribute.AttributeParams != nil {
				valueSpec.Selectors = &models.Selector{AttributeSelectors: attribute.AttributeParams}
			}
			spec.Values = append(spec.Values, valueSpec)
		}
		manifest.Spec = append(manifest.Spec, spec)
	}
	return manifest
}
