/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package application

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/util/settings"
	"testing"

	"go.uber.org/zap"
)

func TestServiceClientImpl_getRolloutStatus(t *testing.T) {
	type fields struct {
		asc    *settings.ArgoCDSettings
		logger *zap.SugaredLogger
	}
	type args struct {
		rolloutManifest     map[string]interface{}
		replicaSetManifests []map[string]interface{}
		serviceManifests    []map[string]interface{}
		resp                *v1alpha1.ApplicationTree
	}
	logger, _ := zap.NewDevelopment()
	data1 := getData1()
	data2 := getData2()
	data3 := getData3()
	data4 := getData4()
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantNewReplicaSet string
		wantStatus        string
	}{
		{
			name: "test1",
			fields: fields{
				asc:    nil,
				logger: logger.Sugar(),
			},
			args: args{
				rolloutManifest:     data1.rollout,
				replicaSetManifests: data1.replicaSet,
				serviceManifests:    data1.service,
				resp:                data1.respTree,
			},
			wantNewReplicaSet: "app-deployment-9-7c797c6d54",
			wantStatus:        Healthy,
		},
		{
			name: "test2",
			fields: fields{
				asc:    nil,
				logger: logger.Sugar(),
			},
			args: args{
				rolloutManifest:     data2.rollout,
				replicaSetManifests: data2.replicaSet,
				serviceManifests:    data2.service,
				resp:                data2.respTree,
			},
			wantNewReplicaSet: "app-deployment-9-7c797c6d54",
			wantStatus:        Degraded,
		},
		{
			name: "test3",
			fields: fields{
				asc:    nil,
				logger: logger.Sugar(),
			},
			args: args{
				rolloutManifest:     data3.rollout,
				replicaSetManifests: data3.replicaSet,
				serviceManifests:    data3.service,
				resp:                data3.respTree,
			},
			wantNewReplicaSet: "app-deployment-9-7c797c6d54",
			wantStatus:        Suspended,
		},
		{
			name: "test4",
			fields: fields{
				asc:    nil,
				logger: logger.Sugar(),
			},
			args: args{
				rolloutManifest:     data4.rollout,
				replicaSetManifests: data4.replicaSet,
				serviceManifests:    data4.service,
				resp:                data4.respTree,
			},
			wantNewReplicaSet: "app-deployment-9-7c797c6d54",
			wantStatus:        Degraded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/*c := ServiceClientImpl{
				settings:    tt.fields.asc,
				logger: tt.fields.logger,
			}
			gotNewReplicaSet, gotStatus := c.getRolloutNewReplicaSetName(tt.args.rolloutManifest, tt.args.replicaSetManifests, tt.args.serviceManifests, tt.args.resp)
			if gotNewReplicaSet != tt.wantNewReplicaSet {
				t.Errorf("ServiceClientImpl.getRolloutNewReplicaSetName() gotNewReplicaSet = %v, want %v", gotNewReplicaSet, tt.wantNewReplicaSet)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("ServiceClientImpl.getRolloutNewReplicaSetName() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}*/
		})
	}
}

/*
desired replica count is met
active service has been switched
old replica set has been brought down
*/
func getData1() *DataTest {
	rolloutManifest := `{"apiVersion":"argoproj.io/v1alpha1","kind":"Rollout","metadata":{"creationTimestamp":"2019-04-16T11:35:42Z","generation":1,"name":"app-deployment-9","namespace":"dont-use2"},"spec":{"pauseForSecondsBeforeSwitchActive":100,"strategy":{"blueGreen":{"activeService":"app-service","previewService":"app-service-preview"}},"waitForSecondsBeforeScalingDown":100},"status":{"blueGreen":{"activeSelector":"7c797c6d54"},"currentPodHash":"7c797c6d54","observedGeneration":"cf98b99fb"}}`
	replicaSetManifests := `[{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"},"creationTimestamp":"2019-04-21T14:21:54Z","generation":27,"name":"app-deployment-9-7c797c6d54","namespace":"dont-use2"},"spec":{"replicas":1,"template":{"metadata":{"labels":{"rollouts-pod-template-hash":"7c797c6d54"}}}},"status":{"availableReplicas":1,"fullyLabeledReplicas":1,"observedGeneration":27,"readyReplicas":1,"replicas":1}},{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"creationTimestamp":"2019-04-16T11:35:43Z","generation":28,"labels":{"rollouts-pod-template-hash":"66b7dd4665"},"name":"app-deployment-9-66b7dd4665","namespace":"dont-use2"},"spec":{"replicas":0,"template":{"metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"66b7dd4665"}}}},"status":{"observedGeneration":28,"replicas":0}}]`
	serviceManifests := `[{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"}}},{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service-preview","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":""}}}]`
	resp := `{"items":[{"group":"argoproj.io","version":"v1alpha1","kind":"Rollout","namespace":"dont-use2","name":"app-deployment-9","children":[{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-66b7dd4665","resourceVersion":"8545875"},{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54","children":[{"version":"v1","kind":"Pod","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54-k5xvs","info":[{"name":"Status Reason","value":"Running"},{"name":"Containers","value":"1/1"}],"resourceVersion":"8545204"}],"resourceVersion":"8545278"}],"resourceVersion":"8545878"}]}`
	rollout, replicaSet, service, respTree := convertDataForTest(rolloutManifest, replicaSetManifests, serviceManifests, resp)
	return &DataTest{
		rollout:    rollout,
		replicaSet: replicaSet,
		service:    service,
		respTree:   respTree,
	}
}

