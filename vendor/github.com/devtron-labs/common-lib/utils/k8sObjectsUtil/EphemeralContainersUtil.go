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

package k8sObjectsUtil

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

const EphemeralContainerStartingShellScriptFileName = "./tmp/%s-devtron.sh"

type EphemeralContainerData struct {
	Name string `json:"name"`
	// IsExternal flag indicates that this ephemeral container is not managed by devtron.(neither created nor can be killed by devtron)
	IsExternal bool `json:"isExternal"`
}

// ExtractEphemeralContainers will only return map of pod_name vs running ephemeral_containers list of this pod.
// Note: pod may contain other ephemeral containers which are not in running state
func ExtractEphemeralContainers(pods []corev1.Pod) map[string][]*EphemeralContainerData {
	ephemeralContainersMap := make(map[string][]*EphemeralContainerData)
	for _, pod := range pods {

		ephemeralContainers := make([]*EphemeralContainerData, 0, len(pod.Spec.EphemeralContainers))
		ephemeralContainerStatusMap := make(map[string]bool)
		for _, c := range pod.Status.EphemeralContainerStatuses {
			//c.state contains three states running,waiting and terminated
			// at any point of time only one state will be there
			if c.State.Running != nil {
				ephemeralContainerStatusMap[c.Name] = true
			}
		}

		for _, ec := range pod.Spec.EphemeralContainers {
			//sending only running ephemeral containers in the list
			if _, ok := ephemeralContainerStatusMap[ec.Name]; ok {
				containerData := EphemeralContainerData{
					Name:       ec.Name,
					IsExternal: IsExternalEphemeralContainer(ec.Command, ec.Name),
				}
				ephemeralContainers = append(ephemeralContainers, &containerData)
			}
		}

		ephemeralContainersMap[pod.Name] = ephemeralContainers

	}
	return ephemeralContainersMap
}

func IsExternalEphemeralContainer(cmds []string, name string) bool {
	isExternal := true
	matchingCmd := fmt.Sprintf("sh "+EphemeralContainerStartingShellScriptFileName, name)
	for _, cmd := range cmds {
		if strings.Contains(cmd, matchingCmd) {
			isExternal = false
			break
		}
	}
	return isExternal
}

func IsPod(kind string, group string) bool {
	return kind == "Pod" && group == ""
}
