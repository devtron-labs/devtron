package types

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/scoop/types"
	"time"
)

type InterceptedEventQueryParams struct {
	Offset                  int
	Size                    int
	SortOrder               string
	SearchString            string
	From                    time.Time
	To                      time.Time
	Watchers                []string
	ClusterIds              []int
	ClusterIdNamespacePairs []*repository.ClusterNamespacePair
	ExecutionStatus         []string
	Actions                 []types.EventType
}

func (params InterceptedEventQueryParams) GetClusterNsPairsQuery() string {
	query := ""
	n := len(params.ClusterIdNamespacePairs)
	for i, pair := range params.ClusterIdNamespacePairs {
		query += fmt.Sprintf("(%d,'%s')", pair.ClusterId, pair.NamespaceName)
		if i < n-1 {
			query += ","
		}
	}

	return query
}

type WatcherQueryParams struct {
	Offset      int    `json:"offset"`
	Search      string `json:"search"`
	Size        int    `json:"size"`
	SortOrder   string `json:"sortOrder"`
	SortOrderBy string `json:"sortOrderBy"`
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
type InterceptedEventData struct {
	InterceptedEventId int         `sql:"intercepted_event_id"`
	ClusterId          int         `sql:"cluster_id"`
	ClusterName        string      `sql:"cluster_name"`
	Namespace          string      `sql:"namespace"`
	Action             string      `sql:"action"`
	Environment        string      `sql:"environment"`
	Metadata           string      `sql:"metadata"`
	InvolvedObjects    string      `sql:"involved_objects"`
	InterceptedAt      time.Time   `sql:"intercepted_at"`
	TriggerExecutionId int         `sql:"trigger_execution_id"`
	Status             Status      `sql:"status"`
	ExecutionMessage   string      `sql:"execution_message"`
	WatcherName        string      `sql:"watcher_name"`
	TriggerId          int         `sql:"trigger_id,pk"`
	TriggerType        TriggerType `sql:"trigger_type"`
	WatcherId          int         `sql:"watcher_id"`
	TriggerData        string      `sql:"trigger_data"`
	TotalCount         int         `sql:"total_count"`
}
