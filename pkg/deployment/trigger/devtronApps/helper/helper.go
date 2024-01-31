package helper

import (
	errors3 "errors"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	errors2 "github.com/juju/errors"
	"time"
)

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
