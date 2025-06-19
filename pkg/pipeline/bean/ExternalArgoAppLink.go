package bean

import (
	"errors"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	chart2 "helm.sh/helm/v3/pkg/chart"
)

type MigrateReleaseValidationRequest struct {
	AppId                      int                        `json:"appId"`
	DeploymentAppName          string                     `json:"deploymentAppName"`
	DeploymentAppType          string                     `json:"deploymentAppType"`
	ApplicationMetadataRequest ApplicationMetadataRequest `json:"applicationMetadata"`
	HelmReleaseMetadataRequest HelmReleaseMetadataRequest `json:"helmReleaseMetadata"`
	FluxReleaseMetadataRequest FluxReleaseMetadataRequest `json:"fluxReleaseMetadata"`
}

type ApplicationMetadataRequest struct {
	ApplicationObjectClusterId int    `json:"applicationObjectClusterId"`
	ApplicationObjectNamespace string `json:"applicationObjectNamespace"`
}

type HelmReleaseMetadataRequest struct {
	ReleaseClusterId int    `json:"releaseClusterId"`
	ReleaseNamespace string `json:"releaseNamespace"`
}

type FluxReleaseMetadataRequest struct {
	ReleaseClusterId int    `json:"releaseClusterId"`
	ReleaseNamespace string `json:"releaseNamespace"`
}

func (h MigrateReleaseValidationRequest) GetHelmReleaseClusterId() int {
	return h.HelmReleaseMetadataRequest.ReleaseClusterId
}

func (h MigrateReleaseValidationRequest) GetHelmReleaseNamespace() string {
	return h.HelmReleaseMetadataRequest.ReleaseNamespace
}

func (h MigrateReleaseValidationRequest) GetFluxReleaseClusterId() int {
	return h.FluxReleaseMetadataRequest.ReleaseClusterId
}

func (h MigrateReleaseValidationRequest) GetFluxReleaseNamespace() string {
	return h.FluxReleaseMetadataRequest.ReleaseNamespace
}

type ExternalAppLinkValidationResponse struct {
	IsLinkable          bool                `json:"isLinkable"`
	ErrorDetail         *ErrorDetail        `json:"errorDetail"`
	ApplicationMetadata ApplicationMetadata `json:"applicationMetadata"`
	HelmReleaseMetadata HelmReleaseMetadata `json:"helmReleaseMetadata"`
	FluxReleaseMetadata FluxReleaseMetadata `json:"fluxReleaseMetadata"`
}

func (a *ApplicationMetadata) UpdateApplicationSpecData(argoApplicationSpec *v1alpha1.Application) {
	if argoApplicationSpec.Spec.Source != nil {

		a.Source = Source{
			RepoURL:   argoApplicationSpec.Spec.Source.RepoURL,
			ChartPath: argoApplicationSpec.Spec.Source.Chart,
			ChartMetadata: ChartMetadata{
				ValuesFilename: argoApplicationSpec.Spec.Source.Helm.ValueFiles[0],
			},
		}
	}
	a.Destination = Destination{
		ClusterServerUrl: argoApplicationSpec.Spec.Destination.Server,
		Namespace:        argoApplicationSpec.Spec.Destination.Namespace,
	}
	a.Status = string(argoApplicationSpec.Status.Health.Status)
}

func (a *ApplicationMetadata) GetTargetClusterURL() string {
	return a.Destination.ClusterServerUrl
}

func (a *ApplicationMetadata) GetTargetClusterNamespace() string {
	return a.Destination.Namespace
}

func (a *ApplicationMetadata) UpdateTargetClusterURL(clusterURL string) {
	a.Destination.ClusterServerUrl = clusterURL
}

func (a *ApplicationMetadata) UpdateTargetClusterName(clusterName string) {
	a.Destination.ClusterName = clusterName
}

func (a *ApplicationMetadata) UpdateClusterData(cluster *bean.ClusterBean) {
	if cluster != nil {
		a.UpdateTargetClusterURL(cluster.ServerUrl)
		a.UpdateTargetClusterName(cluster.ClusterName)
	}
}

func (a *ApplicationMetadata) UpdateHelmChartData(chart *chart2.Chart) {
	a.Source.ChartMetadata.RequiredChartName = chart.Metadata.Name
	a.Source.ChartMetadata.RequiredChartVersion = chart.Metadata.Version
}

func (a *ApplicationMetadata) UpdateChartRefData(chartRef *bean2.ChartRefDto) {
	a.Source.ChartMetadata.SavedChartName = chartRef.Name
}

func (a *ApplicationMetadata) UpdateEnvironmentData(env *repository.Environment) {
	a.Destination.EnvironmentName = env.Name
	a.Destination.EnvironmentId = env.Id
}

func (a *ApplicationMetadata) GetRequiredChartName() string {
	return a.Source.ChartMetadata.RequiredChartName
}

func (a *ApplicationMetadata) GetRequiredChartVersion() string {
	return a.Source.ChartMetadata.RequiredChartVersion
}

func (a *ApplicationMetadata) GetSavedChartName() string {
	return a.Source.ChartMetadata.SavedChartName
}

