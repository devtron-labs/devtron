package repository

import (
	"github.com/devtron-labs/devtron/pkg/remoteConnection/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RemoteConnectionRepository interface {
	Save(model *RemoteConnectionConfig, tx *pg.Tx) error
	Update(model *RemoteConnectionConfig, tx *pg.Tx) error
	GetById(id int) (*RemoteConnectionConfig, error)
	MarkRemoteConnectionConfigDeleted(deleteReq *RemoteConnectionConfig, tx *pg.Tx) error
}

type RemoteConnectionRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewRemoteConnectionRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *RemoteConnectionRepositoryImpl {
	return &RemoteConnectionRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type RemoteConnectionConfig struct {
	tableName        struct{}                    `sql:"remote_connection_config" pg:",discard_unknown_columns"`
	Id               int                         `sql:"id,pk"`
	ConnectionMethod bean.RemoteConnectionMethod `sql:"connection_method"`
	ProxyUrl         string                      `sql:"proxy_url"`
	SSHServerAddress string                      `sql:"ssh_server_address"`
	SSHUsername      string                      `sql:"ssh_username"`
	SSHPassword      string                      `sql:"ssh_password"`
	SSHAuthKey       string                      `sql:"ssh_auth_key"`
	Deleted          bool                        `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *RemoteConnectionRepositoryImpl) Save(model *RemoteConnectionConfig, tx *pg.Tx) error {
	return tx.Insert(model)
}

func (repo *RemoteConnectionRepositoryImpl) Update(model *RemoteConnectionConfig, tx *pg.Tx) error {
	return tx.Update(model)
}

func (repo *RemoteConnectionRepositoryImpl) GetById(id int) (*RemoteConnectionConfig, error) {
	model := &RemoteConnectionConfig{}
	err := repo.dbConnection.Model(model).
		Where("id = ?", id).
		Where("deleted = ?", false).
		Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error in getting remote connection config", "err", err, "id", id)
		return nil, err
	}
	return model, nil
}

func (repo *RemoteConnectionRepositoryImpl) MarkRemoteConnectionConfigDeleted(deleteReq *RemoteConnectionConfig, tx *pg.Tx) error {
	deleteReq.Deleted = true
	return tx.Update(deleteReq)
}
