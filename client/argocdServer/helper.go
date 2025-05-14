/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package argocdServer

import (
	json2 "encoding/json"
	errors3 "errors"
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
