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

package commonService

type GlobalChecklist struct {
	AppChecklist   *AppChecklist   `json:"appChecklist"`
	ChartChecklist *ChartChecklist `json:"chartChecklist"`
	IsAppCreated   bool            `json:"isAppCreated"`
	UserId         int32           `json:"-"`
}

type ChartChecklist struct {
	GitOps      int `json:"gitOps,omitempty"`
	Project     int `json:"project"`
	Environment int `json:"environment"`
}

type FeatureGitOpsVariables struct {
	IsFeatureGitOpsEnabled            bool `json:"isFeatureGitOpsEnabled"`
	IsFeatureUserDefinedGitOpsEnabled bool `json:"isFeatureUserDefinedGitOpsEnabled"`
	IsFeatureArgoCdMigrationEnabled   bool `json:"isFeatureArgoCdMigrationEnabled"`
}

type EnvironmentVariableList struct {
	FeatureGitOpsFlags *FeatureGitOpsVariables `json:"featureGitOpsFlags"`
	EnvironmentVariableListEnt
}

type AppChecklist struct {
	GitOps      int `json:"gitOps,omitempty"`
	Project     int `json:"project"`
	Git         int `json:"git"`
	Environment int `json:"environment"`
	Docker      int `json:"docker"`
	HostUrl     int `json:"hostUrl"`
	//ChartChecklist *ChartChecklist `json:",inline"`
}
