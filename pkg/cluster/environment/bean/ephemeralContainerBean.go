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

import "github.com/devtron-labs/devtron/pkg/cluster/repository"

type EphemeralContainerRequest struct {
	BasicData                   *EphemeralContainerBasicData    `json:"basicData"`
	AdvancedData                *EphemeralContainerAdvancedData `json:"advancedData"`
	Namespace                   string                          `json:"namespace" validate:"required"`
	ClusterId                   int                             `json:"clusterId" validate:"gt=0"`
	PodName                     string                          `json:"podName"   validate:"required"`
	ExternalArgoApplicationName string                          `json:"externalArgoApplicationName,omitempty"`
	UserId                      int32                           `json:"-"`
}

type EphemeralContainerAdvancedData struct {
	Manifest string `json:"manifest"`
}

type EphemeralContainerBasicData struct {
	ContainerName       string `json:"containerName"`
	TargetContainerName string `json:"targetContainerName"`
	Image               string `json:"image"`
}

func (request EphemeralContainerRequest) GetContainerBean() repository.EphemeralContainerBean {
	return repository.EphemeralContainerBean{
		Name:                request.BasicData.ContainerName,
		ClusterId:           request.ClusterId,
		Namespace:           request.Namespace,
		PodName:             request.PodName,
		TargetContainer:     request.BasicData.TargetContainerName,
		Config:              request.AdvancedData.Manifest,
		IsExternallyCreated: false,
	}
}
