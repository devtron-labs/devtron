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

package pipeline

import (
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	"testing"
)

func TestNewGlobalCMCSServiceImpl(t *testing.T) {

	t.Run("testGetterFunctionForGlobalTemplates", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		NewGlobalCMCSServiceImpl(sugaredLogger, nil)

		var globalCmCsConfigs []*bean.GlobalCMCSDto

		data := make(map[string]string)

		data["a"] = "b"
		data["b"] = "c"

		globalCmCsConfigs = append(globalCmCsConfigs, &bean.GlobalCMCSDto{
			Id:                 0,
			ConfigType:         "SECRET",
			Name:               "test-secret-4",
			Type:               "environment",
			Data:               data,
			MountPath:          "",
			Deleted:            false,
			UserId:             0,
			SecretIngestionFor: "CI/CD",
		}, &bean.GlobalCMCSDto{
			Id:                 0,
			ConfigType:         "CONFIGMAP",
			Name:               "test-secret-4",
			Type:               "environment",
			Data:               data,
			MountPath:          "",
			Deleted:            false,
			UserId:             0,
			SecretIngestionFor: "CI/CD",
		}, &bean.GlobalCMCSDto{
			Id:                 0,
			ConfigType:         "SECRET",
			Name:               "test-secret-4",
			Type:               "volumne",
			Data:               data,
			MountPath:          "",
			Deleted:            false,
			UserId:             0,
			SecretIngestionFor: "CI/CD",
		},
		)

		volumeLen := 0
		for _, gc := range globalCmCsConfigs {
			if gc.Type == "volume" {
				volumeLen++
			}
		}

		steps := make([]v1alpha1.ParallelSteps, 0)
		volumes := make([]v12.Volume, 0)
		templates := make([]v1alpha1.Template, 0)

		err = executors.AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs, &steps, &volumes, &templates)
		assert.Nil(t, err)
		assert.Equal(t, len(steps), len(globalCmCsConfigs))
		assert.Equal(t, len(templates), len(globalCmCsConfigs))
		assert.Equal(t, len(volumes), volumeLen)

	})

}
