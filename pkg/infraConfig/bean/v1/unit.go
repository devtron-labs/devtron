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

// Unit represents unitType of a configuration
type Unit struct {
	// Name is unitType name
	Name string `json:"name"`
	// ConversionFactor is used to convert this unitType to the base unitType
	// if ConversionFactor is 1, then this is the base unitType
	ConversionFactor float64 `json:"conversionFactor"`
}
