package drafts

import (
	"errors"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ConfigDraftRepository interface {
	CreateConfigDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error)
	GetLatestDraftVersion(draftId int) (*DraftVersion, error)
	SaveDraftVersionComment(draftVersionComment *DraftVersionComment) error
	SaveDraftVersion(draftVersionDto *DraftVersion) (int, error)
	GetDraftMetadataById(draftId int) (*DraftDto, error)
	UpdateDraftState(draftId int, draftState DraftState, userId int32) error
	GetDraftVersionsMetadata(draftId int) ([]*DraftVersion, error)
	GetDraftVersionComments(draftId int) ([]*DraftVersionComment, error)
	GetDraftVersionCommentsCount(draftId int) (int, error)
	GetLatestConfigDraft(draftId int) (*DraftVersion, error)
	GetLatestConfigDraftByName(appId, envId int, resourceName string, resourceType DraftResourceType) (*DraftVersion, error)
	GetDraftMetadataForAppAndEnv(appId int, envIds []int) ([]*DraftDto, error)
	GetDraftMetadata(appId int, envId int, resourceType DraftResourceType) ([]*DraftDto, error)
	GetDraftVersionById(draftVersionId int) (*DraftVersion, error)
	DeleteComment(draftId int, draftCommentId int, userId int32) (int, error)
	DiscardDrafts(appId int, envId int, userId int32) error
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
	// check draft already exists for this name
	exists, err := repo.checkDraftAlreadyExists(request)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("draft-exists-already")
	}
	draftState := InitDraftState
	metadataDto := request.GetDraftDto()
	err = repo.dbConnection.Insert(metadataDto)
	if err != nil {
		repo.logger.Errorw("error occurred while creating config draft", "err", err, "metadataDto", metadataDto)
		return nil, err
	}

	draftMetadataId := metadataDto.Id
	repo.logger.Debugw("going to save draft version now", "draftId", draftMetadataId)

	draftVersionDto := request.GetDraftVersionDto(draftMetadataId, metadataDto.CreatedOn)
	draftVersionId, err := repo.SaveDraftVersion(draftVersionDto)
	if err != nil {
		return nil, err
	}

	if len(request.UserComment) > 0 {
		draftVersionCommentDto := request.GetDraftVersionComment(draftMetadataId, draftVersionId, metadataDto.CreatedOn)
		err = repo.SaveDraftVersionComment(draftVersionCommentDto)
		if err != nil {
			return nil, err
		}
	}

	return &ConfigDraftResponse{DraftId: draftMetadataId, DraftVersionId: draftVersionId, DraftState: draftState}, nil
}

func (repo *ConfigDraftRepositoryImpl) GetLatestDraftVersion(draftId int) (*DraftVersion, error) {
	draftVersionDto := &DraftVersion{}
	err := repo.dbConnection.Model(draftVersionDto).Column("draft_version.*", "Draft").Where("draft_id = ?", draftId).
		Order("id desc").Limit(1).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			repo.logger.Errorw("no draft version found ", "draftId", draftId)
		} else {
			repo.logger.Errorw("error occurred while fetching latest draft version", "draftId", draftId, "err", err)
		}
	}
	return draftVersionDto, err
}

func (repo *ConfigDraftRepositoryImpl) SaveDraftVersionComment(draftVersionComment *DraftVersionComment) error {
	draftVersionComment.CreatedOn = time.Now()
	err := repo.dbConnection.Insert(draftVersionComment)
	if err != nil {
		repo.logger.Errorw("error occurred while saving draft version comment", "draftVersionId", draftVersionComment.DraftVersionId, "err", err)
	}
	return err
}

func (repo *ConfigDraftRepositoryImpl) SaveDraftVersion(draftVersionDto *DraftVersion) (int, error) {
	draftVersionDto.CreatedOn = time.Now()
	err := repo.dbConnection.Insert(draftVersionDto)
	if err != nil {
		repo.logger.Errorw("error occurred while saving draft version comment", "draftMetadataId", draftVersionDto.DraftsId, "err", err)
	}
	return draftVersionDto.Id, err

}

func (repo *ConfigDraftRepositoryImpl) GetDraftMetadataById(draftId int) (*DraftDto, error) {
	draftMetadataDto := &DraftDto{}
	err := repo.dbConnection.Model(draftMetadataDto).Where("id = ?", draftId).Select()
	if err != nil {
		repo.logger.Errorw("error occurred while fetching draft metadata", "draftId", draftId, "err", err)
		return nil, err
	}
	return draftMetadataDto, err
}

func (repo *ConfigDraftRepositoryImpl) UpdateDraftState(draftId int, draftState DraftState, userId int32) error {
	draftMetadataDto := &DraftDto{}
	result, err := repo.dbConnection.Model(draftMetadataDto).Set("draft_state = ?", draftState).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", userId).Where("id = ?", draftId).Update()
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("no-record-found")
	}
	return nil
}

func (repo *ConfigDraftRepositoryImpl) GetDraftVersionsMetadata(draftId int) ([]*DraftVersion, error) {
	var draftVersions []*DraftVersion
	err := repo.dbConnection.Model(&draftVersions).Column("id", "user_id", "created_on").Where("draft_id = ?", draftId).
		Order("id desc").Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft versions", "draftId", draftId, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return draftVersions, err
}

func (repo *ConfigDraftRepositoryImpl) GetDraftVersionComments(draftId int) ([]*DraftVersionComment, error) {
	var draftComments []*DraftVersionComment
	err := repo.dbConnection.Model(&draftComments).
		Where("draft_id = ?", draftId).
		Where("active = ?", true).
		Order("id desc").Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft comments", "draftId", draftId, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return draftComments, err
}

