package user

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type HelmUserRepository interface {
	CreateHelmUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error)
	UpdateHelmUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error)
}

type HelmUserRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewHelmUserRepositoryImpl(dbConnection *pg.DB, Logger *zap.SugaredLogger) *HelmUserRepositoryImpl {
	return &HelmUserRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type HelmUserModel struct {
	TableName   struct{} `sql:"users"`
	Id          int32    `sql:"id,pk"`
	EmailId     string   `sql:"email_id,notnull"`
	AccessToken string   `sql:"access_token"`
	Active      bool     `sql:"active,notnull"`
	sql.AuditLog
}
type HelmUserRoleModel struct {
	TableName struct{} `sql:"user_roles"`
	Id        int      `sql:"id,pk"`
	UserId    int32    `sql:"user_id,notnull"`
	RoleId    int      `sql:"role_id,notnull"`
	User      HelmUserModel
	sql.AuditLog
}

func (impl HelmUserRepositoryImpl) CreateHelmUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error) {
	err := tx.Insert(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}
	return userModel, nil
}
func (impl HelmUserRepositoryImpl) UpdateHelmUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error) {
	err := tx.Update(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}
	return userModel, nil
}
