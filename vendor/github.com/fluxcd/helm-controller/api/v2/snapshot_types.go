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
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// snapshotStatusDeployed indicates that the release the snapshot was taken
	// from is currently deployed.
	snapshotStatusDeployed = "deployed"
	// snapshotStatusSuperseded indicates that the release the snapshot was taken
	// from has been superseded by a newer release.
	snapshotStatusSuperseded = "superseded"

	// snapshotTestPhaseFailed indicates that the test of the release the snapshot
	// was taken from has failed.
	snapshotTestPhaseFailed = "Failed"
)

// Snapshots is a list of Snapshot objects.
type Snapshots []*Snapshot

// Len returns the number of Snapshots.
func (in Snapshots) Len() int {
	return len(in)
}

// SortByVersion sorts the Snapshots by version, in descending order.
func (in Snapshots) SortByVersion() {
	sort.Slice(in, func(i, j int) bool {
		return in[i].Version > in[j].Version
	})
}

// Latest returns the most recent Snapshot.
func (in Snapshots) Latest() *Snapshot {
	if len(in) == 0 {
		return nil
	}
	in.SortByVersion()
	return in[0]
}

// Previous returns the most recent Snapshot before the Latest that has a
// status of "deployed" or "superseded", or nil if there is no such Snapshot.
// Unless ignoreTests is true, Snapshots with a test in the "Failed" phase are
// ignored.
func (in Snapshots) Previous(ignoreTests bool) *Snapshot {
	if len(in) < 2 {
		return nil
	}
	in.SortByVersion()
	for i := range in[1:] {
		s := in[i+1]
		if s.Status == snapshotStatusDeployed || s.Status == snapshotStatusSuperseded {
			if ignoreTests || !s.HasTestInPhase(snapshotTestPhaseFailed) {
				return s
			}
		}
	}
	return nil
}

// Truncate removes all Snapshots up to the Previous deployed Snapshot.
// If there is no previous-deployed Snapshot, the most recent 5 Snapshots are
// retained.
func (in *Snapshots) Truncate(ignoreTests bool) {
	if in.Len() < 2 {
		return
	}

	in.SortByVersion()
	for i := range (*in)[1:] {
		s := (*in)[i+1]
		if s.Status == snapshotStatusDeployed || s.Status == snapshotStatusSuperseded {
			if ignoreTests || !s.HasTestInPhase(snapshotTestPhaseFailed) {
				*in = (*in)[:i+2]
				return
			}
		}
	}

	if in.Len() > defaultMaxHistory {
		// If none of the Snapshots are deployed or superseded, and there
		// are more than the defaultMaxHistory, truncate to the most recent
		// Snapshots.
		*in = (*in)[:defaultMaxHistory]
	}
}

// Snapshot captures a point-in-time copy of the status information for a Helm release,
// as managed by the controller.
type Snapshot struct {
	// APIVersion is the API version of the Snapshot.
	// Provisional: when the calculation method of the Digest field is changed,
	// this field will be used to distinguish between the old and new methods.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
	// Digest is the checksum of the release object in storage.
	// It has the format of `<algo>:<checksum>`.
	// +required
	Digest string `json:"digest"`
	// Name is the name of the release.
	// +required
	Name string `json:"name"`
	// Namespace is the namespace the release is deployed to.
	// +required
	Namespace string `json:"namespace"`
	// Version is the version of the release object in storage.
	// +required
	Version int `json:"version"`
	// Status is the current state of the release.
	// +required
	Status string `json:"status"`
	// ChartName is the chart name of the release object in storage.
	// +required
	ChartName string `json:"chartName"`
	// ChartVersion is the chart version of the release object in
	// storage.
	// +required
	ChartVersion string `json:"chartVersion"`
	// AppVersion is the chart app version of the release object in storage.
	// +optional
	AppVersion string `json:"appVersion,omitempty"`
	// ConfigDigest is the checksum of the config (better known as
	// "values") of the release object in storage.
	// It has the format of `<algo>:<checksum>`.
	// +required
	ConfigDigest string `json:"configDigest"`
	// FirstDeployed is when the release was first deployed.
	// +required
	FirstDeployed metav1.Time `json:"firstDeployed"`
	// LastDeployed is when the release was last deployed.
	// +required
	LastDeployed metav1.Time `json:"lastDeployed"`
	// Deleted is when the release was deleted.
	// +optional
	Deleted metav1.Time `json:"deleted,omitempty"`
	// TestHooks is the list of test hooks for the release as observed to be
	// run by the controller.
	// +optional
	TestHooks *map[string]*TestHookStatus `json:"testHooks,omitempty"`
	// OCIDigest is the digest of the OCI artifact associated with the release.
	// +optional
	OCIDigest string `json:"ociDigest,omitempty"`
}

// FullReleaseName returns the full name of the release in the format
// of '<namespace>/<name>.<version>
func (in *Snapshot) FullReleaseName() string {
	if in == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s.v%d", in.Namespace, in.Name, in.Version)
}

// VersionedChartName returns the full name of the chart in the format of
// '<name>@<version>'.
func (in *Snapshot) VersionedChartName() string {
	if in == nil {
		return ""
	}
	return fmt.Sprintf("%s@%s", in.ChartName, in.ChartVersion)
}

// HasBeenTested returns true if TestHooks is not nil. This includes an empty
// map, which indicates the chart has no tests.
func (in *Snapshot) HasBeenTested() bool {
	return in != nil && in.TestHooks != nil
}

// GetTestHooks returns the TestHooks for the release if not nil.
func (in *Snapshot) GetTestHooks() map[string]*TestHookStatus {
	if in == nil || in.TestHooks == nil {
		return nil
	}
	return *in.TestHooks
}

// HasTestInPhase returns true if any of the TestHooks is in the given phase.
func (in *Snapshot) HasTestInPhase(phase string) bool {
	if in != nil {
		for _, h := range in.GetTestHooks() {
			if h.Phase == phase {
				return true
			}
		}
	}
	return false
}

// SetTestHooks sets the TestHooks for the release.
func (in *Snapshot) SetTestHooks(hooks map[string]*TestHookStatus) {
	if in == nil || hooks == nil {
		return
	}
	in.TestHooks = &hooks
}

// Targets returns true if the Snapshot targets the given release data.
func (in *Snapshot) Targets(name, namespace string, version int) bool {
	if in != nil {
		return in.Name == name && in.Namespace == namespace && in.Version == version
	}
	return false
}

// TestHookStatus holds the status information for a test hook as observed
// to be run by the controller.
type TestHookStatus struct {
	// LastStarted is the time the test hook was last started.
	// +optional
	LastStarted metav1.Time `json:"lastStarted,omitempty"`
	// LastCompleted is the time the test hook last completed.
	// +optional
	LastCompleted metav1.Time `json:"lastCompleted,omitempty"`
	// Phase the test hook was observed to be in.
	// +optional
	Phase string `json:"phase,omitempty"`
}
