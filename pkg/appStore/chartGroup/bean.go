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

package chartGroup

// / bean for v2
type ChartGroupInstallRequest struct {
	ProjectId                     int                              `json:"projectId"  validate:"required,number"`
	ChartGroupInstallChartRequest []*ChartGroupInstallChartRequest `json:"charts" validate:"dive,required"`
	ChartGroupId                  int                              `json:"chartGroupId"` //optional
	UserId                        int32                            `json:"-"`
}

type ChartGroupInstallChartRequest struct {
	AppName            string `json:"appName,omitempty"  validate:"name-component,max=100" `
	EnvironmentId      int    `json:"environmentId,omitempty" validate:"required,number" `
	AppStoreVersion    int    `json:"appStoreVersion,omitempty,notnull" validate:"required,number" `
	ValuesOverrideYaml string `json:"valuesOverrideYaml,omitempty"` //optional
	ReferenceValueId   int    `json:"referenceValueId, omitempty" validate:"required,number"`
	ReferenceValueKind string `json:"referenceValueKind, omitempty" validate:"oneof=DEFAULT TEMPLATE DEPLOYED"`
	ChartGroupEntryId  int    `json:"chartGroupEntryId"` //optional
}

type ChartGroupInstallAppRes struct {
}
