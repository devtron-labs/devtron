package app

import (
	"testing"

	"github.com/devtron-labs/devtron/api/bean/AppView"
	"github.com/stretchr/testify/require"
)

func TestIsFilteredOnHibernatingStatus(t *testing.T) {
	impl := AppListingServiceImpl{}

	require.True(t, impl.isFilteredOnHibernatingStatus(FetchAppListingRequest{AppStatuses: []string{"Healthy", "HIBERNATING"}}))
	require.False(t, impl.isFilteredOnHibernatingStatus(FetchAppListingRequest{AppStatuses: []string{"Healthy"}}))
	require.False(t, impl.isFilteredOnHibernatingStatus(FetchAppListingRequest{}))
}

// A container whose status was already persisted in db (and therefore already correctly matched and
// counted by the sql query) must never be touched by the in-memory reconciliation pass, even if its
// resolved AppStatus happens not to be an exact string match (defensive: containersWithUnresolvedStatus
// is the only signal used to decide whether a row is eligible for the recheck).
func TestReconcileAppsWithUnresolvedStatus_LeavesResolvedContainersUntouched(t *testing.T) {
	impl := AppListingServiceImpl{}
	dbBacked := &AppView.AppEnvironmentContainer{AppId: 1, AppStatus: "HIBERNATING"}

	envContainers := []*AppView.AppEnvironmentContainer{dbBacked}
	result, appSize := impl.reconcileAppsWithUnresolvedStatus(envContainers, 10, nil, []string{"HIBERNATING"})

	require.Equal(t, envContainers, result)
	require.Equal(t, 10, appSize)
}

// A Helm-type row with no persisted app_status (let through by the db query as an unresolved row) is
// dropped, and appSize is decremented by exactly the number of rows dropped, once its on-the-fly status
// turns out not to match the requested filter.
func TestReconcileAppsWithUnresolvedStatus_DropsNonMatchingUnresolvedContainers(t *testing.T) {
	impl := AppListingServiceImpl{}
	matching := &AppView.AppEnvironmentContainer{AppId: 1, AppStatus: "HIBERNATING"}        // resolved on the fly, matches
	nonMatching := &AppView.AppEnvironmentContainer{AppId: 2, AppStatus: "Healthy"}         // resolved on the fly, doesn't match
	unresolvedStillEmpty := &AppView.AppEnvironmentContainer{AppId: 3, AppStatus: ""}       // on-the-fly fetch was a no-op / found nothing
	dbBackedArgoApp := &AppView.AppEnvironmentContainer{AppId: 4, AppStatus: "HIBERNATING"} // already matched by sql

	envContainers := []*AppView.AppEnvironmentContainer{matching, nonMatching, unresolvedStillEmpty, dbBackedArgoApp}
	containersWithUnresolvedStatus := map[*AppView.AppEnvironmentContainer]bool{
		matching:             true,
		nonMatching:          true,
		unresolvedStillEmpty: true,
	}

	result, appSize := impl.reconcileAppsWithUnresolvedStatus(envContainers, 4, containersWithUnresolvedStatus, []string{"HIBERNATING"})

	require.Equal(t, []*AppView.AppEnvironmentContainer{matching, dbBackedArgoApp}, result)
	require.Equal(t, 2, appSize)
}

func TestReconcileAppsWithUnresolvedStatus_NoUnresolvedContainersIsNoOp(t *testing.T) {
	impl := AppListingServiceImpl{}
	container := &AppView.AppEnvironmentContainer{AppId: 1, AppStatus: "HIBERNATING"}
	envContainers := []*AppView.AppEnvironmentContainer{container}

	result, appSize := impl.reconcileAppsWithUnresolvedStatus(envContainers, 1, map[*AppView.AppEnvironmentContainer]bool{}, []string{"HIBERNATING"})

	require.Equal(t, envContainers, result)
	require.Equal(t, 1, appSize)
}
