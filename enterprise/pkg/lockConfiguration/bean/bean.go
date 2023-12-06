package bean

import "github.com/devtron-labs/devtron/pkg/sql"

type LockConfigRequest struct {
	Allowed bool     `sql:"allowed, notnull"`
	Config  []string `sql:"config, notnull"`
}

type LockConfigResponse struct {
	Id      int      `sql:"id,pk"`
	Allowed bool     `sql:"allowed, notnull"`
	Config  []string `sql:"config, notnull"`
}

type LockConfiguration struct {
	tableName struct{} `sql:"lock_configuration"`
	Id        int      `sql:"id,pk"`
	Allowed   bool     `sql:"allowed, notnull"`
	Config    []string `sql:"config, notnull"`
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

func (impl *LockConfigRequest) ConvertRequestToDBDto() *LockConfiguration {
	return &LockConfiguration{
		Config:  impl.Config,
		Allowed: impl.Allowed,
		Active:  true,
	}
}
