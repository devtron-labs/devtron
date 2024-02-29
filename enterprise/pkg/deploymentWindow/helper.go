package deploymentWindow

import (
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"reflect"
	"time"
)

const searchFieldTypeTag = "isSearchField"

func GetSearchableFields[T any](profile T) []bean.SearchableField {
	var fields []bean.SearchableField

	val := reflect.ValueOf(profile)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get searchFieldType tag value
		tag := fieldType.Tag.Get(searchFieldTypeTag)
		// If the tag is "-" or empty, skip this field
		if tag == "" || tag == "false" {
			continue
		}

		// Determine FieldType from tag value
		//fieldTypeEnum := determineFieldType(tag)

		fields = append(fields, bean.SearchableField{
			FieldName:  fieldType.Name,
			FieldValue: field.Interface(),
			FieldType:  determineFieldType(field.Interface()),
		})
	}

	return fields
}

func determineFieldType(field interface{}) bean.FieldType {

	_, ok1 := field.(date.Time)
	_, ok2 := field.(time.Time)
	if ok1 || ok2 {
		return bean.DateTimeType
	}

	switch reflect.TypeOf(field).Kind() {
	case reflect.String:
		return bean.StringType
	case reflect.Int:
		return bean.NumericType
	case reflect.Bool:
		return bean.BooleanType
	default:
		return bean.StringType
	}
}
