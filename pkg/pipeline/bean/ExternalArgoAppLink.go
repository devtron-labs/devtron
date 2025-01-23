package bean

type ArgoCDAppLinkValidationRequest struct {
	AppId         int    `json:"appId"`
	ClusterId     int    `json:"clusterId"`
	Namespace     string `json:"namespace"`
	ArgoCDAppName string `json:"argoCDAppName"`
}

type ArgoCdAppLinkValidationResponse struct {
	IsLinkable          bool                `json:"isLinkable"`
	ErrorDetail         ErrorDetail         `json:"errorDetail"`
	ApplicationMetadata ApplicationMetadata `json:"applicationMetadata"`
}

type ApplicationMetadata struct {
	TargetCluster     string `json:"targetCluster"`
	TargetEnvironment string `json:"targetEnvironment"`
	TargetNamespace   string `json:"targetNamespace"`
}

func (a *ArgoCdAppLinkValidationResponse) SetErrorDetail(ValidationFailedReason LinkFailedReason, ValidationFailedMessage string) ArgoCdAppLinkValidationResponse {
	a.ErrorDetail = ErrorDetail{
		ValidationFailedReason:  ValidationFailedReason,
		ValidationFailedMessage: ValidationFailedMessage,
	}
	return *a
}

func (a *ArgoCdAppLinkValidationResponse) SetUnknownErrorDetail(err error) ArgoCdAppLinkValidationResponse {
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
)

const (
	ChartTypeMismatchErrorMsg    string = "Argo CD application uses '%s' chart where as this application uses '%s' chart. You can upload your own charts in Global Configuration > Deployment Charts."
	ChartVersionNotFoundErrorMsg string = "Chart version %s not found for %s chart"
	PipelineAlreadyPresentMsg    string = "A pipeline already exist for this environment."
	HelmAppAlreadyPresentMsg     string = "A helm app already exist for this environment."
)
