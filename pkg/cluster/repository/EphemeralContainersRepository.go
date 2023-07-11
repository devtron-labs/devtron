package repository

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type EphemeralContainerBean struct {
	tableName           struct{} `sql:"ephemeral_container" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	Name                string   `sql:"name"`
	ClusterId           int      `sql:"cluster_id"`
	Namespace           string   `sql:"namespace"`
	PodName             string   `sql:"pod_name"`
	TargetContainer     string   `sql:"target_container"`
	Config              string   `sql:"config"`
	IsExternallyCreated bool     `sql:"is_externally_created"`
}

type EphemeralContainerAction struct {
	tableName            struct{}  `sql:"ephemeral_container_actions" pg:",discard_unknown_columns"`
	Id                   int       `sql:"id,pk"`
	EphemeralContainerId int       `sql:"ephemeral_container_id"`
	ActionType           int       `sql:"action_type"`
	PerformedBy          int       `sql:"performed_by"`
	PerformedAt          time.Time `sql:"performed_at"`
}

type EphemeralContainersRepository interface {
	SaveData(tx *pg.Tx, model *EphemeralContainerBean) error
	SaveAction(tx *pg.Tx, model *EphemeralContainerAction) error
	FindContainerByName(clusterID int, namespace, podName, name string) (*EphemeralContainerBean, error)
}

func NewEphemeralContainersRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger, tx *pg.Tx) *EphemeralContainersImpl {
	return &EphemeralContainersImpl{
		dbConnection: dbConnection,
		logger:       logger,
		tx:           tx,
	}
}

type EphemeralContainersImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
	tx           *pg.Tx
}

func (impl EphemeralContainersImpl) SaveData(tx *pg.Tx, model *EphemeralContainerBean) error {
	return tx.Insert(model)
}

func (impl EphemeralContainersImpl) SaveAction(tx *pg.Tx, model *EphemeralContainerAction) error {
	return tx.Insert(model)
}

func (impl EphemeralContainersImpl) FindContainerByName(clusterID int, namespace, podName, name string) (*EphemeralContainerBean, error) {
	container := &EphemeralContainerBean{}
	err := impl.dbConnection.Model(container).
		Where("cluster_id = ?", clusterID).
		Where("namespace = ?", namespace).
		Where("pod_name = ?", podName).
		Where("name = ?", name).
		Select()
	if err != nil {
		return nil, err
	}

	return container, nil
}
