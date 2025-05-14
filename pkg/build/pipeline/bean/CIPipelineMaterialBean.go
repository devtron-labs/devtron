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
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"time"
)

type CiPipelineMaterialResponse struct {
	Id              int                    `json:"id"`
	GitMaterialId   int                    `json:"gitMaterialId"`
	GitMaterialUrl  string                 `json:"gitMaterialUrl"`
	GitMaterialName string                 `json:"gitMaterialName"`
	Type            string                 `json:"type"`
	Value           string                 `json:"value"`
	Active          bool                   `json:"active"`
	History         []*gitSensor.GitCommit `json:"history,omitempty"`
	LastFetchTime   time.Time              `json:"lastFetchTime"`
	IsRepoError     bool                   `json:"isRepoError"`
	RepoErrorMsg    string                 `json:"repoErrorMsg"`
	IsBranchError   bool                   `json:"isBranchError"`
	BranchErrorMsg  string                 `json:"branchErrorMsg"`
	Url             string                 `json:"url"`
	Regex           string                 `json:"regex"`
}

type MaterialTriggerInfo struct {
	GitTriggers map[int]pipelineConfig.GitCommit `json:"gitTriggers"`
	CiMaterials []CiPipelineMaterialResponse     `json:"ciMaterials"`
}
