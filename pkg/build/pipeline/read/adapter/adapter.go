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

package adapter

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/build/pipeline/read/bean"
)

func NewCiPipelineMin(ciPipeline *pipelineConfig.CiPipeline) (*bean.CiPipelineMin, error) {
	if ciPipeline == nil {
		return nil, errors.New("ci pipeline not found")
	}
	dto := &bean.CiPipelineMin{
		Id:               ciPipeline.Id,
		Name:             ciPipeline.Name,
		AppId:            ciPipeline.AppId,
		ParentCiPipeline: ciPipeline.ParentCiPipeline,
		CiPipelineType:   ciPipeline.PipelineType,
	}
	if ciPipeline.App != nil {
		dto.TeamId = ciPipeline.App.TeamId
	}
	return dto, nil
}
