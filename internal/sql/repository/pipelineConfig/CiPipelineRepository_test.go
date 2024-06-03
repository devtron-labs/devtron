/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package pipelineConfig

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"testing"
)

var ciPipelineRepo CiPipelineRepositoryImpl

func setup() {
	cfg, _ := sql.GetConfig()
	con, _ := sql.NewDbConnection(cfg, nil)
	ciPipelineRepo = CiPipelineRepositoryImpl{
		dbConnection: con,
		logger:       nil,
	}
}

func TestCiPipelineRepositoryImpl_FindByAppId(t *testing.T) {
	setup()

	tests := []struct {
		name    string
		appId   int
		wantErr bool
	}{
		{
			name: "test1", appId: 31, wantErr: false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ciPipelineRepo
			gotPipelines, err := impl.FindByAppId(tt.appId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CiPipelineRepositoryImpl.FindByPipelineId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(gotPipelines)
		})
	}
}

func TestCiPipelineRepositoryImpl_PipelineExistsByName(t *testing.T) {
	setup()

	tests := []struct {
		name      string
		names     []string
		wantFound []string
		wantErr   bool
	}{
		{
			name:      "abc",
			names:     []string{"test-nc-ci-qa", "test-fj-ci-qa3"},
			wantFound: []string{"test-nc-ci-qa", "test-fj-ci-qa3"},
			wantErr:   false,
		},
		{
			name:      "ab",
			names:     []string{"test-nc-ci-qa", "test-nc-ci-qacd"},
			wantFound: []string{"test-fc-ci-qa"},
			wantErr:   false,
		},
		{
			name:      "abwedf",
			names:     []string{"test-nc-ci-qaewde", "test-nc-ci-qacdwed"},
			wantFound: nil,
			wantErr:   false,
		},
		{
			name:      "null test",
			names:     nil,
			wantFound: nil,
			wantErr:   false,
		},
		{
			name:      "empty test",
			names:     []string{},
			wantFound: nil,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ciPipelineRepo
			gotFound, err := impl.PipelineExistsByName(tt.names)
			if (err != nil) != tt.wantErr {
				t.Errorf("CiPipelineRepositoryImpl.PipelineExistsByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(gotFound)
		})
	}
}
