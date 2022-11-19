# Deployment

This chart creates a deployment that runs multiple replicas of your application and automatically replaces any instances that fail or become unresponsive. It does not support Blue/Green and Canary deployments.


This is the default deployment chart. You can select `Deployment` chart when you want to use only basic use cases which contain the following:

* Create a Deployment to rollout a ReplicaSet. The ReplicaSet creates Pods in the background. Check the status of the rollout to see if it succeeds or not.
* Declare the new state of the Pods. A new ReplicaSet is created and the Deployment manages moving the Pods from the old ReplicaSet to the new one at a controlled rate. Each new ReplicaSet updates the revision of the Deployment.
* Rollback to an earlier Deployment revision if the current state of the Deployment is not stable. Each rollback updates the revision of the Deployment.
* Scale up the Deployment to facilitate more load.
* Use the status of the Deployment as an indicator that a rollout has stuck.
* Clean up older ReplicaSets that you do not need anymore.


You can define application behavior by providing information in the following sections:

| Key | Descriptions |
| :--- | :--- |
| `Chart version` | Select the Chart Version using which you want to deploy the application.<br> Refer [Chart Version](https://docs.devtron.ai/v/v0.5/usage/applications/creating-application/deployment-template/rollout-deployment#1.-chart-version) section for more detail.</br> |
| `Basic Configuration` | You can select the basic deployment configuration for your application on the **Basic** GUI section instead of configuring the YAML file.<br>Refer [Basic Configuration](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/rollout-deployment#2.-basic-configuration) section for more detail.</br>|
| `Advanced (YAML)` | If you want to do additional configurations, then click **Advanced (YAML)** for modifications.<br>Refer [Advanced (YAML)](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/rollout-deployment#3.-advanced-yaml) section for more detail.</br> |
| `Show application metrics` | You can enable `Show application metrics` to see your application's metrics-CPU Service Monitor usage, Memory Usage, Status, Throughput and Latency.<br>Refer [Application Metrics](https://docs.devtron.ai/v/v0.5/usage/applications/app-details/app-metrics) for more detail.</br> |


