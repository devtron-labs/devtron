package util

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
)

func GetSchemaFromType[T any](t T) (string, error) {
	schema := jsonschema.Reflect(t)
	schemaData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(schemaData), nil
}
