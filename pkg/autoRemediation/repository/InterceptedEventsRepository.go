package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type InterceptedEventExecution struct {
	tableName          struct{}  `sql:"intercepted_event_execution" pg:",discard_unknown_columns"`
	Id                 int       `sql:"id,pk"`
	ClusterId          int       `sql:"cluster_id"`
	Namespace          string    `sql:"namespace"`
	Message            string    `sql:"message"`
	MessageType        string    `sql:"message_type"`
	Event              string    `sql:"event"`
	InvolvedObject     string    `sql:"involved_object"`
	InterceptedAt      time.Time `sql:"intercepted_at"`
	TriggerId          int       `sql:"trigger_id"`
	TriggerExecutionId int       `sql:"trigger_execution_id"`
	Status             Status    `sql:"status"`
	ExecutionMessage   string    `sql:"execution_message"`
	sql.AuditLog
}
type Status string

const (
	Failure     Status = "Failure"
	Success     Status = "Success"
	Progressing Status = "Progressing"
	Errored     Status = "Error"
)

type InterceptedEventsRepository interface {
	Save(interceptedEvents []*InterceptedEventExecution, tx *pg.Tx) ([]*InterceptedEventExecution, error)
	GetAllInterceptedEvents() ([]*InterceptedEventExecution, error)
	// UpdateStatus(status string, interceptedEventId int) error
	FindAllInterceptedEvents(interceptedEventsQueryParams *InterceptedEventQueryParams) ([]*InterceptedEventData, error)
	GetInterceptedEventsByTriggerIds(triggerIds []int) ([]*InterceptedEventExecution, error)
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

func (impl InterceptedEventsRepositoryImpl) Save(interceptedEvents []*InterceptedEventExecution, tx *pg.Tx) ([]*InterceptedEventExecution, error) {
	err := tx.Insert(interceptedEvents)
	if err != nil {
		return interceptedEvents, err
	}
	return interceptedEvents, nil
}

func (impl InterceptedEventsRepositoryImpl) GetAllInterceptedEvents() ([]*InterceptedEventExecution, error) {
	var interceptedEvents []*InterceptedEventExecution
	err := impl.dbConnection.Model(&interceptedEvents).
		Select()
	if err != nil {
		return nil, err
	}
	return interceptedEvents, nil
}

//	func (impl InterceptedEventsRepositoryImpl) UpdateStatus(status string, interceptedEventId int)  error {
//		_, err := impl.dbConnection.Model(&InterceptedEvents{}).Where("id=?", interceptedEventId).Set("status=?", status).Update()
//		if err != nil {
//			return err
//		}
//		return  nil
//
// }
type InterceptedEventData struct {
	ClusterId          int         `sql:"cluster_id"`
	Namespace          string      `sql:"namespace"`
	Message            string      `sql:"message"`
	MessageType        string      `sql:"message_type"`
	Event              string      `sql:"event"`
	InvolvedObject     string      `sql:"involved_object"`
	InterceptedAt      time.Time   `sql:"intercepted_at"`
	TriggerExecutionId int         `sql:"trigger_execution_id"`
	Status             Status      `sql:"status"`
	ExecutionMessage   string      `sql:"execution_message"`
	WatcherName        string      `sql:"watcher_name"`
	TriggerId          int         `sql:"trigger_id,pk"`
	TriggerType        TriggerType `sql:"trigger_type"`
	WatcherId          int         `sql:"watcher_id"`
	TriggerData        string      `sql:"trigger_data"`
}
type InterceptedEventQueryParams struct {
	Offset          int       `json:"offset"`
	Size            int       `json:"size"`
	SortOrder       string    `json:"sortOrder"`
	SearchString    string    `json:"searchString"`
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
	Watchers        []string  `json:"watchers"`
	Clusters        []string  `json:"clusters"`
	Namespaces      []string  `json:"namespaces"`
	ExecutionStatus []string  `json:"execution_status"`
}

func (impl InterceptedEventsRepositoryImpl) FindAllInterceptedEvents(interceptedEventsQueryParams *InterceptedEventQueryParams) ([]*InterceptedEventData, error) {

	var interceptedEvents []*InterceptedEventData

	query := impl.dbConnection.Model().
		Table("intercepted_event_execution").
		ColumnExpr("intercepted_event_execution.cluster_id as cluster_id").
		ColumnExpr("intercepted_event_execution.namespace as namespace").
		ColumnExpr("intercepted_event_execution.message as message").
		ColumnExpr("intercepted_event_execution.message_type as message_type").
		ColumnExpr("intercepted_event_execution.event as event").
		ColumnExpr("intercepted_event_execution.involved_object as involved_object").
		ColumnExpr("intercepted_event_execution.intercepted_at as intercepted_at").
		ColumnExpr("intercepted_event_execution.trigger_execution_id as trigger_execution_id").
		ColumnExpr("intercepted_event_execution.status as status").
		ColumnExpr("intercepted_event_execution.execution_message as execution_message").
		ColumnExpr("watcher.name as watcher_name").
		ColumnExpr("trigger.id as trigger_id").
		ColumnExpr("trigger.type as trigger_type").
		ColumnExpr("trigger.data as trigger_data").
		Join("JOIN trigger ON intercepted_event_execution.trigger_id = trigger.id").
		Join("JOIN watcher ON trigger.watcher_id = watcher.id")

	if !interceptedEventsQueryParams.From.IsZero() && !interceptedEventsQueryParams.To.IsZero() {
		query = query.Where("intercepted_event_execution.intercepted_at BETWEEN ? AND ?", interceptedEventsQueryParams.From, interceptedEventsQueryParams.To)
	}

	if interceptedEventsQueryParams.SearchString != "" {
		query = query.Where("intercepted_event_execution.message ILIKE ? OR intercepted_event_execution.involved_object ILIKE ?", "%"+interceptedEventsQueryParams.SearchString+"%", "%"+interceptedEventsQueryParams.SearchString+"%")
	}

	if len(interceptedEventsQueryParams.Clusters) > 0 {
		query = query.Where("intercepted_event_execution.cluster_id IN (?)", pg.In(interceptedEventsQueryParams.Clusters))
	}

	if len(interceptedEventsQueryParams.Namespaces) > 0 {
		query = query.Where("intercepted_event_execution.namespace IN (?)", pg.In(interceptedEventsQueryParams.Namespaces))
	}

	if len(interceptedEventsQueryParams.Watchers) > 0 {
		query = query.Where("watcher.name IN (?)", pg.In(interceptedEventsQueryParams.Watchers))
	}

	if len(interceptedEventsQueryParams.ExecutionStatus) > 0 {
		query = query.Where("intercepted_event_execution.status IN (?)", pg.In(interceptedEventsQueryParams.ExecutionStatus))
	}
	if interceptedEventsQueryParams.SortOrder == "desc" {
		query = query.Order("intercepted_event_execution.intercepted_at desc")
	} else {
		query = query.Order("intercepted_event_execution.intercepted_at asc")
	}
	err := query.
		Offset(interceptedEventsQueryParams.Offset).
		Limit(interceptedEventsQueryParams.Size).
		Select(&interceptedEvents)
	return interceptedEvents, err
}

func (impl InterceptedEventsRepositoryImpl) GetInterceptedEventsByTriggerIds(triggerIds []int) ([]*InterceptedEventExecution, error) {
	var interceptedEvents []*InterceptedEventExecution
	err := impl.dbConnection.Model(&interceptedEvents).Where("trigger_id IN (?)", pg.In(triggerIds)).Select()
	if err != nil {
		return nil, err
	}
	return interceptedEvents, nil
}
