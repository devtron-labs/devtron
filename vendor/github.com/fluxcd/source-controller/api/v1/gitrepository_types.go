/*
Copyright 2023 The Flux authors

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
	// GitRepositoryKind is the string representation of a GitRepository.
	GitRepositoryKind = "GitRepository"

	// GitProviderGeneric provides support for authentication using
	// credentials specified in secretRef.
	GitProviderGeneric string = "generic"

	// GitProviderAzure provides support for authentication to azure
	// repositories using Managed Identity.
	GitProviderAzure string = "azure"

	// GitProviderGitHub provides support for authentication to git
	// repositories using GitHub App authentication
	GitProviderGitHub string = "github"
)

const (
	// IncludeUnavailableCondition indicates one of the includes is not
	// available. For example, because it does not exist, or does not have an
	// Artifact.
	// This is a "negative polarity" or "abnormal-true" type, and is only
	// present on the resource if it is True.
	IncludeUnavailableCondition string = "IncludeUnavailable"
)

// GitVerificationMode specifies the verification mode for a Git repository.
type GitVerificationMode string

// Valid checks the validity of the Git verification mode.
func (m GitVerificationMode) Valid() bool {
	switch m {
	case ModeGitHEAD, ModeGitTag, ModeGitTagAndHEAD:
		return true
	default:
		return false
	}
}

const (
	// ModeGitHEAD implies that the HEAD of the Git repository (after it has been
	// checked out to the required commit) should be verified.
	ModeGitHEAD GitVerificationMode = "HEAD"
	// ModeGitTag implies that the tag object specified in the checkout configuration
	// should be verified.
	ModeGitTag GitVerificationMode = "Tag"
	// ModeGitTagAndHEAD implies that both the tag object and the commit it points
	// to should be verified.
	ModeGitTagAndHEAD GitVerificationMode = "TagAndHEAD"
)

// GitRepositorySpec specifies the required configuration to produce an
// Artifact for a Git repository.
type GitRepositorySpec struct {
	// URL specifies the Git repository URL, it can be an HTTP/S or SSH address.
	// +kubebuilder:validation:Pattern="^(http|https|ssh)://.*$"
	// +required
	URL string `json:"url"`

	// SecretRef specifies the Secret containing authentication credentials for
	// the GitRepository.
	// For HTTPS repositories the Secret must contain 'username' and 'password'
	// fields for basic auth or 'bearerToken' field for token auth.
	// For SSH repositories the Secret must contain 'identity'
	// and 'known_hosts' fields.
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`

	// Provider used for authentication, can be 'azure', 'github', 'generic'.
	// When not specified, defaults to 'generic'.
	// +kubebuilder:validation:Enum=generic;azure;github
	// +optional
	Provider string `json:"provider,omitempty"`

	// Interval at which the GitRepository URL is checked for updates.
	// This interval is approximate and may be subject to jitter to ensure
	// efficient use of resources.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`

	// Timeout for Git operations like cloning, defaults to 60s.
	// +kubebuilder:default="60s"
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Reference specifies the Git reference to resolve and monitor for
	// changes, defaults to the 'master' branch.
	// +optional
	Reference *GitRepositoryRef `json:"ref,omitempty"`

	// Verification specifies the configuration to verify the Git commit
	// signature(s).
	// +optional
	Verification *GitRepositoryVerification `json:"verify,omitempty"`

	// ProxySecretRef specifies the Secret containing the proxy configuration
	// to use while communicating with the Git server.
	// +optional
	ProxySecretRef *meta.LocalObjectReference `json:"proxySecretRef,omitempty"`

	// Ignore overrides the set of excluded patterns in the .sourceignore format
	// (which is the same as .gitignore). If not provided, a default will be used,
	// consult the documentation for your version to find out what those are.
	// +optional
	Ignore *string `json:"ignore,omitempty"`

	// Suspend tells the controller to suspend the reconciliation of this
	// GitRepository.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// RecurseSubmodules enables the initialization of all submodules within
	// the GitRepository as cloned from the URL, using their default settings.
	// +optional
	RecurseSubmodules bool `json:"recurseSubmodules,omitempty"`

	// Include specifies a list of GitRepository resources which Artifacts
	// should be included in the Artifact produced for this GitRepository.
	// +optional
	Include []GitRepositoryInclude `json:"include,omitempty"`
}

// GitRepositoryInclude specifies a local reference to a GitRepository which
// Artifact (sub-)contents must be included, and where they should be placed.
type GitRepositoryInclude struct {
	// GitRepositoryRef specifies the GitRepository which Artifact contents
	// must be included.
	// +required
	GitRepositoryRef meta.LocalObjectReference `json:"repository"`

	// FromPath specifies the path to copy contents from, defaults to the root
	// of the Artifact.
	// +optional
	FromPath string `json:"fromPath,omitempty"`

	// ToPath specifies the path to copy contents to, defaults to the name of
	// the GitRepositoryRef.
	// +optional
	ToPath string `json:"toPath,omitempty"`
}

// GetFromPath returns the specified FromPath.
func (in *GitRepositoryInclude) GetFromPath() string {
	return in.FromPath
}

// GetToPath returns the specified ToPath, falling back to the name of the
// GitRepositoryRef.
func (in *GitRepositoryInclude) GetToPath() string {
	if in.ToPath == "" {
		return in.GitRepositoryRef.Name
	}
	return in.ToPath
}

// GitRepositoryRef specifies the Git reference to resolve and checkout.
type GitRepositoryRef struct {
	// Branch to check out, defaults to 'master' if no other field is defined.
	// +optional
	Branch string `json:"branch,omitempty"`

	// Tag to check out, takes precedence over Branch.
	// +optional
	Tag string `json:"tag,omitempty"`

	// SemVer tag expression to check out, takes precedence over Tag.
	// +optional
	SemVer string `json:"semver,omitempty"`

	// Name of the reference to check out; takes precedence over Branch, Tag and SemVer.
	//
	// It must be a valid Git reference: https://git-scm.com/docs/git-check-ref-format#_description
	// Examples: "refs/heads/main", "refs/tags/v0.1.0", "refs/pull/420/head", "refs/merge-requests/1/head"
	// +optional
	Name string `json:"name,omitempty"`

	// Commit SHA to check out, takes precedence over all reference fields.
	//
	// This can be combined with Branch to shallow clone the branch, in which
	// the commit is expected to exist.
	// +optional
	Commit string `json:"commit,omitempty"`
}

// GitRepositoryVerification specifies the Git commit signature verification
// strategy.
type GitRepositoryVerification struct {
	// Mode specifies which Git object(s) should be verified.
	//
	// The variants "head" and "HEAD" both imply the same thing, i.e. verify
	// the commit that the HEAD of the Git repository points to. The variant
	// "head" solely exists to ensure backwards compatibility.
	// +kubebuilder:validation:Enum=head;HEAD;Tag;TagAndHEAD
	// +optional
	// +kubebuilder:default:=HEAD
	Mode GitVerificationMode `json:"mode,omitempty"`

	// SecretRef specifies the Secret containing the public keys of trusted Git
	// authors.
	// +required
	SecretRef meta.LocalObjectReference `json:"secretRef"`
}

// GitRepositoryStatus records the observed state of a Git repository.
type GitRepositoryStatus struct {
	// ObservedGeneration is the last observed generation of the GitRepository
	// object.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the GitRepository.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Artifact represents the last successful GitRepository reconciliation.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

	// IncludedArtifacts contains a list of the last successfully included
	// Artifacts as instructed by GitRepositorySpec.Include.
	// +optional
	IncludedArtifacts []*Artifact `json:"includedArtifacts,omitempty"`

	// ObservedIgnore is the observed exclusion patterns used for constructing
	// the source artifact.
	// +optional
	ObservedIgnore *string `json:"observedIgnore,omitempty"`

	// ObservedRecurseSubmodules is the observed resource submodules
	// configuration used to produce the current Artifact.
	// +optional
	ObservedRecurseSubmodules bool `json:"observedRecurseSubmodules,omitempty"`

	// ObservedInclude is the observed list of GitRepository resources used to
	// produce the current Artifact.
	// +optional
	ObservedInclude []GitRepositoryInclude `json:"observedInclude,omitempty"`

	// SourceVerificationMode is the last used verification mode indicating
	// which Git object(s) have been verified.
	// +optional
	SourceVerificationMode *GitVerificationMode `json:"sourceVerificationMode,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

const (
	// GitOperationSucceedReason signals that a Git operation (e.g. clone,
	// checkout, etc.) succeeded.
	GitOperationSucceedReason string = "GitOperationSucceeded"

	// GitOperationFailedReason signals that a Git operation (e.g. clone,
	// checkout, etc.) failed.
	GitOperationFailedReason string = "GitOperationFailed"
)

// GetConditions returns the status conditions of the object.
func (in GitRepository) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *GitRepository) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the GitRepository must be
// reconciled again.
func (in GitRepository) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetArtifact returns the latest Artifact from the GitRepository if present in
// the status sub-resource.
func (in *GitRepository) GetArtifact() *Artifact {
	return in.Status.Artifact
}

// GetProvider returns the Git authentication provider.
func (v *GitRepository) GetProvider() string {
	if v.Spec.Provider == "" {
		return GitProviderGeneric
	}
	return v.Spec.Provider
}

// GetMode returns the declared GitVerificationMode, or a ModeGitHEAD default.
func (v *GitRepositoryVerification) GetMode() GitVerificationMode {
	if v.Mode.Valid() {
		return v.Mode
	}
	return ModeGitHEAD
}

// VerifyHEAD returns if the configured mode instructs verification of the
// Git HEAD.
func (v *GitRepositoryVerification) VerifyHEAD() bool {
	return v.GetMode() == ModeGitHEAD || v.GetMode() == ModeGitTagAndHEAD
}

// VerifyTag returns if the configured mode instructs verification of the
// Git tag.
func (v *GitRepositoryVerification) VerifyTag() bool {
	return v.GetMode() == ModeGitTag || v.GetMode() == ModeGitTagAndHEAD
}

// +genclient
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=gitrepo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// GitRepository is the Schema for the gitrepositories API.
type GitRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GitRepositorySpec `json:"spec,omitempty"`
	// +kubebuilder:default={"observedGeneration":-1}
	Status GitRepositoryStatus `json:"status,omitempty"`
}

// GitRepositoryList contains a list of GitRepository objects.
// +kubebuilder:object:root=true
type GitRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitRepository{}, &GitRepositoryList{})
}
