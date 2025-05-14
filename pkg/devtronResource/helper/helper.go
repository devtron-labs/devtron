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

package helper

import (
	"github.com/devtron-labs/devtron/internal/util"
	"net/http"
	"strings"
)

func GetKindAndSubKindFrom(resourceKindVar string) (kind, subKind string, err error) {
	kindSplits := strings.Split(resourceKindVar, "/")
	if len(kindSplits) == 1 {
		kind = kindSplits[0]
	} else if len(kindSplits) == 2 {
		kind = kindSplits[0]
		subKind = kindSplits[1]
	} else {
		return kind, subKind, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			Code:            "400",
			InternalMessage: "invalid kind!",
			UserMessage:     "invalid kind!",
		}
	}
	return kind, subKind, nil
}
