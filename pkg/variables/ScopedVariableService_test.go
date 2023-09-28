package variables

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/app/mocks"
	util2 "github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	mocks3 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	mocks4 "github.com/devtron-labs/devtron/pkg/devtronResource/mocks"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	mocks5 "github.com/devtron-labs/devtron/pkg/resourceQualifiers/mocks"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/pkg/variables/repository/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
	"time"
)

func TestScopedVariableServiceImpl_CreateVariables(t *testing.T) {
	t.Setenv("VARIABLE_CACHE_ENABLED", "false")
	payload := &models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "Var1",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
			{
				Definition: models.Definition{
					VarName:     "Var2",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 2",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value2",
						},
						AttributeType: "Cluster",
						AttributeParams: map[models.IdentifierType]string{
							"ClusterName": "default_cluster",
						},
					},
				},
			},
		},
		UserId: 2,
	}
	payload1 := &models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "Var1",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Application",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
						},
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Env",
						AttributeParams: map[models.IdentifierType]string{
							"EnvName": "Dev",
						},
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Global",
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Cluster",
						AttributeParams: map[models.IdentifierType]string{
							"ClusterName": "default_cluster",
						},
					},
				},
			},
		},
	}

	varDef := []*repository.VariableDefinition{
		{
			Id:          1,
			Name:        "Var1",
			DataType:    "primitive",
			VarType:     "public",
			Description: "Variable 1",
		},
		{
			Id:          2,
			Name:        "Var2",
			DataType:    "primitive",
			VarType:     "public",
			Description: "Variable 2",
		},
	}

	parentScope := []*models.VariableScope{
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    1,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
				CompositeKey:          "",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    2,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "Dev",
				CompositeKey:          "",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    3,
				ResourceId:            2,
				QualifierId:           3,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "default_cluster",
				CompositeKey:          "",
			},
			Data: "value2",
		},
	}
	parentScope1 := []*models.VariableScope{
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    1,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
				CompositeKey:          "1-dev-test-Dev",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    2,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "Dev",
				CompositeKey:          "1-dev-test-Dev",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    3,
				ResourceId:            1,
				QualifierId:           2,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
				CompositeKey:          "",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    4,
				ResourceId:            1,
				QualifierId:           3,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "Dev",
				CompositeKey:          "",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:            5,
				ResourceId:    1,
				QualifierId:   5,
				IdentifierKey: 7,
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    6,
				ResourceId:            1,
				QualifierId:           4,
				IdentifierKey:         8,
				IdentifierValueInt:    3,
				IdentifierValueString: "default_cluster",
				CompositeKey:          "",
			},
			Data: "value1",
		},
	}
	searchableKeyMap := map[bean.DevtronResourceSearchableKeyName]int{
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME:           1,
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME:           2,
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV:      3,
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH:         4,
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION: 5,
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID:                     6,
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID:                     7,
		bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID:                 8,
	}
	appNameToId := []*app.App{
		{
			Id:      1,
			AppName: "dev-test",
		},
	}
	envNameToId := []*repository2.Environment{
		{
			Id:   1,
			Name: "Dev",
		},
	}
	clusterNameToId := []*repository2.Cluster{
		{
			Id:          1,
			ClusterName: "default_cluster",
		},
	}
	appNameToId1 := []*app.App{
		{
			Id:      1,
			AppName: "dev-test2",
		},
	}
	envNameToId1 := []*repository2.Environment{
		{
			Id:   1,
			Name: "Dev2",
		},
	}
	clusterNameToId1 := []*repository2.Cluster{
		{
			Id:          1,
			ClusterName: "default_cluster2",
		},
	}
	//appNameToIdMap :=map[string]int{
	//
	//}

	type args struct {
		payload *models.Payload
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "Valid case with all data available",
			args: args{
				payload: payload,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Valid case with all scopes in variable  available",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test for error in starting Transaction",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for error in Deletion",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for error in saving data in variable definition",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for error in saving data in variable scope",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for error in saving data in variable data",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for committing transaction",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for error cases in createVariableScopes wrong appNameToId",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for error cases in createVariableScopes wrong envNameToId",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test for error cases in createVariableScopes wrong clusterNameToId",
			args: args{
				payload: payload1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			impl, scopedVariableRepository, appRepository, environmentRepository, devtronResourceService, clusterRepository, _ := InitScopedVariableServiceImpl(t)
			var err error
			tx := &pg.Tx{}
			if tt.name == "Valid case with all data available" {

				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("CommitTx", tx).Return(nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				scopedVariableRepository.On("CreateVariableScope", mock.AnythingOfType("[]*repository.VariableScope"), tx).Return(parentScope, nil)
				scopedVariableRepository.On("CreateVariableData", mock.AnythingOfType("[]*repository.VariableData"), tx).Return(nil)

			}
			if tt.name == "Valid case with all scopes in variable  available" {
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("CommitTx", tx).Return(nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				scopedVariableRepository.On("CreateVariableScope", mock.AnythingOfType("[]*repository.VariableScope"), tx).Return(parentScope1, nil)
				scopedVariableRepository.On("CreateVariableData", mock.AnythingOfType("[]*repository.VariableData"), tx).Return(nil)

			}
			if tt.name == "Test for error in starting Transaction" {
				scopedVariableRepository.On("StartTx").Return(nil, errors.New("error in transaction"))
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)

			}
			if tt.name == "Test for error in Deletion" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(errors.New("error in deletion"))
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)

			}
			if tt.name == "Test for error in saving data in variable definition" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(nil, errors.New("error in saving data in variable definition"))

			}
			if tt.name == "Test for error in saving data in variable scope" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				scopedVariableRepository.On("CreateVariableScope", mock.AnythingOfType("[]*repository.VariableScope"), tx).Return(nil, errors.New("error in saving variable scope"))
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)

			}
			if tt.name == "Test for error in saving data in variable data" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				scopedVariableRepository.On("CreateVariableScope", mock.AnythingOfType("[]*repository.VariableScope"), tx).Return(parentScope1, nil)
				scopedVariableRepository.On("CreateVariableData", mock.AnythingOfType("[]*repository.VariableData"), tx).Return(errors.New("error in saving variable data"))

			}
			if tt.name == "Test for committing transaction" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("CommitTx", tx).Return(errors.New("error in committing transaction"))
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				scopedVariableRepository.On("CreateVariableScope", mock.AnythingOfType("[]*repository.VariableScope"), tx).Return(parentScope1, nil)
				scopedVariableRepository.On("CreateVariableData", mock.AnythingOfType("[]*repository.VariableData"), tx).Return(nil)

			}
			if tt.name == "Test for error cases in createVariableScopes wrong appNameToId" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId1, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)

			}
			if tt.name == "Test for error cases in createVariableScopes wrong envNameToId" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId1, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)

			}
			if tt.name == "Test for error cases in createVariableScopes wrong clusterNameToId" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Maybe().Return(varDef, nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId1, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)

			}

			if err = impl.CreateVariables(*tt.args.payload); (err != nil) != tt.wantErr {
				t.Errorf("CreateVariables() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(1000)
		})
	}
}

