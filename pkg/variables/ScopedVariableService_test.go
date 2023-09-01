package variables

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/app/mocks"
	util2 "github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	mocks3 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	mocks4 "github.com/devtron-labs/devtron/pkg/devtronResource/mocks"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/pkg/variables/repository/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestScopedVariableServiceImpl_CreateVariables(t *testing.T) {

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
	//payload1 := &models.Payload{
	//	Variables: []*models.Variables{
	//		{
	//			Definition: models.Definition{
	//				VarName:     "Var1",
	//				DataType:    "primitive",
	//				VarType:     "public",
	//				Description: "Variable 1",
	//			},
	//			AttributeValues: []models.AttributeValue{
	//				{
	//					VariableValue: models.VariableValue{
	//						Value: "value1",
	//					},
	//					AttributeType: "ApplicationEnv",
	//					AttributeParams: map[models.IdentifierType]string{
	//						"ApplicationName": "dev-test",
	//						"EnvName":         "Dev",
	//					},
	//				},
	//				{
	//					VariableValue: models.VariableValue{
	//						Value: "value1",
	//					},
	//					AttributeType: "Application",
	//					AttributeParams: map[models.IdentifierType]string{
	//						"ApplicationName": "dev-test",
	//					},
	//				},
	//				{
	//					VariableValue: models.VariableValue{
	//						Value: "value1",
	//					},
	//					AttributeType: "Env",
	//					AttributeParams: map[models.IdentifierType]string{
	//						"EnvName": "Dev",
	//					},
	//				},
	//				{
	//					VariableValue: models.VariableValue{
	//						Value: "value1",
	//					},
	//					AttributeType: "Global",
	//				},
	//				{
	//					VariableValue: models.VariableValue{
	//						Value: "value1",
	//					},
	//					AttributeType: "Cluster",
	//					AttributeParams: map[models.IdentifierType]string{
	//						"ClusterName": "default_cluster",
	//					},
	//				},
	//			},
	//		},
	//		{
	//			Definition: models.Definition{
	//				VarName:     "Var2",
	//				DataType:    "primitive",
	//				VarType:     "public",
	//				Description: "Variable 2",
	//			},
	//			AttributeValues: []models.AttributeValue{
	//				{
	//					VariableValue: models.VariableValue{
	//						Value: "value2",
	//					},
	//					AttributeType: "Cluster",
	//					AttributeParams: map[models.IdentifierType]string{
	//						"ClusterName": "default_cluster",
	//					},
	//				},
	//			},
	//		},
	//	},
	//	UserId: 2,
	//}
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
	parentScope := []*repository.VariableScope{
		{
			Id:                    1,
			VariableDefinitionId:  1,
			QualifierId:           1,
			IdentifierKey:         6,
			IdentifierValueInt:    3,
			IdentifierValueString: "dev-test",
			CompositeKey:          "",
			Data:                  "value1",
		},
		{
			Id:                    2,
			VariableDefinitionId:  1,
			QualifierId:           1,
			IdentifierKey:         7,
			IdentifierValueInt:    3,
			IdentifierValueString: "Dev",
			CompositeKey:          "",
			Data:                  "value1",
		},
		{
			Id:                    3,
			VariableDefinitionId:  2,
			QualifierId:           3,
			IdentifierKey:         7,
			IdentifierValueInt:    3,
			IdentifierValueString: "default_cluster",
			CompositeKey:          "",
			Data:                  "value2",
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
	type args struct {
		payload *models.Payload
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{name: "Valid case with all data available",
			args: args{
				payload: payload,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			impl, scopedVariableRepository, appRepository, environmentRepository, devtronResourceService, clusterRepository := InitScopedVariableServiceImpl(t)
			var err error
			tx := &pg.Tx{}
			if tt.name == "Valid case with all data available" {
				scopedVariableRepository.On("DeleteVariables", mock.AnythingOfType("AuditLog"), tx).Return(nil)
				scopedVariableRepository.On("StartTx").Return(tx, nil)
				scopedVariableRepository.On("RollbackTx", tx).Return(nil)
				scopedVariableRepository.On("CommitTx", tx).Return(nil)
				scopedVariableRepository.On("CreateVariableDefinition", mock.AnythingOfType("[]*repository.VariableDefinition"), tx).Return(varDef, nil)
				scopedVariableRepository.On("GetAllVariableMetadata").Return(varDef, nil)
				devtronResourceService.On("GetAllSearchableKeyNameIdMap").Return(searchableKeyMap)
				appRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(appNameToId, nil)
				environmentRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(envNameToId, nil)
				clusterRepository.On("FindByNames", mock.AnythingOfType("[]string")).Return(clusterNameToId, nil)
				scopedVariableRepository.On("CreateVariableScope", mock.AnythingOfType("[]*repository.VariableScope"), tx).Return(parentScope, nil)
				scopedVariableRepository.On("CreateVariableData", mock.AnythingOfType("[]*repository.VariableData"), tx).Return(nil)

			}

			if err = impl.CreateVariables(*tt.args.payload); (err != nil) != tt.wantErr {
				t.Errorf("CreateVariables() error = %v, wantErr %v", err, tt.wantErr)
			}
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
		variableNameConfig: &VariableConfig{
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
	assert.Equal(t, "duplicate variable name", err.Error())

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
	assert.Equal(t, "length of AttributeParams is not valid", err.Error())

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
	assert.Equal(t, "invalid IdentifierType InvalidType for validIdentifierTypeList [ApplicationName EnvName]", err.Error())

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
	assert.Equal(t, "duplicate AttributeParams found for AttributeType ApplicationEnv", err.Error())

	invalidVarNamePayload := models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "_InvalidVariable1", // Contains  invalid
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
	assert.Equal(t, "variable name _InvalidVariable1 doesnot match regex ^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$", err.Error())
	invalidVarNamePayload = models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "InvalidVariable1_", // Contains  invalid
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
	assert.Equal(t, "variable name InvalidVariable1_ doesnot match regex ^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$", err.Error())
	invalidVarNamePayload = models.Payload{
		Variables: []*models.Variables{
			{
				Definition: models.Definition{
					VarName:     "1InvalidVariable1", // Contains  invalid
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
	assert.Equal(t, "variable name 1InvalidVariable1 doesnot match regex ^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$", err.Error())

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
	assert.Equal(t, "variable name -InvalidVariable1 doesnot match regex ^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$", err.Error())

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
	assert.Equal(t, "variable name InvalidVariable1- doesnot match regex ^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$", err.Error())

}

func InitScopedVariableServiceImpl(t *testing.T) (*ScopedVariableServiceImpl, *mocks.ScopedVariableRepository, *mocks2.AppRepository, *mocks3.EnvironmentRepository, *mocks4.DevtronResourceService, *mocks3.ClusterRepository) {
	logger, _ := util2.NewSugardLogger()
	scopedVariableRepository := mocks.NewScopedVariableRepository(t)
	appRepository := mocks2.NewAppRepository(t)
	environmentRepository := mocks3.NewEnvironmentRepository(t)
	devtronResourceService := mocks4.NewDevtronResourceService(t)
	clusterRepository := mocks3.NewClusterRepository(t)

	impl, _ := NewScopedVariableServiceImpl(logger, scopedVariableRepository, appRepository, environmentRepository, devtronResourceService, clusterRepository)
	return impl, scopedVariableRepository, appRepository, environmentRepository, devtronResourceService, clusterRepository
}
