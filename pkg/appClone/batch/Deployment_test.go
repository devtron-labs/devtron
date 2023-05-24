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

package batch

import (
	"encoding/json"
	"fmt"
	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/bean"
	"reflect"
	"strings"
	"testing"
)

func Test_transformStages(t *testing.T) {

	toInt := func(i int32) *int32 { return &i }

	type args struct {
		deploymentStages []v1.Stage
		preOrPost        string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "stages",
			args: args{deploymentStages: []v1.Stage{{
				Name:           "script 1",
				Operation:      "",
				OutputLocation: toString("/tmp/test"),
				Position:       toInt(1),
				Script: toString(`echo date
echo "found"`),
			}},
				preOrPost: "post"},
			want: `cdPipelineConf:
  afterStages:
  - name: script 1
    outputLocation: /tmp/test
    script: |-
      echo date
      echo "found"
version: 0.0.1`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformStages(tt.args.deploymentStages, tt.args.preOrPost)
			fmt.Println(string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("transformStages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if strings.ReplaceAll(string(got), "\n", "") != strings.ReplaceAll(tt.want, "\n", "") {
				t.Errorf("transformStages() got = \n%s, want = \n%s", strings.ReplaceAll(string(got), "\n", ""), strings.ReplaceAll(tt.want, "\n", ""))
			}
		})
	}
}

func Test_transformStrategy(t *testing.T) {
	type args struct {
		deployment *v1.Deployment
	}
	tests := []struct {
		name    string
		args    args
		want    []bean.Strategy
		wantErr bool
	}{
		{
			name: "strategy test",
			args: args{deployment: &v1.Deployment{
				Strategy: v1.DeploymentStrategy{
					BlueGreen: &v1.BlueGreenStrategy{
						AutoPromotionEnabled:  false,
						AutoPromotionSeconds:  30,
						PreviewReplicaCount:   10,
						ScaleDownDelaySeconds: 20,
					},
					Canary:   nil,
					Default:  "BLUE-GREEN",
					Recreate: &v1.RecreateStrategy{},
					Rolling:  nil,
				},
			}},
			want: []bean.Strategy{bean.Strategy{
				DeploymentTemplate: "BLUE-GREEN",
				Config:             []byte(`{"autoPromotionEnabled":false,"autoPromotionSeconds":30,"previewReplicaCount":10,"scaleDownDelaySeconds":20}`),
				Default:            true,
			},
				bean.Strategy{
					DeploymentTemplate: "RECREATE",
					Config:             []byte("{}"),
					Default:            false,
				}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformStrategy(tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("transformStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				d, _ := json.Marshal(got)
				fmt.Printf("%s\n", d)
				t.Errorf("transformStrategy() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
