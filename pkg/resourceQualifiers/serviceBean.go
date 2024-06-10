/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resourceQualifiers

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"golang.org/x/exp/slices"
	"time"
)

type QualifierSelector int

const (
	ApplicationSelector            QualifierSelector = 0
	EnvironmentSelector            QualifierSelector = 1
	ClusterSelector                QualifierSelector = 2
	ApplicationEnvironmentSelector QualifierSelector = 3
	GlobalSelector                 QualifierSelector = 4
)

func (selector QualifierSelector) isCompound() bool {
	return slices.Contains(CompoundQualifiers, selector.toQualifier())
}

func (selector QualifierSelector) toQualifier() Qualifier {
	switch selector {
	case ApplicationSelector:
		return APP_QUALIFIER
	case EnvironmentSelector:
		return ENV_QUALIFIER
	case ClusterSelector:
		return CLUSTER_QUALIFIER
	case ApplicationEnvironmentSelector:
		return APP_AND_ENV_QUALIFIER
	case GlobalSelector:
		return GLOBAL_QUALIFIER
	}
	return Qualifier(0)
}

func GetIdentifierKey(selector QualifierSelector, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) int {
	switch selector {
	case ApplicationSelector:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID]
	case ClusterSelector:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	case EnvironmentSelector:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	default:
		return 0
	}
}

func GetSelectorFromKey(key int, searchableKeyIdNameMap map[int]bean.DevtronResourceSearchableKeyName) QualifierSelector {

	name, ok := searchableKeyIdNameMap[key]
	if !ok {
		return 0
	}

	switch name {
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID:
		return ApplicationSelector
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID:
		return ClusterSelector
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID:
		return EnvironmentSelector
	default:
		return 0
	}
}

func GetQualifierIdForSelector(selector QualifierSelector) Qualifier {
	switch selector {
	case ApplicationEnvironmentSelector:
		return APP_AND_ENV_QUALIFIER
	case ApplicationSelector:
		return APP_QUALIFIER
	case EnvironmentSelector:
		return ENV_QUALIFIER
	case ClusterSelector:
		return CLUSTER_QUALIFIER
	case GlobalSelector:
		return GLOBAL_QUALIFIER
	default:
		return 0
	}
}

func GetValuesFromSelectionIdentifier(selector QualifierSelector, selectionIdentifier *SelectionIdentifier) (int, string) {
	if selectionIdentifier == nil {
		selectionIdentifier = &SelectionIdentifier{}
	}
	if selectionIdentifier.SelectionIdentifierName == nil {
		selectionIdentifier.SelectionIdentifierName = &SelectionIdentifierName{}
	}
	switch selector {
	case ApplicationSelector:
		return selectionIdentifier.AppId, selectionIdentifier.SelectionIdentifierName.AppName
	case EnvironmentSelector:
		return selectionIdentifier.EnvId, selectionIdentifier.SelectionIdentifierName.EnvironmentName
	case ClusterSelector:
		return selectionIdentifier.ClusterId, selectionIdentifier.SelectionIdentifierName.ClusterName
	default:
		return 0, ""
	}
}
func getAuditLog(userid int32) sql.AuditLog {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userid,
		UpdatedOn: time.Now(),
		UpdatedBy: userid,
	}
	return auditLog
}

func (selection *ResourceMappingSelection) toResourceMapping(selector QualifierSelector, resourceKeyMap map[bean.DevtronResourceSearchableKeyName]int, valueInt int, valueString string, compositeString string, userId int32) *QualifierMapping {
	return &QualifierMapping{
		ResourceId:            selection.ResourceId,
		ResourceType:          selection.ResourceType,
		QualifierId:           int(GetQualifierIdForSelector(selection.QualifierSelector)),
		IdentifierKey:         GetIdentifierKey(selector, resourceKeyMap),
		IdentifierValueInt:    valueInt,
		IdentifierValueString: valueString,
		Active:                true,
		CompositeKey:          compositeString,
		AuditLog:              getAuditLog(userId),
	}
}
