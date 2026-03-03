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

package utils

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"golang.org/x/mod/semver"
)

func GetStageType(stageTypeReq string) (int, error) {
	var stageType int
	switch stageTypeReq {
	case repository.CI_STAGE_TYPE:
		stageType = repository.CI
	case repository.CD_STAGE_TYPE:
		stageType = repository.CD
	case repository.CI_CD_STAGE_TYPE:
		stageType = repository.CI_CD
	default:
		return 0, errors.New("stage type not recognised, please add valid stage type in query parameter")
	}
	return stageType, nil
}

// CreateUniqueIdentifier helper func to create plugin identifier
func CreateUniqueIdentifier(pluginName string, pluginId int) string {
	identifier := strings.ToLower(pluginName)
	// Create a regular expression to match any of the special characters
	re := regexp.MustCompile("[" + regexp.QuoteMeta(bean2.SpecialCharsRegex) + "]+")
	// Replace all occurrences of the special characters with a dash
	transformedIdentifier := re.ReplaceAllString(identifier, "-")
	transformedIdentifier = strings.Trim(transformedIdentifier, "-")
	if pluginId > 0 {
		transformedIdentifier = fmt.Sprintf("%s-%d", transformedIdentifier, pluginId)
	}
	return transformedIdentifier
}

func SortParentMetadataDtoSliceByName(pluginParentMetadataDtos []*bean2.PluginParentMetadataDto) {
	sort.Slice(pluginParentMetadataDtos, func(i, j int) bool {
		if strings.Compare(pluginParentMetadataDtos[i].Name, pluginParentMetadataDtos[j].Name) <= 0 {
			return true
		}
		return false
	})
}

func SortPluginsVersionDetailSliceByCreatedOn(pluginsVersionDetail []*bean2.PluginsVersionDetail) {
	sort.Slice(pluginsVersionDetail, func(i, j int) bool {
		if pluginsVersionDetail[i].CreatedOn.After(pluginsVersionDetail[j].CreatedOn) {
			return true
		}
		return false
	})
}

func FetchIconAndCheckSize(url string, maxSize int64) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	iconResp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error in fetching icon : %s", err.Error())
	}
	if iconResp != nil {
		if iconResp.StatusCode >= 200 && iconResp.StatusCode < 300 {
			if iconResp.ContentLength > maxSize {
				return fmt.Errorf("icon size too large")
			}
			iconResp.Body.Close()
		} else {
			return fmt.Errorf("error in fetching icon : %s", iconResp.Status)
		}
	} else {
		return fmt.Errorf("error in fetching icon : empty response")
	}
	return nil
}

func ValidatePluginVersion(version string) error {
	if !strings.Contains(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}
	// semantic versioning validation on plugin's version
	if !semver.IsValid(version) {
		return util.NewApiError(http.StatusBadRequest, bean2.PluginVersionNotSemanticallyCorrectError, bean2.PluginVersionNotSemanticallyCorrectError)
	}
	return nil
}

// StripVersionSuffixFromName removes version suffix patterns like "v1.0.0", "v1.0", "v1" from the end of plugin names.
// This is used during migration to ensure plugins with names like "DockerSlim v1.0.0" and "DockerSlim"
// generate the same identifier and are linked to the same parent metadata.
//
// Examples:
//   - "DockerSlim v1.0.0" -> "DockerSlim"
//   - "Github Release Plugin v1.0" -> "Github Release Plugin"
//   - "Devtron CI Trigger v1.0.0" -> "Devtron CI Trigger"
//   - "DockerSlim" -> "DockerSlim" (no change)
func StripVersionSuffixFromName(pluginName string) string {
	// Pattern to match version suffix like " v1.0.0", " v1.0", " v1" at the end of the name
	// Also handles pre-release versions like " v1.0.0-beta.1"
	versionSuffixPattern := regexp.MustCompile(`\s+v\d+(\.\d+)*(-[a-zA-Z0-9.]+)?$`)
	return strings.TrimSpace(versionSuffixPattern.ReplaceAllString(pluginName, ""))
}
