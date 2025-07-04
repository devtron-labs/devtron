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

package v2

import (
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/fluxcd/pkg/apis/kustomize"
	"github.com/fluxcd/pkg/apis/meta"
)

const (
	// HelmReleaseKind is the kind in string format.
	HelmReleaseKind = "HelmRelease"
	// HelmReleaseFinalizer is set on a HelmRelease when it is first handled by
	// the controller, and removed when this object is deleted.
	HelmReleaseFinalizer = "finalizers.fluxcd.io"
)

const (
	// defaultMaxHistory is the default number of Helm release versions to keep.
	defaultMaxHistory = 5
)

// HelmReleaseSpec defines the desired state of a Helm release.
// +kubebuilder:validation:XValidation:rule="(has(self.chart) && !has(self.chartRef)) || (!has(self.chart) && has(self.chartRef))", message="either chart or chartRef must be set"
type HelmReleaseSpec struct {
	// Chart defines the template of the v1.HelmChart that should be created
	// for this HelmRelease.
	// +optional
	Chart *HelmChartTemplate `json:"chart,omitempty"`

	// ChartRef holds a reference to a source controller resource containing the
	// Helm chart artifact.
	// +optional
	ChartRef *CrossNamespaceSourceReference `json:"chartRef,omitempty"`

	// Interval at which to reconcile the Helm release.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`

	// KubeConfig for reconciling the HelmRelease on a remote cluster.
	// When used in combination with HelmReleaseSpec.ServiceAccountName,
	// forces the controller to act on behalf of that Service Account at the
	// target cluster.
	// If the --default-service-account flag is set, its value will be used as
	// a controller level fallback for when HelmReleaseSpec.ServiceAccountName
	// is empty.
	// +optional
	KubeConfig *meta.KubeConfigReference `json:"kubeConfig,omitempty"`

	// Suspend tells the controller to suspend reconciliation for this HelmRelease,
	// it does not apply to already started reconciliations. Defaults to false.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// ReleaseName used for the Helm release. Defaults to a composition of
	// '[TargetNamespace-]Name'.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=53
	// +kubebuilder:validation:Optional
	// +optional
	ReleaseName string `json:"releaseName,omitempty"`

	// TargetNamespace to target when performing operations for the HelmRelease.
	// Defaults to the namespace of the HelmRelease.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Optional
	// +optional
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// StorageNamespace used for the Helm storage.
	// Defaults to the namespace of the HelmRelease.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Optional
	// +optional
	StorageNamespace string `json:"storageNamespace,omitempty"`

	// DependsOn may contain a meta.NamespacedObjectReference slice with
	// references to HelmRelease resources that must be ready before this HelmRelease
	// can be reconciled.
	// +optional
	DependsOn []meta.NamespacedObjectReference `json:"dependsOn,omitempty"`

	// Timeout is the time to wait for any individual Kubernetes operation (like Jobs
	// for hooks) during the performance of a Helm action. Defaults to '5m0s'.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// MaxHistory is the number of revisions saved by Helm for this HelmRelease.
	// Use '0' for an unlimited number of revisions; defaults to '5'.
	// +optional
	MaxHistory *int `json:"maxHistory,omitempty"`

	// The name of the Kubernetes service account to impersonate
	// when reconciling this HelmRelease.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// PersistentClient tells the controller to use a persistent Kubernetes
	// client for this release. When enabled, the client will be reused for the
	// duration of the reconciliation, instead of being created and destroyed
	// for each (step of a) Helm action.
	//
	// This can improve performance, but may cause issues with some Helm charts
	// that for example do create Custom Resource Definitions during installation
	// outside Helm's CRD lifecycle hooks, which are then not observed to be
	// available by e.g. post-install hooks.
	//
	// If not set, it defaults to true.
	//
	// +optional
	PersistentClient *bool `json:"persistentClient,omitempty"`

	// DriftDetection holds the configuration for detecting and handling
	// differences between the manifest in the Helm storage and the resources
	// currently existing in the cluster.
	// +optional
	DriftDetection *DriftDetection `json:"driftDetection,omitempty"`

	// Install holds the configuration for Helm install actions for this HelmRelease.
	// +optional
	Install *Install `json:"install,omitempty"`

	// Upgrade holds the configuration for Helm upgrade actions for this HelmRelease.
	// +optional
	Upgrade *Upgrade `json:"upgrade,omitempty"`

	// Test holds the configuration for Helm test actions for this HelmRelease.
	// +optional
	Test *Test `json:"test,omitempty"`

	// Rollback holds the configuration for Helm rollback actions for this HelmRelease.
	// +optional
	Rollback *Rollback `json:"rollback,omitempty"`

	// Uninstall holds the configuration for Helm uninstall actions for this HelmRelease.
	// +optional
	Uninstall *Uninstall `json:"uninstall,omitempty"`

	// ValuesFrom holds references to resources containing Helm values for this HelmRelease,
	// and information about how they should be merged.
	ValuesFrom []ValuesReference `json:"valuesFrom,omitempty"`

	// Values holds the values for this Helm release.
	// +optional
	Values *apiextensionsv1.JSON `json:"values,omitempty"`

	// PostRenderers holds an array of Helm PostRenderers, which will be applied in order
	// of their definition.
	// +optional
	PostRenderers []PostRenderer `json:"postRenderers,omitempty"`
}

