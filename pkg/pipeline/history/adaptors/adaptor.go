package adaptors

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/history/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
)

func GetHistoryDetailDto(history *repository.DeploymentTemplateHistory, variableSnapshotMap map[string]string, resolvedTemplate string) *bean.HistoryDetailDto {
	if history == nil {
		return &bean.HistoryDetailDto{}
	}
	return &bean.HistoryDetailDto{
		TemplateName:        history.TemplateName,
		TemplateVersion:     history.TemplateVersion,
		IsAppMetricsEnabled: &history.IsAppMetricsEnabled,
		CodeEditorValue: &bean.HistoryDetailConfig{
			DisplayName:      "values.yaml",
			Value:            history.Template,
			VariableSnapshot: variableSnapshotMap,
			ResolvedValue:    resolvedTemplate,
		},
	}
}
