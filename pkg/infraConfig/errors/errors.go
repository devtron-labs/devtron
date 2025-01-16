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

package errors

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
)

const CPULimReqErrorCompErr = "cpu limit should not be less than cpu request"

const MEMLimReqErrorCompErr = "memory limit should not be less than memory request"

var NoPropertiesFoundError = errors.New("no properties found")

var ProfileIdsRequired = errors.New("profile ids cannot be empty")

const (
	PayloadValidationError               = "payload validation failed"
	InvalidProfileName                   = "profile name is invalid"
	ProfileDoNotExists                   = "profile does not exist"
	InvalidProfileNameChangeRequested    = "invalid profile name change requested"
	ProfileAlreadyExistsErr              = "profile already exists"
	DeletionBlockedForDefaultPlatform    = "cannot delete default platform configuration"
	ConfigurationMissingInGlobalPlatform = "configuration missing in the global Platform"
	UpdatableConfigurationFoundErr       = "updatable configuration not belongs to platform"
)

func InvalidUnitFound(unit string, key v1.ConfigKeyStr) string {
	return fmt.Sprintf("invalid %q unit found for %q", unit, key)
}

func ConfigurationMissingError(missingKey v1.ConfigKeyStr, profileName, platformName string) string {
	return fmt.Sprintf("%q configuration missing in the %q profile %q platform", missingKey, profileName, platformName)
}
