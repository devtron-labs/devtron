package module

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"reflect"
	"testing"
)

func TestModuleServiceImpl_GetAllModuleInfo(t *testing.T) {
	type fields struct {
		logger           *zap.SugaredLogger
		moduleRepository moduleRepo.ModuleRepository
		dbConnection     *pg.DB
	}
	cfg, _ := sql.GetConfig()
	logger, err := utils.NewSugardLogger()
	assert.Nil(t, err)
	dbConnection, err := sql.NewDbConnection(cfg, logger)
	assert.Nil(t, err)
	tests := []struct {
		name    string
		fields  fields
		want    []ModuleInfoDto
		wantErr bool
	}{
		{
			name: "test1_for_data_present_in_db",
			fields: fields{
				logger:           logger,
				dbConnection:     dbConnection,
				moduleRepository: moduleRepo.NewModuleRepositoryImpl(dbConnection),
			},
			want:    getModuleDtoResponse1(),
			wantErr: false,
		},
		{
			name: "test2_for_empty_data_in_db",
			fields: fields{
				logger:           logger,
				dbConnection:     dbConnection,
				moduleRepository: moduleRepo.NewModuleRepositoryImpl(dbConnection),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ModuleServiceImpl{
				logger:           tt.fields.logger,
				moduleRepository: tt.fields.moduleRepository,
			}
			got, err := impl.GetAllModuleInfo()
			if (err != nil) == !tt.wantErr {
				got = nil
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetAllModuleInfo() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				d, _ := json.Marshal(got)
				fmt.Printf("%s\n", d)
				t.Errorf("GetAllModuleInfo() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func getModuleDtoResponse1() []ModuleInfoDto {
	return []ModuleInfoDto{
		{
			Name:   "cicd",
			Status: "installed",
		},
		{
			Name:   "argo-cd",
			Status: "installed",
		},
		{
			Name:   "security.clair",
			Status: "installed",
		},
		{
			Name:   "monitoring.grafana",
			Status: "installed",
		},
		{
			Name:   "notifier",
			Status: "installed",
		},
	}
}
