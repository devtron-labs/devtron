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

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slices"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type credentialsSource = string
type ArtifactsSourceType = string

const (
	GLOBAL_CONTAINER_REGISTRY credentialsSource = "global_container_registry"
)

// List of possible DataSource Type for an artifact
const (
	CI_RUNNER ArtifactsSourceType = "CI-RUNNER"
	WEBHOOK   ArtifactsSourceType = "EXTERNAL" // Currently in use instead of EXT
	PRE_CD    ArtifactsSourceType = "pre_cd"
	POST_CD   ArtifactsSourceType = "post_cd"
	POST_CI   ArtifactsSourceType = "post_ci"
	GOCD      ArtifactsSourceType = "GOCD"
	// deprecated; Handled for backward compatibility
	EXT ArtifactsSourceType = "ext"
	// PRE_CI is not a valid DataSource for an artifact
)

type CiArtifactWithExtraData struct {
	CiArtifact
	PayloadSchema      string
	TotalCount         int
	TriggeredBy        int32
	StartedOn          time.Time
	CdWorkflowRunnerId int
}

type CiArtifact struct {
	tableName             struct{}  `sql:"ci_artifact" pg:",discard_unknown_columns"`
	Id                    int       `sql:"id,pk"`
	PipelineId            int       `sql:"pipeline_id"` // id of the ci pipeline from which this webhook was triggered
	Image                 string    `sql:"image,notnull"`
	ImageDigest           string    `sql:"image_digest,notnull"`
	MaterialInfo          string    `sql:"material_info"`       // git material metadata json array string
	DataSource            string    `sql:"data_source,notnull"` // possible values -> (CI_RUNNER,EXTERNAL,post_ci,pre_cd,post_cd) CI_runner is for normal build ci
	WorkflowId            *int      `sql:"ci_workflow_id"`
	ParentCiArtifact      int       `sql:"parent_ci_artifact"`
	ScanEnabled           bool      `sql:"scan_enabled,notnull"`
	Scanned               bool      `sql:"scanned,notnull"`
	ExternalCiPipelineId  int       `sql:"external_ci_pipeline_id"`
	IsArtifactUploaded    bool      `sql:"is_artifact_uploaded"`
	CredentialsSourceType string    `sql:"credentials_source_type"`
	CredentialSourceValue string    `sql:"credentials_source_value"`
	ComponentId           int       `sql:"component_id"`
	DeployedTime          time.Time `sql:"-"`
	Deployed              bool      `sql:"-"`
	Latest                bool      `sql:"-"`
	RunningOnParent       bool      `sql:"-"`
	sql.AuditLog
}

func (artifact *CiArtifact) IsMigrationRequired() bool {
	validDataSourceTypeList := []string{CI_RUNNER, WEBHOOK, PRE_CD, POST_CD, POST_CI, GOCD}
	if slices.Contains(validDataSourceTypeList, artifact.DataSource) {
		return false
	}
	return true
}

func (artifact *CiArtifact) IsRegistryCredentialMapped() bool {
	return artifact.CredentialsSourceType == GLOBAL_CONTAINER_REGISTRY
}

func (c *CiArtifact) GetMaterialInfo() ([]CiMaterialInfo, error) {
	var ciMaterials []CiMaterialInfo
	err := json.Unmarshal([]byte(c.MaterialInfo), &ciMaterials)
	if err != nil {
		return nil, err
	}
	return ciMaterials, nil
}

