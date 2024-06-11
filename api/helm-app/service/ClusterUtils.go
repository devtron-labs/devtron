/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"strings"
)

const (
	DaemonSetPodDeleteError  = "cannot delete DaemonSet-managed Pods"
	DnsLookupNoSuchHostError = "no such host"
	TimeoutError             = "timeout"
	NotFound                 = "not found"
	ConnectionRefused        = "connection refused"
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
