package audit

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type InfraConfigAuditRepository interface {
	SaveInfraConfigHistorySnapshot(tx *pg.Tx, infraConfigTriggerHistories []*InfraConfigTriggerHistory) error
}

type InfraConfigAuditRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewInfraConfigAuditRepositoryImpl(dbConnection *pg.DB) *InfraConfigAuditRepositoryImpl {
	return &InfraConfigAuditRepositoryImpl{
		dbConnection: dbConnection,
	}
}

type WorkflowType string

const (
	CIWorkflowType WorkflowType = "CI"
)

type InfraConfigTriggerHistory struct {
	tableName    struct{}     `sql:"infra_config_trigger_history" pg:",discard_unknown_columns"`
	Id           int          `sql:"id"`
	Key          v1.ConfigKey `sql:"key"`
	ValueString  string       `sql:"value_string"`
	Platform     string       `sql:"platform"`
	WorkflowId   int          `sql:"workflow_id"`
	WorkflowType WorkflowType `sql:"workflow_type"`
	sql.AuditLog
}

func (i *InfraConfigTriggerHistory) WithPlatform(platform string) *InfraConfigTriggerHistory {
	i.Platform = platform
	return i
}

func (i *InfraConfigTriggerHistory) WithWorkflowId(workflowId int) *InfraConfigTriggerHistory {
	i.WorkflowId = workflowId
	return i
}

func (i *InfraConfigTriggerHistory) WithWorkflowType(workflowType WorkflowType) *InfraConfigTriggerHistory {
	i.WorkflowType = workflowType
	return i
}

func (i *InfraConfigTriggerHistory) WithAuditLog(userId int32) *InfraConfigTriggerHistory {
	i.AuditLog = sql.NewDefaultAuditLog(userId)
	return i
}

func (impl *InfraConfigAuditRepositoryImpl) SaveInfraConfigHistorySnapshot(tx *pg.Tx, infraConfigTriggerHistories []*InfraConfigTriggerHistory) error {
	if len(infraConfigTriggerHistories) == 0 {
		return nil
	}
	_, err := tx.Model(&infraConfigTriggerHistories).Insert()
	if err != nil {
		return err
	}
	return nil
}
