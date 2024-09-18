package repository

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GenericNoteFileBasedRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewGenericNoteFileBasedRepository(connection *sql.SqliteConnection, logger *zap.SugaredLogger) *GenericNoteFileBasedRepositoryImpl {
	genericNote := &GenericNote{}
	connection.Migrator.MigrateEntities(genericNote)
	logger.Debugw("generic note repository file based initialized")
	return &GenericNoteFileBasedRepositoryImpl{logger, connection.DbConnection}
}

func (impl GenericNoteFileBasedRepositoryImpl) StartTx() (*pg.Tx, error) {
	return nil, nil
}

func (impl GenericNoteFileBasedRepositoryImpl) RollbackTx(tx *pg.Tx) error {
	return nil
}

func (impl GenericNoteFileBasedRepositoryImpl) CommitTx(tx *pg.Tx) error {
	return nil
}

func (impl GenericNoteFileBasedRepositoryImpl) Save(tx *pg.Tx, model *GenericNote) error {
	result := impl.dbConnection.Model(model).Create(model)
	return result.Error
}

func (impl GenericNoteFileBasedRepositoryImpl) FindByClusterId(id int) (*GenericNote, error) {
	return impl.FindByIdentifier(id, ClusterType)
}

func (impl GenericNoteFileBasedRepositoryImpl) FindByAppId(id int) (*GenericNote, error) {
	return impl.FindByIdentifier(id, AppType)
}

func (impl GenericNoteFileBasedRepositoryImpl) FindByIdentifier(identifier int, identifierType NoteType) (*GenericNote, error) {
	clusterNote := &GenericNote{}
	result := impl.dbConnection.
		Model(clusterNote).
		Where("identifier =?", identifier).
		Where("identifier_type =?", identifierType).
		First(clusterNote)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	}
	return clusterNote, err
}

func (impl GenericNoteFileBasedRepositoryImpl) Update(tx *pg.Tx, model *GenericNote) error {
	result := impl.dbConnection.Model(model).Updates(model)
	return result.Error
}

func (impl GenericNoteFileBasedRepositoryImpl) GetGenericNotesForAppIds(appIds []int) ([]*GenericNote, error) {
	notes := make([]*GenericNote, 0)
	if len(appIds) == 0 {
		return notes, nil
	}
	result := impl.dbConnection.
		Model(&notes).
		Where("identifier IN ?", appIds).
		Where("identifier_type =?", AppType).
		First(&notes)
	return notes, result.Error
}

func (impl GenericNoteFileBasedRepositoryImpl) GetDescriptionFromAppIds(appIds []int) ([]*GenericNote, error) {
	notes := make([]*GenericNote, 0, len(appIds))
	return notes, nil
}
