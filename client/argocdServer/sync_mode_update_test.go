package argocdServer

import (
	"testing"

	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSyncModeUpdatePreservesApplication(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		app := &v1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "argo-control", ResourceVersion: "42", Labels: map[string]string{"owner": "example"}, Finalizers: []string{"resources-finalizer.argocd.argoproj.io"}},
			Spec: v1alpha1.ApplicationSpec{
				Project:     "applications",
				Sources:     v1alpha1.ApplicationSources{{RepoURL: "https://example.org/config.git", Path: "app", TargetRevision: "main"}},
				Destination: v1alpha1.ApplicationDestination{Server: "https://kubernetes.default.svc", Namespace: "workload"},
				SyncPolicy:  &v1alpha1.SyncPolicy{SyncOptions: v1alpha1.SyncOptions{"CreateNamespace=true"}, Retry: &v1alpha1.RetryStrategy{Limit: 3}},
			},
		}
		before := app.DeepCopy()
		updated := createRequestForArgoCDSyncModeUpdateRequest(app, enabled)
		require.Equal(t, enabled, updated.Spec.SyncPolicy.Automated != nil)
		if enabled {
			require.True(t, *updated.Spec.SyncPolicy.Automated.Prune)
		}
		updated.Spec.SyncPolicy.Automated = before.Spec.SyncPolicy.Automated
		require.Equal(t, before, updated, "Only automated sync may change")
		updated.Labels["owner"] = "changed"
		updated.Spec.Sources[0].Path = "changed"
		require.Equal(t, before, app, "The fetched Application must remain unmodified")
	}
}
