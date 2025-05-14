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
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"time"
)

type EnvConfigOverride struct {
	Id                        int
	ChartId                   int
	TargetEnvironment         int //target environment
	EnvOverrideValues         string
	EnvOverridePatchValues    string
	Status                    models.ChartStatus //new, deployment-in-progress, error, rollbacked, su
	ManualReviewed            bool
	Active                    bool
	Namespace                 string
	Chart                     *chartRepoRepository.Chart
	Environment               *repository2.Environment
	Latest                    bool
	Previous                  bool
	IsOverride                bool
	IsBasicViewLocked         bool
	CurrentViewEditor         models.ChartsViewEditorType
	CreatedOn                 time.Time
	CreatedBy                 int32
	UpdatedOn                 time.Time
	UpdatedBy                 int32
	ResolvedEnvOverrideValues string
	MergeStrategy             models.MergeStrategy
	VariableSnapshot          map[string]string
	//ResolvedEnvOverrideValuesForCM string
	VariableSnapshotForCM map[string]string
	//ResolvedEnvOverrideValuesForCS string
	VariableSnapshotForCS map[string]string
}

func (e *EnvConfigOverride) IsOverridden() bool {
	return e != nil && e.Id != 0 && e.IsOverride
}
