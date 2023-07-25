package mocks

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	repository3 "github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	mocks2 "github.com/devtron-labs/devtron/pkg/user/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSave(t *testing.T) {

}

func TestGetGenericNotesForAppIds(t *testing.T) {
	appId1, appId2 := 1, 2
	var userId1, userId2 int32 = 1, 2
	testAppIds := []int{appId1, appId2}
	testUserIds := []int32{userId1, userId2}
	testUsers := []repository.UserModel{
		{
			Id:      testUserIds[0],
			EmailId: "email-1",
		},
		{
			Id:      testUserIds[1],
			EmailId: "email-2",
		},
	}

	genericNoteResp := &repository3.GenericNote{
		Id:             1,
		Identifier:     appId1,
		IdentifierType: repository3.AppType,
		Description:    "test-response-1",
		AuditLog: sql.AuditLog{
			UpdatedBy: testUserIds[0],
		},
	}

	descriptionResp := &repository3.GenericNote{
		Id:             0,
		Identifier:     appId2,
		IdentifierType: repository3.AppType,
		Description:    "app-description-2",
		AuditLog: sql.AuditLog{
			UpdatedBy: testUserIds[1],
		},
	}

	getGenericNotesForAppIdsResp := []*repository3.GenericNote{genericNoteResp}
	getDescriptionForAppIdsResp := []*repository3.GenericNote{descriptionResp}
	t.Run("Test Error Case, error from GetGenericNotesForAppIds", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(t)

		testErr := errors.New("GetGenericNotesForAppIds error")
		mockedNoteRepo.On("GetGenericNotesForAppIds", mock.AnythingOfType("[]int")).Return(nil, testErr)
		resp, err := genericNoteSvc.GetGenericNotesForAppIds(testAppIds)
		assert.NotNil(tt, resp)
		assert.NotNil(tt, err)
		assert.Equal(tt, 0, len(resp))
		assert.Equal(tt, testErr, err)
	})

	t.Run("Test Error Case, error from GetDescriptionFromAppIds", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(t)
		testErr := errors.New("GetDescriptionFromAppIds error")
		mockedNoteRepo.On("GetGenericNotesForAppIds", mock.AnythingOfType("[]int")).Return(getGenericNotesForAppIdsResp, nil)
		mockedNoteRepo.On("GetDescriptionFromAppIds", mock.AnythingOfType("[]int")).Return(nil, testErr)
		resp, err := genericNoteSvc.GetGenericNotesForAppIds(testAppIds)
		noteRespBean := resp[appId1]
		assert.NotNil(tt, resp)
		assert.Equal(tt, 1, len(resp))
		assert.NotNil(tt, noteRespBean)
		assert.Equal(tt, genericNoteResp.Id, noteRespBean.Id)
		assert.Equal(tt, genericNoteResp.Description, noteRespBean.Description)
		assert.NotNil(tt, err)
		assert.Equal(tt, testErr, err)
	})

	t.Run("Test Error Case, error from GetUsersByIds", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, mockedUserRepo := initGenericNoteService(t)
		testErr := errors.New("GetUsersByIds error")
		mockedNoteRepo.On("GetGenericNotesForAppIds", mock.AnythingOfType("[]int")).Return(getGenericNotesForAppIdsResp, nil)
		mockedNoteRepo.On("GetDescriptionFromAppIds", mock.AnythingOfType("[]int")).Return(getDescriptionForAppIdsResp, nil)
		mockedUserRepo.On("GetByIds", mock.AnythingOfType("[]int32")).Return(nil, testErr)
		resp, err := genericNoteSvc.GetGenericNotesForAppIds(testAppIds)

		assert.NotNil(tt, resp)
		assert.Equal(tt, 1, len(resp))
		assert.NotNil(tt, err)
		assert.Equal(tt, testErr, err)

		noteRespBean := resp[appId1]
		assert.NotNil(tt, noteRespBean)
		assert.Equal(tt, genericNoteResp.Id, noteRespBean.Id)
		assert.Equal(tt, genericNoteResp.Description, noteRespBean.Description)

	})

	t.Run("", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, mockedUserRepo := initGenericNoteService(t)

		mockedNoteRepo.On("GetGenericNotesForAppIds", mock.AnythingOfType("[]int")).Return(getGenericNotesForAppIdsResp, nil)
		mockedNoteRepo.On("GetDescriptionFromAppIds", mock.AnythingOfType("[]int")).Return(getDescriptionForAppIdsResp, nil)
		mockedUserRepo.On("GetByIds", mock.AnythingOfType("[]int32")).Return(testUsers, nil)
		resp, err := genericNoteSvc.GetGenericNotesForAppIds(testAppIds)

		assert.NotNil(tt, resp)
		assert.Equal(tt, 2, len(resp))
		assert.Nil(tt, err)

		noteRespBean := resp[appId1]
		assert.NotNil(tt, noteRespBean)
		assert.Equal(tt, genericNoteResp.Id, noteRespBean.Id)
		assert.Equal(tt, genericNoteResp.Description, noteRespBean.Description)
		assert.Equal(tt, testUsers[0].EmailId, noteRespBean.UpdatedBy)

		descriptionRespBean := resp[appId2]
		assert.NotNil(tt, descriptionRespBean)
		assert.Equal(tt, descriptionResp.Id, descriptionRespBean.Id)
		assert.Equal(tt, descriptionResp.Description, descriptionRespBean.Description)
		assert.Equal(tt, testUsers[1].EmailId, descriptionRespBean.UpdatedBy)
	})
}

func initGenericNoteService(t *testing.T) (*genericNotes.GenericNoteServiceImpl, *mocks.GenericNoteRepository, *GenericNoteHistoryService, *mocks2.UserRepository) {
	logger, err := util.NewSugardLogger()
	if err != nil {
		assert.Fail(t, "error in creating logger")
	}

	mockedNoteRepo := mocks.NewGenericNoteRepository(t)
	mockedHistoryService := NewGenericNoteHistoryService(t)
	mockedUserRepo := mocks2.NewUserRepository(t)
	genericNoteSvc := genericNotes.NewGenericNoteServiceImpl(mockedNoteRepo, mockedHistoryService, mockedUserRepo, logger)
	return genericNoteSvc, mockedNoteRepo, mockedHistoryService, mockedUserRepo
}
