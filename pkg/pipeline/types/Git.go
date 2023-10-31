package types

import "github.com/devtron-labs/devtron/internal/sql/repository"

type GitRegistry struct {
	Id            int                 `json:"id,omitempty" validate:"number"`
	Name          string              `json:"name,omitempty" validate:"required"`
	Url           string              `json:"url,omitempty"`
	UserName      string              `json:"userName,omitempty"`
	Password      string              `json:"password,omitempty"`
	SshPrivateKey string              `json:"sshPrivateKey,omitempty"`
	AccessToken   string              `json:"accessToken,omitempty"`
	AuthMode      repository.AuthMode `json:"authMode,omitempty" validate:"required"`
	Active        bool                `json:"active"`
	UserId        int32               `json:"-"`
	GitHostId     int                 `json:"gitHostId"`
}

type GitHostRequest struct {
	Id              int    `json:"id,omitempty" validate:"number"`
	Name            string `json:"name,omitempty" validate:"required"`
	Active          bool   `json:"active"`
	WebhookUrl      string `json:"webhookUrl"`
	WebhookSecret   string `json:"webhookSecret"`
	EventTypeHeader string `json:"eventTypeHeader"`
	SecretHeader    string `json:"secretHeader"`
	SecretValidator string `json:"secretValidator"`
	UserId          int32  `json:"-"`
}
