package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/imageTagging/mocks"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	mocks3 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImageTaggingServie(t *testing.T) {
	sugaredLogger, err := util.NewSugardLogger()
	assert.True(t, err == nil, err)
	mockedImageTaggingRepo := mocks.NewImageTaggingRepository(t)
	mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
	mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)
	mockedImageTaggingRepo.On("CommitTx", &pg.Tx{}).Return(nil)

	mockedCiPipelineRepo := mocks2.NewCiPipelineRepository(t)
	mockedCdPipelineRepo := mocks2.NewPipelineRepository(t)
	mockedEnvironmentRepo := mocks3.NewEnvironmentRepository(t)
	imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, mockedCiPipelineRepo, mockedCdPipelineRepo, mockedEnvironmentRepo, sugaredLogger)
	t.Run("", func(tt *testing.T) {
		req := &ImageTaggingRequestDTO{}
		res, err := imageTaggingService.CreateOrUpdateImageTagging(0, 0, 0, 0, req)
		assert.Nil(tt, res)
		assert.Nil(tt, err)
	})
}
