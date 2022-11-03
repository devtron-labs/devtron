package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
)

type GitMaterialHistoryService interface {
	CreateMaterialHistory(inputMaterial *pipelineConfig.GitMaterial) error
	//UpdateMaterialHistory(updateMaterialDTO *pipelineConfig.GitMaterial) error
	CreateDeleteMaterialHistory(materials []*pipelineConfig.GitMaterial) error
}

type GitMaterialHistoryServiceImpl struct {
	gitMaterialHistoryRepository repository.GitMaterialHistoryRepository
	logger                       *zap.SugaredLogger
}

func NewGitMaterialHistoryServiceImpl(gitMaterialHistoryRepository repository.GitMaterialHistoryRepository,
	logger *zap.SugaredLogger) *GitMaterialHistoryServiceImpl {

	return &GitMaterialHistoryServiceImpl{
		gitMaterialHistoryRepository: gitMaterialHistoryRepository,
		logger:                       logger,
	}
}

func (impl GitMaterialHistoryServiceImpl) CreateMaterialHistory(inputMaterial *pipelineConfig.GitMaterial) error {

	material := &repository.GitMaterialHistory{
		Url:             inputMaterial.Url,
		AppId:           inputMaterial.AppId,
		Name:            inputMaterial.Name,
		GitProviderId:   inputMaterial.GitProviderId,
		Active:          true,
		CheckoutPath:    inputMaterial.CheckoutPath,
		FetchSubmodules: inputMaterial.FetchSubmodules,
		AuditLog:        sql.AuditLog{UpdatedBy: inputMaterial.UpdatedBy, CreatedBy: inputMaterial.CreatedBy, UpdatedOn: inputMaterial.UpdatedOn, CreatedOn: inputMaterial.CreatedOn},
	}

	err := impl.gitMaterialHistoryRepository.SaveGitMaterialHistory(material)

	if err != nil {
		impl.logger.Errorw("error in saving create/update history for git repository")
		return err
	}

	return err

}

//func (impl GitMaterialHistoryServiceImpl) UpdateMaterialHistory(updateMaterialDTO *pipelineConfig.GitMaterial) error {
//
//	updateMaterial := &repository.GitMaterialHistory{
//		Id:              updateMaterialDTO.Id,
//		AppId:           updateMaterialDTO.AppId,
//		GitProviderId:   updateMaterialDTO.GitProviderId,
//		Active:          updateMaterialDTO.Active,
//		Url:             updateMaterialDTO.Url,
//		Name:            updateMaterialDTO.Name,
//		CheckoutPath:    updateMaterialDTO.CheckoutPath,
//		FetchSubmodules: updateMaterialDTO.FetchSubmodules,
//		AuditLog:        sql.AuditLog{UpdatedBy: updateMaterialDTO.AuditLog.UpdatedBy, CreatedBy: updateMaterialDTO.AuditLog.CreatedBy, UpdatedOn: time.Now(), CreatedOn: time.Now()},
//	}
//
//	err := impl.gitMaterialHistoryRepository.Sa(updateMaterial)
//
//	if err != nil {
//		impl.logger.Errorw("Error in saving history of update action on git material")
//		return err
//	}
//
//	return nil
//}

func (impl GitMaterialHistoryServiceImpl) CreateDeleteMaterialHistory(materials []*pipelineConfig.GitMaterial) error {

	materialsHistory := []*repository.GitMaterialHistory{}

	for _, material := range materials {

		materialHistory := repository.GitMaterialHistory{
			Id:              material.Id,
			AppId:           material.AppId,
			GitProviderId:   material.GitProviderId,
			Active:          material.Active,
			Url:             material.Url,
			Name:            material.Name,
			CheckoutPath:    material.CheckoutPath,
			FetchSubmodules: material.FetchSubmodules,
			AuditLog: sql.AuditLog{
				CreatedOn: material.CreatedOn,
				CreatedBy: material.CreatedBy,
				UpdatedOn: material.UpdatedOn,
				UpdatedBy: material.UpdatedBy,
			},
		}
		materialsHistory = append(materialsHistory, &materialHistory)
	}

	err := impl.gitMaterialHistoryRepository.SaveDeleteMaterialHistory(materialsHistory)

	if err != nil {
		impl.logger.Errorw("Error in saving delete history for git material Repository")
		return err
	}

	return nil

}
