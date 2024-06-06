/*
 * Copyright (c) 2024. Devtron Inc.
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
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ResourceScanExecutionResult struct {
	tableName                   struct{}           `sql:"resource_scan_execution_result" pg:",discard_unknown_columns"`
	Id                          int                `sql:"id,pk"`
	ImageScanExecutionHistoryId int                `sql:"image_scan_execution_history_id"`
	ScanDataJson                string             `sql:"scan_data_json"`
	Format                      ResourceScanFormat `sql:"format"`
	Types                       []ResourceScanType `sql:"types"`
	ScanToolId                  int                `sql:"scan_tool_id"`
}

type ResourceScanFormat int

const (
	CycloneDxSbom ResourceScanFormat = 1 // SBOM
	TrivyJson                        = 2
	Json                             = 3
)

type ResourceScanType int

const (
	Vulnerabilities ResourceScanType = 1
	License                          = 2
	Config                           = 3
	Secrets                          = 4
)

type ResourceScanResultRepository interface {
	SaveInBatch(tx *pg.Tx, models []*ResourceScanExecutionResult) error
}

type ResourceScanResultRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewResourceScanResultRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ResourceScanResultRepositoryImpl {
	return &ResourceScanResultRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl ResourceScanResultRepositoryImpl) SaveInBatch(tx *pg.Tx, models []*ResourceScanExecutionResult) error {
	return tx.Insert(&models)
}
