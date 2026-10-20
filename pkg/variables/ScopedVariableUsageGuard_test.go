/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package variables

import (
	"errors"
	"testing"

	"github.com/devtron-labs/devtron/pkg/sql"
	mocks2 "github.com/devtron-labs/devtron/pkg/variables/mocks"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// scopedVariableRepoStub satisfies ScopedVariableRepository via interface embedding; only the
// methods used by the deletion guard are implemented, so any other call panics the test.
type scopedVariableRepoStub struct {
	repository2.ScopedVariableRepository
	metadata      []*repository2.VariableDefinition
	metadataErr   error
	deleteCalled  bool
	txStartCalled bool
}

func (s *scopedVariableRepoStub) GetAllVariableMetadata() ([]*repository2.VariableDefinition, error) {
	return s.metadata, s.metadataErr
}

func (s *scopedVariableRepoStub) DeleteVariables(auditLog sql.AuditLog, tx *pg.Tx) error {
	s.deleteCalled = true
	return nil
}

func (s *scopedVariableRepoStub) StartTx() (*pg.Tx, error) {
	s.txStartCalled = true
	return nil, errors.New("transaction must not be started when deletion is blocked")
}

func defsFor(names ...string) []*repository2.VariableDefinition {
	defs := make([]*repository2.VariableDefinition, 0, len(names))
	for _, name := range names {
		defs = append(defs, &repository2.VariableDefinition{Name: name, Active: true})
	}
	return defs
}

func payloadFor(names ...string) models.Payload {
	variables := make([]*models.Variables, 0, len(names))
	for _, name := range names {
		variables = append(variables, &models.Variables{
			Definition: models.Definition{VarName: name, DataType: models.PRIMITIVE_TYPE, VarType: models.PUBLIC},
		})
	}
	return models.Payload{Variables: variables, UserId: 1}
}

func usageFor(name string) *models.VariableUsage {
	return &models.VariableUsage{
		VariableName: name,
		UsageType:    models.UsageTypeDeploymentTemplate,
		AppId:        1,
		AppName:      "test-app",
	}
}

func guardTestService(t *testing.T, repoStub *scopedVariableRepoStub) (*ScopedVariableServiceImpl, *mocks2.VariableEntityMappingService) {
	logger, err := zap.NewDevelopment()
	assert.Nil(t, err)
	cfg, err := GetVariableNameConfig()
	assert.Nil(t, err)
	mappingService := mocks2.NewVariableEntityMappingService(t)
	impl := &ScopedVariableServiceImpl{
		logger:                       logger.Sugar(),
		scopedVariableRepository:     repoStub,
		variableEntityMappingService: mappingService,
		VariableNameConfig:           cfg,
	}
	return impl, mappingService
}

func TestDeletionGuard_FirstSaveNoExistingVariables(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: nil}
	impl, mappingService := guardTestService(t, repoStub)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("V1"))

	assert.Nil(t, err)
	mappingService.AssertNotCalled(t, "GetLiveVariableUsage", mock.Anything)
}

func TestDeletionGuard_NoVariableRemoved(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1", "V2")}
	impl, mappingService := guardTestService(t, repoStub)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("V1", "V2", "V3"))

	assert.Nil(t, err)
	mappingService.AssertNotCalled(t, "GetLiveVariableUsage", mock.Anything)
}

func TestDeletionGuard_RemovedVariableUnused(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1", "V2")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", []string{"V2"}).Return([]*models.VariableUsage{}, nil)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("V1"))

	assert.Nil(t, err)
}

func TestDeletionGuard_RemovedVariableInUse(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1", "V2")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", []string{"V2"}).
		Return([]*models.VariableUsage{usageFor("V2")}, nil)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("V1"))

	var blockedErr *models.VariableDeletionBlockedError
	assert.True(t, errors.As(err, &blockedErr))
	assert.Equal(t, []string{"V2"}, blockedErr.BlockedVariables)
	assert.Len(t, blockedErr.Usages, 1)
	assert.Contains(t, blockedErr.Error(), "V2")
	assert.Contains(t, blockedErr.Error(), "currently in use")
}

