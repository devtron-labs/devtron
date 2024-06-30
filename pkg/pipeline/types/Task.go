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

package types

type TaskYaml struct {
	Version          string             `yaml:"version"`
	CdPipelineConfig []CdPipelineConfig `yaml:"cdPipelineConf"`
}

type CdPipelineConfig struct {
	BeforeTasks []*Task `yaml:"beforeStages"`
	AfterTasks  []*Task `yaml:"afterStages"`
}

type Task struct {
	Id             int    `json:"id,omitempty"`
	Index          int    `json:"index,omitempty"`
	Name           string `json:"name" yaml:"name"`
	Script         string `json:"script" yaml:"script"`
	OutputLocation string `json:"outputLocation" yaml:"outputLocation"` // file/dir
	RunStatus      bool   `json:"-,omitempty"`                          // task run was attempted or not
}
