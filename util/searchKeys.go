package util

import (
	"github.com/Azure/go-autorest/autorest/date"
	"reflect"
	"time"
)

type SearchableField struct {
	FieldName  string
	FieldValue interface{}
	FieldType  FieldType
}
type FieldType int

const NumericType FieldType = 1
const StringType FieldType = 2
const DateTimeType FieldType = 3
const BooleanType FieldType = 4

const searchFieldTypeTag = "isSearchField"

func GetSearchableFields[T interface{}](profile T) []SearchableField {
	var fields []SearchableField

	val := reflect.ValueOf(profile)
	typ := reflect.TypeOf(profile)
	kind := typ.Kind()
	if kind != reflect.Struct {
		return nil
	}
	count := typ.NumField()
	for i := 0; i < count; i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get searchFieldType tag value
		tag := fieldType.Tag.Get(searchFieldTypeTag)
		// If the tag is "-" or empty, skip this field
		if tag == "" || tag == "false" {
			continue
		}

		// Determine FieldType from tag value
		// fieldTypeEnum := determineFieldType(tag)

		fields = append(fields, SearchableField{
			FieldName:  fieldType.Name,
			FieldValue: field.Interface(),
			FieldType:  determineFieldType(field.Interface()),
		})
	}

	return fields
}

func determineFieldType(field interface{}) FieldType {

	_, ok1 := field.(date.Time)
	_, ok2 := field.(time.Time)
	if ok1 || ok2 {
		return DateTimeType
	}

	switch reflect.TypeOf(field).Kind() {
	case reflect.String:
		return StringType
	case reflect.Int:
		return NumericType
	case reflect.Bool:
		return BooleanType
	default:
		return StringType
	}
}
