/*
Copyright 2022 The Flux authors

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

package oci

import (
	"github.com/devtron-labs/devtron/pkg/sourceController/bean"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// OCIRepositoryKind is the string representation of a OCIRepository.
	OCIRepositoryKind = "OCIRepository"

	// OCIRepositoryPrefix is the prefix used for OCIRepository URLs.
	OCIRepositoryPrefix = "oci://"

	// GenericOCIProvider provides support for authentication using static credentials
	// for any OCI compatible API such as Docker Registry, GitHub Container Registry,
	// Docker Hub, Quay, etc.
	GenericOCIProvider string = "generic"

	// AmazonOCIProvider provides support for OCI authentication using AWS IRSA.
	AmazonOCIProvider string = "aws"

	// GoogleOCIProvider provides support for OCI authentication using GCP workload identity.
	GoogleOCIProvider string = "gcp"

	// AzureOCIProvider provides support for OCI authentication using a Azure Service Principal,
	// Managed Identity or Shared Key.
	AzureOCIProvider string = "azure"

	// OCILayerExtract defines the operation type for extracting the content from an OCI artifact layer.
	OCILayerExtract = "extract"

	// OCILayerCopy defines the operation type for copying the content from an OCI artifact layer.
	OCILayerCopy = "copy"
)

// OCIRepositorySpec defines the desired state of OCIRepository
type OCIRepositorySpec struct {
	// URL is a reference to an OCI artifact repository hosted
	// on a remote container registry.
	// +kubebuilder:validation:Pattern="^oci://.*$"
	// +required
	URL string `json:"url"`

	// The OCI reference to pull and monitor for changes,
	// defaults to the latest tag.
	// +optional
	Reference *OCIRepositoryRef `json:"ref,omitempty"`

	// LayerSelector specifies which layer should be extracted from the OCI artifact.
	// When not specified, the first layer found in the artifact is selected.
	// +optional
	LayerSelector *OCILayerSelector `json:"layerSelector,omitempty"`

	// The provider used for authentication, can be 'aws', 'azure', 'gcp' or 'generic'.
	// When not specified, defaults to 'generic'.
	// +kubebuilder:validation:Enum=generic;aws;azure;gcp
	// +kubebuilder:default:=generic
	// +optional
	Provider string `json:"provider,omitempty"`

	// SecretRef contains the secret name containing the registry login
	// credentials to resolve image metadata.
	// The secret must be of type kubernetes.io/dockerconfigjson.
	// +optional
	//SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`

	// Verify contains the secret name containing the trusted public keys
	// used to verify the signature and specifies which provider to use to check
	// whether OCI image is authentic.
	// +optional
	Verify *OCIRepositoryVerification `json:"verify,omitempty"`

	// ServiceAccountName is the name of the Kubernetes ServiceAccount used to authenticate
	// the image pull if the service account has attached pull secrets. For more information:
	// https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

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
	// Note: Support for the `caFile`, `certFile` and `keyFile` keys have
	// been deprecated.
	// +optional
	//CertSecretRef *meta.LocalObjectReference `json:"certSecretRef,omitempty"`

	// Interval at which the OCIRepository URL is checked for updates.
	// This interval is approximate and may be subject to jitter to ensure
	// efficient use of resources.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`

	// The timeout for remote OCI Repository operations like pulling, defaults to 60s.
	// +kubebuilder:default="60s"
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Ignore overrides the set of excluded patterns in the .sourceignore format
	// (which is the same as .gitignore). If not provided, a default will be used,
	// consult the documentation for your version to find out what those are.
	// +optional
	Ignore *string `json:"ignore,omitempty"`

	// Insecure allows connecting to a non-TLS HTTP container registry.
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// This flag tells the controller to suspend the reconciliation of this source.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

// OCIRepositoryRef defines the image reference for the OCIRepository's URL
type OCIRepositoryRef struct {
	// Digest is the image digest to pull, takes precedence over SemVer.
	// The value should be in the format 'sha256:<HASH>'.
	// +optional
	Digest string `json:"digest,omitempty"`

	// SemVer is the range of tags to pull selecting the latest within
	// the range, takes precedence over Tag.
	// +optional
	SemVer string `json:"semver,omitempty"`

	// Tag is the image tag to pull, defaults to latest.
	// +optional
	Tag string `json:"tag,omitempty"`
}

// OCILayerSelector specifies which layer should be extracted from an OCI Artifact
type OCILayerSelector struct {
	// MediaType specifies the OCI media type of the layer
	// which should be extracted from the OCI Artifact. The
	// first layer matching this type is selected.
	// +optional
	MediaType string `json:"mediaType,omitempty"`

	// Operation specifies how the selected layer should be processed.
	// By default, the layer compressed content is extracted to storage.
	// When the operation is set to 'copy', the layer compressed content
	// is persisted to storage as it is.
	// +kubebuilder:validation:Enum=extract;copy
	// +optional
	Operation string `json:"operation,omitempty"`
}

// OCIRepositoryVerification verifies the authenticity of an OCI Artifact
type OCIRepositoryVerification struct {
	// Provider specifies the technology used to sign the OCI Artifact.
	// +kubebuilder:validation:Enum=cosign
	// +kubebuilder:default:=cosign
	Provider string `json:"provider"`

	// SecretRef specifies the Kubernetes Secret containing the
	// trusted public keys.
	// +optional
	//SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`
}

// OCIRepositoryStatus defines the observed state of OCIRepository
type OCIRepositoryStatus struct {
	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the OCIRepository.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// URL is the download link for the artifact output of the last OCI Repository sync.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the output of the last successful OCI Repository sync.
	// +optional
	Artifact *bean.Artifact `json:"artifact,omitempty"`

	// ContentConfigChecksum is a checksum of all the configurations related to
	// the content of the source artifact:
	//  - .spec.ignore
	//  - .spec.layerSelector
	// observed in .status.observedGeneration version of the object. This can
	// be used to determine if the content configuration has changed and the
	// artifact needs to be rebuilt.
	// It has the format of `<algo>:<checksum>`, for example: `sha256:<checksum>`.
	//
	// Deprecated: Replaced with explicit fields for observed artifact content
	// config in the status.
	// +optional
	ContentConfigChecksum string `json:"contentConfigChecksum,omitempty"`

	// ObservedIgnore is the observed exclusion patterns used for constructing
	// the source artifact.
	// +optional
	ObservedIgnore *string `json:"observedIgnore,omitempty"`

	// ObservedLayerSelector is the observed layer selector used for constructing
	// the source artifact.
	// +optional
	ObservedLayerSelector *OCILayerSelector `json:"observedLayerSelector,omitempty"`

	//meta.ReconcileRequestStatus `json:",inline"`
}

const (
	// OCIPullFailedReason signals that a pull operation failed.
	OCIPullFailedReason string = "OCIArtifactPullFailed"

	// OCILayerOperationFailedReason signals that an OCI layer operation failed.
	OCILayerOperationFailedReason string = "OCIArtifactLayerOperationFailed"
)

// GetConditions returns the status conditions of the object.
func (in OCIRepository) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *OCIRepository) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the OCIRepository must be
// reconciled again.
func (in OCIRepository) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetArtifact returns the latest Artifact from the OCIRepository if present in
// the status sub-resource.
func (in *OCIRepository) GetArtifact() *bean.Artifact {
	return in.Status.Artifact
}

// GetLayerMediaType returns the media type layer selector if found in spec.
func (in *OCIRepository) GetLayerMediaType() string {
	if in.Spec.LayerSelector == nil {
		return ""
	}

	return in.Spec.LayerSelector.MediaType
}

// GetLayerOperation returns the layer selector operation (defaults to extract).
func (in *OCIRepository) GetLayerOperation() string {
	if in.Spec.LayerSelector == nil || in.Spec.LayerSelector.Operation == "" {
		return OCILayerExtract
	}

	return in.Spec.LayerSelector.Operation
}

// +genclient
// +genclient:Namespaced
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=ocirepo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// OCIRepository is the Schema for the ocirepositories API
type OCIRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec OCIRepositorySpec `json:"spec,omitempty"`
	// +kubebuilder:default={"observedGeneration":-1}
	Status OCIRepositoryStatus `json:"status,omitempty"`
}

// OCIRepositoryList contains a list of OCIRepository
// +kubebuilder:object:root=true
type OCIRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OCIRepository `json:"items"`
}

//func init() {
//	SchemeBuilder.Register(&OCIRepository{}, &OCIRepositoryList{})
//}
