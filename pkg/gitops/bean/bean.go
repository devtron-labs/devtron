package bean

const GitOpsSecretName = "devtron-gitops-secret"

type RepositoryCredentialsDto struct {
	Url            string  `json:"url,omitempty"`
	UsernameSecret *KeyDto `json:"usernameSecret,omitempty"`
	PasswordSecret *KeyDto `json:"passwordSecret,omitempty"`
}

type KeyDto struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
}