func TestScopedVariableServiceImpl_IsValidPayload(t *testing.T) {
	// Create a test case with a valid payload
	validPayload := models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "Var1",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	impl := &ScopedVariableServiceImpl{
		VariableNameConfig: &VariableConfig{
			VariableNameRegex: "^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$",
		},
	}

	// Test the valid payload
	err, isValid := impl.isValidPayload(validPayload)
	assert.NoError(t, err)
	assert.True(t, isValid)

	// Create a test case with a duplicate variable name
	duplicateVarNamePayload := models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "Var1",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
			{
				Definition: models.Definition{
					VarName:     "Var1", // Duplicate VarName
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 2",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value2",
						},
						AttributeType: "Cluster",
						AttributeParams: map[models.IdentifierType]string{
							"ClusterName": "default_cluster",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the duplicate variable name case
	err, isValid = impl.isValidPayload(duplicateVarNamePayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "duplicate variable name Var1", err.Error())

	// Create a test case with invalid AttributeParams length
	invalidAttributeParamsLengthPayload := models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "Var1",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							// Missing a required IdentifierType
							"EnvName": "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case with invalid AttributeParams length
	err, isValid = impl.isValidPayload(invalidAttributeParamsLengthPayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "attribute selectors are not valid for given category ApplicationEnv", err.Error())

	// Create a test case with an invalid IdentifierType
	invalidIdentifierTypePayload := models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "Var1",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"InvalidType":     "Dev", // Invalid IdentifierType
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case with an invalid IdentifierType
	err, isValid = impl.isValidPayload(invalidIdentifierTypePayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "invalid attribute selector key InvalidType", err.Error())

	// Create a test case with duplicate AttributeParams
	duplicateAttributeParamsPayload := models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "Var1",
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case with duplicate AttributeParams
	err, isValid = impl.isValidPayload(duplicateAttributeParamsPayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "duplicate Selectors  found for category ApplicationEnv", err.Error())

	invalidVarNamePayload := models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "_InvalidVariable1", // Contains  invalid varName
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case where the variable name doesn't match the regex pattern
	err, isValid = impl.isValidPayload(invalidVarNamePayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "_InvalidVariable1 does not match the required format (Alphanumeric, 64 characters max, no hyphen/underscore at start/end)", err.Error())
	invalidVarNamePayload = models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "InvalidVariable1_", // Contains  invalid varName
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case where the variable name doesn't match the regex pattern
	err, isValid = impl.isValidPayload(invalidVarNamePayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "InvalidVariable1_ does not match the required format (Alphanumeric, 64 characters max, no hyphen/underscore at start/end)", err.Error())
	invalidVarNamePayload = models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "1InvalidVariable1", // Contains  invalid varName
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case where the variable name doesn't match the regex pattern
	err, isValid = impl.isValidPayload(invalidVarNamePayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "1InvalidVariable1 does not match the required format (Alphanumeric, 64 characters max, no hyphen/underscore at start/end)", err.Error())

	invalidVarNamePayload = models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "-InvalidVariable1", // Contains  invalid
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case where the variable name doesn't match the regex pattern
	err, isValid = impl.isValidPayload(invalidVarNamePayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "-InvalidVariable1 does not match the required format (Alphanumeric, 64 characters max, no hyphen/underscore at start/end)", err.Error())

	invalidVarNamePayload = models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "InvalidVariable1-", // Contains  invalid
					DataType:    "primitive",
					VarType:     "public",
					Description: "Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
				},
			},
		},
		UserId: 2,
	}

	// Test the case where the variable name doesn't match the regex pattern
	err, isValid = impl.isValidPayload(invalidVarNamePayload)
	assert.Error(t, err)
	assert.False(t, isValid)
	assert.Equal(t, "InvalidVariable1- does not match the required format (Alphanumeric, 64 characters max, no hyphen/underscore at start/end)", err.Error())

}

