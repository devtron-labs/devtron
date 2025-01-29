package registry

type Registry string

func (r Registry) String() string {
	return string(r)
}

const (
	DOCKER_REGISTRY_TYPE_ECR        Registry = "ecr"
	DOCKER_REGISTRY_TYPE_ACR        Registry = "acr"
	DOCKER_REGISTRY_TYPE_DOCKERHUB  Registry = "docker-hub"
	DOCKER_REGISTRY_TYPE_OTHER      Registry = "other"
	REGISTRY_TYPE_ARTIFACT_REGISTRY Registry = "artifact-registry"
	REGISTRY_TYPE_GCR               Registry = "gcr"
)

const JSON_KEY_USERNAME = "_json_key"

type RegistryCredential struct {
	RegistryType       Registry `json:"registryType"`
	RegistryURL        string   `json:"registryURL"`
	Username           string   `json:"username"`
	Password           string   `json:"password"`
	AWSAccessKeyId     string   `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string   `json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string   `json:"awsRegion,omitempty"`
}
