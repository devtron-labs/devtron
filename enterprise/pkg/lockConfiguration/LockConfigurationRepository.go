package lockConfiguration

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/go-pg/pg"
	"time"
)

type LockConfigurationRepository interface {
	GetConnection() *pg.DB
	GetLockConfig(int) (*bean.LockConfiguration, error)
	GetActiveLockConfig() (*bean.LockConfiguration, error)
	DeleteActiveLockConfigs(userId int) error
	Create(*bean.LockConfiguration, *pg.Tx) error
	Update(*bean.LockConfiguration, *pg.Tx) error
}

type RepositoryImpl struct {
	dbConnection *pg.DB
}

func NewRepositoryImpl(dbConnection *pg.DB) *RepositoryImpl {
	return &RepositoryImpl{dbConnection: dbConnection}
}

func (impl RepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl RepositoryImpl) GetLockConfig(id int) (*bean.LockConfiguration, error) {
	lockConfig := &bean.LockConfiguration{}
	err := impl.dbConnection.Model(lockConfig).
		Where("id = ?", id).
		Where("active = ?", true).
		Select()
	if err != nil {
		return nil, err
	}
	return lockConfig, nil
}

func (impl RepositoryImpl) GetActiveLockConfig() (*bean.LockConfiguration, error) {
	lockConfig := &bean.LockConfiguration{}
	err := impl.dbConnection.Model(lockConfig).Where("active=?", true).Select()
	if err != nil {
		return nil, err
	}
	return lockConfig, nil
}

func (impl RepositoryImpl) DeleteActiveLockConfigs(userId int) error {
	query := "UPDATE lock_configuration " +
		"SET active=false AND updated_on = ? AND updated_by=? " +
		"WHERE active=true;"
	var lockConfigs []*bean.LockConfiguration
	_, err := impl.dbConnection.Query(&lockConfigs, query, time.Now(), userId)
	return err
}

func (impl RepositoryImpl) Create(lockConfig *bean.LockConfiguration, tx *pg.Tx) error {
	return tx.Insert(lockConfig)
}

func (impl RepositoryImpl) Update(lockConfig *bean.LockConfiguration, tx *pg.Tx) error {
	return tx.Update(lockConfig)
}
