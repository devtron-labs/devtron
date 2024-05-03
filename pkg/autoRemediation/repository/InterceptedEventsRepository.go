package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

type InterceptedEventExecution struct {
	tableName          struct{}        `sql:"intercepted_event_execution" pg:",discard_unknown_columns"`
	Id                 int             `sql:"id,pk"`
	ClusterId          int             `sql:"cluster_id"`
	Namespace          string          `sql:"namespace"`
	Action             watch.EventType `sql:"action"`
	InvolvedObject     string          `sql:"involved_object"`
	Gvk                string          `sql:"gvk"`
	InterceptedAt      time.Time       `sql:"intercepted_at"`
	TriggerId          int             `sql:"trigger_id"`
	TriggerExecutionId int             `sql:"trigger_execution_id"`
	Status             Status          `sql:"status"`
	ExecutionMessage   string          `sql:"execution_message"`
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
	FindAllInterceptedEvents(interceptedEventsQueryParams *InterceptedEventQuery) ([]*InterceptedEventData, int, error)
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
	err := tx.Insert(&interceptedEvents)
	if err != nil {
		return interceptedEvents, err
	}
	return interceptedEvents, nil
}

type InterceptedEventData struct {
	ClusterId          int         `sql:"cluster_id"`
	Namespace          string      `sql:"namespace"`
	Action             string      `sql:"action"`
	Environment        string      `sql:"environment"`
	Gvk                string      `sql:"gvk"`
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

type InterceptedEventQuery struct {
	Offset          int       `json:"offset"`
	Size            int       `json:"size"`
	SortOrder       string    `json:"sortOrder"`
	SearchString    string    `json:"searchString"`
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
	Watchers        []string  `json:"watchers"`
	ClusterIds      []int     `json:"clusters"`
	Namespaces      []string  `json:"namespaces"`
	ExecutionStatus []string  `json:"execution_status"`
}

func (impl InterceptedEventsRepositoryImpl) FindAllInterceptedEvents(interceptedEventsQueryParams *InterceptedEventQuery) ([]*InterceptedEventData, int, error) {

	var interceptedEvents []*InterceptedEventData
	query := impl.dbConnection.Model().
		Table("intercepted_event_execution").
		ColumnExpr("intercepted_event_execution.cluster_id as cluster_id").
		ColumnExpr("intercepted_event_execution.namespace as namespace").
		// ColumnExpr("intercepted_event_execution.message as message").
		ColumnExpr("intercepted_event_execution.action as action").
		ColumnExpr("intercepted_event_execution.gvk as gvk").
		ColumnExpr("intercepted_event_execution.involved_object as involved_object").
		ColumnExpr("intercepted_event_execution.intercepted_at as intercepted_at").
		ColumnExpr("intercepted_event_execution.trigger_execution_id as trigger_execution_id").
		ColumnExpr("intercepted_event_execution.status as status").
		ColumnExpr("intercepted_event_execution.execution_message as execution_message").
		ColumnExpr("environment.environment_name as environment").
		ColumnExpr("k8s_event_watcher.name as watcher_name").
		ColumnExpr("auto_remediation_trigger.id as trigger_id").
		ColumnExpr("auto_remediation_trigger.type as trigger_type").
		ColumnExpr("auto_remediation_trigger.data as trigger_data").
		Join("JOIN auto_remediation_trigger ON intercepted_event_execution.trigger_id = auto_remediation_trigger.id").
		Join("JOIN k8s_event_watcher ON auto_remediation_trigger.watcher_id = k8s_event_watcher.id").
		Join("JOIN environment ON environment.cluster_id = intercepted_event_execution.cluster_id").
		Where("environment.cluster_id=intercepted_event_execution.cluster_id and environment.namespace = intercepted_event_execution.namespace")

	if !interceptedEventsQueryParams.From.IsZero() && !interceptedEventsQueryParams.To.IsZero() {
		query = query.Where("intercepted_event_execution.intercepted_at BETWEEN ? AND ?", interceptedEventsQueryParams.From, interceptedEventsQueryParams.To)
	}

	if interceptedEventsQueryParams.SearchString != "" {
		query = query.Where("intercepted_event_execution.gvk ILIKE ? OR intercepted_event_execution.involved_object ILIKE ?", "%"+interceptedEventsQueryParams.SearchString+"%", "%"+interceptedEventsQueryParams.SearchString+"%")
	}

	if len(interceptedEventsQueryParams.ClusterIds) > 0 {
		query = query.Where("intercepted_event_execution.cluster_id IN (?)", pg.In(interceptedEventsQueryParams.ClusterIds))
	}

	if len(interceptedEventsQueryParams.Namespaces) > 0 {
		query = query.Where("intercepted_event_execution.namespace IN (?)", pg.In(interceptedEventsQueryParams.Namespaces))
	}

	if len(interceptedEventsQueryParams.Watchers) > 0 {
		query = query.Where("k8s_event_watcher.name IN (?)", pg.In(interceptedEventsQueryParams.Watchers))
	}

	if len(interceptedEventsQueryParams.ExecutionStatus) > 0 {
		query = query.Where("intercepted_event_execution.status IN (?)", pg.In(interceptedEventsQueryParams.ExecutionStatus))
	}
	if interceptedEventsQueryParams.SortOrder == "asc" {
		query = query.Order("intercepted_event_execution.intercepted_at asc")
	} else {
		query = query.Order("intercepted_event_execution.intercepted_at desc")
	}
	// Count total number of intercepted events
	total, err := query.Count()
	if err != nil {
		return interceptedEvents, 0, err
	}

	err = query.
		Offset(interceptedEventsQueryParams.Offset).
		Limit(interceptedEventsQueryParams.Size).
		Select(&interceptedEvents)
	return interceptedEvents, total, err
}

func (impl InterceptedEventsRepositoryImpl) GetInterceptedEventsByTriggerIds(triggerIds []int) ([]*InterceptedEventExecution, error) {
	var interceptedEvents []*InterceptedEventExecution
	err := impl.dbConnection.Model(&interceptedEvents).Where("trigger_id IN (?)", pg.In(triggerIds)).Select()
	if err != nil {
		return nil, err
	}
	return interceptedEvents, nil
}
