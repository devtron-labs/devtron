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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/devtron-labs/devtron/pkg/variables/helper"
	"github.com/devtron-labs/devtron/pkg/variables/models"
)

const groupedNameSeparator = ","
const groupedNameJoiner = ", "

// GetGroupedIdentifierForCategory returns the singular selector key of a category together with its
// grouped (plural) form, e.g. Env -> (EnvName, EnvNames). Grouped selectors are supported only for
// categories with exactly one selector key. In OSS only the Global category (no selectors) exists,
// so this always returns ok=false here; the logic is kept in sync with enterprise.
func GetGroupedIdentifierForCategory(category models.AttributeType) (singular models.IdentifierType, grouped models.IdentifierType, ok bool) {
	identifierTypes := helper.GetIdentifierTypeFromAttributeType(category)
	if len(identifierTypes) != 1 {
		return "", "", false
	}
	singular = identifierTypes[0]
	return singular, models.IdentifierType(string(singular) + "s"), true
}

// ExpandAttributeSelectors turns one selector map that may carry a grouped key
// (e.g. EnvNames: "env1, env2") into one selector map per name carrying the singular key
// (EnvName: "env1"). A map without a grouped key is returned unchanged.
func ExpandAttributeSelectors(category models.AttributeType, attributeSelectors map[models.IdentifierType]string) ([]map[models.IdentifierType]string, error) {
	singular, grouped, ok := GetGroupedIdentifierForCategory(category)
	if !ok {
		return []map[models.IdentifierType]string{attributeSelectors}, nil
	}
	groupedValue, hasGrouped := attributeSelectors[grouped]
	if !hasGrouped {
		return []map[models.IdentifierType]string{attributeSelectors}, nil
	}
	if _, hasSingular := attributeSelectors[singular]; hasSingular {
		return nil, fmt.Errorf("selector for category %s cannot use both %s and %s", category, singular, grouped)
	}
	names := make([]string, 0)
	seen := make(map[string]bool)
	for _, name := range strings.Split(groupedValue, groupedNameSeparator) {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("%s for category %s contains an empty name", grouped, category)
		}
		if seen[name] {
			return nil, fmt.Errorf("%s for category %s contains duplicate name %q", grouped, category, name)
		}
		seen[name] = true
		names = append(names, name)
	}
	expanded := make([]map[models.IdentifierType]string, 0, len(names))
	for _, name := range names {
		selectors := make(map[models.IdentifierType]string, len(attributeSelectors))
		for key, value := range attributeSelectors {
			if key != grouped {
				selectors[key] = value
			}
		}
		selectors[singular] = name
		expanded = append(expanded, selectors)
	}
	return expanded, nil
}

func ManifestToPayload(manifest models.ScopedVariableManifest, userId int32) (models.Payload, error) {

	variableList := make([]*models.Variables, 0)

	for _, spec := range manifest.Spec {
		attributes := make([]models.AttributeValue, 0)
		for _, value := range spec.Values {
			selectorsList := []map[models.IdentifierType]string{nil}
			if value.Selectors != nil && value.Selectors.AttributeSelectors != nil {
				expandedSelectors, err := ExpandAttributeSelectors(value.Category, value.Selectors.AttributeSelectors)
				if err != nil {
					return models.Payload{}, models.ValidationError{Err: err}
				}
				selectorsList = expandedSelectors
			}
			for _, selectors := range selectorsList {
				attribute := models.AttributeValue{
					VariableValue: models.VariableValue{Value: value.Value},
					AttributeType: value.Category,
				}
				if selectors != nil {
					attribute.AttributeParams = selectors
				}
				attributes = append(attributes, attribute)
			}
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
	return payload, nil
}

// groupingValueKey renders a value into a stable, type-aware key so that only entries carrying the
// same value are grouped (json.Marshal keeps 5 and "5" distinct and sorts map keys).
func groupingValueKey(value interface{}) string {
	marshaled, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(marshaled)
}

// isGroupableAttribute reports whether an attribute entry may be merged with others of the same
// category and value: single-selector category, exactly the singular key present, and a name that
// itself contains no separator.
func isGroupableAttribute(attribute models.AttributeValue) (name string, ok bool) {
	singular, _, ok := GetGroupedIdentifierForCategory(attribute.AttributeType)
	if !ok || len(attribute.AttributeParams) != 1 {
		return "", false
	}
	name, hasSingular := attribute.AttributeParams[singular]
	if !hasSingular || strings.Contains(name, groupedNameSeparator) {
		return "", false
	}
	return name, true
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
		// entries of the same category carrying the same value are folded into one grouped entry
		groupKeyToValueIndex := make(map[string]int)
		groupNames := make(map[int][]string)
		for _, attribute := range variable.AttributeValues {
			name, groupable := isGroupableAttribute(attribute)
			if groupable {
				groupKey := string(attribute.AttributeType) + "\x00" + groupingValueKey(attribute.VariableValue.Value)
				if valueIndex, found := groupKeyToValueIndex[groupKey]; found {
					groupNames[valueIndex] = append(groupNames[valueIndex], name)
					continue
				}
				groupKeyToValueIndex[groupKey] = len(spec.Values)
				groupNames[len(spec.Values)] = []string{name}
			}
			valueSpec := models.VariableValueSpec{
				Value:    attribute.VariableValue.Value,
				Category: attribute.AttributeType,
			}
			if attribute.AttributeParams != nil {
				valueSpec.Selectors = &models.Selector{AttributeSelectors: attribute.AttributeParams}
			}
			spec.Values = append(spec.Values, valueSpec)
		}
		for valueIndex, names := range groupNames {
			if len(names) < 2 {
				continue
			}
			_, grouped, ok := GetGroupedIdentifierForCategory(spec.Values[valueIndex].Category)
			if !ok {
				continue
			}
			sort.Strings(names)
			spec.Values[valueIndex].Selectors = &models.Selector{AttributeSelectors: map[models.IdentifierType]string{
				grouped: strings.Join(names, groupedNameJoiner),
			}}
		}
		manifest.Spec = append(manifest.Spec, spec)
	}
	return manifest
}
