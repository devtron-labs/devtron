package repository

import (
	"github.com/devtron-labs/devtron/pkg/serverConnection/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ServerConnectionRepository interface {
	Save(model *ServerConnectionConfig, tx *pg.Tx) error
	Update(model *ServerConnectionConfig, tx *pg.Tx) error
	GetById(id int) (*ServerConnectionConfig, error)
	MarkServerConnectionConfigDeleted(deleteReq *ServerConnectionConfig, tx *pg.Tx) error
}

type ServerConnectionRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewServerConnectionRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ServerConnectionRepositoryImpl {
	return &ServerConnectionRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type ServerConnectionConfig struct {
	tableName        struct{}                    `sql:"server_connection_config" pg:",discard_unknown_columns"`
	Id               int                         `sql:"id,pk"`
	ConnectionMethod bean.ServerConnectionMethod `sql:"connection_method"`
	ProxyUrl         string                      `sql:"proxy_url"`
	SSHServerAddress string                      `sql:"ssh_server_address"`
	SSHUsername      string                      `sql:"ssh_username"`
	SSHPassword      string                      `sql:"ssh_password"`
	SSHAuthKey       string                      `sql:"ssh_auth_key"`
	Deleted          bool                        `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *ServerConnectionRepositoryImpl) Save(model *ServerConnectionConfig, tx *pg.Tx) error {
	return tx.Insert(model)
}

func (repo *ServerConnectionRepositoryImpl) Update(model *ServerConnectionConfig, tx *pg.Tx) error {
	return tx.Update(model)
}

func (repo *ServerConnectionRepositoryImpl) GetById(id int) (*ServerConnectionConfig, error) {
	model := &ServerConnectionConfig{}
	err := repo.dbConnection.Model(model).
		Where("id = ?", id).
		Where("deleted = ?", false).
		Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error in getting server connection config", "err", err, "id", id)
		return nil, err
	}
	return model, nil
}

func (repo *ServerConnectionRepositoryImpl) MarkServerConnectionConfigDeleted(deleteReq *ServerConnectionConfig, tx *pg.Tx) error {
	deleteReq.Deleted = true
	return tx.Update(deleteReq)
}
