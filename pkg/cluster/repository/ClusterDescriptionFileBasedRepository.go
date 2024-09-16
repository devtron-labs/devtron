package repository

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ClusterDescriptionFileBasedRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewClusterDescriptionFileBasedRepository(logger *zap.SugaredLogger) *ClusterDescriptionFileBasedRepositoryImpl {
	err, dbPath := createOrCheckClusterDbPath(logger)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	//db, err := sql.Open("sqlite3", "./cluster.db")
	if err != nil {
		logger.Fatal("error occurred while opening db connection", "error", err)
	}
	migrator := db.Migrator()
	clusterDescription := &ClusterDescription{}
	hasTable := migrator.HasTable(clusterDescription)
	if !hasTable {
		err = migrator.CreateTable(clusterDescription)
		if err != nil {
			logger.Fatal("error occurred while creating cluster description table", "error", err)
		}
	}
	logger.Debugw("cluster description repository file based initialized")
	return &ClusterDescriptionFileBasedRepositoryImpl{logger, db}
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
	//_, err := impl.dbConnection.Query(clusterDescription, query)
	result := impl.dbConnection.Raw(query).Find(clusterDescription)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	}
	return clusterDescription, err
}

