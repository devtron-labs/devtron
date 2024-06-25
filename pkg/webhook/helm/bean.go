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

package webhookHelm

type HelmAppCreateUpdateRequest struct {
	ClusterName        string     `json:"clusterName,notnull" validate:"required"`
	Namespace          string     `json:"namespace,omitempty"`
	ReleaseName        string     `json:"releaseName,notnull" validate:"required"`
	ValuesOverrideYaml string     `json:"valuesOverrideYaml,omitempty"`
	Chart              *ChartSpec `json:"chart,notnull" validate:"required"`
}

type ChartSpec struct {
	Repo         *ChartRepoSpec `json:"repo,notnull" validate:"required"`
	ChartName    string         `json:"chartName,notnull" validate:"required"`
	ChartVersion string         `json:"chartVersion,omitempty"`
}

type ChartRepoSpec struct {
	Name       string                   `json:"name,notnull" validate:"required"`
	Identifier *ChartRepoIdentifierSpec `json:"identifier,omitempty"`
}

type ChartRepoIdentifierSpec struct {
	Url      string `json:"url,notnull" validate:"required"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
