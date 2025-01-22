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

// Package v1 implements the infra config with interface values.
package v1

// BuildxK8sDriverMigrated is the marker flag to check if the infra config is migrated completely.
const BuildxK8sDriverMigrated string = "build-infra-migrated"

// Scope refers to the identifier scope for the infra config
//   - Currently, the infra config profile is applied to a specific app only.
type Scope struct {
	AppId int
}

type ConfigKeyPlatformKey struct {
	Key      ConfigKey
	Platform string
}
