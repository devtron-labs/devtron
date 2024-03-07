package resourceFilter

import (
	"fmt"
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
		params := GetParamsFromArtifact(artifact, releaseTags)

		evalReq := CELRequest{
			Expression: "'latest' in imageLabels",
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
		params := GetParamsFromArtifact(artifact, releaseTags)

		evalReq := CELRequest{
			Expression: "'latest' in imageLabels",
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
		params := GetParamsFromArtifact(artifact, releaseTags)

		evalReq := CELRequest{
			Expression: "'latest' in imageLabels",
			ExpressionMetadata: ExpressionMetadata{
				Params: params,
			},
		}
		res, err := celService.EvaluateCELRequest(evalReq)
		assert.Equal(t, false, res)
		assert.Equal(t, nil, err)
	})

	t.Run("test commitDetails", func(tt *testing.T) {
		artifact := "devtron/test:v1beta1"
		var releaseTags []string
		params := GetParamsFromArtifact(artifact, releaseTags, &CommitDetails{"github.com/test", "test commit", "test"})

		evalReq := CELRequest{
			Expression: "commitDetailsMap['github.com/test'].branch == 'test'",
			ExpressionMetadata: ExpressionMetadata{
				Params: params,
			},
		}
		b, err := celService.EvaluateCELRequest(evalReq)
		fmt.Println(b, err)
		assert.Equal(t, nil, err)
	})

}
