/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package chartConfig

import (
	"github.com/devtron-labs/devtron/api/bean"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	deploymentStatus "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"time"
)

type PipelineOverride struct {
	tableName              struct{}              `sql:"pipeline_config_override" pg:",discard_unknown_columns"`
	Id                     int                   `sql:"id,pk"`
	RequestIdentifier      string                `sql:"request_identifier,unique,notnull"`
	EnvConfigOverrideId    int                   `sql:"env_config_override_id, notnull"`
	PipelineOverrideValues string                `sql:"pipeline_override_yaml,notnull"`
	PipelineMergedValues   string                `sql:"merged_values_yaml, notnull"` //merge of appOverride, envOverride, pipelineOverride
	Status                 models.ChartStatus    `sql:"status,notnull"`              // new , deployment-in-progress, success, rollbacked
	GitHash                string                `sql:"git_hash"`
	CommitTime             time.Time             `sql:"commit_time,type:timestamptz"`
	PipelineId             int                   `sql:"pipeline_id"`
	CiArtifactId           int                   `sql:"ci_artifact_id"`
	PipelineReleaseCounter int                   `sql:"pipeline_release_counter"` //built index
	CdWorkflowId           int                   `sql:"cd_workflow_id"`           //built index
	DeploymentType         models.DeploymentType `sql:"deployment_type"`          // deployment type
	sql.AuditLog
	EnvConfigOverride *EnvConfigOverride
	CiArtifact        *repository.CiArtifact
	Pipeline          *pipelineConfig.Pipeline
}

type PipelineConfigOverrideMetadata struct {
	AppId              int
	UserApprovalConfig string
	MergedValuesYaml   string
}

// PipelineOverrideRepository should be used by only manifest.ManifestCreationService and devtronApps.TriggerService
// For read services use configHistory.PipelineConfigOverrideReadService instead
type PipelineOverrideRepository interface {
	Save(*PipelineOverride) error
	UpdateStatusByRequestIdentifier(requestId string, newStatus models.ChartStatus) (int, error)
	GetLatestConfigByRequestIdentifier(requestIdentifier string) (pipelineOverride *PipelineOverride, err error)
	GetLatestConfigByEnvironmentConfigOverrideId(envConfigOverrideId int) (pipelineOverride *PipelineOverride, err error)
	Update(pipelineOverride *PipelineOverride) error
	GetCurrentPipelineReleaseCounter(pipelineId int) (releaseCounter int, err error)
	GetByPipelineIdAndReleaseNo(pipelineId, releaseNo int) (pipelineOverrides []*PipelineOverride, err error)
	GetAllRelease(appId, environmentId int) (pipelineOverrides []*PipelineOverride, err error)
	FindByPipelineTriggerGitHash(gitHash string) (pipelineOverride *PipelineOverride, err error)
	GetLatestRelease(appId, environmentId int) (pipelineOverrides *PipelineOverride, err error)
	FetchLatestNDeployedArtifacts(appId, environmentId, limit int) (pipelineOverrides []*PipelineOverride, err error)
	GetLatestReleaseForAppIds(appIds []int, envId int) (pipelineOverrides []*PipelineConfigOverrideMetadata, err error)
	FindById(id int) (*PipelineOverride, error)
	GetByDeployedImage(appId, environmentId int, images []string) (pipelineOverride *PipelineOverride, err error)
	GetLatestReleaseByPipelineIds(pipelineIds []int) (pipelineOverrides []*PipelineOverride, err error)
	GetLatestReleaseDeploymentType(pipelineIds []int) ([]*PipelineOverride, error)
	FetchHelmTypePipelineOverridesForStatusUpdate() (pipelines []*PipelineOverride, err error)
	FindLatestByAppIdAndEnvId(appId, environmentId int, deploymentAppType string) (pipelineOverrides *PipelineOverride, err error)
	FindLatestByCdWorkflowId(cdWorkflowId int) (pipelineOverride *PipelineOverride, err error)
}

type PipelineOverrideRepositoryImpl struct {
	dbConnection *pg.DB
}

func (impl PipelineOverrideRepositoryImpl) Save(pipelineOverride *PipelineOverride) error {
	return impl.dbConnection.Insert(pipelineOverride)
}

func (impl PipelineOverrideRepositoryImpl) Update(pipelineOverride *PipelineOverride) error {
	_, err := impl.dbConnection.Model(pipelineOverride).WherePK().UpdateNotNull()
	return err
}
func (impl PipelineOverrideRepositoryImpl) UpdateStatusByRequestIdentifier(requestId string, newStatus models.ChartStatus) (int, error) {
	pipelineOverride := &PipelineOverride{RequestIdentifier: requestId, Status: newStatus}
	res, err := impl.dbConnection.Model(pipelineOverride).
		Set("status = ?status").
		Where("request_identifier = ?request_identifier").
		Update()
	return res.RowsAffected(), err
}

