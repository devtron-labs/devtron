package pipelineConfig

import (
	"fmt"
	"github.com/devtron-labs/common-lib-private/utils"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCiWorkflowRepository(t *testing.T) {
	t.SkipNow()
	t.Run("FindBuildTypeAndStatusDataOfLast1Day", func(t *testing.T) {
		cfg, _ := sql.GetConfig()
		logger, err := utils.NewSugardLogger()
		con, _ := sql.NewDbConnection(cfg, logger)
		assert.Nil(t, err)
		workflowRepositoryImpl := NewCiWorkflowRepositoryImpl(con, logger)
		statusData := workflowRepositoryImpl.FindBuildTypeAndStatusDataOfLast1Day()
		for _, statusDatum := range statusData {
			fmt.Println(statusDatum.Count)
		}
	})
}
