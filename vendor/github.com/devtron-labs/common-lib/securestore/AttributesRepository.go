/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package securestore

import (
	sql2 "github.com/devtron-labs/common-lib/utils/sql"
	"github.com/go-pg/pg"
	"time"
)

const ENCRYPTION_KEY string = "encryptionKey" // AES-256 encryption key for sensitive data

type Attributes struct {
	tableName struct{} `sql:"attributes" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Key       string   `sql:"key,notnull"`
	Value     string   `sql:"value,notnull"`
	Active    bool     `sql:"active, notnull"`
	sql2.AuditLog
}

type AttributesRepository interface {
	Save(model *Attributes, tx *pg.Tx) (*Attributes, error)
	Update(model *Attributes, tx *pg.Tx) error
	FindByKey(key string) (*Attributes, error)
	GetConnection() (dbConnection *pg.DB)
	SaveEncryptionKeyIfNotExists(value string) error
	GetEncryptionKey() (string, error)
}

type AttributesRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewAttributesRepositoryImpl(dbConnection *pg.DB) *AttributesRepositoryImpl {
	return &AttributesRepositoryImpl{dbConnection: dbConnection}
}

func (impl *AttributesRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (repo AttributesRepositoryImpl) Save(model *Attributes, tx *pg.Tx) (*Attributes, error) {
	err := tx.Insert(model)
	if err != nil {
		return model, err
	}
	return model, nil
}

func (repo AttributesRepositoryImpl) Update(model *Attributes, tx *pg.Tx) error {
	err := tx.Update(model)
	if err != nil {
		return err
	}
	return nil
}

func (repo AttributesRepositoryImpl) FindByKey(key string) (*Attributes, error) {
	model := &Attributes{}
	err := repo.dbConnection.
		Model(model).
		Where("key = ?", key).
		Where("active = ?", true).
		Select()
	if err != nil {
		return model, err
	}
	return model, nil
}

// SaveEncryptionKeyIfNotExists saves an encryption key in the attributes table if not exists
func (repo AttributesRepositoryImpl) SaveEncryptionKeyIfNotExists(value string) error {
	dbConnection := repo.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//deleting all keys if exists for safe side
	_, err = tx.Model(&Attributes{}).Where("key = ?", ENCRYPTION_KEY).Delete()
	// Create new encryption key entry
	model := &Attributes{
		Key:    ENCRYPTION_KEY,
		Value:  value,
		Active: true,
		AuditLog: sql2.AuditLog{
			CreatedBy: 1,
			UpdatedBy: 1,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	}
	_, err = repo.Save(model, tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetEncryptionKey retrieves the active encryption key from the attributes table
func (repo AttributesRepositoryImpl) GetEncryptionKey() (string, error) {
	model, err := repo.FindByKey(ENCRYPTION_KEY)
	if err != nil {
		return "", err
	}
	return model.Value, nil
}
