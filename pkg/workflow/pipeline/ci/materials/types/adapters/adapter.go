package adapters

import (
	pc "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types"
)

func ConvertToPipelineMaterials(materialEntities []*pc.CiPipelineMaterialEntity) []*types.CiPipelineMaterialModel {
	var models []*types.CiPipelineMaterialModel
	for _, materialEntity := range materialEntities {
		if !materialEntity.Active || !materialEntity.GitMaterial.Active {
			continue
		}
		models = append(models, ConvertToPipelineMaterial(materialEntity))
	}
	return models
}

func ConvertToPipelineMaterial(materialEntity *pc.CiPipelineMaterialEntity) *types.CiPipelineMaterialModel {
	model := &types.CiPipelineMaterialModel{
		CiMaterialId:  materialEntity.Id,
		Type:          materialEntity.Type,
		Value:         materialEntity.Value,
		GitTag:        materialEntity.GitTag,
		GitMaterialId: materialEntity.GitMaterialId,
		Active:        materialEntity.Active,
		GitMaterial:   ConvertToGitMaterialBean(materialEntity.GitMaterial),
		GitOptions: bean.GitOptions{
			UserName:      materialEntity.GitMaterial.GitProvider.UserName,
			Password:      materialEntity.GitMaterial.GitProvider.Password,
			AccessToken:   materialEntity.GitMaterial.GitProvider.AccessToken,
			SshPrivateKey: materialEntity.GitMaterial.GitProvider.SshPrivateKey,
			AuthMode:      materialEntity.GitMaterial.GitProvider.AuthMode,
		},
	}
	return model
}

func ConvertToGitMaterialBean(gitMaterialEntity *pc.GitMaterial) *bean2.GitMaterialModel {
	return &bean2.GitMaterialModel{
		Id:              gitMaterialEntity.Id,
		Name:            gitMaterialEntity.Name,
		Url:             gitMaterialEntity.Url,
		GitProviderId:   gitMaterialEntity.GitProviderId,
		CheckoutPath:    gitMaterialEntity.CheckoutPath,
		FetchSubmodules: gitMaterialEntity.FetchSubmodules,
	}
}
