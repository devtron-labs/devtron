package webhookHelm

type HelmAppCreateUpdateRequest struct {
	ClusterName        string     `json:"clusterName,notnull" validate:"required"`
	Namespace          string     `json:"namespace,omitempty"`
	ReleaseName        string     `json:"releaseName,notnull" validate:"required"`
	ValuesOverrideYaml string     `json:"valuesOverrideYaml,omitempty"`
	Chart              *ChartSpec `json:"chart,notnull" validate:"required"`
}

type ChartSpec struct {
	Repo         *ChartRepoSpec `json:"repo,notnull" validate:"required"`
	ChartName    string         `json:"chartName,notnull" validate:"required"`
	ChartVersion string         `json:"chartVersion,omitempty"`
}

type ChartRepoSpec struct {
	Name       string                   `json:"name,notnull" validate:"required"`
	Identifier *ChartRepoIdentifierSpec `json:"identifier,omitempty"`
}

type ChartRepoIdentifierSpec struct {
	Url      string `json:"url,notnull" validate:"required"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
