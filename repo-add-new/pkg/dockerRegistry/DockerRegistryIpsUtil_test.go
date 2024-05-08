package dockerRegistry

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"testing"
)

func TestAllClusterAccessProvided(t *testing.T) {
	appliedClusterIdsCsv := "-1"
	ignoredClusterIdsCsv := ""
	clusterId := 2
	accessProvided := CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv, ignoredClusterIdsCsv, clusterId)
	assert.True(t, accessProvided)
}

func TestClusterAccessProvidedInApplied(t *testing.T) {
	appliedClusterIdsCsv := "1,2,3"
	ignoredClusterIdsCsv := ""
	clusterId := 2
	accessProvided := CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv, ignoredClusterIdsCsv, clusterId)
	assert.True(t, accessProvided)
}

func TestClusterAccessNotProvidedInApplied(t *testing.T) {
	appliedClusterIdsCsv := "1,2,3"
	ignoredClusterIdsCsv := ""
	clusterId := 4
	accessProvided := CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv, ignoredClusterIdsCsv, clusterId)
	assert.False(t, accessProvided)
}

func TestNoClusterAccessProvided(t *testing.T) {
	appliedClusterIdsCsv := ""
	ignoredClusterIdsCsv := "-1"
	clusterId := 2
	accessProvided := CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv, ignoredClusterIdsCsv, clusterId)
	assert.False(t, accessProvided)
}

func TestClusterAccessProvidedInIgnored(t *testing.T) {
	appliedClusterIdsCsv := ""
	ignoredClusterIdsCsv := "1,2,3"
	clusterId := 2
	accessProvided := CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv, ignoredClusterIdsCsv, clusterId)
	assert.False(t, accessProvided)
}

func TestClusterAccessNotProvidedInIgnored(t *testing.T) {
	appliedClusterIdsCsv := ""
	ignoredClusterIdsCsv := "1,2,3"
	clusterId := 4
	accessProvided := CheckIfImagePullSecretAccessProvided(appliedClusterIdsCsv, ignoredClusterIdsCsv, clusterId)
	assert.True(t, accessProvided)
}

func TestIpsNameForNameType(t *testing.T) {
	ipsVal := "someName"
	ipsName := BuildIpsName("", IPS_CREDENTIAL_TYPE_NAME, ipsVal)
	assert.Equal(t, strings.ToLower(ipsVal), ipsName)
}

func TestIpsNameForNonNameTypeWithHyphen(t *testing.T) {
	dockerRegistryId := "devtron-quay"
	ipsName := BuildIpsName(dockerRegistryId, IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY, "")
	assert.Equal(t, "devtronquay-dtron-ips", ipsName)
}

func TestIpsNameForNonNameTypeWithSpace(t *testing.T) {
	dockerRegistryId := "devtron quay"
	ipsName := BuildIpsName(dockerRegistryId, IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY, "")
	assert.Equal(t, "devtronquay-dtron-ips", ipsName)
}

func TestIpsNameForNonNameTypeWithDot(t *testing.T) {
	dockerRegistryId := "devtron.quay"
	ipsName := BuildIpsName(dockerRegistryId, IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY, "")
	assert.Equal(t, "devtronquay-dtron-ips", ipsName)
}

func TestIpsNameForNonNameTypeWithCase(t *testing.T) {
	dockerRegistryId := "Devtron quay"
	ipsName := BuildIpsName(dockerRegistryId, IPS_CREDENTIAL_TYPE_SAME_AS_REGISTRY, "")
	assert.Equal(t, "devtronquay-dtron-ips", ipsName)
}

func TestBuildIpsData(t *testing.T) {
	dockerRegistryUrl := "someRegistryUrl"
	dockerRegistryUsername := "someUsername"
	dockerRegistryPassword := "somePassword"
	ipsData := BuildIpsData(dockerRegistryUrl, dockerRegistryUsername, dockerRegistryPassword, "")
	expectedIpsAuthsVal := make(map[string]interface{})
	expectedIpsVal := make(map[string]interface{})
	expectedIpsVal[dockerRegistryUrl] = map[string]string{
		"username": dockerRegistryUsername,
		"password": dockerRegistryPassword,
		"auth":     encodeDockerConfigFieldAuth(dockerRegistryUsername, dockerRegistryPassword),
	}
	expectedIpsAuthsVal["auths"] = expectedIpsVal
	actual := make(map[string]interface{})
	json.Unmarshal(ipsData[corev1.DockerConfigJsonKey], &actual)

	expectedIpsAuthsByteArr, _ := json.Marshal(expectedIpsAuthsVal)
	actualByteArr, _ := json.Marshal(actual)

	assert.Equal(t, expectedIpsAuthsByteArr, actualByteArr)
}

func TestDecryptIpsSecret(t *testing.T) {
	dockerRegistryUrl := "someRegistryUrl"
	dockerRegistryUsername := "someUsername"
	dockerRegistryPassword := "somePassword"
	ipsData := BuildIpsData(dockerRegistryUrl, dockerRegistryUsername, dockerRegistryPassword, "")
	username, password := GetUsernamePasswordFromIpsSecret(dockerRegistryUrl, ipsData)
	assert.Equal(t, username, dockerRegistryUsername)
	assert.Equal(t, password, dockerRegistryPassword)
}
