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
	v1alpha12 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"time"
)

type ApplicationDetail struct {
	Application *v1alpha12.Application `json:"application"`
	StatusTime  time.Time              `json:"statusTime"`
}

type ArgoPipelineStatusSyncEvent struct {
	PipelineId            int   `json:"pipelineId"`
	InstalledAppVersionId int   `json:"installedAppVersionId"`
	UserId                int32 `json:"userId"`
	IsAppStoreApplication bool  `json:"isAppStoreApplication"`
}
