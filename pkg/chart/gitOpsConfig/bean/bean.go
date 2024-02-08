package bean

type AppGitOpsConfigRequest struct {
	AppId         int    `json:"appId" validate:"required"`
	GitOpsRepoURL string `json:"gitRepoURL" validate:"required"`
	UserId        int32  `json:"-"`
}

type AppGitOpsConfigResponse struct {
	GitRepoURL string `json:"gitRepoURL"`
	IsEditable bool   `json:"isEditable"`
}