// +kubebuilder:object:generate=false
type ValuesReference = meta.ValuesReference

// Kustomize Helm PostRenderer specification.
type Kustomize struct {
	// Strategic merge and JSON patches, defined as inline YAML objects,
	// capable of targeting objects based on kind, label and annotation selectors.
	// +optional
	Patches []kustomize.Patch `json:"patches,omitempty"`

	// Images is a list of (image name, new name, new tag or digest)
	// for changing image names, tags or digests. This can also be achieved with a
	// patch, but this operator is simpler to specify.
	// +optional
	Images []kustomize.Image `json:"images,omitempty" json:"images,omitempty"`
}

// PostRenderer contains a Helm PostRenderer specification.
type PostRenderer struct {
	// Kustomization to apply as PostRenderer.
	// +optional
	Kustomize *Kustomize `json:"kustomize,omitempty"`
}

// DriftDetectionMode represents the modes in which a controller can detect and
// handle differences between the manifest in the Helm storage and the resources
// currently existing in the cluster.
type DriftDetectionMode string

var (
	// DriftDetectionEnabled instructs the controller to actively detect any
	// changes between the manifest in the Helm storage and the resources
	// currently existing in the cluster.
	// If any differences are detected, the controller will automatically
	// correct the cluster state by performing a Helm upgrade.
	DriftDetectionEnabled DriftDetectionMode = "enabled"

	// DriftDetectionWarn instructs the controller to actively detect any
	// changes between the manifest in the Helm storage and the resources
	// currently existing in the cluster.
	// If any differences are detected, the controller will emit a warning
	// without automatically correcting the cluster state.
	DriftDetectionWarn DriftDetectionMode = "warn"

	// DriftDetectionDisabled instructs the controller to skip detection of
	// differences entirely.
	// This is the default behavior, and the controller will not actively
	// detect or respond to differences between the manifest in the Helm
	// storage and the resources currently existing in the cluster.
	DriftDetectionDisabled DriftDetectionMode = "disabled"
)

var (
	// DriftDetectionMetadataKey is the label or annotation key used to disable
	// the diffing of an object.
	DriftDetectionMetadataKey = GroupVersion.Group + "/driftDetection"
	// DriftDetectionDisabledValue is the value used to disable the diffing of
	// an object using DriftDetectionMetadataKey.
	DriftDetectionDisabledValue = "disabled"
)

// IgnoreRule defines a rule to selectively disregard specific changes during
// the drift detection process.
type IgnoreRule struct {
	// Paths is a list of JSON Pointer (RFC 6901) paths to be excluded from
	// consideration in a Kubernetes object.
	// +required
	Paths []string `json:"paths"`

	// Target is a selector for specifying Kubernetes objects to which this
	// rule applies.
	// If Target is not set, the Paths will be ignored for all Kubernetes
	// objects within the manifest of the Helm release.
	// +optional
	Target *kustomize.Selector `json:"target,omitempty"`
}

// DriftDetection defines the strategy for performing differential analysis and
// provides a way to define rules for ignoring specific changes during this
// process.
type DriftDetection struct {
	// Mode defines how differences should be handled between the Helm manifest
	// and the manifest currently applied to the cluster.
	// If not explicitly set, it defaults to DiffModeDisabled.
	// +kubebuilder:validation:Enum=enabled;warn;disabled
	// +optional
	Mode DriftDetectionMode `json:"mode,omitempty"`

	// Ignore contains a list of rules for specifying which changes to ignore
	// during diffing.
	// +optional
	Ignore []IgnoreRule `json:"ignore,omitempty"`
}

// GetMode returns the DiffMode set on the Diff, or DiffModeDisabled if not
// set.
func (d DriftDetection) GetMode() DriftDetectionMode {
	if d.Mode == "" {
		return DriftDetectionDisabled
	}
	return d.Mode
}

// MustDetectChanges returns true if the DiffMode is set to DiffModeEnabled or
// DiffModeWarn.
func (d DriftDetection) MustDetectChanges() bool {
	return d.GetMode() == DriftDetectionEnabled || d.GetMode() == DriftDetectionWarn
}

// HelmChartTemplate defines the template from which the controller will
// generate a v1.HelmChart object in the same namespace as the referenced
// v1.Source.
type HelmChartTemplate struct {
	// ObjectMeta holds the template for metadata like labels and annotations.
	// +optional
	ObjectMeta *HelmChartTemplateObjectMeta `json:"metadata,omitempty"`

	// Spec holds the template for the v1.HelmChartSpec for this HelmRelease.
	// +required
	Spec HelmChartTemplateSpec `json:"spec"`
}