func InitScopedVariableServiceImpl(t *testing.T) (*ScopedVariableServiceImpl, *mocks.ScopedVariableRepository, *mocks2.AppRepository, *mocks3.EnvironmentRepository, *mocks4.DevtronResourceService, *mocks3.ClusterRepository, *mocks5.QualifierMappingService) {
	logger, _ := util2.NewSugardLogger()
	scopedVariableRepository := mocks.NewScopedVariableRepository(t)
	appRepository := mocks2.NewAppRepository(t)
	environmentRepository := mocks3.NewEnvironmentRepository(t)
	devtronResourceService := mocks4.NewDevtronResourceService(t)
	clusterRepository := mocks3.NewClusterRepository(t)
	qualifierMappingService := mocks5.NewQualifierMappingService(t)
	impl, _ := NewScopedVariableServiceImpl(logger, scopedVariableRepository, appRepository, environmentRepository, devtronResourceService, clusterRepository, qualifierMappingService)

	return impl, scopedVariableRepository, appRepository, environmentRepository, devtronResourceService, clusterRepository, qualifierMappingService
}

func TestScopedVariableServiceImpl_GetScopedVariables(t *testing.T) {
	t.Setenv("VARIABLE_CACHE_ENABLED", "false")
	varDef := []*repository.VariableDefinition{
		{
			Id:          1,
			Name:        "Var1",
			DataType:    "primitive",
			VarType:     "public",
			Description: "Variable 1",
		},
		{
			Id:          2,
			Name:        "Var2",
			DataType:    "primitive",
			VarType:     "public",
			Description: "Variable 2",
		},
	}
	varDef1 := []*repository.VariableDefinition{
		{
			Id:          1,
			Name:        "Var1",
			DataType:    "primitive",
			VarType:     "public",
			Description: "Variable 1",
		},
	}
	varDef2 := []*repository.VariableDefinition{
		{
			Id:          1,
			Name:        "PrivateVar",
			DataType:    "primitive",
			VarType:     "private",
			Description: "Variable 1",
		},
	}
	variableScope := []*models.VariableScope{
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    1,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
			},
			Data: "value-1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    2,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "Dev",
				ParentIdentifier:      1,
			},
			Data: "value-1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    3,
				ResourceId:            1,
				QualifierId:           2,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
				CompositeKey:          "",
			},
			Data: "value-2",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    4,
				ResourceId:            1,
				QualifierId:           3,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "Dev",
				CompositeKey:          "",
			},
			Data: "value-3",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:            5,
				ResourceId:    1,
				QualifierId:   5,
				IdentifierKey: 7,
			},
			Data: "value-4",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    6,
				ResourceId:            1,
				QualifierId:           4,
				IdentifierKey:         8,
				IdentifierValueInt:    3,
				IdentifierValueString: "default_cluster",
				CompositeKey:          "",
			},
			Data: "value-5",
		},
	}
	varData := []*repository.VariableData{
		{
			Id:              1,
			VariableScopeId: 1,
			Data:            "\"" + "value-1" + "\"",
		},
		{
			Id:              2,
			VariableScopeId: 3,
			Data:            "\"" + "value-2" + "\"",
		},
		{
			Id:              3,
			VariableScopeId: 4,
			Data:            "\"" + "value-3" + "\"",
		},
		{
			Id:              4,
			VariableScopeId: 5,
			Data:            "\"" + "value-4" + "\"",
		},
		{
			Id:              5,
			VariableScopeId: 6,
			Data:            "\"" + "value-5" + "\"",
		},
	}
	varData1 := []*repository.VariableData{
		{
			Id:              1,
			VariableScopeId: 1,
			Data:            "\"" + "value-private" + "\"",
		},
	}
	variableScope1 := []*models.VariableScope{
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    1,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
			},
			Data: "value-1",
		},

		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    3,
				ResourceId:            1,
				QualifierId:           2,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
				CompositeKey:          "",
			},
			Data: "value-2",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:            5,
				ResourceId:    1,
				QualifierId:   5,
				IdentifierKey: 7,
			},
			Data: "value-4",
		},
	}
	variableScope2 := []*models.VariableScope{
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    1,
				ResourceId:            1,
				QualifierId:           3,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
			},
			Data: "value-1",
		},
	}

	scope := resourceQualifiers.Scope{AppId: 1, ClusterId: 1, EnvId: 1}
	scope1 := resourceQualifiers.Scope{AppId: 1}
	scopedVariableData := []*models.ScopedVariableData{
		{
			VariableName:     "Var1",
			ShortDescription: "variable 1",
			VariableValue: &models.VariableValue{
				Value: "value-1",
			},
		},
		{
			VariableName:     "Var2",
			ShortDescription: "variable 2",
		},
	}
	scopedVariableData1 := []*models.ScopedVariableData{
		{
			VariableName:     "Var1",
			ShortDescription: "variable 1",
			VariableValue: &models.VariableValue{
				Value: "value-1",
			},
		},
	}
	scopedVariableData2 := []*models.ScopedVariableData{
		{
			VariableName:     "Var1",
			ShortDescription: "variable 1",
			VariableValue: &models.VariableValue{
				Value: "value-2",
			},
		},
	}
	scopedVariableData3 := []*models.ScopedVariableData{
		{
			VariableName:     "PrivateVar",
			ShortDescription: "variable 1",
			VariableValue: &models.VariableValue{
				Value: "",
			},
		},
	}
	type args struct {
		scope        resourceQualifiers.Scope
		varNames     []string
		isSuperAdmin bool
	}
	tests := []struct {
		name                      string
		args                      args
		wantScopedVariableDataObj []*models.ScopedVariableData
		wantErr                   assert.ErrorAssertionFunc
	}{
		{
			name: "NoVariablesFound",
			args: args{
				scope:    scope,
				varNames: []string{"Var1", "Var2"},
			},
			wantErr:                   assert.NoError,
			wantScopedVariableDataObj: nil, // No variables found, so the result should be nil
		},
		{
			name: "get scoped variable data",
			args: args{
				scope:    scope,
				varNames: []string{"Var1", "Var2"},
			},
			wantErr:                   assert.NoError,
			wantScopedVariableDataObj: scopedVariableData1,
		},
		{
			name: "get data when varName is not provided",
			args: args{
				scope: scope,
				//varNames: []string{"Var1", "Var2"},
			},
			wantErr:                   assert.NoError,
			wantScopedVariableDataObj: scopedVariableData,
		},
		{
			name: "test for  error cases  in GetScopedVariableData",
			args: args{
				scope:    scope,
				varNames: []string{"Var1", "Var2"},
			},
			wantErr:                   assert.Error,
			wantScopedVariableDataObj: nil,
		},
		{
			name: "test for  error cases  in GetDataForScopeIds",
			args: args{
				scope:    scope,
				varNames: []string{"Var1", "Var2"},
			},
			wantErr:                   assert.Error,
			wantScopedVariableDataObj: nil,
		},
		{
			name: "get scoped variable data for provided appId",
			args: args{
				scope: scope1,
			},
			wantErr:                   assert.NoError,
			wantScopedVariableDataObj: scopedVariableData2,
		},
		{
			name: "test for private variable",
			args: args{
				scope:        scope1,
				isSuperAdmin: false,
			},
			wantErr:                   assert.NoError,
			wantScopedVariableDataObj: scopedVariableData3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl, scopedVariableRepository, _, _, _, _, qualifierMappingService := InitScopedVariableServiceImpl(t)
			if tt.name == "NoVariablesFound" {
				scopedVariableRepository.On("GetAllVariables").Return(nil, nil)
			}
			if tt.name == "get scoped variable data" {
				var varIds []int
				for _, definition := range varDef {
					varIds = append(varIds, definition.Id)
				}
				var mappings []*resourceQualifiers.QualifierMapping
				for _, varScope := range variableScope {
					mappings = append(mappings, varScope.QualifierMapping)
				}
				qualifierMappingService.On("GetQualifierMappings", resourceQualifiers.Variable, &scope, varIds).Return(mappings, nil)
				scopedVariableRepository.On("GetAllVariables").Return(varDef, nil)
				scopedVariableRepository.On("GetDataForScopeIds", []int{1}).Return(varData, nil)
			}
			if tt.name == "test for  error cases  in GetScopedVariableData" {
				var varIds []int
				for _, definition := range varDef {
					varIds = append(varIds, definition.Id)
				}
				var mappings []*resourceQualifiers.QualifierMapping
				var scopeIds []int
				for _, varScope := range variableScope {
					mappings = append(mappings, varScope.QualifierMapping)
					scopeIds = append(scopeIds, varScope.Id)
				}
				qualifierMappingService.On("GetQualifierMappings", resourceQualifiers.Variable, &scope, varIds).Return(mappings, nil)
				scopedVariableRepository.On("GetAllVariables").Return(varDef, nil)
				scopedVariableRepository.On("GetDataForScopeIds", mock.AnythingOfType("[]int")).Return(nil, errors.New("error in getting varScope"))
			}
			if tt.name == "test for  error cases  in GetDataForScopeIds" {
				var varIds []int
				for _, definition := range varDef {
					varIds = append(varIds, definition.Id)
				}
				var mappings []*resourceQualifiers.QualifierMapping
				for _, varScope := range variableScope {
					mappings = append(mappings, varScope.QualifierMapping)
				}
				qualifierMappingService.On("GetQualifierMappings", resourceQualifiers.Variable, &scope, varIds).Return(mappings, nil)
				scopedVariableRepository.On("GetAllVariables").Return(varDef, nil)
				scopedVariableRepository.On("GetDataForScopeIds", mock.AnythingOfType("[]int")).Return(nil, errors.New("error in getting variable data"))
			}
			if tt.name == "get data when varName is not provided" {
				var varIds []int
				for _, definition := range varDef {
					varIds = append(varIds, definition.Id)
				}
				var mappings []*resourceQualifiers.QualifierMapping
				for _, varScope := range variableScope {
					mappings = append(mappings, varScope.QualifierMapping)
				}
				qualifierMappingService.On("GetQualifierMappings", resourceQualifiers.Variable, &scope, varIds).Return(mappings, nil)
				scopedVariableRepository.On("GetAllVariables").Return(varDef, nil)
				scopedVariableRepository.On("GetDataForScopeIds", mock.AnythingOfType("[]int")).Return(varData, nil)
			}
			if tt.name == "get scoped variable data for provided appId" {
				var varIds []int
				for _, definition := range varDef1 {
					varIds = append(varIds, definition.Id)
				}
				var mappings []*resourceQualifiers.QualifierMapping
				for _, varScope := range variableScope1 {
					mappings = append(mappings, varScope.QualifierMapping)
				}
				qualifierMappingService.On("GetQualifierMappings", resourceQualifiers.Variable, &scope1, varIds).Return(mappings, nil)
				scopedVariableRepository.On("GetAllVariables").Return(varDef1, nil)
				scopedVariableRepository.On("GetDataForScopeIds", mock.AnythingOfType("[]int")).Return(varData, nil)
			}
			if tt.name == "test for private variable" {
				var varIds []int
				for _, definition := range varDef2 {
					varIds = append(varIds, definition.Id)
				}
				var mappings []*resourceQualifiers.QualifierMapping
				for _, varScope := range variableScope2 {
					mappings = append(mappings, varScope.QualifierMapping)
				}
				qualifierMappingService.On("GetQualifierMappings", resourceQualifiers.Variable, &scope1, varIds).Return(mappings, nil)
				scopedVariableRepository.On("GetAllVariables").Return(varDef2, nil)
				scopedVariableRepository.On("GetDataForScopeIds", mock.AnythingOfType("[]int")).Return(varData1, nil)
			}

			gotScopedVariableDataObj, err := impl.GetScopedVariables(tt.args.scope, tt.args.varNames, tt.args.isSuperAdmin)
			if !tt.wantErr(t, err, fmt.Sprintf("GetScopedVariables(%v, %v)", tt.args.scope, tt.args.varNames)) {
				return
			}
			if tt.name == "get scoped variable data" || tt.name == "get data when varName is not provided" || tt.name == "get scoped variable data for provided appId" || tt.name == "test for private variable" {
				assert.Equalf(t, tt.wantScopedVariableDataObj[0].VariableName, gotScopedVariableDataObj[0].VariableName, "GetScopedVariables(%v, %v)", tt.args.scope, tt.args.varNames)
				assert.Equalf(t, tt.wantScopedVariableDataObj[0].VariableValue.Value, gotScopedVariableDataObj[0].VariableValue.Value, "GetScopedVariables(%v, %v)", tt.args.scope, tt.args.varNames)
			} else {
				assert.Equalf(t, tt.wantScopedVariableDataObj, gotScopedVariableDataObj, "GetScopedVariables(%v, %v)", tt.args.scope, tt.args.varNames)
			}

		})
	}
}

