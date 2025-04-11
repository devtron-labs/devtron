package adaptor

import (
	"encoding/json"
	"fmt"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chart/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"strings"
)

// ChartAdaptor converts db object chartRepoRepository.Chart to bean.TemplateRequest
func ChartAdaptor(chartModel *chartRepoRepository.Chart,
	isAppMetricsEnabled bool, deploymentConfig *bean2.DeploymentConfig) (*bean.TemplateRequest, error) {
	if chartModel == nil || chartModel.Id == 0 {
		return &bean.TemplateRequest{}, &util.ApiError{UserMessage: "no chartInput found"}
	}
	gitRepoUrl := ""
	targetRevision := util2.GetDefaultTargetRevision()
	if !apiGitOpsBean.IsGitOpsRepoNotConfigured(deploymentConfig.GetRepoURL()) {
		gitRepoUrl = deploymentConfig.GetRepoURL()
		targetRevision = deploymentConfig.GetTargetRevision()
	}
	templateRequest := &bean.TemplateRequest{
		RefChartTemplate:        chartModel.ReferenceTemplate,
		Id:                      chartModel.Id,
		AppId:                   chartModel.AppId,
		ChartRepositoryId:       chartModel.ChartRepoId,
		DefaultAppOverride:      json.RawMessage(chartModel.GlobalOverride),
		RefChartTemplateVersion: getParentChartVersion(chartModel.ChartVersion),
		Latest:                  chartModel.Latest,
		ChartRefId:              chartModel.ChartRefId,
		IsAppMetricsEnabled:     isAppMetricsEnabled,
		IsBasicViewLocked:       chartModel.IsBasicViewLocked,
		CurrentViewEditor:       chartModel.CurrentViewEditor,
		GitRepoUrl:              gitRepoUrl,
		IsCustomGitRepository:   deploymentConfig.ConfigType == bean2.CUSTOM.String(),
		ImageDescriptorTemplate: chartModel.ImageDescriptorTemplate,
		TargetRevision:          targetRevision,
	}
	if chartModel.Latest {
		templateRequest.LatestChartVersion = chartModel.ChartVersion
	}
	return templateRequest, nil
}

func getParentChartVersion(childVersion string) string {
	placeholders := strings.Split(childVersion, ".")
	return fmt.Sprintf("%s.%s.0", placeholders[0], placeholders[1])
}
