package argocdServer

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRequestForArgoCDSyncModeUpdateRequest(argoApplication *v1alpha1.Application, autoSyncEnabled bool) *v1alpha1.Application {
	// set automated field in update request
	var automated *v1alpha1.SyncPolicyAutomated
	if autoSyncEnabled {
		automated = &v1alpha1.SyncPolicyAutomated{
			Prune: true,
		}
	}
	return &v1alpha1.Application{
		ObjectMeta: v1.ObjectMeta{
			Name:      argoApplication.Name,
			Namespace: DevtronInstalationNs,
		},
		Spec: v1alpha1.ApplicationSpec{
			Destination: argoApplication.Spec.Destination,
			Source:      argoApplication.Spec.Source,
			SyncPolicy: &v1alpha1.SyncPolicy{
				Automated:   automated,
				SyncOptions: argoApplication.Spec.SyncPolicy.SyncOptions,
				Retry:       argoApplication.Spec.SyncPolicy.Retry,
			}}}
}

func isArgoAppSyncModeMigrationNeeded(argoApplication *v1alpha1.Application, acdConfig *ACDConfig) bool {
	if acdConfig.IsManualSyncEnabled() && argoApplication.Spec.SyncPolicy.Automated != nil {
		return true
	} else if acdConfig.IsAutoSyncEnabled() && argoApplication.Spec.SyncPolicy.Automated == nil {
		return true
	}
	return false
}
