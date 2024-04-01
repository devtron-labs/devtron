package bean

import (
	"context"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

type ErrorState int

type FeasibilityErrorState struct {
	ErrorState   ErrorState
	ErrorMessage string
}

const (
	VulnerabilityFoundFeasibilityError ErrorState = iota
)

func (state *FeasibilityErrorState) SetVulnerabilityFeasibilityError() {
	state.ErrorState = VulnerabilityFoundFeasibilityError
}

type FeasibilityDto struct {
	Pipeline    *pipelineConfig.Pipeline
	Artifact    *repository.CiArtifact
	Runner      *pipelineConfig.CdWorkflowRunner
	Context     context.Context
	TriggeredBy int32
}
