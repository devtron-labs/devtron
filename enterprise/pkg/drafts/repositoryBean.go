package drafts

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

type DraftsDto struct {
	tableName    struct{}          `sql:"drafts" pg:",discard_unknown_columns"`
	Id           int               `sql:"id,pk"`
	AppId        int               `sql:"app_id,notnull"`
	EnvId        int               `sql:"env_id"`
	Resource     DraftResourceType `sql:"resource"`
	ResourceName string            `sql:"resource_name,notnull"`
	DraftState   DraftState        `sql:"draft_state"`
	sql.AuditLog
}

type DraftVersionDto struct {
	tableName struct{}       `sql:"draft_versions" pg:",discard_unknown_columns"`
	Id        int            `sql:"id,pk"`
	DraftId   int            `sql:"draft_id,notnull"`
	Data      string         `sql:"data"`
	Action    ResourceAction `sql:"action"`
	UserId    int32          `sql:"user_id"`
	CreatedOn time.Time      `sql:"created_on,type:timestamptz"`
	DraftsDto *DraftsDto
}

type DraftVersionCommentDto struct {
	tableName      struct{}  `sql:"draft_version_comments" pg:",discard_unknown_columns"`
	Id             int       `sql:"id,pk"`
	DraftId        int       `sql:"draft_id,notnull"`
	DraftVersionId int       `sql:"draft_version_id"`
	Comment        string    `sql:"comment"`
	UserId         int32     `sql:"user_id"`
	CreatedOn      time.Time `sql:"created_on,type:timestamptz"`
}

func (dto DraftsDto) ConvertToAppConfigDraft() AppConfigDraft {
	appConfigDraft := AppConfigDraft{
		DraftId:      dto.Id,
		Resource:     dto.Resource,
		ResourceName: dto.ResourceName,
		DraftState:   dto.DraftState,
	}
	return appConfigDraft
}

func (dto DraftVersionDto) ConvertToDraftVersionMetadata() DraftVersionMetadata {
	draftVersionMetadata := DraftVersionMetadata{
		DraftVersionId: dto.Id,
		UserId:         dto.UserId,
		ActivityTime:   dto.CreatedOn,
	}
	return draftVersionMetadata
}

func (dto DraftVersionDto) ConvertToConfigDraft() ConfigDraftResponse {
	configDraftResponse := ConfigDraftResponse{
		DraftId:        dto.DraftId,
		DraftVersionId: dto.Id,
	}
	configDraftResponse.Data = dto.Data
	configDraftResponse.Action = dto.Action
	configDraftResponse.UserId = dto.UserId
	if draftsDto := dto.DraftsDto; draftsDto != nil {
		configDraftResponse.AppId = draftsDto.AppId
		configDraftResponse.EnvId = draftsDto.EnvId
		configDraftResponse.Resource = draftsDto.Resource
		configDraftResponse.ResourceName = draftsDto.ResourceName
		configDraftResponse.DraftState = draftsDto.DraftState
	}
	return configDraftResponse
}

func (dto DraftVersionCommentDto) ConvertToDraftVersionComment() UserCommentMetadata {
	userComment := UserCommentMetadata{
		UserId:      dto.UserId,
		CommentedAt: dto.CreatedOn,
	}
	return userComment
}
