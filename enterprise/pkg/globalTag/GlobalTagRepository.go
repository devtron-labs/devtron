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

package globalTag

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type GlobalTag struct {
	tableName              struct{} `sql:"global_tag"`
	Id                     int      `sql:"id,pk"`
	Key                    string   `sql:"key, notnull"`
	MandatoryProjectIdsCsv string   `sql:"mandatory_project_ids_csv"`
	Propagate              bool     `sql:"propagate"`
	Description            string   `sql:"description, notnull"`
	Active                 bool     `sql:"active"`
	sql.AuditLog
}

type GlobalTagRepository interface {
	GetConnection() *pg.DB
	FindAllActive() ([]*GlobalTag, error)
	CheckKeyExistsForAnyActiveTag(key string) (bool, error)
	CheckKeyExistsForAnyActiveTagExcludeTagId(key string, tagId int) (bool, error)
	FindAllActiveByIds(ids []int) ([]*GlobalTag, error)
	FindActiveById(id int) (*GlobalTag, error)
	Save(globalTags []*GlobalTag, tx *pg.Tx) error
	Update(globalTag *GlobalTag, tx *pg.Tx) error
}

type GlobalTagRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewGlobalTagRepositoryImpl(dbConnection *pg.DB) *GlobalTagRepositoryImpl {
	return &GlobalTagRepositoryImpl{dbConnection: dbConnection}
}

func (impl GlobalTagRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl GlobalTagRepositoryImpl) FindAllActive() ([]*GlobalTag, error) {
	var globalTags []*GlobalTag
	err := impl.dbConnection.Model(&globalTags).
		Where("active IS TRUE").
		Select()
	return globalTags, err
}

func (impl GlobalTagRepositoryImpl) CheckKeyExistsForAnyActiveTag(key string) (bool, error) {
	var globalTag *GlobalTag
	exists, err := impl.dbConnection.Model(globalTag).
		Where("active IS TRUE").
		Where("key = ?", key).
		Exists()
	return exists, err
}

func (impl GlobalTagRepositoryImpl) CheckKeyExistsForAnyActiveTagExcludeTagId(key string, tagId int) (bool, error) {
	var globalTag *GlobalTag
	exists, err := impl.dbConnection.Model(globalTag).
		Where("active IS TRUE").
		Where("key = ?", key).
		Where("id != ?", tagId).
		Exists()
	return exists, err
}

func (impl GlobalTagRepositoryImpl) FindAllActiveByIds(ids []int) ([]*GlobalTag, error) {
	var globalTags []*GlobalTag
	err := impl.dbConnection.Model(&globalTags).
		Where("id in (?)", pg.In(ids)).
		Where("active IS TRUE").
		Select()
	return globalTags, err
}

func (impl GlobalTagRepositoryImpl) FindActiveById(id int) (*GlobalTag, error) {
	globalTag := &GlobalTag{}
	err := impl.dbConnection.Model(globalTag).
		Where("active IS TRUE").
		Where("id = ?", id).
		Select()
	return globalTag, err
}

func (impl GlobalTagRepositoryImpl) Save(globalTags []*GlobalTag, tx *pg.Tx) error {
	return tx.Insert(&globalTags)
}

func (impl GlobalTagRepositoryImpl) Update(globalTag *GlobalTag, tx *pg.Tx) error {
	return tx.Update(globalTag)
}
