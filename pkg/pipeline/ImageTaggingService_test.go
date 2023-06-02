package pipeline

import (
	"errors"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/imageTagging/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	mocks3 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestCreateOrUpdateImageTagging(t *testing.T) {
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

	appId, artifactId, ciPipelineId, userId := 1, 1, 1, 2
	testPayload := &ImageTaggingRequestDTO{
		CreateTags: []*repository.ImageTag{{
			TagName:    "devtron-v1.1.6",
			AppId:      appId,
			ArtifactId: artifactId,
		}},
		SoftDeleteTags: []*repository.ImageTag{{
			Id:         1,
			TagName:    "devtron-v1.1.5",
			AppId:      appId,
			ArtifactId: artifactId,
		}},
		HardDeleteTags: []*repository.ImageTag{{
			Id:         2,
			TagName:    "devtron-v1.1...5",
			AppId:      appId,
			ArtifactId: artifactId,
		}},
		ImageComment: repository.ImageComment{
			Comment:    "hello devtron!",
			ArtifactId: artifactId,
		},
	}

	mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(testPayload.ImageComment, nil)
	//mockedImageTaggingRepo.On("UpdateImageComment", &pg.Tx{}, mock.Anything).Return(nil)
	mockedImageTaggingRepo.On("SaveImageComment", &pg.Tx{}, mock.Anything).Return(nil)
	mockedImageTaggingRepo.On("SaveAuditLogsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
	mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
	mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
	mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
	mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(testPayload.ImageComment, nil)
	mockedImageTaggingRepo.On("GetTagsByAppId", appId).Return(append(testPayload.SoftDeleteTags, testPayload.CreateTags...), nil)
	mockedImageTaggingRepo.On("GetTagsByArtifactId", artifactId).Return(append(testPayload.SoftDeleteTags, testPayload.CreateTags...), nil)

	mockedCiPipelineRepo.On("FindByParentCiPipelineId", ciPipelineId).Return([]*pipelineConfig.CiPipeline{}, nil)

	mockedEnvironmentRepo.On("FindEnvLinkedWithCiPipelines", []int{ciPipelineId}).Return(nil, nil)

	t.Run("Valid Request, No error from repo", func(tt *testing.T) {

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.NotNil(tt, res)
		assert.Nil(tt, err)
		assert.Equal(tt, false, res.ProdEnvExists)
		assert.Equal(tt, len(testPayload.CreateTags)+len(testPayload.SoftDeleteTags), len(res.ImageReleaseTags))
		assert.Equal(tt, len(testPayload.CreateTags)+len(testPayload.SoftDeleteTags), len(res.ImageReleaseTags))
	})

	t.Run("Valid Request, error in SoftDeleting Tags", func(tt *testing.T) {
		softDeleteError := errors.New("error in updating image tags")
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(softDeleteError)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), softDeleteError.Error())
	})

	t.Run("Valid Request, error in HardDeleting Tags", func(tt *testing.T) {
		hardDeleteError := errors.New("error in HardDeleting image tags")
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(hardDeleteError)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), hardDeleteError.Error())
	})
	t.Run("Valid Request, error in saveComment", func(tt *testing.T) {
		saveCommentError := errors.New("error in saveComment")
		mockedImageTaggingRepo.On("SaveImageComment", &pg.Tx{}, mock.Anything).Return(saveCommentError)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), saveCommentError.Error())
	})
	t.Run("Valid Request, error in SaveAuditLogsInBulk", func(tt *testing.T) {
		saveAuditLogsInBulkError := errors.New("error in saveAuditLogsInBulk ")
		mockedImageTaggingRepo.On("SaveAuditLogsInBulk", &pg.Tx{}, mock.Anything).Return(saveAuditLogsInBulkError)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), saveAuditLogsInBulkError.Error())
	})
	t.Run("Valid Request, error in HardDeleting Tags", func(tt *testing.T) {
		t.SkipNow()
		hardDeleteError := errors.New("error in updating image tags")
		mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(testPayload.ImageComment, nil)
		mockedImageTaggingRepo.On("UpdateImageComment", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveImageComment", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveAuditLogsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(hardDeleteError)
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(testPayload.ImageComment, nil)
		mockedImageTaggingRepo.On("GetTagsByAppId", appId).Return(append(testPayload.SoftDeleteTags, testPayload.CreateTags...), nil)
		mockedImageTaggingRepo.On("GetTagsByArtifactId", artifactId).Return(append(testPayload.SoftDeleteTags, testPayload.CreateTags...), nil)

		mockedCiPipelineRepo.On("FindByParentCiPipelineId", ciPipelineId).Return([]*pipelineConfig.CiPipeline{}, nil)

		mockedEnvironmentRepo.On("FindEnvLinkedWithCiPipelines", []int{ciPipelineId}).Return(nil, nil)
		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), hardDeleteError.Error())
	})

}
