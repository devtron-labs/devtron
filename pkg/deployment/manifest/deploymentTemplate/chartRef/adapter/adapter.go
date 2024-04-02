package adapter

import (
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
)

func ConvertChartRefDbObjToBean(ref *chartRepoRepository.ChartRef) *bean.ChartRefDto {
	dto := &bean.ChartRefDto{
		Id:                     ref.Id,
		Location:               ref.Location,
		Version:                ref.Version,
		Default:                ref.Default,
		ChartData:              ref.ChartData,
		ChartDescription:       ref.ChartDescription,
		UserUploaded:           ref.UserUploaded,
		IsAppMetricsSupported:  ref.IsAppMetricsSupported,
		DeploymentStrategyPath: ref.DeploymentStrategyPath,
		JsonPathForStrategy:    ref.JsonPathForStrategy,
	}
	if len(ref.Name) == 0 {
		dto.Name = bean.RolloutChartType
	} else {
		dto.Name = ref.Name
	}
	return dto
}

func ConvertCustomChartRefDtoToDbObj(ref *bean.CustomChartRefDto) *chartRepoRepository.ChartRef {
	return &chartRepoRepository.ChartRef{
		Id:                     ref.Id,
		Location:               ref.Location,
		Version:                ref.Version,
		Active:                 ref.Active,
		Default:                ref.Default,
		Name:                   ref.Name,
		ChartData:              ref.ChartData,
		ChartDescription:       ref.ChartDescription,
		UserUploaded:           ref.UserUploaded,
		IsAppMetricsSupported:  ref.IsAppMetricsSupported,
		DeploymentStrategyPath: ref.DeploymentStrategyPath,
		JsonPathForStrategy:    ref.JsonPathForStrategy,
		AuditLog:               ref.AuditLog,
	}
}
