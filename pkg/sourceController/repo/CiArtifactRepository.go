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

package repository

import (
	"time"

	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiArtifact struct {
	tableName            struct{}  `sql:"ci_artifact" pg:",discard_unknown_columns"`
	Id                   int       `sql:"id,pk"`
	PipelineId           int       `sql:"pipeline_id"` //id of the ci pipeline from which this webhook was triggered
	Image                string    `sql:"image,notnull"`
	ImageDigest          string    `sql:"image_digest,notnull"`
	MaterialInfo         string    `sql:"material_info"` //git material metadata json array string
	DataSource           string    `sql:"data_source,notnull"`
	WorkflowId           *int      `sql:"ci_workflow_id"`
	ParentCiArtifact     int       `sql:"parent_ci_artifact"`
	ScanEnabled          bool      `sql:"scan_enabled,notnull"`
	Scanned              bool      `sql:"scanned,notnull"`
	ExternalCiPipelineId int       `sql:"external_ci_pipeline_id"`
	IsArtifactUploaded   bool      `sql:"is_artifact_uploaded"`
	DeployedTime         time.Time `sql:"-"`
	Deployed             bool      `sql:"-"`
	Latest               bool      `sql:"-"`
	RunningOnParent      bool      `sql:"-"`
	//sql.AuditLog
}

type CiArtifactRepository interface {
	GetByImages(images []string) ([]*CiArtifact, error)
	GetByImageDigests(imageDigests []string) ([]*CiArtifact, error)
}

type CiArtifactRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiArtifactRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiArtifactRepositoryImpl {
	return &CiArtifactRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type Material struct {
	PluginID         string           `json:"plugin-id"`
	GitConfiguration GitConfiguration `json:"git-configuration"`
	ScmConfiguration ScmConfiguration `json:"scm-configuration"`
	Type             string           `json:"type"`
}

type ScmConfiguration struct {
	URL string `json:"url"`
}

type GitConfiguration struct {
	URL string `json:"url"`
}

type Modification struct {
	Revision     string            `json:"revision"`
	ModifiedTime string            `json:"modified-time"`
	Data         map[string]string `json:"data"`
	Author       string            `json:"author"`
	Message      string            `json:"message"`
	Branch       string            `json:"branch"`
	Tag          string            `json:"tag,omitempty"`
	WebhookData  WebhookData       `json:"webhookData,omitempty"`
}

type WebhookData struct {
	Id              int
	EventActionType string
	Data            map[string]string
}

type CiMaterialInfo struct {
	Material      Material       `json:"material"`
	Changed       bool           `json:"changed"`
	Modifications []Modification `json:"modifications"`
}

func (impl CiArtifactRepositoryImpl) GetByImages(images []string) ([]*CiArtifact, error) {
	var artifact []*CiArtifact
	err := impl.dbConnection.Model(&artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.image in (?) ", pg.In(images)).
		Select()
	return artifact, err
}

func (impl CiArtifactRepositoryImpl) GetByImageDigests(imageDigests []string) ([]*CiArtifact, error) {
	var artifact []*CiArtifact
	err := impl.dbConnection.Model(&artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.image_digest in (?) ", imageDigests).
		Select()
	return artifact, err
}
