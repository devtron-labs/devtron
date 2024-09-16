package repository

import (
	"github.com/glebarez/sqlite"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GenericNoteHistoryFileBasedRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewGenericNoteHistoryFileBasedRepositoryImpl(logger *zap.SugaredLogger) *GenericNoteHistoryFileBasedRepositoryImpl {
	err, dbPath := createOrCheckClusterDbPath(logger)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	//db, err := sql.Open("sqlite3", "./cluster.db")
	if err != nil {
		logger.Fatal("error occurred while opening db connection", "error", err)
	}
	migrator := db.Migrator()
	genericNoteHistory := &GenericNoteHistory{}
	err = migrator.AutoMigrate(genericNoteHistory)
	if err != nil {
		logger.Fatal("error occurred while auto-migrating genericNoteHistory", "error", err)
	}
	logger.Debugw("generic note history repository file based initialized")
	return &GenericNoteHistoryFileBasedRepositoryImpl{logger, db}
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) StartTx() (*pg.Tx, error) {
	return nil, nil
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) RollbackTx(tx *pg.Tx) error {
	return nil
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) CommitTx(tx *pg.Tx) error {
	return nil
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) SaveHistory(tx *pg.Tx, model *GenericNoteHistory) error {
	result := impl.dbConnection.Create(model)
	return result.Error
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) FindHistoryByNoteId(id []int) ([]GenericNoteHistory, error) {
	var clusterNoteHistories []GenericNoteHistory
	result := impl.dbConnection.
		Where("note_id =?", id).
		Find(&clusterNoteHistories)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	}
	return clusterNoteHistories, err
}

