package resourceFilter

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluateCELRequest(t *testing.T) {
	logger, _ := util.NewSugardLogger()
	celService := NewCELServiceImpl(logger)
	artifact := "devtron/test:v1beta1"
	releaseTags := []string{"tag1", "latest"}
	params := celService.GetParamsFromArtifact(artifact, releaseTags)

	evalReq := CELRequest{
		Expression: "'latest' in releaseTags",
		ExpressionMetadata: ExpressionMetadata{
			Params: params,
		},
	}
	res, err := celService.EvaluateCELRequest(evalReq)
	assert.Equal(t, true, res)
	assert.Equal(t, nil, err)
}
