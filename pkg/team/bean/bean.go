package bean

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

type TeamRequest struct {
	Id        int       `json:"id,omitempty" validate:"number"`
	Name      string    `json:"name,omitempty" validate:"required"`
	Active    bool      `json:"active"`
	UserId    int32     `json:"-"`
	CreatedOn time.Time `json:"-"`
}

type Team struct {
	tableName struct{} `sql:"team"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
	Active    bool     `sql:"active,notnull"`
	sql.AuditLog
}

type TeamBean struct {
	Id   int    `json:"id"`
	Name string `json:"name,notnull"`
}
