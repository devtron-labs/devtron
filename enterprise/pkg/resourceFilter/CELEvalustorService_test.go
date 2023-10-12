package resourceFilter

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluateCELRequest(t *testing.T) {
	logger, _ := util.NewSugardLogger()
	celService := NewCELServiceImpl(logger)
	t.Run("valid release tags list", func(tt *testing.T) {
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
	})

	t.Run("empty release tags list", func(tt *testing.T) {
		artifact := "devtron/test:v1beta1"
		releaseTags := []string{}
		params := celService.GetParamsFromArtifact(artifact, releaseTags)

		evalReq := CELRequest{
			Expression: "'latest' in releaseTags",
			ExpressionMetadata: ExpressionMetadata{
				Params: params,
			},
		}
		res, err := celService.EvaluateCELRequest(evalReq)
		assert.Equal(t, false, res)
		assert.Equal(t, nil, err)
	})

	t.Run("nil release tags list", func(tt *testing.T) {
		artifact := "devtron/test:v1beta1"
		var releaseTags []string
		params := celService.GetParamsFromArtifact(artifact, releaseTags)

		evalReq := CELRequest{
			Expression: "'latest' in releaseTags",
			ExpressionMetadata: ExpressionMetadata{
				Params: params,
			},
		}
		res, err := celService.EvaluateCELRequest(evalReq)
		assert.Equal(t, false, res)
		assert.Equal(t, nil, err)
	})
}
