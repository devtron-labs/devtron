package repository

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"reflect"
	"testing"
	"time"
)

func getDbConnAndLoggerService(t *testing.T) (*zap.SugaredLogger, *pg.DB) {
	cfg, _ := sql.GetConfig()
	logger, err := utils.NewSugardLogger()
	assert.Nil(t, err)
	dbConnection, err := sql.NewDbConnection(cfg, logger)
	assert.Nil(t, err)

	return logger, dbConnection
}

func getVariableEntityMappingRepositoryImpl(t *testing.T) *VariableEntityMappingRepositoryImpl {
	logger, dbConnection := getDbConnAndLoggerService(t)
	return NewVariableEntityMappingRepository(logger, dbConnection)
}

func TestVariableEntityMappingRepositoryImpl_DeleteAllVariablesForEntities(t *testing.T) {

	type args struct {
		entities []Entity
		userId   int32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getVariableEntityMappingRepositoryImpl(t)
			if err := impl.DeleteAllVariablesForEntities(tt.args.entities, tt.args.userId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllVariablesForEntities() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVariableEntityMappingRepositoryImpl_DeleteVariablesForEntity(t *testing.T) {
	type fields struct {
		logger              *zap.SugaredLogger
		dbConnection        *pg.DB
		TransactionUtilImpl *sql.TransactionUtilImpl
	}
	type args struct {
		tx            *pg.Tx
		variableNames []string
		entity        Entity
		userId        int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &VariableEntityMappingRepositoryImpl{
				logger:              tt.fields.logger,
				dbConnection:        tt.fields.dbConnection,
				TransactionUtilImpl: tt.fields.TransactionUtilImpl,
			}
			if err := impl.DeleteVariablesForEntity(tt.args.tx, tt.args.variableNames, tt.args.entity, tt.args.userId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteVariablesForEntity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVariableEntityMappingRepositoryImpl_GetVariablesForEntities(t *testing.T) {

	type args struct {
		entities []Entity
	}

	tests := []struct {
		name    string
		args    args
		want    []*VariableEntityMapping
		wantErr bool
	}{
		{
			name: "get_variables_when_no_present_returns_empty",
			args: args{entities: []Entity{{
				EntityType: EntityTypeDeploymentTemplateAppLevel,
				EntityId:   1,
			}}},
			want:    make([]*VariableEntityMapping, 0),
			wantErr: false,
		}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getVariableEntityMappingRepositoryImpl(t)
			got, err := impl.GetVariablesForEntities(tt.args.entities)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVariablesForEntities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVariablesForEntities() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariableEntityMappingRepositoryImpl_SaveVariableEntityMappings(t *testing.T) {
	type args struct {
		mappings []*VariableEntityMapping
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "save_multiple_mappings_saves_successfully",
			args: args{
				mappings: []*VariableEntityMapping{{
					VariableName: "test1",
					IsDeleted:    false,
					Entity: Entity{
						EntityType: EntityTypeDeploymentTemplateAppLevel,
						EntityId:   1,
					}, AuditLog: sql.AuditLog{
						CreatedOn: time.Now(),
						CreatedBy: 1,
						UpdatedOn: time.Now(),
						UpdatedBy: 1,
					}},
					{
						VariableName: "test2",
						IsDeleted:    false,
						Entity: Entity{
							EntityType: EntityTypeDeploymentTemplateAppLevel,
							EntityId:   1,
						}, AuditLog: sql.AuditLog{
							CreatedOn: time.Now(),
							CreatedBy: 1,
							UpdatedOn: time.Now(),
							UpdatedBy: 1,
						}},
					{
						VariableName: "test1",
						IsDeleted:    false,
						Entity: Entity{
							EntityType: EntityTypeDeploymentTemplateAppLevel,
							EntityId:   2,
						}, AuditLog: sql.AuditLog{
							CreatedOn: time.Now(),
							CreatedBy: 1,
							UpdatedOn: time.Now(),
							UpdatedBy: 1,
						}},
					{
						VariableName: "test2",
						IsDeleted:    false,
						Entity: Entity{
							EntityType: EntityTypeDeploymentTemplateAppLevel,
							EntityId:   2,
						}, AuditLog: sql.AuditLog{
							CreatedOn: time.Now(),
							CreatedBy: 1,
							UpdatedOn: time.Now(),
							UpdatedBy: 1,
						}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getVariableEntityMappingRepositoryImpl(t)
			tx, _ := impl.StartTx()
			if err := impl.SaveVariableEntityMappings(tx, tt.args.mappings); (err != nil) != tt.wantErr {
				t.Errorf("SaveVariableEntityMappings() error = %v, wantErr %v", err, tt.wantErr)
			}
			impl.CommitTx(tx)
		})
	}
}
