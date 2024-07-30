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

package chart

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/models"
)

var ReservedChartRefNamesList *[]ReservedChartList

type ReservedChartList struct {
	LocationPrefix string
	Name           string
}

type TemplateRequest struct {
	Id                      int                         `json:"id"  validate:"number"`
	AppId                   int                         `json:"appId,omitempty"  validate:"number,required"`
	RefChartTemplate        string                      `json:"refChartTemplate,omitempty"`
	RefChartTemplateVersion string                      `json:"refChartTemplateVersion,omitempty"`
	ChartRepositoryId       int                         `json:"chartRepositoryId,omitempty"`
	ValuesOverride          json.RawMessage             `json:"valuesOverride,omitempty" validate:"required"` //json format user value
	DefaultAppOverride      json.RawMessage             `json:"defaultAppOverride,omitempty"`                 //override values available
	ChartRefId              int                         `json:"chartRefId,omitempty"  validate:"number"`
	Latest                  bool                        `json:"latest"`
	IsAppMetricsEnabled     bool                        `json:"isAppMetricsEnabled"`
	Schema                  json.RawMessage             `json:"schema"`
	Readme                  string                      `json:"readme"`
	IsBasicViewLocked       bool                        `json:"isBasicViewLocked"`
	CurrentViewEditor       models.ChartsViewEditorType `json:"currentViewEditor"` //default "UNDEFINED" in db
	GitRepoUrl              string                      `json:"-"`
	IsCustomGitRepository   bool                        `json:"-"`
	UserId                  int32                       `json:"-"`
	LatestChartVersion      string                      `json:"-"`
	ImageDescriptorTemplate string                      `json:"-"`
}

type ChartUpgradeRequest struct {
	ChartRefId int   `json:"chartRefId"  validate:"number"`
	All        bool  `json:"all"`
	AppIds     []int `json:"appIds"`
	UserId     int32 `json:"-"`
}

type ChartRefChangeRequest struct {
	AppId            int `json:"appId" validate:"required"`
	EnvId            int `json:"envId" validate:"required"`
	TargetChartRefId int `json:"targetChartRefId" validate:"required"`
}

type PipelineConfigRequest struct {
	Id                   int             `json:"id"  validate:"number"`
	AppId                int             `json:"appId,omitempty"  validate:"number,required"`
	EnvConfigOverrideId  int             `json:"envConfigOverrideId,omitempty"`
	PipelineConfigValues json.RawMessage `json:"pipelineConfigValues,omitempty" validate:"required"` //json format user value
	PipelineId           int             `json:"PipelineId,omitempty"`
	Latest               bool            `json:"latest"`
	Previous             bool            `json:"previous"`
	EnvId                int             `json:"envId,omitempty"`
	ManualReviewed       bool            `json:"manualReviewed" validate:"required"`
	UserId               int32           `json:"-"`
}

type PipelineConfigRequestResponse struct {
	LatestPipelineConfigRequest   PipelineConfigRequest `json:"latestPipelineConfigRequest"`
	PreviousPipelineConfigRequest PipelineConfigRequest `json:"previousPipelineConfigRequest"`
}

type AppConfigResponse struct {
	//DefaultAppConfig  json.RawMessage `json:"defaultAppConfig"`
	//AppConfig         TemplateRequest            `json:"appConfig"`
	LatestAppConfig   TemplateRequest `json:"latestAppConfig"`
	PreviousAppConfig TemplateRequest `json:"previousAppConfig"`
}

type DefaultChart string
