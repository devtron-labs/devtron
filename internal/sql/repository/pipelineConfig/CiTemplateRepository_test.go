/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package pipelineConfig

import (
	"testing"

	"github.com/go-pg/pg"
)

func TestCiTemplateRepositoryImpl_FindByAppId(t *testing.T) {
	type fields struct {
		dbConnection *pg.DB
	}
	type args struct {
		appId int
	}
	tests := []struct {
		name           string
		dbConnection   *pg.DB
		appId          int
		wantCiTemplate *CiTemplate
		wantErr        bool
	}{
		//{name: "abc", appId: 20, dbConnection: models.NewDbConnection(nil, nil), wantErr: false, wantCiTemplate: &CiTemplate{Id: 1}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := CiTemplateRepositoryImpl{
				dbConnection: tt.dbConnection,
			}
			gotCiTemplate, err := impl.FindByAppId(tt.appId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CiTemplateRepositoryImpl.FindByPipelineId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotCiTemplate.Id != tt.wantCiTemplate.Id {
				t.Errorf("CiTemplateRepositoryImpl.FindByPipelineId() = %v, want %v", gotCiTemplate, tt.wantCiTemplate)
			}
		})
	}
}
