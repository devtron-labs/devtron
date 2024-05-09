package repository

import (
	"fmt"
	types2 "github.com/devtron-labs/devtron/pkg/autoRemediation/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/scoop/types"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"time"
)

type InterceptedEventExecution struct {
	tableName          struct{}        `sql:"intercepted_event_execution" pg:",discard_unknown_columns"`
	Id                 int             `sql:"id,pk"`
	ClusterId          int             `sql:"cluster_id"`
	Namespace          string          `sql:"namespace"`
	Action             types.EventType `sql:"action"`
	InvolvedObjects    string          `sql:"involved_objects"`
	Metadata           string          `sql:"metadata"`
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
	FindAllInterceptedEvents(interceptedEventsQueryParams *types2.InterceptedEventQueryParams) ([]*types2.InterceptedEventData, int, error)
	GetInterceptedEventsByTriggerIds(triggerIds []int) ([]*InterceptedEventExecution, error)
	GetInterceptedEventById(id int) (*types2.InterceptedEventData, error)
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
	if len(interceptedEvents) == 0 {
		return nil, nil
	}
	err := tx.Insert(&interceptedEvents)
	if err != nil {
		return interceptedEvents, err
	}
	return interceptedEvents, nil
}

func (impl InterceptedEventsRepositoryImpl) FindAllInterceptedEvents(interceptedEventsQueryParams *types2.InterceptedEventQueryParams) ([]*types2.InterceptedEventData, int, error) {

	var interceptedEvents []*types2.InterceptedEventData

	query := impl.buildInterceptEventsListingQuery(interceptedEventsQueryParams)

	err := query.
		Offset(interceptedEventsQueryParams.Offset).
		Limit(interceptedEventsQueryParams.Size).
		Select(&interceptedEvents)

	total := 0
	if len(interceptedEvents) > 0 {
		total = interceptedEvents[0].TotalCount
	}
	return interceptedEvents, total, err
}

func (impl InterceptedEventsRepositoryImpl) GetInterceptedEventsByTriggerIds(triggerIds []int) ([]*InterceptedEventExecution, error) {
	var interceptedEvents []*InterceptedEventExecution
	if len(triggerIds) == 0 {
		return nil, nil
	}
	err := impl.dbConnection.Model(&interceptedEvents).Where("trigger_id IN (?)", pg.In(triggerIds)).Select()
	if err != nil {
		return nil, err
	}
	return interceptedEvents, nil
}

