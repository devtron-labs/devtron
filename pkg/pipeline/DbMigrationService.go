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

package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type DbMigrationConfigBean struct {
	Id            int    `json:"id"`
	DbConfigId    int    `json:"dbConfigId"`
	PipelineId    int    `json:"pipelineId"`
	GitMaterialId int    `json:"gitMaterialId"`
	ScriptSource  string `json:"scriptSource"` //location of file in git. relative to git root
	MigrationTool string `json:"migrationTool"`
	Active        bool   `json:"active"`
	UserId        int32  `json:"-"`
}

type DbMigrationService interface {
	Save(bean *DbMigrationConfigBean) (*DbMigrationConfigBean, error)
	Update(bean *DbMigrationConfigBean) (*DbMigrationConfigBean, error)
	GetByPipelineId(pipelineId int) (*DbMigrationConfigBean, error)
}
type DbMigrationServiceImpl struct {
	logger                      *zap.SugaredLogger
	dbMigrationConfigRepository pipelineConfig.DbMigrationConfigRepository
}

func NewDbMogrationService(logger *zap.SugaredLogger,
	dbMigrationConfigRepository pipelineConfig.DbMigrationConfigRepository) *DbMigrationServiceImpl {
	return &DbMigrationServiceImpl{
		dbMigrationConfigRepository: dbMigrationConfigRepository,
		logger:                      logger,
	}
}

func (impl DbMigrationServiceImpl) Save(bean *DbMigrationConfigBean) (*DbMigrationConfigBean, error) {
	if valid := pipelineConfig.MigrationTool(bean.MigrationTool).IsValid(); !valid {
		return nil, fmt.Errorf("unsupported migration tool %s", bean.MigrationTool)
	}
	migrationConfig := impl.beanToModelAdaptor(bean)
	migrationConfig.AuditLog = models.AuditLog{
		UpdatedOn: time.Now(),
		CreatedOn: time.Now(),
		CreatedBy: bean.UserId,
		UpdatedBy: bean.UserId,
	}
	err := impl.dbMigrationConfigRepository.Save(migrationConfig)
	if err != nil {
		impl.logger.Errorw("error in saving db migration config", "cfg", bean, "err", err)
		return nil, err
	}
	bean.Id = migrationConfig.Id
	return bean, nil
}

func (impl DbMigrationServiceImpl) Update(bean *DbMigrationConfigBean) (*DbMigrationConfigBean, error) {
	if bean.MigrationTool != "" {
		if valid := pipelineConfig.MigrationTool(bean.MigrationTool).IsValid(); !valid {
			return nil, fmt.Errorf("unsupported migration tool %s", bean.MigrationTool)
		}
	}

	migrationConfig := impl.beanToModelAdaptor(bean)
	migrationConfig.AuditLog = models.AuditLog{
		UpdatedOn: time.Now(),
		UpdatedBy: bean.UserId,
	}
	err := impl.dbMigrationConfigRepository.Update(migrationConfig)
	if err != nil {
		impl.logger.Errorw("error in updating db migration config", "cfg", bean, "err", err)
		return nil, err
	}
	bean.Id = migrationConfig.Id
	return bean, nil
}

func (impl DbMigrationServiceImpl) GetByPipelineId(pipelineId int) (*DbMigrationConfigBean, error) {
	cfg, err := impl.dbMigrationConfigRepository.FindByPipelineId(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline db migration config", "id", pipelineId, "err", err)
		return nil, err
	}
	bean := &DbMigrationConfigBean{
		MigrationTool: string(cfg.MigrationTool),
		GitMaterialId: cfg.GitMaterialId,
		PipelineId:    cfg.PipelineId,
		ScriptSource:  cfg.ScriptSource,
		Active:        cfg.Active,
		DbConfigId:    cfg.DbConfigId,
		Id:            cfg.Id,
	}
	return bean, nil

}

func (impl DbMigrationServiceImpl) beanToModelAdaptor(bean *DbMigrationConfigBean) *pipelineConfig.DbMigrationConfig {

	migrationConfig := &pipelineConfig.DbMigrationConfig{
		Id:            bean.Id,
		DbConfigId:    bean.DbConfigId,
		Active:        bean.Active,
		ScriptSource:  bean.ScriptSource,
		PipelineId:    bean.PipelineId,
		GitMaterialId: bean.GitMaterialId,
		MigrationTool: pipelineConfig.MigrationTool(bean.MigrationTool),
	}
	return migrationConfig
}
