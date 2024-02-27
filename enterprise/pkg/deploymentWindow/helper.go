package deploymentWindow

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/app/blackbox"
	"reflect"
)

const searchFieldTypeTag = "searchFieldType"

func GetSearchableFields(profile DeploymentWindowProfile) []blackbox.SearchableField {
	var fields []blackbox.SearchableField

	val := reflect.ValueOf(profile)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get searchFieldType tag value
		tag := fieldType.Tag.Get(searchFieldTypeTag)
		// If the tag is "-" or empty, skip this field
		if tag == "" {
			continue
		}

		// Determine FieldType from tag value
		fieldTypeEnum := determineFieldType(tag)

		fields = append(fields, blackbox.SearchableField{
			FieldName:  fieldType.Name,
			FieldValue: field.Interface(),
			FieldType:  fieldTypeEnum,
		})
	}

	return fields
}

func determineFieldType(fieldTypeStr string) blackbox.FieldType {
	switch fieldTypeStr {
	case "string":
		return blackbox.StringType
	case "numeric":
		return blackbox.NumericType
	case "datetime":
		return blackbox.DateTimeType
	case "boolean":
		return blackbox.BooleanType
	default:
		return blackbox.StringType
	}
}
