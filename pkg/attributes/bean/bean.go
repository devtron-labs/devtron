package bean

const (
	HostUrlKey                     string = "url"
	API_SECRET_KEY                 string = "apiTokenSecret"
	ENFORCE_DEPLOYMENT_TYPE_CONFIG string = "enforceDeploymentTypeConfig"
	PRIORITY_DEPLOYMENT_CONDITION  string = "priorityDeploymentCondition"
)

type AttributesDto struct {
	Id     int    `json:"id"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Active bool   `json:"active"`
	UserId int32  `json:"-"`
}
