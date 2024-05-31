/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package apiToken

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type ApiToken struct {
	tableName    struct{} `sql:"api_token"`
	Id           int      `sql:"id,pk"`
	UserId       int32    `sql:"user_id, notnull"`
	Name         string   `sql:"name, notnull"`
	Version      int      `sql:"version, notnull"`
	Description  string   `sql:"description, notnull"`
	ExpireAtInMs int64    `sql:"expire_at_in_ms"`
	Token        string   `sql:"token, notnull"`
	User         *repository.UserModel
	sql.AuditLog
}

type ApiTokenRepository interface {
	Save(apiToken *ApiToken) error
	Update(apiToken *ApiToken) error
	FindAllActive() ([]*ApiToken, error)
	FindActiveById(id int) (*ApiToken, error)
	FindByName(name string) (*ApiToken, error)
	UpdateIf(apiToken *ApiToken, previousTokenVersion int) error
}

type ApiTokenRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewApiTokenRepositoryImpl(dbConnection *pg.DB) *ApiTokenRepositoryImpl {
	return &ApiTokenRepositoryImpl{dbConnection: dbConnection}
}

func (impl ApiTokenRepositoryImpl) Save(apiToken *ApiToken) error {
	return impl.dbConnection.Insert(apiToken)
}

func (impl ApiTokenRepositoryImpl) Update(apiToken *ApiToken) error {
	return impl.dbConnection.Update(apiToken)
}

func (impl ApiTokenRepositoryImpl) UpdateIf(apiToken *ApiToken, previousTokenVersion int) error {
	res, err := impl.dbConnection.Model(apiToken).
		Where("id = ?", apiToken.Id).
		Where("version = ?", previousTokenVersion).
		Update()
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf(TokenVersionMismatch)
	}
	return nil
}

func (impl ApiTokenRepositoryImpl) FindAllActive() ([]*ApiToken, error) {
	var apiTokens []*ApiToken
	err := impl.dbConnection.Model(&apiTokens).
		Column("api_token.*", "User").
		Relation("User", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("active IS TRUE"), nil
		}).
		Select()
	return apiTokens, err
}

func (impl ApiTokenRepositoryImpl) FindActiveById(id int) (*ApiToken, error) {
	apiToken := &ApiToken{}
	err := impl.dbConnection.Model(apiToken).
		Column("api_token.*", "User").
		Relation("User", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("active IS TRUE"), nil
		}).
		Where("api_token.id = ?", id).
		Select()
	return apiToken, err
}

func (impl ApiTokenRepositoryImpl) FindByName(name string) (*ApiToken, error) {
	apiToken := &ApiToken{}
	err := impl.dbConnection.Model(apiToken).
		Column("api_token.*", "User").
		Where("api_token.name = ?", name).
		Select()
	return apiToken, err
}
