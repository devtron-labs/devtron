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
	bean2 "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/userResource/bean"
)

func BuildUserResourceResponseDto(data interface{}) *bean.UserResourceResponseDto {
	return &bean.UserResourceResponseDto{
		Data: data,
	}
}

func BuildNullDataUserResourceResponseDto() *bean.UserResourceResponseDto {
	return &bean.UserResourceResponseDto{
		Data: nil,
	}
}

func BuildFetchAppListingReqForJobFromDto(reqBean *bean2.ResourceOptionsReqDto) app.FetchAppListingRequest {
	return app.FetchAppListingRequest{
		Teams:     reqBean.TeamIds,
		SortBy:    helper.AppNameSortBy, // default values set
		SortOrder: helper.Asc,           // default values set
	}
}
