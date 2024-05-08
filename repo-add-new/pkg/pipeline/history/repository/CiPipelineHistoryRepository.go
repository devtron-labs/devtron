package repository

import (
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiPipelineTemplateOverrideHistoryDTO struct {
	DockerRegistryId      string `json:"docker_registry_id"`
	DockerRepository      string `json:"docker_repository"`
	DockerfilePath        string `json:"dockerfile_path"`
	Active                bool   `json:"active,notnull"`
	CiBuildConfigId       int    `json:"ci_build_config_id"`
	BuildMetaDataType     string `json:"build_meta_data_type"`
	BuildMetadata         string `json:"build_metadata"`
	IsCiTemplateOverriden bool   `json:"is_ci_template_overriden"`
	sql.AuditLog
}

type CiPipelineMaterialHistoryDTO struct {
	tableName     struct{} `json:"ci_pipeline_material"`
	Id            int      `json:"id"`
	GitMaterialId int      `json:"git_material_id"` //id stored in db GitMaterial( foreign key)
	CiPipelineId  int      `json:"ci_pipeline_id"`
	Path          string   `json:"path"` // defaults to root of git repo
	//depricated was used in gocd remove this
	CheckoutPath string          `json:"checkout_path"` //path where code will be checked out for single source `./` default for multiSource configured by user
	Type         bean.SourceType `json:"type"`
	Value        string          `json:"value"`
	ScmId        string          `json:"scm_id"`      //id of gocd object
	ScmName      string          `json:"scm_name"`    //gocd scm name
	ScmVersion   string          `json:"scm_version"` //gocd scm version
	Active       bool            `json:"active,notnull"`
	Regex        string          `json:"regex"`
	GitTag       string          `json:"-"`
	sql.AuditLog
}

//type CiPipelineMaterialHistoryDTO struct {
//	PipelineMaterialId string `sql:"ci_material_id"`
//	GitMaterialId      int    `sql:"git_material_id"` //id stored in db GitMaterial( foreign key)
//	Path               string `sql:"path"`            // defaults to root of git repo
//	//depricated was used in gocd remove this
//	CheckoutPath string          `sql:"checkout_path"` //path where code will be checked out for single source `./` default for multiSource configured by user
//	Type         bean.SourceType `sql:"type"`
//	Value        string          `sql:"value"`
//	ScmId        string          `sql:"scm_id"`      //id of gocd object
//	ScmName      string          `sql:"scm_name"`    //gocd scm name
//	ScmVersion   string          `sql:"scm_version"` //gocd scm version
//	Regex        string          `json:"regex"`
//	GitTag       string          `sql:"-"`
//	Active       bool   		 `sql:"active,notnull"`
//	sql.AuditLog
//}

type CiPipelineHistory struct {
	tableName                 struct{} `sql:"ci_pipeline_history" pg:",discard_unknown_columns"`
	Id                        int      `sql:"id,pk"`
	CiPipelineId              int      `sql:"ci_pipeline_id"`
	CiTemplateOverrideHistory string   `sql:"ci_template_override_history"`
	CiPipelineMaterialHistory string   `sql:"ci_pipeline_material_history"`
	Trigger                   string   `sql:"trigger"`
	ScanEnabled               bool     `sql:"scan_enabled,notnull"`
	Manual                    bool     `sql:"manual,notnull"`
}

type CiEnvMappingHistory struct {
	tableName     struct{} `sql:"ci_env_mapping_history" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	CiPipelineId  int      `sql:"ci_pipeline_id"`
	EnvironmentId int      `sql:"environment_id"`
	sql.AuditLog
}

const (
	TRIGGER_ADD    = "add"
	TRIGGER_UPDATE = "update"
	TRIGGER_DELETE = "delete"
)

type CiPipelineHistoryRepository interface {
	Save(ciPipelineHistory *CiPipelineHistory) error
	SaveCiEnvMappingHistory(CiEnvMappingHistory *CiEnvMappingHistory) error
}

type CiPipelineHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiPipelineHistoryRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiPipelineHistoryRepositoryImpl {

	return &CiPipelineHistoryRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *CiPipelineHistoryRepositoryImpl) Save(CiPipelineHistory *CiPipelineHistory) error {

	err := impl.dbConnection.Insert(CiPipelineHistory)

	if err != nil {
		impl.logger.Errorw("error in saving history for ci pipeline")
		return err
	}

	return nil
}
func (impl *CiPipelineHistoryRepositoryImpl) SaveCiEnvMappingHistory(CiEnvMappingHistory *CiEnvMappingHistory) error {
	err := impl.dbConnection.Insert(CiEnvMappingHistory)

	if err != nil {
		impl.logger.Errorw("error in saving history for Ci-Env Mapping", "err", err)
		return err
	}

	return nil
}
