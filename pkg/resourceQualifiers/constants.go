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

package resourceQualifiers

type SystemVariableName string

const (
	DevtronNamespace   SystemVariableName = "DEVTRON_NAMESPACE"
	DevtronClusterName SystemVariableName = "DEVTRON_CLUSTER_NAME"
	DevtronEnvName     SystemVariableName = "DEVTRON_ENV_NAME"
	DevtronImageTag    SystemVariableName = "DEVTRON_IMAGE_TAG"
	DevtronImage       SystemVariableName = "DEVTRON_IMAGE"
	DevtronAppName     SystemVariableName = "DEVTRON_APP_NAME"
)

var SystemVariables = []SystemVariableName{
	DevtronNamespace,
	DevtronClusterName,
	DevtronEnvName,
	DevtronImageTag,
	DevtronAppName,
	DevtronImage,
}
