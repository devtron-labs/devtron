package bean

import "github.com/devtron-labs/devtron/pkg/sql"

const (
	DeploymentChartType             = "Deployment"
	RolloutChartType                = "Rollout Deployment"
	ReferenceChart                  = "reference-chart"
	RefChartDirPath                 = "scripts/devtron-reference-helm-charts"
	ChartAlreadyExistsInternalError = "Chart exists already, try uploading another chart"
	ChartNameReservedInternalError  = "Change the name of the chart and try uploading again"
)

type ChartDataInfo struct {
	ChartLocation   string `json:"chartLocation"`
	ChartName       string `json:"chartName"`
	ChartVersion    string `json:"chartVersion"`
	TemporaryFolder string `json:"temporaryFolder"`
	Description     string `json:"description"`
	Message         string `json:"message"`
}

type ChartYamlStruct struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

var ReservedChartRefNamesList *[]ReservedChartList

type ReservedChartList struct {
	LocationPrefix string
	Name           string
}

type ChartRefDto struct {
	Id                     int    `json:"id"`
	Location               string `json:"location"`
	Version                string `json:"version"`
	Default                bool   `json:"isDefault"`
	Name                   string `json:"name"`
	ChartData              []byte `json:"chartData"`
	ChartDescription       string `json:"chartDescription"`
	UserUploaded           bool   `json:"userUploaded,notnull"`
	IsAppMetricsSupported  bool   `json:"isAppMetricsSupported"`
	DeploymentStrategyPath string `json:"deploymentStrategyPath"`
	JsonPathForStrategy    string `json:"jsonPathForStrategy"`
}

// TODO: below objects are created/moved while refactoring to remove db object usage, to remove/replace them with the common objects mentioned above

type CustomChartRefDto struct {
	Id                     int    `sql:"id,pk"`
	Location               string `sql:"location"`
	Version                string `sql:"version"`
	Active                 bool   `sql:"active,notnull"`
	Default                bool   `sql:"is_default,notnull"`
	Name                   string `sql:"name"`
	ChartData              []byte `sql:"chart_data"`
	ChartDescription       string `sql:"chart_description"`
	UserUploaded           bool   `sql:"user_uploaded,notnull"`
	IsAppMetricsSupported  bool   `sql:"is_app_metrics_supported,notnull"`
	DeploymentStrategyPath string `sql:"deployment_strategy_path"`
	JsonPathForStrategy    string `sql:"json_path_for_strategy"`
	sql.AuditLog
}

type ChartRefAutocompleteDto struct {
	Id                    int    `json:"id"`
	Version               string `json:"version"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	UserUploaded          bool   `json:"userUploaded"`
	IsAppMetricsSupported bool   `json:"isAppMetricsSupported"`
}

type ChartRefMetaData struct {
	ChartDescription string `json:"chartDescription"`
}

type ChartRefAutocompleteResponse struct {
	ChartRefs            []ChartRefAutocompleteDto   `json:"chartRefs"`
	LatestChartRef       int                         `json:"latestChartRef"`
	LatestAppChartRef    int                         `json:"latestAppChartRef"`
	LatestEnvChartRef    int                         `json:"latestEnvChartRef,omitempty"`
	ChartsMetadata       map[string]ChartRefMetaData `json:"chartMetadata"` // chartName vs Metadata
	CompatibleChartTypes []string                    `json:"compatibleChartTypes,omitempty"`
}

type ChartDto struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	ChartDescription string `json:"chartDescription"`
	Version          string `json:"version"`
	IsUserUploaded   bool   `json:"isUserUploaded"`
}
