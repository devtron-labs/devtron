package drafts

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type DraftMetadataDto struct {
	tableName    struct{}          `sql:"draft_metadata" pg:",discard_unknown_columns"`
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
	DraftId   int            `sql:"draft_id"`
	Data      string         `sql:"data"`
	Action    ResourceAction `sql:"action"`
	UserId    int32          `sql:"user_id"`
	CreatedOn time.Time      `sql:"created_on,type:timestamptz"`
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

func (dto DraftMetadataDto) ConvertToAppConfigDraft() AppConfigDraft {
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
	return configDraftResponse
}

func (dto DraftVersionCommentDto) ConvertToDraftVersionComment() UserCommentMetadata {
	userComment := UserCommentMetadata{
		UserId:      dto.UserId,
		CommentedAt: dto.CreatedOn,
	}
	return userComment
}

type ConfigDraftRepository interface {
	CreateConfigDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error)
	GetLatestDraftVersionId(draftId int) (int, error)
	SaveDraftVersionComment(draftVersionComment DraftVersionCommentDto) error
	SaveDraftVersion(draftVersionDto DraftVersionDto) (int, error)
	GetDraftMetadataById(draftId int) (*DraftMetadataDto, error)
	UpdateDraftState(draftId int, draftState DraftState, userId int32) error
	GetDraftVersionsMetadata(draftId int) ([]*DraftVersionDto, error)
	GetDraftVersionComments(draftId int) ([]*DraftVersionCommentDto, error)
	GetLatestConfigDraft(draftId int) (*DraftVersionDto, error)
	GetDraftMetadata(appId int, envId int, resourceType DraftResourceType) ([]*DraftMetadataDto, error)
}

type ConfigDraftRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewConfigDraftRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *ConfigDraftRepositoryImpl {
	return &ConfigDraftRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

func (repo *ConfigDraftRepositoryImpl) CreateConfigDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error) {
	metadataDto := &DraftMetadataDto{
		AppId:        request.AppId,
		EnvId:        request.EnvId,
		Resource:     request.Resource,
		ResourceName: request.ResourceName,
		DraftState:   InitDraftState,
	}
	currentTime := time.Now()
	metadataDto.CreatedOn = currentTime
	metadataDto.UpdatedOn = currentTime
	metadataDto.CreatedBy = request.UserId
	metadataDto.UpdatedBy = request.UserId
	err := repo.dbConnection.Insert(metadataDto)
	if err != nil {
		repo.logger.Errorw("error occurred while creating config draft", "err", err, "metadataDto", metadataDto)
		return nil, err
	}

	draftMetadataId := metadataDto.Id
	repo.logger.Debugw("going to save draft version now", "draftId", draftMetadataId)

	draftVersionDto := DraftVersionDto{
		DraftId:   draftMetadataId,
		Action:    request.Action,
		Data:      request.Data,
		UserId:    request.UserId,
		CreatedOn: currentTime,
	}

	draftVersionId, err := repo.SaveDraftVersion(draftVersionDto)
	if len(request.UserComment) > 0 {
		draftVersionCommentDto := DraftVersionCommentDto{}
		draftVersionCommentDto.DraftId = draftMetadataId
		draftVersionCommentDto.DraftVersionId = draftVersionId
		draftVersionCommentDto.Comment = request.UserComment
		draftVersionCommentDto.UserId = request.UserId
		err = repo.SaveDraftVersionComment(draftVersionCommentDto)
		if err != nil {
			return nil, err
		}
	}

	return &ConfigDraftResponse{DraftId: draftMetadataId, DraftVersionId: draftVersionId}, nil
}

func (repo *ConfigDraftRepositoryImpl) GetLatestDraftVersionId(draftId int) (int, error) {
	draftVersionDto := &DraftVersionDto{}
	err := repo.dbConnection.Model(draftVersionDto).Column("id").Where("draft_id = ?", draftId).
		Order("id desc").Limit(1).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			repo.logger.Errorw("no draft version found ", "draftId", draftId)
			return 0, nil
		} else {
			repo.logger.Errorw("error occurred while fetching latest draft version", "draftId", draftId, "err", err)
			return 0, err
		}
	}
	return draftVersionDto.Id, nil
}

