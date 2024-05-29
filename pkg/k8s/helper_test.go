/*
 * Copyright (c) 2024. Devtron Inc.
 */

package k8s

import "testing"

// not removing metadata from k8s version if exists, we only eliminate pre-release from k8s version
const (
	K8sVersionWithPreRelease               = "v1.25.16-eks-b9c9ed7"
	K8sVersionWithPreReleaseAndMetadata    = "v1.25.16-eks-b9c9ed7+acj23-as"
	K8sVersionWithMetadata                 = "v1.25.16+acj23-as"
	K8sVersionWithoutPreReleaseAndMetadata = "v1.25.16"
	InvalidK8sVersion                      = ""
)

func TestStripPrereleaseFromK8sVersion(t *testing.T) {
	type args struct {
		k8sVersion string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test1_K8sVersionWithPreRelease",
			args: args{k8sVersion: K8sVersionWithPreRelease},
			want: K8sVersionWithoutPreReleaseAndMetadata,
		},
		{
			name: "Test2_K8sVersionWithPreReleaseAndMetadata",
			args: args{k8sVersion: K8sVersionWithPreReleaseAndMetadata},
			want: K8sVersionWithMetadata,
		},
		{
			name: "Test3_K8sVersionWithMetadata",
			args: args{k8sVersion: K8sVersionWithMetadata},
			want: K8sVersionWithMetadata,
		},
		{
			name: "Test4_K8sVersionWithoutPrereleaseAndMetadata",
			args: args{k8sVersion: K8sVersionWithoutPreReleaseAndMetadata},
			want: K8sVersionWithoutPreReleaseAndMetadata,
		},
		{
			name: "Test5_EmptyK8sVersion",
			args: args{k8sVersion: InvalidK8sVersion},
			want: InvalidK8sVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripPrereleaseFromK8sVersion(tt.args.k8sVersion)
			if got != tt.want {
				t.Errorf("StripPrereleaseFromK8sVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
