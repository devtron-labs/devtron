package globalTagTests

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/enterprise/globalTag"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"

	//"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/enterprise/globalTag/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGlobalTagService(t *testing.T) {

	t.Run("FetchAllActiveEmptyList", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		globalTagRepositoryMocked := mocks.NewGlobalTagRepository(t)
		var globalTagsFromDb []*globalTag.GlobalTag
		globalTagRepositoryMocked.On("FindAllActive").Return(globalTagsFromDb, nil)
		globalTagServiceImpl := globalTag.NewGlobalTagServiceImpl(sugaredLogger, globalTagRepositoryMocked)
		globalTags, err := globalTagServiceImpl.GetAllActiveTags()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(globalTags))
	})

	t.Run("FetchAllActiveWithError", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		globalTagRepositoryMocked := mocks.NewGlobalTagRepository(t)
		globalTagRepositoryMocked.On("FindAllActive").Return(nil, errors.New("some error occurred"))
		globalTagServiceImpl := globalTag.NewGlobalTagServiceImpl(sugaredLogger, globalTagRepositoryMocked)
		globalTags, err := globalTagServiceImpl.GetAllActiveTags()
		assert.NotNil(t, err)
		assert.Equal(t, 0, len(globalTags))
	})

	t.Run("FetchAllActiveWithData", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		globalTagRepositoryMocked := mocks.NewGlobalTagRepository(t)
		var globalTagsFromDb []*globalTag.GlobalTag
		globalTagsFromDb = append(globalTagsFromDb, &globalTag.GlobalTag{
			Id:                     1,
			Key:                    "key1",
			MandatoryProjectIdsCsv: "1",
			Description:            "someDescription1",
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: 1},
		})
		globalTagsFromDb = append(globalTagsFromDb, &globalTag.GlobalTag{
			Id:                     2,
			Key:                    "key2",
			MandatoryProjectIdsCsv: "2",
			Description:            "someDescription2",
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: 2, UpdatedOn: time.Now()},
		})
		globalTagRepositoryMocked.On("FindAllActive").Return(globalTagsFromDb, nil)
		globalTagServiceImpl := globalTag.NewGlobalTagServiceImpl(sugaredLogger, globalTagRepositoryMocked)
		globalTags, err := globalTagServiceImpl.GetAllActiveTags()
		assert.Nil(t, err)
		assert.Equal(t, 2, len(globalTags))
		assert.Equal(t, 1, globalTags[0].Id)
		assert.Equal(t, 2, globalTags[1].Id)
	})

	t.Run("GetAllActiveTagsForProjectWithEmptyList", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		globalTagRepositoryMocked := mocks.NewGlobalTagRepository(t)
		var globalTagsFromDb []*globalTag.GlobalTag
		globalTagRepositoryMocked.On("FindAllActive").Return(globalTagsFromDb, nil)
		globalTagServiceImpl := globalTag.NewGlobalTagServiceImpl(sugaredLogger, globalTagRepositoryMocked)
		globalTagsForProject, err := globalTagServiceImpl.GetAllActiveTagsForProject(1)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(globalTagsForProject))
	})

	t.Run("GetAllActiveTagsForProjectWithError", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		globalTagRepositoryMocked := mocks.NewGlobalTagRepository(t)
		globalTagRepositoryMocked.On("FindAllActive").Return(nil, errors.New("some error occurred"))
		globalTagServiceImpl := globalTag.NewGlobalTagServiceImpl(sugaredLogger, globalTagRepositoryMocked)
		globalTagsForProject, err := globalTagServiceImpl.GetAllActiveTagsForProject(1)
		assert.NotNil(t, err)
		assert.Equal(t, 0, len(globalTagsForProject))
	})

	t.Run("GetAllActiveTagsForProjectWithEmptyList", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		globalTagRepositoryMocked := mocks.NewGlobalTagRepository(t)
		var globalTagsFromDb []*globalTag.GlobalTag
		globalTagsFromDb = append(globalTagsFromDb, &globalTag.GlobalTag{
			Id:                     1,
			Key:                    "key1",
			MandatoryProjectIdsCsv: "1",
			Description:            "someDescription1",
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: 1},
		})
		globalTagsFromDb = append(globalTagsFromDb, &globalTag.GlobalTag{
			Id:                     2,
			Key:                    "key2",
			MandatoryProjectIdsCsv: "-1",
			Description:            "someDescription2",
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: 2},
		})
		globalTagsFromDb = append(globalTagsFromDb, &globalTag.GlobalTag{
			Id:                     3,
			Key:                    "key3",
			MandatoryProjectIdsCsv: "",
			Description:            "someDescription3",
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: 3},
		})
		globalTagsFromDb = append(globalTagsFromDb, &globalTag.GlobalTag{
			Id:                     4,
			Key:                    "key4",
			MandatoryProjectIdsCsv: "2,3",
			Description:            "someDescription4",
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: 4},
		})
		globalTagsFromDb = append(globalTagsFromDb, &globalTag.GlobalTag{
			Id:                     5,
			Key:                    "key5",
			MandatoryProjectIdsCsv: "1,2",
			Description:            "someDescription5",
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: 5},
		})
		globalTagRepositoryMocked.On("FindAllActive").Return(globalTagsFromDb, nil)
		globalTagServiceImpl := globalTag.NewGlobalTagServiceImpl(sugaredLogger, globalTagRepositoryMocked)
		globalTagsForProject, err := globalTagServiceImpl.GetAllActiveTagsForProject(1)
		assert.Nil(t, err)
		assert.Equal(t, len(globalTagsFromDb), len(globalTagsForProject))
		var expectedMandatoryKeys []string
		expectedMandatoryKeys = append(expectedMandatoryKeys, "key1", "key2", "key5")
		var expectedOptionalKeys []string
		expectedOptionalKeys = append(expectedOptionalKeys, "key3", "key4")

		var actualMandatoryKeys []string
		var actualOptionalKeys []string
		for _, globalTagForProject := range globalTagsForProject {
			if globalTagForProject.IsMandatory {
				actualMandatoryKeys = append(actualMandatoryKeys, globalTagForProject.Key)
			} else {
				actualOptionalKeys = append(actualOptionalKeys, globalTagForProject.Key)
			}
		}
		assert.Equal(t, expectedMandatoryKeys, actualMandatoryKeys)
		assert.Equal(t, expectedOptionalKeys, actualOptionalKeys)
	})

}
