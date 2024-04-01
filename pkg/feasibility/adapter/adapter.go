package adapter

import (
	"context"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/feasibility/bean"
)

func GetFeasibilityDto(artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, runner *pipelineConfig.CdWorkflowRunner, triggeredBy int32, context context.Context) *bean2.FeasibilityDto {
	return &bean2.FeasibilityDto{
		Artifact:    artifact,
		Pipeline:    pipeline,
		Runner:      runner,
		TriggeredBy: triggeredBy,
		Context:     context,
	}
}
