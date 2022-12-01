package chartRepoRepository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type RefChartDir string
type ChartRef struct {
	tableName                  struct{} `sql:"chart_ref" pg:",discard_unknown_columns"`
	Id                         int      `sql:"id,pk"`
	Location                   string   `sql:"location"`
	Version                    string   `sql:"version"`
	Active                     bool     `sql:"active,notnull"`
	Default                    bool     `sql:"is_default,notnull"`
	Name                       string   `sql:"name"`
	ChartData                  []byte   `sql:"chart_data"`
	ChartDescription           string   `sql:"chart_description"`
	UserUploaded               bool     `sql:"user_uploaded,notnull"`
	IsAppMetricsSupported      bool     `sql:"is_app_metrics_supported,notnull"`
	FilePathContainingStrategy string   `sql:"file_path_containing_strategy"`
	JsonPathForStrategy        string   `sql:"json_path_for_strategy"`
	sql.AuditLog
}

type ChartRefMetaData struct {
	tableName        struct{} `sql:"chart_ref_metadata" pg:",discard_unknown_columns"`
	ChartName        string   `sql:"chart_name,pk"`
	ChartDescription string   `sql:"chart_description"`
}

type ChartRefRepository interface {
	Save(chartRepo *ChartRef) error
	GetDefault() (*ChartRef, error)
	FindById(id int) (*ChartRef, error)
	GetAll() ([]*ChartRef, error)
	GetAllChartMetadata() ([]*ChartRefMetaData, error)
	FindByVersionAndName(name, version string) (*ChartRef, error)
	CheckIfDataExists(name string, version string) (bool, error)
	FetchChart(name string) ([]*ChartRef, error)
	FetchChartInfoByUploadFlag(userUploaded bool) ([]*ChartRef, error)
}
type ChartRefRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewChartRefRepositoryImpl(dbConnection *pg.DB) *ChartRefRepositoryImpl {
	return &ChartRefRepositoryImpl{
		dbConnection: dbConnection,
	}
}

func (impl ChartRefRepositoryImpl) Save(chartRepo *ChartRef) error {
	return impl.dbConnection.Insert(chartRepo)
}

func (impl ChartRefRepositoryImpl) GetDefault() (*ChartRef, error) {
	repo := &ChartRef{}
	err := impl.dbConnection.Model(repo).
		Where("is_default = ?", true).
		Where("active = ?", true).Select()
	return repo, err
}

func (impl ChartRefRepositoryImpl) FindById(id int) (*ChartRef, error) {
	repo := &ChartRef{}
	err := impl.dbConnection.Model(repo).
		Where("id = ?", id).
		Where("active = ?", true).Select()
	return repo, err
}
func (impl ChartRefRepositoryImpl) FindByVersionAndName(name, version string) (*ChartRef, error) {
	repo := &ChartRef{}
	var err error
	if len(name) > 0 {
		err = impl.dbConnection.Model(repo).
			Where("name = ?", name).
			Where("version= ?", version).
			Where("active = ?", true).Select()
	} else {
		err = impl.dbConnection.Model(repo).
			Where("name is NULL", name).
			Where("version= ?", version).
			Where("active = ?", true).Select()
	}
	return repo, err
}

func (impl ChartRefRepositoryImpl) GetAll() ([]*ChartRef, error) {
	var chartRefs []*ChartRef
	err := impl.dbConnection.Model(&chartRefs).
		Where("active = ?", true).Select()
	return chartRefs, err
}

func (impl ChartRefRepositoryImpl) GetAllChartMetadata() ([]*ChartRefMetaData, error) {
	var chartRefMetaDatas []*ChartRefMetaData
	err := impl.dbConnection.Model(&chartRefMetaDatas).Select()
	return chartRefMetaDatas, err
}

func (impl ChartRefRepositoryImpl) CheckIfDataExists(name string, version string) (bool, error) {
	repo := &ChartRef{}
	return impl.dbConnection.Model(repo).
		Where("lower(name) = ?", strings.ToLower(name)).
		Where("version = ? ", version).Exists()
}

func (impl ChartRefRepositoryImpl) FetchChart(name string) ([]*ChartRef, error) {
	var chartRefs []*ChartRef
	err := impl.dbConnection.Model(&chartRefs).Where("lower(name) = ?", strings.ToLower(name)).Select()
	if err != nil {
		return nil, err
	}
	return chartRefs, err
}

func (impl ChartRefRepositoryImpl) FetchChartInfoByUploadFlag(userUploaded bool) ([]*ChartRef, error) {
	var repo []*ChartRef
	err := impl.dbConnection.Model(&repo).
		Where("user_uploaded = ?", userUploaded).
		Where("active = ?", true).Select()
	if err != nil {
		return repo, err
	}
	return repo, err
}

// pipeline strategy metadata repository starts here
type DeploymentStrategy string

const (
	DEPLOYMENT_STRATEGY_BLUE_GREEN DeploymentStrategy = "BLUE-GREEN"
	DEPLOYMENT_STRATEGY_ROLLING    DeploymentStrategy = "ROLLING"
	DEPLOYMENT_STRATEGY_CANARY     DeploymentStrategy = "CANARY"
	DEPLOYMENT_STRATEGY_RECREATE   DeploymentStrategy = "RECREATE"
)

type GlobalStrategyMetadata struct {
	tableName   struct{}           `sql:"global_strategy_metadata" pg:",discard_unknown_columns"`
	Id          int                `sql:"id,pk"`
	Name        DeploymentStrategy `sql:"name"`
	Description string             `sql:"description"`
	Deleted     bool               `sql:"deleted,notnull"`
	sql.AuditLog
}

type GlobalStrategyMetadataRepository interface {
	GetByChartRefId(chartRefId int) ([]*GlobalStrategyMetadata, error)
}
type GlobalStrategyMetadataRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewGlobalStrategyMetadataRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *GlobalStrategyMetadataRepositoryImpl {
	return &GlobalStrategyMetadataRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *GlobalStrategyMetadataRepositoryImpl) GetByChartRefId(chartRefId int) ([]*GlobalStrategyMetadata, error) {
	var globalStrategies []*GlobalStrategyMetadata
	err := impl.dbConnection.Model(&globalStrategies).
		Join("INNER JOIN global_strategy_metadata_chart_ref_mapping as gsmcrm on gsmcrm.global_strategy_metadata_id=global_strategy_metadata.id").
		Where("gsmcrm.chart_ref_id = ?", chartRefId).
		Where("gsmcrm.active = ?", true).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("error in getting global strategies metadata by chartRefId", "err", err, "chartRefId", chartRefId)
		return nil, err
	}
	return globalStrategies, err
}

// pipeline strategy metadata and chart_ref mapping repository starts here
type GlobalStrategyMetadataChartRefMapping struct {
	tableName                struct{} `sql:"global_strategy_metadata_chart_ref_mapping" pg:",discard_unknown_columns"`
	Id                       int      `sql:"id,pk"`
	GlobalStrategyMetadataId string   `sql:"global_strategy_metadata_id"`
	ChartRefId               string   `sql:"chart_ref_id"`
	Active                   bool     `sql:"active,notnull"`
	sql.AuditLog
}

type GlobalStrategyMetadataChartRefMappingRepository interface {
}
type GlobalStrategyMetadataChartRefMappingRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewGlobalStrategyMetadataChartRefMappingRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *GlobalStrategyMetadataChartRefMappingRepositoryImpl {
	return &GlobalStrategyMetadataChartRefMappingRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}
