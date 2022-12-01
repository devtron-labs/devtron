package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChartTemplateService(t *testing.T) {
	t.SkipNow()
	t.Run("getValues", func(t *testing.T) {
		logger, err := NewSugardLogger()
		assert.Nil(t, err)
		impl := ChartTemplateServiceImpl{
			logger: logger,
		}
		directory := "github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts"
		pipelineStrategyPath := "pipeline-values.yaml"
		values, err := impl.getValues(directory, pipelineStrategyPath)
		assert.Nil(t, err)
		assert.Nil(t, values)
	})
}
