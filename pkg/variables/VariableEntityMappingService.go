package variables

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
	"time"
)

type VariableEntityMappingService interface {
	UpdateVariablesForEntity(variableNames []string, entity repository.Entity, userId int32) error
	GetAllMappingsForEntities(entities []repository.Entity) (map[int]string, error)
	DeleteMappingsForEntities(entities []repository.Entity, userId int32) error
}

type VariableEntityMappingServiceImpl struct {
	logger                          *zap.SugaredLogger
	variableEntityMappingRepository repository.VariableEntityMappingRepository
}

func NewVariableEntityMappingServiceImpl(variableEntityMappingRepository repository.VariableEntityMappingRepository, logger *zap.SugaredLogger) *VariableEntityMappingServiceImpl {
	return &VariableEntityMappingServiceImpl{
		variableEntityMappingRepository: variableEntityMappingRepository,
		logger:                          logger,
	}
}

func (impl VariableEntityMappingServiceImpl) UpdateVariablesForEntity(variableNames []string, entity repository.Entity, userId int32) error {

	variableMappings, err := impl.variableEntityMappingRepository.GetVariablesForEntities([]repository.Entity{entity})

	existingVarNames := make([]string, 0)
	for _, mapping := range variableMappings {
		existingVarNames = append(existingVarNames, mapping.VariableName)
	}

	existingVarSet := mapset.NewSetFromSlice(ToInterfaceArray(existingVarNames))
	newVarSet := mapset.NewSetFromSlice(ToInterfaceArray(variableNames))

	// If present in existing but not in new, then delete
	variablesToDelete := existingVarSet.Difference(newVarSet).ToSlice()
	// If present in new but not in existing then add
	variableToAdd := newVarSet.Difference(existingVarSet).ToSlice()

	newVariableMappings := make([]*repository.VariableEntityMapping, 0)
	for _, variableId := range variableToAdd {
		variableMappings = append(variableMappings, &repository.VariableEntityMapping{
			VariableName: variableId.(string),
			Entity:       entity,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		})
	}

	connection := impl.variableEntityMappingRepository.GetConnection()
	tx, err := connection.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err != nil {
		return err
	}
	err = impl.variableEntityMappingRepository.DeleteVariablesForEntity(tx, ToStringArray(variablesToDelete), entity, userId)
	if err != nil {
		return err
	}

	err = impl.variableEntityMappingRepository.SaveVariableEntityMappings(tx, newVariableMappings)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl VariableEntityMappingServiceImpl) GetAllMappingsForEntities(entities []repository.Entity) (map[int]string, error) {
	variableEntityMappings, err := impl.variableEntityMappingRepository.GetVariablesForEntities(entities)
	if err != nil {
		return nil, err
	}
	entityIdToVariableNames := make(map[int]string)
	for _, mapping := range variableEntityMappings {
		entityIdToVariableNames[mapping.EntityId] = mapping.VariableName
	}
	return entityIdToVariableNames, nil
}

func (impl VariableEntityMappingServiceImpl) DeleteMappingsForEntities(entities []repository.Entity, userId int32) error {
	err := impl.variableEntityMappingRepository.DeleteAllVariablesForEntities(entities, userId)
	if err != nil {
		return err
	}
	return nil
}

// ToInterfaceArray converts an array of string to an array of interface{}
func ToInterfaceArray(arr []string) []interface{} {
	interfaceArr := make([]interface{}, len(arr))
	for i, v := range arr {
		interfaceArr[i] = v
	}
	return interfaceArr
}

// ToStringArray converts an array of interface{} back to an array of string
func ToStringArray(interfaceArr []interface{}) []string {
	stringArr := make([]string, len(interfaceArr))
	for i, v := range interfaceArr {
		stringArr[i] = v.(string)
	}
	return stringArr
}
