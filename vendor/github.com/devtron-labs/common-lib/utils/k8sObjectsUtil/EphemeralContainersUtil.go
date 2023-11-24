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
