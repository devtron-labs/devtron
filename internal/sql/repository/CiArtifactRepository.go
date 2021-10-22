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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type CiArtifact struct {
	tableName        struct{}  `sql:"ci_artifact" pg:",discard_unknown_columns"`
	Id               int       `sql:"id,pk"`
	PipelineId       int       `sql:"pipeline_id,notnull"` //id of the ci pipeline from which this webhook was triggered
	Image            string    `sql:"image,notnull"`
	ImageDigest      string    `sql:"image_digest,notnull"`
	MaterialInfo     string    `sql:"material_info"` //git material metadata json array string
	DataSource       string    `sql:"data_source,notnull"`
	WorkflowId       *int      `sql:"ci_workflow_id"`
	ParentCiArtifact int       `sql:"parent_ci_artifact"`
	ScanEnabled      bool      `sql:"scan_enabled,notnull"`
	Scanned          bool      `sql:"scanned,notnull"`
	DeployedTime     time.Time `sql:"-"`
	Deployed         bool      `sql:"-"`
	Latest           bool      `sql:"-"`
	models.AuditLog
}

type CiArtifactRepository interface {
	Save(artifact *CiArtifact) error
	Get(id int) (artifact *CiArtifact, err error)
	GetByWfId(wfId int) (artifact *CiArtifact, err error)
	GetArtifactsByCDPipeline(cdPipelineId int) ([]CiArtifact, error)
	FetchArtifactForRollback(cdPipelineId int) ([]CiArtifact, error)

	GetArtifactsByCDPipelineV2(cdPipelineId int) ([]CiArtifact, error)
	GetArtifactsByCDPipelineAndRunnerType(cdPipelineId int, runnerType bean.CdWorkflowType) ([]CiArtifact, error)
	SaveAll(artifacts []*CiArtifact) error
	GetArtifactsByCiPipelineId(ciPipelineId int) ([]CiArtifact, error)
	FinDByParentCiArtifactAndCiId(parentCiArtifact int, ciPipelineIds []int) ([]*CiArtifact, error)
	GetLatest(cdPipelineId int) (int, error)
	GetByImageDigest(imageDigest string) (artifact *CiArtifact, err error)
	GetByIds(ids []int) ([]*CiArtifact, error)
}

type CiArtifactRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiArtifactRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiArtifactRepositoryImpl {
	return &CiArtifactRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl CiArtifactRepositoryImpl) SaveAll(artifacts []*CiArtifact) error {
	err := impl.dbConnection.RunInTransaction(func(tx *pg.Tx) error {
		for _, ciArtifact := range artifacts {
			r, err := tx.Model(ciArtifact).Insert()
			if err != nil {
				return err
			}
			impl.logger.Debugf("total rows saved %d", r.RowsAffected())
		}
		return nil
	})
	return err
}

func (impl CiArtifactRepositoryImpl) Save(artifact *CiArtifact) error {
	return impl.dbConnection.Insert(artifact)
}

func (impl CiArtifactRepositoryImpl) Get(id int) (artifact *CiArtifact, err error) {
	artifact = &CiArtifact{Id: id}
	err = impl.dbConnection.Model(artifact).WherePK().Select()
	return artifact, err
}

