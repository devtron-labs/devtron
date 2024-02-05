package drafts

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

type DraftDto struct {
	tableName    struct{}          `sql:"draft" pg:",discard_unknown_columns"`
	Id           int               `sql:"id,pk"`
	AppId        int               `sql:"app_id,notnull"`
	EnvId        int               `sql:"env_id,notnull"`
	Resource     DraftResourceType `sql:"resource,notnull"`
	ResourceName string            `sql:"resource_name,notnull"`
	DraftState   DraftState        `sql:"draft_state"`
	sql.AuditLog
}

type DraftVersion struct {
	tableName struct{}       `sql:"draft_version" pg:",discard_unknown_columns"`
	Id        int            `sql:"id,pk"`
	DraftsId  int            `sql:"draft_id,notnull"`
	Data      string         `sql:"data,notnull"`
	Action    ResourceAction `sql:"action,notnull"`
	UserId    int32          `sql:"user_id,notnull"`
	CreatedOn time.Time      `sql:"created_on,type:timestamptz"`
	Draft     *DraftDto
}

type DraftVersionComment struct {
	tableName      struct{} `sql:"draft_version_comment" pg:",discard_unknown_columns"`
	Id             int      `sql:"id,pk"`
	DraftId        int      `sql:"draft_id,notnull"`
	DraftVersionId int      `sql:"draft_version_id"`
	Comment        string   `sql:"comment"`
	Active         bool     `sql:"active"`
	sql.AuditLog
}

func (dto DraftDto) ConvertToAppConfigDraft() AppConfigDraft {
	appConfigDraft := AppConfigDraft{
		DraftId:      dto.Id,
		Resource:     dto.Resource,
		ResourceName: dto.ResourceName,
		DraftState:   dto.DraftState,
	}
	return appConfigDraft
}

func (dto DraftVersion) ConvertToDraftVersionMetadata() *DraftVersionMetadata {
	draftVersionMetadata := &DraftVersionMetadata{
		DraftVersionId: dto.Id,
		UserId:         dto.UserId,
		ActivityTime:   dto.CreatedOn,
	}
	return draftVersionMetadata
}

func (dto DraftVersion) ConvertToConfigDraft() *ConfigDraftResponse {
	configDraftResponse := &ConfigDraftResponse{
		DraftId:        dto.DraftsId,
		DraftVersionId: dto.Id,
	}
	configDraftResponse.Data = dto.Data
	configDraftResponse.Action = dto.Action
	configDraftResponse.UserId = dto.UserId
	if draftsDto := dto.Draft; draftsDto != nil {
		configDraftResponse.AppId = draftsDto.AppId
		configDraftResponse.EnvId = draftsDto.EnvId
		configDraftResponse.Resource = draftsDto.Resource
		configDraftResponse.ResourceName = draftsDto.ResourceName
		configDraftResponse.DraftState = draftsDto.DraftState
	}
	return configDraftResponse
}

func (dto DraftVersionComment) ConvertToDraftVersionComment() UserCommentMetadata {
	userComment := UserCommentMetadata{
		CommentId:   dto.Id,
		UserId:      dto.CreatedBy,
		CommentedAt: dto.CreatedOn,
		Comment:     dto.Comment,
	}
	return userComment
}
