package dockerRegistry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/cmd/create"
	"regexp"
	"strconv"
	"strings"
)

const ALL_CLUSTER_ID string = "-1"
const IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY = "SAME_AS_REGISTRY"
const IPS_CREDENTIAL_TYPE_NAME = "NAME"
const IPS_CREDENTIAL_TYPE_CUSTOM_CREDENTIAL = "CUSTOM_CREDENTIAL"
const IMAGE_PULL_SECRET_KEY_IN_VALUES_YAML = "imagePullSecrets"
const IMAGE_PULL_SECRET_DATA_KEY_IN_SECRET = ".dockerconfigjson"

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

func BuildIpsData(dockerRegistryUrl, dockerRegistryUsername, dockerRegistryPassword, ipsCredentialType, ipsCredentialValue string) map[string][]byte {
	config, _ := handleDockerCfgJSONContent(dockerRegistryUrl, dockerRegistryUsername, dockerRegistryPassword)
	data := make(map[string][]byte)
	data[corev1.DockerConfigJsonKey] = config
	return data
}

func GetUsernamePasswordFromIpsSecret(dockerRegistryUrl string, data map[string][]byte) (string, string) {
	dockerConfig := create.DockerConfigJSON{}
	err := json.Unmarshal(data[corev1.DockerConfigJsonKey], &dockerConfig)
	if err == nil {
		val, found := dockerConfig.Auths[dockerRegistryUrl]
		if found {
			return val.Username, val.Password
		}
	}
	return "", ""
}

func handleDockerCfgJSONContent(server, username, password string) ([]byte, error) {
	dockerConfigAuth := create.DockerConfigEntry{
		Username: username,
		Password: password,
		Auth:     encodeDockerConfigFieldAuth(username, password),
	}
	dockerConfigJSON := create.DockerConfigJSON{
		Auths: map[string]create.DockerConfigEntry{server: dockerConfigAuth},
	}

	return json.Marshal(dockerConfigJSON)
}

// encodeDockerConfigFieldAuth returns base64 encoding of the username and password string
func encodeDockerConfigFieldAuth(username, password string) string {
	fieldValue := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(fieldValue))
}

func SetIpsNameInValues(valuesContent []byte, ipsName string) ([]byte, error) {
	valuesMap := make(map[string]interface{})
	err := json.Unmarshal(valuesContent, &valuesMap)
	if err != nil {
		return nil, err
	}

	val, found := valuesMap[IMAGE_PULL_SECRET_KEY_IN_VALUES_YAML]
	var ipsNames []interface{}
	if found {
		ipsNames = val.([]interface{})
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
