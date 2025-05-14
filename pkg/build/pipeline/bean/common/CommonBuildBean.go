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

package common

type PipelineType string

// default PipelineType
const DefaultPipelineType = CI_BUILD

const (
	CI_BUILD  PipelineType = "CI_BUILD"
	LINKED    PipelineType = "LINKED"
	EXTERNAL  PipelineType = "EXTERNAL"
	CI_JOB    PipelineType = "CI_JOB"
	LINKED_CD PipelineType = "LINKED_CD"
)

func (pType PipelineType) ToString() string {
	return string(pType)
}

func (pType PipelineType) IsValidPipelineType() bool {
	switch pType {
	case CI_BUILD, LINKED, EXTERNAL, CI_JOB, LINKED_CD:
		return true
	default:
		return false
	}
}

const (
	BEFORE_DOCKER_BUILD = "BEFORE_DOCKER_BUILD"
	AFTER_DOCKER_BUILD  = "AFTER_DOCKER_BUILD"
)

type RefPluginName = string

const (
	COPY_CONTAINER_IMAGE            RefPluginName = "Copy container image"
	COPY_CONTAINER_IMAGE_VERSION_V1               = "1.0.0"
	COPY_CONTAINER_IMAGE_VERSION_V2               = "2.0.0"
)
