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

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
)

type EnvironmentProperties struct {
	Id                  int                         `json:"id"`
	EnvOverrideValues   json.RawMessage             `json:"envOverrideValues"`
	Status              models.ChartStatus          `json:"status" validate:"number,required"` //default new, when its ready for deployment CHARTSTATUS_SUCCESS
	ManualReviewed      bool                        `json:"manualReviewed" validate:"required"`
	Active              bool                        `json:"active" validate:"required"`
	Namespace           string                      `json:"namespace"`
	EnvironmentId       int                         `json:"environmentId"`
	EnvironmentName     string                      `json:"environmentName"`
	Latest              bool                        `json:"latest"`
	UserId              int32                       `json:"-"`
	AppMetrics          *bool                       `json:"isAppMetricsEnabled"`
	ChartRefId          int                         `json:"chartRefId,omitempty"  validate:"number"`
	IsOverride          bool                        `sql:"isOverride"`
	IsBasicViewLocked   bool                        `json:"isBasicViewLocked"`
	CurrentViewEditor   models.ChartsViewEditorType `json:"currentViewEditor"` //default "UNDEFINED" in db
	Description         string                      `json:"description" validate:"max=40"`
	ClusterId           int                         `json:"clusterId"`
	SaveEligibleChanges bool                        `json:"saveEligibleChanges"`
}

type EnvironmentPropertiesResponse struct {
	EnvironmentConfig EnvironmentProperties `json:"environmentConfig"`
	GlobalConfig      json.RawMessage       `json:"globalConfig"`
	AppMetrics        *bool                 `json:"appMetrics"`
	IsOverride        bool                  `sql:"is_override"`
	GlobalChartRefId  int                   `json:"globalChartRefId,omitempty"  validate:"number"`
	ChartRefId        int                   `json:"chartRefId,omitempty"  validate:"number"`
	Namespace         string                `json:"namespace" validate:"name-space-component"`
	Schema            json.RawMessage       `json:"schema"`
	Readme            string                `json:"readme"`
}

type EnvironmentUpdateResponse struct {
	*EnvironmentProperties
	*bean.LockValidateErrorResponse
}
