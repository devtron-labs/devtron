package variables

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type VariableEntityMappingService interface {
	UpdateVariablesForEntity(variableNames []string, entity repository.Entity, userId int32, tx *pg.Tx) error
	GetAllMappingsForEntities(entities []repository.Entity) (map[repository.Entity][]string, error)
	DeleteMappingsForEntities(entities []repository.Entity, userId int32, tx *pg.Tx) error
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

func (impl VariableEntityMappingServiceImpl) UpdateVariablesForEntity(variableNames []string, entity repository.Entity, userId int32, passedTx *pg.Tx) error {

	variableMappings, err := impl.variableEntityMappingRepository.GetVariablesForEntities([]repository.Entity{entity})

	existingVarNames := make([]string, 0)
	for _, mapping := range variableMappings {
		existingVarNames = append(existingVarNames, mapping.VariableName)
	}

	existingVarSet := mapset.NewSetFromSlice(utils.ToInterfaceArray(existingVarNames))
	newVarSet := mapset.NewSetFromSlice(utils.ToInterfaceArray(variableNames))

	// If present in existing but not in new, then delete
	variablesToDelete := existingVarSet.Difference(newVarSet).ToSlice()
	// If present in new but not in existing then add
	variableToAdd := newVarSet.Difference(existingVarSet).ToSlice()

	newVariableMappings := make([]*repository.VariableEntityMapping, 0)
	for _, variableId := range variableToAdd {
		newVariableMappings = append(newVariableMappings, &repository.VariableEntityMapping{
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

	var tx *pg.Tx
	if passedTx == nil {
		tx, err = impl.variableEntityMappingRepository.StartTx()
		defer func() {
			err = impl.variableEntityMappingRepository.RollbackTx(tx)
			if err != nil {
				impl.logger.Infow("error in rolling back transaction", "err", err)
			}
		}()
	} else {
		tx = passedTx
	}

	if len(variablesToDelete) > 0 {
		err = impl.variableEntityMappingRepository.DeleteVariablesForEntity(tx, utils.ToStringArray(variablesToDelete), entity, userId)
		if err != nil {
			return err
		}
	}

	if len(newVariableMappings) > 0 {
		err = impl.variableEntityMappingRepository.SaveVariableEntityMappings(tx, newVariableMappings)
		if err != nil {
			return err
		}
	}

	if passedTx == nil {
		err = impl.variableEntityMappingRepository.CommitTx(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl VariableEntityMappingServiceImpl) GetAllMappingsForEntities(entities []repository.Entity) (map[repository.Entity][]string, error) {
	variableEntityMappings, err := impl.variableEntityMappingRepository.GetVariablesForEntities(entities)
	if err != nil {
		return nil, err
	}
	entityIdToVariableNames := make(map[repository.Entity][]string)
	for _, mapping := range variableEntityMappings {
		vars := entityIdToVariableNames[mapping.Entity]
		vars = append(vars, mapping.VariableName)
		entityIdToVariableNames[mapping.Entity] = vars
	}
	return entityIdToVariableNames, nil
}

func (impl VariableEntityMappingServiceImpl) DeleteMappingsForEntities(entities []repository.Entity, userId int32, tx *pg.Tx) error {
	err := impl.variableEntityMappingRepository.DeleteAllVariablesForEntities(tx, entities, userId)
	if err != nil {
		return err
	}
	return nil
}
