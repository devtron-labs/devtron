package bean

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Result is a type for creating an abstraction for the controller-runtime
// reconcile Result to simplify the Result values.
type Result int

const (
	HostUrlKey      string = "url"
	WebHookHostUrl  string = "http://devtron-service.devtroncd.svc.cluster.local/orchestrator/webhook/ext-ci/"
	External        string = "ext"
	MaterialTypeGit string = "git"
)

const (
	// ResultEmpty indicates a reconcile result which does not requeue. It is
	// also used when returning an error, since the error overshadows result.
	ResultEmpty Result = iota
	// ResultRequeue indicates a reconcile result which should immediately
	// requeue.
	ResultRequeue
	// ResultSuccess indicates a reconcile success result.
	// For a reconciler that requeues regularly at a fixed interval, runtime
	// result with a fixed RequeueAfter is success result.
	// For a reconciler that doesn't requeue on successful reconciliation,
	// an empty runtime result is success result.
	// It is usually returned at the end of a reconciler/sub-reconciler.
	ResultSuccess
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

	// The provider used for authentication, can be 'aws', 'azure', 'gcp' or 'generic'.
	// When not specified, defaults to 'generic'.
	// +kubebuilder:validation:Enum=generic;aws;azure;gcp
	// +kubebuilder:default:=generic
	// +optional
	Provider string `json:"provider,omitempty"`

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
	//		CertSecretRef *meta.LocalObjectReference `json:"certSecretRef,omitempty"`

	// Interval at which the OCIRepository URL is checked for updates.
	// This interval is approximate and may be subject to jitter to ensure
	// efficient use of resources.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval v1.Duration `json:"interval"`

	// The timeout for remote OCI Repository operations like pulling, defaults to 60s.
	// +kubebuilder:default="60s"
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m))+$"
	// +optional
	Timeout *v1.Duration `json:"timeout,omitempty"`

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

// Artifact represents the output of a Source reconciliation.
type Artifact struct {
	// Path is the relative file path of the Artifact. It can be used to locate
	// the file in the root of the Artifact storage on the local file system of
	// the controller managing the Source.
	// +required
	Path string `json:"path"`

	// URL is the HTTP address of the Artifact as exposed by the controller
	// managing the Source. It can be used to retrieve the Artifact for
	// consumption, e.g. by another controller applying the Artifact contents.
	// +required
	URL string `json:"url"`

	// Revision is a human-readable identifier traceable in the origin source
	// system. It can be a Git commit SHA, Git tag, a Helm chart version, etc.
	// +required
	Revision string `json:"revision"`

	// Digest is the digest of the file in the form of '<algorithm>:<checksum>'.
	// +optional
	// +kubebuilder:validation:Pattern="^[a-z0-9]+(?:[.+_-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$"
	Digest string `json:"digest,omitempty"`

	// Size is the number of bytes in the file.
	// +optional
	Size *int64 `json:"size,omitempty"`

	// Metadata holds upstream information such as OCI annotations.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`
}

// OCIRepositoryStatus defines the observed state of OCIRepository
type OCIRepositoryStatus struct {
	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// URL is the download link for the artifact output of the last OCI Repository sync.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the output of the last successful OCI Repository sync.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

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
}
