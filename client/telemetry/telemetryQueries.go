package telemetry

// EA-mode compatible telemetry queries - no imports needed for current methods

func (impl *TelemetryEventClientImpl) getHelmAppCount() int {
	if impl.installedAppReadService == nil {
		impl.logger.Warnw("installedAppReadService not available for helm app count")
		return -1
	}
	count, err := impl.installedAppReadService.GetActiveInstalledAppCount()
	if err != nil {
		impl.logger.Errorw("error getting helm app count", "err", err)
		return -1
	}
	return count
}

// EA-mode compatible telemetry methods

func (impl *TelemetryEventClientImpl) getClusterCounts() (physicalCount int, isolatedCount int) {
	clusters, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error getting cluster counts", "err", err)
		return -1, -1
	}

	physicalCount = 0
	isolatedCount = 0

	for _, cluster := range clusters {
		if cluster.IsVirtualCluster {
			isolatedCount++
		} else {
			physicalCount++
		}
	}

	return physicalCount, isolatedCount
}

// Note: FULL-mode specific methods like getDevtronAppCount, getJobCount, etc.
// are now implemented in TelemetryEventClientImplExtended in telemetryQueriesExtended.go

func (impl *TelemetryEventClientImpl) getActiveUsersLast30Days() int {
	if impl.userAuditService == nil {
		impl.logger.Warnw("userAuditService not available for active users count")
		return -1
	}

	count, err := impl.userAuditService.GetActiveUsersCountInLast30Days()
	if err != nil {
		impl.logger.Errorw("error getting active users count in last 30 days", "err", err)
		return -1
	}

	impl.logger.Debugw("counted active users in last 30 days", "count", count)
	return count
}