func TestScopedVariableServiceImpl_GetJsonForVariables(t *testing.T) {
	t.Setenv("VARIABLE_CACHE_ENABLED", "false")
	searchableKeyMap := map[int]bean.DevtronResourceSearchableKeyName{
		1: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_APP_NAME,
		2: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ENV_NAME,
		3: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_IS_ALL_PRODUCTION_ENV,
		4: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_BRANCH,
		5: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CI_PIPELINE_TRIGGER_ACTION,
		6: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID,
		7: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID,
		8: bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID,
	}
	payload1 := &models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:          "Var1",
					DataType:         "primitive",
					VarType:          "public",
					Description:      "Variable 1",
					ShortDescription: "ShortDescription-Variable 1",
				},
				AttributeValues: []models.AttributeValue{
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "ApplicationEnv",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
							"EnvName":         "Dev",
						},
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Application",
						AttributeParams: map[models.IdentifierType]string{
							"ApplicationName": "dev-test",
						},
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Env",
						AttributeParams: map[models.IdentifierType]string{
							"EnvName": "Dev",
						},
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Global",
					},
					{
						VariableValue: models.VariableValue{
							Value: "value1",
						},
						AttributeType: "Cluster",
						AttributeParams: map[models.IdentifierType]string{
							"ClusterName": "default_cluster",
						},
					},
				},
			},
		},
	}
	variableDataArray := []*repository.VariableData{
		{
			Id:              1,
			VariableScopeId: 1,
			Data:            "value1",
		},
		{
			Id:              2,
			VariableScopeId: 3,
			Data:            "value1",
		},
		{
			Id:              3,
			VariableScopeId: 4,
			Data:            "value1",
		},
		{
			Id:              4,
			VariableScopeId: 5,
			Data:            "value1",
		},
		{
			Id:              5,
			VariableScopeId: 6,
			Data:            "value1",
		},
	}
	variableScope := []*models.VariableScope{
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    1,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    2,
				ResourceId:            1,
				QualifierId:           1,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "Dev",
				ParentIdentifier:      1,
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    3,
				ResourceId:            1,
				QualifierId:           2,
				IdentifierKey:         6,
				IdentifierValueInt:    3,
				IdentifierValueString: "dev-test",
				CompositeKey:          "",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    4,
				ResourceId:            1,
				QualifierId:           3,
				IdentifierKey:         7,
				IdentifierValueInt:    3,
				IdentifierValueString: "Dev",
				CompositeKey:          "",
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:          5,
				ResourceId:  1,
				QualifierId: 5,
			},
			Data: "value1",
		},
		{
			QualifierMapping: &resourceQualifiers.QualifierMapping{
				Id:                    6,
				ResourceId:            1,
				QualifierId:           4,
				IdentifierKey:         8,
				IdentifierValueInt:    3,
				IdentifierValueString: "default_cluster",
				CompositeKey:          "",
			},
			Data: "value1",
		},
	}

	variableDefinition := []*repository.VariableDefinition{
		{
			Id:               1,
			Name:             "Var1",
			DataType:         "primitive",
			VarType:          "public",
			Description:      "Variable 1",
			ShortDescription: "ShortDescription-Variable 1",
			Active:           true,
		},
	}

	tests := []struct {
		name string

		want    *models.Payload
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "test for getting json",
			want:    payload1,
			wantErr: assert.NoError,
		},
		{
			name:    "test for error cases in GetAllVariableDefinition",
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name:    "test for empty payload",
			want:    nil,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl, scopedVariableRepository, _, _, devtronResourceService, _, qualifierMappingService := InitScopedVariableServiceImpl(t)
			if tt.name == "test for getting json" {
				scopedVariableRepository.On("GetAllVariableDefinition").Return(variableDefinition, nil)

				var varDefIds []int
				for _, definition := range variableDefinition {
					varDefIds = append(varDefIds, definition.Id)
				}
				var qualifierMappings []*resourceQualifiers.QualifierMapping
				var scopeIds []int
				for _, varScope := range variableScope {
					qualifierMappings = append(qualifierMappings, varScope.QualifierMapping)
					scopeIds = append(scopeIds, varScope.Id)
				}
				qualifierMappingService.On("GetQualifierMappings", resourceQualifiers.Variable, mock.AnythingOfType("*resourceQualifiers.Scope"), varDefIds).Return(qualifierMappings, nil)
				scopedVariableRepository.On("GetDataForScopeIds", scopeIds).Return(variableDataArray, nil)
				devtronResourceService.On("GetAllSearchableKeyIdNameMap").Return(searchableKeyMap)
			}
			if tt.name == "test for error cases in GetAllVariableDefinition" {
				scopedVariableRepository.On("GetAllVariableDefinition").Return(nil, errors.New("error in getting data for json"))
			}
			if tt.name == "test for empty payload" {
				scopedVariableRepository.On("GetAllVariableDefinition").Return(nil, nil)
				devtronResourceService.On("GetAllSearchableKeyIdNameMap").Return(searchableKeyMap)
			}
			got, err := impl.GetJsonForVariables()
			if !tt.wantErr(t, err, fmt.Sprintf("GetJsonForVariables()")) {
				return
			}
			assert.True(t, reflect.DeepEqual(tt.want, got), tt.want, got)
		})
	}
}