// HelmChartTemplateObjectMeta defines the template for the ObjectMeta of a
// v1.HelmChart.
type HelmChartTemplateObjectMeta struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// HelmChartTemplateSpec defines the template from which the controller will
// generate a v1.HelmChartSpec object.
type HelmChartTemplateSpec struct {
	// The name or path the Helm chart is available at in the SourceRef.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +required
	Chart string `json:"chart"`

	// Version semver expression, ignored for charts from v1.GitRepository and
	// v1beta2.Bucket sources. Defaults to latest when omitted.
	// +kubebuilder:default:=*
	// +optional
	Version string `json:"version,omitempty"`

	// The name and namespace of the v1.Source the chart is available at.
	// +required
	SourceRef CrossNamespaceObjectReference `json:"sourceRef"`

	// Interval at which to check the v1.Source for updates. Defaults to
	// 'HelmReleaseSpec.Interval'.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Interval *metav1.Duration `json:"interval,omitempty"`

	// Determines what enables the creation of a new artifact. Valid values are
	// ('ChartVersion', 'Revision').
	// See the documentation of the values for an explanation on their behavior.
	// Defaults to ChartVersion when omitted.
	// +kubebuilder:validation:Enum=ChartVersion;Revision
	// +kubebuilder:default:=ChartVersion
	// +optional
	ReconcileStrategy string `json:"reconcileStrategy,omitempty"`

	// Alternative list of values files to use as the chart values (values.yaml
	// is not included by default), expected to be a relative path in the SourceRef.
	// Values files are merged in the order of this list with the last file overriding
	// the first. Ignored when omitted.
	// +optional
	ValuesFiles []string `json:"valuesFiles,omitempty"`

	// IgnoreMissingValuesFiles controls whether to silently ignore missing values files rather than failing.
	// +optional
	IgnoreMissingValuesFiles bool `json:"ignoreMissingValuesFiles,omitempty"`

	// Verify contains the secret name containing the trusted public keys
	// used to verify the signature and specifies which provider to use to check
	// whether OCI image is authentic.
	// This field is only supported for OCI sources.
	// Chart dependencies, which are not bundled in the umbrella chart artifact,
	// are not verified.
	// +optional
	Verify *HelmChartTemplateVerification `json:"verify,omitempty"`
}

// GetInterval returns the configured interval for the v1.HelmChart,
// or the given default.
func (in HelmChartTemplate) GetInterval(defaultInterval metav1.Duration) metav1.Duration {
	if in.Spec.Interval == nil {
		return defaultInterval
	}
	return *in.Spec.Interval
}

// GetNamespace returns the namespace targeted namespace for the
// v1.HelmChart, or the given default.
func (in HelmChartTemplate) GetNamespace(defaultNamespace string) string {
	if in.Spec.SourceRef.Namespace == "" {
		return defaultNamespace
	}
	return in.Spec.SourceRef.Namespace
}

// HelmChartTemplateVerification verifies the authenticity of an OCI Helm chart.
type HelmChartTemplateVerification struct {
	// Provider specifies the technology used to sign the OCI Helm chart.
	// +kubebuilder:validation:Enum=cosign;notation
	// +kubebuilder:default:=cosign
	Provider string `json:"provider"`

	// SecretRef specifies the Kubernetes Secret containing the
	// trusted public keys.
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`
}

// Remediation defines a consistent interface for InstallRemediation and
// UpgradeRemediation.
// +kubebuilder:object:generate=false
type Remediation interface {
	GetRetries() int
	MustIgnoreTestFailures(bool) bool
	MustRemediateLastFailure() bool
	GetStrategy() RemediationStrategy
	GetFailureCount(hr *HelmRelease) int64
	IncrementFailureCount(hr *HelmRelease)
	RetriesExhausted(hr *HelmRelease) bool
}

// Install holds the configuration for Helm install actions performed for this
// HelmRelease.
type Install struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm install action. Defaults to
	// 'HelmReleaseSpec.Timeout'.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Remediation holds the remediation configuration for when the Helm install
	// action for the HelmRelease fails. The default is to not perform any action.
	// +optional
	Remediation *InstallRemediation `json:"remediation,omitempty"`

	// DisableTakeOwnership disables taking ownership of existing resources
	// during the Helm install action. Defaults to false.
	// +optional
	DisableTakeOwnership bool `json:"disableTakeOwnership,omitempty"`

	// DisableWait disables the waiting for resources to be ready after a Helm
	// install has been performed.
	// +optional
	DisableWait bool `json:"disableWait,omitempty"`

	// DisableWaitForJobs disables waiting for jobs to complete after a Helm
	// install has been performed.
	// +optional
	DisableWaitForJobs bool `json:"disableWaitForJobs,omitempty"`

	// DisableHooks prevents hooks from running during the Helm install action.
	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`

	// DisableOpenAPIValidation prevents the Helm install action from validating
	// rendered templates against the Kubernetes OpenAPI Schema.
	// +optional
	DisableOpenAPIValidation bool `json:"disableOpenAPIValidation,omitempty"`

	// DisableSchemaValidation prevents the Helm install action from validating
	// the values against the JSON Schema.
	// +optional
	DisableSchemaValidation bool `json:"disableSchemaValidation,omitempty"`

	// Replace tells the Helm install action to re-use the 'ReleaseName', but only
	// if that name is a deleted release which remains in the history.
	// +optional
	Replace bool `json:"replace,omitempty"`

	// SkipCRDs tells the Helm install action to not install any CRDs. By default,
	// CRDs are installed if not already present.
	//
	// Deprecated use CRD policy (`crds`) attribute with value `Skip` instead.
	//
	// +deprecated
	// +optional
	SkipCRDs bool `json:"skipCRDs,omitempty"`

	// CRDs upgrade CRDs from the Helm Chart's crds directory according
	// to the CRD upgrade policy provided here. Valid values are `Skip`,
	// `Create` or `CreateReplace`. Default is `Create` and if omitted
	// CRDs are installed but not updated.
	//
	// Skip: do neither install nor replace (update) any CRDs.
	//
	// Create: new CRDs are created, existing CRDs are neither updated nor deleted.
	//
	// CreateReplace: new CRDs are created, existing CRDs are updated (replaced)
	// but not deleted.
	//
	// By default, CRDs are applied (installed) during Helm install action.
	// With this option users can opt in to CRD replace existing CRDs on Helm
	// install actions, which is not (yet) natively supported by Helm.
	// https://helm.sh/docs/chart_best_practices/custom_resource_definitions.
	//
	// +kubebuilder:validation:Enum=Skip;Create;CreateReplace
	// +optional
	CRDs CRDsPolicy `json:"crds,omitempty"`

	// CreateNamespace tells the Helm install action to create the
	// HelmReleaseSpec.TargetNamespace if it does not exist yet.
	// On uninstall, the namespace will not be garbage collected.
	// +optional
	CreateNamespace bool `json:"createNamespace,omitempty"`
}

