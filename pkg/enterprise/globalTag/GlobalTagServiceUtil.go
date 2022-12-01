/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package globalTag

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation"
	"strconv"
	"strings"
)

const ALL_PROJECT_ID string = "-1"

func CheckIfTagIsMandatoryForProject(mandatoryProjectIdsCsv string, projectId int) bool {
	projectIdStr := strconv.Itoa(projectId)
	if len(mandatoryProjectIdsCsv) > 0 {
		mandatoryProjectIds := strings.Split(mandatoryProjectIdsCsv, ",")
		for _, mandatoryProjectId := range mandatoryProjectIds {
			if mandatoryProjectId == ALL_PROJECT_ID || mandatoryProjectId == projectIdStr {
				return true
			}
		}
	}
	return false
}

func CheckIfValidLabels(labels map[string]string, globalTags []*GlobalTagDtoForProject) error {
	// check mandatory labels with values
	for _, globalTag := range globalTags {
		if globalTag.IsMandatory {
			key := globalTag.Key
			if _, ok := labels[key]; !ok {
				return errors.New(fmt.Sprintf("Validation error - Mandatory tag - %s not found in labels", key))
			}
		}
	}

	// check labels keys and values validation
	for labelKey, labelValue := range labels {
		errs := validation.IsQualifiedName(labelKey)
		if len(errs) > 0 {
			return errors.New(fmt.Sprintf("Validation error - label key - %s is not satisfying the label key criteria", labelKey))
		}

		errs = validation.IsValidLabelValue(labelValue)
		if len(errs) > 0 {
			return errors.New(fmt.Sprintf("Validation error - label value - %s is not satisfying the label value criteria for label key - %s", labelValue, labelKey))
		}
	}

	return nil
}
