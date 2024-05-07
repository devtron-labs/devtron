package types

import (
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
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
}
type WatcherQueryParams struct {
	Offset      int    `json:"offset"`
	Search      string `json:"search"`
	Size        int    `json:"size"`
	SortOrder   string `json:"sortOrder"`
	SortOrderBy string `json:"sortOrderBy"`
}
