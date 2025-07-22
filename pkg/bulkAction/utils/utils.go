package utils

import (
	"fmt"
)

// GenerateIdentifierKey returns the appKey and pipelineKey for a given pipeline.
// It assumes that the Pipeline struct has the fields AppId, Id, and Name,
// and that it contains an embedded App struct with an AppName field.
func GenerateIdentifierKey(id int, name string) (appKey string) {
	appKey = fmt.Sprintf("%d_%s", id, name)
	return appKey
}