func (r *HelmReleaseMetadata) UpdateReleaseData(release *gRPC.DeployedAppDetail) {
	r.Name = release.AppName
	r.Info = HelmReleaseInfo{
		Status: release.ReleaseStatus,
	}
	r.Destination = Destination{
		Namespace: release.EnvironmentDetail.Namespace,
	}
	r.Chart.HelmReleaseChartMetadata = HelmReleaseChartMetadata{
		RequiredChartName:    release.ChartName,
		Home:                 release.Home,
		RequiredChartVersion: release.ChartVersion,
		Icon:                 release.ChartAvatar,
	}
}

func (r *Destination) UpdateClusterData(cluster *bean.ClusterBean) {
	r.ClusterName = cluster.ClusterName
	r.ClusterServerUrl = cluster.ServerUrl
}

func (r *Destination) UpdateEnvironmentMetadata(environment *repository.Environment) {
	r.EnvironmentName = environment.Name
	r.EnvironmentId = environment.Id
	r.Namespace = environment.Namespace
}

func (r *HelmReleaseMetadata) UpdateChartRefData(chartRef *bean2.ChartRefDto) {
	r.Chart.HelmReleaseChartMetadata.SavedChartName = chartRef.Name
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

type FluxReleaseMetadata struct {
	RepoUrl              string      `json:"repoUrl"`
	RequiredChartName    string      `json:"requiredChartName"`
	SavedChartName       string      `json:"savedChartName"`
	RequiredChartVersion string      `json:"requiredChartVersion"`
	Destination          Destination `json:"destination"`
}

type HelmReleaseChart struct {
	HelmReleaseChartMetadata HelmReleaseChartMetadata `json:"metadata"`
}

type HelmReleaseChartMetadata struct {
	RequiredChartName    string `json:"requiredChartName"`
	SavedChartName       string `json:"savedChartName"`
	Home                 string `json:"home"`
	RequiredChartVersion string `json:"requiredChartVersion"`
	Icon                 string `json:"icon"`
	ApiVersion           string `json:"apiVersion"`
	Deprecated           bool   `json:"deprecated"`
}

type HelmReleaseInfo struct {
	Status string `json:"status"`
}

func (a *ExternalAppLinkValidationResponse) SetErrorDetail(err error) ExternalAppLinkValidationResponse {
	var linkFailedError LinkFailedError
	if errors.As(err, &linkFailedError) {
		a.ErrorDetail = &ErrorDetail{
			ValidationFailedReason:  linkFailedError.Reason,
			ValidationFailedMessage: linkFailedError.UserMessage,
		}
	}
	return *a
}

func (a *ExternalAppLinkValidationResponse) SetUnknownErrorDetail(err error) ExternalAppLinkValidationResponse {
	a.ErrorDetail = &ErrorDetail{
		ValidationFailedReason:  InternalServerError,
		ValidationFailedMessage: err.Error(),
	}
	return *a
}

type LinkFailedError struct {
	Reason      LinkFailedReason
	UserMessage string
}

type LinkFailedReason string

type ErrorDetail struct {
	ValidationFailedReason  LinkFailedReason `json:"validationFailedReason"`
	ValidationFailedMessage string           `json:"validationFailedMessage"`
}

func (l LinkFailedError) Error() string {
	return fmt.Sprintf("%s: %s", l.Reason, l.UserMessage)
}

const (
	ClusterNotFound                      LinkFailedReason = "ClusterNotFound"
	EnvironmentNotFound                  LinkFailedReason = "EnvironmentNotFound"
	ApplicationAlreadyPresent            LinkFailedReason = "ApplicationAlreadyPresent"
	UnsupportedApplicationSpec           LinkFailedReason = "UnsupportedApplicationSpec"
	ChartTypeMismatch                    LinkFailedReason = "ChartTypeMismatch"
	ChartVersionNotFound                 LinkFailedReason = "ChartVersionNotFound"
	GitOpsNotFound                       LinkFailedReason = "GitOpsNotFound"
	GitOpsOrganisationMismatch           LinkFailedReason = "GitOpsOrganisationMismatch"
	GitOpsRepoUrlAlreadyUsedInAnotherApp LinkFailedReason = "GitOpsRepoUrlAlreadyUsedInAnotherApp"
	InternalServerError                  LinkFailedReason = "InternalServerError"
	EnvironmentAlreadyPresent            LinkFailedReason = "EnvironmentAlreadyPresent"
	EnforcedPolicyViolation              LinkFailedReason = "EnforcedPolicyViolation"
	UnsupportedFluxHelmReleaseSpec       LinkFailedReason = "UnsupportedFluxHelmReleaseSpec"
)

const (
	ChartTypeMismatchErrorMsg    string = "External application uses '%s' chart where as this application uses '%s' chart. You can upload your own charts in Global Configuration > Deployment Charts."
	ChartVersionNotFoundErrorMsg string = "Chart version %s not found for %s chart"
	PipelineAlreadyPresentMsg    string = "A pipeline already exist for this environment."
	HelmAppAlreadyPresentMsg     string = "A helm app already exist for this environment."
)
