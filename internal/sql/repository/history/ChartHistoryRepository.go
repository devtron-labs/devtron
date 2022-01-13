package history

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ChartHistoryRepository interface {
	CreateHistory(chart *ChartsHistory) (*ChartsHistory, error)
	UpdateHistory(chart *ChartsHistory) (*ChartsHistory, error)
	CreateHistoryWithTxn(chart *ChartsHistory, tx *pg.Tx) (*ChartsHistory, error)
}

type ChartHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewChartHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *ChartHistoryRepositoryImpl {
	return &ChartHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type ChartsHistory struct {
	tableName               struct{}  `sql:"charts_history" pg:",discard_unknown_columns"`
	Id                      int       `sql:"id,pk"`
	PipelineId              int       `sql:"pipeline_id, notnull"`
	ImageDescriptorTemplate string    `sql:"image_descriptor_template"`
	Template                string    `sql:"template"`
	TargetEnvironment       int       `sql:"target_environment"`
	Deployed                bool      `sql:"deployed"`
	DeployedOn              time.Time `sql:"deployed_on"`
	DeployedBy              int32     `sql:"deployed_by"`
	sql.AuditLog
}

func (impl ChartHistoryRepositoryImpl) CreateHistory(chart *ChartsHistory) (*ChartsHistory, error) {
	err := impl.dbConnection.Insert(chart)
	if err != nil {
		impl.logger.Errorw("err in creating chart history entry", "err", err, "history", chart)
		return chart, err
	}
	return chart, nil
}
func (impl ChartHistoryRepositoryImpl) UpdateHistory(chart *ChartsHistory) (*ChartsHistory, error) {
	err := impl.dbConnection.Update(chart)
	if err != nil {
		impl.logger.Errorw("err in updating chart history entry", "err", err)
		return chart, err
	}
	return chart, nil
}

func (impl ChartHistoryRepositoryImpl) CreateHistoryWithTxn(chart *ChartsHistory, tx *pg.Tx) (*ChartsHistory, error) {
	err := tx.Insert(chart)
	if err != nil {
		impl.logger.Errorw("err in creating chart history entry", "err", err, "history", chart)
		return chart, err
	}
	return chart, nil
}
