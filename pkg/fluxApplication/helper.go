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

package fluxApplication

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"strconv"
	"strings"
)

/*
* appIdString contains four fields, separated by '|':
* 1. clusterId: The ID of the cluster, which is an integer value
* 2. namespace: The namespace, which is a string.
* 3. appName: The name of the Flux application (either Kustomization or HelmRelease), which is a string.
* 4. isKustomize: A boolean value indicating whether the application is of type Kustomization (true) or HelmRelease (false).
*
*
* Example: "123|my-namespace|my-app|true"
* - clusterId: "123"
* - namespace: "my-namespace"
* - appName: "my-app"
* - isKustomization: true
 */

func DecodeFluxExternalAppId(appId string) (*bean.FluxAppIdentifier, error) {
	component := strings.Split(appId, "|")
	if len(component) != 4 {
		return nil, fmt.Errorf("malformed app id %s", appId)
	}
	clusterId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	isKustomizeApp, err := strconv.ParseBool(component[3])
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, fmt.Errorf("target cluster is not provided")
	}
	return &bean.FluxAppIdentifier{
		ClusterId:      clusterId,
		Namespace:      component[1],
		Name:           component[2],
		IsKustomizeApp: isKustomizeApp,
	}, nil
}