func (repo *ConfigDraftRepositoryImpl) SaveDraftVersionComment(draftVersionComment DraftVersionCommentDto) error {
	draftVersionComment.CreatedOn = time.Now()
	err := repo.dbConnection.Insert(&draftVersionComment)
	if err != nil {
		repo.logger.Errorw("error occurred while saving draft version comment", "draftVersionId", draftVersionComment.DraftVersionId, "err", err)
	}
	return err
}

func (repo *ConfigDraftRepositoryImpl) SaveDraftVersion(draftVersionDto DraftVersionDto) (int, error) {
	draftVersionDto.CreatedOn = time.Now()
	err := repo.dbConnection.Insert(&draftVersionDto)
	if err != nil {
		repo.logger.Errorw("error occurred while saving draft version comment", "draftMetadataId", draftVersionDto.DraftId, "err", err)
	}
	return draftVersionDto.Id, err

}

func (repo *ConfigDraftRepositoryImpl) GetDraftMetadataById(draftId int) (*DraftMetadataDto, error) {
	draftMetadataDto := &DraftMetadataDto{}
	err := repo.dbConnection.Model(draftMetadataDto).Where("id = ?", draftId).Select()
	if err != nil {
		repo.logger.Errorw("error occurred while fetching draft metadata", "draftId", draftId, "err", err)
		return nil, err
	}
	return draftMetadataDto, err
}

func (repo *ConfigDraftRepositoryImpl) UpdateDraftState(draftId int, draftState DraftState, userId int32) error {
	draftMetadataDto := &DraftMetadataDto{}
	result, err := repo.dbConnection.Model(draftMetadataDto).Set("draft_state", draftState).Set("updated_on", time.Now()).
		Set("updated_by", userId).Where("id = ?", draftId).Update()
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("no-record-found")
	}
	return nil
}

func (repo *ConfigDraftRepositoryImpl) GetDraftVersionsMetadata(draftId int) ([]*DraftVersionDto, error) {
	var draftVersions []*DraftVersionDto
	err := repo.dbConnection.Model(&draftVersions).Column("id, user_id, created_on").Where("draft_id = ?", draftId).
		Order("id desc").Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft versions", "draftId", draftId, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return draftVersions, err
}

func (repo *ConfigDraftRepositoryImpl) GetDraftVersionComments(draftId int) ([]*DraftVersionCommentDto, error) {
	var draftComments []*DraftVersionCommentDto
	err := repo.dbConnection.Model(&draftComments).Where("draft_id = ?", draftId).
		Order("id desc").Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft comments", "draftId", draftId, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return draftComments, err
}

func (repo *ConfigDraftRepositoryImpl) GetLatestConfigDraft(draftId int) (*DraftVersionDto, error) {
	var draftVersion *DraftVersionDto
	err := repo.dbConnection.Model(draftVersion).Where("draft_id = ?", draftId).
		Order("id desc").Limit(1).Select()
	if err != nil {
		repo.logger.Errorw("error occurred while fetching latest draft version", "draftId", draftId, "err", err)
		return nil, err
	}
	return draftVersion, nil
}

func (repo *ConfigDraftRepositoryImpl) GetDraftMetadata(appId int, envId int, resourceType DraftResourceType) ([]*DraftMetadataDto, error) {
	var draftMetadataDtos []*DraftMetadataDto
	err := repo.dbConnection.Model(&draftMetadataDtos).Where("app_id = ?", appId).Where("env_id = ?", envId).
		Where("resource = ?", resourceType).Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft metadata", "appId", appId, "envId", envId, "resourceType", resourceType, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return draftMetadataDtos, err
}



