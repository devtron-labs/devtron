/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipeline

import (
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	"log"
	"testing"
)

func Test_getConfigMapsAndSecrets(t *testing.T) {
	t.SkipNow()
	type args struct {
		workflowRequest   *types.WorkflowRequest
		existingConfigMap *bean3.ConfigMapJson
		existingSecrets   *bean3.ConfigSecretJson
	}
	workflowRequest := &types.WorkflowRequest{
		WorkflowId: 123,
	}
	existingConfigMap := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
			},
		},
	}
	existingSecrets := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
			},
		},
	}
	existingSecrets1 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: true,
			},
		},
	}
	tests := []struct {
		name    string
		args    args
		want    bean3.ConfigMapJson
		want1   bean3.ConfigSecretJson
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Empty existingConfigMap and existingSecrets",
			args: args{
				workflowRequest:   workflowRequest,
				existingConfigMap: &bean3.ConfigMapJson{},
				existingSecrets:   &bean3.ConfigSecretJson{},
			},
			want:    bean3.ConfigMapJson{},
			want1:   bean3.ConfigSecretJson{},
			wantErr: assert.NoError,
		},
		{
			name: "non empty  existingConfigMap and empty existingSecrets",
			args: args{
				workflowRequest:   workflowRequest,
				existingConfigMap: existingConfigMap,
				existingSecrets:   &bean3.ConfigSecretJson{},
			},
			want:    bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{{Name: "job-map-123-ci", Data: []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"), External: false}}},
			want1:   bean3.ConfigSecretJson{},
			wantErr: assert.NoError,
		},
		{
			name: "non empty  existingConfigMap and non empty existingSecrets",
			args: args{
				workflowRequest:   workflowRequest,
				existingConfigMap: existingConfigMap,
				existingSecrets:   existingSecrets,
			},
			want:    bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{{Name: "job-map-123-ci", Data: []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"), External: false}}},
			want1:   bean3.ConfigSecretJson{Secrets: []*bean3.ConfigSecretMap{{Name: "job-secret-123-ci", Data: []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"), External: false}}},
			wantErr: assert.NoError,
		},
		{
			name: "non empty  existingConfigMap and non empty existingSecrets with external ",
			args: args{
				workflowRequest:   workflowRequest,
				existingConfigMap: existingConfigMap,
				existingSecrets:   existingSecrets1,
			},
			want:    bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{{Name: "job-map-123-ci", Data: []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"), External: false}}},
			want1:   bean3.ConfigSecretJson{},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getConfigMapsAndSecrets(tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
			if !tt.wantErr(t, err, fmt.Sprintf("getConfigMapsAndSecrets(%v, %v, %v)", tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getConfigMapsAndSecrets(%v, %v, %v)", tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
			assert.Equalf(t, tt.want1, got1, "getConfigMapsAndSecrets(%v, %v, %v)", tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
			if tt.args.existingSecrets == existingSecrets {
				assert.Equalf(t, tt.want1.Secrets[0].Name, got1.Secrets[0].Name, "getConfigMapsAndSecrets(%v, %v, %v)", tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)

			}
		})
	}
}

func Test_getCiTemplateWithConfigMapsAndSecrets(t *testing.T) {
	t.SkipNow()
	type args struct {
		configMaps        *bean3.ConfigMapJson
		secrets           *bean3.ConfigSecretJson
		ciTemplate        v1alpha1.Template
		existingConfigMap *bean3.ConfigMapJson
		existingSecrets   *bean3.ConfigSecretJson
	}
	configMaps := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
			},
		},
	}
	secrets := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
			},
		},
	}

	ciTemplate := v1alpha1.Template{
		Name: CI_WORKFLOW_NAME,
		Container: &v12.Container{
			EnvFrom: []v12.EnvFromSource{},
		},
	}
	ciTemplate1 := v1alpha1.Template{
		Name: CI_WORKFLOW_NAME,
		Container: &v12.Container{
			VolumeMounts: []v12.VolumeMount{},
		},
	}

	existingConfigMap := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
			},
		},
	}
	existingSecrets := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\":\"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
			},
		},
	}
	configMaps1 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
				Type:     "environment",
			},
		},
	}
	secrets1 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
				Type:     "environment",
			},
		},
	}
	existingConfigMap1 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
				Type:     "environment",
			},
		},
	}
	existingSecrets1 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\":\"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
				Type:     "environment",
			},
		},
	}
	configMaps2 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
				Type:     "volume",
			},
		},
	}
	secrets2 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
				Type:     "volume",
			},
		},
	}
	existingConfigMap2 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
				Type:     "volume",
			},
		},
	}
	existingSecrets2 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\":\"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
				Type:     "volume",
			},
		},
	}
	configMaps3 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: true,
				Type:     "volume",
			},
		},
	}
	secrets3 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: true,
				Type:     "volume",
			},
		},
	}
	existingConfigMap3 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: true,
				Type:     "volume",
			},
		},
	}
	existingSecrets3 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\":\"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: true,
				Type:     "volume",
			},
		},
	}
	configMaps4 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: true,
				Type:     "environment",
			},
		},
	}
	secrets4 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: true,
				Type:     "environment",
			},
		},
	}
	existingConfigMap4 := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: true,
				Type:     "environment",
			},
		},
	}
	existingSecrets4 := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\":\"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: true,
				Type:     "environment",
			},
		},
	}
	tests := []struct {
		name    string
		args    args
		want    v1alpha1.Template
		wantErr assert.ErrorAssertionFunc
	}{

		{name: "test1",
			args: args{
				configMaps:        configMaps,
				secrets:           secrets,
				ciTemplate:        ciTemplate,
				existingSecrets:   existingSecrets,
				existingConfigMap: existingConfigMap,
			},
			want:    ciTemplate,
			wantErr: assert.NoError,
		},
		{name: "test2",
			args: args{
				configMaps:        configMaps1,
				secrets:           secrets1,
				ciTemplate:        ciTemplate,
				existingSecrets:   existingSecrets1,
				existingConfigMap: existingConfigMap1,
			},
			want:    ciTemplate,
			wantErr: assert.NoError,
		},
		{name: "test3",
			args: args{
				configMaps:        configMaps2,
				secrets:           secrets2,
				ciTemplate:        ciTemplate1,
				existingSecrets:   existingSecrets2,
				existingConfigMap: existingConfigMap2,
			},
			want:    ciTemplate1,
			wantErr: assert.NoError,
		},
		{name: "test4",
			args: args{
				configMaps:        configMaps3,
				secrets:           secrets3,
				ciTemplate:        ciTemplate1,
				existingSecrets:   existingSecrets3,
				existingConfigMap: existingConfigMap3,
			},
			want:    ciTemplate1,
			wantErr: assert.NoError,
		},
		{name: "test5",
			args: args{
				configMaps:        configMaps4,
				secrets:           secrets4,
				ciTemplate:        ciTemplate,
				existingSecrets:   existingSecrets4,
				existingConfigMap: existingConfigMap4,
			},
			want:    ciTemplate,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCiTemplateWithConfigMapsAndSecrets(tt.args.configMaps, tt.args.secrets, tt.args.ciTemplate)
			if !tt.wantErr(t, err, fmt.Sprintf("getCiTemplateWithConfigMapsAndSecrets(%v, %v, %v, %v, %v)", tt.args.configMaps, tt.args.secrets, tt.args.ciTemplate, tt.args.existingConfigMap, tt.args.existingSecrets)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getCiTemplateWithConfigMapsAndSecrets(%v, %v, %v, %v, %v)", tt.args.configMaps, tt.args.secrets, tt.args.ciTemplate, tt.args.existingConfigMap, tt.args.existingSecrets)
		})
	}
}

