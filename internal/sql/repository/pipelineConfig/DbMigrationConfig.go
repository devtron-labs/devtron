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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type MigrationTool string

const MIGRATION_TOOL_MIGRATE MigrationTool = "migrate"

func (t MigrationTool) IsValid() bool {
	types := map[string]MigrationTool{"migrate": MIGRATION_TOOL_MIGRATE}
	_, ok := types[string(t)]
	return ok
}

type DbMigrationConfig struct {
	tableName     struct{}      `sql:"db_migration_config" pg:",discard_unknown_columns"`
	Id            int           `sql:"id"`
	DbConfigId    int           `sql:"db_config_id"`
	PipelineId    int           `sql:"pipeline_id"`
	GitMaterialId int           `sql:"git_material_id"`
	ScriptSource  string        `sql:"script_source"` //location of file in git. relative to git root
	MigrationTool MigrationTool `sql:"migration_tool"`
	Active        bool          `sql:"active"`
	sql.AuditLog
	DbConfig    *repository.DbConfig
	GitMaterial *GitMaterial
}

type DbMigrationConfigRepository interface {
	Save(config *DbMigrationConfig) error
	FindByPipelineId(pipelineId int) (config *DbMigrationConfig, err error)
	Update(config *DbMigrationConfig) error
}

type DbMigrationConfigRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewDbMigrationConfigRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DbMigrationConfigRepositoryImpl {
	return &DbMigrationConfigRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl DbMigrationConfigRepositoryImpl) Save(config *DbMigrationConfig) error {
	return impl.dbConnection.Insert(config)
}
func (impl DbMigrationConfigRepositoryImpl) Update(config *DbMigrationConfig) error {
	_, err := impl.dbConnection.Model(config).WherePK().UpdateNotNull()
	return err
}

func (impl DbMigrationConfigRepositoryImpl) FindByPipelineId(pipelineId int) (config *DbMigrationConfig, err error) {
	config = &DbMigrationConfig{}
	err = impl.dbConnection.Model(config).
		Column("db_migration_config.*", "DbConfig", "GitMaterial", "GitMaterial.GitProvider").
		Where("db_migration_config.pipeline_id =?", pipelineId).
		Where("db_migration_config.active =? ", true).
		Select()
	return config, err
}
