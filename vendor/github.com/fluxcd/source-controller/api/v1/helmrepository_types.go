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

	"github.com/fluxcd/pkg/apis/acl"
	"github.com/fluxcd/pkg/apis/meta"
)

const (
	// HelmRepositoryKind is the string representation of a HelmRepository.
	HelmRepositoryKind = "HelmRepository"
	// HelmRepositoryURLIndexKey is the key used for indexing HelmRepository
	// objects by their HelmRepositorySpec.URL.
	HelmRepositoryURLIndexKey = ".metadata.helmRepositoryURL"
	// HelmRepositoryTypeDefault is the default HelmRepository type.
	// It is used when no type is specified and corresponds to a Helm repository.
	HelmRepositoryTypeDefault = "default"
	// HelmRepositoryTypeOCI is the type for an OCI repository.
	HelmRepositoryTypeOCI = "oci"
)

// HelmRepositorySpec specifies the required configuration to produce an
// Artifact for a Helm repository index YAML.
type HelmRepositorySpec struct {
	// URL of the Helm repository, a valid URL contains at least a protocol and
	// host.
	// +kubebuilder:validation:Pattern="^(http|https|oci)://.*$"
	// +required
	URL string `json:"url"`

	// SecretRef specifies the Secret containing authentication credentials
	// for the HelmRepository.
	// For HTTP/S basic auth the secret must contain 'username' and 'password'
	// fields.
	// Support for TLS auth using the 'certFile' and 'keyFile', and/or 'caFile'
	// keys is deprecated. Please use `.spec.certSecretRef` instead.
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`

	// CertSecretRef can be given the name of a Secret containing
	// either or both of
	//
	// - a PEM-encoded client certificate (`tls.crt`) and private
	// key (`tls.key`);
	// - a PEM-encoded CA certificate (`ca.crt`)
	//
	// and whichever are supplied, will be used for connecting to the
	// registry. The client cert and key are useful if you are
	// authenticating with a certificate; the CA cert is useful if
	// you are using a self-signed server certificate. The Secret must
	// be of type `Opaque` or `kubernetes.io/tls`.
	//
	// It takes precedence over the values specified in the Secret referred
	// to by `.spec.secretRef`.
	// +optional
	CertSecretRef *meta.LocalObjectReference `json:"certSecretRef,omitempty"`

	// PassCredentials allows the credentials from the SecretRef to be passed
	// on to a host that does not match the host as defined in URL.
	// This may be required if the host of the advertised chart URLs in the
	// index differ from the defined URL.
	// Enabling this should be done with caution, as it can potentially result
	// in credentials getting stolen in a MITM-attack.
	// +optional
	PassCredentials bool `json:"passCredentials,omitempty"`

	// Interval at which the HelmRepository URL is checked for updates.
	// This interval is approximate and may be subject to jitter to ensure
	// efficient use of resources.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Interval metav1.Duration `json:"interval,omitempty"`

	// Insecure allows connecting to a non-TLS HTTP container registry.
	// This field is only taken into account if the .spec.type field is set to 'oci'.
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// Timeout is used for the index fetch operation for an HTTPS helm repository,
	// and for remote OCI Repository operations like pulling for an OCI helm
	// chart by the associated HelmChart.
	// Its default value is 60s.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Suspend tells the controller to suspend the reconciliation of this
	// HelmRepository.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// AccessFrom specifies an Access Control List for allowing cross-namespace
	// references to this object.
	// NOTE: Not implemented, provisional as of https://github.com/fluxcd/flux2/pull/2092
	// +optional
	AccessFrom *acl.AccessFrom `json:"accessFrom,omitempty"`

	// Type of the HelmRepository.
	// When this field is set to  "oci", the URL field value must be prefixed with "oci://".
	// +kubebuilder:validation:Enum=default;oci
	// +optional
	Type string `json:"type,omitempty"`

	// Provider used for authentication, can be 'aws', 'azure', 'gcp' or 'generic'.
	// This field is optional, and only taken into account if the .spec.type field is set to 'oci'.
	// When not specified, defaults to 'generic'.
	// +kubebuilder:validation:Enum=generic;aws;azure;gcp
	// +kubebuilder:default:=generic
	// +optional
	Provider string `json:"provider,omitempty"`
}

// HelmRepositoryStatus records the observed state of the HelmRepository.
type HelmRepositoryStatus struct {
	// ObservedGeneration is the last observed generation of the HelmRepository
	// object.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the HelmRepository.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// URL is the dynamic fetch link for the latest Artifact.
	// It is provided on a "best effort" basis, and using the precise
	// HelmRepositoryStatus.Artifact data is recommended.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the last successful HelmRepository reconciliation.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

const (
	// IndexationFailedReason signals that the HelmRepository index fetch
	// failed.
	IndexationFailedReason string = "IndexationFailed"
)

// GetConditions returns the status conditions of the object.
func (in HelmRepository) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *HelmRepository) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the source must be
// reconciled again.
func (in HelmRepository) GetRequeueAfter() time.Duration {
	if in.Spec.Interval.Duration != 0 {
		return in.Spec.Interval.Duration
	}
	return time.Minute
}

// GetTimeout returns the timeout duration used for various operations related
// to this HelmRepository.
func (in HelmRepository) GetTimeout() time.Duration {
	if in.Spec.Timeout != nil {
		return in.Spec.Timeout.Duration
	}
	return time.Minute
}

// GetArtifact returns the latest artifact from the source if present in the
// status sub-resource.
func (in *HelmRepository) GetArtifact() *Artifact {
	return in.Status.Artifact
}

// +genclient
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=helmrepo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// HelmRepository is the Schema for the helmrepositories API.
type HelmRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HelmRepositorySpec `json:"spec,omitempty"`
	// +kubebuilder:default={"observedGeneration":-1}
	Status HelmRepositoryStatus `json:"status,omitempty"`
}

// HelmRepositoryList contains a list of HelmRepository objects.
// +kubebuilder:object:root=true
type HelmRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmRepository{}, &HelmRepositoryList{})
}
