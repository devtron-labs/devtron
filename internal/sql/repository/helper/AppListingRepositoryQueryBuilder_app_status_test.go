package helper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// When HIBERNATING is requested, Helm pipelines - whose status isn't persisted in app_status and is
// instead resolved on the fly at listing time - must not be excluded outright by the db-side predicate
// just because they have no app_status row yet (aps.status IS NULL). See
// AppListingServiceImpl.reconcileAppsWithUnresolvedStatus for how such rows are rechecked after fetch.
func TestBuildAppListingWhereCondition_HibernatingLetsUnresolvedHelmStatusThrough(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())

	whereClause, queryParams, err := queryBuilder.buildAppListingWhereCondition(AppListingFilter{
		AppStatuses: []string{"HIBERNATING"},
	})
	require.NoError(t, err)

	require.Contains(t, whereClause, "aps.status IN (?) or (p.deployment_app_type = ? and aps.status IS NULL)")
	require.Len(t, queryParams, 4)
	require.Equal(t, true, queryParams[0])
	require.Equal(t, CustomApp, queryParams[1])
	require.Equal(t, "helm", queryParams[3])
}

// A status filter that does not include HIBERNATING keeps matching purely against the persisted
// app_status column, as before - unresolved (NULL) rows are not specially let through for it.
func TestBuildAppListingWhereCondition_NonHibernatingStatusOnlyMatchesPersistedStatus(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())

	whereClause, queryParams, err := queryBuilder.buildAppListingWhereCondition(AppListingFilter{
		AppStatuses: []string{"Healthy"},
	})
	require.NoError(t, err)

	require.Contains(t, whereClause, "aps.status IN (?)")
	require.NotContains(t, whereClause, "deployment_app_type")
	require.Len(t, queryParams, 3)
}

// HIBERNATING combined with NOT DEPLOYED still lets unresolved Helm rows through in the "or aps.status
// IN (...)" branch of the NOT DEPLOYED condition.
func TestBuildAppListingWhereCondition_HibernatingWithNotDeployed(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())

	whereClause, queryParams, err := queryBuilder.buildAppListingWhereCondition(AppListingFilter{
		AppStatuses: []string{"HIBERNATING", "NOT DEPLOYED"},
	})
	require.NoError(t, err)

	require.Contains(t, whereClause, "(aps.status IN (?) or (p.deployment_app_type = ? and aps.status IS NULL))")
	require.Len(t, queryParams, 7)
	require.Equal(t, "helm", queryParams[6])
}