func TestDeletionGuard_MixedRemoval_OnlyUsedOnesBlock(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1", "V2", "V3")}
	impl, mappingService := guardTestService(t, repoStub)
	// V2 and V3 are removed; only V3 is in use
	mappingService.On("GetLiveVariableUsage", mock.MatchedBy(func(names []string) bool {
		return len(names) == 2
	})).Return([]*models.VariableUsage{usageFor("V3"), usageFor("V3")}, nil)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("V1"))

	var blockedErr *models.VariableDeletionBlockedError
	assert.True(t, errors.As(err, &blockedErr))
	assert.Equal(t, []string{"V3"}, blockedErr.BlockedVariables)
}

func TestDeletionGuard_EmptyPayloadChecksAllVariables(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1", "V2")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", mock.MatchedBy(func(names []string) bool {
		return len(names) == 2
	})).Return([]*models.VariableUsage{usageFor("V2"), usageFor("V1")}, nil)

	err := impl.checkVariablesToBeDeletedForUsage(models.Payload{UserId: 1})

	var blockedErr *models.VariableDeletionBlockedError
	assert.True(t, errors.As(err, &blockedErr))
	// blocked variables are reported sorted
	assert.Equal(t, []string{"V1", "V2"}, blockedErr.BlockedVariables)
}

func TestDeletionGuard_RenameIsTreatedAsDeletion(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("DB_HOST")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", []string{"DB_HOST"}).
		Return([]*models.VariableUsage{usageFor("DB_HOST")}, nil)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("DB_HOST_NEW"))

	var blockedErr *models.VariableDeletionBlockedError
	assert.True(t, errors.As(err, &blockedErr))
	assert.Equal(t, []string{"DB_HOST"}, blockedErr.BlockedVariables)
}

func TestDeletionGuard_NamesAreCaseSensitive(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("DB_HOST")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", []string{"DB_HOST"}).
		Return([]*models.VariableUsage{usageFor("DB_HOST")}, nil)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("db_host"))

	var blockedErr *models.VariableDeletionBlockedError
	assert.True(t, errors.As(err, &blockedErr))
}

func TestDeletionGuard_MetadataFetchErrorPropagates(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadataErr: errors.New("db down")}
	impl, _ := guardTestService(t, repoStub)

	err := impl.checkVariablesToBeDeletedForUsage(payloadFor("V1"))

	assert.NotNil(t, err)
	assert.False(t, repoStub.deleteCalled)
}

func TestDeletionGuard_UsageLookupErrorPropagates(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", []string{"V1"}).Return(nil, errors.New("db down"))

	err := impl.checkVariablesToBeDeletedForUsage(models.Payload{UserId: 1})

	assert.NotNil(t, err)
	var blockedErr *models.VariableDeletionBlockedError
	assert.False(t, errors.As(err, &blockedErr))
}

// CreateVariables must reject the whole request before starting the delete-and-recreate
// transaction when a removed variable is still in use.
func TestCreateVariables_BlockedBeforeAnythingIsDeleted(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1", "V2")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", []string{"V1"}).
		Return([]*models.VariableUsage{usageFor("V1")}, nil)

	err := impl.CreateVariables(payloadFor("V2"))

	var blockedErr *models.VariableDeletionBlockedError
	assert.True(t, errors.As(err, &blockedErr))
	assert.False(t, repoStub.txStartCalled, "transaction must not start when deletion is blocked")
	assert.False(t, repoStub.deleteCalled, "no variable may be deleted when deletion is blocked")
}

func TestCreateVariables_PayloadValidationRunsBeforeUsageCheck(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1")}
	impl, mappingService := guardTestService(t, repoStub)

	// duplicate variable names fail payload validation before the usage check
	err := impl.CreateVariables(payloadFor("V2", "V2"))

	assert.NotNil(t, err)
	assert.True(t, errors.As(err, &models.ValidationError{}))
	mappingService.AssertNotCalled(t, "GetLiveVariableUsage", mock.Anything)
}

