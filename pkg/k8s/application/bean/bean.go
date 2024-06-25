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
	"github.com/devtron-labs/common-lib/utils/k8s"
)

const (
	DEFAULT_NAMESPACE = "default"
	EVENT_K8S_KIND    = "Event"
	LIST_VERB         = "list"
	Delete            = "delete"
)

const (
	// App Type Identifiers
	DevtronAppType = 0 // Identifier for Devtron Apps
	HelmAppType    = 1 // Identifier for Helm Apps
	ArgoAppType    = 2
	// Deployment Type Identifiers
	HelmInstalledType = 0 // Identifier for Helm deployment
	ArgoInstalledType = 1 // Identifier for ArgoCD deployment
)

const (
	LastEventID                         = "Last-Event-ID"
	TimestampOffsetToAvoidDuplicateLogs = 1
	IntegerBase                         = 10
	IntegerBitSize                      = 64
)

const (
	LocalTimezoneInGMT = "GMT+0530"
	LocalTimeOffset    = 5*60*60 + 30*60
)

type ResourceInfo struct {
	PodName string `json:"podName"`
}

type DevtronAppIdentifier struct {
	ClusterId int `json:"clusterId"`
	AppId     int `json:"appId"`
	EnvId     int `json:"envId"`
}

type Response struct {
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	PointsTo string   `json:"pointsTo"`
	Urls     []string `json:"urls"`
}

type RotatePodResourceResponse struct {
	k8s.ResourceIdentifier
	ErrorResponse string `json:"errorResponse"`
}
