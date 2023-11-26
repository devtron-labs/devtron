package chartRepo

import "github.com/devtron-labs/devtron/internal/sql/repository"

const ValidationSuccessMsg = "Configurations are validated successfully"

type ChartRepoDto struct {
	Id                      int                 `json:"id,omitempty" validate:"number"`
	Name                    string              `json:"name,omitempty" validate:"required"`
	Url                     string              `json:"url,omitempty"`
	UserName                string              `json:"userName,omitempty"`
	Password                string              `json:"password,omitempty"`
	SshKey                  string              `json:"sshKey,omitempty"`
	AccessToken             string              `json:"accessToken,omitempty"`
	AuthMode                repository.AuthMode `json:"authMode,omitempty" validate:"required"`
	Active                  bool                `json:"active"`
	Default                 bool                `json:"default"`
	UserId                  int32               `json:"-"`
	AllowInsecureConnection bool                `json:"allow_insecure_connection"`
}

type ChartRepoWithIsEditableDto struct {
	ChartRepoDto
	IsEditable bool `json:"isEditable"`
}

type DetailedErrorHelmRepoValidation struct {
	CustomErrMsg string `json:"customErrMsg"`
	ActualErrMsg string `json:"actualErrMsg"`
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

const (
	GIT_REPO_DEFAULT        = "Default"
	GIT_REPO_NOT_CONFIGURED = "NOT_CONFIGURED" // The value of the constant has been used in the migration script for `custom_gitops_repo_url`; Need to add another migration script if the value is updated.
)
