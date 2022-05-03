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

package module

import (
	"github.com/go-pg/pg"
	"time"
)

type Module struct {
	tableName struct{}  `sql:"module"`
	Id        int       `sql:"id,pk"`
	Name      string    `sql:"name, notnull"`
	Version   string    `sql:"version, notnull"`
	Status    string    `sql:"status,notnull"`
	UpdatedOn time.Time `sql:"updated_on"`
}

type ModuleRepository interface {
	Save(module *Module) error
	FindOne(name string) (*Module, error)
	Update(module *Module) error
	FindAll() ([]Module, error)
}

type ModuleRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewModuleRepositoryImpl(dbConnection *pg.DB) *ModuleRepositoryImpl {
	return &ModuleRepositoryImpl{dbConnection: dbConnection}
}

func (impl ModuleRepositoryImpl) Save(module *Module) error {
	return impl.dbConnection.Insert(module)
}

func (impl ModuleRepositoryImpl) FindOne(name string) (*Module, error) {
	module := &Module{}
	err := impl.dbConnection.Model(module).
		Where("name = ?", name).Select()
	return module, err
}

func (impl ModuleRepositoryImpl) Update(module *Module) error {
	return impl.dbConnection.Update(module)
}

func (impl ModuleRepositoryImpl) FindAll() ([]Module, error) {
	var modules []Module
	err := impl.dbConnection.Model(&modules).
		Select()
	return modules, err
}
