package drafts

import "time"

type DraftResourceType uint8

var (
	CmDraftResource            DraftResourceType = 0
	CsDraftResource            DraftResourceType = 1
	DeploymentTemplateResource DraftResourceType = 2
)

type ResourceAction uint8

var (
	AddResourceAction    ResourceAction = 0
	UpdateResourceAction ResourceAction = 1
	DeleteResourceAction ResourceAction = 2
)

type DraftState uint8

var (
	InitDraftState          DraftState = 0
	DiscardedDraftState     DraftState = 1
	PublishedDraftState     DraftState = 2
	AwaitApprovalDraftState DraftState = 3
)

func (state DraftState) IsTerminal() bool {
	return state == DiscardedDraftState || state == PublishedDraftState
}

type ConfigDraftRequest struct {
	AppId        int               `json:"appId" validate:"number,required"`
	EnvId        int               `json:"envId"`
	Resource     DraftResourceType `json:"resource"`
	ResourceName string            `json:"resourceName"`
	Action       ResourceAction    `json:"action"`
	Data         string            `json:"data" validate:"min=1"`
	UserComment  string            `json:"userComment"`
	UserId       int32             `json:"-"`
}

type ConfigDraftResponse struct {
	ConfigDraftRequest
	DraftId        int `json:"draftId"`
	DraftVersionId int `json:"draftVersionId"`
}

type ConfigDraftVersionRequest struct {
	DraftId            int            `json:"draftId" validate:"number,required"`
	LastDraftVersionId int            `json:"lastDraftVersionId" validate:"number,required"`
	Action             ResourceAction `json:"action"`
	Data               string         `json:"data"`
	UserComment        string         `json:"userComment"`
	UserId             int32          `json:"-"`
}

type DraftVersionMetadataResponse struct {
	DraftId       int                    `json:"draftId"`
	DraftVersions []DraftVersionMetadata `json:"versionMetadata"`
}

type DraftVersionMetadata struct {
	DraftVersionId int       `json:"draftVersionId"`
	UserId         int32     `json:"userId"`
	UserEmail      string    `json:"userEmail"`
	ActivityTime   time.Time `json:"activityTime"`
}

type DraftVersionCommentResponse struct {
	DraftId              int                   `json:"draftId"`
	DraftVersionComments []DraftVersionComment `json:"versionComments"`
}

type DraftVersionComment struct {
	DraftVersionId int                   `json:"draftVersionId"`
	UserComments   []UserCommentMetadata `json:"userComments"`
}

type UserCommentMetadata struct {
	UserId      int32     `json:"userId"`
	UserEmail   string    `json:"userEmail"`
	CommentedAt time.Time `json:"commentedAt"`
}

type AppConfigDraft struct {
	DraftId      int               `json:"draftId"`
	Resource     DraftResourceType `json:"resourceType"`
	ResourceName string            `json:"resourceName"`
	DraftState   DraftState        `json:"draftState"`
}
