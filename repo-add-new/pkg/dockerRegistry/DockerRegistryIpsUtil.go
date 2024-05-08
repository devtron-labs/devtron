package dockerRegistry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

type DockerIpsCustomCredential struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv string, ignoredClusterIdsCsv string, clusterId int, isVirtualEnv bool) bool {
	if isVirtualEnv {
		return false
	}
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
	var ipsName string
	if ipsCredentialType == IPS_CREDENTIAL_TYPE_NAME {
		ipsName = ipsCredentialValue
	} else {
		ipsName = fmt.Sprintf("%s-%s-%s", nonAlphanumericRegex.ReplaceAllString(dockerRegistryId, ""), "dtron", "ips")
	}
	if len(ipsName) > 0 {
		ipsName = strings.ToLower(ipsName)
	}
	return ipsName
}

func BuildIpsData(dockerRegistryUrl, dockerRegistryUsername, dockerRegistryPassword, dockerRegistryEmail string) map[string][]byte {
	config, _ := handleDockerCfgJSONContent(dockerRegistryUrl, dockerRegistryUsername, dockerRegistryPassword, dockerRegistryEmail)
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

func handleDockerCfgJSONContent(server, username, password, email string) ([]byte, error) {
	dockerConfigAuth := create.DockerConfigEntry{
		Username: username,
		Password: password,
		Auth:     encodeDockerConfigFieldAuth(username, password),
	}
	if len(email) > 0 {
		dockerConfigAuth.Email = email
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

// returns username and password
func CreateCredentialForEcr(awsRegion, awsAccessKey, awsSecretKey string) (string, string, error) {
	creds := credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")
	sess, err := session.NewSession(&aws.Config{
		Region:      &awsRegion,
		Credentials: creds,
	})
	if err != nil {
		return "", "", err
	}
	svc := ecr.New(sess)
	input := &ecr.GetAuthorizationTokenInput{}
	authData, err := svc.GetAuthorizationToken(input)
	if err != nil {
		return "", "", err
	}

	// decode token
	token := authData.AuthorizationData[0].AuthorizationToken
	decodedToken, err := base64.StdEncoding.DecodeString(*token)
	if err != nil {
		return "", "", err
	}
	credsSlice := strings.Split(string(decodedToken), ":")
	username := credsSlice[0]
	pwd := credsSlice[1]

	return username, pwd, nil
}
