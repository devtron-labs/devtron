/*
Copyright 2020 The Flux authors

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

package meta

const (
	// ReconcileRequestAnnotation is the annotation used for triggering a reconciliation
	// outside of a defined interval. The value is interpreted as a token, and any change
	// in value SHOULD trigger a reconciliation.
	ReconcileRequestAnnotation string = "reconcile.fluxcd.io/requestedAt"

	// ForceRequestAnnotation is the annotation used for triggering a one-off forced
	// reconciliation, for example, of a HelmRelease when there are no new changes,
	// or of something that runs on a schedule when the schedule is not due at the moment.
	// The specific conditions for triggering a forced reconciliation depend on the
	// specific controller implementation, but the annotation is used to standardize
	// the mechanism across controllers. The value is interpreted as a token, and must
	// equal the value of ReconcileRequestAnnotation in order to trigger a release.
	ForceRequestAnnotation string = "reconcile.fluxcd.io/forceAt"
)

// ReconcileAnnotationValue returns a value for the reconciliation request annotation, which can be used to detect
// changes, and a boolean indicating whether the annotation was set.
func ReconcileAnnotationValue(annotations map[string]string) (string, bool) {
	requestedAt, ok := annotations[ReconcileRequestAnnotation]
	return requestedAt, ok
}

// ReconcileRequestStatus is a struct to embed in a status type, so that all types using the mechanism have the same
// field. Use it like this:
//
//		type FooStatus struct {
//	 	meta.ReconcileRequestStatus `json:",inline"`
//	 	// other status fields...
//		}
type ReconcileRequestStatus struct {
	// LastHandledReconcileAt holds the value of the most recent
	// reconcile request value, so a change of the annotation value
	// can be detected.
	// +optional
	LastHandledReconcileAt string `json:"lastHandledReconcileAt,omitempty"`
}

// GetLastHandledReconcileRequest returns the most recent reconcile request value from the ReconcileRequestStatus.
func (in ReconcileRequestStatus) GetLastHandledReconcileRequest() string {
	return in.LastHandledReconcileAt
}

// SetLastHandledReconcileRequest sets the most recent reconcile request value in the ReconcileRequestStatus.
func (in *ReconcileRequestStatus) SetLastHandledReconcileRequest(token string) {
	in.LastHandledReconcileAt = token
}

// StatusWithHandledReconcileRequest describes a status type which holds the value of the most recent
// ReconcileAnnotationValue.
// +k8s:deepcopy-gen=false
type StatusWithHandledReconcileRequest interface {
	GetLastHandledReconcileRequest() string
}

// StatusWithHandledReconcileRequestSetter describes a status with a setter for the most ReconcileAnnotationValue.
// +k8s:deepcopy-gen=false
type StatusWithHandledReconcileRequestSetter interface {
	SetLastHandledReconcileRequest(token string)
}

// ForceRequestStatus is a struct to embed in a status type, so that all types using the mechanism have the same
// field. Use it like this:
//
//		type FooStatus struct {
//	 	meta.ForceRequestStatus `json:",inline"`
//	 	// other status fields...
//		}
type ForceRequestStatus struct {
	// LastHandledForceAt holds the value of the most recent
	// force request value, so a change of the annotation value
	// can be detected.
	// +optional
	LastHandledForceAt string `json:"lastHandledForceAt,omitempty"`
}

// ShouldHandleForceRequest returns true if the object has a force request
// annotation, and the value of the annotation matches the value of the
// ReconcileRequestAnnotation annotation.
//
// To ensure that the force request is handled only once, the value of
// <ObjectType>Status.LastHandledForceAt is updated to match the value of the
// force request annotation (even if the force request is not handled because
// the value of the ReconcileRequestAnnotation annotation does not match).
func ShouldHandleForceRequest(obj interface {
	ObjectWithAnnotationRequests
	GetLastHandledForceRequestStatus() *string
}) bool {
	return HandleAnnotationRequest(obj, ForceRequestAnnotation, obj.GetLastHandledForceRequestStatus())
}

// ObjectWithAnnotationRequests is an interface that describes an object
// that has annotations and a status with a last handled reconcile request.
// +k8s:deepcopy-gen=false
type ObjectWithAnnotationRequests interface {
	GetAnnotations() map[string]string
	StatusWithHandledReconcileRequest
}

// HandleAnnotationRequest returns true if the object has a request annotation, and
// the value of the annotation matches the value of the ReconcileRequestAnnotation
// annotation.
//
// The lastHandled argument is used to ensure that the request is handled only
// once, and is updated to match the value of the request annotation (even if
// the request is not handled because the value of the ReconcileRequestAnnotation
// annotation does not match).
func HandleAnnotationRequest(obj ObjectWithAnnotationRequests, annotation string, lastHandled *string) bool {
	requestAt, requestOk := obj.GetAnnotations()[annotation]
	reconcileAt, reconcileOk := ReconcileAnnotationValue(obj.GetAnnotations())

	var lastHandledRequest string
	if requestOk {
		lastHandledRequest = *lastHandled
		*lastHandled = requestAt
	}

	if requestOk && reconcileOk && requestAt == reconcileAt {
		lastHandledReconcile := obj.GetLastHandledReconcileRequest()
		if lastHandledReconcile != reconcileAt && lastHandledRequest != requestAt {
			return true
		}
	}
	return false
}
