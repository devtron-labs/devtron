package repository

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CustomTag struct {
	tableName            struct{} `sql:"custom_tag" pg:",discard_unknown_columns"`
	Id                   int      `sql:"id"`
	EntityKey            int      `sql:"entity_key"`
	EntityValue          string   `sql:"entity_value"`
	TagPattern           string   `sql:"tag_pattern"`
	AutoIncreasingNumber int      `sql:"auto_increasing_number"`
	Metadata             string   `sql:"metadata"`
}

type ImagePathReservation struct {
	tableName   struct{} `sql:"image_path_reservation" pg:",discard_unknown_columns"`
	Id          int      `sql:"id"`
	ImagePath   string   `sql:"image_path"`
	CustomTagId int      `sql:"custom_tag_id"`
	active      bool     `sql:"active"`
}

type ImageTagRepository interface {
	GetConnection() *pg.DB
	CreateImageTag(customTagData *CustomTag) error
	FetchCustomTagData(entityType int, entityValue string) (*CustomTag, error)
	IncrementAndFetchByEntityKeyAndValue(tx *pg.Tx, entityKey int, entityValue string) (*CustomTag, error)
	FindByImagePath(tx *pg.Tx, path string) ([]*ImagePathReservation, error)
	InsertImagePath(tx *pg.Tx, reservation ImagePathReservation) error
	UpdateImageTag(customTag *CustomTag) error
	DeleteByEntityKeyAndValue(entityKey int, entityValue string) error
}

type ImageTagRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewImageTagRepository(dbConnection *pg.DB, logger *zap.SugaredLogger) *ImageTagRepositoryImpl {
	return &ImageTagRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *ImageTagRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl *ImageTagRepositoryImpl) CreateImageTag(customTagData *CustomTag) error {
	return impl.dbConnection.Insert(customTagData)
}

func (impl *ImageTagRepositoryImpl) UpdateImageTag(customTag *CustomTag) error {
	return impl.dbConnection.Update(customTag)
}

func (impl *ImageTagRepositoryImpl) DeleteByEntityKeyAndValue(entityKey int, entityValue string) error {
	query := `delete from table custom_tag where entity_key = ? and entity_value = ?`
	_, err := impl.dbConnection.Exec(query, entityKey, entityValue)
	return err
}

func (impl *ImageTagRepositoryImpl) FetchCustomTagData(entityType int, entityValue string) (*CustomTag, error) {
	var customTagData CustomTag
	err := impl.dbConnection.Model(&customTagData).
		Where("entity_type = ?", entityType).
		Where("entity_value = ?", entityValue).Select()
	return &customTagData, err
}

func (impl *ImageTagRepositoryImpl) IncrementAndFetchByEntityKeyAndValue(tx *pg.Tx, entityKey int, entityValue string) (*CustomTag, error) {
	var customTag CustomTag
	query := `update custom_tag set auto_increasing_number=auto_increasing_number+1 where entity_key=? and entity_value=? returning id, custom_tag_format, auto_increasing_number, metadata, entity_key, entity_value`
	_, err := tx.Query(&customTag, query, entityKey, entityValue)
	return &customTag, err
}

func (impl *ImageTagRepositoryImpl) FindByImagePath(tx *pg.Tx, path string) ([]*ImagePathReservation, error) {
	var imagePaths []*ImagePathReservation
	err := tx.Model(&imagePaths).
		Where("image_path = ?", path).
		Where("active = ?", true).Select()
	return imagePaths, err
}

func (impl *ImageTagRepositoryImpl) InsertImagePath(tx *pg.Tx, reservation ImagePathReservation) error {
	return tx.Insert(reservation)
}