func (impl InterceptedEventsRepositoryImpl) buildInterceptEventsListingQuery(interceptedEventsQueryParams *types2.InterceptedEventQueryParams) *orm.Query {
	query := impl.dbConnection.Model().
		Table("intercepted_event_execution").
		Column("intercepted_event_execution.cluster_id").
		Column("intercepted_event_execution.namespace").
		Column("intercepted_event_execution.action").
		Column("intercepted_event_execution.metadata").
		Column("intercepted_event_execution.involved_objects").
		Column("intercepted_event_execution.intercepted_at").
		Column("intercepted_event_execution.trigger_execution_id").
		Column("intercepted_event_execution.status").
		Column("intercepted_event_execution.execution_message").
		ColumnExpr("intercepted_event_execution.id as intercepted_event_id").
		// ColumnExpr("intercepted_event_execution.message as message").
		ColumnExpr("environment.environment_name as environment").
		ColumnExpr("k8s_event_watcher.name as watcher_name").
		ColumnExpr("auto_remediation_trigger.id as trigger_id").
		ColumnExpr("auto_remediation_trigger.type as trigger_type").
		ColumnExpr("auto_remediation_trigger.data as trigger_data").
		ColumnExpr("COUNT(*) OVER() as total_count").
		Join("INNER JOIN auto_remediation_trigger ON intercepted_event_execution.trigger_id = auto_remediation_trigger.id").
		Join("INNER JOIN k8s_event_watcher ON auto_remediation_trigger.watcher_id = k8s_event_watcher.id").
		Join("INNER JOIN environment ON environment.cluster_id = intercepted_event_execution.cluster_id").
		Where("environment.cluster_id=intercepted_event_execution.cluster_id and environment.namespace = intercepted_event_execution.namespace")

	if !interceptedEventsQueryParams.From.IsZero() && !interceptedEventsQueryParams.To.IsZero() {
		query = query.Where("intercepted_event_execution.intercepted_at BETWEEN ? AND ?", interceptedEventsQueryParams.From, interceptedEventsQueryParams.To)
	}

	if interceptedEventsQueryParams.SearchString != "" {
		query = query.Where("intercepted_event_execution.metadata ILIKE ?", "%"+interceptedEventsQueryParams.SearchString+"%")
	}

	if len(interceptedEventsQueryParams.ClusterIds) > 0 || len(interceptedEventsQueryParams.ClusterIdNamespacePairs) > 0 {
		query = query.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			if len(interceptedEventsQueryParams.ClusterIds) > 0 {
				q = q.WhereOr("intercepted_event_execution.cluster_id IN (?)", pg.In(interceptedEventsQueryParams.ClusterIds))
			}

			if len(interceptedEventsQueryParams.ClusterIdNamespacePairs) > 0 {
				clusterNamespaceGroupQuery := fmt.Sprintf("(intercepted_event_execution.cluster_id,intercepted_event_execution.namespace) IN (%s)", interceptedEventsQueryParams.GetClusterNsPairsQuery())
				q = q.WhereOr(clusterNamespaceGroupQuery)
			}
			return q, nil
		})
	}

	if len(interceptedEventsQueryParams.Watchers) > 0 {
		query = query.Where("k8s_event_watcher.name IN (?)", pg.In(interceptedEventsQueryParams.Watchers))
	}

	if len(interceptedEventsQueryParams.ExecutionStatus) > 0 {
		query = query.Where("intercepted_event_execution.status IN (?)", pg.In(interceptedEventsQueryParams.ExecutionStatus))
	}

	if len(interceptedEventsQueryParams.Actions) > 0 {
		query = query.Where("intercepted_event_execution.action IN (?)", pg.In(interceptedEventsQueryParams.Actions))
	}

	if interceptedEventsQueryParams.SortOrder == "asc" {
		query = query.Order("intercepted_event_execution.intercepted_at asc")
	} else {
		query = query.Order("intercepted_event_execution.intercepted_at desc")
	}

	return query
}
func (impl InterceptedEventsRepositoryImpl) GetInterceptedEventById(id int) (*types2.InterceptedEventData, error) {
	var interceptedEvents = types2.InterceptedEventData{}
	err := impl.dbConnection.Model().Table("intercepted_event_execution").
		Column("intercepted_event_execution.cluster_id").
		Column("intercepted_event_execution.namespace").
		Column("intercepted_event_execution.action").
		Column("intercepted_event_execution.metadata").
		Column("intercepted_event_execution.involved_objects").
		Column("intercepted_event_execution.intercepted_at").
		Column("intercepted_event_execution.trigger_execution_id").
		Column("intercepted_event_execution.status").
		Column("intercepted_event_execution.execution_message").
		ColumnExpr("cluster.cluster_name as cluster_name").
		ColumnExpr("intercepted_event_execution.id as intercepted_event_id").
		ColumnExpr("auto_remediation_trigger.id as trigger_id").
		Join("INNER JOIN auto_remediation_trigger ON intercepted_event_execution.trigger_id = auto_remediation_trigger.id").
		Join("INNER JOIN cluster ON cluster.id = intercepted_event_execution.cluster_id").
		Where("intercepted_event_execution.id = ?", id).
		Select(&interceptedEvents)
	if err != nil {
		return nil, err
	}
	return &interceptedEvents, nil
}
