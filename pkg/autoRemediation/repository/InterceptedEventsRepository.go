package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type InterceptedEvents struct {
	tableName          struct{}  `sql:"intercepted_events" pg:",discard_unknown_columns"`
	Id                 int       `sql:"id,pk"`
	ClusterName        string    `sql:"cluster_name"`
	Namespace          string    `sql:"namespace"`
	Message            string    `sql:"message"`
	Event              string    `sql:"event"`
	InvolvedObject     string    `sql:"involved_object"`
	InterceptedAt      time.Time `sql:"intercepted_at"`
	TriggerId          int       `sql:"trigger_id"`
	TriggerExecutionId int       `sql:"trigger_execution_id"`
	Status             Status    `sql:"status"`
}
type Status string

const (
	Failure     Status = "Failure"
	Success     Status = "Success"
	Progressing Status = "Progressing"
)

type InterceptedEventsRepository interface {
	Save(interceptedEvents *InterceptedEvents, tx *pg.Tx) (*InterceptedEvents, error)
	GetAllInterceptedEvents() ([]*InterceptedEvents, error)
	//UpdateStatus(status string, interceptedEventId int) error
	sql.TransactionWrapper
}
type InterceptedEventsRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
	*sql.TransactionUtilImpl
}

func NewInterceptedEventsRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *InterceptedEventsRepositoryImpl {
	TransactionUtilImpl := sql.NewTransactionUtilImpl(dbConnection)
	return &InterceptedEventsRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: TransactionUtilImpl,
	}
}

func (impl InterceptedEventsRepositoryImpl) Save(interceptedEvents *InterceptedEvents, tx *pg.Tx) (*InterceptedEvents, error) {
	_, err := tx.Model(interceptedEvents).Insert()
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return interceptedEvents, nil
}
func (impl InterceptedEventsRepositoryImpl) GetAllInterceptedEvents() ([]*InterceptedEvents, error) {
	var interceptedEvents []*InterceptedEvents
	err := impl.dbConnection.Model(&interceptedEvents).
		Select()
	if err != nil {
		return nil, err
	}
	return interceptedEvents, nil
}

//func (impl InterceptedEventsRepositoryImpl) UpdateStatus(status string, interceptedEventId int)  error {
//	_, err := impl.dbConnection.Model(&InterceptedEvents{}).Where("id=?", interceptedEventId).Set("status=?", status).Update()
//	if err != nil {
//		return err
//	}
//	return  nil
//
//}
