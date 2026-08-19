package adapter

import (
	bean2 "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/app"
	bean3 "github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
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

func ArgoAppToExternalGitOpsApp(applications []*bean3.ArgoApplicationListDto) []*bean.ExternalGitOpsAppDto {
	result := make([]*bean.ExternalGitOpsAppDto, 0, len(applications))
	for _, application := range applications {
		appDto := bean.ExternalGitOpsAppDto{
			AppName:     application.Name,
			Namespace:   application.Namespace,
			ClusterId:   application.ClusterId,
			ClusterName: application.ClusterName,
		}
		result = append(result, &appDto)
	}

	return result
}

func FluxAppToExternalGitOpsApp(applications []bean4.FluxApplication) []*bean.ExternalGitOpsAppDto {
	result := make([]*bean.ExternalGitOpsAppDto, 0, len(applications))
	for _, application := range applications {
		appDto := bean.ExternalGitOpsAppDto{
			AppName:     application.Name,
			Namespace:   application.Namespace,
			ClusterId:   application.ClusterId,
			ClusterName: application.ClusterName,
		}
		result = append(result, &appDto)
	}

	return result
}
