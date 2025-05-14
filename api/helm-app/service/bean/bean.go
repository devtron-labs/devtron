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
	"fmt"
	"strconv"
)

type AppIdentifier struct {
	ClusterId   int    `json:"clusterId"`
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"releaseName"`
}

// GetUniqueAppNameIdentifier returns unique app name identifier, we store all helm releases in kubelink cache with key
// as what is returned from this func, this is the case where an app across diff namespace or cluster can have same name,
// so to identify then uniquely below implementation would serve as good unique identifier for an external app.
func (r *AppIdentifier) GetUniqueAppNameIdentifier() string {
	return fmt.Sprintf("%s-%s-%s", r.ReleaseName, r.Namespace, strconv.Itoa(r.ClusterId))
}

func (r *AppIdentifier) GetUniqueAppIdentifierForGivenNamespaceAndCluster(namespace, clusterId string) string {
	return fmt.Sprintf("%s-%s-%s", r.ReleaseName, namespace, clusterId)
}

type ExternalHelmAppListingResult struct {
	ReleaseName   string `json:"releaseName"`
	ClusterId     int    `json:"clusterId"`
	Namespace     string `json:"namespace"`
	EnvironmentId string `json:"environmentId"`
	Status        string `json:"status"`
	ChartAvatar   string `json:"chartAvatar"`
}
