package resourceQualifiers

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"strings"
)

func GetQualifierMappingsForCompoundQualifier(selection *ResourceMappingSelection, resourceKeyMap map[bean.DevtronResourceSearchableKeyName]int, userId int32) (*QualifierMapping, []*QualifierMapping) {
	switch selection.QualifierSelector {
	case ApplicationEnvironmentSelector:
		return GetMappingsForAppEnv(selection, resourceKeyMap, userId)
	}
	return nil, nil
}

func GetMappingsForAppEnv(selection *ResourceMappingSelection, resourceKeyMap map[bean.DevtronResourceSearchableKeyName]int, userId int32) (*QualifierMapping, []*QualifierMapping) {
	appId, appName := GetValuesFromSelectionIdentifier(ApplicationSelector, selection.SelectionIdentifier)
	envId, envName := GetValuesFromSelectionIdentifier(EnvironmentSelector, selection.SelectionIdentifier)
	compositeString := getCompositeString(selection.ResourceId, appId, envId)
	parent := selection.toResourceMapping(ApplicationSelector, resourceKeyMap, appId, appName, compositeString, userId)
	children := selection.toResourceMapping(EnvironmentSelector, resourceKeyMap, envId, envName, compositeString, userId)
	return parent, []*QualifierMapping{children}
}

func getCompositeString(ids ...int) string {
	return fmt.Sprintf(strings.Repeat("%v-", len(ids)), ids)
}

func getCompositeStringsAppEnvSelection(selectionIdentifiers []*SelectionIdentifier) mapset.Set {
	compositeSet := mapset.NewSet()
	for _, selectionIdentifier := range selectionIdentifiers {
		compositeSet.Add(getCompositeString(selectionIdentifier.AppId, selectionIdentifier.EnvId))
	}
	return compositeSet
}

func getSelectionIdentifierForAppEnv(appId int, envId int, names *SelectionIdentifierName) *SelectionIdentifier {
	return &SelectionIdentifier{
		AppId:                   appId,
		EnvId:                   envId,
		SelectionIdentifierName: names,
	}
}

func getIdentifierNamesForAppEnv(envName string, appName string) *SelectionIdentifierName {
	return &SelectionIdentifierName{
		EnvironmentName: envName,
		AppName:         appName,
	}
}