/*
actual replica count is zero
irrespective of active service
pause switch is set but has been running for too long hence degraded performance
*/
func getData2() *DataTest {
	rolloutManifest := `{"apiVersion":"argoproj.io/v1alpha1","kind":"Rollout","metadata":{"creationTimestamp":"2019-04-16T11:35:42Z","generation":1,"name":"app-deployment-9","namespace":"dont-use2"},"spec":{"pauseForSecondsBeforeSwitchActive":100,"strategy":{"blueGreen":{"activeService":"app-service","previewService":"app-service-preview"}},"waitForSecondsBeforeScalingDown":100},"status":{"blueGreen":{"activeSelector":"7c797c6d54"},"currentPodHash":"7c797c6d54","observedGeneration":"cf98b99fb"}}`
	replicaSetManifests := `[{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"},"creationTimestamp":"2019-04-21T14:21:54Z","generation":27,"name":"app-deployment-9-7c797c6d54","namespace":"dont-use2"},"spec":{"replicas":1,"template":{"metadata":{"labels":{"rollouts-pod-template-hash":"7c797c6d54"}}}},"status":{"availableReplicas":1,"fullyLabeledReplicas":1,"observedGeneration":27,"readyReplicas":1,"replicas":1}},{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"creationTimestamp":"2019-04-16T11:35:43Z","generation":28,"labels":{"rollouts-pod-template-hash":"66b7dd4665"},"name":"app-deployment-9-66b7dd4665","namespace":"dont-use2"},"spec":{"replicas":1,"template":{"metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"66b7dd4665"}}}},"status":{"observedGeneration":28,"replicas":0}}]`
	serviceManifests := `[{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"}}},{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service-preview","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":""}}}]`
	resp := `{"items":[{"group":"argoproj.io","version":"v1alpha1","kind":"Rollout","namespace":"dont-use2","name":"app-deployment-9","children":[{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-66b7dd4665","resourceVersion":"8545875"},{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54","children":[{"version":"v1","kind":"Pod","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54-k5xvs","info":[{"name":"Status Reason","value":"Running"},{"name":"Containers","value":"1/1"}],"resourceVersion":"8545204"}],"resourceVersion":"8545278"}],"resourceVersion":"8545878"}]}`
	rollout, replicaSet, service, respTree := convertDataForTest(rolloutManifest, replicaSetManifests, serviceManifests, resp)
	return &DataTest{
		rollout:    rollout,
		replicaSet: replicaSet,
		service:    service,
		respTree:   respTree,
	}
}

/*
actual replica count is equal to desired
active not switched
pause switch is missing which means its manual, should be waiting for switch to active
*/
func getData3() *DataTest {
	rolloutManifest := `{"apiVersion":"argoproj.io/v1alpha1","kind":"Rollout","metadata":{"creationTimestamp":"2019-04-16T11:35:42Z","generation":1,"name":"app-deployment-9","namespace":"dont-use2"},"spec":{"strategy":{"blueGreen":{"activeService":"app-service","previewService":"app-service-preview"}},"waitForSecondsBeforeScalingDown":100},"status":{"blueGreen":{"activeSelector":"7c797c6d54"},"currentPodHash":"7c797c6d54","observedGeneration":"cf98b99fb"}}`
	replicaSetManifests := `[{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"},"creationTimestamp":"2019-04-21T14:21:54Z","generation":27,"name":"app-deployment-9-7c797c6d54","namespace":"dont-use2"},"spec":{"replicas":1,"template":{"metadata":{"labels":{"rollouts-pod-template-hash":"7c797c6d54"}}}},"status":{"availableReplicas":1,"fullyLabeledReplicas":1,"observedGeneration":27,"readyReplicas":1,"replicas":1}},{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"creationTimestamp":"2019-04-16T11:35:43Z","generation":28,"labels":{"rollouts-pod-template-hash":"66b7dd4665"},"name":"app-deployment-9-66b7dd4665","namespace":"dont-use2"},"spec":{"replicas":0,"template":{"metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"66b7dd4665"}}}},"status":{"observedGeneration":28,"replicas":0}}]`
	serviceManifests := `[{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":""}}},{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service-preview","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"}}}]`
	resp := `{"items":[{"group":"argoproj.io","version":"v1alpha1","kind":"Rollout","namespace":"dont-use2","name":"app-deployment-9","children":[{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-66b7dd4665","resourceVersion":"8545875"},{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54","children":[{"version":"v1","kind":"Pod","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54-k5xvs","info":[{"name":"Status Reason","value":"Running"},{"name":"Containers","value":"1/1"}],"resourceVersion":"8545204"}],"resourceVersion":"8545278"}],"resourceVersion":"8545878"}]}`
	rollout, replicaSet, service, respTree := convertDataForTest(rolloutManifest, replicaSetManifests, serviceManifests, resp)
	return &DataTest{
		rollout:    rollout,
		replicaSet: replicaSet,
		service:    service,
		respTree:   respTree,
	}
}