type CiArtifactRepository interface {
	Save(artifact *CiArtifact) error
	Delete(artifact *CiArtifact) error

	// Get returns the CiArtifact of the given id.
	// Note: Use Get along with MigrateToWebHookDataSourceType. For webhook artifacts, migration is required for column DataSource from 'ext' to 'EXTERNAL'
	Get(id int) (artifact *CiArtifact, err error)
	GetArtifactParentCiAndWorkflowDetailsByIdsInDesc(ids []int) ([]*CiArtifact, error)
	GetByWfId(wfId int) (artifact *CiArtifact, err error)
	GetArtifactsByCDPipeline(cdPipelineId, limit int, parentId int, searchString string, parentType bean.WorkflowType) ([]*CiArtifact, error)
	GetArtifactsByCDPipelineV3(listingFilterOpts *bean.ArtifactsListFilterOptions, isApprovalNode bool) ([]*CiArtifact, int, error)
	GetLatestArtifactTimeByCiPipelineIds(ciPipelineIds []int) ([]*CiArtifact, error)
	GetLatestArtifactTimeByCiPipelineId(ciPipelineId int) (*CiArtifact, error)
	GetArtifactsByCDPipelineV2(cdPipelineId int) ([]CiArtifact, error)
	GetArtifactsByCDPipelineAndRunnerType(cdPipelineId int, runnerType bean.WorkflowType) ([]CiArtifact, error)
	SaveAll(artifacts []*CiArtifact) ([]*CiArtifact, error)
	GetArtifactsByCiPipelineId(ciPipelineId int) ([]CiArtifact, error)
	GetAllArtifactsForWfComponents(artifactIds, ciPipelineIds, externalCiPipelineIds, cdPipelineIds []int,
		searchArtifactTag string, offset, limit int) ([]CiArtifact, int, error)
	GetArtifactByImageTagAndAppId(imageTag string, appId int) (CiArtifact, error)
	GetArtifactsByCiPipelineIds(ciPipelineIds []int) ([]CiArtifact, error)
	FinDByParentCiArtifactAndCiId(parentCiArtifact int, ciPipelineIds []int) ([]*CiArtifact, error)
	GetLatest(cdPipelineId int) (int, error)
	GetByImageDigest(imageDigest string) (artifact *CiArtifact, err error)
	GetByImageAndDigestAndPipelineId(ctx context.Context, image string, imageDigest string, ciPipelineId int) (artifact *CiArtifact, err error)
	GetByIds(ids []int) ([]*CiArtifact, error)
	GetByIdsAndArtifactTag(ids []int, searchArtifactTag string, limit, offset int) ([]CiArtifact, int, error)
	GetArtifactByCdWorkflowId(cdWorkflowId int) (artifact *CiArtifact, err error)
	GetArtifactsByParentCiWorkflowId(parentCiWorkflowId int) ([]string, error)
	FetchArtifactsByCdPipelineIdV2(listingFilterOptions bean.ArtifactsListFilterOptions) ([]CiArtifactWithExtraData, int, error)
	FindArtifactByListFilter(listingFilterOptions *bean.ArtifactsListFilterOptions, isApprovalNode bool) ([]*CiArtifact, int, error)
	FetchApprovedArtifactsForRollback(listingFilterOptions bean.ArtifactsListFilterOptions) ([]CiArtifactWithExtraData, int, error)
	FindApprovedArtifactsWithFilter(listingFilterOpts *bean.ArtifactsListFilterOptions) ([]*CiArtifact, int, error)
	GetArtifactsByDataSourceAndComponentId(dataSource string, componentId int) ([]CiArtifact, error)
	FindCiArtifactByImagePaths(images []string) ([]CiArtifact, error)
	GetPreviousArtifactOfId(ctx context.Context, ciPipelineId int, artifactId int) (int, error)
	// MigrateToWebHookDataSourceType is used for backward compatibility. It'll migrate the deprecated DataSource type
	MigrateToWebHookDataSourceType(id int) error
	UpdateLatestTimestamp(artifactIds []int) error
	FindDeployedArtifactsOnPipeline(artifactsListingFilterOps *bean.CdNodeMaterialParams) ([]CiArtifact, int, error)
	FindArtifactsByCIPipelineId(artifactsListingFilterOps *bean.CiNodeMaterialParams) ([]CiArtifact, int, error)
	FindArtifactsByExternalCIPipelineId(artifactsListingFilterOps *bean.ExtCiNodeMaterialParams) ([]CiArtifact, int, error)
	FindArtifactsPendingForPromotion(request *bean.PromotionPendingNodeMaterialParams) ([]CiArtifact, int, error)
	IsArtifactAvailableForDeployment(pipelineId, parentPipelineId, artifactId int, parentStage, pluginStage string, deployStage string) (bool, error)
}

type CiArtifactRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiArtifactRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiArtifactRepositoryImpl {
	return &CiArtifactRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl CiArtifactRepositoryImpl) SaveAll(artifacts []*CiArtifact) ([]*CiArtifact, error) {
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
	return artifacts, err
}

func (impl CiArtifactRepositoryImpl) MigrateToWebHookDataSourceType(id int) error {
	_, err := impl.dbConnection.Model(&CiArtifact{}).
		Set("data_source = ?", WEBHOOK).
		Where("id = ?", id).
		Update()
	return err
}

func (impl CiArtifactRepositoryImpl) GetPreviousArtifactOfId(ctx context.Context, ciPipelineId int, artifactId int) (int, error) {
	_, span := otel.Tracer("CiArtifactRepository").Start(ctx, "GetPreviousArtifactOfId")
	defer span.End()
	artifact := &CiArtifact{}
	err := impl.dbConnection.Model(artifact).
		Column("ci_artifact.id").
		Where("ci_artifact.pipeline_id = ? ", ciPipelineId).
		Where("ci_artifact.id < ?", artifactId).
		Order("ci_artifact.id desc").Limit(1).Select()
	return artifact.Id, err
}
func (impl CiArtifactRepositoryImpl) UpdateLatestTimestamp(artifactIds []int) error {
	if len(artifactIds) == 0 {
		impl.logger.Debug("UpdateLatestTimestamp empty list of artifacts, not updating")
		return nil
	}
	_, err := impl.dbConnection.Model(&CiArtifact{}).
		Set("updated_on = ?", time.Now()).
		Where("id IN (?)", pg.In(artifactIds)).
		Update()
	return err
}

func (impl CiArtifactRepositoryImpl) Save(artifact *CiArtifact) error {
	return impl.dbConnection.Insert(artifact)
}
func (impl CiArtifactRepositoryImpl) Delete(artifact *CiArtifact) error {
	return impl.dbConnection.Delete(artifact)
}

func (impl CiArtifactRepositoryImpl) Get(id int) (artifact *CiArtifact, err error) {
	artifact = &CiArtifact{Id: id}
	err = impl.dbConnection.Model(artifact).WherePK().Select()
	return artifact, err
}

