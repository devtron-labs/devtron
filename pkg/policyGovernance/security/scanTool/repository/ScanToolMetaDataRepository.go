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

package repository

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ScanToolMetadata struct {
	tableName                struct{}            `sql:"scan_tool_metadata" pg:",discard_unknown_columns"`
	Id                       int                 `sql:"id,pk"`
	Name                     string              `sql:"name"`
	Version                  string              `sql:"version"`
	ServerBaseUrl            string              `sql:"server_base_url"`
	ResultDescriptorTemplate string              `sql:"result_descriptor_template"`
	ScanTarget               bean.ScanTargetType `sql:"scan_target"`
	Active                   bool                `sql:"active,notnull"`
	Deleted                  bool                `sql:"deleted,notnull"`
	ToolMetaData             string              `sql:"tool_metadata"`
	PluginId                 int                 `sql:"plugin_id"`
	IsPreset                 bool                `sql:"is_preset,notnull"`
	Url                      string              `sql:"url"`
	sql.AuditLog
}

type ScanToolMetadataRepository interface {
	FindActiveToolByScanTarget(scanTarget bean.ScanTargetType) (*ScanToolMetadata, error)
	FindByNameAndVersion(name, version string) (*ScanToolMetadata, error)
	FindActiveById(id int) (*ScanToolMetadata, error)
	Save(tx *pg.Tx, model *ScanToolMetadata) (*ScanToolMetadata, error)
	Update(model *ScanToolMetadata) (*ScanToolMetadata, error)
	MarkToolDeletedById(id int) error
	FindAllActiveTools() ([]*ScanToolMetadata, error)
	MarkToolAsActive(toolName, version string, tx *pg.Tx) error
	MarkOtherToolsInActive(toolName string, tx *pg.Tx, version string) error
	FindActiveTool() (*ScanToolMetadata, error)
	FindNameAndUrlById(id int) (string, string, error)
}

type ScanToolMetadataRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewScanToolMetadataRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *ScanToolMetadataRepositoryImpl {
	return &ScanToolMetadataRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo *ScanToolMetadataRepositoryImpl) FindActiveToolByScanTarget(scanTargetType bean.ScanTargetType) (*ScanToolMetadata, error) {
	var model ScanToolMetadata
	err := repo.dbConnection.Model(&model).Where("active = ?", true).
		Where("scan_target = ?", scanTargetType).
		Where("deleted = ?", false).Limit(1).Select()
	if err != nil {
		repo.logger.Errorw("error in getting active tool for scan target", "err", err, "scanTargetType", scanTargetType)
		return nil, err
	}
	return &model, nil
}

func (repo *ScanToolMetadataRepositoryImpl) FindByNameAndVersion(name, version string) (*ScanToolMetadata, error) {
	model := &ScanToolMetadata{}
	err := repo.dbConnection.Model(model).Where("active = ?", true).
		Where("name = ?", name).Where("version = ?", version).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting tool by name and version", "err", err, "name", name, "version", version)
		return nil, err
	}
	return model, nil
}

func (repo *ScanToolMetadataRepositoryImpl) FindActiveById(id int) (*ScanToolMetadata, error) {
	model := &ScanToolMetadata{}
	err := repo.dbConnection.Model(model).Where("id = ?", id).
		Where("active = ?", true).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting active by id", "err", err, "id", id)
		return nil, err
	}
	return model, nil
}

func (repo *ScanToolMetadataRepositoryImpl) Save(tx *pg.Tx, model *ScanToolMetadata) (*ScanToolMetadata, error) {
	if tx != nil {
		err := tx.Insert(model)
		if err != nil {
			repo.logger.Errorw("error in saving scan tool metadata using transaction", "model", model, "err", err)
			return nil, err
		}
	} else {
		err := repo.dbConnection.Insert(model)
		if err != nil {
			repo.logger.Errorw("error in saving scan tool metadata", "model", model, "err", err)
			return nil, err
		}
	}

	return model, nil
}

func (repo *ScanToolMetadataRepositoryImpl) Update(model *ScanToolMetadata) (*ScanToolMetadata, error) {
	err := repo.dbConnection.Update(model)
	if err != nil {
		repo.logger.Errorw("error in updating scan tool metadata", "err", err, "model", model)
		return nil, err
	}
	return model, nil
}

func (repo *ScanToolMetadataRepositoryImpl) MarkToolDeletedById(id int) error {
	model := &ScanToolMetadata{}
	_, err := repo.dbConnection.Model(model).Set("deleted = ?", true).
		Where("id = ?", id).Update()
	if err != nil {
		repo.logger.Errorw("error in marking tool entry deleted by id", "err", err, "id", id)
		return err
	}
	return nil
}
func (repo *ScanToolMetadataRepositoryImpl) FindAllActiveTools() ([]*ScanToolMetadata, error) {
	var models []*ScanToolMetadata
	err := repo.dbConnection.Model(&models).Where("active = ?", true).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting active tool for scan target", "err", err)
		return nil, err
	}
	return models, nil

}
func (repo *ScanToolMetadataRepositoryImpl) MarkToolAsActive(toolName, version string, tx *pg.Tx) error {
	model := &ScanToolMetadata{}
	_, err := tx.Model(model).Set("active = ?", true).Where("name = ?", toolName).Where("version = ?", version).Update()

	if err != nil {
		repo.logger.Errorw("error in marking tool active for scan target", "toolName", toolName, "err", err)
		return err
	}
	return nil
}
func (repo *ScanToolMetadataRepositoryImpl) MarkOtherToolsInActive(toolName string, tx *pg.Tx, version string) error {
	model := &ScanToolMetadata{}
	_, err := tx.Model(model).Set("active = ?", false).Where("name != ?", toolName).Where("version != ?", version).Update()

	if err != nil {
		repo.logger.Errorw("error in marking tool active for scan target", "toolName", toolName, "err", err)
		return err
	}
	return nil
}
func (repo *ScanToolMetadataRepositoryImpl) FindActiveTool() (*ScanToolMetadata, error) {
	model := &ScanToolMetadata{}
	err := repo.dbConnection.Model(model).Where("active = ?", true).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting active tool for scan target", "err", err)
		return nil, err
	}
	return model, nil

}

func (repo *ScanToolMetadataRepositoryImpl) FindNameAndUrlById(id int) (string, string, error) {
	model := &ScanToolMetadata{}
	err := repo.dbConnection.Model(model).Column("name", "url").Where("id = ?", id).Select()
	if err != nil {
		repo.logger.Errorw("error in getting tool name by id", "err", err, "id", id)
		return "", "", err
	}
	return model.Name, model.Url, nil
}
