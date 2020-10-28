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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	v1 "github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"reflect"
	"testing"
)

var toString func(string) *string = func(str string) *string { return &str }

func Test_deleteDataKeys(t *testing.T) {
	type args struct {
		dataType   string
		holder     *v1.DataHolder
		configData *pipeline.ConfigDataRequest
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		expectedData map[string]interface{}
	}{
		{
			name: "configmap test",
			args: args{
				dataType: v1.ConfigMap,
				holder: &v1.DataHolder{
					ApiVersion: "",
					Data:       map[string]interface{}{"key1": "value1", "key2": "value2"},
					Destination: &v1.ResourcePath{
						App:         nil,
						Configmap:   toString("test1"),
						Environment: nil,
						Pipeline:    nil,
						Secret:      nil,
						Uid:         nil,
						Workflow:    nil,
					},
					External:     nil,
					ExternalType: nil,
					Global:       nil,
					MountPath:    nil,
					Operation:    "",
					Source:       nil,
					Type:         nil,
				},
				configData: &pipeline.ConfigDataRequest{
					Id:            0,
					AppId:         0,
					EnvironmentId: 0,
					ConfigData: []pipeline.ConfigData{pipeline.ConfigData{
						Name:               "test1",
						Type:               "",
						External:           false,
						MountPath:          "",
						Data:               []byte(`{"key1":"v1", "key3":"v3","key2":"v2", "key4": "v4"}`),
						DefaultData:        nil,
						DefaultMountPath:   "",
						Global:             false,
						ExternalSecretType: "",
					}},
					UserId: 0,
				},
			},
			wantErr:      false,
			expectedData: map[string]interface{}{"key3": "v3", "key4": "v4"},
		},
		{
			name: "secret test",
			args: args{
				dataType: v1.Secret,
				holder: &v1.DataHolder{
					ApiVersion: "",
					Data:       map[string]interface{}{"key1": "value1", "key2": "value2"},
					Destination: &v1.ResourcePath{
						App:         nil,
						Configmap:   nil,
						Environment: nil,
						Pipeline:    nil,
						Secret:      toString("test1"),
						Uid:         nil,
						Workflow:    nil,
					},
					External:     nil,
					ExternalType: nil,
					Global:       nil,
					MountPath:    nil,
					Operation:    "",
					Source:       nil,
					Type:         nil,
				},
				configData: &pipeline.ConfigDataRequest{
					Id:            0,
					AppId:         0,
					EnvironmentId: 0,
					ConfigData: []pipeline.ConfigData{pipeline.ConfigData{
						Name:               "test1",
						Type:               "",
						External:           false,
						MountPath:          "",
						Data:               []byte(`{"key1":"v1", "key3":"v3","key2":"v2", "key4": "v4"}`),
						DefaultData:        nil,
						DefaultMountPath:   "",
						Global:             false,
						ExternalSecretType: "",
					}},
					UserId: 0,
				},
			},
			wantErr:      false,
			expectedData: map[string]interface{}{"key3": "v3", "key4": "v4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteDataKeys(tt.args.dataType, tt.args.holder, tt.args.configData); (err != nil) != tt.wantErr {
				t.Errorf("deleteDataKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, cf := range tt.args.configData.ConfigData {
				var name string
				if tt.args.dataType == v1.ConfigMap {
					name = *tt.args.holder.Destination.Configmap
				} else if tt.args.dataType == v1.Secret {
					name = *tt.args.holder.Destination.Secret
				}
				if cf.Name == name {
					m := make(map[string]interface{}, 0)
					json.Unmarshal(cf.Data, &m)
					same := reflect.DeepEqual(m, tt.expectedData)
					if !same {
						t.Errorf("expected %+v, found %+v\n", tt.expectedData, m)
					}
				}
			}
		})
	}
}

