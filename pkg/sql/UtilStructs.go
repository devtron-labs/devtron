/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package sql

import (
	"time"
)

type AuditLog struct {
	CreatedOn time.Time `sql:"created_on,type:timestamptz"`
	CreatedBy int32     `sql:"created_by,type:integer"`
	UpdatedOn time.Time `sql:"updated_on,type:timestamptz"`
	UpdatedBy int32     `sql:"updated_by,type:integer"`
}

func NewDefaultAuditLog(userId int32) AuditLog {
	return AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userId,
		UpdatedOn: time.Now(),
		UpdatedBy: userId,
	}
}

// CreateAuditLog can be used by any repository to create AuditLog for insert operation
func (model *AuditLog) CreateAuditLog(userId int32) {
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	model.CreatedBy = userId
	model.UpdatedBy = userId
}

// UpdateAuditLog can be used by any repository to update AuditLog for update operation
func (model *AuditLog) UpdateAuditLog(userId int32) {
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
}