func TestGetVariableUsage_ForSingleVariable(t *testing.T) {
	repoStub := &scopedVariableRepoStub{}
	impl, mappingService := guardTestService(t, repoStub)
	expected := []*models.VariableUsage{usageFor("V1")}
	mappingService.On("GetLiveVariableUsage", []string{"V1"}).Return(expected, nil)

	usages, err := impl.GetVariableUsage("V1")

	assert.Nil(t, err)
	assert.Equal(t, expected, usages)
}

func TestGetVariableUsage_AllVariablesWhenNameEmpty(t *testing.T) {
	repoStub := &scopedVariableRepoStub{metadata: defsFor("V1", "V2")}
	impl, mappingService := guardTestService(t, repoStub)
	mappingService.On("GetLiveVariableUsage", []string{"V1", "V2"}).
		Return([]*models.VariableUsage{usageFor("V1")}, nil)

	usages, err := impl.GetVariableUsage("")

	assert.Nil(t, err)
	assert.Len(t, usages, 1)
}

func TestGetLiveVariableUsage_MapsEntityTypesAndDedupes(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.Nil(t, err)
	repoMock := mocks2.NewVariableEntityMappingRepository(t)
	service := NewVariableEntityMappingServiceImpl(repoMock, logger.Sugar())

	rows := []*repository2.VariableUsageRow{
		{VariableName: "V1", EntityType: repository2.EntityTypeDeploymentTemplateAppLevel, EntityId: 10, AppId: 1, AppName: "app-1"},
		// duplicate mapping row for the same entity must be deduped
		{VariableName: "V1", EntityType: repository2.EntityTypeDeploymentTemplateAppLevel, EntityId: 10, AppId: 1, AppName: "app-1"},
		{VariableName: "V1", EntityType: repository2.EntityTypeSecretEnvLevel, EntityId: 20, AppId: 1, AppName: "app-1", EnvId: 3, EnvName: "env-3"},
		{VariableName: "V2", EntityType: repository2.EntityTypeConfigMapAppLevel, EntityId: 30, AppId: 2, AppName: "app-2"},
		{VariableName: "V2", EntityType: repository2.EntityTypePipelineStage, EntityId: 40, AppId: 2, AppName: "app-2", PipelineName: "ci-pipeline", StageType: "PRE_CI"},
		// two different variables on the same entity must both be reported
		{VariableName: "V3", EntityType: repository2.EntityTypeDeploymentTemplateAppLevel, EntityId: 10, AppId: 1, AppName: "app-1"},
	}
	repoMock.On("GetLiveUsagesForVariableNames", []string{"V1", "V2", "V3"}).Return(rows, nil)

	usages, err := service.GetLiveVariableUsage([]string{"V1", "V2", "V3"})

	assert.Nil(t, err)
	assert.Len(t, usages, 5)
	usageTypeByKey := make(map[string]models.VariableUsageType)
	for _, usage := range usages {
		usageTypeByKey[usage.VariableName+"/"+string(usage.UsageType)] = usage.UsageType
	}
	assert.Equal(t, models.UsageTypeDeploymentTemplate, usageTypeByKey["V1/deploymentTemplate"])
	assert.Equal(t, models.UsageTypeSecret, usageTypeByKey["V1/secret"])
	assert.Equal(t, models.UsageTypeConfigMap, usageTypeByKey["V2/configMap"])
	assert.Equal(t, models.UsageTypePipelineStage, usageTypeByKey["V2/pipelineStage"])
	assert.Equal(t, models.UsageTypeDeploymentTemplate, usageTypeByKey["V3/deploymentTemplate"])
}

func TestGetLiveVariableUsage_EmptyInputSkipsRepository(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.Nil(t, err)
	repoMock := mocks2.NewVariableEntityMappingRepository(t)
	service := NewVariableEntityMappingServiceImpl(repoMock, logger.Sugar())

	usages, err := service.GetLiveVariableUsage(nil)

	assert.Nil(t, err)
	assert.Empty(t, usages)
	repoMock.AssertNotCalled(t, "GetLiveUsagesForVariableNames", mock.Anything)
}
