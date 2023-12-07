package lockConfiguration

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/go-pg/pg"
)

type LockConfigurationRepository interface {
	GetConnection() *pg.DB
	GetLockConfig(int) (*bean.LockConfiguration, error)
	GetActiveLockConfig() (*bean.LockConfiguration, error)
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

func (impl RepositoryImpl) Create(lockConfig *bean.LockConfiguration, tx *pg.Tx) error {
	return tx.Insert(lockConfig)
}

func (impl RepositoryImpl) Update(lockConfig *bean.LockConfiguration, tx *pg.Tx) error {
	return tx.Update(lockConfig)
}
