package repository

import (
	"errors"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/glebarez/sqlite"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"path"
)

type GenericNoteFileBasedRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func createOrCheckClusterDbPath(logger *zap.SugaredLogger) (error, string) {
	err, devtronDirPath := util2.CheckOrCreateDevtronDir()
	if err != nil {
		logger.Errorw("error occurred while creating devtron dir ", "err", err)
		return err, ""
	}
	clusterDbPath := path.Join(devtronDirPath, "./client.db")
	return nil, clusterDbPath
}

func NewGenericNoteFileBasedRepository(logger *zap.SugaredLogger) *GenericNoteFileBasedRepositoryImpl {
	err, dbPath := createOrCheckClusterDbPath(logger)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	//db, err := sql.Open("sqlite3", "./cluster.db")
	if err != nil {
		logger.Fatal("error occurred while opening db connection", "error", err)
	}
	migrator := db.Migrator()
	genericNote := &GenericNote{}
	err = migrator.AutoMigrate(genericNote)
	if err != nil {
		logger.Fatal("error occurred while auto-migrating genericNote", "error", err)
	}
	logger.Debugw("generic note repository file based initialized")
	return &GenericNoteFileBasedRepositoryImpl{logger, db}
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
