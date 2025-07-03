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

import "github.com/fluxcd/pkg/apis/meta"

const (
	// ForceRequestAnnotation is the annotation used for triggering a one-off forced
	// Helm release, even when there are no new changes in the HelmRelease.
	// The value is interpreted as a token, and must equal the value of
	// meta.ReconcileRequestAnnotation in order to trigger a release.
	ForceRequestAnnotation string = "reconcile.fluxcd.io/forceAt"

	// ResetRequestAnnotation is the annotation used for resetting the failure counts
	// of a HelmRelease, so that it can be retried again.
	// The value is interpreted as a token, and must equal the value of
	// meta.ReconcileRequestAnnotation in order to reset the failure counts.
	ResetRequestAnnotation string = "reconcile.fluxcd.io/resetAt"
)

// ShouldHandleResetRequest returns true if the HelmRelease has a reset request
// annotation, and the value of the annotation matches the value of the
// meta.ReconcileRequestAnnotation annotation.
//
// To ensure that the reset request is handled only once, the value of
// HelmReleaseStatus.LastHandledResetAt is updated to match the value of the
// reset request annotation (even if the reset request is not handled because
// the value of the meta.ReconcileRequestAnnotation annotation does not match).
func ShouldHandleResetRequest(obj *HelmRelease) bool {
	return handleRequest(obj, ResetRequestAnnotation, &obj.Status.LastHandledResetAt)
}

// ShouldHandleForceRequest returns true if the HelmRelease has a force request
// annotation, and the value of the annotation matches the value of the
// meta.ReconcileRequestAnnotation annotation.
//
// To ensure that the force request is handled only once, the value of
// HelmReleaseStatus.LastHandledForceAt is updated to match the value of the
// force request annotation (even if the force request is not handled because
// the value of the meta.ReconcileRequestAnnotation annotation does not match).
func ShouldHandleForceRequest(obj *HelmRelease) bool {
	return handleRequest(obj, ForceRequestAnnotation, &obj.Status.LastHandledForceAt)
}

// handleRequest returns true if the HelmRelease has a request annotation, and
// the value of the annotation matches the value of the meta.ReconcileRequestAnnotation
// annotation.
//
// The lastHandled argument is used to ensure that the request is handled only
// once, and is updated to match the value of the request annotation (even if
// the request is not handled because the value of the meta.ReconcileRequestAnnotation
// annotation does not match).
func handleRequest(obj *HelmRelease, annotation string, lastHandled *string) bool {
	requestAt, requestOk := obj.GetAnnotations()[annotation]
	reconcileAt, reconcileOk := meta.ReconcileAnnotationValue(obj.GetAnnotations())

	var lastHandledRequest string
	if requestOk {
		lastHandledRequest = *lastHandled
		*lastHandled = requestAt
	}

	if requestOk && reconcileOk && requestAt == reconcileAt {
		lastHandledReconcile := obj.Status.GetLastHandledReconcileRequest()
		if lastHandledReconcile != reconcileAt && lastHandledRequest != requestAt {
			return true
		}
	}
	return false
}
