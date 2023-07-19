package protect

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ResourceProtectionRepository interface {
	ConfigureResourceProtection(appId int, envId int, state ProtectionState, userId int32) error
	GetResourceProtectMetadata(appId int) ([]*ResourceProtectionDto, error)
}

type ResourceProtectionDto struct {
	tableName struct{}        `sql:"resource_protection" pg:",discard_unknown_columns"` // make sure to make unique constraint on app_id, env_id, resource_type & state
	Id        int             `sql:"id,pk"`
	AppId     int             `sql:"app_id"`
	EnvId     int             `sql:"env_id"`
	Resource  ResourceType    `sql:"resource"`
	State     ProtectionState `sql:"protection_state"`
	sql.AuditLog
}

type ResourceProtectionHistoryDto struct {
	tableName struct{}        `sql:"resource_protection_history" pg:",discard_unknown_columns"`
	Id        int             `sql:"id,pk"`
	AppId     int             `sql:"app_id"`
	EnvId     int             `sql:"env_id"`
	Resource  ResourceType    `sql:"resource"`
	State     ProtectionState `sql:"protection_state"`
	UpdatedOn time.Time       `sql:"updated_on,type:timestamptz"`
	UpdatedBy int32           `sql:"updated_by,type:integer"`
}

type ResourceProtectionRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewResourceProtectionRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *ResourceProtectionRepositoryImpl {
	return &ResourceProtectionRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

func (repo *ResourceProtectionRepositoryImpl) ConfigureResourceProtection(appId int, envId int, state ProtectionState, userId int32) error {
	// check whether entry exists before or not
	// if exists then update and make entry in history table or else make a new entry
	protectionStateDto, err := repo.GetResourceProtectionState(appId, envId)
	if err != nil {
		return err
	}
	currentTime := time.Now()
	if protectionStateDto == nil {
		// data not found, make new entry in table
		protectionStateDto = &ResourceProtectionDto{
			AppId:    appId,
			EnvId:    envId,
			State:    state,
			Resource: ConfigProtectionResourceType,
		}
		protectionStateDto.CreatedOn = currentTime
		protectionStateDto.UpdatedOn = currentTime
		protectionStateDto.CreatedBy = userId
		protectionStateDto.UpdatedBy = userId
		_, err = repo.dbConnection.Model(protectionStateDto).Insert()
		if err != nil {
			repo.logger.Errorw("error occurred while inserting protection dto", "appId", appId, "envId", envId, "err", err)
			return err
		}
	} else {
		// make entry in history table and update current entry
		_, err = repo.createProtectionHistoryDto(protectionStateDto)
		if err != nil {
			return err
		}
		result, err := repo.dbConnection.Model(protectionStateDto).Set("protection_state = ?", state).
			Set("updated_on = ?", currentTime).Set("updated_by = ?", userId).
			Where("app_id = ?", appId).
			Where("env_id = ?", envId).
			Update()
		if err != nil {
			repo.logger.Errorw("error occurred while updating protection state", "appId", appId, "env_id", envId, "err", err)
			return err
		}
		if result.RowsAffected() == 0 {
			return errors.New("data-not-updated")
		}
	}
	return nil
}

func (repo *ResourceProtectionRepositoryImpl) GetResourceProtectionState(appId int, envId int) (*ResourceProtectionDto, error) {
	protectionDto := &ResourceProtectionDto{}
	err := repo.dbConnection.Model(protectionDto).Where("app_id = ?", appId).Where("env_id = ?", envId).
		Where("resource = ?", ConfigProtectionResourceType).Select()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		repo.logger.Errorw("error occurred while fetching resource protection data", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return protectionDto, nil
}

func (repo *ResourceProtectionRepositoryImpl) createProtectionHistoryDto(dto *ResourceProtectionDto) (*ResourceProtectionHistoryDto, error) {
	history := &ResourceProtectionHistoryDto{}
	envId := dto.EnvId
	appId := dto.AppId
	history.EnvId = envId
	history.AppId = appId
	history.State = dto.State
	history.Resource = dto.Resource
	history.UpdatedBy = dto.UpdatedBy
	history.UpdatedOn = dto.UpdatedOn
	_, err := repo.dbConnection.Model(history).Insert()
	if err != nil {
		repo.logger.Errorw("error occurred while creating history dto", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	return history, nil
}

func (repo *ResourceProtectionRepositoryImpl) GetResourceProtectMetadata(appId int) ([]*ResourceProtectionDto, error) {
	var resourceProtectionDtos []*ResourceProtectionDto
	err := repo.dbConnection.Model(&resourceProtectionDtos).Where("app_id = ?", appId).Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error occurred while fetching resource protection", "appId", appId, "err", err)
	} else {
		err = nil
	}
	return resourceProtectionDtos, err
}

