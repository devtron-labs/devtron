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
