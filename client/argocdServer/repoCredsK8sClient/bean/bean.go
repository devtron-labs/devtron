package bean

const (
	REPOSITORY_SECRET_NAME_KEY       = "name"
	REPOSITORY_SECRET_URL_KEY        = "url"
	REPOSITORY_SECRET_USERNAME_KEY   = "username"
	REPOSITORY_SECRET_PASSWORD_KEY   = "password"
	REPOSITORY_SECRET_ENABLE_OCI_KEY = "enableOCI"
	REPOSITORY_SECRET_TYPE_KEY       = "type"
	REPOSITORY_TYPE_HELM             = "helm"
	ARGOCD_REPOSITORY_SECRET_KEY     = "argocd.argoproj.io/secret-type"
	ARGOCD_REPOSITORY_SECRET_VALUE   = "repository"
	REPOSITORY_SECRET_INSECURE_KEY   = "insecure"
)

type ChartRepositoryAddRequest struct {
	Name                    string
	Username                string
	Password                string
	URL                     string
	AllowInsecureConnection bool
	IsPrivateChart          bool
}

type ChartRepositoryUpdateRequest struct {
	PreviousName            string
	PreviousURL             string
	Name                    string
	AuthMode                string
	Username                string
	Password                string
	SSHKey                  string
	URL                     string
	AllowInsecureConnection bool
	IsPrivateChart          bool
}

type KeyDto struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
	Url  string `json:"url,omitempty"`
}

type AcdConfigMapRepositoriesDto struct {
	Type           string  `json:"type,omitempty"`
	Name           string  `json:"name,omitempty"`
	Url            string  `json:"url,omitempty"`
	UsernameSecret *KeyDto `json:"usernameSecret,omitempty"`
	PasswordSecret *KeyDto `json:"passwordSecret,omitempty"`
	CaSecret       *KeyDto `json:"caSecret,omitempty"`
	CertSecret     *KeyDto `json:"certSecret,omitempty"`
	KeySecret      *KeyDto `json:"keySecret,omitempty"`
}
