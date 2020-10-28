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

package util

import (
	"testing"
)

var client *K8sUtil
var clusterConfig *ClusterConfig

func init() {
	client = NewK8sUtil(NewSugardLogger())
	clusterConfig = &ClusterConfig{
		Host:        "",
		BearerToken: "",
	}
}

func TestK8sUtil_checkIfNsExists(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "test-kube-system",
			namespace:  "kube-system",
			wantErr:    false,
			wantExists: true,
		}, {
			name:       "test-randum",
			namespace:  "test-rand-laknd-kwejdwiu",
			wantExists: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := client
			k8s, _ := impl.getClient(clusterConfig)
			gotExists, err := impl.checkIfNsExists(tt.namespace, k8s)
			if (err != nil) != tt.wantErr {
				t.Errorf("K8sUtil.checkIfNsExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("K8sUtil.checkIfNsExists() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func TestK8sUtil_CreateNsIfNotExists(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
	}{
		{
			name:      "create test",
			namespace: "createtestns",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := client
			if err := impl.CreateNsIfNotExists(tt.namespace, clusterConfig); (err != nil) != tt.wantErr {
				t.Errorf("K8sUtil.CreateNsIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			k8s, _ := impl.getClient(clusterConfig)
			if err := impl.deleteNs(tt.namespace, k8s); (err != nil) != tt.wantErr {
				t.Errorf("K8sUtil.deleteNs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
