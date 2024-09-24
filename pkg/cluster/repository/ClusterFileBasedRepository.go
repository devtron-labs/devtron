package repository

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"

	"github.com/go-pg/pg"
	//_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ClusterFileBasedRepository struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

type ClusterEntity struct {
	ID                int
	ClusterName       string
	ServerUrl         string
	Active            *bool
	Config            string
	K8sVersion        string
	ErrorInConnecting string
	Description string
	PrometheusEndpoint     string
	CdArgoSetup            bool
	PUserName              string
	PPassword              string
	PTlsClientCert         string
	PTlsClientKey          string
	AgentInstallationStage int
	IsVirtualCluster       bool
	InsecureSkipTlsVerify *bool
	sql.AuditLog
}

func NewClusterRepositoryFileBased(connection *sql.SqliteConnection, logger *zap.SugaredLogger) *ClusterFileBasedRepository {
	clusterEntity := &ClusterEntity{}
	connection.Migrator.MigrateEntities(clusterEntity)
	logger.Debugw("cluster repository file based initialized")
	return &ClusterFileBasedRepository{logger, connection.DbConnection}
}

func (impl *ClusterFileBasedRepository) FindAllActiveExceptVirtual() ([]Cluster, error) {
	var clusterEntities []ClusterEntity
	result := impl.dbConnection.
		Where("active=?", true).
		Where("is_virtual_cluster=? OR is_virtual_cluster IS NULL", false).
		Find(&clusterEntities)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	}
	if err != nil {
		impl.logger.Errorw("error occurred while finding all cluster data", "err", err)
		return nil, err
	}
	clusters := impl.ConvertEntitiesToModel(clusterEntities)
	return clusters, nil
}

func (impl *ClusterFileBasedRepository) SetDescription(id int, description string, userId int32) error {
	cluster, err := impl.FindById(id)
	if err != nil {
		return err
	}
	err, clusterEntity := impl.convertToEntity(cluster)
	if err != nil {
		impl.logger.Errorw("error occurred while converting model to entity", "error", err)
		return errors.New("failed to update cluster")
	}
	clusterEntity.Description = description
	clusterEntity.UpdatedBy = userId
	clusterEntity.UpdatedOn = time.Now()
	result := impl.dbConnection.Model(clusterEntity).Updates(clusterEntity)
	err = result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while updating cluster description", "clusterId", id, "err", err)
		return errors.New("failed to update cluster description")
	}
	return err
}

func (impl *ClusterFileBasedRepository) FindActiveClusters() ([]Cluster, error) {
	var clusterEntities []ClusterEntity
	result := impl.dbConnection.
		Where("active=?", true).
		Select("id, cluster_name, active").
		Find(&clusterEntities)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while finding all active cluster data", "err", err)
		return nil, err
	}
	clusters := impl.ConvertEntitiesToModel(clusterEntities)
	return clusters, nil
}

func (impl *ClusterFileBasedRepository) ConvertEntitiesToModel(clusterEntities []ClusterEntity) []Cluster {
	var clusters []Cluster
	for _, clusterEntity := range clusterEntities {
		clusterBean, err := impl.convertToModel(&clusterEntity)
		if err != nil {
			impl.logger.Errorw("error occurred while converting entity to model bean", "err", err)
			continue
		}
		clusters = append(clusters, *clusterBean)
	}
	return clusters
}

func (impl *ClusterFileBasedRepository) SaveAll(models []*Cluster) error {
	var clusterEntities []*ClusterEntity
	for _, cluster := range models {
		_, clusterEntity := impl.convertToEntity(cluster)
		clusterEntities = append(clusterEntities, clusterEntity)
	}
	result := impl.dbConnection.Create(&clusterEntities)
	return result.Error
}

func (impl *ClusterFileBasedRepository) FindByNames(clusterNames []string) ([]*Cluster, error) {
	var clusterEntities []ClusterEntity
	result := impl.dbConnection.
		Where("active=?", true).
		Where("cluster_name in ?", clusterNames).
		Find(&clusterEntities)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while finding all cluster data", "err", err)
		return nil, err
	}
	var clusters []*Cluster
	for _, clusterEntity := range clusterEntities {
		clusterBean, err := impl.convertToModel(&clusterEntity)
		if err != nil {
			impl.logger.Errorw("error occurred while converting entity to model bean", "err", err)
			continue
		}
		clusters = append(clusters, clusterBean)
	}
	return clusters, nil
}

