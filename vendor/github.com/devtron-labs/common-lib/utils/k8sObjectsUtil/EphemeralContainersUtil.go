package k8sObjectsUtil

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

const EphemeralContainerStartingShellScriptFileName = "./tmp/%s-devtron.sh"

type EphemeralContainerData struct {
	Name string `json:"name"`
	// IsExternal flag indicates that this ephemeral container is not managed by devtron.(neither created nor can be killed by devtron)
	IsExternal bool `json:"isExternal"`
}

// SetEphemeralContainersInManifestResponse will extract out all the running ephemeral containers of the given pod manifest and sets in manifestResponse.EphemeralContainers
// if given manifest is not of pod kind, it just returns the manifestResponse arg
func SetEphemeralContainersInManifestResponse(manifestResponse *k8s.ManifestResponse) (*k8s.ManifestResponse, error) {
	if manifestResponse != nil {
		if podManifest := manifestResponse.Manifest; isPod(podManifest.GetKind(), podManifest.GroupVersionKind().Group) {
			pod := corev1.Pod{}
			// Convert the unstructured object to a Pod object
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(podManifest.Object, &pod)
			if err != nil {
				return manifestResponse, err
			}
			runningEphemeralContainers := ExtractEphemeralContainers([]corev1.Pod{pod})
			manifestResponse.EphemeralContainers = runningEphemeralContainers[pod.Name]
		}
	}
	return manifestResponse, nil
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

func isPod(kind string, group string) bool {
	return kind == "Pod" && group == ""
}
