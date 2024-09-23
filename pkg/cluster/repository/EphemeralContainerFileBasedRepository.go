package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type EphemeralContainerEntity struct {
	Id                  int
	Name                string
	ClusterId           int
	Namespace           string
	PodName             string
	TargetContainer     string
	Config              string
	IsExternallyCreated bool
}

type EphemeralContainerActionEntity struct {
	Id                   int
	EphemeralContainerId int
	ActionType           ContainerAction
	PerformedBy          int32
	PerformedAt          time.Time
}

type EphemeralContainerFileBasedRepositoryImpl struct {
	*sql.NoopTransactionUtilImpl
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewEphemeralContainerFileBasedRepository(connection *sql.SqliteConnection, logger *zap.SugaredLogger, transactionWrapper *sql.NoopTransactionUtilImpl) *EphemeralContainerFileBasedRepositoryImpl {
	ephemeralContainerEntity := &EphemeralContainerBean{}
	ephemeralContainerActionEntity := &EphemeralContainerAction{}
	connection.Migrator.MigrateEntities(ephemeralContainerEntity, ephemeralContainerActionEntity)
	logger.Debugw("ephemeralContainer repository file based initialized")
	return &EphemeralContainerFileBasedRepositoryImpl{transactionWrapper, logger, connection.DbConnection}
}

func (impl EphemeralContainerFileBasedRepositoryImpl) SaveEphemeralContainerData(tx *pg.Tx, model *EphemeralContainerBean) error {
	result := impl.dbConnection.Create(model)
	return result.Error
}

func (impl EphemeralContainerFileBasedRepositoryImpl) SaveEphemeralContainerActionAudit(tx *pg.Tx, model *EphemeralContainerAction) error {
	result := impl.dbConnection.Create(model)
	return result.Error
}

func (impl EphemeralContainerFileBasedRepositoryImpl) FindContainerByName(clusterID int, namespace, podName, name string) (*EphemeralContainerBean, error) {
	ephemeralContainerEntity := &EphemeralContainerBean{}
	result := impl.dbConnection.
		Where("cluster_id = ?", clusterID).
		Where("namespace = ?", namespace).
		Where("pod_name = ?", podName).
		Where("name = ?", name).
		Find(ephemeralContainerEntity)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	} else if err != nil {
		impl.logger.Errorw("error occurred while finding ephemeral container data ", "cluster_id", clusterID, "err", err)
		return nil, errors.New("failed to fetch ephemeral container")
	}
	if errors.Is(err, pg.ErrNoRows) || ephemeralContainerEntity.Id == 0 {
		ephemeralContainerEntity = nil
	}
	return ephemeralContainerEntity, nil
}
