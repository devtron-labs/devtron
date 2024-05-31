/*
 * Copyright (c) 2024. Devtron Inc.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AutoRemediationTrigger struct {
	tableName struct{}    `sql:"auto_remediation_trigger" pg:",discard_unknown_columns"`
	Id        int         `sql:"id,pk"`
	Type      TriggerType `sql:"type"`
	WatcherId int         `sql:"watcher_id"`
	Data      string      `sql:"data"`
	Active    bool        `sql:"active,notnull"`
	sql.AuditLog
}
type TriggerType string

const (
	DEVTRON_JOB TriggerType = "DEVTRON_JOB"
)

type TriggerRepository interface {
	Save(trigger *AutoRemediationTrigger, tx *pg.Tx) (*AutoRemediationTrigger, error)
	SaveInBulk(trigger []*AutoRemediationTrigger, tx *pg.Tx) ([]*AutoRemediationTrigger, error)
	Update(trigger *AutoRemediationTrigger) (*AutoRemediationTrigger, error)
	GetTriggerByWatcherId(watcherId int) ([]*AutoRemediationTrigger, error)
	GetTriggerByWatcherIds(watcherIds []int) ([]*AutoRemediationTrigger, error)
	GetTriggerById(id int) (*AutoRemediationTrigger, error)
	DeleteTriggerByWatcherId(tx *pg.Tx, watcherId int) error
	GetWatcherByTriggerId(triggerId int) (*K8sEventWatcher, error)
	GetAllTriggers() ([]*AutoRemediationTrigger, error)
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

func (impl TriggerRepositoryImpl) Save(trigger *AutoRemediationTrigger, tx *pg.Tx) (*AutoRemediationTrigger, error) {
	_, err := tx.Model(trigger).Insert()
	if err != nil {
		return nil, err
	}
	return trigger, nil
}
func (impl TriggerRepositoryImpl) SaveInBulk(triggers []*AutoRemediationTrigger, tx *pg.Tx) ([]*AutoRemediationTrigger, error) {
	if len(triggers) == 0 {
		return nil, nil
	}
	_, err := tx.Model(&triggers).Insert()
	if err != nil {
		return nil, err
	}
	return triggers, nil
}

func (impl TriggerRepositoryImpl) Update(trigger *AutoRemediationTrigger) (*AutoRemediationTrigger, error) {
	_, err := impl.dbConnection.Model(&trigger).Update()
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (impl TriggerRepositoryImpl) GetTriggerByWatcherId(watcherId int) ([]*AutoRemediationTrigger, error) {
	var trigger []*AutoRemediationTrigger
	err := impl.dbConnection.Model(&trigger).
		Where("watcher_id = ? and active =?", watcherId, true).
		Select()
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (impl TriggerRepositoryImpl) GetTriggerByWatcherIds(watcherIds []int) ([]*AutoRemediationTrigger, error) {
	var trigger []*AutoRemediationTrigger
	if len(watcherIds) == 0 {
		return nil, nil
	}
	err := impl.dbConnection.Model(&trigger).
		Where(" watcher_id IN (?) ", pg.In(watcherIds)).
		Where(" active = ? ", true).
		Select()
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (impl TriggerRepositoryImpl) DeleteTriggerByWatcherId(tx *pg.Tx, watcherId int) error {
	_, err := tx.Model((*AutoRemediationTrigger)(nil)).
		Set("active = ?", false).
		Where("watcher_id = ?", watcherId).
		Update()
	return err
}

func (impl TriggerRepositoryImpl) GetTriggerById(id int) (*AutoRemediationTrigger, error) {
	var trigger AutoRemediationTrigger
	err := impl.dbConnection.Model(&trigger).
		Where("id = ? and active =?", id, true).
		Select()
	if err != nil {
		return &AutoRemediationTrigger{}, err
	}
	return &trigger, nil
}

func (impl TriggerRepositoryImpl) GetWatcherByTriggerId(triggerId int) (*K8sEventWatcher, error) {
	var trigger AutoRemediationTrigger
	err := impl.dbConnection.Model(&trigger).Where("id = ? and active =?", triggerId, true).Select()
	if err != nil {
		impl.logger.Error(err)
		return &K8sEventWatcher{}, err
	}
	var watcher K8sEventWatcher
	err = impl.dbConnection.Model(&watcher).Where("id = ? and active =?", trigger.WatcherId, true).Select()
	if err != nil {
		return &K8sEventWatcher{}, err
	}
	return &watcher, nil
}
func (impl TriggerRepositoryImpl) GetAllTriggers() ([]*AutoRemediationTrigger, error) {
	var trigger []*AutoRemediationTrigger
	err := impl.dbConnection.Model(&trigger).Where("active =?", true).Select()
	if err != nil {
		return trigger, err
	}
	return trigger, nil
}
