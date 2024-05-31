/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
