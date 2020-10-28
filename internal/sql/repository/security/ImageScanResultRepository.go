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

type ImageScanExecutionResult struct {
	tableName                   struct{} `sql:"image_scan_execution_result" pg:",discard_unknown_columns"`
	Id                          int      `sql:"id,pk"`
	CveStoreName                string   `sql:"cve_store_name,notnull"`
	ImageScanExecutionHistoryId int      `sql:"image_scan_execution_history_id"`
	CveStore                    CveStore
	ImageScanExecutionHistory   ImageScanExecutionHistory
}

type ImageScanResultRepository interface {
	Save(model *ImageScanExecutionResult) error
	FindAll() ([]*ImageScanExecutionResult, error)
	FindOne(id int) (*ImageScanExecutionResult, error)
	FindByCveName(name string) ([]*ImageScanExecutionResult, error)
	Update(model *ImageScanExecutionResult) error
	FetchByScanExecutionId(id int) ([]*ImageScanExecutionResult, error)
	FetchByScanExecutionIds(ids []int) ([]*ImageScanExecutionResult, error)
	FindByImageDigest(imageDigest string) ([]*ImageScanExecutionResult, error)
	FindByImageDigests(digest []string) ([]*ImageScanExecutionResult, error)
	FindByImage(image string) ([]*ImageScanExecutionResult, error)
}

type ImageScanResultRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewImageScanResultRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ImageScanResultRepositoryImpl {
	return &ImageScanResultRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl ImageScanResultRepositoryImpl) Save(model *ImageScanExecutionResult) error {
	err := impl.dbConnection.Insert(model)
	return err
}

func (impl ImageScanResultRepositoryImpl) FindAll() ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FindOne(id int) (*ImageScanExecutionResult, error) {
	var model *ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).
		Where("id = ?", id).Select()
	return model, err
}

func (impl ImageScanResultRepositoryImpl) FindByCveName(name string) ([]*ImageScanExecutionResult, error) {
	var model []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).
		Where("cve_store_name = ?", name).Select()
	return model, err
}

func (impl ImageScanResultRepositoryImpl) Update(team *ImageScanExecutionResult) error {
	err := impl.dbConnection.Update(team)
	return err
}

func (impl ImageScanResultRepositoryImpl) FetchByScanExecutionId(scanExecutionId int) ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	/*err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "cs.*").
	Join("inner join cve_store cs on cs.name=image_scan_execution_result.cve_name").
	Where("image_scan_execution_result.scan_execution_id = ?", id).Select()
	*/

	err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "CveStore").
		Where("image_scan_execution_result.image_scan_execution_history_id = ?", scanExecutionId).
		Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FetchByScanExecutionIds(ids []int) ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_scan_execution_result.image_scan_execution_history_id in(?)", pg.In(ids)).
		Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FindByImageDigest(imageDigest string) ([]*ImageScanExecutionResult, error) {
	var model []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_scan_execution_history.image_hash = ?", imageDigest).Order("image_scan_execution_history.execution_time desc").Select()
	return model, err
}

func (impl ImageScanResultRepositoryImpl) FindByImageDigests(digest []string) ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_hash in (?)", pg.In(digest)).Order("execution_time desc").Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FindByImage(image string) ([]*ImageScanExecutionResult, error) {
	var model []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_scan_execution_history.image = ?", image).Order("image_scan_execution_history.execution_time desc").Select()
	return model, err
}
