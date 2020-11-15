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

package chartConfig

import (
	"reflect"
	"testing"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
)

func getPor() PipelineOverrideRepository {
	//return NewPipelineOverrideRepository(models.GetDbConnection())
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
		AuditLog:               models.AuditLog{CreatedBy: 1, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: 1},
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
		//	fields:               fields{dbConnection: models.GetDbConnection()},
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
