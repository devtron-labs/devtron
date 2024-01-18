package pipeline

import (
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	"testing"
)

func TestNewGlobalCMCSServiceImpl(t *testing.T) {

	t.Run("testGetterFunctionForGlobalTemplates", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		NewGlobalCMCSServiceImpl(sugaredLogger, nil)

		var globalCmCsConfigs []*bean.GlobalCMCSDto

		data := make(map[string]string)

		data["a"] = "b"
		data["b"] = "c"

		globalCmCsConfigs = append(globalCmCsConfigs, &bean.GlobalCMCSDto{
			Id:                 0,
			ConfigType:         "SECRET",
			Name:               "test-secret-4",
			Type:               "environment",
			Data:               data,
			MountPath:          "",
			Deleted:            false,
			UserId:             0,
			SecretIngestionFor: "CI/CD",
		}, &bean.GlobalCMCSDto{
			Id:                 0,
			ConfigType:         "CONFIGMAP",
			Name:               "test-secret-4",
			Type:               "environment",
			Data:               data,
			MountPath:          "",
			Deleted:            false,
			UserId:             0,
			SecretIngestionFor: "CI/CD",
		}, &bean.GlobalCMCSDto{
			Id:                 0,
			ConfigType:         "SECRET",
			Name:               "test-secret-4",
			Type:               "volumne",
			Data:               data,
			MountPath:          "",
			Deleted:            false,
			UserId:             0,
			SecretIngestionFor: "CI/CD",
		},
		)

		volumeLen := 0
		for _, gc := range globalCmCsConfigs {
			if gc.Type == "volume" {
				volumeLen++
			}
		}

		steps := make([]v1alpha1.ParallelSteps, 0)
		volumes := make([]v12.Volume, 0)
		templates := make([]v1alpha1.Template, 0)

		err = executors.AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs, &steps, &volumes, &templates)
		assert.Nil(t, err)
		assert.Equal(t, len(steps), len(globalCmCsConfigs))
		assert.Equal(t, len(templates), len(globalCmCsConfigs))
		assert.Equal(t, len(volumes), volumeLen)

	})

}
