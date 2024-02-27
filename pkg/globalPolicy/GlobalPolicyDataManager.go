package globalPolicy

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GlobalPolicyDataManager interface {
	CreatePolicy(globalPolicyDataModel *bean.GlobalPolicyDataModel, tx *pg.Tx) (*bean.GlobalPolicyDataModel, error)

	GetPolicyById(policyId int) (*bean.GlobalPolicyBaseModel, error)
	GetPolicyByName(policyName string) (*bean.GlobalPolicyBaseModel, error)
	GetPolicyByIds(policyIds []int) ([]*bean.GlobalPolicyBaseModel, error)
	GetAllActiveByType(policyType *bean.GlobalPolicyType) (*bean.GlobalPolicyBaseModel, error)

	UpdatePolicy(globalPolicyDataModel *bean.GlobalPolicyDataModel, tx *pg.Tx) (*bean.GlobalPolicyDataModel, error)
	//UpdatePolicyByName(PolicyName string, globalPolicyDataModel *bean.GlobalPolicyDataModel) (*bean.GlobalPolicyDataModel, error)

	DeletePolicyById(policyId int, userId int32) error
	DeletePolicyByName(policyName string, userId int32) error

	GetPolicyMetadataByFields(policyIds []int, fields []*bean.SearchableField) (map[int][]*bean.SearchableField, error)
	//GetPoliciesBySearchableFields(policyIds []int,fields []*SearchableField) ([]*GlobalPolicyBaseModel, error)
}

type GlobalPolicyDataManagerImpl struct {
	logger                                *zap.SugaredLogger
	globalPolicyRepository                repository.GlobalPolicyRepository
	globalPolicySearchableFieldRepository repository.GlobalPolicySearchableFieldRepository
}

func NewGlobalPolicyDataManagerImpl(logger *zap.SugaredLogger, globalPolicyRepository repository.GlobalPolicyRepository,
	globalPolicySearchableFieldRepository repository.GlobalPolicySearchableFieldRepository,
) *GlobalPolicyDataManagerImpl {
	return &GlobalPolicyDataManagerImpl{
		logger:                                logger,
		globalPolicyRepository:                globalPolicyRepository,
		globalPolicySearchableFieldRepository: globalPolicySearchableFieldRepository,
	}
}

func (impl *GlobalPolicyDataManagerImpl) CreatePolicy(globalPolicyDataModel *bean.GlobalPolicyDataModel, tx *pg.Tx) (*bean.GlobalPolicyDataModel, error) {
	var err error
	if tx == nil {
		tx, err = impl.globalPolicyRepository.GetDbTransaction()
		if err != nil {
			impl.logger.Errorw("error in initiating transaction", "err", err)
			return nil, err
		}
	}
	// Rollback tx on error.
	defer func() {
		err = impl.globalPolicyRepository.RollBackTransaction(tx)
		if err != nil {
			impl.logger.Errorw("error in rolling back transaction", "err", err)
		}
	}()
	globalPolicy := impl.getGlobalPolicyDto(globalPolicyDataModel)
	err = impl.globalPolicyRepository.Create(globalPolicy, tx)
	if err != nil {
		impl.logger.Errorw("error, CreatePolicy", "err", err, "globalPolicy", globalPolicy)
	}
	searchableKeyEntriesTotal := impl.getSearchableKeyEntries(globalPolicyDataModel)
	err = impl.globalPolicySearchableFieldRepository.CreateInBatchWithTxn(searchableKeyEntriesTotal, tx)
	if err != nil {
		impl.logger.Errorw("error in creating global policy searchable fields entry", "err", err, "searchableKeyEntriesTotal", searchableKeyEntriesTotal)
		// TODO KB: Why are not we returning from here ??
	}
	err = impl.globalPolicyRepository.CommitTransaction(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return globalPolicyDataModel, err
	}
	globalPolicyDataModel.Id = globalPolicy.Id
	return globalPolicyDataModel, nil
}
func (impl *GlobalPolicyDataManagerImpl) UpdatePolicy(globalPolicyDataModel *bean.GlobalPolicyDataModel, tx *pg.Tx) (*bean.GlobalPolicyDataModel, error) {
	var err error
	if tx != nil {
		tx, err = impl.globalPolicyRepository.GetDbTransaction()
		if err != nil {
			impl.logger.Errorw("error in initiating transaction", "err", err)
			return nil, err
		}
	}
	// Rollback tx on error.
	defer func() {
		err = impl.globalPolicyRepository.RollBackTransaction(tx)
		if err != nil {
			impl.logger.Errorw("error in rolling back transaction", "err", err)
		}
	}()
	globalPolicy := impl.getGlobalPolicyDto(globalPolicyDataModel)
	globalPolicy.Id = globalPolicyDataModel.Id
	err = impl.globalPolicyRepository.Update(globalPolicy, tx)
	if err != nil {
		impl.logger.Errorw("error, UpdatePolicy", "err", err, "globalPolicy", globalPolicy)
	}
	err = impl.globalPolicySearchableFieldRepository.DeleteByPolicyId(globalPolicy.Id, tx)
	if err != nil {
		impl.logger.Errorw("error in  deleting Policy Searchable key", "globalPolicyDataModel", globalPolicyDataModel, "err", err)
	}
	searchableKeyEntriesTotal := impl.getSearchableKeyEntries(globalPolicyDataModel)
	err = impl.globalPolicySearchableFieldRepository.CreateInBatchWithTxn(searchableKeyEntriesTotal, tx)
	if err != nil {
		impl.logger.Errorw("error in creating global policy searchable fields entry", "err", err, "searchableKeyEntriesTotal", searchableKeyEntriesTotal)
	}
	err = impl.globalPolicyRepository.CommitTransaction(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return globalPolicyDataModel, err
	}
	globalPolicyDataModel.Id = globalPolicy.Id
	return globalPolicyDataModel, nil
}

