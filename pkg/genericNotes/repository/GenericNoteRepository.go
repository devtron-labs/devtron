/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package repository

import (
	"fmt"
	repository1 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type NoteType int

const ClusterType NoteType = 0
const AppType NoteType = 1

type GenericNote struct {
	tableName      struct{} `sql:"generic_note" json:"-" pg:",discard_unknown_columns"`
	Id             int      `sql:"id,pk" json:"id"`
	Identifier     int      `sql:"identifier" json:"identifier" validate:"required"`
	IdentifierType NoteType `sql:"identifier_type,notnull" json:"identifierType"`
	Description    string   `sql:"description" json:"description"`
	sql.AuditLog
}

type GenericNoteRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	Save(tx *pg.Tx, model *GenericNote) error
	FindByClusterId(id int) (*GenericNote, error)
	FindByAppId(id int) (*GenericNote, error)
	FindByIdentifier(identifier int, identifierType NoteType) (*GenericNote, error)
	Update(tx *pg.Tx, model *GenericNote) error
	GetGenericNotesForAppIds(appIds []int) ([]*GenericNote, error)
	GetDescriptionFromAppIds(appIds []int) ([]*GenericNote, error)
}

func NewGenericNoteRepositoryImpl(dbConnection *pg.DB) *GenericNoteRepositoryImpl {
	TransactionUtilImpl := sql.NewTransactionUtilImpl(dbConnection)
	return &GenericNoteRepositoryImpl{
		dbConnection:        dbConnection,
		TransactionUtilImpl: TransactionUtilImpl,
	}
}

type GenericNoteRepositoryImpl struct {
	*sql.TransactionUtilImpl
	dbConnection *pg.DB
}

func (impl GenericNoteRepositoryImpl) Save(tx *pg.Tx, model *GenericNote) error {
	return tx.Insert(model)
}

func (impl GenericNoteRepositoryImpl) FindByClusterId(id int) (*GenericNote, error) {
	return impl.FindByIdentifier(id, ClusterType)
}

func (impl GenericNoteRepositoryImpl) Update(tx *pg.Tx, model *GenericNote) error {
	return tx.Update(model)
}

func (impl GenericNoteRepositoryImpl) FindByAppId(id int) (*GenericNote, error) {
	return impl.FindByIdentifier(id, AppType)
}

func (impl GenericNoteRepositoryImpl) FindByIdentifier(identifier int, identifierType NoteType) (*GenericNote, error) {
	clusterNote := &GenericNote{}
	err := impl.dbConnection.
		Model(clusterNote).
		Where("identifier =?", identifier).
		Where("identifier_type =?", identifierType).
		Limit(1).
		Select()
	return clusterNote, err
}

func (impl GenericNoteRepositoryImpl) GetGenericNotesForAppIds(appIds []int) ([]*GenericNote, error) {
	notes := make([]*GenericNote, 0)
	if len(appIds) == 0 {
		return notes, nil
	}
	err := impl.dbConnection.
		Model(&notes).
		Where("identifier IN (?)", pg.In(appIds)).
		Where("identifier_type =?", AppType).
		Limit(1).
		Select()
	return notes, err
}

func (impl GenericNoteRepositoryImpl) GetDescriptionFromAppIds(appIds []int) ([]*GenericNote, error) {
	apps := make([]*repository1.App, 0)
	if len(appIds) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf("SELECT * "+
		"FROM app WHERE id IN (%s)", helper.GetCommaSepratedString(appIds))
	_, err := impl.dbConnection.Query(&apps, query)
	if err != nil {
		return nil, err
	}
	notes := make([]*GenericNote, 0, len(apps))
	for _, app := range apps {
		note := &GenericNote{}
		note.Id = 0
		note.Identifier = int(AppType)
		note.Description = app.Description
		note.UpdatedOn = app.UpdatedOn
		note.UpdatedBy = app.UpdatedBy
		notes = append(notes, note)
	}
	return notes, nil
}
