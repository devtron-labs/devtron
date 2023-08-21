package repository

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
)

type VariableSnapshotHistory struct {
	tableName struct{} `sql:"variable-snapshot_history" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	VariableSnapshotHistoryBean
	sql.AuditLog
}

type HistoryReference struct {
	HistoryReferenceId   int                  `sql:"history_reference_id"`
	HistoryReferenceType HistoryReferenceType `sql:"history_reference_type"`
}

type VariableSnapshotHistoryBean struct {
	VariableSnapshot json.RawMessage
	HistoryReference
}

type HistoryReferenceType string

const (
	HistoryReferenceTypeDeploymentTemplate HistoryReferenceType = "DEPLOYMENT_TEMPLATE"
)
