package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GenericNoteHistoryFileBasedRepositoryImpl struct {
	*sql.NoopTransactionUtilImpl
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewGenericNoteHistoryFileBasedRepositoryImpl(connection *sql.SqliteConnection, logger *zap.SugaredLogger, transactionWrapper *sql.NoopTransactionUtilImpl) *GenericNoteHistoryFileBasedRepositoryImpl {
	genericNoteHistory := &GenericNoteHistory{}
	connection.Migrator.MigrateEntities(genericNoteHistory)
	logger.Debugw("generic note history repository file based initialized")
	return &GenericNoteHistoryFileBasedRepositoryImpl{transactionWrapper, logger, connection.DbConnection}
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

