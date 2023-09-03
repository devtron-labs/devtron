package sql

import "time"

type AuditLog struct {
	CreatedOn time.Time `sql:"created_on"`
	CreatedBy int32     `sql:"created_by"`
	UpdatedOn time.Time `sql:"updated_on"`
	UpdatedBy int32     `sql:"updated_by"`
}
