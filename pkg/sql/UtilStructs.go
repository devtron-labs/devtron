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
