package util

import (
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"regexp"
	"strings"
)

type K8sUtilConfig struct {
	EphemeralServerVersionRegex string `env:"EPHEMERAL_SERVER_VERSION_REGEX" envDefault:"v[1-9]\\.\\b(2[3-9]|[3-9][0-9])\\b.*"`
}

func CheckIfValidLabel(labelKey string, labelValue string) error {
	labelKey = strings.TrimSpace(labelKey)
	labelValue = strings.TrimSpace(labelValue)

	errs := validation.IsQualifiedName(labelKey)
	if len(labelKey) == 0 || len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label key - %s is not satisfying the label key criteria", labelKey))
	}

	errs = validation.IsValidLabelValue(labelValue)
	if len(labelValue) == 0 || len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label value - %s is not satisfying the label value criteria for label key - %s", labelValue, labelKey))
	}
	return nil
}

type EphemeralContainerData struct {
	Name       string `json:"name"`
	IsExternal bool   `json:"isExternal"`
}

func ExtractEphemeralContainers(pods []corev1.Pod) map[string][]*EphemeralContainerData {
	ephemeralContainersMap := make(map[string][]*EphemeralContainerData)
	for _, pod := range pods {

		ephemeralContainers := make([]*EphemeralContainerData, 0, len(pod.Spec.EphemeralContainers))
		for _, ec := range pod.Spec.EphemeralContainers {
			ephemeralContainerStatusMap := make(map[string]bool)
			for _, c := range pod.Status.EphemeralContainerStatuses {
				//c.state contains three states running,waiting and terminated
				// at any point of time only one state will be there
				if c.State.Running != nil {
					ephemeralContainerStatusMap[c.Name] = true
				}
			}

			//sending only running ephemeral containers in the list
			if _, ok := ephemeralContainerStatusMap[ec.Name]; ok {
				containerData := EphemeralContainerData{
					Name:       ec.Name,
					IsExternal: isExternalEphemeralContainer(ec.Command, ec.Name),
				}
				ephemeralContainers = append(ephemeralContainers, &containerData)
			}
		}
		ephemeralContainersMap[pod.Name] = ephemeralContainers

	}
	return ephemeralContainersMap
}

func isExternalEphemeralContainer(cmds []string, name string) bool {
	isExternal := true
	matchingCmd := fmt.Sprintf("sh %s-devtron.sh", name)
	for _, cmd := range cmds {
		if strings.Contains(cmd, matchingCmd) {
			isExternal = false
			break
		}
	}
	return isExternal
}

func MatchRegex(exp string, text string) (bool, error) {
	rExp, err := regexp.Compile(exp)
	if err != nil {
		return false, err
	}
	matched := rExp.Match([]byte(text))
	return matched, nil
}
