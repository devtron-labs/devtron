package types

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
)

type DockerArtifactStoreBean struct {
	Id                      string                       `json:"id" validate:"required"`
	PluginId                string                       `json:"pluginId,omitempty" validate:"required"`
	RegistryURL             string                       `json:"registryUrl" validate:"required"`
	RegistryType            repository.RegistryType      `json:"registryType" validate:"required"`
	IsOCICompliantRegistry  bool                         `json:"isOCICompliantRegistry"`
	OCIRegistryConfig       map[string]string            `json:"ociRegistryConfig,omitempty"`
	IsPublic                bool                         `json:"isPublic"`
	RepositoryList          []string                     `json:"repositoryList,omitempty"`
	AWSAccessKeyId          string                       `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey      string                       `json:"awsSecretAccessKey,omitempty"`
	AWSRegion               string                       `json:"awsRegion,omitempty"`
	Username                string                       `json:"username,omitempty"`
	Password                string                       `json:"password,omitempty"`
	IsDefault               bool                         `json:"isDefault"`
	Connection              string                       `json:"connection"`
	Cert                    string                       `json:"cert"`
	Active                  bool                         `json:"active"`
	DisabledFields          []DisabledFields             `json:"disabledFields"`
	User                    int32                        `json:"-"`
	DockerRegistryIpsConfig *DockerRegistryIpsConfigBean `json:"ipsConfig,omitempty"`
}

func LoadFromEntity(dockerArtifactEntity *repository.DockerArtifactStore) *DockerArtifactStoreBean {
	artifactBean := &DockerArtifactStoreBean{
		Id:           dockerArtifactEntity.Id,
		RegistryURL:  dockerArtifactEntity.RegistryURL,
		RegistryType: dockerArtifactEntity.RegistryType,
	}
	return artifactBean
}

type DockerRegistryIpsConfigBean struct {
	Id                   int                                        `json:"id"`
	CredentialType       repository.DockerRegistryIpsCredentialType `json:"credentialType,omitempty" validate:"oneof=SAME_AS_REGISTRY NAME CUSTOM_CREDENTIAL"`
	CredentialValue      string                                     `json:"credentialValue,omitempty"`
	AppliedClusterIdsCsv string                                     `json:"appliedClusterIdsCsv,omitempty"`
	IgnoredClusterIdsCsv string                                     `json:"ignoredClusterIdsCsv,omitempty"`
	Active               bool                                       `json:"active,omitempty"`
}

type DisabledFields string
