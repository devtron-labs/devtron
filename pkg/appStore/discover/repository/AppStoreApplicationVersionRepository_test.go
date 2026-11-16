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
			filter: appStoreBean.AppStoreFilter{ChartRepoId: []int{1, 2}},
			contains: []string{
				"ch.name as chart_name", "AND ch.id IN (1,2)",
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
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			queryError := errors.New("query transport unavailable")
			dialCalls := 0
			db := pg.Connect(&pg.Options{
				Dialer: func(_, _ string) (net.Conn, error) {
					dialCalls++
					return nil, queryError
				},
			})
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			columns, _ := updateFindWithFilterQuery(&tc.filter, QUERY_COLUMN_UPDATE)
			join, params := updateFindWithFilterQuery(&tc.filter, QUERY_JOIN_UPDTAE)
			query := string(db.FormatQuery(nil, columns+join, params...))
			repository := NewAppStoreApplicationVersionRepositoryImpl(zap.NewNop().Sugar(), db)

			versions, err := repository.FindWithFilter(&tc.filter)

			require.ErrorIs(t, err, queryError)
			require.Nil(t, versions)
			require.Equal(t, 1, dialCalls)
			for _, fragment := range tc.contains {
				require.Contains(t, query, fragment)
			}
			if tc.excludes != "" {
				require.NotContains(t, query, tc.excludes)
			}
		})
	}
}