func (impl CiArtifactRepositoryImpl) GetArtifactParentCiAndWorkflowDetailsByIdsInDesc(ids []int) ([]*CiArtifact, error) {
	artifacts := make([]*CiArtifact, 0)
	if len(ids) == 0 {
		return artifacts, nil
	}

	err := impl.dbConnection.Model(&artifacts).
		Column("ci_artifact.id", "ci_artifact.ci_workflow_id", "ci_artifact.parent_ci_artifact", "ci_artifact.external_ci_pipeline_id", "ci_artifact.pipeline_id").
		Where("ci_artifact.id in (?)", pg.In(ids)).
		Order("ci_artifact.id DESC").
		Select()

	if err != nil {
		impl.logger.Errorw("failed to get artifact parent ci and workflow details",
			"ids", ids,
			"err", err)
		return nil, err
	}
	return artifacts, nil
}

func (impl CiArtifactRepositoryImpl) GetByWfId(wfId int) (*CiArtifact, error) {
	artifact := &CiArtifact{}
	err := impl.dbConnection.Model(artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.ci_workflow_id = ? ", wfId).
		Select()
	return artifact, err
}

// this method takes CD Pipeline id and Returns List of Artifacts Latest By last deployed
func (impl CiArtifactRepositoryImpl) GetArtifactsByCDPipeline(cdPipelineId, limit int, parentId int, searchString string, parentType bean.WorkflowType) ([]*CiArtifact, error) {
	artifacts := make([]*CiArtifact, 0, limit)
	searchStringFinal := "%" + searchString + "%"

	if parentType == bean.WEBHOOK_WORKFLOW_TYPE {
		// WEBHOOK type parent
		err := impl.dbConnection.Model(&artifacts).
			Column("ci_artifact.id", "ci_artifact.material_info", "ci_artifact.data_source", "ci_artifact.image", "ci_artifact.image_digest", "ci_artifact.scan_enabled", "ci_artifact.scanned").
			Where("ci_artifact.external_ci_pipeline_id = ?", parentId).
			Where("ci_artifact.image LIKE ?", searchStringFinal).
			Order("ci_artifact.id DESC").
			Limit(limit).
			Select()

		if err != nil {
			impl.logger.Errorw("error while fetching artifacts for cd pipeline from db",
				"cdPipelineId", cdPipelineId,
				"parentId", parentId,
				"err", err)

			return nil, err
		}

	} else if parentType == bean.CI_WORKFLOW_TYPE {
		// CI type parent
		err := impl.dbConnection.Model(&artifacts).
			Column("ci_artifact.id", "ci_artifact.material_info", "ci_artifact.data_source", "ci_artifact.image", "ci_artifact.image_digest", "ci_artifact.scan_enabled", "ci_artifact.scanned").
			Join("INNER JOIN ci_pipeline cp on cp.id=ci_artifact.pipeline_id").
			Join("INNER JOIN pipeline p on p.ci_pipeline_id = cp.id").
			Where("p.id = ?", cdPipelineId).
			Where("ci_artifact.image LIKE ?", searchStringFinal).
			Order("ci_artifact.id DESC").
			Limit(limit).
			Select()

		if err != nil {
			impl.logger.Errorw("error while fetching artifacts for cd pipeline from db",
				"cdPipelineId", cdPipelineId,
				"err", err)

			return nil, err
		}
	}

	artifactsDeployed := make([]*CiArtifact, 0, limit)
	query := "" +
		" SELECT cia.id, pco.created_on as created_on" +
		" FROM ci_artifact cia" +
		" INNER JOIN pipeline_config_override pco ON pco.ci_artifact_id=cia.id" +
		" WHERE pco.pipeline_id = ? ORDER BY pco.ci_artifact_id DESC, pco.created_on ASC" +
		" LIMIT ?;"

	_, err := impl.dbConnection.Query(&artifactsDeployed, query, cdPipelineId, limit)

	if err != nil {
		impl.logger.Errorw("error while fetching deployed artifacts for cd pipeline from db",
			"cdPipelineId", cdPipelineId,
			"err", err)

		return nil, err
	}

	// find latest deployed entry
	lastDeployedArtifactId := 0
	if len(artifactsDeployed) > 0 {
		createdOn := artifactsDeployed[0].CreatedOn

		for _, artifact := range artifactsDeployed {

			// Need artifact id of the most recent created one
			if createdOn.After(artifact.CreatedOn) {
				lastDeployedArtifactId = artifact.Id
				createdOn = artifact.CreatedOn
			}
		}
	}

	if err != nil {
		impl.logger.Errorw("error while fetching latest deployed artifact from db",
			"cdPipelineId", cdPipelineId,
			"err", err)

		return nil, err
	}

	artifactsAll := make([]*CiArtifact, 0, limit)
	mapData2 := make(map[int]time.Time)
	for _, a := range artifactsDeployed {
		mapData2[a.Id] = a.CreatedOn
	}
	for _, a := range artifacts {
		if val, ok := mapData2[a.Id]; ok {
			a.Deployed = true
			a.DeployedTime = val
			if lastDeployedArtifactId == a.Id {
				a.Latest = true
			}
		}
		artifactsAll = append(artifactsAll, a)
	}
	return artifactsAll, err
}

func (impl CiArtifactRepositoryImpl) GetArtifactsByCDPipelineV3(listingFilterOpts *bean.ArtifactsListFilterOptions, isApprovalNode bool) ([]*CiArtifact, int, error) {

	if listingFilterOpts.ParentStageType != bean.CI_WORKFLOW_TYPE && listingFilterOpts.ParentStageType != bean.WEBHOOK_WORKFLOW_TYPE {
		return nil, 0, nil
	}

	artifactsResp := make([]*CiArtifactWithExtraData, 0, listingFilterOpts.Limit)
	var artifacts []*CiArtifact
	totalCount := 0
	finalQuery := BuildQueryForParentTypeCIOrWebhook(*listingFilterOpts, isApprovalNode)
	_, err := impl.dbConnection.Query(&artifactsResp, finalQuery)
	if err != nil {
		return nil, totalCount, err
	}
	artifacts = make([]*CiArtifact, len(artifactsResp))
	for i, _ := range artifactsResp {
		artifacts[i] = &artifactsResp[i].CiArtifact
		totalCount = artifactsResp[i].TotalCount
	}

	if len(artifacts) == 0 {
		return artifacts, totalCount, nil
	}
	artifacts, err = impl.setDeployedDataInArtifacts(listingFilterOpts.PipelineId, artifacts)
	return artifacts, totalCount, err
}

func (impl CiArtifactRepositoryImpl) setDeployedDataInArtifacts(pipelineId int, artifacts []*CiArtifact) ([]*CiArtifact, error) {
	// processing
	artifactsMap := make(map[int]*CiArtifact)
	artifactsIds := make([]int, 0, len(artifacts))
	for _, artifact := range artifacts {
		artifactsMap[artifact.Id] = artifact
		artifactsIds = append(artifactsIds, artifact.Id)
	}

	// (this will fetch all the artifacts that were deployed on the given pipeline atleast once in new->old deployed order)
	artifactsDeployed := make([]*CiArtifact, 0, len(artifactsIds))
	query := "SELECT pco.ci_artifact_id AS id,pco.created_on AS created_on " +
		" FROM pipeline_config_override pco " +
		" WHERE pco.pipeline_id = ? " +
		" AND pco.ci_artifact_id IN (?) "

	_, err := impl.dbConnection.Query(&artifactsDeployed, query, pipelineId, pg.In(artifactsIds))
	if err != nil {
		return artifacts, err
	}

	// set deployed time and latest deployed artifact
	for _, deployedArtifact := range artifactsDeployed {
		artifactId := deployedArtifact.Id
		if _, ok := artifactsMap[artifactId]; ok {
			artifactsMap[artifactId].Deployed = true
			artifactsMap[artifactId].DeployedTime = deployedArtifact.CreatedOn
		}
	}

	return artifacts, nil
}

func (impl CiArtifactRepositoryImpl) GetLatestArtifactTimeByCiPipelineIds(ciPipelineIds []int) ([]*CiArtifact, error) {
	artifacts := make([]*CiArtifact, 0)
	query := "select cws.pipeline_id, cws.created_on from " +
		"(SELECT  pipeline_id, MAX(created_on) created_on " +
		"FROM ci_artifact " +
		"GROUP BY pipeline_id) cws " +
		"where cws.pipeline_id IN (" + helper.GetCommaSepratedString(ciPipelineIds) + "); "

	_, err := impl.dbConnection.Query(&artifacts, query)
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

// GetLatestArtifactTimeByCiPipelineId will fetch latest ci artifact time(created) against that ci pipeline
func (impl CiArtifactRepositoryImpl) GetLatestArtifactTimeByCiPipelineId(ciPipelineId int) (*CiArtifact, error) {
	artifacts := &CiArtifact{}
	query := "select cws.pipeline_id, cws.created_on from " +
		"(SELECT  pipeline_id, MAX(created_on) created_on " +
		"FROM ci_artifact " +
		"GROUP BY pipeline_id) cws " +
		"where cws.pipeline_id = ? ; "

	_, err := impl.dbConnection.Query(artifacts, query, ciPipelineId)
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

func (impl CiArtifactRepositoryImpl) GetArtifactsByCDPipelineAndRunnerType(cdPipelineId int, runnerType bean.WorkflowType) ([]CiArtifact, error) {
	var artifactsA []CiArtifact
	var artifactsAB []CiArtifact

	queryFetchArtifacts := ""
	/*	queryFetchArtifacts = "SELECT cia.id, cia.data_source, cia.image FROM ci_artifact cia" +
		" INNER JOIN ci_pipeline cp on cp.id=cia.pipeline_id" +
		" INNER JOIN pipeline p on p.ci_pipeline_id = cp.id" +
		" INNER JOIN cd_workflow wf on wf.pipeline_id=p.id" +
		" INNER JOIN cd_workflow_runner wfr on wfr.cd_workflow_id = wf.id" +
		" WHERE p.id= ? and wfr.workflow_type = ? GROUP BY cia.id, cia.data_source, cia.image ORDER BY cia.id DESC"*/

	// this query gets details for status = Succeeded, this status is only valid
	// for pre stages & post stages, for deploy stage status will be healthy, degraded, aborted, missing etc
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

	// find latest deployed entry
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
func (artifact *CiArtifact) ParseMaterialInfo() (map[string]string, error) {
	if artifact.DataSource != GOCD && artifact.DataSource != CI_RUNNER && artifact.DataSource != WEBHOOK && artifact.DataSource != EXT {
		return nil, fmt.Errorf("datasource: %s not supported", artifact.DataSource)
	}
	var ciMaterials []*CiMaterialInfo
	err := json.Unmarshal([]byte(artifact.MaterialInfo), &ciMaterials)
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
	if source != GOCD && source != CI_RUNNER && source != WEBHOOK && source != POST_CI && source != PRE_CD && source != POST_CD && source != EXT {
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

func (impl CiArtifactRepositoryImpl) GetAllArtifactsForWfComponents(artifactIds, ciPipelineIds, externalCiPipelineIds, cdPipelineIds []int,
	searchArtifactTag string, offset, limit int) ([]CiArtifact, int, error) {
	var artifacts []CiArtifact
	query := impl.dbConnection.Model(&artifacts).Column("ci_artifact.*")
	query = query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		if len(ciPipelineIds) > 0 {
			query = query.Where("ci_artifact.pipeline_id IN (?)", pg.In(ciPipelineIds))
		} else if len(externalCiPipelineIds) > 0 {
			query = query.Where("ci_artifact.external_ci_pipeline_id IN (?)", pg.In(externalCiPipelineIds))
		}
		if len(cdPipelineIds) > 0 {
			query = query.
				WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
					query = query.
						Where("ci_artifact.component_id in (?)", pg.In(cdPipelineIds)).
						WhereGroup(func(query *orm.Query) (*orm.Query, error) {
							query = query.
								WhereOr("ci_artifact.data_source = ?", PRE_CD).
								WhereOr("ci_artifact.data_source = ?", POST_CD)
							return query, nil
						})
					return query, nil
				}).
				WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
					subQuery := impl.dbConnection.
						Model().
						Table("cd_workflow").
						ColumnExpr("DISTINCT cd_workflow.ci_artifact_id").
						Where("cd_workflow.pipeline_id IN (?)", pg.In(cdPipelineIds))
					query = query.
						Where("ci_artifact.id in (?)", subQuery)
					return query, nil
				})
		}
		return query, nil
	})
	if len(searchArtifactTag) > 0 {
		query = query.Where("ci_artifact.image LIKE ?", fmt.Sprint("%:%", searchArtifactTag, "%"))
	}
	if len(artifactIds) > 0 {
		query = query.Where("ci_artifact.id in (?)", pg.In(artifactIds))
	}
	totalQuery, err := query.Count()
	if err != nil {
		return nil, 0, err
	}
	err = query.
		Order("ci_artifact.id DESC").
		Offset(offset).
		Limit(limit).
		Select()

	return artifacts, totalQuery, err
}