func (impl *GlobalPolicyDataManagerImpl) getSearchableKeyEntries(globalPolicyDataModel *bean.GlobalPolicyDataModel) []*repository.GlobalPolicySearchableField {
	searchableKeyEntriesTotal := make([]*repository.GlobalPolicySearchableField, 0)
	for _, field := range globalPolicyDataModel.SearchableFields {
		searchableKeyEntries := &repository.GlobalPolicySearchableField{
			GlobalPolicyId: globalPolicyDataModel.Id,
			IsRegex:        false,
			FieldName:      field.FieldName,
		}
		switch field.FieldType {
		case bean.NumericType:
			searchableKeyEntries.ValueInt = field.FieldValue.(int)
		case bean.StringType:
			searchableKeyEntries.Value = field.FieldValue.(string)
		case bean.DateTimeType:
			searchableKeyEntries.ValueTimeStamp = field.FieldValue.(time.Time)
		}

		searchableKeyEntries.CreateAuditLog(globalPolicyDataModel.UserId)
		searchableKeyEntriesTotal = append(searchableKeyEntriesTotal, searchableKeyEntries)
	}
	return searchableKeyEntriesTotal
}
func (impl *GlobalPolicyDataManagerImpl) getGlobalPolicyDto(globalPolicyDataModel *bean.GlobalPolicyDataModel) *repository.GlobalPolicy {
	globalPolicy := &repository.GlobalPolicy{
		Name:        globalPolicyDataModel.Name,
		PolicyOf:    string(globalPolicyDataModel.PolicyOf),
		Version:     string(bean.GLOBAL_POLICY_VERSION_V1),
		Description: globalPolicyDataModel.Description,
		PolicyJson:  globalPolicyDataModel.JsonData,
		Enabled:     true,
		Deleted:     false,
	}
	globalPolicy.CreateAuditLog(globalPolicyDataModel.UserId)
	return globalPolicy
}

func (impl *GlobalPolicyDataManagerImpl) GetPolicyByName(policyName string) (*bean.GlobalPolicyBaseModel, error) {
	globalPolicy, err := impl.globalPolicyRepository.GetByName(policyName)
	if err != nil {
		impl.logger.Errorw("error in fetching global policy", "policyName", policyName, "err", err)
		return nil, err
	}
	return globalPolicy.GetGlobalPolicyBaseModel(), nil
}

func (impl *GlobalPolicyDataManagerImpl) GetPolicyMetadataByFields(policyIds []int, fields []*bean.SearchableField) (map[int][]*bean.SearchableField, error) {
	var policyIdToSearchableField map[int][]*bean.SearchableField
	GlobalPolicySearchableFields, err := impl.globalPolicySearchableFieldRepository.GetSearchableFieldByIds(policyIds)
	if err != nil {
		impl.logger.Errorw("error in fetching GlobalPolicySearchableFields", "err", err)
		return nil, err
	}
	fieldNames := make(map[string]bool)
	for _, field := range fields {
		fieldNames[field.FieldName] = true
	}
	for _, searchableField := range GlobalPolicySearchableFields {
		if _, ok := fieldNames[searchableField.FieldName]; ok {
			fieldValue, fieldType := impl.setFieldValueAndType(searchableField)
			policyIdToSearchableField = impl.setPolicyIdToSearchableFieldMap(searchableField, fieldType, fieldValue)
		}
	}
	return policyIdToSearchableField, nil
}

