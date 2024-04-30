package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type Watcher struct {
	tableName        struct{} `sql:"watcher" pg:",discard_unknown_columns"`
	Id               int      `sql:"id,pk"`
	Name             string   `sql:"name,notnull"`
	Description      string   `sql:"description"`
	FilterExpression string   `sql:"filter_expression,notnull"`
	Gvks             string   `sql:"gvks"`
	Active           bool     `sql:"active,notnull"`
	sql.AuditLog
}
type WatcherQueryParams struct {
	Offset      int    `json:"offset"`
	Size        int    `json:"size"`
	SortOrder   string `json:"sortOrder"`
	SortOrderBy string `json:"sortOrderBy"`
	Search      string `json:"Search"`
}

type WatcherRepository interface {
	Save(watcher *Watcher, tx *pg.Tx) (*Watcher, error)
	Update(tx *pg.Tx, watcher *Watcher, userId int32) error
	Delete(watcher *Watcher) error
	GetWatcherById(id int) (*Watcher, error)
	DeleteWatcherById(id int) error
	FindAllWatchersByQueryName(params WatcherQueryParams) ([]*Watcher, int, error)
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
func (impl WatcherRepositoryImpl) Update(tx *pg.Tx, watcher *Watcher, userId int32) error {
	_, err := tx.
		Model((*Watcher)(nil)).
		Set("name = ?", watcher.Name).
		Set("description = ?", watcher.Description).
		Set("filter_expression = ?", watcher.FilterExpression).
		Set("gvks = ?", watcher.Gvks).
		Set("updated_by = ?", userId).
		Set("updated_on = ?", time.Now()).
		Where("active = ?", true).
		Where("id = ?", watcher.Id).
		Update()
	return err
}

func (impl WatcherRepositoryImpl) Delete(watcher *Watcher) error {
	err := impl.dbConnection.Delete(&watcher)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
func (impl WatcherRepositoryImpl) GetWatcherById(id int) (*Watcher, error) {
	var watcher Watcher
	err := impl.dbConnection.Model(&watcher).Where("id = ? and active = ?", id, true).Select()
	if err != nil {
		impl.logger.Error(err)
		return &Watcher{}, err
	}
	return &watcher, nil
}
func (impl WatcherRepositoryImpl) DeleteWatcherById(id int) error {
	_, err := impl.dbConnection.Model(&Watcher{}).Set("active = ?", false).Where("id = ?", id).Update()
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl WatcherRepositoryImpl) FindAllWatchersByQueryName(params WatcherQueryParams) ([]*Watcher, int, error) {
	var watcher []*Watcher
	query := impl.dbConnection.Model(&watcher)
	if params.Search != "" {
		query = query.Where("name ILIKE ? ", "%"+params.Search+"%")
	}
	if params.SortOrderBy == "name" {
		if params.SortOrder == "desc" {
			query = query.Order("name desc")
		} else {
			query = query.Order("name asc")
		}
	}
	// Count total number of watchers
	total, err := query.Count()
	if err != nil {
		return []*Watcher{}, 0, err
	}
	err = query.Where("active = ?", true).Offset(params.Offset).Limit(params.Size).Select()
	if err != nil {
		return []*Watcher{}, 0, err
	}
	return watcher, total, nil
}
