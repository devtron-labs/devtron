package bean

import "github.com/devtron-labs/devtron/pkg/sql"

type LockConfigResponse struct {
	Id      int    `sql:"id,pk"`
	Allowed bool   `sql:"allowed, notnull"`
	Config  string `sql:"config"`
}

type LockConfiguration struct {
	tableName struct{} `sql:"lock_configuration"`
	Id        int      `sql:"id,pk"`
	Allowed   bool     `sql:"allowed, notnull"`
	Config    string   `sql:"config"`
	Active    bool     `sql:"active"`
	sql.AuditLog
}

func (impl *LockConfiguration) ConvertDBDtoToResponse() *LockConfigResponse {
	return &LockConfigResponse{
		Id:      impl.Id,
		Allowed: impl.Allowed,
		Config:  impl.Config,
	}
}

func (impl *LockConfigResponse) ConvertResponseToDBDto() *LockConfiguration {
	return &LockConfiguration{
		Config:  impl.Config,
		Allowed: impl.Allowed,
		Active:  true,
	}
}
