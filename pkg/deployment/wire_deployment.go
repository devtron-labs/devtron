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

package deployment

import (
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	"github.com/devtron-labs/devtron/pkg/deployment/providerConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger"
	"github.com/google/wire"
)

// TODO: add separate wire sets for full and ea mode when reached that level of transparency

var DeploymentWireSet = wire.NewSet(
	gitOps.GitOpsWireSet,
	manifest.DeploymentManifestWireSet,
	trigger.DeploymentTriggerWireSet,
	deployedApp.DeployedAppWireSet,
	providerConfig.DeploymentProviderConfigWireSet,
)
