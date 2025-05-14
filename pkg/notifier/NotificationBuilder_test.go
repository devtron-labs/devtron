/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package notifier

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"k8s.io/utils/pointer"
	"maps"
	"testing"
)

var logger, _ = util.NewSugardLogger()

func generateCombinations(testData beans.NotificationConfigRequest) []*beans.NotificationConfigRequest {
	var result []*beans.NotificationConfigRequest

	// List of all fields with empty and non-empty combinations
	combinations := []struct {
		teamIds    []*int
		appIds     []*int
		envIds     []*int
		clusterIds []*int
		pipelineId *int
	}{
		{testData.TeamId, testData.AppId, testData.EnvId, testData.ClusterId, testData.PipelineId}, // All non-empty
		{nil, testData.AppId, testData.EnvId, testData.ClusterId, testData.PipelineId},             // TeamId empty
		{testData.TeamId, nil, testData.EnvId, testData.ClusterId, testData.PipelineId},            // AppId empty
		{testData.TeamId, testData.AppId, nil, testData.ClusterId, testData.PipelineId},            // EnvId empty
		{testData.TeamId, testData.AppId, testData.EnvId, nil, testData.PipelineId},                // ClusterId empty
		{nil, nil, testData.EnvId, testData.ClusterId, testData.PipelineId},                        // TeamId and AppId empty
		{nil, testData.AppId, nil, testData.ClusterId, testData.PipelineId},                        // TeamId and EnvId empty
		{nil, testData.AppId, testData.EnvId, nil, testData.PipelineId},                            // TeamId and ClusterId empty
		{testData.TeamId, nil, nil, testData.ClusterId, testData.PipelineId},                       // AppId and EnvId empty
		{testData.TeamId, nil, testData.EnvId, nil, testData.PipelineId},                           // AppId and ClusterId empty
		{testData.TeamId, testData.AppId, nil, nil, testData.PipelineId},                           // EnvId and ClusterId empty
		{nil, nil, nil, testData.ClusterId, testData.PipelineId},                                   // TeamId, AppId, and EnvId empty
		{nil, nil, testData.EnvId, nil, testData.PipelineId},                                       // TeamId, AppId, and ClusterId empty
		{nil, testData.AppId, nil, nil, testData.PipelineId},                                       // TeamId, EnvId, and ClusterId empty
		{testData.TeamId, nil, nil, nil, testData.PipelineId},                                      // AppId, EnvId, and ClusterId empty
		{nil, nil, nil, nil, testData.PipelineId},                                                  // All empty
	}

	// Generate Test structs for all combinations
	for _, combo := range combinations {
		result = append(result, &beans.NotificationConfigRequest{
			TeamId:     combo.teamIds,
			AppId:      combo.appIds,
			EnvId:      combo.envIds,
			ClusterId:  combo.clusterIds,
			PipelineId: testData.PipelineId, // Keep PipelineId same for all
		})
	}

	return result
}

func StringifyLocalRequest(lr *beans.LocalRequest) string {
	params := make([]interface{}, 0)
	if lr.TeamId != nil {
		params = append(params, *lr.TeamId)
	}
	if lr.AppId != nil {
		params = append(params, *lr.AppId)
	}
	if lr.EnvId != nil {
		params = append(params, *lr.EnvId)
	}
	if lr.ClusterId != nil {
		params = append(params, *lr.ClusterId)
	}

	if lr.PipelineId != nil {
		params = append(params, *lr.PipelineId)
	}

	return fmt.Sprintf("TeamId: %v, AppId: %v, EnvId: %v,ClusterId: %v, PipelineId: %v", params...)
}

func Equal(settingsV1, settingsV2 []*beans.LocalRequest) bool {
	settingsV1Set := make(map[string]bool)
	settingsV2Set := make(map[string]bool)

	for _, setting := range settingsV1 {
		settingsV1Set[StringifyLocalRequest(setting)] = true
	}

	for _, setting := range settingsV2 {
		settingsV2Set[StringifyLocalRequest(setting)] = true
	}

	return maps.Equal(settingsV1Set, settingsV2Set)
}

// TestGenerateSettings tests the generation of settings with old and new data
func TestGenerateSettings(t *testing.T) {
	testData := beans.NotificationConfigRequest{
		TeamId:     []*int{pointer.Int(1), pointer.Int(2)},
		ClusterId:  []*int{pointer.Int(-1), pointer.Int(21)},
		EnvId:      []*int{pointer.Int(-1), pointer.Int(-2), pointer.Int(56)},
		AppId:      []*int{pointer.Int(11), pointer.Int(12), pointer.Int(13)},
		PipelineId: pointer.Int(100),
	}

	// Generate all possible combinations
	combinations := generateCombinations(testData)
	// Test each combination
	for i, combo := range combinations {
		t.Run(fmt.Sprintf("combination-%v", i), func(tt *testing.T) {
			settingsV1 := combo.GenerateSettingCombinationsV1()
			settingsV2 := combo.GenerateSettingCombinations()

			if len(settingsV2) == 0 {
				logger.Errorw("settingsV2 cannot be empty for the request", "request", *combo)
				tt.Fail()
			}

			if combo.ClusterId != nil {
				if len(settingsV2) != (len(testData.ClusterId))*len(settingsV1) {
					tt.Fail()
				}
			} else {
				if !Equal(settingsV1, settingsV2) {
					tt.Fail()
				}
			}
		})
	}
}
