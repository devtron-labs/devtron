package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type DeploymentTemplateHistoryRepository interface {
	CreateHistory(chart *DeploymentTemplateHistory) (*DeploymentTemplateHistory, error)
	CreateHistoryWithTxn(chart *DeploymentTemplateHistory, tx *pg.Tx) (*DeploymentTemplateHistory, error)
	GetHistoryForDeployedTemplateById(id, pipelineId int) (*DeploymentTemplateHistory, error)
	GetDeploymentDetailsForDeployedTemplateHistory(pipelineId, offset, limit int) ([]*DeploymentTemplateHistory, error)
	GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId int) (*DeploymentTemplateHistory, error)
	GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*DeploymentTemplateHistory, error)
}

type DeploymentTemplateHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewDeploymentTemplateHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *DeploymentTemplateHistoryRepositoryImpl {
	return &DeploymentTemplateHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type DeploymentTemplateHistory struct {
	tableName               struct{}  `sql:"deployment_template_history" pg:",discard_unknown_columns"`
	Id                      int       `sql:"id,pk"`
	PipelineId              int       `sql:"pipeline_id"`
	AppId                   int       `sql:"app_id"`
	ImageDescriptorTemplate string    `sql:"image_descriptor_template"`
	Template                string    `sql:"template"`
	TargetEnvironment       int       `sql:"target_environment"`
	TemplateName            string    `sql:"template_name"`
	TemplateVersion         string    `sql:"template_version"`
	IsAppMetricsEnabled     bool      `sql:"is_app_metrics_enabled,notnull"`
	Deployed                bool      `sql:"deployed"`
	DeployedOn              time.Time `sql:"deployed_on"`
	DeployedBy              int32     `sql:"deployed_by"`
	sql.AuditLog
	//getting below data from cd_workflow_runner and users join
	DeploymentStatus  string `sql:"-"`
	DeployedByEmailId string `sql:"-"`
}

func (impl DeploymentTemplateHistoryRepositoryImpl) CreateHistory(chart *DeploymentTemplateHistory) (*DeploymentTemplateHistory, error) {
	err := impl.dbConnection.Insert(chart)
	if err != nil {
		impl.logger.Errorw("err in creating deployment template history entry", "err", err, "history", chart)
		return chart, err
	}
	return chart, nil
}

func (impl DeploymentTemplateHistoryRepositoryImpl) CreateHistoryWithTxn(chart *DeploymentTemplateHistory, tx *pg.Tx) (*DeploymentTemplateHistory, error) {
	err := tx.Insert(chart)
	if err != nil {
		impl.logger.Errorw("err in creating deployment template history entry", "err", err, "history", chart)
		return chart, err
	}
	return chart, nil
}

func (impl DeploymentTemplateHistoryRepositoryImpl) GetHistoryForDeployedTemplateById(id, pipelineId int) (*DeploymentTemplateHistory, error) {
	var history DeploymentTemplateHistory
	err := impl.dbConnection.Model(&history).Where("id = ?", id).
		Where("pipeline_id = ?", pipelineId).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error in getting deployment template history", "err", err)
		return &history, err
	}
	return &history, nil
}

func (impl DeploymentTemplateHistoryRepositoryImpl) GetDeploymentDetailsForDeployedTemplateHistory(pipelineId, offset, limit int) ([]*DeploymentTemplateHistory, error) {
	var histories []*DeploymentTemplateHistory
	err := impl.dbConnection.Model(&histories).Where("pipeline_id = ?", pipelineId).
		Where("deployed = ?", true).
		Offset(offset).Limit(limit).Select()
	if err != nil {
		impl.logger.Errorw("error in getting deployment template history", "err", err)
		return histories, err
	}
	return histories, nil
}

func (impl DeploymentTemplateHistoryRepositoryImpl) GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId int) (*DeploymentTemplateHistory, error) {
	var history DeploymentTemplateHistory
	err := impl.dbConnection.Model(&history).Join("INNER JOIN cd_workflow_runner cwr ON cwr.started_on = deployment_template_history.deployed_on").
		Where("deployment_template_history.pipeline_id = ?", pipelineId).
		Where("deployment_template_history.deployed = ?", true).
		Where("cwr.id = ?", wfrId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting deployment template history by pipelineId & wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return &history, err
	}
	return &history, nil
}

func (impl DeploymentTemplateHistoryRepositoryImpl) GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*DeploymentTemplateHistory, error) {
	var histories []*DeploymentTemplateHistory
	query := "SELECT dth.id, dth.deployed_on, dth.deployed_by, cwr.status as deployment_status, users.email_id as deployed_by_email_id" +
		" FROM deployment_template_history dth" +
		" INNER JOIN cd_workflow_runner cwr ON cwr.started_on = dth.deployed_on" +
		" INNER JOIN users ON users.id = dth.deployed_by" +
		" WHERE dth.pipeline_id = ? AND dth.deployed = true AND dth.id <= ?" +
		" ORDER BY dth.id DESC;"
	_, err := impl.dbConnection.Query(&histories, query, pipelineId, baseConfigId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment template history list by pipelineId", "err", err, "pipelineId", pipelineId)
		return histories, err
	}
	return histories, nil
}