// GetTimeout returns the configured timeout for the Helm install action,
// or the given default.
func (in Install) GetTimeout(defaultTimeout metav1.Duration) metav1.Duration {
	if in.Timeout == nil {
		return defaultTimeout
	}
	return *in.Timeout
}

// GetRemediation returns the configured Remediation for the Helm install action.
func (in Install) GetRemediation() Remediation {
	if in.Remediation == nil {
		return InstallRemediation{}
	}
	return *in.Remediation
}

// InstallRemediation holds the configuration for Helm install remediation.
type InstallRemediation struct {
	// Retries is the number of retries that should be attempted on failures before
	// bailing. Remediation, using an uninstall, is performed between each attempt.
	// Defaults to '0', a negative integer equals to unlimited retries.
	// +optional
	Retries int `json:"retries,omitempty"`

	// IgnoreTestFailures tells the controller to skip remediation when the Helm
	// tests are run after an install action but fail. Defaults to
	// 'Test.IgnoreFailures'.
	// +optional
	IgnoreTestFailures *bool `json:"ignoreTestFailures,omitempty"`

	// RemediateLastFailure tells the controller to remediate the last failure, when
	// no retries remain. Defaults to 'false'.
	// +optional
	RemediateLastFailure *bool `json:"remediateLastFailure,omitempty"`
}

// GetRetries returns the number of retries that should be attempted on
// failures.
func (in InstallRemediation) GetRetries() int {
	return in.Retries
}

// MustIgnoreTestFailures returns the configured IgnoreTestFailures or the given
// default.
func (in InstallRemediation) MustIgnoreTestFailures(def bool) bool {
	if in.IgnoreTestFailures == nil {
		return def
	}
	return *in.IgnoreTestFailures
}

// MustRemediateLastFailure returns whether to remediate the last failure when
// no retries remain.
func (in InstallRemediation) MustRemediateLastFailure() bool {
	if in.RemediateLastFailure == nil {
		return false
	}
	return *in.RemediateLastFailure
}

// GetStrategy returns the strategy to use for failure remediation.
func (in InstallRemediation) GetStrategy() RemediationStrategy {
	return UninstallRemediationStrategy
}

// GetFailureCount gets the failure count.
func (in InstallRemediation) GetFailureCount(hr *HelmRelease) int64 {
	return hr.Status.InstallFailures
}

// IncrementFailureCount increments the failure count.
func (in InstallRemediation) IncrementFailureCount(hr *HelmRelease) {
	hr.Status.InstallFailures++
}

// RetriesExhausted returns true if there are no remaining retries.
func (in InstallRemediation) RetriesExhausted(hr *HelmRelease) bool {
	return in.Retries >= 0 && in.GetFailureCount(hr) > int64(in.Retries)
}

// CRDsPolicy defines the install/upgrade approach to use for CRDs when
// installing or upgrading a HelmRelease.
type CRDsPolicy string

const (
	// Skip CRDs do neither install nor replace (update) any CRDs.
	Skip CRDsPolicy = "Skip"
	// Create CRDs which do not already exist, do not replace (update) already existing
	// CRDs and keep (do not delete) CRDs which no longer exist in the current release.
	Create CRDsPolicy = "Create"
	// Create CRDs which do not already exist, Replace (update) already existing CRDs
	// and keep (do not delete) CRDs which no longer exist in the current release.
	CreateReplace CRDsPolicy = "CreateReplace"
)

