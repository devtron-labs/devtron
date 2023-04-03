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

func CheckIfMandatoryLabelsProvided(labels map[string]string, globalTags []*GlobalTagDtoForProject) error {
	// check if mandatory label provided
	for _, globalTag := range globalTags {
		if !globalTag.IsMandatory {
			continue
		}
		key := globalTag.Key
		labelValue, found := labels[key]
		if !found {
			return errors.New(fmt.Sprintf("Validation error - Mandatory tag - %s not found in labels", key))
		}
		if len(labelValue) == 0 {
			return errors.New(fmt.Sprintf("Validation error - value for mandatory tag - %s found empty", key))
		}
	}
	return nil
}