func (impl CiArtifactRepositoryImpl) GetArtifactByImageTagAndAppId(imageTag string, appId int) (CiArtifact, error) {
	var artifact CiArtifact
	err := impl.dbConnection.Model(&artifact).
		Column("ci_artifact.*").
		Join("INNER JOIN release_tags rt").
		JoinOn("rt.artifact_id = ci_artifact.id").
		Where("rt.app_id IN (?)", appId).
		Where("rt.tag_name = ?", imageTag).
		Order("ci_artifact.id DESC").
		Limit(1).
		Select()
	return artifact, err
}

func (impl CiArtifactRepositoryImpl) GetArtifactsByCiPipelineIds(ciPipelineIds []int) ([]CiArtifact, error) {
	var artifacts []CiArtifact
	if len(ciPipelineIds) == 0 {
		impl.logger.Debug("GetArtifactsByCiPipelineIds empty list of ids, returning empty list of artifacts")
		return artifacts, nil
	}
	err := impl.dbConnection.
		Model(&artifacts).
		Column("ci_artifact.*").
		Join("INNER JOIN ci_pipeline cp on cp.id=ci_artifact.pipeline_id").
		Where("cp.id in (?)", pg.In(ciPipelineIds)).
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
	// find latest deployed entry
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

func (impl CiArtifactRepositoryImpl) GetByImageAndDigestAndPipelineId(ctx context.Context, image string, imageDigest string, ciPipelineId int) (*CiArtifact, error) {
	_, span := otel.Tracer("CiArtifactRepository").Start(ctx, "GetByImageAndDigestAndPipelineId")
	defer span.End()
	artifact := &CiArtifact{}
	err := impl.dbConnection.Model(artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.image_digest = ? ", imageDigest).
		Where("ci_artifact.image = ?", image).
		Where("ci_artifact.pipeline_id = ?", ciPipelineId).
		Order("ci_artifact.id desc").Limit(1).
		Select()
	return artifact, err
}

func (impl CiArtifactRepositoryImpl) GetByIds(ids []int) ([]*CiArtifact, error) {
	if len(ids) == 0 {
		return nil, pg.ErrNoRows
	}
	var artifact []*CiArtifact
	err := impl.dbConnection.Model(&artifact).
		Column("ci_artifact.*").
		Where("ci_artifact.id in (?) ", pg.In(ids)).
		Select()
	return artifact, err
}

func (impl CiArtifactRepositoryImpl) GetByIdsAndArtifactTag(ids []int, searchArtifactTag string, limit, offset int) ([]CiArtifact, int, error) {
	if len(ids) == 0 {
		return nil, 0, pg.ErrNoRows
	}
	var artifacts []CiArtifact
	query := impl.dbConnection.Model(&artifacts).
		Column("ci_artifact.*").
		Where("ci_artifact.id in (?) ", pg.In(ids))
	if len(searchArtifactTag) > 0 {
		query = query.Where("ci_artifact.image LIKE ?", fmt.Sprint("%:%", searchArtifactTag, "%"))
	}
	totalCount, err := query.Count()
	if err != nil {
		return nil, 0, err
	}
	err = query.Limit(limit).Offset(offset).Select()
	return artifacts, totalCount, err
}

func (impl CiArtifactRepositoryImpl) GetArtifactByCdWorkflowId(cdWorkflowId int) (artifact *CiArtifact, err error) {
	artifact = &CiArtifact{}
	err = impl.dbConnection.Model(artifact).
		Column("ci_artifact.*").
		Join("INNER JOIN cd_workflow cdwf on cdwf.ci_artifact_id = ci_artifact.id").
		Where("cdwf.id = ? ", cdWorkflowId).
		Select()
	return artifact, err
}

// GetArtifactsByParentCiWorkflowId will get all artifacts of child workflow sorted by descending order to fetch latest at top, child workflow required for handling container image polling plugin as there can be multiple images from a single parent workflow, which are accommodated in child workflows
func (impl CiArtifactRepositoryImpl) GetArtifactsByParentCiWorkflowId(parentCiWorkflowId int) ([]string, error) {
	var artifacts []string
	query := "SELECT cia.image FROM ci_artifact cia where cia.ci_workflow_id in (SELECT wf.id from ci_workflow wf where wf.parent_ci_workflow_id = ? ) ORDER BY cia.created_on DESC ;"
	_, err := impl.dbConnection.Query(&artifacts, query, parentCiWorkflowId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching artifacts for parent ci workflow id", "err", err)
		return nil, err
	}
	return artifacts, err
}

func (impl CiArtifactRepositoryImpl) FindArtifactByListFilter(listingFilterOptions *bean.ArtifactsListFilterOptions, isApprovalNode bool) ([]*CiArtifact, int, error) {

	var ciArtifactsResp []CiArtifactWithExtraData
	ciArtifacts := make([]*CiArtifact, 0)
	totalCount := 0
	finalQuery := BuildQueryForArtifactsForCdStage(*listingFilterOptions, isApprovalNode)
	_, err := impl.dbConnection.Query(&ciArtifactsResp, finalQuery)
	if err == pg.ErrNoRows || len(ciArtifactsResp) == 0 {
		return ciArtifacts, totalCount, nil
	}

	if listingFilterOptions.UseCdStageQueryV2 {
		for i, _ := range ciArtifactsResp {
			cia := ciArtifactsResp[i]
			totalCount = cia.TotalCount
			ciArtifacts = append(ciArtifacts, &cia.CiArtifact)
		}
		ciArtifacts, err = impl.setDeployedDataInArtifacts(listingFilterOptions.PipelineId, ciArtifacts)
		return ciArtifacts, totalCount, err
	}

	artifactIds := make([]int, len(ciArtifactsResp))
	for i, af := range ciArtifactsResp {
		artifactIds[i] = af.Id
		totalCount = af.TotalCount
	}

	if len(artifactIds) > 0 {
		err = impl.dbConnection.
			Model(&ciArtifacts).
			Where("id IN (?) ", pg.In(artifactIds)).
			Select()
		if err != nil {
			if err == pg.ErrNoRows {
				err = nil
			}
			return ciArtifacts, totalCount, err
		}
	}

	ciArtifacts, err = impl.setDeployedDataInArtifacts(listingFilterOptions.PipelineId, ciArtifacts)
	return ciArtifacts, totalCount, err
}

func (impl CiArtifactRepositoryImpl) FetchArtifactsByCdPipelineIdV2(listingFilterOptions bean.ArtifactsListFilterOptions) ([]CiArtifactWithExtraData, int, error) {
	var wfrList []CiArtifactWithExtraData
	totalCount := 0
	finalQuery := BuildQueryForArtifactsForRollback(listingFilterOptions)
	_, err := impl.dbConnection.Query(&wfrList, finalQuery)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting Wfrs and ci artifacts by pipelineId", "err", err, "pipelineId", listingFilterOptions.PipelineId)
		return nil, totalCount, err
	}
	if len(wfrList) > 0 {
		totalCount = wfrList[0].TotalCount
	}
	return wfrList, totalCount, nil
}

func (impl CiArtifactRepositoryImpl) FindApprovedArtifactsWithFilter(listingFilterOpts *bean.ArtifactsListFilterOptions) ([]*CiArtifact, int, error) {
	artifacts := make([]*CiArtifactWithExtraData, 0)
	totalCount := 0

	finalQuery := BuildApprovedOnlyArtifactsWithFilter(*listingFilterOpts)
	_, err := impl.dbConnection.Query(&artifacts, finalQuery)
	artifactsResp := make([]*CiArtifact, len(artifacts))
	for i, _ := range artifacts {
		artifactsResp[i] = &artifacts[i].CiArtifact
		totalCount = artifacts[i].TotalCount
	}
	return artifactsResp, totalCount, err
}

func (impl CiArtifactRepositoryImpl) FetchApprovedArtifactsForRollback(listingFilterOpts bean.ArtifactsListFilterOptions) ([]CiArtifactWithExtraData, int, error) {
	var results []CiArtifactWithExtraData
	totalCount := 0
	finalQuery := BuildQueryForApprovedArtifactsForRollback(listingFilterOpts)
	_, err := impl.dbConnection.Query(&results, finalQuery)
	if err != nil && err != pg.ErrNoRows {
		return results, totalCount, err
	}
	if len(results) > 0 {
		totalCount = results[0].TotalCount
	}
	return results, totalCount, nil
}

func (impl CiArtifactRepositoryImpl) GetArtifactsByDataSourceAndComponentId(dataSource string, componentId int) ([]CiArtifact, error) {
	var ciArtifacts []CiArtifact
	err := impl.dbConnection.
		Model(&ciArtifacts).
		Where(" data_source=? and component_id=? ", dataSource, componentId).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ci artifacts by data_source and component_id")
		return ciArtifacts, err
	}
	return ciArtifacts, nil
}

