package argocdServer

import (
	json2 "encoding/json"
	errors3 "errors"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"k8s.io/utils/ptr"
)

func createRequestForArgoCDSyncModeUpdateRequest(argoApplication *v1alpha1.Application, autoSyncEnabled bool) *v1alpha1.Application {
	// set automated field in update request
	var automated *v1alpha1.SyncPolicyAutomated
	if autoSyncEnabled {
		automated = &v1alpha1.SyncPolicyAutomated{
			// argo-cd v3.4 changed these fields to *bool; leaving Enabled nil
			// keeps automated sync on, matching the pre-v3.4 behaviour.
			Prune: ptr.To(true),
		}
	}
	updated := argoApplication.DeepCopy()
	if updated.Spec.SyncPolicy == nil {
		updated.Spec.SyncPolicy = &v1alpha1.SyncPolicy{}
	}
	updated.Spec.SyncPolicy.Automated = automated
	return updated
}

func isArgoAppSyncModeMigrationNeeded(argoApplication *v1alpha1.Application, acdConfig *ACDConfig) bool {
	if acdConfig.IsManualSyncEnabled() && argoApplication.Spec.SyncPolicy.Automated != nil {
		return true
	} else if acdConfig.IsAutoSyncEnabled() && argoApplication.Spec.SyncPolicy.Automated == nil {
		return true
	}
	return false
}

func GetAppObject(appMapObj map[string]interface{}) (*v1alpha1.Application, error) {
	if appMapObj == nil {
		return nil, errors3.New("found empty application object")
	}
	appJson, err := json2.Marshal(appMapObj)
	if err != nil {
		return nil, err
	}
	var app v1alpha1.Application
	err = json2.Unmarshal(appJson, &app)
	if err != nil {
		return nil, err
	}
	return &app, err
}
