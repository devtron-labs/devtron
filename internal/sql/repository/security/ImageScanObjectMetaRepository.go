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

package security

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ImageScanObjectMeta struct {
	tableName struct{} `sql:"image_scan_object_meta" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
	Image     string   `sql:"image,notnull"`
	Active    bool     `sql:"active"`
}

type ImageScanObjectMetaRepository interface {
	Save(model *ImageScanObjectMeta) error
	FindAll() ([]*ImageScanObjectMeta, error)
	FindOne(id int) (*ImageScanObjectMeta, error)
	FindByNameAndType(name string, types string) ([]*ImageScanObjectMeta, error)
	Update(model *ImageScanObjectMeta) error
}

type ImageScanObjectMetaRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewImageScanObjectMetaRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ImageScanObjectMetaRepositoryImpl {
	return &ImageScanObjectMetaRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl ImageScanObjectMetaRepositoryImpl) Save(model *ImageScanObjectMeta) error {
	err := impl.dbConnection.Insert(model)
	return err
}

func (impl ImageScanObjectMetaRepositoryImpl) FindAll() ([]*ImageScanObjectMeta, error) {
	var models []*ImageScanObjectMeta
	err := impl.dbConnection.Model(&models).Where("active=?", true).Select()
	return models, err
}

func (impl ImageScanObjectMetaRepositoryImpl) FindOne(id int) (*ImageScanObjectMeta, error) {
	var model *ImageScanObjectMeta
	err := impl.dbConnection.Model(&model).
		Where("id = ?", id).Select()
	return model, err
}

func (impl ImageScanObjectMetaRepositoryImpl) FindByNameAndType(name string, types string) ([]*ImageScanObjectMeta, error) {
	var models []*ImageScanObjectMeta
	err := impl.dbConnection.Model(&models).
		Where("cve_name = ?", name).Select()
	return models, err
}

func (impl ImageScanObjectMetaRepositoryImpl) Update(team *ImageScanObjectMeta) error {
	err := impl.dbConnection.Update(team)
	return err
}
