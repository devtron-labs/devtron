package cluster

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockEphemeralContainersRepository struct {
	mock.Mock
}

func (m MockEphemeralContainersRepository) StartTx() (*pg.Tx, error) {
	return &pg.Tx{}, nil
}

func (m MockEphemeralContainersRepository) RollbackTx(tx *pg.Tx) error {
	return tx.Rollback()
}

func (m MockEphemeralContainersRepository) CommitTx(tx *pg.Tx) error {
	return tx.Commit()
}

func (m MockEphemeralContainersRepository) SaveData(tx *pg.Tx, model *repository.EphemeralContainerBean) error {
	return tx.Insert(model)
}

func (m MockEphemeralContainersRepository) SaveAction(tx *pg.Tx, model *repository.EphemeralContainerAction) error {
	return tx.Insert(model)
}

func (m MockEphemeralContainersRepository) FindContainerByName(clusterID int, namespace, podName, name string) (*repository.EphemeralContainerBean, error) {
	container := repository.EphemeralContainerBean{
		ClusterId: clusterID,
		Namespace: namespace,
		PodName:   podName,
		Name:      name,
	}

	return &container, nil
}

func TestSaveEphemeralContainer_Success(t *testing.T) {

	repository := &MockEphemeralContainersRepository{}

	// Set up the expected repository method calls and return values
	repository.On("FindContainerByName", mock.AnythingOfType("int"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, nil)
	repository.On("StartTx").Return(&pg.Tx{}, nil)
	repository.On("SaveData", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerBean")).Return(nil)
	repository.On("SaveAction", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerAction")).Return(nil)
	repository.On("CommitTx", mock.AnythingOfType("*pg.Tx")).Return(nil)

	service := EphemeralContainerServiceImpl{
		repository: repository,
		logger:     nil,
	}

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
	repository.AssertCalled(t, "SaveData", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerBean"))
	repository.AssertCalled(t, "SaveAction", mock.AnythingOfType("*pg.Tx"), mock.AnythingOfType("*repository.EphemeralContainerAction"))
	repository.AssertCalled(t, "CommitTx", mock.AnythingOfType("*pg.Tx"))
}

func TestSaveEphemeralContainer_FindContainerError(t *testing.T) {

	repository := &MockEphemeralContainersRepository{}

	repository.On("FindContainerByName", mock.AnythingOfType("int"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, errors.New("error finding container"))

	service := EphemeralContainerServiceImpl{
		repository: repository,
		logger:     nil,
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
	repository := &MockEphemeralContainersRepository{}

	repository.On("FindContainerByName", mock.AnythingOfType("int"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&EphemeralContainerBean{}, nil)

	service := EphemeralContainerServiceImpl{
		repository: repository,
		logger:     nil,
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
