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

package app

import (
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"strings"
)

// LabelMatchingRegex is the official k8s label matching regex, pls refer https://github.com/kubernetes/apimachinery/blob/bfd2aff97e594f6aad77acbe2cbbe190acc93cbc/pkg/util/validation/validation.go#L167
const LabelMatchingRegex = "^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"

// MergeChildMapToParentMap merges child map of generic type map into parent map of generic type
// and returns merged mapping, if parentMap is nil then nil is returned.
func MergeChildMapToParentMap[T comparable, R any](parentMap map[T]R, toMergeMap map[T]R) map[T]R {
	if parentMap == nil {
		return nil
	}
	for key, value := range toMergeMap {
		if _, ok := parentMap[key]; !ok {
			parentMap[key] = value
		}
	}
	return parentMap
}

func sanitizeLabels(extraAppLabels map[string]string) map[string]string {
	for lkey, lvalue := range extraAppLabels {
		if strings.Contains(lvalue, " ") {
			extraAppLabels[lkey] = strings.ReplaceAll(lvalue, " ", "_")
		}
	}
	return extraAppLabels
}

// identifyDuplicateApps identifies the earliest created app and the most recent duplicate app.
func identifyDuplicateApps(apps []*appRepository.App) (earliestApp *appRepository.App, duplicatedApp *appRepository.App) {
	if len(apps) == 0 {
		return nil, nil
	}
	earliestApp = apps[0]
	duplicatedApp = apps[0]
	for _, app := range apps[1:] {
		if app.AuditLog.CreatedOn.Before(earliestApp.AuditLog.CreatedOn) {
			earliestApp = app
		}
		if app.AuditLog.CreatedOn.After(duplicatedApp.AuditLog.CreatedOn) {
			duplicatedApp = app
		}
	}
	return earliestApp, duplicatedApp
}
