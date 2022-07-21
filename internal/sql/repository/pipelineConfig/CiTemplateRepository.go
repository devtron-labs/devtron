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
	"github.com/juju/errors"
	"go.uber.org/zap"
)

type CiTemplate struct {
	tableName         struct{} `sql:"ci_template" pg:",discard_unknown_columns"`
	Id                int      `sql:"id"`
	AppId             int      `sql:"app_id"`             //foreign key of app
	DockerRegistryId  string   `sql:"docker_registry_id"` //foreign key of registry
	DockerRepository  string   `sql:"docker_repository"`
	DockerfilePath    string   `sql:"dockerfile_path"`
	Args              string   `sql:"args"` //json string format of map[string]string
	TargetPlatform    string   `sql:"target_platform,notnull"`
	BeforeDockerBuild string   `sql:"before_docker_build"` //json string  format of []*Task
	AfterDockerBuild  string   `sql:"after_docker_build"`  //json string  format of []*Task
	TemplateName      string   `sql:"template_name"`
	Version           string   `sql:"version"` //gocd etage
	Active            bool     `sql:"active,notnull"`
	GitMaterialId     int      `sql:"git_material_id"`
	sql.AuditLog
	App            *app.App
	DockerRegistry *repository.DockerArtifactStore
	GitMaterial    *GitMaterial
}

type CiTemplateRepository interface {
	Save(material *CiTemplate) error
	FindByAppId(appId int) (ciTemplate *CiTemplate, err error)
	Update(material *CiTemplate) error
	FindByDockerRegistryId(dockerRegistryId string) (ciTemplates []*CiTemplate, err error)
	FindNumberOfAppsWithDockerConfigured(appIds []int) (int, error)
}

type CiTemplateRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiTemplateRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiTemplateRepositoryImpl {
	return &CiTemplateRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl CiTemplateRepositoryImpl) Save(material *CiTemplate) error {
	return impl.dbConnection.Insert(material)
}

func (impl CiTemplateRepositoryImpl) Update(material *CiTemplate) error {
	r, err := impl.dbConnection.Model(material).WherePK().UpdateNotNull()
	impl.logger.Infof("total rows saved %d", r.RowsAffected())
	return err
}
func (impl CiTemplateRepositoryImpl) FindByAppId(appId int) (ciTemplate *CiTemplate, err error) {
	template := &CiTemplate{}
	err = impl.dbConnection.Model(template).
		Where("app_id =? ", appId).
		Column("ci_template.*", "App", "DockerRegistry").
		Select()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return template, err
}

func (impl CiTemplateRepositoryImpl) FindByDockerRegistryId(dockerRegistryId string) (ciTemplates []*CiTemplate, err error) {
	err = impl.dbConnection.Model(&ciTemplates).
		Where("docker_registry_id =? ", dockerRegistryId).
		Where("active = ?", true).
		Select()
	return ciTemplates, err
}

func (impl CiTemplateRepositoryImpl) FindNumberOfAppsWithDockerConfigured(appIds []int) (int, error) {
	var ciTemplates []*CiTemplate
	count, err := impl.dbConnection.
		Model(&ciTemplates).
		ColumnExpr("DISTINCT app_id").
		Where("app_id in (?)", pg.In(appIds)).
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}