// Upgrade holds the configuration for Helm upgrade actions for this
// HelmRelease.
type Upgrade struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm upgrade action. Defaults to
	// 'HelmReleaseSpec.Timeout'.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Remediation holds the remediation configuration for when the Helm upgrade
	// action for the HelmRelease fails. The default is to not perform any action.
	// +optional
	Remediation *UpgradeRemediation `json:"remediation,omitempty"`

	// DisableTakeOwnership disables taking ownership of existing resources
	// during the Helm upgrade action. Defaults to false.
	// +optional
	DisableTakeOwnership bool `json:"disableTakeOwnership,omitempty"`

	// DisableWait disables the waiting for resources to be ready after a Helm
	// upgrade has been performed.
	// +optional
	DisableWait bool `json:"disableWait,omitempty"`

	// DisableWaitForJobs disables waiting for jobs to complete after a Helm
	// upgrade has been performed.
	// +optional
	DisableWaitForJobs bool `json:"disableWaitForJobs,omitempty"`

	// DisableHooks prevents hooks from running during the Helm upgrade action.
	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`

	// DisableOpenAPIValidation prevents the Helm upgrade action from validating
	// rendered templates against the Kubernetes OpenAPI Schema.
	// +optional
	DisableOpenAPIValidation bool `json:"disableOpenAPIValidation,omitempty"`

	// DisableSchemaValidation prevents the Helm upgrade action from validating
	// the values against the JSON Schema.
	// +optional
	DisableSchemaValidation bool `json:"disableSchemaValidation,omitempty"`

	// Force forces resource updates through a replacement strategy.
	// +optional
	Force bool `json:"force,omitempty"`

	// PreserveValues will make Helm reuse the last release's values and merge in
	// overrides from 'Values'. Setting this flag makes the HelmRelease
	// non-declarative.
	// +optional
	PreserveValues bool `json:"preserveValues,omitempty"`

	// CleanupOnFail allows deletion of new resources created during the Helm
	// upgrade action when it fails.
	// +optional
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`

	// CRDs upgrade CRDs from the Helm Chart's crds directory according
	// to the CRD upgrade policy provided here. Valid values are `Skip`,
	// `Create` or `CreateReplace`. Default is `Skip` and if omitted
	// CRDs are neither installed nor upgraded.
	//
	// Skip: do neither install nor replace (update) any CRDs.
	//
	// Create: new CRDs are created, existing CRDs are neither updated nor deleted.
	//
	// CreateReplace: new CRDs are created, existing CRDs are updated (replaced)
	// but not deleted.
	//
	// By default, CRDs are not applied during Helm upgrade action. With this
	// option users can opt-in to CRD upgrade, which is not (yet) natively supported by Helm.
	// https://helm.sh/docs/chart_best_practices/custom_resource_definitions.
	//
	// +kubebuilder:validation:Enum=Skip;Create;CreateReplace
	// +optional
	CRDs CRDsPolicy `json:"crds,omitempty"`
}

// GetTimeout returns the configured timeout for the Helm upgrade action, or the
// given default.
func (in Upgrade) GetTimeout(defaultTimeout metav1.Duration) metav1.Duration {
	if in.Timeout == nil {
		return defaultTimeout
	}
	return *in.Timeout
}

// GetRemediation returns the configured Remediation for the Helm upgrade
// action.
func (in Upgrade) GetRemediation() Remediation {
	if in.Remediation == nil {
		return UpgradeRemediation{}
	}
	return *in.Remediation
}

// UpgradeRemediation holds the configuration for Helm upgrade remediation.
type UpgradeRemediation struct {
	// Retries is the number of retries that should be attempted on failures before
	// bailing. Remediation, using 'Strategy', is performed between each attempt.
	// Defaults to '0', a negative integer equals to unlimited retries.
	// +optional
	Retries int `json:"retries,omitempty"`

	// IgnoreTestFailures tells the controller to skip remediation when the Helm
	// tests are run after an upgrade action but fail.
	// Defaults to 'Test.IgnoreFailures'.
	// +optional
	IgnoreTestFailures *bool `json:"ignoreTestFailures,omitempty"`

	// RemediateLastFailure tells the controller to remediate the last failure, when
	// no retries remain. Defaults to 'false' unless 'Retries' is greater than 0.
	// +optional
	RemediateLastFailure *bool `json:"remediateLastFailure,omitempty"`

	// Strategy to use for failure remediation. Defaults to 'rollback'.
	// +kubebuilder:validation:Enum=rollback;uninstall
	// +optional
	Strategy *RemediationStrategy `json:"strategy,omitempty"`
}

// GetRetries returns the number of retries that should be attempted on
// failures.
func (in UpgradeRemediation) GetRetries() int {
	return in.Retries
}

// MustIgnoreTestFailures returns the configured IgnoreTestFailures or the given
// default.
func (in UpgradeRemediation) MustIgnoreTestFailures(def bool) bool {
	if in.IgnoreTestFailures == nil {
		return def
	}
	return *in.IgnoreTestFailures
}

// MustRemediateLastFailure returns whether to remediate the last failure when
// no retries remain.
func (in UpgradeRemediation) MustRemediateLastFailure() bool {
	if in.RemediateLastFailure == nil {
		return in.Retries > 0
	}
	return *in.RemediateLastFailure
}

// GetStrategy returns the strategy to use for failure remediation.
func (in UpgradeRemediation) GetStrategy() RemediationStrategy {
	if in.Strategy == nil {
		return RollbackRemediationStrategy
	}
	return *in.Strategy
}

// GetFailureCount gets the failure count.
func (in UpgradeRemediation) GetFailureCount(hr *HelmRelease) int64 {
	return hr.Status.UpgradeFailures
}

