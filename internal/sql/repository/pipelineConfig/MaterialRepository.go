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

package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type SourceType string

const (
	SOURCE_TYPE_BRANCH_FIXED SourceType = "SOURCE_TYPE_BRANCH_FIXED"
	SOURCE_TYPE_BRANCH_REGEX SourceType = "SOURCE_TYPE_BRANCH_REGEX"
	SOURCE_TYPE_TAG_ANY      SourceType = "SOURCE_TYPE_TAG_ANY"
	SOURCE_TYPE_WEBHOOK      SourceType = "WEBHOOK"
)

//TODO: add support for submodule
type GitMaterial struct {
	tableName       struct{} `sql:"git_material" pg:",discard_unknown_columns"`
	Id              int      `sql:"id,pk"`
	AppId           int      `sql:"app_id,notnull"`
	GitProviderId   int      `sql:"git_provider_id,notnull"`
	Active          bool     `sql:"active,notnull"`
	Url             string   `sql:"url,omitempty"`
	Name            string   `sql:"name, omitempty"`
	CheckoutPath    string   `sql:"checkout_path, omitempty"`
	FetchSubmodules bool     `sql:"fetch_submodules,notnull"`
	sql.AuditLog
	App         *app.App
	GitProvider *repository.GitProvider
}

type MaterialRepository interface {
	MaterialExists(url string) (bool, error)
	SaveMaterial(material *GitMaterial) error
	UpdateMaterial(material *GitMaterial) error
	Update(materials []*GitMaterial) error
	FindByAppId(appId int) ([]*GitMaterial, error)
	FindById(Id int) (*GitMaterial, error)
	UpdateMaterialScmId(material *GitMaterial) error
	FindByAppIdAndCheckoutPath(appId int, checkoutPath string) (*GitMaterial, error)
}
type MaterialRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewMaterialRepositoryImpl(dbConnection *pg.DB) *MaterialRepositoryImpl {
	return &MaterialRepositoryImpl{dbConnection: dbConnection}
}

func (repo MaterialRepositoryImpl) FindByAppId(appId int) ([]*GitMaterial, error) {
	var materials []*GitMaterial
	err := repo.dbConnection.Model(&materials).
		Column("git_material.*", "GitProvider").
		Where("app_id =? ", appId).
		Where("git_material.active =? ", true).
		Select()
	return materials, err
}

func (repo MaterialRepositoryImpl) FindById(Id int) (*GitMaterial, error) {
	material := &GitMaterial{}
	err := repo.dbConnection.Model(material).
		Column("git_material.*", "GitProvider").
		Where("git_material.id =? ", Id).
		Where("git_material.active =? ", true).
		Select()
	return material, err
}

func (repo MaterialRepositoryImpl) MaterialExists(url string) (bool, error) {
	material := &GitMaterial{}
	exists, err := repo.dbConnection.
		Model(material).
		Where("url = ?", url).
		Exists()
	return exists, err
}

func (repo MaterialRepositoryImpl) SaveMaterial(material *GitMaterial) error {
	return repo.dbConnection.Insert(material)
}

func (repo MaterialRepositoryImpl) UpdateMaterial(material *GitMaterial) error {
	return repo.dbConnection.Update(material)
}

func (repo MaterialRepositoryImpl) UpdateMaterialScmId(material *GitMaterial) error {
	panic(nil)
	/*	_, err := repo.dbConnection.Model(material).
		Set("ci_scm_id =? ", material.CiScmId).
		Set("ct_scm_id =? ", material.CtScmId).
		Set("production_scm_id =? ", material.ProductionScmId).
		Where("id =? ", material.GitMaterialId).
		Update()*/
	return nil
}

func (impl MaterialRepositoryImpl) Update(materials []*GitMaterial) error {
	err := impl.dbConnection.RunInTransaction(func(tx *pg.Tx) error {
		for _, material := range materials {
			_, err := tx.Model(material).WherePK().UpdateNotNull()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (repo MaterialRepositoryImpl) FindByAppIdAndCheckoutPath(appId int, checkoutPath string) (*GitMaterial, error) {
	material := &GitMaterial{}
	err := repo.dbConnection.Model(material).
		Column("git_material.*", "GitProvider").
		Where("app_id = ? ", appId).
		Where("checkout_path = ?", checkoutPath).
		Where("git_material.active =? ", true).
		Select()
	return material, err
}
