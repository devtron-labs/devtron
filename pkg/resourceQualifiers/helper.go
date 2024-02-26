package resourceQualifiers

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
)

func GetQualifierMappingsForCompoundQualifier(selection *ResourceMappingSelection, resourceKeyMap map[bean.DevtronResourceSearchableKeyName]int, userId int32) (*QualifierMapping, []*QualifierMapping) {
	switch selection.QualifierSelector {
	case ApplicationEnvironmentSelector:
		return GetMappingsForAppEnv(selection, resourceKeyMap, userId)
	}
	return nil, nil
}

func GetMappingsForAppEnv(selection *ResourceMappingSelection, resourceKeyMap map[bean.DevtronResourceSearchableKeyName]int, userId int32) (*QualifierMapping, []*QualifierMapping) {
	appId, appName := GetValuesFromScope(ApplicationSelector, selection.Scope)
	envId, envName := GetValuesFromScope(EnvironmentSelector, selection.Scope)
	compositeString := fmt.Sprintf("%v-%s-%s", selection.ResourceId, appName, envName)

	parent := selection.toResourceMapping(resourceKeyMap, appId, appName, compositeString, userId)
	children := selection.toResourceMapping(resourceKeyMap, envId, envName, compositeString, userId)
	return parent, []*QualifierMapping{children}
}