func (impl CiArtifactRepositoryImpl) FindCiArtifactByImagePaths(images []string) ([]CiArtifact, error) {
	var ciArtifacts []CiArtifact
	err := impl.dbConnection.
		Model(&ciArtifacts).
		Where(" image in (?) ", pg.In(images)).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ci artifacts by data_source and component_id")
		return ciArtifacts, err
	}
	return ciArtifacts, nil
}

func (impl CiArtifactRepositoryImpl) FindDeployedArtifactsOnPipeline(artifactsListingFilterOps *bean.CdNodeMaterialParams) ([]CiArtifact, int, error) {

	var ciArtifacts []CiArtifact
	var ciArtifactsResp []CiArtifactWithExtraData

	query := fmt.Sprintf(" select ci_artifact.*, count(ci_artifact.id) over() as total_count from ci_artifact where ci_artifact.id in "+
		"( select distinct(cdw.ci_artifact_id) from cd_workflow cdw inner join cd_workflow_runner cdwr ON cdw.id = cdwr.cd_workflow_id and cdw.pipeline_id = %d  and cdwr.workflow_type = 'DEPLOY' and cdwr.status IN ('Healthy','Succeeded')  )", artifactsListingFilterOps.GetCDPipelineId())

	listingOptions := artifactsListingFilterOps.GetListingOptions()
	searchRegex := listingOptions.GetSearchStringRegex()
	if searchRegex != EmptyLikeRegex {
		query = query + fmt.Sprintf(" and ci_artifact.image like '%s' ", searchRegex)
	}

	limitOffSetQuery := fmt.Sprintf(" order by ci_artifact.id desc LIMIT %v OFFSET %v", listingOptions.Limit, listingOptions.Offset)
	query = query + limitOffSetQuery

	_, err := impl.dbConnection.Query(&ciArtifactsResp, query)
	if err != nil {
		return ciArtifacts, 0, err
	}

	var totalCount int
	if len(ciArtifactsResp) > 0 {
		totalCount = ciArtifactsResp[0].TotalCount
	}
	for _, ciArtifact := range ciArtifactsResp {
		ciArtifacts = append(ciArtifacts, ciArtifact.CiArtifact)
	}
	return ciArtifacts, totalCount, nil
}

