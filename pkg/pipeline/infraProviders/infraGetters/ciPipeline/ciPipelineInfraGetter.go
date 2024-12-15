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

package ciPipeline

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service"
)

// CiInfraGetter gets infra config for ci workflows
type CiInfraGetter struct {
	infraConfigService service.InfraConfigService
}

func NewCiInfraGetter(infraConfigService service.InfraConfigService) *CiInfraGetter {
	return &CiInfraGetter{infraConfigService: infraConfigService}
}

// GetInfraConfigurationsByScope gets infra config for ci workflows using the scope
func (ciInfraGetter CiInfraGetter) GetInfraConfigurationsByScopeAndPlatform(scope *bean.Scope, platform string) (*bean.InfraConfig, error) {
	return ciInfraGetter.infraConfigService.GetInfraConfigurationsByScopeAndPlatform(*scope, platform)
}
