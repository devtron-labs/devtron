/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package bean

import (
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
)

type RegisterScanToolsDto struct {
	ScanToolMetadata       *ScanToolsMetadataDto      `json:"scanToolMetadata" validate:"dive,required"`
	ScanToolPluginMetadata *ScanToolPluginMetadataDto `json:"scanToolPluginMetadata,omitempty" validate:"dive,required"`
}

type ScanToolsMetadataDto struct {
	ScanToolId               int            `json:"scanToolId,omitempty"`
	Name                     string         `json:"name,omitempty" validate:"required"`
	Version                  string         `json:"version,omitempty" validate:"required"`
	ServerBaseUrl            string         `json:"serverBaseUrl,omitempty"`
	ResultDescriptorTemplate string         `json:"resultDescriptorTemplate,omitempty"`
	ScanTarget               ScanTargetType `json:"scanTarget,omitempty"`
	ToolMetaData             string         `json:"toolMetadata,omitempty"`
	ScanToolUrl              string         `json:"scanToolUrl,omitempty"`
}

type ScanToolPluginMetadataDto struct {
	Name             string                  `json:"name" validate:"required,min=3,max=100"`
	PluginIdentifier string                  `json:"pluginIdentifier" validate:"required,min=3,max=100,global-entity-name"`
	Description      string                  `json:"description" validate:"max=300"`
	PluginSteps      []*bean2.PluginStepsDto `json:"pluginSteps,omitempty"`
}

const (
	DevtronImageScanningIntegratorPluginIdentifier = "devtron-image-scanning-integrator"
)

type ScanTargetType string

const (
	ScanTargetTypeImage ScanTargetType = "IMAGE"
)
