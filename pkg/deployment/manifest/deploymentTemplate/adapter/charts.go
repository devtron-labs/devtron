package adapter

import (
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
)

func ChartsDBObjToDTO(chartDbObj *chartRepoRepository.Chart) *bean.Charts {
	return &bean.Charts{
		Id:                      chartDbObj.Id,
		AppId:                   chartDbObj.AppId,
		ChartRepoId:             chartDbObj.ChartRepoId,
		ChartName:               chartDbObj.ChartName,
		ChartVersion:            chartDbObj.ChartVersion,
		ChartRepo:               chartDbObj.ChartRepo,
		ChartRepoUrl:            chartDbObj.ChartRepoUrl,
		Values:                  chartDbObj.Values,
		GlobalOverride:          chartDbObj.GlobalOverride,
		ReleaseOverride:         chartDbObj.ReleaseOverride,
		PipelineOverride:        chartDbObj.PipelineOverride,
		Status:                  chartDbObj.Status,
		Active:                  chartDbObj.Active,
		GitRepoUrl:              chartDbObj.GitRepoUrl,
		ChartLocation:           chartDbObj.ChartLocation,
		ReferenceTemplate:       chartDbObj.ReferenceTemplate,
		ImageDescriptorTemplate: chartDbObj.ImageDescriptorTemplate,
		ChartRefId:              chartDbObj.ChartRefId,
		Latest:                  chartDbObj.Latest,
		Previous:                chartDbObj.Previous,
		ReferenceChart:          chartDbObj.ReferenceChart,
		IsBasicViewLocked:       chartDbObj.IsBasicViewLocked,
		CurrentViewEditor:       chartDbObj.CurrentViewEditor,
		IsCustomGitRepository:   chartDbObj.IsCustomGitRepository,
		ResolvedGlobalOverride:  chartDbObj.ResolvedGlobalOverride,
		CreatedOn:               chartDbObj.CreatedOn,
		CreatedBy:               chartDbObj.CreatedBy,
		UpdatedOn:               chartDbObj.UpdatedOn,
		UpdatedBy:               chartDbObj.UpdatedBy,
	}
}

func ChartsDTOToDBObj(chartsDTO *bean.Charts) *chartRepoRepository.Chart {
	return &chartRepoRepository.Chart{
		Id:                      chartsDTO.Id,
		AppId:                   chartsDTO.AppId,
		ChartRepoId:             chartsDTO.ChartRepoId,
		ChartName:               chartsDTO.ChartName,
		ChartVersion:            chartsDTO.ChartVersion,
		ChartRepo:               chartsDTO.ChartRepo,
		ChartRepoUrl:            chartsDTO.ChartRepoUrl,
		Values:                  chartsDTO.Values,
		GlobalOverride:          chartsDTO.GlobalOverride,
		ReleaseOverride:         chartsDTO.ReleaseOverride,
		PipelineOverride:        chartsDTO.PipelineOverride,
		Status:                  chartsDTO.Status,
		Active:                  chartsDTO.Active,
		GitRepoUrl:              chartsDTO.GitRepoUrl,
		ChartLocation:           chartsDTO.ChartLocation,
		ReferenceTemplate:       chartsDTO.ReferenceTemplate,
		ImageDescriptorTemplate: chartsDTO.ImageDescriptorTemplate,
		ChartRefId:              chartsDTO.ChartRefId,
		Latest:                  chartsDTO.Latest,
		Previous:                chartsDTO.Previous,
		ReferenceChart:          chartsDTO.ReferenceChart,
		IsBasicViewLocked:       chartsDTO.IsBasicViewLocked,
		CurrentViewEditor:       chartsDTO.CurrentViewEditor,
		IsCustomGitRepository:   chartsDTO.IsCustomGitRepository,
		ResolvedGlobalOverride:  chartsDTO.ResolvedGlobalOverride,
		AuditLog: sql.AuditLog{
			CreatedOn: chartsDTO.CreatedOn,
			CreatedBy: chartsDTO.CreatedBy,
			UpdatedOn: chartsDTO.UpdatedOn,
			UpdatedBy: chartsDTO.UpdatedBy,
		},
	}
}
