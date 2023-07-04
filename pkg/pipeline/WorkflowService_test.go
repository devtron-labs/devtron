package pipeline

import (
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	"log"
	"testing"
)

func Test_getConfigMapsAndSecrets(t *testing.T) {
	t.SkipNow()

	type args struct {
		impl              *WorkflowServiceImpl
		workflowRequest   *WorkflowRequest
		existingConfigMap *bean3.ConfigMapJson
		existingSecrets   *bean3.ConfigSecretJson
	}
	workflowRequest := &WorkflowRequest{
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
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
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
				impl:              &WorkflowServiceImpl{Logger: logger},
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
				impl:              &WorkflowServiceImpl{Logger: logger},
				workflowRequest:   workflowRequest,
				existingConfigMap: existingConfigMap,
				existingSecrets:   &bean3.ConfigSecretJson{},
			},
			want:    bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{{Name: "job-map-123-ci", Data: []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"), External: false}}},
			want1:   bean3.ConfigSecretJson{},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getConfigMapsAndSecrets(tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
			if !tt.wantErr(t, err, fmt.Sprintf("getConfigMapsAndSecrets(%v, %v, %v, %v)", tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getConfigMapsAndSecrets(%v, %v, %v, %v)", tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
			assert.Equalf(t, tt.want1, got1, "getConfigMapsAndSecrets(%v, %v, %v, %v)", tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
		})
	}
}
func Test_getConfigMapsAndSecrets1(t *testing.T) {
	t.SkipNow()

	type args struct {
		impl              *WorkflowServiceImpl
		workflowRequest   *WorkflowRequest
		existingConfigMap *bean3.ConfigMapJson
		existingSecrets   *bean3.ConfigSecretJson
	}
	workflowRequest := &WorkflowRequest{
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
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatalf("error in logger initialization %s,%s", "err", err)
	}
	tests := []struct {
		name    string
		args    args
		want    bean3.ConfigMapJson
		want1   bean3.ConfigSecretJson
		wantErr assert.ErrorAssertionFunc
	}{

		{
			name: "empty  existingConfigMap and non empty existingSecrets",
			args: args{
				impl:              &WorkflowServiceImpl{Logger: logger},
				workflowRequest:   workflowRequest,
				existingConfigMap: existingConfigMap,
				existingSecrets:   existingSecrets,
			},
			want:    bean3.ConfigMapJson{Maps: []bean3.ConfigSecretMap{{Name: "job-map-123-ci", Data: []byte("{\"abcd\": \"aditya-cm-1-job-test-cm1\"}"), External: false}}},
			want1:   bean3.ConfigSecretJson{Secrets: []*bean3.ConfigSecretMap{{Name: "job-secret-123-ci", Data: []byte("{\"abcd\": \"XCJhZGl0eWEtY20tMS1qb2ItdGVzdC1jbTFcIn0i\"}"), External: false}}},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getConfigMapsAndSecrets(tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
			if !tt.wantErr(t, err, fmt.Sprintf("getConfigMapsAndSecrets(%v, %v, %v, %v)", tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getConfigMapsAndSecrets(%v, %v, %v, %v)", tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
			assert.Equalf(t, tt.want1.Secrets[0].Name, got1.Secrets[0].Name, "getConfigMapsAndSecrets(%v, %v, %v, %v)", tt.args.impl, tt.args.workflowRequest, tt.args.existingConfigMap, tt.args.existingSecrets)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCiTemplateWithConfigMapsAndSecrets(tt.args.configMaps, tt.args.secrets, tt.args.ciTemplate, tt.args.existingConfigMap, tt.args.existingSecrets)
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
			tt.wantErr(t, processConfigMapsAndSecrets(tt.args.impl, tt.args.configMaps, tt.args.secrets, tt.args.entryPoint, tt.args.steps, tt.args.volumes, tt.args.templates), fmt.Sprintf("processConfigMapsAndSecrets(%v, %v, %v, %v, %v, %v, %v)", tt.args.impl, tt.args.configMaps, tt.args.secrets, tt.args.entryPoint, tt.args.steps, tt.args.volumes, tt.args.templates))
		})
	}
}
