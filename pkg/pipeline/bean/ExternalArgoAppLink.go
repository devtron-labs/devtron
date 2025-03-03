package bean

type MigrateReleaseValidationRequest struct {
	AppId                      int                        `json:"appId"`
	DeploymentAppName          string                     `json:"deploymentAppName"`
	DeploymentAppType          string                     `json:"deploymentAppType"`
	ApplicationMetadataRequest ApplicationMetadataRequest `json:"applicationMetadata"`
	HelmReleaseMetadataRequest HelmReleaseMetadataRequest `json:"helmReleaseMetadata"`
}

type ApplicationMetadataRequest struct {
	ApplicationObjectClusterId int    `json:"applicationObjectClusterId"`
	ApplicationObjectNamespace string `json:"applicationObjectNamespace"`
}

type HelmReleaseMetadataRequest struct {
	ReleaseClusterId int    `json:"releaseClusterId"`
	ReleaseNamespace string `json:"releaseNamespace"`
}

type ExternalAppLinkValidationResponse struct {
	IsLinkable          bool                `json:"isLinkable"`
	ErrorDetail         ErrorDetail         `json:"errorDetail"`
	ApplicationMetadata ApplicationMetadata `json:"applicationMetadata"`
	HelmReleaseMetadata HelmReleaseMetadata `json:"helmReleaseMetadata"`
}

type ApplicationMetadata struct {
	Source      Source      `json:"source"`
	Destination Destination `json:"destination"`
	Status      string      `json:"status"`
}

func NewEmptyApplicationMetadata() ApplicationMetadata {
	return ApplicationMetadata{}
}

type Source struct {
	RepoURL       string        `json:"repoURL"`
	ChartPath     string        `json:"chartPath"`
	ChartMetadata ChartMetadata `json:"chartMetadata"`
}

type ChartMetadata struct {
	RequiredChartVersion string `json:"requiredChartVersion"`
	SavedChartName       string `json:"savedChartName"`
	ValuesFilename       string `json:"valuesFilename"`
	RequiredChartName    string `json:"requiredChartName"`
}

type Destination struct {
	ClusterName      string `json:"clusterName"`
	ClusterServerUrl string `json:"clusterServerUrl"`
	Namespace        string `json:"namespace"`
	EnvironmentName  string `json:"environmentName"`
	EnvironmentId    int    `json:"environmentId"`
}

type HelmReleaseMetadata struct {
	Name        string           `json:"name"`
	Info        HelmReleaseInfo  `json:"info"`
	Chart       HelmReleaseChart `json:"chart"`
	Destination Destination      `json:"destination"`
}

type HelmReleaseChart struct {
	HelmReleaseChartMetadata HelmReleaseChartMetadata `json:"metadata"`
}

type HelmReleaseChartMetadata struct {
	RequiredChartName string `json:"requiredChartName"`
	SavedChartName    string `json:"savedChartName"`
	Home              string `json:"home"`
	Version           string `json:"version"`
	Icon              string `json:"icon"`
	ApiVersion        string `json:"apiVersion"`
	Deprecated        bool   `json:"deprecated"`
}

type HelmReleaseInfo struct {
	Status string `json:"status"`
}

func (a *ExternalAppLinkValidationResponse) SetErrorDetail(ValidationFailedReason LinkFailedReason, ValidationFailedMessage string) ExternalAppLinkValidationResponse {
	a.ErrorDetail = ErrorDetail{
		ValidationFailedReason:  ValidationFailedReason,
		ValidationFailedMessage: ValidationFailedMessage,
	}
	return *a
}

func (a *ExternalAppLinkValidationResponse) SetUnknownErrorDetail(err error) ExternalAppLinkValidationResponse {
	a.ErrorDetail = ErrorDetail{
		ValidationFailedReason:  InternalServerError,
		ValidationFailedMessage: err.Error(),
	}
	return *a
}

type LinkFailedReason string

type ErrorDetail struct {
	ValidationFailedReason  LinkFailedReason `json:"validationFailedReason"`
	ValidationFailedMessage string           `json:"validationFailedMessage"`
}

const (
	ClusterNotFound            LinkFailedReason = "ClusterNotFound"
	EnvironmentNotFound        LinkFailedReason = "EnvironmentNotFound"
	ApplicationAlreadyPresent  LinkFailedReason = "ApplicationAlreadyPresent"
	UnsupportedApplicationSpec LinkFailedReason = "UnsupportedApplicationSpec"
	ChartTypeMismatch          LinkFailedReason = "ChartTypeMismatch"
	ChartVersionNotFound       LinkFailedReason = "ChartVersionNotFound"
	GitOpsNotFound             LinkFailedReason = "GitOpsNotFound"
	InternalServerError        LinkFailedReason = "InternalServerError"
	EnvironmentAlreadyPresent  LinkFailedReason = "EnvironmentAlreadyPresent"
	EnforcedPolicyViolation    LinkFailedReason = "EnforcedPolicyViolation"
)

const (
	ChartTypeMismatchErrorMsg    string = "Argo CD application uses '%s' chart where as this application uses '%s' chart. You can upload your own charts in Global Configuration > Deployment Charts."
	ChartVersionNotFoundErrorMsg string = "Chart version %s not found for %s chart"
	PipelineAlreadyPresentMsg    string = "A pipeline already exist for this environment."
	HelmAppAlreadyPresentMsg     string = "A helm app already exist for this environment."
)
