/*
 * Copyright (c) 2024. Devtron Inc.
 */

package appStoreDiscoverRepository

import (
	"errors"
	"net"
	"testing"

	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFindWithFilter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		filter   appStoreBean.AppStoreFilter
		contains []string
		excludes string
	}{
		{
			name: "all repositories",
			contains: []string{
				"ch.name as chart_name, das.id as docker_artifact_store_id",
				"LEFT JOIN chart_repo ch", "LEFT JOIN docker_artifact_store das",
			},
			excludes: "AND ch.id IN",
		},
		{
			name:   "chart repositories",
			filter: appStoreBean.AppStoreFilter{ChartRepoId: []int{1, 2}, Size: 10, Offset: 20},
			contains: []string{
				"ch.name as chart_name", "AND ch.id IN (1,2)", "OFFSET 20 LIMIT 10",
			},
			excludes: "LEFT JOIN docker_artifact_store das",
		},
		{
			name:     "registries",
			filter:   appStoreBean.AppStoreFilter{RegistryId: []string{"registry-1"}},
			contains: []string{"das.id as docker_artifact_store_id", "AND das.id IN ('registry-1')"},
			excludes: "LEFT JOIN chart_repo ch",
		},
		{
			name:   "chart repositories and registries",
			filter: appStoreBean.AppStoreFilter{ChartRepoId: []int{1, 2}, RegistryId: []string{"registry-1"}},
			contains: []string{
				"ch.name as chart_name, das.id as docker_artifact_store_id",
				"AND (ch.id IN (1,2) OR das.id IN ('registry-1'))",
			},
			excludes: "OFFSET",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			queryError := errors.New("query transport unavailable")
			db := pg.Connect(&pg.Options{
				Dialer: func(_, _ string) (net.Conn, error) { return nil, queryError },
			})
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			var queries []string
			var formattingError error
			db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
				query, err := event.FormattedQuery()
				formattingError = err
				queries = append(queries, query)
			})
			repository := NewAppStoreApplicationVersionRepositoryImpl(zap.NewNop().Sugar(), db)

			versions, err := repository.FindWithFilter(&tc.filter)

			require.ErrorIs(t, err, queryError)
			require.Nil(t, versions)
			require.NoError(t, formattingError)
			require.Len(t, queries, 1)
			for _, fragment := range tc.contains {
				require.Contains(t, queries[0], fragment)
			}
			require.NotContains(t, queries[0], tc.excludes)
		})
	}
}
