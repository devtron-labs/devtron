package drafts

import (
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"time"
)

const (
	LastVersionOutdated         = "last-version-outdated"
	DraftAlreadyInTerminalState = "already-in-terminal-state"
	ApprovalRequestNotRaised    = "approval-request-not-raised"
	UserContributedToDraft      = "user-committed-to-draft"
	TemplateOutdated            = "template-outdated"
	FailedToDeleteComment       = "failed to delete comment"
	ConfigProtectionDisabled    = "config-protection-disabled"
)

type DraftResourceType uint8

const (
	CMDraftResource            DraftResourceType = 1
	CSDraftResource            DraftResourceType = 2
	DeploymentTemplateResource DraftResourceType = 3
)

func (draftType DraftResourceType) GetDraftResourceType() client.ResourceType {
	switch draftType {
	case CMDraftResource:
		return client.CM
	case CSDraftResource:
		return client.CS
	case DeploymentTemplateResource:
		return client.DeploymentTemplate
	}
	return ""
}

type ResourceAction uint8

const (
	AddResourceAction    ResourceAction = 1
	UpdateResourceAction ResourceAction = 2
	DeleteResourceAction ResourceAction = 3
)

type DraftState uint8

const (
	InitDraftState          DraftState = 1
	DiscardedDraftState     DraftState = 2
	PublishedDraftState     DraftState = 3
	AwaitApprovalDraftState DraftState = 4
)

func (state DraftState) IsTerminal() bool {
	return state == DiscardedDraftState || state == PublishedDraftState
}

func GetNonTerminalDraftStates() []int {
	return []int{int(InitDraftState), int(AwaitApprovalDraftState)}
}

type ConfigDraftRequest struct {
	AppId                     int                       `json:"appId" validate:"number,required"`
	EnvId                     int                       `json:"envId"`
	Resource                  DraftResourceType         `json:"resource"`
	ResourceName              string                    `json:"resourceName"`
	Action                    ResourceAction            `json:"action"`
	Data                      string                    `json:"data" validate:"min=1"`
	UserComment               string                    `json:"userComment"`
	ChangeProposed            bool                      `json:"changeProposed"`
	UserId                    int32                     `json:"-"`
	ProtectNotificationConfig ProtectNotificationConfig `json:"protectNotificationConfig"`
}
type ProtectNotificationConfig struct {
	EmailIds []string `json:"emailIds"`
}

func (request ConfigDraftRequest) TransformDraftRequestForNotification() client.ConfigDataForNotification {
	return client.ConfigDataForNotification{
		AppId:        request.AppId,
		EnvId:        request.EnvId,
		Resource:     request.Resource.GetDraftResourceType(),
		ResourceName: request.ResourceName,
		UserComment:  request.UserComment,
		UserId:       request.UserId,
		EmailIds:     request.ProtectNotificationConfig.GetEmailIdsForProtectConfig(),
	}
}

func (protectNotificationConfig ProtectNotificationConfig) GetEmailIdsForProtectConfig() []string {
	return protectNotificationConfig.EmailIds
}
func (request ConfigDraftRequest) GetDraftDto() *DraftDto {
	draftState := InitDraftState
	if proposed := request.ChangeProposed; proposed {
		draftState = AwaitApprovalDraftState
	}
	metadataDto := &DraftDto{
		AppId:        request.AppId,
		EnvId:        request.EnvId,
		Resource:     request.Resource,
		ResourceName: request.ResourceName,
		DraftState:   draftState,
	}
	currentTime := time.Now()
	metadataDto.CreatedOn = currentTime
	metadataDto.UpdatedOn = currentTime
	metadataDto.CreatedBy = request.UserId
	metadataDto.UpdatedBy = request.UserId
	return metadataDto
}

func (request ConfigDraftRequest) GetDraftVersionDto(draftMetadataId int, timestamp time.Time) *DraftVersion {
	draftVersionDto := &DraftVersion{
		DraftsId:  draftMetadataId,
		Action:    request.Action,
		Data:      request.Data,
		UserId:    request.UserId,
		CreatedOn: timestamp,
	}
	return draftVersionDto
}

func (request ConfigDraftRequest) GetDraftVersionComment(draftMetadataId, draftVersionId int, timestamp time.Time) *DraftVersionComment {
	draftVersionCommentDto := &DraftVersionComment{}
	draftVersionCommentDto.DraftId = draftMetadataId
	draftVersionCommentDto.DraftVersionId = draftVersionId
	draftVersionCommentDto.Comment = request.UserComment
	draftVersionCommentDto.Active = true
	draftVersionCommentDto.CreatedBy = request.UserId
	draftVersionCommentDto.UpdatedBy = request.UserId
	draftVersionCommentDto.CreatedOn = timestamp
	draftVersionCommentDto.UpdatedOn = timestamp
	return draftVersionCommentDto
}

