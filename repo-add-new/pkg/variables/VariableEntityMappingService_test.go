package variables

import (
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
	"reflect"
	"testing"
)

func TestNewVariableEntityMappingServiceImpl(t *testing.T) {
	type args struct {
		variableEntityMappingRepository repository.VariableEntityMappingRepository
		logger                          *zap.SugaredLogger
	}
	tests := []struct {
		name string
		args args
		want *VariableEntityMappingServiceImpl
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewVariableEntityMappingServiceImpl(tt.args.variableEntityMappingRepository, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewVariableEntityMappingServiceImpl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariableEntityMappingServiceImpl_DeleteMappingsForEntities(t *testing.T) {
	type fields struct {
		logger                          *zap.SugaredLogger
		variableEntityMappingRepository repository.VariableEntityMappingRepository
	}
	type args struct {
		entities []repository.Entity
		userId   int32
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
			impl := VariableEntityMappingServiceImpl{
				logger:                          tt.fields.logger,
				variableEntityMappingRepository: tt.fields.variableEntityMappingRepository,
			}
			if err := impl.DeleteMappingsForEntities(tt.args.entities, tt.args.userId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMappingsForEntities() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVariableEntityMappingServiceImpl_GetAllMappingsForEntities(t *testing.T) {
	type fields struct {
		logger                          *zap.SugaredLogger
		variableEntityMappingRepository repository.VariableEntityMappingRepository
	}
	type args struct {
		entities []repository.Entity
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[repository.Entity][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := VariableEntityMappingServiceImpl{
				logger:                          tt.fields.logger,
				variableEntityMappingRepository: tt.fields.variableEntityMappingRepository,
			}
			got, err := impl.GetAllMappingsForEntities(tt.args.entities)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllMappingsForEntities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllMappingsForEntities() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariableEntityMappingServiceImpl_UpdateVariablesForEntity(t *testing.T) {
	type fields struct {
		logger                          *zap.SugaredLogger
		variableEntityMappingRepository repository.VariableEntityMappingRepository
	}
	type args struct {
		variableNames []string
		entity        repository.Entity
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
			impl := VariableEntityMappingServiceImpl{
				logger:                          tt.fields.logger,
				variableEntityMappingRepository: tt.fields.variableEntityMappingRepository,
			}
			if err := impl.UpdateVariablesForEntity(tt.args.variableNames, tt.args.entity, tt.args.userId); (err != nil) != tt.wantErr {
				t.Errorf("UpdateVariablesForEntity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
