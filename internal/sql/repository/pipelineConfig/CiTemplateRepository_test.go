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
