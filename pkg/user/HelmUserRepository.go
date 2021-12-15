package user

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	bean2 "github.com/devtron-labs/devtron/pkg/user/dto"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type HelmUserRepository interface {
	CreateUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error)
	UpdateUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error)
	GetById(id int32) (*HelmUserModel, error)
	GetByIdIncludeDeleted(id int32) (*HelmUserModel, error)
	GetAll() ([]HelmUserModel, error)
	FetchActiveUserByEmail(email string) (bean2.UserInfo, error)
	FetchUserDetailByEmail(email string) (bean2.UserInfo, error)
	GetByIds(ids []int32) ([]HelmUserModel, error)
	GetConnection() (dbConnection *pg.DB)
	FetchUserMatchesByEmailId(email string) ([]HelmUserModel, error)
	FetchActiveOrDeletedUserByEmail(email string) (*HelmUserModel, error)
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

func (impl *HelmUserRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl HelmUserRepositoryImpl) CreateUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error) {
	err := tx.Insert(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}
	//TODO - Create Entry In UserRole With Default Role for User
	return userModel, nil
}
func (impl HelmUserRepositoryImpl) UpdateUser(userModel *HelmUserModel, tx *pg.Tx) (*HelmUserModel, error) {
	err := tx.Update(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}

	//TODO - Create Entry In UserRole With Default Role for User

	return userModel, nil
}
func (impl HelmUserRepositoryImpl) GetById(id int32) (*HelmUserModel, error) {
	var model HelmUserModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Where("active = ?", true).Select()
	return &model, err
}

func (impl HelmUserRepositoryImpl) GetByIdIncludeDeleted(id int32) (*HelmUserModel, error) {
	var model HelmUserModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	return &model, err
}

func (impl HelmUserRepositoryImpl) GetAll() ([]HelmUserModel, error) {
	var userModel []HelmUserModel
	err := impl.dbConnection.Model(&userModel).Where("active = ?", true).Order("updated_on desc").Select()
	return userModel, err
}

func (impl HelmUserRepositoryImpl) FetchActiveUserByEmail(email string) (bean2.UserInfo, error) {
	var users bean2.UserInfo

	query := "SELECT u.id, u.email_id, u.access_token FROM users u" +
		" WHERE u.active = true and u.email_id ILIKE ? order by u.updated_on desc"
	_, err := impl.dbConnection.Query(&users, query, email)
	if err != nil {
		impl.Logger.Error("Exception caught:", err)
		return users, err
	}

	return users, nil
}

func (impl HelmUserRepositoryImpl) FetchUserDetailByEmail(email string) (bean2.UserInfo, error) {
	//impl.Logger.Info("reached at FetchUserDetailByEmail:")
	var users []bean.UserRole
	var userFinal bean2.UserInfo

	query := "SELECT u.id, u.email_id, r.role FROM users u" +
		" INNER JOIN user_roles ur ON ur.user_id=u.id" +
		" INNER JOIN roles r ON r.id=ur.role_id" +
		" WHERE u.email_id= ? and u.active = true" +
		" ORDER BY u.updated_on desc;"
	_, err := impl.dbConnection.Query(&users, query, email)
	if err != nil {
		return userFinal, err
	}

	var role []string
	for _, item := range users {
		userFinal.Id = item.Id
		userFinal.EmailId = item.EmailId
		role = append(role, item.Role)
	}
	userFinal.Roles = role
	return userFinal, nil
}
func (impl HelmUserRepositoryImpl) GetByIds(ids []int32) ([]HelmUserModel, error) {
	var model []HelmUserModel
	err := impl.dbConnection.Model(&model).Where("id in (?)", pg.In(ids)).Where("active = ?", true).Select()
	return model, err
}

func (impl HelmUserRepositoryImpl) FetchUserMatchesByEmailId(email string) ([]HelmUserModel, error) {
	var model []HelmUserModel
	err := impl.dbConnection.Model(&model).Where("email_id like (?)", "%"+email+"%").Where("active = ?", true).Select()
	return model, err
}

func (impl HelmUserRepositoryImpl) FetchActiveOrDeletedUserByEmail(email string) (*HelmUserModel, error) {
	var model HelmUserModel
	err := impl.dbConnection.Model(&model).Where("email_id ILIKE (?)", email).Limit(1).Select()
	return &model, err
}