func Test_updateKeys(t *testing.T) {
	type args struct {
		dataType   string
		holder     *v1.DataHolder
		configData *pipeline.ConfigDataRequest
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		expectedData map[string]interface{}
	}{
		{
			name: "configmap test",
			args: args{
				dataType: v1.ConfigMap,
				holder: &v1.DataHolder{
					ApiVersion: "",
					Data:       map[string]interface{}{"key1": "value1", "key2": "value2", "key5": "v5", "key6": ""},
					Destination: &v1.ResourcePath{
						App:         nil,
						Configmap:   toString("test1"),
						Environment: nil,
						Pipeline:    nil,
						Secret:      nil,
						Uid:         nil,
						Workflow:    nil,
					},
					External:     nil,
					ExternalType: nil,
					Global:       nil,
					MountPath:    nil,
					Operation:    "",
					Source:       nil,
					Type:         nil,
				},
				configData: &pipeline.ConfigDataRequest{
					Id:            0,
					AppId:         0,
					EnvironmentId: 0,
					ConfigData: []pipeline.ConfigData{pipeline.ConfigData{
						Name:               "test1",
						Type:               "",
						External:           false,
						MountPath:          "",
						Data:               []byte(`{"key1":"v1", "key3":"v3","key2":"v2", "key4": "v4", "key6":"v6"}`),
						DefaultData:        nil,
						DefaultMountPath:   "",
						Global:             false,
						ExternalSecretType: "",
					},
						pipeline.ConfigData{
							Name:               "test2",
							Type:               "",
							External:           false,
							MountPath:          "",
							Data:               []byte(`{"key1":"v1", "key3":"v3","key2":"v2", "key4": "v4", "key7":"v6"}`),
							DefaultData:        nil,
							DefaultMountPath:   "",
							Global:             false,
							ExternalSecretType: "",
						}},
					UserId: 0,
				},
			},
			wantErr:      false,
			expectedData: map[string]interface{}{"key1": "value1", "key2": "value2", "key5": "v5", "key3": "v3", "key4": "v4"},
		},
		{
			name: "secret test",
			args: args{
				dataType: v1.Secret,
				holder: &v1.DataHolder{
					ApiVersion: "",
					Data:       map[string]interface{}{"key1": "value1", "key2": "value2", "key5": "v5"},
					Destination: &v1.ResourcePath{
						App:         nil,
						Configmap:   nil,
						Environment: nil,
						Pipeline:    nil,
						Secret:      toString("test1"),
						Uid:         nil,
						Workflow:    nil,
					},
					External:     nil,
					ExternalType: nil,
					Global:       nil,
					MountPath:    nil,
					Operation:    "",
					Source:       nil,
					Type:         nil,
				},
				configData: &pipeline.ConfigDataRequest{
					Id:            0,
					AppId:         0,
					EnvironmentId: 0,
					ConfigData: []pipeline.ConfigData{pipeline.ConfigData{
						Name:               "test1",
						Type:               "",
						External:           false,
						MountPath:          "",
						Data:               []byte(`{"key1":"v1", "key3":"v3","key2":"v2", "key4": "v4"}`),
						DefaultData:        nil,
						DefaultMountPath:   "",
						Global:             false,
						ExternalSecretType: "",
					}},
					UserId: 0,
				},
			},
			wantErr:      false,
			expectedData: map[string]interface{}{"key1": "value1", "key2": "value2", "key5": "v5", "key3": "v3", "key4": "v4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateKeys(tt.args.dataType, tt.args.holder, tt.args.configData); (err != nil) != tt.wantErr {
				t.Errorf("updateKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, cf := range tt.args.configData.ConfigData {
				var name string
				if tt.args.dataType == v1.ConfigMap {
					name = *tt.args.holder.Destination.Configmap
				} else if tt.args.dataType == v1.Secret {
					name = *tt.args.holder.Destination.Secret
				}
				if cf.Name == name {
					m := make(map[string]interface{}, 0)
					json.Unmarshal(cf.Data, &m)
					same := reflect.DeepEqual(m, tt.expectedData)
					if !same {
						t.Errorf("expected %+v, found %+v\n", tt.expectedData, m)
					}
				}
			}
		})
	}
}

func TestDataHolderActionImpl_Execute(t *testing.T) {

	secretConfig := `apiVersion: github.com/devtron-labs/v1
operation: create
destination:
  app: app1
  configMap: configmap1
data:
  fName: prashant
  lName: Ghildiyal`
	dh := v1.DataHolder{}
	err := yaml.Unmarshal([]byte(secretConfig), &dh)
	if err != nil {
		panic(err)
	}
	type fields struct {
		logger           *zap.SugaredLogger
		appRepo          pipelineConfig.AppRepository
		configMapService pipeline.ConfigMapService
		envService       cluster.EnvironmentService
	}
	type args struct {
		holder   *v1.DataHolder
		props    v1.InheritedProps
		dataType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Secret test",
			fields: fields{
				logger:           &LoggerMock,
				appRepo:          AppRepositoryMock{},
				configMapService: ConfigMapServiceMock{},
				envService:       EnvironmentServiceMock{},
			},
			args: args{
				holder:   &dh,
				props:    v1.InheritedProps{},
				dataType: v1.ConfigMap,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := DataHolderActionImpl{
				logger:           tt.fields.logger,
				appRepo:          tt.fields.appRepo,
				configMapService: tt.fields.configMapService,
				envService:       tt.fields.envService,
			}
			if err := impl.Execute(tt.args.holder, tt.args.props, tt.args.dataType); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
