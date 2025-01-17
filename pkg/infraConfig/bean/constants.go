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

package bean

type ConfigKey int
type ConfigKeyStr string
type ProfileType string

const NORMAL ProfileType = "NORMAL"
const InvalidUnit = "invalid %s unit found in %s "
const InvalidTypeValue = "invalid value found in %s with value %s "
const GLOBAL_PROFILE_NAME = "global"

// TODO Asutosh: Backward compatibility for default profile is compromised. revisit this.
const DEFAULT_PROFILE_NAME = "default"
const DEFAULT_PROFILE_EXISTS = "default profile exists"
const NO_PROPERTIES_FOUND = "no properties found"
const DEFAULT ProfileType = "DEFAULT"
const GLOBAL ProfileType = "GLOBAL"
const InvalidProfileName = "profile name is invalid"
const PayloadValidationError = "payload validation failed"
const CPULimReqErrorCompErr = "cpu limit should not be less than cpu request"
const MEMLimReqErrorCompErr = "memory limit should not be less than memory request"
const InvalidValueType = "invalid Value type Found"

const CPULimitKey ConfigKey = 1
const CPURequestKey ConfigKey = 2
const MemoryLimitKey ConfigKey = 3
const MemoryRequestKey ConfigKey = 4
const TimeOutKey ConfigKey = 5

// whenever new constant gets added here ,
// we need to add it in GetDefaultConfigKeysMap method as well

const CPU_LIMIT ConfigKeyStr = "cpu_limit"
const CPU_REQUEST ConfigKeyStr = "cpu_request"
const MEMORY_LIMIT ConfigKeyStr = "memory_limit"
const MEMORY_REQUEST ConfigKeyStr = "memory_request"
const TIME_OUT ConfigKeyStr = "timeout"

// internal-platforms
const RUNNER_PLATFORM = "runner"
const QualifiedProfileMaxLength = 253
const QualifiedDescriptionMaxLength = 350
const QualifiedPlatformMaxLength = 50
const ConfigurationMissingInGlobalPlatform = "configuration missing in the global Platform"
