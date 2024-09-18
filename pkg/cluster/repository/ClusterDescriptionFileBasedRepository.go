package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ClusterDescriptionFileBasedRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewClusterDescriptionFileBasedRepository(connection *sql.SqliteConnection, logger *zap.SugaredLogger) *ClusterDescriptionFileBasedRepositoryImpl {

	clusterDescription := &ClusterDescription{}
	connection.Migrator.MigrateEntities(clusterDescription)
	logger.Debugw("cluster description repository file based initialized")
	return &ClusterDescriptionFileBasedRepositoryImpl{logger, connection.DbConnection}
}

func (impl ClusterDescriptionFileBasedRepositoryImpl) FindByClusterIdWithClusterDetails(clusterId int) (*ClusterDescription, error) {
	clusterDescription := &ClusterDescription{}
	query := "SELECT cl.id AS cluster_id, cl.cluster_name AS cluster_name, cl.description AS cluster_description,  cl.server_url, cl.created_on AS cluster_created_on, cl.created_by AS cluster_created_by, gn.id AS note_id, gn.description AS note, gn.created_by, gn.created_on, gn.updated_by, gn.updated_on FROM" +
		" cluster_entities cl LEFT JOIN" +
		" generic_notes gn " +
		" ON cl.id=gn.identifier AND (gn.identifier_type = %d OR gn.identifier_type IS NULL)" +
		" WHERE cl.id=%d AND cl.active=true " +
		" LIMIT 1 OFFSET 0;"
	query = fmt.Sprintf(query, 0, clusterId) //0 is for cluster type description
	result := impl.dbConnection.Raw(query).Find(clusterDescription)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	}
	return clusterDescription, err
}

