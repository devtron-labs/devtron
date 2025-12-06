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

// Package v0 implements the infra config with float64 values only.
//
// Deprecated: v0 is functionally broken and should not be used
// except for compatibility with legacy systems. Use v1 instead.
//
// This package is frozen and no new functionality will be added.
package v0

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
)

// Deprecated: ConfigurationBeanV0 is deprecated in favor of v1.ConfigurationBean
type ConfigurationBeanV0 struct {
	v1.ConfigurationBeanAbstract
	Value float64 `json:"value" validate:"required,gt=0"`
}