/*
actual replica count (Running) not equal to desired
active not switched
pause switch is missing which means its manual, should be degraded
*/
func getData4() *DataTest {
	rolloutManifest := `{"apiVersion":"argoproj.io/v1alpha1","kind":"Rollout","metadata":{"creationTimestamp":"2019-04-16T11:35:42Z","generation":1,"name":"app-deployment-9","namespace":"dont-use2"},"spec":{"strategy":{"blueGreen":{"activeService":"app-service","previewService":"app-service-preview"}},"waitForSecondsBeforeScalingDown":100},"status":{"blueGreen":{"activeSelector":"7c797c6d54"},"currentPodHash":"7c797c6d54","observedGeneration":"cf98b99fb"}}`
	replicaSetManifests := `[{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"},"creationTimestamp":"2019-04-21T14:21:54Z","generation":27,"name":"app-deployment-9-7c797c6d54","namespace":"dont-use2"},"spec":{"replicas":1,"template":{"metadata":{"labels":{"rollouts-pod-template-hash":"7c797c6d54"}}}},"status":{"availableReplicas":1,"fullyLabeledReplicas":1,"observedGeneration":27,"readyReplicas":1,"replicas":1}},{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"creationTimestamp":"2019-04-16T11:35:43Z","generation":28,"labels":{"rollouts-pod-template-hash":"66b7dd4665"},"name":"app-deployment-9-66b7dd4665","namespace":"dont-use2"},"spec":{"replicas":0,"template":{"metadata":{"labels":{"app":"app-deploy","rollouts-pod-template-hash":"66b7dd4665"}}}},"status":{"observedGeneration":28,"replicas":0}}]`
	serviceManifests := `[{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":""}}},{"apiVersion":"v1","kind":"Service","metadata":{"name":"app-service-preview","namespace":"dont-use2"},"spec":{"selector":{"app":"app-deploy","rollouts-pod-template-hash":"7c797c6d54"}}}]`
	resp := `{"items":[{"group":"argoproj.io","version":"v1alpha1","kind":"Rollout","namespace":"dont-use2","name":"app-deployment-9","children":[{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-66b7dd4665","resourceVersion":"8545875"},{"group":"apps","version":"v1","kind":"ReplicaSet","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54","children":[{"version":"v1","kind":"Pod","namespace":"dont-use2","name":"app-deployment-9-7c797c6d54-k5xvs","info":[{"name":"Status Reason","value":"Sdsd"},{"name":"Containers","value":"1/1"}],"resourceVersion":"8545204"}],"resourceVersion":"8545278"}],"resourceVersion":"8545878"}]}`
	rollout, replicaSet, service, respTree := convertDataForTest(rolloutManifest, replicaSetManifests, serviceManifests, resp)
	return &DataTest{
		rollout:    rollout,
		replicaSet: replicaSet,
		service:    service,
		respTree:   respTree,
	}
}

func convertDataForTest(rolloutManifest string, replicaSetManifests string, serviceManifests string, resp string) (rollout map[string]interface{}, replicaSet []map[string]interface{},
	service []map[string]interface{}, respTree *v1alpha1.ApplicationTree) {
	e := json.Unmarshal([]byte(rolloutManifest), &rollout)
	if e != nil {
		fmt.Printf("error %+v\n", e)
	}
	e = json.Unmarshal([]byte(replicaSetManifests), &replicaSet)
	if e != nil {
		fmt.Printf("error %+v\n", e)
	}
	e = json.Unmarshal([]byte(serviceManifests), &service)
	if e != nil {
		fmt.Printf("error %+v\n", e)
	}
	e = json.Unmarshal([]byte(resp), &respTree)
	if e != nil {
		fmt.Printf("error %+v\n", e)
	}
	return
}

type DataTest struct {
	rollout    map[string]interface{}
	replicaSet []map[string]interface{}
	service    []map[string]interface{}
	respTree   *v1alpha1.ApplicationTree
}
