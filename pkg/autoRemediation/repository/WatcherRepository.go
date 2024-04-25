package repository

import (
	autoRemediation2 "github.com/devtron-labs/devtron/api/restHandler/autoRemediation"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type Watcher struct {
	tableName        struct{} `sql:"watcher" pg:",discard_unknown_columns"`
	Id               int      `sql:"id,pk"`
	Name             string   `sql:"name,notnull"`
	Desc             string   `sql:"desc"`
	FilterExpression string   `sql:"filter_expression,notnull"`
	Gvks             []string `sql:"gvks"`
	Active           bool     `sql:"active,notnull"`
	sql.AuditLog
}
type WatcherRepository interface {
	Save(watcher *Watcher, tx *pg.Tx) (*Watcher, error)
	Update(watcher *Watcher) (*Watcher, error)
	Delete(watcher *Watcher) error
	GetWatcherById(id int) (*Watcher, error)
	DeleteWatcherById(id int) error
	FindAllWatchersByQueryName(params autoRemediation2.WatcherQueryParams) ([]*Watcher, error)
	sql.TransactionWrapper
}
type WatcherRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
	*sql.TransactionUtilImpl
}

func NewWatcherRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *WatcherRepositoryImpl {
	TransactionUtilImpl := sql.NewTransactionUtilImpl(dbConnection)
	return &WatcherRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: TransactionUtilImpl,
	}
}

func (impl WatcherRepositoryImpl) Save(watcher *Watcher, tx *pg.Tx) (*Watcher, error) {
	_, err := tx.Model(watcher).Insert()
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return watcher, nil
}
func (impl WatcherRepositoryImpl) Update(watcher *Watcher) (*Watcher, error) {
	_, err := impl.dbConnection.Model(watcher).Update()
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return watcher, nil
}
func (impl WatcherRepositoryImpl) Delete(watcher *Watcher) error {
	err := impl.dbConnection.Delete(watcher)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
func (impl WatcherRepositoryImpl) GetWatcherById(id int) (*Watcher, error) {
	var watcher Watcher
	err := impl.dbConnection.Model(&watcher).Where("watcher_id = ? and active =?", id, true).Select()
	if err != nil {
		impl.logger.Error(err)
		return &Watcher{}, err
	}
	return &watcher, nil
}
func (impl WatcherRepositoryImpl) DeleteWatcherById(id int) error {
	var watcher Watcher
	err := impl.dbConnection.Model(watcher).Where("id = ?", id).Select()
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	watcher.Active = false
	_, err = impl.Update(&watcher)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
func (impl WatcherRepositoryImpl) FindAllWatchersByQueryName(params autoRemediation2.WatcherQueryParams) ([]*Watcher, error) {
	var watcher []*Watcher
	query := impl.dbConnection.Model(&watcher)
	if params.Search != "" {
		query = query.Where("name LIKE ?", "%"+params.Search+"%")
	}
	if params.SortOrderBy == "name" {
		query = query.Order("name ?", params.SortOrder)
	}
	err := query.Offset(params.Offset).Limit(params.Size).Select()
	if err != nil {
		return []*Watcher{}, err
	}

	return watcher, nil
}
