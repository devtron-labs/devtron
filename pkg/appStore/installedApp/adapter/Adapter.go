package adapter

import (
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"path"
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

func ParseChartCreateRequest(appName string) *util.ChartCreateRequest {
	chartPath := getRefProxyChartPath()
	return &util.ChartCreateRequest{
		ChartMetaData: &chart.Metadata{
			Name:    appName,
			Version: "1.0.1",
		},
		ChartPath: chartPath,
	}
}

func getRefProxyChartPath() string {
	template := appStoreBean.CHART_PROXY_TEMPLATE
	return path.Join(appStoreBean.RefChartProxyDirPath, template)
}