func (impl CiArtifactRepositoryImpl) FindArtifactsByCIPipelineId(request *bean.CiNodeMaterialParams) ([]CiArtifact, int, error) {

	query := impl.dbConnection.Model((*CiArtifact)(nil)).
		Column("ci_artifact.*").ColumnExpr("COUNT(ci_artifact.id) OVER() AS total_count").
		Join("join ci_pipeline ON ( ci_pipeline.id = ci_artifact.pipeline_id OR (ci_pipeline.id=ci_artifact.component_id AND ci_artifact.data_source= ? ) )", POST_CI).
		Where("ci_pipeline.active=true and ci_pipeline.id = ? ", request.GetCiPipelineId())

	listingOptions := request.GetListingOptions()
	if len(listingOptions.SearchString) > 0 {
		query = query.Where("ci_artifact.image like ?", listingOptions.GetSearchStringRegex())
	}
	query = query.OrderExpr("ci_artifact.id DESC ").Limit(listingOptions.Limit).Offset(listingOptions.Offset)
	ciArtifacts, totalCount, err := impl.executePromotionNodeQuery(query)
	if err != nil {
		return ciArtifacts, 0, err
	}
	return ciArtifacts, totalCount, nil
}

func (impl CiArtifactRepositoryImpl) FindArtifactsByExternalCIPipelineId(request *bean.ExtCiNodeMaterialParams) ([]CiArtifact, int, error) {

	query := impl.dbConnection.Model((*CiArtifact)(nil)).
		Column("ci_artifact.*").ColumnExpr("COUNT(ci_artifact.id) OVER() AS total_count").
		Join("join external_ci_pipeline on external_ci_pipeline.id = ci_artifact.external_ci_pipeline_id ").
		Where("external_ci_pipeline.active=true and external_ci_pipeline.id = ? ", request.GetExtCiPipelineId())

	listingOptions := request.GetListingOptions()
	if len(listingOptions.SearchString) > 0 {
		query = query.Where("ci_artifact.image like ?", listingOptions.GetSearchStringRegex())
	}
	query = query.OrderExpr("ci_artifact.id DESC").Limit(listingOptions.Limit).Offset(listingOptions.Offset)
	ciArtifacts, totalCount, err := impl.executePromotionNodeQuery(query)
	if err != nil {
		return ciArtifacts, 0, err
	}
	return ciArtifacts, totalCount, nil

}

