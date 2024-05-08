package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiTemplateHistory struct {
	tableName          struct{} `sql:"ci_template_history" pg:",discard_unknown_columns"`
	Id                 int      `sql:"id,pk"`
	CiTemplateId       int      `sql:"ci_template_id"`
	AppId              int      `sql:"app_id"`             //foreign key of app
	DockerRegistryId   *string  `sql:"docker_registry_id"` //foreign key of registry
	DockerRepository   string   `sql:"docker_repository"`
	DockerfilePath     string   `sql:"dockerfile_path"`
	Args               string   `sql:"args"` //json string format of map[string]string
	TargetPlatform     string   `sql:"target_platform,notnull"`
	BeforeDockerBuild  string   `sql:"before_docker_build"` //json string  format of []*Task
	AfterDockerBuild   string   `sql:"after_docker_build"`  //json string  format of []*Task
	TemplateName       string   `sql:"template_name"`
	Version            string   `sql:"version"` //gocd etage
	Active             bool     `sql:"active,notnull"`
	GitMaterialId      int      `sql:"git_material_id"`
	DockerBuildOptions string   `sql:"docker_build_options"` //json string format of map[string]string
	CiBuildConfigId    int      `sql:"ci_build_config_id"`
	BuildMetaDataType  string   `sql:"build_meta_data_type"`
	BuildMetadata      string   `sql:"build_metadata"`
	Trigger            string   `sql:"trigger"`
	sql.AuditLog
	App            *app.App
	DockerRegistry *repository2.DockerArtifactStore
}

type CiTemplateHistoryRepository interface {
	Save(material *CiTemplateHistory) error
}

type CiTemplateHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiTemplateHistoryRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiTemplateHistoryRepositoryImpl {
	return &CiTemplateHistoryRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl CiTemplateHistoryRepositoryImpl) Save(material *CiTemplateHistory) error {

	err := impl.dbConnection.Insert(material)

	if err != nil {
		impl.logger.Errorw("error in saving history for ci template history")
		return err
	}

	return nil

}
