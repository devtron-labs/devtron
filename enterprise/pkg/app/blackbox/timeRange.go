package blackbox

import (
	"time"
)

//type GlobalPolicyDataManager interface {
//	// fetch data only from GlobalPolicyRepository
//	GetPolicyById(policyId int) (*GlobalPolicyBaseModel, error)
//	GetAllActiveByType() ([]*GlobalPolicyBaseModel, error)
//	GetPolicyByName(policyName string) (*GlobalPolicyBaseModel, error)
//	GetPolicyByIds(policyIds []int) ([]*GlobalPolicyBaseModel, error)
//
//	// save data using both GlobalPolicyRepository & GlobalPolicySearchableFieldRepository
//	// but perform operation in single Tx
//	CreatePolicy(tx *pg.Tx, globalPolicyDataModel *GlobalPolicyDataModel) (*GlobalPolicyDataModel, error)
//	UpdatePolicy(tx *pg.Tx, globalPolicyDataModel *GlobalPolicyDataModel) (*GlobalPolicyDataModel, error)
//	DeletePolicyById(tx *pg.Tx, policyId int) error
//
//	// fetch data only from GlobalPolicySearchableFieldRepository
//	GetPolicyMetadataByFields(policyIds []int, fields []*SearchableField) (map[int][]*SearchableField, error)
//	//
//	GetPoliciesBySearchableFields(policyIds []int, fields []*SearchableField) ([]*GlobalPolicyBaseModel, error)
//}
//
//type GlobalPolicyBaseModel struct {
//	Id            int
//	Name          string
//	Description   string
//	Enabled       bool
//	PolicyOf      bean.GlobalPolicyType
//	PolicyVersion bean.GlobalPolicyVersion
//	JsonData      string
//	Active        bool
//	UserId        int32
//}
//
//type GlobalPolicyDataModel struct {
//	GlobalPolicyBaseModel
//	SearchableFields []SearchableField
//}
//
//type SearchableField struct {
//	FieldName  string
//	FieldValue interface{}
//	FieldType  FieldType
//}
//type FieldType int
//
//const NumericType FieldType = 1
//const StringType FieldType = 2
//const DateTimeType FieldType = 3
//const BooleanType FieldType = 4

type TimeRange struct {
	TimeFrom       time.Time
	TimeTo         time.Time
	HourMinuteFrom string
	HourMinuteTo   string
	DayFrom        int
	DayTo          int
	WeekdayFrom    time.Weekday
	WeekdayTo      time.Weekday
	Weekdays       []time.Weekday
	Frequency      Frequency
}

func (timeRange TimeRange) GetScheduleSpec(targetTime time.Time) (time.Time, bool) {
	return time.Time{}, true
}

type Frequency string

const (
	FIXED        Frequency = "FIXED"
	DAILY        Frequency = "DAILY"
	WEEKLY       Frequency = "WEEKLY"
	WEEKLY_RANGE Frequency = "WEEKLY_RANGE"
	MONTHLY      Frequency = "MONTHLY"
)
