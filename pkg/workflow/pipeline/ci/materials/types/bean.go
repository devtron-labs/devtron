package types

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
)

type CiPipelineMaterialModel struct {
	//TODO KB: check model bean would work or not ??
	//CiPipeline *bean.CiPipeline // not needed

	// fields needed are metadata along with GitProvider
	GitMaterialId int
	GitMaterial   *bean.GitMaterial
	GitProvider   *repository.GitProvider //TODO KB: should not be sql struct

}
