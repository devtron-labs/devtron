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

package helper

import (
	errors3 "errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	errors2 "github.com/juju/errors"
	"time"
)

func GetValuesFileForEnv(environmentId int) string {
	return fmt.Sprintf("_%d-values.yaml", environmentId) //-{envId}-values.yaml
}

func GetTriggerEvent(deploymentAppType string, triggeredAt time.Time, deployedBy int32) bean.TriggerEvent {
	// trigger event will decide whether to perform GitOps or deployment for a particular deployment app type
	triggerEvent := bean.TriggerEvent{
		TriggeredBy: deployedBy,
		TriggerdAt:  triggeredAt,
	}
	switch deploymentAppType {
	case bean.ArgoCd:
		triggerEvent.PerformChartPush = true
		triggerEvent.PerformDeploymentOnCluster = true
		triggerEvent.GetManifestInResponse = false
		triggerEvent.DeploymentAppType = bean.ArgoCd
		triggerEvent.ManifestStorageType = bean2.ManifestStorageGit
	case bean.Helm:
		triggerEvent.PerformChartPush = false
		triggerEvent.PerformDeploymentOnCluster = true
		triggerEvent.GetManifestInResponse = false
		triggerEvent.DeploymentAppType = bean.Helm
	}
	return triggerEvent
}

func ValidateTriggerEvent(triggerEvent bean.TriggerEvent) (bool, error) {
	switch triggerEvent.DeploymentAppType {
	case bean.ArgoCd:
		if !triggerEvent.PerformChartPush {
			return false, errors2.New("For deployment type ArgoCd, PerformChartPush flag expected value = true, got false")
		}
	case bean.Helm:
		return true, nil
	case bean.GitOpsWithoutDeployment:
		if triggerEvent.PerformDeploymentOnCluster {
			return false, errors2.New("For deployment type GitOpsWithoutDeployment, PerformDeploymentOnCluster flag expected value = false, got value = true")
		}
	case bean.ManifestDownload:
		if triggerEvent.PerformChartPush {
			return false, errors3.New("For deployment type ManifestDownload,  PerformChartPush flag expected value = false, got true")
		}
		if triggerEvent.PerformDeploymentOnCluster {
			return false, errors3.New("For deployment type ManifestDownload,  PerformDeploymentOnCluster flag expected value = false, got true")
		}
	}
	return true, nil
}
