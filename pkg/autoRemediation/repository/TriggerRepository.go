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
	WatcherId int             `sql:"watcher_id"`
	Data      json.RawMessage `sql:"data"`
	Active    bool            `sql:"active,notnull"`
	sql.AuditLog
}
type TriggerType string

const (
	DEVTRON_JOB TriggerType = "DEVTRON_JOB"
)

type TriggerRepository interface {
	Save(trigger *Trigger, tx *pg.Tx) (*Trigger, error)
	SaveInBulk(trigger []*Trigger, tx *pg.Tx) ([]*Trigger, error)
	Update(trigger *Trigger) (*Trigger, error)
	Delete(trigger *Trigger) error
	GetTriggerByWatcherId(watcherId int) ([]*Trigger, error)
	GetTriggerByWatcherIds(watcherIds []int) ([]*Trigger, error)
	GetTriggerById(id int) (*Trigger, error)
	DeleteTriggerByWatcherId(watcherId int) error
	GetWatcherByTriggerId(triggerId int) (*Watcher, error)
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
func (impl TriggerRepositoryImpl) SaveInBulk(triggers []*Trigger, tx *pg.Tx) ([]*Trigger, error) {
	_, err := tx.Model(&triggers).Insert()
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return triggers, nil
}

func (impl TriggerRepositoryImpl) Update(trigger *Trigger) (*Trigger, error) {
	_, err := impl.dbConnection.Model(&trigger).Update()
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return trigger, nil
}

func (impl TriggerRepositoryImpl) Delete(trigger *Trigger) error {
	err := impl.dbConnection.Delete(&trigger)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl TriggerRepositoryImpl) GetTriggerByWatcherId(watcherId int) ([]*Trigger, error) {
	var trigger []*Trigger
	err := impl.dbConnection.Model(&trigger).
		Where("watcher_id = ? and active =?", watcherId, true).
		Select()
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (impl TriggerRepositoryImpl) GetTriggerByWatcherIds(watcherIds []int) ([]*Trigger, error) {
	var trigger []*Trigger
	err := impl.dbConnection.Model(&trigger).
		Where(" watcher_id IN (?) ", pg.In(watcherIds)).
		Where(" active = ? ", true).
		Select()
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (impl TriggerRepositoryImpl) DeleteTriggerByWatcherId(watcherId int) error {
	var trigger []*Trigger
	_, err := impl.dbConnection.Model(&trigger).Set("active = ?", false).Where("watcher_id = ?", watcherId).Update()
	if err != nil {
		impl.logger.Error(err)
		return err
	}

	return nil
}

func (impl TriggerRepositoryImpl) GetTriggerById(id int) (*Trigger, error) {
	var trigger Trigger
	err := impl.dbConnection.Model(&trigger).Where("id = ? and active =?", id, true).Select()
	if err != nil {
		impl.logger.Error(err)
		return &Trigger{}, err
	}
	return &trigger, nil
}
func (impl TriggerRepositoryImpl) GetWatcherByTriggerId(triggerId int) (*Watcher, error) {
	var trigger Trigger
	err := impl.dbConnection.Model(&trigger).Where("id = ? and active =?", triggerId, true).Select()
	if err != nil {
		impl.logger.Error(err)
		return &Watcher{}, err
	}
	var watcher Watcher
	err = impl.dbConnection.Model(&watcher).Where("id = ? and active =?", trigger.WatcherId, true).Select()
	if err != nil {
		impl.logger.Error(err)
		return &Watcher{}, err
	}
	return &watcher, nil
}