func (impl *ClusterFileBasedRepository) Save(model *Cluster) error {

	err, clusterEntity := impl.convertToEntity(model)
	if err != nil {
		return err
	}
	result := impl.dbConnection.Model(clusterEntity).Create(clusterEntity)
	err = result.Error

	if err != nil {
		impl.logger.Errorw("error occurred while executing insert statement", "err", err)
		return err
	}
	model.Id = clusterEntity.ID
	return nil
}

func (impl *ClusterFileBasedRepository) convertToEntity(model *Cluster) (error, *ClusterEntity) {
	configJson, err := json.Marshal(model.Config)
	if err != nil {
		impl.logger.Errorw("error occurred while converting to entity", "err", err)
		return errors.New("failed to process cluster data"), nil
	}
	clusterEntity := &ClusterEntity{
		ID:                model.Id,
		ClusterName:       model.ClusterName,
		ServerUrl:         model.ServerUrl,
		Config:            string(configJson),
		Active:            &model.Active,
		K8sVersion:        model.K8sVersion,
		ErrorInConnecting: model.ErrorInConnecting,
		PrometheusEndpoint:     model.PrometheusEndpoint,
		AgentInstallationStage: model.AgentInstallationStage,
		InsecureSkipTlsVerify: &model.InsecureSkipTlsVerify,
		IsVirtualCluster:       model.IsVirtualCluster,
		PUserName:              model.PUserName,
		PPassword:              model.PPassword,
		PTlsClientCert:         model.PTlsClientCert,
		PTlsClientKey:          model.PTlsClientKey,
		AuditLog:          sql.AuditLog{UpdatedOn: model.UpdatedOn, CreatedOn: model.CreatedOn, UpdatedBy: model.UpdatedBy, CreatedBy: model.CreatedBy},
	}
	return err, clusterEntity
}

func (impl *ClusterFileBasedRepository) convertToModel(entity *ClusterEntity) (*Cluster, error) {
	clusterConfig := make(map[string]string)
	if len(entity.Config) > 0 {
		err := json.Unmarshal([]byte(entity.Config), &clusterConfig)
		if err != nil {
			impl.logger.Errorw("error occured while unmarshalling cluster config ", "error", err)
			return nil, errors.New("failed to process cluster data")
		}
	}
	isActive := false
	if entity.Active != nil {
		isActive = *entity.Active
	}
	insecureSkipTlsVerify := true
	if entity.InsecureSkipTlsVerify != nil && *entity.InsecureSkipTlsVerify == false {
		insecureSkipTlsVerify = false
	}
	clusterBean := &Cluster{
		Id:                entity.ID,
		ClusterName:       entity.ClusterName,
		ServerUrl:         entity.ServerUrl,
		Config:            clusterConfig,
		K8sVersion:        entity.K8sVersion,
		ErrorInConnecting: entity.ErrorInConnecting,
		AuditLog:          entity.AuditLog,
		Active:            isActive,
		PrometheusEndpoint:     entity.PrometheusEndpoint,
		AgentInstallationStage: entity.AgentInstallationStage,
		InsecureSkipTlsVerify: insecureSkipTlsVerify,
		IsVirtualCluster:       entity.IsVirtualCluster,
		PUserName:              entity.PUserName,
		PPassword:              entity.PPassword,
		PTlsClientCert:         entity.PTlsClientCert,
		PTlsClientKey:          entity.PTlsClientKey,
	}
	return clusterBean, nil
}

func (impl *ClusterFileBasedRepository) FindOne(clusterName string) (*Cluster, error) {
	return impl.FindOneActive(clusterName)
}

