package repository

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type Trigger struct {
	tableName struct{}        `sql:"trigger" pg:",discard_unknown_columns"`
	Id        int             `sql:"id,pk"`
	Type      TriggerType     `sql:"type"`
	WatcherId string          `sql:"watcher_id"`
	Data      json.RawMessage `sql:"data"`
	sql.AuditLog
}
type TriggerType string

const (
	DEVTRON_JOB TriggerType = "DEVTRON_JOB"
)

type TriggerRepository interface {
	Save(trigger *Trigger, tx *pg.Tx) (*Trigger, error)
	Update(trigger *Trigger) (*Trigger, error)
	Delete(trigger *Trigger) error
	sql.TransactionWrapper
}
type TriggerRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
	*sql.TransactionUtilImpl
}

func NewTriggerRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *TriggerRepositoryImpl {
	TransactionUtilImpl := sql.NewTransactionUtilImpl(dbConnection)
	return &TriggerRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: TransactionUtilImpl,
	}
}

func (impl TriggerRepositoryImpl) Save(trigger *Trigger, tx *pg.Tx) (*Trigger, error) {
	_, err := tx.Model(trigger).Insert()
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return trigger, nil
}
func (impl TriggerRepositoryImpl) Update(trigger *Trigger) (*Trigger, error) {
	_, err := impl.dbConnection.Model(trigger).Update()
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return trigger, nil
}
func (impl TriggerRepositoryImpl) Delete(trigger *Trigger) error {
	err := impl.dbConnection.Delete(trigger)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
