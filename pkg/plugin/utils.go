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

package plugin

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
)

func getStageType(stageTypeReq string) (int, error) {
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
