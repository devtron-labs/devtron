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

// Package v1 implements the infra config with interface values.
package v1

type PlatformResponse struct {
	Platforms []string `json:"platforms"`
}

// internal-platforms
const (
	// RUNNER_PLATFORM is the name of the default platform; a reserved name
	RUNNER_PLATFORM = "runner"

	// Deprecated: use RUNNER_PLATFORM instead
	// CI_RUNNER_PLATFORM is earlier used as the name of the default platform
	CI_RUNNER_PLATFORM = "ci-runner"
)
