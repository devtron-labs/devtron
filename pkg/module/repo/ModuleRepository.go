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

package moduleRepo

import (
	"github.com/go-pg/pg"
	"time"
)

type Module struct {
	tableName  struct{}  `sql:"module"`
	Id         int       `sql:"id,pk"`
	Name       string    `sql:"name, notnull"`
	Version    string    `sql:"version, notnull"`
	Status     string    `sql:"status,notnull"`
	UpdatedOn  time.Time `sql:"updated_on"`
	Enabled    bool      `sql:"enabled"`
	ModuleType string    `sql:"module_type"`
}

type ModuleRepository interface {
	Save(module *Module) error
	FindOne(name string) (*Module, error)
	Update(module *Module) error
	FindAll() ([]Module, error)
	ModuleExists() (bool, error)
	GetInstalledModuleNames() ([]string, error)
	FindByModuleTypeAndStatus(moduleType string, status string) error
	MarkModuleAsEnabledWithTransaction(moduleName string, tx *pg.Tx) error
	GetConnection() (dbConnection *pg.DB)
	MarkOtherModulesDisabledOfSameType(moduleName, moduleType string, tx *pg.Tx) error
	UpdateWithTransaction(module *Module, tx *pg.Tx) error
	SaveWithTransaction(module *Module, tx *pg.Tx) error
	MarkModuleAsEnabled(moduleName string) error
}

type ModuleRepositoryImpl struct {
	dbConnection *pg.DB
}

func (impl ModuleRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func NewModuleRepositoryImpl(dbConnection *pg.DB) *ModuleRepositoryImpl {
	return &ModuleRepositoryImpl{dbConnection: dbConnection}
}

func (impl ModuleRepositoryImpl) Save(module *Module) error {
	return impl.dbConnection.Insert(module)
}

func (impl ModuleRepositoryImpl) SaveWithTransaction(module *Module, tx *pg.Tx) error {
	return tx.Insert(module)
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

func (impl ModuleRepositoryImpl) UpdateWithTransaction(module *Module, tx *pg.Tx) error {
	return tx.Update(module)
}
func (impl ModuleRepositoryImpl) FindAllByStatus(status string) ([]Module, error) {
	var modules []Module
	err := impl.dbConnection.Model(&modules).
		Where("status = ?", status).
		Select()
	return modules, err
}

func (impl ModuleRepositoryImpl) FindAll() ([]Module, error) {
	var modules []Module
	err := impl.dbConnection.Model(&modules).
		Select()
	return modules, err
}

func (impl ModuleRepositoryImpl) ModuleExists() (bool, error) {
	module := &Module{}
	exists, err := impl.dbConnection.Model(module).
		Exists()
	return exists, err
}

func (impl ModuleRepositoryImpl) GetInstalledModuleNames() ([]string, error) {
	modules, err := impl.FindAllByStatus("installed")
	var moduleNames []string
	if err != nil && err != pg.ErrNoRows {
		return moduleNames, err
	}

	for _, module := range modules {
		moduleNames = append(moduleNames, module.Name)
	}
	return moduleNames, nil
}

func (impl ModuleRepositoryImpl) FindByModuleTypeAndStatus(moduleType string, status string) error {
	module := &Module{}
	err := impl.dbConnection.Model(module).
		Where("module_type = ?", moduleType).
		Where("status = ?", status).
		Select()
	return err
}

func (impl ModuleRepositoryImpl) MarkModuleAsEnabledWithTransaction(moduleName string, tx *pg.Tx) error {
	module := &Module{}
	_, err := tx.Model(module).Set("enabled = ?", true).Where("name = ?", moduleName).Update()
	return err
}
func (impl ModuleRepositoryImpl) MarkModuleAsEnabled(moduleName string) error {
	module := &Module{}
	_, err := impl.dbConnection.Model(module).Set("enabled = ?", true).Where("name = ?", moduleName).Update()
	return err
}

func (impl ModuleRepositoryImpl) MarkOtherModulesDisabledOfSameType(moduleName, moduleType string, tx *pg.Tx) error {
	module := &Module{}
	_, err := tx.Model(module).Set("enabled = ?", false).Where("name != ?", moduleName).Where("module_type = ?", moduleType).Update()
	return err
}