func (impl PipelineOverrideRepositoryImpl) GetLatestConfigByRequestIdentifier(requestIdentifier string) (pipelineOverride *PipelineOverride, err error) {
	pipelineOverride = &PipelineOverride{RequestIdentifier: requestIdentifier}
	err = impl.dbConnection.Model(pipelineOverride).
		Where("request_identifier = ?request_identifier").
		Order("id DESC").
		First()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return pipelineOverride, err
}

func (impl PipelineOverrideRepositoryImpl) GetLatestConfigByEnvironmentConfigOverrideId(envConfigOverrideId int) (pipelineOverride *PipelineOverride, err error) {
	pipelineOverride = &PipelineOverride{EnvConfigOverrideId: envConfigOverrideId}
	err = impl.dbConnection.Model(pipelineOverride).
		Where("env_config_override_id = ?env_config_override_id").
		Order("id DESC").
		First()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return pipelineOverride, err
}

func (impl PipelineOverrideRepositoryImpl) GetCurrentPipelineReleaseCounter(pipelineId int) (releaseCounter int, err error) {
	var counter int
	err = impl.dbConnection.Model((*PipelineOverride)(nil)).
		Column("pipeline_release_counter").
		Where("pipeline_id =? ", pipelineId).
		Order("id DESC").
		Limit(1).
		Select(&counter)
	if err != nil && util.IsErrNoRows(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else {
		return counter, nil
	}
}

func (impl PipelineOverrideRepositoryImpl) GetByPipelineIdAndReleaseNo(pipelineId, releaseNo int) (pipelineOverrides []*PipelineOverride, err error) {
	var overrides []*PipelineOverride
	err = impl.dbConnection.Model(&overrides).
		Where("pipeline_id =? ", pipelineId).
		Where("pipeline_release_counter =? ", releaseNo).
		Order("id ASC").
		Select()
	return overrides, err
}

func NewPipelineOverrideRepository(dbConnection *pg.DB) *PipelineOverrideRepositoryImpl {
	return &PipelineOverrideRepositoryImpl{dbConnection: dbConnection}
}

func (impl PipelineOverrideRepositoryImpl) GetAllRelease(appId, environmentId int) (pipelineOverrides []*PipelineOverride, err error) {
	var overrides []*PipelineOverride
	err = impl.dbConnection.Model(&overrides).
		Column("pipeline_override.*", "Pipeline", "CiArtifact").
		Where("pipeline.app_id =? ", appId).
		Where("pipeline.environment_id =?", environmentId).
		Order("id ASC").
		Select()
	return overrides, err
}

func (impl PipelineOverrideRepositoryImpl) GetByDeployedImage(appId, environmentId int, images []string) (pipelineOverride *PipelineOverride, err error) {
	override := &PipelineOverride{}
	err = impl.dbConnection.Model(override).
		Column("pipeline_override.*", "Pipeline", "CiArtifact").
		Where("pipeline.app_id =? ", appId).
		Where("pipeline.environment_id =?", environmentId).
		Where("ci_artifact.image in (?)", pg.In(images)).
		Order("id Desc").
		Limit(1).
		Select()
	return override, err
}

func (impl PipelineOverrideRepositoryImpl) GetLatestRelease(appId, environmentId int) (pipelineOverrides *PipelineOverride, err error) {
	overrides := &PipelineOverride{}
	err = impl.dbConnection.Model(overrides).
		Column("pipeline_override.*", "Pipeline", "CiArtifact").
		Where("pipeline.app_id =? ", appId).
		Where("pipeline.environment_id =?", environmentId).
		Order("id DESC").
		Limit(1).
		Select()
	return overrides, err
}

func (impl PipelineOverrideRepositoryImpl) FetchLatestNDeployedArtifacts(appId, environmentId, limit int) (pipelineOverrides []*PipelineOverride, err error) {
	var overrides []*PipelineOverride
	err = impl.dbConnection.Model(&overrides).
		Column("pipeline_override.ci_artifact_id").
		Where("pipeline.app_id =? ", appId).
		Where("pipeline.environment_id =?", environmentId).
		Order("id DESC").
		Limit(limit).
		Select()
	return overrides, err
}

func (impl PipelineOverrideRepositoryImpl) GetLatestReleaseForAppIds(appIds []int, envId int) (pipelineOverrideMetadata []*PipelineConfigOverrideMetadata, err error) {
	var OverrideMetadata []*PipelineConfigOverrideMetadata
	if len(appIds) == 0 {
		return nil, nil
	}
	query := "WITH temp_pipeline AS (" +
		"     SELECT p.id,p.app_id,p.user_approval_config " +
		"     FROM pipeline p " +
		"     WHERE p.environment_id = ? " +
		"     AND p.app_id IN (?) " +
		"     AND p.deleted = false " +
		"     AND p.deployment_app_created = true " +
		"     AND p.deployment_app_delete_request = false) " +
		" SELECT pco.merged_values_yaml,p.app_id,p.user_approval_config " +
		" FROM pipeline_config_override pco " +
		" INNER JOIN temp_pipeline p ON p.id = pco.pipeline_id " +
		" WHERE pco.id IN " +
		"      ( SELECT max(pco.id) as pco_id " +
		"         FROM pipeline_config_override pco " +
		"         WHERE pco.pipeline_id IN (SELECT id FROM temp_pipeline) " +
		"         GROUP BY pco.pipeline_id " +
		"       );"
	_, err = impl.dbConnection.
		Query(&OverrideMetadata, query, envId, pg.In(appIds))
	return OverrideMetadata, err
}
func (impl PipelineOverrideRepositoryImpl) GetLatestReleaseByPipelineIds(pipelineIds []int) (pipelineOverrides []*PipelineOverride, err error) {
	var overrides []*PipelineOverride
	err = impl.dbConnection.Model(&overrides).
		Column("pipeline_override.*").
		Where("pipeline_override.pipeline_id in (?) ", pg.In(pipelineIds)).
		Order("id DESC").
		Select()
	return overrides, err
}

func (impl PipelineOverrideRepositoryImpl) GetLatestReleaseDeploymentType(pipelineIds []int) ([]*PipelineOverride, error) {
	var overrides []*PipelineOverride
	query := "select pco.pipeline_id,pco.deployment_type, max(id) as id from pipeline_config_override pco" +
		" where pco.pipeline_id in (?) " +
		" group by pco.pipeline_id, pco.deployment_type order by id desc"
	_, err := impl.dbConnection.Query(&overrides, query, pg.In(pipelineIds))
	if err != nil {
		return overrides, err
	}
	return overrides, err
}

func (impl PipelineOverrideRepositoryImpl) FindByPipelineTriggerGitHash(gitHash string) (pipelineOverride *PipelineOverride, err error) {
	pipelineOverride = &PipelineOverride{}
	err = impl.dbConnection.Model(pipelineOverride).
		Column("pipeline_override.*", "Pipeline", "CiArtifact").
		Where("pipeline_override.git_hash =?", gitHash).
		Order("id DESC").Limit(1).
		Select()
	return pipelineOverride, err
}

func (impl PipelineOverrideRepositoryImpl) FindById(id int) (*PipelineOverride, error) {
	var pipelineOverride PipelineOverride
	err := impl.dbConnection.Model(&pipelineOverride).
		Column("pipeline_override.*", "Pipeline", "CiArtifact").
		Where("pipeline_override.id =?", id).
		Select()
	return &pipelineOverride, err
}

func (impl PipelineOverrideRepositoryImpl) FetchHelmTypePipelineOverridesForStatusUpdate() (pipelines []*PipelineOverride, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("pipeline_override.*", "Pipeline", "Pipeline.App", "Pipeline.Environment").
		Join("inner join pipeline p on p.id = pipeline_override.pipeline_id").
		Join("inner join cd_workflow cdwf on cdwf.pipeline_id = p.id").
		Join("inner join cd_workflow_runner cdwfr on cdwfr.cd_workflow_id = cdwf.id").
		Where("p.deployment_app_type = ?", util.PIPELINE_DEPLOYMENT_TYPE_HELM).
		Where("cdwfr.status not in (?)", pg.In([]string{argoApplication.Degraded, argoApplication.HIBERNATING, argoApplication.Healthy, deploymentStatus.WorkflowFailed, deploymentStatus.WorkflowAborted})).
		Where("cdwfr.workflow_type = ?", bean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("p.deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineOverrideRepositoryImpl) FindLatestByAppIdAndEnvId(appId, environmentId int, deploymentAppType string) (pipelineOverrides *PipelineOverride, err error) {
	var override PipelineOverride
	err = impl.dbConnection.Model(&override).
		Column("pipeline_override.*", "Pipeline").
		Join("inner join pipeline p on p.id = pipeline_override.pipeline_id").
		Where("pipeline.app_id =? ", appId).
		Where("pipeline.environment_id =?", environmentId).
		Where("p.deployment_app_type = ?", deploymentAppType).
		Where("p.deleted = ?", false).
		Order("id DESC").Limit(1).
		Select()
	return &override, err
}

func (impl PipelineOverrideRepositoryImpl) FindLatestByCdWorkflowId(cdWorkflowId int) (*PipelineOverride, error) {
	var override PipelineOverride
	err := impl.dbConnection.Model(&override).
		Column("pipeline_override.*", "Pipeline").
		Where("cd_workflow_id=?", cdWorkflowId).
		Order("id DESC").Limit(1).
		Select()
	return &override, err
}
