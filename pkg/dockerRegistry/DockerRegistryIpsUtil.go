package dockerRegistry

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const ALL_CLUSTER_ID string = "-1"
const IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY = "SAME_AS_REGISTRY"
const IPS_CREDENTIAL_TYPE_NAME = "NAME"
const IPS_CREDENTIAL_TYPE_CUSTOM_CREDENTIAL = "CUSTOM_CREDENTIAL"
const IMAGE_PULL_SECRET_KEY_IN_VALUES_YAML = "imagePullSecrets"

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv string, ignoredClusterIdsCsv string, clusterId int) bool {
	clusterIdStr := strconv.Itoa(clusterId)
	if len(appliedClusterIdsCsv) > 0 {
		appliedClusterIds := strings.Split(appliedClusterIdsCsv, ",")
		for _, appliedClusterId := range appliedClusterIds {
			if appliedClusterId == ALL_CLUSTER_ID || appliedClusterId == clusterIdStr {
				return true
			}
		}
		return false
	}
	if len(ignoredClusterIdsCsv) > 0 {
		ignoredClusterIds := strings.Split(ignoredClusterIdsCsv, ",")
		for _, ignoredClusterId := range ignoredClusterIds {
			if ignoredClusterId == ALL_CLUSTER_ID || ignoredClusterId == clusterIdStr {
				return false
			}
		}
		return true
	}
	return false
}

// eg: quayio-dtron-ips
func BuildIpsName(dockerRegistryId string, ipsCredentialType string, ipsCredentialValue string) string {
	if ipsCredentialType == IPS_CREDENTIAL_TYPE_NAME {
		return ipsCredentialValue
	}
	return fmt.Sprintf("%s-%s-%s", nonAlphanumericRegex.ReplaceAllString(dockerRegistryId, ""), "dtron", "ips")
}

func BuildIpsData(dockerRegistryType string, dockerRegistryUsername string, dockerRegistryPassword string, ipsCredentialType string, ipsCredentialValue string) map[string][]byte {
	return nil
}

func SetIpsNameInValues(valuesContent []byte, ipsName string) ([]byte, error) {
	valuesMap := make(map[string]interface{})
	err := json.Unmarshal(valuesContent, &valuesMap)
	if err != nil {
		return nil, err
	}

	val, found := valuesMap[IMAGE_PULL_SECRET_KEY_IN_VALUES_YAML]
	var ipsNames []string
	if found {
		ipsNames = val.([]string)
	}

	var ipsNameFound bool
	for _, ipsNameVal := range ipsNames {
		if ipsNameVal == ipsName {
			ipsNameFound = true
			break
		}
	}

	if !ipsNameFound {
		ipsNames = append(ipsNames, ipsName)
		valuesMap[IMAGE_PULL_SECRET_KEY_IN_VALUES_YAML] = ipsNames
	}

	updatedValuesContent, err := json.Marshal(&valuesMap)
	if err != nil {
		return nil, err
	}
	return updatedValuesContent, nil
}