func (impl CiArtifactRepositoryImpl) FindArtifactsPendingForPromotion(request *bean.PromotionPendingNodeMaterialParams) ([]CiArtifact, int, error) {

	if len(request.GetCDPipelineIds()) == 0 {
		return []CiArtifact{}, 0, nil
	}
	awaitingRequestQuery := impl.dbConnection.Model((*repository.ArtifactPromotionApprovalRequest)(nil)).
		ColumnExpr("distinct(artifact_id)").
		Where("destination_pipeline_id in (?) and status = ? ", pg.In(request.GetCDPipelineIds()), constants.AWAITING_APPROVAL)

	excludedArtifactIds := request.GetExcludedArtifactIds()
	if len(excludedArtifactIds) > 0 {
		awaitingRequestQuery = awaitingRequestQuery.Where("artifact_id not in (?)", pg.In(excludedArtifactIds))
	}

	query := impl.dbConnection.Model((*CiArtifact)(nil)).
		Column("ci_artifact.*").ColumnExpr("COUNT(id) OVER() AS total_count").
		Where("ci_artifact.id in ( ? ) ", awaitingRequestQuery)

	listingOptions := request.GetListingOptions()
	if len(listingOptions.SearchString) > 0 {
		query = query.Where("ci_artifact.image like ?", listingOptions.GetSearchStringRegex())
	}
	query = query.OrderExpr("ci_artifact.id DESC").Limit(listingOptions.Limit).Offset(listingOptions.Offset)

	ciArtifacts, totalCount, err := impl.executePromotionNodeQuery(query)
	if err != nil {
		return ciArtifacts, 0, err
	}
	return ciArtifacts, totalCount, nil

}

