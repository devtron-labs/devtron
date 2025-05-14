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

package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"time"
)

type Charts struct {
	Id                      int
	AppId                   int
	ChartRepoId             int
	ChartName               string
	ChartVersion            string
	ChartRepo               string
	ChartRepoUrl            string
	Values                  string             //json format // used at for release. this should be always updated
	GlobalOverride          string             //json format    // global overrides visible to user only
	ReleaseOverride         string             //json format   //image descriptor template used for injecting tigger metadata injection
	PipelineOverride        string             //json format  // pipeline values -> strategy values
	Status                  models.ChartStatus //(new , deployment-in-progress, deployed-To-production, error )
	Active                  bool
	GitRepoUrl              string // Deprecated;  use deployment_config table instead   //git repository where chart is stored
	ChartLocation           string //location within git repo where current chart is pointing
	ReferenceTemplate       string
	ImageDescriptorTemplate string
	ChartRefId              int
	Latest                  bool
	Previous                bool
	ReferenceChart          []byte
	IsBasicViewLocked       bool
	CurrentViewEditor       models.ChartsViewEditorType
	IsCustomGitRepository   bool // Deprecated;  use deployment_config table instead
	ResolvedGlobalOverride  string
	CreatedOn               time.Time
	CreatedBy               int32
	UpdatedOn               time.Time
	UpdatedBy               int32
}
