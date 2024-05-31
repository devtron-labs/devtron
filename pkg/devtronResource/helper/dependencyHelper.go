/*
 * Copyright (c) 2024. Devtron Inc.
 */

package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
	"strings"
)

func DecodeDependencyInfoString(dependencyInfo string) (resourceIdentifier *bean.ResourceIdentifier, err error) {
	resourceIdentifier = &bean.ResourceIdentifier{}
	dependencyInfoSplits := strings.Split(dependencyInfo, "|")
	if len(dependencyInfoSplits) != 3 && len(dependencyInfoSplits) != 4 {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryDependencyInfo, bean.InvalidQueryDependencyInfo)
	}
	kind, subKind, err := GetKindAndSubKindFrom(dependencyInfoSplits[0])
	if err != nil {
		return nil, err
	}
	resourceIdentifier.ResourceKind = bean.DevtronResourceKind(kind)
	resourceIdentifier.ResourceSubKind = bean.DevtronResourceKind(subKind)
	resourceIdentifier.ResourceVersion = bean.DevtronResourceVersion(dependencyInfoSplits[1])
	if len(dependencyInfoSplits) == 3 && dependencyInfoSplits[2] == bean.AllIdentifierQueryString {
		resourceIdentifier.Identifier = bean.AllIdentifierQueryString
	} else if len(dependencyInfoSplits) == 4 {
		if dependencyInfoSplits[2] == bean.IdentifierQueryString {
			resourceIdentifier.Identifier = dependencyInfoSplits[3]
		} else if dependencyInfoSplits[2] == bean.IdQueryString {
			resourceIdentifier.Id, err = strconv.Atoi(dependencyInfoSplits[3])
			if err != nil {
				return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryDependencyInfo, err.Error())
			}
		}
	} else {
		return nil, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean.InvalidQueryDependencyInfo, bean.InvalidQueryDependencyInfo)
	}
	return resourceIdentifier, nil
}

func GetDependencyOldObjectIdsForSpecificType(dependencies []*bean.DevtronResourceDependencyBean, typeOfDependency bean.DevtronResourceDependencyType) []int {
	dependencyIds := make([]int, len(dependencies))
	for _, dependency := range dependencies {
		if dependency.TypeOfDependency != typeOfDependency {
			continue
		}
		dependencyIds = append(dependencyIds, dependency.OldObjectId)
	}
	return dependencyIds
}

func CheckIfDependencyTypeToBeValidated(dependencyType bean.DevtronResourceDependencyType) bool {
	return dependencyType != bean.DevtronResourceDependencyTypeLevel
}

func CheckIfDependencyIsDependentOnRemovedDependency(updatedDependencies []*bean.DevtronResourceDependencyBean, indexOfDependencyRemoved int) bool {
	for _, dependency := range updatedDependencies {
		isDependent := slices.Contains(dependency.DependentOnIndexes, indexOfDependencyRemoved)
		if isDependent {
			return isDependent
		}
	}
	return false
}

func GetKeyForADependencyMap(oldObjectId, devtronResourceSchemaId int) string {
	// key can be "oldObjectId-schemaId" or "name-schemaId"
	return fmt.Sprintf("%d-%d", oldObjectId, devtronResourceSchemaId)
}

func IsApplicationDependency(devtronResourceTypeReq *bean.DevtronResourceTypeReq) bool {
	if devtronResourceTypeReq == nil {
		return false
	}
	return devtronResourceTypeReq.ResourceKind ==
		BuildExtendedResourceKindUsingKindAndSubKind(bean.DevtronResourceApplication.ToString(),
			bean.DevtronResourceDevtronApplication.ToString())
}
