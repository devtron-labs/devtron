package adapter

import (
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/remote/bean"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

func ParseChartGitPushRequest(installAppRequestDTO *appStoreBean.InstallAppVersionDTO, repoURl string, tempRefChart string) *bean.PushChartToGitRequestDTO {
	return &bean.PushChartToGitRequestDTO{
		AppName:           installAppRequestDTO.AppName,
		EnvName:           installAppRequestDTO.Environment.Name,
		ChartAppStoreName: installAppRequestDTO.AppStoreName,
		RepoURL:           repoURl,
		TempChartRefDir:   tempRefChart,
		UserId:            installAppRequestDTO.UserId,
	}
}

func ParseChartCreateRequest(installAppRequestDTO *appStoreBean.InstallAppVersionDTO, chartPath string) *util.ChartCreateRequest {
	return &util.ChartCreateRequest{ChartMetaData: &chart.Metadata{
		Name:    installAppRequestDTO.AppName,
		Version: "1.0.1",
	}, ChartPath: chartPath}
}
