package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitMaterialService(t *testing.T) {

	t.Run("Save", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		mockedGitMaterialHistoryRepository := mocks.NewGitMaterialHistoryRepository(t)

		GitHistoryServiceImpl := NewGitMaterialHistoryServiceImpl(mockedGitMaterialHistoryRepository, sugaredLogger)

		GitMaterial := &pipelineConfig.GitMaterial{
			Id:              1,
			Url:             "https://github.com/devtron-labs/ci-runner",
			AppId:           49,
			Name:            "1-devtron-test",
			Active:          true,
			CheckoutPath:    "./",
			FetchSubmodules: false,
			AuditLog:        sql.AuditLog{},
		}

		mockedMaterial := &repository.GitMaterialHistory{
			GitMaterialId:   1,
			Url:             "https://github.com/devtron-labs/ci-runner",
			AppId:           49,
			Name:            "1-devtron-test",
			Active:          true,
			CheckoutPath:    "./",
			FetchSubmodules: false,
			AuditLog:        sql.AuditLog{},
		}

		mockedGitMaterialHistoryRepository.On("SaveGitMaterialHistory", mockedMaterial).Return(nil)

		err = GitHistoryServiceImpl.CreateMaterialHistory(GitMaterial)

		assert.Nil(t, err)

	})

	t.Run("MarkMaterialDelete", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		mockedGitMaterialHistoryRepository := mocks.NewGitMaterialHistoryRepository(t)

		GitHistoryServiceImpl := NewGitMaterialHistoryServiceImpl(mockedGitMaterialHistoryRepository, sugaredLogger)

		mockedMaterial := &repository.GitMaterialHistory{
			GitMaterialId:   1,
			Url:             "https://github.com/devtron-labs/ci-runner",
			AppId:           49,
			Name:            "1-devtron-test",
			Active:          false,
			CheckoutPath:    "./",
			FetchSubmodules: false,
			AuditLog:        sql.AuditLog{},
		}

		GitMaterial := &pipelineConfig.GitMaterial{
			Id:              1,
			Url:             "https://github.com/devtron-labs/ci-runner",
			AppId:           49,
			Name:            "1-devtron-test",
			Active:          true,
			CheckoutPath:    "./",
			FetchSubmodules: false,
			AuditLog:        sql.AuditLog{},
		}

		mockedGitMaterialHistoryRepository.On("SaveGitMaterialHistory", mockedMaterial).Return(nil)

		err = GitHistoryServiceImpl.MarkMaterialDeletedAndCreateHistory(GitMaterial)

		assert.Nil(t, err)

	})

}
