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

package models

type VariableRequest struct {
	Manifest ScopedVariableManifest `json:"manifest"`
	UserId   int32                  `json:"-"`
}
type ScopedVariableManifest struct {
	ApiVersion string         `json:"apiVersion" validate:"oneof=devtron.ai/v1beta1"`
	Kind       string         `json:"kind" validate:"oneof=Variable"`
	Spec       []VariableSpec `json:"spec" validate:"required,dive"`
}

type VariableSpec struct {
	Notes            string              `json:"notes"`
	ShortDescription string              `json:"shortDescription" validate:"max=120"`
	IsSensitive      bool                `json:"isSensitive"`
	Name             string              `json:"name" validate:"required"`
	Values           []VariableValueSpec `json:"values" validate:"dive"`
}

type VariableValueSpec struct {
	Category  AttributeType `json:"category" validate:"oneof=ApplicationEnv Application Env Cluster Global"`
	Value     interface{}   `json:"value" validate:"required"`
	Selectors *Selector     `json:"selectors,omitempty"`
}

type Selector struct {
	AttributeSelectors map[IdentifierType]string `json:"attributeSelectors"`
}
