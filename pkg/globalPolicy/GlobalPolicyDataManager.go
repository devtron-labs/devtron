package globalPolicy

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
)

type GlobalPolicyBaseModel struct {
	Id            int
	Name          string
	Description   string
	Enabled       bool
	PolicyOf      bean.GlobalPolicyType
	PolicyVersion bean.GlobalPolicyVersion
	JsonData      string
	Active        bool
	UserId        int32
}

type GlobalPolicyDataModel struct {
	GlobalPolicyBaseModel
	SearchableFields []SearchableField
}

type SearchableField struct {
	FieldName  string
	FieldValue interface{}
	FieldType  FieldType
}

type FieldType int

const NumericType FieldType = 1
const StringType FieldType = 2
const DateTimeType FieldType = 3

type GlobalPolicyDataManager interface {
	// fetch data only from GlobalPolicyRepository
	GetPolicyById(policyId int) (*GlobalPolicyBaseModel, error)
	GetPolicyByName(policyName string) (*GlobalPolicyBaseModel, error)
	GetPolicyByIds(policyIds []int) ([]*GlobalPolicyBaseModel, error)

	// save data using both GlobalPolicyRepository & GlobalPolicySearchableFieldRepository
	// but perform operation in single Tx
	CreatePolicy(globalPolicyDataModel *GlobalPolicyDataModel) error
	UpdatePolicy(globalPolicyDataModel *GlobalPolicyDataModel) error
	DeletePolicyById(policyId int) error

	// fetch data only from GlobalPolicySearchableFieldRepository
	GetPolicyMetadataByFields(fields []*SearchableField) (map[int][]*SearchableField, error)

	GetPoliciesBySearchableFields(fields []*SearchableField) ([]*GlobalPolicyBaseModel, error)
}
