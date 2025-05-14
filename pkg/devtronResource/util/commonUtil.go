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

package util

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"net/http"
	"strings"
)

func DecodeFilterCriteriaString(criteria string) (*bean.FilterCriteriaDecoder, error) {
	objs := strings.Split(criteria, "|")
	if len(objs) != 3 {
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			Code:            "400",
			InternalMessage: "invalid format filter criteria!",
			UserMessage:     "invalid format filter criteria!",
		}
	}
	criteriaDecoder := adapter.BuildFilterCriteriaDecoder(objs[0], objs[1], objs[2])
	return criteriaDecoder, nil
}