// IncrementFailureCount increments the failure count.
func (in UpgradeRemediation) IncrementFailureCount(hr *HelmRelease) {
	hr.Status.UpgradeFailures++
}

// RetriesExhausted returns true if there are no remaining retries.
func (in UpgradeRemediation) RetriesExhausted(hr *HelmRelease) bool {
	return in.Retries >= 0 && in.GetFailureCount(hr) > int64(in.Retries)
}

// RemediationStrategy returns the strategy to use to remediate a failed install
// or upgrade.
type RemediationStrategy string

const (
	// RollbackRemediationStrategy represents a Helm remediation strategy of Helm
	// rollback.
	RollbackRemediationStrategy RemediationStrategy = "rollback"

	// UninstallRemediationStrategy represents a Helm remediation strategy of Helm
	// uninstall.
	UninstallRemediationStrategy RemediationStrategy = "uninstall"
)

// Test holds the configuration for Helm test actions for this HelmRelease.
type Test struct {
	// Enable enables Helm test actions for this HelmRelease after an Helm install
	// or upgrade action has been performed.
	// +optional
	Enable bool `json:"enable,omitempty"`

	// Timeout is the time to wait for any individual Kubernetes operation during
	// the performance of a Helm test action. Defaults to 'HelmReleaseSpec.Timeout'.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// IgnoreFailures tells the controller to skip remediation when the Helm tests
	// are run but fail. Can be overwritten for tests run after install or upgrade
	// actions in 'Install.IgnoreTestFailures' and 'Upgrade.IgnoreTestFailures'.
	// +optional
	IgnoreFailures bool `json:"ignoreFailures,omitempty"`

	// Filters is a list of tests to run or exclude from running.
	Filters *[]Filter `json:"filters,omitempty"`
}

// GetTimeout returns the configured timeout for the Helm test action,
// or the given default.
func (in Test) GetTimeout(defaultTimeout metav1.Duration) metav1.Duration {
	if in.Timeout == nil {
		return defaultTimeout
	}
	return *in.Timeout
}

// Filter holds the configuration for individual Helm test filters.
type Filter struct {
	// Name is the name of the test.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Name string `json:"name"`
	// Exclude specifies whether the named test should be excluded.
	// +optional
	Exclude bool `json:"exclude,omitempty"`
}

// GetFilters returns the configured filters for the Helm test action/
func (in Test) GetFilters() []Filter {
	if in.Filters == nil {
		var filters []Filter
		return filters
	}
	return *in.Filters
}

// Rollback holds the configuration for Helm rollback actions for this
// HelmRelease.
type Rollback struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm rollback action. Defaults to
	// 'HelmReleaseSpec.Timeout'.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// DisableWait disables the waiting for resources to be ready after a Helm
	// rollback has been performed.
	// +optional
	DisableWait bool `json:"disableWait,omitempty"`

	// DisableWaitForJobs disables waiting for jobs to complete after a Helm
	// rollback has been performed.
	// +optional
	DisableWaitForJobs bool `json:"disableWaitForJobs,omitempty"`

	// DisableHooks prevents hooks from running during the Helm rollback action.
	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`

	// Recreate performs pod restarts for the resource if applicable.
	// +optional
	Recreate bool `json:"recreate,omitempty"`

	// Force forces resource updates through a replacement strategy.
	// +optional
	Force bool `json:"force,omitempty"`

	// CleanupOnFail allows deletion of new resources created during the Helm
	// rollback action when it fails.
	// +optional
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`
}

// GetTimeout returns the configured timeout for the Helm rollback action, or
// the given default.
func (in Rollback) GetTimeout(defaultTimeout metav1.Duration) metav1.Duration {
	if in.Timeout == nil {
		return defaultTimeout
	}
	return *in.Timeout
}

