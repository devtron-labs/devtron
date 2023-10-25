package repository

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
)

type VariableSnapshotHistory struct {
	tableName struct{} `sql:"variable_snapshot_history" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	VariableSnapshotHistoryBean
	sql.AuditLog
}

type HistoryReference struct {
	HistoryReferenceId   int                  `sql:"history_reference_id"`
	HistoryReferenceType HistoryReferenceType `sql:"history_reference_type"`
}

type VariableSnapshotHistoryBean struct {
	VariableSnapshot json.RawMessage `sql:"variable_snapshot"`
	HistoryReference
}

type VariableSnapshotHistoryBeanRaw struct {
	VariableSnapshot map[string]string
	HistoryReference
}

func GetSnapshotBean(referenceId int, referenceType HistoryReferenceType, snapshot map[string]string) *VariableSnapshotHistoryBean {
	//todo Aditya have to handle error
	if snapshot != nil && len(snapshot) > 0 {
		variableMapBytes, _ := json.Marshal(snapshot)
		return &VariableSnapshotHistoryBean{
			VariableSnapshot: variableMapBytes,
			HistoryReference: HistoryReference{
				HistoryReferenceId:   referenceId,
				HistoryReferenceType: referenceType,
			},
		}
	}
	return nil
}

type HistoryReferenceType int

const (
	HistoryReferenceTypeDeploymentTemplate HistoryReferenceType = 1
	HistoryReferenceTypeCIWORKFLOW         HistoryReferenceType = 2
	HistoryReferenceTypeCDWORKFLOWRUNNER   HistoryReferenceType = 3
	HistoryReferenceTypeConfigMap          HistoryReferenceType = 4
	HistoryReferenceTypeSecret             HistoryReferenceType = 5
)
