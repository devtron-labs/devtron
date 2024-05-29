/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package security

import (
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ImageScanExecutionHistory struct {
	tableName                       struct{}      `sql:"image_scan_execution_history" pg:",discard_unknown_columns"`
	Id                              int           `sql:"id,pk"`
	Image                           string        `sql:"image,notnull"`
	ImageHash                       string        `sql:"image_hash,notnull"` // TODO Migrate to request metadata
	ExecutionTime                   time.Time     `sql:"execution_time"`
	ExecutedBy                      int           `sql:"executed_by,notnull"`
	SourceMetadataJson              string        `sql:"source_metadata_json"` // to have relevant info to process a scan for a given source type and subtype
	SourceType                      SourceType    `sql:"source_type"`
	SourceSubType                   SourceSubType `sql:"source_sub_type"`
	ScanToolExecutionHistoryMapping *ScanToolExecutionHistoryMapping
}

func (ed *ExecutionData) IsBuiltImage() bool {
	return ed.SourceType == SourceTypeImage && ed.SourceSubType == SourceSubTypeCi
}

func (ed *ExecutionData) IsManifestImage() bool {
	return ed.SourceType == SourceTypeImage && ed.SourceSubType == SourceSubTypeManifest
}

func (ed *ExecutionData) IsManifest() bool {
	return ed.SourceType == SourceTypeCode && ed.SourceSubType == SourceSubTypeManifest
}

func (ed *ExecutionData) IsCode() bool {
	return ed.SourceType == SourceTypeCode && ed.SourceSubType == SourceSubTypeCi
}

func (ed *ExecutionData) ContainsType(typeToCheck ResourceScanType) bool {
	for _, scanType := range ed.Types {
		if scanType == int(typeToCheck) {
			return true
		}
	}
	return false
}

type ExecutionData struct {
	Image         string
	ScanDataJson  string
	StartedOn     time.Time
	ScanToolName  string
	SourceType    SourceType
	SourceSubType SourceSubType
	Types         []int `sql:"types" pg:",array"`
	Status        serverBean.ScanExecutionProcessState
}

// multiple history rows for one source event
type SourceType int

const (
	SourceTypeImage SourceType = 1
	SourceTypeCode  SourceType = 2
	SourceTypeSbom  SourceType = 3 // can be used in future for direct sbom scanning
)

type SourceSubType int

const (
	SourceSubTypeCi       SourceSubType = 1 // relevant for ci code(2,1) or ci built image(1,1)
	SourceSubTypeManifest SourceSubType = 2 // relevant for devtron app deployment manifest/helm app manifest(2,2) or images retrieved from manifest(1,2))
)

type ImageScanHistoryRepository interface {
	Save(model *ImageScanExecutionHistory) error
	FindAll() ([]*ImageScanExecutionHistory, error)
	FindOne(id int) (*ImageScanExecutionHistory, error)
	FindByImageAndDigest(imageDigest string, image string) (*ImageScanExecutionHistory, error)
	Update(model *ImageScanExecutionHistory) error
	FindByImage(image string) (*ImageScanExecutionHistory, error)
	FindByImages(images []string) ([]*ImageScanExecutionHistory, error)
	FindByIds(ids []int) ([]*ImageScanExecutionHistory, error)
	FetchWithHistoryIds(historyIds []int) ([]*ExecutionData, error)
}

type ImageScanHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewImageScanHistoryRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ImageScanHistoryRepositoryImpl {
	return &ImageScanHistoryRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl ImageScanHistoryRepositoryImpl) Save(model *ImageScanExecutionHistory) error {
	err := impl.dbConnection.Insert(model)
	return err
}

func (impl ImageScanHistoryRepositoryImpl) FindAll() ([]*ImageScanExecutionHistory, error) {
	var models []*ImageScanExecutionHistory
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl ImageScanHistoryRepositoryImpl) FindOne(id int) (*ImageScanExecutionHistory, error) {
	var model ImageScanExecutionHistory
	err := impl.dbConnection.Model(&model).
		Where("id = ?", id).Select()
	return &model, err
}

func (impl ImageScanHistoryRepositoryImpl) FindByImageAndDigest(imageDigest string, image string) (*ImageScanExecutionHistory, error) {
	var model ImageScanExecutionHistory
	err := impl.dbConnection.Model(&model).
		Where("image_hash = ?", imageDigest).
		Where("image = ?", image).
		Where("source_type is null or source_type = 0").
		Order("execution_time desc").Limit(1).Select()
	return &model, err
}

func (impl ImageScanHistoryRepositoryImpl) FindByImageDigests(digest []string) ([]*ImageScanExecutionHistory, error) {
	var models []*ImageScanExecutionHistory
	err := impl.dbConnection.Model(&models).
		Where("image_hash in (?)", pg.In(digest)).
		Where("source_type is null or source_type = 0").
		Order("execution_time desc").Select()
	return models, err
}

func (impl ImageScanHistoryRepositoryImpl) Update(team *ImageScanExecutionHistory) error {
	err := impl.dbConnection.Update(team)
	return err
}

func (impl ImageScanHistoryRepositoryImpl) FindByImage(image string) (*ImageScanExecutionHistory, error) {
	var model ImageScanExecutionHistory
	err := impl.dbConnection.Model(&model).
		Where("image = ?", image).
		Order("execution_time desc").
		Where("source_type is null or source_type = 0").
		Limit(1).
		Select()
	return &model, err
}

// TODO need to change the behavior to use image hash instead of image
func (impl ImageScanHistoryRepositoryImpl) FindByImages(images []string) ([]*ImageScanExecutionHistory, error) {
	var model []*ImageScanExecutionHistory
	err := impl.dbConnection.Model(&model).
		Where("image IN (?)", pg.In(images)).Where("source_type is null or source_type = 0").Select()
	if err == pg.ErrNoRows {
		return model, nil
	}
	return model, err
}

func (impl ImageScanHistoryRepositoryImpl) FindByIds(ids []int) ([]*ImageScanExecutionHistory, error) {
	var models = make([]*ImageScanExecutionHistory, 0)
	err := impl.dbConnection.Model(&models).
		Where("id IN (? )", ids).Select()
	return models, err
}

func (impl ImageScanHistoryRepositoryImpl) FetchWithHistoryIds(historyIds []int) ([]*ExecutionData, error) {
	var models []*ExecutionData
	query := " SELECT iseh.image,iseh.execution_time AS started_on,iseh.source_type,iseh.source_sub_type,rser.scan_data_json,stehm.state AS status,stm.name AS scan_tool_name, rser.types AS types" +
		" FROM image_scan_execution_history iseh " +
		" INNER JOIN resource_scan_execution_result rser ON iseh.id = rser.image_scan_execution_history_id " +
		" INNER JOIN scan_tool_execution_history_mapping stehm ON iseh.id = stehm.image_scan_execution_history_id " +
		" INNER JOIN scan_tool_metadata stm ON stehm.scan_tool_id = stm.id " +
		" WHERE iseh.id IN (?)" +
		" AND rser.format = ?" +
		" ORDER BY iseh.id"
	// order needed as multiple scans can be performed on a single resource, in runime we will only parse latest entry for each resource

	_, err := impl.dbConnection.Query(&models, query, pg.In(historyIds), Json)
	return models, err

}
