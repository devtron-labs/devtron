/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"time"
)

type UserAttributes struct {
	tableName struct{} `sql:"user_attributes" pg:",discard_unknown_columns"`
	EmailId   string   `sql:"email_id,pk"`
	UserData  string   `sql:"user_data,notnull"`
	sql.AuditLog
}

type UserAttributesDao struct {
	EmailId string `json:"emailId"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	UserId  int32  `json:"-"`
}

type UserAttributesRepository interface {
	GetConnection() (dbConnection *pg.DB)
	AddUserAttribute(attrDto *UserAttributesDao) (*UserAttributesDao, error)
	UpdateDataValByKey(attrDto *UserAttributesDao) error
	GetDataValueByKey(attrDto *UserAttributesDao) (string, error)
	GetUserDataByEmailId(emailId string) (string, error)
}

type UserAttributesRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewUserAttributesRepositoryImpl(dbConnection *pg.DB) *UserAttributesRepositoryImpl {
	return &UserAttributesRepositoryImpl{dbConnection: dbConnection}
}

func (impl *UserAttributesRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (repo UserAttributesRepositoryImpl) AddUserAttribute(attrDto *UserAttributesDao) (*UserAttributesDao, error) {
	userDataMap := make(map[string]interface{})
	userDataMap[attrDto.Key] = attrDto.Value
	userAttr := UserAttributes{}
	userAttr.EmailId = attrDto.EmailId
	userDataJson, err := json.Marshal(userDataMap)
	if err != nil {
		return nil, err
	}
	userAttr.UserData = string(userDataJson)
	userAttr.CreatedBy = attrDto.UserId
	userAttr.UpdatedBy = attrDto.UserId
	userAttr.CreatedOn = time.Now()
	userAttr.UpdatedOn = time.Now()

	err = repo.dbConnection.Insert(&userAttr)
	if err != nil {
		return nil, err
	}

	return attrDto, nil
}

func (repo UserAttributesRepositoryImpl) UpdateDataValByKey(attrDto *UserAttributesDao) error {
	var userAttr = &UserAttributes{}
	keyValMap := make(map[string]string)
	keyValMap[attrDto.Key] = attrDto.Value
	updatedValJson, err := json.Marshal(keyValMap)
	if err != nil {
		return err
	}
	query := "update user_attributes SET user_data = user_data::jsonb - ? || ? where email_id = ?"

	_, err = repo.dbConnection.
		Query(userAttr, query, attrDto.Key, string(updatedValJson), attrDto.EmailId)
	return err
}

func (repo UserAttributesRepositoryImpl) GetDataValueByKey(attrDto *UserAttributesDao) (string, error) {
	model := &UserAttributes{}
	err := repo.dbConnection.Model(model).Where("email_id = ?", attrDto.EmailId).
		Select()
	if err != nil {
		return "", err
	}
	data := model.UserData
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(data), &jsonMap)
	if err != nil {
		return "", err
	}
	dataVal := jsonMap[attrDto.Key]
	var response = ""
	if dataVal != nil {
		response = dataVal.(string)
	}
	return response, err
}

func (repo UserAttributesRepositoryImpl) GetUserDataByEmailId(emailId string) (string, error) {
	model := &UserAttributes{}
	err := repo.dbConnection.Model(model).Where("email_id = ?", emailId).
		Select()
	if err != nil {
		return "", err
	}
	return model.UserData, err
}
