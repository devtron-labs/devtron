package service

import (
	"strings"
)

const (
	DaemonSetPodDeleteError     = "cannot delete DaemonSet-managed Pods"
	DnsLookupNoSuchHostError    = "no such host"
	TimeoutError                = "timeout"
	NotFound                    = "not found"
	ConnectionRefused           = "connection refused"
	AppNotLinkedToChartStoreErr = "app not yet linked with any chart from chart-store in devtron, hence can't perform update operation from devtron ui"
)

func IsClusterUnReachableError(err error) bool {
	if strings.Contains(err.Error(), DnsLookupNoSuchHostError) || strings.Contains(err.Error(), TimeoutError) ||
		strings.Contains(err.Error(), ConnectionRefused) {
		return true
	}
	return false
}

func IsNodeNotFoundError(err error) bool {
	if strings.Contains(err.Error(), NotFound) {
		return true
	}
	return false
}

func IsDaemonSetPodDeleteError(err error) bool {
	if strings.Contains(err.Error(), DaemonSetPodDeleteError) {
		return true
	}
	return false
}