func (repo *ConfigDraftRepositoryImpl) GetDraftVersionCommentsCount(draftId int) (int, error) {
	count, err := repo.dbConnection.Model(&DraftVersionComment{}).
		Where("draft_id = ?", draftId).
		Where("active = ?", true).
		Order("id desc").Count()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft comments", "draftId", draftId, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return count, err
}

func (repo *ConfigDraftRepositoryImpl) GetLatestConfigDraftByName(appId, envId int, resourceName string, resourceType DraftResourceType) (*DraftVersion, error) {
	draftVersion := &DraftVersion{}
	err := repo.dbConnection.Model(draftVersion).Column("draft_version.*", "Draft").
		//Join("INNER JOIN draft ON draft.id = draft_version.draft_id").
		Where("draft.app_id = ?", appId).
		Where("draft.env_id = ?", envId).
		Where("draft.resource_name = ?", resourceName).
		Where("draft.resource = ?", resourceType).
		Order("draft_version.id desc").Limit(1).Select()
	if err != nil {
		repo.logger.Errorw("error occurred while fetching latest draft version", "resourceName", resourceName,
			"resourceType", resourceType, "err", err)
		return nil, err
	}
	return draftVersion, nil
}

func (repo *ConfigDraftRepositoryImpl) GetLatestConfigDraft(draftId int) (*DraftVersion, error) {
	draftVersion := &DraftVersion{}
	err := repo.dbConnection.Model(draftVersion).Column("draft_version.*", "Draft").Where("draft_id = ?", draftId).
		Order("id desc").Limit(1).Select()
	if err != nil {
		repo.logger.Errorw("error occurred while fetching latest draft version", "draftId", draftId, "err", err)
		return nil, err
	}
	return draftVersion, nil
}

func (repo *ConfigDraftRepositoryImpl) GetDraftMetadataForAppAndEnv(appId int, envIds []int) ([]*DraftDto, error) {
	var draftMetadataDtos []*DraftDto
	err := repo.dbConnection.Model(&draftMetadataDtos).Where("app_id = ?", appId).Where("env_id in (?)", pg.In(envIds)).
		Where("draft_state in (?)", pg.In(GetNonTerminalDraftStates())).Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft metadata", "appId", appId, "envIds", envIds, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return draftMetadataDtos, err
}

func (repo *ConfigDraftRepositoryImpl) GetDraftMetadata(appId int, envId int, resourceType DraftResourceType) ([]*DraftDto, error) {
	var draftMetadataDtos []*DraftDto
	err := repo.dbConnection.Model(&draftMetadataDtos).Where("app_id = ?", appId).Where("env_id = ?", envId).
		Where("resource = ?", resourceType).Where("draft_state in (?)", pg.In(GetNonTerminalDraftStates())).Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching draft metadata", "appId", appId, "envId", envId, "resourceType", resourceType, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return draftMetadataDtos, err
}

func (repo *ConfigDraftRepositoryImpl) GetDraftVersionById(draftVersionId int) (*DraftVersion, error) {
	var draftVersion *DraftVersion
	err := repo.dbConnection.Model(draftVersion).Column("draft_version.*", "Draft").Where("id = ?", draftVersionId).
		Order("id desc").Select()
	if err != nil {
		repo.logger.Errorw("error occurred while fetching draft version", "draftVersionId", draftVersionId, "err", err)
		return nil, err
	}
	return draftVersion, nil
}

func (repo *ConfigDraftRepositoryImpl) DeleteComment(draftId int, draftCommentId int, userId int32) (int, error) {
	draftVersionComment := &DraftVersionComment{}
	result, err := repo.dbConnection.Model(draftVersionComment).Set("active = ?", false).Set("updated_on = ?", time.Now()).
		Where("id = ?", draftCommentId).
		Where("draft_id = ?", draftId).
		Where("created_by = ?", userId).
		Update()
	if err != nil {
		repo.logger.Errorw("error occurred while deleting draft", "draftId", draftId, "draftCommentId", draftCommentId, "err", err)
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (repo *ConfigDraftRepositoryImpl) DiscardDrafts(appId int, envId int, userId int32) error {
	draftsDto := &DraftDto{}
	_, err := repo.dbConnection.Model(draftsDto).Set("draft_state = ?", DiscardedDraftState).
		Set("updated_on = ?", time.Now()).Set("updated_by = ?", userId).
		Where("app_id = ?", appId).Where("env_id = ?", envId).
		Where("draft_state in (?)", pg.In(GetNonTerminalDraftStates())).
		Update()
	if err != nil {
		repo.logger.Errorw("error occurred while discarding drafts", "appId", appId, "envId", envId, "err", err)
	}
	return err
}

func (repo *ConfigDraftRepositoryImpl) checkDraftAlreadyExists(request ConfigDraftRequest) (bool, error) {
	resourceName := request.ResourceName
	resourceType := request.Resource
	appId := request.AppId
	envId := request.EnvId
	count, err := repo.dbConnection.Model(&DraftDto{}).
		Where("resource = ?", resourceType).
		Where("resource_name = ?", resourceName).
		Where("app_id = ?", appId).
		Where("env_id = ?", envId).
		Where("draft_state in (?)", pg.In(GetNonTerminalDraftStates())).
		Count()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while checking draft already exists", "appId", appId, "envId", envId, "err", err)
	} else {
		err = nil //ignoring noRows Error
	}
	return count > 0, err
}






