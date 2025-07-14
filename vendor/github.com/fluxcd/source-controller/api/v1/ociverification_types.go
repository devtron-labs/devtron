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
	"github.com/fluxcd/pkg/apis/meta"
)

// OCIRepositoryVerification verifies the authenticity of an OCI Artifact
type OCIRepositoryVerification struct {
	// Provider specifies the technology used to sign the OCI Artifact.
	// +kubebuilder:validation:Enum=cosign;notation
	// +kubebuilder:default:=cosign
	Provider string `json:"provider"`

	// SecretRef specifies the Kubernetes Secret containing the
	// trusted public keys.
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`

	// MatchOIDCIdentity specifies the identity matching criteria to use
	// while verifying an OCI artifact which was signed using Cosign keyless
	// signing. The artifact's identity is deemed to be verified if any of the
	// specified matchers match against the identity.
	// +optional
	MatchOIDCIdentity []OIDCIdentityMatch `json:"matchOIDCIdentity,omitempty"`
}

// OIDCIdentityMatch specifies options for verifying the certificate identity,
// i.e. the issuer and the subject of the certificate.
type OIDCIdentityMatch struct {
	// Issuer specifies the regex pattern to match against to verify
	// the OIDC issuer in the Fulcio certificate. The pattern must be a
	// valid Go regular expression.
	// +required
	Issuer string `json:"issuer"`
	// Subject specifies the regex pattern to match against to verify
	// the identity subject in the Fulcio certificate. The pattern must
	// be a valid Go regular expression.
	// +required
	Subject string `json:"subject"`
}
