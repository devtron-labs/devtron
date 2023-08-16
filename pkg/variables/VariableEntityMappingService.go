package variables

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
	"time"
)

type VariableEntityMappingService interface {
	UpdateVariablesForEntity(variableIds []int, entity repository.Entity, userId int32) error
	GetAllMappingsForEntities(entities []repository.Entity) (map[string]int, error)
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

func (impl VariableEntityMappingServiceImpl) UpdateVariablesForEntity(variableIds []int, entity repository.Entity, userId int32) error {

	variableMappings, err := impl.variableEntityMappingRepository.GetVariablesForEntities([]repository.Entity{entity})

	existingVarIds := make([]int, 0)
	for _, mapping := range variableMappings {
		existingVarIds = append(existingVarIds, mapping.VariableId)
	}

	existingVarSet := mapset.NewSetFromSlice(ToInterfaceArray(existingVarIds))
	newVarSet := mapset.NewSetFromSlice(ToInterfaceArray(variableIds))

	// If present in existing but not in new, then delete
	variablesToDelete := existingVarSet.Difference(newVarSet).ToSlice()
	// If present in new but not in existing then add
	variableToAdd := newVarSet.Difference(existingVarSet).ToSlice()

	newVariableMappings := make([]*repository.VariableEntityMapping, 0)
	for _, variableId := range variableToAdd {
		variableMappings = append(variableMappings, &repository.VariableEntityMapping{
			VariableId: variableId.(int),
			Entity:     entity,
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
	err = impl.variableEntityMappingRepository.DeleteVariablesForEntity(tx, ToIntArray(variablesToDelete), entity, userId)
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

func (impl VariableEntityMappingServiceImpl) GetAllMappingsForEntities(entities []repository.Entity) (map[string]int, error) {
	variableEntityMappings, err := impl.variableEntityMappingRepository.GetVariablesForEntities(entities)
	if err != nil {
		return nil, err
	}
	entityIdToVariableIds := make(map[string]int)
	for _, mapping := range variableEntityMappings {
		entityIdToVariableIds[mapping.EntityId] = mapping.VariableId
	}
	return entityIdToVariableIds, nil
}

func (impl VariableEntityMappingServiceImpl) DeleteMappingsForEntities(entities []repository.Entity, userId int32) error {
	err := impl.variableEntityMappingRepository.DeleteAllVariablesForEntities(entities, userId)
	if err != nil {
		return err
	}
	return nil
}

// ToInterfaceArray converts an array of int to an array of interface{}
func ToInterfaceArray(arr []int) []interface{} {
	interfaceArr := make([]interface{}, len(arr))
	for i, v := range arr {
		interfaceArr[i] = v
	}
	return interfaceArr
}

// ToIntArray converts an array of interface{} back to an array of int
func ToIntArray(interfaceArr []interface{}) []int {
	intArr := make([]int, len(interfaceArr))
	for i, v := range interfaceArr {
		intArr[i] = v.(int)
	}
	return intArr
}
