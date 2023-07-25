package mocks

import (
	"errors"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	repository3 "github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	mocks2 "github.com/devtron-labs/devtron/pkg/user/repository/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

var appId1, appId2 = 1, 2
var userId1, userId2 int32 = 1, 2
var testAppIds = []int{appId1, appId2}
var testUserIds = []int32{userId1, userId2}

var testUsers = []repository.UserModel{
	{
		Id:      testUserIds[0],
		EmailId: "email-1",
	},
	{
		Id:      testUserIds[1],
		EmailId: "email-2",
	},
}

func TestSave(t *testing.T) {

	testErrorAssertions := func(tt *testing.T, resp *bean2.GenericNoteResponseBean, testErr string, err error) {
		assert.NotNil(tt, err)
		assert.Nil(tt, resp)
		assert.Equal(tt, testErr, err.Error())
	}

	t.Run("Test Error Case, error in FindByIdentifier method", func(tt *testing.T) {
		req := &repository3.GenericNote{
			Identifier:     appId1,
			IdentifierType: repository3.AppType,
			Description:    "test-description",
		}
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}
		testErr := "FindByIdentifier error"
		mockedNoteRepo.On("FindByIdentifier", mock.AnythingOfType("int"), mock.AnythingOfType("repository.NoteType")).Return(nil, errors.New(testErr))

		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)

	})

	t.Run("Test Error Case, entry already exists for the given identifier and identifier type ", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}
		clusterNoteNotFoundError := "cluster note already exists"

		mockedNoteRepo.On("FindByIdentifier", mock.AnythingOfType("int"), mock.AnythingOfType("repository.NoteType")).Return(&repository3.GenericNote{Id: 1}, nil)
		req := &repository3.GenericNote{}
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, clusterNoteNotFoundError, err)
	})

	t.Run("Test Error Case, error in saving genericNote entry", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}
		testErr := "GenericNote Save error"
		mockedNoteRepo.On("FindByIdentifier", mock.AnythingOfType("int"), mock.AnythingOfType("repository.NoteType")).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, mock.AnythingOfType("*repository.GenericNote")).Return(errors.New(testErr))
		req := &repository3.GenericNote{}
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)
	})

	t.Run("Test Error Case, error in saving genericNote history audit", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}
		req := &repository3.GenericNote{}
		testErr := "saving genericNote history audit error"
		mockedNoteRepo.On("FindByIdentifier", mock.AnythingOfType("int"), mock.AnythingOfType("repository.NoteType")).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, mock.AnythingOfType("*repository.GenericNote")).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), mock.AnythingOfType("int32")).Return(nil, errors.New(testErr))
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)
	})

	t.Run("Test Error Case, error in getting user by Id", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, mockedUserRepo := initGenericNoteService(t)
		tx := &pg.Tx{}
		mockedNoteRepo.On("FindByIdentifier", mock.AnythingOfType("int"), mock.AnythingOfType("repository.NoteType")).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, mock.AnythingOfType("*repository.GenericNote")).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), mock.AnythingOfType("int32")).Return(nil, nil)
		req := &repository3.GenericNote{
			Description: "test description",
			AuditLog: sql.AuditLog{
				UpdatedBy: userId1,
			},
		}
		testErr := "GetUserById error"
		mockedUserRepo.On("GetById", req.UpdatedBy).Return(nil, errors.New(testErr))
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)
	})

	t.Run("Test Success Case", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, mockedUserRepo := initGenericNoteService(t)
		tx := &pg.Tx{}
		mockedNoteRepo.On("FindByIdentifier", mock.AnythingOfType("int"), mock.AnythingOfType("repository.NoteType")).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, mock.AnythingOfType("*repository.GenericNote")).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), mock.AnythingOfType("int32")).Return(nil, nil)
		req := &repository3.GenericNote{

			Description: "test description",
			AuditLog: sql.AuditLog{
				UpdatedBy: userId1,
			},
		}

		testUser := testUsers[0]
		mockedUserRepo.On("GetById", req.UpdatedBy).Return(&testUser, nil)
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		assert.NotNil(tt, resp)
		assert.Nil(tt, err)
		assert.Equal(tt, testUser.EmailId, resp.UpdatedBy)
		assert.Equal(tt, req.Description, resp.Description)
	})
}

func TestUpdate(t *testing.T) {

}

func TestGetGenericNotesForAppIds(t *testing.T) {

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

	t.Run("Test Success, Get Newly edited/created Description and old descriptions", func(tt *testing.T) {
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