func (impl CiArtifactRepositoryImpl) executePromotionNodeQuery(query *orm.Query) ([]CiArtifact, int, error) {
	var ciArtifactArrResp []CiArtifactWithExtraData
	var ciArtifacts []CiArtifact
	err := query.Select(&ciArtifactArrResp)
	if err != nil {
		return ciArtifacts, 0, err
	}
	var totalCount int
	if len(ciArtifactArrResp) > 0 {
		totalCount = ciArtifactArrResp[0].TotalCount
	}
	for _, ciArtifactResp := range ciArtifactArrResp {
		ciArtifacts = append(ciArtifacts, ciArtifactResp.CiArtifact)
	}
	return ciArtifacts, totalCount, nil
}

func (impl CiArtifactRepositoryImpl) IsArtifactAvailableForDeployment(pipelineId, parentPipelineId, artifactId int, parentStage, pluginStage string, deployStage string) (bool, error) {
	//1) check if artifact is present on current pipeline
	//2) check if artifact is successfully deployed on parent pipeline
	//3) check if artifact is generated by parent plugin (skopeo)
	//4) check if artifact is promoted to this environment
	var Available bool
	query := fmt.Sprintf(
		`SELECT count(*) > 0 as Available
    FROM ci_artifact 
    WHERE (
    ( --  checking if artifact is successfully deployed at parent or is deployed on current cd 
        ci_artifact.id = %d -- doing this for optimization, only checking for current artifact 
        AND id IN (
            SELECT DISTINCT(cd_workflow.ci_artifact_id) as ci_artifact_id 
            FROM cd_workflow_runner 
            INNER JOIN cd_workflow ON cd_workflow.id = cd_workflow_runner.cd_workflow_id   
            AND (
                cd_workflow.pipeline_id in (%d,%d) AND cd_workflow.ci_artifact_id = %d --optimization 
            )  
            WHERE (
                ( 	-- current CD check
                	cd_workflow.pipeline_id = %d AND cd_workflow_runner.workflow_type = '%s'
                )   
                OR  
                (	-- parent CD check
                    cd_workflow.pipeline_id = %d AND cd_workflow_runner.workflow_type = '%s' AND cd_workflow_runner.status IN ('Healthy','Succeeded') 
                )
            )
        )
    ) 
    OR ( -- checking if artifact is created at parent stage by plugin (eg: skopeo )
        ci_artifact.id=%d  
        AND ci_artifact.component_id = %d   
        AND ci_artifact.data_source = '%s' 
    ) 
    OR id IN (
        -- check if artifact is promoted directly to this pipeline  
        SELECT artifact_id 
        FROM artifact_promotion_approval_request 
        WHERE status = %d 
        AND destination_pipeline_id = %d and artifact_id=%d
    ))`,
		artifactId, pipelineId, parentPipelineId, artifactId, pipelineId, deployStage, parentPipelineId, parentStage, artifactId, parentPipelineId, pluginStage, constants.PROMOTED, pipelineId, artifactId)

	_, err := impl.dbConnection.Query(&Available, query)
	if err != nil {
		return false, err
	}
	return Available, nil
}