func (impl CiArtifactRepositoryImpl) GetByWfId(wfId int) (*CiArtifact, error) {
	artifact := &CiArtifact{}
	err := impl.dbConnection.Model(artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.ci_workflow_id = ? ", wfId).
		Select()
	return artifact, err
}

//this method takes CD Pipeline id and Returns List of Artifacts Latest By last deployed
func (impl CiArtifactRepositoryImpl) GetArtifactsByCDPipeline(cdPipelineId int) ([]CiArtifact, error) {
	var artifactsA []CiArtifact
	var artifactsAB []CiArtifact

	queryFetchArtifacts := ""
	queryFetchArtifacts = "SELECT cia.id, cia.data_source, cia.image, cia.image_digest FROM ci_artifact cia" +
		" INNER JOIN ci_pipeline cp on cp.id=cia.pipeline_id" +
		" INNER JOIN pipeline p on p.ci_pipeline_id = cp.id" +
		" WHERE p.id= ? ORDER BY cia.id DESC"
	_, err := impl.dbConnection.Query(&artifactsA, queryFetchArtifacts, cdPipelineId)
	if err != nil {
		impl.logger.Debugw("Error", err)
		return nil, err
	}

	// fetching material info separately because it gives error with fetching other (check its json) - FIXME
	type Object struct {
		Id           int    `json:"id"`
		MaterialInfo string `json:"material_info"`
	}

	var artifactsB []Object
	var queryTemp string = "SELECT cia.id, cia.material_info FROM ci_artifact cia" +
		" INNER JOIN ci_pipeline cp on cp.id=cia.pipeline_id" +
		" INNER JOIN pipeline p on p.ci_pipeline_id = cp.id" +
		" WHERE p.id= ? ORDER BY cia.id DESC"
	_, err = impl.dbConnection.Query(&artifactsB, queryTemp, cdPipelineId)
	if err != nil {
		return nil, err
	}

	mapData := make(map[int]string)
	for _, a := range artifactsB {
		mapData[a.Id] = a.MaterialInfo
	}
	for _, a := range artifactsA {
		a.MaterialInfo = mapData[a.Id]
		artifactsAB = append(artifactsAB, a)
	}

	var artifactsDeployed []CiArtifact
	query := "" +
		" SELECT cia.id, pco.created_on as created_on" +
		" FROM ci_artifact cia" +
		" INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=cia.id" +
		" WHERE pco.pipeline_id = ? ORDER BY pco.ci_artifact_id DESC, pco.created_on ASC;"

	_, err = impl.dbConnection.Query(&artifactsDeployed, query, cdPipelineId)
	if err != nil {
		impl.logger.Debugw("Error", err)
		return nil, err
	}

	//find latest deployed entry
	latestObj := Object{}
	latestDeployedQuery := "SELECT cia.id FROM ci_artifact cia" +
		" INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=cia.id" +
		" WHERE pco.pipeline_id = ? ORDER BY pco.created_on DESC LIMIT 1"

	_, err = impl.dbConnection.Query(&latestObj, latestDeployedQuery, cdPipelineId)
	if err != nil {
		impl.logger.Debugw("Error", err)
		return nil, err
	}

	var artifactsAll []CiArtifact
	mapData2 := make(map[int]time.Time)
	for _, a := range artifactsDeployed {
		mapData2[a.Id] = a.CreatedOn
	}
	for _, a := range artifactsAB {
		if val, ok := mapData2[a.Id]; ok {
			a.Deployed = true
			a.DeployedTime = val
			if latestObj.Id == a.Id {
				a.Latest = true
			}
		}
		artifactsAll = append(artifactsAll, a)
		if len(artifactsAll) >= 10 {
			break
		}
	}
	return artifactsAll, err
}

func (impl CiArtifactRepositoryImpl) GetArtifactsByCDPipelineAndRunnerType(cdPipelineId int, runnerType bean.CdWorkflowType) ([]CiArtifact, error) {
	var artifactsA []CiArtifact
	var artifactsAB []CiArtifact

	queryFetchArtifacts := ""
	/*	queryFetchArtifacts = "SELECT cia.id, cia.data_source, cia.image FROM ci_artifact cia" +
		" INNER JOIN ci_pipeline cp on cp.id=cia.pipeline_id" +
		" INNER JOIN pipeline p on p.ci_pipeline_id = cp.id" +
		" INNER JOIN cd_workflow wf on wf.pipeline_id=p.id" +
		" INNER JOIN cd_workflow_runner wfr on wfr.cd_workflow_id = wf.id" +
		" WHERE p.id= ? and wfr.workflow_type = ? GROUP BY cia.id, cia.data_source, cia.image ORDER BY cia.id DESC"*/
	queryFetchArtifacts = "SELECT cia.id, cia.data_source, cia.image, cia.image_digest FROM cd_workflow_runner wfr" +
		" INNER JOIN cd_workflow wf on wf.id=wfr.cd_workflow_id" +
		" INNER JOIN pipeline p on p.id = wf.pipeline_id" +
		" INNER JOIN ci_artifact cia on cia.id=wf.ci_artifact_id" +
		" WHERE p.id= ? and wfr.workflow_type = ? and wfr.status = ?" +
		" GROUP BY cia.id, cia.data_source, cia.image, cia.image_digest ORDER BY cia.id DESC"
	_, err := impl.dbConnection.Query(&artifactsA, queryFetchArtifacts, cdPipelineId, runnerType, "Succeeded")
	if err != nil {
		impl.logger.Debugw("Error", err)
		return nil, err
	}

	// fetching material info separately because it gives error with fetching other (check its json) - FIXME
	type Object struct {
		Id           int    `json:"id"`
		MaterialInfo string `json:"material_info"`
	}

	var artifactsB []Object
	var queryTemp string = "SELECT cia.id, cia.material_info FROM ci_artifact cia" +
		" INNER JOIN ci_pipeline cp on cp.id=cia.pipeline_id" +
		" INNER JOIN pipeline p on p.ci_pipeline_id = cp.id" +
		" WHERE p.id= ? ORDER BY cia.id DESC"
	_, err = impl.dbConnection.Query(&artifactsB, queryTemp, cdPipelineId)
	if err != nil {
		return nil, err
	}

	mapData := make(map[int]string)
	for _, a := range artifactsB {
		mapData[a.Id] = a.MaterialInfo
	}
	for _, a := range artifactsA {
		a.MaterialInfo = mapData[a.Id]
		artifactsAB = append(artifactsAB, a)
	}

	var artifactsDeployed []CiArtifact
	query := "" +
		" SELECT cia.id, pco.created_on as created_on" +
		" FROM ci_artifact cia" +
		" INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=cia.id" +
		" WHERE pco.pipeline_id = ? ORDER BY pco.ci_artifact_id DESC, pco.created_on ASC;"

	_, err = impl.dbConnection.Query(&artifactsDeployed, query, cdPipelineId)
	if err != nil {
		impl.logger.Debugw("Error", err)
		return nil, err
	}

	//find latest deployed entry
	latestObj := Object{}
	latestDeployedQuery := "SELECT cia.id FROM ci_artifact cia" +
		" INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=cia.id" +
		" WHERE pco.pipeline_id = ? ORDER BY pco.created_on DESC LIMIT 1"

	_, err = impl.dbConnection.Query(&latestObj, latestDeployedQuery, cdPipelineId)
	if err != nil {
		impl.logger.Debugw("Error", err)
		return nil, err
	}

	var artifactsAll []CiArtifact
	mapData2 := make(map[int]time.Time)
	for _, a := range artifactsDeployed {
		mapData2[a.Id] = a.CreatedOn
	}
	for _, a := range artifactsAB {
		if val, ok := mapData2[a.Id]; ok {
			a.Deployed = true
			a.DeployedTime = val
			if latestObj.Id == a.Id {
				a.Latest = true
			}
		}
		artifactsAll = append(artifactsAll, a)
		if len(artifactsAll) >= 10 {
			break
		}
	}
	return artifactsAll, err
}

// return map of gitUrl:hash
func (info *CiArtifact) ParseMaterialInfo() (map[string]string, error) {
	if info.DataSource != "GOCD" && info.DataSource != "CI-RUNNER" && info.DataSource != "EXTERNAL" {
		return nil, fmt.Errorf("datasource: %s not supported", info.DataSource)
	}
	var ciMaterials []*CiMaterialInfo
	err := json.Unmarshal([]byte(info.MaterialInfo), &ciMaterials)
	scmMap := map[string]string{}
	for _, material := range ciMaterials {
		var url string
		if material.Material.Type == "git" {
			url = material.Material.GitConfiguration.URL
		} else if material.Material.Type == "scm" {
			url = material.Material.ScmConfiguration.URL
		} else {
			return nil, fmt.Errorf("unknown material type:%s ", material.Material.Type)
		}
		revision := material.Modifications[0].Revision
		url = strings.TrimSpace(url)
		scmMap[url] = revision
	}
	return scmMap, err
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

func (impl CiArtifactRepositoryImpl) FetchArtifactForRollback(cdPipelineId int) ([]CiArtifact, error) {
	var artifactResults []CiArtifact
	var artifacts []CiArtifact
	query := "SELECT cia.id, cia.data_source, cia.image, pco.created_on as created_on" +
		" FROM pipeline_config_override pco" +
		" INNER JOIN ci_artifact cia on cia.id = pco.ci_artifact_id" +
		" WHERE pco.pipeline_id=? and cia.id < (SELECT MAX(cia.id) FROM ci_artifact cia INNER JOIN pipeline_config_override pco on cia.id = pco.ci_artifact_id where pco.pipeline_id=?)" +
		" ORDER BY pco.created_on desc;"
	_, err := impl.dbConnection.Query(&artifacts, query, cdPipelineId, cdPipelineId)
	if err != nil {
		return artifacts, err
	}

	// fetching material info separately because it gives error with fetching other (check its json) - FIXME
	type Object struct {
		Id           int    `json:"id"`
		MaterialInfo string `json:"material_info"`
	}

	var artifactsTemp []Object
	queryTemp := "SELECT cia.id, cia.material_info FROM pipeline_config_override pco" +
		" INNER JOIN ci_artifact cia on cia.id = pco.ci_artifact_id" +
		" WHERE pco.pipeline_id=?"
	_, err = impl.dbConnection.Query(&artifactsTemp, queryTemp, cdPipelineId)
	if err != nil {
		return nil, err
	}

	mapData := make(map[int]string)
	for _, a := range artifactsTemp {
		mapData[a.Id] = a.MaterialInfo
	}
	for _, a := range artifacts {
		a.MaterialInfo = mapData[a.Id]
		artifactResults = append(artifactResults, a)
	}

	return artifactResults, nil
}

func (impl CiArtifactRepositoryImpl) GetArtifactsByCDPipelineV2(cdPipelineId int) ([]CiArtifact, error) {
	var artifacts []CiArtifact
	err := impl.dbConnection.
		Model(&artifacts).
		Column("ci_artifact.*").
		Join("INNER JOIN ci_pipeline cp on cp.id=ci_artifact.pipeline_id").
		Join("INNER JOIN pipeline p on p.ci_pipeline_id = cp.id").
		Where("p.id = ?", cdPipelineId).
		Where("p.deleted = ?", false).
		Order("ci_artifact.id DESC").
		Select()

	var artifactsDeployed []CiArtifact
	/*	err = impl.dbConnection.
		Model(&artifactsDeployed).
		Column("ci_artifact.id as id, pco.created_on as created_on").
		Join("INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=ci_artifact.id").
		Where("pco.pipeline_id = ?", cdPipelineId).
		Order("ci_artifact.id DESC").
		Select()*/

	query := "" +
		" SELECT cia.id, pco.created_on as created_on" +
		" FROM ci_artifact cia" +
		" INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=cia.id" +
		" WHERE pco.pipeline_id = ? ORDER BY pco.ci_artifact_id DESC;"

	_, err = impl.dbConnection.Query(&artifactsDeployed, query, cdPipelineId)
	if err != nil {
		impl.logger.Debugw("Error", err)
		return nil, err
	}

	var responseCiArtifacts []CiArtifact
	mapData := make(map[int]time.Time)
	for _, a := range artifactsDeployed {
		mapData[a.Id] = a.CreatedOn
	}

	for _, artifact := range artifacts {
		if val, ok := mapData[artifact.Id]; ok {
			artifact.Deployed = true
			artifact.DeployedTime = val
		}
		responseCiArtifacts = append(responseCiArtifacts, artifact)
	}

	return responseCiArtifacts, err
}

func GetCiMaterialInfo(materialInfo string, source string) ([]CiMaterialInfo, error) {
	if source != "GOCD" && source != "CI-RUNNER" && source != "EXTERNAL" {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	var ciMaterials []CiMaterialInfo
	err := json.Unmarshal([]byte(materialInfo), &ciMaterials)
	if err != nil {
		println("material info", materialInfo)
		println("unmarshal error for material info", "err", err)
	}

	return ciMaterials, err
}

func (impl CiArtifactRepositoryImpl) GetArtifactsByCiPipelineId(ciPipelineId int) ([]CiArtifact, error) {
	var artifacts []CiArtifact
	err := impl.dbConnection.
		Model(&artifacts).
		Column("ci_artifact.*").
		Join("INNER JOIN ci_pipeline cp on cp.id=ci_artifact.pipeline_id").
		Where("cp.id = ?", ciPipelineId).
		Where("cp.deleted = ?", false).
		Order("ci_artifact.id DESC").
		Select()

	return artifacts, err
}

func (impl CiArtifactRepositoryImpl) FinDByParentCiArtifactAndCiId(parentCiArtifact int, ciPipelineIds []int) ([]*CiArtifact, error) {
	var CiArtifacts []*CiArtifact
	err := impl.dbConnection.
		Model(&CiArtifacts).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr("parent_ci_artifact =?", parentCiArtifact).
				WhereOr("id = ?", parentCiArtifact)
			return q, nil
		}).
		Where("pipeline_id in (?)", pg.In(ciPipelineIds)).
		Select()
	return CiArtifacts, err

}

func (impl CiArtifactRepositoryImpl) GetLatest(cdPipelineId int) (int, error) {
	//find latest deployed entry
	type Object struct {
		Id           int    `json:"id"`
		MaterialInfo string `json:"material_info"`
	}
	latestObj := Object{}
	latestDeployedQuery := "SELECT cia.id FROM ci_artifact cia" +
		" INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=cia.id" +
		" WHERE pco.pipeline_id = ? ORDER BY pco.created_on DESC LIMIT 1"

	_, err := impl.dbConnection.Query(&latestObj, latestDeployedQuery, cdPipelineId)
	if err != nil {
		impl.logger.Debugw("Error", err)
		return 0, err
	}
	return latestObj.Id, nil
}

func (impl CiArtifactRepositoryImpl) GetByImageDigest(imageDigest string) (*CiArtifact, error) {
	artifact := &CiArtifact{}
	err := impl.dbConnection.Model(artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.image_digest = ? ", imageDigest).
		Order("ci_artifact.id desc").Limit(1).
		Select()
	return artifact, err
}

func (impl CiArtifactRepositoryImpl) GetByIds(ids []int) ([]*CiArtifact, error) {
	var artifact []*CiArtifact
	err := impl.dbConnection.Model(&artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.id in (?) ", pg.In(ids)).
		Select()
	return artifact, err
}
