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

package ciPipeline

type LinkedCIDetails struct {
	AppName         string `sql:"app_name"`
	EnvironmentName string `sql:"environment_name"`
	TriggerMode     string `sql:"trigger_mode"`
	PipelineId      int    `sql:"pipeline_id"`
	AppId           int    `sql:"app_id"`
	EnvironmentId   int    `sql:"environment_id"`
}

type CiPipelinesMap struct {
	Id               int `json:"id"`
	ParentCiPipeline int `json:"parentCiPipeline"`
}
