package tests

import (
	"errors"
	"testing"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	mocks2 "github.com/devtron-labs/devtron/pkg/auth/user/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	mocks3 "github.com/devtron-labs/devtron/pkg/genericNotes/mocks"
	repository3 "github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testAppId1, testAppId2 = 1, 2
var userId1, userId2 int32 = 1, 2
var testAppIds = []int{testAppId1, testAppId2}
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

	req := &repository3.GenericNote{
		Identifier:     testAppId1,
		IdentifierType: repository3.AppType,
		Description:    "test-description",
		AuditLog: sql.AuditLog{
			UpdatedBy: userId1,
		},
	}

	t.Run("Test Error Case, error in FindByIdentifier method", func(tt *testing.T) {

		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}
		testErr := "FindByIdentifier error"
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(nil, errors.New(testErr))

		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)

	})

	t.Run("Test Error Case, entry already exists for the given identifier and identifier type ", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}
		clusterNoteNotFoundError := "cluster note already exists"
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(&repository3.GenericNote{Id: 1}, nil)
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, clusterNoteNotFoundError, err)
	})

	t.Run("Test Error Case, error in saving genericNote entry", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}
		testErr := "GenericNote Save error"
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, req).Return(errors.New(testErr))

		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)
	})

	t.Run("Test Error Case, error in saving genericNote history audit", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, _ := initGenericNoteService(tt)
		tx := &pg.Tx{}

		testErr := "saving genericNote history audit error"
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, req).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), mock.AnythingOfType("int32")).Return(nil, errors.New(testErr))
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)
	})

	t.Run("Test Error Case, error in getting user by Id", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, mockedUserRepo := initGenericNoteService(t)
		tx := &pg.Tx{}

		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, req).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), mock.AnythingOfType("int32")).Return(nil, nil)

		testErr := "GetUserById error"
		mockedUserRepo.On("GetById", req.UpdatedBy).Return(nil, errors.New(testErr))
		resp, err := genericNoteSvc.Save(tx, req, userId1)
		testErrorAssertions(tt, resp, testErr, err)
	})

	t.Run("Test Success Case for Save", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, mockedUserRepo := initGenericNoteService(t)
		tx := &pg.Tx{}

		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(&repository3.GenericNote{}, nil)
		mockedNoteRepo.On("Save", tx, req).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), mock.AnythingOfType("int32")).Return(nil, nil)

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
	req := &repository3.GenericNote{
		Identifier:     testAppId1,
		IdentifierType: repository3.AppType,
		Description:    "test-description",
		AuditLog: sql.AuditLog{
			UpdatedBy: userId1,
		},
	}

	t.Run("Test Error Case, error in starting transaction", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(t)
		tx := &pg.Tx{}

		testErr := errors.New("start tnx error")
		mockedNoteRepo.On("StartTx").Return(tx, testErr)
		mockedNoteRepo.On("RollbackTx", tx).Return(errors.New("rollback error"))
		res, err := genericNoteSvc.Update(req, userId1)
		testErrorAssertions(tt, res, testErr.Error(), err)
	})

	t.Run("Test Error Case, error in starting transaction", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(t)
		tx := &pg.Tx{}

		testErr := errors.New("find note by identifier error")
		mockedNoteRepo.On("StartTx").Return(tx, nil)
		mockedNoteRepo.On("RollbackTx", tx).Return(nil)
		dbModel := &repository3.GenericNote{}
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(dbModel, testErr)
		res, err := genericNoteSvc.Update(req, userId1)
		testErrorAssertions(tt, res, testErr.Error(), err)
	})

	t.Run("Test Error Case, error updating generic note data", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, _ := initGenericNoteService(t)
		tx := &pg.Tx{}
		testErr := errors.New("note update error")
		mockedNoteRepo.On("StartTx").Return(tx, nil)
		mockedNoteRepo.On("RollbackTx", tx).Return(nil)
		dbModel := &repository3.GenericNote{}
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(dbModel, nil)
		mockedNoteRepo.On("Update", tx, dbModel).Return(testErr)
		res, err := genericNoteSvc.Update(req, userId1)
		testErrorAssertions(tt, res, testErr.Error(), err)
	})

	t.Run("Test Error Case, error in saving generic_note update audit data", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, _ := initGenericNoteService(t)
		tx := &pg.Tx{}
		testErr := errors.New("save generic_note update audit error")
		mockedNoteRepo.On("StartTx").Return(tx, nil)
		mockedNoteRepo.On("RollbackTx", tx).Return(nil)
		dbModel := &repository3.GenericNote{}
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(dbModel, nil)
		mockedNoteRepo.On("Update", tx, dbModel).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), userId1).Return(nil, testErr)

		res, err := genericNoteSvc.Update(req, userId1)
		testErrorAssertions(tt, res, testErr.Error(), err)
	})

	t.Run("Test Error Case, error in finding user by userId", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, mockedUserRepo := initGenericNoteService(t)
		tx := &pg.Tx{}
		testErr := errors.New("find user by userId error")
		mockedNoteRepo.On("StartTx").Return(tx, nil)
		mockedNoteRepo.On("RollbackTx", tx).Return(nil)
		dbModel := &repository3.GenericNote{}
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(dbModel, nil)
		mockedNoteRepo.On("Update", tx, dbModel).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), userId1).Return(nil, nil)

		mockedUserRepo.On("GetById", req.UpdatedBy).Return(nil, testErr)
		res, err := genericNoteSvc.Update(req, userId1)
		testErrorAssertions(tt, res, testErr.Error(), err)
	})

	t.Run("Test Error Case, error in committing db transaction", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, mockedUserRepo := initGenericNoteService(t)
		tx := &pg.Tx{}
		testErr := errors.New("transaction commit error")
		mockedNoteRepo.On("StartTx").Return(tx, nil)
		mockedNoteRepo.On("RollbackTx", tx).Return(nil)
		dbModel := &repository3.GenericNote{}
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(dbModel, nil)
		mockedNoteRepo.On("Update", tx, dbModel).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), userId1).Return(nil, nil)

		testUser := testUsers[0]
		mockedUserRepo.On("GetById", req.UpdatedBy).Return(&testUser, nil)
		mockedNoteRepo.On("CommitTx", tx).Return(testErr)
		res, err := genericNoteSvc.Update(req, userId1)
		testErrorAssertions(tt, res, testErr.Error(), err)
	})

	t.Run("Test Success Case for Update", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, mockedHistorySvc, mockedUserRepo := initGenericNoteService(t)
		tx := &pg.Tx{}

		mockedNoteRepo.On("StartTx").Return(tx, nil)
		mockedNoteRepo.On("RollbackTx", tx).Return(nil)
		dbModel := &repository3.GenericNote{}
		mockedNoteRepo.On("FindByIdentifier", req.Identifier, req.IdentifierType).Return(dbModel, nil)
		mockedNoteRepo.On("Update", tx, dbModel).Return(nil)
		mockedHistorySvc.On("Save", tx, mock.AnythingOfType("*genericNotes.GenericNoteHistoryBean"), userId1).Return(nil, nil)

		testUser := testUsers[0]
		mockedUserRepo.On("GetById", req.UpdatedBy).Return(&testUser, nil)
		mockedNoteRepo.On("CommitTx", tx).Return(nil)
		resp, err := genericNoteSvc.Update(req, userId1)
		assert.NotNil(tt, resp)
		assert.Nil(tt, err)
		assert.Equal(tt, testUser.EmailId, resp.UpdatedBy)
		assert.Equal(tt, req.Description, resp.Description)
	})

}
func TestGetGenericNotesForAppIds(t *testing.T) {

	genericNoteResp := &repository3.GenericNote{
		Id:             1,
		Identifier:     testAppId1,
		IdentifierType: repository3.AppType,
		Description:    "test-response-1",
		AuditLog: sql.AuditLog{
			UpdatedBy: testUserIds[0],
		},
	}

	descriptionResp := &repository3.GenericNote{
		Id:             0,
		Identifier:     testAppId2,
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
		noteRespBean := resp[testAppId1]
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

		noteRespBean := resp[testAppId1]
		assert.NotNil(tt, noteRespBean)
		assert.Equal(tt, genericNoteResp.Id, noteRespBean.Id)
		assert.Equal(tt, genericNoteResp.Description, noteRespBean.Description)

	})

	t.Run("Test Success, Get Newly edited/created GenericNote and old descriptions", func(tt *testing.T) {
		genericNoteSvc, mockedNoteRepo, _, mockedUserRepo := initGenericNoteService(t)

		mockedNoteRepo.On("GetGenericNotesForAppIds", mock.AnythingOfType("[]int")).Return(getGenericNotesForAppIdsResp, nil)
		mockedNoteRepo.On("GetDescriptionFromAppIds", mock.AnythingOfType("[]int")).Return(getDescriptionForAppIdsResp, nil)
		mockedUserRepo.On("GetByIds", mock.AnythingOfType("[]int32")).Return(testUsers, nil)
		resp, err := genericNoteSvc.GetGenericNotesForAppIds(testAppIds)

		assert.NotNil(tt, resp)
		assert.Equal(tt, 2, len(resp))
		assert.Nil(tt, err)

		noteRespBean := resp[testAppId1]
		assert.NotNil(tt, noteRespBean)
		assert.Equal(tt, genericNoteResp.Id, noteRespBean.Id)
		assert.Equal(tt, genericNoteResp.Description, noteRespBean.Description)
		assert.Equal(tt, testUsers[0].EmailId, noteRespBean.UpdatedBy)

		descriptionRespBean := resp[testAppId2]
		assert.NotNil(tt, descriptionRespBean)
		assert.Equal(tt, descriptionResp.Id, descriptionRespBean.Id)
		assert.Equal(tt, descriptionResp.Description, descriptionRespBean.Description)
		assert.Equal(tt, testUsers[1].EmailId, descriptionRespBean.UpdatedBy)
	})
}

func testErrorAssertions(tt *testing.T, resp *bean2.GenericNoteResponseBean, testErr string, err error) {
	assert.NotNil(tt, err)
	assert.Nil(tt, resp)
	assert.Equal(tt, testErr, err.Error())
}

func initGenericNoteService(t *testing.T) (*genericNotes.GenericNoteServiceImpl, *mocks.GenericNoteRepository, *mocks3.GenericNoteHistoryService, *mocks2.UserRepository) {
	logger, err := util.NewSugardLogger()
	if err != nil {
		assert.Fail(t, "error in creating logger")
	}

	mockedNoteRepo := mocks.NewGenericNoteRepository(t)
	mockedHistoryService := mocks3.NewGenericNoteHistoryService(t)
	mockedUserRepo := mocks2.NewUserRepository(t)
	genericNoteSvc := genericNotes.NewGenericNoteServiceImpl(mockedNoteRepo, mockedHistoryService, mockedUserRepo, logger)
	return genericNoteSvc, mockedNoteRepo, mockedHistoryService, mockedUserRepo
}