// Uninstall holds the configuration for Helm uninstall actions for this
// HelmRelease.
type Uninstall struct {
	// Timeout is the time to wait for any individual Kubernetes operation (like
	// Jobs for hooks) during the performance of a Helm uninstall action. Defaults
	// to 'HelmReleaseSpec.Timeout'.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// DisableHooks prevents hooks from running during the Helm rollback action.
	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`

	// KeepHistory tells Helm to remove all associated resources and mark the
	// release as deleted, but retain the release history.
	// +optional
	KeepHistory bool `json:"keepHistory,omitempty"`

	// DisableWait disables waiting for all the resources to be deleted after
	// a Helm uninstall is performed.
	// +optional
	DisableWait bool `json:"disableWait,omitempty"`

	// DeletionPropagation specifies the deletion propagation policy when
	// a Helm uninstall is performed.
	// +kubebuilder:default=background
	// +kubebuilder:validation:Enum=background;foreground;orphan
	// +optional
	DeletionPropagation *string `json:"deletionPropagation,omitempty"`
}

// GetTimeout returns the configured timeout for the Helm uninstall action, or
// the given default.
func (in Uninstall) GetTimeout(defaultTimeout metav1.Duration) metav1.Duration {
	if in.Timeout == nil {
		return defaultTimeout
	}
	return *in.Timeout
}

// GetDeletionPropagation returns the configured deletion propagation policy
// for the Helm uninstall action, or 'background'.
func (in Uninstall) GetDeletionPropagation() string {
	if in.DeletionPropagation == nil {
		return "background"
	}
	return *in.DeletionPropagation
}

// ReleaseAction is the action to perform a Helm release.
type ReleaseAction string

const (
	// ReleaseActionInstall represents a Helm install action.
	ReleaseActionInstall ReleaseAction = "install"
	// ReleaseActionUpgrade represents a Helm upgrade action.
	ReleaseActionUpgrade ReleaseAction = "upgrade"
)

// HelmReleaseStatus defines the observed state of a HelmRelease.
type HelmReleaseStatus struct {
	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ObservedPostRenderersDigest is the digest for the post-renderers of
	// the last successful reconciliation attempt.
	// +optional
	ObservedPostRenderersDigest string `json:"observedPostRenderersDigest,omitempty"`

	// LastAttemptedGeneration is the last generation the controller attempted
	// to reconcile.
	// +optional
	LastAttemptedGeneration int64 `json:"lastAttemptedGeneration,omitempty"`

	// Conditions holds the conditions for the HelmRelease.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// HelmChart is the namespaced name of the HelmChart resource created by
	// the controller for the HelmRelease.
	// +optional
	HelmChart string `json:"helmChart,omitempty"`

	// StorageNamespace is the namespace of the Helm release storage for the
	// current release.
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// +optional
	StorageNamespace string `json:"storageNamespace,omitempty"`

	// History holds the history of Helm releases performed for this HelmRelease
	// up to the last successfully completed release.
	// +optional
	History Snapshots `json:"history,omitempty"`

	// LastAttemptedReleaseAction is the last release action performed for this
	// HelmRelease. It is used to determine the active remediation strategy.
	// +kubebuilder:validation:Enum=install;upgrade
	// +optional
	LastAttemptedReleaseAction ReleaseAction `json:"lastAttemptedReleaseAction,omitempty"`

	// Failures is the reconciliation failure count against the latest desired
	// state. It is reset after a successful reconciliation.
	// +optional
	Failures int64 `json:"failures,omitempty"`

	// InstallFailures is the install failure count against the latest desired
	// state. It is reset after a successful reconciliation.
	// +optional
	InstallFailures int64 `json:"installFailures,omitempty"`

	// UpgradeFailures is the upgrade failure count against the latest desired
	// state. It is reset after a successful reconciliation.
	// +optional
	UpgradeFailures int64 `json:"upgradeFailures,omitempty"`

	// LastAttemptedRevision is the Source revision of the last reconciliation
	// attempt. For OCIRepository  sources, the 12 first characters of the digest are
	// appended to the chart version e.g. "1.2.3+1234567890ab".
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// LastAttemptedRevisionDigest is the digest of the last reconciliation attempt.
	// This is only set for OCIRepository sources.
	// +optional
	LastAttemptedRevisionDigest string `json:"lastAttemptedRevisionDigest,omitempty"`

	// LastAttemptedValuesChecksum is the SHA1 checksum for the values of the last
	// reconciliation attempt.
	// Deprecated: Use LastAttemptedConfigDigest instead.
	// +optional
	LastAttemptedValuesChecksum string `json:"lastAttemptedValuesChecksum,omitempty"`

	// LastReleaseRevision is the revision of the last successful Helm release.
	// Deprecated: Use History instead.
	// +optional
	LastReleaseRevision int `json:"lastReleaseRevision,omitempty"`

	// LastAttemptedConfigDigest is the digest for the config (better known as
	// "values") of the last reconciliation attempt.
	// +optional
	LastAttemptedConfigDigest string `json:"lastAttemptedConfigDigest,omitempty"`

	// LastHandledForceAt holds the value of the most recent force request
	// value, so a change of the annotation value can be detected.
	// +optional
	LastHandledForceAt string `json:"lastHandledForceAt,omitempty"`

	// LastHandledResetAt holds the value of the most recent reset request
	// value, so a change of the annotation value can be detected.
	// +optional
	LastHandledResetAt string `json:"lastHandledResetAt,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

// ClearHistory clears the History.
func (in *HelmReleaseStatus) ClearHistory() {
	in.History = nil
}

// ClearFailures clears the failure counters.
func (in *HelmReleaseStatus) ClearFailures() {
	in.Failures = 0
	in.InstallFailures = 0
	in.UpgradeFailures = 0
}

// GetHelmChart returns the namespace and name of the HelmChart.
func (in HelmReleaseStatus) GetHelmChart() (string, string) {
	if in.HelmChart == "" {
		return "", ""
	}
	if split := strings.Split(in.HelmChart, string(types.Separator)); len(split) > 1 {
		return split[0], split[1]
	}
	return "", ""
}

func (in *HelmReleaseStatus) GetLastAttemptedRevision() string {
	return in.LastAttemptedRevision
}

const (
	// SourceIndexKey is the key used for indexing HelmReleases based on
	// their sources.
	SourceIndexKey string = ".metadata.source"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=hr
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// HelmRelease is the Schema for the helmreleases API
type HelmRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HelmReleaseSpec `json:"spec,omitempty"`
	// +kubebuilder:default:={"observedGeneration":-1}
	Status HelmReleaseStatus `json:"status,omitempty"`
}

// GetDriftDetection returns the configuration for detecting and handling
// differences between the manifest in the Helm storage and the resources
// currently existing in the cluster.
func (in *HelmRelease) GetDriftDetection() DriftDetection {
	if in.Spec.DriftDetection == nil {
		return DriftDetection{}
	}
	return *in.Spec.DriftDetection
}

// GetInstall returns the configuration for Helm install actions for the
// HelmRelease.
func (in *HelmRelease) GetInstall() Install {
	if in.Spec.Install == nil {
		return Install{}
	}
	return *in.Spec.Install
}

// GetUpgrade returns the configuration for Helm upgrade actions for this
// HelmRelease.
func (in *HelmRelease) GetUpgrade() Upgrade {
	if in.Spec.Upgrade == nil {
		return Upgrade{}
	}
	return *in.Spec.Upgrade
}

// GetTest returns the configuration for Helm test actions for this HelmRelease.
func (in *HelmRelease) GetTest() Test {
	if in.Spec.Test == nil {
		return Test{}
	}
	return *in.Spec.Test
}

// GetRollback returns the configuration for Helm rollback actions for this
// HelmRelease.
func (in *HelmRelease) GetRollback() Rollback {
	if in.Spec.Rollback == nil {
		return Rollback{}
	}
	return *in.Spec.Rollback
}

// GetUninstall returns the configuration for Helm uninstall actions for this
// HelmRelease.
func (in *HelmRelease) GetUninstall() Uninstall {
	if in.Spec.Uninstall == nil {
		return Uninstall{}
	}
	return *in.Spec.Uninstall
}

// GetActiveRemediation returns the active Remediation configuration for the
// HelmRelease.
func (in HelmRelease) GetActiveRemediation() Remediation {
	switch in.Status.LastAttemptedReleaseAction {
	case ReleaseActionInstall:
		return in.GetInstall().GetRemediation()
	case ReleaseActionUpgrade:
		return in.GetUpgrade().GetRemediation()
	default:
		return nil
	}
}

// GetRequeueAfter returns the duration after which the HelmRelease
// must be reconciled again.
func (in HelmRelease) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetValues unmarshals the raw values to a map[string]interface{} and returns
// the result.
func (in HelmRelease) GetValues() map[string]interface{} {
	var values map[string]interface{}
	if in.Spec.Values != nil {
		_ = yaml.Unmarshal(in.Spec.Values.Raw, &values)
	}
	return values
}

// GetReleaseName returns the configured release name, or a composition of
// '[TargetNamespace-]Name'.
func (in HelmRelease) GetReleaseName() string {
	if in.Spec.ReleaseName != "" {
		return in.Spec.ReleaseName
	}
	if in.Spec.TargetNamespace != "" {
		return strings.Join([]string{in.Spec.TargetNamespace, in.Name}, "-")
	}
	return in.Name
}

// GetReleaseNamespace returns the configured TargetNamespace, or the namespace
// of the HelmRelease.
func (in HelmRelease) GetReleaseNamespace() string {
	if in.Spec.TargetNamespace != "" {
		return in.Spec.TargetNamespace
	}
	return in.Namespace
}

// GetStorageNamespace returns the configured StorageNamespace for helm, or the namespace
// of the HelmRelease.
func (in HelmRelease) GetStorageNamespace() string {
	if in.Spec.StorageNamespace != "" {
		return in.Spec.StorageNamespace
	}
	return in.Namespace
}

// GetHelmChartName returns the name used by the controller for the HelmChart creation.
func (in HelmRelease) GetHelmChartName() string {
	return strings.Join([]string{in.Namespace, in.Name}, "-")
}

// GetTimeout returns the configured Timeout, or the default of 300s.
func (in HelmRelease) GetTimeout() metav1.Duration {
	if in.Spec.Timeout == nil {
		return metav1.Duration{Duration: 300 * time.Second}
	}
	return *in.Spec.Timeout
}

// GetMaxHistory returns the configured MaxHistory, or the default of 5.
func (in HelmRelease) GetMaxHistory() int {
	if in.Spec.MaxHistory == nil {
		return defaultMaxHistory
	}
	return *in.Spec.MaxHistory
}

// UsePersistentClient returns the configured PersistentClient, or the default
// of true.
func (in HelmRelease) UsePersistentClient() bool {
	if in.Spec.PersistentClient == nil {
		return true
	}
	return *in.Spec.PersistentClient
}

// GetDependsOn returns the list of dependencies across-namespaces.
func (in HelmRelease) GetDependsOn() []meta.NamespacedObjectReference {
	return in.Spec.DependsOn
}

// GetConditions returns the status conditions of the object.
func (in HelmRelease) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *HelmRelease) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// HasChartRef returns true if the HelmRelease has a ChartRef.
func (in *HelmRelease) HasChartRef() bool {
	return in.Spec.ChartRef != nil
}

// HasChartTemplate returns true if the HelmRelease has a ChartTemplate.
func (in *HelmRelease) HasChartTemplate() bool {
	return in.Spec.Chart != nil
}

// +kubebuilder:object:root=true

// HelmReleaseList contains a list of HelmRelease objects.
type HelmReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmRelease{}, &HelmReleaseList{})
}
