package cluster

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSaveEphemeralContainer_Success(t *testing.T) {
	repository := mocks.NewEphemeralContainersRepository(t)

	// Set up the expected repository method calls and return values
	repository.On("FindContainerByName", 1, "namespace-1", "pod-1", "container-1").Return(nil, nil)
	repository.On("StartTx").Return(&pg.Tx{}, nil)
	repository.On("RollbackTx", mock.AnythingOfType("*pg.Tx")).Return(nil) // Add this expectation
	repository.On("SaveData", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerBean")).Return(nil)
	repository.On("SaveAction", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerAction")).Return(nil)
	repository.On("CommitTx", mock.AnythingOfType("*pg.Tx")).Return(nil)
	logger, _ := util.NewSugardLogger()
	service := NewEphemeralContainerServiceImpl(repository, logger)

	// Create a sample EphemeralContainerRequest
	request := EphemeralContainerRequest{
		BasicData: &EphemeralContainerBasicData{
			ContainerName:       "container-1",
			TargetContainerName: "target-container-1",
			Image:               "image-1",
		},
		AdvancedData: &EphemeralContainerAdvancedData{
			Manifest: "manifest-1",
		},
		Namespace: "namespace-1",
		ClusterId: 1,
		PodName:   "pod-1",
		UserId:    123,
	}

	err := service.SaveEphemeralContainer(request)

	assert.NoError(t, err)

	repository.AssertCalled(t, "FindContainerByName", 1, "namespace-1", "pod-1", "container-1")
	repository.AssertCalled(t, "StartTx")
	repository.AssertCalled(t, "RollbackTx", mock.AnythingOfType("*pg.Tx")) // Add this assertion
	repository.AssertCalled(t, "SaveData", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerBean"))
	repository.AssertCalled(t, "SaveAction", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerAction"))
	repository.AssertCalled(t, "CommitTx", mock.AnythingOfType("*pg.Tx"))
}

func TestSaveEphemeralContainer_FindContainerError(t *testing.T) {

	repository := mocks.NewEphemeralContainersRepository(t)

	repository.On("FindContainerByName", mock.AnythingOfType("int"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, errors.New("error finding container"))

	logger, _ := util.NewSugardLogger()
	service := EphemeralContainerServiceImpl{
		repository: repository,
		logger:     logger,
	}

	request := EphemeralContainerRequest{
		BasicData: &EphemeralContainerBasicData{
			ContainerName:       "container-1",
			TargetContainerName: "target-container-1",
			Image:               "image-1",
		},
		AdvancedData: &EphemeralContainerAdvancedData{
			Manifest: "manifest-1",
		},
		Namespace: "namespace-1",
		ClusterId: 1,
		PodName:   "pod-1",
		UserId:    123,
	}

	err := service.SaveEphemeralContainer(request)

	assert.Error(t, err)

	repository.AssertCalled(t, "FindContainerByName", 1, "namespace-1", "pod-1", "container-1")
}

type EphemeralContainerBean struct {
}

func TestSaveEphemeralContainer_ContainerAlreadyPresent(t *testing.T) {
	repository := mocks.NewEphemeralContainersRepository(t)

	repository.On("FindContainerByName", mock.AnythingOfType("int"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&EphemeralContainerBean{}, nil)

	logger, _ := util.NewSugardLogger()
	service := EphemeralContainerServiceImpl{
		repository: repository,
		logger:     logger,
	}

	request := EphemeralContainerRequest{
		BasicData: &EphemeralContainerBasicData{
			ContainerName:       "container-1",
			TargetContainerName: "target-container-1",
			Image:               "image-1",
		},
		AdvancedData: &EphemeralContainerAdvancedData{
			Manifest: "manifest-1",
		},
		Namespace: "namespace-1",
		ClusterId: 1,
		PodName:   "pod-1",
		UserId:    123,
	}

	err := service.SaveEphemeralContainer(request)

	assert.Error(t, err)

	repository.AssertCalled(t, "FindContainerByName", 1, "namespace-1", "pod-1", "container-1")
}
