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

package adaptor

import (
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/bean"
	"time"
)

func GetPluginMetadataAndStepsDetail(scanToolPluginMetadataDto *bean.ScanToolPluginMetadataDto, scanToolUrl string, version string) *bean2.PluginParentMetadataDto {
	pluginParentObj := &bean2.PluginParentMetadataDto{
		Name:             scanToolPluginMetadataDto.Name,
		PluginIdentifier: scanToolPluginMetadataDto.PluginIdentifier,
		Description:      scanToolPluginMetadataDto.Description,
		Type:             bean2.SHARED.ToString(),
		Icon:             scanToolUrl,
	}
	pluginMetadataDto := &bean2.PluginMetadataDto{
		Tags:        []string{"Security"},
		PluginStage: "SCANNER",
		PluginSteps: scanToolPluginMetadataDto.PluginSteps,
	}
	pluginVersionDetail := &bean2.PluginsVersionDetail{
		PluginMetadataDto: pluginMetadataDto,
		Version:           version,
		IsLatest:          true,
		CreatedOn:         time.Now(),
	}
	pluginParentObj.Versions = &bean2.PluginVersions{DetailedPluginVersionData: []*bean2.PluginsVersionDetail{pluginVersionDetail}}
	return pluginParentObj
}
