/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipeline

import (
	"errors"
	"fmt"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/imageTagging/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	repository1 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	mocks3 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestImageTaggingService(t *testing.T) {

	//test data and mocks intialisation
	sugaredLogger, err := util.NewSugardLogger()
	assert.True(t, err == nil, err)

	appId, artifactId, ciPipelineId, userId := 1, 3, 1, 2
	testPayload := &types.ImageTaggingRequestDTO{
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
		ExternalCi: false,
	}

	initRepos := func() (ImageTaggingService, *mocks.ImageTaggingRepository, *mocks2.CiPipelineRepository, *mocks2.PipelineRepository, *mocks3.EnvironmentRepository) {
		mockedImageTaggingRepo := mocks.NewImageTaggingRepository(t)
		mockedCiPipelineRepo := mocks2.NewCiPipelineRepository(t)
		mockedCdPipelineRepo := mocks2.NewPipelineRepository(t)
		mockedEnvironmentRepo := mocks3.NewEnvironmentRepository(t)
		imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, mockedCiPipelineRepo, mockedCdPipelineRepo, mockedEnvironmentRepo, sugaredLogger)
		//mockedImageTaggingRepo.On("UpdateImageComment", &pg.Tx{}, mock.Anything).Return(nil)
		return imageTaggingService, mockedImageTaggingRepo, mockedCiPipelineRepo, mockedCdPipelineRepo, mockedEnvironmentRepo
	}

	//test cases starts here
	t.Run("CreateOrUpdateImageTagging,Valid payload Request,expected: valid response with nil error", func(tt *testing.T) {
		//tt.SkipNow()

		imageTaggingService, mockedImageTaggingRepo, mockedCiPipelineRepo, _, mockedEnvironmentRepo := initRepos()
		mockedImageTaggingRepo.On("SaveImageComment", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveAuditLogsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(testPayload.ImageComment, nil)
		mockedImageTaggingRepo.On("GetTagsByAppId", appId).Return(append(testPayload.SoftDeleteTags, testPayload.CreateTags...), nil)
		mockedImageTaggingRepo.On("getTagsByArtifactId", artifactId).Return(append(testPayload.SoftDeleteTags, testPayload.CreateTags...), nil)
		mockedCiPipelineRepo.On("FindByParentCiPipelineId", ciPipelineId).Return([]*pipelineConfig.CiPipeline{}, nil)
		mockedEnvironmentRepo.On("FindEnvLinkedWithCiPipelines", testPayload.ExternalCi, []int{ciPipelineId}).Return(nil, nil)
		mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
		mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)
		mockedImageTaggingRepo.On("CommitTx", &pg.Tx{}).Return(nil)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.NotNil(tt, res)
		assert.Nil(tt, err)
		assert.Equal(tt, false, res.ProdEnvExists)
		assert.Equal(tt, len(testPayload.CreateTags)+len(testPayload.SoftDeleteTags), len(res.ImageReleaseTags))
		assert.Equal(tt, len(testPayload.CreateTags)+len(testPayload.SoftDeleteTags), len(res.ImageReleaseTags))
	})

	t.Run("CreateOrUpdateImageTagging,Valid Request payload, error in SoftDeleting Tags,expected: nil response with non nil error", func(tt *testing.T) {
		//tt.SkipNow()

		imageTaggingService, mockedImageTaggingRepo, _, _, _ := initRepos()
		softDeleteError := errors.New("error in updating image tags")
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(softDeleteError)
		mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
		mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), softDeleteError.Error())
	})

	t.Run("Valid Request, error in HardDeleting Tags", func(tt *testing.T) {
		//tt.SkipNow()

		imageTaggingService, mockedImageTaggingRepo, _, _, _ := initRepos()
		hardDeleteError := errors.New("error in HardDeleting image tags")
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(hardDeleteError)
		mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
		mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), hardDeleteError.Error())
	})
	t.Run("Valid Request, error in saveComment", func(tt *testing.T) {
		//tt.SkipNow()

		imageTaggingService, mockedImageTaggingRepo, _, _, _ := initRepos()
		saveCommentError := errors.New("error in saveComment")
		mockedImageTaggingRepo.On("SaveImageComment", &pg.Tx{}, mock.Anything).Return(saveCommentError)
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
		mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(repository.ImageComment{}, nil)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), saveCommentError.Error())
	})
	t.Run("Valid Request, error in SaveAuditLogsInBulk", func(tt *testing.T) {
		//tt.SkipNow()

		imageTaggingService, mockedImageTaggingRepo, _, _, _ := initRepos()
		saveAuditLogsInBulkError := errors.New("error in saveAuditLogsInBulk ")
		mockedImageTaggingRepo.On("SaveAuditLogsInBulk", &pg.Tx{}, mock.Anything).Return(saveAuditLogsInBulkError)
		mockedImageTaggingRepo.On("SaveImageComment", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
		mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(repository.ImageComment{}, nil)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), saveAuditLogsInBulkError.Error())
	})

	t.Run("Valid Request, error in GetTagsByAppId", func(tt *testing.T) {
		//tt.SkipNow()

		imageTaggingService, mockedImageTaggingRepo, _, _, _ := initRepos()
		GetTagsByAppIdError := errors.New("error in GetTagsByAppId ")
		mockedImageTaggingRepo.On("GetTagsByAppId", appId).Return(nil, GetTagsByAppIdError)
		mockedImageTaggingRepo.On("SaveAuditLogsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveImageComment", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
		mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)
		mockedImageTaggingRepo.On("CommitTx", &pg.Tx{}).Return(nil)
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(repository.ImageComment{}, nil)

		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		assert.NotNil(tt, res)
		assert.Equal(tt, 0, len(res.AppReleaseTags))
		assert.Nil(tt, res.ImageComment)
		assert.Equal(tt, 0, len(res.ImageReleaseTags))
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), GetTagsByAppIdError.Error())
	})

	t.Run("Valid Request, error in GetImageComment", func(tt *testing.T) {
		//tt.SkipNow()
		imageTaggingService, mockedImageTaggingRepo, _, _, _ := initRepos()
		GetImageCommentError := errors.New("error in GetImageComment ")
		mockedImageTaggingRepo.On("GetImageComment", artifactId).Return(repository.ImageComment{}, GetImageCommentError)
		mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("StartTx").Return(&pg.Tx{}, nil)
		mockedImageTaggingRepo.On("RollbackTx", &pg.Tx{}).Return(nil)
		mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", &pg.Tx{}, mock.Anything).Return(nil)
		mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", &pg.Tx{}, mock.Anything).Return(nil)

		testPayload.ImageComment.Id = 1
		res, err := imageTaggingService.CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId, testPayload)
		testPayload.ImageComment.Id = 0
		assert.Nil(tt, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, err.Error(), GetImageCommentError.Error())
	})
	t.Run("TestGetProdEnvByCdPipelineId error check", func(tt *testing.T) {
		mockedCdPipelineRepo := mocks2.NewPipelineRepository(tt)
		testPipelineId := 0
		testError := fmt.Sprintf("error in fetching pipeline by pipelineId %v", testPipelineId)
		mockedCdPipelineRepo.On("FindById", testPipelineId).Return(nil, errors.New(testError))
		imageTaggingService := NewImageTaggingServiceImpl(nil, nil, mockedCdPipelineRepo, nil, sugaredLogger)
		res, err := imageTaggingService.GetProdEnvByCdPipelineId(testPipelineId)
		assert.Equal(tt, false, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, testError, err.Error())
	})

	t.Run("TestGetProdEnvByCdPipelineId cd env with default true", func(tt *testing.T) {
		mockedCdPipelineRepo := mocks2.NewPipelineRepository(tt)
		testPipelineId := 1
		testCiPipelineId := 1
		testPipeline := &pipelineConfig.Pipeline{
			Id:           testPipelineId,
			CiPipelineId: testCiPipelineId,
			Environment: repository1.Environment{
				Default: true,
			},
		}
		//testError := fmt.Sprintf("error in fetching pipeline by pipelineId %v", testPipelineId)
		mockedCdPipelineRepo.On("FindById", testPipelineId).Return(testPipeline, nil)
		imageTaggingService := NewImageTaggingServiceImpl(nil, nil, mockedCdPipelineRepo, nil, sugaredLogger)
		res, err := imageTaggingService.GetProdEnvByCdPipelineId(testPipelineId)
		assert.Equal(tt, true, res)
		assert.Nil(tt, err)
	})

	t.Run("TestGetProdEnvByCdPipelineId,FindByParentCiPipelineId throws error", func(tt *testing.T) {
		mockedCdPipelineRepo := mocks2.NewPipelineRepository(tt)
		mockedCiPipelineRepo := mocks2.NewCiPipelineRepository(tt)
		//mockedEnvRepo := mocks3.NewEnvironmentRepository(tt)
		testPipelineId := 1
		testCiPipelineId := 2
		testPipeline := &pipelineConfig.Pipeline{
			Id:           testPipelineId,
			CiPipelineId: testCiPipelineId,
			Environment: repository1.Environment{
				Default: false,
			},
		}
		testError := fmt.Sprintf("error in fetching ciPipelines with parent ciPipelineId %v", testCiPipelineId)
		//testCiPipelinesResponse := []bean.CiPipeline
		mockedCdPipelineRepo.On("FindById", testPipelineId).Return(testPipeline, nil)
		mockedCiPipelineRepo.On("FindByParentCiPipelineId", testCiPipelineId).Return(nil, errors.New(testError))
		imageTaggingService := NewImageTaggingServiceImpl(nil, mockedCiPipelineRepo, mockedCdPipelineRepo, nil, sugaredLogger)
		res, err := imageTaggingService.GetProdEnvByCdPipelineId(testPipelineId)
		assert.Equal(tt, false, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, testError, err.Error())
	})

	t.Run("TestGetProdEnvByCdPipelineId,FindByParentCiPipelineId throws error", func(tt *testing.T) {
		mockedCdPipelineRepo := mocks2.NewPipelineRepository(tt)
		mockedCiPipelineRepo := mocks2.NewCiPipelineRepository(tt)
		mockedEnvRepo := mocks3.NewEnvironmentRepository(tt)
		testPipelineId := 1
		testCiPipelineId := 2
		testPipeline := &pipelineConfig.Pipeline{
			Id:           testPipelineId,
			CiPipelineId: testCiPipelineId,
			Environment: repository1.Environment{
				Default: false,
			},
		}

		testCipipelinesResp := []*pipelineConfig.CiPipeline{
			{
				Id:   4,
				Name: "test-ci-pipeline-4",
			}, {
				Id:   5,
				Name: "test-ci-pipeline-5",
			},
			{
				Id:   6,
				Name: "test-ci-pipeline-6",
			},
		}

		testCipipelineIds := []int{testCiPipelineId}
		for _, cipip := range testCipipelinesResp {
			testCipipelineIds = append(testCipipelineIds, cipip.Id)
		}
		ciPipelineIdsString := "4,5,6"
		testError := fmt.Sprintf("error in fetching environemts for ciPipelineIds %v", ciPipelineIdsString)
		//testCiPipelinesResponse := []bean.CiPipeline
		mockedCdPipelineRepo.On("FindById", testPipelineId).Return(testPipeline, nil)
		mockedCiPipelineRepo.On("FindByParentCiPipelineId", testCiPipelineId).Return(testCipipelinesResp, nil)
		mockedEnvRepo.On("FindEnvLinkedWithCiPipelines", false, testCipipelineIds).Return(nil, errors.New(testError))

		imageTaggingService := NewImageTaggingServiceImpl(nil, mockedCiPipelineRepo, mockedCdPipelineRepo, mockedEnvRepo, sugaredLogger)
		res, err := imageTaggingService.GetProdEnvByCdPipelineId(testPipelineId)
		assert.Equal(tt, false, res)
		assert.NotNil(tt, err)
		assert.Equal(tt, testError, err.Error())
	})

	t.Run("TestGetProdEnvByCdPipelineId,valid flow with no prod env", func(tt *testing.T) {
		mockedCdPipelineRepo := mocks2.NewPipelineRepository(tt)
		mockedCiPipelineRepo := mocks2.NewCiPipelineRepository(tt)
		mockedEnvRepo := mocks3.NewEnvironmentRepository(tt)
		testPipelineId := 1
		testCiPipelineId := 2
		testPipeline := &pipelineConfig.Pipeline{
			Id:           testPipelineId,
			CiPipelineId: testCiPipelineId,
			Environment: repository1.Environment{
				Default: false,
			},
		}

		testCipipelinesResp := []*pipelineConfig.CiPipeline{
			{
				Id:   4,
				Name: "test-ci-pipeline-4",
			}, {
				Id:   5,
				Name: "test-ci-pipeline-5",
			},
			{
				Id:   6,
				Name: "test-ci-pipeline-6",
			},
		}
		testEnvs := []*repository1.Environment{
			{
				Id:      101,
				Default: false,
			},
			{
				Id:      102,
				Default: false,
			},
		}
		testCipipelineIds := []int{testCiPipelineId}
		for _, cipip := range testCipipelinesResp {
			testCipipelineIds = append(testCipipelineIds, cipip.Id)
		}
		//mock functions
		mockedCdPipelineRepo.On("FindById", testPipelineId).Return(testPipeline, nil)
		mockedCiPipelineRepo.On("FindByParentCiPipelineId", testCiPipelineId).Return(testCipipelinesResp, nil)
		mockedEnvRepo.On("FindEnvLinkedWithCiPipelines", false, testCipipelineIds).Return(testEnvs, nil)

		//test the service
		imageTaggingService := NewImageTaggingServiceImpl(nil, mockedCiPipelineRepo, mockedCdPipelineRepo, mockedEnvRepo, sugaredLogger)
		res, err := imageTaggingService.GetProdEnvByCdPipelineId(testPipelineId)
		assert.Equal(tt, false, res)
		assert.Nil(tt, err)
	})

	t.Run("TestGetProdEnvByCdPipelineId,valid flow with prod env", func(tt *testing.T) {
		mockedCdPipelineRepo := mocks2.NewPipelineRepository(tt)
		mockedCiPipelineRepo := mocks2.NewCiPipelineRepository(tt)
		mockedEnvRepo := mocks3.NewEnvironmentRepository(tt)
		testPipelineId := 1
		testCiPipelineId := 2
		testPipeline := &pipelineConfig.Pipeline{
			Id:           testPipelineId,
			CiPipelineId: testCiPipelineId,
			Environment: repository1.Environment{
				Default: false,
			},
		}

		testCipipelinesResp := []*pipelineConfig.CiPipeline{
			{
				Id:               4,
				Name:             "test-ci-pipeline-4",
				ParentCiPipeline: testCiPipelineId,
			}, {
				Id:               5,
				Name:             "test-ci-pipeline-5",
				ParentCiPipeline: testCiPipelineId,
			},
			{
				Id:               6,
				Name:             "test-ci-pipeline-6",
				ParentCiPipeline: testCiPipelineId,
			},
		}
		testEnvs := []*repository1.Environment{
			{
				Id:      101,
				Default: false,
			},
			{
				Id:      102,
				Default: true,
			},
		}
		testCipipelineIds := []int{testCiPipelineId}
		for _, cipip := range testCipipelinesResp {
			testCipipelineIds = append(testCipipelineIds, cipip.Id)
		}
		mockedCdPipelineRepo.On("FindById", testPipelineId).Return(testPipeline, nil)
		mockedCiPipelineRepo.On("FindByParentCiPipelineId", testCiPipelineId).Return(testCipipelinesResp, nil)
		mockedEnvRepo.On("FindEnvLinkedWithCiPipelines", false, testCipipelineIds).Return(testEnvs, nil)
		imageTaggingService := NewImageTaggingServiceImpl(nil, mockedCiPipelineRepo, mockedCdPipelineRepo, mockedEnvRepo, sugaredLogger)
		res, err := imageTaggingService.GetProdEnvByCdPipelineId(testPipelineId)
		assert.Equal(tt, true, res)
		assert.Nil(tt, err)
	})

	t.Run("tagNameValidation", func(tt *testing.T) {
		errString := "tag name should be max of 128 characters long,tag name should not start with '.' and '-'"
		tt.Run("valid tag name", func(ttt *testing.T) {
			//valid tag name
			testTagName := "v1.1"
			returnErr := tagNameValidation(testTagName)
			assert.Nil(tt, returnErr)
		})

		tt.Run("invalid tag name ,starts with '.'", func(ttt *testing.T) {
			//invalid tag name 1
			testTagName := ".django"
			returnErr := tagNameValidation(testTagName)
			assert.NotNil(tt, returnErr)
			assert.Equal(tt, errString, returnErr.Error())
		})

		tt.Run("invalid tag name, starts with '-'", func(ttt *testing.T) {
			//invalid tag name 2
			testTagName := "-django"
			returnErr := tagNameValidation(testTagName)
			assert.NotNil(tt, returnErr)
			assert.Equal(tt, errString, returnErr.Error())
		})

		tt.Run("invalid tag name, empty tag", func(ttt *testing.T) {
			//invalid tag name 3
			testTagName := ""
			returnErr := tagNameValidation(testTagName)
			assert.NotNil(tt, returnErr)
			assert.Equal(tt, errString, returnErr.Error())
		})

		tt.Run("invalid tag name, tag have more than 128 chars", func(ttt *testing.T) {
			//invalid tag name 4

			testTagName := ""
			for i := 0; i < 129; i++ {
				testTagName += "v"
			}
			returnErr := tagNameValidation(testTagName)
			assert.NotNil(tt, returnErr)
			assert.Equal(tt, errString, returnErr.Error())
		})

	})

	t.Run("ValidateImageTaggingRequest", func(tt *testing.T) {
		testImageTaggingRequest := types.ImageTaggingRequestDTO{
			ImageComment: repository.ImageComment{
				Id:      1,
				Comment: "test Comment",
			},
			CreateTags: []*repository.ImageTag{{
				TagName:    "v1",
				AppId:      appId,
				ArtifactId: artifactId,
			}},
			SoftDeleteTags: []*repository.ImageTag{{
				Id:      1,
				TagName: "v1",
			}},
			HardDeleteTags: []*repository.ImageTag{{
				Id:      1,
				TagName: "v1",
			}},
		}
		testTagValidationResp := func(ttt *testing.T, err error, valid bool) {
			assert.NotNil(ttt, err)
			assert.Equal(ttt, false, valid)
		}
		imageTaggingService := NewImageTaggingServiceImpl(nil, nil, nil, nil, sugaredLogger)
		tt.Run("valid payload", func(ttt *testing.T) {

			valid, err := imageTaggingService.ValidateImageTaggingRequest(&testImageTaggingRequest, appId, artifactId)

			assert.Nil(tt, err)
			assert.Equal(tt, true, valid)
		})

		tt.Run("inValid payload,CreateTags", func(ttt *testing.T) {
			testImageTaggingRequest.CreateTags[0].Id = 1
			valid, err := imageTaggingService.ValidateImageTaggingRequest(&testImageTaggingRequest, appId, artifactId)
			testImageTaggingRequest.CreateTags[0].Id = 0
			testTagValidationResp(ttt, err, valid)
		})

		tt.Run("inValid payload,SoftDeleteTags", func(ttt *testing.T) {
			testImageTaggingRequest.SoftDeleteTags[0].Id = 0
			valid, err := imageTaggingService.ValidateImageTaggingRequest(&testImageTaggingRequest, appId, artifactId)
			testImageTaggingRequest.SoftDeleteTags[0].Id = 1
			testTagValidationResp(ttt, err, valid)
		})

		tt.Run("inValid payload,HardDeleteTags", func(ttt *testing.T) {
			testImageTaggingRequest.HardDeleteTags[0].Id = 0
			valid, err := imageTaggingService.ValidateImageTaggingRequest(&testImageTaggingRequest, appId, artifactId)
			testImageTaggingRequest.HardDeleteTags[0].Id = 1
			testTagValidationResp(ttt, err, valid)
		})

	})

	t.Run("GetTaggingDataMapByAppId", func(tt *testing.T) {
		testArtifactId1 := 1
		testTags := []*repository.ImageTag{
			{
				Id:         1,
				AppId:      appId,
				ArtifactId: testArtifactId1,
			},
		}

		testComments := []*repository.ImageComment{
			{
				Id:         2,
				ArtifactId: testArtifactId1,
				Comment:    "hello test1",
			},
		}

		tt.Run("GetTagsDataMapByAppId, no error case", func(ttt *testing.T) {
			mockedImageTaggingRepo := mocks.NewImageTaggingRepository(tt)
			mockedImageTaggingRepo.On("GetTagsByAppId", appId).Return(testTags, nil)
			imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, nil, nil, nil, sugaredLogger)
			resMap, err := imageTaggingService.GetTagsDataMapByAppId(appId)
			assert.NotNil(ttt, resMap)
			assert.Nil(ttt, err)
			assert.NotNil(ttt, resMap[testArtifactId1])
		})

		tt.Run("GetTagsDataMapByAppId,GetTagsByAppId repo func throws error, return same error", func(ttt *testing.T) {
			testErr := "error in GetTagsByAppId"
			mockedImageTaggingRepo := mocks.NewImageTaggingRepository(tt)
			mockedImageTaggingRepo.On("GetTagsByAppId", appId).Return(nil, errors.New(testErr))
			imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, nil, nil, nil, sugaredLogger)
			resMap, err := imageTaggingService.GetTagsDataMapByAppId(appId)
			assert.Nil(ttt, resMap)
			assert.NotNil(ttt, err)
			assert.Equal(ttt, testErr, err.Error())
			assert.Nil(ttt, resMap[testArtifactId1])
		})

		tt.Run("GetImageCommentsDataMapByArtifactIds,no error case", func(ttt *testing.T) {
			mockedImageTaggingRepo := mocks.NewImageTaggingRepository(tt)
			mockedImageTaggingRepo.On("GetImageCommentsByArtifactIds", []int{testArtifactId1}).Return(testComments, nil)
			imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, nil, nil, nil, sugaredLogger)
			resMap, err := imageTaggingService.GetImageCommentsDataMapByArtifactIds([]int{testArtifactId1})
			assert.NotNil(ttt, resMap)
			assert.Nil(ttt, err)
			assert.NotNil(ttt, resMap[testArtifactId1])
		})

		tt.Run("GetImageCommentsDataMapByArtifactIds,GetImageCommentsByAppId repo func throws error,return same error", func(ttt *testing.T) {
			testErr := "error in GetImageCommentsByAppId"
			mockedImageTaggingRepo := mocks.NewImageTaggingRepository(tt)
			mockedImageTaggingRepo.On("GetImageCommentsByArtifactIds", []int{testArtifactId1}).Return(nil, errors.New(testErr))
			imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, nil, nil, nil, sugaredLogger)
			resMap, err := imageTaggingService.GetImageCommentsDataMapByArtifactIds([]int{testArtifactId1})
			assert.Nil(ttt, resMap)
			assert.NotNil(ttt, err)
			assert.Equal(ttt, testErr, err.Error())
			assert.Nil(ttt, resMap[testArtifactId1])
		})

	})

	t.Run("getTagsByArtifactId, getTagsByArtifactId throws error", func(tt *testing.T) {
		testErr := "error in getTagsByArtifactId"
		mockedImageTaggingRepo := mocks.NewImageTaggingRepository(tt)
		mockedImageTaggingRepo.On("getTagsByArtifactId", artifactId).Return(nil, errors.New(testErr))
		imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, nil, nil, nil, sugaredLogger)
		tags, err := imageTaggingService.getTagsByArtifactId(artifactId)
		assert.Nil(tt, tags)
		assert.Equal(tt, 0, len(tags))
		assert.NotNil(tt, err)
		assert.Equal(tt, testErr, err.Error())
	})

	t.Run("performTagOperationsAndGetAuditList", func(tt *testing.T) {

		testImageTaggingRequest := types.ImageTaggingRequestDTO{
			ImageComment: repository.ImageComment{
				Id:      1,
				Comment: "test Comment",
			},
			CreateTags: []*repository.ImageTag{{
				TagName:    "v1",
				AppId:      appId,
				ArtifactId: artifactId,
			}},
			SoftDeleteTags: []*repository.ImageTag{{
				Id:      1,
				TagName: "v1",
			}},
			HardDeleteTags: []*repository.ImageTag{{
				Id:      1,
				TagName: "v1",
			}},
		}

		tt.Run("SaveReleaseTagsInBulk throws error", func(ttt *testing.T) {
			testErr := "error in SaveReleaseTagsInBulk"
			tx := &pg.Tx{}
			mockedImageTaggingRepo := mocks.NewImageTaggingRepository(ttt)
			mockedImageTaggingRepo.On("SaveReleaseTagsInBulk", tx, testImageTaggingRequest.CreateTags).Return(errors.New(testErr))
			mockedImageTaggingRepo.On("UpdateReleaseTagInBulk", tx, testImageTaggingRequest.SoftDeleteTags).Return(nil)
			mockedImageTaggingRepo.On("DeleteReleaseTagInBulk", tx, testImageTaggingRequest.HardDeleteTags).Return(nil)
			imageTaggingService := NewImageTaggingServiceImpl(mockedImageTaggingRepo, nil, nil, nil, sugaredLogger)
			res, err := imageTaggingService.performTagOperationsAndGetAuditList(tx, appId, artifactId, userId, &testImageTaggingRequest)
			assert.Nil(ttt, res)
			assert.NotNil(ttt, err)
			assert.Equal(ttt, testErr, err.Error())
		})

	})

}