func Test_processConfigMapsAndSecrets(t *testing.T) {
	t.SkipNow()
	type args struct {
		impl       *WorkflowServiceImpl
		configMaps *bean3.ConfigMapJson
		secrets    *bean3.ConfigSecretJson
		entryPoint *string
		steps      *[]v1alpha1.ParallelSteps
		volumes    *[]v12.Volume
		templates  *[]v1alpha1.Template
	}
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
	}
	impl := &WorkflowServiceImpl{Logger: logger}
	configMaps := &bean3.ConfigMapJson{
		Enabled: true,
		Maps: []bean3.ConfigSecretMap{
			{Name: "job-map",
				Data:     []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"),
				External: false,
			},
		},
	}
	secrets := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"),
				External: false,
			},
		},
	}
	entryPoint := "\"ci-stages-with-env\""
	steps := &[]v1alpha1.ParallelSteps{}
	volumes := &[]v12.Volume{}
	templates := &[]v1alpha1.Template{}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "TEST 1",
			args: args{
				impl:       impl,
				configMaps: configMaps,
				secrets:    secrets,
				entryPoint: &entryPoint,
				steps:      steps,
				volumes:    volumes,
				templates:  templates,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, processConfigMapsAndSecrets(tt.args.impl, tt.args.configMaps, tt.args.secrets, tt.args.entryPoint, tt.args.steps, tt.args.templates), fmt.Sprintf("processConfigMapsAndSecrets(%v, %v, %v, %v, %v, %v, %v)", tt.args.impl, tt.args.configMaps, tt.args.secrets, tt.args.entryPoint, tt.args.steps, tt.args.volumes, tt.args.templates))
		})
	}
}

func Test_processConfigMapsAndSecrets1(t *testing.T) {
	t.SkipNow()

	type args struct {
		impl       *WorkflowServiceImpl
		configMaps *bean3.ConfigMapJson
		secrets    *bean3.ConfigSecretJson
		entryPoint *string
		steps      *[]v1alpha1.ParallelSteps
		volumes    *[]v12.Volume
		templates  *[]v1alpha1.Template
	}

	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
	}
	impl := &WorkflowServiceImpl{Logger: logger}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "NoError",
			args: args{
				impl:       impl,
				configMaps: &bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{}},
				secrets:    &bean3.ConfigSecretJson{Secrets: []*bean3.ConfigSecretMap{}},
				entryPoint: new(string),
				steps:      &[]v1alpha1.ParallelSteps{},
				volumes:    &[]v12.Volume{},
				templates:  &[]v1alpha1.Template{},
			},
			wantErr: false,
		},
		{
			name: "ProcessConfigMapError",
			args: args{
				impl:       impl,
				configMaps: &bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{{Name: "job-map"}}},
				secrets:    &bean3.ConfigSecretJson{Secrets: []*bean3.ConfigSecretMap{}},
				entryPoint: new(string),
				steps:      &[]v1alpha1.ParallelSteps{},
				volumes:    &[]v12.Volume{},
				templates:  &[]v1alpha1.Template{},
			},
			wantErr: true,
		},
		{
			name: "ProcessSecretsError",
			args: args{
				impl:       impl,
				configMaps: &bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{}},
				secrets:    &bean3.ConfigSecretJson{Secrets: []*bean3.ConfigSecretMap{{Name: "job-secret"}}}, // Simulate an error scenario
				entryPoint: new(string),
				steps:      &[]v1alpha1.ParallelSteps{},
				volumes:    &[]v12.Volume{},
				templates:  &[]v1alpha1.Template{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processConfigMapsAndSecrets(tt.args.impl, tt.args.configMaps, tt.args.secrets, tt.args.entryPoint, tt.args.steps, tt.args.templates)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
