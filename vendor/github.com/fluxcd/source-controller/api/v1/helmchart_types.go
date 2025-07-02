/*
Copyright 2024 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluxcd/pkg/apis/meta"
)

// HelmChartKind is the string representation of a HelmChart.
const HelmChartKind = "HelmChart"

// HelmChartSpec specifies the desired state of a Helm chart.
type HelmChartSpec struct {
	// Chart is the name or path the Helm chart is available at in the
	// SourceRef.
	// +required
	Chart string `json:"chart"`

	// Version is the chart version semver expression, ignored for charts from
	// GitRepository and Bucket sources. Defaults to latest when omitted.
	// +kubebuilder:default:=*
	// +optional
	Version string `json:"version,omitempty"`

	// SourceRef is the reference to the Source the chart is available at.
	// +required
	SourceRef LocalHelmChartSourceReference `json:"sourceRef"`

	// Interval at which the HelmChart SourceRef is checked for updates.
	// This interval is approximate and may be subject to jitter to ensure
	// efficient use of resources.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`

	// ReconcileStrategy determines what enables the creation of a new artifact.
	// Valid values are ('ChartVersion', 'Revision').
	// See the documentation of the values for an explanation on their behavior.
	// Defaults to ChartVersion when omitted.
	// +kubebuilder:validation:Enum=ChartVersion;Revision
	// +kubebuilder:default:=ChartVersion
	// +optional
	ReconcileStrategy string `json:"reconcileStrategy,omitempty"`

	// ValuesFiles is an alternative list of values files to use as the chart
	// values (values.yaml is not included by default), expected to be a
	// relative path in the SourceRef.
	// Values files are merged in the order of this list with the last file
	// overriding the first. Ignored when omitted.
	// +optional
	ValuesFiles []string `json:"valuesFiles,omitempty"`

	// IgnoreMissingValuesFiles controls whether to silently ignore missing values
	// files rather than failing.
	// +optional
	IgnoreMissingValuesFiles bool `json:"ignoreMissingValuesFiles,omitempty"`

	// Suspend tells the controller to suspend the reconciliation of this
	// source.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// Verify contains the secret name containing the trusted public keys
	// used to verify the signature and specifies which provider to use to check
	// whether OCI image is authentic.
	// This field is only supported when using HelmRepository source with spec.type 'oci'.
	// Chart dependencies, which are not bundled in the umbrella chart artifact, are not verified.
	// +optional
	Verify *OCIRepositoryVerification `json:"verify,omitempty"`
}

const (
	// ReconcileStrategyChartVersion reconciles when the version of the Helm chart is different.
	ReconcileStrategyChartVersion string = "ChartVersion"

	// ReconcileStrategyRevision reconciles when the Revision of the source is different.
	ReconcileStrategyRevision string = "Revision"
)

// LocalHelmChartSourceReference contains enough information to let you locate
// the typed referenced object at namespace level.
type LocalHelmChartSourceReference struct {
	// APIVersion of the referent.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referent, valid values are ('HelmRepository', 'GitRepository',
	// 'Bucket').
	// +kubebuilder:validation:Enum=HelmRepository;GitRepository;Bucket
	// +required
	Kind string `json:"kind"`

	// Name of the referent.
	// +required
	Name string `json:"name"`
}

// HelmChartStatus records the observed state of the HelmChart.
type HelmChartStatus struct {
	// ObservedGeneration is the last observed generation of the HelmChart
	// object.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ObservedSourceArtifactRevision is the last observed Artifact.Revision
	// of the HelmChartSpec.SourceRef.
	// +optional
	ObservedSourceArtifactRevision string `json:"observedSourceArtifactRevision,omitempty"`

	// ObservedChartName is the last observed chart name as specified by the
	// resolved chart reference.
	// +optional
	ObservedChartName string `json:"observedChartName,omitempty"`

	// ObservedValuesFiles are the observed value files of the last successful
	// reconciliation.
	// It matches the chart in the last successfully reconciled artifact.
	// +optional
	ObservedValuesFiles []string `json:"observedValuesFiles,omitempty"`

	// Conditions holds the conditions for the HelmChart.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// URL is the dynamic fetch link for the latest Artifact.
	// It is provided on a "best effort" basis, and using the precise
	// BucketStatus.Artifact data is recommended.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the output of the last successful reconciliation.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

const (
	// ChartPullSucceededReason signals that the pull of the Helm chart
	// succeeded.
	ChartPullSucceededReason string = "ChartPullSucceeded"

	// ChartPackageSucceededReason signals that the package of the Helm
	// chart succeeded.
	ChartPackageSucceededReason string = "ChartPackageSucceeded"
)

// GetConditions returns the status conditions of the object.
func (in HelmChart) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *HelmChart) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the source must be
// reconciled again.
func (in HelmChart) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetArtifact returns the latest artifact from the source if present in the
// status sub-resource.
func (in *HelmChart) GetArtifact() *Artifact {
	return in.Status.Artifact
}

// GetValuesFiles returns a merged list of HelmChartSpec.ValuesFiles.
func (in *HelmChart) GetValuesFiles() []string {
	return in.Spec.ValuesFiles
}

// +genclient
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=hc
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Chart",type=string,JSONPath=`.spec.chart`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
// +kubebuilder:printcolumn:name="Source Kind",type=string,JSONPath=`.spec.sourceRef.kind`
// +kubebuilder:printcolumn:name="Source Name",type=string,JSONPath=`.spec.sourceRef.name`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// HelmChart is the Schema for the helmcharts API.
type HelmChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HelmChartSpec `json:"spec,omitempty"`
	// +kubebuilder:default={"observedGeneration":-1}
	Status HelmChartStatus `json:"status,omitempty"`
}

// HelmChartList contains a list of HelmChart objects.
// +kubebuilder:object:root=true
type HelmChartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmChart `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmChart{}, &HelmChartList{})
}
