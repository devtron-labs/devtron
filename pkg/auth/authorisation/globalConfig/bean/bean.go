package bean

type GlobalAuthorisationConfigType string

const (
	DevtronSystemManaged       GlobalAuthorisationConfigType = "devtron-system-managed"
	DevtronSelfRegisteredGroup GlobalAuthorisationConfigType = "devtron-self-registered-group"
	GroupClaims                GlobalAuthorisationConfigType = "group-claims" // GroupClaims are currently used for Active directory and LDAP
)

type GlobalAuthorisationConfig struct {
	ConfigTypes []string `json:"configTypes" validate:"required"`
	UserId      int32    `json:"userId"` //for Internal Use
}

type GlobalAuthorisationConfigResponse struct {
	Id         int    `json:"id"`
	ConfigType string `json:"configType"`
	Active     bool   `json:"active"`
}
