package repository

import (
	_ "github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CICD struct {
	tableName struct{} `sql:"id_create_cicd" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
}

type CicdRepository interface {
	Save(cicd *CICD) error
	FindByAppId(cicdId int) (*CICD, error)
	Update(cicd *CICD) error
	Delete(cicd *CICD) error
}

type CicdRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCicdRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CicdRepositoryImpl {
	return &CicdRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *CicdRepositoryImpl) Save(cicd *CICD) error {
	return impl.dbConnection.Insert(cicd)
}

func (impl *CicdRepositoryImpl) FindByAppId(cicdId int) (*CICD, error) {
	cicd := &CICD{}
	err := impl.dbConnection.Model(cicd).Where("id = ? ", cicdId).Select()
	if err != nil {
		println(err)
	}
	return cicd, err
}

func (impl *CicdRepositoryImpl) Update(cicd *CICD) error {
	err := impl.dbConnection.Update(cicd)
	return err
}

func (impl *CicdRepositoryImpl) Delete(cicd *CICD) error {
	err := impl.dbConnection.Delete(cicd)
	return err
}
