package bean

type UserInfo struct {
	Id          int32        `json:"id" validate:"number"`
	EmailId     string       `json:"email_id" validate:"required"`
	RoleFilters []RoleFilter `json:"roleFilters"`
	Groups      []string     `json:"groups"`
	SuperAdmin  bool         `json:"superAdmin,notnull"`
	Status      string       `json:"status,omitempty"`
	AccessToken string       `json:"access_token,omitempty"`
	Roles       []string     `json:"roles,omitempty"`
	AccessType  string       `json:"accessType,omitempty"` //dawf, hawf, iam, scluster
	UserId      int32        `json:"-"`                    // created or modified user id
}

type RoleFilter struct {
	Entity      string `json:"entity"` // "awf", "chart-group", "hawf", "cluster", [""=awf] existing
	Team        string `json:"team"`   // project name
	Cluster     string `json:"cluster"`
	EntityName  string `json:"entityName"` // app or chart name
	Environment string `json:"environment"`
	Namespace   string `json:"namespace"`
	Action      string `json:"action"` // view, trigger, edit, admin
	Deny        bool   `json:"deny"`
}