func (impl *ClusterFileBasedRepository) FindOneActive(clusterName string) (*Cluster, error) {
	clusterEntity := &ClusterEntity{}
	result := impl.dbConnection.
		Where("cluster_name = ?", clusterName).
		Where("active = ?", true).
		Find(clusterEntity).
		Limit(1)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while finding cluster data ", "clusterName", clusterName, "err", err)
		return nil, errors.New("failed to fetch cluster")
	}
	//queryRow := impl.dbConnection.QueryRow("SELECT * FROM clusterEntity WHERE cluster_name = ? and active = ?")
	//err := queryRow.Scan()
	clusterBean, err := impl.convertToModel(clusterEntity)
	if err != nil {
		impl.logger.Errorw("error occurred while converting cluster data to  model ", "clusterName", clusterName, "err", err)
		return nil, errors.New("failed to fetch cluster")
	}
	return clusterBean, nil
}

func (impl *ClusterFileBasedRepository) FindAll() ([]Cluster, error) {
	return impl.FindAllActive()
}

func (impl *ClusterFileBasedRepository) FindAllActive() ([]Cluster, error) {
	var clusterEntities []ClusterEntity
	result := impl.dbConnection.
		Where("active = ?", true).
		Find(&clusterEntities)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while finding all cluster data", "err", err)
		return nil, err
	}
	clusters := impl.ConvertEntitiesToModel(clusterEntities)
	return clusters, nil
}

func (impl *ClusterFileBasedRepository) FindById(id int) (*Cluster, error) {
	clusterEntity := &ClusterEntity{}
	result := impl.dbConnection.
		Where("id =?", id).
		Where("active = ?", true).
		Find(clusterEntity).
		Limit(1)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	}
	if err != nil {
		impl.logger.Errorw("error occurred while finding cluster data ", "id", id, "err", err)
		return nil, err
	}
	//queryRow := impl.dbConnection.QueryRow("SELECT * FROM clusterEntity WHERE cluster_name = ? and active = ?")
	//err := queryRow.Scan()
	clusterBean, err := impl.convertToModel(clusterEntity)
	if err != nil {
		impl.logger.Errorw("error occurred while converting cluster data to  model ", "id", id, "err", err)
		return nil, errors.New("failed to fetch cluster")
	}
	return clusterBean, nil
}

func (impl *ClusterFileBasedRepository) FindByIds(id []int) ([]Cluster, error) {

	var clusterEntities []ClusterEntity
	result := impl.dbConnection.
		Where("id in ?", id).
		Where("active = ?", true).
		Find(&clusterEntities)
	err := result.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = pg.ErrNoRows
	}
	if err != nil {
		impl.logger.Errorw("error occurred while finding all cluster data", "err", err)
		return nil, err
	}
	clusters := impl.ConvertEntitiesToModel(clusterEntities)
	return clusters, nil
}

func (impl *ClusterFileBasedRepository) Update(model *Cluster) error {
	err, entity := impl.convertToEntity(model)
	if err != nil {
		impl.logger.Errorw("error occurred while converting model to entity", "error", err)
		return errors.New("failed to update cluster")
	}
	result := impl.dbConnection.Model(entity).Updates(entity)
	err = result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while updating cluster", "error", err)
		return errors.New("failed to update cluster")
	}
	return nil
	//return impl.dbConnection.Update(model)
}

func (impl *ClusterFileBasedRepository) Delete(model *Cluster) error {
	err, entity := impl.convertToEntity(model)
	if err != nil {
		impl.logger.Errorw("error occurred while converting model to entity", "error", err)
		return errors.New("failed to delete cluster")
	}
	result := impl.dbConnection.Delete(entity)
	err = result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while deleting cluster", "err", err)
		return errors.New("failed to delete cluster")
	}
	return nil
}

func (impl *ClusterFileBasedRepository) MarkClusterDeleted(model *Cluster) error {
	model.Active = false
	return impl.Update(model)
}

func (impl *ClusterFileBasedRepository) UpdateClusterConnectionStatus(clusterId int, errorInConnecting string) error {

	result := impl.dbConnection.Model(&ClusterEntity{}).Where("id = ?", clusterId).Update("error_in_connecting = ?", errorInConnecting)
	err := result.Error
	if err != nil {
		impl.logger.Errorw("error occurred while updating cluster connection status", "clusterId", clusterId, "error", errorInConnecting, "err", err)
		return errors.New("failed to update cluster status")
	}
	return nil
}