type ConfigDraftResponse struct {
	ConfigDraftRequest
	*bean.LockValidateErrorResponse
	DraftId        int        `json:"draftId"`
	DraftVersionId int        `json:"draftVersionId"`
	DraftState     DraftState `json:"draftState"`
	Approvers      []string   `json:"approvers"`
	CanApprove     *bool      `json:"canApprove,omitempty"`
	CommentsCount  int        `json:"commentsCount"`
	DataEncrypted  bool       `json:"dataEncrypted"`
	IsAppAdmin     bool       `json:"isAppAdmin"`
}

//type LockValidateError struct {
//	*ResponseError
//	LockedOverride json.RawMessage `json:"lockedOverride"`
//}
//
//type ResponseError struct {
//	ErrorMessage string
//	ErrorCode    string
//}

type DraftCountResponse struct {
	AppId       int `json:"appId"`
	EnvId       int `json:"envId"`
	DraftsCount int `json:"draftsCount"`
}

type ConfigDraftVersionRequest struct {
	DraftId                   int                       `json:"draftId" validate:"number,required"`
	LastDraftVersionId        int                       `json:"lastDraftVersionId" validate:"number,required"`
	Action                    ResourceAction            `json:"action"`
	Data                      string                    `json:"data"`
	UserComment               string                    `json:"userComment"`
	ChangeProposed            bool                      `json:"changeProposed"`
	UserId                    int32                     `json:"-"`
	ProtectNotificationConfig ProtectNotificationConfig `json:"protectNotificationConfig"`
}

func (request ConfigDraftVersionRequest) GetDraftVersionDto(currentTime time.Time) *DraftVersion {
	draftVersionDto := &DraftVersion{}
	draftVersionDto.DraftsId = request.DraftId
	draftVersionDto.Data = request.Data
	draftVersionDto.Action = request.Action
	draftVersionDto.UserId = request.UserId
	draftVersionDto.CreatedOn = currentTime
	return draftVersionDto
}

func (request ConfigDraftVersionRequest) GetDraftVersionComment(lastDraftVersionId int, currentTime time.Time) *DraftVersionComment {
	draftVersionCommentDto := &DraftVersionComment{}
	draftVersionCommentDto.DraftId = request.DraftId
	draftVersionCommentDto.DraftVersionId = lastDraftVersionId
	draftVersionCommentDto.Comment = request.UserComment
	draftVersionCommentDto.Active = true
	draftVersionCommentDto.CreatedBy = request.UserId
	draftVersionCommentDto.UpdatedBy = request.UserId
	draftVersionCommentDto.CreatedOn = currentTime
	draftVersionCommentDto.UpdatedOn = currentTime
	return draftVersionCommentDto
}

type DraftVersionMetadataResponse struct {
	DraftId       int                     `json:"draftId"`
	DraftVersions []*DraftVersionMetadata `json:"versionMetadata"`
}

type DraftVersionMetadata struct {
	DraftVersionId int       `json:"draftVersionId"`
	UserId         int32     `json:"userId"`
	UserEmail      string    `json:"userEmail"`
	ActivityTime   time.Time `json:"activityTime"`
}

type DraftVersionCommentResponse struct {
	DraftId              int                       `json:"draftId"`
	DraftVersionComments []DraftVersionCommentBean `json:"versionComments"`
}

type DraftVersionCommentBean struct {
	DraftVersionId int                   `json:"draftVersionId"`
	UserComments   []UserCommentMetadata `json:"userComments"`
}

type UserCommentMetadata struct {
	CommentId   int       `json:"commentId"`
	UserId      int32     `json:"userId"`
	UserEmail   string    `json:"userEmail"`
	CommentedAt time.Time `json:"commentedAt"`
	Comment     string    `json:"comment"`
}

type AppConfigDraft struct {
	DraftId      int               `json:"draftId"`
	Resource     DraftResourceType `json:"resourceType"`
	ResourceName string            `json:"resourceName"`
	DraftState   DraftState        `json:"draftState"`
}

type DraftVersionResponse struct {
	DraftVersionId                  int `json:"draftVersionId"`
	*bean.LockValidateErrorResponse     // check if got error of lock config
}
