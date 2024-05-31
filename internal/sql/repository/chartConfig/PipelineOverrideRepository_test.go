/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package chartConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"reflect"
	"testing"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
)

func getPor() PipelineOverrideRepository {
	//return NewPipelineOverrideRepository(models.GetDbTransaction())
	return nil
}

func TestPipelineOverrideRepositoryImpl_Save(t *testing.T) {
	envConfigOverride, _ := getEcr().Get(5)
	po := &PipelineOverride{
		EnvConfigOverrideId:    envConfigOverride.Id,
		Status:                 models.CHARTSTATUS_NEW,
		PipelineMergedValues:   "{}",
		PipelineOverrideValues: "{}",
		RequestIdentifier:      "request-1",
		AuditLog:               sql.AuditLog{CreatedBy: 1, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: 1},
	}
	err := getPor().Save(po)
	assert.NoError(t, err)

}

func TestPipelineOverrideRepositoryImpl_UpdateStatus(t *testing.T) {
	count, err := getPor().UpdateStatusByRequestIdentifier("request-1", models.CHARTSTATUS_UNKNOWN)
	assert.NoError(t, err)
	assert.Equal(t, count, 1)
}

func TestPipelineOverrideRepositoryImpl_GetLatestConfigByEnvironmentConfigOverrideId(t *testing.T) {
	type fields struct {
		dbConnection *pg.DB
	}
	type args struct {
		envConfigOverrideId int
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantPipelineOverride *PipelineOverride
		wantErr              bool
	}{{
		name: "err_test",
		//	fields:               fields{dbConnection: models.GetDbTransaction()},
		args:                 args{envConfigOverrideId: 2222},
		wantErr:              true,
		wantPipelineOverride: nil,
	},
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := PipelineOverrideRepositoryImpl{
				dbConnection: tt.fields.dbConnection,
			}
			gotPipelineOverride, err := impl.GetLatestConfigByEnvironmentConfigOverrideId(tt.args.envConfigOverrideId)
			if (err != nil) != tt.wantErr {
				t.Errorf("PipelineOverrideRepositoryImpl.GetLatestConfigByEnvironmentConfigOverrideId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPipelineOverride, tt.wantPipelineOverride) {
				t.Errorf("PipelineOverrideRepositoryImpl.GetLatestConfigByEnvironmentConfigOverrideId() = %v, want %v", gotPipelineOverride, tt.wantPipelineOverride)
			}
		})
	}
}
