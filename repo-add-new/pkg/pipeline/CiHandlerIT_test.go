package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCiHandlerImpl_FetchArtifactsForCiJob(t *testing.T) {
	t.SkipNow()
	ciHandler := initCiHandler()

	t.Run("Fetch Ci Artifacts For Ci Job type", func(tt *testing.T) {
		buildId := 304 // Mocked because child workflows are only created dynamic based on number of images which are available after polling
		time.Sleep(5 * time.Second)
		_, err := ciHandler.FetchArtifactsForCiJob(buildId)
		assert.Nil(t, err)

	})
}

func initCiHandler() *CiHandlerImpl {
	sugaredLogger, _ := util.InitLogger()
	config, _ := sql.GetConfig()
	db, _ := sql.NewDbConnection(config, sugaredLogger)
	ciArtifactRepositoryImpl := repository.NewCiArtifactRepositoryImpl(db, sugaredLogger)
	ciHandlerImpl := NewCiHandlerImpl(sugaredLogger, nil, nil, nil, nil, nil, nil, ciArtifactRepositoryImpl, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return ciHandlerImpl
}
