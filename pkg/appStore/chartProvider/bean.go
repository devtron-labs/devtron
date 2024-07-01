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

package chartProvider

import repository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"

type ChartProviderResponseDto struct {
	Id               string                  `json:"id" validate:"required"`
	Name             string                  `json:"name" validate:"required"`
	Active           bool                    `json:"active" validate:"required"`
	IsEditable       bool                    `json:"isEditable"`
	IsOCIRegistry    bool                    `json:"isOCIRegistry"`
	RegistryProvider repository.RegistryType `json:"registryProvider"`
	UserId           int32                   `json:"-"`
}

type ChartProviderRequestDto struct {
	Id            string `json:"id" validate:"required"`
	IsOCIRegistry bool   `json:"isOCIRegistry"`
	Active        bool   `json:"active,omitempty"`
	UserId        int32  `json:"-"`
	ChartRepoId   int    `json:"-"`
}
