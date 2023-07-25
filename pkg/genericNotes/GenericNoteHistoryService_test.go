package genericNotes

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestSave(t *testing.T) {

	testInput := &GenericNoteHistoryBean{
		NoteId:      1,
		Description: "test description",
		CreatedBy:   1,
		CreatedOn:   time.Now(),
	}
	testUserId := int32(1)

	t.Run("Test save successfully", func(tt *testing.T) {
		noteHistoryService, mockNoteHistoryRepo := InitNoteHistoryService(t)
		tx := &pg.Tx{}
		mockNoteHistoryRepo.On("SaveHistory", tx, mock.AnythingOfType("*repository.GenericNoteHistory")).Return(nil)
		resp, err := noteHistoryService.Save(tx, testInput, testUserId)
		assert.Nil(tt, err)
		assert.NotNil(tt, resp)
		if resp == nil {
			assert.Fail(tt, "nil response from save, expecting non nil response ")
		}
		assert.Equal(tt, testInput.NoteId, resp.NoteId)
		assert.Equal(tt, testInput.Description, resp.Description)
	})

	t.Run("Test save Fail", func(tt *testing.T) {
		noteHistoryService, mockNoteHistoryRepo := InitNoteHistoryService(t)
		tx := &pg.Tx{}
		expectedError := errors.New("some error occurred")
		mockNoteHistoryRepo.On("SaveHistory", tx, mock.AnythingOfType("*repository.GenericNoteHistory")).Return(expectedError)

		resp, err := noteHistoryService.Save(tx, testInput, testUserId)
		assert.Nil(tt, resp)
		assert.NotNil(tt, err)
		if err != nil {
			assert.Equal(tt, expectedError, err)
		}
	})

}

func InitNoteHistoryService(t *testing.T) (*GenericNoteHistoryServiceImpl, *mocks.GenericNoteHistoryRepository) {
	logger, err := util.NewSugardLogger()
	if err != nil {
		assert.Fail(t, "error in creating logger", "err", err)
	}
	noteHistoryRepo := mocks.NewGenericNoteHistoryRepository(t)

	return NewGenericNoteHistoryServiceImpl(noteHistoryRepo, logger), noteHistoryRepo
}
