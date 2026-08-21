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

package rbac

import (
	"fmt"

	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
)

// EnforcerUtilGitOps builds RBAC objects for external Argo CD and Flux CD applications.
//
// The object is two segments — <clusterName>__<namespace>/<appName> — where the first segment
// is the environment_identifier convention used throughout Devtron. There is no project
// segment: external GitOps applications have no Devtron project, so unlike external Helm apps
// there is no team name and no "unassigned" placeholder to fill.
//
// Both segments are always populated, which matters because matchKeyByPart rejects an empty
// segment on either side unconditionally.
type EnforcerUtilGitOps interface {
	// GetExternalGitOpsAppObject returns the RBAC object for an external Argo/Flux application.
	// Returns an empty string if the cluster cannot be resolved, which callers must treat as
	// a denial rather than as a wildcard.
	GetExternalGitOpsAppObject(clusterId int, namespace string, appName string) string
	// GetExternalGitOpsAppObjectByClusterName is the same, for callers that already hold the
	// cluster name and can avoid the lookup.
	GetExternalGitOpsAppObjectByClusterName(clusterName string, namespace string, appName string) string
}

type EnforcerUtilGitOpsImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
}

func NewEnforcerUtilGitOpsImpl(logger *zap.SugaredLogger,
	clusterRepository repository.ClusterRepository) *EnforcerUtilGitOpsImpl {
	return &EnforcerUtilGitOpsImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
	}
}

func (impl EnforcerUtilGitOpsImpl) GetExternalGitOpsAppObject(clusterId int, namespace string, appName string) string {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error on fetching cluster for rbac object", "err", err, "clusterId", clusterId)
		return ""
	}
	return impl.GetExternalGitOpsAppObjectByClusterName(cluster.ClusterName, namespace, appName)
}

func (impl EnforcerUtilGitOpsImpl) GetExternalGitOpsAppObjectByClusterName(clusterName string, namespace string, appName string) string {
	if len(clusterName) == 0 || len(namespace) == 0 || len(appName) == 0 {
		impl.logger.Errorw("incomplete identifier for rbac object, denying",
			"clusterName", clusterName, "namespace", namespace, "appName", appName)
		return ""
	}
	return fmt.Sprintf("%s__%s/%s", clusterName, namespace, appName)
}
