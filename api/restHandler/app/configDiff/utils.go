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

package configDiff

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/config/configDiff/bean"
)

var validConfigCategories = map[string]bool{bean.Secret.ToString(): true, bean.ConfigMap.ToString(): true, bean.DeploymentTemplate.ToString(): true, bean.PipelineStrategy.ToString(): true}
var ErrInvalidConfigCategory = errors.New("invalid config category provided")
var ErrInvalidComparisonItems = errors.New("invalid comparison items, only 2 items are supported for comparison")
var ErrInvalidIndexValInComparisonItems = errors.New("invalid index values in comparison items")

func validateComparisonRequest(configCategory string, comparisonRequestDto bean.ComparisonRequestDto) error {
	if ok := validConfigCategories[configCategory]; !ok {
		return ErrInvalidConfigCategory
	}
	// comparison items expects exactly two items
	if len(comparisonRequestDto.ComparisonItems) != 2 {
		return ErrInvalidComparisonItems
	}
	// if index value is other than 0 or 1 then throw invalid index error
	if len(comparisonRequestDto.ComparisonItems) > 1 && (comparisonRequestDto.ComparisonItems[0].Index != 0 && comparisonRequestDto.ComparisonItems[1].Index != 1) {
		return ErrInvalidIndexValInComparisonItems
	}
	return nil
}
