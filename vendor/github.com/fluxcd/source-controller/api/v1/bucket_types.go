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

const (
	// BucketKind is the string representation of a Bucket.
	BucketKind = "Bucket"
)

const (
	// BucketProviderGeneric for any S3 API compatible storage Bucket.
	BucketProviderGeneric string = "generic"
	// BucketProviderAmazon for an AWS S3 object storage Bucket.
	// Provides support for retrieving credentials from the AWS EC2 service.
	BucketProviderAmazon string = "aws"
	// BucketProviderGoogle for a Google Cloud Storage Bucket.
	// Provides support for authentication using a workload identity.
	BucketProviderGoogle string = "gcp"
	// BucketProviderAzure for an Azure Blob Storage Bucket.
	// Provides support for authentication using a Service Principal,
	// Managed Identity or Shared Key.
	BucketProviderAzure string = "azure"
)

// BucketSpec specifies the required configuration to produce an Artifact for
// an object storage bucket.
// +kubebuilder:validation:XValidation:rule="self.provider == 'aws' || self.provider == 'generic' || !has(self.sts)", message="STS configuration is only supported for the 'aws' and 'generic' Bucket providers"
// +kubebuilder:validation:XValidation:rule="self.provider != 'aws' || !has(self.sts) || self.sts.provider == 'aws'", message="'aws' is the only supported STS provider for the 'aws' Bucket provider"
// +kubebuilder:validation:XValidation:rule="self.provider != 'generic' || !has(self.sts) || self.sts.provider == 'ldap'", message="'ldap' is the only supported STS provider for the 'generic' Bucket provider"
// +kubebuilder:validation:XValidation:rule="!has(self.sts) || self.sts.provider != 'aws' || !has(self.sts.secretRef)", message="spec.sts.secretRef is not required for the 'aws' STS provider"
// +kubebuilder:validation:XValidation:rule="!has(self.sts) || self.sts.provider != 'aws' || !has(self.sts.certSecretRef)", message="spec.sts.certSecretRef is not required for the 'aws' STS provider"
type BucketSpec struct {
	// Provider of the object storage bucket.
	// Defaults to 'generic', which expects an S3 (API) compatible object
	// storage.
	// +kubebuilder:validation:Enum=generic;aws;gcp;azure
	// +kubebuilder:default:=generic
	// +optional
	Provider string `json:"provider,omitempty"`

	// BucketName is the name of the object storage bucket.
	// +required
	BucketName string `json:"bucketName"`

	// Endpoint is the object storage address the BucketName is located at.
	// +required
	Endpoint string `json:"endpoint"`

	// STS specifies the required configuration to use a Security Token
	// Service for fetching temporary credentials to authenticate in a
	// Bucket provider.
	//
	// This field is only supported for the `aws` and `generic` providers.
	// +optional
	STS *BucketSTSSpec `json:"sts,omitempty"`

	// Insecure allows connecting to a non-TLS HTTP Endpoint.
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// Region of the Endpoint where the BucketName is located in.
	// +optional
	Region string `json:"region,omitempty"`

	// Prefix to use for server-side filtering of files in the Bucket.
	// +optional
	Prefix string `json:"prefix,omitempty"`

	// SecretRef specifies the Secret containing authentication credentials
	// for the Bucket.
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
	// bucket. The client cert and key are useful if you are
	// authenticating with a certificate; the CA cert is useful if
	// you are using a self-signed server certificate. The Secret must
	// be of type `Opaque` or `kubernetes.io/tls`.
	//
	// This field is only supported for the `generic` provider.
	// +optional
	CertSecretRef *meta.LocalObjectReference `json:"certSecretRef,omitempty"`

	// ProxySecretRef specifies the Secret containing the proxy configuration
	// to use while communicating with the Bucket server.
	// +optional
	ProxySecretRef *meta.LocalObjectReference `json:"proxySecretRef,omitempty"`

	// Interval at which the Bucket Endpoint is checked for updates.
	// This interval is approximate and may be subject to jitter to ensure
	// efficient use of resources.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`

	// Timeout for fetch operations, defaults to 60s.
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

	// Suspend tells the controller to suspend the reconciliation of this
	// Bucket.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

// BucketSTSSpec specifies the required configuration to use a Security Token
// Service for fetching temporary credentials to authenticate in a Bucket
// provider.
type BucketSTSSpec struct {
	// Provider of the Security Token Service.
	// +kubebuilder:validation:Enum=aws;ldap
	// +required
	Provider string `json:"provider"`

	// Endpoint is the HTTP/S endpoint of the Security Token Service from
	// where temporary credentials will be fetched.
	// +required
	// +kubebuilder:validation:Pattern="^(http|https)://.*$"
	Endpoint string `json:"endpoint"`

	// SecretRef specifies the Secret containing authentication credentials
	// for the STS endpoint. This Secret must contain the fields `username`
	// and `password` and is supported only for the `ldap` provider.
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
	// STS endpoint. The client cert and key are useful if you are
	// authenticating with a certificate; the CA cert is useful if
	// you are using a self-signed server certificate. The Secret must
	// be of type `Opaque` or `kubernetes.io/tls`.
	//
	// This field is only supported for the `ldap` provider.
	// +optional
	CertSecretRef *meta.LocalObjectReference `json:"certSecretRef,omitempty"`
}

// BucketStatus records the observed state of a Bucket.
type BucketStatus struct {
	// ObservedGeneration is the last observed generation of the Bucket object.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the Bucket.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// URL is the dynamic fetch link for the latest Artifact.
	// It is provided on a "best effort" basis, and using the precise
	// BucketStatus.Artifact data is recommended.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the last successful Bucket reconciliation.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

	// ObservedIgnore is the observed exclusion patterns used for constructing
	// the source artifact.
	// +optional
	ObservedIgnore *string `json:"observedIgnore,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

const (
	// BucketOperationSucceededReason signals that the Bucket listing and fetch
	// operations succeeded.
	BucketOperationSucceededReason string = "BucketOperationSucceeded"

	// BucketOperationFailedReason signals that the Bucket listing or fetch
	// operations failed.
	BucketOperationFailedReason string = "BucketOperationFailed"
)

// GetConditions returns the status conditions of the object.
func (in *Bucket) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *Bucket) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the source must be reconciled again.
func (in *Bucket) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetArtifact returns the latest artifact from the source if present in the status sub-resource.
func (in *Bucket) GetArtifact() *Artifact {
	return in.Status.Artifact
}

// +genclient
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.spec.endpoint`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// Bucket is the Schema for the buckets API.
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BucketSpec `json:"spec,omitempty"`
	// +kubebuilder:default={"observedGeneration":-1}
	Status BucketStatus `json:"status,omitempty"`
}

// BucketList contains a list of Bucket objects.
// +kubebuilder:object:root=true
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}