func (impl *GlobalPolicyDataManagerImpl) setPolicyIdToSearchableFieldMap(searchableField *repository.GlobalPolicySearchableField, fieldType bean.FieldType, fieldValue interface{}) map[int][]*bean.SearchableField {
	policyIdToSearchableField := make(map[int][]*bean.SearchableField, 0)
	if policyIdToSearchableField[searchableField.GlobalPolicyId] != nil {
		policyIdToSearchableField[searchableField.GlobalPolicyId] = append(policyIdToSearchableField[searchableField.GlobalPolicyId], &bean.SearchableField{
			FieldName:  searchableField.FieldName,
			FieldType:  fieldType,
			FieldValue: fieldValue,
		})
	} else {
		policyIdToSearchableField[searchableField.GlobalPolicyId] = []*bean.SearchableField{
			{
				FieldName:  searchableField.FieldName,
				FieldType:  fieldType,
				FieldValue: fieldValue,
			},
		}
	}
	return policyIdToSearchableField
}

func (impl *GlobalPolicyDataManagerImpl) setFieldValueAndType(searchableField *repository.GlobalPolicySearchableField) (interface{}, bean.FieldType) {
	var fieldValue interface{}
	var fieldType bean.FieldType
	if searchableField.Value != "" {
		fieldValue = searchableField.Value
		fieldType = bean.StringType
	} else if searchableField.ValueInt != 0 {
		fieldValue = searchableField.ValueInt
		fieldType = bean.NumericType
	} else if !searchableField.ValueTimeStamp.IsZero() {
		fieldValue = searchableField.ValueTimeStamp
		fieldType = bean.DateTimeType

	}
	return fieldValue, fieldType
}

func (impl *GlobalPolicyDataManagerImpl) GetPolicyById(policyId int) (*bean.GlobalPolicyBaseModel, error) {
	globalPolicy, err := impl.globalPolicyRepository.GetById(policyId)
	if err != nil {
		impl.logger.Errorw("error in fetching global policy", "policyId", policyId, "err", err)
		return nil, err
	}
	return globalPolicy.GetGlobalPolicyBaseModel(), nil
}

func (impl *GlobalPolicyDataManagerImpl) GetPolicyByIds(policyIds []int) ([]*bean.GlobalPolicyBaseModel, error) {
	GlobalPolicyBaseModels := make([]*bean.GlobalPolicyBaseModel, 0)
	if len(policyIds) == 0 {
		return GlobalPolicyBaseModels, nil
	}
	globalPolicies, err := impl.globalPolicyRepository.GetByIds(policyIds)
	if err != nil {
		impl.logger.Errorw("error in fetching global policy", "policyIds", policyIds, "err", err)
		return nil, err
	}
	for _, policy := range globalPolicies {
		GlobalPolicyBaseModels = append(GlobalPolicyBaseModels, policy.GetGlobalPolicyBaseModel())
	}
	return GlobalPolicyBaseModels, nil
}
func (impl *GlobalPolicyDataManagerImpl) GetAllActiveByType(policyType *bean.GlobalPolicyType) (*bean.GlobalPolicyBaseModel, error) {
	globalPolicy, err := impl.globalPolicyRepository.GetPolicyByType(policyType)
	if err != nil {
		impl.logger.Errorw("error in fetching global policy", "policyType", policyType, "err", err)
		return nil, err
	}
	return globalPolicy.GetGlobalPolicyBaseModel(), nil
}
func (impl *GlobalPolicyDataManagerImpl) DeletePolicyById(policyId int, userId int32) error {
	err := impl.globalPolicyRepository.DeletedById(policyId, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting policies", "err", err, "policyId", policyId)
	}
	return err
}
func (impl *GlobalPolicyDataManagerImpl) DeletePolicyByName(policyName string, userId int32) error {
	err := impl.globalPolicyRepository.DeletedByName(policyName, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting policies", "err", err, "policyName", policyName)
	}
	return err

}
