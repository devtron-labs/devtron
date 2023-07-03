package pipeline

import (
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"log"

	"github.com/stretchr/testify/assert"
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
	existingSecrets := &bean3.ConfigSecretJson{
		Enabled: true,
		Secrets: []*bean3.ConfigSecretMap{
			{Name: "job-secret",
				Data:     []byte("{\"abcd\": \"aditya-cs-1-job-test-cs1\"}"),
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
		{
			name: "empty  existingConfigMap and non empty existingSecrets",
			args: args{
				impl:              &WorkflowServiceImpl{Logger: logger},
				workflowRequest:   workflowRequest,
				existingConfigMap: &bean3.ConfigMapJson{},
				existingSecrets:   existingSecrets,
			},
			want:    bean3.ConfigMapJson{},
			want1:   bean3.ConfigSecretJson{Secrets: []*bean3.ConfigSecretMap{{Name: "job-secret", Data: []byte("{\"abcd\": \"aditya-cs-1-job-test-cs1\"}"), External: false}}},
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
